# Nooks — design system

The specification the UI is built to. The values here are exact, not approximate. If a component in
the app disagrees with this file, the component is wrong.

Enforcement lives in `apps/web/src/index.css`: every value below exists there as a custom property,
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

Names here are the **real CSS custom properties** in `apps/web/src/index.css`. Where a colour fills one
of shadcn's semantics it carries shadcn's name; where shadcn has no name for it, it is Nooks' own and
is exposed as a utility through `@theme inline`. §16 explains the contract; this is the palette.

Hex is given because that is what the canvas specifies and what a designer reads. The CSS is OKLCH —
conversions are in §16.

### Light — surfaces and text

| Hex | Token | Job |
| --- | --- | --- |
| `#FFFFFF` | `--background`, `--card`, `--popover` | Paper: panels, main content, dialogs |
| `#FBFBFA` | `--sidebar` | Sidebar and inset panels |
| `#F2F2F0` | `--desk` | The surface behind the app frame |
| `#F0F0EE` | `--hair` | **Row rule** — separates rows |
| `#EFEFED` | `--secondary`, `--muted`, `--accent` | Selected row, hover, active toolbar button |
| `#EDEDEA` | `--chip` | Avatar and chip fill |
| `#E6E6E3` | `--border` | **Block rule** — separates blocks inside a section |
| `#E0E0DC` | `--toggle-off` | Toggle track, off |
| `#D6D6D1` | `--input` | Input border at rest — deliberately darker than `--border` |
| `#8F8F8A` | `--control` | Checkbox and radio borders at rest |
| `#5F5F5B` | `--muted-foreground` | Muted text: counts, timestamps, attribution |
| `#3F3F3C` | `--secondary-foreground` | Secondary text: body copy, explanations |
| `#1F1F1E` | `--foreground`, `--primary` | Ink: primary text, primary button fill |

### Light — the four meaning colours

Each has exactly one job. Nothing else is coloured.

| Hex | Token | Job |
| --- | --- | --- |
| `#37618F` | `--shared`, `--ring` | Shared, links, focus ring, quiet buttons, due dates |
| `#F1F6FB` | `--shared-bg` | Accent fill: badges, the secret panel, a selected share option |
| `#C9DAEC` | `--shared-line` | Accent panel border |
| `#2E7D5B` | `--done` | Done, and the presence of another Member |
| `#F4F8F4` | `--done-bg` | A row just ticked by someone else |
| `#9A6B14` | `--offline` | Offline dot |
| `#FBF6E9` / `#EDE1C4` / `#6B4A0E` | `--offline-bg` / `--offline-line` / `--offline-text` | The offline banner |
| `#A03A2E` | `--overdue` | Overdue dates |
| `#8E3B31` | `--destructive` | Error text, destructive button label |
| `#E4C2BC` | `--destructive-line` | Error and destructive borders |
| `#FCF5F4` | `--destructive-bg` | Rejected input, errored row |

Plus `--knob` (`#FFFFFF`, the toggle knob) and `--scrim` (ink at 34%).

### Dark

A full palette swap, never a filter. Same tokens, different values.

| Token | Light | Dark |
| --- | --- | --- |
| `--desk` | `#F2F2F0` | `#111110` |
| `--background` | `#FFFFFF` | `#191918` |
| `--sidebar` | `#FBFBFA` | `#151514` |
| `--secondary` / `--muted` / `--accent` | `#EFEFED` | `#232321` |
| `--hair` | `#F0F0EE` | `#242422` |
| `--chip` | `#EDEDEA` | `#2A2A27` |
| `--border` / `--input` | `#E6E6E3` / `#D6D6D1` | `#2E2E2B` |
| `--toggle-off` | `#E0E0DC` | `#33332F` |
| `--control` | `#8F8F8A` | `#7A7A73` |
| `--muted-foreground` | `#5F5F5B` | `#9A9A95` |
| `--secondary-foreground` | `#3F3F3C` | `#C8C8C3` |
| `--foreground` / `--primary` | `#1F1F1E` | `#ECECEA` |
| `--shared` / `--ring` | `#37618F` | `#7FA8D4` |
| `--shared-bg` | `#F1F6FB` | `#1B2129` |
| `--shared-line` | `#C9DAEC` | `#33465C` |
| `--done` / `--done-bg` | `#2E7D5B` / `#F4F8F4` | `#5FA98A` / `#18241F` |
| `--offline` | `#9A6B14` | `#D9A93C` |
| `--offline-bg` / `-line` / `-text` | `#FBF6E9` / `#EDE1C4` / `#6B4A0E` | `#2A2413` / `#3D3520` / `#E4C67E` |
| `--overdue` / `--destructive` | `#A03A2E` / `#8E3B31` | `#E08B7C` / `#D98A7E` |
| `--knob` | `#FFFFFF` | `#0F0F0E` |
| `--scrim` | ink at 34% | black at 55% |

**Visitor variant.** On the Public list the checkbox is not interactive, so its border uses a lighter
control: `#B4B4AE` light, `#5A5A55` dark.

**Theme selection.** Light / Dark / System, a three-segment control in Settings → Appearance. System is
the default, and follows the machine live rather than being read once at load.

**Language.** A select beside it in the same Appearance section, listing each language by its own
name, with a note saying which are translated so far. It sets the words, how dates read, and how
numbers and money are written — one choice, not three. An Admin also sets an Instance **default
language** under Instance settings, used for public lists, printed sheets, and anyone who has not
chosen their own.

---

## 3. Type

Two families, **self-hosted and served from the binary** — an Instance may run on a machine with no
internet, and a CDN call would also tell a third party that the Instance exists:

- **Public Sans** (variable) — everything. `--font-sans`.
- **IBM Plex Mono** 400 / 500 — quantities, keycaps, tokens, addresses. `--font-mono`.
  **Never for prose.**

Utility names are given below. A type token must never share its name with a colour
token: Tailwind resolves `text-<name>` to a colour when one exists, so a size called
`input` beside a colour called `input` silently paints the text in the border colour.
That is why the 15px step is `field` and the 13.5px step is `meta`.

| Role | Utility | Size / weight / tracking | Where |
| --- | --- | --- | --- |
| Display | `text-display` | 33 / 600 / −0.025em / lh 1.1 | List titles, view headings |
| Page title | `text-page` | 27 / 600 / −0.02em | Settings pages, states |
| Dialog title | `text-dialog` | 21 / 600 / −0.015em | Dialog headers |
| Section heading | `text-section` | 19 / 600 / −0.01em | Sections within a page |
| Note heading | `text-note-heading` | 18 / 600 / −0.01em / lh 1.4 | Note blocks, full screen |
| Empty title | `text-empty` | 17 / 500 | Empty states |
| Body | `text-body` | 16 / 400 / lh 1.4 | **Item labels — the floor for member-written text** |
| Note body | `text-note` | 15.5 / 400 / lh 1.6 | Note paragraphs, full screen |
| Field | `text-field` | 15 / 400 | Field values |
| Note body, sheet | `text-note-sheet` | 14.5 / 400 / lh 1.6 | Note paragraphs in the side sheet |
| Chrome | `text-chrome` | 14 / 500 | Sidebar items, buttons, field labels |
| Meta | `text-meta` | 13.5 / 400 | Explanations, sidebar rows, hints |
| Small | `text-small` | 13 / 400 | Table cells, meta values |
| Micro | `text-micro` | 12.5 / 400 | Breadcrumbs, attribution, counts |
| Section head | `text-label` | 11.5 / 600 / 0.08em / uppercase | Page section labels |
| Column head | `text-label` | 11.5 / 600 / 0.04em / uppercase | Table columns |
| Mono badge | `font-mono text-label` | 11.5 | Quantities |
| Mono keycap | `font-mono text-keycap` | 10.5 | ⌘K, ↵ |
| Mono secret | `font-mono text-meta` | 13.5 / break-all | Access tokens |

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
| 2px | `--border` | Sections of a page |
| 1px | `--border` | Blocks inside a section |
| 1px | `--hair` | Rows |

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

- Checkbox **17 × 17**, 1.5px `--control` border, radius 4. Hit area is **44px** regardless.
- Label 16px, truncates with ellipsis. **Truncation happens in the label**, never in the metadata, so
  the label always starts at the same x and the row never grows.
- Metadata is right-aligned in the third column, `gap: 12px`, `white-space: nowrap`: quantity badge,
  list name, date, `note` marker, attribution.
- Quantity is a mono 11.5 badge: 1px `--border`, radius 4, padding `1px 5px`.
- A due date takes `--shared`; an overdue date takes `--overdue`.
- Hover fills `--secondary` and darkens the checkbox border to `--muted-foreground`.
- Done: checkbox filled `--muted-foreground` with a white tick; label `--muted-foreground` with
  `line-through` in `#C4C4BE`.
- Just ticked by someone else: fill `--done-bg`, checkbox filled `--done`, attribution
  `--done`. Holds for one second, then settles. **No toast, no sound.**
- A Note previews as one line below the row: 13.5px `--secondary-foreground`, on a 2px `--border` left rule
  indented 10px, in the label column, with `+N lines` after it.

**Add row.** Same grid, `border-top: 1px solid var(--hair)`, a `+` in the checkbox column, placeholder
in `--muted-foreground`, then the date control and a mono `↵` keycap at the right. Enter adds and
keeps focus for the next one, because the common case is adding several.

Nothing but a name is needed: the row must accept `Milk` and Enter.

*Chips.* A quantity or a date written into the sentence is lifted out of it and drawn where the words
were — the sentence and its chips are one field, not a tray beside it. A chip is radius 4 on
`--shared-bg`, the value in `--shared` (mono for a quantity, not for a date), and the kind — `quantity`,
`due` — after it in `--muted-foreground` at 10.5.

*What is recognised.* A quantity is a number with an optional unit (`2`, `1kg`, `250 g`), normalised
to `value unit`. A date is a near-day word, a weekday, or a written date (`tomorrow`, `sat`, `30/8`,
`30 aug`), day first, a bare weekday meaning the **next** one with today excluded. At most one of
each. Only the trailing words are read, and the scan stops at the first word it does not recognise —
which is what keeps `Call 2 plumbers` a name.

*When.* A word becomes a chip only once a space or Enter follows it, never mid-word, so `1` does not
become a quantity while `1kg` is still being typed. Backspace immediately after a chip returns it to
text. A quoted token is never read. **When a token is ambiguous, leave it as text** — a wrong silent
chip is worse than no chip, because the Member has to notice it before they can fix it.

Adding a language translates the near-day words and the units. Weekday and month names come from the
language's own calendar, so nobody retypes a calendar into a locale file.

*Date control.* Always in the row, whether or not a date was typed — parsing is the shortcut, not the
requirement. Three appearances: **empty**, the calendar glyph and "Add a date" in
`--secondary-foreground` on a 1px `--border`; **the view's default**, the word the view supplies
(Today on Today, Tomorrow on Upcoming), which applies unless the Member clears or overrides it; and
**set**, the resolved date in `--shared` on `--shared-bg`, borderless. It and the typed date are one
value, and the last action wins.

Quantity has no permanent control — only the date earned one, because a date is the one field whose
absence changes where the Item appears. Notes, assignment and moving between Lists are all sheet-level.

---

## 7. Controls

### Buttons

| Variant | Spec |
| --- | --- |
| Primary | `--foreground` fill, `--background` text, 500 |
| Secondary | 1px `--border` border, `--secondary-foreground` text |
| Quiet | no border, `--shared` text |
| Destructive | 1px `--destructive`/border border, `--destructive` text |

| Size | Height / padding / radius / text | Allowed in |
| --- | --- | --- |
| Default | 34 / `0 16` / 7 / 14 | Page and section headers, dialog footers |
| Compact | 32 / `0 14` / 7 / 13.5 | Inside a table row or settings row |
| Toolbar | 26 / `0 10` / 6 / 12.5 | The 44px chrome bar only |

**Focus is 2px `--shared` outline at 2px offset**, everywhere, on every focusable thing.

### Inputs

38px high, radius 7, 15px text. Rest 1px `--input`; focus **1.5px** `--shared`;
rejected 1px `--destructive`/border on `--destructive`/bg with `--destructive`. Settings-row inputs are
36px and 280px wide. Labels are 13 / 500 `--secondary-foreground`; hints sit under the field in 12.5
`--muted-foreground`.

### Toggle

34 × 20 track, radius 10, 16px knob, 2px padding. On: `--shared` track, knob right. Off:
`--toggle-off` track, knob left. The word `On` / `Off` sits to its left in 13 `--muted-foreground`.

### Chips and permission lists

Chips pick many of one kind of thing: 30px high, radius 15, 13px, a 13px square swatch at 3px radius
inside. Selected fills `--secondary`; unselected takes a 1px `--border` border.

Use the bordered list instead when each choice needs a sentence explaining what it costs: 1px
`--border`, radius 8, rows padded `12px 14px` split by 1px `--hair`, a 15px checkbox, label 14
and explanation 12.5 `--muted-foreground`.

---

## 8. Page furniture

**Chrome bar** — 44px, 1px `--hair` bottom, padding `0 16px 0 22px`. Breadcrumb left in 12.5
(`--muted-foreground`, current segment `--secondary-foreground`), toolbar buttons right, then the `···` menu at 26 × 26.
It carries location and view-level actions only — **never the page's primary action**.

**Page header** — 27px title, one 14px line of orientation capped at 540px, primary action pinned
right and top-aligned with the title. The action never sits at the bottom of the page.

**Section header** — 2px `--border` above, 18px padding, 19px heading, 13px explanation, secondary
action pinned right.

**Table** — 11.5 uppercase column heads over a 1px `--border` rule; rows min 56px split by 1px
`--hair`; compact buttons inside; the last column is always the `···` menu at 24px. A row that has
never been used greys its name and its last column, **not the whole row**.

**Card** — 1px `--border`, radius 8, padding `15px 16px`. Name row, body, then one
`--hair`-separated footer fact. Cards go two-up in a 12px grid; they never stretch to a
single full-width column.

**Settings row** — `190px 1fr` grid, 24px gap, min-height 48px, split by 1px `--hair`. Label 14
with a 12.5 `--muted-foreground` explanation under it; control right-aligned. The controls a row
carries are an input, a segment, a toggle, a select, or a button with a note to its left. A **select**
is 280 × 36, radius 7, 1px `--border`, the value at 14.5 with a chevron in `--control` at the right.

**Empty state** — 1px **dashed** `--border`, radius 8, padding `30px 26px`, **left-aligned, never
centred**. 17 / 500 statement of fact, then one 14 / 1.6 sentence capped at 440px saying what the thing
is for. No illustration, no centred hero, and **no button if the page header already has one**.

---

## 9. Overlays

**Dialog** — 560px, radius 10, 1px `--border`, shadow `0 24px 60px rgba(0,0,0,0.22)`, over the
scrim. Header `24px 26px 0`: 21px title then one 14 / 1.6 explaining line. Body. Then a
`--hair`-topped footer at `14px 26px`: **consequence text left, Cancel then confirm right.**

Say what happens: *"Ignore removes the request silently and never notifies the sender."* Not *"Are you
sure?"*

**Menu** — 248px, radius 9 (`rounded-menu`, the one step outside the sm–2xl scale), 1px `--border`, 6px padding; items 32px at radius 6 with the shortcut
right in 12 `--muted-foreground`; groups split by a 1px `--hair` rule inset 8px. Destructive items take
`--overdue`.

**Side sheet** — 520px, pinned below the chrome bar, 1px `--border` left border. Same order as full
screen: chrome bar, checkbox and title, detail row, hairline, Note, footer bar. Title 24 / 600, fields
in a wrapping row of 12.5 label + 13 value pairs. **Opening an Item never replaces the List with a
page.**

**Secret shown once** — accent panel (`--shared-bg` on 1px `--shared-line`, radius 8),
mono 13.5 with `word-break: break-all`, a primary Copy button, and a sentence saying plainly that it
cannot be shown again. **Never a warning triangle.**

**Layers.** Three, and no others. A component that needs to sit above something reaches for the layer
that describes it rather than for a bigger number.

| Layer | Where |
| --- | --- |
| `z-10` | A row's own controls, above the invisible target that opens the row |
| `z-20` | The side sheet |
| `z-50` | Anything over a scrim: dialogs, ⌘K |

Nothing is ever raised by inventing a value between them: two things that need to be told apart
belong to different layers, and if they do not, one of them should not be raised at all.

---

## 10. Note blocks

The same markup renders in the side sheet and full screen; only the scale changes. Markdown shorthand
converts a block as it is typed but is **never displayed back to the Member**.

| Block | Full screen | Sheet | Spec |
| --- | --- | --- | --- |
| Heading | 18 / 600 / lh 1.4 | 13.5 | padding `18px 0 3px` |
| Paragraph | 15.5 / lh 1.6 | 14.5 | padding `5px 0` |
| Checklist | 15.5 | 14.5 | 15px box in a 26px gutter, padding `3px 0` |
| Quote | 15.5 `--secondary-foreground` | 14.5 | 2px `--border` left rule, 13px indent, padding `10px 0` |
| Code | mono 13 `--secondary-foreground` | 12.5 | padding `10px 0` |

**A block's shorthand is never shown**, not even under the caret. `### ` disappears the moment the
space that made it a heading is typed, and a line inserted from the `/` menu never shows one at all.
It stays removable: one Backspace takes the space back, the line stops being a heading, and the `###`
is ordinary text again.

**Inline markup is markup too.** `**bold**`, `*italic*`, `~~struck~~`, `` `code` `` and
`[text](url)` are drawn as what they mean and the punctuation is taken away — including a link's
address. These do reappear, because they have no line of their own to leave: a pair shows while the
caret is inside the phrase it wraps, and both ends show together.

**The `/` menu** — 300px, radius 9, 1px `--border`, 6px padding, shadow `0 12px 32px rgba(0,0,0,0.14)`,
opened by `/` on an empty line and filtered as typing continues. An 11.5 uppercase `--muted-foreground`
heading — *Add to the note* — then 34px rows at radius 6, `0 10px`, 11px gap: the glyph in a 20px
centred column at 12.5 `--secondary-foreground`, the label at 13.5, and the markdown shorthand at the
right in mono `--muted-foreground`. The selected row fills `--secondary`. **The menu teaches the
shortcut rather than replacing it.**

---

## 11. States

**Loading** — skeleton rows hold the **exact height of real rows**, so nothing jumps when they arrive.
Blocks are `--secondary` fading to `--hair` down the list. After 4s: *"Still waiting on the
server — 4s. It's probably just waking up."*

**Offline** — a banner above the content: `--offline-bg` on 1px `--offline-line`, a 7px
`--offline` dot, 13.5 `--offline-text`, and a Retry on the right. Ticks made offline carry a
`not synced` badge. Reading, ticking and adding all still work.

**Error on save** — the errored row gets 1px `--destructive`/border on `--destructive`/bg with a `Try
again` in `--destructive`. Below it, a plain sentence saying what the server said and what became
of the Member's text, then real options. Their text is never discarded.

**Conflict** — only competing *text* asks a question, and only on the row it affects. **Ticks never
conflict** — a tick is a tick whoever made it. The two versions are shown as stacked options, the
Member's outlined in accent, with `Keep mine` and `Keep both`.

**Motion** — movement explains where something came from, and is used nowhere else. There
are four:

| Token | Where | Duration |
| --- | --- | --- |
| `animate-sheet-in` | The side sheet, from the edge it is anchored to | 180ms |
| `animate-panel-in` | ⌘K and dialogs, a short drop with the scrim | 160ms |
| `animate-scrim-in` | The ground behind an overlay | 120ms |
| `animate-settle` | Somebody else's tick, holding then settling | 1s |

Hover and focus changes are `transition-colors` at 150ms. Nothing bounces, nothing
springs, and nothing waits for an animation before it responds — every one of these is
decoration on a state that has already changed. A Member whose machine asks for reduced
motion gets none of it.

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
   `--primary`, the `--sidebar-*` family. There is no `--nooks-*` prefixed layer underneath — the tables in §2
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

@theme inline { /* --color-* mappings, incl. Nooks' own */ }

@layer base { /* base styles */ }
```

### The collision, and how it is resolved

**shadcn's `--accent` is a subtle hover surface, not a brand colour.** Nooks' accent is a meaning
colour. Reusing the name would silently restyle every shadcn hover state.

The Foundations artboard labels that swatch **"Shared / accent"**, so Nooks' takes the name **`shared`**
and shadcn keeps `--accent` for its own job. Everywhere §2 says *accent*, the token is `shared`.

### Mapping onto shadcn's semantics

| shadcn variable | Nooks source | Note |
| --- | --- | --- |
| `--background` / `--foreground` | paper / ink | |
| `--card`, `--popover` (+ `-foreground`) | paper / ink | |
| `--primary` / `--primary-foreground` | ink / paper | The primary button is ink on paper |
| `--secondary` / `--secondary-foreground` | selected / sub | |
| `--muted` | selected | shadcn `--muted` is a **surface** |
| `--muted-foreground` | muted `#5F5F5B` | Nooks' "muted" is a **text** colour |
| `--accent` / `--accent-foreground` | selected / ink | shadcn's hover surface — **not** Nooks' accent |
| `--destructive` | error-text | Nooks' destructive button is outlined, not filled |
| `--border` | line | |
| `--input` | input-line `#D6D6D1` | Deliberately darker than `--border` |
| `--ring` | shared | The 2px focus ring |
| `--sidebar` / `--sidebar-foreground` | sidebar / sub | |
| `--sidebar-accent` / `-foreground` | selected / ink | |
| `--sidebar-border` / `--sidebar-ring` | line / shared | |
| `--chart-1` … `--chart-5` | shared, done, offline, overdue, control | Unused; set so shadcn defaults never leak |

### Nooks' own tokens

Added the documented way — declared in `:root` and `.dark`, exposed through `@theme inline` as
`--color-<name>`, which is what generates the `bg-*` / `text-*` / `border-*` utilities:

`shared`, `shared-bg`, `shared-line`, `done`, `done-bg`, `offline`, `offline-bg`,
`offline-line`, `offline-text`, `overdue`, `hair`, `control`, `chip`, `desk`, `toggle-off`, `knob`.

### Radius

shadcn's own components reach for `rounded-sm` / `md` / `lg` / `xl`, and its preset derives those from
one `--radius` by multiplication. Rather than fight that, **Nooks' five radii are written directly onto
that scale**, so a shadcn component gets the right corner without being touched:

| Tailwind | Value | DESIGN.md role |
| --- | --- | --- |
| `rounded-sm` | 4px | Checkbox, small badge |
| `rounded-md` | 6px | Row, toolbar button, menu item |
| `rounded-lg` | 7px | Control: button, input, toggle track |
| `rounded-xl` | 8px | Inner surface: card, empty state, banner |
| `rounded-2xl` | 10px | Surface: panel, dialog |

`--radius` is set to `7px` for anything reading it directly. A chip is a 30px pill — `rounded-full`.

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
