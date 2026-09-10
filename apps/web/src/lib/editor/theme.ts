import { EditorView } from "@codemirror/view"
import type { Extension } from "@codemirror/state"

/**
 * How a Note's blocks are drawn, per DESIGN.md §10.
 *
 * Every value is a token: the editor is part of the app, not a widget with a look of
 * its own. The sizes are the sheet scale — the same blocks render one step larger at
 * full width.
 */
export const noteTheme: Extension = EditorView.theme({
  "&": {
    fontFamily: "var(--font-sans)",
    fontSize: "var(--text-note-sheet)",
    color: "var(--foreground)",
    backgroundColor: "transparent",
  },
  "&.cm-focused": {
    // Focus is the page's own 2px ring, not the editor's.
    outline: "none",
  },
  ".cm-content": {
    padding: "0",
    caretColor: "var(--shared)",
    lineHeight: "1.6",
  },
  ".cm-line": {
    padding: "5px 0",
  },
  ".cm-nooks-heading": {
    fontSize: "var(--text-note-heading)",
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
    fontSize: "var(--text-small)",
    color: "var(--secondary-foreground)",
    padding: "10px 0",
  },
  // A checklist line is indented to leave room for the box drawn in its gutter.
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
