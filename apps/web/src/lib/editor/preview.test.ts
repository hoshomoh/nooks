import { describe, expect, it } from "vitest"

import { previewOf } from "./preview"

/** said is the words of a preview, with the marks dropped. */
function said(markdown: string): string {
  return previewOf(markdown)
    .runs.map((run) => run.text)
    .join("")
}

describe("what a row shows of a Note", () => {
  it("shows the words, never the shorthand", () => {
    // Every one of these came back with its markers still on when the row was built by
    // taking prefixes off the markdown.
    expect(said("### Where\nSaturday market")).toBe("Where")
    expect(said("- [ ] tomatoes")).toBe("tomatoes")
    expect(said("> they pack up around two")).toBe("they pack up around two")
  })

  it("leaves a bullet alone, because a Note has no bullet lists", () => {
    // DESIGN.md §10 names five blocks and a bullet list is not one of them, so "- milk"
    // is a line that starts with a dash rather than a list nobody can make.
    expect(said("- milk")).toBe("- milk")
  })

  it("shows nothing for a checklist item with no words on it", () => {
    // This is what put "[ ]" on a row: a marker with nothing after it, stripped by
    // guesswork down to the box it was written with.
    expect(said("- [ ]")).toBe("")
    expect(said("- [ ] ")).toBe("")
  })

  it("skips past an empty item to the first one that says something", () => {
    expect(said("- [ ]\n- [ ] tomatoes")).toBe("tomatoes")
  })

  it("does not count the empty ones toward what is left", () => {
    expect(previewOf("- [ ] tomatoes\n- [ ]\n- [ ]").remaining).toBe(0)
  })

  it("counts the blocks that do say something", () => {
    expect(previewOf("### Where\nSaturday market\n\n> around two").remaining).toBe(2)
  })

  it("keeps inline markup as marks rather than as punctuation", () => {
    // The row renders these; it does not print the asterisks the Member never typed
    // as asterisks.
    const [run] = previewOf("**loud**").runs

    expect(run?.text).toBe("loud")
    expect(run?.marks).toContain("bold")
  })

  it("keeps a link's address, so the row can be one", () => {
    const [run] = previewOf("[the docs](https://nooks.test)").runs

    expect(run?.text).toBe("the docs")
    expect(run?.href).toBe("https://nooks.test")
  })

  it("says nothing about a Note that is not there", () => {
    expect(previewOf("")).toEqual({ runs: [], remaining: 0 })
    expect(previewOf("   \n\n  ")).toEqual({ runs: [], remaining: 0 })
  })
})
