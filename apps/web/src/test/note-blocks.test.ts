import { readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

const DESIGN = "../../DESIGN.md"
const EXTENSIONS = "src/lib/editor/extensions.ts"

/**
 * The editor offers the blocks the design draws, and says how many there are.
 *
 * `extensions.ts` turns every other StarterKit block off rather than styling it, because
 * a block a Member can reach but the design has no drawing for looks wrong the first
 * time somebody uses it. The number is written out twice in that file, and it had
 * already drifted: one comment said "not among the five" directly under a header saying
 * the design names seven.
 *
 * It is written a third time in Go, where the MCP `update_item` tool lists the blocks for
 * an assistant writing a note. A number in prose is a copy of a fact, and a copy drifts.
 */
describe("the blocks a note is made of", () => {
  const rows = blockRowsInDesign()

  it("is the table in DESIGN.md that says how many", () => {
    // A floor, so a table that stopped being found does not quietly agree with anything.
    expect(rows.length, "found no block table in DESIGN.md §10").toBeGreaterThan(4)
  })

  it("is the same number the editor says, everywhere it says it", () => {
    const counted = numberWord(rows.length)
    const said = countsSaidIn(readFileSync(EXTENSIONS, "utf8"))

    // A floor: the file states the number more than once, which is why it drifted.
    expect(said.length, "extensions.ts states no block count at all").toBeGreaterThan(1)
    for (const one of said) {
      expect(
        one,
        `extensions.ts calls the set of blocks "${one}", and DESIGN.md §10 draws ` +
          `${rows.length}: ${rows.join(", ")}`,
      ).toBe(counted)
    }
  })

  it("leaves out the lists the design has no drawing for", () => {
    const source = readFileSync(EXTENSIONS, "utf8")
    for (const off of ["bulletList", "orderedList", "listItem"]) {
      expect(source, `${off} is not turned off`).toContain(`${off}: false`)
    }
  })
})

/** blockRowsInDesign is the Block column of the table under "## 10. Note blocks". */
function blockRowsInDesign(): string[] {
  const design = readFileSync(DESIGN, "utf8")
  const section = design.split("## 10. Note blocks")[1]?.split("\n## ")[0] ?? ""

  const names: string[] = []
  for (const line of section.split("\n")) {
    const cells = line.split("|").map((cell) => cell.trim())
    // A row of the table proper: leading and trailing pipe, and not the header or the
    // dashes under it.
    if (cells.length < 4 || cells[1] === "" || cells[1] === "Block" || cells[1].startsWith("---")) {
      continue
    }
    names.push(cells[1])
  }
  return names
}

/**
 * countsSaidIn is every number the prose uses for the set of blocks.
 *
 * Every occurrence, not the first: the drift this catches was one comment saying "five"
 * six lines under a header saying "seven", so a check satisfied by finding the right
 * word somewhere would have passed over it. It did, until this was rewritten.
 */
function countsSaidIn(source: string): string[] {
  return [...source.matchAll(/(?:among the|one of the|names)\s+(\w+)/g)].map((m) => m[1])
}

/** numberWord is how the comments write a small count. */
function numberWord(count: number): string {
  const words = ["zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"]
  return words[count] ?? String(count)
}
