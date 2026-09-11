/**
 * The tick a Visitor reached for before they had an account.
 *
 * Somebody opens the public list, taps a box, and is told they need to sign in. What
 * they were trying to do should not be lost on the way: they meant to tick that thing,
 * and making them find it again afterwards is the app forgetting on their behalf.
 *
 * Kept in the browser because that is where the intention was formed, and because the
 * Instance has no idea who they are yet. It survives a sign-in that takes days, which
 * is why it is not simply held in memory.
 */
export interface PendingTickStore {
  /** Remembers what they reached for, replacing anything older. */
  remember: (itemUid: string) => void
  /**
   * Returns what was remembered and forgets it.
   *
   * Taken rather than read: it is worth applying once, and a tick that failed because
   * the Member cannot reach that List should not be tried again on every sign-in.
   */
  take: () => string
}

export interface PendingTickStoreDeps {
  storage: Pick<Storage, "getItem" | "setItem" | "removeItem">
}

/** PENDING_TICK_KEY is where the intention waits. */
export const PENDING_TICK_KEY = "nooks.pending-tick"

export function createPendingTickStore(deps: PendingTickStoreDeps): PendingTickStore {
  return {
    remember(itemUid) {
      try {
        deps.storage.setItem(PENDING_TICK_KEY, itemUid)
      } catch {
        // A Visitor in a private window simply signs in and ticks it themselves.
      }
    },
    take() {
      try {
        const held = deps.storage.getItem(PENDING_TICK_KEY) ?? ""
        deps.storage.removeItem(PENDING_TICK_KEY)
        return held
      } catch {
        return ""
      }
    },
  }
}

/** The application's store. */
export const pendingTickStore: PendingTickStore =
  typeof window === "undefined"
    ? { remember: () => {}, take: () => "" }
    : createPendingTickStore({ storage: window.localStorage })
