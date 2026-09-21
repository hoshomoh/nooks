import { globSync, readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

/** GENERATED is upstream's and is not ours to hold to this. */
const GENERATED = "src/components/ui/"

/**
 * An effect that is the only way to do something, and the reason it is.
 *
 * Empty, and that is the point: the app has none. An entry here is a decision somebody
 * made and wrote down, not a default somebody reached for.
 */
const UNAVOIDABLE: Record<string, string> = {}

/**
 * The app synchronises with React rather than around it.
 *
 * STANDARDS §5: useEffect is a last resort. Most of what it gets used for has a better
 * answer — derived state is computed during render, a subscription outside React is
 * read with useSyncExternalStore, and something that should happen when a Member does
 * something belongs in the handler for the thing they did.
 *
 * The app currently has none at all, which is worth keeping: effects arrive one at a
 * time, each with a reason, and the fifth one is what makes a screen re-render twice
 * and fetch what it already had.
 */
describe("effects", () => {
  const files = ours()

  it("is reading the app", () => {
    expect(files.length).toBeGreaterThan(80)
  })

  it("are not how the app does things", () => {
    const reaching = files
      .filter((file) => /\buse(Layout)?Effect\s*\(/.test(readFileSync(file, "utf8")))
      .filter((file) => !(file in UNAVOIDABLE))

    expect(reaching, "see STANDARDS §5 — or add it to UNAVOIDABLE with the reason").toEqual([])
  })

  // An entry left behind after the effect went would quietly excuse the next one.
  it("excuses nothing that is not there", () => {
    for (const [file, why] of Object.entries(UNAVOIDABLE)) {
      expect(files, `${file} is listed and is not here`).toContain(file)
      expect(why.length, `say why ${file} needs one`).toBeGreaterThan(20)
    }
  })
})

/**
 * ours is the app's own source: the generated components are upstream's.
 *
 * Filtered after the glob rather than inside it. The exclude callback is given a path
 * whose shape is the glob's business, and a predicate that silently matches nothing
 * would leave this test reading files it meant to skip — which is how it was written
 * the first time.
 */
function ours(): string[] {
  return globSync("src/**/*.{ts,tsx}")
    .map((file) => String(file))
    .filter((file) => !file.includes(".test."))
    .filter((file) => !file.replaceAll("\\", "/").includes(GENERATED))
}
