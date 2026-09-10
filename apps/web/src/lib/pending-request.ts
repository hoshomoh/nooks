/**
 * The identifier of a request this browser has made, remembered locally.
 *
 * Nooks sends no email, so there is no link to come back on: the browser that asked is
 * the thing that remembers, and the Member returns to the same address to check.
 */
export type PendingRequestKind = "join" | "reset"

const KEY: Record<PendingRequestKind, string> = {
  join: "nooks.join-request",
  reset: "nooks.reset-request",
}

/** readPendingRequest returns the remembered identifier, or null. */
export function readPendingRequest(
  kind: PendingRequestKind,
  storage: Pick<Storage, "getItem"> = window.localStorage,
): string | null {
  try {
    return storage.getItem(KEY[kind])
  } catch {
    // A browser that refuses storage simply cannot check back; asking still works.
    return null
  }
}

export function rememberPendingRequest(
  kind: PendingRequestKind,
  uid: string,
  storage: Pick<Storage, "setItem"> = window.localStorage,
): void {
  try {
    storage.setItem(KEY[kind], uid)
  } catch {
    // Nothing to do: the request was still made.
  }
}

export function forgetPendingRequest(
  kind: PendingRequestKind,
  storage: Pick<Storage, "removeItem"> = window.localStorage,
): void {
  try {
    storage.removeItem(KEY[kind])
  } catch {
    // Nothing to do.
  }
}
