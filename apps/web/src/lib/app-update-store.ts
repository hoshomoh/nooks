import { registerSW } from "virtual:pwa-register"

/**
 * Whether a newer build of the app is waiting.
 *
 * The service worker holds the app, so a new release does not arrive by itself.
 * Something has to say so.
 *
 * It says so and then waits. A Note is saved when the typing settles, and a reload a
 * second before that takes somebody's words away to give the app a version number.
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
