#!/usr/bin/env bash
#
# Every RPC's comment names the RPC it is above.
#
# The proto is the source these comments are read from: buf generates them into the Go
# stubs, the TypeScript client and openapi.yaml, and openapi.yaml is what the published
# API reference is built from. So a comment above the wrong rpc is wrong in four places
# at once, and one of them is the page a stranger reads.
#
# It happens the ordinary way: an rpc is added directly beneath a comment that already
# belonged to the one below, which inherits it and leaves the other bare. SetListArchived
# spent its life documented as "DeleteList removes a List and the Items on it", and
# DeleteList had nothing — the same slip that had already been found and fixed in the Go
# service for the same two methods.
#
# internal/conventions holds Go to this and cannot see it here: it parses Go, and skips
# the generated tree where these end up.
set -euo pipefail

cd "$(dirname "$0")/.."

wrong=()

for proto in proto/nooks/api/v1/*.proto; do
  # awk walks the file once, remembering the last comment line it saw before an rpc.
  while IFS= read -r problem; do
    wrong+=("$(basename "$proto"): $problem")
  done < <(awk '
    # A comment line: keep the first word of the first line of the block.
    /^[[:space:]]*\/\// {
      if (!inblock) { opens = $2; inblock = 1 }
      next
    }
    /^[[:space:]]*rpc [A-Za-z]+/ {
      match($0, /rpc [A-Za-z]+/)
      name = substr($0, RSTART + 4, RLENGTH - 4)
      if (!inblock) {
        print name " has no comment"
      } else if (opens != name) {
        print name " is documented as \"" opens "\""
      }
      inblock = 0
      next
    }
    # Anything else ends a comment block.
    { inblock = 0 }
  ' "$proto")
done

if [ ${#wrong[@]} -gt 0 ]; then
  printf '\033[1;31m✗ these RPC comments do not name their RPC:\033[0m\n' >&2
  printf '  %s\n' "${wrong[@]}" >&2
  printf '\nThe comment above an rpc is generated into the public API reference.\n' >&2
  exit 1
fi
