import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { getRouteApi, Link, useNavigate } from "@tanstack/react-router"

import { AuthShell } from "@/components/ds/auth-shell"
import { Button } from "@/components/ds/button"
import { Field } from "@/components/ds/field"
import { FormError } from "@/components/ds/form-error"
import { authClient } from "@/lib/api"
import { applyPendingTick } from "@/lib/apply-pending-tick"
import { messageFrom } from "@/lib/errors"

const route = getRouteApi("/sign-in")

export function SignIn() {
  const { instance } = route.useLoaderData()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { t } = useTranslation()

  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")

  const signIn = useMutation({
    mutationFn: () => authClient.signIn({ email, password }),
    onSuccess: async () => {
      // Before anything is read, so the List they land on already shows it ticked.
      await applyPendingTick()
      await queryClient.invalidateQueries()
      await navigate({ to: "/" })
    },
  })

  return (
    <AuthShell
      eyebrow={t("auth.signIn.eyebrow")}
      title={instance.name}
      blurb={t("auth.signIn.blurb")}
      footerLeft={
        <Link to="/forgot-password" className="text-shared">
          {t("auth.signIn.forgot")}
        </Link>
      }
      footerRight={
        <Link to="/join" className="text-shared">
          {t("auth.signIn.join")}
        </Link>
      }
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          signIn.mutate()
        }}
      >
        <Field
          label={t("auth.setup.email")}
          type="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          autoComplete="email"
          required
        />
        <Field
          label={t("auth.setup.password")}
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          autoComplete="current-password"
          required
        />

        <FormError message={messageFrom(signIn.error)} />

        <Button type="submit" disabled={signIn.isPending} className="self-start">
          {signIn.isPending ? t("auth.signIn.submitting") : t("action.signIn")}
        </Button>
      </form>
    </AuthShell>
  )
}
