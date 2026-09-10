import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { useNavigate } from "@tanstack/react-router"
import { cn } from "cn"
import type { Activity } from "@nooks/api"
import { ActivityKind } from "@nooks/api"

import { ActivityPanel } from "./activity-panel"
import { activityClient } from "@/lib/api"
import { activityQuery } from "@/lib/activity-queries"
import { useMomentLabel } from "@/lib/use-moment-label"

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

  const follow = (entry: Activity) => {
    setOpen(false)
    if (entry.kind === ActivityKind.LIST_SHARED) {
      void navigate({ to: "/lists/$listUid", params: { listUid: entry.targetUid } })
    }
  }

  return (
    <>
      <button
        type="button"
        onClick={isOpen ? () => setOpen(false) : open}
        className={cn(
          "flex h-control-toolbar items-center gap-1.5 rounded-md px-2.5 text-micro transition-colors",
          isOpen ? "bg-secondary text-foreground" : "text-secondary-foreground hover:bg-secondary",
        )}
      >
        <span>{t("activity.action")}</span>
        {unread > 0 && <span className="size-[5px] rounded-full bg-shared" />}
      </button>

      {isOpen && (
        <ActivityPanel
          activity={activity.data?.activity ?? []}
          timeOf={timeOf}
          onOpen={follow}
          onClose={() => setOpen(false)}
        />
      )}
    </>
  )
}
