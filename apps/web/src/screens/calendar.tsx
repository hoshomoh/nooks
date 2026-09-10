import { useSuspenseQuery } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { AppShell } from "@/components/ds/app-shell"
import { ChromeBar } from "@/components/ds/chrome-bar"
import { Button } from "@/components/ds/button"
import { buildMonth, byDay, type CalendarDay } from "@/lib/calendar"
import { datedRangeQuery } from "@/lib/dated-queries"
import { monthHeading, monthWindow, today, toStored } from "@/lib/dates"
import { useCommandPalette } from "@/lib/use-command-palette"
import { useLocale } from "@/lib/use-locale"
import { useSignedInData } from "@/lib/use-signed-in-data"

/** The days of a week, in the order the grid draws them. */
const WEEKDAY_KEYS = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"] as const

/**
 * The calendar: the same dated Items as Upcoming, laid out by day.
 *
 * Undated Items never appear — a calendar is a lens on dates, not a second home for
 * Lists.
 */
export function CalendarScreen() {
  const { instanceName, member, lists } = useSignedInData()
  const { t } = useTranslation()
  const { dateLocale } = useLocale()
  const palette = useCommandPalette()

  const from = today()
  const window = monthWindow(from)
  const dated = useSuspenseQuery(datedRangeQuery(window.start, window.end)).data
  const month = buildMonth(from)
  const itemsByDay = byDay(dated.items, (entry) => entry.item?.dueOn ?? "")
  const currentDay = toStored(from)

  return (
    <AppShell
      instanceName={instanceName}
      memberName={member?.name ?? ""}
      lists={lists}
      onSearch={palette.open}
      onAddList={palette.openAddList}
    >
      <ChromeBar
        crumbs={[t("views.upcoming"), t("views.calendar")]}
        actions={
          <Link to="/upcoming">
            <Button tone="quiet" scale="toolbar">
              {t("views.list")}
            </Button>
          </Link>
        }
      />

      <div className="flex justify-center px-5.5 pt-14 pb-22">
        <div className="w-full max-w-calendar">
          <header className="mb-8.5 flex flex-col gap-3.5">
            <h1 className="text-display">{monthHeading(month.month, dateLocale)}</h1>
            <div className="flex items-center gap-3 text-meta text-secondary-foreground">
              <span>{t("views.calendarSubtitle")}</span>
              <span className="h-3 w-px bg-border" />
              <span>{t("views.datedCount", { count: dated.items.length })}</span>
            </div>
          </header>

          <div className="grid grid-cols-7 border-b border-border">
            {WEEKDAY_KEYS.map((key) => (
              <span key={key} className="px-2 pb-2 text-label uppercase text-muted-foreground">
                {t(`views.weekdays.${key}`)}
              </span>
            ))}
          </div>

          {month.weeks.map((week) => (
            <div key={week.key} className="grid grid-cols-7">
              {week.days.map((day) => (
                <DayCell
                  key={day.date}
                  day={day}
                  isToday={day.date === currentDay}
                  labels={(itemsByDay.get(day.date) ?? []).map((entry) => ({
                    uid: entry.item?.uid ?? "",
                    label: entry.item?.label ?? "",
                    listUid: entry.listUid,
                  }))}
                />
              ))}
            </div>
          ))}
        </div>
      </div>
    </AppShell>
  )
}

export type DayCellLabel = {
  uid: string
  label: string
  listUid: string
}

export type DayCellProps = {
  day: CalendarDay
  isToday: boolean
  labels: DayCellLabel[]
}

/** One day of the grid. Its height is fixed so the weeks stay even. */
function DayCell({ day, isToday, labels }: DayCellProps) {
  return (
    <div
      className={
        isToday
          ? "flex min-h-[104px] flex-col gap-1.5 border-r border-b border-hair bg-secondary p-2"
          : "flex min-h-[104px] flex-col gap-1.5 border-r border-b border-hair p-2"
      }
    >
      <span
        className={
          isToday
            ? "text-micro font-semibold text-shared"
            : day.inMonth
              ? "text-micro text-secondary-foreground"
              : "text-micro text-muted-foreground"
        }
      >
        {day.dayOfMonth}
      </span>
      {labels.map((entry) => (
        <Link
          key={entry.uid}
          to="/lists/$listUid"
          params={{ listUid: entry.listUid }}
          className="truncate text-micro leading-[1.35] hover:underline"
        >
          {entry.label}
        </Link>
      ))}
    </div>
  )
}
