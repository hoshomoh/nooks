import { describe, expect, it } from "vitest"

import { copyText, type CopyWays } from "./copy-text"

/** ways builds the two routes to the clipboard, each recording whether it was used. */
function ways(clipboard: "works" | "refuses" | "missing", bySelecting: boolean) {
  const used: string[] = []
  const built: CopyWays = {
    clipboard:
      clipboard === "missing"
        ? undefined
        : {
            writeText: async () => {
              used.push("clipboard")
              if (clipboard === "refuses") {
                throw new DOMException("Document is not focused.", "NotAllowedError")
              }
            },
          },
    bySelecting: () => {
      used.push("bySelecting")
      return bySelecting
    },
  }
  return { built, used }
}

describe("copying text", () => {
  it("uses the clipboard where the browser offers it", async () => {
    const { built, used } = ways("works", true)
    expect(await copyText("hunter2", built)).toBe(true)
    expect(used).toEqual(["clipboard"])
  })

  // A plain HTTP page at a LAN address has no clipboard API, which is where the Copy
  // beside a temporary password did nothing at all.
  it("still copies on a page with no clipboard API", async () => {
    const { built, used } = ways("missing", true)
    expect(await copyText("hunter2", built)).toBe(true)
    expect(used).toEqual(["bySelecting"])
  })

  it("tries the other way when the clipboard refuses", async () => {
    const { built, used } = ways("refuses", true)
    expect(await copyText("hunter2", built)).toBe(true)
    expect(used).toEqual(["clipboard", "bySelecting"])
  })

  it("says so when nothing worked", async () => {
    const { built } = ways("missing", false)
    expect(await copyText("hunter2", built)).toBe(false)
  })
})
