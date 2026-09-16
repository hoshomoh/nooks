import { Link, getRouteApi, useNavigate } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import { cn } from "cn"
import { Sharing, type List } from "@nooks/api"

import { AppShell } from "@/components/ds/app-shell"
import { Button } from "@/components/ds/button"
import { ActivityControl } from "@/components/ds/activity-control"
import { ChromeBar } from "@/components/ds/chrome-bar"
import { COVERING, INERT } from "@/components/ds/covering"
import { EmptyState } from "@/components/ds/empty-state"
import { ListActions } from "@/components/ds/list-actions"
import { Icon } from "@/components/ds/icon"
import { Menu, MenuItem } from "@/components/ds/menu"
import { SegmentedControl, type Segment } from "@/components/ds/segmented"
import {
  DEFAULT_SORT,
  LIST_SORTS,
  LIST_STATUSES,
  filterLists,
  pageOf,
  sortLists,
  type ListStatus,
} from "@/lib/list-table"
import { isArchived } from "@/lib/list-groups"
import type { Translate } from "@/lib/translate"
import { useCommandPalette } from "@/lib/use-command-palette"
import { useLocale } from "@/lib/use-locale"
import { useMomentLabel } from "@/lib/use-moment-label"
import { useSignedInData } from "@/lib/use-signed-in-data"
import type { ListsSearch } from "@/routes/index"

const route = getRouteApi("/")

/**
 * All lists — where a Member lands, and the cold start for a fresh account.
 *
 * The sidebar is what somebody is working in; this is everywhere else. It is the only
 * place a finished List can be found once it has left the groups above, which is why it
 * carries the filter rather than being a plain index.
 *
 * An empty instance says what a List is for rather than apologising, per DESIGN.md §11.
 */
export function Home() {
  const { instanceName, member, lists } = useSignedInData()
  const palette = useCommandPalette()
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { status = "all", sort = DEFAULT_SORT, page = 1 } = route.useSearch()
  const { code: locale } = useLocale()

  const shown = pageOf(sortLists(filterLists(lists, status), sort, locale), page)

  /*
   * Changing what is shown goes back to the first page.
   *
   * Page 3 of a filter is not page 3 of the next one, and staying put means narrowing
   * the set and landing somewhere that looks empty.
   */
  const show = (next: Partial<ListsSearch>) =>
    void navigate({ to: ".", search: (old) => ({ ...old, ...next, page: undefined }) })

  return (
    <AppShell
      instanceName={instanceName}
      memberName={member?.name ?? ""}
      lists={lists}
      onSearch={palette.open}
      onAddList={palette.openAddList}
    >
      <ChromeBar crumbs={[t("list.allLists")]} actions={<ActivityControl />} />

      <div className="flex min-h-0 flex-1 justify-center overflow-y-auto px-5.5 pt-14">
        <div className="w-full max-w-content self-start pb-22">
          <header className="mb-8.5 flex flex-col gap-3.5">
            <h1 className="text-display">{t("list.allLists")}</h1>
            <p className="text-meta text-secondary-foreground">
              {t("list.signedInAs", { name: member?.name ?? "" })}
            </p>
          </header>

          {lists.length === 0 ? (
            <div className="flex flex-col gap-6">
              <EmptyState title={t("list.coldStartTitle")} body={t("list.coldStartBody")} />
              <Button onClick={palette.openAddList} className="self-start">
                {t("sidebar.addList")}
              </Button>
            </div>
          ) : (
            <>
              <div className="mb-5 flex flex-wrap items-center gap-3">
                <SegmentedControl
                  label={t("list.statusLabel")}
                  options={LIST_STATUSES.map<Segment<ListStatus>>((one) => ({
                    value: one,
                    label: t(`list.status${capitalise(one)}`),
                  }))}
                  chosen={status}
                  onChoose={(next) => show({ status: next })}
                />

                {/* A menu rather than a select, per the design: it says what the
                    order is now and opens the three it could be, and it is not the
                    280px a settings row gives its select. */}
                <Menu
                  trigger={
                    <button
                      type="button"
                      aria-label={t("list.sortLabel")}
                      className="ml-auto flex h-control-settings shrink-0 items-center gap-1.5 rounded-lg border border-border px-3 text-meta text-secondary-foreground transition-colors hover:bg-secondary"
                    >
                      <span>{t(`list.sort${capitalise(sort)}`)}</span>
                      <Icon name="collapse" size="small" className="size-3 text-control" />
                    </button>
                  }
                >
                  {LIST_SORTS.map((one) => (
                    <MenuItem
                      key={one}
                      shortcut={one === sort ? "✓" : undefined}
                      onSelect={() => show({ sort: one })}
                    >
                      {t(`list.sort${capitalise(one)}`)}
                    </MenuItem>
                  ))}
                </Menu>
              </div>

              {shown.total === 0 ? (
                <EmptyState title={t("list.noneMatching")} body={t("list.noneMatchingBody")} />
              ) : (
                <>
                  <div className="flex flex-col">
                    {shown.rows.map((list) => (
                      <ListTableRow key={list.uid} list={list} instanceName={instanceName} />
                    ))}
                  </div>

                  {/* Only once there is more than a page. A household with nine Lists
                      should not be told it is looking at 1–9 of 9. */}
                  {shown.pages > 1 && (
                    <div className="mt-4 flex items-center gap-1.5">
                      <span className="text-micro text-muted-foreground">
                        {t("list.pageRange", {
                          from: shown.from,
                          to: shown.to,
                          total: shown.total,
                        })}
                      </span>
                      <span className="ml-auto flex gap-1.5">
                        <PageButton
                          label={t("list.previousPage")}
                          disabled={shown.page === 1}
                          onSelect={() =>
                            void navigate({
                              to: ".",
                              search: (old) => ({ ...old, page: shown.page - 1 }),
                            })
                          }
                        />
                        <PageButton
                          label={t("list.nextPage")}
                          disabled={shown.page === shown.pages}
                          onSelect={() =>
                            void navigate({
                              to: ".",
                              search: (old) => ({ ...old, page: shown.page + 1 }),
                            })
                          }
                        />
                      </span>
                    </div>
                  )}
                </>
              )}
            </>
          )}
        </div>
      </div>
    </AppShell>
  )
}

interface PageButtonProps {
  label: string
  disabled: boolean
  onSelect: () => void
}

/** One step through the table, per the design: 28px, radius 6, a hairline border. */
function PageButton({ label, disabled, onSelect }: PageButtonProps) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onSelect}
      className={cn(
        "h-7 rounded-md border border-border px-2.5 text-micro transition-colors",
        disabled ? "text-muted-foreground opacity-50" : "text-secondary-foreground hover:bg-secondary",
      )}
    >
      {label}
    </button>
  )
}

interface ListTableRowProps {
  list: List
  instanceName: string
}

/**
 * One List: what it is called, how much is left on it, and when it last changed.
 *
 * The whole row opens it, with the `···` raised above that so the menu still answers a
 * click — the same arrangement every other row in the app uses.
 */
function ListTableRow({ list, instanceName }: ListTableRowProps) {
  const { t } = useTranslation()
  const moment = useMomentLabel()

  return (
    <div className="group/list relative grid min-h-13 grid-cols-[1fr_96px_26px] items-center gap-4 border-b border-hair px-2 -mx-2 hover:bg-secondary">
      <Link
        to="/lists/$listUid"
        params={{ listUid: list.uid }}
        aria-label={list.name}
        className={COVERING}
      />

      <span className={cn(INERT, "flex min-w-0 flex-col gap-0.5")}>
        <span className="flex items-center gap-2.5">
          {list.sharing !== Sharing.PRIVATE && (
            <span className="size-[5px] shrink-0 rounded-full bg-shared" />
          )}
          <span className={cn("truncate text-body", isArchived(list) && "text-secondary-foreground")}>
            {list.name}
          </span>
        </span>
        <span className="text-micro text-muted-foreground">{metaFor(list, t)}</span>
      </span>

      <span className={cn(INERT, "text-micro whitespace-nowrap text-muted-foreground")}>
        {moment(list.updatedAt)}
      </span>

      <span className="relative z-10 opacity-0 transition-opacity group-hover/list:opacity-100 focus-within:opacity-100 has-[[data-popup-open]]:opacity-100">
        <ListActions list={list} instanceName={instanceName} />
      </span>
    </div>
  )
}

/** metaFor is the line under the name: who put it away, or how much is left on it. */
function metaFor(list: List, t: Translate): string {
  if (isArchived(list)) {
    return t("list.archivedBy", { name: list.archivedByName })
  }
  if (list.openCount > 0) {
    return t("list.openCount", { count: list.openCount })
  }
  return t("list.nothingOpen")
}

/** capitalise makes a value into the tail of a translation key. */
function capitalise(word: string): string {
  return word.charAt(0).toUpperCase() + word.slice(1)
}
