import { dehydrate, hydrate, type DehydratedState, type QueryClient } from "@tanstack/react-query"

import { debounce, type DebounceTimers } from "./debounce"

/**
 * What the browser already knows, kept across a reload.
 *
 * Without this, reading offline works right up until somebody closes the tab. The
 * cache is in memory, so a Member who opens the app on the train is shown a spinner
 * over data their machine had all along.
 *
 * What is written is what the cache already held: nothing is asked for in order to
 * store it, and nothing is stored that the Member has not already been shown.
 */

/** CACHE_KEY is where the last known answers wait. */
export const CACHE_KEY = "nooks.cache"

/**
 * How old a stored cache may be before it is thrown away rather than shown.
 *
 * A day. Long enough that a commute is covered, short enough that nobody is handed a
 * shopping list from last week as though it were current.
 */
export const CACHE_MAX_AGE_MS = 24 * 60 * 60 * 1000

/** How long writing waits for the cache to settle, so a burst is written once. */
const SETTLE_MS = 1000

/** What is written, with enough to know whether it is still worth reading. */
interface StoredCache {
  /** When it was written, as epoch milliseconds. */
  at: number
  state: DehydratedState
}

export interface CacheStoreDeps {
  storage: Pick<Storage, "getItem" | "setItem" | "removeItem">
  now: () => number
  timers?: DebounceTimers
}

/**
 * restoreCache puts the last known answers back, if they are recent enough.
 *
 * Called before the first render so a loader reading with `ensureQueryData` finds
 * something rather than reaching for the network.
 */
export function restoreCache(queryClient: QueryClient, deps: CacheStoreDeps): void {
  const stored = read(deps)
  if (!stored) {
    return
  }
  if (deps.now() - stored.at > CACHE_MAX_AGE_MS) {
    forgetCache(deps)
    return
  }
  hydrate(queryClient, stored.state)
}

/**
 * persistCache writes the cache as it changes, and answers how to stop.
 *
 * Debounced: ticking five things off writes once rather than five times, and the
 * write is the expensive half of this.
 */
export function persistCache(queryClient: QueryClient, deps: CacheStoreDeps): () => void {
  const write = debounce(
    () => {
      try {
        const stored: StoredCache = { at: deps.now(), state: dehydrate(queryClient) }
        deps.storage.setItem(CACHE_KEY, JSON.stringify(stored))
      } catch {
        // A private window, or a cache larger than the quota. Reading offline stops
        // working; nothing else does, so there is nothing to tell the Member.
      }
    },
    SETTLE_MS,
    deps.timers,
  )

  const stop = queryClient.getQueryCache().subscribe(() => write.call())

  // Once up front as well as on every change: a cache that was restored and then read
  // without changing would otherwise never be written back, and would age out of the
  // window on a machine that is only ever used offline.
  write.call()

  return () => {
    stop()
    write.cancel()
  }
}

/**
 * forgetCache throws away what was stored.
 *
 * Called when the session changes. The cache holds what somebody was shown while they
 * were signed in, and the next person at this browser is not necessarily them.
 */
export function forgetCache(deps: CacheStoreDeps): void {
  try {
    deps.storage.removeItem(CACHE_KEY)
  } catch {
    // Nothing was stored, so there is nothing to forget.
  }
}

/** read parses what is stored, or answers nothing rather than throwing. */
function read(deps: CacheStoreDeps): StoredCache | null {
  try {
    const raw = deps.storage.getItem(CACHE_KEY)
    if (!raw) {
      return null
    }
    const parsed: unknown = JSON.parse(raw)
    if (typeof parsed !== "object" || parsed === null || !("at" in parsed) || !("state" in parsed)) {
      return null
    }
    return parsed as StoredCache
  } catch {
    return null
  }
}

/** The surroundings the application uses. */
export const browserCacheDeps: CacheStoreDeps = {
  storage:
    typeof window === "undefined"
      ? { getItem: () => null, setItem: () => {}, removeItem: () => {} }
      : window.localStorage,
  now: () => Date.now(),
}
