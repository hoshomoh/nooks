import { describe, expect, it } from "vitest"

import { createConnectionStore, type ConnectionStoreDeps } from "./connection-store"
import { wireOnline, type OnlineManagerLike } from "./online-wiring"

/** A connection store with no browser behind it, and a handle to drive it. */
function connection(startsOnline = true) {
  const deps: ConnectionStoreDeps = {
    isOnline: () => startsOnline,
    events: { addEventListener: () => {}, removeEventListener: () => {} },
    now: () => new Date("2026-03-14T09:00:00Z"),
  }
  return createConnectionStore(deps)
}

/** A stand-in for TanStack's online manager that records what it was told. */
function manager() {
  const told: boolean[] = []
  let setOnline: ((online: boolean) => void) | null = null

  const fake: OnlineManagerLike = {
    setOnline: (online) => void told.push(online),
    setEventListener: (setup) => {
      setOnline = (online) => told.push(online)
      setup(setOnline)
    },
  }
  return { fake, told, wired: () => setOnline !== null }
}

describe("what the query layer is told about the connection", () => {
  it("is told the current state before anything changes", () => {
    const { fake, told } = manager()

    wireOnline(connection(false), fake)

    // A mutation can be made before any event fires, and an unset manager assumes
    // online — which would fail the change instead of pausing it.
    expect(told[0]).toBe(false)
  })

  it("subscribes for what happens next", () => {
    const { fake, wired } = manager()

    wireOnline(connection(), fake)

    expect(wired()).toBe(true)
  })

  it("passes on a request that never arrived", () => {
    const { fake, told } = manager()
    const store = connection()
    wireOnline(store, fake)
    const before = told.length

    store.markUnreachable()

    expect(told.slice(before)).toContain(false)
  })

  it("passes on contact coming back", () => {
    const { fake, told } = manager()
    const store = connection()
    wireOnline(store, fake)
    store.markUnreachable()
    const before = told.length

    store.markSeen()

    expect(told.slice(before)).toContain(true)
  })
})
