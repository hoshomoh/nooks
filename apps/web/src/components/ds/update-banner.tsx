import { useTranslation } from "react-i18next"
import { cn } from "cn"

import { Button } from "./button"
import { useAppUpdate } from "@/lib/use-app-update"

/**
 * What the app says when a newer build is waiting.
 *
 * The same strip as the offline banner, and for the same reason: it pushes the page
 * down rather than covering it, and it is never modal.
 *
 * It asks rather than reloading, because a Note is saved when the typing settles.
 */
export function UpdateBanner() {
  const { waiting, apply } = useAppUpdate()
  const { t } = useTranslation()

  return (
    <div
      className={cn(
        "grid transition-[grid-template-rows] duration-(--duration-push) ease-sheet",
        waiting ? "grid-rows-[1fr]" : "grid-rows-[0fr]",
      )}
    >
      <div className="overflow-hidden" inert={!waiting}>
        <div className="flex items-center gap-2.5 border-b border-shared-line bg-shared-bg px-5.5 py-2.5">
          <span aria-hidden className="size-[7px] shrink-0 rounded-full bg-shared" />
          <p role="status" aria-live="polite" className="min-w-0 flex-1 text-meta text-shared">
            {waiting ? t("app.updateReady") : ""}
          </p>
          <Button tone="secondary" scale="compact" onClick={apply}>
            {t("app.updateReload")}
          </Button>
        </div>
      </div>
    </div>
  )
}
