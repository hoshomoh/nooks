import { Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import { cn } from "cn"

/** Where the switch can send a Member. Only routes that exist are listed. */
export type ViewRoute = "/today" | "/upcoming" | "/calendar"

/** One side of the switch. */
interface ViewChoice {
  to: ViewRoute
  labelKey: string
}

export interface ViewSwitchProps {
  /** The dated view this page belongs to, for the list side of the switch. */
  listRoute: Extract<ViewRoute, "/today" | "/upcoming">
  /** Which side is showing. */
  showing: "list" | "calendar"
}

/**
 * List or Calendar, in the chrome bar of a dated view.
 *
 * Two sides rather than one button, because the calendar is a lens on the same Items
 * rather than somewhere else to go — and a one-way link out of it leaves a Member
 * hunting for the way back.
 */
export function ViewSwitch({ listRoute, showing }: ViewSwitchProps) {
  const { t } = useTranslation()

  const choices: ViewChoice[] = [
    { to: listRoute, labelKey: "views.list" },
    { to: "/calendar", labelKey: "views.calendar" },
  ]

  return (
    <div className="flex items-center gap-0.5">
      {choices.map((choice, index) => (
        <Link
          key={choice.to}
          to={choice.to}
          className={cn(
            "flex h-control-toolbar items-center rounded-md px-2.5 text-micro transition-colors",
            (index === 0) === (showing === "list")
              ? "bg-secondary font-medium text-foreground"
              : "text-secondary-foreground hover:bg-secondary",
          )}
        >
          {t(choice.labelKey)}
        </Link>
      ))}
    </div>
  )
}
