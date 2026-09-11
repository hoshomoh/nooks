import { useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "./button"
import { Field } from "./field"
import { DIALOG_SURFACE } from "./dialog-surface"
import { Dialog, DialogContent } from "@/components/ui/dialog"

export interface PromptDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  /** One line of orientation under the title. */
  blurb: string
  /** What the field is called. Always present: a placeholder is not a label. */
  label: string
  /** What the field starts with. */
  initialValue: string
  confirmLabel: string
  onConfirm: (value: string) => void
}

/**
 * A dialog that asks for one line of text.
 *
 * Renaming a List is the case it exists for. The field is keyed on what it started
 * with, so opening it again starts from what is there now rather than from what was
 * typed the last time — and nothing has to be synchronised to keep that true.
 */
export function PromptDialog({
  open,
  onOpenChange,
  title,
  blurb,
  label,
  initialValue,
  confirmLabel,
  onConfirm,
}: PromptDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton={false}
        className={DIALOG_SURFACE}
      >
        <PromptForm
          key={initialValue}
          title={title}
          blurb={blurb}
          label={label}
          initialValue={initialValue}
          confirmLabel={confirmLabel}
          onCancel={() => onOpenChange(false)}
          onConfirm={(value) => {
            onConfirm(value)
            onOpenChange(false)
          }}
        />
      </DialogContent>
    </Dialog>
  )
}

interface PromptFormProps {
  title: string
  blurb: string
  label: string
  initialValue: string
  confirmLabel: string
  onCancel: () => void
  onConfirm: (value: string) => void
}

/** The form itself, which owns what has been typed so far. */
function PromptForm({
  title,
  blurb,
  label,
  initialValue,
  confirmLabel,
  onCancel,
  onConfirm,
}: PromptFormProps) {
  const { t } = useTranslation()
  const [value, setValue] = useState(initialValue)

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault()
        if (value.trim()) {
          onConfirm(value.trim())
        }
      }}
    >
      <header className="flex flex-col gap-2 px-6.5 pt-6">
        <h2 className="text-dialog">{title}</h2>
        <p className="text-field text-secondary-foreground">{blurb}</p>
      </header>

      <div className="px-6.5 pt-5 pb-6">
        <Field
          label={label}
          value={value}
          onChange={(event) => setValue(event.target.value)}
          autoFocus
        />
      </div>

      <footer className="flex items-center gap-3 border-t border-hair px-6.5 py-3.5">
        <span className="flex-1" />
        <Button tone="secondary" type="button" onClick={onCancel}>
          {t("action.cancel")}
        </Button>
        <Button type="submit">{confirmLabel}</Button>
      </footer>
    </form>
  )
}
