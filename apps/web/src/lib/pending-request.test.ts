import { describe, expect, it } from "vitest"

import {
  forgetPendingRequest,
  readPendingRequest,
  rememberPendingRequest,
} from "./pending-request"

/** held is a browser's storage, standing in for the one the screens reach for. */
function held(initial: Record<string, string> = {}) {
  const store = new Map(Object.entries(initial))
  return {
    getItem: (key: string) => store.get(key) ?? null,
    setItem: (key: string, value: string) => void store.set(key, value),
    removeItem: (key: string) => void store.delete(key),
  }
}

/**
 * The identifier a browser remembers is the thing that finishes the request.
 *
 * nooks sends no mail, so there is no link to come back on: this identifier is what
 * turns an approval into an account, or into a new password. It is kept on disk because
 * an Admin may take days to answer.
 *
 * Which is why "ask again" has to end it. It used to clear the screen's own state and
 * leave the identifier where it was, so somebody who said they were starting over came
 * back to the old request on their next visit, and the old one was still live.
 */
describe("a remembered request", () => {
  it("is the one the browser comes back to", () => {
    const storage = held()
    rememberPendingRequest("join", "req_one", storage)

    expect(readPendingRequest("join", storage)).toBe("req_one")
  })

  it("is gone once it is forgotten", () => {
    const storage = held()
    rememberPendingRequest("join", "req_one", storage)
    forgetPendingRequest("join", storage)

    expect(readPendingRequest("join", storage)).toBeNull()
  })

  // Two kinds, two keys: asking to join again does not end a password reset.
  it("keeps the two kinds apart", () => {
    const storage = held()
    rememberPendingRequest("join", "req_join", storage)
    rememberPendingRequest("reset", "req_reset", storage)

    forgetPendingRequest("join", storage)

    expect(readPendingRequest("join", storage)).toBeNull()
    expect(readPendingRequest("reset", storage)).toBe("req_reset")
  })

  // A browser that refuses storage can still ask; it simply cannot check back.
  it("does not throw where there is nowhere to remember", () => {
    const refused = {
      getItem: () => {
        throw new Error("denied")
      },
      setItem: () => {
        throw new Error("denied")
      },
      removeItem: () => {
        throw new Error("denied")
      },
    }

    expect(() => rememberPendingRequest("join", "req_one", refused)).not.toThrow()
    expect(readPendingRequest("join", refused)).toBeNull()
    expect(() => forgetPendingRequest("join", refused)).not.toThrow()
  })
})
