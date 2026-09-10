import { describe, expect, it } from "vitest"

import {
  applyTheme,
  DEFAULT_THEME,
  isTheme,
  readStoredTheme,
  resolveTheme,
  storeTheme,
  THEME_STORAGE_KEY,
} from "./theme"

describe("resolveTheme", () => {
  it("takes an explicit choice regardless of the machine", () => {
    expect(resolveTheme("light", true)).toBe("light")
    expect(resolveTheme("dark", false)).toBe("dark")
  })

  it("follows the machine when the choice is system", () => {
    expect(resolveTheme("system", true)).toBe("dark")
    expect(resolveTheme("system", false)).toBe("light")
  })
})

describe("isTheme", () => {
  it("accepts the three themes and nothing else", () => {
    expect(isTheme("light")).toBe(true)
    expect(isTheme("dark")).toBe(true)
    expect(isTheme("system")).toBe(true)
    expect(isTheme("solarized")).toBe(false)
    expect(isTheme(null)).toBe(false)
  })
})

describe("readStoredTheme", () => {
  it("returns what was stored", () => {
    expect(readStoredTheme({ getItem: () => "dark" })).toBe("dark")
  })

  it("falls back to the default when nothing is stored", () => {
    expect(readStoredTheme({ getItem: () => null })).toBe(DEFAULT_THEME)
  })

  it("falls back when the stored value is not a theme", () => {
    expect(readStoredTheme({ getItem: () => "solarized" })).toBe(DEFAULT_THEME)
  })

  // A private window throws on access; the app must still render.
  it("falls back when storage throws", () => {
    expect(
      readStoredTheme({
        getItem: () => {
          throw new Error("storage is disabled")
        },
      }),
    ).toBe(DEFAULT_THEME)
  })
})

describe("storeTheme", () => {
  it("writes under the theme key", () => {
    const written: Array<[string, string]> = []
    storeTheme({ setItem: (key, value) => void written.push([key, value]) }, "dark")
    expect(written).toEqual([[THEME_STORAGE_KEY, "dark"]])
  })

  it("does not throw when storage refuses", () => {
    expect(() =>
      storeTheme(
        {
          setItem: () => {
            throw new Error("storage is disabled")
          },
        },
        "dark",
      ),
    ).not.toThrow()
  })
})

describe("applyTheme", () => {
  it("adds the dark class for dark and removes it for light", () => {
    const calls: Array<[string, boolean | undefined]> = []
    const root = { classList: { toggle: (t: string, f?: boolean) => (calls.push([t, f]), true) } }

    applyTheme(root, "dark")
    applyTheme(root, "light")

    expect(calls).toEqual([
      ["dark", true],
      ["dark", false],
    ])
  })
})
