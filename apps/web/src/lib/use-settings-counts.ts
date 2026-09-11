import { useQuery } from "@tanstack/react-query"

import type { SettingsCounts } from "@/components/ds/settings-shell"
import { groupsQuery, membersQuery } from "./sharing-queries"
import { tokensQuery } from "./token-queries"

/**
 * The numbers beside the settings nav entries.
 *
 * useQuery rather than useSuspenseQuery: these are furniture, and a settings page should
 * render before a count it does not need has arrived. A Member is also refused the
 * Members and Groups reads, so those simply stay undefined and their entries — which
 * they cannot see anyway — carry no number.
 */
export function useSettingsCounts(): SettingsCounts {
  const tokens = useQuery(tokensQuery)
  const members = useQuery(membersQuery)
  const groups = useQuery(groupsQuery)

  return {
    tokens: tokens.data?.tokens.length,
    members: members.data?.members.length,
    groups: groups.data?.groups.length,
  }
}
