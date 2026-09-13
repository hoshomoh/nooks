/**
 * What the markdown shorthand at the start of a line means.
 *
 * DESIGN.md §10: typing the shorthand converts the line as the Member goes, and the
 * marker is never displayed back to them. Deciding *what* a line is has to be pure so
 * the rules can be read and tested in one place; hiding the marker is the editor's job.
 */

/** The block types a Note line can be. */
export type BlockKind = "paragraph" | "heading" | "todo" | "todo-done" | "quote" | "code"

/** What was found at the start of a line. */
export type LineMarker = {
  kind: BlockKind
  /** How many characters the marker occupies, including its trailing space. */
  length: number
}

/**
 * MARKERS is every shorthand, longest first.
 *
 * Order matters: "- [ ] " has to be tested before "- ", or a checklist line would be
 * read as a bullet and keep its box.
 */
const MARKERS: ReadonlyArray<{ prefix: string; kind: BlockKind }> = [
  { prefix: "- [ ] ", kind: "todo" },
  { prefix: "- [x] ", kind: "todo-done" },
  { prefix: "- [X] ", kind: "todo-done" },
  { prefix: "###### ", kind: "heading" },
  { prefix: "##### ", kind: "heading" },
  { prefix: "#### ", kind: "heading" },
  { prefix: "### ", kind: "heading" },
  { prefix: "## ", kind: "heading" },
  { prefix: "# ", kind: "heading" },
  { prefix: "> ", kind: "quote" },
]

/**
 * BARE is a checklist marker with nothing after it.
 *
 * The marker is written with a trailing space, so an item whose words are deleted is
 * "- [ ] " and still reads as one. Anything that trims trailing whitespace on the way
 * past — a store, a script writing markdown by hand — leaves "- [ ]", which then reads
 * as an ordinary line and shows the Member the box they typed as text.
 */
const BARE: ReadonlyArray<{ line: string; kind: BlockKind }> = [
  { line: "- [ ]", kind: "todo" },
  { line: "- [x]", kind: "todo-done" },
  { line: "- [X]", kind: "todo-done" },
]

/** markerOf reads the shorthand at the start of a line. */
export function markerOf(line: string): LineMarker {
  for (const { prefix, kind } of MARKERS) {
    if (line.startsWith(prefix)) {
      return { kind, length: prefix.length }
    }
  }
  // Only when it is the whole line: "- [ ]x" is not an empty checklist item, it is a
  // line somebody typed that happens to start the same way.
  for (const { line: bare, kind } of BARE) {
    if (line.trimEnd() === bare) {
      return { kind, length: line.length }
    }
  }
  // A fence opens and closes a code block; the line itself carries no text.
  if (line.startsWith("```")) {
    return { kind: "code", length: line.length }
  }
  return { kind: "paragraph", length: 0 }
}

/** The shorthand each block type is written with, shown in the / menu. */
export const SHORTHAND: Record<Exclude<BlockKind, "paragraph" | "todo-done">, string> = {
  heading: "###",
  todo: "- [ ]",
  quote: ">",
  code: "```",
}

/** What inserting a block type from the / menu puts at the start of the line. */
export const INSERTION: Record<Exclude<BlockKind, "paragraph" | "todo-done">, string> = {
  heading: "### ",
  todo: "- [ ] ",
  quote: "> ",
  code: "```\n\n```",
}

/**
 * toggleTodo flips a checklist line between done and not done, leaving anything else
 * alone. Ticking in the Note is the same act as ticking in the list.
 */
export function toggleTodo(line: string): string {
  if (line.startsWith("- [ ] ")) {
    return `- [x] ${line.slice(6)}`
  }
  if (line.startsWith("- [x] ") || line.startsWith("- [X] ")) {
    return `- [ ] ${line.slice(6)}`
  }
  return line
}

/**
 * continueList returns what a new line should start with when Enter is pressed on this
 * one — so a checklist keeps going without retyping the marker, and stops when the
 * Member leaves an empty one.
 */
export function continueList(line: string): string | null {
  const marker = markerOf(line)
  if (marker.kind !== "todo" && marker.kind !== "todo-done" && marker.kind !== "quote") {
    return null
  }
  // An empty marked line means they are done with the list.
  if (line.slice(marker.length).trim() === "") {
    return ""
  }
  return marker.kind === "quote" ? "> " : "- [ ] "
}
