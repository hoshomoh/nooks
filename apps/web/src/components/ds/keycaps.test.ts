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
