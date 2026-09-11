export interface FormErrorProps {
  /** What went wrong, or nothing. */
  message?: string
}

/**
 * What a form says when it could not do the thing.
 *
 * The line is always there, whether or not it says anything. A message that appears out
 * of nowhere grows the form, and on a page whose column is centred that moves every
 * field under the Member's cursor at the moment they are reading why it failed.
 *
 * Announced politely, because somebody using a screen reader is told nothing at all by
 * a line that simply appears.
 */
export function FormError({ message }: FormErrorProps) {
  return (
    <p aria-live="polite" className="min-h-5 text-meta text-destructive">
      {message}
    </p>
  )
}
