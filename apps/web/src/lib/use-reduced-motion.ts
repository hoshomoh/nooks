import { useSyncExternalStore } from "react"

const QUERY = "(prefers-reduced-motion: reduce)"

// Absent in some runtimes, jsdom among them, where nothing asked for less motion.
function media(): MediaQueryList | null {
  return typeof window.matchMedia === "function" ? window.matchMedia(QUERY) : null
}

function subscribe(onChange: () => void): () => void {
  const list = media()
  list?.addEventListener("change", onChange)
  return () => list?.removeEventListener("change", onChange)
}

function snapshot(): boolean {
  return media()?.matches ?? false
}

/**
 * Whether the Member asked for less motion.
 *
 * For motion drawn from script, which the stylesheet's own reduced-motion rule cannot
 * reach: a row that slides into place is animated by the drag library, not by CSS.
 */
export function useReducedMotion(): boolean {
  return useSyncExternalStore(subscribe, snapshot, () => false)
}
