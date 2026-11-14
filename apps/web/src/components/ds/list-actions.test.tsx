import { afterEach, beforeAll, describe, expect, it, vi } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { RouterProvider, createRootRoute, createRouter } from "@tanstack/react-router"
import { Sharing, type GetListResponse, type Item, type List } from "@nooks/api"

import { listClient } from "@/lib/api"
import { readyForEnglish } from "@/test/i18n"
import { ListActions } from "./list-actions"

beforeAll(readyForEnglish)
afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

/** BlobURL is the URL jsdom leaves out: a file handed to the browser to save. */
class BlobURL extends URL {
  static createObjectURL = () => "blob:groceries"
  static revokeObjectURL = () => {}
}

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

/**
 * openMenu renders the actions inside a router, because opening a List is one of them
 * and a menu that cannot navigate is not the thing being tested.
 */
async function openMenu(isOwner = true, withControls = false) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const router = createRouter({
    routeTree: createRootRoute({
      component: () => (
        <ListActions
          list={groceries(isOwner)}
          instanceName="Brunnen Street"
          withControls={withControls}
        />
      ),
    }),
  })

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )

  await userEvent.click(await screen.findByRole("button", { name: "More" }))
}

describe("what can be done to a List", () => {
  it("offers everything that happens to a List, from one menu", async () => {
    await openMenu()

    for (const label of ["Open list", "Rename", "Pin", "Share…", "Print", "Duplicate", "Export", "Delete list"]) {
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

  // Every list menu offers the same things, so one away from the List reads its Items.
  it("exports from wherever the menu is", async () => {
    const milk = { uid: "item_milk", label: "Milk", done: false } as Item
    const read = vi
      .spyOn(listClient, "getList")
      .mockResolvedValue({ list: groceries(true), items: [milk] } as GetListResponse)
    vi.stubGlobal("URL", BlobURL)
    const saved: string[] = []
    vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(function (this: HTMLAnchorElement) {
      saved.push(this.download)
    })
    await openMenu(true)

    await userEvent.click(await screen.findByRole("menuitem", { name: /Export/ }))

    expect(read).toHaveBeenCalledWith({ listUid: "list_groceries" })
    await waitFor(() => expect(saved).toEqual(["groceries.md"]))
  })

  // The same action twice in one place reads as two different actions.
  it("drops what the chrome bar already shows as a control", async () => {
    await openMenu(true, true)

    expect(screen.getByRole("button", { name: "Share" })).toBeInTheDocument()
    expect(screen.queryByRole("menuitem", { name: /Share…/ })).not.toBeInTheDocument()
  })

  it("asks before deleting, and says what goes", async () => {
    await openMenu()

    await userEvent.click(await screen.findByRole("menuitem", { name: /Delete list/ }))

    expect(await screen.findByText(/Delete “Groceries”\?/)).toBeInTheDocument()
    expect(screen.getByText(/Nobody it was shared with keeps a copy/)).toBeInTheDocument()
  })
})
