import { useState } from "react"
import { useMutation, useQueryClient, useSuspenseQuery } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import type { InstanceSettings } from "@nooks/api"

import { Field } from "@/components/ds/field"
import { SaveButton } from "@/components/ds/save-button"
import { SelectField } from "@/components/ds/select-field"
import { SettingsRow, Toggle } from "@/components/ds/settings-row"
import { SettingsShell } from "@/components/ds/settings-shell"
import { LOCALES, DEFAULT_LOCALE } from "@/i18n/locales"
import { instanceClient } from "@/lib/api"
import { instanceSettingsQuery } from "@/lib/instance-queries"

/**
 * The Instance itself: what it is called, who may join, and the language it falls back
 * to. Admins only — the server refuses anybody else, and the nav entry is hidden.
 *
 * Held until Save, unlike Appearance. These decide what people outside this browser
 * see, so a half-typed name should not reach the sidebar of everybody here.
 */
export function SettingsInstanceScreen() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const saved = useSuspenseQuery(instanceSettingsQuery).data.settings

  const [draft, setDraft] = useState<InstanceSettings | undefined>(saved)

  const save = useMutation({
    mutationFn: (settings: InstanceSettings) =>
      instanceClient.updateInstanceSettings({ settings }),
    onSuccess: () => queryClient.invalidateQueries(),
  })

  const set = (fields: Partial<InstanceSettings>) =>
    setDraft({ ...(draft as InstanceSettings), ...fields })

  return (
    <SettingsShell active="/settings/instance" crumb={t("settings.instance")}>
      <header className="flex flex-col gap-1.5">
        <h1 className="text-page">{t("instance.title")}</h1>
        <p className="max-w-135 text-chrome leading-[1.55] text-secondary-foreground">
          {t("instance.blurb")}
        </p>
      </header>

      <div className="mt-3.5 flex flex-col">
        <SettingsRow label={t("instance.name")} blurb={t("instance.nameBlurb")}>
          <Field
            label={t("instance.name")}
            hideLabel
            value={draft?.name ?? ""}
            onChange={(event) => set({ name: event.target.value })}
            className="w-70"
          />
        </SettingsRow>

        <SettingsRow label={t("instance.publicSignup")} blurb={t("instance.publicSignupBlurb")}>
          <Toggle
            on={draft?.publicSignup ?? false}
            onChange={(publicSignup) => set({ publicSignup })}
            label={t("instance.publicSignup")}
          />
        </SettingsRow>

        <SettingsRow label={t("instance.language")} blurb={t("instance.languageBlurb")}>
          <SelectField
            label={t("instance.language")}
            options={LOCALES.map((locale) => ({
              value: locale.code,
              label: locale.nativeName,
            }))}
            value={draft?.defaultLocale || DEFAULT_LOCALE}
            onValueChange={(defaultLocale) => set({ defaultLocale })}
          />
        </SettingsRow>
      </div>

      <div className="mt-4 flex">
        <span className="flex-1" />
        <SaveButton
          unchanged={sameInstance(draft, saved)}
          pending={save.isPending}
          succeeded={save.isSuccess}
          disabled={!draft?.name.trim()}
          onClick={() => draft && save.mutate(draft)}
        />
      </div>
    </SettingsShell>
  )
}

/**
 * sameInstance reports whether the draft still says what was saved.
 *
 * Only the fields this page owns: the public list settings live on their own page, and
 * carrying them into the comparison would make this Save light up when that one is used.
 */
function sameInstance(
  draft: InstanceSettings | undefined,
  saved: InstanceSettings | undefined,
): boolean {
  return (
    (draft?.name ?? "") === (saved?.name ?? "") &&
    (draft?.publicSignup ?? false) === (saved?.publicSignup ?? false) &&
    (draft?.defaultLocale ?? "") === (saved?.defaultLocale ?? "")
  )
}
