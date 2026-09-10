import { useId } from "react"
import { cn } from "cn"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

type FieldProps = Omit<React.ComponentProps<typeof Input>, "id"> & {
  /** The label above the field. Always present: a placeholder is not a label. */
  label: string
  /**
   * One line under the field saying what is wanted, e.g. "Twelve characters or more.
   * No other rules." Rendered as a sentence, with a full stop.
   */
  hint?: string
  /** What went wrong. Replaces the hint and re-colours the field. */
  error?: string
}

/**
 * A labelled input, per DESIGN.md §7: 38px high, radius 7, 15px text, with the label
 * at 13/500 and the hint at 12.5 in muted text.
 *
 * useId rather than a caller-supplied id, so a field is always wired to its label and
 * two of them on one page cannot collide.
 */
export function Field({ label, hint, error, className, ...props }: FieldProps) {
  const id = useId()
  const describedBy = hint || error ? `${id}-description` : undefined

  return (
    <div className="flex flex-col gap-1.5">
      <Label htmlFor={id} className="text-secondary-foreground text-small font-medium">
        {label}
      </Label>

      <Input
        id={id}
        aria-describedby={describedBy}
        aria-invalid={error ? true : undefined}
        className={cn(
          "h-input rounded-lg border-input px-3 text-input focus-visible:ring-0",
          // Focus thickens the border rather than adding a ring, so the control does
          // not grow and shift the form.
          "focus-visible:border-[length:1.5px] focus-visible:border-ring",
          error && "border-destructive-line bg-destructive-bg text-destructive",
          className,
        )}
        {...props}
      />

      {(error || hint) && (
        <p
          id={describedBy}
          className={cn("text-micro", error ? "text-destructive" : "text-muted-foreground")}
        >
          {error ?? hint}
        </p>
      )}
    </div>
  )
}
