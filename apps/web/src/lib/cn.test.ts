import { readFileSync } from "node:fs"
import { fileURLToPath } from "node:url"
import { describe, expect, it } from "vitest"

import { cn, NOOKS_COLORS, NOOKS_CONTAINERS, NOOKS_SPACING, NOOKS_TEXT_SIZES } from "./cn"

// Read rather than imported: Vitest stubs CSS imports, so `?raw` would hand back an
// empty string and the check would pass without checking anything.
/*
 * The foundations, not the app's own stylesheet.
 *
 * The tokens live in @nooks/tokens because the website has to look like the app, and
 * two copies of a palette are two palettes. This test follows them there: what it is
 * guarding is that every token the scale declares is one `cn` knows how to merge.
 */
const stylesheet = readFileSync(
  fileURLToPath(new URL("../../../../packages/tokens/foundations.css", import.meta.url)),
  "utf8",
)

/** tokensNamed lists every token declared under a namespace in the stylesheet. */
function tokensNamed(namespace: string): string[] {
  const pattern = new RegExp(`^\\s*--${namespace}-([a-z0-9-]+):`, "gm")
  const found = new Set<string>()
  for (const match of stylesheet.matchAll(pattern)) {
    const name = match[1] ?? ""
    // A modifier such as --text-display--line-height belongs to its token, not beside it.
    if (!name.includes("--")) {
      found.add(name)
    }
  }
  return [...found]
}

// The lists in cn.ts are what stops a size and a colour of the same name being read as
// one utility, so a token added to the stylesheet has to reach them.
describe("the tokens cn is told about", () => {
  it("covers every step of the type scale", () => {
    expect([...NOOKS_TEXT_SIZES].sort()).toEqual(tokensNamed("text").sort())
  })

  it("covers every width the layout is measured in", () => {
    expect([...NOOKS_CONTAINERS].sort()).toEqual(tokensNamed("container").sort())
  })

  it("covers every height a control is measured in", () => {
    expect([...NOOKS_SPACING].sort()).toEqual(tokensNamed("spacing").sort())
  })

  it("covers every colour that shadcn does not already name", () => {
    const shadcnColors = tokensNamed("color").filter((name) => !NOOKS_COLORS.includes(name))
    // Whatever is left must be a name cn already knows, or merging it would drop it.
    expect(cn(`text-small text-${shadcnColors[0] ?? "foreground"}`)).toContain("text-small")
  })
})

describe("merging Nooks' own classes", () => {
  it("keeps a size and a colour that are both spelled text-", () => {
    expect(cn("text-small text-shared")).toBe("text-small text-shared")
  })

  it("still lets a later size win over an earlier one", () => {
    expect(cn("text-small text-body")).toBe("text-body")
  })

  it("still lets a later colour win over an earlier one", () => {
    expect(cn("text-shared text-overdue")).toBe("text-overdue")
  })

  // A generated shadcn component sets its own width; a Nooks one overrides it.
  it("lets a Nooks width win over the one a shadcn component sets", () => {
    expect(cn("max-w-[calc(100%-2rem)] sm:max-w-sm", "max-w-dialog sm:max-w-dialog")).toBe(
      "max-w-dialog sm:max-w-dialog",
    )
  })
})
