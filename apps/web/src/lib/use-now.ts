import { useSyncExternalStore } from "react"

import { clockStore } from "./clock-store"

/** Reads the time from its store, so a label measured against now moves on by itself. */
export function useNow(): Date {
  return useSyncExternalStore(clockStore.subscribe, clockStore.getState, clockStore.getState)
}
