import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import type { TokenAbilities } from "@nooks/api"
import {
  ASSISTANT_CLIENTS,
  BLOCK_KIND,
  onlyThisMachine,
  setupBlock,
  SITE,
  TOKEN_PLACEHOLDER,
  type AssistantClient,
} from "@nooks/shared"

import { Button } from "@/components/ds/button"
import { CopyButton } from "@/components/ds/copy-button"
import { Field } from "@/components/ds/field"
import { FormError } from "@/components/ds/form-error"
import { SegmentedControl, type Segment } from "@/components/ds/segmented"
import { SelectField, type SelectOption } from "@/components/ds/select-field"
import { SettingsRow } from "@/components/ds/settings-row"
import { SettingsShell } from "@/components/ds/settings-shell"
import { tokenClient } from "@/lib/api"
import { messageFrom } from "@/lib/errors"
import { tokensQuery } from "@/lib/token-queries"
import type { Translate } from "@/lib/translate"
import { useSettingsCounts } from "@/lib/use-settings-counts"
import { useSignedInData } from "@/lib/use-signed-in-data"

/** DOCS is where the long version lives, for the setup that did not work. */
const DOCS = `${SITE}/docs/mcp`

/** How much of the instance an assistant is being given. */
type Permission = "read" | "readAndAdd" | "everything"

const PERMISSIONS: Permission[] = ["read", "readAndAdd", "everything"]

/** ABILITIES says what each answer means on the wire. */
const ABILITIES: Record<Permission, TokenAbilities> = {
  read: { read: true, write: false, delete: false } as TokenAbilities,
  readAndAdd: { read: true, write: true, delete: false } as TokenAbilities,
  everything: { read: true, write: true, delete: true } as TokenAbilities,
}

/** ALL_LISTS is the scope answer that is not one List's uid. */
const ALL_LISTS = "all"

/** The token once it exists, which is the only moment its secret can be read. */
interface MadeToken {
  secret: string
  /** What it is for, said in one line under the title. */
  summary: string
}

/**
 * The Assistants page: one screen that ends in a block to copy across.
 *
 * Setting an assistant up is the same three facts written five different ways: the
 * address, the transport, one header. A Member should not have to cut a token on one page,
 * find their address themselves, and then translate all three into their client's
 * format from a documentation site. This page asks the two questions that change the
 * answer and writes the block out finished.
 */
export function SettingsAssistantsScreen() {
  const { t } = useTranslation()
  const counts = useSettingsCounts()
  const queryClient = useQueryClient()
  const { lists } = useSignedInData()

  const here = typeof window === "undefined" ? "" : window.location.origin

  const [client, setClient] = useState<AssistantClient>("claudeCode")
  const [permission, setPermission] = useState<Permission>("readAndAdd")
  const [scope, setScope] = useState(ALL_LISTS)
  const [address, setAddress] = useState(here)
  const [draft, setDraft] = useState(here)
  // An address only this machine can reach is asked about before it is copied, rather
  // than after the block has failed somewhere else.
  const [asking, setAsking] = useState(onlyThisMachine(here))
  const [made, setMade] = useState<MadeToken | null>(null)

  const scopeLabel = () =>
    scope === ALL_LISTS
      ? t("assistants.allLists")
      : (lists.find((list) => list.uid === scope)?.name ?? "")

  const create = useMutation({
    mutationFn: () =>
      tokenClient.createAccessToken({
        name: t(clientKey(client)),
        abilities: ABILITIES[permission],
        listUids: scope === ALL_LISTS ? [] : [scope],
        allLists: scope === ALL_LISTS,
        // No expiry: this page never asks for one, and an assistant that stops working
        // on a day nobody was told about is worse than one that keeps working.
        expiresAt: "",
      }),
    onSuccess: async (res) => {
      await queryClient.invalidateQueries({ queryKey: tokensQuery.queryKey })
      setMade({ secret: res.secret, summary: summarise(client, scopeLabel(), permission, t) })
    },
  })

  const block = setupBlock({ client, address, token: made?.secret ?? TOKEN_PLACEHOLDER })

  return (
    <SettingsShell active="/settings/assistants" crumb={t("assistants.title")} counts={counts}>
      {/*
       * The shell puts 14px between its children, which on a page of five sections is
       * not enough to tell a section break from the gap under its own label. One child
       * instead, and the rhythm is set here: 28 between sections, 8 within.
       */}
      <div className="flex flex-col gap-7">
        <header className="min-w-0">
          <h1 className="mb-1.5 text-page">{t("assistants.title")}</h1>
          <p className="max-w-160 text-chrome leading-[1.6] text-secondary-foreground">
            {t("assistants.blurb")}
          </p>
        </header>

        <section className="flex flex-col gap-2">
          <span className="text-label text-muted-foreground uppercase">
            {t("assistants.which")}
          </span>
          <SegmentedControl
            label={t("assistants.which")}
            options={ASSISTANT_CLIENTS.map<Segment<AssistantClient>>((one) => ({
              value: one,
              label: t(clientKey(one)),
            }))}
            chosen={client}
            onChoose={setClient}
          />
          <p className="text-meta leading-[1.55] text-secondary-foreground">
            {t(`assistants.note${capitalise(client)}`)}
          </p>
        </section>

        <section className="flex flex-col gap-2">
          <span className="text-label text-muted-foreground uppercase">
            {t("assistants.tokenSection")}
          </span>

          <div className="flex flex-col gap-2.5 rounded-xl border border-border px-4.5 py-3">
            <div className="flex min-w-0 flex-col gap-0.5">
              <span className="text-field font-medium">{t("assistants.newToken")}</span>
              <span className="text-micro leading-[1.5] text-muted-foreground">
                {made ? t("assistants.madeTokenBlurb") : t("assistants.newTokenBlurb")}
              </span>
            </div>

            {made ? (
              <MadeTokenPanel made={made} />
            ) : (
              <>
                <div className="flex flex-col border-t border-hair">
                  <SettingsRow label={t("assistants.lists")}>
                    <SelectField
                      label={t("assistants.lists")}
                      options={[
                        { value: ALL_LISTS, label: t("assistants.allLists") },
                        ...lists.map<SelectOption>((list) => ({
                          value: list.uid,
                          label: list.name,
                        })),
                      ]}
                      value={scope}
                      onValueChange={setScope}
                    />
                  </SettingsRow>

                  <SettingsRow label={t("assistants.mayDo")}>
                    <SegmentedControl
                      label={t("assistants.mayDo")}
                      options={PERMISSIONS.map<Segment<Permission>>((one) => ({
                        value: one,
                        label: t(`assistants.${one}`),
                      }))}
                      chosen={permission}
                      onChoose={setPermission}
                    />
                  </SettingsRow>
                </div>

                <div className="flex flex-col gap-1">
                  <div className="flex items-center gap-3">
                    <Button disabled={create.isPending} onClick={() => create.mutate()}>
                      {create.isPending ? t("assistants.creating") : t("assistants.create")}
                    </Button>
                    <span className="text-micro text-muted-foreground">
                      {t("assistants.createNote")}
                    </span>
                  </div>
                  {create.isError && <FormError message={messageFrom(create.error)} />}
                </div>
              </>
            )}
          </div>
        </section>

        <section className="flex flex-col gap-3.5">
          {asking ? (
            <div className="flex flex-col gap-2">
              <p className="max-w-150 text-meta leading-[1.55] text-secondary-foreground">
                {t("assistants.addressLocal", { address: here })}
              </p>
              <form
                className="flex items-end gap-2.5"
                onSubmit={(event) => {
                  event.preventDefault()
                  setAddress(draft.trim())
                  setAsking(false)
                }}
              >
                <Field
                  label={t("assistants.address")}
                  hideLabel
                  className="w-90 font-mono"
                  value={draft}
                  onChange={(event) => setDraft(event.target.value)}
                />
                <Button type="submit" tone="secondary">
                  {t("assistants.useAddress")}
                </Button>
              </form>
            </div>
          ) : (
            <p className="text-meta leading-[1.6] text-secondary-foreground">
              {t("assistants.connectsTo", { address })}{" "}
              <button
                type="button"
                className="text-shared underline-offset-2 hover:underline"
                onClick={() => {
                  setDraft(address)
                  setAsking(true)
                }}
              >
                {t("assistants.differentAddress")}
              </button>
            </p>
          )}

          <div className="flex flex-col gap-2">
            <div className="overflow-hidden rounded-xl border border-border">
              <div className="flex min-h-10 items-center gap-3 border-b border-hair bg-sidebar py-2 pr-2.5 pl-3.5">
                <span className="shrink-0 text-micro text-muted-foreground">
                  {t(`assistants.${BLOCK_KIND[client]}`)}
                </span>
                {BLOCK_KIND[client] === "file" && (
                  <span className="font-mono text-micro leading-[1.45] break-all text-secondary-foreground">
                    {t(`assistants.where${capitalise(client)}`)}
                  </span>
                )}
                <CopyButton
                  key={block}
                  text={block}
                  label={t("action.copy")}
                  tone="secondary"
                  scale="compact"
                  className="ml-auto shrink-0"
                />
              </div>
              <pre className="overflow-x-auto px-3.5 py-3 font-mono text-micro leading-[1.55] whitespace-pre-wrap text-foreground">
                {block}
              </pre>
            </div>

            <div className="flex flex-col gap-2 text-micro leading-[1.55] text-muted-foreground">
              <span>
                {made
                  ? t(`assistants.blockNote${capitalise(client)}`)
                  : t("assistants.placeholderNote")}{" "}
                <a href={DOCS} target="_blank" rel="noreferrer" className="text-shared">
                  {t("assistants.troubleshooting")}
                </a>
              </span>
              {!made && <span>{t("assistants.ownToken")}</span>}
            </div>
          </div>
        </section>
      </div>
    </SettingsShell>
  )
}

interface MadeTokenPanelProps {
  made: MadeToken
}

/** The secret, in the one moment it can be read, per DESIGN.md §9. */
function MadeTokenPanel({ made }: MadeTokenPanelProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-2 rounded-lg border border-shared-line bg-shared-bg px-4 py-3">
      <div className="flex items-baseline gap-2.5">
        <span className="text-field font-medium">{t("assistants.secretTitle")}</span>
        <span className="text-micro text-secondary-foreground">{made.summary}</span>
      </div>

      <div className="flex items-center gap-3.5">
        <code className="min-w-0 font-mono text-meta leading-[1.6] break-all">{made.secret}</code>
        <CopyButton
          text={made.secret}
          label={t("assistants.copyToken")}
          className="ml-auto shrink-0"
        />
      </div>

      <span className="text-micro leading-[1.55] text-secondary-foreground">
        {t("assistants.secretBlurb")}
      </span>
    </div>
  )
}

/** clientKey is where a client's own name lives, since it is the same in every locale. */
function clientKey(client: AssistantClient): string {
  return `assistants.client${capitalise(client)}`
}

/** summarise says what the token that was just cut is for, in one line. */
function summarise(
  client: AssistantClient,
  lists: string,
  permission: Permission,
  t: Translate,
): string {
  return t("assistants.secretFor", {
    client: t(clientKey(client)),
    lists,
    permission: t(`assistants.${permission}`),
  })
}

/** capitalise makes a camelCase name into the tail of a translation key. */
function capitalise(word: string): string {
  return word.charAt(0).toUpperCase() + word.slice(1)
}
