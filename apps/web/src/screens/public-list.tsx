import { useState, type ReactNode } from "react"
import { useSuspenseQuery } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import { cn } from "cn"
import type { PublicItem } from "@nooks/api"

import { Button } from "@/components/ds/button"
import { EmptyState } from "@/components/ds/empty-state"
import { Mark } from "@nooks/design/mark"
import { pendingTickStore } from "@/lib/pending-tick-store"
import { publicListQuery } from "@/lib/public-queries"
import { useDueLabel } from "@/lib/use-due-label"
import { useMomentLabel } from "@/lib/use-moment-label"

/**
 * The public page: one List, no account, no sidebar.
 *
 * A Visitor is not a Member with fewer buttons — they are somebody who was sent a link
 * to a shopping list. So there is no sidebar, no search, no other List, and nothing on
 * the page that hints at what else the Instance holds. The only thing here is the List
 * they were sent.
 */
export function PublicListScreen() {
  const { t } = useTranslation()
  const page = useSuspenseQuery(publicListQuery).data
  const due = useDueLabel()
  const moment = useMomentLabel()

  // Which row a Visitor reached for. The prompt appears under that row rather than as
  // a wall across the page: they were trying to do one thing, and being told "sign in"
  // before they touched anything would be the page refusing a request nobody made.
  const [reachedFor, setReachedFor] = useState<number | null>(null)

  if (!page.published) {
    return (
      <PublicShell instanceName="">
        <EmptyState title={t("public.nothingTitle")} body={t("public.nothingBody")} />
      </PublicShell>
    )
  }

  return (
    <PublicShell instanceName={page.instanceName} allowJoin={page.allowJoin}>
      <header className="mb-8 flex flex-col gap-3">
        <h1 className="text-display">{page.listName}</h1>
        <div className="flex items-center gap-3 text-meta text-secondary-foreground">
          <span>{t("public.openCount", { count: page.openCount })}</span>
          {page.updatedAt && (
            <>
              <span className="h-3 w-px bg-border" />
              <span>{t("public.updated", { when: moment(page.updatedAt) })}</span>
            </>
          )}
          <span className="h-3 w-px bg-border" />
          <span>{t("public.readOnly")}</span>
        </div>
      </header>

      {page.items.length === 0 ? (
        <EmptyState
          title={t("public.emptyTitle")}
          body={t("public.emptyBody", { name: page.instanceName })}
        />
      ) : (
        <div className="flex flex-col">
          {page.items.map((item, index) => (
            <div key={`${item.label}-${index}`} className="flex flex-col">
              <PublicRow
                item={item}
                dueLabel={due.label(item.dueOn)}
                reached={reachedFor === index}
                onReach={() => {
                  // What they meant to do, kept until they have an account to do it
                  // with.
                  pendingTickStore.remember(item.uid)
                  setReachedFor(index)
                }}
              />
              {reachedFor === index && (
                <SignInPrompt
                  label={item.label}
                  allowJoin={page.allowJoin}
                  instanceName={page.instanceName}
                  onDismiss={() => setReachedFor(null)}
                />
              )}
            </div>
          ))}
        </div>
      )}

      {/* What a Visitor cannot do here, said once at the foot rather than beside every
          row they cannot use. */}
      <div className="mt-6.5 flex flex-wrap items-center gap-3.5 border-t border-hair pt-4.5 text-meta">
        <span className="text-secondary-foreground">{t("public.needsAccount")}</span>
        {page.allowJoin && (
          <Link to="/join" className="text-shared hover:underline">
            {t("public.askToJoinInstance", { name: page.instanceName })}
          </Link>
        )}
      </div>
    </PublicShell>
  )
}

interface PublicRowProps {
  item: PublicItem
  /** The date as it reads, or empty when the Instance does not show dates. */
  dueLabel: string
  /** Whether this is the row the prompt below belongs to. */
  reached: boolean
  onReach: () => void
}

/**
 * One line of the public list.
 *
 * The box is a real button even though a Visitor cannot tick anything: reaching for it
 * is how they find out they need an account, and a box that ignores the pointer would
 * leave them clicking at nothing.
 */
function PublicRow({ item, dueLabel, reached, onReach }: PublicRowProps) {
  return (
    <div
      className={cn(
        "grid min-h-row grid-cols-[20px_1fr_auto] items-center gap-3.5 rounded-md border px-2 py-1.5 -mx-2",
        // The row they touched is what the prompt is about, so the row is what is
        // marked. The prompt itself stays quiet: two accents would be two questions.
        reached ? "border-[length:1.5px] border-shared bg-shared-bg" : "border-transparent",
      )}
    >
      <button
        type="button"
        onClick={onReach}
        aria-label={item.label}
        className={cn(
          "size-[17px] rounded-sm border-[length:1.5px]",
          // Lighter than a Member's, because it is not interactive in the way theirs is.
          item.done ? "border-muted-foreground bg-muted-foreground" : "border-toggle-off",
          reached && !item.done && "border-shared",
        )}
      />

      <span className="flex min-w-0 items-baseline gap-2.5">
        <span className={cn("truncate text-body", item.done && "text-muted-foreground line-through")}>
          {item.label}
        </span>
        {item.quantity && (
          <span className="shrink-0 rounded-sm border border-border px-1.5 py-px font-mono text-[11.5px] text-muted-foreground">
            {item.quantity}
          </span>
        )}
      </span>

      <span className="flex items-center gap-3 whitespace-nowrap text-micro text-muted-foreground">
        {dueLabel && <span className="text-shared">{dueLabel}</span>}
        {item.addedByName && <span>{item.addedByName}</span>}
      </span>
    </div>
  )
}

interface SignInPromptProps {
  /** The Item they reached for, named back to them. */
  label: string
  allowJoin: boolean
  instanceName: string
  onDismiss: () => void
}

/**
 * What a Visitor is told when they reach for a row.
 *
 * Under the row they touched, not across the page: it answers the thing they just
 * tried, and leaves everything else readable.
 */
function SignInPrompt({ label, allowJoin, instanceName, onDismiss }: SignInPromptProps) {
  const { t } = useTranslation()

  return (
    <div className="mt-2.5 mb-4.5 flex flex-col gap-3.5 rounded-lg border border-border bg-sidebar px-5 py-4.5 -mx-2">
      <div className="flex flex-col gap-1">
        <span className="text-field font-medium">{t("public.signInToTick", { label })}</span>
        <span className="text-meta leading-[1.6] text-secondary-foreground">
          {t("public.signInBlurb", { name: instanceName })}
        </span>
      </div>

      <div className="flex flex-wrap items-center gap-2.5">
        <Link to="/sign-in">
          <Button scale="compact">{t("public.signIn")}</Button>
        </Link>
        {allowJoin && (
          <Link to="/join">
            <Button tone="secondary" scale="compact">
              {t("public.askToJoin")}
            </Button>
          </Link>
        )}
        <span className="flex-1" />
        <button
          type="button"
          onClick={onDismiss}
          className="text-meta text-muted-foreground transition-colors hover:text-foreground"
        >
          {t("public.notNow")}
        </button>
      </div>

      {/* Not a promise about signing in — a statement about what already happened. The
          tick is in this browser's storage the moment they reach for the row. */}
      <span className="text-micro text-muted-foreground">{t("public.tickRemembered")}</span>
    </div>
  )
}

interface PublicShellProps {
  instanceName: string
  /** Whether the Instance takes requests for an account from this page. */
  allowJoin?: boolean
  children: ReactNode
}

/** The page a Visitor sees: a mark, a name, and the List. Nothing else. */
function PublicShell({ instanceName, allowJoin, children }: PublicShellProps) {
  const { t } = useTranslation()

  return (
    <div className="flex min-h-dvh flex-col bg-background">
      <header className="flex h-chrome-auth items-center gap-3 border-b border-hair px-6">
        <Mark size={20} />
        <span className="text-chrome font-medium">{instanceName}</span>
        {instanceName && (
          <span className="text-micro text-muted-foreground">{t("public.label")}</span>
        )}
        <span className="flex-1" />
        <Link to="/sign-in">
          <Button tone="quiet" scale="compact">
            {t("public.signIn")}
          </Button>
        </Link>
        {allowJoin && (
          <Link to="/join">
            <Button scale="compact">{t("public.askToJoin")}</Button>
          </Link>
        )}
      </header>

      <div className="flex flex-1 justify-center px-5.5 pt-14 pb-22">
        <div className="w-full max-w-content">{children}</div>
      </div>
    </div>
  )
}
