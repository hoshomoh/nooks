import { useMutationState } from "@tanstack/react-query"

import { queuedChanges, type QueuedChanges, type QueuedMutation } from "./queued-changes"

/**
 * What the Member has done that the Instance has not seen yet.
 *
 * Read from the mutation cache rather than tracked separately: TanStack Query already
 * knows which changes are paused, and a second list of them is a second thing that can
 * be wrong about it.
 */
export function useQueuedChanges(): QueuedChanges {
  const waiting = useMutationState<QueuedMutation>({
    filters: { status: "pending" },
    select: (mutation) => ({
      isPaused: mutation.state.isPaused,
      variables: mutation.state.variables,
    }),
  })

  return queuedChanges(waiting)
}
