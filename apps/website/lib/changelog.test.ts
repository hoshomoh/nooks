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
