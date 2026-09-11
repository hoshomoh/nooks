import type { ComponentProps } from "react"
import { cn } from "cn"

import { Icon } from "./icon"

/** One of the answers a SelectField offers. */
export interface SelectOption {
  value: string
  label: string
}

export interface SelectFieldProps
  extends Omit<ComponentProps<"select">, "children" | "value" | "onChange"> {
  /** What a screen reader calls it — usually the settings row's own label. */
  label: string
  options: SelectOption[]
  value: string
  onValueChange: (value: string) => void
}

/**
 * The select a settings row carries, per DESIGN.md §8: 280 × 36, radius 7.
 *
 * The browser's own select, restyled rather than rebuilt: it is one of the few controls
 * where the native one is better on a phone than anything drawn by hand, and a Member
 * changing a language is not helped by a bespoke listbox.
 */
export function SelectField({
  label,
  options,
  value,
  onValueChange,
  className,
  ...props
}: SelectFieldProps) {
  return (
    <div className="relative w-70">
      <select
        aria-label={label}
        value={value}
        onChange={(event) => onValueChange(event.target.value)}
        className={cn(
          "h-control-settings w-full appearance-none rounded-lg border border-border bg-transparent",
          "pr-9 pl-3 text-field text-foreground transition-colors hover:bg-secondary",
          className,
        )}
        {...props}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>

      {/* Decorative: the select underneath already announces itself and its value. */}
      <Icon
        name="collapse"
        size="small"
        className="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-control"
      />
    </div>
  )
}
