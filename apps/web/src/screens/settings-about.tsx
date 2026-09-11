import { useState } from "react"
import { useMutation, useQueryClient, useSuspenseQuery } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { Role } from "@nooks/api"
import { useTranslation } from "react-i18next"
import type { ReactNode } from "react"

import { Button } from "@/components/ds/button"
import {
  IrreversibleDialog,
  type IrreversibleConfirmation,
} from "@/components/ds/irreversible-dialog"
import { SettingsRow } from "@/components/ds/settings-row"
import { SettingsShell } from "@/components/ds/settings-shell"
import { aboutQuery } from "@/lib/about-queries"
import { instanceClient } from "@/lib/api"
import { readableBytes } from "@/lib/bytes"
import { messageFrom } from "@/lib/errors"
import { useSignedInData } from "@/lib/use-signed-in-data"
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
  const { member } = useSignedInData()

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

      {member?.role === Role.ADMIN && <DeleteInstance name={about.instanceName} />}
    </SettingsShell>
  )
}

interface DeleteInstanceProps {
  /** What the Instance is called, which is what has to be typed back. */
  name: string
}

/**
 * The last thing on the page, because it is the last thing anybody wants.
 *
 * It takes every Member, List, Item, Note, Access token and setting with it and returns
 * the Instance to first run. Any Admin may do it: on a household Instance the people
 * with the keys are the people who share the shopping, and a rule that only the founder
 * could wipe it would strand a household whose founder has left.
 */
function DeleteInstance({ name }: DeleteInstanceProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [asking, setAsking] = useState(false)

  const remove = useMutation({
    mutationFn: (confirmation: IrreversibleConfirmation) =>
      instanceClient.deleteInstance({
        password: confirmation.password,
        instanceName: confirmation.typedName,
      }),
    onSuccess: async () => {
      queryClient.clear()
      await navigate({ to: "/" })
    },
  })

  return (
    <section className="mt-8 flex flex-col gap-3.5 border-t border-hair pt-5">
      <div className="flex items-center gap-6">
        <div className="flex flex-col gap-0.5">
          <span className="text-chrome">{t("danger.deleteTitle")}</span>
          <span className="max-w-135 text-micro leading-[1.5] text-muted-foreground">
            {t("danger.deleteBlurb")}
          </span>
        </div>
        <span className="flex-1" />
        <Button tone="destructive" scale="compact" onClick={() => setAsking(true)}>
          {t("danger.deleteAction")}
        </Button>
      </div>

      <IrreversibleDialog
        open={asking}
        onOpenChange={(open) => {
          setAsking(open)
          if (!open) {
            remove.reset()
          }
        }}
        title={t("danger.confirmTitle", { name })}
        blurb={
          <>
            <span>{t("danger.confirmGoes")}</span>
            <span>{t("danger.confirmKept")}</span>
          </>
        }
        name={name}
        nameLabel={t("danger.typeName", { name })}
        confirmLabel={t("danger.deleteAction")}
        onConfirm={(confirmation) => remove.mutate(confirmation)}
        error={remove.error ? messageFrom(remove.error) : undefined}
        pending={remove.isPending}
      />
    </section>
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
