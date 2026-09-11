import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { useNavigate } from "@tanstack/react-router"
import { cn } from "cn"
import type { Activity } from "@nooks/api"
import { ActivityKind } from "@nooks/api"

import { ActivityPanel } from "./activity-panel"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { activityClient, requestClient } from "@/lib/api"
import { activityQuery } from "@/lib/activity-queries"
import { useMomentLabel } from "@/lib/use-moment-label"

/** What the decision mutation is told. */
interface DecideVariables {
  entry: Activity
  approve: boolean
}

/**
 * The Activity control in the chrome bar, and the panel behind it.
 *
 * The dot is the only unread indicator in Nooks: there is no mail, no badge on a tab
 * and no toast, so this is where everything waits.
 */
export function ActivityControl() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [isOpen, setOpen] = useState(false)
  const timeOf = useMomentLabel()

  const activity = useQuery(activityQuery)
  const unread = activity.data?.unreadCount ?? 0

  // Opening the panel is reading it: what is on screen has been seen.
  const markRead = useMutation({
    mutationFn: () => activityClient.markActivityRead({}),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["activity"] }),
  })

  const open = () => {
    setOpen(true)
    if (unread > 0) {
      markRead.mutate()
    }
  }

  // Approving a request and ignoring it are the same call with a different answer.
  // Ignoring is silent: nothing is sent, and the sender is never told.
  const decide = useMutation({
    // Neither call answers with anything, so neither answer is kept.
    mutationFn: async ({ entry, approve }: DecideVariables): Promise<void> => {
      const requestUid = entry.targetUid
      if (entry.kind === ActivityKind.JOIN_REQUEST) {
        await requestClient.decideJoinRequest({ requestUid, approve })
        return
      }
      await requestClient.decideResetRequest({ requestUid, approve })
    },
    // Deciding a request changes who is a Member, so the panel is not the only thing
    // that has to catch up.
    onSuccess: () => queryClient.invalidateQueries(),
  })

  const follow = (entry: Activity) => {
    setOpen(false)
    if (entry.kind === ActivityKind.LIST_SHARED) {
      void navigate({ to: "/lists/$listUid", params: { listUid: entry.targetUid } })
    }
    // An entry about a key points at the page where keys are managed. The token itself
    // has no page of its own, and never will: there is nothing to read.
    if (entry.kind === ActivityKind.TOKEN_USED) {
      void navigate({ to: "/settings/tokens" })
    }
  }

  return (
    <Popover open={isOpen} onOpenChange={(next) => (next ? open() : setOpen(false))}>
      <PopoverTrigger
        render={
          <button
            type="button"
            className={cn(
              "flex h-control-toolbar items-center gap-1.5 rounded-md px-2.5 text-micro transition-colors",
              isOpen
                ? "bg-secondary text-foreground"
                : "text-secondary-foreground hover:bg-secondary",
            )}
          >
            <span>{t("activity.action")}</span>
            {unread > 0 && <span className="size-[5px] rounded-full bg-shared" />}
          </button>
        }
      />

      {/* A popover rather than a panel positioned by hand: the browser closes it on a
          click outside or Escape, and a portal puts it above the side sheet rather
          than beside it in the same stacking context. */}
      <PopoverContent
        align="end"
        sideOffset={6}
        className="w-95 gap-0 rounded-2xl border border-border p-2 shadow-[0_12px_32px_rgba(0,0,0,0.14)]"
      >
        <ActivityPanel
          activity={activity.data?.activity ?? []}
          timeOf={timeOf}
          onOpen={follow}
          onDecide={(entry, approve) => decide.mutate({ entry, approve })}
          deciding={decide.isPending ? decide.variables?.entry.uid : undefined}
        />
      </PopoverContent>
    </Popover>
  )
}
