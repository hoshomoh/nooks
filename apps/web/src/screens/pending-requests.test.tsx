import { beforeAll, describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import type { PendingJoinRequest, PendingResetRequest } from "@nooks/api"

import { readyForEnglish } from "@/test/i18n"
import { PendingRequests } from "./pending-requests"

beforeAll(readyForEnglish)

const ASKED_AT = "2026-09-10T18:44:00Z"

/** joining is somebody asking for an account. */
function joining(message = ""): PendingJoinRequest {
  return {
    requestUid: "req_join",
    name: "Til",
    email: "til@example.com",
    message,
    createdAt: ASKED_AT,
  } as PendingJoinRequest
}

/** resetting is a Member who has forgotten their password. */
function resetting(): PendingResetRequest {
  return {
    requestUid: "req_reset",
    member: { uid: "mem_jonas", name: "Jonas", email: "jonas@brunnen.lan" },
    createdAt: ASKED_AT,
  } as PendingResetRequest
}

/** show draws the section and hands back what it reports. */
function show(joins: PendingJoinRequest[], resets: PendingResetRequest[]) {
  const onDecideJoin = vi.fn()
  const onDecideReset = vi.fn()
  render(
    <PendingRequests
      joins={joins}
      resets={resets}
      onDecideJoin={onDecideJoin}
      onDecideReset={onDecideReset}
    />,
  )
  return { onDecideJoin, onDecideReset }
}

describe("who is waiting", () => {
  // Nothing waiting is not an empty section with a heading: it is no section.
  it("is nothing at all when nobody is", () => {
    const { container } = render(
      <PendingRequests joins={[]} resets={[]} onDecideJoin={vi.fn()} onDecideReset={vi.fn()} />,
    )

    expect(container).toBeEmptyDOMElement()
  })

  /*
   * Both kinds, in one section. They are the same job — somebody an Admin has to
   * recognise before letting them in — and a reset used to appear only in the Activity
   * panel, so an Admin who cleared the panel had nowhere left to find one.
   */
  it("counts both kinds together", () => {
    show([joining()], [resetting()])

    expect(screen.getByText("2 waiting")).toBeInTheDocument()
    expect(screen.getByText("Til")).toBeInTheDocument()
    expect(screen.getByText("Jonas")).toBeInTheDocument()
  })

  // nooks sends no mail, so the Admin is the only check there is on a reset.
  it("says to check a reset is really them", () => {
    show([], [resetting()])

    expect(screen.getByText(/check it is really them/)).toBeInTheDocument()
  })

  // An empty message is worth an Admin's suspicion, so its absence is said out loud
  // rather than left as a gap.
  it("says when somebody wrote nothing", () => {
    show([joining()], [])

    expect(screen.getByText(/no message/)).toBeInTheDocument()
  })

  it("shows what somebody did write", () => {
    show([joining("It's Til, from upstairs")], [])

    expect(screen.getByText(/It's Til, from upstairs/)).toBeInTheDocument()
  })

  // The two kinds are answered by two different calls, and sending a reset to the join
  // endpoint would answer somebody else's request or nobody's.
  it("answers each kind through its own call", async () => {
    const { onDecideJoin, onDecideReset } = show([joining()], [resetting()])

    const approve = screen.getAllByRole("button", { name: "Approve" })
    await userEvent.click(approve[0])
    expect(onDecideJoin).toHaveBeenCalledWith("req_join", true)
    expect(onDecideReset).not.toHaveBeenCalled()

    await userEvent.click(approve[1])
    expect(onDecideReset).toHaveBeenCalledWith("req_reset", true)
  })

  // Ignoring is silent: the sender is never told, so there is nothing to confirm.
  it("ignores without asking again", async () => {
    const { onDecideJoin } = show([joining()], [])

    await userEvent.click(screen.getByRole("button", { name: "Ignore" }))

    expect(onDecideJoin).toHaveBeenCalledWith("req_join", false)
  })
})
