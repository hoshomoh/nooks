import { beforeAll, describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { ActivityKind, ActivityOutcome, type Activity } from "@nooks/api"

import { readyForEnglish } from "@/test/i18n"
import { ActivityPanel } from "./activity-panel"

beforeAll(readyForEnglish)

/** entry builds one thing waiting for attention. */
function entry(kind: ActivityKind, text: string, outcome = ActivityOutcome.UNSPECIFIED): Activity {
  return {
    uid: "act_1",
    kind,
    text,
    targetUid: "req_1",
    createdAt: "2026-09-10T18:44:00Z",
    unread: true,
    outcome,
  } as Activity
}

/** waitingJoins builds a queue of people asking for an account, none of them decided. */
function waitingJoins(howMany: number): Activity[] {
  return Array.from({ length: howMany }, (_, index) => ({
    ...entry(ActivityKind.JOIN_REQUEST, `Asker ${index} asked to join`),
    uid: `join-${index}`,
  }))
}

/** show renders the panel with one entry. */
function show(activity: Activity[], onDecide = vi.fn(), onOpen = vi.fn(), onSeeRequests = vi.fn()) {
  render(
    <ActivityPanel
      activity={activity}
      timeOf={() => "18:44"}
      onOpen={onOpen}
      onDecide={onDecide}
      onSeeRequests={onSeeRequests}
    />,
  )
  return { onDecide, onOpen, onSeeRequests }
}

describe("the activity panel", () => {
  it("says there is no mail server, because there is not", () => {
    show([])
    expect(screen.getByText(/no mail server/)).toBeInTheDocument()
  })

  it("offers both answers to a request", () => {
    show([entry(ActivityKind.JOIN_REQUEST, "Til asked to join")])

    expect(screen.getByRole("button", { name: "Approve" })).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Ignore" })).toBeInTheDocument()
  })

  it("approves a request", async () => {
    const joined = entry(ActivityKind.JOIN_REQUEST, "Til asked to join")
    const { onDecide } = show([joined])

    await userEvent.click(screen.getByRole("button", { name: "Approve" }))

    expect(onDecide).toHaveBeenCalledWith(joined, true)
  })

  // Ignoring is silent: nothing is sent, and there is nothing to confirm.
  it("ignores a request without asking again", async () => {
    const joined = entry(ActivityKind.JOIN_REQUEST, "Til asked to join")
    const { onDecide } = show([joined])

    await userEvent.click(screen.getByRole("button", { name: "Ignore" }))

    expect(onDecide).toHaveBeenCalledWith(joined, false)
  })

  // An Admin who clicks Approve and sees nothing change cannot tell whether it worked.
  it("says what happened once a request has been decided", () => {
    show([entry(ActivityKind.JOIN_REQUEST, "Til asked to join", ActivityOutcome.APPROVED)])

    expect(screen.getByText("Approved")).toBeInTheDocument()
    expect(screen.queryByRole("button", { name: "Approve" })).not.toBeInTheDocument()
  })

  // Ignoring is silent to the sender, not to the Admin who did it.
  it("says so when a request was ignored", () => {
    show([entry(ActivityKind.JOIN_REQUEST, "Til asked to join", ActivityOutcome.IGNORED)])

    expect(screen.getByText("Ignored, silently")).toBeInTheDocument()
  })

  // An entry with nothing to decide is a statement.
  it("gives a shared List somewhere to go and nothing to decide", () => {
    show([entry(ActivityKind.LIST_SHARED, "Jonas shared “Flat jobs” with you")])

    expect(screen.getByRole("button", { name: "Open" })).toBeInTheDocument()
    expect(screen.queryByRole("button", { name: "Approve" })).not.toBeInTheDocument()
  })

  // Three people asking is the ordinary case, and one-click Approve stays where it is.
  it("decides a handful of requests without sending anybody anywhere", () => {
    show(waitingJoins(3))

    expect(screen.getAllByRole("button", { name: "Approve" })).toHaveLength(3)
    expect(screen.queryByRole("button", { name: /more in Members/ })).not.toBeInTheDocument()
  })

  // Past a handful the first three still answer in place and the rest become one line,
  // so somebody with a script cannot bury everything else that was waiting.
  it("sends the rest to the Members screen", async () => {
    const { onSeeRequests } = show(waitingJoins(9))

    expect(screen.getAllByRole("button", { name: "Approve" })).toHaveLength(3)

    await userEvent.click(screen.getByRole("button", { name: "See 6 more in Members" }))
    expect(onSeeRequests).toHaveBeenCalled()
  })

  // Only requests nobody has answered are held back. A decided one is a statement, and
  // counting it would say more people are waiting than are.
  it("counts only what is still waiting", () => {
    show([
      ...waitingJoins(4),
      entry(ActivityKind.JOIN_REQUEST, "Jonas asked to join", ActivityOutcome.APPROVED),
    ])

    expect(screen.getByRole("button", { name: "See 1 more in Members" })).toBeInTheDocument()
  })

  // A shared List is not a request, so it is drawn whatever else is waiting.
  it("keeps everything that is not a request", () => {
    show([...waitingJoins(9), entry(ActivityKind.LIST_SHARED, "Jonas shared a list")])

    expect(screen.getByText("Jonas shared a list")).toBeInTheDocument()
  })
})
