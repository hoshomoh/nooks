import { useSyncExternalStore } from "react"

import { themeStore } from "./theme-store"
import type { ResolvedTheme, Theme } from "./theme"

/**
 * Reads the theme from its store.
 *
 * useSyncExternalStore rather than an effect: the machine's colour preference is an
 * external mutable source, and this is the API React provides for exactly that. Each
 * snapshot is a string, so it is stable by value and needs no memoisation.
 */
export function useTheme(): {
  theme: Theme
  resolved: ResolvedTheme
  setTheme: (theme: Theme) => void
} {
  const theme = useSyncExternalStore(
    themeStore.subscribe,
    themeStore.getTheme,
    themeStore.getTheme,
  )
  const resolved = useSyncExternalStore(
    themeStore.subscribe,
    themeStore.getResolvedTheme,
    themeStore.getResolvedTheme,
  )
  return { theme, resolved, setTheme: themeStore.setTheme }
}
