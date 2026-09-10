import { useTranslation } from "react-i18next"

export interface LoadingRowsProps {
  /** How many rows to hold the space of. */
  count?: number
}

/**
 * What a screen shows while its data is on the way, per DESIGN.md §11.
 *
 * The rows hold the exact height of real ones, so nothing jumps when they arrive. The
 * line about a slow server appears only after four seconds — a home server that has
 * gone to sleep is the common case, and saying so is kinder than an empty page. The
 * delay is in the animation rather than in a timer, so there is nothing to clean up.
 */
export function LoadingRows({ count = 6 }: LoadingRowsProps) {
  const { t } = useTranslation()

  return (
    <div className="flex w-full max-w-content flex-col">
      {Array.from({ length: count }, (_, index) => (
        <div key={index} className="grid min-h-row grid-cols-[20px_1fr] items-center gap-3.5 px-2">
          <span className="size-[17px] rounded-sm bg-secondary" />
          <span
            className="h-3.5 rounded-sm bg-secondary"
            // Rows shorten down the list, so the block reads as text rather than as a
            // table of identical bars.
            style={{ width: `${72 - index * 7}%` }}
          />
        </div>
      ))}

      <p className="mt-6 animate-slow-note text-meta text-muted-foreground">
        {t("app.stillWaiting")}
      </p>
    </div>
  )
}

/** The whole-screen wait: centred, in the same column the content will use. */
export function RoutePending() {
  return (
    <div className="flex min-h-dvh justify-center px-5.5 pt-14">
      <LoadingRows />
    </div>
  )
}
