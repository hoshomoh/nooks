/**
 * What a function needs in order to write words a Member can read.
 *
 * Named once rather than written into each signature, so a helper that is not a
 * component can still be given the Member's language without taking a dependency on
 * react-i18next.
 */
export type Translate = (key: string, options?: Record<string, unknown>) => string
