import { useState, type ReactNode } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { ConnectError, Code } from "@connectrpc/connect"

import { Button } from "@/components/ds/button"
import { Field } from "@/components/ds/field"
import { SettingsShell } from "@/components/ds/settings-shell"
import { authClient, memberClient } from "@/lib/api"
import type { Translate } from "@/lib/translate"
import { useSignedInData } from "@/lib/use-signed-in-data"

/**
 * The account page: who a Member is here, and the password they get in with.
 *
 * Two forms rather than one, because they fail differently. A taken email and a wrong
 * current password are different problems, and a single Save that could mean either
 * would leave a Member guessing which half went wrong.
 */
export function SettingsAccountScreen() {
  const { t } = useTranslation()

  return (
    <SettingsShell active="/settings/account" crumb={t("settings.account")}>
      <header className="flex flex-col gap-1.5">
        <h1 className="text-page">{t("account.title")}</h1>
        <p className="max-w-135 text-chrome leading-[1.55] text-secondary-foreground">
          {t("account.blurb")}
        </p>
      </header>

      <ProfileForm />
      <PasswordForm />
    </SettingsShell>
  )
}

/** What a Member is called here, and the address they sign in with. */
function ProfileForm() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { member } = useSignedInData()

  const [name, setName] = useState(member?.name ?? "")
  const [email, setEmail] = useState(member?.email ?? "")

  const save = useMutation({
    mutationFn: () => memberClient.updateOwnProfile({ name: name.trim(), email: email.trim() }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["current-member"] })
      await queryClient.invalidateQueries({ queryKey: ["members"] })
    },
  })

  const unchanged = name.trim() === member?.name && email.trim() === member?.email

  return (
    <Section title={t("account.profile")} blurb={t("account.profileBlurb")}>
      <form
        className="flex flex-col gap-4.5"
        onSubmit={(event) => {
          event.preventDefault()
          save.mutate()
        }}
      >
        <Field
          label={t("account.name")}
          value={name}
          onChange={(event) => setName(event.target.value)}
          className="max-w-70"
        />
        <Field
          label={t("account.email")}
          type="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          hint={t("account.emailHint")}
          error={profileErrorOf(t, save.error)}
          className="max-w-70"
        />

        <Footer
          saved={save.isSuccess && !save.isPending && unchanged}
          pending={save.isPending}
          disabled={unchanged || !name.trim() || !email.trim()}
        />
      </form>
    </Section>
  )
}

/** Replacing the password, which needs the current one. */
function PasswordForm() {
  const { t } = useTranslation()
  const [current, setCurrent] = useState("")
  const [next, setNext] = useState("")

  const replace = useMutation({
    mutationFn: () =>
      authClient.replacePassword({ currentPassword: current, newPassword: next }),
    onSuccess: () => {
      setCurrent("")
      setNext("")
    },
  })

  return (
    <Section title={t("account.password")} blurb={t("account.passwordBlurb")}>
      <form
        className="flex flex-col gap-4.5"
        onSubmit={(event) => {
          event.preventDefault()
          replace.mutate()
        }}
      >
        <Field
          label={t("account.currentPassword")}
          type="password"
          value={current}
          onChange={(event) => setCurrent(event.target.value)}
          error={passwordErrorOf(t, replace.error)}
          className="max-w-70"
        />
        <Field
          label={t("account.newPassword")}
          type="password"
          value={next}
          onChange={(event) => setNext(event.target.value)}
          hint={t("auth.passwordRule")}
          className="max-w-70"
        />

        <Footer
          saved={replace.isSuccess && !replace.isPending}
          pending={replace.isPending}
          disabled={!current || !next}
          label={t("account.replacePassword")}
        />
      </form>
    </Section>
  )
}

interface SectionProps {
  title: string
  blurb: string
  children: ReactNode
}

/** One block of the page, split from the next by a rule. */
function Section({ title, blurb, children }: SectionProps) {
  return (
    <section className="flex flex-col gap-4 border-t border-hair pt-6">
      <div className="flex flex-col gap-1">
        <h2 className="text-chrome font-semibold">{title}</h2>
        <p className="max-w-135 text-micro leading-[1.5] text-muted-foreground">{blurb}</p>
      </div>
      {children}
    </section>
  )
}

interface FooterProps {
  saved: boolean
  pending: boolean
  disabled: boolean
  /** What the button says. Defaults to Save. */
  label?: string
}

/** The submit row, left-aligned: these are forms, not dialogs. */
function Footer({ saved, pending, disabled, label }: FooterProps) {
  const { t } = useTranslation()

  return (
    <div className="flex items-center gap-3">
      <Button type="submit" disabled={disabled || pending}>
        {label ?? t("action.save")}
      </Button>
      {saved && <span className="text-micro text-muted-foreground">{t("account.saved")}</span>}
    </div>
  )
}

/** profileErrorOf reads a failed save as the one thing that usually goes wrong. */
function profileErrorOf(t: Translate, error: Error | null): string | undefined {
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
  return ConnectError.from(error).code === Code.InvalidArgument
    ? ConnectError.from(error).rawMessage
    : t("error.somethingWentWrong")
}
