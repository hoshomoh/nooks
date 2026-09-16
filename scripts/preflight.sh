#!/usr/bin/env bash
# Run CI against exactly what is committed, in a clean checkout.
#
# ./scripts/ci.sh checks the working tree, which is not what gets pushed. A fix made
# after CI passed, or a file never added, both leave a green local run and a red one on
# GitHub. `tsc -b` is incremental too, so a local run can skip files a fresh checkout
# will not.
#
# The checkout is kept between runs rather than thrown away, so its node_modules is
# installed once instead of on every run.
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

# A linked worktree's .git is a file, not a directory, so test for either.
# Resolved here, in the repository being pushed. Inside the worktree "HEAD" is the
# worktree's own, so checking out HEAD there verifies whatever it was last on.
target=$(git rev-parse HEAD)

# --force because ci.sh builds inside this checkout, and a later checkout refuses to
# overwrite what it wrote. Nothing here is worth keeping.
if [ -e "$worktree/.git" ]; then
  git -C "$worktree" checkout -q --detach --force "$target"
else
  git worktree add --detach "$worktree" "$target"
fi

# shellcheck source=lib/groups.sh
. "$root/scripts/lib/groups.sh"

# What this push would add to the branch it is going to.
#
# CI runs every group whatever happens, so the only thing skipping costs is the chance
# to learn about a break here rather than there. It is worth taking when a group
# provably cannot be affected — and groups.sh resolves every doubt the other way.
base=$(git merge-base origin/main "$target" 2>/dev/null || true)
if [ -n "$base" ]; then
  changed=$(git diff --name-only "$base" "$target")
else
  # No remote to compare with: a fresh clone, or a branch nobody has pushed. Run the lot.
  changed=""
fi

if [ -n "$changed" ]; then
  groups=$(groupsFor $changed)
  skipped=$(for g in $ALL_GROUPS; do case " $groups " in *" $g "*) ;; *) printf '%s ' "$g" ;; esac; done)
else
  groups="$ALL_GROUPS"
  skipped=""
fi

if [ "${1:-}" = "--explain" ]; then
  echo "changed since origin/main:"
  echo "$changed" | sed 's/^/  /'
  echo "groups: ${groups:-none}"
  echo "skipped: ${skipped:-none}"
  exit 0
fi

# The mapping decides what is allowed to go unchecked, so it is checked first.
"$root/scripts/lib/groups.test.sh"

echo "preflight: checking $(git rev-parse --short "$target") in a clean checkout"
if [ -n "$skipped" ]; then
  printf '\033[2m  nothing changed under %s; CI still runs them\033[0m\n' "${skipped% }"
fi

if [ -z "$groups" ]; then
  printf '\n\033[1;32m✓ %s changes nothing any check covers\033[0m\n' "$(git rev-parse --short "$target")"
  exit 0
fi

# shellcheck disable=SC2086
(cd "$worktree" && ./scripts/ci.sh $groups)

printf '\n\033[1;32m✓ %s is what it says it is\033[0m\n' "$(git rev-parse --short "$target")"
