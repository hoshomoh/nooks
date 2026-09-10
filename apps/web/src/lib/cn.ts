import { createCn } from "cn/config"

/**
 * Class merging that knows Nooks' own tokens.
 *
 * `cn` resolves conflicts from a table of Tailwind's built-in names. Nooks adds its own
 * colours and its own type scale, and without them in that table a size and a colour
 * that are both spelled `text-…` look like the same utility: `cn("text-small
 * text-shared")` quietly drops the size. Listing them here is what keeps the design
 * system's names as safe to combine as Tailwind's.
 *
 * The lists are checked against `index.css` by `cn.test.ts`, so a token added there
 * cannot be forgotten here.
 */

/** Every colour Nooks defines beyond the ones shadcn ships. */
export const NOOKS_COLORS = [
  "desk",
  "hair",
  "control",
  "chip",
  "shared",
  "shared-bg",
  "shared-line",
  "done",
  "done-bg",
  "offline",
  "offline-bg",
  "offline-line",
  "offline-text",
  "overdue",
  "destructive-line",
  "destructive-bg",
  "toggle-off",
  "knob",
]

/** Every step of the type scale — DESIGN.md §3. */
export const NOOKS_TEXT_SIZES = [
  "display",
  "page",
  "dialog",
  "section",
  "note-heading",
  "empty",
  "body",
  "note",
  "field",
  "note-sheet",
  "chrome",
  "meta",
  "small",
  "micro",
  "label",
  "keycap",
]

/** Every width the layout is measured in — DESIGN.md §5. */
export const NOOKS_CONTAINERS = [
  "content",
  "calendar",
  "sheet",
  "dialog",
  "form",
]

/** Every height a control is measured in — DESIGN.md §4 and §7. */
export const NOOKS_SPACING = [
  "control",
  "control-compact",
  "control-toolbar",
  "input",
  "row",
  "chrome",
  "chrome-auth",
  "sheet-footer",
]

/** The two together: a spacing token is a valid width as well as a valid height. */
const SIZES = [...NOOKS_CONTAINERS, ...NOOKS_SPACING]

export const cn = createCn({
  extend: {
    classGroups: {
      "font-size": [{ text: NOOKS_TEXT_SIZES }],
      "text-color": [{ text: NOOKS_COLORS }],
      "bg-color": [{ bg: NOOKS_COLORS }],
      "border-color": [{ border: NOOKS_COLORS }],
      w: [{ w: SIZES }],
      "min-w": [{ "min-w": SIZES }],
      "max-w": [{ "max-w": SIZES }],
      h: [{ h: SIZES }],
      "min-h": [{ "min-h": SIZES }],
      "max-h": [{ "max-h": SIZES }],
      size: [{ size: SIZES }],
    },
  },
})
