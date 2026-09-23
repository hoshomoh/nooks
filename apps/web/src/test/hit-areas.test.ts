import { readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

const DESIGN = "../../DESIGN.md"
const TOKENS = "../../packages/design/foundations.css"
const ICON_BUTTON = "src/components/ds/icon-button.tsx"
const CHECKBOX = "src/components/ds/checkbox.tsx"

/**
 * The hit areas DESIGN.md §4 states are the hit areas the app draws.
 *
 * This is where the rule and the code had already come apart once. STANDARDS said "44px
 * hit areas" flatly while icon-only buttons were 22px squares, and the sentence was
 * short rather than the code being wrong: a 44px target on a 31px row reaches into the
 * row above and takes its taps. §4 now says what is actually true, and this holds it
 * true, because a number written in prose beside a number written in CSS is two numbers
 * that drift.
 */
describe("hit areas", () => {
  const design = readFileSync(DESIGN, "utf8")
  const tokens = readFileSync(TOKENS, "utf8")

  // A regex that matched nothing would agree with every number there is.
  it("is reading the rule", () => {
    expect(design).toContain("as easy to hit as the row is tall")
  })

  it("are the numbers the tokens carry", () => {
    expect(pxOf(tokens, "--spacing-row")).toBe(saidIn(design, "The list row is (\\d+)px"))
    expect(pxOf(tokens, "--spacing-control-compact")).toBe(
      saidIn(design, "takes (\\d+)px, the height of a settings row"),
    )
  })

  // Prose agreeing with a token nothing reads is prose agreeing with itself.
  it("are the tokens the controls read", () => {
    expect(readFileSync(ICON_BUTTON, "utf8")).toContain("after:h-control-compact")
    expect(readFileSync(CHECKBOX, "utf8")).toContain("h-row")
  })
})

/** pxOf is the pixel value of one token, as the stylesheet declares it. */
function pxOf(css: string, token: string): number {
  const found = css.match(new RegExp(`${token}:\\s*(\\d+)px`))
  expect(found, `${token} is not declared in ${TOKENS}`).not.toBeNull()
  return Number(found?.[1])
}

/** saidIn is the number a sentence in the design states. */
function saidIn(prose: string, pattern: string): number {
  const found = prose.match(new RegExp(pattern))
  expect(found, `${DESIGN} no longer says "${pattern}"`).not.toBeNull()
  return Number(found?.[1])
}
