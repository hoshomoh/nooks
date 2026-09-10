import { useCallback, useMemo, useRef, useState } from "react"
import type { CSSProperties } from "react"
import { useTranslation } from "react-i18next"
import { EditorState } from "@codemirror/state"
import { EditorView, keymap } from "@codemirror/view"
import { defaultKeymap, history, historyKeymap } from "@codemirror/commands"
import { markdown } from "@codemirror/lang-markdown"

import { liveMarkers } from "@/lib/editor/live-markers"
import { slashMenu, type SlashOption } from "@/lib/editor/slash-menu"
import { buildNoteTheme, type NoteScale } from "@/lib/editor/theme"

export type NoteEditorProps = {
  /**
   * The Note the editor opens with, as markdown.
   *
   * Read once, when the editor mounts: CodeMirror owns the document from then on, and
   * feeding it back on every render would fight the Member for the cursor. Give the
   * component a `key` to point it at a different Note.
   */
  initialValue: string
  /** Called on every change. Whoever owns the saving decides when to act on it. */
  onChange: (markdown: string) => void
  readOnly?: boolean
  /** Which of the design's two sizes to render the blocks at. */
  scale?: NoteScale
}

/**
 * The Note editor.
 *
 * A Note is markdown, and this shows it as blocks: the shorthand converts the line as
 * it is typed and is then hidden, so a Member sees a heading rather than `### Where`.
 * The marker reappears on the line the cursor is on, because otherwise there would be
 * no way to take it off again.
 *
 * The view is created once and driven imperatively. CodeMirror owns its own DOM and
 * state; re-creating it on every render would lose the cursor mid-sentence.
 */
/** The custom properties the editor hands down to its own stylesheet. */
type EditorStyle = CSSProperties & Record<"--nooks-slash-title", string>

/**
 * slashTitleStyle carries the menu's heading into CSS.
 *
 * `content` takes a quoted string, so the words are quoted here rather than in the
 * stylesheet, where they could not be translated.
 */
function slashTitleStyle(title: string): EditorStyle {
  return { "--nooks-slash-title": JSON.stringify(title) }
}

export function NoteEditor({ initialValue, onChange, readOnly, scale = "sheet" }: NoteEditorProps) {
  const { t } = useTranslation()
  const viewRef = useRef<EditorView | null>(null)

  // Captured once. The prop changes after every save, as the Item comes back from the
  // server, and reacting to that would rebuild the editor under the Member's cursor.
  const [initialDoc] = useState(initialValue)

  const options: SlashOption[] = useMemo(
    () => [
      { kind: "todo", label: t("note.blocks.todo"), glyph: "☐" },
      { kind: "heading", label: t("note.blocks.heading"), glyph: "H" },
      { kind: "quote", label: t("note.blocks.quote"), glyph: "❝" },
      { kind: "code", label: t("note.blocks.code"), glyph: "{}" },
    ],
    [t],
  )

  /**
   * Mounts the editor when the element appears and tears it down when it goes.
   *
   * A callback ref rather than an effect: the element's arrival is the event, and React
   * hands it over directly. There is nothing to synchronise afterwards.
   *
   * Its dependencies are all stable, so in practice it runs once: the initial document
   * is captured with useState, the menu options are memoised, and the caller passes a
   * stable onChange. A different Note means a different `key`, and a fresh editor.
   */
  const mount = useCallback(
    (element: HTMLDivElement | null) => {
      if (!element) {
        viewRef.current?.destroy()
        viewRef.current = null
        return
      }

      const view = new EditorView({
        parent: element,
        state: EditorState.create({
          doc: initialDoc,
          extensions: [
            history(),
            keymap.of([...defaultKeymap, ...historyKeymap]),
            markdown(),
            liveMarkers,
            slashMenu({ options }),
            buildNoteTheme(scale),
            EditorView.lineWrapping,
            EditorState.readOnly.of(Boolean(readOnly)),
            EditorView.updateListener.of((update) => {
              if (update.docChanged) {
                onChange(update.state.doc.toString())
              }
            }),
          ],
        }),
      })
      viewRef.current = view
    },
    [initialDoc, onChange, options, readOnly, scale],
  )

  return (
    <div
      ref={mount}
      className="flex-1"
      // The / menu's heading is drawn by CSS, which cannot read a translation. Handing
      // it down as a custom property keeps the words in the Member's language.
      style={slashTitleStyle(t("note.addToNote"))}
    />
  )
}
