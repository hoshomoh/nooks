#!/usr/bin/env bash
#
# Every RPC must have a gateway adapter, or its REST route answers "unimplemented".
#
# The adapters in server/router/gateway/services.go each embed an
# UnimplementedXServiceServer, so an RPC with no adapter compiles and fails only when
# somebody calls it over REST. The app never does — it speaks Connect — so the gap
# shows up for REST clients and nobody else. This is what the compiler does not do.
set -euo pipefail

cd "$(dirname "$0")/.."

adapters="server/router/gateway/services.go"
missing=""

for proto in proto/nooks/api/v1/*_service.proto; do
  # Every rpc the file declares, and the service it belongs to.
  while read -r rpc; do
    [ -n "$rpc" ] || continue
    if ! grep -q "^func (g [a-zA-Z]*) $rpc(" "$adapters"; then
      missing="$missing  $rpc  (from $(basename "$proto"))
"
    fi
  done < <(grep -oE '^[[:space:]]*rpc [A-Za-z]+' "$proto" | awk '{print $2}')
done

if [ -n "$missing" ]; then
  echo "these RPCs have no gateway adapter, so REST answers unimplemented:" >&2
  printf '%s' "$missing" >&2
  echo "add one to $adapters" >&2
  exit 1
fi

echo "every RPC has a gateway adapter"
