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

  it("anchors the side sheet to something that does not grow", () => {
    // The sheet is absolute. Its containing block has to be the frame, not the column
    // that scrolls — an absolute child of a scroller scrolls with it.
    const list = source("../../screens/list.tsx")
    expect(list).toMatch(/<div className="relative flex min-h-0 flex-1">/)
    expect(list).not.toMatch(/relative[^"]*overflow-y-auto/)
  })
})
