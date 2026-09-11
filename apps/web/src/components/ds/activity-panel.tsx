import { useTranslation } from "react-i18next"
import { cn } from "cn"
import { ActivityKind, ActivityOutcome, type Activity } from "@nooks/api"

export interface ActivityPanelProps {
  activity: Activity[]
  /** How each entry's time reads, already in the Member's language. */
  timeOf: (createdAt: string) => string
  /** Follows an entry to whatever it points at. */
  onOpen: (entry: Activity) => void
  /** Decides a request. Ignoring is silent and never tells the sender. */
  onDecide: (entry: Activity, approve: boolean) => void
  /** The request being decided right now, if any. */
  deciding?: string
}

/**
 * The Activity panel, per DESIGN.md §9.
 *
 * Nooks has no mail server, so this is the only place a join request, a reset request
 * or a share surfaces. That is why every entry carries its own words rather than a
 * code, and why the panel says so at the bottom: a Member who is waiting for an email
 * should find out here that none is coming.
 */
export function ActivityPanel({
  activity,
  timeOf,
  onOpen,
  onDecide,
  deciding,
}: ActivityPanelProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col">
      <header className="flex items-baseline justify-between px-2.5 pt-2 pb-2.5">
        <span className="text-meta font-semibold">{t("activity.title")}</span>
        <span className="text-micro text-muted-foreground">{t("activity.inAppOnly")}</span>
      </header>

      {activity.length === 0 ? (
        <p className="px-2.5 py-3 text-meta text-muted-foreground">{t("activity.empty")}</p>
      ) : (
        activity.map((entry) => (
          <Entry
            key={entry.uid}
            entry={entry}
            when={timeOf(entry.createdAt)}
            onOpen={onOpen}
            onDecide={onDecide}
            deciding={deciding === entry.uid}
          />
        ))
      )}

      <p className="mt-1.5 border-t border-hair px-2.5 py-2.5 text-micro text-muted-foreground">
        {t("activity.noMailServer")}
      </p>
    </div>
  )
}

interface EntryProps {
  entry: Activity
  /** When it happened, in words. */
  when: string
  onOpen: (entry: Activity) => void
  onDecide: (entry: Activity, approve: boolean) => void
  /** Whether this entry is the one being decided right now. */
  deciding: boolean
}

/** One thing waiting for attention. */
function Entry({ entry, when, onOpen, onDecide, deciding }: EntryProps) {
  const { t } = useTranslation()

  return (
    <div
      className={cn(
        "grid grid-cols-[6px_1fr] items-start gap-2.5 rounded-md p-2.5",
        entry.unread && "bg-secondary",
      )}
    >
      <span
        className={cn("mt-1.5 size-[5px] rounded-full", entry.unread ? "bg-shared" : "bg-transparent")}
      />
      <div className="flex flex-col gap-1">
        <span className="text-chrome leading-[1.45]">{entry.text}</span>
        <span className="flex items-center gap-2.5">
          <span className="text-micro text-muted-foreground">{when}</span>

          {/* Once anybody has decided, the entry says what happened instead of
              offering to decide again — whichever Admin got there first. */}
          {isRequest(entry.kind) && entry.outcome !== ActivityOutcome.UNSPECIFIED ? (
            <span className="text-micro text-muted-foreground">
              {t(
                entry.outcome === ActivityOutcome.APPROVED
                  ? "activity.approved"
                  : "activity.ignored",
              )}
            </span>
          ) : isRequest(entry.kind) ? (
            <>
              <Action
                label={deciding ? t("activity.deciding") : t("activity.approve")}
                disabled={deciding}
                onClick={() => onDecide(entry, true)}
              />
              {/* Ignoring is silent: the sender is never told, so there is nothing to
                  confirm and nothing to undo. */}
              <Action
                label={t("activity.ignore")}
                quiet
                disabled={deciding}
                onClick={() => onDecide(entry, false)}
              />
            </>
          ) : (
            entry.kind === ActivityKind.LIST_SHARED && (
              <Action label={t("activity.open")} onClick={() => onOpen(entry)} />
            )
          )}
        </span>
      </div>
    </div>
  )
}

interface ActionProps {
  label: string
  /** A second action beside the first one, in muted text rather than the accent. */
  quiet?: boolean
  /** While the answer is on its way, so a Member cannot send it twice. */
  disabled?: boolean
  onClick: () => void
}

/**
 * One thing an entry offers to do about itself.
 *
 * Only what can actually be acted on gets a word: an entry with nothing to do is a
 * statement, and giving it a button would be a button that goes nowhere.
 */
function Action({ label, quiet, disabled, onClick }: ActionProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className={cn(
        "text-micro transition-opacity hover:underline disabled:opacity-50",
        quiet ? "text-muted-foreground" : "text-shared",
      )}
    >
      {label}
    </button>
  )
}

/** isRequest reports whether an entry is somebody waiting on an Admin's decision. */
function isRequest(kind: ActivityKind): boolean {
  return kind === ActivityKind.JOIN_REQUEST || kind === ActivityKind.RESET_REQUEST
}
