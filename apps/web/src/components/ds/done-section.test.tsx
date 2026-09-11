/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import "@/test/dom"
import { readyForEnglish } from "@/test/i18n"
import { DoneSection } from "./done-section"

beforeAll(readyForEnglish)

describe("the completed items", () => {
  // A List that doubles in length as the week goes on is one nobody scrolls through.
  it("starts collapsed, behind a count", () => {
    render(
      <DoneSection label="3 done today">
        <span>Washing-up liquid</span>
      </DoneSection>,
    )

    expect(screen.getByRole("button", { name: /3 done today/ })).toHaveAttribute(
      "aria-expanded",
      "false",
    )
    expect(screen.queryByText("Washing-up liquid")).not.toBeInTheDocument()
  })

  it("opens when the Member asks for it", async () => {
    render(
      <DoneSection label="3 done today">
        <span>Washing-up liquid</span>
      </DoneSection>,
    )

    await userEvent.click(screen.getByRole("button", { name: /3 done today/ }))

    expect(screen.getByText("Washing-up liquid")).toBeInTheDocument()
  })
})
