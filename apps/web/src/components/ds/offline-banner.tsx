import { useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"

import { Button } from "./button"
import { useConnection } from "@/lib/use-connection"
import { useMomentLabel } from "@/lib/use-moment-label"

/**
 * What the app says when it cannot reach the Instance, per DESIGN.md §11.
 *
 * It reads the connection itself rather than taking it as a prop, so that a screen
 * cannot forget to show it: AppShell places it once and every signed-in view has it.
 *
 * Reading, ticking and adding all still work while this is up. The banner is not a
 * barrier, which is why it is a strip above the content rather than anything modal.
 */
export function OfflineBanner() {
  const { online, lastSeenAt } = useConnection()
  const { t } = useTranslation()
  const formatMoment = useMomentLabel()
  const queryClient = useQueryClient()

  if (online) {
    return null
  }

  // Asking again is the whole of Retry: a refetch that succeeds marks the Instance
  // seen through the ordinary path, and one that fails leaves the banner where it is.
  const retry = () => {
    void queryClient.refetchQueries()
  }

  const seen = lastSeenAt ? formatMoment(lastSeenAt) : ""

  return (
    <div
      role="status"
      aria-live="polite"
      className="flex items-center gap-2.5 border-b border-offline-line bg-offline-bg px-5.5 py-2.5"
    >
      <span aria-hidden className="size-[7px] shrink-0 rounded-full bg-offline" />
      <p className="min-w-0 flex-1 text-meta text-offline-text">
        {seen ? t("app.offlineSince", { seen }) : t("app.offline")}
      </p>
      <Button tone="secondary" scale="compact" onClick={retry}>
        {t("action.retry")}
      </Button>
    </div>
  )
}
