import type { List } from "@nooks/api"

/**
 * How the sidebar divides the Lists a Member can reach.
 *
 * Derived rather than stored, so a List moves the moment it is pinned, shared or
 * finished — there is no state to keep in step.
 */
export interface ListGroups {
  pinned: List[]
  mine: List[]
  shared: List[]
  completed: List[]
}

/**
 * isCompleted reports whether everything on a List has been ticked.
 *
 * Both counts, not just the open one. A List nobody has put anything on yet also has
 * nothing open, and calling that finished would file every new List under Completed the
 * moment it was made.
 */
export function isCompleted(list: List): boolean {
  return list.doneCount > 0 && list.openCount === 0
}

/**
 * groupLists sorts the Lists into the four the sidebar draws.
 *
 * Pinned wins over finished: pinning is something a Member did on purpose, and a List
 * they asked to keep in front of them should not move because they ticked the last
 * thing on it.
 */
export function groupLists(lists: readonly List[]): ListGroups {
  const pinned = lists.filter((list) => list.isPinned)
  const rest = lists.filter((list) => !list.isPinned)

  return {
    pinned,
    mine: rest.filter((list) => list.isOwner && !isCompleted(list)),
    shared: rest.filter((list) => !list.isOwner && !isCompleted(list)),
    // Whoever owns it. A finished List is finished, and splitting them by owner would
    // make two short groups of the same thing.
    completed: rest.filter((list) => isCompleted(list)),
  }
}
