import { useDeferredValue, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { useNavigate } from "@tanstack/react-router"
import { SearchHitKind } from "@nooks/api"

import { Button } from "./button"
import { Field } from "./field"
import { listClient } from "@/lib/api"
import { searchQuery } from "@/lib/list-queries"
import { refreshLists } from "@/lib/refresh"
import { useCommandPalette } from "@/lib/use-command-palette"

/**
 * ⌘K: jump to anything, and add a List.
 *
 * The results are whatever the server returns, which already excludes anything the
 * Member could not open — search is not a second permission system.
 */
export function CommandPalette() {
  const palette = useCommandPalette()

  if (palette.mode === "closed") {
    return null
  }

  return (
    <div
      className="fixed inset-0 z-50 flex animate-scrim-in items-start justify-center bg-[var(--scrim)] pt-[12vh] px-6"
      onClick={palette.close}
    >
      <div
        className="w-full max-w-[560px] animate-panel-in overflow-hidden rounded-2xl border border-border bg-popover shadow-[0_24px_60px_rgba(0,0,0,0.22)]"
        onClick={(event) => event.stopPropagation()}
      >
        {palette.mode === "search" ? <SearchPanel /> : <AddListPanel />}
      </div>
    </div>
  )
}

function SearchPanel() {
  const [query, setQuery] = useState("")
  const navigate = useNavigate()
  const palette = useCommandPalette()
  const { t } = useTranslation()

  // The field stays exactly as fast as the typing; the search follows a beat behind.
  // React's own deferral rather than a timer, so there is nothing to clean up and
  // nothing to synchronise.
  const searching = useDeferredValue(query)
  const results = useQuery(searchQuery(searching))

  const go = async (listUid: string) => {
    palette.close()
    await navigate({ to: "/lists/$listUid", params: { listUid } })
  }

  const hits = results.data?.hits ?? []

  return (
    <div className="flex flex-col">
      <input
        autoFocus
        value={query}
        onChange={(event) => setQuery(event.target.value)}
        placeholder={t("palette.searchPlaceholder")}
        className="h-input border-b border-hair px-4 text-field outline-none placeholder:text-muted-foreground"
      />

      <div className="max-h-[50vh] overflow-y-auto p-1.5">
        {query.trim() === "" && (
          <p className="px-3 py-3 text-meta text-muted-foreground">
            {t("palette.searchHint")}
          </p>
        )}
        {/* Only once the answer is in: saying "nothing" while still looking is worse
            than saying nothing at all. */}
        {searching.trim() !== "" && hits.length === 0 && results.isSuccess && (
          <p className="px-3 py-3 text-meta text-muted-foreground">
            {t("palette.noResults", { query: searching })}
          </p>
        )}
        {hits.map((hit) => (
          <button
            key={`${hit.kind}-${hit.itemUid || hit.listUid}`}
            type="button"
            onClick={() => go(hit.listUid)}
            className="flex w-full items-baseline gap-3 rounded-md px-3 py-2 text-left hover:bg-secondary"
          >
            <span className="truncate text-chrome">{hit.text}</span>
            {hit.kind === SearchHitKind.ITEM && (
              <span className="ml-auto shrink-0 text-micro text-muted-foreground">
                {hit.listName}
              </span>
            )}
          </button>
        ))}
      </div>
    </div>
  )
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
        <p className="text-chrome leading-[1.6] text-secondary-foreground">
          {t("palette.addListBlurb")}
        </p>
      </div>

      <Field
        label={t("palette.name")}
        value={name}
        onChange={(event) => setName(event.target.value)}
        autoFocus
        required
      />

      <div className="flex items-center gap-3">
        <Button type="submit" disabled={addList.isPending}>
          {addList.isPending ? t("palette.adding") : t("palette.addListSubmit")}
        </Button>
        <Button type="button" tone="secondary" onClick={palette.close}>
          {t("action.cancel")}
        </Button>
      </div>
    </form>
  )
}
