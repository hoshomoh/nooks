import { cn } from "cn"

import { Checkbox as ShadcnCheckbox } from "@/components/ui/checkbox"

/**
 * The tick box, per DESIGN.md §6.
 *
 * The box is 17px but the hit area is the row's full height, so a thumb on a phone in a
 * shop hits it. The height is on a wrapper rather than the control, because growing the
 * control would push the label out of line — and it grows only downwards and upwards:
 * a 44px square would reach into the label and take its clicks.
 */
export type CheckboxProps = React.ComponentProps<typeof ShadcnCheckbox> & {
  /** Someone else ticked it a moment ago. Holds for a second, then settles. */
  justTicked?: boolean
}

export function Checkbox({ className, justTicked, ...props }: CheckboxProps) {
  return (
    <span className="relative z-10 grid h-row w-full -my-3 place-items-center">
      <ShadcnCheckbox
        className={cn(
          "size-[17px] rounded-sm border-[length:1.5px] border-control",
          "data-[checked]:border-muted-foreground data-[checked]:bg-muted-foreground",
          "focus-visible:ring-0",
          justTicked && "border-done bg-done",
          className,
        )}
        {...props}
      />
    </span>
  )
}
