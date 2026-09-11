import { useTranslation } from "react-i18next"

import { DotsButton } from "./dots-button"
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
      trigger={<DotsButton aria-label={t("itemMenu.open")} />}
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
