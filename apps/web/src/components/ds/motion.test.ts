import { readFileSync, readdirSync } from "node:fs"
import { fileURLToPath } from "node:url"
import { describe, expect, it } from "vitest"

/*
Every movement in the app is one the design named.

DESIGN.md §11 lists what moves and for how long, and the foundations declare each of
them once. A class naming an animation that was never declared is not an error anywhere:
Tailwind emits nothing, the browser has nothing to run, and the component simply does
not move — which looks exactly like a component that was never meant to. That has cost
this repo a whole afternoon once already, with a font size that silently did nothing.

So the two halves are compared here. Anything reaching for `animate-x` has to find an
`--animate-x` waiting for it, and every `--animate-x` has to be reached for by somebody.
*/

const WEB = fileURLToPath(new URL("../..", import.meta.url))
const FOUNDATIONS = fileURLToPath(
  new URL("../../../../../packages/design/foundations.css", import.meta.url),
)

/** Every source file in the web app, as text. */
function sources(dir: string): string[] {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = `${dir}/${entry.name}`
    if (entry.isDirectory()) {
      return sources(path)
    }
    return /\.tsx?$/.test(entry.name) && !/\.test\.tsx?$/.test(entry.name)
      ? [readFileSync(path, "utf8")]
      : []
  })
}

/**
 * The two the app does not declare because it did not write them.
 *
 * `animate-in` and `animate-out` come from tw-animate-css, which the foundations
 * import for the generated shadcn overlays — the one family of components in here that
 * nobody edits by hand.
 */
const FROM_LIBRARY = ["in", "out"]

/** The animations the foundations declare. */
function declared(): string[] {
  const css = readFileSync(FOUNDATIONS, "utf8")
  return [...css.matchAll(/--animate-([a-z-]+):/g)].map((match) => match[1]).sort()
}

/** The animations the app asks for. */
function used(): string[] {
  const asked = new Set<string>()
  for (const source of sources(WEB)) {
    for (const match of source.matchAll(/\banimate-([a-z-]+)\b/g)) {
      asked.add(match[1])
    }
  }
  return [...asked].filter((name) => !FROM_LIBRARY.includes(name)).sort()
}

describe("the motion vocabulary", () => {
  it("declares everything the app reaches for", () => {
    expect(used().filter((name) => !declared().includes(name))).toEqual([])
  })

  it("has nothing declared that nothing uses", () => {
    // A token the app stopped using is a decision nobody made. Either something should
    // be moving and is not, or the design has one fewer movement than it says.
    expect(declared().filter((name) => !used().includes(name))).toEqual([])
  })

  it("gives a Member who asked for less movement none of it", () => {
    // The only reason nothing in here handles reduced motion itself.
    expect(readFileSync(FOUNDATIONS, "utf8")).toMatch(
      /@media \(prefers-reduced-motion: reduce\)/,
    )
  })
})
