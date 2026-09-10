# Nooks — plan and status

Everything in the design canvas, broken into shippable milestones. This file is both the plan and the
status board: update the checkbox in the same commit as the work.

Terms are from `CONTEXT.md`. UI work is specified in `DESIGN.md`. Code rules are in `STANDARDS.md`.

**Legend** — `[ ]` not started · `[~]` in progress · `[x]` done

---

## Status

| Milestone | State |
| --- | --- |
| M0 · Groundwork | `[x]` |
| M1 · Skeleton that runs | `[~]` all but TanStack Router |
| M2 · Auth and first run | `[ ]` |
| M3 · Lists and Items — the core | `[ ]` |
| M4 · Views: Today, Upcoming, Calendar | `[ ]` |
| M5 · Notes | `[ ]` |
| M6 · Sharing, presence, Activity | `[ ]` |
| M7 · Members and Groups | `[ ]` |
| M8 · Public list | `[ ]` |
| M9 · Print | `[ ]` |
| M10 · Access tokens, REST, MCP | `[ ]` |
| M11 · Settings, export, import | `[ ]` |
| M12 · Website: marketing, docs, API docs | `[ ]` |
| M13 · Offline and conflicts | `[ ]` |
| M14 · Ship | `[ ]` |

---

## M0 · Groundwork

- [x] Clone `usememos/memos` into `reference/`, gitignored, as a worked example of this stack
- [x] Install toolchain: Go 1.27.1, Node 24.21.0 LTS, pnpm 12.3.4, buf 1.72.0
- [x] Pin buf as a workspace devDependency — no global install for contributors
- [x] `CONTEXT.md` — domain glossary, 26 terms, each with what not to call it
- [x] `DESIGN.md` — the design system spec, exact values, OKLCH, shadcn contract
- [x] `STANDARDS.md` — code standards
- [x] `AGENTS.md` — repo rules, commands, code map
- [x] `TODO.md` — this file
- [x] Decide the Go module path — `github.com/hoshomoh/nooks`
- [x] Decide v1 database drivers — SQLite and Postgres
- [x] Decide the live-update transport — SSE
- [x] `go.mod` (`github.com/hoshomoh/nooks`), `cmd/nooks/`, `internal/version/`, `.editorconfig`
- [x] Licence — AGPL-3.0
- [x] Commit convention — Conventional Commits, no attribution trailers

---

## M1 · Skeleton that runs

The goal is a binary that serves an empty page, and a `buf generate` that produces both Go and
TypeScript. No features. Everything after this is filling in.

- [x] `proto/` module: `buf.yaml`, `buf.gen.yaml` — Go, Connect, TS via `@bufbuild/es`.
      Gateway and OpenAPI deferred to M10, when REST is actually specified
- [x] First service definition end to end — `InstanceService`, generating Go and TypeScript
- [x] Store interface + SQLite and Postgres drivers, migration runner, `LATEST.sql` for each.
      The suite runs against both; CI provides a Postgres service so the driver cannot rot
- [x] Config: flags and env (port, data directory, driver, DSN), parsed as a pure function
- [x] HTTP server: routing, graceful shutdown, request logging
- [x] `apps/web`: Vite + React + TypeScript
- [x] Tailwind v4 + `shadcn init` (Base UI, Nova). Verified against `DESIGN.md` §16 and the spec
      amended: `@import "shadcn/tailwind.css"` is real, and the five radii map onto `sm`–`2xl`
- [x] Tokens: full light and dark palettes in OKLCH, Nooks' own tokens, radii
- [x] Self-host Public Sans and IBM Plex Mono, bundled and served from the binary
- [x] Theme: light / dark / system, `.dark` class, persisted, follows the machine live
- [~] TanStack Query wired; TanStack Router lands with the first real routes in M2
- [x] Connect-web transport wired to the generated client
- [x] SPA embedded into the binary; one process serves API and app, with SPA fallback
- [x] CI: `go build`/`vet`/`test -race`/`gofmt`, `buf lint` and `buf format -d`, web typecheck,
      lint and build

---

## M2 · Auth and first run

From `Nooks-Auth-Onboarding.dc.html`.

- [ ] Member and Instance schema; password hashing; sessions
- [ ] Password rule: twelve characters or more, **no other rules**
- [ ] First run — create the first Admin and name the Instance, in one form
- [ ] Sign in
- [ ] Temporary password set by an Admin, and forced replacement on first sign-in
- [ ] Join request: request → Admin approves → Member chooses a password
- [ ] Reset request: request → Admin approves out of band → Member sets a new password.
      **Approval expires in an hour.**
- [ ] Signed-out shell: 52px chrome bar, centred card, eyebrow / title / blurb / fields / footer links
- [ ] Guard: no route reachable before first run completes

---

## M3 · Lists and Items — the core

From `Nooks-List-View.dc.html`. The single most important milestone; everything else is furniture.

- [ ] List and Item schema, ordering, soft delete
- [ ] Sidebar: Instance switcher, Search, Today / Upcoming / All lists, Pinned / My lists / Shared with
      me, Add a list, Member footer
- [ ] App shell: 258px sidebar, 44px chrome bar, 660px content column
- [ ] List view: title, avatars, orientation line, count, the hairline rule
- [ ] **The list row** — 44px, `20px 1fr auto`, truncation in the label, metadata right-aligned
- [ ] Checkbox states: rest, hover, focus, done, just ticked by someone else
- [ ] Add row: `+`, placeholder, `↵` keycap, stays focused after adding
- [ ] Quantity as free text, shown as a mono badge
- [ ] Due dates, and date parsing from the typed text (`milk friday`)
- [ ] Tick and untick, with attribution
- [ ] Completed Items: placement and the "3 done today" row
- [ ] Empty states: empty List, cold-start All lists
- [ ] List menu (`···`): rename, pin, duplicate, print, sort, completed placement, export as plain
      text, delete
- [ ] Pinning, per Member
- [ ] Keyboard: `↵`, `↑` `↓`, `space`, `⌘K` jump-to-list

---

## M4 · Views: Today, Upcoming, Calendar

- [ ] Today: overdue and due-today sections, gathered from every reachable List
- [ ] Upcoming: next two weeks, grouped by day
- [ ] Calendar: month grid, 1060px column, **dated Items only**
- [ ] Add row that targets a List and a date from context
- [ ] Empty states: nothing due today says what is next, and does not apologise

---

## M5 · Notes

- [ ] Note and block schema
- [ ] Side sheet at 520px — opening an Item never replaces the List
- [ ] Full-screen Note view, and Esc back to the List
- [ ] Blocks: paragraph, heading, checklist, quote, code — same markup, two scales
- [ ] Markdown shorthand converts as typed, and is never displayed back
- [ ] `/` block menu, filtered, showing each shortcut
- [ ] Row preview: the Note's own first line plus `+N lines`. **Never a summary.**
- [ ] Autosave, and the "Saving… / Saved" footer state

---

## M6 · Sharing, presence, Activity

From `Nooks-Sharing-Team.dc.html`.

- [ ] Sharing model: private / everyone on the Instance / specific Members and Groups
- [ ] Can-edit toggle; read-only means see and print, not tick or add
- [ ] Share dialog, and the specific-people dialog (Groups first, then individuals)
- [ ] Copy list address
- [ ] Live updates — someone else's tick lands with a one-second highlight, then settles.
      **No toast, no sound.**
- [ ] Presence: "Jonas is here", avatar ring in `done`
- [ ] Activity panel: join requests, reset requests, shares, conflicts, token use
- [ ] Approve and ignore actions. **Ignore is silent and never notifies the sender.**
- [ ] Unread state and the dot on the Activity control

---

## M7 · Members and Groups

- [ ] Members page: table, roles, added dates, `···` menu
- [ ] Add a member — sets a temporary password, read out once
- [ ] A Member who has never signed in greys their name and last column, not the whole row
- [ ] Groups page: cards two-up, membership, lists shared with the Group
- [ ] Add a group; add and remove Members
- [ ] Removing someone from a Group takes away the lists they got through it, and nothing else

---

## M8 · Public list

From `Nooks-Public-Access.dc.html`.

- [ ] At most one Public list per Instance, at a stable address, no password, no account
- [ ] Public page: no sidebar, no attribution, no other List reachable
- [ ] Settings: which List, show contributor names, show quantities and dates, let visitors ask to join
- [ ] Sign-in prompt appears **under the row the Visitor touched**, not as a wall
- [ ] The attempted tick is remembered and applied once they are in
- [ ] Ask to join dialog, and the request-sent state
- [ ] Empty state explains the page rather than asking for work
- [ ] Auto-refresh — it is meant to stay open on the way to the shop

---

## M9 · Print

From `Nooks-Print.dc.html`.

- [ ] `@page` A4, `18mm 18mm 14mm`, print stylesheet
- [ ] Header: eyebrow, title, date, counts, and the 0.7pt ink rule
- [ ] Rows with real 6mm checkboxes; quantity and attribution
- [ ] Three blank dashed rows at the end
- [ ] Two-up for long lists, type down one step, never below 13pt
- [ ] Footer with page numbers
- [ ] `⌘P` prints the List you are looking at
- [ ] Resolve the 40pt/27pt conflict *(open question 2)*

---

## M10 · Access tokens, REST, MCP

- [ ] Access token schema: hashed, scoped to named Lists, permissioned, expiring
- [ ] Tokens page — a Member's own; an Admin also sees that others' exist
- [ ] An Admin can revoke another Member's token but **cannot read it or make one in their name**
- [ ] Add a token dialog: name, List scope picker, permissions, expiry
- [ ] Secret shown once, with a plain sentence that it cannot be shown again
- [ ] Unpicked Lists are invisible to a token — it cannot see that they exist
- [ ] Token activity log
- [ ] REST API at `/api/v1`
- [ ] MCP server at `/mcp`, same token
- [ ] Changes made by a token are attributed to the token in List history

---

## M11 · Settings, export, import

- [ ] Account: name, email, change password
- [ ] Appearance: theme segment
- [ ] Instance (admins only): Instance name, public signup toggle
- [ ] About: version, storage, instance age, counts, licence, **Telemetry: None**
- [ ] Export everything — one JSON file plus the printed pages as PDFs
- [ ] Import a backup — the same file the exporter writes, so moving machines is copy and restore

---

## M12 · Website: marketing, docs, API docs

`apps/website`, Next.js + Fumadocs. Part of v1 — a self-hosted app nobody can read about does not
ship. It reuses `DESIGN.md`, so the site looks like the product rather than like a template.

- [ ] Next.js app in `apps/website` with Tailwind and the Nooks tokens
- [ ] Landing page: what Nooks is, the printed page, one screenshot, how to run it
- [ ] Docs with Fumadocs MDX: install, configure, back up, upgrade, reverse proxy
- [ ] API docs generated from the OpenAPI spec via `fumadocs-openapi` — **needs buf's
      gnostic-openapi plugin, pulled forward from M10**
- [ ] MCP docs: what a token reaches, how to point a client at `/mcp`
- [ ] Dark mode, sharing the app's `.dark` contract
- [ ] Deploy target and a preview build in CI

---

## M13 · Offline and conflicts

- [ ] Offline banner, retry, last-seen time
- [ ] Reading, ticking and adding all work offline; queued changes carry `not synced`
- [ ] Sync on reconnect
- [ ] **Ticks never conflict** — last write wins, whoever made it
- [ ] Competing text prompts, on the row it affects only: Keep mine / Keep both
- [ ] Error on save keeps the Member's text and offers real options
- [ ] Loading skeletons that hold the exact height of real rows
- [ ] "Still waiting on the server" after 4s

---

## M14 · Ship

- [ ] Docker image, `docker-compose.yml`, and a one-line run command
- [ ] Release workflow, versioned binaries
- [x] `README.md`: what it is, how to run it, how to back it up
- [ ] `CONTRIBUTING.md` pointing at `STANDARDS.md`, `CONTEXT.md`, `DESIGN.md`
- [x] Licence — AGPL-3.0
- [ ] Accessibility pass: focus order, labels, contrast, 44px hit areas
- [ ] Seed data for a believable first run

---

## Decisions

| Decision | Choice | Consequence |
| --- | --- | --- |
| Go module path | `github.com/hoshomoh/nooks` | Clone path and module path agree |
| Database drivers, v1 | **SQLite and Postgres** | Every schema change ships migrations and `LATEST.sql` for both, plus driver tests. SQLite stays the default — the About screen promises "one file you can copy" |
| Live updates | **SSE** | One-directional, plain HTTP, self-reconnecting, fine behind a reverse proxy. Enough for ticks, presence and Activity |
| buf | Workspace devDependency | `pnpm exec buf`; no global install for contributors |
| Data access | **Bun** | SQL-first, and the only option giving both dialects from one query set. memos writes every query three times, once per driver |
| Token naming | shadcn's names; Nooks' accent is `shared` | Avoids colliding with shadcn's `--accent` hover surface |
| Repository shape | **Monorepo**: `apps/*` and `packages/*`, Go at the root | Room for the website now and Expo later; shared code lives in a package rather than inside one app |
| Generated TS client | `packages/api` (`@nooks/api`) | Two known consumers — the web app now, Expo later — so it is a package from the start rather than a later extraction |
| Website stack | **Next.js + Fumadocs**, in v1 | Same React/Tailwind/shadcn stack as the app, so `DESIGN.md` carries over. `fumadocs-openapi` turns the spec into API docs |
| Mobile app | **Expo**, after v1 | Shares `@nooks/api` and tokens, not DOM components |

## Open questions

None outstanding.

Resolved so far: the print type scale (canvas now says 27/16/13, matching the Print artboard); the
accent tint (collapsed to one value, `#F1F6FB`); `@import "shadcn/tailwind.css"` (confirmed — it is
what `shadcn init` emits); fonts offline (Public Sans and IBM Plex Mono are self-hosted via
`@fontsource`, bundled and served from the binary).

---

## Explicitly not in v1

Named so nobody wonders whether they were forgotten. None of these appear in the design.

- Recurring Items · Reminders and push notifications · File attachments · Sub-lists or nesting beyond
  Note checklists · Full-text search across Notes (⌘K jumps to Lists only) · Multiple Instances behind
  one deployment · Email, of any kind

The **mobile app is planned but not v1**: Expo, in `apps/mobile`, sharing `@nooks/api` and the design
tokens. The layout already has room for it.
