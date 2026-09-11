import { describe, expect, it } from "vitest"

import { documentFrom, markdownFrom } from "./markdown"

/** roundTrip is what a Note becomes after being opened and saved again. */
function roundTrip(markdown: string): string {
  return markdownFrom(documentFrom(markdown))
}

// The property that matters: opening a Note and saving it without touching anything
// must not change a word of it. A storage format the editor quietly rewrites is a
// storage format nobody can trust.
describe("markdown survives being edited", () => {
  it("keeps a paragraph", () => {
    expect(roundTrip("Saturday market, second row.")).toBe("Saturday market, second row.")
  })

  it("keeps a heading", () => {
    expect(roundTrip("### Where")).toBe("### Where")
  })

  it("keeps a quote", () => {
    expect(roundTrip("> They pack up around two.")).toBe("> They pack up around two.")
  })

  it("keeps a checklist, ticked and unticked", () => {
    const list = "- [ ] Ethiopian, whole bean\n- [x] Got it"
    expect(roundTrip(list)).toBe(list)
  })

  it("keeps a code block, including what is inside it", () => {
    const code = "```\ngrind: filter, 250 g\ncash only\n```"
    expect(roundTrip(code)).toBe(code)
  })

  it("keeps a whole Note with every block in it", () => {
    const note = [
      "### Where",
      "Saturday market, second row.",
      "",
      "- [ ] Ethiopian, whole bean",
      "- [x] Ask about the light roast",
      "",
      "> They pack up around two.",
      "",
      "```",
      "grind: filter",
      "```",
    ].join("\n")
    expect(roundTrip(note)).toBe(note)
  })

  it("keeps the blank lines somebody put between paragraphs", () => {
    expect(roundTrip("One.\n\nTwo.")).toBe("One.\n\nTwo.")
  })
})

describe("markdown the design does not have", () => {
  // A Member's words matter more than the shape they arrived in.
  it("keeps the text of a block type Nooks cannot draw", () => {
    expect(roundTrip("| a | table |")).toBe("| a | table |")
  })

  // Every heading is the one size the design has, so a bigger one comes back as that.
  it("brings any heading back as the one heading there is", () => {
    expect(roundTrip("# Where")).toBe("### Where")
  })

  it("does not lose what is inside a fence nobody closed", () => {
    expect(roundTrip("```\ngrind: filter")).toContain("grind: filter")
  })
})

describe("a Note with nothing in it", () => {
  // There has to be somewhere to put the caret.
  it("is one empty paragraph, not nothing at all", () => {
    expect(documentFrom("").content).toHaveLength(1)
  })

  it("writes back out as nothing", () => {
    expect(roundTrip("")).toBe("")
  })
})
