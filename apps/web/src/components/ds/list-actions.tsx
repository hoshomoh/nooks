import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import type { List } from "@nooks/api"

import { Button } from "./button"
import { ConfirmDialog } from "./confirm-dialog"
import { IconButton } from "./icon-button"
import { Menu, MenuItem, MenuSeparator } from "./menu"
import { PromptDialog } from "./prompt-dialog"
import { ShareDialog, type ShareDecision } from "./share-dialog"
import { listClient } from "@/lib/api"
import { afterDeletingList } from "@/lib/after-deleting-list"
import { download } from "@/lib/download"
import { exportFileFor } from "@/lib/list-export"
import { listQuery } from "@/lib/list-queries"
import { isArchived } from "@/lib/list-state"
import { refreshLists } from "@/lib/refresh"

export interface ListActionsProps {
  list: List
  /** What this Instance calls itself, for the share dialog. */
  instanceName: string
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
export function ListActions({ list, instanceName, withControls }: ListActionsProps) {
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

  const archive = useMutation({
    mutationFn: (archived: boolean) => listClient.setListArchived({ listUid, archived }),
    onSuccess: refresh,
  })

  // Read from the cache when the List is open, and fetched when the menu is somewhere
  // else, so every list menu can export rather than only the one on the List itself.
  const exportList = useMutation({
    mutationFn: () => queryClient.ensureQueryData(listQuery(listUid)),
    onSuccess: ({ items }) => download(exportFileFor(list, items)),
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
    onSuccess: () => afterDeletingList(queryClient, listUid, () => navigate({ to: "/" })),
  })

  return (
    <>
      {withControls && list.isOwner && (
        <Button tone="quiet" scale="toolbar" onClick={() => setSharing(true)}>
          {t("share.action")}
        </Button>
      )}
      {withControls && (
        <Button tone="quiet" scale="toolbar" onClick={() => window.print()}>
          {t("sidebarMenu.print")}
        </Button>
      )}

      <Menu
        trigger={<IconButton name="more" scale="compact" label={t("sidebarMenu.open")} />}
      >
        {/* No keycap either: Enter is what the menu itself does to whichever entry is
            highlighted, so printing it against one of them says the wrong thing. */}
        <MenuItem onSelect={() => void navigate({ to: "/lists/$listUid", params: { listUid } })}>
          {t("sidebarMenu.openList")}
        </MenuItem>
        {/* Rename carries no keycap: F2 is the brightness key on a Mac unless somebody
            has changed it, and the combinations a browser leaves free are ones it may
            want back. Renaming is rare enough to be worth a click. */}
        {list.isOwner && (
          <MenuItem onSelect={() => setRenaming(true)}>{t("sidebarMenu.rename")}</MenuItem>
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

        <MenuItem onSelect={() => exportList.mutate()}>{t("listMenu.export")}</MenuItem>

        {/* Grouped with Duplicate and above Delete, per the design. Archiving is not
            deleting — it keeps everything and comes back — so it is not destructive and
            does not sit with the one thing that is. */}
        {list.isOwner && (
          <MenuItem shortcut="⌘⇧A" onSelect={() => archive.mutate(!isArchived(list))}>
            {t(isArchived(list) ? "sidebarMenu.restore" : "sidebarMenu.archive")}
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
