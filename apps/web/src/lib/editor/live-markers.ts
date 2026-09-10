import { syntaxTree } from "@codemirror/language"
import { Decoration, EditorView, ViewPlugin } from "@codemirror/view"
import type { DecorationSet, ViewUpdate } from "@codemirror/view"
import type { EditorState, Range } from "@codemirror/state"

import { markerOf, type BlockKind } from "./markers"

/**
 * Renders a Note's markdown as the thing it describes, and hides the markup.
 *
 * DESIGN.md §10: "Markdown shorthand converts a block as it is typed but is never
 * displayed back to the Member." So a line reads as a heading rather than as `###
 * Where`, and a phrase reads as bold rather than as `**this**`.
 *
 * A marker reappears only while the caret is touching that marker itself — not the
 * whole line. That is the narrowest reveal that still leaves a way to take a marker off
 * again: typing `###` shows it, and it disappears as soon as the heading's first word
 * is typed.
 *
 * Only the visible ranges are decorated, so a long Note costs nothing to scroll.
 */
export const liveMarkers = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet

    constructor(view: EditorView) {
      this.decorations = build(view)
    }

    update(update: ViewUpdate) {
      // The caret moving changes which marker is revealed, so selection counts.
      if (update.docChanged || update.viewportChanged || update.selectionSet) {
        this.decorations = build(update.view)
      }
    }
  },
  { decorations: (plugin) => plugin.decorations },
)

/** The class each block type is drawn with. Styling lives in the theme. */
const BLOCK_CLASS: Record<BlockKind, string> = {
  paragraph: "cm-nooks-paragraph",
  heading: "cm-nooks-heading",
  todo: "cm-nooks-todo",
  "todo-done": "cm-nooks-todo-done",
  quote: "cm-nooks-quote",
  code: "cm-nooks-code",
}

/**
 * The markup nodes markdown puts around a phrase, and what the phrase becomes.
 *
 * Every one of these is punctuation the Member wrote to mean something; the meaning is
 * drawn and the punctuation is taken away.
 */
const INLINE_STYLE: Record<string, string> = {
  StrongEmphasis: "cm-nooks-strong",
  Emphasis: "cm-nooks-emphasis",
  Strikethrough: "cm-nooks-struck",
  InlineCode: "cm-nooks-inline-code",
  Link: "cm-nooks-link",
}

/**
 * The node types that are pure punctuation, and the address of a link — which is
 * markup as much as the brackets around it are.
 */
const INLINE_MARKS = new Set([
  "EmphasisMark",
  "CodeMark",
  "StrikethroughMark",
  "LinkMark",
  "URL",
])

/** hidden replaces a marker with nothing at all. */
const hidden = Decoration.replace({})

function build(view: EditorView): DecorationSet {
  const marks: Range<Decoration>[] = []
  for (const { from, to } of view.visibleRanges) {
    addBlockMarks(view.state, from, to, marks)
    addInlineMarks(view.state, from, to, marks)
  }
  // Sorted by CodeMirror rather than by hand: the two passes interleave, and line,
  // mark and replace decorations at one position have an order of their own.
  return Decoration.set(marks, true)
}

/** addBlockMarks draws each line as its block type and hides the shorthand. */
function addBlockMarks(
  state: EditorState,
  from: number,
  to: number,
  marks: Range<Decoration>[],
): void {
  let position = from
  while (position <= to) {
    const line = state.doc.lineAt(position)
    const marker = markerOf(line.text)

    marks.push(Decoration.line({ class: BLOCK_CLASS[marker.kind] }).range(line.from))

    const markerEnd = line.from + marker.length
    if (marker.length > 0 && !touching(state, line.from, markerEnd)) {
      marks.push(hidden.range(line.from, markerEnd))
    }

    position = line.to + 1
  }
}

/** addInlineMarks styles emphasised phrases and hides the punctuation around them. */
function addInlineMarks(
  state: EditorState,
  from: number,
  to: number,
  marks: Range<Decoration>[],
): void {
  // How far the phrase the caret is in reaches. Both ends of a pair are revealed
  // together: seeing `**light` with no closing pair would be worse than seeing neither.
  let revealedTo = -1

  syntaxTree(state).iterate({
    from,
    to,
    enter: (node) => {
      const style = INLINE_STYLE[node.name]
      if (style) {
        marks.push(Decoration.mark({ class: style }).range(node.from, node.to))
        if (touching(state, node.from, node.to)) {
          revealedTo = Math.max(revealedTo, node.to)
        }
        return
      }
      if (!INLINE_MARKS.has(node.name) || node.from === node.to) {
        return
      }
      // A fence's own line is hidden whole by the block pass; hiding its backticks
      // again would be two decorations over the same text.
      if (state.doc.lineAt(node.from).text.startsWith("```")) {
        return
      }
      // The punctuation is hidden unless the caret is in the phrase it belongs to,
      // which is the only way to see what is there and take it off again.
      if (node.from < revealedTo || touching(state, node.from, node.to)) {
        return
      }
      marks.push(hidden.range(node.from, node.to))
    },
  })
}

/** touching reports whether a selection reaches into a range, its edges included. */
function touching(state: EditorState, from: number, to: number): boolean {
  return state.selection.ranges.some((range) => range.from <= to && range.to >= from)
}
