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

/*
The examples the docs give a Member, held to what the parser does.

These were wrong once. The website design showed "2 oat milk saturday" becoming three
things, the docs repeated it, and the parser had never done that — DESIGN.md says only
the trailing words are read, which is what keeps "Call 2 plumbers" a name. A promise
made to a reader is worth a test, so the next person to change either finds out here.
*/
describe("what the docs promise", () => {
  it("reads a quantity and a date off the end of the line", () => {
    expect(parse("oat milk 2 saturday")).toMatchObject({ name: "oat milk", quantity: "2" })
    expect(parse("oat milk 2 saturday").due).not.toBe("")
  })

  it("normalises a number and its unit", () => {
    expect(parse("rye flour 1kg")).toMatchObject({ name: "rye flour", quantity: "1 kg" })
    expect(parse("tomatoes 250 g")).toMatchObject({ name: "tomatoes", quantity: "250 g" })
  })

  it("reads a date on its own", () => {
    expect(parse("call the plumber tomorrow").name).toBe("call the plumber")
    expect(parse("bins thursday").name).toBe("bins")
  })

  it("leaves a number that is not at the end alone", () => {
    // The rule that earns the trailing-only scan. Two plumbers is not a quantity.
    expect(parse("Call 2 plumbers")).toMatchObject({ name: "Call 2 plumbers", quantity: "" })
    expect(parse("2 oat milk")).toMatchObject({ name: "2 oat milk", quantity: "" })
  })

  it("leaves a quoted token alone", () => {
    expect(parse('Buy "2kg" bag').quantity).toBe("")
  })
})
