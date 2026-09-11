import { useSyncExternalStore } from "react"
import type { List } from "@nooks/api"

import { lastListStore } from "./last-list-store"

/** Which List a new Item should land on, and how to remember a different one. */
export interface LastList {
  /** The List to add to, or undefined when the Member has none at all. */
  target: List | undefined
  remember: (listUid: string) => void
}

/**
 * Picks the List a cross-list view adds to.
 *
 * The one last added to, when it is still reachable — a List that was deleted or
 * unshared cannot be the answer. Otherwise the first one, which is at least a List the
 * Member has.
 */
export function useLastList(lists: List[]): LastList {
  const remembered = useSyncExternalStore(
    lastListStore.subscribe,
    lastListStore.getUid,
    lastListStore.getUid,
  )

  const target = lists.find((list) => list.uid === remembered) ?? lists[0]
  return { target, remember: lastListStore.remember }
}
