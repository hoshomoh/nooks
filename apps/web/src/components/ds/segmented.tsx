import { cn } from "cn"

/** One of the answers a Segmented offers. */
export interface Segment<T> {
  value: T
  label: string
}

export interface SegmentedProps<T> {
  /** What a screen reader calls the group — usually the settings row's own label. */
  label: string
  options: Segment<T>[]
  chosen: T
  onChoose: (value: T) => void
}

/**
 * A few answers side by side, one of them on.
 *
 * For a choice small enough to read at a glance, where opening a control to see the
 * options would hide the thing that makes the choice easy. Three is the size it was
 * drawn for; more than four belongs in a select.
 */
export function Segmented<T extends string>({
  label,
  options,
  chosen,
  onChoose,
}: SegmentedProps<T>) {
  return (
    <div
      role="radiogroup"
      aria-label={label}
      className="flex h-control-settings items-center gap-0.5 rounded-lg border border-border p-0.5"
    >
      {options.map((option) => (
        <button
          key={option.value}
          type="button"
          role="radio"
          aria-checked={option.value === chosen}
          onClick={() => onChoose(option.value)}
          className={cn(
            "h-full rounded-md px-3.5 text-meta transition-colors",
            option.value === chosen
              ? "bg-secondary font-medium text-foreground"
              : "text-secondary-foreground hover:text-foreground",
          )}
        >
          {option.label}
        </button>
      ))}
    </div>
  )
}
