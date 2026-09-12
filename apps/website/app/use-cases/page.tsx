import type { Metadata } from "next"
import Link from "next/link"

export const metadata: Metadata = {
  title: "Use cases — nooks",
  description: "Who Nooks is for, and what they actually do with it.",
}

/** One kind of household or team, and what they get out of it. */
interface Band {
  id: string
  name: string
  text: string
  points: string[]
}

/*
 * Four, not nine.
 *
 * A note-taking app can claim a journalist and a student because "notes" does not
 * discriminate between them. A household list app is the opposite: who is in the house
 * is the product, so the honest list is short and each entry is different work.
 */
const BANDS: Band[] = [
  {
    id: "household",
    name: "The shared house",
    text: "One instance on a machine in the hallway, one list per room or errand, and a printed sheet on the fridge that matches what is on the phone.",
    points: [
      "Groceries with quantities that read as a person wrote them — “1 kg”, “a bunch” — rather than a number and a unit somebody had to pick",
      "Chores split between the people who live there, with the history saying who actually did it",
      "A read-only link for whoever needs to see the list without having an account on your server",
    ],
  },
  {
    id: "self-hosters",
    name: "Self-hosters",
    text: "One Go binary and one file. Nothing to keep an eye on beyond the backup you already run.",
    points: [
      "SQLite by default, Postgres when you want it, and the suite runs against both",
      "Reverse-proxy friendly, with live updates over one long-lived connection rather than polling",
      "No telemetry, no outbound calls, no licence server — and export is the database file itself",
    ],
  },
  {
    id: "small-teams",
    name: "Small teams",
    text: "Lists that are genuinely lists: stock to reorder, a run of checks, things to bring to the site. Not a tracker pretending to be one.",
    points: [
      "Groups grant access to a set of lists at once, without becoming a hierarchy to maintain",
      "Print a run sheet that is legible at arm’s length, with a real checkbox and three blank rows",
      "Every action available over REST, so whatever you already run can drive it",
    ],
  },
  {
    id: "developers",
    name: "Building on it",
    text: "Everything the interface does is an endpoint, and the reference is generated from the same definitions the server is built from.",
    points: [
      "Bearer tokens scoped to named lists, with read, write and delete as separate abilities and an expiry",
      "An MCP server that reaches exactly what the token reaches, with every change attributed to it",
      "A check in CI that refuses an endpoint which ships to the browser only",
    ],
  },
]

export default function UseCases() {
  return (
    <main className="mx-auto max-w-band px-7 pt-19">
      <div className="flex flex-col gap-5">
        <span className="text-label text-muted-foreground uppercase">Use cases</span>
        <h1 className="text-hero max-w-[18ch] text-pretty">Four households, one primitive.</h1>
        <p className="max-w-[54ch] text-lede text-pretty text-secondary-foreground">
          Nooks is one list, shared four ways. The shape of the work changes; the thing you are
          looking at does not.
        </p>
      </div>

      <div className="flex flex-col">
        {BANDS.map((band) => (
          <section
            key={band.id}
            id={band.id}
            className="flex scroll-mt-20 flex-wrap gap-x-14 gap-y-6 border-t border-border pt-10 pb-12 first:mt-14"
          >
            <div className="flex min-w-0 flex-[1_1_300px] flex-col gap-3">
              <h2 className="text-band text-pretty">{band.name}</h2>
              <p className="max-w-[42ch] text-field text-secondary-foreground">{band.text}</p>
            </div>
            <ul className="flex min-w-0 flex-[1_1_380px] flex-col gap-4">
              {band.points.map((point) => (
                <li key={point} className="flex gap-3 text-field text-secondary-foreground">
                  <span
                    aria-hidden
                    className="mt-2 size-[5px] flex-none rounded-full bg-shared"
                  />
                  <span className="min-w-0">{point}</span>
                </li>
              ))}
            </ul>
          </section>
        ))}
      </div>

      <section className="flex flex-wrap items-center gap-6 rounded-xl border border-border bg-sidebar px-8 py-9">
        <div className="flex min-w-0 flex-1 flex-col gap-1.5">
          <h2 className="text-section">None of these quite you?</h2>
          <p className="text-field text-secondary-foreground">
            It is one list and a printed page. That covers more than it sounds like it does.
          </p>
        </div>
        <Link
          href="/docs"
          className="flex h-8.5 items-center rounded-md bg-primary px-4 text-chrome font-medium text-primary-foreground"
        >
          Try it
        </Link>
      </section>
    </main>
  )
}
