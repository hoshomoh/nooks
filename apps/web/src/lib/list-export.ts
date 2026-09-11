import type { Item, List } from "@nooks/api"

/**
 * A List as plain text, per DESIGN.md §13: the same markdown a Note is written in.
 *
 * Export is not a format of its own. A Member who exports a List gets something they
 * can paste into a message, a Note, or any other app that reads a checklist — which is
 * what "export" is for. Nothing generated is added: no header, no date, no tool name.
 */
export function listAsText(list: List, items: readonly Item[]): string {
  const lines = [`# ${list.name}`, ""]
  for (const item of items) {
    lines.push(`- [${item.done ? "x" : " "}] ${withQuantity(item)}`)
  }
  return `${lines.join("\n")}\n`
}

/** withQuantity writes an Item the way its row reads it. */
function withQuantity(item: Item): string {
  return item.quantity ? `${item.label} — ${item.quantity}` : item.label
}

/** A file to hand to the Member, named and typed. */
export interface TextFile {
  name: string
  /** The text itself, ready to be written. */
  body: string
}

/**
 * exportFileFor names the file a List is saved as.
 *
 * Pure, so the naming rules are testable without a browser: whoever is holding a
 * document decides what to do with it.
 */
export function exportFileFor(list: List, items: readonly Item[]): TextFile {
  return { name: `${slugOf(list.name)}.md`, body: listAsText(list, items) }
}

/** slugOf makes a name safe to be a filename, in any language. */
function slugOf(name: string): string {
  const cleaned = name
    .toLocaleLowerCase()
    .replace(/[^\p{L}\p{N}]+/gu, "-")
    .replace(/^-+|-+$/g, "")
  return cleaned || "list"
}
