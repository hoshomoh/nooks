# Nooks — design system

The specification the UI is built to. Values here are extracted from the design canvas
(`docs/design/README.md`) and are exact, not approximate. If a component in the app disagrees with
this file, the component is wrong.

Enforcement lives in `web/src/styles/tokens.css`: every value below exists there as a custom property,
and no colour, radius or size may be written inline in a component. Use `CONTEXT.md` terms for the
things being displayed.

---

## 1. Principles

- **A document, not a dashboard.** One list primitive, one action to add something. The signature
  element is the hairline rule under the list title — everything hangs off it, and it is the one mark
  that survives onto paper.
- **Four meaning colours, one job each**: accent/shared, done/presence, offline, overdue/error.
  Nothing is coloured for decoration. A screen at rest shows only ink, paper, and one blue dot per
  shared List.
- **A row is 44px** whether it carries one piece of metadata or five.
- **16px is the floor** for anything a Member wrote.
- **Never summarise the Member.** A long Note previews as its own first line plus a count of what is
  left. Nothing is generated on their behalf.
- **Nooks sends no email.** Anything that would need one waits in Activity.

---

## 2. Colour

### Light

| Token | Value | Job |
| --- | --- | --- |
| `--nooks-desk` | `#F2F2F0` | The surface behind the app frame |
| `--nooks-paper` | `#FFFFFF` | Panels, main content, dialogs |
| `--nooks-sidebar` | `#FBFBFA` | Sidebar, inset panels |
| `--nooks-selected` | `#EFEFED` | Selected row, active toolbar button |
| `--nooks-line` | `#E6E6E3` | Block rule — separates blocks inside a section |
| `--nooks-hair` | `#F0F0EE` | Row rule — separates rows |
| `--nooks-control` | `#8F8F8A` | Checkbox and radio borders at rest |
| `--nooks-muted` | `#5F5F5B` | Muted text — counts, timestamps, attribution |
| `--nooks-sub` | `#3F3F3C` | Secondary text — body copy, explanations |
| `--nooks-ink` | `#1F1F1E` | Primary text, primary button fill, monogram |
| `--nooks-chip` | `#EDEDEA` | Avatar and chip fill |
| `--nooks-input-line` | `#D6D6D1` | Input border at rest (darker than `--nooks-line`) |

### Light — meaning colours

| Token | Value | Job |
| --- | --- | --- |
| `--nooks-accent` | `#37618F` | Shared, links, focus ring, quiet buttons, due dates |
| `--nooks-accent-bg` | `#F1F6FB` | Accent fill — badges, the secret panel, a selected share option |
| `--nooks-accent-line` | `#C9DAEC` | Accent panel border |
| `--nooks-done` | `#2E7D5B` | Done, and presence of another Member |
| `--nooks-done-bg` | `#F4F8F4` | Row just ticked by someone else |
| `--nooks-offline` | `#9A6B14` | Offline dot |
| `--nooks-offline-bg` | `#FBF6E9` | Offline banner fill |
| `--nooks-offline-line` | `#EDE1C4` | Offline banner border |
| `--nooks-offline-text` | `#6B4A0E` | Offline banner text |
| `--nooks-error` | `#A03A2E` | Overdue dates |
| `--nooks-error-text` | `#8E3B31` | Error text, destructive button label |
| `--nooks-error-line` | `#E4C2BC` | Error and destructive borders |
| `--nooks-error-bg` | `#FCF5F4` | Rejected input, errored row |
| `--nooks-scrim` | `rgba(31,31,30,0.34)` | Behind dialogs — ink at 34% |
| `--nooks-knob` | `#FFFFFF` | Toggle knob |
| `--nooks-toggle-off` | `#E0E0DC` | Toggle track, off |

### Dark

| Token | Value |
| --- | --- |
| `--nooks-desk` | `#111110` |
| `--nooks-paper` | `#191918` |
| `--nooks-sidebar` | `#151514` |
| `--nooks-selected` | `#232321` |
| `--nooks-line` | `#2E2E2B` |
| `--nooks-hair` | `#242422` |
| `--nooks-control` | `#7A7A73` |
| `--nooks-muted` | `#9A9A95` |
| `--nooks-sub` | `#C8C8C3` |
| `--nooks-ink` | `#ECECEA` |
| `--nooks-chip` | `#2A2A27` |
| `--nooks-accent` | `#7FA8D4` |
| `--nooks-accent-bg` | `#1B2129` |
| `--nooks-accent-line` | `#33465C` |
| `--nooks-done` | `#5FA98A` |
| `--nooks-done-bg` | `#18241F` |
| `--nooks-offline` | `#D9A93C` |
| `--nooks-offline-bg` | `#2A2413` |
| `--nooks-offline-line` | `#3D3520` |
| `--nooks-offline-text` | `#E4C67E` |
| `--nooks-error` | `#E08B7C` |
| `--nooks-error-text` | `#D98A7E` |
| `--nooks-error-line` | `#5C3A34` |
| `--nooks-scrim` | `rgba(0,0,0,0.55)` — ink at 55% |
| `--nooks-knob` | `#0F0F0E` |
| `--nooks-toggle-off` | `#33332F` |

**Visitor variant.** On the Public list the checkbox is not interactive, so its border uses a lighter
control: `#B4B4AE` light, `#5A5A55` dark.

**Theme selection.** Light / Dark / System, a three-segment control in Settings → Appearance. System is
the default. Dark is a full palette swap, never a filter.

---

## 3. Type

Two families, loaded from Google Fonts with a real fallback stack:

- **Public Sans** 400 / 500 / 600 — everything.
- **IBM Plex Mono** 400 / 500 — quantities, keycaps, tokens, addresses. **Never for prose.**

| Role | Size / weight / tracking | Where |
| --- | --- | --- |
| Display | 33 / 600 / −0.025em / lh 1.1 | List titles, view headings |
| Page title | 27 / 600 / −0.02em | Settings pages, states |
| Dialog title | 21 / 600 / −0.015em | Dialog headers |
| Section heading | 19 / 600 / −0.01em | Sections within a page |
| Note heading | 18 / 600 / −0.01em / lh 1.4 | Note blocks, full screen |
| Empty title | 17 / 500 | Empty states |
| Body | 16 / 400 / lh 1.4 | **Item labels — the floor for member-written text** |
| Note body | 15.5 / 400 / lh 1.6 | Note paragraphs, full screen |
| Input | 15 / 400 | Field values |
| Note body, sheet | 14.5 / 400 / lh 1.6 | Note paragraphs in the side sheet |
| Chrome | 14 / 500 | Sidebar items, buttons, field labels |
| Secondary | 13.5 / 400 | Explanations, sidebar rows, hints |
| Small | 13 / 400 | Table cells, meta values |
| Micro | 12.5 / 400 | Breadcrumbs, attribution, counts |
| Section head | 11.5 / 600 / 0.08em / uppercase | Page section labels |
| Column head | 11.5 / 600 / 0.04em / uppercase | Table columns |
| Mono badge | 11.5 | Quantities |
| Mono keycap | 10.5 | ⌘K, ↵ |
| Mono secret | 13.5 / break-all | Access tokens |

Uppercase is reserved for the 11.5 section and column heads. Nothing else is uppercased.

---

## 4. Spacing, radius, rules

**Spacing scale** — a 2px base with a loose ratio: `2 · 4 · 6 · 8 · 12 · 16 · 22 · 34 · 56`.

**Radius means something:**

| Radius | Used for |
| --- | --- |
| 4 | Checkbox, small badge |
| 6 | Row, toolbar button, menu item |
| 7 | Control — button, input, toggle track |
| 8 | Inner surface — card, empty state, banner |
| 10 | Surface — panel, dialog |
| 15 | Chip (a 30px pill) |
| 50% | Avatar |

**Three rule weights, three meanings — do not substitute:**

| Weight | Colour | Separates |
| --- | --- | --- |
| 2px | `--nooks-line` | Sections of a page |
| 1px | `--nooks-line` | Blocks inside a section |
| 1px | `--nooks-hair` | Rows |

---

## 5. Layout

| Element | Size |
| --- | --- |
| App frame (design reference) | 1440 × 900 |
| Sidebar | 258px |
| Chrome bar | 44px (52px on signed-out pages) |
| Content column | max-width 660px, centred |
| Calendar column | max-width 1060px |
| Side sheet | 520px |
| Dialog | 560px |
| Content padding | 56px top / 22px sides / 90px bottom |
| Sheet footer bar | 42px |

When the side sheet is open the content pane keeps its place — right padding becomes 544px rather than
the list being replaced.

---

## 6. The list row

The core component. Everything else is furniture around it.

```
grid-template-columns: 20px 1fr auto;
align-items: center;
gap: 14px;
min-height: 44px;
padding: 6px 8px;
margin: 0 -8px;
border-radius: 6px;
```

- Checkbox **17 × 17**, 1.5px `--nooks-control` border, radius 4. Hit area is **44px** regardless.
- Label 16px, truncates with ellipsis. **Truncation happens in the label**, never in the metadata, so
  the label always starts at the same x and the row never grows.
- Metadata is right-aligned in the third column, `gap: 12px`, `white-space: nowrap`: quantity badge,
  list name, date, `note` marker, attribution.
- Quantity is a mono 11.5 badge: 1px `--nooks-line`, radius 4, padding `1px 5px`.
- A due date takes `--nooks-accent`; an overdue date takes `--nooks-error`.
- Hover fills `--nooks-selected` and darkens the checkbox border to `--nooks-muted`.
- Done: checkbox filled `--nooks-muted` with a white tick; label `--nooks-muted` with
  `line-through` in `#C4C4BE`.
- Just ticked by someone else: fill `--nooks-done-bg`, checkbox filled `--nooks-done`, attribution
  `--nooks-done`. Holds for one second, then settles. **No toast, no sound.**
- A Note previews as one line below the row: 13.5px `--nooks-sub`, on a 2px `--nooks-line` left rule
  indented 10px, in the label column, with `+N lines` after it.

**Add row.** Same grid, `border-top: 1px solid --nooks-hair`, a `+` in the checkbox column, placeholder
in `--nooks-muted`, and a mono `↵` keycap at the right. Enter adds and keeps focus for the next one.

---

## 7. Controls

### Buttons

| Variant | Spec |
| --- | --- |
| Primary | `--nooks-ink` fill, `--nooks-paper` text, 500 |
| Secondary | 1px `--nooks-line` border, `--nooks-sub` text |
| Quiet | no border, `--nooks-accent` text |
| Destructive | 1px `--nooks-error-line` border, `--nooks-error-text` text |

| Size | Height / padding / radius / text | Allowed in |
| --- | --- | --- |
| Default | 34 / `0 16` / 7 / 14 | Page and section headers, dialog footers |
| Compact | 32 / `0 14` / 7 / 13.5 | Inside a table row or settings row |
| Toolbar | 26 / `0 10` / 6 / 12.5 | The 44px chrome bar only |

**Focus is 2px `--nooks-accent` outline at 2px offset**, everywhere, on every focusable thing.

### Inputs

38px high, radius 7, 15px text. Rest 1px `--nooks-input-line`; focus **1.5px** `--nooks-accent`;
rejected 1px `--nooks-error-line` on `--nooks-error-bg` with `--nooks-error-text`. Settings-row inputs are
36px and 280px wide. Labels are 13 / 500 `--nooks-sub`; hints sit under the field in 12.5
`--nooks-muted`.

### Toggle

34 × 20 track, radius 10, 16px knob, 2px padding. On: `--nooks-accent` track, knob right. Off:
`--nooks-toggle-off` track, knob left. The word `On` / `Off` sits to its left in 13 `--nooks-muted`.

### Chips and permission lists

Chips pick many of one kind of thing: 30px high, radius 15, 13px, a 13px square swatch at 3px radius
inside. Selected fills `--nooks-selected`; unselected takes a 1px `--nooks-line` border.

Use the bordered list instead when each choice needs a sentence explaining what it costs: 1px
`--nooks-line`, radius 8, rows padded `12px 14px` split by 1px `--nooks-hair`, a 15px checkbox, label 14
and explanation 12.5 `--nooks-muted`.

---

## 8. Page furniture

**Chrome bar** — 44px, 1px `--nooks-hair` bottom, padding `0 16px 0 22px`. Breadcrumb left in 12.5
(`--nooks-muted`, current segment `--nooks-sub`), toolbar buttons right, then the `···` menu at 26 × 26.
It carries location and view-level actions only — **never the page's primary action**.

**Page header** — 27px title, one 14px line of orientation capped at 540px, primary action pinned
right and top-aligned with the title. The action never sits at the bottom of the page.

**Section header** — 2px `--nooks-line` above, 18px padding, 19px heading, 13px explanation, secondary
action pinned right.

**Table** — 11.5 uppercase column heads over a 1px `--nooks-line` rule; rows min 56px split by 1px
`--nooks-hair`; compact buttons inside; the last column is always the `···` menu at 24px. A row that has
never been used greys its name and its last column, **not the whole row**.

**Card** — 1px `--nooks-line`, radius 8, padding `15px 16px`. Name row, body, then one
`--nooks-hair`-separated footer fact. Cards go two-up in a 12px grid; they never stretch to a
single full-width column.

**Settings row** — `190px 1fr` grid, 24px gap, min-height 48px, split by 1px `--nooks-hair`. Label 14
with a 12.5 `--nooks-muted` explanation under it; control right-aligned.

**Empty state** — 1px **dashed** `--nooks-line`, radius 8, padding `30px 26px`, **left-aligned, never
centred**. 17 / 500 statement of fact, then one 14 / 1.6 sentence capped at 440px saying what the thing
is for. No illustration, no centred hero, and **no button if the page header already has one**.

---

## 9. Overlays

**Dialog** — 560px, radius 10, 1px `--nooks-line`, shadow `0 24px 60px rgba(0,0,0,0.22)`, over the
scrim. Header `24px 26px 0`: 21px title then one 14 / 1.6 explaining line. Body. Then a
`--nooks-hair`-topped footer at `14px 26px`: **consequence text left, Cancel then confirm right.**

Say what happens: *"Ignore removes the request silently and never notifies the sender."* Not *"Are you
sure?"*

**Menu** — 248px, radius 9, 1px `--nooks-line`, 6px padding; items 32px at radius 6 with the shortcut
right in 12 `--nooks-muted`; groups split by a 1px `--nooks-hair` rule inset 8px. Destructive items take
`--nooks-error`.

**Side sheet** — 520px, pinned below the chrome bar, 1px `--nooks-line` left border. Same order as full
screen: chrome bar, checkbox and title, detail row, hairline, Note, footer bar. Title 24 / 600, fields
in a wrapping row of 12.5 label + 13 value pairs. **Opening an Item never replaces the List with a
page.**

**Secret shown once** — accent panel (`--nooks-accent-bg` on 1px `--nooks-accent-line`, radius 8),
mono 13.5 with `word-break: break-all`, a primary Copy button, and a sentence saying plainly that it
cannot be shown again. **Never a warning triangle.**

---

## 10. Note blocks

The same markup renders in the side sheet and full screen; only the scale changes. Markdown shorthand
converts a block as it is typed but is **never displayed back to the Member**.

| Block | Full screen | Sheet | Spec |
| --- | --- | --- | --- |
| Heading | 18 / 600 / lh 1.4 | 13.5 | padding `18px 0 3px` |
| Paragraph | 15.5 / lh 1.6 | 14.5 | padding `5px 0` |
| Checklist | 15.5 | 14.5 | 15px box in a 26px gutter, padding `3px 0` |
| Quote | 15.5 `--nooks-sub` | 14.5 | 2px `--nooks-line` left rule, 13px indent, padding `10px 0` |
| Code | mono 13 `--nooks-sub` | 12.5 | padding `10px 0` |

Typing `/` on an empty line opens a 300px block menu, filtered as typing continues. Every entry shows
its markdown shorthand, so **the menu teaches the shortcut rather than replacing it**.

---

## 11. States

**Loading** — skeleton rows hold the **exact height of real rows**, so nothing jumps when they arrive.
Blocks are `--nooks-selected` fading to `--nooks-hair` down the list. After 4s: *"Still waiting on the
server — 4s. It's probably just waking up."*

**Offline** — a banner above the content: `--nooks-offline-bg` on 1px `--nooks-offline-line`, a 7px
`--nooks-offline` dot, 13.5 `--nooks-offline-text`, and a Retry on the right. Ticks made offline carry a
`not synced` badge. Reading, ticking and adding all still work.

**Error on save** — the errored row gets 1px `--nooks-error-line` on `--nooks-error-bg` with a `Try
again` in `--nooks-error-text`. Below it, a plain sentence saying what the server said and what became
of the Member's text, then real options. Their text is never discarded.

**Conflict** — only competing *text* asks a question, and only on the row it affects. **Ticks never
conflict** — a tick is a tick whoever made it. The two versions are shown as stacked options, the
Member's outlined in accent, with `Keep mine` and `Keep both`.

---

## 12. Keyboard

| Key | Does |
| --- | --- |
| `⌘K` | Jump to any List |
| `↵` | Add — stays focused for the next one |
| `↑` `↓` | Move between Items |
| `space` | Tick |
| `⌘P` | Print the List you are looking at |
| `Esc` | Return to the list from a Note |

---

## 13. Words

- **One verb for creating things, everywhere: Add.** Add a member · Add a group · Add a token · Add a
  list. Never New, Create, Issue, or Generate.
- Sentence case. Buttons and labels take no full stop; explanatory lines are sentences and take one.
- Say what happens, in the footer of the thing that causes it.
- Never summarise the Member.

---

## 14. Identity

The mark is the design's own signature element drawn small: a list title, the hairline rule under it,
and two items. It is not a letter, and it is never set in a typeface.

Canonical geometry, in a `24 × 24` viewBox, `fill="none"`:

| Path | Stroke | Cap | Colour | Is |
| --- | --- | --- | --- | --- |
| `M3 6h9` | 2.2 | square | ink | The list title |
| `M3 11h18` | 1 | butt | ink | **The hairline rule** |
| `M3 15.5h14  M3 19.5h9` | 1.6 | square | control | Two items |

**At 16px the mark simplifies to two paths** — the four-line version fills in at that size. This is an
optical size, not a scale: `M4 8h8` at stroke 3, and `M4 14h16` at stroke 1.5.

**Wordmark** is lowercase — `nooks`, 600, −0.02em. It sits beside the mark, never above it.

| Context | Mark | Wordmark |
| --- | --- | --- |
| Wordmark, light | 32 | 30 |
| Wordmark, dark | 28 | 26 |
| Sidebar | 20 | 14 |

**App icon** is the mark reversed out of an ink square: 64 / radius 14 / glyph 38 · 32 / 7 / 20 ·
16 / 4 / 11 (the simplified two-path mark).

**Colour by ground:**

| Ground | Title and rule | Items |
| --- | --- | --- |
| Paper | ink | control |
| Dark paper | `#ECECEA` | `#8B8B85` |
| Ink (app icon) | `#FFFFFF` | `#9A9A94` |

Tab title is `<List> · nooks`.

The product is **Nooks** in prose and **nooks** as a logotype. Sentences capitalise it; the wordmark,
the tab title and the app icon do not.

---

## 15. Print

The Print sheet is a deliverable, not a screenshot. A4, 210 × 297mm, margins `18mm 18mm 14mm`.

- Eyebrow: mono 8.5pt, 0.14em, uppercase — `<Instance> · shopping list`.
- Title 27pt / 600 / −0.02em, with the date and an `N items · N people` line right-aligned.
- **A 0.7pt solid ink rule under the header** — the one mark that survives from screen to paper.
- Items: `11mm 1fr auto` grid, 4.4mm vertical padding, split by 0.4pt `#E2E2DC`. A real **6mm
  checkbox** at 0.7pt. Label 16pt, quantity mono 10.5pt, attribution 9.5pt.
- **Three blank dashed rows** at the end, for whatever gets remembered in the shop.
- Long lists go two-up: `column-count: 2`, 12mm gap, a 0.4pt column rule, type down one step to 13pt
  and 5.5mm boxes. **Never below 13pt.**
- Footer: a sentence left, and `Nooks · A4 210 × 297 mm · n/N` right in mono 8.5pt.



---

## 16. Tokens — the shadcn contract

The UI is built on shadcn/ui, so tokens follow **shadcn's paradigm, not a parallel one**. Three rules
come from that and are not ours to vary:

1. **Names are shadcn's, unprefixed.** `--background`, `--foreground`, `--border`, `--input`, `--ring`,
   `--primary`, the `--sidebar-*` family. There is no `--nooks-*` layer underneath — the tables in §2
   are the *source values*, and they are written directly into shadcn's variables.
2. **Values are OKLCH.** Hex appears nowhere in the CSS. §2 keeps hex because that is what the canvas
   specifies and what a designer reads; the conversions are in the table below.
3. **Dark mode is the `.dark` class**, declared `@custom-variant dark (&:is(.dark *))`. Not
   `[data-theme]`, not a bare `prefers-color-scheme` swap. Settings → Appearance writes the choice, and
   `System` resolves it against `matchMedia` and toggles the same class.

`web/src/styles/index.css` is therefore laid out exactly as the docs prescribe:

```css
@import "tailwindcss";
@import "shadcn/tailwind.css";

@custom-variant dark (&:is(.dark *));

:root  { /* light values */ }
.dark  { /* dark values  */ }

@theme inline { /* --color-* mappings, incl. Nooks's own */ }

@layer base { /* base styles */ }
```

### The collision, and how it is resolved

**shadcn's `--accent` is a subtle hover surface, not a brand colour.** Nooks's accent is a meaning
colour. Reusing the name would silently restyle every shadcn hover state.

The Foundations artboard labels that swatch **"Shared / accent"**, so Nooks's takes the name **`shared`**
and shadcn keeps `--accent` for its own job. Everywhere §2 says *accent*, the token is `shared`.

### Mapping onto shadcn's semantics

| shadcn variable | Nooks source | Note |
| --- | --- | --- |
| `--background` / `--foreground` | paper / ink | |
| `--card`, `--popover` (+ `-foreground`) | paper / ink | |
| `--primary` / `--primary-foreground` | ink / paper | The primary button is ink on paper |
| `--secondary` / `--secondary-foreground` | selected / sub | |
| `--muted` | selected | shadcn `--muted` is a **surface** |
| `--muted-foreground` | muted `#5F5F5B` | Nooks's "muted" is a **text** colour |
| `--accent` / `--accent-foreground` | selected / ink | shadcn's hover surface — **not** Nooks's accent |
| `--destructive` | error-text | Nooks's destructive button is outlined, not filled |
| `--border` | line | |
| `--input` | input-line `#D6D6D1` | Deliberately darker than `--border` |
| `--ring` | shared | The 2px focus ring |
| `--sidebar` / `--sidebar-foreground` | sidebar / sub | |
| `--sidebar-accent` / `-foreground` | selected / ink | |
| `--sidebar-border` / `--sidebar-ring` | line / shared | |
| `--chart-1` … `--chart-5` | shared, done, offline, overdue, control | Unused; set so shadcn defaults never leak |

### Nooks's own tokens

Added the documented way — declared in `:root` and `.dark`, exposed through `@theme inline` as
`--color-<name>`, which is what generates the `bg-*` / `text-*` / `border-*` utilities:

`shared`, `shared-bg`, `shared-line`, `done`, `done-bg`, `offline`, `offline-bg`,
`offline-line`, `offline-text`, `overdue`, `hair`, `control`, `chip`, `desk`, `toggle-off`, `knob`.

### Radius

shadcn derives `--radius-sm/md/lg/xl` from one `--radius` by ±4px, which cannot express Nooks's
4 / 6 / 7 / 8 / 10. So `--radius: 7px` is set for shadcn's own components, and Nooks's five radii are
explicit tokens: `--radius-checkbox` 4, `--radius-row` 6, `--radius-control` 7, `--radius-inner` 8,
`--radius-surface` 10.

### OKLCH conversions

**Light**

| Token | Hex | OKLCH |
| --- | --- | --- |
| paper | `#FFFFFF` | `oklch(1 0 0)` |
| sidebar | `#FBFBFA` | `oklch(0.9878 0.0013 106.42)` |
| desk | `#F2F2F0` | `oklch(0.9606 0.0027 106.45)` |
| hair | `#F0F0EE` | `oklch(0.9546 0.0027 106.45)` |
| selected | `#EFEFED` | `oklch(0.9516 0.0027 106.45)` |
| chip | `#EDEDEA` | `oklch(0.9453 0.0040 106.48)` |
| line | `#E6E6E3` | `oklch(0.9241 0.0040 106.48)` |
| toggle-off | `#E0E0DC` | `oklch(0.9056 0.0054 106.51)` |
| input-line | `#D6D6D1` | `oklch(0.8747 0.0068 106.54)` |
| control | `#8F8F8A` | `oklch(0.6485 0.0073 106.60)` |
| muted | `#5F5F5B` | `oklch(0.4842 0.0063 106.63)` |
| sub | `#3F3F3C` | `oklch(0.3666 0.0050 106.65)` |
| ink | `#1F1F1E` | `oklch(0.2389 0.0019 106.54)` |
| shared | `#37618F` | `oklch(0.4836 0.0883 252.04)` |
| shared-bg | `#F1F6FB` | `oklch(0.9708 0.0086 247.91)` |
| shared-line | `#C9DAEC` | `oklch(0.8811 0.0309 249.71)` |
| done | `#2E7D5B` | `oklch(0.5314 0.0946 161.86)` |
| done-bg | `#F4F8F4` | `oklch(0.9750 0.0068 145.52)` |
| offline | `#9A6B14` | `oklch(0.5629 0.1115 76.44)` |
| offline-bg | `#FBF6E9` | `oklch(0.9735 0.0180 89.36)` |
| offline-line | `#EDE1C4` | `oklch(0.9118 0.0405 88.20)` |
| offline-text | `#6B4A0E` | `oklch(0.4350 0.0843 76.76)` |
| overdue | `#A03A2E` | `oklch(0.4906 0.1380 29.59)` |
| error-text | `#8E3B31` | `oklch(0.4630 0.1151 29.14)` |
| error-line | `#E4C2BC` | `oklch(0.8416 0.0398 29.71)` |
| error-bg | `#FCF5F4` | `oklch(0.9753 0.0076 27.23)` |

**Dark**

| Token | Hex | OKLCH |
| --- | --- | --- |
| desk | `#111110` | `oklch(0.1772 0.0020 106.60)` |
| knob | `#0F0F0E` | `oklch(0.1680 0.0020 106.62)` |
| sidebar | `#151514` | `oklch(0.1953 0.0020 106.58)` |
| paper | `#191918` | `oklch(0.2130 0.0019 106.56)` |
| selected | `#232321` | `oklch(0.2554 0.0037 106.66)` |
| hair | `#242422` | `oklch(0.2596 0.0037 106.65)` |
| chip | `#2A2A27` | `oklch(0.2839 0.0054 106.74)` |
| line | `#2E2E2B` | `oklch(0.3001 0.0053 106.71)` |
| toggle-off | `#33332F` | `oklch(0.3197 0.0069 106.79)` |
| control | `#7A7A73` | `oklch(0.5774 0.0105 106.72)` |
| muted | `#9A9A95` | `oklch(0.6847 0.0072 106.59)` |
| sub | `#C8C8C3` | `oklch(0.8314 0.0069 106.55)` |
| ink | `#ECECEA` | `oklch(0.9425 0.0027 106.45)` |
| shared | `#7FA8D4` | `oklch(0.7183 0.0785 250.71)` |
| shared-bg | `#1B2129` | `oklch(0.2454 0.0177 255.69)` |
| shared-line | `#33465C` | `oklch(0.3878 0.0448 252.57)` |
| done | `#5FA98A` | `oklch(0.6776 0.0882 165.01)` |
| done-bg | `#18241F` | `oklch(0.2472 0.0193 167.23)` |
| offline | `#D9A93C` | `oklch(0.7600 0.1343 84.13)` |
| offline-bg | `#2A2413` | `oklch(0.2621 0.0301 90.72)` |
| offline-line | `#3D3520` | `oklch(0.3315 0.0357 89.46)` |
| offline-text / error-text | `#E4C67E` / `#D98A7E` | `oklch(0.8363 0.0970 87.68)` / `oklch(0.7215 0.1070 30.68)` |
| overdue | `#E08B7C` | `oklch(0.7215 0.1070 30.68)` |
| error-line | `#5C3A34` | `oklch(0.3868 0.0499 30.46)` |
