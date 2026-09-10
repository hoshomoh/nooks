/**
 * A trailing debounce that can be flushed.
 *
 * Autosave needs three things: not to fire on every keystroke, to fire once the typing
 * settles, and to fire immediately when the Member closes the sheet — otherwise the
 * last sentence they wrote is lost to a timer that never ran.
 *
 * The clock is injected so the behaviour can be tested without waiting.
 */
export type Debounced<T extends unknown[]> = {
  /** Schedule a call, replacing any pending one. */
  call: (...args: T) => void
  /** Run a pending call now, if there is one. */
  flush: () => void
  /** Drop a pending call without running it. */
  cancel: () => void
  /** Whether a call is waiting. */
  pending: () => boolean
}

export type DebounceTimers = {
  set: (run: () => void, delay: number) => number
  clear: (handle: number) => void
}

const REAL_TIMERS: DebounceTimers = {
  set: (run, delay) => window.setTimeout(run, delay),
  clear: (handle) => window.clearTimeout(handle),
}

export function debounce<T extends unknown[]>(
  run: (...args: T) => void,
  delay: number,
  timers: DebounceTimers = REAL_TIMERS,
): Debounced<T> {
  let handle: number | null = null
  let latest: T | null = null

  const cancel = () => {
    if (handle !== null) {
      timers.clear(handle)
      handle = null
    }
    latest = null
  }

  const fire = () => {
    handle = null
    if (latest) {
      const args = latest
      latest = null
      run(...args)
    }
  }

  return {
    call(...args: T) {
      latest = args
      if (handle !== null) {
        timers.clear(handle)
      }
      handle = timers.set(fire, delay)
    },
    flush() {
      if (handle !== null) {
        timers.clear(handle)
        fire()
      }
    },
    cancel,
    pending: () => handle !== null,
  }
}
