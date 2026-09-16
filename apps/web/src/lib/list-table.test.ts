import { describe, expect, it } from "vitest"
import type { List } from "@nooks/api"

import { filterLists, pageOf, sortLists } from "./list-table"

function list(name: string, over: Partial<List> = {}): List {
  return {
    uid: name,
    name,
    isOwner: true,
    isPinned: false,
    openCount: 1,
    doneCount: 0,
    archivedAt: "",
    archivedByName: "",
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

describe("archived Lists", () => {
  const lists = [
    list("Groceries", { openCount: 3 }),
    list("Party", { openCount: 0, doneCount: 12 }),
    list("Move — Kreuzberg", { archivedAt: "2026-08-04T10:00:00Z", archivedByName: "Anna" }),
  ]

  /*
   * Archiving is for a List somebody has finished with but does not want to lose.
   *
   * Leaving it in All would defeat that: the point is that it stops turning up where
   * somebody is looking for what they are working on.
   */
  it("keeps an archived List out of every other filter", () => {
    expect(names(filterLists(lists, "all"))).toEqual(["Groceries", "Party"])
    expect(names(filterLists(lists, "active"))).toEqual(["Groceries"])
    expect(names(filterLists(lists, "completed"))).toEqual(["Party"])
  })

  it("shows them under Archived, which is the only place they are", () => {
    expect(names(filterLists(lists, "archived"))).toEqual(["Move — Kreuzberg"])
  })

  // Archived wins over finished: a List can be both, and it is put away either way.
  it("files a finished List that was archived under Archived", () => {
    const both = [list("Done and gone", { openCount: 0, doneCount: 5, archivedAt: "2026-08-04T10:00:00Z" })]
    expect(names(filterLists(both, "completed"))).toEqual([])
    expect(names(filterLists(both, "archived"))).toEqual(["Done and gone"])
  })
})

describe("cutting the table into pages", () => {
  const many = (count: number) => Array.from({ length: count }, (_, i) => list(`List ${i + 1}`))

  it("fills a page and says where it sits", () => {
    const { rows, from, to, total, pages } = pageOf(many(60), 1, 25)

    expect(rows).toHaveLength(25)
    expect([from, to, total, pages]).toEqual([1, 25, 60, 3])
  })

  it("gives the last page what is left rather than a full one", () => {
    const { rows, from, to } = pageOf(many(60), 3, 25)

    expect(rows).toHaveLength(10)
    expect([from, to]).toEqual([51, 60])
  })

  /*
   * The page is in the address, so it can outlive what it was counting.
   *
   * Narrowing the filter while on page 3 would otherwise answer with an empty table and
   * nothing to say why, which reads as "you have no lists".
   */
  it("clamps a page that is past the end", () => {
    const { page, rows } = pageOf(many(10), 99, 25)

    expect(page).toBe(1)
    expect(rows).toHaveLength(10)
  })

  it("clamps a page below the first", () => {
    expect(pageOf(many(10), 0, 25).page).toBe(1)
    expect(pageOf(many(10), -4, 25).page).toBe(1)
  })

  it("has one page and no range when there is nothing to show", () => {
    expect(pageOf([], 1, 25)).toMatchObject({ rows: [], page: 1, pages: 1, from: 0, to: 0, total: 0 })
  })

  // Everything fits, so there is nothing to page through.
  it("has one page for a household with a handful of lists", () => {
    expect(pageOf(many(9), 1, 25).pages).toBe(1)
  })
})
