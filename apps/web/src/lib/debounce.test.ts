import { describe, expect, it } from "vitest"

import { debounce, type DebounceTimers } from "./debounce"

/** A clock a test drives by hand. */
function fakeTimers() {
  let next = 1
  const scheduled = new Map<number, () => void>()
  const timers: DebounceTimers = {
    set: (run) => {
      const handle = next++
      scheduled.set(handle, run)
      return handle
    },
    clear: (handle) => void scheduled.delete(handle),
  }
  return {
    timers,
    /** Run everything currently scheduled. */
    tick() {
      for (const run of [...scheduled.values()]) {
        run()
      }
      scheduled.clear()
    },
  }
}

describe("debounce", () => {
  it("does not run until the delay passes", () => {
    const clock = fakeTimers()
    const calls: string[] = []
    const debounced = debounce((value: string) => calls.push(value), 500, clock.timers)

    debounced.call("a")
    expect(calls).toEqual([])

    clock.tick()
    expect(calls).toEqual(["a"])
  })

  // Typing quickly should save once, with what they ended up with.
  it("keeps only the last call", () => {
    const clock = fakeTimers()
    const calls: string[] = []
    const debounced = debounce((value: string) => calls.push(value), 500, clock.timers)

    debounced.call("a")
    debounced.call("ab")
    debounced.call("abc")
    clock.tick()

    expect(calls).toEqual(["abc"])
  })

  // Closing the sheet must not lose the last sentence to a timer that never ran.
  it("flushes a pending call immediately", () => {
    const clock = fakeTimers()
    const calls: string[] = []
    const debounced = debounce((value: string) => calls.push(value), 500, clock.timers)

    debounced.call("abc")
    debounced.flush()

    expect(calls).toEqual(["abc"])
  })

  it("does nothing when flushed with nothing pending", () => {
    const clock = fakeTimers()
    const calls: string[] = []
    const debounced = debounce((value: string) => calls.push(value), 500, clock.timers)

    debounced.flush()
    expect(calls).toEqual([])
  })

  it("drops a pending call when cancelled", () => {
    const clock = fakeTimers()
    const calls: string[] = []
    const debounced = debounce((value: string) => calls.push(value), 500, clock.timers)

    debounced.call("abc")
    debounced.cancel()
    clock.tick()

    expect(calls).toEqual([])
  })

  it("reports whether something is waiting", () => {
    const clock = fakeTimers()
    const debounced = debounce(() => {}, 500, clock.timers)

    expect(debounced.pending()).toBe(false)
    debounced.call()
    expect(debounced.pending()).toBe(true)
    clock.tick()
    expect(debounced.pending()).toBe(false)
  })
})
