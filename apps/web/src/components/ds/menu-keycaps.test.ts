import { readFileSync, readdirSync } from "node:fs"
import { fileURLToPath } from "node:url"
import { describe, expect, it } from "vitest"

import { readKeycap } from "@/lib/keycaps"

/*
A keycap a menu prints is one the menu answers.

Menu reads the keys off the items it was given, so a keycap it cannot read is drawn and
then ignored — which is how "⌘D" and "F2" came to sit in a menu for months doing
nothing. Checked against the source because the failure is a keycap that parses to
nothing, and a rendered menu shows it either way.
*/
const DS = fileURLToPath(new URL(".", import.meta.url))

function menus(): { file: string; content: string }[] {
  return readdirSync(DS)
    .filter((name) => name.endsWith(".tsx") && !name.endsWith(".test.tsx"))
    .map((name) => ({ file: name, content: readFileSync(DS + name, "utf8") }))
}

describe("what a menu promises", () => {
  it("prints no keycap it cannot read as a key", () => {
    const unreadable = menus().flatMap(({ file, content }) =>
      [...content.matchAll(/shortcut="([^"]+)"/g)]
        .map((match) => match[1] ?? "")
        .filter((drawn) => readKeycap(drawn) === null)
        .map((drawn) => `${file}: ${drawn}`),
    )

    expect(unreadable).toEqual([])
  })

  /*
   * Enter belongs to the menu, which uses it on whichever entry is highlighted.
   *
   * Binding it to one entry as well means two things happen at once: the entry the
   * Member had chosen, and the one that printed the keycap.
   */
  it("leaves Enter to the menu itself", () => {
    const claimed = menus().flatMap(({ file, content }) =>
      /shortcut="↵"/.test(content) ? [file] : [],
    )

    expect(claimed).toEqual([])
  })
})
