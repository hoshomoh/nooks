import { Code, ConnectError } from "@connectrpc/connect"

/**
 * What went wrong with a change to one row, and what can be done about it.
 *
 * DESIGN.md §11 draws two different things here and they are different on purpose. A
 * conflict is not a failure — somebody else wrote first, both versions exist, and only
 * a person can say which should survive. An error is a failure, and the one thing that
 * must not happen is the Member's text being thrown away while they are told about it.
 */

/** Somebody else rewrote the text while this Member was writing. */
export interface TextConflict {
  kind: "conflict"
  itemUid: string
  /** What this Member wrote. Never discarded, whatever they choose. */
  mine: string
}

/** The Instance refused the change for some other reason. */
export interface SaveError {
  kind: "error"
  itemUid: string
  mine: string
  /** What the Instance said, in its own words. */
  said: string
}

export type RowTrouble = TextConflict | SaveError

/**
 * troubleFrom reads a failed save into what the row should offer.
 *
 * ABORTED is the one the Instance uses for competing text — see textUnchanged in
 * item.go. Everything else is a failure rather than a disagreement, and is shown as one.
 */
export function troubleFrom(itemUid: string, mine: string, error: unknown): RowTrouble {
  if (error instanceof ConnectError && error.code === Code.Aborted) {
    return { kind: "conflict", itemUid, mine }
  }
  return { kind: "error", itemUid, mine, said: messageOf(error) }
}

/**
 * bothOf is what "Keep both" leaves behind when two people wrote different things.
 *
 * Theirs first, because theirs is what is already on the list and what everybody else
 * has seen. Nothing is merged word by word: a household list is short, and a machine
 * guessing at a sentence somebody wrote is worse than showing them both.
 */
export function bothOf(theirs: string, mine: string): string {
  if (!theirs.trim()) {
    return mine
  }
  if (theirs.trim() === mine.trim()) {
    return theirs
  }
  return `${theirs} / ${mine}`
}

/** messageOf is what the Instance said, or a sentence when it said nothing readable. */
function messageOf(error: unknown): string {
  if (error instanceof ConnectError) {
    return error.rawMessage
  }
  if (error instanceof Error && error.message) {
    return error.message
  }
  return ""
}
