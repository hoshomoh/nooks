import { Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { AppShell } from "@/components/ds/app-shell"
import { Button } from "@/components/ds/button"
import { ChromeBar } from "@/components/ds/chrome-bar"
import { EmptyState } from "@/components/ds/empty-state"
import { useCommandPalette } from "@/lib/use-command-palette"
import { useSignedInData } from "@/lib/use-signed-in-data"

/**
 * All lists — where a Member lands, and the cold start for a fresh account.
 *
 * An empty instance says what a List is for rather than apologising, per DESIGN.md §11.
 */
export function Home() {
  const { instanceName, member, lists } = useSignedInData()
  const palette = useCommandPalette()
  const { t } = useTranslation()

  return (
    <AppShell
      instanceName={instanceName}
      memberName={member?.name ?? ""}
      lists={lists}
      onSearch={palette.open}
      onAddList={palette.openAddList}
    >
      <ChromeBar crumbs={[t("list.allLists")]} />

      <div className="flex justify-center px-5.5 pt-14 pb-22">
        <div className="w-full max-w-content">
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
            <div className="flex flex-col">
              {lists.map((list) => (
                <Link
                  key={list.uid}
                  to="/lists/$listUid"
                  params={{ listUid: list.uid }}
                  className="grid min-h-row grid-cols-[1fr_auto] items-center gap-3.5 rounded-md border-b border-hair px-2 -mx-2 hover:bg-secondary"
                >
                  <span className="flex items-center gap-2.5">
                    {list.sharing !== 1 && (
                      <span className="size-[5px] shrink-0 rounded-full bg-shared" />
                    )}
                    <span className="truncate text-body">{list.name}</span>
                  </span>
                  <span className="text-micro text-muted-foreground">
                    {t("list.openCount", { count: list.openCount })}
                  </span>
                </Link>
              ))}
            </div>
          )}
        </div>
      </div>
    </AppShell>
  )
}
