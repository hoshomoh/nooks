/** @vitest-environment jsdom */
import { afterEach, describe, expect, it } from "vitest"
import { EditorState } from "@codemirror/state"
import { EditorView } from "@codemirror/view"
import { markdown } from "@codemirror/lang-markdown"

import { liveMarkers } from "./live-markers"

let view: EditorView | null = null

afterEach(() => {
  view?.destroy()
  view = null
})

/** shown is what a Member actually reads, with every hidden marker taken out. */
function shown(doc: string, caretAt = doc.length): string {
  view?.destroy()
  view = new EditorView({
    parent: document.body,
    state: EditorState.create({
      doc,
      selection: { anchor: caretAt },
      extensions: [markdown(), liveMarkers],
    }),
  })
  return view.contentDOM.textContent ?? ""
}

// DESIGN.md §10: markdown converts a block as it is typed but is never displayed back.
describe("what a Note shows", () => {
  it("draws a heading without its hashes", () => {
    expect(shown("### Where\nSaturday market.", 20)).toBe("WhereSaturday market.")
  })

  it("draws a checklist without its brackets", () => {
    expect(shown("- [ ] Ethiopian, whole bean\n\n", 29)).toContain("Ethiopian, whole bean")
    expect(shown("- [ ] Ethiopian, whole bean\n\n", 29)).not.toContain("- [ ]")
  })

  it("draws a quote without its angle bracket", () => {
    expect(shown("> They pack up around two.\n\n", 28)).not.toContain(">")
  })

  it("draws bold without its asterisks", () => {
    expect(shown("Ask for the **light roast**.\n\n", 30)).toBe("Ask for the light roast.")
  })

  it("draws inline code without its backticks", () => {
    expect(shown("Say `250 g` at the counter.\n\n", 29)).toBe("Say 250 g at the counter.")
  })

  it("draws a link as its words, not its address", () => {
    expect(shown("See [the map](https://brunnen.lan/map).\n\n", 41)).toBe("See the map.")
  })
})

// A marker has to be reachable, or there would be no way to take it off again.
describe("what the caret reveals", () => {
  it("shows the marker it is sitting in", () => {
    expect(shown("### Where", 2)).toBe("### Where")
  })

  it("hides it again once the heading is being written", () => {
    expect(shown("### Where", 9)).toBe("Where")
  })

  it("reveals only the marker the caret touches, not the whole line", () => {
    expect(shown("The **light** and *dark* roasts", 6)).toBe("The **light** and dark roasts")
  })
})
