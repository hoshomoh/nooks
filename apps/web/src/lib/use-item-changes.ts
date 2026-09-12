import { useMutation, type UseMutationResult } from "@tanstack/react-query"
import type { ConnectError } from "@connectrpc/connect"

import type { CreateItemVariables, SetDoneVariables } from "./item-mutations"
import { ADD_MUTATION, TICK_MUTATION } from "./queued-changes"

/**
 * What a screen reaches for to tick or add.
 *
 * The hooks carry a key and nothing else: what ticking an Item actually does is
 * registered on the client in item-changes.ts, so a change restored from disk after a
 * reload runs the same code as one made a second ago.
 *
 * Ticking happens on a List, in Today, in Upcoming and in the calendar, and adding
 * happens on a List and from a dated view. Written out at each of those, the offline
 * behaviour would be six pieces of code that have to agree — and the first one somebody
 * forgot would be a tick that looked saved and was not.
 */

/** Everything a Member does to an Item returns this, so a screen reads one shape. */
export type ItemChange<Variables> = UseMutationResult<unknown, ConnectError, Variables, unknown>

/** Ticking an Item off, or putting it back. */
export function useSetDone(): ItemChange<SetDoneVariables> {
  return useMutation<unknown, ConnectError, SetDoneVariables>({ mutationKey: TICK_MUTATION })
}

/** Putting something on a List. */
export function useAddItem(): ItemChange<CreateItemVariables> {
  return useMutation<unknown, ConnectError, CreateItemVariables>({ mutationKey: ADD_MUTATION })
}
