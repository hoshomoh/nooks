import { registerSW } from "virtual:pwa-register"

/**
 * Whether a newer build of the app is waiting.
 *
 * The service worker holds the app, so a new release does not arrive by itself. This
 * says one is there and then waits: a Note saves when the typing settles, and reloading
 * a second before that trades somebody's words for a version number.
 *
 * Outside a secure context there is no service worker, so this never fires.
 */
export interface AppUpdateStore {
  subscribe: (listener: () => void) => () => void
  /** Whether a new version is waiting. */
  getState: () => boolean
  /** Take it: the page reloads into the new build. */
  apply: () => void
}

export function createAppUpdateStore(
  register: typeof registerSW = registerSW,
): AppUpdateStore {
  let waiting = false
  const listeners = new Set<() => void>()

  const update = register({
    onNeedRefresh() {
      waiting = true
      for (const listener of listeners) {
        listener()
      }
    },
  })

  return {
    subscribe(listener) {
      listeners.add(listener)
      return () => listeners.delete(listener)
    },
    getState: () => waiting,
    apply: () => void update(true),
  }
}

export const appUpdateStore: AppUpdateStore = createAppUpdateStore()
