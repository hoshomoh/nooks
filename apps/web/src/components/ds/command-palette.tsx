import { useDeferredValue, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { useNavigate } from "@tanstack/react-router"
import { SearchHitKind, type SearchHit } from "@nooks/api"

import { cn } from "cn"

import { Button } from "./button"
import { DIALOG_SURFACE } from "./dialog-surface"
import { Field } from "./field"
import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandShortcut,
} from "@/components/ui/command"
import { listClient } from "@/lib/api"
import { listsQuery, searchQuery } from "@/lib/list-queries"
import { refreshLists } from "@/lib/refresh"
import { useCommandPalette } from "@/lib/use-command-palette"

/**
 * ⌘K: go anywhere, and find anything.
 *
 * One field for both, because a Member looking for "Groceries" does not know or care
 * whether it is a place in the app or a line inside one. Destinations are matched here;
 * content is matched by the server, which already excludes anything the Member could
 * not open — search is not a second permission system.
 */
export function CommandPalette() {
  const palette = useCommandPalette()

  return (
    <CommandDialog
      open={palette.mode !== "closed"}
      onOpenChange={(open) => {
        if (!open) {
          palette.close()
        }
      }}
      // The generated command dialog sets its own width and radius, the radius with
      // !important, so both have to be answered in kind.
      className={cn(DIALOG_SURFACE, "rounded-2xl!")}
    >
      {palette.mode === "add-list" ? <AddListPanel /> : <SearchPanel />}
    </CommandDialog>
  )
}

/** Where a Member can go that is not a List. */
interface Destination {
  to: "/today" | "/upcoming" | "/calendar" | "/"
  labelKey: string
}

const DESTINATIONS: Destination[] = [
  { to: "/today", labelKey: "views.today" },
  { to: "/upcoming", labelKey: "views.upcoming" },
  { to: "/calendar", labelKey: "views.calendar" },
  { to: "/", labelKey: "list.allLists" },
]

function SearchPanel() {
  const [query, setQuery] = useState("")
  const navigate = useNavigate()
  const palette = useCommandPalette()
  const { t } = useTranslation()

  // The field stays exactly as fast as the typing; the search follows a beat behind.
  // React's own deferral rather than a timer, so there is nothing to clean up.
  const searching = useDeferredValue(query)
  const results = useQuery(searchQuery(searching))
  const lists = useQuery(listsQuery)

  const hits = results.data?.hits ?? []
  const destinations = DESTINATIONS.filter((place) => matches(t(place.labelKey), query))
  const reachable = (lists.data?.lists ?? []).filter((list) => matches(list.name, query))

  const go = async (to: Destination["to"]) => {
    palette.close()
    await navigate({ to })
  }

  const openList = async (listUid: string) => {
    palette.close()
    await navigate({ to: "/lists/$listUid", params: { listUid } })
  }

  return (
    // The server decides which content matches; destinations and Lists are matched
    // here. One of the two has to be turned off, and it is easier to read when both
    // are matched the same way.
    <Command shouldFilter={false}>
      <CommandInput
        autoFocus
        value={query}
        onValueChange={setQuery}
        placeholder={t("palette.placeholder")}
      />
      <CommandList>
        <CommandEmpty>{t("palette.empty")}</CommandEmpty>

        {destinations.length > 0 && (
          <CommandGroup heading={t("palette.goTo")}>
            {destinations.map((place) => (
              <CommandItem key={place.to} onSelect={() => void go(place.to)}>
                {t(place.labelKey)}
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {reachable.length > 0 && (
          <CommandGroup heading={t("palette.lists")}>
            {reachable.map((list) => (
              <CommandItem key={list.uid} onSelect={() => void openList(list.uid)}>
                {list.name}
                {list.openCount > 0 && <CommandShortcut>{list.openCount}</CommandShortcut>}
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {hits.length > 0 && (
          <CommandGroup heading={t("palette.results")}>
            {hits.map((hit) => (
              <CommandItem
                key={`${hit.kind}-${hit.itemUid || hit.listUid}`}
                onSelect={() => void openList(hit.listUid)}
              >
                <span className="truncate">{hit.text}</span>
                {namesItsList(hit) && <CommandShortcut>{hit.listName}</CommandShortcut>}
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        <CommandGroup heading={t("palette.actions")}>
          <CommandItem onSelect={palette.openAddList}>{t("palette.addListTitle")}</CommandItem>
        </CommandGroup>
      </CommandList>
    </Command>
  )
}

/** matches is how a destination or a List name is compared with what was typed. */
function matches(name: string, query: string): boolean {
  return name.toLocaleLowerCase().includes(query.trim().toLocaleLowerCase())
}

/** namesItsList reports whether a hit needs to say which List it came from. */
function namesItsList(hit: SearchHit): boolean {
  return hit.kind !== SearchHitKind.LIST
}

function AddListPanel() {
  const [name, setName] = useState("")
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const palette = useCommandPalette()
  const { t } = useTranslation()

  const addList = useMutation({
    mutationFn: () => listClient.createList({ name }),
    onSuccess: async (res) => {
      palette.close()
      await refreshLists(queryClient)
      if (res.list) {
        await navigate({ to: "/lists/$listUid", params: { listUid: res.list.uid } })
      }
    },
  })

  return (
    <form
      className="flex flex-col gap-5 p-6"
      onSubmit={(event) => {
        event.preventDefault()
        if (name.trim()) {
          addList.mutate()
        }
      }}
    >
      <div className="flex flex-col gap-2">
        <h2 className="text-dialog">{t("palette.addListTitle")}</h2>
        <p className="text-field text-secondary-foreground">{t("palette.addListBlurb")}</p>
      </div>

      <Field
        label={t("palette.name")}
        value={name}
        onChange={(event) => setName(event.target.value)}
        autoFocus
      />

      <div className="flex items-center gap-3">
        <span className="flex-1" />
        <Button tone="secondary" type="button" onClick={palette.close}>
          {t("action.cancel")}
        </Button>
        <Button type="submit" disabled={addList.isPending}>
          {addList.isPending ? t("palette.adding") : t("palette.addListSubmit")}
        </Button>
      </div>
    </form>
  )
}
