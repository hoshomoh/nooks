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

step "web lint and typecheck"
pnpm --filter @nooks/web lint

step "web test"
pnpm --filter @nooks/web test

step "web build"
pnpm --filter @nooks/web build

step "release lands where go:embed reads"
pnpm --filter @nooks/web release
test -d server/router/frontend/dist/assets
test ! -e apps/server

printf '\n\033[1;32m✓ everything CI runs passes\033[0m\n'
