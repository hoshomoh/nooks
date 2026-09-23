import { useTranslation } from "react-i18next"
import type { PendingJoinRequest, PendingResetRequest } from "@nooks/api"

import { Avatar } from "@/components/ds/avatar"
import { Button } from "@/components/ds/button"
import { useMomentLabel } from "@/lib/use-moment-label"
import { initialsOf } from "@/lib/initials"

export interface PendingRequestsProps {
  joins: PendingJoinRequest[]
  resets: PendingResetRequest[]
  /** Answers somebody asking for an account. */
  onDecideJoin: (requestUid: string, approve: boolean) => void
  /** Answers somebody asking to replace a forgotten password. */
  onDecideReset: (requestUid: string, approve: boolean) => void
}

/**
 * Who is waiting for an answer, under the Members table.
 *
 * Both kinds together. They are the same job — somebody an Admin has to recognise
 * before letting them in — and splitting them would be two rules about which kind
 * appears where. The reset requests used to appear only in the Activity panel, so an
 * Admin who read the panel and cleared it had nowhere left to find one.
 *
 * A join request shows what the sender wrote, or says plainly that they wrote nothing:
 * an empty message is worth an Admin's suspicion, and hiding its absence would take
 * that signal away. A reset request says to check it is really them, because nooks
 * sends no mail and there is nothing else to verify a reset against.
 *
 * **Ignore is silent.** The sender is never told, so there is nothing to confirm and
 * nothing to undo.
 */
export function PendingRequests({
  joins,
  resets,
  onDecideJoin,
  onDecideReset,
}: PendingRequestsProps) {
  const { t } = useTranslation()
  const timeOf = useMomentLabel()

  const waiting = joins.length + resets.length
  if (waiting === 0) {
    return null
  }

  return (
    <section className="flex flex-col gap-3 border-t border-border pt-5">
      <div className="flex items-baseline gap-2.5">
        <span className="text-field font-semibold">{t("members.requests")}</span>
        <span className="text-micro text-muted-foreground">
          {t("members.requestsWaiting", { count: waiting })}
        </span>
      </div>

      {joins.map((request) => (
        <RequestRow
          key={request.requestUid}
          name={request.name}
          detail={`${request.email} · ${request.message || t("members.noMessage")} · ${timeOf(request.createdAt)}`}
          onDecide={(approve) => onDecideJoin(request.requestUid, approve)}
        />
      ))}

      {resets.map((request) => (
        <RequestRow
          key={request.requestUid}
          name={request.member?.name ?? ""}
          detail={`${request.member?.email ?? ""} · ${t("members.resetAsked")} · ${timeOf(request.createdAt)}`}
          onDecide={(approve) => onDecideReset(request.requestUid, approve)}
        />
      ))}
    </section>
  )
}

interface RequestRowProps {
  name: string
  /** The second line: their address, what they want, and when they asked. */
  detail: string
  onDecide: (approve: boolean) => void
}

/** One person waiting, whichever they are waiting for. */
function RequestRow({ name, detail, onDecide }: RequestRowProps) {
  const { t } = useTranslation()

  return (
    <div className="grid min-h-13 grid-cols-[1fr_auto] items-center gap-4 border-t border-hair">
      <span className="flex min-w-0 items-center gap-3">
        <Avatar badge={initialsOf(name)} size="large" />
        <span className="flex min-w-0 flex-col">
          <span className="text-field">{name}</span>
          <span className="truncate text-micro text-muted-foreground">{detail}</span>
        </span>
      </span>

      <span className="flex shrink-0 items-center gap-2">
        <Button tone="secondary" scale="compact" onClick={() => onDecide(false)}>
          {t("activity.ignore")}
        </Button>
        <Button scale="compact" onClick={() => onDecide(true)}>
          {t("activity.approve")}
        </Button>
      </span>
    </div>
  )
}
