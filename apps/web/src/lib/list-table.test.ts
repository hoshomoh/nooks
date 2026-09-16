import { describe, expect, it } from "vitest"
import { ListOrder, ListStatus as StatusWire } from "@nooks/api"

import type { ListListsResponse } from "@nooks/api"

import { LIST_SORTS, LIST_STATUSES, boundsOf, sortOnTheWire, statusOnTheWire } from "./list-table"

/** answered is one page as the server sends it back. */
function answered(
  total: number,
  page: number,
  shown: number,
  over: Partial<ListListsResponse> = {},
): ListListsResponse {
  return {
    total,
    page,
    pageSize: 25,
    lists: Array.from({ length: shown }),
    atLeast: false,
    ...over,
  } as ListListsResponse
}

describe("the words in the address and the enums on the wire", () => {
  /*
   * All is the absence of a filter, not a fourth one.
   *
   * Sending an enum of its own would need the server to know a word that means "do not
   * narrow", which is what an unset field already means.
   */
  it("asks for nothing in particular under All", () => {
    expect(statusOnTheWire("all")).toBe(StatusWire.UNSPECIFIED)
  })

  it("has an enum for every filter the address can hold", () => {
    for (const status of LIST_STATUSES) {
      expect(statusOnTheWire(status)).not.toBeUndefined()
    }
  })

  it("has an enum for every order the address can hold", () => {
    for (const sort of LIST_SORTS) {
      expect(sortOnTheWire(sort)).not.toBeUndefined()
    }
  })

  it("orders by name when the address says so", () => {
    expect(sortOnTheWire("name")).toBe(ListOrder.NAME)
  })
})

describe("where a page sits in the set", () => {
  it("says the range a full first page covers, and that there is more", () => {
    expect(boundsOf(answered(60, 1, 25))).toEqual({ more: true, from: 1, to: 25 })
  })

  it("gives the last page the range that is left, and no more after it", () => {
    expect(boundsOf(answered(60, 3, 10))).toEqual({ more: false, from: 51, to: 60 })
  })

  // Everything fits, so there is nothing to page through.
  it("offers no next page to a household with a handful of lists", () => {
    expect(boundsOf(answered(9, 1, 9)).more).toBe(false)
  })

  it("has no range and nothing after it when there is nothing to show", () => {
    expect(boundsOf(answered(0, 1, 0))).toEqual({ more: false, from: 0, to: 0 })
  })

  // The server's size, not the one that was asked for: a caller asking for a thousand
  // rows is answered with the ceiling, and the range has to count what it got.
  it("counts in the page size the server answered with", () => {
    expect(boundsOf(answered(600, 3, 100, { pageSize: 100 }))).toMatchObject({ from: 201, to: 300 })
  })

  /*
   * Past the counting ceiling the total is a floor, so it cannot say where the end is.
   *
   * A full page is the signal instead: there is more while the server keeps filling one,
   * and the first short page is the last.
   */
  it("keeps offering the next page while the total is only a floor", () => {
    expect(boundsOf(answered(1000, 40, 25, { atLeast: true })).more).toBe(true)
  })

  it("stops at the first page that came back short, floor or not", () => {
    expect(boundsOf(answered(1000, 41, 9, { atLeast: true })).more).toBe(false)
  })

  /*
   * A page past the end still has to say something.
   *
   * The page is in the address, so it can outlive what it was counting: ticking the
   * last thing off the last page leaves somebody standing on a page that is no longer
   * there. Printing no range, and letting Previous stay live, is what gets them back.
   */
  it("prints no range for a page past the end", () => {
    expect(boundsOf(answered(10, 99, 0))).toMatchObject({ more: false, from: 0 })
  })
})
