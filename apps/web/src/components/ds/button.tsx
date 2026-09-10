import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "cn"

import { Button as ShadcnButton } from "@/components/ui/button"

/**
 * Nooks' button, per DESIGN.md §7.
 *
 * It wraps the generated shadcn button rather than replacing it, so that Base UI's
 * behaviour — focus handling, disabled state, slots — comes for free and
 * `shadcn add button` can still overwrite the primitive.
 *
 * The base class turns off shadcn's own focus ring: DESIGN.md specifies one focus
 * treatment, a 2px outline at 2px offset, and it is applied globally in index.css.
 */
const nooksButton = cva(
  "font-sans whitespace-nowrap transition-colors focus-visible:ring-0 focus-visible:border-transparent",
  {
    variants: {
      tone: {
        /** Ink on paper. The page's one primary action. */
        primary: "bg-primary text-primary-foreground font-medium hover:bg-primary/90",
        /** A hairline border, for the action beside the primary one. */
        secondary:
          "border border-border bg-transparent text-secondary-foreground hover:bg-secondary",
        /** No border at all, in the shared colour. For links and low-stakes actions. */
        quiet: "border-transparent bg-transparent text-shared hover:bg-secondary",
        /** Outlined, never filled — deleting is not the page's primary action. */
        destructive:
          "border border-destructive-line bg-transparent text-destructive hover:bg-destructive-bg",
      },
      /** Three sizes, and DESIGN.md §7 says where each one is allowed. */
      scale: {
        /** 34px — page and section headers, dialog footers. */
        default: "h-control rounded-lg px-4 text-chrome",
        /** 32px — inside a table row or a settings row. */
        compact: "h-control-compact rounded-lg px-3.5 text-meta",
        /** 26px — the 44px chrome bar only. */
        toolbar: "h-control-toolbar rounded-md px-2.5 text-micro",
      },
    },
    defaultVariants: { tone: "primary", scale: "default" },
  },
)

export type ButtonProps = React.ComponentProps<typeof ShadcnButton> &
  VariantProps<typeof nooksButton>

export function Button({ className, tone, scale, ...props }: ButtonProps) {
  return <ShadcnButton className={cn(nooksButton({ tone, scale }), className)} {...props} />
}
