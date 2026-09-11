import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { EditorContent, useEditor, type JSONContent } from "@tiptap/react"
import { cn } from "cn"

import { SlashMenu } from "./slash-menu"
import { noteExtensions } from "@/lib/editor/extensions"
import { documentFrom, markdownFrom } from "@/lib/editor/markdown"
import { matching, slashItems } from "@/lib/editor/slash-items"
import { slashSuggestion, type SlashState } from "@/lib/editor/slash-suggestion"

/**
 * Which of the design's two sizes to draw the blocks at.
 *
 * DESIGN.md §10 gives two scales for the same blocks, so the editor is built once and
 * told which it is rather than written twice.
 */
export type NoteScale = "sheet" | "full"

export interface NoteEditorProps {
  /**
   * The Note the editor opens with, as markdown.
   *
   * Read once, when the editor mounts: the editor owns the document from then on, and
   * feeding it back on every render would fight the Member for the cursor. Give the
   * component a `key` to point it at a different Note.
   */
  initialValue: string
  /** Called on every change, as markdown. Whoever owns the saving decides when to act. */
  onChange: (markdown: string) => void
  readOnly?: boolean
  scale?: NoteScale
}

/**
 * The Note editor.
 *
 * A Note is a document, and this edits it as one: a heading is a heading, a checklist
 * is a checklist, and the markdown that describes them is never on screen because the
 * document does not contain any. Markdown is what a Note is *stored* as, and the
 * conversion happens at the two moments that matter — opening and saving.
 *
 * That is the whole reason this is ProseMirror rather than a text editor with the
 * markup painted over: there is no markup to hide.
 */
export function NoteEditor({
  initialValue,
  onChange,
  readOnly,
  scale = "sheet",
}: NoteEditorProps) {
  const { t } = useTranslation()

  // Captured once. The prop changes after every save, as the Item comes back from the
  // server, and reacting to that would rebuild the document under the Member's cursor.
  const [initialDocument] = useState<JSONContent>(() => documentFrom(initialValue))

  // What the `/` menu is showing, if anything. It is the editor that decides — the
  // slash, the query and the caret's rectangle all come from the document.
  const [slash, setSlash] = useState<SlashState | null>(null)

  const entries = useMemo(() => slashItems(t), [t])

  const extensions = useMemo(
    () => [
      ...noteExtensions({ placeholder: t("note.placeholder") }),
      slashSuggestion({
        onChange: setSlash,
        countFor: (query) => matching(entries, query).length,
        resolve: (query, index) => matching(entries, query)[index]?.kind ?? null,
      }),
    ],
    [entries, t],
  )

  const editor = useEditor({
    extensions,
    content: initialDocument,
    editable: !readOnly,
    onUpdate: ({ editor: changed }) => onChange(markdownFrom(changed.getJSON())),
    editorProps: { attributes: { class: "nooks-note outline-none" } },
  })

  const visible = slash ? matching(entries, slash.query) : []

  return (
    <div
      className={cn(
        "relative flex flex-1 flex-col",
        scale === "full" ? "nooks-note-full" : "nooks-note-sheet",
      )}
      // A Note is a page, so the whole of it is the page. Clicking the space below the
      // last line puts the caret at the end, rather than doing nothing because the
      // document happens to be shorter than the panel it sits in.
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          event.preventDefault()
          editor?.commands.focus("end")
        }
      }}
    >
      <EditorContent editor={editor} className="flex-1" />

      {slash && visible.length > 0 && (
        <div
          className="fixed z-50"
          style={{ left: slash.rect.left, top: slash.rect.bottom + 6 }}
          // The editor keeps the focus: clicking an entry must not take it away, or the
          // block would be inserted nowhere.
          onMouseDown={(event) => event.preventDefault()}
        >
          <SlashMenu
            items={visible}
            selected={slash.selected % visible.length}
            onChoose={(item) => slash.choose(item.kind)}
          />
        </div>
      )}
    </div>
  )
}
