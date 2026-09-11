#!/usr/bin/env bash
#
# Every RPC must be reachable over REST.
#
# The rule is the product's, not the compiler's: anything a Member can do in the app, a
# script or an alternative client must be able to do too. Nothing enforces that except
# this, so a new RPC that ships UI-only stops the build rather than quietly leaving the
# REST API a version behind.
set -euo pipefail

cd "$(dirname "$0")/.."

# EXEMPT lists RPCs that deliberately have no HTTP binding, each with the reason.
# Keep it short: an entry here is a hole in the rule above.
EXEMPT=(
  # These four hand over a session, which means a Set-Cookie the gateway adapter cannot
  # carry — it returns the message and nothing else. A REST client reaches the same
  # ground a different way: it signs in over Connect once, or it is handed an Access
  # token, and then it holds the short-lived access token these return in their body.
  "SignIn"
  "CompleteSetup"
  "CompleteJoin"
  "CompletePasswordReset"
  # Signing out clears a cookie, which is a browser's business. A REST client stops
  # using its access token, and the token expires within the hour either way.
  "SignOut"
)

missing=()

for proto in proto/nooks/api/v1/*.proto; do
  # An RPC is annotated when google.api.http appears before the next rpc or the end of
  # the service block. awk walks the file once and remembers the last rpc it saw.
  while IFS= read -r rpc; do
    exempt=false
    for name in "${EXEMPT[@]}"; do
      [ "$rpc" = "$name" ] && exempt=true && break
    done
    $exempt || missing+=("$(basename "$proto"): $rpc")
  done < <(awk '
    /^[[:space:]]*rpc [A-Za-z]+/ {
      if (pending != "") { print pending }
      match($0, /rpc [A-Za-z]+/)
      pending = substr($0, RSTART + 4, RLENGTH - 4)
      next
    }
    /google\.api\.http/ { pending = "" }
    END { if (pending != "") print pending }
  ' "$proto")
done

if [ ${#missing[@]} -gt 0 ]; then
  printf '\033[1;31m✗ these RPCs have no google.api.http annotation:\033[0m\n' >&2
  printf '  %s\n' "${missing[@]}" >&2
  printf '\nAnything the app can do, a REST client must be able to do.\n' >&2
  printf 'Add the annotation, or add the RPC to EXEMPT in %s with a reason.\n' "$0" >&2
  exit 1
fi
