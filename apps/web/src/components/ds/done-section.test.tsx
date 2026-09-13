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

  it("opens by pushing the foot of the List down rather than by appearing", async () => {
    const { container } = render(
      <DoneSection label="3 done today">
        <span>Washing-up liquid</span>
      </DoneSection>,
    )

    const opening = container.querySelector(".transition-\\[grid-template-rows\\]")
    expect(opening).toHaveClass("grid-rows-[0fr]")

    await userEvent.click(screen.getByRole("button", { name: /3 done today/ }))

    expect(opening).toHaveClass("grid-rows-[1fr]")
  })

  it("keeps what it has closed over out of the keyboard's way", async () => {
    render(
      <DoneSection label="3 done today">
        <button type="button">Washing-up liquid</button>
      </DoneSection>,
    )

    const toggle = screen.getByRole("button", { name: /3 done today/ })
    await userEvent.click(toggle)
    await userEvent.click(toggle)

    // Built once and kept, so the second opening has nothing left to build — which
    // means a row nobody can see is still in the page, and must not be reachable.
    expect(screen.getByText("Washing-up liquid").closest("[inert]")).not.toBeNull()
  })

  describe("aiming at it", () => {
    it("is the whole line, not the words on it", () => {
      // Sized to its label it was a target somebody had to find, sitting under rows
      // that are each clickable across their full width.
      render(
        <DoneSection label="3 done today">
          <span>Washing-up liquid</span>
        </DoneSection>,
      )

      expect(screen.getByRole("button", { name: /3 done today/ })).not.toHaveClass("self-start")
    })

    it("says it can be pressed before it is", () => {
      render(
        <DoneSection label="3 done today">
          <span>Washing-up liquid</span>
        </DoneSection>,
      )

      expect(screen.getByRole("button", { name: /3 done today/ })).toHaveClass("hover:bg-secondary")
    })

    it("is as tall as the rows above it", () => {
      render(
        <DoneSection label="3 done today">
          <span>Washing-up liquid</span>
        </DoneSection>,
      )

      expect(screen.getByRole("button", { name: /3 done today/ })).toHaveClass("min-h-row")
    })
  })
})
