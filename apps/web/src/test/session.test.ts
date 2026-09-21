import { globSync, readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

/** OWNER is the one file allowed to end a session, and HERE is this one. */
const OWNER = "src/lib/new-session.ts"
const HERE = "src/test/session.test.ts"

/**
 * Ending a session has two halves and only one of them is obvious.
 *
 * `queryClient.clear()` empties the tab. The cache is also written to localStorage so
 * that reading offline survives a reload, and that copy holds whatever the Member was
 * shown: their lists, their items, their notes, and everybody's names and addresses.
 * Clearing one without the other leaves all of it on the machine for a day.
 *
 * Signing out did exactly that. The debounced writer would usually overwrite the stored
 * copy a second later, which is not a guarantee — closing the lid straight after
 * signing out is the ordinary way to hand a tablet over.
 *
 * So `startNewSession` does both, and this is what keeps the halves together.
 */
describe("who may end a session", () => {
  // Filtered after the glob, not by its exclude callback: that is given a file's name
  // rather than its path, so a predicate written against a path silently matches
  // nothing and the scan quietly reads what it meant to skip.
  const sources = globSync("src/**/*.{ts,tsx}")
    .map((file) => String(file))
    .filter((file) => !file.includes(".test."))

  // A move or a rename that left the names above pointing at nothing would turn this
  // into furniture that passes because it is looking at an empty list.
  it("is looking at the app, and can see the one file that may", () => {
    expect(sources).toContain(OWNER)
    expect(sources.length).toBeGreaterThan(50)
    expect(globSync(`${HERE}`)).toEqual([HERE])
  })

  it("keeps clearing the tab and clearing the disk in the same place", () => {
    const clearing = sources
      .filter((file) => file !== OWNER)
      .filter((file) => readFileSync(file, "utf8").includes("queryClient.clear()"))

    expect(clearing, `call startNewSession from ${OWNER} instead`).toEqual([])
  })
})
