#!/usr/bin/env bash
# The mapping in groups.sh, checked.
#
# It decides what ./scripts/preflight.sh is allowed to skip, so a mistake here is a
# green local run and a red CI — exactly what preflight exists to prevent. Every case
# below is a path shape somebody will actually commit.
set -euo pipefail

self="$(cd "$(dirname "$0")" && pwd)/$(basename "$0")"
cd "$(dirname "$0")"
# shellcheck source=/dev/null
. ./groups.sh

failures=0

# expect names the groups that should come back for the given paths.
expect() {
  local want="$1"; shift
  local got
  got=$(groupsFor "$@")
  if [ "$got" != "$want" ]; then
    printf '  \033[31m✗\033[0m %s\n      want: %s\n      got:  %s\n' "$*" "$want" "$got"
    failures=$((failures + 1))
  fi
}

# One app, one group.
expect "web" "apps/web/src/screens/list.tsx"
expect "website" "apps/website/content/docs/mcp.mdx"
expect "go" "server/router/api/v1/list.go"
expect "go" "store/item.go"
expect "go" "cmd/nooks/main.go"
expect "go" "internal/password/password.go"
expect "go" "go.mod"

# The contract reaches everything generated from it.
expect "go proto shared web website" "proto/nooks/api/v1/list_service.proto"

# Shared code reaches whatever compiles it. The Go server does not.
expect "shared web website" "packages/shared/src/mcp.ts"
expect "shared web website" "packages/api/src/index.ts"
expect "web website" "packages/design/foundations.css"

# The changelog page reads this file at build time, so it is not just prose.
expect "website" "CHANGELOG.md"

# Prose no build step reads.
expect "" "README.md"
expect "" "DESIGN.md" "STANDARDS.md"

# Anything that decides what the checks are, or that nothing above claimed.
expect "go proto shared web website" "pnpm-lock.yaml"
expect "go proto shared web website" ".github/workflows/ci.yml"
expect "go proto shared web website" "scripts/ci.sh"
expect "go proto shared web website" "Dockerfile"
expect "go proto shared web website" "some/new/place.txt"

# Several paths at once: the union, deduplicated, in ci.sh's order.
expect "go web" "server/router/api/v1/item.go" "apps/web/src/screens/list.tsx"
expect "web" "apps/web/a.tsx" "apps/web/b.tsx"
expect "go proto shared web website" "README.md" "pnpm-lock.yaml"

# Nothing changed is not the same as everything changed.
expect ""

if [ "$failures" -gt 0 ]; then
  printf '\n\033[31m%s case(s) wrong in groups.sh\033[0m\n' "$failures"
  exit 1
fi
printf '\033[2m   group mapping: %s cases\033[0m\n' "$(grep -c '^expect ' "$self")"
