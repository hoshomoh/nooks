import { describe, expect, it } from "vitest"

import { enGB } from "date-fns/locale"

import { dayHeading, daysUntil, dueLabel, isOverdue, parseDue, rangeFrom, toStored } from "./dates"
import type { DateLabelOptions } from "./dates"

// A Tuesday, matching the design's "Tuesday, 25 August".
const today = new Date(2026, 7, 25)

// The near-day words come from the caller, so a test supplies them directly.
const labels: DateLabelOptions = {
  from: today,
  locale: enGB,
  todayWord: "Today",
  tomorrowWord: "Tomorrow",
}

describe("parseDue", () => {
  it("accepts a stored date", () => {
    expect(toStored(parseDue("2026-08-25") as Date)).toBe("2026-08-25")
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
    expect(dueLabel("2026-08-25", labels)).toBe("Today")
    expect(dueLabel("2026-08-26", labels)).toBe("Tomorrow")
  })

  it("uses a weekday within the coming week", () => {
    expect(dueLabel("2026-08-28", labels)).toBe("Fri")
  })

  it("adds the date beyond it", () => {
    expect(dueLabel("2026-09-05", labels)).toBe("Sat 5 Sep")
  })

  it("shows nothing for an Item with no due date", () => {
    expect(dueLabel("", labels)).toBe("")
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
      expect(dueLabel(due, labels)).toBe(want)
    }
  })
})


describe("dayHeading", () => {
  it("names the near days and uses a weekday beyond them", () => {
    expect(dayHeading("2026-08-25", labels)).toBe("Today")
    expect(dayHeading("2026-08-26", labels)).toBe("Tomorrow")
    expect(dayHeading("2026-08-28", labels)).toBe("Friday")
  })
})

describe("rangeFrom", () => {
  it("bounds a window of days in stored form", () => {
    expect(rangeFrom(today, 14)).toEqual({ start: "2026-08-25", end: "2026-09-08" })
  })
})
