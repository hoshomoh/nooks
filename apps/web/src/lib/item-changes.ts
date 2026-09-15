import type { QueryClient } from "@tanstack/react-query"

import { listClient } from "./api"
import type { CreateItemVariables, SetDoneVariables } from "./item-mutations"
import { lastListStore } from "./last-list-store"
import { ADD_MUTATION, provisionalItem, TICK_MUTATION, withItem, withTick } from "./queued-changes"
import { refreshLists } from "./refresh"

/**
 * The two changes a Member makes most, defined on the client rather than in a hook.
 *
 * This is what lets a change survive a reload. A paused change is written out with the
 * cache as a record of what was asked for, not the function that does it, so the client
 * needs the function registered under a key to pick it back up. Defined in a
 * `useMutation` instead, a tick made in a supermarket basement comes back unrunnable.
 *
 * The screens still use hooks, in use-item-changes.ts; they carry the key and nothing
 * else.
 *
 * A refused change converges rather than rolls back. A tick queued offline can return
 * to an Item somebody has deleted, and the honest answer is what the Instance has now,
 * not the value this browser held before.
 */
export function registerItemChanges(queryClient: QueryClient): void {
  queryClient.setMutationDefaults(TICK_MUTATION, {
    mutationFn: ({ itemUid, done }: SetDoneVariables) => listClient.setItemDone({ itemUid, done }),
    onMutate: ({ itemUid, done }: SetDoneVariables) => {
      // Everywhere the Item is cached, not just the screen it was ticked on: the same
      // Item is on its List and in every dated view, and a tick that only lands on one
      // comes back the moment the Member changes screen.
      for (const queryKey of [["list"], ["dated"], ["search"]]) {
        queryClient.setQueriesData({ queryKey }, (data) => withTick(data, itemUid, done))
      }
    },
    onSuccess: () => refreshLists(queryClient),
    onError: () => refreshLists(queryClient),
  })

  queryClient.setMutationDefaults(ADD_MUTATION, {
    mutationFn: ({ listUid, label, quantity, dueOn }: CreateItemVariables) =>
      listClient.createItem({ listUid, label, quantity, dueOn }),
    onMutate: ({ listUid, label, quantity, dueOn }: CreateItemVariables) => {
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
    onError: () => refreshLists(queryClient),
  })
}

/**
 * resumeQueuedChanges sends whatever was waiting when the tab was last closed.
 *
 * Called once, after the cache has been restored. Reconnecting while the app is open
 * needs no help — the query layer resumes paused changes itself the moment it is told
 * the Instance is reachable again — but nothing resumes a change that was restored
 * from disk, because as far as the client is concerned it never paused.
 */
export function resumeQueuedChanges(queryClient: QueryClient): void {
  void queryClient.resumePausedMutations()
}
