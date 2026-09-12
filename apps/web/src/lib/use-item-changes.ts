import { useMutation, useQueryClient, type UseMutationResult } from "@tanstack/react-query"
import type { ConnectError } from "@connectrpc/connect"

import { listClient } from "./api"
import type { CreateItemVariables, SetDoneVariables } from "./item-mutations"
import {
  ADD_MUTATION,
  provisionalItem,
  TICK_MUTATION,
  withItem,
  withTick,
} from "./queued-changes"
import { lastListStore } from "./last-list-store"
import { refreshLists } from "./refresh"

/**
 * The two changes a Member makes most, in one place.
 *
 * Ticking happens on a List, in Today, in Upcoming and in the calendar, and adding
 * happens on a List and from a dated view. Written out at each of those, the offline
 * behaviour would be six separate pieces of code that have to agree — and the first
 * one somebody forgot would be a tick that looked saved and was not.
 *
 * Each writes the answer into the cache before sending it. That is what makes ticking
 * and adding work with no connection at all: the change is shown, the mutation is
 * paused rather than failed, and the row it affects says it has not been sent.
 */

/** Everything a Member does to an Item returns this, so a screen reads one shape. */
export type ItemChange<Variables> = UseMutationResult<unknown, ConnectError, Variables, unknown>

/** Ticking an Item off, or putting it back. */
export function useSetDone(): ItemChange<SetDoneVariables> {
  const queryClient = useQueryClient()

  return useMutation({
    // Keyed so a paused tick can be found again and shown as unsent.
    mutationKey: TICK_MUTATION,
    mutationFn: ({ itemUid, done }: SetDoneVariables) => listClient.setItemDone({ itemUid, done }),
    onMutate: ({ itemUid, done }) => {
      // Everywhere the Item is cached, not just the screen it was ticked on: the same
      // Item is on its List and in every dated view, and a tick that only lands on one
      // comes back the moment the Member changes screen.
      for (const queryKey of [["list"], ["dated"], ["search"]]) {
        queryClient.setQueriesData({ queryKey }, (data) => withTick(data, itemUid, done))
      }
    },
    onSuccess: () => refreshLists(queryClient),
  })
}

/** Putting something on a List. */
export function useAddItem(): ItemChange<CreateItemVariables> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationKey: ADD_MUTATION,
    mutationFn: ({ listUid, label, quantity, dueOn }: CreateItemVariables) =>
      listClient.createItem({ listUid, label, quantity, dueOn }),
    onMutate: ({ listUid, label, quantity, dueOn }) => {
      // Shown at the foot of the List straight away, with a uid this browser invented.
      // The refresh once it lands replaces it with the Instance's own.
      const provisional = provisionalItem({ label, quantity, dueOn })
      queryClient.setQueryData(["list", listUid], (data: unknown) => withItem(data, provisional))

      // Adding to a List is what makes it the one Today and Upcoming will add to next.
      // Remembered on the way out rather than on the way back, so it holds even while
      // the change is still waiting to be sent.
      lastListStore.remember(listUid)
    },
    onSuccess: () => refreshLists(queryClient),
  })
}
