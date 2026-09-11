/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import "@/test/dom"
import { readyForEnglish } from "@/test/i18n"
import { ItemMenu, type ItemMenuActions } from "./item-menu"

beforeAll(readyForEnglish)

/** openMenu renders the menu for an Item and opens it. */
async function openMenu(quantity = "", dueOn = "") {
  const actions: ItemMenuActions = {
    onView: vi.fn(),
    onSetDate: vi.fn(),
    onSetQuantity: vi.fn(),
    onDuplicate: vi.fn(),
    onDelete: vi.fn(),
  }
  render(<ItemMenu quantity={quantity} dueOn={dueOn} actions={actions} />)
  await userEvent.click(screen.getByRole("button", { name: "More" }))
  return actions
}

describe("an Item's menu", () => {
  // Clicking the label renames it, so the sheet needs a way in that is not a hunt for
  // dead space on the row.
  it("opens the Item, rather than making a Member find where to click", async () => {
    const actions = await openMenu()

    await userEvent.click(await screen.findByRole("button", { name: /Open/ }))

    expect(actions.onView).toHaveBeenCalledOnce()
  })

  // The entry names the field, so the entry is where the field should be.
  it("sets a quantity inside the menu", async () => {
    const actions = await openMenu("")

    await userEvent.click(await screen.findByRole("button", { name: /Set a quantity/ }))
    await userEvent.type(screen.getByRole("textbox", { name: "Quantity" }), "1 kg")
    await userEvent.click(screen.getByRole("button", { name: "Save" }))

    expect(actions.onSetQuantity).toHaveBeenCalledWith("1 kg")
  })

  it("opens on the quantity the Item already has", async () => {
    await openMenu("2")

    await userEvent.click(await screen.findByRole("button", { name: /Set a quantity/ }))

    expect(screen.getByRole("textbox", { name: "Quantity" })).toHaveValue("2")
  })

  it("sets a date inside the menu", async () => {
    await openMenu()

    await userEvent.click(await screen.findByRole("button", { name: /Set a date/ }))

    expect(await screen.findByRole("grid")).toBeInTheDocument()
  })

  it("offers to take a date off only when there is one", async () => {
    await openMenu("", "2026-08-30")

    await userEvent.click(await screen.findByRole("button", { name: /Set a date/ }))

    expect(screen.getByRole("button", { name: "No date" })).toBeInTheDocument()
  })

  it("comes back from a field to the actions", async () => {
    await openMenu()

    await userEvent.click(await screen.findByRole("button", { name: /Set a quantity/ }))
    await userEvent.click(screen.getByRole("button", { name: "Back" }))

    expect(screen.getByRole("button", { name: /Duplicate/ })).toBeInTheDocument()
  })
})
