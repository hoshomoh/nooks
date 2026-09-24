import { beforeAll, describe, expect, it, vi } from "vitest"
import { render } from "@testing-library/react"

import { readyForEnglish } from "@/test/i18n"
import { NoteEditor } from "./note-editor"

beforeAll(readyForEnglish)

/** show renders a Note and returns what the editor drew. */
function show(markdown: string) {
  const { container } = render(<NoteEditor initialValue={markdown} onChange={vi.fn()} />)
  return container
}

// The whole reason for ProseMirror rather than a text editor with the markup painted
// over: there is no markup in the document to leak.
describe("what a Note draws", () => {
  it("draws a heading as a heading, not as its hashes", () => {
    const container = show("### Where")

    expect(container.querySelector("h3")?.textContent).toBe("Where")
    expect(container.textContent).not.toContain("###")
  })

  it("draws a quote as a quote", () => {
    const container = show("> They pack up around two.")

    expect(container.querySelector("blockquote")).not.toBeNull()
    expect(container.textContent).not.toContain(">")
  })

  it("draws a code block as one block, with what is inside it", () => {
    const container = show("```\ngrind: filter\n```")

    expect(container.querySelector("pre")?.textContent).toContain("grind: filter")
    expect(container.textContent).not.toContain("```")
  })

  it("draws a checklist as boxes, ticked and unticked", () => {
    const container = show("- [ ] Whole bean\n- [x] Got it")

    const boxes = container.querySelectorAll<HTMLInputElement>('input[type="checkbox"]')
    expect(boxes).toHaveLength(2)
    expect(boxes[0]?.checked).toBe(false)
    expect(boxes[1]?.checked).toBe(true)
    expect(container.textContent).not.toContain("- [")
  })

  /*
   * The editor is the other place a Note is drawn, and the only one that draws real
   * anchors: the preview draws spans, so an address there has nowhere to be an address.
   *
   * What refuses a hostile scheme here is TipTap, not anything written in this
   * repository, which `note-preview.test.tsx` said in a comment and nothing checked. A
   * dependency default is a fine thing to rely on and a poor thing to assume: this is
   * here so that the day it changes is a failing test rather than a live
   * `javascript:` link in somebody's Note.
   */
  it("gives a hostile address nowhere to be an address", () => {
    const container = show("[click me](javascript:alert(1))")

    expect(container.textContent).toContain("click me")
    for (const anchor of container.querySelectorAll("a")) {
      expect(anchor.getAttribute("href") ?? "").not.toContain("javascript:")
    }
    expect(container.innerHTML).not.toContain("javascript:")
  })

  // The ordinary case, so the test above cannot pass by the editor drawing no links.
  it("draws an ordinary address as a link", () => {
    const container = show("[the shop](https://nooks.example)")

    expect(container.querySelector("a")?.getAttribute("href")).toBe("https://nooks.example")
  })

  it("says what an empty Note is for", () => {
    const container = show("")
    expect(container.querySelector("[data-placeholder]")).not.toBeNull()
  })
})

describe("a Note nobody may change", () => {
  it("is not editable", () => {
    const { container } = render(
      <NoteEditor initialValue="Saturday market." onChange={vi.fn()} readOnly />,
    )
    expect(container.querySelector('[contenteditable="false"]')).not.toBeNull()
  })
})
