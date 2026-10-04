import { describe, expect, it } from "vitest"
import { Code, ConnectError } from "@connectrpc/connect"

import { worthAskingAgain } from "./retry"

describe("whether a failed read is asked again", () => {
  // These were retried twice with a pause between, so a List deleted while somebody
  // looked at it took seconds to say so.
  it("never asks again for an answer that cannot change", () => {
    for (const code of [Code.NotFound, Code.PermissionDenied, Code.Unauthenticated]) {
      expect(worthAskingAgain(0, new ConnectError("", code)), Code[code]).toBe(false)
    }
  })

  it("asks again when the Instance was busy or out of reach", () => {
    expect(worthAskingAgain(0, new ConnectError("", Code.Unavailable))).toBe(true)
    expect(worthAskingAgain(0, new TypeError("Failed to fetch"))).toBe(true)
  })

  it("stops after two more tries", () => {
    const busy = new ConnectError("", Code.Unavailable)
    expect(worthAskingAgain(1, busy)).toBe(true)
    expect(worthAskingAgain(2, busy)).toBe(false)
  })
})
