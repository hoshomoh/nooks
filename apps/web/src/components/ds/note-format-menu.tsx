import { useTranslation } from "react-i18next"
import { BubbleMenu } from "@tiptap/react/menus"
import type { Editor } from "@tiptap/react"
import { cn } from "cn"

import { Icon, type IconName } from "./icon"

/** The three marks a Member reaches for on a phrase they have already written. */
const MARKS: { name: "bold" | "italic" | "strike"; icon: IconName; label: string }[] = [
  { name: "bold", icon: "bold", label: "note.bold" },
  { name: "italic", icon: "italic", label: "note.italic" },
  { name: "strike", icon: "struck", label: "note.struck" },
]

export interface NoteFormatMenuProps {
  /** The editor the selection belongs to, or nothing before it has mounted. */
  editor: Editor | null
}

/**
 * What appears over a phrase somebody has selected.
 *
 * DESIGN.md §10 says inline markup is markup: the punctuation is taken away and the
 * text is drawn as what it means. That leaves nowhere to type `**`, so selecting a
 * phrase is how a mark is applied — and the menu comes to the selection rather than
 * the Member going to a toolbar and losing it on the way.
 *
 * Only over a selection. With the caret merely resting somewhere there is nothing to
 * mark, and a menu that hovers over an empty caret is a menu in the way of the words.
 */
export function NoteFormatMenu({ editor }: NoteFormatMenuProps) {
  const { t } = useTranslation()

  if (!editor) {
    return null
  }

  return (
    <BubbleMenu
      editor={editor}
      // A code span is literal, so marking text inside one would write punctuation the
      // reader is meant to see as punctuation.
      shouldShow={({ editor: current, from, to }) =>
        current.isEditable && from !== to && !current.isActive("code")
      }
    >
      <div className="flex items-center gap-0.5 rounded-lg border border-border bg-background p-1 shadow-[0_12px_32px_rgba(0,0,0,0.14)]">
        {MARKS.map((mark) => {
          const active = editor.isActive(mark.name)
          return (
            <button
              key={mark.name}
              type="button"
              aria-label={t(mark.label)}
              aria-pressed={active}
              // The editor keeps the focus: taking it would collapse the selection the
              // Member is trying to mark.
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => editor.chain().focus().toggleMark(mark.name).run()}
              className={cn(
                "grid size-7 place-items-center rounded-md transition-colors",
                active
                  ? "bg-secondary text-foreground"
                  : "text-secondary-foreground hover:bg-secondary",
              )}
            >
              <Icon name={mark.icon} size="small" />
            </button>
          )
        })}
      </div>
    </BubbleMenu>
  )
}
