import { readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

const CATALOGUE = "src/i18n/locales/en.json"

/**
 * One verb for creating things, which DESIGN.md §13 names and nothing held.
 *
 * "Add a member · Add a group · Add a token · Add a list. Never New, Create, Issue, or
 * Generate." It is about what a Member reads, which is what §13 is: it sits beside
 * "sentence case" and "never summarise the Member". Go constructors are ordinary Go and
 * this has nothing to say about them.
 *
 * Unheld, it had drifted twice. The token button said "Create the token" directly
 * against §13's own example of "Add a token", and it is fixed. The other is below.
 *
 * "New" as an adjective is left alone: "New password" describes a password rather than
 * asking for one to be made, and "New here? Ask to join." is not about creating at all.
 * What is caught is the verb at the front of something a Member reads.
 */
describe("the verb for creating things", () => {
  /*
   * The first-run button, still to be decided rather than quietly exempt.
   *
   * §13 forbids "Create" and names no replacement for this one, and the instance is not
   * added to anything the way a member or a list is — "Add the instance" reads wrong for
   * the screen that brings it into being. So the copy is somebody's call, and until it
   * is made this names the string so the decision cannot be lost. Whoever makes it
   * deletes this entry.
   */
  const undecided = new Set(["auth.setup.submit", "auth.setup.submitting"])

  const catalogue = flatten(JSON.parse(readFileSync(CATALOGUE, "utf8")) as Catalogue)

  // A floor, so a catalogue that stopped being read does not agree with everything.
  it("is reading the catalogue", () => {
    expect(catalogue.size).toBeGreaterThan(400)
  })

  it("is Add, in everything a Member reads", () => {
    const wrong = [...catalogue]
      .filter(([key]) => !undecided.has(key))
      .filter(([, said]) => opensWithCreating(said))
      .map(([key, said]) => `${key} = "${said}"`)

    expect(wrong, "DESIGN.md §13: one verb for creating things, and it is Add").toEqual([])
  })

  // The exception list is part of the check: an entry left behind after the string was
  // changed would quietly stop holding that string.
  it("excuses only strings that are still wrong", () => {
    for (const key of undecided) {
      const said = catalogue.get(key)
      expect(said, `${key} is excused and is not in the catalogue`).toBeDefined()
      expect(
        said !== undefined && opensWithCreating(said),
        `${key} = "${said}" no longer breaks the rule, so take it off the list`,
      ).toBe(true)
    }
  })
})

/**
 * opensWithCreating reports whether a string leads with a verb §13 rules out.
 *
 * The participle counts. "Creating…" is what the button says while it works, and it is
 * as much a word a Member reads as the one before it; leaving it out would let half a
 * button be changed and the other half go on saying the old verb.
 */
function opensWithCreating(said: string): boolean {
  return (
    /^(Create|Creating|Issue|Issuing|Generate|Generating)\b/.test(said) ||
    /^New (a|an|the) /.test(said)
  )
}

type Catalogue = { [key: string]: string | string[] | Catalogue }

/** flatten is the catalogue as dotted keys, the way the app asks for them. */
function flatten(node: Catalogue, path: string[] = [], into = new Map<string, string>()) {
  for (const [key, value] of Object.entries(node)) {
    const at = [...path, key]
    if (typeof value === "string") {
      into.set(at.join("."), value)
    } else if (!Array.isArray(value)) {
      flatten(value, at, into)
    }
  }
  return into
}
