/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"

import "@/test/dom"
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
