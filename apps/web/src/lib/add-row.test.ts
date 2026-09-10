import { describe, expect, it } from "vitest"
import { enGB } from "date-fns/locale"

import { parseAddRow, type AddRowParseOptions, type AddRowVocabulary } from "./add-row"

const vocabulary: AddRowVocabulary = {
  todayWords: ["today", "tonight"],
  tomorrowWords: ["tomorrow"],
  units: ["kg", "g", "l", "ml", "x"],
}

// Tuesday, 25 August 2026 — the day the artboard is drawn on.
const from = new Date(2026, 7, 25)

const options: AddRowParseOptions = { from, locale: enGB, vocabulary }

/** parse reads a line on the fixed day. */
function parse(text: string, taken?: AddRowParseOptions["taken"]) {
  return parseAddRow(text, { ...options, taken })
}

describe("what an item needs", () => {
  it("accepts a bare name", () => {
    expect(parse("Milk")).toMatchObject({ name: "Milk", quantity: "", due: "" })
  })

  it("leaves a sentence alone when nothing is recognised", () => {
    expect(parse("Coffee — the good beans from the market").name).toBe(
      "Coffee — the good beans from the market",
    )
  })
})

describe("quantities", () => {
  it("normalises a unit written against the number", () => {
    expect(parse("Tomatoes 1kg")).toMatchObject({ name: "Tomatoes", quantity: "1 kg" })
  })

  it("reads a unit written as its own word", () => {
    expect(parse("Flour 250 g")).toMatchObject({ name: "Flour", quantity: "250 g" })
  })

  it("accepts a bare number", () => {
    expect(parse("Milk 2")).toMatchObject({ name: "Milk", quantity: "2" })
  })

  it("keeps an unknown unit as part of the name", () => {
    expect(parse("Tomatoes 3 punnets").name).toBe("Tomatoes 3 punnets")
  })

  // The scan stops at the first word it does not recognise, which is what keeps a
  // number in the middle of a sentence part of the name.
  it("never reaches a number that is not at the end", () => {
    expect(parse("Call 2 plumbers")).toMatchObject({ name: "Call 2 plumbers", quantity: "" })
  })

  it("takes only the first quantity", () => {
    expect(parse("Milk 2 3")).toMatchObject({ name: "Milk 2", quantity: "3" })
  })
})

describe("dates", () => {
  it("reads today", () => {
    expect(parse("Bins tonight").due).toBe("2026-08-25")
  })

  it("reads tomorrow", () => {
    expect(parse("Bins tomorrow").due).toBe("2026-08-26")
  })

  it("resolves a weekday to its next occurrence", () => {
    expect(parse("Bins sat").due).toBe("2026-08-29")
  })

  // Naming a day means the one still to come; today would have been called "today".
  it("skips today when today is the weekday named", () => {
    expect(parse("Bins tue").due).toBe("2026-09-01")
  })

  it("reads a day and month", () => {
    expect(parse("Ferry 30 aug").due).toBe("2026-08-30")
  })

  it("reads a numeric date day first", () => {
    expect(parse("Ferry 30/8").due).toBe("2026-08-30")
  })

  it("rolls a date that has passed into next year", () => {
    expect(parse("Ferry 1/8").due).toBe("2027-08-01")
  })

  it("ignores a day the month never has", () => {
    expect(parse("Ferry 31/9").name).toBe("Ferry 31/9")
  })
})

describe("both at once", () => {
  it("reads a quantity and a date from one sentence", () => {
    expect(parse("Tomatoes 1kg sat")).toMatchObject({
      name: "Tomatoes",
      quantity: "1 kg",
      due: "2026-08-29",
    })
  })

  it("puts the chips in the order they were typed", () => {
    expect(parse("Tomatoes 1kg sat").chips.map((chip) => chip.kind)).toEqual([
      "quantity",
      "due",
    ])
  })
})

describe("leaving text alone", () => {
  it("never parses a quoted token", () => {
    expect(parse('Buy the "2kg"').name).toBe('Buy the "2kg"')
  })

  it("never consumes the whole line", () => {
    expect(parse("tomorrow")).toMatchObject({ name: "tomorrow", due: "" })
  })

  it("skips a kind that was already lifted out", () => {
    expect(parse("Tomatoes sat", ["due"]).name).toBe("Tomatoes sat")
  })
})
