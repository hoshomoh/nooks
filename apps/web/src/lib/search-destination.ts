import type { SearchHit } from "@nooks/api"

/**
 * Where a search result takes a Member.
 *
 * Pure, so what a result opens can be read and tested without a router — and so the
 * decision lives somewhere other than inside a click handler.
 */
export interface SearchDestination {
  listUid: string
  /** The Item to open the sheet on, when the hit was about one. */
  itemUid?: string
}

/**
 * destinationOf answers what was found, not merely where it lives.
 *
 * A hit on an Item, or on words in its Note, is about that Item, so the sheet opens on
 * it. Landing on the List instead makes a Member search twice for something already
 * found. A hit on a List is the List: there is nothing narrower to open.
 */
export function destinationOf(hit: SearchHit): SearchDestination {
  return hit.itemUid ? { listUid: hit.listUid, itemUid: hit.itemUid } : { listUid: hit.listUid }
}
