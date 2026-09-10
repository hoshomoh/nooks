import {
  applyTheme,
  DEFAULT_THEME,
  readStoredTheme,
  resolveTheme,
  storeTheme,
  type ClassListLike,
  type ResolvedTheme,
  type Theme,
} from "./theme"

/**
 * The theme, as an external store.
 *
 * React reads this through useSyncExternalStore and never synchronises it with an
 * effect. The module owns both halves of the job — tracking the machine's preference
 * and putting the class on <html> — so that changing the theme is one synchronous call
 * with no render-then-patch step in between.
 */
export type ThemeStore = {
  subscribe: (listener: () => void) => () => void
  getTheme: () => Theme
  getResolvedTheme: () => ResolvedTheme
  setTheme: (theme: Theme) => void
}

/** What the store needs from its surroundings, so a test can supply all of it. */
export type ThemeStoreDeps = {
  storage: Pick<Storage, "getItem" | "setItem">
  /** The machine's dark-mode preference, as a subscribable source. */
  prefersDark: {
    matches: boolean
    addEventListener: (type: "change", listener: (event: MediaQueryListEvent) => void) => void
    removeEventListener: (type: "change", listener: (event: MediaQueryListEvent) => void) => void
  }
  /** The element the `dark` class goes on. */
  root: { classList: ClassListLike }
}

/**
 * createThemeStore builds a store over the given surroundings and applies the current
 * theme immediately, so the DOM is correct before anything renders.
 */
export function createThemeStore(deps: ThemeStoreDeps): ThemeStore {
  let theme: Theme = readStoredTheme(deps.storage)
  let resolved: ResolvedTheme = resolveTheme(theme, deps.prefersDark.matches)
  const listeners = new Set<() => void>()

  applyTheme(deps.root, resolved)

  /** update recomputes the resolved theme, touches the DOM, and notifies React. */
  const update = (next: Theme, prefersDark: boolean) => {
    const nextResolved = resolveTheme(next, prefersDark)
    if (next === theme && nextResolved === resolved) {
      return
    }
    theme = next
    resolved = nextResolved
    applyTheme(deps.root, resolved)
    for (const listener of listeners) {
      listener()
    }
  }

  // "System" means follow the machine, so the machine is watched for as long as the
  // page lives. There is no unsubscribe because there is no point at which the app
  // stops caring.
  deps.prefersDark.addEventListener("change", (event) => update(theme, event.matches))

  return {
    subscribe(listener) {
      listeners.add(listener)
      return () => listeners.delete(listener)
    },
    getTheme: () => theme,
    getResolvedTheme: () => resolved,
    setTheme(next) {
      storeTheme(deps.storage, next)
      update(next, deps.prefersDark.matches)
    },
  }
}

/** A store for a document with no browser behind it, so nothing has to null-check. */
function createInertStore(): ThemeStore {
  return {
    subscribe: () => () => {},
    getTheme: () => DEFAULT_THEME,
    getResolvedTheme: () => "light",
    setTheme: () => {},
  }
}

/** The application's store. Created at import time, before anything renders. */
export const themeStore: ThemeStore =
  typeof window === "undefined"
    ? createInertStore()
    : createThemeStore({
        storage: window.localStorage,
        prefersDark: window.matchMedia("(prefers-color-scheme: dark)"),
        root: document.documentElement,
      })
