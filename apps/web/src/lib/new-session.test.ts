import { describe, expect, it } from "vitest"
import { QueryClient } from "@tanstack/react-query"

import { CACHE_KEY } from "./cache-store"
import { startNewSession } from "./new-session"
import { PENDING_TICK_KEY, createPendingTickStore } from "./pending-tick-store"

/** held is a browser's storage, with whatever was left in it. */
function held(initial: Record<string, string>) {
  const store = new Map(Object.entries(initial))
  return {
    getItem: (key: string) => store.get(key) ?? null,
    setItem: (key: string, value: string) => void store.set(key, value),
    removeItem: (key: string) => void store.delete(key),
    has: (key: string) => store.has(key),
  }
}

/**
 * A session change ends what the last person left behind.
 *
 * Signing out on a shared tablet is the moment this matters: everything the browser is
 * holding belonged to whoever is leaving. The cache was already forgotten here. A tick
 * somebody reached for on the public list was not, and it is kept on disk on purpose,
 * because signing in can take days — so it outlived the sign-out and was applied by
 * whoever signed in next, under their name rather than the name of the person who
 * reached for it.
 */
describe("starting a new session", () => {
  it("forgets the answers the last Member saw", () => {
    const storage = held({ [CACHE_KEY]: "{}" })
    startNewSession(new QueryClient(), { storage, now: () => 0 })

    expect(storage.has(CACHE_KEY)).toBe(false)
  })

  it("forgets a tick the last person at this browser reached for", () => {
    const storage = held({})
    const ticks = createPendingTickStore({ storage })
    ticks.remember("item_bread")

    startNewSession(new QueryClient(), { storage, now: () => 0 }, ticks)

    expect(storage.has(PENDING_TICK_KEY)).toBe(false)
  })

  /*
   * Which is why the two callers that mean to apply it take it first.
   *
   * Signing in and joining both call applyPendingTick before starting the session. The
   * other way round, the tick is forgotten before anybody can act on it and the feature
   * quietly does nothing, which is a thing a reordering would cause and nothing else
   * would notice.
   */
  it("is too late to apply a tick after a session has started", () => {
    const storage = held({})
    const ticks = createPendingTickStore({ storage })
    ticks.remember("item_bread")

    startNewSession(new QueryClient(), { storage, now: () => 0 }, ticks)

    expect(ticks.take()).toBe("")
  })
})
