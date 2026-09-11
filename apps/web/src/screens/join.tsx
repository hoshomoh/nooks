import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { RequestStatus } from "@nooks/api"

import { AuthShell } from "@/components/ds/auth-shell"
import { Button } from "@/components/ds/button"
import { Field } from "@/components/ds/field"
import { FormError } from "@/components/ds/form-error"
import { NotePanel } from "@/components/ds/note-panel"
import { authClient } from "@/lib/api"
import { applyPendingTick } from "@/lib/apply-pending-tick"
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

type AskToJoinProps = {
  /** Called with the request identifier once it has been sent. */
  onSent: (requestUid: string) => void
}

function AskToJoin({ onSent }: AskToJoinProps) {
  const { t } = useTranslation()
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
      title={t("auth.join.title")}
      blurb={t("auth.join.blurb")}
      footerRight={t("auth.join.footer")}
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          requestJoin.mutate()
        }}
      >
        <Field
          label={t("auth.setup.yourName")}
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder={t("auth.join.namePlaceholder")}
          autoComplete="name"
          required
        />
        <Field
          label={t("auth.setup.email")}
          type="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          placeholder={t("auth.join.emailPlaceholder")}
          autoComplete="email"
          required
        />
        <Field
          label={t("auth.join.message")}
          value={message}
          onChange={(event) => setMessage(event.target.value)}
          placeholder={t("auth.join.messagePlaceholder")}
        />

        <FormError message={messageFrom(requestJoin.error)} />

        <Button type="submit" disabled={requestJoin.isPending} className="self-start">
          {requestJoin.isPending ? t("auth.join.submitting") : t("auth.join.submit")}
        </Button>
      </form>
    </AuthShell>
  )
}

type CheckJoinRequestProps = {
  requestUid: string
  /** Forget this request and ask again from the start. */
  onStartOver: () => void
}

function CheckJoinRequest({ requestUid, onStartOver }: CheckJoinRequestProps) {
  const { t } = useTranslation()
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
      // Somebody who asked to join from the public list reached for something first.
      await applyPendingTick()
      await queryClient.invalidateQueries()
      await navigate({ to: "/" })
    },
  })

  const approved = request.data?.status === RequestStatus.APPROVED

  if (!approved) {
    return (
      <AuthShell
        eyebrow={t("auth.join.sentEyebrow")}
        title={t("auth.join.sentTitle")}
        blurb={t("auth.join.sentBlurb")}
        footerLeft={
          <button type="button" onClick={onStartOver} className="text-shared">
            {t("auth.join.askAgain")}
          </button>
        }
        footerRight={t("auth.join.sentFooter")}
      >
        <Button
          tone="secondary"
          onClick={() => request.refetch()}
          disabled={request.isFetching}
          className="self-start"
        >
          {request.isFetching ? t("auth.join.checking") : t("auth.join.checkAgain")}
        </Button>
      </AuthShell>
    )
  }

  return (
    <AuthShell
      eyebrow={t("auth.join.approvedEyebrow")}
      title={t("auth.join.approvedTitle")}
      blurb={t("auth.join.approvedBlurb")}
      footerRight={t("auth.join.approvedFooter")}
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          completeJoin.mutate()
        }}
      >
        <NotePanel label={t("auth.join.joiningAs")}>{request.data?.email}</NotePanel>

        <Field
          label={t("palette.name")}
          value={name || (request.data?.name ?? "")}
          onChange={(event) => setName(event.target.value)}
          hint={t("auth.join.nameHint")}
          autoComplete="name"
        />
        <Field
          label={t("auth.setup.password")}
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          hint={t("auth.setup.passwordHint")}
          autoComplete="new-password"
          required
        />

        <FormError message={messageFrom(completeJoin.error)} />

        <Button type="submit" disabled={completeJoin.isPending} className="self-start">
          {completeJoin.isPending ? t("auth.join.joining") : t("auth.join.submitJoin")}
        </Button>
      </form>
    </AuthShell>
  )
}
