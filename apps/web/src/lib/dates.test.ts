import { describe, expect, it } from "vitest"

import { daysUntil, dueLabel, isOverdue, parseDue } from "./dates"

// A Tuesday, matching the design's "Tuesday, 25 August".
const today = new Date("2026-08-25T00:00:00Z")

describe("parseDue", () => {
  it("accepts a stored date", () => {
    expect(parseDue("2026-08-25")?.toISOString()).toBe("2026-08-25T00:00:00.000Z")
  })

  it("rejects anything that is not one", () => {
    for (const value of ["", "next friday", "25/08/2026", "2026-13-99"]) {
      expect(parseDue(value)).toBeNull()
    }
  })
})

describe("daysUntil", () => {
  it("counts whole days either way", () => {
    expect(daysUntil("2026-08-25", today)).toBe(0)
    expect(daysUntil("2026-08-26", today)).toBe(1)
    expect(daysUntil("2026-08-22", today)).toBe(-3)
  })
})

describe("isOverdue", () => {
  // Something due today is not late yet.
  it("does not call today overdue", () => {
    expect(isOverdue("2026-08-25", today)).toBe(false)
  })

  it("calls the past overdue", () => {
    expect(isOverdue("2026-08-22", today)).toBe(true)
  })
})

describe("dueLabel", () => {
  it("names the near days rather than dating them", () => {
    expect(dueLabel("2026-08-25", today)).toBe("Today")
    expect(dueLabel("2026-08-26", today)).toBe("Tomorrow")
  })

  it("uses a weekday within the coming week", () => {
    expect(dueLabel("2026-08-28", today)).toBe("Fri")
  })

  it("adds the date beyond it", () => {
    expect(dueLabel("2026-09-05", today)).toBe("Sat 5 Sep")
  })

  it("shows nothing for an Item with no due date", () => {
    expect(dueLabel("", today)).toBe("")
  })
})

// Month names are written out rather than taken from Intl, which renders September as
// "Sept" under en-GB and changes between runtime versions.
describe("month names", () => {
  it("abbreviates every month to three letters", () => {
    const cases: Array<[string, string]> = [
      ["2026-01-10", "Sat 10 Jan"],
      ["2026-09-05", "Sat 5 Sep"],
      ["2026-12-25", "Fri 25 Dec"],
    ]
    for (const [due, want] of cases) {
      expect(dueLabel(due, today)).toBe(want)
    }
  })
})
