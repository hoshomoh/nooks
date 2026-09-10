import { autocompletion } from "@codemirror/autocomplete"
import type { CompletionContext, CompletionResult, Completion } from "@codemirror/autocomplete"
import type { Extension } from "@codemirror/state"

import { INSERTION, SHORTHAND } from "./markers"

/** What the / menu offers, and how each entry reads. */
export type SlashOption = {
  /** The block type, used to look up its shorthand and what to insert. */
  kind: keyof typeof SHORTHAND
  /** The entry's label, already translated. */
  label: string
  /** The glyph shown at the left, per the design's block menu. */
  glyph: string
}

export type SlashMenuOptions = {
  options: readonly SlashOption[]
}

/**
 * The `/` menu, per DESIGN.md §10.
 *
 * Every entry shows the markdown shorthand beside it, so the menu teaches the shortcut
 * rather than replacing it — a Member who learns `### ` stops needing the menu.
 *
 * It only opens on a `/` that begins a line: a slash inside a sentence is a slash.
 */
export function slashMenu({ options }: SlashMenuOptions): Extension {
  const completions: Completion[] = options.map((option) => ({
    label: `${option.glyph}  ${option.label}`,
    // The shorthand appears at the right of each entry.
    detail: SHORTHAND[option.kind],
    apply: INSERTION[option.kind],
    type: "keyword",
  }))

  return autocompletion({
    override: [(context) => slashCompletions(context, completions)],
    // The menu is the only completion source here, so it may open on its own.
    activateOnTyping: true,
    icons: false,
  })
}

/** slashCompletions offers the block types when a line begins with a slash. */
function slashCompletions(
  context: CompletionContext,
  completions: Completion[],
): CompletionResult | null {
  const line = context.state.doc.lineAt(context.pos)
  const before = line.text.slice(0, context.pos - line.from)

  // Only a slash that starts the line opens the menu; one inside a sentence is a slash.
  const match = /^\/(\w*)$/.exec(before)
  if (!match) {
    return null
  }

  const typed = (match[1] ?? "").toLowerCase()
  return {
    // Replacing from the slash means picking an entry removes it.
    from: line.from,
    options: completions.filter((option) => option.label.toLowerCase().includes(typed)),
    // Matching is ours: CodeMirror would score the entries against the "/" as well, and
    // no block type is spelled with one.
    filter: false,
    validFor: /^\/\w*$/,
  }
}
