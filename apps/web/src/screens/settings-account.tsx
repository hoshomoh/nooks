import { useState, type ReactNode } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import { ConnectError, Code } from "@connectrpc/connect"

import { Button } from "@/components/ds/button"
import { SaveButton } from "@/components/ds/save-button"
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
      <SignOut />
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

        <SaveButton
          unchanged={unchanged}
          pending={save.isPending}
          succeeded={save.isSuccess}
          disabled={!name.trim() || !email.trim()}
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

        <SaveButton
          // Cleared on success, so an empty form is both "nothing to send" and "done".
          unchanged={!current && !next}
          pending={replace.isPending}
          succeeded={replace.isSuccess}
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

/**
 * Leaving.
 *
 * On this page because this page is the account, and because the design gives the
 * sidebar's member row one label and it says Settings. Two clicks away is further than
 * signing out should be; if it wants to be one, the sidebar needs a menu the design
 * does not have yet.
 */
function SignOut() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const signOut = useMutation({
    mutationFn: () => authClient.signOut({}),
    onSuccess: async () => {
      // Everything in the cache belongs to the Member who is leaving.
      queryClient.clear()
      await navigate({ to: "/sign-in" })
    },
  })

  return (
    <Section title={t("account.signOut")} blurb={t("account.signOutBlurb")}>
      <div>
        <Button tone="secondary" disabled={signOut.isPending} onClick={() => signOut.mutate()}>
          {t("account.signOutAction")}
        </Button>
      </div>
    </Section>
  )
}
