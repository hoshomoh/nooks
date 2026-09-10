import { Decoration, EditorView, ViewPlugin } from "@codemirror/view"
import type { DecorationSet, ViewUpdate } from "@codemirror/view"
import { RangeSetBuilder } from "@codemirror/state"

import { markerOf, type BlockKind } from "./markers"

/**
 * Renders a Note's lines as blocks, and hides the shorthand that made them.
 *
 * DESIGN.md §10: "Markdown shorthand converts a block as the Member types but is never
 * displayed back to them." So a line reads as a heading rather than as `### Where` —
 * except on the line the cursor is on, where hiding the marker would make the text
 * jump under the caret and leave no way to remove it.
 *
 * Only the visible ranges are decorated, because a long Note should not cost anything
 * to scroll.
 */
export const liveMarkers = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet

    constructor(view: EditorView) {
      this.decorations = build(view)
    }

    update(update: ViewUpdate) {
      // The cursor moving changes which marker is revealed, so selection counts.
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

/** hidden replaces a marker with nothing at all. */
const hidden = Decoration.replace({})

function build(view: EditorView): DecorationSet {
  const builder = new RangeSetBuilder<Decoration>()
  const cursorLine = view.state.doc.lineAt(view.state.selection.main.head).number

  for (const { from, to } of view.visibleRanges) {
    let position = from
    while (position <= to) {
      const line = view.state.doc.lineAt(position)
      const marker = markerOf(line.text)

      builder.add(
        line.from,
        line.from,
        Decoration.line({ class: BLOCK_CLASS[marker.kind] }),
      )

      // The marker stays visible on the line being edited: hiding it there would move
      // the text under the caret and leave no way to take the marker off again.
      if (marker.length > 0 && line.number !== cursorLine) {
        builder.add(line.from, line.from + marker.length, hidden)
      }

      position = line.to + 1
    }
  }
  return builder.finish()
}
