import { useCallback, useSyncExternalStore } from "react"

/**
 * Runs an action when Escape is pressed.
 *
 * useSyncExternalStore rather than an effect: a key press is an external event, and
 * subscribing is exactly what this API is for. The snapshot never changes — nothing
 * about the key is rendered — so the component does not re-render on a press.
 */
export function useEscape(onEscape: () => void): void {
  const subscribe = useCallback(
    (notify: () => void) => {
      const onKeyDown = (event: KeyboardEvent) => {
        if (event.key === "Escape") {
          onEscape()
          notify()
        }
      }
      window.addEventListener("keydown", onKeyDown)
      return () => window.removeEventListener("keydown", onKeyDown)
    },
    [onEscape],
  )

  useSyncExternalStore(
    subscribe,
    () => null,
    () => null,
  )
}
