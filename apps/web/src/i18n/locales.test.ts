import { describe, expect, it } from "vitest"

import { DEFAULT_LOCALE, matchLocale } from "./locales"

describe("matchLocale", () => {
  const supported = ["en", "en-GB", "de", "pt-BR"]

  it("takes an exact match", () => {
    expect(matchLocale(["en-GB"], supported)).toBe("en-GB")
    expect(matchLocale(["pt-BR"], supported)).toBe("pt-BR")
  })

  it("ignores case", () => {
    expect(matchLocale(["EN-gb"], supported)).toBe("en-GB")
  })

  // Someone asking for Austrian German should read German, not English.
  it("falls back to the base language", () => {
    expect(matchLocale(["de-AT"], supported)).toBe("de")
  })

  it("tries each preference in order", () => {
    expect(matchLocale(["fr", "de"], supported)).toBe("de")
  })

  it("falls back to the default when nothing matches", () => {
    expect(matchLocale(["ja", "ko"], supported)).toBe(DEFAULT_LOCALE)
  })

  it("handles an empty preference list", () => {
    expect(matchLocale([], supported)).toBe(DEFAULT_LOCALE)
  })
})
