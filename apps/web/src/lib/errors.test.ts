import { describe, expect, it } from "vitest"
import { ConnectError, Code } from "@connectrpc/connect"

import { messageFrom } from "./errors"

describe("reading an error for a Member", () => {
  // A form that has not been submitted has nothing to apologise for. Returning the
  // fallback here put "something went wrong" on every sign-in page before it was used.
  it("says nothing when there is no error", () => {
    expect(messageFrom(null)).toBe("")
    expect(messageFrom(undefined)).toBe("")
  })

  // The server writes in the Instance's own words, so its message is the one to show.
  it("uses what the server said", () => {
    const failure = new ConnectError("that is not your current password", Code.InvalidArgument)

    expect(messageFrom(failure)).toBe("that is not your current password")
  })

  it("falls back only when something unrecognisable arrives", () => {
    expect(messageFrom("a string nobody expected")).toContain("went wrong")
  })
})
