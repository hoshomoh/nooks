import { describe, expect, it } from "vitest"

import { createThemeStore, type ThemeStoreDeps } from "./theme-store"

/** A fake of everything the store touches, so no browser is involved. */
function fakeDeps(overrides: { stored?: string | null; prefersDark?: boolean } = {}) {
  const stored = new Map<string, string>()
  if (overrides.stored != null) {
    stored.set("nooks.theme", overrides.stored)
  }

  let matches = overrides.prefersDark ?? false
  const changeListeners = new Set<(event: MediaQueryListEvent) => void>()
  const classes = new Set<string>()

  const deps: ThemeStoreDeps = {
    storage: {
      getItem: (key) => stored.get(key) ?? null,
      setItem: (key, value) => void stored.set(key, value),
    },
    prefersDark: {
      get matches() {
        return matches
      },
      addEventListener: (_type, listener) => void changeListeners.add(listener),
      removeEventListener: (_type, listener) => void changeListeners.delete(listener),
    },
    root: {
      classList: {
        toggle: (token: string, force?: boolean) => {
          const on = force ?? !classes.has(token)
          if (on) classes.add(token)
          else classes.delete(token)
          return on
        },
      },
    },
  }

  /** setSystemDark simulates the machine's preference changing while the page is open. */
  const setSystemDark = (next: boolean) => {
    matches = next
    for (const listener of changeListeners) {
      listener({ matches: next } as MediaQueryListEvent)
    }
  }

  return { deps, classes, stored, setSystemDark }
}

describe("createThemeStore", () => {
  it("applies the theme as soon as it is created, before anything renders", () => {
    const { deps, classes } = fakeDeps({ stored: "dark" })
    createThemeStore(deps)
    expect(classes.has("dark")).toBe(true)
  })

  it("follows the machine when the choice is system", () => {
    const { deps, classes } = fakeDeps({ prefersDark: true })
    const store = createThemeStore(deps)
    expect(store.getTheme()).toBe("system")
    expect(store.getResolvedTheme()).toBe("dark")
    expect(classes.has("dark")).toBe(true)
  })

  it("reacts when the machine's preference changes while the page is open", () => {
    const { deps, classes, setSystemDark } = fakeDeps({ prefersDark: false })
    const store = createThemeStore(deps)

    let notified = 0
    store.subscribe(() => void notified++)

    setSystemDark(true)

    expect(store.getResolvedTheme()).toBe("dark")
    expect(classes.has("dark")).toBe(true)
    expect(notified).toBe(1)
  })

  it("ignores the machine once a Member chooses explicitly", () => {
    const { deps, classes, setSystemDark } = fakeDeps({ prefersDark: false })
    const store = createThemeStore(deps)

    store.setTheme("light")
    setSystemDark(true)

    expect(store.getResolvedTheme()).toBe("light")
    expect(classes.has("dark")).toBe(false)
  })

  it("remembers the choice", () => {
    const { deps, stored } = fakeDeps()
    createThemeStore(deps).setTheme("dark")
    expect(stored.get("nooks.theme")).toBe("dark")
  })

  it("notifies subscribers once per real change, and not otherwise", () => {
    const { deps } = fakeDeps({ prefersDark: false })
    const store = createThemeStore(deps)

    let notified = 0
    const unsubscribe = store.subscribe(() => void notified++)

    store.setTheme("dark")
    store.setTheme("dark") // no change
    expect(notified).toBe(1)

    unsubscribe()
    store.setTheme("light")
    expect(notified).toBe(1)
  })
})
