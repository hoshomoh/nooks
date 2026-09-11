import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import type { Item, List } from "@nooks/api"

import { Button } from "./button"
import { ConfirmDialog } from "./confirm-dialog"
import { DotsButton } from "./dots-button"
import { Menu, MenuItem, MenuSeparator } from "./menu"
import { PromptDialog } from "./prompt-dialog"
import { ShareDialog, type ShareDecision } from "./share-dialog"
import { listClient } from "@/lib/api"
import { download } from "@/lib/download"
import { exportFileFor } from "@/lib/list-export"
import { refreshLists } from "@/lib/refresh"

export interface ListActionsProps {
  list: List
  /** What this Instance calls itself, for the share dialog. */
  instanceName: string
  /**
   * The Items, when the caller already has them.
   *
   * Exporting needs them; the sidebar does not have them and does not offer it. A
   * menu entry that would need a second round trip to answer is not worth the entry.
   */
  items?: Item[]
  /**
   * Whether to render Share and Print as permanent controls beside the menu.
   *
   * The chrome bar does, per DESIGN.md §8; a sidebar row has no room and keeps them in
   * the menu. Either way the dialogs are owned here, so there is one of each.
   */
  withControls?: boolean
}

/**
 * Everything that can be done to a List, behind one `···`.
 *
 * One component for both places the design puts it — the sidebar row and the List's own
 * chrome bar — because the same List should offer the same things wherever it is
 * reached from. It owns its dialogs, so a menu entry never has to hand its work back to
 * a screen that happens to be mounted.
 */
export function ListActions({ list, instanceName, items, withControls }: ListActionsProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [renaming, setRenaming] = useState(false)
  const [sharing, setSharing] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const refresh = () => refreshLists(queryClient)
  const listUid = list.uid

  const rename = useMutation({
    mutationFn: (name: string) => listClient.renameList({ listUid, name }),
    onSuccess: refresh,
  })

  const pin = useMutation({
    mutationFn: (pinned: boolean) => listClient.setListPinned({ listUid, pinned }),
    onSuccess: refresh,
  })

  const share = useMutation({
    mutationFn: (decision: ShareDecision) => listClient.setListSharing({ listUid, ...decision }),
    onSuccess: refresh,
  })

  const duplicate = useMutation({
    mutationFn: () => listClient.duplicateList({ listUid }),
    onSuccess: async (res) => {
      await refresh()
      if (res.list) {
        void navigate({ to: "/lists/$listUid", params: { listUid: res.list.uid } })
      }
    },
  })

  const remove = useMutation({
    mutationFn: () => listClient.deleteList({ listUid }),
    onSuccess: async () => {
      await refresh()
      void navigate({ to: "/" })
    },
  })

  return (
    <>
      {withControls && list.isOwner && (
        <Button tone="secondary" scale="toolbar" onClick={() => setSharing(true)}>
          {t("share.action")}
        </Button>
      )}
      {withControls && (
        <Button tone="quiet" scale="toolbar" onClick={() => window.print()}>
          {t("sidebarMenu.print")}
        </Button>
      )}

      <Menu
        trigger={<DotsButton aria-label={t("sidebarMenu.open")} />}
      >
        <MenuItem
          shortcut="↵"
          onSelect={() => void navigate({ to: "/lists/$listUid", params: { listUid } })}
        >
          {t("sidebarMenu.openList")}
        </MenuItem>
        {list.isOwner && (
          <MenuItem shortcut="F2" onSelect={() => setRenaming(true)}>
            {t("sidebarMenu.rename")}
          </MenuItem>
        )}
        <MenuItem shortcut="⌘D" onSelect={() => pin.mutate(!list.isPinned)}>
          {t(list.isPinned ? "sidebarMenu.unpin" : "sidebarMenu.pin")}
        </MenuItem>

        <MenuSeparator />

        {/* Absent when the chrome bar already offers them as permanent controls: the
            same action twice in one place reads as two different actions. */}
        {!withControls && list.isOwner && (
          <MenuItem onSelect={() => setSharing(true)}>{t("sidebarMenu.share")}</MenuItem>
        )}
        {!withControls && (
          <MenuItem shortcut="⌘P" onSelect={() => window.print()}>
            {t("sidebarMenu.print")}
          </MenuItem>
        )}
        <MenuItem onSelect={() => duplicate.mutate()}>{t("sidebarMenu.duplicate")}</MenuItem>

        {items && (
          <MenuItem onSelect={() => download(exportFileFor(list, items))}>
            {t("listMenu.export")}
          </MenuItem>
        )}

        {list.isOwner && (
          <>
            <MenuSeparator />
            <MenuItem destructive shortcut="⌫" onSelect={() => setDeleting(true)}>
              {t("sidebarMenu.delete")}
            </MenuItem>
          </>
        )}
      </Menu>

      <PromptDialog
        open={renaming}
        onOpenChange={setRenaming}
        title={t("sidebarMenu.renameTitle", { name: list.name })}
        blurb={t("sidebarMenu.renameBlurb")}
        label={t("palette.name")}
        initialValue={list.name}
        confirmLabel={t("action.save")}
        onConfirm={(name) => rename.mutate(name)}
      />

      <ShareDialog
        list={list}
        instanceName={instanceName}
        open={sharing}
        onOpenChange={setSharing}
        onSave={(decision) => share.mutate(decision)}
      />

      <ConfirmDialog
        open={deleting}
        onOpenChange={setDeleting}
        title={t("sidebarMenu.deleteTitle", { name: list.name })}
        blurb={t("sidebarMenu.deleteBlurb")}
        confirmLabel={t("sidebarMenu.deleteConfirm")}
        destructive
        onConfirm={() => remove.mutate()}
      />
    </>
  )
}
