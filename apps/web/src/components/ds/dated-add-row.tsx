import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import type { List } from "@nooks/api"

import { AddRow, type AddRowSubmission } from "./add-row"
import { listClient } from "@/lib/api"
import type { DueDate } from "@/lib/dates"
import { refreshLists } from "@/lib/refresh"
import { useDueLabel } from "@/lib/use-due-label"
import { useLastList } from "@/lib/use-last-list"

export interface DatedAddRowProps {
  /** Every List the Member can reach, to pick the one an Item lands on. */
  lists: List[]
  /** The date this view gives a new Item: today on Today, tomorrow on Upcoming. */
  defaultDue: DueDate
  /** Whether a rule separates the row from what is above it. See AddRow. */
  divided?: boolean
}

/**
 * The add row on a view that gathers Items from every List, per the add row's spec.
 *
 * An Item added here has to land somewhere, and the somewhere a Member means is the
 * List they last added to. The placeholder says so in words — "Add to Groceries, due
 * today" — because a default that is not stated is a default nobody can trust.
 *
 * Nothing renders when the Member has no Lists at all: there would be nowhere to put
 * what they typed, and a row that swallows an Item is worse than no row.
 */
export function DatedAddRow({ lists, defaultDue, divided }: DatedAddRowProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const due = useDueLabel()
  const { target, remember } = useLastList(lists)

  const addItem = useMutation({
    mutationFn: ({ label, quantity, dueOn }: AddRowSubmission) =>
      listClient.createItem({ listUid: target?.uid ?? "", label, quantity, dueOn }),
    onSuccess: () => refreshLists(queryClient),
  })

  if (!target) {
    return null
  }

  return (
    <AddRow
      placeholder={t("list.addToListDue", {
        name: target.name,
        date: due.label(defaultDue).toLocaleLowerCase(),
      })}
      defaultDue={defaultDue}
      divided={divided}
      disabled={addItem.isPending}
      onAdd={(item) => {
        remember(target.uid)
        addItem.mutate(item)
      }}
    />
  )
}
