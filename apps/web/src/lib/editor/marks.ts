import type { JSONContent } from "@tiptap/react"

/**
 * Inline markup, in both directions.
 *
 * DESIGN.md §10: `**bold**`, `*italic*`, `~~struck~~` and `` `code` `` are markup like
 * any other, drawn as what they mean with the punctuation taken away. A Note is stored
 * as markdown, so the punctuation has to come back on the way out — a Member who bolds
 * a phrase and finds it plain the next morning has been told the app forgot.
 *
 * Written here rather than reached for from a library: the document only ever holds
 * the four marks the design names, and a full CommonMark implementation would parse
 * three more that the editor cannot draw.
 */

/** The four marks a Note can carry, longest delimiter first so `**` beats `*`. */
const MARKS = [
  { name: "code", token: "`" },
  { name: "bold", token: "**" },
  { name: "strike", token: "~~" },
  { name: "italic", token: "*" },
] as const

/** One of the four. Named so a caller cannot ask for a mark the editor cannot draw. */
export type MarkName = (typeof MARKS)[number]["name"]

/**
 * A link, which is inline markup of a different shape.
 *
 * DESIGN.md §10 names it alongside the other four: drawn as what it means, with the
 * punctuation taken away — including the address.
 */
const LINK = /\[([^\]]+)\]\(([^)\s]+)\)/

/** One mark on a run: a name, and for a link the address it points at. */
interface MarkSpec {
  type: string
  attrs?: Record<string, unknown>
}

/** The order marks are written in, so the same document always writes the same way. */
const WRITE_ORDER: MarkName[] = ["bold", "italic", "strike", "code"]

/**
 * inlineFrom reads one line of markdown into the text runs it describes.
 *
 * Answers a single plain run when there is no markup, which is the ordinary case.
 */
export function inlineFrom(text: string): JSONContent[] {
  return runsIn(text, [])
}

/**
 * inlineTo writes text runs back out as markdown.
 *
 * The inverse of inlineFrom for everything inlineFrom produces, which is the property
 * marks.test.ts holds it to.
 */
export function inlineTo(nodes: readonly JSONContent[]): string {
  return nodes.map(writeRun).join("")
}

/** PAIRING are the inline tokens, which only mean anything when one closes another. */
const PAIRING = ["*", "`", "~"] as const

/*
escapeText keeps punctuation somebody typed from reading as markup next time.

Only where it would. A delimiter on its own is a character somebody typed: "2 * 3 items"
has always come back as itself and should go on doing so, and a Note full of backslashes
is read by whoever calls the API and by any assistant holding one. So a pairing token is
protected when the same one appears again and could close it, and a bracket when there is
a `](` after it for a link to end with.

A backslash is always protected. Without that, somebody typing `\*` as two characters
would find it had become an escape, which is the same loss one turn later.

`]`, `(` and `)` are left alone: with no unescaped `[` in front of them a link cannot
begin.
*/
function escapeText(text: string): string {
  const pairs = new Set(
    PAIRING.filter((token) => text.indexOf(token) !== text.lastIndexOf(token)),
  )

  let out = ""
  for (let at = 0; at < text.length; at++) {
    const here = text[at] ?? ""
    if (here === "\\" || pairs.has(here as (typeof PAIRING)[number]) || opensALink(text, at)) {
      out += "\\"
    }
    out += here
  }
  return out
}

/** opensALink reports whether a bracket here has somewhere to close. */
function opensALink(text: string, at: number): boolean {
  return text[at] === "[" && text.slice(at + 1).includes("](")
}

/** unescapeText is the inverse, applied where a run of literal text is made. */
function unescapeText(text: string): string {
  return text.replace(/\\([\\`*~[])/g, "$1")
}

/*
masked hides what a backslash protects from the search for markup, keeping the length so
that every position found in it is a position in the text itself.

Searching the text directly would find the `*` in `\*this\*` and read it as emphasis,
which is the fault this exists to stop.
*/
function masked(text: string): string {
  let out = ""
  for (let at = 0; at < text.length; at++) {
    if (text[at] === "\\" && at + 1 < text.length) {
      out += "\0\0"
      at++
      continue
    }
    out += text[at]
  }
  return out
}

/** runsIn parses text into runs, carrying the marks already open around it. */
function runsIn(text: string, open: readonly MarkSpec[]): JSONContent[] {
  const search = masked(text)
  const link = LINK.exec(search)
  const found = firstPair(search)

  // Whichever starts first wins, so `**[a](b)**` and `[**a**](b)` both read correctly
  // rather than depending on which rule was tried first.
  if (link && (!found || link.index < found.start)) {
    return linkRuns(text, link, open)
  }
  if (!found) {
    return text ? [run(unescapeText(text), open)] : []
  }

  const before = text.slice(0, found.start)
  const inside = text.slice(found.start + found.token.length, found.end)
  const after = text.slice(found.end + found.token.length)

  return [
    ...(before ? [run(unescapeText(before), open)] : []),
    // Code is literal: markup inside a code span is text, which is the whole point of
    // writing something in a code span.
    ...(found.name === "code"
      ? [run(inside, [...open, { type: "code" }])]
      : runsIn(inside, [...open, { type: found.name }])),
    ...runsIn(after, open),
  ]
}

/** Where a delimiter opens and closes, for the earliest pair in the text. */
interface Pair {
  name: MarkName
  token: string
  start: number
  end: number
}

/**
 * firstPair finds the earliest delimiter that actually closes.
 *
 * An unmatched `*` is a star somebody typed, not the start of anything, so it is left
 * alone rather than swallowing the rest of the line.
 */
function firstPair(text: string): Pair | null {
  let earliest: Pair | null = null

  for (const { name, token } of MARKS) {
    const start = text.indexOf(token)
    if (start < 0) {
      continue
    }
    const end = text.indexOf(token, start + token.length)
    if (end < 0 || end === start + token.length) {
      // Nothing between the pair is not emphasis; `**` on its own is two stars.
      continue
    }
    if (!earliest || start < earliest.start) {
      earliest = { name, token, start, end }
    }
  }

  return earliest
}

/** linkRuns reads `[text](url)` and whatever surrounds it. */
function linkRuns(text: string, link: RegExpExecArray, open: readonly MarkSpec[]): JSONContent[] {
  const before = text.slice(0, link.index)
  const after = text.slice(link.index + link[0].length)

  // Cut from the text rather than taken from the match, because the match was made
  // against the masked copy and would carry the mask's own characters. The positions
  // are the same in both, which is what masked() preserves the length for.
  const labelAt = link.index + "[".length
  const label = text.slice(labelAt, labelAt + (link[1] ?? "").length)
  const hrefAt = labelAt + label.length + "](".length
  const href = unescapeText(text.slice(hrefAt, hrefAt + (link[2] ?? "").length))

  return [
    ...(before ? runsIn(before, open) : []),
    ...runsIn(label, [...open, { type: "link", attrs: { href } }]),
    ...runsIn(after, open),
  ]
}

/** run is one stretch of text and the marks covering it. */
function run(text: string, marks: readonly MarkSpec[]): JSONContent {
  const node: JSONContent = { type: "text", text }
  if (marks.length > 0) {
    node.marks = marks.map((mark) => ({ ...mark }))
  }
  return node
}

/** writeRun puts the punctuation back around one run. */
function writeRun(node: JSONContent): string {
  const text = node.text ?? ""
  if (!text) {
    return ""
  }

  const carried = node.marks ?? []
  const names = new Set(carried.map((mark) => mark.type))

  // Everything but a code span, where markdown does not read a backslash as protecting
  // anything and the point of the span is that what is inside it is literal.
  const written = names.has("code") ? text : escapeText(text)

  // Written outermost first and closed in reverse, so a bold-italic phrase reads the
  // same every time rather than depending on which mark was applied first.
  const marked = WRITE_ORDER.filter((name) => names.has(name)).reduceRight(
    (inner, name) => wrap(inner, name),
    written,
  )

  // The link goes outside the rest: `[**loud**](url)` rather than `**[loud](url)**`,
  // which is what every markdown reader writes.
  const href = carried.find((mark) => mark.type === "link")?.attrs?.href
  return typeof href === "string" ? `[${marked}](${href})` : marked
}

/** wrap puts one mark's delimiters around what it covers. */
function wrap(text: string, name: MarkName): string {
  const token = MARKS.find((mark) => mark.name === name)?.token ?? ""
  return `${token}${text}${token}`
}
