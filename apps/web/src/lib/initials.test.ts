import { describe, expect, it } from "vitest"

import { initialsOf } from "./initials"

describe("the letters an avatar carries", () => {
  it("is the first two, in capitals", () => {
    expect(initialsOf("Anna")).toBe("AN")
  })

  // One name or two, the same person gets the same avatar.
  it("does not read the second word", () => {
    expect(initialsOf("Anna Schmidt")).toBe("AN")
  })

  it("gives back what little there is of a short name", () => {
    expect(initialsOf("A")).toBe("A")
    expect(initialsOf("")).toBe("")
  })
})
