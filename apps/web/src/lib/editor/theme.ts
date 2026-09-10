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
      color: "var(--foreground)",
      backgroundColor: "transparent",
    },
    // CodeMirror's own base theme puts `font-family: monospace` on the scroller, which
    // outranks anything set on the editor root. The app's font has to be set here.
    ".cm-scroller": {
      fontFamily: "var(--font-sans)",
      fontSize: body,
      lineHeight: "1.6",
    },
    "&.cm-focused": { outline: "none" },
    ".cm-content": {
      padding: "0",
      caretColor: "var(--shared)",
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

    // Emphasis is drawn, never spelled: the markup that asked for it is hidden.
    ".cm-nooks-strong": { fontWeight: "600" },
    ".cm-nooks-emphasis": { fontStyle: "italic" },
    ".cm-nooks-struck": {
      textDecoration: "line-through",
      color: "var(--muted-foreground)",
    },
    ".cm-nooks-inline-code": {
      fontFamily: "var(--font-mono)",
      fontSize: "0.92em",
      backgroundColor: "var(--chip)",
      borderRadius: "var(--radius-sm)",
      padding: "1px 4px",
    },
    ".cm-nooks-link": {
      color: "var(--shared)",
      textDecoration: "underline",
      textUnderlineOffset: "2px",
    },
    // The / menu, per DESIGN.md §10: 300px, its own heading, and 34px rows.
    ".cm-tooltip.cm-tooltip-autocomplete": {
      width: "300px",
      border: "1px solid var(--border)",
      borderRadius: "var(--radius-menu)",
      backgroundColor: "var(--popover)",
      boxShadow: "0 12px 32px rgba(0,0,0,0.14)",
      padding: "6px",
    },
    // The heading is set as a custom property by the editor, so it stays translated.
    ".cm-tooltip.cm-tooltip-autocomplete::before": {
      content: "var(--nooks-slash-title, none)",
      display: "block",
      padding: "7px 10px 8px",
      fontFamily: "var(--font-sans)",
      fontSize: "var(--text-label)",
      fontWeight: "600",
      letterSpacing: "0.04em",
      textTransform: "uppercase",
      color: "var(--muted-foreground)",
    },
    ".cm-tooltip-autocomplete > ul": {
      maxHeight: "none",
      fontFamily: "var(--font-sans)",
    },
    ".cm-tooltip-autocomplete > ul > li": {
      display: "flex",
      alignItems: "center",
      gap: "11px",
      height: "34px",
      padding: "0 10px",
      borderRadius: "var(--radius-md)",
      fontFamily: "var(--font-sans)",
      fontSize: "var(--text-meta)",
      color: "var(--foreground)",
    },
    ".cm-tooltip-autocomplete > ul > li[aria-selected]": {
      backgroundColor: "var(--secondary)",
      color: "var(--foreground)",
    },
    // The glyph sits in a column of its own so every label starts at the same x.
    ".cm-nooks-slash-glyph": {
      flex: "none",
      width: "20px",
      textAlign: "center",
      fontSize: "var(--text-micro)",
      color: "var(--secondary-foreground)",
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
