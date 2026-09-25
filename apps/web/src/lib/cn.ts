import { createCn } from "cn/config"

/**
 * Class merging that knows nooks' own tokens.
 *
 * `cn` resolves conflicts from a table of Tailwind's built-in names. Without nooks'
 * colours and type scale in that table, a size and a colour both spelled `text-…` look
 * like the same utility, and `cn("text-small text-shared")` quietly drops the size.
 *
 * cn.test.ts checks the lists against index.css, so a token added there cannot be
 * forgotten here.
 */

/** Every colour nooks defines beyond the ones shadcn ships. */
export const NOOKS_COLORS = [
  "desk",
  "hair",
  "control",
  "chip",
  "shared",
  "shared-bg",
  "shared-line",
  "done",
  "done-strike",
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
  "note-heading-sheet",
  "sheet-title",
  "empty",
  "body",
  "note",
  "field",
  "note-sheet",
  "chrome",
  "meta",
  "small",
  "micro",
  "badge",
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
  "settings",
]

/**
 * Every named spacing token — DESIGN.md §4 and §7.
 *
 * Mostly the heights a control is measured in, and one that is not: the room the
 * content pane keeps clear of the side sheet. tailwind-merge has to know all of them
 * or it cannot tell that two classes built from them are the same property, and the
 * test beside this list fails when a token is added to the design and not to here.
 */
export const NOOKS_SPACING = [
  "control",
  "control-compact",
  "control-settings",
  "control-toolbar",
  "input",
  "row",
  "chrome",
  "chrome-auth",
  "sheet-footer",
  "sheet-clear",
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
