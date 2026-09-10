import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Link, useNavigate } from "@tanstack/react-router"
import { RequestStatus } from "@nooks/api"

import { AuthShell } from "@/components/ds/auth-shell"
import { Button } from "@/components/ds/button"
import { Field } from "@/components/ds/field"
import { authClient } from "@/lib/api"
import { messageFrom } from "@/lib/errors"
import {
  forgetPendingRequest,
  readPendingRequest,
  rememberPendingRequest,
} from "@/lib/pending-request"

/**
 * A forgotten password. Nooks sends no email: an Admin checks it is really you however
 * they like, approves, and you set the new password yourself.
 */
export function ForgotPassword() {
  const [requestUid, setRequestUid] = useState(() => readPendingRequest("reset"))

  if (!requestUid) {
    return <AskForReset onSent={setRequestUid} />
  }
  return <CheckResetRequest requestUid={requestUid} onStartOver={() => setRequestUid(null)} />
}

const backToSignIn = (
  <Link to="/sign-in" className="text-shared">
    Back to sign in
  </Link>
)

function AskForReset({ onSent }: { onSent: (uid: string) => void }) {
  const [emailOrName, setEmailOrName] = useState("")

  const requestReset = useMutation({
    mutationFn: () => authClient.requestPasswordReset({ emailOrName }),
    onSuccess: (res) => {
      rememberPendingRequest("reset", res.requestUid)
      onSent(res.requestUid)
    },
  })

  return (
    <AuthShell
      title="Ask for a reset"
      blurb="Nooks sends no email. An admin sees it in Activity, checks it's really you however they like, and approves it. Then you set a new password yourself."
      footerLeft={backToSignIn}
      footerRight="No email, ever"
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          requestReset.mutate()
        }}
      >
        <Field
          label="Your email or name"
          value={emailOrName}
          onChange={(event) => setEmailOrName(event.target.value)}
          hint="So they know whose account to unlock."
          autoComplete="email"
          required
        />

        {requestReset.isError && (
          <p className="text-destructive text-secondary">{messageFrom(requestReset.error)}</p>
        )}

        <Button type="submit" disabled={requestReset.isPending} className="self-start">
          {requestReset.isPending ? "Sending…" : "Send the request"}
        </Button>
      </form>
    </AuthShell>
  )
}

function CheckResetRequest({
  requestUid,
  onStartOver,
}: {
  requestUid: string
  onStartOver: () => void
}) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [newPassword, setNewPassword] = useState("")

  const request = useQuery({
    queryKey: ["reset-request", requestUid],
    queryFn: () => authClient.getResetRequest({ requestUid }),
  })

  const completeReset = useMutation({
    mutationFn: () => authClient.completePasswordReset({ requestUid, newPassword }),
    onSuccess: async () => {
      forgetPendingRequest("reset")
      await queryClient.invalidateQueries()
      await navigate({ to: "/" })
    },
  })

  if (request.data?.status !== RequestStatus.APPROVED) {
    return (
      <AuthShell
        eyebrow="Request sent"
        title="An admin has it"
        blurb="They'll ask you something only you would know, then approve it. Come back to this address afterwards — the prompt to choose a password will be waiting."
        footerLeft={backToSignIn}
        footerRight="Nothing will arrive by email"
      >
        <div className="flex items-center gap-3">
          <Button
            tone="secondary"
            onClick={() => request.refetch()}
            disabled={request.isFetching}
          >
            {request.isFetching ? "Checking…" : "Check again"}
          </Button>
          <button type="button" onClick={onStartOver} className="text-muted-foreground text-small">
            Ask again
          </button>
        </div>
      </AuthShell>
    )
  }

  return (
    <AuthShell
      eyebrow="An admin approved your request"
      title="Set a new password"
      blurb="Nobody else sees what you choose — not the admin, not the server logs. The approval expires in an hour if you don't use it."
      footerRight=""
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          completeReset.mutate()
        }}
      >
        <Field
          label="New password"
          type="password"
          value={newPassword}
          onChange={(event) => setNewPassword(event.target.value)}
          hint="Twelve characters or more. No other rules."
          autoComplete="new-password"
          required
        />

        {completeReset.isError && (
          <p className="text-destructive text-secondary">{messageFrom(completeReset.error)}</p>
        )}

        <Button type="submit" disabled={completeReset.isPending} className="self-start">
          {completeReset.isPending ? "Saving…" : "Save and sign in"}
        </Button>
      </form>
    </AuthShell>
  )
}
