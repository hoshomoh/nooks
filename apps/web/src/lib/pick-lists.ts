import { SearchHitKind, type List, type SearchHit } from "@nooks/api"

/** A List as a picker names it: enough to show a row and to save the answer. */
export interface PickableList {
  uid: string
  name: string
}

/**
 * listsAmong keeps the Lists out of a set of search results.
 *
 * Search answers with Items and Notes as well, and the same List can be named by
 * several of them. A picker wants each List once and only because it is called what was
 * typed, which is what a hit of kind LIST already means.
 */
export function listsAmong(hits: readonly SearchHit[]): PickableList[] {
  const seen = new Set<string>()
  const out: PickableList[] = []

  for (const hit of hits) {
    if (hit.kind !== SearchHitKind.LIST || seen.has(hit.listUid)) {
      continue
    }
    seen.add(hit.listUid)
    out.push({ uid: hit.listUid, name: hit.listName || hit.text })
  }

  return out
}

/** pickable keeps only what a picker needs of a List it was handed. */
export function pickable(lists: readonly List[]): PickableList[] {
  return lists.map((list) => ({ uid: list.uid, name: list.name }))
}
