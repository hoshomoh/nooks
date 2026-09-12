# syntax=docker/dockerfile:1
# Nooks is one binary. This builds it, and the image around it holds nothing else.
#
# Three stages: the app, the binary with the app baked into it, and a runtime with
# neither toolchain in it. The result is a static binary and a CA bundle — no shell, no
# package manager, and nothing to keep patched on somebody's machine in a hallway.

# The app. Pinned to the same Node the repository pins, so a container build and a
# contributor's build are the same build.
#
# Built on whatever architecture is doing the building, never emulated: the output is
# JavaScript, which is the same bytes either way, and emulating a Node build to produce
# identical files is minutes spent for nothing.
FROM --platform=$BUILDPLATFORM node:24.21.0-alpine AS app
WORKDIR /src

RUN corepack enable

# The manifests alone first, so a change to the source does not re-resolve every
# dependency. Every workspace package needs its manifest for the lockfile to install.
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/web/package.json apps/web/
COPY apps/website/package.json apps/website/
COPY packages/api/package.json packages/api/
COPY packages/design/package.json packages/design/

# The app and what it depends on, and explicitly not the workspace root: its one
# devDependency is buf, a 38MB proto compiler, and the protos are generated in the
# repository rather than here. An image that downloads a compiler it never runs is an
# image that fails to build on a slow connection for no reason.
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile --filter "@nooks/web..." --filter "!nooks"

COPY . .
RUN pnpm --filter @nooks/web release

# The binary, with the app inside it.
#
# Also on the building architecture. Go cross-compiles, so an arm64 image is built by a
# native toolchain told to emit arm64 rather than by an emulated one — the difference
# between a minute and most of an hour on a release that builds both.
FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
# The app built above, landing where go:embed reads it.
COPY --from=app /src/server/router/frontend/dist ./server/router/frontend/dist

# Static: the SQLite driver is pure Go, so nothing here wants a C library and the
# runtime stage needs no libc at all. Symbols and DWARF stripped — a stack trace from a
# stripped binary is still a stack trace, and this halves the image.
ARG VERSION=dev
ARG COMMIT=""
# Set by buildx for whatever platform is being produced. Defaulted so that a plain
# `docker build` with no buildx still works and simply builds for the host.
ARG TARGETOS=linux
ARG TARGETARCH
# The module and build caches are mounted rather than copied, so a rebuild compiles
# what changed instead of the whole dependency tree. Without them every build is a cold
# one, which on a small machine is the difference between a minute and ten.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -trimpath \
    -ldflags="-s -w \
      -X github.com/hoshomoh/nooks/internal/version.Version=${VERSION} \
      -X github.com/hoshomoh/nooks/internal/version.Commit=${COMMIT}" \
    -o /out/nooks ./cmd/nooks

# An empty directory to carry across. VOLUME on a path that does not exist creates it
# at container start owned by root, and the process does not run as root — so the very
# first thing it does is fail to open its own database. Made here because the runtime
# image has no shell to make it with.
RUN mkdir -p /out/data

# What ships.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build --chown=65532:65532 /out/nooks /usr/local/bin/nooks

# One directory, and everything is in it. Copy it and you have copied the instance.
#
# Copied in already owned by the runtime user rather than declared and left to Docker.
# The uid is written as a number: --chown resolves a name against the target image's
# passwd file, and depending on distroless shipping one is a dependency worth not having.
COPY --from=build --chown=65532:65532 /out/data /var/lib/nooks
VOLUME ["/var/lib/nooks"]
ENV NOOKS_DATA=/var/lib/nooks
EXPOSE 8081

USER nonroot
ENTRYPOINT ["/usr/local/bin/nooks"]
