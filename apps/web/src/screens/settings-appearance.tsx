import { useTranslation } from "react-i18next"

import { Segmented } from "@/components/ds/segmented"
import { SelectField } from "@/components/ds/select-field"
import { SettingsRow } from "@/components/ds/settings-row"
import { SettingsShell } from "@/components/ds/settings-shell"
import { LOCALES } from "@/i18n/locales"
import type { Theme } from "@/lib/theme"
import { useLocale } from "@/lib/use-locale"
import { useTheme } from "@/lib/use-theme"

/** The themes, in the order DESIGN.md §2 lists them. */
const THEMES: Theme[] = ["light", "dark", "system"]

/**
 * Appearance: how Nooks looks, and what language it speaks.
 *
 * Both take effect as they are chosen rather than on a Save. There is nothing to get
 * half-right here — a Member can see immediately whether they meant it, and a theme
 * behind a Save button would be a preview nobody asked for.
 */
export function SettingsAppearanceScreen() {
  const { t } = useTranslation()
  const { theme, setTheme } = useTheme()
  const { code, setCode } = useLocale()

  return (
    <SettingsShell active="/settings/appearance" crumb={t("settings.appearance")}>
      <header className="flex flex-col gap-1.5">
        <h1 className="text-page">{t("appearance.title")}</h1>
        <p className="max-w-135 text-chrome leading-[1.55] text-secondary-foreground">
          {t("appearance.blurb")}
        </p>
      </header>

      <div className="mt-3.5 flex flex-col">
        <SettingsRow label={t("appearance.theme")} blurb={t("appearance.themeBlurb")}>
          <Segmented
            label={t("appearance.theme")}
            options={THEMES.map((value) => ({ value, label: t(`appearance.${value}`) }))}
            chosen={theme}
            onChoose={setTheme}
          />
        </SettingsRow>

        <SettingsRow label={t("appearance.language")} blurb={t("appearance.languageBlurb")}>
          <SelectField
            label={t("appearance.language")}
            options={LOCALES.map((locale) => ({
              value: locale.code,
              label: locale.nativeName,
            }))}
            value={code}
            onValueChange={setCode}
          />
        </SettingsRow>
      </div>

      {/* Said once, at the foot: a Member looking for their own language should find out
          here why it is not listed, rather than concluding Nooks cannot do it. */}
      <p className="mt-4 max-w-135 text-micro leading-[1.6] text-muted-foreground">
        {t("appearance.translationsNote", { count: LOCALES.length })}
      </p>
    </SettingsShell>
  )
}
