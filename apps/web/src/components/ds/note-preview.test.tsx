import { describe, expect, it } from "vitest"
import { render, screen } from "@testing-library/react"

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

  /*
   * The reason a link is not a target here is that the row is already one. The reason it
   * matters is this: a Note is written by one Member and read by everybody the List
   * reaches, and markdown will carry any address at all.
   *
   * `documentFrom` keeps `javascript:` in the document — it parses, it does not sanitise
   * — so what stops it being a live anchor is that this draws spans. TipTap's own render
   * refuses a disallowed scheme too, and the editor is the only other place a Note is
   * drawn, but that is a dependency's default rather than anything written down here.
   */
  it("gives a hostile address nowhere to be an address", () => {
    const { container } = render(
      <NotePreview runs={previewOf("[click me](javascript:alert(1))").runs} />,
    )

    expect(screen.getByText("click me")).toBeInTheDocument()
    expect(screen.queryByRole("link")).not.toBeInTheDocument()
    expect(container.querySelector("[href]")).toBeNull()
    expect(container.innerHTML).not.toContain("javascript:")
  })

  it("shows nothing at all for a Note that says nothing", () => {
    const { container } = render(<NotePreview runs={previewOf("- [ ] ").runs} />)

    expect(container.textContent).toBe("")
  })
})
