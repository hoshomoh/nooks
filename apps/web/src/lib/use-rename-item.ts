import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"

import { listClient } from "./api"
import { refreshLists } from "./refresh"
import { bothOf, troubleFrom, type RowTrouble } from "./row-trouble"

/** Renaming one Item, with what the Member believed the name was. */
export interface RenameRequest {
  itemUid: string
  label: string
  /**
   * What was on screen when they started typing.
   *
   * Left out to write regardless, which is what a Member answering "Keep mine" has
   * asked for: they have seen the other version and decided.
   */
  expected?: string
}

/** What a screen needs to offer a Member whose rename did not land. */
export interface RenameItem {
  rename: (request: RenameRequest) => void
  /** The row in trouble, or nothing. At most one at a time. */
  trouble: RowTrouble | null
  /** Send it again, unconditionally: this Member's words win. */
  keepMine: () => void
  /** Keep both, one after the other, rather than guessing which was meant. */
  keepBoth: (theirs: string) => void
  /** Only for a failure: the same change again, on the same terms. */
  tryAgain: () => void
  /** Give up on this Member's text. Only ever because they said so. */
  discard: () => void
}

/**
 * Renaming an Item, and what happens when somebody else got there first.
 *
 * The rename says what it believed the name was, so the Instance can refuse it rather
 * than write over a change this Member never saw — see textUnchanged in item.go. A
 * refusal is not an error to report and move on from: both versions are somebody's
 * words, and DESIGN.md §11 says the row asks.
 *
 * One row at a time. Two rows arguing at once has never happened to a household, and a
 * queue of questions is a worse thing to come back to than one.
 */
export function useRenameItem(): RenameItem {
  const queryClient = useQueryClient()
  const [trouble, setTrouble] = useState<RowTrouble | null>(null)

  const send = useMutation({
    mutationFn: ({ itemUid, label, expected }: RenameRequest) =>
      listClient.updateItem({ itemUid, label, expectedLabel: expected }),
    onSuccess: () => {
      setTrouble(null)
      return refreshLists(queryClient)
    },
    onError: (error, request) => {
      setTrouble(troubleFrom(request.itemUid, request.label, error))
      // What is actually on the List is what the panel shows beside the Member's own,
      // so it has to be read again before the question can be asked honestly.
      return refreshLists(queryClient)
    },
  })

  /** resend writes the given text regardless of what is there now. */
  const resend = (label: string) => {
    if (!trouble) {
      return
    }
    send.mutate({ itemUid: trouble.itemUid, label })
  }

  return {
    trouble,
    rename: (request) => send.mutate(request),
    keepMine: () => resend(trouble?.mine ?? ""),
    keepBoth: (theirs) => resend(bothOf(theirs, trouble?.mine ?? "")),
    tryAgain: () => resend(trouble?.mine ?? ""),
    discard: () => setTrouble(null),
  }
}
