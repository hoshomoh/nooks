import { useSuspenseQuery } from "@tanstack/react-query"
import type { List, Member } from "@nooks/api"

import { listsQuery } from "./list-queries"
import { currentMemberQuery, instanceQuery } from "./queries"

/** The three reads every signed-in screen renders from. */
export interface SignedInData {
  instanceName: string
  /** Null only while signing out; a route loader has already redirected by then. */
  member: Member | null
  lists: List[]
}

/**
 * Reads the shell's data from the query cache rather than from the route's loader.
 *
 * The loader still fetches it, so the screen never renders empty — but a loader's
 * result is a snapshot taken once, and a Member who adds a List has to see it appear.
 * Reading the same queries here subscribes the screen to them, so invalidating after a
 * mutation updates the sidebar without a reload and without an effect.
 */
export function useSignedInData(): SignedInData {
  const instance = useSuspenseQuery(instanceQuery).data
  const member = useSuspenseQuery(currentMemberQuery).data
  const lists = useSuspenseQuery(listsQuery).data

  return { instanceName: instance.name, member, lists: lists.lists }
}
