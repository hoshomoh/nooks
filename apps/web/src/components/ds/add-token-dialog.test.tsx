/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { Permission, Sharing, type List } from "@nooks/api"

import "@/test/dom"
import { readyForEnglish } from "@/test/i18n"
import { AddTokenDialog, type NewToken } from "./add-token-dialog"

beforeAll(readyForEnglish)

/** list builds one of the Member's Lists. */
function list(uid: string, name: string): List {
  return {
    uid,
    name,
    sharing: Sharing.PRIVATE,
    canEdit: true,
    isOwner: true,
    isPinned: false,
    openCount: 0,
  } as List
}

const LISTS = [list("list_groceries", "Groceries"), list("list_bike", "Bike")]

/** show opens the dialog and answers with what it reports. */
function show(onAdd: (token: NewToken) => void = vi.fn()) {
  render(
    <AddTokenDialog
      open
      onOpenChange={vi.fn()}
      lists={LISTS}
      onAdd={onAdd}
      onSecretRead={vi.fn()}
    />,
  )
}

describe("cutting an access token", () => {
  // The common case: a client somebody builds wants everything, including the Lists
  // they make next week.
  it("reaches every list by default, without making anybody tick them all", async () => {
    const onAdd = vi.fn()
    show(onAdd)

    await userEvent.type(screen.getByRole("textbox", { name: "Name" }), "My client")
    await userEvent.click(screen.getByRole("button", { name: "Add" }))

    expect(onAdd).toHaveBeenCalledWith(
      expect.objectContaining({ allLists: true, listUids: [] }),
    )
  })

  // A token that names no List reaches none, so there is nothing to cut yet.
  it("will not cut a token that names nothing once lists are being picked", async () => {
    show()

    await userEvent.type(screen.getByRole("textbox", { name: "Name" }), "Kitchen tablet")
    await userEvent.click(screen.getByRole("radio", { name: /Only the lists I pick/ }))
    expect(screen.getByRole("button", { name: "Add" })).toBeDisabled()

    await userEvent.click(screen.getByRole("checkbox", { name: "Groceries" }))
    expect(screen.getByRole("button", { name: "Add" })).toBeEnabled()
  })

  it("reports the name, the lists and the permission that were picked", async () => {
    const onAdd = vi.fn()
    show(onAdd)

    await userEvent.type(screen.getByRole("textbox", { name: "Name" }), "Kitchen tablet")
    await userEvent.click(screen.getByRole("radio", { name: /Only the lists I pick/ }))
    await userEvent.click(screen.getByRole("checkbox", { name: "Groceries" }))
    await userEvent.click(screen.getByRole("radio", { name: /Read only/ }))
    await userEvent.click(screen.getByRole("button", { name: "Add" }))

    expect(onAdd).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "Kitchen tablet",
        permission: Permission.READ,
        listUids: ["list_groceries"],
        allLists: false,
      }),
    )
  })

  // A control that explains only the answer already chosen asks a Member to pick first
  // and understand afterwards.
  it("says what each permission means before one is chosen", () => {
    show()

    expect(screen.getByText(/cannot tick anything off/)).toBeInTheDocument()
    expect(screen.getByText(/plus ticking off, adding/)).toBeInTheDocument()
  })

  // A token that does not expire is a deliberate answer, not the absence of one.
  it("gives a token an expiry unless one is turned off", async () => {
    const onAdd = vi.fn()
    show(onAdd)

    await userEvent.type(screen.getByRole("textbox", { name: "Name" }), "Kitchen tablet")
    await userEvent.click(screen.getByRole("button", { name: "Add" }))

    expect(onAdd.mock.calls[0][0].expiresAt).not.toBe("")
  })

  it("takes the expiry off when asked", async () => {
    const onAdd = vi.fn()
    show(onAdd)

    await userEvent.type(screen.getByRole("textbox", { name: "Name" }), "Kitchen tablet")
    await userEvent.click(screen.getByRole("radio", { name: "Never" }))
    await userEvent.click(screen.getByRole("button", { name: "Add" }))

    expect(onAdd.mock.calls[0][0].expiresAt).toBe("")
  })

  // The quick answers cover most of it; a day of their own covers the rest.
  it("waits for a day when one is being picked", async () => {
    show()

    await userEvent.type(screen.getByRole("textbox", { name: "Name" }), "Kitchen tablet")
    await userEvent.click(screen.getByRole("radio", { name: "Pick a day" }))

    expect(screen.getByRole("button", { name: "Add" })).toBeDisabled()
  })

  // The one moment the secret exists outside the caller's hands.
  it("shows the secret instead of the form once there is one", () => {
    render(
      <AddTokenDialog
        open
        onOpenChange={vi.fn()}
        lists={LISTS}
        onAdd={vi.fn()}
        secret={{ tokenName: "Kitchen tablet", secret: "cedar-lantern-pebble" }}
        onSecretRead={vi.fn()}
      />,
    )

    expect(screen.getByText("cedar-lantern-pebble")).toBeInTheDocument()
    expect(screen.queryByRole("textbox", { name: "Name" })).not.toBeInTheDocument()
  })
})
