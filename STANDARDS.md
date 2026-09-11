# Code standards

Binding on every contributor and every agent working in this repo. `AGENTS.md` points here; reviews
enforce it. Where a rule below and a rule in `AGENTS.md` disagree, `AGENTS.md` wins — it is closer to
the commands.

Nooks is a small app that people run on their own machines and that other developers will extend after
v1. Both facts push the same way: **plain code that a newcomer can read, test, and change without
first learning our abstractions.**

---

## 1. Architecture

**Single responsibility.** Each function, type, or package does exactly one thing and has one reason to
change. A `store` package persists; it does not decide policy. A service method decides policy; it does
not format HTTP.

**KISS.** No abstract framework, no plugin system, no premature optimisation, unless a task explicitly
calls for it. Two similar call sites do not justify an abstraction — the third might. Prefer the
obvious implementation over the clever one; someone will read it at 11pm on a home server that will not
start.

**DRY, applied to knowledge rather than characters.** Extract genuinely shared logic into a named
helper. Do **not** merge two things that merely look alike today — duplication is cheaper than the
wrong abstraction.

**Dependencies point inwards.** `internal/` packages know nothing of `store` or `server`. `store` knows
nothing of `server`. Nothing imports `reference/`.

---

## 2. Testability

**Pure functions by default.** A function given the same input returns the same output, with no hidden
side effects. Anything non-deterministic — the clock, random values, IDs, the filesystem, the network —
is passed in, not reached for. This is what makes the date parser, the markdown shorthand converter,
and the recurrence logic testable without a database.

**Dependency injection, explicitly.** A database handle, an HTTP client, a config struct, and a clock
are **arguments**, never things a constructor reaches out and creates. In Go, accept an interface and
return a concrete type. In TypeScript, take the dependency as a parameter or a hook.

**Small functions.** Aim under 20–30 lines. If a function needs a section comment to explain its
second half, that half is a function.

**Test what can break.** Every bug fix arrives with the test that would have caught it. Pure logic gets
unit tests; store code gets driver tests; the layer in between gets one test that proves the wiring.

---

## 3. Readability

**Intention-revealing names.** `calculateMonthlyRevenue`, not `calcRev`. `overdueItemsForMember`, not
`getData`. Use the terms from `CONTEXT.md` exactly — a variable holding a List is `list`, never `board`
or `collection`.

**Explicit types on every signature.** Go: no naked `interface{}`/`any` at a package boundary.
TypeScript: `strict` is on, exported functions and components carry full parameter and return types,
and `any` needs a comment justifying it.

**Comment *why*, never *what*.** The code says what it does. A comment explains a non-obvious decision,
a workaround, a spec quirk, or a constraint that is not visible locally — for example why ticks are
last-write-wins while text conflicts prompt.

---

## 4. Errors

**Guard clauses, fail early.** Handle edge cases at the top and return. No deeply nested `if`/`else`;
the happy path stays at the left margin.

**Standardised handling, never silent.**

- Go: return `error` as the last value, wrap with `%w` and context (`fmt.Errorf("load list %s: %w", uid,
  err)`). Never `_ = err`. Panics are for programmer errors at startup, never for request handling.
- TypeScript: throw `Error` subclasses or return a typed result; never swallow in an empty `catch`.
  React Query owns retry and error state — do not hand-roll it per component.

**Error copy is design copy.** User-facing errors follow `DESIGN.md` §11 and §13: say what happened and
what became of the Member's work. Never "Something went wrong".

---

## 5. Frontend specifics

**Generated shadcn components are never edited.** `src/components/ui/` is output from
`pnpm dlx shadcn@latest add <component>` and must stay byte-for-byte what the generator
produces, so that adding or updating a component never clobbers a local change. Nooks'
own components live in `src/components/ds/` and compose those primitives. When something
needs to look or behave differently, in this order:

1. Wrap it in `ds/`, passing `className` or props.
2. Change a token in `src/index.css` — which fixes every instance rather than one.
3. Only if neither works, discuss it before touching `ui/`.

**`useEffect` is a last resort, not a default.** Most of what it gets used for has a
better tool:

| Instead of an effect that… | Use |
| --- | --- |
| Subscribes to something outside React (media queries, storage, sockets) | `useSyncExternalStore` |
| Derives a value from props or state | Compute it during render |
| Resets state when a prop changes | A `key`, or compute it during render |
| Fetches data | TanStack Query |
| Responds to a user action | The event handler that caused it |
| Writes to the DOM the module already owns | That module, synchronously |

What is left — genuinely synchronising with an external system, and nothing else — is
what `useEffect` is for. If a case seems to need one, raise it rather than reaching for
it quietly.

**Types are named, never inline.** A prop object written into a signature cannot be
imported, extended, or read at a glance:

```tsx
// No.
function AppShell({ instanceName, lists }: { instanceName: string; lists: List[] }) {}

// Yes.
export type AppShellProps = {
  instanceName: string
  lists: List[]
}

function AppShell({ instanceName, lists }: AppShellProps) {}
```

The same applies to mutation variables, hook returns, and anything else with a shape.
Export the type when anything outside the file could want it.

**Dates go through `lib/dates.ts`, and nothing else.** It wraps date-fns; no component
parses, formats or compares a date itself. `Intl` is not used for dates: it renders
September as "Sept" under `en-GB` where the design says "Sep", and its data changes with
the runtime's ICU version, so the same build would look different on different machines.
`Intl` **is** right for numbers, currency and lists — those are in `lib/format.ts`.

**Every string a Member reads comes from a locale file.** `src/i18n/locales/en.json` is
the source language and holds the design's copy; adding a language is adding a JSON file
and one entry in `i18n/locales.ts`, never a code change. No English in a component.
Dates follow the same language: `lib/locale-store.ts` owns both the i18next language and
the date-fns locale, so the two cannot drift.

Otherwise:

- Presentation and data fetching stay separate: a component that renders a List row does
  not know the transport. Server state is TanStack Query's; local UI state is `useState`.
  There is no third store.
- Components follow `DESIGN.md`. No colour, radius, or size literal in a component —
  tokens only.
- Accessibility is not a later pass: real `<button>`s, labelled inputs, visible 2px focus
  rings, 44px hit areas, and keyboard paths for everything in `DESIGN.md` §12.

## 6. Commits and review

**Before every commit**, run `./scripts/ci.sh`. It is exactly what CI runs, in the same
order. Discovering a break after pushing is the same information, slower and noisier.

**Before every push**, run `./scripts/preflight.sh`. It checks out `HEAD` somewhere
else, installs from the lockfile and runs the same script there — so what passes is what
you are about to push, rather than a working tree that has moved on since.

The difference is not pedantry. `ci.sh` checks the tree you are looking at, and two
ordinary things break that: a fix made *after* the run passed, and a file that was never
`git add`ed. Both leave a green run locally and a red one on GitHub. Note also that
`tsc -b` is incremental, so a local run can skip files a fresh checkout will not.

**Conventional Commits, and keep them short.**

```
<type>(<optional scope>): <description>
```

Types: `feat` · `fix` · `docs` · `refactor` · `test` · `chore` · `build` · `ci` · `perf` · `style`.
Scope is a package or area — `store`, `auth`, `list-view`, `proto`.

- Description in the imperative, lower case, no full stop: `feat(store): add sqlite migration runner`.
- **No body by default.** Add one only when the *why* is not obvious from the diff, and then keep it to
  a line or two — not a summary of what changed, which the diff already says.
- `!` after the type, or a `BREAKING CHANGE:` footer, for anything that breaks the API or the schema.
- No tool or assistant attribution trailers.

Otherwise:

- One logical change per commit.
- Generated files (`proto/gen/`, `web/src/types/proto/`) are committed but never hand-edited.
- A PR that changes behaviour changes a test. A PR that changes the schema adds migrations for every
  driver.
