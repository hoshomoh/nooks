import { afterEach, beforeAll, describe, expect, it, vi } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { Sharing, type List, type PurgeListResponse, type RestoreListResponse } from "@nooks/api"

import { listClient } from "@/lib/api"
import { readyForEnglish } from "@/test/i18n"
import { DeletedListActions } from "./deleted-list-actions"

beforeAll(readyForEnglish)
afterEach(() => vi.restoreAllMocks())

const deleted = {
  uid: "list_party",
  name: "Party, 30th",
  sharing: Sharing.PRIVATE,
  isOwner: true,
  openCount: 3,
  doneCount: 11,
} as List

async function openMenu() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={queryClient}>
      <DeletedListActions list={deleted} />
    </QueryClientProvider>,
  )
  await userEvent.click(screen.getByRole("button", { name: "More" }))
}

describe("a deleted List's menu", () => {
  it("brings the List back", async () => {
    const restore = vi
      .spyOn(listClient, "restoreList")
      .mockResolvedValue({} as RestoreListResponse)
    await openMenu()

    await userEvent.click(await screen.findByRole("menuitem", { name: "Restore" }))

    expect(restore).toHaveBeenCalledWith({ listUid: "list_party" })
  })

  // A deleted List cannot be opened, renamed or shared, so nothing offers to.
  it("offers nothing a deleted List cannot do", async () => {
    await openMenu()

    const offered = await screen.findAllByRole("menuitem")
    expect(offered.map((entry) => entry.textContent)).toEqual(["Restore", "Delete forever⌫"])
  })

  // The one thing that loses something for certain says so, and asks first.
  it("asks before deleting it forever", async () => {
    const purge = vi.spyOn(listClient, "purgeList").mockResolvedValue({} as PurgeListResponse)
    await openMenu()

    await userEvent.click(await screen.findByRole("menuitem", { name: /Delete forever/ }))
    expect(purge).not.toHaveBeenCalled()
    expect(await screen.findByText("Delete “Party, 30th” forever?")).toBeInTheDocument()
    expect(screen.getByText(/its 14 items are removed for everyone/)).toBeInTheDocument()

    await userEvent.click(screen.getByRole("button", { name: "Delete forever" }))
    expect(purge).toHaveBeenCalledWith({ listUid: "list_party" })
  })

  it("keeps it when the question is turned down", async () => {
    const purge = vi.spyOn(listClient, "purgeList").mockResolvedValue({} as PurgeListResponse)
    await openMenu()

    await userEvent.click(await screen.findByRole("menuitem", { name: /Delete forever/ }))
    await userEvent.click(await screen.findByRole("button", { name: "Cancel" }))

    expect(purge).not.toHaveBeenCalled()
  })
})
