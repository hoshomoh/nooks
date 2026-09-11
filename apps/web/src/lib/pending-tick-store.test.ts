import { describe, expect, it } from "vitest"

import { createPendingTickStore } from "./pending-tick-store"

/** memory is a Storage a test can look inside. */
function memory(initial: Record<string, string> = {}) {
  const held: Record<string, string> = { ...initial }
  return {
    held,
    getItem: (key: string) => held[key] ?? null,
    setItem: (key: string, value: string) => {
      held[key] = value
    },
    removeItem: (key: string) => {
      delete held[key]
    },
  }
}

describe("the tick a Visitor reached for", () => {
  it("is nothing when they reached for nothing", () => {
    expect(createPendingTickStore({ storage: memory() }).take()).toBe("")
  })

  it("survives until they have an account to use it with", () => {
    const storage = memory()
    createPendingTickStore({ storage }).remember("item_milk")

    // A different store, as a later visit would be.
    expect(createPendingTickStore({ storage }).take()).toBe("item_milk")
  })

  // Worth applying once. A tick that failed because they cannot reach that List should
  // not be tried again on every sign-in they ever make.
  it("is forgotten once it has been taken", () => {
    const store = createPendingTickStore({ storage: memory() })
    store.remember("item_milk")

    expect(store.take()).toBe("item_milk")
    expect(store.take()).toBe("")
  })

  it("keeps only the last thing they reached for", () => {
    const store = createPendingTickStore({ storage: memory() })
    store.remember("item_milk")
    store.remember("item_oats")

    expect(store.take()).toBe("item_oats")
  })

  // A Visitor in a private window signs in and ticks it themselves.
  it("does not fail when the browser will not store anything", () => {
    const store = createPendingTickStore({
      storage: {
        getItem: () => {
          throw new Error("blocked")
        },
        setItem: () => {
          throw new Error("blocked")
        },
        removeItem: () => {
          throw new Error("blocked")
        },
      },
    })

    store.remember("item_milk")
    expect(store.take()).toBe("")
  })
})
