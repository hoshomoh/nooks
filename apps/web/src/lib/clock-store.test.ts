import { describe, expect, it } from "vitest"

import { createClockStore, TICK_MS, type ClockStoreDeps } from "./clock-store"

/** FakeClock is a clock a test moves by hand, and a timer it fires by hand. */
type FakeClock = {
  deps: ClockStoreDeps
  advance: (ms: number) => void
  fire: () => void
  running: () => boolean
}

function fakeClock(start: string): FakeClock {
  let now = new Date(start)
  let tick: (() => void) | null = null
  return {
    deps: {
      now: () => now,
      every: (callback) => {
        tick = callback
        return () => {
          tick = null
        }
      },
    },
    advance: (ms) => {
      now = new Date(now.getTime() + ms)
    },
    fire: () => tick?.(),
    running: () => tick !== null,
  }
}

describe("the clock", () => {
  it("gives the same moment to every read between ticks", () => {
    const clock = fakeClock("2026-11-13T23:59:30")
    const store = createClockStore(clock.deps)
    store.subscribe(() => {})

    const first = store.getState()
    clock.advance(10_000)
    expect(store.getState()).toBe(first)
  })

  it("moves on with each tick and tells whoever is reading", () => {
    const clock = fakeClock("2026-11-13T23:59:30")
    const store = createClockStore(clock.deps)
    let told = 0
    store.subscribe(() => told++)

    clock.advance(TICK_MS)
    clock.fire()

    expect(told).toBe(1)
    expect(store.getState().toISOString()).toBe(new Date("2026-11-14T00:00:30").toISOString())
  })

  it("only runs while something reads it", () => {
    const clock = fakeClock("2026-11-13T12:00:00")
    const store = createClockStore(clock.deps)
    expect(clock.running()).toBe(false)

    const stopA = store.subscribe(() => {})
    const stopB = store.subscribe(() => {})
    expect(clock.running()).toBe(true)

    stopA()
    expect(clock.running()).toBe(true)
    stopB()
    expect(clock.running()).toBe(false)
  })

  it("gives the first reader after a quiet spell the time it is", () => {
    const clock = fakeClock("2026-11-13T12:00:00")
    const store = createClockStore(clock.deps)

    clock.advance(8 * 60 * 60 * 1000)

    expect(store.getState().toISOString()).toBe(new Date("2026-11-13T20:00:00").toISOString())
  })

  it("gives the time it is after the clock is set back", () => {
    const clock = fakeClock("2026-11-13T12:00:00")
    const store = createClockStore(clock.deps)

    clock.advance(-2 * 60 * 60 * 1000)

    expect(store.getState().toISOString()).toBe(new Date("2026-11-13T10:00:00").toISOString())
  })
})
