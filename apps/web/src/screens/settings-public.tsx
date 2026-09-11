import { useState } from "react"
import { useMutation, useQueryClient, useSuspenseQuery } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import type { InstanceSettings, PublicListSettings } from "@nooks/api"

import { SaveButton } from "@/components/ds/save-button"
import { SettingsRow, Toggle } from "@/components/ds/settings-row"
import { SettingsShell } from "@/components/ds/settings-shell"
import { instanceClient } from "@/lib/api"
import { instanceSettingsQuery } from "@/lib/instance-queries"
import { listsQuery } from "@/lib/list-queries"

/**
 * The public list settings: which List, and how much of it a Visitor sees.
 *
 * Everything here is off by default. A page anybody can open is a decision somebody
 * makes deliberately, and each thing it reveals is a second one.
 */
export function SettingsPublicScreen() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const saved = useSuspenseQuery(instanceSettingsQuery).data.settings
  const lists = useSuspenseQuery(listsQuery).data.lists

  // Held until Save, so a half-made decision never reaches a Visitor.
  const [draft, setDraft] = useState<PublicListSettings | undefined>(saved?.publicList)

  const save = useMutation({
    mutationFn: (settings: InstanceSettings) =>
      instanceClient.updateInstanceSettings({ settings }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["instance-settings"] }),
  })

  const set = (fields: Partial<PublicListSettings>) =>
    setDraft({ ...(draft as PublicListSettings), ...fields })

  const published = Boolean(draft?.listUid)

  return (
    <SettingsShell active="/settings/public" crumb={t("settings.publicList")}>
      <header className="flex flex-col gap-1.5">
        <h1 className="text-page">{t("publicSettings.title")}</h1>
        <p className="max-w-135 text-chrome leading-[1.55] text-secondary-foreground">
          {t("publicSettings.blurb")}
        </p>
      </header>

      <div className="mt-3.5 flex flex-col">
        <SettingsRow label={t("publicSettings.whichList")} blurb={t("publicSettings.whichListBlurb")}>
          <select
            value={draft?.listUid ?? ""}
            onChange={(event) => set({ listUid: event.target.value })}
            aria-label={t("publicSettings.whichList")}
            className="h-input w-70 rounded-lg border border-border bg-transparent px-3 text-field"
          >
            <option value="">{t("publicSettings.none")}</option>
            {lists.map((list) => (
              <option key={list.uid} value={list.uid}>
                {list.name}
              </option>
            ))}
          </select>
        </SettingsRow>

        {/* The rest only mean anything once something is published. */}
        {published && (
          <>
            <SettingsRow
              label={t("publicSettings.showNames")}
              blurb={t("publicSettings.showNamesBlurb")}
            >
              <Toggle
                on={draft?.showNames ?? false}
                onChange={(showNames) => set({ showNames })}
                label={t("publicSettings.showNames")}
              />
            </SettingsRow>

            <SettingsRow
              label={t("publicSettings.showMeta")}
              blurb={t("publicSettings.showMetaBlurb")}
            >
              <Toggle
                on={draft?.showMeta ?? false}
                onChange={(showMeta) => set({ showMeta })}
                label={t("publicSettings.showMeta")}
              />
            </SettingsRow>

            <SettingsRow
              label={t("publicSettings.allowJoin")}
              blurb={t("publicSettings.allowJoinBlurb")}
            >
              <Toggle
                on={draft?.allowJoin ?? false}
                onChange={(allowJoin) => set({ allowJoin })}
                label={t("publicSettings.allowJoin")}
              />
            </SettingsRow>
          </>
        )}
      </div>

      <div className="mt-4 flex items-center gap-3">
        {published && (
          <span className="text-micro text-muted-foreground">
            {t("publicSettings.address", { url: window.location.origin })}
          </span>
        )}
        <span className="flex-1" />
        <SaveButton
          unchanged={samePublicList(draft, saved?.publicList)}
          pending={save.isPending}
          succeeded={save.isSuccess}
          onClick={() =>
            saved && draft && save.mutate({ ...saved, publicList: draft } as InstanceSettings)
          }
        />
      </div>
    </SettingsShell>
  )
}

/**
 * samePublicList reports whether the draft still says what was saved.
 *
 * Field by field rather than by identity: the answer comes back from the server as a
 * fresh object every time, so two settings that agree are never the same value.
 */
function samePublicList(
  draft: PublicListSettings | undefined,
  saved: PublicListSettings | undefined,
): boolean {
  return (
    (draft?.listUid ?? "") === (saved?.listUid ?? "") &&
    (draft?.showNames ?? false) === (saved?.showNames ?? false) &&
    (draft?.showMeta ?? false) === (saved?.showMeta ?? false) &&
    (draft?.allowJoin ?? false) === (saved?.allowJoin ?? false)
  )
}
