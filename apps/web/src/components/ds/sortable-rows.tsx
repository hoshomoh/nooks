import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { Accessibility } from "@dnd-kit/dom"
import {
  DragDropProvider,
  type DragEndEvent,
  type DragOverEvent,
  type DragStartEvent,
} from "@dnd-kit/react"
import { isSortable, useSortable } from "@dnd-kit/react/sortable"
import type { Item } from "@nooks/api"

import { placeAfterDrop, type MovePlace } from "@/lib/item-order"
import type { Translate } from "@/lib/translate"
import { useReducedMotion } from "@/lib/use-reduced-motion"

import { ListRow, type ListRowProps } from "./list-row"

export interface SortableRowsProps {
  /** The open Items, in the order they are drawn. */
  items: readonly Item[]
  /** Where a drop put an Item, in the terms the Instance moves one in. */
  onMove: (place: MovePlace) => void
  children: ReactNode
}

/**
 * The open rows of a List, each of which can be picked up by its grip and dropped.
 *
 * Only around the open rows: completed Items do not move, and the done section is not a
 * place to drop one. The rows the carried one passes slide apart to open its place, and
 * Esc puts it back, both handled by the library rather than written here.
 */
export function SortableRows({ items, onMove, children }: SortableRowsProps) {
  const { t } = useTranslation()
  const labelOf: LabelOf = (id) => items.find((item) => item.uid === id)?.label ?? ""

  const dropped = (event: DragEndEvent) => {
    const source = event.operation.source
    if (event.canceled || !isSortable(source)) {
      return
    }
    const place = placeAfterDrop(items, source.sortable.initialIndex, source.sortable.index)
    if (place) {
      onMove(place)
    }
  }

  return (
    <DragDropProvider
      plugins={(defaults) => [
        ...defaults,
        Accessibility.configure({
          screenReaderInstructions: { draggable: t("list.dragInstructions") },
          announcements: announcements(t, labelOf, items.length),
        }),
      ]}
      onDragEnd={dropped}
    >
      {children}
    </DragDropProvider>
  )
}

/** What an Item is called, given the id the library tracks it by. */
type LabelOf = (id: unknown) => string

/** What a screen reader is told at each moment of a carry, or nothing. */
interface RowAnnouncements {
  dragstart: (event: DragStartEvent) => string | undefined
  dragover: (event: DragOverEvent) => string | undefined
  dragend: (event: DragEndEvent) => string | undefined
}

/**
 * announcements is what a screen reader hears while a row is carried.
 *
 * The library's own name the Item by its uid, which nobody reads. These name it by its
 * label and say its place in words, the way the row says it on screen.
 */
function announcements(t: Translate, labelOf: LabelOf, count: number): RowAnnouncements {
  return {
    dragstart: ({ operation: { source } }) =>
      source ? t("list.dragPickedUp", { label: labelOf(source.id) }) : undefined,
    dragover: ({ operation: { source } }) =>
      isSortable(source)
        ? t("list.dragOver", { label: labelOf(source.id), place: source.sortable.index + 1, count })
        : undefined,
    dragend: ({ operation: { source }, canceled }) => {
      if (!isSortable(source)) {
        return undefined
      }
      const label = labelOf(source.id)
      return canceled
        ? t("list.dragCanceled", { label })
        : t("list.dragDropped", { label, place: source.sortable.index + 1, count })
    },
  }
}

export interface SortableListRowProps extends ListRowProps {
  /** The Item's uid, which is what the library tracks it by. */
  id: string
  /** Where it sits among the open rows. */
  index: number
  /** Whether the Member may pick it up, which is whether they may change the List. */
  movable: boolean
}

/**
 * An open Item's row, which slides when its place changes.
 *
 * Sortable for everyone, so a move made by somebody else on a shared List slides the
 * rows into their new places for whoever is watching, rather than jumping. Only a
 * Member who may change the List gets the grip to make a move of their own.
 */
export function SortableListRow({ id, index, movable, ...row }: SortableListRowProps) {
  const { t } = useTranslation()
  const reduced = useReducedMotion()
  const { ref, handleRef, isDragging } = useSortable({
    id,
    index,
    disabled: !movable,
    // The rows that make way slide; for somebody who asked for less motion they step.
    transition: reduced ? null : undefined,
  })

  return (
    <ListRow
      {...row}
      drag={{
        ref,
        lifted: isDragging,
        grip: movable ? { handleRef, label: t("list.dragGrip") } : undefined,
      }}
    />
  )
}
