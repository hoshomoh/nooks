import { globSync, readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

/**
 * The split in vite.config.ts is by file extension: .test.tsx renders and gets jsdom
 * with the cleanup in ./dom.ts, .test.ts does not and gets node.
 *
 * That is a rule nothing enforces at the type level, and getting it wrong is quiet: a
 * .test.ts that renders finds no document and fails with something that reads like a
 * broken component. This is the enforcement.
 */
describe("which project a test file lands in", () => {
  const unit = globSync("src/**/*.test.ts", { cwd: process.cwd() })

  it("finds the unit tests", () => {
    expect(unit.length).toBeGreaterThan(20)
  })

  it("keeps rendering out of the tests that run without a DOM", () => {
    const rendering = unit.filter((file) =>
      readFileSync(file, "utf8").includes("@testing-library/react"),
    )
    expect(rendering, "rename these to .test.tsx").toEqual([])
  })
})
