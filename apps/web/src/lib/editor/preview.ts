import type { JSONContent } from "@tiptap/react"

import { documentFrom } from "./markdown"

/**
 * What a Note looks like on a row, taken from the Note rather than from its source.
 *
 * A Note is stored as markdown, and for a while the first line of a row was produced by
 * looking at that markdown and taking the shorthand off the front of it. That is a
 * second, worse parser: it guesses by prefix, so an empty checklist item came out as
 * "[ ]", a search result read "- [ ] sample content", and anything with `**bold**` in it
 * showed the asterisks.
 *
 * This reads the document instead — through the same function the editor renders from,
 * so what a row shows and what the Note shows cannot disagree. Nothing is summarised:
 * DESIGN.md §13 is explicit that Nooks never summarises a Member, so this is their own
 * first block, and a count of the blocks after it.
 */

/** One run of text in a preview, with whatever was on it. */
export interface PreviewRun {
  text: string
  /** `bold`, `italic`, `strike`, `code`, `link` — whatever the document carried. */
  marks: string[]
  /** Where a link points, when the run is one. */
  href?: string
}

/** A Note as a row shows it. */
export interface NotePreview {
  /** The first block that says anything, as the runs it is made of. */
  runs: PreviewRun[]
  /** How many further blocks say something, for the "+N lines" after it. */
  remaining: number
}

/** NOTHING is a Note with nothing in it, or no Note at all. */
const NOTHING: NotePreview = { runs: [], remaining: 0 }

export function previewOf(markdown: string): NotePreview {
  if (!markdown.trim()) {
    return NOTHING
  }

  const blocks = (documentFrom(markdown).content ?? []).flatMap(saying)
  const [first, ...rest] = blocks
  if (!first) {
    return NOTHING
  }
  return { runs: first, remaining: rest.length }
}

/**
 * saying is the runs a block is made of, or nothing when it says nothing.
 *
 * A checklist is several blocks rather than one, because each item is its own line on
 * the page — and an item with no words is not a line anybody wrote, so it is dropped
 * here rather than shown as an empty row or counted toward what is left.
 */
function saying(node: JSONContent): PreviewRun[][] {
  if (node.type === "taskList") {
    return (node.content ?? []).flatMap(saying)
  }

  const runs = runsIn(node)
  return runs.length > 0 ? [runs] : []
}

/** runsIn walks a block for its text, keeping the marks each run carries. */
function runsIn(node: JSONContent): PreviewRun[] {
  if (node.type === "text") {
    const text = node.text ?? ""
    if (!text) {
      return []
    }
    const marks = node.marks ?? []
    const link = marks.find((mark) => mark.type === "link")
    const href = link?.attrs?.href
    return [
      {
        text,
        marks: marks.map((mark) => mark.type),
        ...(typeof href === "string" ? { href } : {}),
      },
    ]
  }
  return (node.content ?? []).flatMap(runsIn)
}
