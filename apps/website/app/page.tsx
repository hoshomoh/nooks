/**
 * The landing page.
 *
 * Placeholder while the milestone is built out: what Nooks is, the printed page, one
 * screenshot and how to run it all arrive with their own items. What it proves today is
 * that the website draws from the same tokens the app does.
 */
export default function Home() {
  return (
    <main className="mx-auto flex min-h-dvh max-w-content flex-col justify-center gap-5 px-5.5">
      <h1 className="text-display">nooks</h1>
      <p className="max-w-135 text-chrome leading-[1.55] text-secondary-foreground">
        A household todo app that looks like a document, not a dashboard. One list
        primitive, one action to add something, and a printed page that is a real
        deliverable rather than a fallback.
      </p>
    </main>
  )
}
