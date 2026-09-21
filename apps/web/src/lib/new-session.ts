import type { QueryClient } from "@tanstack/react-query"

import { browserCacheDeps, forgetCache, type CacheStoreDeps } from "./cache-store"

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
 */
export function startNewSession(queryClient: QueryClient, deps: CacheStoreDeps = browserCacheDeps): void {
  queryClient.clear()
  forgetCache(deps)
}
