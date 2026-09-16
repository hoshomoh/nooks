import { describe, expect, it } from "vitest"
import type { List } from "@nooks/api"

import { filterLists, sortLists } from "./list-table"

function list(name: string, over: Partial<List> = {}): List {
  return {
    uid: name,
    name,
    isOwner: true,
    isPinned: false,
    openCount: 1,
    doneCount: 0,
    updatedAt: "2026-01-01T00:00:00Z",
    ...over,
  } as List
}

const names = (lists: List[]) => lists.map((l) => l.name)

describe("which Lists the table shows", () => {
  const lists = [
    list("Groceries", { openCount: 3 }),
    list("Party", { openCount: 0, doneCount: 12 }),
    list("Brand new", { openCount: 0, doneCount: 0 }),
  ]

  it("shows everything under All", () => {
    expect(names(filterLists(lists, "all"))).toEqual(["Groceries", "Party", "Brand new"])
  })

  // An empty List is not finished, so Active is where somebody finds the one they just
  // made and have not filled in yet.
  it("counts a List nobody has filled in as active", () => {
    expect(names(filterLists(lists, "active"))).toEqual(["Groceries", "Brand new"])
  })

  it("shows only the finished ones under Completed", () => {
    expect(names(filterLists(lists, "completed"))).toEqual(["Party"])
  })

  it("does not hand back the array it was given", () => {
    const all = filterLists(lists, "all")
    all.reverse()
    expect(names(lists)).toEqual(["Groceries", "Party", "Brand new"])
  })
})

describe("how the table is ordered", () => {
  it("puts what changed last at the top", () => {
    const sorted = sortLists(
      [
        list("Old", { updatedAt: "2026-01-01T00:00:00Z" }),
        list("Newest", { updatedAt: "2026-09-16T00:00:00Z" }),
        list("Middle", { updatedAt: "2026-05-01T00:00:00Z" }),
      ],
      "updated",
    )
    expect(names(sorted)).toEqual(["Newest", "Middle", "Old"])
  })

  /*
   * The order a person reads, not the order bytes sort in.
   *
   * Comparing strings directly files every capital above every lowercase, so "avocados"
   * lands under "Groceries" instead of at the top.
   */
  it("reads names the way a person does", () => {
    const sorted = sortLists(
      [list("Groceries"), list("avocados"), list("Flat jobs"), list("bike parts")],
      "name",
      "en",
    )
    expect(names(sorted)).toEqual(["avocados", "bike parts", "Flat jobs", "Groceries"])
  })

  it("puts the fullest List first, and settles ties by name", () => {
    const sorted = sortLists(
      [
        list("Quiet", { openCount: 1 }),
        list("Busy", { openCount: 9 }),
        list("Also one", { openCount: 1 }),
      ],
      "open",
      "en",
    )
    expect(names(sorted)).toEqual(["Busy", "Also one", "Quiet"])
  })

  it("does not reorder the array it was given", () => {
    const original = [list("B"), list("A")]
    sortLists(original, "name", "en")
    expect(names(original)).toEqual(["B", "A"])
  })
})
