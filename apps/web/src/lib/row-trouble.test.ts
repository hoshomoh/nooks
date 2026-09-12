import { Code, ConnectError } from "@connectrpc/connect"
import { describe, expect, it } from "vitest"

import { bothOf, troubleFrom } from "./row-trouble"

describe("reading a rename that did not land", () => {
  it("calls a refused write a conflict", () => {
    // ABORTED is what the Instance uses for competing text — see textUnchanged in
    // item.go. It is the one failure that is a question rather than an error.
    const trouble = troubleFrom(
      "item_milk",
      "Whole milk",
      new ConnectError("somebody else renamed this", Code.Aborted),
    )

    expect(trouble).toEqual({ kind: "conflict", itemUid: "item_milk", mine: "Whole milk" })
  })

  it("calls anything else an error, and says what was said", () => {
    const trouble = troubleFrom(
      "item_milk",
      "Whole milk",
      new ConnectError("only an admin can do that", Code.PermissionDenied),
    )

    expect(trouble).toMatchObject({ kind: "error", said: "only an admin can do that" })
  })

  it("keeps the Member's text whichever it was", () => {
    // The one rule both halves share: what somebody wrote is never thrown away on
    // their behalf.
    for (const error of [
      new ConnectError("gone", Code.Aborted),
      new ConnectError("no", Code.PermissionDenied),
      new Error("the network"),
      "something unreadable",
    ]) {
      expect(troubleFrom("item_milk", "Whole milk", error).mine).toBe("Whole milk")
    }
  })

  it("survives something that is not an error at all", () => {
    expect(troubleFrom("item_milk", "Whole milk", undefined)).toMatchObject({ kind: "error" })
  })
})

describe("keeping both", () => {
  it("puts what is already on the list first", () => {
    // Theirs is what everybody else has seen. Nothing is merged word by word: a
    // machine guessing at somebody's sentence is worse than showing them both.
    expect(bothOf("Oat milk", "Whole milk")).toBe("Oat milk / Whole milk")
  })

  it("does not repeat a version that says the same thing", () => {
    expect(bothOf("Milk", "Milk")).toBe("Milk")
    expect(bothOf("Milk", " Milk ")).toBe("Milk")
  })

  it("answers the Member's own when there is nothing to keep", () => {
    expect(bothOf("", "Whole milk")).toBe("Whole milk")
    expect(bothOf("   ", "Whole milk")).toBe("Whole milk")
  })
})
