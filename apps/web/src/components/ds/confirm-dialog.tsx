import { useTranslation } from "react-i18next"

import { Button } from "./button"
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
  onConfirm,
}: ConfirmDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton={false}
        className="w-full max-w-dialog gap-0 rounded-2xl p-0 sm:max-w-dialog"
      >
        <header className="flex flex-col gap-2 px-6.5 pt-6 pb-5">
          <h2 className="text-dialog">{title}</h2>
          <p className="text-field text-secondary-foreground">{blurb}</p>
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
