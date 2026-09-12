import { describe, expect, it } from "vitest"

import {
  createConnectionStore,
  type ConnectionStore,
  type ConnectionStoreDeps,
} from "./connection-store"

const AT_NINE = new Date("2026-03-14T09:00:00Z")
const AT_TEN = new Date("2026-03-14T10:00:00Z")

/** A fake of everything the store touches, so no browser is involved. */
function fakeDeps(startsOnline = true) {
  let online = startsOnline
  let now = AT_NINE
  const listeners = new Map<string, Set<() => void>>()

  const deps: ConnectionStoreDeps = {
    isOnline: () => online,
    events: {
      addEventListener: (type, listener) => {
        const set = listeners.get(type) ?? new Set()
        set.add(listener)
        listeners.set(type, set)
      },
      removeEventListener: (type, listener) => void listeners.get(type)?.delete(listener),
    },
    now: () => now,
  }

  return {
    deps,
    listenerCount: () => [...listeners.values()].reduce((total, set) => total + set.size, 0),
    fire: (type: "online" | "offline") => {
      online = type === "online"
      for (const listener of listeners.get(type) ?? []) {
        listener()
      }
    },
    setNow: (at: Date) => void (now = at),
  }
}

/** watching subscribes the way React does, so the window listeners are attached. */
function watching(store: ConnectionStore): () => void {
  return store.subscribe(() => {})
}

describe("the connection store", () => {
  it("starts from what the machine says, and counts that as contact", () => {
    const { deps } = fakeDeps()
    const store = createConnectionStore(deps)

    expect(store.getState()).toEqual({ online: true, lastSeenAt: AT_NINE.toISOString() })
  })

  it("has never seen the Instance when the machine starts offline", () => {
    const { deps } = fakeDeps(false)
    const store = createConnectionStore(deps)

    expect(store.getState()).toEqual({ online: false, lastSeenAt: null })
  })

  it("keeps the last good moment after the network drops", () => {
    const { deps, fire } = fakeDeps()
    const store = createConnectionStore(deps)
    watching(store)

    fire("offline")

    // The moment is what the banner has to say about how stale the screen might be, so
    // going offline must not clear it.
    expect(store.getState()).toEqual({ online: false, lastSeenAt: AT_NINE.toISOString() })
  })

  it("moves the moment on when contact comes back", () => {
    const { deps, fire, setNow } = fakeDeps()
    const store = createConnectionStore(deps)
    watching(store)

    fire("offline")
    setNow(AT_TEN)
    fire("online")

    expect(store.getState()).toEqual({ online: true, lastSeenAt: AT_TEN.toISOString() })
  })

  it("goes offline on a request that never arrived, without waiting for the machine", () => {
    const { deps } = fakeDeps()
    const store = createConnectionStore(deps)

    store.markUnreachable()

    expect(store.getState().online).toBe(false)
  })

  it("tells a listener once per change", () => {
    const { deps, fire } = fakeDeps()
    const store = createConnectionStore(deps)
    let told = 0
    store.subscribe(() => void (told += 1))

    fire("offline")
    fire("offline")

    expect(told).toBe(1)
  })

  it("hands back the same snapshot while nothing changes", () => {
    const { deps } = fakeDeps()
    const store = createConnectionStore(deps)

    // useSyncExternalStore compares snapshots by identity: a fresh object per read
    // would re-render for ever.
    expect(store.getState()).toBe(store.getState())
  })

  it("listens to the window only while somebody is watching", () => {
    const { deps, listenerCount } = fakeDeps()
    const store = createConnectionStore(deps)

    const stop = watching(store)
    expect(listenerCount()).toBe(2)

    stop()
    expect(listenerCount()).toBe(0)
  })
})
