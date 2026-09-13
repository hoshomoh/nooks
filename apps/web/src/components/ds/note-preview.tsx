import { cn } from "cn"

import type { PreviewRun } from "@/lib/editor/preview"

export interface NotePreviewProps {
  /** The first block of the Note, as the runs it is made of. */
  runs: PreviewRun[]
  className?: string
}

/**
 * A Note's own first line, drawn the way the Note draws it.
 *
 * Marks are rendered rather than stripped, because a row showing `**loud**` is showing
 * the Member punctuation they never typed. A link is not a link here: the row is already
 * one target that opens the Item, and a second thing to hit inside it is a row nobody
 * can aim at. It still reads as a link, in the shared colour, so it is recognisable as
 * what it will be once opened.
 */
export function NotePreview({ runs, className }: NotePreviewProps) {
  return (
    <span className={cn("truncate", className)}>
      {runs.map((run, index) => (
        <span
          // Runs have no identity of their own; their position in the line is what they
          // are. Nothing reorders inside a preview, so the index is stable.
          key={index}
          className={cn(
            run.marks.includes("bold") && "font-semibold",
            run.marks.includes("italic") && "italic",
            run.marks.includes("strike") && "line-through",
            run.marks.includes("code") && "rounded-sm bg-secondary px-1 font-mono text-[0.92em]",
            run.marks.includes("link") && "text-shared",
          )}
        >
          {run.text}
        </span>
      ))}
    </span>
  )
}
