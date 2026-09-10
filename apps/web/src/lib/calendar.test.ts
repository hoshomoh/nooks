import { describe, expect, it } from "vitest"

import { buildMonth, byDay } from "./calendar"

describe("buildMonth", () => {
  // August 2026 starts on a Saturday and ends on a Monday — both edges need padding.
  it("pads to whole weeks starting on Monday", () => {
    const { weeks } = buildMonth(new Date(2026, 7, 1))

    expect(weeks[0].days[0].date).toBe("2026-07-27")
    expect(weeks[0].days).toHaveLength(7)
    expect(weeks.at(-1)?.days.at(-1)?.date).toBe("2026-09-06")
  })

  it("marks the padding days as outside the month", () => {
    const { weeks } = buildMonth(new Date(2026, 7, 1))

    expect(weeks[0].days[0].inMonth).toBe(false)
    expect(weeks[0].days.find((day) => day.date === "2026-08-01")?.inMonth).toBe(true)
  })

  it("covers every day of the month exactly once", () => {
    const { weeks } = buildMonth(new Date(2026, 7, 1))
    const inMonth = weeks.flatMap((week) => week.days).filter((day) => day.inMonth)

    expect(inMonth).toHaveLength(31)
    expect(new Set(inMonth.map((day) => day.date)).size).toBe(31)
  })

  // February 2027 starts on a Monday, so the first week needs no padding at all.
  it("handles a month that starts on a Monday", () => {
    const { weeks } = buildMonth(new Date(2027, 1, 1))
    expect(weeks[0].days[0].date).toBe("2027-02-01")
  })
})

describe("byDay", () => {
  it("buckets entries by their due date", () => {
    const entries = [
      { due: "2026-08-25", label: "Kettle" },
      { due: "2026-08-25", label: "Bike" },
      { due: "2026-08-26", label: "Heating" },
    ]

    const map = byDay(entries, (entry) => entry.due)

    expect(map.get("2026-08-25")).toHaveLength(2)
    expect(map.get("2026-08-26")).toHaveLength(1)
  })

  // An undated Item never appears on a calendar.
  it("drops entries with no due date", () => {
    const map = byDay([{ due: "" }], (entry) => entry.due)
    expect(map.size).toBe(0)
  })
})
