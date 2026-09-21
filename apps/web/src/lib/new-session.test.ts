import { QueryClient } from "@tanstack/react-query"
import { describe, expect, it } from "vitest"

import { CACHE_KEY, type CacheStoreDeps } from "./cache-store"
import { startNewSession } from "./new-session"

/** A storage that remembers, so the test can look at what is left in it. */
function storageHolding(value: string): { deps: CacheStoreDeps; read: () => string | null } {
  const held = new Map<string, string>([[CACHE_KEY, value]])
  return {
    deps: {
      storage: {
        getItem: (key) => held.get(key) ?? null,
        setItem: (key, next) => void held.set(key, next),
        removeItem: (key) => void held.delete(key),
      },
      now: () => 0,
    },
    read: () => held.get(CACHE_KEY) ?? null,
  }
}

describe("starting a new session", () => {
  it("empties the tab and the disk, not one of them", () => {
    const client = new QueryClient()
    client.setQueryData(["current-member"], { name: "Anna", email: "anna@brunnen.lan" })
    const stored = storageHolding('{"at":0,"state":{"queries":[]}}')

    startNewSession(client, stored.deps)

    expect(client.getQueryData(["current-member"])).toBeUndefined()
    expect(stored.read(), "the next person at this browser reads this").toBeNull()
  })
})
