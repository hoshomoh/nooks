import { QueryClient } from "@tanstack/react-query"
import { describe, expect, it } from "vitest"

import { actOn } from "./live-wiring"
import type { LiveEvent } from "./live-store"

/** invalidatedBy runs one event and answers which query keys were re-read. */
async function invalidatedBy(event: LiveEvent): Promise<string[]> {
  const queryClient = new QueryClient()
  const asked: string[] = []
  queryClient.invalidateQueries = (filters?: { queryKey?: readonly unknown[] }) => {
    asked.push(JSON.stringify(filters?.queryKey ?? []))
    return Promise.resolve()
  }

  await actOn(queryClient, event)
  return asked
}

describe("what a live event re-reads", () => {
  /*
   * Who we are signed in as is read once and held for the life of the tab — staleTime
   * Infinity, no refetch on focus — because it does not change because somebody ticked
   * the milk. So when an Admin does change somebody's role, nothing else in the app
   * would ever notice: a Member promoted to Admin would not see the screens they were
   * just given, and one demoted would go on being offered buttons the Instance refuses,
   * both until they happened to reload.
   */
  it("re-reads who we are when the account changed, and nothing else does", async () => {
    expect(await invalidatedBy({ kind: "member.changed" })).toEqual(['["current-member"]'])

    for (const kind of ["list.changed", "lists.changed", "activity"] as const) {
      expect(await invalidatedBy({ kind }), `${kind} should not re-read the Member`).not.toContain(
        '["current-member"]',
      )
    }
  })

  it("re-reads the activity panel when something arrives in it", async () => {
    expect(await invalidatedBy({ kind: "activity" })).toEqual(['["activity"]'])
  })

  it("reads nothing back for presence, which the store already holds", async () => {
    expect(await invalidatedBy({ kind: "presence" })).toEqual([])
  })

  it("re-reads the lists when one changed", async () => {
    const asked = await invalidatedBy({ kind: "list.changed", listUid: "lst_1" })
    expect(asked).toContain('["lists"]')
    expect(asked).toContain('["sidebar"]')
  })
})
