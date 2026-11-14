import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import type { List } from "@nooks/api"

import { ConfirmDialog } from "./confirm-dialog"
import { IconButton } from "./icon-button"
import { Menu, MenuItem, MenuSeparator } from "./menu"
import { listClient } from "@/lib/api"
import { refreshLists } from "@/lib/refresh"

export interface DeletedListActionsProps {
  list: List
}

/**
 * What can be done to a List its owner deleted, behind the same `···` as every other row.
 *
 * A deleted List cannot be opened, renamed or shared, so its menu holds only bringing it
 * back and letting it go now rather than at the end of its window. Everything is re-read
 * afterwards, so it leaves this filter either way.
 */
export function DeletedListActions({ list }: DeletedListActionsProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [purging, setPurging] = useState(false)

  const restore = useMutation({
    mutationFn: () => listClient.restoreList({ listUid: list.uid }),
    onSuccess: () => refreshLists(queryClient),
  })

  const purge = useMutation({
    mutationFn: () => listClient.purgeList({ listUid: list.uid }),
    onSuccess: () => refreshLists(queryClient),
  })

  return (
    <>
      <Menu trigger={<IconButton name="more" scale="compact" label={t("sidebarMenu.open")} />}>
        <MenuItem onSelect={() => restore.mutate()}>{t("list.restore")}</MenuItem>
        <MenuSeparator />
        <MenuItem destructive shortcut="⌫" onSelect={() => setPurging(true)}>
          {t("list.purge")}
        </MenuItem>
      </Menu>

      <ConfirmDialog
        open={purging}
        onOpenChange={setPurging}
        title={t("list.purgeTitle", { name: list.name })}
        blurb={t("list.purgeBlurb", { count: list.openCount + list.doneCount })}
        confirmLabel={t("list.purge")}
        destructive
        onConfirm={() => purge.mutate()}
      />
    </>
  )
}
