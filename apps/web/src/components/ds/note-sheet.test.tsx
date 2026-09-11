/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import type { Item } from "@nooks/api"

import "@/test/dom"
import { readyForEnglish } from "@/test/i18n"
import { NoteSheet } from "./note-sheet"

beforeAll(readyForEnglish)

/** milk is an Item with nothing on it but a name. */
const milk = {
  uid: "item_milk",
  label: "Milk",
  quantity: "",
  dueOn: "",
  done: false,
  addedByName: "Anna",
  note: "",
  noteFirstLine: "",
  noteRemainingLines: 0,
} as Item

/** show renders the sheet, returning what it was asked to do. */
function show(item: Item = milk, canEdit = true) {
  const actions = {
    onNoteChange: vi.fn(),
    onToggleDone: vi.fn(),
    onRename: vi.fn(),
    onQuantityChange: vi.fn(),
    onDueChange: vi.fn(),
    onClose: vi.fn(),
    onOpenFull: vi.fn(),
  }
  render(<NoteSheet item={item} crumbs={["Groceries", "Item"]} listName="Groceries" canEdit={canEdit} {...actions} />)
  return actions
}

describe("the item sheet", () => {
  // The crumb already says where this is; a second word beside "Open full" would read
  // as a second destination.
  it("closes with an icon, named for a screen reader", () => {
    const actions = show()
    const close = screen.getByRole("button", { name: "Close" })

    expect(close).toHaveTextContent("")
    expect(close.querySelector("svg")).not.toBeNull()
    expect(screen.getByRole("button", { name: /Open full/ })).toBeInTheDocument()
    void actions
  })

  // Every field here is the control that changes it: there is no edit mode.
  it("changes the quantity where it is read", async () => {
    const actions = show()

    const quantity = screen.getByRole("textbox", { name: "Quantity" })
    await userEvent.type(quantity, "1 kg{Enter}")

    expect(actions.onQuantityChange).toHaveBeenCalledWith("1 kg")
  })

  it("says what an empty quantity is for", () => {
    show()
    expect(screen.getByRole("textbox", { name: "Quantity" })).toHaveAttribute(
      "placeholder",
      "Add",
    )
  })

  // That it is the design system's calendar rather than the browser's is asserted
  // where the control lives, in add-row.test.tsx. Here the question is only whether
  // the sheet opens it at all.
  it("opens the date control", async () => {
    show()

    const control = screen.getByRole("button", { name: "Due date" })
    expect(control).toHaveAttribute("aria-expanded", "false")

    await userEvent.click(control)

    expect(control).toHaveAttribute("aria-expanded", "true")
  })

  it("is read-only for somebody who may not change it", () => {
    show(milk, false)
    expect(screen.queryByRole("textbox", { name: "Quantity" })).not.toBeInTheDocument()
  })
})
