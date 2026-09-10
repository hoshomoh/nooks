import type { ReactNode } from "react"

import { Mark } from "@/components/mark"

type AuthShellProps = {
  /** Small line above the title, e.g. "Signing in to" or "Anna approved your request". */
  eyebrow?: string
  title: string
  /** One paragraph of orientation under the title. */
  blurb: string
  children: ReactNode
  /** A quiet link on the left of the footer, e.g. "Back to sign in". */
  footerLeft?: ReactNode
  /** A statement of fact on the right, e.g. "No email, ever". */
  footerRight?: ReactNode
}

/**
 * The signed-out layout: a 52px chrome bar, then a centred column.
 *
 * Every screen in Nook-Auth-Onboarding shares this shape, so it is one component
 * rather than a shape repeated five times.
 */
export function AuthShell({
  eyebrow,
  title,
  blurb,
  children,
  footerLeft,
  footerRight,
}: AuthShellProps) {
  return (
    <div className="bg-background min-h-dvh">
      <header className="border-hair flex h-[52px] items-center gap-3 border-b px-6">
        <Mark size={20} />
        <span className="text-[14px] font-medium tracking-[-0.02em]">nooks</span>
      </header>

      <main className="mx-auto flex w-full max-w-[420px] flex-col gap-8 px-6 pt-20 pb-16">
        <div className="flex flex-col gap-3">
          {eyebrow && <p className="text-muted-foreground text-[13px]">{eyebrow}</p>}
          <h1 className="text-[33px] leading-[1.1] font-semibold tracking-[-0.025em]">{title}</h1>
          <p className="text-secondary-foreground text-[15px] leading-[1.65]">{blurb}</p>
        </div>

        {children}

        {(footerLeft || footerRight) && (
          <div className="border-hair flex items-center gap-4 border-t pt-5 text-[13px]">
            {footerLeft}
            <span className="flex-1" />
            <span className="text-muted-foreground">{footerRight}</span>
          </div>
        )}
      </main>
    </div>
  )
}
