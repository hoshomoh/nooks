import { describe, expect, it } from "vitest"

import { readReleases } from "./changelog"

/*
The page is only as good as what it reads.

release-please writes this file, so its shape is not ours to choose — it arrives with
commit links appended, scopes in bold, and breaking changes under a heading with a
warning sign in it. Every one of those has to survive the trip to a page that draws its
own hierarchy.
*/
const CHANGELOG = `# Changelog

## [0.2.0](https://github.com/hoshomoh/nooks/compare/v0.1.0...v0.2.0) (2026-09-12)

### ⚠ BREAKING CHANGES

* token hashes require one migration on first start

### New

* **web:** read how far away a day is ([50fdb21](https://github.com/hoshomoh/nooks/commit/50fdb21))

### Fixed

* stop skipping deploys ([661b495](https://github.com/hoshomoh/nooks/commit/661b495))

## [0.1.0](https://github.com/hoshomoh/nooks/compare/v0.0.0...v0.1.0) (2026-08-21)

### New

* the first one ([9d8bde5](https://github.com/hoshomoh/nooks/commit/9d8bde5))
`

describe("reading the changelog", () => {
  it("finds every release, newest first", () => {
    expect(readReleases(CHANGELOG).map((release) => release.version)).toEqual([
      "v0.2.0",
      "v0.1.0",
    ])
  })

  it("writes the date the way the page reads dates", () => {
    expect(readReleases(CHANGELOG)[0]?.date).toBe("12 September 2026")
  })

  it("marks a release that changes stored data", () => {
    // The one thing on the page somebody has to act on before upgrading.
    expect(readReleases(CHANGELOG)[0]?.migration).toBe(true)
    expect(readReleases(CHANGELOG)[1]?.migration).toBe(false)
  })

  it("keeps a breaking change as a line rather than only a badge", () => {
    expect(readReleases(CHANGELOG)[0]?.changes).toContainEqual({
      kind: "Changed",
      text: "token hashes require one migration on first start",
    })
  })

  it("strips the commit link and the bold scope", () => {
    // The page draws its own hierarchy; a second one arriving in the text fights it.
    expect(readReleases(CHANGELOG)[0]?.changes).toContainEqual({
      kind: "New",
      text: "web: read how far away a day is",
    })
  })

  it("says nothing rather than guessing when there is no file yet", () => {
    expect(readReleases("")).toEqual([])
  })
})

describe("how much the page carries", () => {
  /** A changelog with `count` releases in it, newest first. */
  function many(count: number): string {
    const entry = (n: number) =>
      [`## [1.${n}.0](https://x/compare) (2026-09-${String(n + 1).padStart(2, "0")})`, "", "### New", "", `* thing ${n}`, ""].join("\n")
    return ["# Changelog", ""].concat(
      Array.from({ length: count }, (_, i) => entry(count - i)),
    ).join("\n")
  }

  it("reads every release it is given", () => {
    // The page decides how many to show; the parser does not get to lose any, because
    // the count of what is left is what the box at the bottom is for.
    expect(readReleases(many(9))).toHaveLength(9)
  })

  it("keeps them newest first", () => {
    const versions = readReleases(many(3)).map((release) => release.version)
    expect(versions).toEqual(["v1.3.0", "v1.2.0", "v1.1.0"])
  })

  /*
   * A version reads as the git tag of the same release.
   *
   * The changelog page shows the first few lines of a big release and links the rest to
   * that release on GitHub, at `/releases/tag/<version>`. So the string this parser
   * builds is half of an address, and the other half is what release-please tagged.
   * Drop the `v` and every one of those links is a 404, with nothing failing here or at
   * build time to say so: `check-links.mjs` compares internal links against built pages
   * and never leaves the site.
   *
   * The release-please heading is written two ways in the real file, with a compare
   * link and without, and 1.0.0 is the one without. Both have to come out tagged.
   */
  it("reads as the tag the release was made under", () => {
    const written = [
      "# Changelog",
      "",
      "## [1.4.1](https://github.com/hoshomoh/nooks/compare/v1.4.0...v1.4.1) (2026-09-16)",
      "",
      "### Fixed",
      "",
      "* a thing",
      "",
      "## 1.0.0 (2026-09-13)",
      "",
      "### New",
      "",
      "* the first one",
      "",
    ].join("\n")

    const versions = readReleases(written).map((release) => release.version)
    expect(versions).toEqual(["v1.4.1", "v1.0.0"])
    for (const version of versions) {
      expect(version, `${version} is not shaped like a tag`).toMatch(/^v\d+\.\d+\.\d+$/)
    }
  })
})
