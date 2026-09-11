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
  due: "Add a date",
}

describe("the list row", () => {
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
})
