import { Mark } from "@nooks/design/mark"

import { ThemeToggle } from "@/components/theme-toggle"

/** What somebody has to type to have Nooks running, and nothing more. */
const RUN = `go build -o nooks ./cmd/nooks
./nooks --data ./data`

/**
 * The landing page.
 *
 * It answers three questions in the order somebody actually asks them: what is this,
 * why is it different, and how do I run it. No pricing, no signup, no waitlist —
 * there is nothing to sell, and the only call to action is a command you can paste.
 */
export default function Home() {
  return (
    <main className="mx-auto flex max-w-content flex-col gap-16 px-5.5 pt-20 pb-24">
      <header className="flex flex-col gap-5">
        <span className="flex items-center gap-2.5">
          <Mark size={28} />
          <span className="text-page">nooks</span>
          <span className="flex-1" />
          <ThemeToggle />
        </span>

        <h1 className="text-display max-w-135">A household todo app you run on your own machine.</h1>

        <p className="max-w-135 text-body leading-[1.65] text-secondary-foreground">
          One list primitive, one action to add something, and a printed page that is a real
          deliverable rather than a fallback. No email server to configure, no telemetry, and
          no account anywhere but yours.
        </p>
      </header>

      <Section
        title="It looks like a document, not a dashboard"
        body="The signature element is the hairline rule under a list title: everything on screen hangs off it, and it is the one mark that survives onto paper. There are no cards, no widgets and no charts, because a shopping list is not a metric."
      />

      <Section
        title="The printed page is the point"
        body="Most apps print a screenshot of themselves. Nooks prints an A4 sheet with a real checkbox, the quantities a shopper needs, and three blank rows for whatever gets remembered in the shop. The sidebar and the chrome do not come with it."
      />

      <Section
        title="Everything stays on your machine"
        body="One binary serving the API and the app from the same process, and one SQLite file you can copy. Export is that file — nothing is converted, so there is no format of ours to fall behind the schema. There is no telemetry to turn off, because there is none."
      />

      <section className="flex flex-col gap-4">
        <h2 className="text-page">Running it</h2>
        <p className="max-w-135 text-body leading-[1.65] text-secondary-foreground">
          Nooks is one binary. The first person to arrive creates their account and names the
          instance; everyone else asks to join, and an admin approves.
        </p>
        <pre className="overflow-x-auto rounded-xl border border-border bg-sidebar px-5 py-4 font-mono text-meta text-secondary-foreground">
          <code>{RUN}</code>
        </pre>
        <p className="text-micro text-muted-foreground">
          Then open localhost:8081. Postgres instead of SQLite is a flag, not a rewrite.
        </p>
      </section>

      <footer className="flex flex-wrap items-center gap-3.5 border-t border-hair pt-5 text-meta text-secondary-foreground">
        <span>AGPL-3.0. The source is yours, and so is anything you change.</span>
        <a
          href="https://github.com/hoshomoh/nooks"
          className="text-shared hover:underline"
        >
          Read the source
        </a>
      </footer>
    </main>
  )
}

interface SectionProps {
  title: string
  body: string
}

/** One claim, and the sentence that earns it. */
function Section({ title, body }: SectionProps) {
  return (
    <section className="flex flex-col gap-2.5">
      <h2 className="text-page">{title}</h2>
      <p className="max-w-135 text-body leading-[1.65] text-secondary-foreground">{body}</p>
    </section>
  )
}
