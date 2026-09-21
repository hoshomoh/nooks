import { describe, expect, it } from "vitest"
import type { Item } from "@nooks/api"

import { groupDone } from "./done-groups"

/** ticked is an Item completed at a moment, which is all this reads. */
function ticked(label: string, doneAt: string): Item {
  return { uid: label, label, done: true, doneAt } as Item
}

/** labels is what came back, flattened, so a test reads as an expectation. */
function labels(days: { items: Item[] }[]): string[][] {
  return days.map((day) => day.items.map((item) => item.label))
}

describe("gathering what was ticked", () => {
  it("puts each Item under the day it was ticked, newest day first", () => {
    const { days } = groupDone([
      ticked("Bread", "2026-09-16T08:12:00Z"),
      ticked("Kitchen roll", "2026-09-15T18:26:00Z"),
      ticked("Bin bags", "2026-09-16T09:40:00Z"),
    ])

    expect(days).toHaveLength(2)
    expect(labels(days)).toEqual([["Bin bags", "Bread"], ["Kitchen roll"]])
  })

  // The most recent tick is the one a Member is looking for.
  it("puts the most recent tick at the top of its day", () => {
    const { days } = groupDone([
      ticked("Washing powder", "2026-09-16T13:05:00Z"),
      ticked("Bread", "2026-09-16T08:12:00Z"),
    ])

    expect(labels(days)).toEqual([["Washing powder", "Bread"]])
  })

  /*
   * A List a household has used for a year holds a year of ticks.
   *
   * Two days keep their heading and everything older goes behind one line, so opening
   * the section never unrolls a page nobody asked for.
   */
  it("keeps the last two days and counts the rest", () => {
    const { days, earlier } = groupDone([
      ticked("Today one", "2026-09-16T08:00:00Z"),
      ticked("Yesterday one", "2026-09-15T08:00:00Z"),
      ticked("Older one", "2026-09-14T08:00:00Z"),
      ticked("Older two", "2026-09-14T09:00:00Z"),
      ticked("Oldest", "2026-01-02T09:00:00Z"),
    ])

    expect(labels(days)).toEqual([["Today one"], ["Yesterday one"]])
    expect(earlier).toBe(3)
  })

  /*
   * The line that says how many are left has to be able to reach them.
   *
   * It used to be a plain span: "23 completed earlier", with nothing behind it and no
   * way anywhere in the app to see those 23. Ticking is meant to put an Item somewhere
   * rather than delete it, and that line was where they went to disappear.
   */
  it("keeps every day when asked for every day, so nothing is left behind a count", () => {
    const ticks = [
      ticked("Today one", "2026-09-16T08:00:00Z"),
      ticked("Yesterday one", "2026-09-15T08:00:00Z"),
      ticked("Older one", "2026-09-14T08:00:00Z"),
      ticked("Oldest", "2026-01-02T09:00:00Z"),
    ]

    const { days, earlier } = groupDone(ticks, ticks.length)

    expect(earlier).toBe(0)
    expect(labels(days).flat()).toHaveLength(ticks.length)
  })

  it("counts an Item nobody can date rather than inventing a day for it", () => {
    const { days, earlier } = groupDone([
      ticked("Bread", "2026-09-16T08:12:00Z"),
      ticked("From somewhere", ""),
    ])

    expect(labels(days)).toEqual([["Bread"]])
    expect(earlier).toBe(1)
  })

  it("has nothing to say about a List nobody has ticked anything on", () => {
    expect(groupDone([])).toEqual({ days: [], earlier: 0 })
  })

  // A tick at 23:30 belongs to the day the Member was having, not the UTC one.
  it("groups by the day where the Member is", () => {
    const late = new Date(2026, 8, 16, 23, 30)
    const soonAfter = new Date(2026, 8, 16, 23, 45)
    const { days } = groupDone([
      ticked("Late", late.toISOString()),
      ticked("Later", soonAfter.toISOString()),
    ])

    expect(days).toHaveLength(1)
  })
})
