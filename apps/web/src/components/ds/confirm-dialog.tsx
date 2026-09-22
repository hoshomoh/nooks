import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "./button"
import { DIALOG_SURFACE } from "./dialog-surface"
import { Dialog, DialogContent } from "@/components/ui/dialog"

export interface ConfirmDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  /**
   * What will happen, in words.
   *
   * DESIGN.md §9: say what happens, not "are you sure?". A Member who is told what a
   * button does can decide; a Member who is asked whether they are sure can only guess.
   */
  blurb: string
  confirmLabel: string
  /** Outlined, never filled — deleting is not the page's primary action. */
  destructive?: boolean
  /**
   * A control the decision needs, drawn under the blurb.
   *
   * Most confirmations are a question with two answers and carry none. Removing a Member
   * is the one that cannot be: what becomes of the Lists they started is a second
   * decision, and asking it anywhere but here would be asking it after the fact.
   */
  children?: ReactNode
  onConfirm: () => void
}

/** A dialog that asks once, per DESIGN.md §9: consequence left, Cancel then confirm right. */
export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  blurb,
  confirmLabel,
  destructive,
  children,
  onConfirm,
}: ConfirmDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton={false}
        className={DIALOG_SURFACE}
      >
        <header className="flex flex-col gap-2 px-6.5 pt-6 pb-5">
          <h2 className="text-dialog">{title}</h2>
          <p className="text-field text-secondary-foreground">{blurb}</p>
          {children}
        </header>

        <footer className="flex items-center gap-3 border-t border-hair px-6.5 py-3.5">
          <span className="flex-1" />
          <Button tone="secondary" onClick={() => onOpenChange(false)}>
            {t("action.cancel")}
          </Button>
          <Button
            tone={destructive ? "destructive" : "primary"}
            onClick={() => {
              onConfirm()
              onOpenChange(false)
            }}
          >
            {confirmLabel}
          </Button>
        </footer>
      </DialogContent>
    </Dialog>
  )
}
