import type { QueryClient } from "@tanstack/react-query"

import { browserCacheDeps, forgetCache } from "./cache-store"

/**
 * Forgets everything the cache learned before the session changed.
 *
 * Router loaders read with `ensureQueryData`, which hands back a cached answer even
 * once it is stale. Invalidating alone therefore leaves a loader reading "nobody is
 * signed in", or "this instance needs setting up", a moment after neither is true —
 * and bouncing the Member straight back to the page they just finished.
 *
 * Called after anything that changes who the browser is: first run, signing in,
 * joining, and setting a new password. One function so that the next screen to do it
 * does not have to rediscover why invalidating is not enough.
 *
 * The copy on disk goes with it. It holds what somebody was shown while they were
 * signed in, and the next person at this browser is not necessarily them.
 */
export function startNewSession(queryClient: QueryClient): void {
  queryClient.clear()
  forgetCache(browserCacheDeps)
}
