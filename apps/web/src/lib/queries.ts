import { ConnectError, Code } from "@connectrpc/connect"
import { queryOptions } from "@tanstack/react-query"
import type { Member } from "@nooks/api"

import { authClient, instanceClient } from "./api"

/**
 * What this Instance is, and whether it has been set up. Reachable without signing in,
 * because the sign-in page and the Public list both need it.
 */
export const instanceQuery = queryOptions({
  queryKey: ["instance"],
  queryFn: () => instanceClient.getInstance({}),
  // The Instance's name and version change about as often as the binary does.
  staleTime: 5 * 60 * 1000,
})

/**
 * Whoever the session cookie belongs to, or null when nobody is signed in.
 *
 * Anonymous is an answer, not a failure: first run, sign-in and the Public list are all
 * legitimately signed out. Only unauthenticated is folded into null — anything else is
 * a real error and must still surface.
 */
export const currentMemberQuery = queryOptions<Member | null>({
  queryKey: ["current-member"],
  queryFn: async () => {
    try {
      const res = await authClient.getCurrentMember({})
      return res.member ?? null
    } catch (error) {
      if (error instanceof ConnectError && error.code === Code.Unauthenticated) {
        return null
      }
      throw error
    }
  },
  // Signing in or out invalidates this explicitly, so it need not be refetched on a
  // window focus.
  staleTime: Infinity,
})
