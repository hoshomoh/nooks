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
  // A token that names no List reaches none, so there is nothing to cut yet.
  it("will not cut a token that names nothing", async () => {
    show()

    await userEvent.type(screen.getByRole("textbox", { name: "Name" }), "Kitchen tablet")
    expect(screen.getByRole("button", { name: "Add" })).toBeDisabled()

    await userEvent.click(screen.getByRole("checkbox", { name: "Groceries" }))
    expect(screen.getByRole("button", { name: "Add" })).toBeEnabled()
  })

  it("reports the name, the lists and the permission that were picked", async () => {
    const onAdd = vi.fn()
    show(onAdd)

    await userEvent.type(screen.getByRole("textbox", { name: "Name" }), "Kitchen tablet")
    await userEvent.click(screen.getByRole("checkbox", { name: "Groceries" }))
    await userEvent.click(screen.getByRole("radio", { name: "Read" }))
    await userEvent.click(screen.getByRole("button", { name: "Add" }))

    expect(onAdd).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "Kitchen tablet",
        permission: Permission.READ,
        listUids: ["list_groceries"],
      }),
    )
  })

  // A token that does not expire is a deliberate answer, not the absence of one.
  it("gives a token an expiry unless one is turned off", async () => {
    const onAdd = vi.fn()
    show(onAdd)

    await userEvent.type(screen.getByRole("textbox", { name: "Name" }), "Kitchen tablet")
    await userEvent.click(screen.getByRole("checkbox", { name: "Groceries" }))
    await userEvent.click(screen.getByRole("button", { name: "Add" }))

    expect(onAdd.mock.calls[0][0].expiresAt).not.toBe("")
  })

  it("takes the expiry off when asked", async () => {
    const onAdd = vi.fn()
    show(onAdd)

    await userEvent.type(screen.getByRole("textbox", { name: "Name" }), "Kitchen tablet")
    await userEvent.click(screen.getByRole("checkbox", { name: "Groceries" }))
    await userEvent.click(screen.getByRole("radio", { name: "Never" }))
    await userEvent.click(screen.getByRole("button", { name: "Add" }))

    expect(onAdd.mock.calls[0][0].expiresAt).toBe("")
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
