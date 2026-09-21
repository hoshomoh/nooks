import { globSync, readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

const CATALOGUE = "src/i18n/locales/en.json"

/**
 * Every key the app asks for is one the catalogue has.
 *
 * i18next answers a key it does not know with the key itself, so a missing one is not
 * an error anywhere: it renders, in the middle of the screen, as `list.addToList`. The
 * app compiles, the tests pass, and the first person to find out is whoever opened that
 * screen.
 *
 * Keys are read from inside `t(...)` rather than from any key-shaped string, because
 * plenty of strings are that shape without being keys — the live event kinds are
 * `list.changed` and `lists.changed`.
 */
describe("the strings the app asks for", () => {
  const catalogue = flatten(JSON.parse(readFileSync(CATALOGUE, "utf8")) as Catalogue)
  const asked = keysAskedFor()

  // A regex that stopped matching would leave this passing over nothing.
  it("is reading the app and the catalogue", () => {
    expect(catalogue.size).toBeGreaterThan(400)
    expect(asked.size).toBeGreaterThan(300)
  })

  it("has every one of them", () => {
    const missing = [...asked].filter((key) => !catalogue.has(key)).sort()
    expect(missing, `add these to ${CATALOGUE} or they render as themselves`).toEqual([])
  })
})

type Catalogue = { [key: string]: string | string[] | Catalogue }

/**
 * flatten names every leaf, keeping arrays whole.
 *
 * An array is read in one go with returnObjects — the add row's list of words for
 * "tomorrow" — so the array is the key and its items are not.
 */
function flatten(node: Catalogue, prefix = ""): Set<string> {
  const keys = new Set<string>()
  for (const [name, value] of Object.entries(node)) {
    if (Array.isArray(value) || typeof value !== "object") {
      keys.add(prefix + name)
      continue
    }
    for (const nested of flatten(value, `${prefix}${name}.`)) {
      keys.add(nested)
    }
  }
  return keys
}

/**
 * keysAskedFor is every key passed to t() as a literal.
 *
 * The whole call is read rather than only its first argument, so a key chosen inside it
 * — `t(saved ? "account.saved" : "account.save")` — is found too. A key built from a
 * variable cannot be checked this way and is left alone; the count assertion above is
 * what notices if that becomes most of them.
 */
function keysAskedFor(): Set<string> {
  const asked = new Set<string>()
  // Filtered here rather than by the glob's exclude callback, which is given a file's
  // name and not its path — a predicate written against a path matches nothing and says
  // nothing about it.
  const ours = globSync("src/**/*.{ts,tsx}")
    .map((file) => String(file))
    .filter((file) => !file.includes(".test."))

  for (const file of ours) {
    const source = readFileSync(file, "utf8")
    for (const call of source.matchAll(/\bt\(([^()]*)\)/g)) {
      for (const quoted of call[1].matchAll(/"([a-z][A-Za-z0-9]*(?:\.[A-Za-z0-9_]+)+)"/g)) {
        asked.add(quoted[1])
      }
    }
  }
  return asked
}
