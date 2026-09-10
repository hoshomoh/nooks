import { createContext, useContext } from "react"

import type { ResolvedTheme, Theme } from "./theme"

export type ThemeContextValue = {
  /** What the Member chose. */
  theme: Theme
  /** What that choice resolves to right now. */
  resolved: ResolvedTheme
  setTheme: (theme: Theme) => void
}

export const ThemeContext = createContext<ThemeContextValue | null>(null)

export function useTheme(): ThemeContextValue {
  const value = useContext(ThemeContext)
  if (!value) {
    throw new Error("useTheme must be used inside a ThemeProvider")
  }
  return value
}
