import { QueryClient } from "@tanstack/react-query"
import { describe, expect, it } from "vitest"

import {
  CACHE_KEY,
  CACHE_MAX_AGE_MS,
  forgetCache,
  persistCache,
  restoreCache,
  type CacheStoreDeps,
} from "./cache-store"
import type { DebounceTimers } from "./debounce"

const AT_NOON = new Date("2026-03-14T12:00:00Z").getTime()

/** A fake of everything the store touches, with a clock and timers a test can drive. */
function fakeDeps(seed?: string) {
  const held = new Map<string, string>()
  if (seed !== undefined) {
    held.set(CACHE_KEY, seed)
  }
  let now = AT_NOON
  const due: (() => void)[] = []

  const timers: DebounceTimers = {
    set: (run) => due.push(run),
    clear: () => {},
  }

  const deps: CacheStoreDeps = {
    storage: {
      getItem: (key) => held.get(key) ?? null,
      setItem: (key, value) => void held.set(key, value),
      removeItem: (key) => void held.delete(key),
    },
    now: () => now,
    timers,
  }

  return {
    deps,
    held,
    setNow: (at: number) => void (now = at),
    settle: () => {
      // The debounce hands its work to `set`; running it is the timer firing.
      for (const run of due.splice(0)) {
        run()
      }
    },
  }
}

/** withGroceries is a client that has already been shown something. */
function withGroceries(): QueryClient {
  const queryClient = new QueryClient()
  queryClient.setQueryData(["lists"], [{ uid: "list_groceries", name: "Groceries" }])
  return queryClient
}

describe("keeping what the browser knows", () => {
  it("writes the cache once the changes settle", () => {
    const { deps, held, settle } = fakeDeps()
    const queryClient = withGroceries()

    persistCache(queryClient, deps)
    queryClient.setQueryData(["lists"], [{ uid: "list_weekend", name: "Weekend" }])
    settle()

    expect(held.get(CACHE_KEY)).toContain("Weekend")
  })

  it("reads it back into a fresh client", () => {
    const { deps, held, settle } = fakeDeps()
    persistCache(withGroceries(), deps)
    settle()

    const next = new QueryClient()
    restoreCache(next, { ...deps, storage: deps.storage })

    expect(next.getQueryData(["lists"])).toEqual([{ uid: "list_groceries", name: "Groceries" }])
    expect(held.has(CACHE_KEY)).toBe(true)
  })

  it("throws away a cache old enough to mislead", () => {
    const { deps, held, settle, setNow } = fakeDeps()
    persistCache(withGroceries(), deps)
    settle()

    // A shopping list from last week, shown as though it were current, is worse than
    // a spinner.
    setNow(AT_NOON + CACHE_MAX_AGE_MS + 1)
    const next = new QueryClient()
    restoreCache(next, deps)

    expect(next.getQueryData(["lists"])).toBeUndefined()
    expect(held.has(CACHE_KEY)).toBe(false)
  })

  it("stops writing once it is told to", () => {
    const { deps, held, settle } = fakeDeps()
    const queryClient = withGroceries()

    const stop = persistCache(queryClient, deps)
    stop()
    queryClient.setQueryData(["lists"], [{ uid: "list_weekend", name: "Weekend" }])
    settle()

    expect(held.has(CACHE_KEY)).toBe(false)
  })

  it("survives a storage that refuses to answer", () => {
    const deps: CacheStoreDeps = {
      storage: {
        getItem: () => {
          throw new Error("private window")
        },
        setItem: () => {
          throw new Error("quota")
        },
        removeItem: () => {},
      },
      now: () => AT_NOON,
    }

    // Reading offline stops working. Nothing else does, so nothing may throw.
    expect(() => restoreCache(new QueryClient(), deps)).not.toThrow()
    expect(() => forgetCache(deps)).not.toThrow()
  })

  it("ignores a stored cache that is not one", () => {
    const { deps } = fakeDeps("not json at all")
    const next = new QueryClient()

    expect(() => restoreCache(next, deps)).not.toThrow()
    expect(next.getQueryData(["lists"])).toBeUndefined()
  })

  it("forgets on request, so a new session does not read the last one's lists", () => {
    const { deps, held, settle } = fakeDeps()
    persistCache(withGroceries(), deps)
    settle()

    forgetCache(deps)

    expect(held.has(CACHE_KEY)).toBe(false)
  })
})
