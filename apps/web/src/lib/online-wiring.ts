import { onlineManager } from "@tanstack/react-query"

import type { ConnectionStore } from "./connection-store"

/**
 * Tells TanStack Query what Nooks means by offline.
 *
 * Its own answer is the browser's, which is the wrong one twice over: a machine on
 * hotel wifi that reaches nothing still reports online, and a machine that reports
 * offline may still reach an Instance on the same network. Nooks already has a better
 * answer — what the last request actually found — so the query layer is pointed at it
 * rather than left with a second opinion.
 *
 * What this buys, once wired: a mutation made while offline is paused rather than
 * failed, and paused is a thing that can still be sent later. A failure is not.
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
