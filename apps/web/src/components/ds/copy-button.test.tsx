import { beforeAll, describe, expect, it } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import { readyForEnglish } from "@/test/i18n"
import { CopyButton } from "./copy-button"

beforeAll(readyForEnglish)

describe("the copy button", () => {
  it("says copied once the text is on the clipboard", async () => {
    render(<CopyButton text="hunter2" label="Copy" copy={async () => true} />)
    await userEvent.click(screen.getByRole("button", { name: "Copy" }))
    expect(await screen.findByRole("button", { name: "Copied" })).toBeInTheDocument()
  })

  // It said Copied before the copy happened, so a temporary password that never reached
  // the clipboard looked as if it had.
  it("never says copied when nothing was", async () => {
    render(<CopyButton text="hunter2" label="Copy" copy={async () => false} />)
    await userEvent.click(screen.getByRole("button", { name: "Copy" }))
    expect(await screen.findByRole("button", { name: "Select it to copy" })).toBeInTheDocument()
    expect(screen.queryByRole("button", { name: "Copied" })).not.toBeInTheDocument()
  })
})
