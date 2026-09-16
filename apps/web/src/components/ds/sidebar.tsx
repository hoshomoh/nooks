import { useState } from "react"
import { Link, useRouterState } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import { cn } from "cn"
import { Sharing, type GetSidebarResponse, type List, type SidebarGroup } from "@nooks/api"

import { COVERING, INERT, RAISED } from "./covering"
import { Icon } from "./icon"
import { ListActions } from "./list-actions"
import { SETTINGS_HOME } from "./settings-sections"
import { reachableCount } from "@/lib/sidebar-groups"

import { Mark } from "@nooks/design/mark"

export type SidebarViewCounts = {
  /** How many Items are overdue or due today. */
  today: number
  /** How many are due in the next two weeks. */
  upcoming: number
}

/** The three destinations above the Lists. */
export type ViewTarget = "/today" | "/upcoming" | "/"

export type SidebarProps = {
  instanceName: string
  memberName: string
  /** The four groups, as the server capped them. */
  groups: GetSidebarResponse
  /** The List currently open, so it can be marked. */
  activeListUid?: string
  /** The numbers beside Today and Upcoming. */
  counts?: SidebarViewCounts
  onSearch: () => void
  onAddList: () => void
}

/**
 * The sidebar, per DESIGN.md §5: 258px, grouped as Pinned, My lists and Shared with me.
 *
 * Every group is capped by the server and says how many there are, so the column is the
 * same handful of rows whether a Member has ten Lists or ten thousand. The rest are one
 * row away, in All lists.
 */
export function Sidebar({
  instanceName,
  memberName,
  groups,
  activeListUid,
  counts,
  onSearch,
  onAddList,
}: SidebarProps) {
  const { t } = useTranslation()
  const path = useRouterState({ select: (state) => state.location.pathname })
  const active = viewForPath(path)

  return (
    <aside
      // A full-height column with its own scroller. The shell is the window's height,
      // so this stays where it is and a long list of Lists scrolls inside it rather
      // than taking the views and the account off the top of the screen.
      className="flex h-full w-[258px] shrink-0 flex-col gap-4.5 overflow-y-auto border-r border-border bg-sidebar px-2 pt-3 pb-4"
    >
      <div className="flex h-8.5 items-center gap-2.5 px-2">
        <Mark size={20} />
        <span className="truncate text-chrome font-medium">{instanceName}</span>
      </div>

      <div className="flex flex-col gap-px">
        <button
          type="button"
          onClick={onSearch}
          className="flex h-7.5 items-center rounded-md px-2 text-meta text-muted-foreground hover:bg-secondary"
        >
          {t("action.search")}
          <span className="ml-auto rounded-sm border border-border px-1.5 py-px font-mono text-keycap">
            ⌘K
          </span>
        </button>
      </div>

      <div className="flex flex-col gap-px">
        <ViewLink
          to="/today"
          label={t("views.today")}
          count={counts?.today}
          active={active === "/today"}
          accent
        />
        <ViewLink
          to="/upcoming"
          label={t("views.upcoming")}
          count={counts?.upcoming}
          active={active === "/upcoming"}
        />
        <ViewLink
          to="/"
          label={t("list.allLists")}
          count={reachableCount(groups)}
          active={active === "/"}
        />
      </div>

      <ListGroup
        label={t("sidebar.pinned")}
        group={groups.pinned}
        activeListUid={activeListUid}
        instanceName={instanceName}
      />
      <ListGroup
        label={t("sidebar.myLists")}
        group={groups.mine}
        activeListUid={activeListUid}
        instanceName={instanceName}
      />
      <ListGroup
        label={t("sidebar.sharedWithMe")}
        group={groups.shared}
        activeListUid={activeListUid}
        instanceName={instanceName}
      />
      <CompletedGroup group={groups.completed} activeListUid={activeListUid} />

      <button
        type="button"
        onClick={onAddList}
        className="flex h-7.5 items-center gap-2.5 rounded-md px-2 text-meta text-muted-foreground hover:bg-secondary"
      >
        <span>+</span>
        <span>{t("sidebar.addList")}</span>
      </button>

      <Link
        to={SETTINGS_HOME}
        className="mt-auto flex h-8.5 items-center gap-2.5 rounded-md px-2 transition-colors hover:bg-secondary"
      >
        <span className="grid size-5.5 place-items-center rounded-full bg-chip text-[10px] text-secondary-foreground">
          {initialsOf(memberName)}
        </span>
        <span className="truncate text-small text-secondary-foreground">{memberName}</span>
        <span className="ml-auto text-micro text-muted-foreground">{t("settings.title")}</span>
      </Link>
    </aside>
  )
}

/**
 * viewForPath says which view a path belongs to.
 *
 * The calendar is a lens on the same dated Items as Upcoming rather than a place of its
 * own, so it keeps Upcoming marked. Deciding it here rather than in each link is what
 * keeps the rule in one readable place.
 */
function viewForPath(path: string): ViewTarget | null {
  if (path === "/today") {
    return "/today"
  }
  if (path === "/upcoming" || path === "/calendar") {
    return "/upcoming"
  }
  return path === "/" ? "/" : null
}

type ViewLinkProps = {
  to: ViewTarget
  label: string
  count?: number
  /** Whether this is the view being read. */
  active: boolean
  /** Today's count is the one number in the sidebar that takes the shared colour. */
  accent?: boolean
}

/** One of the three views above the Lists. */
function ViewLink({ to, label, count, active, accent }: ViewLinkProps) {
  return (
    <Link
      to={to}
      className={cn(
        "flex h-7.5 items-center rounded-md px-2 text-meta text-secondary-foreground",
        "transition-colors hover:bg-secondary",
        active && "bg-secondary font-medium text-foreground",
      )}
    >
      {label}
      {count !== undefined && count > 0 && (
        <span className={accent ? "ml-auto text-badge text-shared" : "ml-auto text-badge text-muted-foreground"}>
          {count}
        </span>
      )}
    </Link>
  )
}

interface SeeAllProps {
  /** How many there are altogether, which is what the row offers to show. */
  total: number
  /** Which filter All lists should open on. */
  status?: "completed"
}

/** The row under a group the server had to cut short. */
function SeeAll({ total, status }: SeeAllProps) {
  const { t } = useTranslation()

  return (
    <Link
      to="/"
      search={status ? { status } : {}}
      className="flex h-7.5 items-center rounded-md px-2 text-badge text-shared hover:bg-secondary"
    >
      {t("sidebar.seeAll", { count: total })}
    </Link>
  )
}

interface CompletedGroupProps {
  group?: SidebarGroup
  activeListUid?: string
}

/**
 * The finished Lists, last and open.
 *
 * They leave My lists rather than sitting in both, so the groups above stay what a
 * Member is working on. Open by default because a List finished this morning is still
 * part of this week — it is the ones from March that nobody needs in front of them, and
 * those are behind the count.
 *
 * No `···` on these rows. What a Member does to a finished List — rename it, share it,
 * delete it — is done from All lists, where the row carries the menu.
 */
function CompletedGroup({ group, activeListUid }: CompletedGroupProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(true)

  const lists = group?.lists ?? []
  if (lists.length === 0) {
    return null
  }
  const total = group?.total ?? lists.length

  return (
    <div className="flex flex-col gap-0.5">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        aria-expanded={open}
        aria-label={t("sidebar.toggleCompleted")}
        className={cn(
          "flex h-6.5 items-center gap-1 rounded-md px-2 text-badge font-semibold tracking-[0.03em]",
          "text-muted-foreground transition-colors hover:bg-secondary",
        )}
      >
        <Icon name={open ? "collapse" : "expand"} size="small" className="size-3" />
        <span>{t("sidebar.completed")}</span>
        <span className="ml-auto font-normal">{total}</span>
      </button>

      {open && (
        <>
          {lists.map((list) => (
            <div
              key={list.uid}
              className={cn(
                "relative flex h-7.5 items-center gap-2.5 rounded-md px-2",
                list.uid === activeListUid ? "bg-secondary font-medium" : "hover:bg-secondary",
              )}
            >
              <Link
                to="/lists/$listUid"
                params={{ listUid: list.uid }}
                aria-label={list.name}
                className={COVERING}
              />
              {/* Muted, no count and no dot: there is nothing left on it to be shared
                  about or to count, and the row is a record rather than a destination. */}
              <span className={cn(INERT, "truncate text-chrome text-muted-foreground")}>
                {list.name}
              </span>
            </div>
          ))}

          {total > lists.length && <SeeAll total={total} status="completed" />}
        </>
      )}
    </div>
  )
}

interface ListGroupProps {
  label: string
  group?: SidebarGroup
  activeListUid?: string
  /** What this Instance calls itself, for the share dialog behind each row's menu. */
  instanceName: string
}

/** One labelled group of Lists. An empty group is not rendered at all. */
function ListGroup({ label, group, activeListUid, instanceName }: ListGroupProps) {
  const lists = group?.lists ?? []
  if (lists.length === 0) {
    return null
  }
  const total = group?.total ?? lists.length

  return (
    <div className="flex flex-col gap-0.5">
      <div className="flex h-6.5 items-center px-2 text-badge font-semibold tracking-[0.03em] text-muted-foreground">
        {label}
      </div>
      {lists.map((list) => (
        <SidebarRow
          key={list.uid}
          list={list}
          active={list.uid === activeListUid}
          instanceName={instanceName}
        />
      ))}
      {total > lists.length && <SeeAll total={total} />}
    </div>
  )
}

interface SidebarRowProps {
  list: List
  active: boolean
  instanceName: string
}

/** One List in the sidebar: whether it is shared, what it is called, what is left on it. */
function SidebarRow({ list, active, instanceName }: SidebarRowProps) {
  return (
    <div
      className={cn(
        "group/list relative flex h-7.5 items-center gap-2.5 rounded-md px-2",
        active ? "bg-secondary font-medium" : "hover:bg-secondary",
      )}
    >
      <Link
        to="/lists/$listUid"
        params={{ listUid: list.uid }}
        aria-label={list.name}
        className={COVERING}
      />

      {/* Everything that only shows the List is inert, so the whole row stays one
          target and the link behind it is what answers a click. */}
      {list.sharing !== Sharing.PRIVATE && (
        <span className={cn(INERT, "size-[5px] shrink-0 rounded-full bg-shared")} />
      )}
      <span className={cn(INERT, "truncate text-chrome")}>{list.name}</span>

      {list.openCount > 0 && (
        <span className={cn(INERT, "ml-auto text-badge text-muted-foreground")}>
          {list.openCount}
        </span>
      )}

      {/* The `···` has a slot of its own, always the same size, so nothing moves
          when the pointer crosses the row. It is hidden by opacity rather than by
          display: a trigger that stops being rendered takes with it the element its
          menu is positioned against, and the menu jumps to the corner of the page
          the moment the pointer leaves the row to reach it. */}
      <span
        className={cn(
          RAISED,
          "-mr-1 shrink-0 opacity-0 transition-opacity",
          "group-hover/list:opacity-100 focus-within:opacity-100",
          "has-[[data-popup-open]]:opacity-100",
          list.openCount > 0 || "ml-auto",
        )}
      >
        <ListActions list={list} instanceName={instanceName} />
      </span>
    </div>
  )
}

/** initialsOf is the two-letter avatar fallback, e.g. "Anna" → "AN". */
function initialsOf(name: string): string {
  return name.slice(0, 2).toUpperCase()
}
