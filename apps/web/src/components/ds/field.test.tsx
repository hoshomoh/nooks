import { beforeAll, describe, expect, it } from "vitest"
import { render, screen } from "@testing-library/react"

import { readyForEnglish } from "@/test/i18n"
import { Field } from "./field"

beforeAll(readyForEnglish)

/**
 * A bounded field says how much room is left, and only when that is news.
 *
 * The Instance refuses anything longer either way, so this is about when somebody finds
 * out: while they are typing, or after they press save and lose what they wrote past the
 * end. A counter under every field all the time is the other failure, which is a form
 * that looks like a form somebody has to be careful with.
 */
describe("a field with a limit", () => {
  it("says nothing while the end is far off", () => {
    render(<Field label="Name" limit={100} value="Anna" onChange={() => {}} />)

    expect(screen.queryByText(/characters? left/)).not.toBeInTheDocument()
  })

  it("counts down as the end comes close", () => {
    render(<Field label="Name" limit={10} value="Anna" onChange={() => {}} />)

    expect(screen.getByText("6 characters left")).toBeInTheDocument()
  })

  // One is a character, not characters.
  it("counts the last one properly", () => {
    render(<Field label="Name" limit={5} value="Anna" onChange={() => {}} />)

    expect(screen.getByText("1 character left")).toBeInTheDocument()
  })

  /*
   * A value already past the limit can only be one stored before the limit existed,
   * because typing is stopped at it. "Minus twelve characters left" is not a sentence.
   */
  it("never counts below nothing", () => {
    render(<Field label="Name" limit={3} value="Annaliese" onChange={() => {}} />)

    expect(screen.getByText("0 characters left")).toBeInTheDocument()
  })

  // Counted the way the Instance counts, which is characters rather than bytes.
  it("counts a character that takes two bytes as one", () => {
    render(<Field label="Name" limit={10} value="Jonas 🧦" onChange={() => {}} />)

    expect(screen.getByText("3 characters left")).toBeInTheDocument()
  })

  // Saying it and not enforcing it is a suggestion; enforcing it and not saying it looks
  // like a broken field.
  it("stops the typing as well as counting it", () => {
    render(<Field label="Name" limit={10} value="Anna" onChange={() => {}} />)

    expect(screen.getByLabelText("Name")).toHaveAttribute("maxlength", "10")
  })

  // A count nothing points at is a count a screen reader never reaches.
  it("is part of what describes the field", () => {
    render(<Field label="Name" limit={10} value="Anna" onChange={() => {}} />)

    expect(screen.getByLabelText("Name")).toHaveAccessibleDescription(/6 characters left/)
  })
})
