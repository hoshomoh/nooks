import { globSync, readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

/** HERE is this file, which names the import below in order to look for it. */
const HERE = "src/test/projects.test.ts"

/**
 * The split in vite.config.ts is by file extension: .test.tsx renders and gets jsdom
 * with the cleanup in ./dom.ts, .test.ts does not and gets node.
 *
 * That is a rule nothing enforces at the type level, and getting it wrong is quiet: a
 * .test.ts that renders finds no document and fails with something that reads like a
 * broken component. This is the enforcement.
 */
describe("which project a test file lands in", () => {
  const found = globSync("src/**/*.test.ts")

  // A rename that left the exclusion below pointing at nothing would quietly stop this
  // file from checking itself out of the list, which is how a guard becomes furniture.
  it("is looking at the unit tests, itself included", () => {
    expect(found).toContain(HERE)
    expect(found.length).toBeGreaterThan(20)
  })

  it("keeps rendering out of the tests that run without a DOM", () => {
    const rendering = found
      .filter((file) => file !== HERE)
      .filter((file) => readFileSync(file, "utf8").includes("@testing-library/react"))

    expect(rendering, "rename these to .test.tsx").toEqual([])
  })
})
