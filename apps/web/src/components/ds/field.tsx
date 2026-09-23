import { useId } from "react"
import { useTranslation } from "react-i18next"
import { cn } from "cn"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export type FieldProps = Omit<React.ComponentProps<typeof Input>, "id"> & {
  /** The label above the field. Always present: a placeholder is not a label. */
  label: string
  /**
   * One line under the field saying what is wanted, e.g. "Twelve characters or more.
   * No other rules." Rendered as a sentence, with a full stop.
   */
  hint?: string
  /** What went wrong. Replaces the hint and re-colours the field. */
  error?: string
  /**
   * Hides the label without removing it.
   *
   * For a settings row, which already prints the label in its left column. The label
   * stays in the markup because a field with no label is a field a screen reader cannot
   * announce — it is the seeing of it that is redundant, not the having.
   */
  hideLabel?: boolean
  /**
   * The most that may be typed here, from `lib/limits.ts`.
   *
   * It both stops the typing and says how much room is left, because either alone is
   * half an answer: a field that silently stops accepting letters looks broken, and a
   * count with nothing enforcing it is a suggestion.
   */
  limit?: number
}

/**
 * How close the limit has to be before the count appears.
 *
 * Counted in characters left rather than as a fraction of the limit, because that is
 * what the number has to mean: with twenty left you are about to run out whether the
 * field holds fifty characters or sixty-four thousand. A tenth of the way from the end
 * would put a counter under a note nobody is near filling, and under a quantity only
 * once it was too late.
 */
const COUNTER_SHOWS_AT = 20

/**
 * A labelled input, per DESIGN.md §7: 38px high, radius 7, 15px text, with the label
 * at 13/500 and the hint at 12.5 in muted text.
 *
 * useId rather than a caller-supplied id, so a field is always wired to its label and
 * two of them on one page cannot collide.
 */
export function Field({ label, hint, error, hideLabel, limit, className, ...props }: FieldProps) {
  const { t } = useTranslation()
  const id = useId()
  const left = roomLeft(limit, props.value)
  // The count is read out too, so it has to be inside what describes the field rather
  // than beside it.
  const describedBy = hint || error || left !== undefined ? `${id}-description` : undefined

  return (
    <div className="flex flex-col gap-1.5">
      <Label
        htmlFor={id}
        className={cn("text-secondary-foreground text-small font-medium", hideLabel && "sr-only")}
      >
        {label}
      </Label>

      <Input
        id={id}
        maxLength={limit}
        aria-describedby={describedBy}
        aria-invalid={error ? true : undefined}
        className={cn(
          "h-input rounded-lg border-input px-3 text-field focus-visible:ring-0",
          // Focus thickens the border rather than adding a ring, so the control does
          // not grow and shift the form.
          "focus-visible:border-[length:1.5px] focus-visible:border-ring",
          error && "border-destructive-line bg-destructive-bg text-destructive",
          className,
        )}
        {...props}
      />

      {(error || hint || left !== undefined) && (
        <p
          id={describedBy}
          className={cn(
            "flex items-baseline gap-3 text-micro",
            error ? "text-destructive" : "text-muted-foreground",
          )}
        >
          <span className="flex-1">{error ?? hint}</span>
          {left !== undefined && (
            <span className="shrink-0 tabular-nums">
              {t("field.charactersLeft", { count: left })}
            </span>
          )}
        </p>
      )}
    </div>
  )
}

/**
 * How much room is left, or nothing at all while the end is still far off.
 *
 * Never negative. A value longer than the limit can only be one that was stored before
 * the limit existed, since typing is stopped at it, and "minus twelve characters left"
 * is not a sentence. Nought is true: no more may be added, and the Instance will say so
 * on save.
 */
function roomLeft(limit: number | undefined, value: FieldProps["value"]): number | undefined {
  if (limit === undefined || typeof value !== "string") {
    return undefined
  }
  const left = limit - [...value].length
  if (left > COUNTER_SHOWS_AT) {
    return undefined
  }
  return Math.max(left, 0)
}
