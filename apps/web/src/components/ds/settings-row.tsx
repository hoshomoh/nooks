import type { ReactNode } from "react"
import { cn } from "cn"

export interface SettingsRowProps {
  label: string
  /** One line under the label saying what the setting does. */
  blurb?: string
  children: ReactNode
}

/**
 * One setting, per DESIGN.md §8: a `190px 1fr` grid, 48px minimum, split by a hairline.
 *
 * The control is right-aligned and the explanation sits under the label, so a column of
 * settings reads as a list of sentences rather than a form.
 */
export function SettingsRow({ label, blurb, children }: SettingsRowProps) {
  return (
    <div className="grid min-h-12 grid-cols-[190px_1fr] items-center gap-6 border-b border-hair py-1.5">
      <div className="flex flex-col gap-0.5">
        <span className="text-chrome">{label}</span>
        {blurb && <span className="text-micro leading-[1.5] text-muted-foreground">{blurb}</span>}
      </div>
      <div className="flex items-center justify-end gap-3">{children}</div>
    </div>
  )
}

export interface ToggleProps {
  on: boolean
  onChange: (on: boolean) => void
  /** What a screen reader calls it — usually the row's own label. */
  label: string
}

/** The toggle, per DESIGN.md §7: a 34×20 track with a 16px knob. */
export function Toggle({ on, onChange, label }: ToggleProps) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={on}
      aria-label={label}
      onClick={() => onChange(!on)}
      className={cn(
        "flex h-5 w-8.5 shrink-0 items-center rounded-full p-0.5 transition-colors",
        on ? "justify-end bg-shared" : "justify-start bg-toggle-off",
      )}
    >
      <span className="size-4 rounded-full bg-knob" />
    </button>
  )
}
