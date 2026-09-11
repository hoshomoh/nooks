#!/usr/bin/env bash
# Run CI against exactly what is committed, in a clean checkout.
#
# ./scripts/ci.sh checks the working tree, which is not what gets pushed. A fix made
# after CI passed, or a file never added, both leave a green local run and a red one on
# GitHub — which is how b13ab29 went out with a type error in it.
#
# This checks out HEAD somewhere else, installs from the lockfile, and runs the same
# script there. If it passes, what you are about to push passes.
set -euo pipefail

cd "$(dirname "$0")/.."

if [ -n "$(git status --porcelain)" ]; then
  echo "preflight: commit first — this checks HEAD, not the working tree" >&2
  git status --short >&2
  exit 1
fi

worktree=$(mktemp -d)
cleanup() { git worktree remove --force "$worktree" >/dev/null 2>&1 || true; }
trap cleanup EXIT

git worktree add -q --detach "$worktree" HEAD
echo "preflight: checking $(git rev-parse --short HEAD) in a clean checkout"

(cd "$worktree" && ./scripts/ci.sh)

printf '\n\033[1;32m✓ %s is what it says it is\033[0m\n' "$(git rev-parse --short HEAD)"
