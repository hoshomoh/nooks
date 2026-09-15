import { beforeAll, describe, expect, it } from "vitest"
import { render, screen } from "@testing-library/react"

import { readyForEnglish } from "@/test/i18n"
import { ChromeBar } from "./chrome-bar"

beforeAll(readyForEnglish)

const LONG = "TripZapp — Checkout, Success & Homepage redesign"

describe("a long name in the chrome bar", () => {
  it("truncates the name rather than squeezing the controls", () => {
    // Flex items shrink by default, so the button was what gave way: a long list name
    // squeezed "Open full" until its two words stacked inside a 44px bar.
    render(
      <ChromeBar crumbs={[LONG, "Item"]} actions={<button type="button">Open full</button>} />,
    )

    const crumbs = screen.getByText(LONG).parentElement
    expect(crumbs).toHaveClass("truncate")
    expect(crumbs).toHaveClass("min-w-0")
  })

  it("keeps the controls at their own size", () => {
    render(
      <ChromeBar crumbs={[LONG]} actions={<button type="button">Open full</button>} />,
    )

    expect(screen.getByRole("button", { name: "Open full" }).parentElement).toHaveClass("shrink-0")
  })

  it("adds nothing when there is nothing to add", () => {
    const { container } = render(<ChromeBar crumbs={["Groceries"]} />)

    expect(container.querySelectorAll("span.shrink-0")).toHaveLength(0)
  })
})
