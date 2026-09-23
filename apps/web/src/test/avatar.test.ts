import { globSync, readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

/** GENERATED is upstream's and is not ours to hold to this. */
const GENERATED = "src/components/ui/"

/** The one component allowed to draw the chip. */
const AVATAR = "src/components/ds/avatar.tsx"

/**
 * The avatar chip is one component, not six spans.
 *
 * It had been written out by hand in six places at three sizes, four of them in the UI
 * font and two in mono, which is what a shared thing looks like when there is nowhere
 * shared to put it. Nothing was broken; the sizes had simply drifted apart one call at
 * a time, and drifting again costs one more hand-written span.
 *
 * `bg-chip` is the fill DESIGN.md §2 gives it, and it appears nowhere else in the app,
 * which is what makes it the thing to check for: a chip drawn by hand has to reach for
 * that token to look like a chip at all.
 */
describe("the avatar chip", () => {
  const files = ours()

  it("is reading the app", () => {
    expect(files.length).toBeGreaterThan(80)
  })

  it("is drawn in one place", () => {
    const drawing = files
      .filter((file) => file !== AVATAR)
      .filter((file) => readFileSync(file, "utf8").includes("bg-chip"))

    expect(drawing, `use <Avatar> from ${AVATAR} rather than writing the chip again`).toEqual([])
  })

  // A check that reads no file agrees with everything, and so does one whose subject
  // has been renamed out from under it.
  it("is checking for something that is there", () => {
    expect(readFileSync(AVATAR, "utf8")).toContain("bg-chip")
  })
})

/** ours is the app's own source: the generated components are upstream's. */
function ours(): string[] {
  return globSync("src/**/*.{ts,tsx}")
    .map((file) => String(file))
    .filter((file) => !file.includes(".test."))
    .filter((file) => !file.replaceAll("\\", "/").includes(GENERATED))
}
