import { Fragment } from "react"
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
  /** Opens the Members screen, where every waiting request is listed. */
  onSeeRequests: () => void
  /** The request being decided right now, if any. */
  deciding?: string
}

/**
 * How many waiting requests the panel offers to decide before sending the rest on.
 *
 * One person asking is the ordinary case, and it keeps one-click Approve where it is.
 * A queueful is somebody with a script, and drawing all of them buries everything else
 * that was waiting.
 */
const REQUESTS_SHOWN = 3

/**
 * The Activity panel, per DESIGN.md §9.
 *
 * nooks has no mail server, so this is the only place a join request, a reset request
 * or a share surfaces. That is why every entry carries its own words rather than a
 * code, and why the panel says so at the bottom: a Member who is waiting for an email
 * should find out here that none is coming.
 *
 * The entries scroll and the two lines around them do not. Without that the popover
 * grows to fit fifty entries, and the footer saying no mail is coming ends up below the
 * bottom of the screen with nothing to scroll it back.
 */
export function ActivityPanel({
  activity,
  timeOf,
  onOpen,
  onDecide,
  onSeeRequests,
  deciding,
}: ActivityPanelProps) {
  const { t } = useTranslation()
  const { hidden, lineAfter, moreWaiting } = collapse(activity)

  return (
    <div className="flex flex-col">
      <header className="flex items-baseline justify-between px-2.5 pt-2 pb-2.5">
        <span className="text-meta font-semibold">{t("activity.title")}</span>
        <span className="text-micro text-muted-foreground">{t("activity.inAppOnly")}</span>
      </header>

      {activity.length === 0 ? (
        <p className="px-2.5 py-3 text-meta text-muted-foreground">{t("activity.empty")}</p>
      ) : (
        <div className="flex max-h-(--size-floating-list) flex-col overflow-y-auto">
          {activity.map((entry) =>
            hidden.has(entry.uid) ? null : (
              <Fragment key={entry.uid}>
                <Entry
                  entry={entry}
                  when={timeOf(entry.createdAt)}
                  onOpen={onOpen}
                  onDecide={onDecide}
                  deciding={deciding === entry.uid}
                />
                {entry.uid === lineAfter && (
                  <button
                    type="button"
                    onClick={onSeeRequests}
                    className="px-2.5 py-2 text-left text-micro text-shared hover:underline"
                  >
                    {t("activity.moreRequests", { count: moreWaiting })}
                  </button>
                )}
              </Fragment>
            ),
          )}
        </div>
      )}

      <p className="mt-1.5 border-t border-hair px-2.5 py-2.5 text-micro text-muted-foreground">
        {t("activity.noMailServer")}
      </p>
    </div>
  )
}

/** Which entries the panel holds back, and where the line standing for them goes. */
interface Collapsed {
  /** Entries not drawn, by identifier. */
  hidden: Set<string>
  /** The entry the standing-in line follows, empty when nothing is held back. */
  lineAfter: string
  /** How many are held back. */
  moreWaiting: number
}

/**
 * Works out which waiting join requests the panel keeps.
 *
 * Only join requests, because the line stands in for them by opening the Members
 * screen, which is where they are listed. A reset request has no such screen yet and
 * cannot be flooded either: one waiting reset per Member bounds it by how many Members
 * there are.
 */
function collapse(activity: Activity[]): Collapsed {
  const waiting = activity.filter(isWaitingJoin)
  const held = waiting.slice(REQUESTS_SHOWN)
  return {
    hidden: new Set(held.map((entry) => entry.uid)),
    lineAfter: held.length === 0 ? "" : waiting[REQUESTS_SHOWN - 1].uid,
    moreWaiting: held.length,
  }
}

/** isWaitingJoin reports whether an entry is somebody asking for an account, undecided. */
function isWaitingJoin(entry: Activity): boolean {
  return entry.kind === ActivityKind.JOIN_REQUEST && entry.outcome === ActivityOutcome.UNSPECIFIED
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
        "grid shrink-0 grid-cols-[6px_1fr] items-start gap-2.5 rounded-md p-2.5",
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
            followable(entry.kind) && (
              <Action label={t(openLabelOf(entry.kind))} onClick={() => onOpen(entry)} />
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

/** followable reports whether an entry points somewhere a Member can go. */
function followable(kind: ActivityKind): boolean {
  return kind === ActivityKind.LIST_SHARED || kind === ActivityKind.TOKEN_USED
}

/** openLabelOf names where following an entry leads, rather than saying "open". */
function openLabelOf(kind: ActivityKind): string {
  return kind === ActivityKind.TOKEN_USED ? "activity.openTokens" : "activity.open"
}

/** isRequest reports whether an entry is somebody waiting on an Admin's decision. */
function isRequest(kind: ActivityKind): boolean {
  return kind === ActivityKind.JOIN_REQUEST || kind === ActivityKind.RESET_REQUEST
}
