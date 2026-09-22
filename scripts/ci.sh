#!/usr/bin/env bash
# Everything .github/workflows/ci.yml runs, in the same order.
#
# Run this before pushing. CI failing after a push is a slower, noisier way to learn
# the same thing.
#
#   ./scripts/ci.sh              every group, which is what CI does
#   ./scripts/ci.sh go web       only those, for when the rest cannot be affected
#
# The groups are the workflow's jobs. Naming a subset is ./scripts/preflight.sh's doing,
# not something to reach for by hand: CI runs the lot either way, and the only thing a
# narrower local run saves is your own minutes.
set -euo pipefail

export PATH="/usr/local/go/bin:$PATH"
export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
# shellcheck disable=SC1091
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"

ALL_GROUPS="go proto shared web website"
groups="${*:-$ALL_GROUPS}"

for name in $groups; do
  case " $ALL_GROUPS " in
    *" $name "*) ;;
    *) echo "ci: no such group: $name (have: $ALL_GROUPS)" >&2; exit 2 ;;
  esac
done

wants() {
  case " $groups " in *" $1 "*) return 0 ;; *) return 1 ;; esac
}

# Each step says how long it took, so a slow run can be pointed at rather than guessed at.
started=$(date +%s)
step_at=$started
step_name=""
step() {
  now=$(date +%s)
  if [ -n "$step_name" ]; then
    printf '\033[2m   %s: %ss\033[0m\n' "$step_name" "$((now - step_at))"
  fi
  step_name="$1"
  step_at=$now
  printf '\n\033[1m→ %s\033[0m\n' "$1"
}

step "install (frozen lockfile, as CI does)"
pnpm install --frozen-lockfile

if wants go; then
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
fi

if wants proto; then
  step "buf lint"
  pnpm proto:lint

  step "buf format"
  if [ -n "$(cd proto && pnpm exec buf format -d)" ]; then
    echo "proto files are not formatted; run pnpm proto:format"; exit 1
  fi

  # Two halves of one rule: anything the app can do, a script must be able to do. The
  # annotation is what gives an RPC a REST route at all; the adapter is what answers on
  # it. An RPC can have either without the other, and then the REST API is a version
  # behind with nothing saying so.
  step "rest bindings"
  ./scripts/check-annotations.sh
  ./scripts/check-rpc-comments.sh

  step "gateway adapters"
  ./scripts/check-gateway.sh

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
fi

if wants shared; then
  step "shared package lint and test"
  pnpm --filter @nooks/shared lint
  pnpm --filter @nooks/shared test
fi

if wants web; then
  step "web lint and typecheck"
  pnpm --filter @nooks/web lint

  step "web test"
  pnpm --filter @nooks/web test
fi

if wants website; then
  step "website typecheck and build"
  pnpm --filter @nooks/website build

  step "website test and links"
  pnpm --filter @nooks/website test
  pnpm --filter @nooks/website check-links
fi

if wants web; then
  step "web build, landing where go:embed reads"
  # One build, not two: `release` is the same Vite build as `build` with its output
  # pointed at the binary, so running both proves the same thing twice.
  pnpm --filter @nooks/web release
  test -d server/router/frontend/dist/assets
  test ! -e apps/server

  step "what a first visit costs"
  pnpm --filter @nooks/web check-bundle ../../server/router/frontend/dist/assets
fi

printf '\033[2m   %s: %ss\033[0m\n' "$step_name" "$(($(date +%s) - step_at))"

if [ "$groups" = "$ALL_GROUPS" ]; then
  # Named rather than implied. Two CI jobs cannot run here: the image needs Docker, and
  # the dependency scan needs the network and a tool this script has no business
  # installing. A banner that said "everything" would be wrong about both.
  printf '\n\033[1;32m✓ every check that runs here passes in %ss\033[0m\n' "$(($(date +%s) - started))"
  printf '\033[2m  CI also builds the image and scans dependencies.\033[0m\n'
else
  printf '\n\033[1;32m✓ %s passes in %ss\033[0m\n' "$groups" "$(($(date +%s) - started))"
  printf '\033[2m  CI still runs the rest.\033[0m\n'
fi
