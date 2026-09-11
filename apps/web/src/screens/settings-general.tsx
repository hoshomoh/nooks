import { useState, type ReactNode } from "react"
import { useMutation, useQueryClient, useSuspenseQuery } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import { ConnectError, Code } from "@connectrpc/connect"
import { Role, type InstanceSettings } from "@nooks/api"

import { Button } from "@/components/ds/button"
import {
  ChangePasswordDialog,
  type PasswordChange,
} from "@/components/ds/change-password-dialog"
import { Field } from "@/components/ds/field"
import { SaveButton } from "@/components/ds/save-button"
import { SegmentedControl } from "@/components/ds/segmented"
import { SelectField } from "@/components/ds/select-field"
import { SettingsRow, Toggle } from "@/components/ds/settings-row"
import { SettingsShell } from "@/components/ds/settings-shell"
import { LOCALES, DEFAULT_LOCALE } from "@/i18n/locales"
import { authClient, instanceClient, memberClient } from "@/lib/api"
import { instanceSettingsQuery } from "@/lib/instance-queries"
import type { Theme } from "@/lib/theme"
import type { Translate } from "@/lib/translate"
import { useLocale } from "@/lib/use-locale"
import { useSettingsCounts } from "@/lib/use-settings-counts"
import { useSignedInData } from "@/lib/use-signed-in-data"
import { useTheme } from "@/lib/use-theme"

/** The themes, in the order DESIGN.md §2 lists them. */
const THEMES: Theme[] = ["light", "dark", "system"]

/**
 * General: everything about you, and — for an Admin — about the Instance.
 *
 * One page with sections rather than three pages, because these are all things a
 * Member changes in passing. Splitting them would turn one glance down a page into
 * three navigations.
 */
export function SettingsGeneralScreen() {
  const { t } = useTranslation()
  const counts = useSettingsCounts()
  const { instanceName, member } = useSignedInData()
  const isAdmin = member?.role === Role.ADMIN

  return (
    <SettingsShell active="/settings/general" crumb={t("settings.general")} counts={counts}>
      <header className="flex flex-col gap-1.5">
        <h1 className="text-page">{t("settings.general")}</h1>
        <p className="text-chrome text-secondary-foreground">
          {t(isAdmin ? "general.signedInAdmin" : "general.signedIn", {
            name: member?.name ?? "",
            instance: instanceName,
          })}
        </p>
      </header>

      <AccountSection />
      <AppearanceSection />
      {isAdmin && <InstanceSection />}

      {!isAdmin && (
        <p className="mt-2 text-micro text-muted-foreground">{t("general.footerNote")}</p>
      )}
    </SettingsShell>
  )
}

/** Who you are here, and the password you get in with. */
function AccountSection() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const { member } = useSignedInData()

  const [name, setName] = useState(member?.name ?? "")
  const [email, setEmail] = useState(member?.email ?? "")
  const [changing, setChanging] = useState(false)

  const save = useMutation({
    mutationFn: () => memberClient.updateOwnProfile({ name: name.trim(), email: email.trim() }),
    onSuccess: () => queryClient.invalidateQueries(),
  })

  const replace = useMutation({
    mutationFn: (change: PasswordChange) => authClient.replacePassword(change),
    onSuccess: () => setChanging(false),
  })

  const signOut = useMutation({
    mutationFn: () => authClient.signOut({}),
    onSuccess: async () => {
      // Everything in the cache belongs to the Member who is leaving.
      queryClient.clear()
      await navigate({ to: "/sign-in" })
    },
  })

  const unchanged = name.trim() === member?.name && email.trim() === member?.email

  return (
    <Section label={t("general.account")}>
      <SettingsRow label={t("account.name")} blurb={t("general.nameBlurb")}>
        <Field
          label={t("account.name")}
          hideLabel
          value={name}
          onChange={(event) => setName(event.target.value)}
          className="w-70"
        />
      </SettingsRow>

      <SettingsRow label={t("account.email")} blurb={t("account.emailHint")}>
        <Field
          label={t("account.email")}
          hideLabel
          type="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          error={emailErrorOf(t, save.error)}
          className="w-70"
        />
      </SettingsRow>

      <SettingsRow label={t("account.password")} blurb={t("general.passwordBlurb")}>
        <Button tone="secondary" scale="compact" onClick={() => setChanging(true)}>
          {t("account.changePassword")}
        </Button>
      </SettingsRow>

      <SettingsRow label={t("account.signOut")} blurb={t("account.signOutBlurb")}>
        <Button
          tone="secondary"
          scale="compact"
          disabled={signOut.isPending}
          onClick={() => signOut.mutate()}
        >
          {t("account.signOutAction")}
        </Button>
      </SettingsRow>

      <div className="mt-4 flex">
        <span className="flex-1" />
        <SaveButton
          unchanged={unchanged}
          pending={save.isPending}
          succeeded={save.isSuccess}
          disabled={!name.trim() || !email.trim()}
          onClick={() => save.mutate()}
        />
      </div>

      <ChangePasswordDialog
        open={changing}
        onOpenChange={(open) => {
          setChanging(open)
          if (!open) {
            replace.reset()
          }
        }}
        onChange={(change) => replace.mutate(change)}
        error={passwordErrorOf(t, replace.error)}
        pending={replace.isPending}
      />
    </Section>
  )
}

/** How Nooks looks, and what language it speaks. Applies as chosen. */
function AppearanceSection() {
  const { t } = useTranslation()
  const { theme, setTheme } = useTheme()
  const { code, setCode } = useLocale()

  return (
    <Section label={t("general.appearance")}>
      <SettingsRow label={t("appearance.theme")} blurb={t("appearance.themeBlurb")}>
        <SegmentedControl
          label={t("appearance.theme")}
          options={THEMES.map((value) => ({ value, label: t(`appearance.${value}`) }))}
          chosen={theme}
          onChoose={setTheme}
        />
      </SettingsRow>

      <SettingsRow label={t("appearance.language")} blurb={t("general.languageNote")}>
        <SelectField
          label={t("appearance.language")}
          options={LOCALES.map((locale) => ({ value: locale.code, label: locale.nativeName }))}
          value={code}
          onValueChange={setCode}
        />
      </SettingsRow>
    </Section>
  )
}

/** The Instance itself. Admins only, and the badge says so. */
function InstanceSection() {
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
    <Section label={t("general.instance")} badge={t("general.adminsOnly")}>
      <SettingsRow label={t("instance.name")} blurb={t("instance.nameBlurb")}>
        <Field
          label={t("instance.name")}
          hideLabel
          value={draft?.name ?? ""}
          onChange={(event) => set({ name: event.target.value })}
          className="w-70"
        />
      </SettingsRow>

      <SettingsRow label={t("instance.language")} blurb={t("instance.languageBlurb")}>
        <SelectField
          label={t("instance.language")}
          options={LOCALES.map((locale) => ({ value: locale.code, label: locale.nativeName }))}
          value={draft?.defaultLocale || DEFAULT_LOCALE}
          onValueChange={(defaultLocale) => set({ defaultLocale })}
        />
      </SettingsRow>

      <SettingsRow label={t("instance.publicSignup")} blurb={t("instance.publicSignupBlurb")}>
        <span className="text-micro text-muted-foreground">
          {t(draft?.publicSignup ? "general.on" : "general.off")}
        </span>
        <Toggle
          on={draft?.publicSignup ?? false}
          onChange={(publicSignup) => set({ publicSignup })}
          label={t("instance.publicSignup")}
        />
      </SettingsRow>

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
    </Section>
  )
}

interface SectionProps {
  label: string
  /** Who the section is for, when it is not for everybody. */
  badge?: string
  children: ReactNode
}

/** One block of the page, split from the next by a rule. */
function Section({ label, badge, children }: SectionProps) {
  return (
    <section className="mt-3.5 flex flex-col">
      <div className="flex items-baseline gap-2.5 border-b border-border pb-2.5">
        <span className="text-label text-muted-foreground uppercase">{label}</span>
        {badge && <span className="text-micro text-muted-foreground">{badge}</span>}
      </div>
      {children}
    </section>
  )
}

/**
 * sameInstance reports whether the draft still says what was saved.
 *
 * Only the fields this section owns: the public list settings live on their own page,
 * and carrying them in would make this Save light up when that page is used.
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

/** emailErrorOf reads a failed save as the one thing that usually goes wrong. */
function emailErrorOf(t: Translate, error: Error | null): string | undefined {
  if (!error) {
    return undefined
  }
  return ConnectError.from(error).code === Code.AlreadyExists
    ? t("account.emailTaken")
    : t("error.somethingWentWrong")
}

/** passwordErrorOf reads a failed replacement the same way. */
function passwordErrorOf(t: Translate, error: Error | null): string | undefined {
  if (!error) {
    return undefined
  }
  const failure = ConnectError.from(error)
  return failure.code === Code.InvalidArgument
    ? failure.rawMessage
    : t("error.somethingWentWrong")
}
