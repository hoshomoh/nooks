import { readFileSync } from "node:fs"
import { fileURLToPath } from "node:url"
import { describe, expect, it } from "vitest"

/*
The window is a frame; the content inside it scrolls.

The app used to be one long document: the shell was `min-h-dvh` and grew with the List,
and everything anchored to it grew too. The sidebar slid away as you scrolled, and the
side sheet — absolutely positioned inside a container as tall as the List — was as tall
as the List and travelled with it, so its own scroller never had anything to do.

DESIGN.md draws the app as a fixed frame with a 258px column beside the content, which
is the same thing said in CSS: the shell is the height of the window, and the middle
column is the only part that scrolls.
*/

/** source reads one of our components. */
function source(name: string): string {
  return readFileSync(fileURLToPath(new URL(name, import.meta.url)), "utf8")
}

describe("what scrolls", () => {
  it("gives the shell the window's height and no scroll of its own", () => {
    const shell = source("./app-shell.tsx")
    expect(shell).toMatch(/h-dvh overflow-hidden/)
    expect(shell).not.toMatch(/min-h-dvh/)
  })

  it("lets the content column be shorter than its content", () => {
    // Without min-h-0 a flex child refuses to shrink below its content, and the frame
    // grows instead of the column scrolling — the failure looks like nothing happened.
    expect(source("./app-shell.tsx")).toMatch(/<main className="flex min-h-0 min-w-0 flex-1/)
    expect(source("./settings-shell.tsx")).toMatch(/<main className="flex min-h-0 min-w-0 flex-1/)
  })

  it("keeps both sidebars full height with a scroller of their own", () => {
    expect(source("./sidebar.tsx")).toMatch(/h-full[^"]*overflow-y-auto/)
    expect(source("./settings-shell.tsx")).toMatch(/h-full[^"]*overflow-y-auto/)
  })

  it("keeps the content clear of the sheet rather than under it", () => {
    /*
     * DESIGN.md §5: with the sheet open the pane's right padding becomes 544px.
     *
     * Without it a 520px panel is simply drawn over the rows. It reads as ordinary
     * truncation, which is why it survived so long — the row is still there, it is
     * just underneath. It only stops overlapping above about 1958px of window, so
     * every laptop had it.
     */
    const list = source("../../screens/list.tsx")
    expect(list).toMatch(/sheetOpen\s*\?\s*"pr-sheet-clear/)
    expect(list).toMatch(/:\s*"pr-5\.5 /)

    const tokens = readFileSync(
      fileURLToPath(new URL("../../../../../packages/design/foundations.css", import.meta.url)),
      "utf8",
    )
    expect(tokens).toMatch(/--spacing-sheet-clear:\s*544px/)
  })

  // The gap and the sheet are one movement. A pane that resizes after the sheet has
  // gone reads as two things happening.
  it("opens and closes the gap on the sheet's own clock", () => {
    const list = source("../../screens/list.tsx")
    expect(list).toMatch(/transition-\[padding-right\][^"]*ease-sheet/)
    expect(list).toMatch(/duration-\(--duration-sheet-in\)/)
    expect(list).toMatch(/duration-\(--duration-sheet-out\)/)

    const tokens = readFileSync(
      fileURLToPath(new URL("../../../../../packages/design/foundations.css", import.meta.url)),
      "utf8",
    )
    // The sheet's own animation reads the same two numbers, so they cannot drift.
    expect(tokens).toMatch(/--animate-sheet-in:[^;]*var\(--duration-sheet-in\)/)
    expect(tokens).toMatch(/--animate-sheet-out:[^;]*var\(--duration-sheet-out\)/)
  })

  it("anchors the side sheet to something that does not grow", () => {
    // The sheet is absolute. Its containing block has to be the frame, not the column
    // that scrolls — an absolute child of a scroller scrolls with it.
    const list = source("../../screens/list.tsx")
    expect(list).toMatch(/<div className="relative flex min-h-0 flex-1">/)
    expect(list).not.toMatch(/relative[^"]*overflow-y-auto/)
  })



  /*
   * There is room under the last thing on a screen that scrolls.
   *
   * Two traps, and the app fell into both. Padding on the scroller does nothing,
   * because a scrolling flex container drops the padding on its end edge. And padding
   * on the column inside does nothing either while that column is a stretched flex
   * item: it is sized to the container rather than to its contents, so its padding
   * sits above where the content actually ends.
   *
   * The column has to size itself — self-start — and carry the room.
   */
  it("leaves room under the last thing on every screen that scrolls", () => {
    const screens: Array<[string, string]> = [
      ["../../screens/list.tsx", "list"],
      ["../../screens/home.tsx", "all lists"],
      ["../../screens/today.tsx", "today"],
      ["../../screens/upcoming.tsx", "upcoming"],
      ["../../screens/calendar.tsx", "calendar"],
      ["../../screens/note.tsx", "the full note"],
    ]

    for (const [file, what] of screens) {
      const text = source(file)
      const scroller = /overflow-y-auto[^"]*/.exec(text)?.[0] ?? ""
      expect(scroller, `${what} pads its scroller, which does nothing`).not.toMatch(/\bpb-\d/)
      expect(text, `${what} has no column that sizes itself and holds the room`).toMatch(
        /self-start[^"]*\bpb-\d|\bpb-\d[^"]*self-start/,
      )
    }

    // The sheet scrolls a column with no single wrapper, so its room is a spacer.
    expect(source("./note-sheet.tsx")).toMatch(/after:h-\d/)
  })
})
