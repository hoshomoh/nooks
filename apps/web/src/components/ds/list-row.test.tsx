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
  // The person is a circle with one letter, and the whole name is what it says.
  it("ends in the circle of whoever added it", () => {
    render(<ListRow label="Milk" person={{ name: "Anna", said: "Anna" }} labels={labels} />)

    const circle = screen.getByRole("img", { name: "Anna" })
    expect(circle).toHaveTextContent("A")
  })

  // A token is somebody's access narrowed, never an identity of its own, so the circle
  // is the Member's and the token is said after their name.
  it("names the Member as well as the token an Item came through", () => {
    render(
      <ListRow
        label="Milk"
        person={{ name: "Anna", said: "Anna via Kitchen tablet", viaToken: true }}
        labels={labels}
      />,
    )

    expect(screen.getByRole("img", { name: "Anna via Kitchen tablet" })).toBeInTheDocument()
  })

  // A quantity is plain words, not a badge, and a long one is cut rather than widening
  // the row, with the whole of it a hover away.
  it("says a quantity as plain text, cut short when it is long", () => {
    render(<ListRow label="Coffee" quantity="H · E-invoice, the long estimate" labels={labels} />)

    const quantity = screen.getByText("H · E-invoice, the long estimate")
    expect(quantity).toHaveAttribute("title", "H · E-invoice, the long estimate")
    expect(quantity).toHaveClass("truncate", "max-w-30")
    expect(quantity.className).not.toMatch(/border|font-mono/)
  })

  it("gives a quantity less room while the sheet is open", () => {
    render(<ListRow label="Coffee" quantity="250 g" layout={{ narrow: true }} labels={labels} />)

    expect(screen.getByText("250 g")).toHaveClass("max-w-24")
  })

  // DESIGN.md §6: the note slot is kept on every row of a List with any Note, so the
  // marks line up, and it is the last thing on the row.
  it("keeps the note slot on a row with no Note, where the List has some", () => {
    const { container } = render(<ListRow label="Milk" layout={{ notes: true }} labels={labels} />)
    const last = container.querySelector(".grid-cols-\\[20px_1fr_auto\\] > :last-child > :last-child")

    expect(last).toHaveClass("w-4")
    expect(last?.querySelector("svg")).toBeNull()
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
    // The `···` takes the person's place on hover, and nothing else moves: the date
    // stays where it was read.
    it("trades places with the person, and leaves the rest where it was", () => {
      render(
        <ListRow
          label="Milk"
          person={{ name: "Anna", said: "Anna" }}
          dueLabel="Fri"
          labels={labels}
          menu={<button type="button">More</button>}
        />,
      )

      const circle = screen.getByRole("img", { name: "Anna" })
      const menu = screen.getByRole("button", { name: "More" })
      expect(circle.parentElement?.parentElement).toBe(menu.parentElement?.parentElement)
      expect(circle.parentElement).toHaveClass("group-hover:opacity-0")
      expect(menu.parentElement).toHaveClass("opacity-0", "group-hover:opacity-100")
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

describe("moving a row by its keys", () => {
  // DESIGN.md §12: ⌥↑ and ⌥↓ move the focused Item, wherever in the row focus is.
  it("moves up and down with ⌥ and the arrows", async () => {
    const onMove = vi.fn()
    render(<ListRow label="Milk" onRename={vi.fn()} onMove={onMove} labels={labels} />)

    await userEvent.click(screen.getByRole("textbox", { name: "Item name" }))
    await userEvent.keyboard("{Alt>}{ArrowUp}{/Alt}")
    await userEvent.keyboard("{Alt>}{ArrowDown}{/Alt}")

    expect(onMove.mock.calls).toEqual([["up"], ["down"]])
  })

  // A bare arrow is the caret's, and moving between rows; only ⌥ moves the Item.
  it("leaves the arrows alone without ⌥", async () => {
    const onMove = vi.fn()
    render(<ListRow label="Milk" onRename={vi.fn()} onMove={onMove} labels={labels} />)

    await userEvent.click(screen.getByRole("textbox", { name: "Item name" }))
    await userEvent.keyboard("{ArrowUp}{ArrowDown}")

    expect(onMove).not.toHaveBeenCalled()
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
