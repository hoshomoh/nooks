import { QueryClient } from "@tanstack/react-query"
import { describe, expect, it, vi } from "vitest"

import { registerItemChanges } from "./item-changes"
import { ADD_MUTATION, TICK_MUTATION } from "./queued-changes"

/**
 * The surroundings the query layer hands every callback.
 *
 * Built here rather than taken from a running mutation: these tests call the registered
 * callbacks directly, which is the point — what is being checked is what the client
 * will do, not whether TanStack calls it.
 */
function surroundings(queryClient: QueryClient) {
  return { client: queryClient, meta: undefined }
}

/** registered is what the client will run for a change with that key. */
function registered(queryClient: QueryClient, key: readonly string[]) {
  const defaults = queryClient.getMutationDefaults([...key])
  if (!defaults) {
    throw new Error(`nothing registered for ${key.join("/")}`)
  }
  return defaults
}

/** A client that has already been shown a List with one unticked Item on it. */
function withMilk(): QueryClient {
  const queryClient = new QueryClient()
  registerItemChanges(queryClient)
  queryClient.setQueryData(["list", "list_groceries"], {
    items: [{ uid: "item_milk", label: "Milk", done: false }],
  })
  return queryClient
}

/** itemsOf reads the cached List back. */
function itemsOf(queryClient: QueryClient) {
  return (queryClient.getQueryData(["list", "list_groceries"]) as { items: { done: boolean }[] })
    .items
}

describe("what the client will run for a queued change", () => {
  it("knows how to tick and how to add", () => {
    // This is what lets a change restored from a previous visit be sent. Without it the
    // client has a record of the intention and no way to act on it.
    const queryClient = new QueryClient()
    registerItemChanges(queryClient)

    expect(registered(queryClient, TICK_MUTATION).mutationFn).toBeTypeOf("function")
    expect(registered(queryClient, ADD_MUTATION).mutationFn).toBeTypeOf("function")
  })

  it("shows the tick before it has been sent", async () => {
    const queryClient = withMilk()
    const { onMutate } = registered(queryClient, TICK_MUTATION)

    await onMutate?.({ itemUid: "item_milk", done: true }, surroundings(queryClient))

    expect(itemsOf(queryClient)[0]?.done).toBe(true)
  })

  it("shows an added Item before it has been sent", async () => {
    const queryClient = withMilk()
    const { onMutate } = registered(queryClient, ADD_MUTATION)

    await onMutate?.(
      { listUid: "list_groceries", label: "Rye flour", quantity: "", dueOn: "" },
      surroundings(queryClient),
    )

    expect(itemsOf(queryClient)).toHaveLength(2)
  })

  it("goes back to the Instance when a change is refused", async () => {
    // A tick queued offline can come back to an Item somebody deleted. Leaving the
    // optimistic row where it is would have the app claiming something it never did,
    // and putting back what was on screen would be an older wrong answer.
    const queryClient = withMilk()
    const asked = vi.spyOn(queryClient, "invalidateQueries")
    const { onError } = registered(queryClient, TICK_MUTATION)

    await onError?.(
      new Error("no such item"),
      { itemUid: "item_milk", done: true },
      undefined,
      surroundings(queryClient),
    )

    expect(asked).toHaveBeenCalled()
  })
})
