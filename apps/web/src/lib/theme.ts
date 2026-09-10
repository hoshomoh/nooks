/**
 * Theme selection: Light, Dark or System, per DESIGN.md §2.
 *
 * The resolution rules are pure functions so they can be tested without a browser.
 * Only applyTheme touches the DOM, and it takes the element to act on.
 */

/** What a Member chose in Settings → Appearance. */
export type Theme = "light" | "dark" | "system"

/** What that choice resolves to right now. */
export type ResolvedTheme = "light" | "dark"

/** Where the choice is remembered. */
export const THEME_STORAGE_KEY = "nooks.theme"

/** System is the default: Nooks follows the machine until told otherwise. */
export const DEFAULT_THEME: Theme = "system"

/** resolveTheme turns a choice plus the machine's preference into a concrete theme. */
export function resolveTheme(theme: Theme, prefersDark: boolean): ResolvedTheme {
  if (theme === "system") {
    return prefersDark ? "dark" : "light"
  }
  return theme
}

/** isTheme narrows unknown stored text to a Theme. */
export function isTheme(value: unknown): value is Theme {
  return value === "light" || value === "dark" || value === "system"
}

/**
 * readStoredTheme returns the remembered choice, or the default when nothing is
 * stored or storage is unavailable — a private window must still render.
 */
export function readStoredTheme(storage: Pick<Storage, "getItem">): Theme {
  try {
    const stored = storage.getItem(THEME_STORAGE_KEY)
    return isTheme(stored) ? stored : DEFAULT_THEME
  } catch {
    return DEFAULT_THEME
  }
}

/** storeTheme remembers a choice, and does nothing if storage refuses. */
export function storeTheme(storage: Pick<Storage, "setItem">, theme: Theme): void {
  try {
    storage.setItem(THEME_STORAGE_KEY, theme)
  } catch {
    // A Member in a private window still gets the theme for this session.
  }
}

/**
 * applyTheme puts the resolved theme on the document. shadcn's dark variant is
 * `&:is(.dark *)`, so the class is the whole contract — see DESIGN.md §16.
 */
export function applyTheme(root: Element, resolved: ResolvedTheme): void {
  root.classList.toggle("dark", resolved === "dark")
}
