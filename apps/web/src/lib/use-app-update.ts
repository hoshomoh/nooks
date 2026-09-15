import { useSyncExternalStore } from "react"

import { appUpdateStore, type AppUpdateStore } from "./app-update-store"

/** Whether a newer build is waiting, and how to take it. */
export interface AppUpdate {
  waiting: boolean
  apply: () => void
}

/** Reads the update store, which lives outside React as the connection does. */
export function useAppUpdate(store: AppUpdateStore = appUpdateStore): AppUpdate {
  const waiting = useSyncExternalStore(
    store.subscribe,
    store.getState,
    // Nothing is waiting on the server; there is no service worker there.
    () => false,
  )
  return { waiting, apply: store.apply }
}
