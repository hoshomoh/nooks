import { MutationCache, QueryCache, QueryClient } from "@tanstack/react-query"

import { worthAskingAgain } from "./retry"
import { endedSession, forgetSession } from "./session-ended"

/** What the app does once the Instance says the session is over. */
export type SessionEnded = () => void

/**
 * createQueryClient is the one query client the app reads and writes through.
 *
 * Every read and every change passes its failure through here first, so a session that
 * ended is noticed by whichever request finds out, rather than by a check each screen
 * would have to remember.
 */
export function createQueryClient(onSessionEnded: SessionEnded): QueryClient {
  const ended = (error: unknown) => {
    if (endedSession(error)) {
      forgetSession(queryClient)
      onSessionEnded()
    }
  }

  const queryClient: QueryClient = new QueryClient({
    queryCache: new QueryCache({ onError: ended }),
    mutationCache: new MutationCache({ onError: ended }),
    defaultOptions: {
      queries: {
        retry: worthAskingAgain,
        refetchOnWindowFocus: false,
      },
    },
  })
  return queryClient
}
