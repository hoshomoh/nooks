import { readFileSync, readdirSync } from "node:fs"
import { fileURLToPath } from "node:url"
import { describe, expect, it } from "vitest"

/*
One focus treatment, on everything focusable.

DESIGN.md §7: a 2px `--shared` outline at 2px offset, everywhere. It is applied once, to
`:focus-visible` in the foundations, so nothing has to remember it — which means the only
way to lose it is for a component to turn it off.

Turning it off is sometimes right. What is never right is turning it off without saying
why, so a component that does has to appear below with a reason. Somebody adding
`outline-none` to make a field look tidier finds out here that they have taken the
keyboard's only indication of where it is.
*/

const DS = fileURLToPath(new URL(".", import.meta.url))

/** Where the outline is deliberately suppressed, and what stands in for it. */
const ALLOWED: Record<string, string> = {
  "add-row.tsx":
    "The row around the field takes the ring with focus-within, because the sentence " +
    "and its chips are drawn as one field. The input's own outline would be a second " +
    "ring inside the first.",
  "note-editor.tsx":
    "The Note body is a document, not a control. Its focus indicator is the caret, and " +
    "an outline around a page of prose would be drawn around the whole page.",
}

/** suppressors lists the components that turn the outline off. */
function suppressors(): string[] {
  return readdirSync(DS)
    .filter((name) => name.endsWith(".tsx") && !name.endsWith(".test.tsx"))
    .filter((name) => /outline-none|outline-hidden/.test(withoutComments(readFileSync(DS + name, "utf8"))))
    .sort()
}

/**
 * withoutComments is the code, without what is written about it.
 *
 * A comment explaining why the outline is kept mentions the class that would remove it,
 * and a scanner that cannot tell those apart reports the file that got it right.
 */
function withoutComments(source: string): string {
  return source.replace(/\/\*[\s\S]*?\*\//g, "").replace(/\/\/[^\n]*/g, "")
}

describe("where the focus ring is turned off", () => {
  it("is only where there is a reason written down", () => {
    const unexplained = suppressors().filter((name) => !(name in ALLOWED))

    expect(unexplained, "these suppress the focus outline with no reason given").toEqual([])
  })

  it("does not list a component that no longer does it", () => {
    // A reason for something that stopped happening is a reason nobody can check.
    const stale = Object.keys(ALLOWED).filter((name) => !suppressors().includes(name))

    expect(stale, "these are excused from something they no longer do").toEqual([])
  })
})
