# Nooks

A household todo app you run on your own machine.

One list primitive, one action to add something, and a printed page that is a real deliverable rather
than a fallback. No email server to configure, no telemetry, no account anywhere but yours.

> Nooks is in early development. Nothing here is releasable yet — see [`TODO.md`](TODO.md) for where it
> is up to.

## Running it

Nooks is one binary. It serves the API and the app from the same process and keeps everything in one
SQLite file you can copy.

```bash
go build -o nooks ./cmd/nooks
./nooks --data ./data
```

Then open <http://localhost:8081>. The first person to arrive creates their account and names the
Instance; everyone else joins by asking, and an Admin approves.

### Options

| Flag | Environment | Default | Meaning |
| --- | --- | --- | --- |
| `--addr` | `NOOKS_ADDR` | `:8081` | Host and port to listen on |
| `--data` | `NOOKS_DATA` | `./data` | Directory holding the SQLite file |
| `--driver` | `NOOKS_DRIVER` | `sqlite` | `sqlite` or `postgres` |
| `--dsn` | `NOOKS_DSN` | — | Postgres connection string; unused for SQLite |
| `--mode` | `NOOKS_MODE` | `prod` | `dev` expects the Vite server to serve the app |

### Backing it up

Stop Nooks and copy `data/nooks.db`. That is the whole Instance.

## Developing

Requires Go, Node and pnpm. `pnpm install` brings its own pinned `buf` — there is nothing to install
globally.

```bash
pnpm install                        # dependencies, including buf

go run ./cmd/nooks --mode dev        # API on :8081
pnpm --filter @nooks/web dev         # app on :3001, proxying to the API

go test ./...                       # Go tests
pnpm --filter @nooks/web lint        # typecheck and lint
pnpm generate                       # regenerate Go and TypeScript from proto/
```

The store suite covers both drivers. Postgres cases skip unless you point them at a server:

```bash
docker run -d --name nooks-pg -e POSTGRES_USER=nooks -e POSTGRES_PASSWORD=nooks \
  -e POSTGRES_DB=nooks -p 55432:5432 postgres:17-alpine
NOOKS_TEST_POSTGRES_DSN='postgres://nooks:nooks@localhost:55432/nooks?sslmode=disable' go test ./store/
```

To bake the app into the binary:

```bash
pnpm --filter @nooks/web release     # builds into server/router/frontend/dist
go build -o nooks ./cmd/nooks         # go:embed picks it up
```

## Before you contribute

Four documents, each with one job. They are short, and reading them first will save you a review
round.

| File | Job |
| --- | --- |
| [`CONTEXT.md`](CONTEXT.md) | The language. Every term has one name and a list of what not to call it |
| [`DESIGN.md`](DESIGN.md) | The design system. Exact values, not approximate |
| [`STANDARDS.md`](STANDARDS.md) | Code standards |
| [`TODO.md`](TODO.md) | The plan and the current status |

## Licence

[AGPL-3.0](LICENSE).
