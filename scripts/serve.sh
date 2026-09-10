#!/usr/bin/env bash
# Build the app into the binary and serve it, replacing whatever holds the port.
#
# Exists because `go run` spawns a child that survives killing its parent, so a stale
# server keeps the port and the next one silently fails to bind — you then debug a
# server that is not the one you just built.
set -euo pipefail

PORT="${PORT:-8081}"
DATA="${DATA:-./data}"

for pid in $(ss -ltnpH "sport = :$PORT" 2>/dev/null | grep -oP 'pid=\K[0-9]+' | sort -u); do
  echo "stopping pid $pid on :$PORT"
  kill "$pid" 2>/dev/null || true
done

pnpm --filter @nooks/web release
go build -o ./nooks ./cmd/nooks
exec ./nooks --data "$DATA" --addr ":$PORT"
