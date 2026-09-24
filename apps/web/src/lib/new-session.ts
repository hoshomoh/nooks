import type { QueryClient } from "@tanstack/react-query"

import { browserCacheDeps, forgetCache, type CacheStoreDeps } from "./cache-store"
import { pendingTickStore, type PendingTickStore } from "./pending-tick-store"

/**
 * Forgets everything the cache learned before the session changed.
 *
 * Router loaders read with `ensureQueryData`, which hands back a stale answer. So
 * invalidating alone leaves a loader reading "nobody is signed in" a moment after
 * somebody is, and bouncing the Member back to the page they just finished.
 *
 * Called after anything that changes who the browser is: first run, signing in,
 * joining, setting a new password, signing out, and wiping the Instance. The copy on
 * disk goes too, since the next person at this browser is not necessarily the last —
 * which is the whole reason somebody signs out of a shared tablet.
 *
 * A tick somebody reached for on the public list goes with it, for the same reason and
 * by the same argument. It is kept on disk because signing in can take days, so it
 * survived a sign-out and was applied by whoever signed in next, under their name. The
 * two callers that mean to apply it do so before calling this, so forgetting it here
 * takes nothing from them.
 */
export function startNewSession(
  queryClient: QueryClient,
  deps: CacheStoreDeps = browserCacheDeps,
  ticks: PendingTickStore = pendingTickStore,
): void {
  queryClient.clear()
  forgetCache(deps)
  ticks.forget()
}
