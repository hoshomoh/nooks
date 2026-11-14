import { Navigate, useRouter, type ErrorComponentProps } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { endedSession } from "@/lib/session-ended"

import { Button } from "./button"
import { EmptyState } from "./empty-state"

/**
 * What a screen shows when its data will not load, per DESIGN.md §11.
 *
 * A plain sentence and a way to try again, rather than a stack trace or a blank page.
 * A home server that was asleep, restarted or is behind a proxy that dropped the
 * connection is the ordinary case, and retrying usually is the fix.
 *
 * A refusal because the session ended is not that case: trying again is refused again,
 * so it goes to sign in. The query client has already let go of the session by the time
 * a failed read reaches here; this only stops the page saying the server did not answer.
 */
export function RouteError({ error }: ErrorComponentProps) {
  const { t } = useTranslation()
  const router = useRouter()

  if (endedSession(error)) {
    return <Navigate to="/sign-in" replace />
  }

  return (
    <div className="flex min-h-dvh justify-center px-5.5 pt-14">
      <div className="flex w-full max-w-content flex-col items-start gap-6">
        <EmptyState title={t("error.screenTitle")} body={t("error.screenBody")} />
        <Button onClick={() => void router.invalidate()}>{t("error.tryAgain")}</Button>
      </div>
    </div>
  )
}
