import { useSyncExternalStore } from "react"

import { liveStore, type LiveState } from "./live-store"

/** Reads what the live stream has produced. */
export function useLive(): LiveState {
  return useSyncExternalStore(liveStore.subscribe, liveStore.getState, liveStore.getState)
}
