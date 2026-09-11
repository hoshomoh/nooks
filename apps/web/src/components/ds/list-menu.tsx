import { useState } from "react"
import { useTranslation } from "react-i18next"
import type { Item, List } from "@nooks/api"

import { Button } from "./button"
import { ConfirmDialog } from "./confirm-dialog"
import { Menu, MenuItem, MenuSeparator } from "./menu"
import { download } from "@/lib/download"
import { exportFileFor } from "@/lib/list-export"

export interface ListMenuActions {
  onRename: () => void
  onPin: (pinned: boolean) => void
  onDuplicate: () => void
  onDelete: () => void
}

export interface ListMenuProps {
  list: List
  /** The Items, so the List can be exported without asking for them again. */
  items: Item[]
  actions: ListMenuActions
}

/**
 * The `···` menu, per DESIGN.md §9.
 *
 * Everything that is not worth a permanent control: renaming, pinning, duplicating,
 * printing, exporting and deleting. Only an owner may rename or delete, so those are
 * absent rather than disabled for anyone else — a control that can never be used is
 * noise, not information.
 */
export function ListMenu({ list, items, actions }: ListMenuProps) {
  const { t } = useTranslation()
  const [confirming, setConfirming] = useState(false)

  return (
    <>
      <Menu
        trigger={
          <Button tone="quiet" scale="toolbar" aria-label={t("listMenu.open")}>
            ···
          </Button>
        }
      >
        {list.isOwner && (
          <MenuItem onSelect={actions.onRename}>{t("listMenu.rename")}</MenuItem>
        )}
        <MenuItem onSelect={() => actions.onPin(!list.isPinned)}>
          {t(list.isPinned ? "listMenu.unpin" : "listMenu.pin")}
        </MenuItem>
        <MenuItem onSelect={actions.onDuplicate}>{t("listMenu.duplicate")}</MenuItem>

        <MenuSeparator />

        <MenuItem shortcut="⌘P" onSelect={() => window.print()}>
          {t("listMenu.print")}
        </MenuItem>
        <MenuItem onSelect={() => download(exportFileFor(list, items))}>
          {t("listMenu.export")}
        </MenuItem>

        {list.isOwner && (
          <>
            <MenuSeparator />
            <MenuItem destructive onSelect={() => setConfirming(true)}>
              {t("listMenu.delete")}
            </MenuItem>
          </>
        )}
      </Menu>

      <ConfirmDialog
        open={confirming}
        onOpenChange={setConfirming}
        title={t("listMenu.deleteTitle", { name: list.name })}
        blurb={t("listMenu.deleteBlurb")}
        confirmLabel={t("listMenu.deleteConfirm")}
        destructive
        onConfirm={actions.onDelete}
      />
    </>
  )
}
