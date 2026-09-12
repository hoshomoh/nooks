#!/usr/bin/env bash
# Everything .github/workflows/ci.yml runs, in the same order.
#
# Run this before pushing. CI failing after a push is a slower, noisier way to learn
# the same thing.
set -euo pipefail

export PATH="/usr/local/go/bin:$PATH"
export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
# shellcheck disable=SC1091
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"

step() { printf '\n\033[1m→ %s\033[0m\n' "$1"; }

step "install (frozen lockfile, as CI does)"
pnpm install --frozen-lockfile

step "go build"
go build ./...

step "go vet"
go vet ./...

step "go test"
# -race needs cgo, and a machine without a C compiler cannot run it. CI has one, so it
# runs there; here we fall back rather than skipping the tests altogether.
if [ "$(go env CGO_ENABLED)" = "1" ]; then
  go test -race ./...
else
  echo "note: no cgo on this machine, running without -race (CI still runs it)"
  go test ./...
fi

step "gofmt"
unformatted=$(gofmt -l cmd internal server store)
if [ -n "$unformatted" ]; then
  echo "not gofmt'd:"; echo "$unformatted"; exit 1
fi

step "buf lint"
pnpm proto:lint

step "buf format"
if [ -n "$(cd proto && pnpm exec buf format -d)" ]; then
  echo "proto files are not formatted; run pnpm proto:format"; exit 1
fi

# The protos are the contract every client is built against: the web app, a mobile
# client, and anything anybody writes against the REST API. Checking against main means
# a change that would break one of those stops here rather than in somebody's build.
#
# It is skipped where there is no main to compare with — a fresh clone with no remote,
# or the first commit on a new branch that has not been pushed.
step "buf breaking"
if git rev-parse --verify --quiet main >/dev/null; then
  (cd proto && pnpm exec buf breaking --against "../.git#branch=main,subdir=proto")
else
  echo "no main to compare against; skipped"
fi

step "web lint and typecheck"
pnpm --filter @nooks/web lint

step "web test"
pnpm --filter @nooks/web test

step "website typecheck and build"
pnpm --filter @nooks/website build

step "web build, landing where go:embed reads"
# One build, not two: `release` is the same Vite build as `build` with the output
# pointed at the binary, so running both proved the same thing twice and was the
# heaviest thing in this script. A project that expects to be self-hosted should not
# need a large machine to test itself.
pnpm --filter @nooks/web release
test -d server/router/frontend/dist/assets
test ! -e apps/server

printf '\n\033[1;32m✓ everything CI runs passes\033[0m\n'
