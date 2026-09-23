import { describe, expect, it } from "vitest"

import { listUidIn, openItemIn } from "./live-wiring"

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

/*
The Note held open is read from the search, and the List from the path.

Opening a Note is a state of the List screen rather than a place of its own, so it is a
search parameter. The claim on it lives on the live connection, which is what makes it
safe: closing the sheet navigates, the connection reconnects without it, and the Note is
free. A laptop closed mid-edit is the same thing more slowly.
*/
describe("what the stream is told about a screen", () => {
  it("reads the open note out of the search", () => {
    expect(openItemIn("?item=item_bread")).toBe("item_bread")
    expect(openItemIn("?item=item_bread&other=1")).toBe("item_bread")
  })

  it("says nothing when no note is open", () => {
    expect(openItemIn("")).toBe("")
    expect(openItemIn("?sort=name")).toBe("")
  })
})
