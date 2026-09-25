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
      "",
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

  it("reads a blank line as a separator, not as an empty paragraph", () => {
    // Read as a block it is an empty paragraph, a whole line of height on top of the
    // padding the paragraphs already have, and anything writing ordinary markdown
    // comes out double spaced.
    const blocks = documentFrom("One.\n\nTwo.").content ?? []

    expect(blocks).toHaveLength(2)
    expect(blocks.every((block) => (block.content ?? []).length > 0)).toBe(true)
  })

  it("does not care how many blank lines there were", () => {
    expect(roundTrip("One.\n\n\n\nTwo.")).toBe("One.\n\nTwo.")
  })
})

describe("markdown the design does not have", () => {
  /*
   * A Member's words matter more than the shape they arrived in.
   *
   * A bullet list, because that is the block type nooks has no drawing for. `markerOf`
   * has no `- ` among its prefixes, so `- milk` falls through to a paragraph whose
   * marker length is zero, and the dash stays part of the words. A `- ` prefix added
   * there would consume the marker and the Member would lose it.
   *
   * This used to say `| a | table |` and stopped meaning anything when tables became
   * one of the seven blocks the design draws. It passed either way, because a single
   * pipe row with no separator under it is not a table to begin with: it is text with
   * pipes in it, which would survive whatever the design did.
   *
   * The MCP `update_item` tool promises this to an assistant in as many words, so that
   * an assistant writing a shopping list knows the characters are kept rather than
   * dropped: "Bullet and numbered lists are not rendered and survive as the literal
   * characters typed, so use a checklist instead."
   */
  it("keeps the text of a block type nooks cannot draw", () => {
    expect(documentFrom("- milk").content?.[0]?.type).toBe("paragraph")
    expect(roundTrip("- milk")).toBe("- milk")
    expect(roundTrip("1. milk")).toBe("1. milk")
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

describe("dividers", () => {
  it("reads a line of three dashes as a rule", () => {
    expect(documentFrom("Above\n---\nBelow").content?.map((node) => node.type)).toEqual([
      "paragraph",
      "horizontalRule",
      "paragraph",
    ])
  })

  it("writes one back", () => {
    expect(markdownFrom(documentFrom("Above\n\n---\n\nBelow"))).toBe("Above\n\n---\n\nBelow")
  })

  it("leaves dashes that are part of a sentence alone", () => {
    // "--- and then some words" is a sentence about dashes, not a rule.
    expect(documentFrom("--- ish").content?.[0]?.type).toBe("paragraph")
  })
})

describe("tables", () => {
  const markdown = ["| Bean | Grind | Price |", "| --- | --- | --- |", "| House | Filter | 11,00 |"].join("\n")

  it("reads a heading row and a body row", () => {
    const table = documentFrom(markdown).content?.[0]
    expect(table?.type).toBe("table")
    expect(table?.content?.[0]?.content?.map((cell) => cell.type)).toEqual([
      "tableHeader",
      "tableHeader",
      "tableHeader",
    ])
    expect(table?.content?.[1]?.content?.[0]?.type).toBe("tableCell")
  })

  it("writes it back exactly", () => {
    expect(markdownFrom(documentFrom(markdown))).toBe(markdown)
  })

  it("needs the separator row before it is a table", () => {
    // One row on its own is a sentence somebody started with a pipe.
    expect(documentFrom("| not a table").content?.[0]?.type).toBe("paragraph")
  })

  it("accepts alignment colons and forgets them", () => {
    // DESIGN.md §10 decides alignment, so there is nothing here to store.
    const aligned = ["| A | B |", "| :--- | ---: |", "| 1 | 2 |"].join("\n")
    expect(markdownFrom(documentFrom(aligned))).toBe(
      ["| A | B |", "| --- | --- |", "| 1 | 2 |"].join("\n"),
    )
  })

  it("pads a row that is short rather than refusing it", () => {
    const ragged = ["| A | B |", "| --- | --- |", "| 1 |"].join("\n")
    expect(markdownFrom(documentFrom(ragged))).toBe(
      ["| A | B |", "| --- | --- |", "| 1 |  |"].join("\n"),
    )
  })

  it("keeps a pipe somebody wrote as a pipe", () => {
    const piped = ["| A |", "| --- |", "| a \\| b |"].join("\n")
    expect(markdownFrom(documentFrom(piped))).toBe(piped)
  })
})
