import { describe, expect, it } from "vitest"

import { listUidIn } from "./live-wiring"

describe("which List a path is showing", () => {
  it("reads the List being looked at", () => {
    expect(listUidIn("/lists/list_groceries")).toBe("list_groceries")
  })

  // Reading an Item's Note is still standing on that List.
  it("reads it from an Item's Note too", () => {
    expect(listUidIn("/lists/list_groceries/items/item_milk")).toBe("list_groceries")
  })

  it("is nothing anywhere else", () => {
    expect(listUidIn("/today")).toBe("")
    expect(listUidIn("/")).toBe("")
    expect(listUidIn("/lists")).toBe("")
  })

  it("reads an identifier that had to be escaped", () => {
    expect(listUidIn("/lists/list%20one")).toBe("list one")
  })
})
