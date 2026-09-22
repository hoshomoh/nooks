import { ConnectError } from "@connectrpc/connect"

import type { Translate } from "./translate"

/** What the Instance sends instead of the words for a failure inside itself. */
const KIND = "nooks-error-kind"
const REF = "nooks-error-ref"

/**
 * The message to show a Member for a failed request.
 *
 * Connect puts the code in front of the message ("invalid_argument: …"), which is for
 * the developer rather than the person reading it, so the plain sentence is taken out.
 *
 * A failure inside the Instance carries no sentence at all. It used to carry the cause,
 * which meant a table name or the path to the database file could be drawn on screen,
 * and it was English wherever it was read. What comes now is a kind and a reference: the
 * kind chooses a line from the catalogue in the Member's own language, and the reference
 * is what they quote to whoever runs the Instance, who can find the matching log line.
 *
 * Refusals the Instance writes on purpose still arrive as their own words, and those
 * words are still English. That is a smaller version of the same problem and is written
 * down in the defense log rather than pretended away here.
 */
export function messageFrom(t: Translate, error: unknown): string {
  // No error is not a failure to describe: returning something here would have every
  // form on screen already apologising.
  if (error === null || error === undefined) {
    return ""
  }

  if (error instanceof ConnectError) {
    const kind = error.metadata.get(KIND)
    if (kind) {
      return t(keyFor(kind), { ref: error.metadata.get(REF) ?? "" })
    }
    return error.rawMessage
  }

  if (error instanceof Error && error.message !== "") {
    return error.message
  }
  return t("error.unreachable")
}

/** keyFor is the catalogue entry a kind is drawn from, and what an unknown one falls to. */
function keyFor(kind: string): string {
  switch (kind) {
    case "save-failed":
      return "error.saveFailed"
    case "load-failed":
      return "error.loadFailed"
    // An Instance newer than this tab can send a kind it has never heard of, the way the
    // live stream can send a kind it has no case for. Saying less is better than saying
    // the key.
    default:
      return "error.internal"
  }
}
