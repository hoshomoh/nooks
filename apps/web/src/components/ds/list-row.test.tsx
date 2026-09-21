import { beforeAll, describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import { readyForEnglish } from "@/test/i18n"
import { ListRow, type ListRowLabels } from "./list-row"

beforeAll(readyForEnglish)

const labels: ListRowLabels = {
  name: "Item name",
  open: "Open item",
  quantity: "Quantity",
  due: "Add a date", notSynced: "not synced",
}

describe("the list row", () => {
  // A token is somebody's access narrowed, never an identity of its own, so a row says
  // who did it and only then what through.
  it("names the Member as well as the token an Item came through", () => {
    render(<ListRow label="Milk" addedByName="Anna" addedVia="via Kitchen tablet" labels={labels} />)

    expect(screen.getByText("Anna")).toBeInTheDocument()
    expect(screen.getByText("via Kitchen tablet")).toBeInTheDocument()
  })

  it("says nothing about a token when a browser added the Item", () => {
    render(<ListRow label="Milk" addedByName="Anna" labels={labels} />)

    expect(screen.getByText("Anna")).toBeInTheDocument()
    expect(screen.queryByText(/via/)).not.toBeInTheDocument()
  })

  it("opens the Item from anywhere in the row, not only its label", async () => {
    const onOpen = vi.fn()
    render(<ListRow label="Milk" onOpen={onOpen} labels={labels} />)

    await userEvent.click(screen.getByRole("button", { name: "Open item" }))

    expect(onOpen).toHaveBeenCalledOnce()
  })

  it("ticks without opening the Item", async () => {
    const onOpen = vi.fn()
    const onToggle = vi.fn()
    render(<ListRow label="Milk" onOpen={onOpen} onToggle={onToggle} labels={labels} />)

    await userEvent.click(screen.getByRole("checkbox"))

    // Base UI hands the handler the event alongside the new value, so only the first
    // argument is what this is about.
    expect(onToggle.mock.calls[0]?.[0]).toBe(true)
    expect(onOpen).not.toHaveBeenCalled()
  })

  /*
   * A List shared read-only shows what was done and does not offer to change it.
   *
   * The checkbox drew itself whatever it was given. On a read-only List the Member
   * could press it, the request went, and the Instance refused — a control that looked
   * live and was not. The note sheet and the note page had both been gating theirs on
   * the same answer all along; the row was the one that had not.
   */
  it("does not offer to tick when there is nothing to tick with", async () => {
    render(<ListRow label="Milk" done labels={labels} />)

    // Base UI draws a checkbox as a span with the state in ARIA rather than as a
    // native input, so this is what "disabled" is on one.
    const box = screen.getByRole("checkbox")
    expect(box).toHaveAttribute("aria-disabled", "true")

    // Still says what happened to the Item, which is what read-only is for.
    expect(box).toBeChecked()

    await userEvent.click(box)
    expect(box).toBeChecked()
  })

  // The settle animation is for somebody else's tick, so without this the Member's own
  // is the one action in the app that answers with nothing.
  it("answers the Member's own tick, and only that much", () => {
    render(<ListRow label="Milk" onToggle={vi.fn()} labels={labels} />)

    const box = screen.getByRole("checkbox")
    expect(box).toHaveClass("active:scale-[0.97]")
    expect(box.className).toMatch(/transition-\[background-color,border-color,transform\]/)
    expect(box.className).not.toMatch(/animate-/)
  })

  it("renames the Item where it is read", async () => {
    const onRename = vi.fn()
    render(<ListRow label="Milk" onRename={onRename} labels={labels} />)

    const name = screen.getByRole("textbox", { name: "Item name" })
    await userEvent.clear(name)
    await userEvent.type(name, "Oat milk{Enter}")

    expect(onRename).toHaveBeenCalledWith("Oat milk")
  })

  it("is plain text when the Member may not change it", () => {
    render(<ListRow label="Milk" labels={labels} />)

    expect(screen.queryByRole("textbox")).not.toBeInTheDocument()
    expect(screen.getByText("Milk")).toBeInTheDocument()
  })

  // Two regressions this catches, both from making the `···` appear on hover.
  describe("the row's own menu", () => {
    it("keeps the metadata on screen, rather than swapping it for the menu", () => {
      render(
        <ListRow
          label="Milk"
          addedByName="Anna"
          dueLabel="Fri"
          labels={labels}
          menu={<button type="button">More</button>}
        />,
      )

      expect(screen.getByText("Anna")).toBeInTheDocument()
      expect(screen.getByText("Fri")).toBeInTheDocument()
    })

    // A trigger that is only mounted on hover takes with it the element its menu is
    // positioned against, and the menu jumps to the corner of the page.
    it("keeps the trigger mounted whether or not the pointer is over the row", () => {
      render(
        <ListRow label="Milk" labels={labels} menu={<button type="button">More</button>} />,
      )

      expect(screen.getByRole("button", { name: "More" })).toBeInTheDocument()
    })
  })

  /*
   * A Note is marked on the row, not quoted on it.
   *
   * The row used to carry the Note's first line and a count of the rest, which made
   * every row with a Note two lines tall and a List of them hard to run an eye down.
   * The Note is read in the sheet; the row only says there is one.
   */
  it("marks an Item that carries a Note, without quoting it", () => {
    const { container } = render(<ListRow label="Coffee" hasNote labels={labels} />)

    expect(container.querySelector("svg")).not.toBeNull()
    expect(screen.getByText("Coffee")).toBeInTheDocument()
  })

  it("leaves the mark off an Item with no Note", () => {
    const { container } = render(<ListRow label="Milk" labels={labels} />)

    expect(container.querySelector("svg")).toBeNull()
  })

  describe("aiming at a row", () => {
    it("opens the Item when the label itself is clicked", async () => {
      // The label is most of a 44px row. Asking a Member to find the gap beside it
      // wastes the target and is the thing that made rows hard to hit.
      let opened = 0
      render(
        <ListRow
          label="Milk"
          labels={labels}
          onOpen={() => (opened += 1)}
          onRename={() => {}}
        />,
      )

      await userEvent.click(screen.getByRole("textbox", { name: labels.name }))

      expect(opened).toBe(1)
    })

    it("does not take the caret on a single click", async () => {
      render(<ListRow label="Milk" labels={labels} onOpen={() => {}} onRename={() => {}} />)
      const field = screen.getByRole("textbox", { name: labels.name })

      await userEvent.click(field)

      expect(field).not.toHaveFocus()
    })

    it("hands over the field on a double click", async () => {
      render(<ListRow label="Milk" labels={labels} onOpen={() => {}} onRename={() => {}} />)
      const field = screen.getByRole("textbox", { name: labels.name })

      await userEvent.dblClick(field)

      expect(field).toHaveFocus()
    })

    it("does not select the whole line on a double click", async () => {
      // Somebody who double-clicked a word wants to edit at that word. Selecting
      // everything means their next keystroke replaces the lot.
      render(<ListRow label="Milk and bread" labels={labels} onOpen={() => {}} onRename={() => {}} />)
      const field = screen.getByRole("textbox", { name: labels.name }) as HTMLInputElement

      await userEvent.dblClick(field)

      expect(field.selectionStart === 0 && field.selectionEnd === field.value.length).toBe(false)
    })

    it("says with the pointer that a click opens the row", () => {
      // An I-beam over something that does not take the caret is the control lying.
      render(<ListRow label="Milk" labels={labels} onOpen={() => {}} onRename={() => {}} />)

      expect(screen.getByRole("textbox", { name: labels.name })).toHaveClass("cursor-pointer")
    })

    it("still edits on a single click where there is nothing to open", async () => {
      // A List title is the only thing on its line, so a click there means edit.
      render(<ListRow label="Milk" labels={labels} onRename={() => {}} />)
      const field = screen.getByRole("textbox", { name: labels.name })

      await userEvent.click(field)

      expect(field).toHaveFocus()
    })
  })

  it("says when a change has not reached the Instance", () => {
    render(<ListRow label="Milk" labels={labels} notSynced />)

    expect(screen.getByText(labels.notSynced)).toBeInTheDocument()
  })
})

describe("a long label on a row", () => {
  it("stays one line, because a row never grows", () => {
    // DESIGN.md §6. A title in the sheet wraps; a row cannot, or the list changes
    // height as somebody types and stops being something you can aim at.
    render(
      <ListRow
        label="Ethiopian whole bean, the light roast from the corner shop"
        onRename={() => {}}
        labels={labels}
      />,
    )

    const label = screen.getByRole("textbox", { name: "Item name" })
    expect(label.tagName).toBe("INPUT")
  })
})
