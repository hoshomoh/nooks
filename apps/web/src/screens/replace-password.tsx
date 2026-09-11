import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { useNavigate } from "@tanstack/react-router"

import { AuthShell } from "@/components/ds/auth-shell"
import { Button } from "@/components/ds/button"
import { Field } from "@/components/ds/field"
import { FormError } from "@/components/ds/form-error"
import { authClient } from "@/lib/api"
import { messageFrom } from "@/lib/errors"

export function ReplacePassword() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { t } = useTranslation()

  const [currentPassword, setCurrentPassword] = useState("")
  const [newPassword, setNewPassword] = useState("")

  const replacePassword = useMutation({
    mutationFn: () => authClient.replacePassword({ currentPassword, newPassword }),
    onSuccess: async () => {
      await queryClient.invalidateQueries()
      await navigate({ to: "/" })
    },
  })

  return (
    <AuthShell
      eyebrow={t("auth.replacePassword.eyebrow")}
      title={t("auth.replacePassword.title")}
      blurb={t("auth.replacePassword.blurb")}
      footerRight={t("auth.replacePassword.footer")}
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          replacePassword.mutate()
        }}
      >
        <Field
          label={t("auth.replacePassword.temporary")}
          type="password"
          value={currentPassword}
          onChange={(event) => setCurrentPassword(event.target.value)}
          autoComplete="current-password"
          hint={t("auth.replacePassword.temporaryHint")}
          required
        />
        <Field
          label={t("auth.replacePassword.yours")}
          type="password"
          value={newPassword}
          onChange={(event) => setNewPassword(event.target.value)}
          autoComplete="new-password"
          hint={t("auth.setup.passwordHint")}
          required
        />

        <FormError message={messageFrom(replacePassword.error)} />

        <Button type="submit" disabled={replacePassword.isPending} className="self-start">
          {replacePassword.isPending ? t("auth.replacePassword.submitting") : t("auth.replacePassword.submit")}
        </Button>
      </form>
    </AuthShell>
  )
}
