import { cn } from "cn"

import { Checkbox as ShadcnCheckbox } from "@/components/ui/checkbox"

/**
 * The tick box, per DESIGN.md §6.
 *
 * The box is 17px but the hit area is 44px — the row's own height — so a thumb on a
 * phone in a shop hits it. That is why the padding is on a wrapper rather than the
 * control: growing the control would push the label out of line.
 */
export function Checkbox({
  className,
  justTicked,
  ...props
}: React.ComponentProps<typeof ShadcnCheckbox> & {
  /** Someone else ticked it a moment ago. Holds for a second, then settles. */
  justTicked?: boolean
}) {
  return (
    <span className="grid size-row -m-3 place-items-center">
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
