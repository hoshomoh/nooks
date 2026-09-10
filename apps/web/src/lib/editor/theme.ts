import { EditorView } from "@codemirror/view"
import type { Extension } from "@codemirror/state"

/**
 * How a Note's blocks are drawn, per DESIGN.md §10.
 *
 * Every value is a token: the editor is part of the app, not a widget with a look of
 * its own.
 *
 * Scale is the only difference between the sheet and the full-screen view.
 *
 * DESIGN.md §10 gives two sizes for the same blocks, so the editor is built once and
 * told which it is rather than written twice.
 */
export type NoteScale = "sheet" | "full"

/** buildNoteTheme renders the blocks at one of the two scales. */
export function buildNoteTheme(scale: NoteScale): Extension {
  const body = scale === "full" ? "var(--text-note)" : "var(--text-note-sheet)"
  const heading = scale === "full" ? "var(--text-note-heading)" : "var(--text-small)"
  const code = scale === "full" ? "var(--text-small)" : "var(--text-micro)"

  return EditorView.theme({
    "&": {
      fontFamily: "var(--font-sans)",
      fontSize: body,
      color: "var(--foreground)",
      backgroundColor: "transparent",
    },
    "&.cm-focused": { outline: "none" },
    ".cm-content": {
      padding: "0",
      caretColor: "var(--shared)",
      lineHeight: "1.6",
    },
    ".cm-line": { padding: "5px 0" },
    ".cm-nooks-heading": {
      fontSize: heading,
      fontWeight: "600",
      letterSpacing: "-0.01em",
      lineHeight: "1.4",
      padding: "18px 0 3px",
    },
    ".cm-nooks-quote": {
      color: "var(--secondary-foreground)",
      borderLeft: "2px solid var(--border)",
      paddingLeft: "13px",
      margin: "10px 0",
    },
    ".cm-nooks-code": {
      fontFamily: "var(--font-mono)",
      fontSize: code,
      color: "var(--secondary-foreground)",
      padding: "10px 0",
    },
    ".cm-nooks-todo, .cm-nooks-todo-done": {
      paddingLeft: "26px",
      position: "relative",
    },
    ".cm-nooks-todo::before, .cm-nooks-todo-done::before": {
      content: '""',
      position: "absolute",
      left: "0",
      top: "9px",
      width: "15px",
      height: "15px",
      borderRadius: "var(--radius-sm)",
      border: "1.5px solid var(--control)",
    },
    ".cm-nooks-todo-done": {
      color: "var(--muted-foreground)",
      textDecoration: "line-through",
    },
    ".cm-nooks-todo-done::before": {
      backgroundColor: "var(--muted-foreground)",
      borderColor: "var(--muted-foreground)",
    },
    ".cm-tooltip.cm-tooltip-autocomplete": {
      border: "1px solid var(--border)",
      borderRadius: "var(--radius-2xl)",
      backgroundColor: "var(--popover)",
      boxShadow: "0 12px 32px rgba(0,0,0,0.14)",
      padding: "6px",
    },
    ".cm-tooltip-autocomplete > ul > li": {
      borderRadius: "var(--radius-md)",
      padding: "7px 10px",
      fontFamily: "var(--font-sans)",
      fontSize: "var(--text-secondary)",
    },
    ".cm-tooltip-autocomplete > ul > li[aria-selected]": {
      backgroundColor: "var(--secondary)",
      color: "var(--foreground)",
    },
    ".cm-completionDetail": {
      marginLeft: "auto",
      paddingLeft: "16px",
      fontFamily: "var(--font-mono)",
      fontSize: "var(--text-keycap)",
      fontStyle: "normal",
      color: "var(--muted-foreground)",
    },
  })
}
