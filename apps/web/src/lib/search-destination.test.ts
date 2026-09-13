import { SearchHitKind, type SearchHit } from "@nooks/api"
import { describe, expect, it } from "vitest"

import { destinationOf } from "./search-destination"

/** hit builds a result of the given kind, the way the Instance sends one. */
function hit(kind: SearchHitKind, itemUid = ""): SearchHit {
  return {
    $typeName: "nooks.api.v1.SearchHit",
    kind,
    listUid: "list_groceries",
    listName: "Groceries",
    text: "tomatoes",
    itemUid,
  }
}

describe("where a result takes a Member", () => {
  it("opens the Item a result is about", () => {
    expect(destinationOf(hit(SearchHitKind.ITEM, "item_tomatoes"))).toEqual({
      listUid: "list_groceries",
      itemUid: "item_tomatoes",
    })
  })

  it("opens the Item when the words were inside its Note", () => {
    // A Note belongs to an Item. Landing on the List would leave the Member looking for
    // the row that already matched.
    expect(destinationOf(hit(SearchHitKind.NOTE, "item_tomatoes"))).toEqual({
      listUid: "list_groceries",
      itemUid: "item_tomatoes",
    })
  })

  it("opens the List for a List, because there is nothing narrower", () => {
    expect(destinationOf(hit(SearchHitKind.LIST))).toEqual({ listUid: "list_groceries" })
  })

  it("never asks for an Item it was not given", () => {
    // The Instance leaves itemUid empty on a List hit. Sending "" as the open sheet
    // would be asking the List screen to open an Item that does not exist.
    expect(destinationOf(hit(SearchHitKind.LIST)).itemUid).toBeUndefined()
  })
})
