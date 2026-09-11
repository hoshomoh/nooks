import { useSuspenseQuery } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import type { ReactNode } from "react"

import { SettingsRow } from "@/components/ds/settings-row"
import { SettingsShell } from "@/components/ds/settings-shell"
import { aboutQuery } from "@/lib/about-queries"
import { readableBytes } from "@/lib/bytes"
import { useMomentLabel } from "@/lib/use-moment-label"
import { useSettingsCounts } from "@/lib/use-settings-counts"

/**
 * About: what this copy of Nooks is, and how much it is holding.
 *
 * Facts in the same rows every other settings page uses. There is no dashboard here
 * and nothing is charted — somebody self-hosting wants to know the thing is real and
 * roughly how big it is before they decide whether to back it up.
 */
export function SettingsAboutScreen() {
  const { t } = useTranslation()
  const counts = useSettingsCounts()
  const moment = useMomentLabel()
  const about = useSuspenseQuery(aboutQuery).data

  const size = readableBytes(Number(about.storageBytes))

  return (
    <SettingsShell active="/settings/about" crumb={t("settings.about")} counts={counts}>
      <header className="flex flex-col gap-1.5">
        <h1 className="text-page">{t("settings.about")}</h1>
        <p className="max-w-135 text-chrome leading-[1.55] text-secondary-foreground">
          {t("about.blurb")}
        </p>
      </header>

      <div className="mt-3.5 flex flex-col">
        <SettingsRow label={t("about.version")}>
          <Value>{t("about.versionValue", { version: about.version })}</Value>
        </SettingsRow>

        {size && (
          <SettingsRow label={t("about.storage")} blurb={t("about.storageBlurb")}>
            <Value mono>{`${about.storageDriver} · ${size}`}</Value>
          </SettingsRow>
        )}

        <SettingsRow
          label={t("about.instance")}
          blurb={about.startedAt ? t("about.since", { when: moment(about.startedAt) }) : undefined}
        >
          <Value>{about.instanceName}</Value>
        </SettingsRow>

        <SettingsRow label={t("about.inIt")}>
          <Value>
            {[
              t("about.listCount", { count: about.listCount }),
              t("about.itemCount", { count: about.itemCount }),
              t("about.memberCount", { count: about.memberCount }),
            ].join(" · ")}
          </Value>
        </SettingsRow>

        <SettingsRow label={t("about.licence")} blurb={t("about.licenceBlurb")}>
          <Value>{about.licence}</Value>
        </SettingsRow>

        {/* Stated as a fact rather than as a setting with a toggle off. There is nothing
            to turn off, and a toggle would imply there was. */}
        <SettingsRow label={t("about.telemetry")} blurb={t("about.telemetryBlurb")}>
          <Value>{t("about.none")}</Value>
        </SettingsRow>
      </div>
    </SettingsShell>
  )
}

interface ValueProps {
  /** Mono for anything measured, so two sizes line up under each other. */
  mono?: boolean
  children: ReactNode
}

/** The right-hand side of an About row: a fact, never a control. */
function Value({ mono, children }: ValueProps) {
  return (
    <span className={mono ? "font-mono text-meta text-secondary-foreground" : "text-field"}>
      {children}
    </span>
  )
}
