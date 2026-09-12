# The Nooks website

The landing page and the documentation, including the API reference generated from
`proto/gen/openapi.yaml`. Next.js and Fumadocs, sharing `packages/design` with the app so a
colour changed in `DESIGN.md` changes both.

## Running it

```sh
pnpm install
pnpm --filter @nooks/website dev
```

`dev` and `build` both regenerate `content/docs/reference` from the OpenAPI spec first. Those
pages are not committed — they are derived from the protos, and committing them would be
committing the same API description twice.

## How it is deployed

Hosting is Vercel, as one project pointed at this directory. Everything that can be recorded
in the repository is in `vercel.json`; the rest is project settings, listed here because a
setting nobody wrote down is a setting nobody can check.

| Setting | Value | Why |
| --- | --- | --- |
| Root Directory | `apps/website` | The Next.js app. Vercel's Next builder expects to find `next.config.ts` at the root of the project |
| Include source files outside of the Root Directory | on | The build reads `../../proto/gen/openapi.yaml` and transpiles `@nooks/design`. On by default for projects created since 2020, but it is the one setting that silently breaks the build if it is ever turned off |
| Skip deployment (unaffected projects) | on | Free, and it uses the pnpm workspace graph rather than a guess |
| Node.js Version | from `engines.node` | Pinned to `24.x` in `package.json`, matching `.nvmrc`. The dashboard dropdown is overridden by that field |
| Install Command | Vercel's default | Left alone deliberately: Vercel detects the pnpm workspace and installs from its root with a frozen lockfile, which is what CI does too |

`ignoreCommand` cancels a build when the commit touched nothing the site is built from. Most
commits here are Go, and a docs site does not need rebuilding because the store grew a column.
The pathspecs are `:/`-prefixed so they resolve against the repository root wherever Vercel runs
the command from, and a missing `HEAD^` — a first deployment, a shallow clone — exits non-zero,
which means build. Failing towards a build is the safe direction.

CI builds this site from cold on every push and uploads `.next` as an artifact, so a change to
the docs can be read rather than diffed. That runs whether or not the deployment does.
