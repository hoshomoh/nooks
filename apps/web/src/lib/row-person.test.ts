import { beforeAll, describe, expect, it } from "vitest"
import i18n from "@/i18n"

import { readyForEnglish } from "@/test/i18n"
import type { Translate } from "./translate"
import { addedBy, tickedBy } from "./row-person"

beforeAll(readyForEnglish)

const t: Translate = (key, options) => i18n.t(key, options)

describe("whose circle a row ends in", () => {
  it("is whoever added it, said in full on hover", () => {
    expect(addedBy(t, "Anna", "")).toEqual({ name: "Anna", said: "Anna" })
  })

  it("names the Member before the token they added it through", () => {
    expect(addedBy(t, "Anna", "Kitchen tablet")).toEqual({
      name: "Anna",
      said: "Anna via Kitchen tablet",
      viaToken: true,
    })
  })

  // Somebody removed leaves no name, and nobody else's is put in its place.
  it("is nobody's when nobody can be named", () => {
    expect(addedBy(t, "", "Kitchen tablet")).toBeUndefined()
    expect(tickedBy("")).toBeUndefined()
  })

  it("is whoever ticked it, for a completed Item", () => {
    expect(tickedBy("Jonas")).toEqual({ name: "Jonas", said: "Jonas" })
  })
})
