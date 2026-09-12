import { cn } from "cn"

import { Button } from "./button"
import type { RowTrouble as Trouble } from "@/lib/row-trouble"

/** The words the panel needs that are not the Member's own. */
export interface RowTroubleLabels {
  /** Heading when somebody else wrote first, e.g. "Somebody else renamed this". */
  conflict: string
  /** What the Member wrote, e.g. "Yours". */
  mine: string
  /** What is on the List now, e.g. "On the list now". */
  theirs: string
  keepMine: string
  keepBoth: string
  /** Heading when the Instance refused, e.g. "That did not save". */
  failed: string
  /** Reassurance that nothing was lost, e.g. "Your text is still here." */
  kept: string
  tryAgain: string
  discard: string
}

export interface RowTroubleProps {
  trouble: Trouble
  /** What is on the List now — the row's own current text. */
  theirs: string
  onKeepMine: () => void
  onKeepBoth: () => void
  onTryAgain: () => void
  onDiscard: () => void
  labels: RowTroubleLabels
}

/**
 * What a row says when a change did not land, per DESIGN.md §11.
 *
 * Two different things, and they are drawn differently because they are different. A
 * conflict is not a failure: both versions exist, both are somebody's words, and only a
 * person can say which should survive — so it asks, with the Member's own outlined in
 * accent. A failure is a failure, and takes the destructive treatment.
 *
 * What they share is the rule that matters: the Member's text is on screen in both, and
 * every button here keeps it. Nothing offered discards what somebody wrote without
 * their saying so.
 */
export function RowTrouble({
  trouble,
  theirs,
  onKeepMine,
  onKeepBoth,
  onTryAgain,
  onDiscard,
  labels,
}: RowTroubleProps) {
  if (trouble.kind === "conflict") {
    return (
      <div
        role="alert"
        className="mt-1 mb-2 flex flex-col gap-2.5 rounded-lg border border-border bg-sidebar px-3.5 py-3"
      >
        <p className="text-meta text-secondary-foreground">{labels.conflict}</p>

        <Version label={labels.theirs} text={theirs} />
        <Version label={labels.mine} text={trouble.mine} mine />

        <div className="flex flex-wrap gap-2 pt-0.5">
          <Button tone="primary" scale="compact" onClick={onKeepMine}>
            {labels.keepMine}
          </Button>
          <Button tone="secondary" scale="compact" onClick={onKeepBoth}>
            {labels.keepBoth}
          </Button>
          <Button tone="quiet" scale="compact" onClick={onDiscard}>
            {labels.discard}
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div
      role="alert"
      className="mt-1 mb-2 flex flex-col gap-2 rounded-lg border border-destructive-line bg-destructive-bg px-3.5 py-3"
    >
      <p className="text-meta text-destructive">{labels.failed}</p>
      {trouble.said && <p className="text-meta text-secondary-foreground">{trouble.said}</p>}
      <p className="text-meta text-secondary-foreground">
        {labels.kept} <span className="text-foreground">“{trouble.mine}”</span>
      </p>

      <div className="flex flex-wrap gap-2 pt-0.5">
        <Button tone="destructive" scale="compact" onClick={onTryAgain}>
          {labels.tryAgain}
        </Button>
        <Button tone="quiet" scale="compact" onClick={onDiscard}>
          {labels.discard}
        </Button>
      </div>
    </div>
  )
}

/** One of the two versions, stacked. The Member's own is outlined in accent. */
function Version({ label, text, mine }: { label: string; text: string; mine?: boolean }) {
  return (
    <div
      className={cn(
        "flex flex-col gap-0.5 rounded-md border px-2.5 py-1.5",
        mine ? "border-shared-line bg-shared-bg" : "border-hair bg-background",
      )}
    >
      <span className="text-micro text-muted-foreground">{label}</span>
      <span className="text-meta text-foreground">{text}</span>
    </div>
  )
}
