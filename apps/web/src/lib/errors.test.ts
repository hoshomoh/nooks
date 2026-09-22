import { describe, expect, it } from "vitest"
import { ConnectError, Code } from "@connectrpc/connect"

import { messageFrom } from "./errors"

/** A translator that shows which key was chosen, and what was put in it. */
const t = (key: string, options?: Record<string, unknown>) =>
  options && "ref" in options ? `${key}:${String(options.ref)}` : key

describe("reading an error for a Member", () => {
  // A form that has not been submitted has nothing to apologise for. Returning a message
  // here put "something went wrong" on every sign-in page before it was used.
  it("says nothing when there is no error", () => {
    expect(messageFrom(t, null)).toBe("")
    expect(messageFrom(t, undefined)).toBe("")
  })

  // A refusal the Instance writes on purpose is its own sentence.
  it("uses what the server said when the server said something", () => {
    const failure = new ConnectError("that is not your current password", Code.InvalidArgument)

    expect(messageFrom(t, failure)).toBe("that is not your current password")
  })

  /*
   * A failure inside the Instance carries a kind and a reference instead of a cause.
   *
   * It used to carry the cause, so a table name, the path to the database file, or
   * `dial tcp 10.0.0.5:5432: connection refused` could be drawn on screen, and it was
   * English wherever it was read. What is asserted here is that neither travels: the
   * kind picks a line from the catalogue, and the reference goes into it.
   */
  it("draws a failure inside the instance from the catalogue", () => {
    const failure = new ConnectError("could not read the member (ref 7f3a2c)", Code.Internal)
    failure.metadata.set("nooks-error-kind", "load-failed")
    failure.metadata.set("nooks-error-ref", "7f3a2c")

    expect(messageFrom(t, failure)).toBe("error.loadFailed:7f3a2c")
  })

  it("says a save did not happen when that is what happened", () => {
    const failure = new ConnectError("could not create the list (ref 991b)", Code.Internal)
    failure.metadata.set("nooks-error-kind", "save-failed")
    failure.metadata.set("nooks-error-ref", "991b")

    expect(messageFrom(t, failure)).toBe("error.saveFailed:991b")
  })

  // An Instance newer than this tab can send a kind this one has never heard of.
  it("says less rather than the key when the kind is unknown", () => {
    const failure = new ConnectError("could not do something new (ref aa11)", Code.Internal)
    failure.metadata.set("nooks-error-kind", "something-invented-later")

    expect(messageFrom(t, failure)).toBe("error.internal:")
  })

  it("falls back to the catalogue when something unrecognisable arrives", () => {
    expect(messageFrom(t, "a string nobody expected")).toBe("error.unreachable")
  })
})
