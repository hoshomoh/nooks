import { describe, expect, it } from "vitest"

import { groupByDay } from "./dated-queries"

type Entry = { item?: { dueOn: string }; label: string }

describe("groupByDay", () => {
  it("buckets by due date, keeping the order they arrived in", () => {
    const entries: Entry[] = [
      { item: { dueOn: "2026-08-25" }, label: "Descale the kettle" },
      { item: { dueOn: "2026-08-25" }, label: "Bike service" },
      { item: { dueOn: "2026-08-26" }, label: "Heating" },
    ]

    const groups = groupByDay(entries)

    expect(groups.map((group) => group.day)).toEqual(["2026-08-25", "2026-08-26"])
    expect(groups[0].items.map((entry) => entry.label)).toEqual([
      "Descale the kettle",
      "Bike service",
    ])
  })

  it("returns nothing for nothing", () => {
    expect(groupByDay([])).toEqual([])
  })
})
