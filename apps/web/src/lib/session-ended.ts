import { Code, ConnectError } from "@connectrpc/connect"
import type { QueryClient } from "@tanstack/react-query"

import { browserCacheDeps, forgetCache, type CacheStoreDeps } from "./cache-store"
import { currentMemberQuery } from "./queries"

/**
 * endedSession reports whether an error is the Instance saying the session is gone.
 *
 * It expired, the password was changed somewhere else, the Member was removed, or this
 * browser is talking to a different Instance than the one it signed in to.
 */
export function endedSession(error: unknown): boolean {
  return error instanceof ConnectError && error.code === Code.Unauthenticated
}

/**
 * forgetSession lets go of who this browser thought was signed in.
 *
 * Who is signed in is cached, and kept on disk with everything else so a List still opens
 * offline, so nothing re-asks the Instance on its own. Every other read then fails as
 * refused while the page guards go on believing the cached answer. The first refusal is
 * where that belief ends.
 *
 * Only what was read goes. Changes still waiting to be sent are kept: signing back in is
 * what decides what happens to them, as it does after any sign-in.
 */
export function forgetSession(queryClient: QueryClient, deps: CacheStoreDeps = browserCacheDeps): void {
  queryClient.removeQueries()
  queryClient.setQueryData(currentMemberQuery.queryKey, null)
  forgetCache(deps)
}
