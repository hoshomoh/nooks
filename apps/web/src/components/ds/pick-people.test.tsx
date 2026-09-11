/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import type { Member } from "@nooks/api"

import "@/test/dom"
import { readyForEnglish } from "@/test/i18n"
import { PickPeople } from "./pick-people"

beforeAll(readyForEnglish)

/** everyone is the household the dialog picks from. */
const everyone = [
  { uid: "mem_anna", name: "Anna", email: "anna@brunnen.lan" },
  { uid: "mem_jonas", name: "Jonas", email: "jonas@brunnen.lan" },
  { uid: "mem_mira", name: "Mira", email: "mira@brunnen.lan" },
] as Member[]

/** show renders the dialog with some people already in. */
function show(picked: string[]) {
  const onConfirm = vi.fn()
  render(
    <PickPeople
      open
      onOpenChange={vi.fn()}
      title="Who is in Flatmates?"
      blurb="Taking somebody out takes away the lists they reached through this group."
      members={everyone}
      picked={picked}
      confirmLabel="Save"
      onConfirm={onConfirm}
    />,
  )
  return onConfirm
}

describe("choosing who is in a group", () => {
  it("starts from who is in it now", async () => {
    show(["mem_jonas"])

    expect(await screen.findByRole("checkbox", { name: /Jonas/ })).toBeChecked()
    expect(screen.getByRole("checkbox", { name: /Anna/ })).not.toBeChecked()
  })

  // Membership is one decision: what the dialog sends replaces what was there.
  it("sends everyone who ends up picked, not what changed", async () => {
    const onConfirm = show(["mem_jonas"])

    await userEvent.click(await screen.findByRole("checkbox", { name: /Anna/ }))
    await userEvent.click(screen.getByRole("button", { name: "Save" }))

    expect(onConfirm).toHaveBeenCalledWith(["mem_jonas", "mem_anna"])
  })

  it("takes somebody out", async () => {
    const onConfirm = show(["mem_jonas", "mem_anna"])

    await userEvent.click(await screen.findByRole("checkbox", { name: /Jonas/ }))
    await userEvent.click(screen.getByRole("button", { name: "Save" }))

    expect(onConfirm).toHaveBeenCalledWith(["mem_anna"])
  })

  // Nothing happens until Save: a dialog somebody closes changes nothing.
  it("changes nothing when it is cancelled", async () => {
    const onConfirm = show(["mem_jonas"])

    await userEvent.click(await screen.findByRole("checkbox", { name: /Anna/ }))
    await userEvent.click(screen.getByRole("button", { name: "Cancel" }))

    expect(onConfirm).not.toHaveBeenCalled()
  })
})
