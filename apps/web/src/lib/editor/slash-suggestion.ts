import { Extension, type Editor, type Range } from "@tiptap/react"
import Suggestion from "@tiptap/suggestion"

import { insert, type SlashKind } from "./slash-items"

/** What the `/` menu needs to draw itself, or null when it is closed. */
export interface SlashState {
  /** What has been typed after the slash. */
  query: string
  /** Where the caret is on the page, so the menu can sit under it. */
  rect: DOMRect
  /** Which entry the keyboard is on. */
  selected: number
  /** Turns the line into the chosen block. */
  choose: (kind: SlashKind) => void
}

export interface SlashSuggestionOptions {
  /** Called whenever the menu opens, changes or closes. */
  onChange: (state: SlashState | null) => void
  /** How many entries the query currently matches, for wrapping the selection. */
  countFor: (query: string) => number
  /**
   * Which block the nth matching entry is, or null when there is no nth.
   *
   * The plugin does not know what the menu is showing — the entries are matched where
   * they are drawn, so the words being compared are the ones a Member can read.
   */
  resolve: (query: string, index: number) => SlashKind | null
}

/**
 * The `/` trigger, per DESIGN.md §10: a slash that **begins a line** opens the block
 * menu; one inside a sentence is a slash.
 *
 * Built on Tiptap's suggestion utility rather than by watching the document, because
 * knowing where the caret is on the page, what has been typed since the trigger, and
 * which range to replace are three things it already does — and three more things to
 * get wrong by hand.
 */
export function slashSuggestion(options: SlashSuggestionOptions): Extension {
  return Extension.create({
    name: "nooksSlashMenu",

    addProseMirrorPlugins() {
      // Held across the plugin's callbacks: which entry the keyboard is on, and where
      // the menu would insert. Neither belongs in the document.
      let selected = 0
      let latest: { editor: Editor; range: Range; query: string } | null = null
      // Where the caret was when the menu last moved. Arrow keys change which entry is
      // chosen without moving the caret, so the rectangle has to outlive them.
      let rect: DOMRect | null = null

      const publish = () => {
        if (!latest || !rect) {
          options.onChange(null)
          return
        }
        const { editor, range, query } = latest
        options.onChange({
          query,
          rect,
          selected,
          choose: (kind) => insert(editor, range, kind),
        })
      }

      return [
        Suggestion({
          editor: this.editor,
          char: "/",
          startOfLine: true,
          // The entries are matched by whoever draws them, so that the words being
          // compared are the ones a Member can read.
          items: () => [],

          render: () => ({
            onStart: (props) => {
              selected = 0
              latest = { editor: props.editor, range: props.range, query: props.query }
              rect = props.clientRect?.() ?? null
              publish()
            },
            onUpdate: (props) => {
              latest = { editor: props.editor, range: props.range, query: props.query }
              rect = props.clientRect?.() ?? rect
              publish()
            },
            onKeyDown: (props) => moveOrChoose(props.event),
            onExit: () => {
              latest = null
              rect = null
              options.onChange(null)
            },
          }),
        }),
      ]

      /** moveOrChoose answers the keys the menu owns, and nothing else. */
      function moveOrChoose(event: KeyboardEvent): boolean {
        if (!latest) {
          return false
        }
        const count = options.countFor(latest.query)
        if (count === 0) {
          return false
        }

        if (event.key === "ArrowDown") {
          selected = (selected + 1) % count
          publish()
          return true
        }
        if (event.key === "ArrowUp") {
          selected = (selected - 1 + count) % count
          publish()
          return true
        }
        if (event.key === "Enter") {
          const kind = options.resolve(latest.query, selected % count)
          if (!kind) {
            return false
          }
          insert(latest.editor, latest.range, kind)
          return true
        }
        return false
      }
    },
  })
}
