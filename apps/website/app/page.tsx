import Link from "next/link"

import { AppPreview } from "@/components/app-preview"
import { SOURCE } from "@/lib/site"

/** What somebody has to type to have Nooks running, and nothing more. */
const RUN = `go build -o nooks ./cmd/nooks
./nooks --data ./data`

/** The four claims under the fold, each one a sentence that earns it. */
const CLAIMS = [
  {
    title: "One primitive",
    body: "A nook is a list. No boards, no projects, no hierarchy to maintain. Today and Upcoming filter what is already there rather than holding copies of it.",
  },
  {
    title: "One way in",
    body: "Type the line. A quantity and a date are read off the end of it, so “oat milk 2 saturday” becomes one item with two chips attached — and “Call 2 plumbers” stays exactly what you wrote.",
  },
  {
    title: "Made to be shared",
    body: "Share a list with the household, publish a read-only page, or print it. The paper version is designed, not screenshotted.",
  },
  {
    title: "Yours to run",
    body: "One binary, one SQLite file. Nothing leaves the machine, because there is nothing in it that phones anywhere.",
  },
] as const

/** The questions a self-hoster asks before they run anything. */
const FAQS = [
  {
    q: "Is Nooks free?",
    a: "Yes. AGPL-3.0, with no paid tier and no seat pricing. You pay for whatever you run it on, and anything you change stays yours on the same terms.",
  },
  {
    q: "Can people without accounts read a list?",
    a: "A list can be published as a read-only page at a stable link, which anybody can open and print. Ticking needs an account, and the link is revocable.",
  },
  {
    q: "What happens when the connection drops?",
    a: "The app says so, keeps showing what it last read, and tells you when it last had contact. Queuing changes made while offline is the piece being built now.",
  },
  {
    q: "Where does my data live?",
    a: "In the database you configured, on the host you chose. There is no cloud component and no telemetry to turn off, because there is none to begin with.",
  },
] as const

/**
 * The landing page.
 *
 * It answers the questions in the order somebody actually asks them: what is this, why
 * is it different, how do I run it, and what can I build on it. No pricing, no signup,
 * no waitlist — there is nothing to sell, and the call to action is a command you paste.
 */
export default function Home() {
  return (
    <main>
      <section className="mx-auto flex max-w-band flex-wrap items-start gap-14 px-7 pt-19">
        <div className="flex min-w-0 flex-[1_1_360px] flex-col gap-5.5">
          <p className="text-small text-muted-foreground">
            Access tokens and the MCP server are in.
          </p>
          <h1 className="text-hero text-pretty">A list you can hand to someone.</h1>
          <p className="max-w-[46ch] text-lede text-pretty text-secondary-foreground">
            Nooks is a self-hosted list app for a household. One line to add something, one page
            to read it on, and a printed sheet that is a real deliverable rather than a fallback.
          </p>
          <div className="flex flex-wrap gap-2.5 pt-0.5">
            <Link
              href="/docs"
              className="flex h-8.5 items-center rounded-md bg-primary px-4 text-chrome font-medium text-primary-foreground"
            >
              Install Nooks
            </Link>
            <Link
              href="/features"
              className="flex h-8.5 items-center rounded-md border border-border px-3.5 text-chrome text-secondary-foreground hover:text-foreground"
            >
              See what it does
            </Link>
          </div>
          <ul className="flex flex-wrap gap-4.5 pt-1.5 text-small text-muted-foreground">
            <li>Self-hosted</li>
            <li>AGPL-3.0</li>
            <li>SQLite or Postgres</li>
            <li>No telemetry</li>
          </ul>
        </div>

        <div className="min-w-0 flex-[1_1_420px]">
          <AppPreview />
        </div>
      </section>

      <Band title="Not a project tracker.">
        <div className="grid gap-x-10 gap-y-9 sm:grid-cols-2">
          {CLAIMS.map((claim) => (
            <div key={claim.title} className="flex flex-col gap-2">
              <h3 className="text-note-heading">{claim.title}</h3>
              <p className="text-field text-secondary-foreground">{claim.body}</p>
            </div>
          ))}
        </div>
      </Band>

      <Band title="Your server, your machine.">
        <div className="flex flex-wrap items-start gap-10">
          <div className="flex min-w-0 flex-[1_1_320px] flex-col gap-4">
            <p className="text-field text-secondary-foreground">
              One binary serves the API and the app from the same process, and keeps everything
              in one SQLite file you can copy. Postgres instead is a flag, not a rewrite.
            </p>
            <Link href="/docs" className="w-fit text-meta text-shared hover:underline">
              How to install it →
            </Link>
          </div>
          <div className="min-w-0 flex-[1_1_380px] overflow-hidden rounded-xl border border-border bg-sidebar">
            <div className="border-b border-hair px-4 py-2 text-micro text-muted-foreground">
              From source
            </div>
            <pre className="overflow-x-auto px-4 py-3.5 font-mono text-meta text-secondary-foreground">
              <code>{RUN}</code>
            </pre>
          </div>
        </div>
      </Band>

      <Band title="Built for scripts as much as people.">
        <div className="grid gap-x-10 gap-y-9 sm:grid-cols-2">
          <div className="flex flex-col gap-2">
            <h3 className="text-note-heading">REST API</h3>
            <p className="text-field text-secondary-foreground">
              Everything the app does is an endpoint under <Code>/api/v1</Code>, authenticated
              with a bearer token you issue yourself. The reference is generated from the same
              definitions the server is built from, so it cannot describe an endpoint that
              is not there.
            </p>
            <Link href="/docs/api" className="w-fit text-meta text-shared hover:underline">
              API reference →
            </Link>
          </div>
          <div className="flex flex-col gap-2">
            <h3 className="text-note-heading">MCP server</h3>
            <p className="text-field text-secondary-foreground">
              Point an assistant at your instance and it reaches exactly what its token reaches
              — the same rules the browser gets, not a second set. Every change it makes is
              attributed to the token in the list’s history.
            </p>
            <Link href="/docs/mcp" className="w-fit text-meta text-shared hover:underline">
              MCP setup →
            </Link>
          </div>
        </div>
      </Band>

      <Band title="A few useful answers.">
        <div className="grid gap-x-10 gap-y-9 sm:grid-cols-2">
          {FAQS.map((faq) => (
            <div key={faq.q} className="flex flex-col gap-2">
              <h3 className="text-note-heading">{faq.q}</h3>
              <p className="text-field text-secondary-foreground">{faq.a}</p>
            </div>
          ))}
        </div>
      </Band>

      <section className="mx-auto mt-20 flex max-w-band flex-wrap items-center gap-6 rounded-xl border border-border bg-sidebar px-8 py-9">
        <div className="flex min-w-0 flex-1 flex-col gap-1.5">
          <h2 className="text-section">Start with one list.</h2>
          <p className="text-field text-secondary-foreground">
            The first person to arrive creates their account and names the instance.
          </p>
        </div>
        <div className="flex flex-wrap gap-2.5">
          <Link
            href="/docs/deploy"
            className="flex h-8.5 items-center rounded-md bg-primary px-4 text-chrome font-medium text-primary-foreground"
          >
            Install Nooks
          </Link>
          <a
            href={SOURCE}
            className="flex h-8.5 items-center rounded-md border border-border px-3.5 text-chrome text-secondary-foreground hover:text-foreground"
          >
            Read the source
          </a>
        </div>
      </section>
    </main>
  )
}

/** One band of the page: a heading on a rule, and whatever it introduces. */
function Band({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="mx-auto mt-22 max-w-band px-7">
      <div className="mb-8 flex items-center gap-5">
        <h2 className="text-band flex-none text-pretty">{title}</h2>
        <span className="h-px flex-1 bg-border" />
      </div>
      {children}
    </section>
  )
}

/** An inline path or flag, at the weight the rest of the sentence reads at. */
function Code({ children }: { children: React.ReactNode }) {
  return (
    <span className="rounded border border-hair bg-sidebar px-1 py-0.5 font-mono text-meta">
      {children}
    </span>
  )
}
