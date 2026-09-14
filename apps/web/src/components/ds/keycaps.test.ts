import { readFileSync, readdirSync } from "node:fs"
import { fileURLToPath } from "node:url"
import { describe, expect, it } from "vitest"

/*
CommandShortcut draws keycaps, and only keycaps.

It is a generated component, and what it does is `ml-auto`, 12px and 0.1em of tracking —
right for ⌘K and ↵, wrong for anything made of words. Given a list name it letter-spaces
it and lets it wrap, which makes one row taller than its neighbours and leaves a column
of names that no longer lines up. That is what happened to the search results.

DESIGN.md §3 says what the alternatives are: `text-micro` for attribution and
breadcrumbs, `font-mono text-badge` for counts. Neither is this.
*/

const DS = fileURLToPath(new URL(".", import.meta.url))

/** Every component that hands CommandShortcut something to draw. */
function users(): { file: string; content: string }[] {
  return readdirSync(DS)
    .filter((name) => name.endsWith(".tsx") && !name.endsWith(".test.tsx"))
    .map((name) => ({ file: name, content: readFileSync(DS + name, "utf8") }))
    .filter(({ content }) => /<CommandShortcut>/.test(content))
}

describe("telling one row from another", () => {
  it("gives every row in the palette an identity that is not its words", () => {
    /*
     * cmdk identifies a row by the words in it unless it is told otherwise.
     *
     * A search that matched an Item's label and its Note produced two rows reading the
     * same thing, and cmdk treated them as one: hovering either highlighted both. The
     * duplicate itself is fixed in the service, which now returns one hit per Item —
     * but two Lists may still share a name, so every row says what it is about.
     *
     * Checked here rather than by rendering, because the collision needs two rows with
     * identical words and the palette's rows come from the server.
     */
    const palette = readFileSync(DS + "command-palette.tsx", "utf8")

    const rows = [...palette.matchAll(/<CommandItem\b([\s\S]*?)>/g)].map((match) => match[1] ?? "")
    expect(rows.length).toBeGreaterThan(0)
    expect(rows.filter((row) => !/\bvalue=/.test(row))).toEqual([])
  })
})

describe("what a keycap is for", () => {
  it("is never handed a value that is words or a count", () => {
    const wrong = users().flatMap(({ file, content }) =>
      [...content.matchAll(/<CommandShortcut>\{([^}]+)\}/g)].map(
        (match) => `${file}: ${match[1]}`,
      ),
    )

    // A literal keycap — <CommandShortcut>⌘K</CommandShortcut> — is what it is for, and
    // is not an expression, so it is not caught here.
    expect(wrong).toEqual([])
  })
})
