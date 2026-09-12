import { describe, expect, it } from "vitest"

import { inlineFrom, inlineTo } from "./marks"

/** roundTrip is what a line becomes after being read and written back. */
function roundTrip(line: string): string {
  return inlineTo(inlineFrom(line))
}

describe("reading inline markup", () => {
  it("leaves ordinary prose as one run", () => {
    expect(inlineFrom("Rye flour")).toEqual([{ type: "text", text: "Rye flour" }])
  })

  it("reads the four marks the design names", () => {
    for (const [line, mark] of [
      ["**loud**", "bold"],
      ["*soft*", "italic"],
      ["~~gone~~", "strike"],
      ["`code`", "code"],
    ] as const) {
      expect(inlineFrom(line)).toEqual([
        { type: "text", text: line.replaceAll(/[*~`]/g, ""), marks: [{ type: mark }] },
      ])
    }
  })

  it("keeps what surrounds a marked phrase", () => {
    expect(inlineFrom("get **rye** flour")).toEqual([
      { type: "text", text: "get " },
      { type: "text", text: "rye", marks: [{ type: "bold" }] },
      { type: "text", text: " flour" },
    ])
  })

  it("prefers the longer delimiter", () => {
    // Otherwise `**bold**` reads as an italic star, which is not what anybody wrote.
    expect(inlineFrom("**loud**")).toEqual([
      { type: "text", text: "loud", marks: [{ type: "bold" }] },
    ])
  })

  it("leaves a star that never closes alone", () => {
    // Somebody typed a star. It is not the start of anything, and must not swallow the
    // rest of the line.
    expect(inlineFrom("2 * 3 items")).toEqual([{ type: "text", text: "2 * 3 items" }])
  })

  it("treats markup inside code as text, which is the point of code", () => {
    expect(inlineFrom("`**not bold**`")).toEqual([
      { type: "text", text: "**not bold**", marks: [{ type: "code" }] },
    ])
  })

  it("reads a link, address and all", () => {
    expect(inlineFrom("see [the docs](https://nooks.test/docs)")).toEqual([
      { type: "text", text: "see " },
      {
        type: "text",
        text: "the docs",
        marks: [{ type: "link", attrs: { href: "https://nooks.test/docs" } }],
      },
    ])
  })

  it("reads a marked phrase inside a link", () => {
    expect(inlineFrom("[**loud**](https://nooks.test)")).toEqual([
      {
        type: "text",
        text: "loud",
        marks: [{ type: "link", attrs: { href: "https://nooks.test" } }, { type: "bold" }],
      },
    ])
  })

  it("leaves brackets that are not a link alone", () => {
    expect(inlineFrom("milk [the good one]")).toEqual([
      { type: "text", text: "milk [the good one]" },
    ])
  })

  it("nests one mark inside another", () => {
    expect(inlineFrom("**~~both~~**")).toEqual([
      { type: "text", text: "both", marks: [{ type: "bold" }, { type: "strike" }] },
    ])
  })
})

describe("writing it back", () => {
  it("returns every line it can read", () => {
    for (const line of [
      "Rye flour",
      "**loud**",
      "*soft*",
      "~~gone~~",
      "`code`",
      "get **rye** flour",
      "2 * 3 items",
      "`**not bold**`",
      "see [the docs](https://nooks.test/docs)",
      "[**loud**](https://nooks.test)",
      "milk [the good one]",
    ]) {
      expect(roundTrip(line)).toBe(line)
    }
  })

  it("writes a phrase carrying two marks the same way every time", () => {
    // The order marks were applied in is not the order they are written in, or the
    // same Note would save differently depending on which button was pressed first.
    const applied = [{ type: "text", text: "both", marks: [{ type: "strike" }, { type: "bold" }] }]
    const reversed = [{ type: "text", text: "both", marks: [{ type: "bold" }, { type: "strike" }] }]

    expect(inlineTo(applied)).toBe(inlineTo(reversed))
  })

  it("drops an empty run rather than writing bare punctuation", () => {
    expect(inlineTo([{ type: "text", text: "", marks: [{ type: "bold" }] }])).toBe("")
  })
})
