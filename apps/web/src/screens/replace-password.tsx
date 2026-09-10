import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"

import { AuthShell } from "@/components/ds/auth-shell"
import { Button } from "@/components/ds/button"
import { Field } from "@/components/ds/field"
import { authClient } from "@/lib/api"
import { messageFrom } from "@/lib/errors"

export function ReplacePassword() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

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
      eyebrow="An admin made you an account"
      title="Replace the temporary password"
      blurb="They had to pick the first one, so it can't stay. Choose your own now and theirs stops working."
      footerRight="Required before you can use Nooks"
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          replacePassword.mutate()
        }}
      >
        <Field
          label="Temporary password"
          type="password"
          value={currentPassword}
          onChange={(event) => setCurrentPassword(event.target.value)}
          autoComplete="current-password"
          hint="The one they gave you."
          required
        />
        <Field
          label="Your password"
          type="password"
          value={newPassword}
          onChange={(event) => setNewPassword(event.target.value)}
          autoComplete="new-password"
          hint="Twelve characters or more. No other rules."
          required
        />

        {replacePassword.isError && (
          <p className="text-destructive text-secondary">
            {messageFrom(replacePassword.error)}
          </p>
        )}

        <Button type="submit" disabled={replacePassword.isPending} className="self-start">
          {replacePassword.isPending ? "Saving…" : "Save and continue"}
        </Button>
      </form>
    </AuthShell>
  )
}
