/**
 * The tick a Visitor reached for before they had an account.
 *
 * Somebody taps a box on the public list and is told to sign in. Making them find it
 * again afterwards is the app forgetting on their behalf.
 *
 * Kept in the browser, since the Instance has no idea who they are yet, and on disk
 * rather than in memory because signing in can take days.
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
  /**
   * Forgets it without applying it, for when the browser changes hands.
   *
   * An intention belongs to whoever had it. It outlives a tab on purpose, because
   * signing in can take days, and that is exactly why signing out has to end it: the
   * next person at a shared tablet did not reach for anything.
   */
  forget: () => void
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
    forget() {
      try {
        deps.storage.removeItem(PENDING_TICK_KEY)
      } catch {
        // Nothing was stored, so there is nothing to end.
      }
    },
  }
}

export const pendingTickStore: PendingTickStore =
  typeof window === "undefined"
    ? { remember: () => {}, take: () => "", forget: () => {} }
    : createPendingTickStore({ storage: window.localStorage })
