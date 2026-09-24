import { describe, expect, it } from "vitest"

import { documentFrom, markdownFrom } from "./markdown"

/** typed is a Note holding one paragraph of exactly these characters. */
function typed(text: string) {
  return { type: "doc", content: [{ type: "paragraph", content: [{ type: "text", text }] }] }
}

/** saved writes a Note out and reads it back, which is what closing and reopening does. */
function saved(text: string): string {
  const back = documentFrom(markdownFrom(typed(text)))
  return JSON.stringify(back)
}

/**
 * Punctuation somebody typed is still punctuation when they come back.
 *
 * A Note is written out as markdown and read back in, so text that looks like markup is
 * markup the second time unless something says otherwise. `say *this* aloud` came back
 * with the stars gone and the middle word in italic: not a rendering quirk, the
 * characters were no longer in the document. It is the Member's own words, and it is the
 * one kind of loss nothing else in the app can undo.
 *
 * Read as text rather than as markdown, because what matters is what the document holds
 * rather than how it was spelled on the way there.
 */
describe("punctuation somebody typed", () => {
  it("does not become emphasis", () => {
    expect(saved("say *this* aloud")).toContain("say *this* aloud")
    expect(saved("say *this* aloud")).not.toContain("italic")
  })

  it("does not become bold", () => {
    expect(saved("**not** shouting")).toContain("**not** shouting")
    expect(saved("**not** shouting")).not.toContain("bold")
  })

  it("does not become a code span", () => {
    expect(saved("use `ls` there")).toContain("use `ls` there")
    expect(saved("use `ls` there")).not.toContain("code")
  })

  it("does not become a strike", () => {
    expect(saved("a ~~b~~ c")).toContain("a ~~b~~ c")
    expect(saved("a ~~b~~ c")).not.toContain("strike")
  })

  it("does not become a link", () => {
    expect(saved("see [here](there) for it")).toContain("see [here](there) for it")
    expect(saved("see [here](there) for it")).not.toContain("link")
  })

  // Without this, protecting the punctuation above would eat the next turn's backslash.
  it("keeps a backslash that was typed", () => {
    expect(saved("C:\\Users")).toContain("C:\\\\Users")
  })

  /*
   * A delimiter on its own is a character, and was never at risk. Left unprotected so
   * that a Note read over the API or by an assistant is the prose somebody wrote rather
   * than prose with backslashes through it.
   */
  it("is left alone when nothing could close it", () => {
    expect(markdownFrom(typed("2 * 3 items"))).toBe("2 * 3 items")
    expect(markdownFrom(typed("a ~ b"))).toBe("a ~ b")
  })
})

/** The markup that is meant to be markup still is, which is the other half. */
describe("markup that was meant", () => {
  it("still reads as emphasis", () => {
    expect(JSON.stringify(documentFrom("say *this* aloud"))).toContain("italic")
  })

  it("still reads as a link", () => {
    const back = JSON.stringify(documentFrom("see [here](https://nooks.example) for it"))
    expect(back).toContain("link")
    expect(back).toContain("https://nooks.example")
  })

  it("still survives being written and read again", () => {
    const once = markdownFrom(documentFrom("**loud** and *quiet*"))
    expect(markdownFrom(documentFrom(once))).toBe(once)
  })
})
