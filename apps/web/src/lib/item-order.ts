import type { Item } from "@nooks/api"

/** Which way a Member moved an Item, one place at a time. */
export type MoveDirection = "up" | "down"

/** What moving an Item one place does, given the way it went. */
export type Mover = (direction: MoveDirection) => void

/** Where a move puts an Item: after another, or at the top. */
export interface MovePlace {
  itemUid: string
  /** The Item to place it after, or empty for the top, as MoveItem reads it. */
  afterItemUid: string
}

/**
 * placeFor is where moving an open Item one step lands it, or null at either end.
 *
 * Measured among the open Items, because those are the ones a Member sees in a row: a
 * ticked Item between two open ones is not a place anybody can see to move past.
 */
export function placeFor(open: readonly Item[], itemUid: string, direction: MoveDirection): MovePlace | null {
  const at = open.findIndex((item) => item.uid === itemUid)
  if (at === -1) {
    return null
  }
  if (direction === "up") {
    if (at === 0) {
      return null
    }
    return { itemUid, afterItemUid: open[at - 2]?.uid ?? "" }
  }
  const next = open[at + 1]
  return next ? { itemUid, afterItemUid: next.uid } : null
}

/**
 * movedAfter is the Items with one moved, the way the Instance will order them.
 *
 * So the row lands where it was put the moment it is put there, rather than when the
 * answer comes back; the answer then replaces this with what is really stored.
 */
export function movedAfter(items: readonly Item[], place: MovePlace): Item[] {
  const moving = items.find((item) => item.uid === place.itemUid)
  if (!moving) {
    return [...items]
  }
  const rest = items.filter((item) => item.uid !== place.itemUid)
  const at = place.afterItemUid === "" ? 0 : rest.findIndex((item) => item.uid === place.afterItemUid) + 1
  return [...rest.slice(0, at), moving, ...rest.slice(at)]
}

/**
 * placeAfterDrop is where dropping an open Item at a new index lands it, or null when it
 * was put back where it started.
 *
 * Read off the order as it will be once the Item is out of its old place and in its new
 * one, so the Item it lands after is whichever ends up just above it.
 */
export function placeAfterDrop(open: readonly Item[], from: number, to: number): MovePlace | null {
  const moving = open[from]
  if (!moving || from === to) {
    return null
  }
  const rest = open.filter((item) => item.uid !== moving.uid)
  return { itemUid: moving.uid, afterItemUid: to === 0 ? "" : (rest[to - 1]?.uid ?? "") }
}
