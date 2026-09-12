"use client"

// From Fumadocs rather than next-themes directly: it re-exports the same hook, and the
// provider that owns the state is already its. One dependency instead of two on the
// same thing.
import { useTheme } from "fumadocs-ui/provider/base"
import { useSyncExternalStore } from "react"

/** The three choices, in the order DESIGN.md §2 lists them. */
const THEMES = ["light", "dark", "system"] as const

/** Nothing to subscribe to: whether we have hydrated only ever changes once. */
const NEVER_CHANGES = () => () => {}

/**
 * Whether the browser has taken over from the server-rendered HTML.
 *
 * useSyncExternalStore rather than an effect, which is the rule everywhere in Nooks:
 * it takes a server snapshot and a client one, which is precisely the question being
 * asked. An effect would answer it by rendering the wrong thing first and correcting
 * itself.
 */
function useHydrated(): boolean {
  return useSyncExternalStore(
    NEVER_CHANGES,
    () => true,
    () => false,
  )
}

/**
 * Light, dark or system — the same three the app offers, in the same order.
 *
 * Nothing is marked until the browser has hydrated. The server cannot know which theme
 * this reader chose, so marking one on the way out would be marking the wrong one for
 * anybody who chose differently.
 */
export function ThemeToggle() {
  const { theme, setTheme } = useTheme()
  const hydrated = useHydrated()

  return (
    <div
      role="radiogroup"
      aria-label="Theme"
      className="flex h-control-settings items-center gap-0.5 rounded-lg border border-border p-0.5"
    >
      {THEMES.map((choice) => {
        const chosen = hydrated && theme === choice
        return (
          <button
            key={choice}
            type="button"
            role="radio"
            aria-checked={chosen}
            onClick={() => setTheme(choice)}
            className={`h-full rounded-md px-3 text-meta capitalize transition-colors ${
              chosen
                ? "bg-secondary font-medium text-foreground"
                : "text-secondary-foreground hover:text-foreground"
            }`}
          >
            {choice}
          </button>
        )
      })}
    </div>
  )
}
