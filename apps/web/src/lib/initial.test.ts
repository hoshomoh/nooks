import { describe, expect, it } from "vitest"

import { initialOf } from "./initial"

describe("initialOf", () => {
  it("is the first letter of the name", () => {
    expect(initialOf("anna")).toBe("A")
    expect(initialOf("Anna Schmidt")).toBe("A")
  })

  // Two UTF-16 units, one character: half of it is not a letter at all.
  it("keeps a character that is wider than one unit whole", () => {
    expect(initialOf("😀 Kitchen")).toBe("😀")
  })

  it("is empty for an empty name", () => {
    expect(initialOf("")).toBe("")
  })
})
