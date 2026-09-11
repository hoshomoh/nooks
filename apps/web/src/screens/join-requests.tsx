import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import type { PendingJoinRequest } from "@nooks/api"

import { Button } from "@/components/ds/button"
import { requestClient } from "@/lib/api"
import { useMomentLabel } from "@/lib/use-moment-label"

export interface JoinRequestsProps {
  requests: PendingJoinRequest[]
}

/**
 * Who is waiting to join, under the Members table.
 *
 * A request shows what the sender wrote, or says plainly that they wrote nothing — an
 * empty message is worth an Admin's suspicion, and hiding its absence would take that
 * signal away.
 *
 * **Ignore is silent.** The sender is never told, so there is nothing to confirm and
 * nothing to undo.
 */
export function JoinRequests({ requests }: JoinRequestsProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const timeOf = useMomentLabel()

  const decide = useMutation({
    mutationFn: ({ requestUid, approve }: DecideVariables) =>
      requestClient.decideJoinRequest({ requestUid, approve }),
    onSuccess: () => queryClient.invalidateQueries(),
  })

  if (requests.length === 0) {
    return null
  }

  return (
    <section className="flex flex-col gap-3 border-t border-border pt-5">
      <div className="flex items-baseline gap-2.5">
        <span className="text-field font-semibold">{t("members.requests")}</span>
        <span className="text-micro text-muted-foreground">
          {t("members.requestsWaiting", { count: requests.length })}
        </span>
      </div>

      {requests.map((request) => (
        <div
          key={request.requestUid}
          className="grid min-h-13 grid-cols-[1fr_auto] items-center gap-4 border-t border-hair"
        >
          <span className="flex min-w-0 items-center gap-3">
            <span className="grid size-7 shrink-0 place-items-center rounded-full bg-chip text-[11px] text-secondary-foreground">
              {initialsOf(request.name)}
            </span>
            <span className="flex min-w-0 flex-col">
              <span className="text-field">{request.name}</span>
              <span className="truncate text-micro text-muted-foreground">
                {request.email} · {request.message || t("members.noMessage")} ·{" "}
                {timeOf(request.createdAt)}
              </span>
            </span>
          </span>

          <span className="flex shrink-0 items-center gap-2">
            <Button
              tone="secondary"
              scale="compact"
              onClick={() =>
                decide.mutate({ requestUid: request.requestUid, approve: false })
              }
            >
              {t("activity.ignore")}
            </Button>
            <Button
              scale="compact"
              onClick={() => decide.mutate({ requestUid: request.requestUid, approve: true })}
            >
              {t("activity.approve")}
            </Button>
          </span>
        </div>
      ))}
    </section>
  )
}

/** What the decision mutation is told. */
interface DecideVariables {
  requestUid: string
  approve: boolean
}

/** initialsOf is the two letters an avatar carries. */
function initialsOf(name: string): string {
  return name.slice(0, 2).toUpperCase()
}
