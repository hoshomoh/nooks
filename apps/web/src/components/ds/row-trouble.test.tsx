/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import "@/test/dom"
import { readyForEnglish } from "@/test/i18n"
import { RowTrouble, type RowTroubleLabels } from "./row-trouble"
import type { RowTrouble as Trouble } from "@/lib/row-trouble"

beforeAll(readyForEnglish)

const labels: RowTroubleLabels = {
  conflict: "Somebody else renamed this",
  mine: "Yours",
  theirs: "On the list now",
  keepMine: "Keep mine",
  keepBoth: "Keep both",
  failed: "That did not save",
  kept: "What you wrote is still here:",
  tryAgain: "Try again",
  discard: "Discard mine",
}

function show(trouble: Trouble, handlers: Partial<Record<string, () => void>> = {}) {
  const noop = () => {}
  return render(
    <RowTrouble
      trouble={trouble}
      theirs="Oat milk"
      onKeepMine={handlers.keepMine ?? noop}
      onKeepBoth={handlers.keepBoth ?? noop}
      onTryAgain={handlers.tryAgain ?? noop}
      onDiscard={handlers.discard ?? noop}
      labels={labels}
    />,
  )
}

const CONFLICT: Trouble = { kind: "conflict", itemUid: "item_milk", mine: "Whole milk" }
const FAILED: Trouble = {
  kind: "error",
  itemUid: "item_milk",
  mine: "Whole milk",
  said: "only an admin can do that",
}

describe("a competing rename", () => {
  it("shows both versions, so a person can choose", () => {
    show(CONFLICT)

    expect(screen.getByText("Oat milk")).toBeInTheDocument()
    expect(screen.getByText("Whole milk")).toBeInTheDocument()
  })

  it("offers keeping either, or both", async () => {
    const chose: string[] = []
    show(CONFLICT, {
      keepMine: () => chose.push("mine"),
      keepBoth: () => chose.push("both"),
    })

    await userEvent.click(screen.getByRole("button", { name: "Keep mine" }))
    await userEvent.click(screen.getByRole("button", { name: "Keep both" }))

    expect(chose).toEqual(["mine", "both"])
  })

  it("says so out loud, because the row changed under the Member", () => {
    show(CONFLICT)

    expect(screen.getByRole("alert")).toBeInTheDocument()
  })
})

describe("a rename the Instance refused", () => {
  it("says what the Instance said", () => {
    show(FAILED)

    expect(screen.getByText("only an admin can do that")).toBeInTheDocument()
  })

  it("keeps the Member's text on screen", () => {
    // The rule that matters: their words are never thrown away while they are being
    // told about it.
    show(FAILED)

    expect(screen.getByText(/Whole milk/)).toBeInTheDocument()
  })

  it("offers the change again", async () => {
    const again = vi.fn()
    show(FAILED, { tryAgain: again })

    await userEvent.click(screen.getByRole("button", { name: "Try again" }))

    expect(again).toHaveBeenCalled()
  })

  it("only discards when the Member says so", async () => {
    const discarded = vi.fn()
    show(FAILED, { discard: discarded })

    expect(discarded).not.toHaveBeenCalled()
    await userEvent.click(screen.getByRole("button", { name: "Discard mine" }))
    expect(discarded).toHaveBeenCalled()
  })
})
