/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import "@/test/dom"
import { readyForEnglish } from "@/test/i18n"
import { CommandPalette } from "./command-palette"
import { commandPaletteStore } from "@/lib/command-palette-store"

beforeAll(readyForEnglish)

/** openPalette renders the palette with its queries answered by nothing at all. */
function openPalette() {
  commandPaletteStore.open()
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <CommandPalette />
    </QueryClientProvider>,
  )
}

describe("⌘K", () => {
  it("offers the places a Member can go, not only a search box", async () => {
    openPalette()

    expect(await screen.findByRole("option", { name: /Today/ })).toBeInTheDocument()
    expect(screen.getByRole("option", { name: /Upcoming/ })).toBeInTheDocument()
    expect(screen.getByRole("option", { name: /Calendar/ })).toBeInTheDocument()
    expect(screen.getByRole("option", { name: /All lists/ })).toBeInTheDocument()
  })

  it("narrows the destinations as the Member types", async () => {
    openPalette()

    await userEvent.type(await screen.findByRole("combobox"), "upc")

    expect(screen.getByRole("option", { name: /Upcoming/ })).toBeInTheDocument()
    expect(screen.queryByRole("option", { name: /^Today/ })).not.toBeInTheDocument()
  })

  it("offers adding a list without leaving the palette", async () => {
    openPalette()

    await userEvent.click(await screen.findByRole("option", { name: /Add a list/ }))

    expect(await screen.findByLabelText("Name")).toBeInTheDocument()
  })
})
