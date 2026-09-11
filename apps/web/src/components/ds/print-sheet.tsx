import { useTranslation } from "react-i18next"
import type { Item } from "@nooks/api"

export interface PrintSheetProps {
  /** What the Instance calls itself, for the eyebrow. */
  instanceName: string
  listName: string
  /** The date the sheet was printed, already in the Member's language. */
  printedOn: string
  items: Item[]
}

/** BLANK_ROWS is how many dashed lines the sheet ends with. */
const BLANK_ROWS = 3

/**
 * The printed List, per DESIGN.md §15.
 *
 * A deliverable, not a screenshot. It is rendered into the page and hidden on screen,
 * so printing takes the List rather than the browser's idea of the app — the sidebar,
 * the chrome bar and the menus are all things a person standing in a shop has no use
 * for.
 *
 * It renders **every** Item, ticked ones included: a printed sheet is a snapshot of the
 * List, and somebody who ticked something on the way out still wants to see they did.
 */
export function PrintSheet({ instanceName, listName, printedOn, items }: PrintSheetProps) {
  const { t } = useTranslation()

  return (
    <article className="nooks-print-sheet" aria-hidden>
      <header className="nooks-print-header">
        <p className="nooks-print-eyebrow">{t("print.eyebrow", { instance: instanceName })}</p>
        <div className="nooks-print-title-row">
          <h1 className="nooks-print-title">{listName}</h1>
          <div className="nooks-print-meta">
            <span>{printedOn}</span>
            {/* §15 also asks for a count of people. The browser does not have that
                number — who can reach a List is decided on the server — and a wrong
                one on paper is worse than none, so it waits for the server to say. */}
            <span>{t("print.itemCount", { count: items.length })}</span>
          </div>
        </div>
      </header>

      <ul className="nooks-print-items">
        {items.map((item) => (
          <li key={item.uid} className="nooks-print-item">
            <span className="nooks-print-box" data-done={item.done ? "" : undefined} />
            <span className="nooks-print-label">
              {item.label}
              {item.quantity && <span className="nooks-print-quantity">{item.quantity}</span>}
            </span>
            <span className="nooks-print-who">{item.addedByName}</span>
          </li>
        ))}

        {/* For whatever gets remembered in the shop. */}
        {Array.from({ length: BLANK_ROWS }, (_, index) => (
          <li key={`blank-${index}`} className="nooks-print-item nooks-print-blank">
            <span className="nooks-print-box" />
            <span className="nooks-print-label" />
            <span />
          </li>
        ))}
      </ul>

      <footer className="nooks-print-footer">
        <span>{t("print.footer")}</span>
        <span className="nooks-print-sheet-name">{t("print.sheet")}</span>
      </footer>
    </article>
  )
}
