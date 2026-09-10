import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { RequestStatus } from "@nooks/api"

import { AuthShell } from "@/components/ds/auth-shell"
import { Button } from "@/components/ds/button"
import { Field } from "@/components/ds/field"
import { NotePanel } from "@/components/ds/note-panel"
import { authClient } from "@/lib/api"
import { messageFrom } from "@/lib/errors"
import {
  forgetPendingRequest,
  readPendingRequest,
  rememberPendingRequest,
} from "@/lib/pending-request"

/**
 * Asking for an account, and coming back to see whether an Admin has decided.
 *
 * Which of the three states shows is derived from what the browser remembers and what
 * the server says — there is no step counter to keep in sync.
 */
export function Join() {
  const [requestUid, setRequestUid] = useState(() => readPendingRequest("join"))

  if (!requestUid) {
    return <AskToJoin onSent={setRequestUid} />
  }
  return <CheckJoinRequest requestUid={requestUid} onStartOver={() => setRequestUid(null)} />
}

function AskToJoin({ onSent }: { onSent: (uid: string) => void }) {
  const [name, setName] = useState("")
  const [email, setEmail] = useState("")
  const [message, setMessage] = useState("")

  const requestJoin = useMutation({
    mutationFn: () => authClient.requestJoin({ name, email, message }),
    onSuccess: (res) => {
      rememberPendingRequest("join", res.requestUid)
      onSent(res.requestUid)
    },
  })

  return (
    <AuthShell
      title="Ask to join"
      blurb="An admin approves requests. Public signup is off, so this is the way in."
      footerRight="Nothing is sent by email — the request waits in Nooks"
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          requestJoin.mutate()
        }}
      >
        <Field
          label="Your name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder="What should people call you?"
          autoComplete="name"
          required
        />
        <Field
          label="Email"
          type="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          placeholder="So they know who you are"
          autoComplete="email"
          required
        />
        <Field
          label="Anything to add — optional"
          value={message}
          onChange={(event) => setMessage(event.target.value)}
          placeholder="“It's Til, from upstairs”"
        />

        {requestJoin.isError && (
          <p className="text-destructive text-secondary">{messageFrom(requestJoin.error)}</p>
        )}

        <Button type="submit" disabled={requestJoin.isPending} className="self-start">
          {requestJoin.isPending ? "Sending…" : "Send the request"}
        </Button>
      </form>
    </AuthShell>
  )
}

function CheckJoinRequest({
  requestUid,
  onStartOver,
}: {
  requestUid: string
  onStartOver: () => void
}) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [name, setName] = useState("")
  const [password, setPassword] = useState("")

  const request = useQuery({
    queryKey: ["join-request", requestUid],
    queryFn: () => authClient.getJoinRequest({ requestUid }),
  })

  const completeJoin = useMutation({
    mutationFn: () => authClient.completeJoin({ requestUid, name, password }),
    onSuccess: async () => {
      forgetPendingRequest("join")
      await queryClient.invalidateQueries()
      await navigate({ to: "/" })
    },
  })

  const approved = request.data?.status === RequestStatus.APPROVED

  if (!approved) {
    return (
      <AuthShell
        eyebrow="Request sent"
        title="An admin has it"
        blurb="They'll approve it or they won't. This instance can't send email, so nothing will arrive in your inbox — ask them directly if it's taking a while."
        footerLeft={
          <button type="button" onClick={onStartOver} className="text-shared">
            Ask again
          </button>
        }
        footerRight="Come back to this address once it's approved"
      >
        <Button
          tone="secondary"
          onClick={() => request.refetch()}
          disabled={request.isFetching}
          className="self-start"
        >
          {request.isFetching ? "Checking…" : "Check again"}
        </Button>
      </AuthShell>
    )
  }

  return (
    <AuthShell
      eyebrow="An admin approved your request"
      title="Choose a password"
      blurb="That's the last step. Nobody else sees what you pick — not the admin, not the server logs."
      footerRight="Welcome in"
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          completeJoin.mutate()
        }}
      >
        <NotePanel label="Joining as">{request.data?.email}</NotePanel>

        <Field
          label="Name"
          value={name || (request.data?.name ?? "")}
          onChange={(event) => setName(event.target.value)}
          hint="From your request. Change it if you like."
          autoComplete="name"
        />
        <Field
          label="Password"
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          hint="Twelve characters or more. No other rules."
          autoComplete="new-password"
          required
        />

        {completeJoin.isError && (
          <p className="text-destructive text-secondary">{messageFrom(completeJoin.error)}</p>
        )}

        <Button type="submit" disabled={completeJoin.isPending} className="self-start">
          {completeJoin.isPending ? "Joining…" : "Join"}
        </Button>
      </form>
    </AuthShell>
  )
}
