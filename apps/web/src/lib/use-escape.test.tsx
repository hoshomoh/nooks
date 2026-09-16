import { describe, expect, it, vi } from "vitest"
import { fireEvent, render, screen } from "@testing-library/react"

import { useEscape } from "./use-escape"

/** screenWith renders something that backs out on Escape, over a field that also takes it. */
function screenWith(onEscape: () => void, handledByTheField: boolean) {
  function Screen() {
    useEscape(onEscape)
    return (
      <input
        aria-label="Name"
        onKeyDown={(event) => {
          if (event.key === "Escape" && handledByTheField) {
            event.preventDefault()
          }
        }}
      />
    )
  }
  render(<Screen />)
  return screen.getByRole("textbox", { name: "Name" })
}

describe("backing out with Escape", () => {
  it("leaves the screen when nothing else wanted the key", () => {
    const onEscape = vi.fn()
    const field = screenWith(onEscape, false)

    fireEvent.keyDown(field, { key: "Escape" })

    expect(onEscape).toHaveBeenCalledOnce()
  })

  /*
   * Escape means "back out of the thing I am in", and the innermost thing wins.
   *
   * Without this, Escape out of a rename put the old name back and left the screen as
   * well, which is two answers to one key and only one of them asked for.
   */
  it("stops at the control that took it first", () => {
    const onEscape = vi.fn()
    const field = screenWith(onEscape, true)

    fireEvent.keyDown(field, { key: "Escape" })

    expect(onEscape).not.toHaveBeenCalled()
  })

  it("ignores every other key", () => {
    const onEscape = vi.fn()
    const field = screenWith(onEscape, false)

    fireEvent.keyDown(field, { key: "Enter" })
    fireEvent.keyDown(field, { key: "a" })

    expect(onEscape).not.toHaveBeenCalled()
  })
})
