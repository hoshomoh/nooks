import { Link, useRouterState } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import { cn } from "cn"
import { Sharing, type List } from "@nooks/api"

import { ListActions } from "./list-actions"

import { Mark } from "@/components/mark"

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
  lists: List[]
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
 * The grouping is derived from the Lists themselves rather than stored, so a List moves
 * between groups the moment it is pinned or shared.
 */
export function Sidebar({
  instanceName,
  memberName,
  lists,
  activeListUid,
  counts,
  onSearch,
  onAddList,
}: SidebarProps) {
  const { t } = useTranslation()
  const path = useRouterState({ select: (state) => state.location.pathname })
  const active = viewForPath(path)
  const pinned = lists.filter((list) => list.isPinned)
  const mine = lists.filter((list) => !list.isPinned && list.isOwner)
  const shared = lists.filter((list) => !list.isPinned && !list.isOwner)

  return (
    <aside className="flex w-[258px] flex-col gap-4.5 border-r border-border bg-sidebar px-2 pt-3 pb-4">
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
          count={lists.length}
          active={active === "/"}
        />
      </div>

      <ListGroup
        label={t("sidebar.pinned")}
        lists={pinned}
        activeListUid={activeListUid}
        instanceName={instanceName}
      />
      <ListGroup
        label={t("sidebar.myLists")}
        lists={mine}
        activeListUid={activeListUid}
        instanceName={instanceName}
      />
      <ListGroup
        label={t("sidebar.sharedWithMe")}
        lists={shared}
        activeListUid={activeListUid}
        instanceName={instanceName}
      />

      <button
        type="button"
        onClick={onAddList}
        className="flex h-7.5 items-center gap-2.5 rounded-md px-2 text-meta text-muted-foreground hover:bg-secondary"
      >
        <span>+</span>
        <span>{t("sidebar.addList")}</span>
      </button>

      <Link
        to="/settings/members"
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
        <span className={accent ? "ml-auto text-[11.5px] text-shared" : "ml-auto text-[11.5px] text-muted-foreground"}>
          {count}
        </span>
      )}
    </Link>
  )
}

interface ListGroupProps {
  label: string
  lists: List[]
  activeListUid?: string
  /** What this Instance calls itself, for the share dialog behind each row's menu. */
  instanceName: string
}

/** One labelled group of Lists. An empty group is not rendered at all. */
function ListGroup({ label, lists, activeListUid, instanceName }: ListGroupProps) {
  if (lists.length === 0) {
    return null
  }

  return (
    <div className="flex flex-col gap-0.5">
      <div className="flex h-6.5 items-center px-2 text-[11.5px] font-semibold tracking-[0.03em] text-muted-foreground">
        {label}
      </div>
      {lists.map((list) => (
        <div
          key={list.uid}
          className={cn(
            "group/list relative flex h-7.5 items-center gap-2.5 rounded-md px-2",
            list.uid === activeListUid ? "bg-secondary font-medium" : "hover:bg-secondary",
          )}
        >
          <Link
            to="/lists/$listUid"
            params={{ listUid: list.uid }}
            aria-label={list.name}
            className="absolute inset-0 rounded-md"
          />

          {/* Everything that only shows the List is inert, so the whole row stays one
              target and the link behind it is what answers a click. */}
          {list.sharing !== Sharing.PRIVATE && (
            <span className="pointer-events-none relative size-[5px] shrink-0 rounded-full bg-shared" />
          )}
          <span className="pointer-events-none relative truncate text-chrome">{list.name}</span>

          {list.openCount > 0 && (
            <span className="pointer-events-none relative ml-auto text-[11.5px] text-muted-foreground">
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
              "relative z-10 -mr-1 shrink-0 opacity-0 transition-opacity",
              "group-hover/list:opacity-100 focus-within:opacity-100",
              "has-[[data-popup-open]]:opacity-100",
              list.openCount > 0 || "ml-auto",
            )}
          >
            <ListActions list={list} instanceName={instanceName} />
          </span>
        </div>
      ))}
    </div>
  )
}

/** initialsOf is the two-letter avatar fallback, e.g. "Anna" → "AN". */
function initialsOf(name: string): string {
  return name.slice(0, 2).toUpperCase()
}
