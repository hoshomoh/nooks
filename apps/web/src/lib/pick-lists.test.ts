import { describe, expect, it } from "vitest"
import { SearchHitKind, type SearchHit } from "@nooks/api"

import { listsAmong, pickable } from "./pick-lists"

/** hit is one search result, of whichever kind. */
function hit(kind: SearchHitKind, listUid: string, listName: string): SearchHit {
  return { kind, listUid, listName, itemUid: "", text: listName } as SearchHit
}

describe("the Lists among a set of search results", () => {
  it("keeps the ones that are Lists", () => {
    const found = listsAmong([
      hit(SearchHitKind.LIST, "list_groceries", "Groceries"),
      hit(SearchHitKind.ITEM, "list_bike", "Bike parts"),
      hit(SearchHitKind.NOTE, "list_move", "Move"),
    ])

    expect(found).toEqual([{ uid: "list_groceries", name: "Groceries" }])
  })

  /*
   * A List is one answer however many times it was matched.
   *
   * Search is asked what a word appears in, not which Lists are called it, so a word in
   * a List's name and again in its Note comes back twice.
   */
  it("names a List once", () => {
    const found = listsAmong([
      hit(SearchHitKind.LIST, "list_move", "Move"),
      hit(SearchHitKind.LIST, "list_move", "Move"),
    ])

    expect(found).toHaveLength(1)
  })

  it("has nothing to offer for a word nothing is called", () => {
    expect(listsAmong([])).toEqual([])
  })
})

describe("what a picker needs of a List", () => {
  it("keeps what it shows and what it saves, and nothing else", () => {
    expect(pickable([{ uid: "list_groceries", name: "Groceries", openCount: 7 }] as never)).toEqual([
      { uid: "list_groceries", name: "Groceries" },
    ])
  })
})
