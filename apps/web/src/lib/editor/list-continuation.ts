import { keymap } from "@codemirror/view"
import type { EditorView } from "@codemirror/view"
import type { Extension } from "@codemirror/state"

import { continueList } from "./markers"

/**
 * Enter, on a line that is part of a list.
 *
 * A checklist keeps going without retyping the marker, and **stops when the Member
 * leaves an empty one** — the marker on that line is taken away rather than another
 * being added under it. Without this, ending a quote leaves a stray `>` behind, which
 * is markup a Member never typed and cannot see the point of.
 *
 * Bound above the default keymap, so it decides first and falls through when the line
 * is not part of a list at all.
 */
export const listContinuation: Extension = keymap.of([
  { key: "Enter", run: continueMarkedLine },
])

/** continueMarkedLine handles Enter, reporting whether it did anything. */
function continueMarkedLine(view: EditorView): boolean {
  const range = view.state.selection.main
  if (!range.empty) {
    return false
  }

  const line = view.state.doc.lineAt(range.head)
  const next = continueList(line.text)
  if (next === null) {
    return false
  }

  if (next === "") {
    // An empty marked line means they are done with the list. The line becomes an
    // ordinary empty line, which is what pressing Enter twice has always meant.
    view.dispatch({
      changes: { from: line.from, to: line.to, insert: "" },
      selection: { anchor: line.from },
    })
    return true
  }

  view.dispatch({
    changes: { from: range.head, to: range.head, insert: `\n${next}` },
    selection: { anchor: range.head + 1 + next.length },
    scrollIntoView: true,
  })
  return true
}
