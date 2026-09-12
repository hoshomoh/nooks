import { describe, expect, it } from "vitest"

import { starterItems } from "./starter-list"

/** t returns the key, so a test reads which string was asked for. */
const t = (key: string) => key

// Tuesday, 25 August 2026.
const now = new Date(2026, 7, 25)

describe("what a new Instance opens onto", () => {
  // A Member who has never seen Nooks cannot tell that an Item carries these unless
  // something on screen does.
  it("shows a quantity, a date and a Note, each at least once", () => {
    const items = starterItems(t, now)

    expect(items.some((item) => item.quantity !== "")).toBe(true)
    expect(items.some((item) => item.dueOn !== "")).toBe(true)
    expect(items.some((item) => item.note !== "")).toBe(true)
  })

  it("dates the one dated Item relative to the day the Instance was made", () => {
    const dated = starterItems(t, now).find((item) => item.dueOn !== "")
    expect(dated?.dueOn).toBe("2026-08-27")
  })

  // Every word comes from the locale file, so the first List is in the Member's
  // language rather than in English.
  it("takes every word from the language", () => {
    for (const item of starterItems(t, now)) {
      expect(item.label.startsWith("starter.")).toBe(true)
    }
  })

  it("is short enough to read at a glance", () => {
    expect(starterItems(t, now).length).toBeLessThanOrEqual(6)
  })

  // The completed section collapses at the foot of a List, and a Member cannot discover
  // a section that is not there. It also answers the question people ask first: no,
  // ticking something off does not take it away.
  it("arrives with one thing already ticked", () => {
    const items = starterItems(t, now)

    expect(items.filter((item) => item.done)).toHaveLength(1)
    expect(items.filter((item) => !item.done).length).toBeGreaterThan(3)
  })

  it("puts the ticked one last, where the List will collapse it", () => {
    const items = starterItems(t, now)

    expect(items.at(-1)?.done).toBe(true)
  })
})
