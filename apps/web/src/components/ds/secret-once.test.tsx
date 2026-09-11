/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import "@/test/dom"
import { readyForEnglish } from "@/test/i18n"
import { SecretOnce } from "./secret-once"

beforeAll(readyForEnglish)

describe("a secret shown once", () => {
  it("shows the secret itself, not a masked version of it", () => {
    render(
      <SecretOnce
        secret="cedar-lantern-pebble-thistle"
        title="Til can sign in with this"
        blurb="It cannot be shown again."
        onDone={vi.fn()}
      />,
    )

    expect(screen.getByText("cedar-lantern-pebble-thistle")).toBeInTheDocument()
  })

  // Nothing has gone wrong here, and dressing a normal step as a hazard teaches a
  // Member to ignore the warnings that matter.
  it("says it plainly, without a warning", () => {
    render(
      <SecretOnce
        secret="cedar-lantern-pebble-thistle"
        title="Til can sign in with this"
        blurb="It cannot be shown again — nooks keeps only a hash of it."
        onDone={vi.fn()}
      />,
    )

    expect(screen.getByText(/cannot be shown again/)).toBeInTheDocument()
    expect(screen.queryByText(/⚠|warning|danger/i)).not.toBeInTheDocument()
  })

  it("is finished when the Admin says so", async () => {
    const onDone = vi.fn()
    render(
      <SecretOnce secret="x" title="t" blurb="b" onDone={onDone} />,
    )

    await userEvent.click(screen.getByRole("button", { name: "Done" }))

    expect(onDone).toHaveBeenCalledOnce()
  })
})
