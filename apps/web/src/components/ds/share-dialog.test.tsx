/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it, vi } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { Sharing, type List } from "@nooks/api"

import "@/test/dom"
import { readyForEnglish } from "@/test/i18n"
import { ShareDialog, type ShareDecision } from "./share-dialog"

beforeAll(readyForEnglish)

/** groceries is a private List that Anna owns. */
const groceries = {
  uid: "list_groceries",
  name: "Groceries",
  sharing: Sharing.PRIVATE,
  canEdit: true,
  isOwner: true,
  isPinned: false,
  openCount: 7,
} as List

/** openDialog renders the dialog with its queries answered by nothing at all. */
function openDialog(onSave: (decision: ShareDecision) => void) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ShareDialog
        list={groceries}
        instanceName="Brunnen Street"
        open
        onOpenChange={() => {}}
        onSave={onSave}
      />
    </QueryClientProvider>,
  )
}

describe("the share dialog", () => {
  it("offers the three ways a List can be shared", async () => {
    openDialog(vi.fn())

    expect(await screen.findByRole("radio", { name: "Private" })).toBeInTheDocument()
    expect(screen.getByRole("radio", { name: "Everyone at Brunnen Street" })).toBeInTheDocument()
    expect(screen.getByRole("radio", { name: "Specific people or a group" })).toBeInTheDocument()
  })

  it("starts on what the List already is", async () => {
    openDialog(vi.fn())

    expect(await screen.findByRole("radio", { name: "Private" })).toBeChecked()
  })

  // Nothing happens until Save: a dialog a Member closes changes nothing.
  it("saves the choice that was made", async () => {
    const onSave = vi.fn()
    openDialog(onSave)

    await userEvent.click(
      await screen.findByRole("radio", { name: "Everyone at Brunnen Street" }),
    )
    await userEvent.click(screen.getByRole("button", { name: "Save" }))

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({ sharing: Sharing.INSTANCE, canEdit: true }),
    )
  })

  it("carries the read-only decision with it", async () => {
    const onSave = vi.fn()
    openDialog(onSave)

    await userEvent.click(await screen.findByRole("switch", { name: "They can edit" }))
    await userEvent.click(screen.getByRole("button", { name: "Save" }))

    expect(onSave).toHaveBeenCalledWith(expect.objectContaining({ canEdit: false }))
  })

  // Picking "specific people" is a step, not a longer first step.
  it("opens the people step when specific people are chosen", async () => {
    openDialog(vi.fn())

    await userEvent.click(
      await screen.findByRole("radio", { name: "Specific people or a group" }),
    )

    expect(await screen.findByText("Specific people")).toBeInTheDocument()
  })

  it("sends no names when the List is not shared by name", async () => {
    const onSave = vi.fn()
    openDialog(onSave)

    await userEvent.click(await screen.findByRole("button", { name: "Save" }))

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({ memberUids: [], groupUids: [] }),
    )
  })
})
