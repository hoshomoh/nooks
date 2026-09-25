# nooks

A household todo app you run on your own machine.

One list primitive, one action to add something, and a printed page that is a real deliverable rather
than a fallback. No email server to configure, no telemetry, no account anywhere but yours.

**[Documentation](https://usenooks.vercel.app/docs)** — installing, configuring, the REST API and the
MCP server.

## Running it

nooks is one binary. It serves the API and the app from the same process and keeps everything in one
SQLite file you can copy.

With Docker:

```bash
docker run -d --name nooks -p 8081:8081 -v nooks-data:/var/lib/nooks ghcr.io/hoshomoh/nooks
```

Or with the `docker-compose.yml` in this repository:

```bash
docker compose up -d
```

Or from a clone. The app is baked into the binary, so it has to be built first —
`go build` alone gives you a binary that serves nothing:

```bash
pnpm install && pnpm --filter @nooks/web release && go build -o nooks ./cmd/nooks
```

Then:

```bash
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
| `--secure-cookies` | `NOOKS_SECURE_COOKIES` | off | Marks session cookies `Secure`. Turn it on behind TLS |
| `--log-level` | `NOOKS_LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |
| `--db-max-conns` | `NOOKS_DB_MAX_CONNS` | `10` | How many connections to open to Postgres at once. Unused for SQLite |

### Backing it up

Stop nooks and copy `data/nooks.db`. That is the whole Instance.

## Developing

Requires Go, Node and pnpm. `pnpm install` brings its own pinned `buf` — there is nothing to install
globally.

```bash
./scripts/ci.sh                     # every check that runs locally — do this before committing

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

[`CONTRIBUTING.md`](CONTRIBUTING.md) is how to get it running and what a change is expected to do.

Behind it are four documents, each with one job. They are short, and reading them first will save
you a review round.

| File | Job |
| --- | --- |
| [`CONTEXT.md`](CONTEXT.md) | The language. Every term has one name and a list of what not to call it |
| [`DESIGN.md`](DESIGN.md) | The design system. Exact values, not approximate |
| [`STANDARDS.md`](STANDARDS.md) | Code standards |

## Not in v1

Named so nobody wonders whether they were forgotten. None of them appear in the design,
and none of them are ruled out later:

- **Reminders and push.** Reminders are push with a time on them, so they are one
  feature rather than two. Browser push can be added on its own, and the planned mobile
  app brings the obvious way to deliver them.
- **File attachments.**
- **Sub-lists**, or nesting beyond a Note's checklists.
- **Multiple Instances behind one deployment.**

## Two standing decisions

These are not scheduling. They shape what nooks is, and changing either would change
the app rather than extend it:

- **No email, of any kind.** No mail server to configure on a machine in a hallway, and
  nothing to intercept: an account is created by an Admin handing over a temporary
  password, and a forgotten one is reset the same way. The whole joining and reset flow
  is built on its absence.
- **No remote images in a Note.** One fetched from somebody else's server would tell
  that server the Note was read, and from which address, on an app whose claim is that
  nothing leaves the machine. An image uploaded to your own Instance would not have that
  problem, and is in the first list rather than this one.

A **mobile app is planned**: Expo, in `apps/mobile`, sharing `@nooks/api` and the design
tokens. The layout already has room for it.

## Licence

[AGPL-3.0](LICENSE).
