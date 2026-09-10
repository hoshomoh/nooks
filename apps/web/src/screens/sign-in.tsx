import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { getRouteApi, useNavigate } from "@tanstack/react-router"

import { AuthShell } from "@/components/ds/auth-shell"
import { Button } from "@/components/ds/button"
import { Field } from "@/components/ds/field"
import { authClient } from "@/lib/api"
import { messageFrom } from "@/lib/errors"

const route = getRouteApi("/sign-in")

export function SignIn() {
  const { instance } = route.useLoaderData()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")

  const signIn = useMutation({
    mutationFn: () => authClient.signIn({ email, password }),
    onSuccess: async () => {
      await queryClient.invalidateQueries()
      await navigate({ to: "/" })
    },
  })

  return (
    <AuthShell
      eyebrow="Signing in to"
      title={instance.name}
      blurb="Your lists are behind this."
      footerLeft={<span className="text-shared">Forgot your password?</span>}
      footerRight="New here? Ask to join from the list."
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          signIn.mutate()
        }}
      >
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
          autoComplete="current-password"
          required
        />

        {signIn.isError && (
          <p className="text-destructive text-[13.5px] leading-[1.5]">
            {messageFrom(signIn.error)}
          </p>
        )}

        <Button type="submit" disabled={signIn.isPending} className="self-start">
          {signIn.isPending ? "Signing in…" : "Sign in"}
        </Button>
      </form>
    </AuthShell>
  )
}
