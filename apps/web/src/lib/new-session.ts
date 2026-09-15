import type { QueryClient } from "@tanstack/react-query"

import { browserCacheDeps, forgetCache } from "./cache-store"

/**
 * Forgets everything the cache learned before the session changed.
 *
 * Router loaders read with `ensureQueryData`, which hands back a stale answer. So
 * invalidating alone leaves a loader reading "nobody is signed in" a moment after
 * somebody is, and bouncing the Member back to the page they just finished.
 *
 * Called after anything that changes who the browser is: first run, signing in, joining
 * and setting a new password. The copy on disk goes too, since the next person at this
 * browser is not necessarily the last.
 */
export function startNewSession(queryClient: QueryClient): void {
  queryClient.clear()
  forgetCache(browserCacheDeps)
}
