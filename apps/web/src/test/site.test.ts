import { readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"
import { SITE, SOURCE } from "@nooks/shared"

const README = "../../README.md"

/**
 * The README agrees with the app about where the documentation is.
 *
 * `SITE` says it once and the app imports it. The README cannot import anything, so it
 * writes the address out, and a second copy of a fact is a fact that drifts — which is
 * what `site.ts` says of itself: "A site that disagrees with its own repository about
 * where it is has told somebody the wrong thing."
 *
 * The person who finds a moved address through a stale README is somebody who could not
 * get something working and went looking for help.
 *
 * Checked from here rather than beside `site.ts`, because that package compiles for the
 * app and for the site and has no node types on purpose; a test that reads a file does
 * not belong in it.
 */
describe("where the project says it lives", () => {
  it("is one address, in the README and in the app", () => {
    const readme = readFileSync(README, "utf8")
    expect(readme, `README points somewhere other than ${SITE}`).toContain(SITE)
  })

  // Both constants satisfying "is in the README" by being empty would be a test that
  // passes over nothing.
  it("is an address at all", () => {
    for (const [name, url] of Object.entries({ SITE, SOURCE })) {
      expect(url, `${name} is not a URL`).toMatch(/^https:\/\/[^/]+/)
    }
  })
})
