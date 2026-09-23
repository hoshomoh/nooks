import { useDeferredValue, useState, type ReactNode } from "react"
import { useQuery } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { cn } from "cn"

import { Field } from "./field"
import { Icon } from "./icon"
import { TickBox } from "./tick-box"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { searchQuery } from "@/lib/list-queries"
import { listsAmong, type PickableList } from "@/lib/pick-lists"

export interface ListPickerProps {
  /** The Lists to offer before anything is typed, usually the sidebar's own. */
  suggested: PickableList[]
  /** Which are picked now, by identifier. */
  chosen: string[]
  onToggle: (list: PickableList) => void
  /** A row above the Lists, for an answer that is not one of them. */
  lead?: ReactNode
}

/**
 * The list of Lists a dialog picks from: a few to hand, and the rest by typing.
 *
 * An Instance can hold far more Lists than anybody wants scrolled past, so the rows
 * start as the handful somebody is working in and the field asks the server for the
 * rest. Nothing is sent until something is typed, so the common case is no round trip.
 */
export function ListPicker({ suggested, chosen, onToggle, lead }: ListPickerProps) {
  const { t } = useTranslation()
  const [query, setQuery] = useState("")

  // The field stays exactly as fast as the typing; the search follows a beat behind.
  const searching = useDeferredValue(query)
  const found = useQuery(searchQuery(searching))

  const typed = searching.trim().length > 0
  const shown = typed ? listsAmong(found.data?.hits ?? []) : suggested

  return (
    <div className="flex flex-col gap-2">
      <Field
        label={t("pickList.find")}
        value={query}
        onChange={(event) => setQuery(event.target.value)}
        placeholder={t("pickList.placeholder")}
      />

      {lead}

      {shown.length === 0 ? (
        <p className="py-2 text-field text-muted-foreground">
          {typed ? t("pickList.noneFound") : t("pickList.noneYet")}
        </p>
      ) : (
        <div className="flex max-h-(--size-floating-list) flex-col gap-0.5 overflow-y-auto">
          {shown.map((list) => (
            <button
              key={list.uid}
              type="button"
              role="checkbox"
              aria-checked={chosen.includes(list.uid)}
              onClick={() => onToggle(list)}
              className={cn(
                "flex min-h-row items-center gap-3 rounded-md px-2 py-1.5 text-left transition-colors",
                chosen.includes(list.uid) ? "bg-secondary" : "hover:bg-secondary",
              )}
            >
              <span className="truncate text-field">{list.name}</span>
              <TickBox picked={chosen.includes(list.uid)} className="ml-auto" />
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

export interface PickOneListProps {
  /** What a screen reader calls it — usually the settings row's own label. */
  label: string
  suggested: PickableList[]
  /** The List picked now, or undefined when none is. */
  picked?: PickableList
  /** What the control reads when none is: "All lists", "No public page". */
  noneLabel: string
  onPick: (list: PickableList | undefined) => void
}

/**
 * One List, or none, chosen from a control the width of a settings select.
 *
 * A popover rather than a select because the answers are not a fixed handful: a select
 * has to hold every option at once, and an Instance's Lists are not a list of languages.
 */
export function PickOneList({ label, suggested, picked, noneLabel, onPick }: PickOneListProps) {
  const [open, setOpen] = useState(false)

  const choose = (list: PickableList | undefined) => {
    onPick(list)
    setOpen(false)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        // Both, because the text in the control is the answer and not the question: a
        // button that only reads "Bike" says nothing about what Bike was picked for.
        aria-label={`${label}: ${picked?.name ?? noneLabel}`}
        className={cn(
          "flex h-control-settings w-70 items-center gap-2 rounded-lg border border-border",
          "px-3 text-field transition-colors hover:bg-secondary",
        )}
      >
        <span className={cn("truncate", picked ? "text-foreground" : "text-secondary-foreground")}>
          {picked?.name ?? noneLabel}
        </span>
        <Icon name="collapse" size="small" className="ml-auto shrink-0 text-control" />
      </PopoverTrigger>

      <PopoverContent align="end" className="w-80">
        <ListPicker
          suggested={suggested}
          chosen={picked ? [picked.uid] : []}
          onToggle={choose}
          lead={
            <button
              type="button"
              onClick={() => choose(undefined)}
              className={cn(
                "flex min-h-row items-center rounded-md px-2 py-1.5 text-left text-field transition-colors",
                picked ? "hover:bg-secondary" : "bg-secondary",
              )}
            >
              {noneLabel}
            </button>
          }
        />
      </PopoverContent>
    </Popover>
  )
}
