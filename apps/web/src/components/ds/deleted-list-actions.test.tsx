import { afterEach, beforeAll, describe, expect, it, vi } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { Sharing, type List, type RestoreListResponse } from "@nooks/api"

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

    expect(await screen.findAllByRole("menuitem")).toHaveLength(1)
  })
})
