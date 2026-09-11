import { listClient } from "./api"
import { pendingTickStore } from "./pending-tick-store"

/**
 * Ticks whatever a Visitor reached for before they signed in.
 *
 * Called once a session exists. It never throws: the Item may have been deleted, or the
 * List may not be one this Member can reach after all, and neither is a reason to
 * interrupt somebody who has just successfully signed in.
 */
export async function applyPendingTick(): Promise<void> {
  const itemUid = pendingTickStore.take()
  if (!itemUid) {
    return
  }

  try {
    await listClient.setItemDone({ itemUid, done: true })
  } catch {
    // They can tick it themselves; the List is in front of them now.
  }
}
