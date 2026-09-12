import { Code, ConnectError, type Interceptor } from "@connectrpc/connect"

import type { ConnectionStore } from "./connection-store"

/**
 * Tells the connection store what every request found out.
 *
 * One interceptor rather than a check at each call site: the app makes requests from
 * dozens of places, and a banner that is only right on the screens somebody remembered
 * to wire is worse than no banner.
 *
 * A refusal still counts as contact. "You may not do that" is the Instance answering,
 * so only a request that never reached it marks the connection lost.
 */
export function reportReachability(store: ConnectionStore): Interceptor {
  return (next) => async (request) => {
    try {
      const response = await next(request)
      store.markSeen()
      return response
    } catch (error) {
      if (isUnreachable(error)) {
        store.markUnreachable()
      } else {
        store.markSeen()
      }
      throw error
    }
  }
}

/**
 * isUnreachable reports whether the request never got an answer.
 *
 * Connect turns a failed fetch into Unavailable, which is also what a server that is
 * up but shutting down returns. Both mean the same thing to a Member holding a list:
 * ask again later.
 */
function isUnreachable(error: unknown): boolean {
  if (error instanceof ConnectError) {
    return error.code === Code.Unavailable
  }
  return error instanceof TypeError
}
