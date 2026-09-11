import { useTranslation } from "react-i18next"

import { Button } from "./button"
import { Menu, MenuItem, MenuSeparator } from "./menu"

export interface ItemMenuActions {
  onEdit: () => void
  onSetDate: () => void
  onSetQuantity: () => void
  onDuplicate: () => void
  onDelete: () => void
}

export interface ItemMenuProps {
  actions: ItemMenuActions
}

/**
 * The `···` on an Item's row, per the design's row menu.
 *
 * Three of the entries the design lists are not here: moving between Lists, assigning
 * to somebody, and the note shortcut all wait for the milestones that build them. An
 * entry that does nothing teaches a Member that the menu is decoration.
 */
export function ItemMenu({ actions }: ItemMenuProps) {
  const { t } = useTranslation()

  return (
    <Menu
      trigger={
        <Button tone="quiet" scale="toolbar" aria-label={t("itemMenu.open")}>
          <Dots />
        </Button>
      }
    >
      <MenuItem shortcut="↵" onSelect={actions.onEdit}>
        {t("itemMenu.edit")}
      </MenuItem>
      <MenuItem shortcut="D" onSelect={actions.onSetDate}>
        {t("itemMenu.setDate")}
      </MenuItem>
      <MenuItem shortcut="Q" onSelect={actions.onSetQuantity}>
        {t("itemMenu.setQuantity")}
      </MenuItem>

      <MenuSeparator />

      <MenuItem onSelect={actions.onDuplicate}>{t("itemMenu.duplicate")}</MenuItem>
      <MenuItem destructive shortcut="⌫" onSelect={actions.onDelete}>
        {t("itemMenu.delete")}
      </MenuItem>
    </Menu>
  )
}

/** The three dots, drawn rather than typed: `···` is punctuation, not an icon. */
function Dots() {
  return (
    <svg width="13" height="13" viewBox="0 0 14 14" fill="currentColor" aria-hidden>
      <circle cx="3" cy="7" r="1.3" />
      <circle cx="7" cy="7" r="1.3" />
      <circle cx="11" cy="7" r="1.3" />
    </svg>
  )
}
