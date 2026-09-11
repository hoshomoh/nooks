import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { useNavigate } from "@tanstack/react-router"

import { AuthShell } from "@/components/ds/auth-shell"
import { Button } from "@/components/ds/button"
import { Field } from "@/components/ds/field"
import { FormError } from "@/components/ds/form-error"
import { authClient } from "@/lib/api"
import { createStarterList } from "@/lib/starter-list"
import { messageFrom } from "@/lib/errors"
import { startNewSession } from "@/lib/new-session"

export function Setup() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { t } = useTranslation()

  const [name, setName] = useState("")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [instanceName, setInstanceName] = useState("")

  const completeSetup = useMutation({
    mutationFn: () => authClient.completeSetup({ name, email, password, instanceName }),
    onSuccess: async () => {
      // A first List, so the Instance opens onto something that shows what an Item can
      // carry. It is a courtesy: if it fails, the Admin gets an empty Instance, which
      // is the thing they asked for and can fill themselves.
      await createStarterList(t).catch(() => undefined)

      // The Instance is now named, there is a session, and there is a List.
      startNewSession(queryClient)
      await navigate({ to: "/" })
    },
  })

  return (
    <AuthShell
      title={t("auth.setup.title")}
      blurb={t("auth.setup.blurb")}
      footerRight={t("auth.setup.footer")}
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          completeSetup.mutate()
        }}
      >
        <Field
          label={t("auth.setup.yourName")}
          value={name}
          onChange={(event) => setName(event.target.value)}
          autoComplete="name"
          required
        />
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
          autoComplete="new-password"
          hint={t("auth.setup.passwordHint")}
          required
        />
        <Field
          label={t("auth.setup.instanceName")}
          value={instanceName}
          onChange={(event) => setInstanceName(event.target.value)}
          hint={t("auth.setup.instanceNameHint")}
          required
        />

        <FormError message={messageFrom(completeSetup.error)} />

        <Button type="submit" disabled={completeSetup.isPending} className="self-start">
          {completeSetup.isPending ? t("auth.setup.submitting") : t("auth.setup.submit")}
        </Button>
      </form>
    </AuthShell>
  )
}
