/** @vitest-environment jsdom */
import { describe, expect, it } from "vitest"
import { render, screen } from "@testing-library/react"

import "@/test/dom"
import { NotePreview } from "./note-preview"
import { previewOf } from "@/lib/editor/preview"

describe("a Note on a row", () => {
  it("shows the words without the punctuation they were written with", () => {
    // "- [ ] sample content" in a search result and "[ ]" on a row were both this.
    render(<NotePreview runs={previewOf("- [ ] sample content").runs} />)

    expect(screen.getByText("sample content")).toBeInTheDocument()
    expect(screen.queryByText(/\[ \]/)).not.toBeInTheDocument()
  })

  it("draws a mark rather than printing it", () => {
    render(<NotePreview runs={previewOf("**loud**").runs} />)

    const run = screen.getByText("loud")
    expect(run).toHaveClass("font-semibold")
    expect(screen.queryByText(/\*\*/)).not.toBeInTheDocument()
  })

  it("draws a link as one without making it a target", () => {
    // The row is already one target. A link inside it is a second thing to hit.
    render(<NotePreview runs={previewOf("[the docs](https://nooks.test)").runs} />)

    expect(screen.getByText("the docs")).toHaveClass("text-shared")
    expect(screen.queryByRole("link")).not.toBeInTheDocument()
  })

  it("shows nothing at all for a Note that says nothing", () => {
    const { container } = render(<NotePreview runs={previewOf("- [ ] ").runs} />)

    expect(container.textContent).toBe("")
  })
})
