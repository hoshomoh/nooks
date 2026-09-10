import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"

import { AuthShell } from "@/components/ds/auth-shell"
import { Button } from "@/components/ds/button"
import { Field } from "@/components/ds/field"
import { authClient } from "@/lib/api"
import { messageFrom } from "@/lib/errors"

export function Setup() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [name, setName] = useState("")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [instanceName, setInstanceName] = useState("")

  const completeSetup = useMutation({
    mutationFn: () => authClient.completeSetup({ name, email, password, instanceName }),
    onSuccess: async () => {
      // The Instance is now named and there is a session, so both answers are stale.
      await queryClient.invalidateQueries()
      await navigate({ to: "/" })
    },
  })

  return (
    <AuthShell
      title="Nooks is running."
      blurb="Make yourself an account and name the instance. Everything else can wait until someone asks for it."
      footerRight="People ask to join from the public list; you approve them"
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          completeSetup.mutate()
        }}
      >
        <Field
          label="Your name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          autoComplete="name"
          required
        />
        <Field
          label="Email"
          type="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          autoComplete="email"
          required
        />
        <Field
          label="Password"
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          autoComplete="new-password"
          hint="Twelve characters or more. No other rules."
          required
        />
        <Field
          label="What should this instance be called?"
          value={instanceName}
          onChange={(event) => setInstanceName(event.target.value)}
          hint="Shows in the sidebar and on printed lists."
          required
        />

        {completeSetup.isError && (
          <p className="text-destructive text-secondary">
            {messageFrom(completeSetup.error)}
          </p>
        )}

        <Button type="submit" disabled={completeSetup.isPending} className="self-start">
          {completeSetup.isPending ? "Creating…" : "Create the instance"}
        </Button>
      </form>
    </AuthShell>
  )
}
