import { describe, expect, it, vi } from "vitest"

import { createLastListStore, LAST_LIST_STORAGE_KEY } from "./last-list-store"

/** memory is a Storage that a test can look inside. */
function memory(initial: Record<string, string> = {}) {
  const held = { ...initial }
  return {
    held,
    getItem: (key: string) => held[key] ?? null,
    setItem: (key: string, value: string) => {
      held[key] = value
    },
  }
}

describe("the List last added to", () => {
  it("is nothing before anything has been added", () => {
    expect(createLastListStore({ storage: memory() }).getUid()).toBe("")
  })

  it("comes back from the last time the Member was here", () => {
    const storage = memory({ [LAST_LIST_STORAGE_KEY]: "list_groceries" })
    expect(createLastListStore({ storage }).getUid()).toBe("list_groceries")
  })

  it("is remembered when something is added to it", () => {
    const storage = memory()
    const store = createLastListStore({ storage })

    store.remember("list_bike")

    expect(store.getUid()).toBe("list_bike")
    expect(storage.held[LAST_LIST_STORAGE_KEY]).toBe("list_bike")
  })

  it("tells whoever is reading it that it changed", () => {
    const store = createLastListStore({ storage: memory() })
    const listener = vi.fn()
    store.subscribe(listener)

    store.remember("list_bike")

    expect(listener).toHaveBeenCalledOnce()
  })

  it("says nothing when the answer has not changed", () => {
    const store = createLastListStore({ storage: memory() })
    store.remember("list_bike")
    const listener = vi.fn()
    store.subscribe(listener)

    store.remember("list_bike")

    expect(listener).not.toHaveBeenCalled()
  })

  // A Member in a private window still gets it for the session they are in.
  it("keeps working when the browser will not store anything", () => {
    const store = createLastListStore({
      storage: {
        getItem: () => {
          throw new Error("blocked")
        },
        setItem: () => {
          throw new Error("blocked")
        },
      },
    })

    store.remember("list_bike")

    expect(store.getUid()).toBe("list_bike")
  })
})
