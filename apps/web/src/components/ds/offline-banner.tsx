import { useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { cn } from "cn"

import { Button } from "./button"
import { useConnection } from "@/lib/use-connection"
import { useMomentLabel } from "@/lib/use-moment-label"

/**
 * What the app says when it cannot reach the Instance, per DESIGN.md §11.
 *
 * It reads the connection itself rather than taking a prop, so a screen cannot forget
 * to show it. Reading, ticking and adding still work while it is up, which is why it is
 * a strip rather than anything modal.
 *
 * The strip is always in the page and opens to its own height. Moving the height rather
 * than inserting the element keeps the app from shoving down a row under a Member's
 * cursor, and a live region that was already there when its text changed is announced
 * by every screen reader, where one that arrives carrying its message is not.
 */
export function OfflineBanner() {
  const { online, lastSeenAt } = useConnection()
  const { t } = useTranslation()
  const formatMoment = useMomentLabel()
  const queryClient = useQueryClient()

  // Asking again is the whole of Retry: a refetch that succeeds marks the Instance
  // seen through the ordinary path, and one that fails leaves the banner where it is.
  const retry = () => {
    void queryClient.refetchQueries()
  }

  const seen = lastSeenAt ? formatMoment(lastSeenAt) : ""
  const trouble = seen ? t("app.offlineSince", { seen }) : t("app.offline")

  return (
    <div
      className={cn(
        "grid transition-[grid-template-rows] duration-(--duration-push) ease-sheet",
        online ? "grid-rows-[0fr]" : "grid-rows-[1fr]",
      )}
    >
      {/* Clipped while closed, and inert with it: a Retry the Member cannot see is
          still a stop on the way to the content if the keyboard can reach it. */}
      <div className="overflow-hidden" inert={online}>
        <div className="flex items-center gap-2.5 border-b border-offline-line bg-offline-bg px-5.5 py-2.5">
          <span aria-hidden className="size-[7px] shrink-0 rounded-full bg-offline" />
          {/* The region stays; its message does not. */}
          <p
            role="status"
            aria-live="polite"
            className="min-w-0 flex-1 text-meta text-offline-text"
          >
            {online ? "" : trouble}
          </p>
          <Button tone="secondary" scale="compact" onClick={retry}>
            {t("action.retry")}
          </Button>
        </div>
      </div>
    </div>
  )
}
