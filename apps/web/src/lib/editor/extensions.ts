import StarterKit from "@tiptap/starter-kit"
import Placeholder from "@tiptap/extension-placeholder"
import TaskItem from "@tiptap/extension-task-item"
import TaskList from "@tiptap/extension-task-list"
import type { Extension, Node } from "@tiptap/react"

/**
 * The blocks a Note is made of, and nothing else.
 *
 * DESIGN.md §10 names five: paragraph, heading, checklist, quote, code. Every other
 * block StarterKit offers is turned off here rather than styled — a block type a
 * Member can reach but the design has no drawing for is a block type that will look
 * wrong the first time somebody uses it.
 *
 * Headings have exactly one level, because the design has exactly one heading.
 */
export interface NoteExtensionOptions {
  /** What an empty Note says, already translated. */
  placeholder: string
}

export function noteExtensions(options: NoteExtensionOptions): (Extension | Node)[] {
  return [
    StarterKit.configure({
      heading: { levels: [3] },
      // Bullet and ordered lists are not among the five. A checklist is the list Nooks
      // has, and it is the one that means something on a household todo.
      bulletList: false,
      orderedList: false,
      listItem: false,
      // Nothing in the design draws a rule across a Note.
      horizontalRule: false,
      // Links are named in DESIGN.md §10 alongside the other inline markup, drawn as
      // what they mean with the address taken away. They open in a new tab rather than
      // navigating the app out from under an unsaved Note, and never on a click inside
      // the editor, where a click means "put the caret here".
      link: {
        openOnClick: false,
        autolink: true,
        defaultProtocol: "https",
        HTMLAttributes: { rel: "noopener noreferrer", target: "_blank" },
      },
    }),
    TaskList,
    TaskItem.configure({ nested: false }),
    Placeholder.configure({ placeholder: options.placeholder }),
  ]
}
