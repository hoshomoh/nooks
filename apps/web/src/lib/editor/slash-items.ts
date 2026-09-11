import type { Editor, Range } from "@tiptap/react"

import type { Translate } from "@/lib/translate"

/** One entry in the `/` menu. */
export interface SlashItem {
  /** The block type, which is also its key. */
  kind: SlashKind
  /** What the entry reads, already translated. */
  label: string
  /** The glyph at the left, per the design's block menu. */
  glyph: string
  /** The markdown shorthand shown at the right, so the menu teaches the shortcut. */
  shorthand: string
}

/** The block types the `/` menu offers. */
export type SlashKind = "todo" | "heading" | "quote" | "code"

/**
 * The blocks a Member can reach from `/`, in the order the design lists them.
 *
 * Pure, so what the menu offers can be read here rather than run — and so the words
 * come from the Member's language rather than from this file.
 */
export function slashItems(t: Translate): SlashItem[] {
  return [
    { kind: "todo", label: t("note.blocks.todo"), glyph: "☐", shorthand: "- [ ]" },
    { kind: "heading", label: t("note.blocks.heading"), glyph: "H", shorthand: "###" },
    { kind: "quote", label: t("note.blocks.quote"), glyph: "❝", shorthand: ">" },
    { kind: "code", label: t("note.blocks.code"), glyph: "{}", shorthand: "```" },
  ]
}

/** matching narrows the menu to what has been typed after the slash. */
export function matching(items: SlashItem[], query: string): SlashItem[] {
  const wanted = query.trim().toLocaleLowerCase()
  return items.filter((item) => item.label.toLocaleLowerCase().includes(wanted))
}

/**
 * insert turns the line into the block that was chosen.
 *
 * The slash and whatever was typed after it go with it: they were the instruction, not
 * the Member's words.
 */
export function insert(editor: Editor, range: Range, kind: SlashKind): void {
  const chain = editor.chain().focus().deleteRange(range)

  switch (kind) {
    case "todo":
      chain.toggleTaskList().run()
      return
    case "heading":
      chain.setNode("heading", { level: 3 }).run()
      return
    case "quote":
      chain.toggleBlockquote().run()
      return
    case "code":
      chain.toggleCodeBlock().run()
      return
  }
}
