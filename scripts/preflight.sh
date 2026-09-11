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

# So a caller who pipes this to `tail` still sees a failure.
set -o pipefail

cd "$(dirname "$0")/.."
root=$(pwd)
worktree="$root/.preflight"

if [ -n "$(git status --porcelain)" ]; then
  echo "preflight: commit first — this checks HEAD, not the working tree" >&2
  git status --short >&2
  exit 1
fi

# A linked worktree's .git is a file, not a directory, so test for either — getting
# this wrong meant the script tried to create a worktree that was already there, failed,
# and only looked like it had run.
# Resolved here, in the repository being pushed. Inside the worktree "HEAD" means the
# worktree's own HEAD, so checking out HEAD there is a no-op that silently verifies
# whatever it happened to be on last time.
target=$(git rev-parse HEAD)

if [ -e "$worktree/.git" ]; then
  git -C "$worktree" checkout -q --detach "$target"
else
  git worktree add --detach "$worktree" "$target"
fi

echo "preflight: checking $(git rev-parse --short "$target") in a clean checkout"
(cd "$worktree" && ./scripts/ci.sh)

printf '\n\033[1;32m✓ %s is what it says it is\033[0m\n' "$(git rev-parse --short "$target")"
