import { getRouteApi, Link } from "@tanstack/react-router"

import { AppShell } from "@/components/ds/app-shell"
import { Button } from "@/components/ds/button"
import { ChromeBar } from "@/components/ds/chrome-bar"
import { EmptyState } from "@/components/ds/empty-state"
import { useCommandPalette } from "@/lib/use-command-palette"

const route = getRouteApi("/")

/**
 * All lists — where a Member lands, and the cold start for a fresh account.
 *
 * An empty instance says what a List is for rather than apologising, per DESIGN.md §11.
 */
export function Home() {
  const { instance, member, lists } = route.useLoaderData()
  const palette = useCommandPalette()

  return (
    <AppShell
      instanceName={instance.name}
      memberName={member.name}
      lists={lists.lists}
      onSearch={palette.open}
      onAddList={palette.openAddList}
    >
      <ChromeBar crumbs={["All lists"]} />

      <div className="flex justify-center px-5.5 pt-14 pb-22">
        <div className="w-full max-w-content">
          <header className="mb-8.5 flex flex-col gap-3.5">
            <h1 className="text-display">All lists</h1>
            <p className="text-secondary text-secondary-foreground">
              Signed in as {member.name}
            </p>
          </header>

          {lists.lists.length === 0 ? (
            <div className="flex flex-col gap-6">
              <EmptyState
                title="Start with one list"
                body="Groceries, flat jobs, a book you keep meaning to find. A list is a name and a first item; sharing it with the household comes later."
              />
              <Button onClick={palette.openAddList} className="self-start">
                Add a list
              </Button>
            </div>
          ) : (
            <div className="flex flex-col">
              {lists.lists.map((list) => (
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
                    {list.openCount === 0 ? "nothing open" : `${list.openCount} open`}
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
