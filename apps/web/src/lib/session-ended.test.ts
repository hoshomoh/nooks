import { describe, expect, it, vi } from "vitest"
import { Code, ConnectError } from "@connectrpc/connect"
import type { Member } from "@nooks/api"

import { CACHE_KEY } from "./cache-store"
import { currentMemberQuery } from "./queries"
import { createQueryClient } from "./query-client"
import { endedSession, forgetSession } from "./session-ended"

/** held is a browser's storage, with whatever was left in it. */
function held(initial: Record<string, string>) {
  const store = new Map(Object.entries(initial))
  return {
    getItem: (key: string) => store.get(key) ?? null,
    setItem: (key: string, value: string) => void store.set(key, value),
    removeItem: (key: string) => void store.delete(key),
    has: (key: string) => store.has(key),
  }
}

const refused = new ConnectError("sign in first", Code.Unauthenticated)

describe("telling a session that ended from any other failure", () => {
  it("is a refusal for want of a session", () => {
    expect(endedSession(refused)).toBe(true)
  })

  // A server that is down or a List that is gone is worth trying again, or saying so.
  it("is not a failure of any other kind", () => {
    expect(endedSession(new ConnectError("gone", Code.NotFound))).toBe(false)
    expect(endedSession(new ConnectError("down", Code.Unavailable))).toBe(false)
    expect(endedSession(new Error("network"))).toBe(false)
  })
})

describe("letting go of a session that ended", () => {
  it("stops believing anybody is signed in, and drops what was read", () => {
    const storage = held({ [CACHE_KEY]: "{}" })
    const queryClient = createQueryClient(() => {})
    queryClient.setQueryData(currentMemberQuery.queryKey, { uid: "m1", name: "Ada" } as Member)
    queryClient.setQueryData(["list", "l1"], { items: [] })

    forgetSession(queryClient, { storage, now: () => 0 })

    expect(queryClient.getQueryData(currentMemberQuery.queryKey)).toBeNull()
    expect(queryClient.getQueryData(["list", "l1"])).toBeUndefined()
    expect(storage.has(CACHE_KEY)).toBe(false)
  })

  // Nothing typed is lost to this: what to do with it is decided by signing back in.
  it("keeps the changes still waiting to be sent", () => {
    const queryClient = createQueryClient(() => {})
    queryClient.getMutationCache().build(queryClient, { mutationKey: ["tick"] })

    forgetSession(queryClient, { storage: held({}), now: () => 0 })

    expect(queryClient.getMutationCache().getAll()).toHaveLength(1)
  })
})

describe("finding out the session ended", () => {
  it("ends it when a read is refused", async () => {
    const ended = vi.fn()
    const queryClient = createQueryClient(ended)
    queryClient.setQueryData(currentMemberQuery.queryKey, { uid: "m1", name: "Ada" } as Member)

    await queryClient.fetchQuery({ queryKey: ["sidebar"], queryFn: () => Promise.reject(refused) }).catch(() => {})

    expect(ended).toHaveBeenCalledOnce()
    expect(queryClient.getQueryData(currentMemberQuery.queryKey)).toBeNull()
  })

  it("ends it when a change is refused", async () => {
    const ended = vi.fn()
    const queryClient = createQueryClient(ended)

    await queryClient
      .getMutationCache()
      .build(queryClient, { mutationFn: () => Promise.reject(refused) })
      .execute(undefined)
      .catch(() => {})

    expect(ended).toHaveBeenCalledOnce()
  })

  // Anything else is shown as itself, not taken for a sign-out.
  it("leaves the session alone when a read fails for another reason", async () => {
    const ended = vi.fn()
    const queryClient = createQueryClient(ended)
    queryClient.setQueryData(currentMemberQuery.queryKey, { uid: "m1", name: "Ada" } as Member)

    await queryClient
      .fetchQuery({ queryKey: ["list", "l9"], queryFn: () => Promise.reject(new ConnectError("gone", Code.NotFound)) })
      .catch(() => {})

    expect(ended).not.toHaveBeenCalled()
    expect(queryClient.getQueryData(currentMemberQuery.queryKey)).not.toBeNull()
  })
})
