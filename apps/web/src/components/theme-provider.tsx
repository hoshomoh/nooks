import { useCallback, useEffect, useMemo, useState } from "react"
import type { ReactNode } from "react"

import { ThemeContext } from "@/lib/theme-context"
import {
  applyTheme,
  DEFAULT_THEME,
  readStoredTheme,
  resolveTheme,
  storeTheme,
  type Theme,
} from "@/lib/theme"

const DARK_QUERY = "(prefers-color-scheme: dark)"

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(() =>
    typeof window === "undefined" ? DEFAULT_THEME : readStoredTheme(window.localStorage),
  )
  const [prefersDark, setPrefersDark] = useState<boolean>(() =>
    typeof window === "undefined" ? false : window.matchMedia(DARK_QUERY).matches,
  )

  // System means "follow the machine", so the machine is watched, not read once.
  useEffect(() => {
    const query = window.matchMedia(DARK_QUERY)
    const onChange = (event: MediaQueryListEvent) => setPrefersDark(event.matches)
    query.addEventListener("change", onChange)
    return () => query.removeEventListener("change", onChange)
  }, [])

  const resolved = useMemo(() => resolveTheme(theme, prefersDark), [theme, prefersDark])

  useEffect(() => {
    applyTheme(document.documentElement, resolved)
  }, [resolved])

  const setTheme = useCallback((next: Theme) => {
    setThemeState(next)
    storeTheme(window.localStorage, next)
  }, [])

  const value = useMemo(() => ({ theme, resolved, setTheme }), [theme, resolved, setTheme])

  return <ThemeContext value={value}>{children}</ThemeContext>
}
