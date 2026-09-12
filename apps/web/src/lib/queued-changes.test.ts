import { describe, expect, it } from "vitest"

import {
  isProvisional,
  provisionalItem,
  queuedChanges,
  withItem,
  withTick,
  type QueuedMutation,
} from "./queued-changes"

/** paused is a change that has been made and not yet sent. */
function paused(variables: unknown): QueuedMutation {
  return { isPaused: true, variables }
}

/** inFlight is a change that is on its way, which is not the same thing. */
function inFlight(variables: unknown): QueuedMutation {
  return { isPaused: false, variables }
}

describe("what is waiting to be sent", () => {
  it("collects the items whose tick is paused", () => {
    const waiting = queuedChanges([paused({ itemUid: "item_milk", done: true })])

    expect([...waiting.itemUids]).toEqual(["item_milk"])
  })

  it("ignores a change that is on its way", () => {
    // In flight is not waiting: the row should not claim it is unsent while it is
    // being sent.
    const waiting = queuedChanges([inFlight({ itemUid: "item_milk", done: true })])

    expect(waiting.itemUids.size).toBe(0)
  })

  it("collects the lists something was added to", () => {
    const waiting = queuedChanges([paused({ listUid: "list_groceries", label: "Rye flour" })])

    expect([...waiting.listUids]).toEqual(["list_groceries"])
  })

  it("survives a mutation whose variables are not ours", () => {
    expect(() => queuedChanges([paused(undefined), paused("what"), paused(null)])).not.toThrow()
  })
})

describe("showing a tick before it is sent", () => {
  it("flips the item on a list answer", () => {
    const data = { items: [{ uid: "item_milk", done: false }] }

    expect(withTick(data, "item_milk", true)).toEqual({ items: [{ uid: "item_milk", done: true }] })
  })

  it("flips the item inside a dated answer", () => {
    // The same Item, wrapped with the List it came from. Today and Upcoming read this
    // shape, and a tick that only landed on one would come back on the other.
    const data = { items: [{ listName: "Groceries", item: { uid: "item_milk", done: false } }] }

    expect(withTick(data, "item_milk", true)).toEqual({
      items: [{ listName: "Groceries", item: { uid: "item_milk", done: true } }],
    })
  })

  it("leaves everything else alone", () => {
    const data = { items: [{ uid: "item_bread", done: false }] }

    expect(withTick(data, "item_milk", true)).toEqual(data)
  })

  it("answers an unread shape unchanged rather than throwing", () => {
    expect(withTick(undefined, "item_milk", true)).toBeUndefined()
    expect(withTick({ lists: [] }, "item_milk", true)).toEqual({ lists: [] })
  })
})

describe("showing an item before it is sent", () => {
  it("puts it at the end of the list", () => {
    const data = { items: [{ uid: "item_milk" }] }
    const added = provisionalItem({ label: "Rye flour", quantity: "1 kg", dueOn: "" })

    const next = withItem(data, added) as { items: { uid: string }[] }

    expect(next.items.at(-1)?.uid).toBe(added.uid)
  })

  it("marks the item as one the instance has never heard of", () => {
    const added = provisionalItem({ label: "Rye flour", quantity: "", dueOn: "" })

    // The uid is this browser's invention, and is never sent anywhere: the refresh
    // after the change lands replaces the whole Item.
    expect(isProvisional(added.uid)).toBe(true)
    expect(isProvisional("item_milk")).toBe(false)
  })
})

describe("a change that outlived the tab", () => {
  it("is written out with the cache, so it can be sent later", async () => {
    const { QueryClient, dehydrate } = await import("@tanstack/react-query")
    const queryClient = new QueryClient()
    queryClient.setMutationDefaults(["item", "set-done"], { mutationFn: () => Promise.resolve() })

    // Paused rather than failed is the whole point: TanStack writes a paused change
    // out with the cache, and a failed one is gone.
    const cache = queryClient.getMutationCache()
    const mutation = cache.build(queryClient, {
      mutationKey: ["item", "set-done"],
      mutationFn: () => Promise.resolve(),
    })
    mutation.state.isPaused = true
    mutation.state.variables = { itemUid: "item_milk", done: true }

    expect(dehydrate(queryClient).mutations).toHaveLength(1)
  })
})
