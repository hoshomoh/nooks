import { readFileSync } from "node:fs"
import { fileURLToPath } from "node:url"
import { describe, expect, it } from "vitest"

import { contrastRatio, parseOklch, toRgb, type Rgb } from "./contrast"

/*
Whether the palette can be read.

DESIGN.md picks colours in OKLCH, which is the right space to pick them in and the wrong
one to measure them in. These are the pairs that actually appear — text on the ground it
sits on — measured the way WCAG defines it, in both themes.

Read from @nooks/design rather than imported: Vitest stubs CSS imports, so `?raw` would
hand back an empty string and this would pass without checking anything.
*/
const STYLESHEET = readFileSync(
  fileURLToPath(new URL("../../../../packages/design/foundations.css", import.meta.url)),
  "utf8",
)

/** WCAG AA: body text. */
const BODY = 4.5

/** WCAG AA: large text, and the parts of a control that say where it is. */
const LARGE = 3

/**
 * Where each theme's tokens are declared, so a token can be read per theme.
 *
 * Split on the blocks themselves rather than on the word: `.dark` appears first in a
 * Tailwind variant declaration near the top of the file, and slicing there gives a
 * "light palette" containing no colours at all — which every pair then passes, having
 * measured nothing.
 */
const LIGHT_AT = STYLESHEET.indexOf(":root {")
const DARK_AT = STYLESHEET.indexOf(".dark {")

const THEMES = {
  light: STYLESHEET.slice(LIGHT_AT, DARK_AT),
  dark: STYLESHEET.slice(DARK_AT),
}

/** colour reads one token out of a theme and converts it. */
function colour(theme: keyof typeof THEMES, token: string): Rgb {
  const found = new RegExp(`--${token}:\\s*([^;]+);`).exec(THEMES[theme])
  if (!found?.[1]) {
    throw new Error(`no --${token} in the ${theme} palette`)
  }
  const parsed = parseOklch(found[1])
  if (!parsed) {
    throw new Error(`--${token} in ${theme} is not a colour: ${found[1]}`)
  }
  return toRgb(parsed)
}

/** Text, the ground it is read on, and what it has to clear. */
const PAIRS: [ink: string, ground: string, floor: number][] = [
  ["foreground", "background", BODY],
  ["foreground", "sidebar", BODY],
  ["secondary-foreground", "background", BODY],
  ["muted-foreground", "background", BODY],
  ["muted-foreground", "sidebar", BODY],
  ["shared", "background", BODY],
  ["destructive", "background", BODY],
  ["destructive", "destructive-bg", BODY],
  ["offline-text", "offline-bg", BODY],
  ["primary-foreground", "primary", BODY],
  // The ring says where the keyboard is, which is a control's state rather than prose.
  ["ring", "background", LARGE],
  ["done", "background", LARGE],
]

describe("the palette, measured the way it is read", () => {
  for (const theme of ["light", "dark"] as const) {
    describe(theme, () => {
      for (const [ink, ground, floor] of PAIRS) {
        it(`${ink} on ${ground} clears ${floor}:1`, () => {
          const ratio = contrastRatio(colour(theme, ink), colour(theme, ground))
          expect(Number(ratio.toFixed(2))).toBeGreaterThanOrEqual(floor)
        })
      }
    })
  }
})

describe("the palette this reads", () => {
  it("found both blocks, and they are not the same block", () => {
    // The whole check is worthless if the slices are wrong, and a wrong slice fails
    // silently by containing nothing to measure.
    expect(LIGHT_AT).toBeGreaterThan(-1)
    expect(DARK_AT).toBeGreaterThan(LIGHT_AT)
    expect(THEMES.light).toContain("--foreground:")
    expect(THEMES.dark).toContain("--foreground:")
    expect(THEMES.light).not.toContain(".dark {")
  })
})

describe("the conversion itself", () => {
  it("puts black and white where they belong", () => {
    // 21:1 is the whole range. If the maths is wrong this is the first thing to move.
    const white = toRgb({ l: 1, c: 0, h: 0 })
    const black = toRgb({ l: 0, c: 0, h: 0 })

    expect(contrastRatio(white, black)).toBeCloseTo(21, 1)
  })

  it("reads a token as written, percent or not", () => {
    expect(parseOklch("oklch(0.4836 0.0883 252.04)")).toEqual({ l: 0.4836, c: 0.0883, h: 252.04 })
    expect(parseOklch("oklch(48.36% 0.0883 252.04)")?.l).toBeCloseTo(0.4836, 4)
  })

  it("says so when a value is not a colour", () => {
    expect(parseOklch("44px")).toBeNull()
    expect(parseOklch("var(--ring)")).toBeNull()
  })
})
