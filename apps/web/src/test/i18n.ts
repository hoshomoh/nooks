import i18n from "@/i18n"

/**
 * Waits for the app's own i18next to have English in hand.
 *
 * The real pipeline, lazy locale file and all, rather than a second instance set up for
 * tests: a test that renders words asserts on the words a Member would see, so a key
 * that was never translated fails here rather than shipping as `list.addItem`.
 */
export async function readyForEnglish(): Promise<void> {
  await i18n.changeLanguage("en")
}
