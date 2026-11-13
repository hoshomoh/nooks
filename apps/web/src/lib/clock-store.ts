/**
 * The time, as an external store that moves on once a minute, which is as fine as any
 * label reads it. Read during render instead, "Tomorrow" stayed "Tomorrow" past
 * midnight until something else happened to render.
 */

export interface ClockStore {
  subscribe: (listener: () => void) => () => void
  getState: () => Date
}

export interface ClockStoreDeps {
  now: () => Date
  /** every calls tick each period until the function it returns is called. */
  every: (tick: () => void, periodMs: number) => () => void
}

export const TICK_MS = 60_000

export function createClockStore(deps: ClockStoreDeps): ClockStore {
  let state = deps.now()
  const listeners = new Set<() => void>()
  let stop: (() => void) | null = null

  const tick = () => {
    state = deps.now()
    for (const listener of listeners) {
      listener()
    }
  }

  return {
    subscribe: (listener) => {
      listeners.add(listener)
      if (!stop) {
        stop = deps.every(tick, TICK_MS)
      }
      return () => {
        listeners.delete(listener)
        if (listeners.size === 0 && stop) {
          stop()
          stop = null
        }
      }
    },
    getState: () => {
      // Nothing was listening, so nothing kept the time. Either way round, because a
      // clock can be set back.
      const drift = Math.abs(deps.now().getTime() - state.getTime())
      if (!stop && drift >= TICK_MS) {
        state = deps.now()
      }
      return state
    },
  }
}

export const clockStore: ClockStore = createClockStore({
  now: () => new Date(),
  every: (tick, periodMs) => {
    const id = setInterval(tick, periodMs)
    return () => clearInterval(id)
  },
})
