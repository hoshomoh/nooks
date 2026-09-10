/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import "@/test/dom"
import { readyForEnglish } from "@/test/i18n"
import { AddRow, type AddRowSubmission } from "./add-row"

beforeAll(readyForEnglish)

/** field is the add row's text input, whose placeholder clears once a chip appears. */
function field() {
  return screen.getByRole("textbox", { name: "Add an item" })
}

/** typeInto puts text in the add row's field. */
async function typeInto(text: string) {
  await userEvent.type(field(), text)
}

describe("the add row", () => {
  it("files an item with nothing but a name", async () => {
    const onAdd = vi.fn()
    render(<AddRow placeholder="Add an item" onAdd={onAdd} />)

    await typeInto("Milk{Enter}")

    expect(onAdd).toHaveBeenCalledWith<[AddRowSubmission]>({
      label: "Milk",
      quantity: "",
      dueOn: "",
    })
  })

  it("lifts a quantity out of the sentence once a space follows it", async () => {
    const onAdd = vi.fn()
    render(<AddRow placeholder="Add an item" onAdd={onAdd} />)

    await typeInto("Tomatoes 1kg ")

    expect(screen.getByText("1 kg")).toBeInTheDocument()
    expect(screen.getByText("Tomatoes")).toBeInTheDocument()

    await typeInto("{Enter}")
    expect(onAdd).toHaveBeenCalledWith(expect.objectContaining({ label: "Tomatoes", quantity: "1 kg" }))
  })

  it("returns a chip to the text on backspace", async () => {
    render(<AddRow placeholder="Add an item" onAdd={vi.fn()} />)

    await typeInto("Tomatoes 1kg ")
    expect(screen.getByText("1 kg")).toBeInTheDocument()

    await typeInto("{Backspace}")
    expect(screen.queryByText("1 kg")).not.toBeInTheDocument()
    expect(field()).toHaveValue("Tomatoes 1kg")
  })

  it("keeps the date control in the row whether or not a date was typed", () => {
    render(<AddRow placeholder="Add an item" onAdd={vi.fn()} />)
    expect(screen.getByText("Add a date")).toBeInTheDocument()
  })

  it("shows the view's date until the Member says otherwise", () => {
    render(<AddRow placeholder="Add an item" defaultDue="2026-08-25" onAdd={vi.fn()} />)
    expect(screen.queryByText("Add a date")).not.toBeInTheDocument()
  })

  // The browser's own date picker looks different in every browser, and this control
  // is on every row a Member adds.
  it("opens the design system's calendar rather than the browser's", async () => {
    const { container } = render(<AddRow placeholder="Add an item" onAdd={vi.fn()} />)

    expect(container.querySelector("input[type=date]")).toBeNull()

    await userEvent.click(screen.getByRole("button", { name: "Due date" }))

    expect(await screen.findByRole("grid")).toBeInTheDocument()
  })

  it("files the day picked in the calendar", async () => {
    const onAdd = vi.fn()
    render(<AddRow placeholder="Add an item" onAdd={onAdd} />)

    await typeInto("Milk")
    await userEvent.click(screen.getByRole("button", { name: "Due date" }))
    const days = await screen.findAllByRole("gridcell")
    const pickable = days.find((day) => day.querySelector("button:not([disabled])"))
    await userEvent.click(pickable?.querySelector("button") as HTMLElement)
    await typeInto("{Enter}")

    expect(onAdd).toHaveBeenCalledWith(
      expect.objectContaining({ label: "Milk", dueOn: expect.stringMatching(/^\d{4}-\d{2}-\d{2}$/) }),
    )
  })

  it("clears the row after filing, so the next item can be typed", async () => {
    render(<AddRow placeholder="Add an item" onAdd={vi.fn()} />)

    await typeInto("Milk{Enter}")

    expect(field()).toHaveValue("")
  })

  it("files nothing when there is no name", async () => {
    const onAdd = vi.fn()
    render(<AddRow placeholder="Add an item" onAdd={onAdd} />)

    await typeInto("{Enter}")

    expect(onAdd).not.toHaveBeenCalled()
  })
})
