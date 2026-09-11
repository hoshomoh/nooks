/** @vitest-environment jsdom */
import { describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import "@/test/dom"
import { Segmented } from "./segmented"

const THEMES = [
  { value: "light", label: "Light" },
  { value: "dark", label: "Dark" },
  { value: "system", label: "System" },
]

describe("a segmented control", () => {
  it("says which answer is on", () => {
    render(<Segmented label="Theme" options={THEMES} chosen="system" onChoose={vi.fn()} />)

    expect(screen.getByRole("radio", { name: "System" })).toBeChecked()
    expect(screen.getByRole("radio", { name: "Light" })).not.toBeChecked()
  })

  it("reports the answer that was pressed", async () => {
    const onChoose = vi.fn()
    render(<Segmented label="Theme" options={THEMES} chosen="system" onChoose={onChoose} />)

    await userEvent.click(screen.getByRole("radio", { name: "Dark" }))

    expect(onChoose).toHaveBeenCalledWith("dark")
  })

  // The group carries the name, so a screen reader says what the answers are for.
  it("names the group", () => {
    render(<Segmented label="Theme" options={THEMES} chosen="light" onChoose={vi.fn()} />)

    expect(screen.getByRole("radiogroup", { name: "Theme" })).toBeInTheDocument()
  })
})
