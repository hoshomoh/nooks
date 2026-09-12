/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import "@/test/dom"
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

  it("shows the Note's own first line under the row", () => {
    render(
      <ListRow
        label="Coffee"
        note={{ firstLine: "Saturday market, second row.", remainingLines: 3 }}
        moreLinesLabel={(count) => `+${count} lines`}
        labels={labels}
      />,
    )

    expect(screen.getByText("Saturday market, second row.")).toBeInTheDocument()
    expect(screen.getByText("+3 lines")).toBeInTheDocument()
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
