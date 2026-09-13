import { readFileSync } from "node:fs"
import { fileURLToPath } from "node:url"
import { describe, expect, it } from "vitest"

/*
The checklist row, held to what DESIGN.md §10 draws.

This has been reported twice. Both times the cause was the same: ProseMirror renders a
paragraph inside every checklist item, so the paragraph rule lands on it, pads it, and
pushes the words down past the box that was meant to sit beside them. The box looks
high; the row looks broken; nothing in the checklist rules is wrong.

A stylesheet cannot be laid out here, so what is held is the rule that stops it — and
the two numbers the design names.
*/
const CSS = readFileSync(
  fileURLToPath(new URL("../index.css", import.meta.url)),
  "utf8",
)

/** ruleFor is the body of the first rule with that exact selector. */
function ruleFor(selector: string): string {
  const at = CSS.indexOf(`${selector} {`)
  if (at < 0) {
    throw new Error(`no rule for ${selector}`)
  }
  return CSS.slice(at, CSS.indexOf("}", at))
}

const ITEM = '.nooks-note ul[data-type="taskList"] li'

describe("a checklist row", () => {
  it("keeps the paragraph rule off the words beside the box", () => {
    // Without this the words sit 5px below the box and the row reads as misaligned.
    expect(ruleFor(`${ITEM} > div > p`)).toMatch(/padding:\s*0/)
  })

  it("still pads the rows the way the design says", () => {
    expect(ruleFor(ITEM)).toMatch(/padding:\s*3px 0/)
  })

  it("puts the box in the gutter the design names", () => {
    // DESIGN.md §10: a 15px box in a 26px gutter.
    expect(ruleFor(`${ITEM} > label`)).toMatch(/width:\s*26px/)
    expect(ruleFor('.nooks-note input[type="checkbox"]')).toMatch(/width:\s*15px/)
  })

  it("centres the box on the first line, at any type size", () => {
    // 1lh rather than a fixed nudge: a nudge is right at one font size and wrong at
    // the other, and the Note is set at two.
    const label = ruleFor(`${ITEM} > label`)
    expect(label).toMatch(/height:\s*1lh/)
    expect(label).toMatch(/align-items:\s*center/)
  })

  it("does not reintroduce a gap beside the gutter", () => {
    // A gutter plus a gap is a gutter of some other width than the one named.
    expect(ruleFor(ITEM)).not.toMatch(/gap:/)
  })
})
