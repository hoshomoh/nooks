import type { QueryClient } from "@tanstack/react-query"

/**
 * What a change to a List or an Item can affect.
 *
 * Named once rather than invalidating everything: an Instance's name and the signed-in
 * Member do not change because somebody ticked the milk, and refetching them after
 * every keystroke of a Note is a round trip a home server does not need to serve.
 */
const AFFECTED_BY_A_CHANGE = [["lists"], ["list"], ["dated"], ["search"]] as const

/** refreshLists re-reads everything a List or Item change can have altered. */
export async function refreshLists(queryClient: QueryClient): Promise<void> {
  await Promise.all(
    AFFECTED_BY_A_CHANGE.map((queryKey) => queryClient.invalidateQueries({ queryKey })),
  )
}
