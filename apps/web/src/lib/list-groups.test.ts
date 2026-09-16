import { describe, expect, it } from "vitest"
import type { List } from "@nooks/api"

import { groupLists, isCompleted } from "./list-groups"

/** list is a List as the sidebar reads one: who owns it, and how much is left. */
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
    ...over,
  } as List
}

/** names is what came back, so a test reads as an expectation. */
const names = (lists: List[]) => lists.map((l) => l.name)

describe("whether a List is finished", () => {
  it("is finished when everything on it is ticked", () => {
    expect(isCompleted(list("Party", { openCount: 0, doneCount: 12 }))).toBe(true)
  })

  /*
   * The distinction the whole group rests on.
   *
   * A List nobody has put anything on has nothing open either. Calling that finished
   * files every List under Completed the moment it is made.
   */
  it("is not finished when nobody has put anything on it", () => {
    expect(isCompleted(list("Brand new", { openCount: 0, doneCount: 0 }))).toBe(false)
  })

  it("is not finished while anything is still open", () => {
    expect(isCompleted(list("Groceries", { openCount: 3, doneCount: 40 }))).toBe(false)
  })
})

describe("how the sidebar divides them", () => {
  it("takes a finished List out of My lists", () => {
    const { mine, completed } = groupLists([
      list("Groceries"),
      list("Party", { openCount: 0, doneCount: 12 }),
    ])

    expect(names(mine)).toEqual(["Groceries"])
    expect(names(completed)).toEqual(["Party"])
  })

  it("takes a finished List out of Shared with me too", () => {
    const { shared, completed } = groupLists([
      list("Landlord questions", { isOwner: false }),
      list("Move", { isOwner: false, openCount: 0, doneCount: 4 }),
    ])

    expect(names(shared)).toEqual(["Landlord questions"])
    expect(names(completed)).toEqual(["Move"])
  })

  // Pinning is deliberate. Ticking the last thing off should not move a List somebody
  // asked to keep in front of them.
  it("leaves a pinned List pinned, finished or not", () => {
    const { pinned, completed } = groupLists([
      list("Flat jobs", { isPinned: true, openCount: 0, doneCount: 9 }),
    ])

    expect(names(pinned)).toEqual(["Flat jobs"])
    expect(completed).toEqual([])
  })

  it("gathers finished Lists whoever owns them", () => {
    const { completed } = groupLists([
      list("Mine", { openCount: 0, doneCount: 1 }),
      list("Theirs", { isOwner: false, openCount: 0, doneCount: 1 }),
    ])

    expect(names(completed)).toEqual(["Mine", "Theirs"])
  })

  it("has four empty groups for a Member with no Lists", () => {
    expect(groupLists([])).toEqual({ pinned: [], mine: [], shared: [], completed: [] })
  })
})

describe("what archiving takes out of the sidebar", () => {
  // The sidebar is exactly what archiving removes a List from. It is still sent, so
  // that All lists can show it under the filter.
  it("leaves an archived List out of every group", () => {
    const groups = groupLists([
      list("Groceries"),
      list("Move", { archivedAt: "2026-08-04T10:00:00Z" }),
      list("Party", { isPinned: true, archivedAt: "2026-08-04T10:00:00Z" }),
    ])

    expect(names(groups.mine)).toEqual(["Groceries"])
    expect(groups.pinned).toEqual([])
    expect(groups.completed).toEqual([])
  })
})
