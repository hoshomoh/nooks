import { Code, ConnectError } from "@connectrpc/connect"
import { describe, expect, it } from "vitest"

import type { ConnectionStore } from "./connection-store"
import { reportReachability } from "./reachability"

/** A store that only records what it was told, so the interceptor is what is tested. */
function recorder() {
  const told: string[] = []
  const store: ConnectionStore = {
    subscribe: () => () => {},
    getState: () => ({ online: true, lastSeenAt: null }),
    markSeen: () => void told.push("seen"),
    markUnreachable: () => void told.push("unreachable"),
  }
  return { store, told }
}

/** run puts one request through the interceptor, with the given outcome. */
async function run(store: ConnectionStore, outcome: () => Promise<unknown>): Promise<unknown> {
  const interceptor = reportReachability(store)
  // The interceptor only ever awaits what it is given and reads the error, so the
  // request itself can be anything.
  const call = interceptor(outcome as never)
  return call({} as never)
}

describe("what a request reports about the connection", () => {
  it("counts an answer as contact", async () => {
    const { store, told } = recorder()

    await run(store, () => Promise.resolve({ ok: true }))

    expect(told).toEqual(["seen"])
  })

  it("counts a refusal as contact", async () => {
    const { store, told } = recorder()
    const refused = new ConnectError("only an admin can do that", Code.PermissionDenied)

    await expect(run(store, () => Promise.reject(refused))).rejects.toThrow(refused)

    // The Instance answering "no" is the Instance answering.
    expect(told).toEqual(["seen"])
  })

  it("counts a request that never arrived as lost", async () => {
    const { store, told } = recorder()
    const gone = new ConnectError("failed to fetch", Code.Unavailable)

    await expect(run(store, () => Promise.reject(gone))).rejects.toThrow(gone)

    expect(told).toEqual(["unreachable"])
  })

  it("counts a failed fetch as lost", async () => {
    const { store, told } = recorder()

    await expect(run(store, () => Promise.reject(new TypeError("Failed to fetch")))).rejects.toThrow(
      TypeError,
    )

    expect(told).toEqual(["unreachable"])
  })
})
