import { beforeAll, describe, expect, it, vi } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import { readyForEnglish } from "@/test/i18n"
import { ListPicker, PickOneList } from "./list-picker"

beforeAll(readyForEnglish)

const SUGGESTED = [
  { uid: "list_groceries", name: "Groceries" },
  { uid: "list_bike", name: "Bike" },
]

/** show renders a picker with a query client, which its search needs. */
function show(node: React.ReactNode) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(<QueryClientProvider client={queryClient}>{node}</QueryClientProvider>)
}

describe("picking Lists", () => {
  // The whole point of the change: the rows are a handful to hand, not every List an
  // Instance holds, and the field is how the rest are reached.
  it("offers the Lists it was handed before anything is typed", () => {
    show(<ListPicker suggested={SUGGESTED} chosen={[]} onToggle={vi.fn()} />)

    expect(screen.getByRole("checkbox", { name: "Groceries" })).toBeInTheDocument()
    expect(screen.getByRole("checkbox", { name: "Bike" })).toBeInTheDocument()
  })

  it("says which are picked", () => {
    show(<ListPicker suggested={SUGGESTED} chosen={["list_bike"]} onToggle={vi.fn()} />)

    expect(screen.getByRole("checkbox", { name: "Bike" })).toBeChecked()
    expect(screen.getByRole("checkbox", { name: "Groceries" })).not.toBeChecked()
  })

  it("answers with the List, not only its identifier", async () => {
    const toggle = vi.fn()
    show(<ListPicker suggested={SUGGESTED} chosen={[]} onToggle={toggle} />)

    await userEvent.click(screen.getByRole("checkbox", { name: "Groceries" }))

    expect(toggle).toHaveBeenCalledWith({ uid: "list_groceries", name: "Groceries" })
  })

  it("says so when the Member has no Lists at all", () => {
    show(<ListPicker suggested={[]} chosen={[]} onToggle={vi.fn()} />)

    expect(screen.getByText("You have no lists yet.")).toBeInTheDocument()
  })
})

describe("picking one List, or none", () => {
  it("reads as the one that is picked", () => {
    show(
      <PickOneList
        label="Which list"
        suggested={SUGGESTED}
        picked={{ uid: "list_bike", name: "Bike" }}
        noneLabel="No public page"
        onPick={vi.fn()}
      />,
    )

    expect(screen.getByRole("button", { name: "Which list: Bike" })).toBeInTheDocument()
  })

  it("reads as the answer that is not a List when none is picked", () => {
    show(
      <PickOneList
        label="Which list"
        suggested={SUGGESTED}
        noneLabel="No public page"
        onPick={vi.fn()}
      />,
    )

    expect(screen.getByRole("button", { name: "Which list: No public page" })).toBeInTheDocument()
  })

  it("answers with the List somebody picked, and closes", async () => {
    const pick = vi.fn()
    show(
      <PickOneList
        label="Which list"
        suggested={SUGGESTED}
        noneLabel="No public page"
        onPick={pick}
      />,
    )

    await userEvent.click(screen.getByRole("button", { name: "Which list: No public page" }))
    await userEvent.click(await screen.findByRole("checkbox", { name: "Groceries" }))

    expect(pick).toHaveBeenCalledWith({ uid: "list_groceries", name: "Groceries" })
    expect(screen.queryByRole("checkbox", { name: "Groceries" })).not.toBeInTheDocument()
  })
})
