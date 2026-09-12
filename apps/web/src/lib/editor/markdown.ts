import type { JSONContent } from "@tiptap/react"

import { inlineFrom, inlineTo } from "./marks"

import { markerOf, type BlockKind } from "./markers"

/**
 * Markdown in, document out, and back again.
 *
 * A Note is stored as markdown and edited as a document. That is deliberate: markdown
 * is what a Member exports, what the conflict rule in M13 compares, and what survives
 * Nooks being uninstalled — while a document tree is what an editor can render without
 * ever showing the markup. So the two meet here, at the save boundary, and nowhere else.
 *
 * Only the five block types the design names are represented. Anything else in the
 * markdown becomes a paragraph rather than being dropped: a Member's words matter more
 * than the shape they arrived in.
 */

/** TASK_LIST and friends are the node names the editor's extensions register. */
const NODE = {
  doc: "doc",
  paragraph: "paragraph",
  heading: "heading",
  taskList: "taskList",
  taskItem: "taskItem",
  quote: "blockquote",
  code: "codeBlock",
  text: "text",
} as const

/** HEADING_LEVEL is the one size of heading the design has. */
const HEADING_LEVEL = 3

/** FENCE is the shorthand that opens and closes a code block. */
const FENCE = "```"

/** documentFrom parses stored markdown into what the editor renders. */
export function documentFrom(markdown: string): JSONContent {
  const content: JSONContent[] = []
  const lines = markdown.split("\n")

  let index = 0
  while (index < lines.length) {
    const line = lines[index] ?? ""

    if (line.startsWith(FENCE)) {
      const [node, next] = readFence(lines, index)
      content.push(node)
      index = next
      continue
    }

    const marker = markerOf(line)
    if (marker.kind === "todo" || marker.kind === "todo-done") {
      const [node, next] = readTasks(lines, index)
      content.push(node)
      index = next
      continue
    }

    content.push(blockFrom(line, marker.kind, marker.length))
    index += 1
  }

  // A document with no blocks at all cannot be edited: there is nowhere to put the
  // caret. An empty Note is one empty paragraph.
  return { type: NODE.doc, content: content.length > 0 ? content : [paragraph("")] }
}

/** readFence reads a code block, and where the document carries on. */
function readFence(lines: string[], start: number): [JSONContent, number] {
  const body: string[] = []
  let index = start + 1

  while (index < lines.length && !(lines[index] ?? "").startsWith(FENCE)) {
    body.push(lines[index] ?? "")
    index += 1
  }

  // An unclosed fence still ends the document; the text inside it is not lost.
  return [textNode(NODE.code, body.join("\n")), index + 1]
}

/** readTasks reads a run of checklist lines as one list, and where it ends. */
function readTasks(lines: string[], start: number): [JSONContent, number] {
  const items: JSONContent[] = []
  let index = start

  while (index < lines.length) {
    const marker = markerOf(lines[index] ?? "")
    if (marker.kind !== "todo" && marker.kind !== "todo-done") {
      break
    }
    items.push({
      type: NODE.taskItem,
      attrs: { checked: marker.kind === "todo-done" },
      content: [paragraph((lines[index] ?? "").slice(marker.length))],
    })
    index += 1
  }

  return [{ type: NODE.taskList, content: items }, index]
}

/** blockFrom turns one line into the block its shorthand asked for. */
function blockFrom(line: string, kind: BlockKind, markerLength: number): JSONContent {
  const text = line.slice(markerLength)

  if (kind === "heading") {
    return { ...textNode(NODE.heading, text), attrs: { level: HEADING_LEVEL } }
  }
  if (kind === "quote") {
    return { type: NODE.quote, content: [paragraph(text)] }
  }
  return paragraph(text)
}

/** paragraph is one line of ordinary prose. */
function paragraph(text: string): JSONContent {
  return textNode(NODE.paragraph, text)
}

/**
 * textNode is a block holding the runs its line describes, or nothing at all.
 *
 * The line is read for inline markup on the way in, so `**bold**` arrives as bold text
 * rather than as four stars the Member has to look at.
 */
function textNode(type: string, text: string): JSONContent {
  return text ? { type, content: inlineFrom(text) } : { type }
}

/**
 * markdownFrom writes a document back out as markdown.
 *
 * The inverse of documentFrom for everything documentFrom produces, which is the
 * property markdown.test.ts holds it to.
 */
export function markdownFrom(document: JSONContent): string {
  const lines = (document.content ?? []).flatMap(linesOf)
  return lines.join("\n")
}

/** linesOf writes one block as the lines it occupies. */
function linesOf(node: JSONContent): string[] {
  switch (node.type) {
    case NODE.heading:
      return [`${"#".repeat(HEADING_LEVEL)} ${markedText(node)}`]
    case NODE.quote:
      return (node.content ?? []).map((child) => `> ${markedText(child)}`)
    case NODE.code:
      return [FENCE, ...plainText(node).split("\n"), FENCE]
    case NODE.taskList:
      return (node.content ?? []).map(taskLine)
    default:
      return [markedText(node)]
  }
}

/** taskLine writes one checklist item, ticked or not. */
function taskLine(item: JSONContent): string {
  const box = item.attrs?.checked === true ? "x" : " "
  return `- [${box}] ${(item.content ?? []).map(markedText).join(" ")}`
}

/**
 * markedText is everything written inside a node, with its inline markup put back.
 *
 * Used everywhere except a code block, where the contents are literal: markup inside a
 * code span is text, which is the whole point of writing something in one.
 */
function markedText(node: JSONContent): string {
  if (node.type === NODE.text) {
    return inlineTo([node])
  }
  return inlineTo(node.content ?? [])
}

/** plainText is everything written inside a node, with its marks left off. */
function plainText(node: JSONContent): string {
  if (node.type === NODE.text) {
    return node.text ?? ""
  }
  return (node.content ?? []).map(plainText).join("")
}
