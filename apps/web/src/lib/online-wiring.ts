import { onlineManager } from "@tanstack/react-query"

import type { ConnectionStore } from "./connection-store"

/**
 * Tells TanStack Query what nooks means by offline.
 *
 * Its own answer is the browser's, which is wrong both ways: hotel wifi that reaches
 * nothing still reports online, and a device that reports offline may still reach an
 * Instance on the same network. What the last request found is the better answer.
 *
 * Once wired, a mutation made offline is paused rather than failed, and a paused one
 * can still be sent later.
 */

/** The part of TanStack's online manager this wiring uses. */
export interface OnlineManagerLike {
  setOnline: (online: boolean) => void
  setEventListener: (setup: (setOnline: (online: boolean) => void) => () => void) => void
}

export function wireOnline(
  store: ConnectionStore,
  manager: OnlineManagerLike = onlineManager,
): void {
  const tell = (setOnline: (online: boolean) => void) => setOnline(store.getState().online)

  // Set before subscribing: the first mutation may be made before anything changes,
  // and until something does, an unset manager assumes online.
  manager.setOnline(store.getState().online)
  manager.setEventListener((setOnline) => {
    tell(setOnline)
    return store.subscribe(() => tell(setOnline))
  })
}
