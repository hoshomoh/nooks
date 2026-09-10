/**
 * An empty state, per DESIGN.md §8: dashed, left-aligned, never centred, no
 * illustration. A statement of fact, then one sentence saying what the thing is for.
 */
export type EmptyStateProps = {
  /** A statement of fact, not an apology. */
  title: string
  /** One sentence saying what the thing is for. */
  body: string
}

export function EmptyState({ title, body }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-start gap-2.5 rounded-xl border border-dashed border-border px-6 py-7">
      <span className="text-empty">{title}</span>
      <span className="max-w-[440px] text-chrome leading-[1.6] text-secondary-foreground">{body}</span>
    </div>
  )
}
