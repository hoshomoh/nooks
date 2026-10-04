import { describe, expect, it } from "vitest"
import { QueryClient, QueryObserver } from "@tanstack/react-query"

import { afterDeletingList } from "./after-deleting-list"

/**
 * Deleting the List on screen leaves it without asking for it again.
 *
 * It used to re-read everything a List change touches before leaving, and the List on
 * screen was one of those things. The Instance answered not found, the read was retried,
 * and the page stayed on a deleted List for about ten seconds.
 */
describe("after deleting a List", () => {
  it("never asks for the deleted List again", async () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

    let asked = 0
    const onScreen = new QueryObserver(queryClient, {
      queryKey: ["list", "list_gone"],
      queryFn: async () => {
        asked++
        return { name: "Groceries" }
      },
    })
    const unmount = onScreen.subscribe(() => {})
    await expect.poll(() => asked).toBe(1)

    await afterDeletingList(queryClient, "list_gone", async () => unmount())

    expect(asked).toBe(1)
    expect(queryClient.getQueryData(["list", "list_gone"])).toBeUndefined()
  })

  it("leaves every other List where it was", async () => {
    const queryClient = new QueryClient()
    queryClient.setQueryData(["list", "list_kept"], { name: "Hardware" })

    await afterDeletingList(queryClient, "list_gone", async () => {})

    expect(queryClient.getQueryData(["list", "list_kept"])).toEqual({ name: "Hardware" })
  })
})
