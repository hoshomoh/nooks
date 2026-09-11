#!/usr/bin/env bash
# Run CI against exactly what is committed, in a clean checkout.
#
# ./scripts/ci.sh checks the working tree, which is not what gets pushed. A fix made
# after CI passed, or a file never added, both leave a green local run and a red one on
# GitHub — which is how b13ab29 went out with a type error in it. `tsc -b` is incremental
# too, so a local run can skip files a fresh checkout will not.
#
# The checkout is kept between runs rather than thrown away, so its node_modules is
# installed once. A throwaway worktree meant a full install every time, which is slow
# and, on a machine doing anything else, enough to get the run killed.
set -euo pipefail

cd "$(dirname "$0")/.."
root=$(pwd)
worktree="$root/.preflight"

if [ -n "$(git status --porcelain)" ]; then
  echo "preflight: commit first — this checks HEAD, not the working tree" >&2
  git status --short >&2
  exit 1
fi

if [ ! -d "$worktree/.git" ]; then
  git worktree add -q --detach "$worktree" HEAD
else
  git -C "$worktree" checkout -q --detach HEAD
fi

echo "preflight: checking $(git rev-parse --short HEAD) in a clean checkout"
(cd "$worktree" && ./scripts/ci.sh)

printf '\n\033[1;32m✓ %s is what it says it is\033[0m\n' "$(git rev-parse --short HEAD)"
