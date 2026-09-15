import { useState } from "react"
import { beforeAll, describe, expect, it, vi } from "vitest"
import { fireEvent, render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import type { Item } from "@nooks/api"

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
} as Item

/**
 * show renders the sheet the way the List does, returning what it was asked to do.
 *
 * The List owns whether the sheet is leaving, so holding it in a harness is what makes
 * these tests the whole loop: the press, the exit, and the List being told it is over.
 */
function show(item: Item = milk, canEdit = true) {
  const actions = {
    onNoteChange: vi.fn(),
    onToggleDone: vi.fn(),
    onRename: vi.fn(),
    onQuantityChange: vi.fn(),
    onDueChange: vi.fn(),
    onLeave: vi.fn(),
    onClose: vi.fn(),
    onOpenFull: vi.fn(),
  }

  function Harness() {
    const [leaving, setLeaving] = useState(false)
    return (
      <NoteSheet
        item={item}
        crumbs={["Groceries", "Item"]}
        listName="Groceries"
        canEdit={canEdit}
        {...actions}
        leaving={leaving}
        onLeave={() => {
          actions.onLeave()
          setLeaving(true)
        }}
      />
    )
  }

  render(<Harness />)
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

  describe("leaving", () => {
    // The List moves with it, so it has to hear about the close when it starts.
    it("says so the moment it is asked, not when it has finished", async () => {
      const actions = show()

      await userEvent.click(screen.getByRole("button", { name: "Close" }))

      expect(actions.onLeave).toHaveBeenCalled()
      expect(actions.onClose).not.toHaveBeenCalled()
    })

    it("goes back the way it came rather than vanishing", async () => {
      const actions = show()

      await userEvent.click(screen.getByRole("button", { name: "Close" }))

      const sheet = screen.getByRole("complementary")
      expect(sheet).toHaveClass("animate-sheet-out")
      expect(sheet).not.toHaveClass("animate-sheet-in")
      // Still there: a panel that took 180ms to arrive cannot leave between frames.
      expect(actions.onClose).not.toHaveBeenCalled()
    })

    it("tells the List it has gone once the movement is over", async () => {
      const actions = show()

      await userEvent.click(screen.getByRole("button", { name: "Close" }))
      // A real animationend bubbles, and React listens at the root.
      fireEvent.animationEnd(screen.getByRole("complementary"), { bubbles: true })

      expect(actions.onClose).toHaveBeenCalled()
    })

    it("is not the editor settling, or anything else inside it", async () => {
      const actions = show()

      await userEvent.click(screen.getByRole("button", { name: "Close" }))
      // Animation events bubble. The sheet closing on one raised by a control inside
      // it would close under a Member who was still using it.
      fireEvent.animationEnd(screen.getByRole("button", { name: /Open full/ }), {
        bubbles: true,
      })

      expect(actions.onClose).not.toHaveBeenCalled()
    })

    it("stops answering while it is on its way out", async () => {
      show()

      await userEvent.click(screen.getByRole("button", { name: "Close" }))

      expect(screen.getByRole("complementary")).toHaveClass("pointer-events-none")
    })
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

describe("a long name", () => {
  it("wraps in the sheet rather than scrolling out of sight", () => {
    const long = "Ethiopian whole bean, the light roast from the place on the corner"
    show({ ...milk, label: long } as Item)

    const title = screen.getByRole("textbox", { name: "Item name" })
    // A textarea that sizes to its content, not a one-line input that hides the rest.
    expect(title.tagName).toBe("TEXTAREA")
    expect(title).toHaveClass("field-sizing-content")
    expect(title).toHaveValue(long)
  })
})
