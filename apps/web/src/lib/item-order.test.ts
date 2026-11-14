import { describe, expect, it } from "vitest"
import type { Item } from "@nooks/api"

import { movedAfter, placeFor } from "./item-order"

/** shelf is Items in their order, by label, which is all this reads. */
function shelf(...labels: string[]): Item[] {
  return labels.map((label) => ({ uid: `item_${label}`, label }) as Item)
}

const labelsOf = (items: readonly Item[]) => items.map((item) => item.label)

describe("moving an Item one place", () => {
  const open = shelf("milk", "oats", "eggs", "bread")

  it("goes up by landing after the one two above it", () => {
    expect(placeFor(open, "item_eggs", "up")).toEqual({ itemUid: "item_eggs", afterItemUid: "item_milk" })
  })

  it("goes to the top from second place", () => {
    expect(placeFor(open, "item_oats", "up")).toEqual({ itemUid: "item_oats", afterItemUid: "" })
  })

  it("goes down by landing after the one below it", () => {
    expect(placeFor(open, "item_oats", "down")).toEqual({ itemUid: "item_oats", afterItemUid: "item_eggs" })
  })

  // Nothing to move past, so nothing is sent.
  it("stays put at either end", () => {
    expect(placeFor(open, "item_milk", "up")).toBeNull()
    expect(placeFor(open, "item_bread", "down")).toBeNull()
  })
})

describe("showing a move before it is answered", () => {
  const items = shelf("milk", "oats", "eggs", "bread")

  it("puts the Item after the one named", () => {
    expect(labelsOf(movedAfter(items, { itemUid: "item_milk", afterItemUid: "item_eggs" }))).toEqual([
      "oats",
      "eggs",
      "milk",
      "bread",
    ])
  })

  it("puts the Item first when nothing is named", () => {
    expect(labelsOf(movedAfter(items, { itemUid: "item_bread", afterItemUid: "" }))).toEqual([
      "bread",
      "milk",
      "oats",
      "eggs",
    ])
  })

  it("leaves the List alone for an Item it does not hold", () => {
    expect(labelsOf(movedAfter(items, { itemUid: "item_gone", afterItemUid: "" }))).toEqual(labelsOf(items))
  })
})
