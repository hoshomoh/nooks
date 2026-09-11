/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { Sharing, type Item, type List } from "@nooks/api"

import "@/test/dom"
import { readyForEnglish } from "@/test/i18n"
import { ListMenu, type ListMenuActions } from "./list-menu"

beforeAll(readyForEnglish)

/** groceries is a List, owned or not. */
function groceries(isOwner: boolean): List {
  return {
    uid: "list_groceries",
    name: "Groceries",
    sharing: Sharing.PRIVATE,
    canEdit: true,
    isOwner,
    isPinned: false,
    openCount: 2,
  } as List
}

/** openMenu renders the menu and opens it. */
async function openMenu(isOwner = true, items: Item[] = []) {
  const actions: ListMenuActions = {
    onRename: vi.fn(),
    onPin: vi.fn(),
    onDuplicate: vi.fn(),
    onDelete: vi.fn(),
  }
  render(<ListMenu list={groceries(isOwner)} items={items} actions={actions} />)
  await userEvent.click(screen.getByRole("button", { name: "More" }))
  return actions
}

describe("the list menu", () => {
  it("holds everything that is not worth a permanent control", async () => {
    await openMenu()

    for (const label of ["Rename", "Pin to sidebar", "Duplicate", "Print", "Export as plain text"]) {
      expect(await screen.findByRole("menuitem", { name: new RegExp(label) })).toBeInTheDocument()
    }
  })

  // A control that can never be used is noise, not information.
  it("leaves out what somebody who does not own it cannot do", async () => {
    await openMenu(false)

    expect(await screen.findByRole("menuitem", { name: /Duplicate/ })).toBeInTheDocument()
    expect(screen.queryByRole("menuitem", { name: /Rename/ })).not.toBeInTheDocument()
    expect(screen.queryByRole("menuitem", { name: /Delete list/ })).not.toBeInTheDocument()
  })

  it("pins a List that is not pinned", async () => {
    const actions = await openMenu()

    await userEvent.click(await screen.findByRole("menuitem", { name: /Pin to sidebar/ }))

    expect(actions.onPin).toHaveBeenCalledWith(true)
  })

  // Deleting a List asks once, and says what will happen rather than "are you sure?".
  it("asks before deleting, and says what goes", async () => {
    const actions = await openMenu()

    await userEvent.click(await screen.findByRole("menuitem", { name: /Delete list/ }))

    expect(await screen.findByText(/Delete “Groceries”\?/)).toBeInTheDocument()
    expect(screen.getByText(/Nobody it was shared with keeps a copy/)).toBeInTheDocument()
    expect(actions.onDelete).not.toHaveBeenCalled()

    await userEvent.click(screen.getByRole("button", { name: "Delete" }))
    expect(actions.onDelete).toHaveBeenCalledOnce()
  })
})
