import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import type { List } from "@nooks/api"

import { IconButton } from "./icon-button"
import { Menu, MenuItem } from "./menu"
import { listClient } from "@/lib/api"
import { refreshLists } from "@/lib/refresh"

export interface DeletedListActionsProps {
  list: List
}

/**
 * What can be done to a List its owner deleted, behind the same `···` as every other row.
 *
 * A deleted List cannot be opened, renamed or shared, so its menu holds only what brings
 * it back. Everything is re-read afterwards: it reappears under whatever filter it
 * belongs to now.
 */
export function DeletedListActions({ list }: DeletedListActionsProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const restore = useMutation({
    mutationFn: () => listClient.restoreList({ listUid: list.uid }),
    onSuccess: () => refreshLists(queryClient),
  })

  return (
    <Menu trigger={<IconButton name="more" scale="compact" label={t("sidebarMenu.open")} />}>
      <MenuItem onSelect={() => restore.mutate()}>{t("list.restore")}</MenuItem>
    </Menu>
  )
}
