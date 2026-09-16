import { useCallback, useSyncExternalStore } from "react"

/**
 * Runs an action when Escape is pressed, unless something nearer the key already dealt
 * with it.
 *
 * Escape means "back out of the thing I am in", and a screen is the outermost of those.
 * A control that takes the key first — a title being renamed, a date picker, a menu —
 * says so by calling preventDefault, and this listens on the window, so by the time it
 * runs that has happened. Without the check, Escape out of a rename also leaves the
 * screen behind it.
 *
 * useSyncExternalStore rather than an effect: a key press is an external event, and
 * subscribing is exactly what this API is for. The snapshot never changes, so the
 * component does not re-render on a press.
 */
export function useEscape(onEscape: () => void): void {
  const subscribe = useCallback(
    (notify: () => void) => {
      const onKeyDown = (event: KeyboardEvent) => {
        if (event.key === "Escape" && !event.defaultPrevented) {
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
