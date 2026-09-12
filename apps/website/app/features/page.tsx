import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Features — nooks",
  description: "Everything Nooks does today, grouped by what it is for.",
}

/** One thing Nooks does. `building` marks what is in flight rather than shipped. */
interface Feature {
  name: string
  text: string
  building?: boolean
}

interface FeatureGroup {
  id: string
  number: string
  name: string
  blurb: string
  items: Feature[]
}

/*
 * What is on this page is what runs.
 *
 * Nothing here is a plan. The one exception carries a badge saying so, because a
 * features page that quietly lists intentions is the thing a self-hoster finds out
 * about after they have moved their household onto it.
 */
const GROUPS: FeatureGroup[] = [
  {
    id: "capture",
    number: "01",
    name: "Capture",
    blurb: "The shortest path from noticing something to it being on the list.",
    items: [
      {
        name: "One-line add",
        text: "The composer sits at the bottom of every list and keeps focus after Enter, so ten things go on in ten lines without touching the mouse.",
      },
      {
        name: "Inline quantities",
        text: "Type 1kg or x3 and it becomes a quantity beside the label rather than part of it. Quantities are free text, because “a bunch” is a quantity.",
      },
      {
        name: "Inline dates",
        text: "“saturday”, “tomorrow” and “14 oct” are read as you type, in the language the reader set, with a control to correct what was understood.",
      },
      {
        name: "Notes",
        text: "A Markdown note attaches to an item without lengthening the row: the first line previews under it, the rest opens in a sheet.",
      },
      {
        name: "Command palette",
        text: "Everything reachable from the keyboard, including jumping to a list by name.",
      },
      {
        name: "Offline queue",
        text: "Adds and ticks made with no signal land when it returns.",
        building: true,
      },
    ],
  },
  {
    id: "read",
    number: "02",
    name: "Read",
    blurb: "Views filter and sort the lists you already have. Nothing is duplicated.",
    items: [
      {
        name: "Today and Upcoming",
        text: "Cross-list views of everything dated, with whatever is overdue gathered into Today rather than hidden behind it.",
      },
      {
        name: "Calendar",
        text: "The same dated items by month, for deciding when something should happen rather than what is next.",
      },
      {
        name: "Search",
        text: "Full-text across labels and notes, over every list the reader can reach. Built in from the first release, not added once it got slow.",
      },
      {
        name: "Live updates",
        text: "A tick somebody else makes arrives without a refresh, and the list shows who else is reading it.",
      },
      {
        name: "Light and dark",
        text: "Per-member, with a system option. The palette is one file shared with this site, so they cannot drift apart.",
      },
      {
        name: "Your language",
        text: "Localised from the first release, dates included — not retrofitted once somebody asked.",
      },
    ],
  },
  {
    id: "share",
    number: "03",
    name: "Share",
    blurb: "Getting the list to the person who needs it, on screen or on paper.",
    items: [
      {
        name: "Household sharing",
        text: "Invite people to the instance, then share a list with all of them, with named people, or with a group.",
      },
      {
        name: "Groups",
        text: "A group is a shortcut for sharing and nothing else, so it never becomes a second hierarchy to maintain.",
      },
      {
        name: "Public pages",
        text: "Publish a list as a read-only page at a stable link. A visitor cannot see that any other list exists, and the link is revocable.",
      },
      {
        name: "Print design",
        text: "A real print stylesheet: 27pt title, 16pt items, a drawn checkbox, and the quantities a shopper needs. The sidebar does not come with it.",
      },
      {
        name: "Export",
        text: "The database itself, not a format of ours. On SQLite that is one file you can copy, which is the whole instance and cannot fall behind the schema.",
      },
      {
        name: "History",
        text: "Who added a thing and who ticked it, including when a script or an assistant did it through a token.",
      },
    ],
  },
  {
    id: "run",
    number: "04",
    name: "Run",
    blurb: "One binary, a database you choose, and nothing phoning home.",
    items: [
      {
        name: "Self-hosted",
        text: "One binary serving the API and the app from the same process. No queue, no mail server, no sidecars.",
      },
      {
        name: "SQLite or Postgres",
        text: "SQLite for a household, Postgres once several people write at once. The test suite runs against both, so the second one cannot rot.",
      },
      {
        name: "Access tokens",
        text: "Scoped to named lists, with read, write and delete as separate abilities, an expiry, and an audit trail. A list a token was not given is invisible to it.",
      },
      {
        name: "REST API",
        text: "Everything the app does, under /api/v1 with a bearer token. A check in CI refuses an endpoint that ships to the browser only.",
      },
      {
        name: "MCP server",
        text: "Built in. An assistant reaches exactly what its token reaches — a test checks that set against the service definitions, so the two cannot drift.",
      },
      {
        name: "No telemetry",
        text: "There is nothing to turn off. Nothing leaves the machine unless you point it somewhere yourself.",
      },
    ],
  },
]

export default function Features() {
  return (
    <main className="mx-auto max-w-band px-7 pt-19">
      <div className="flex flex-col gap-5">
        <span className="text-label text-muted-foreground uppercase">Features</span>
        <h1 className="text-hero max-w-[18ch] text-pretty">Everything hangs off the list.</h1>
        <p className="max-w-[54ch] text-lede text-pretty text-secondary-foreground">
          Each of these exists because a household hit the wall without it. What is not here is
          not here — there is no roadmap wearing a product’s clothes on this page.
        </p>
      </div>

      <nav className="mt-10 grid gap-px overflow-hidden rounded-xl border border-border bg-border sm:grid-cols-4">
        {GROUPS.map((group) => (
          <a
            key={group.id}
            href={`#${group.id}`}
            className="flex items-center gap-2.5 bg-sidebar px-4 py-3.5 text-chrome text-secondary-foreground hover:text-foreground"
          >
            <span className="font-mono text-micro text-muted-foreground">{group.number}</span>
            {group.name}
            <span className="flex-1" />
            <span className="font-mono text-micro text-muted-foreground">
              {String(group.items.length).padStart(2, "0")}
            </span>
          </a>
        ))}
      </nav>

      {GROUPS.map((group) => (
        <section key={group.id} id={group.id} className="scroll-mt-20 pt-16">
          <div className="flex flex-col gap-2 border-b border-border pb-6">
            <span className="font-mono text-micro text-muted-foreground">{group.number}</span>
            <h2 className="text-band">{group.name}</h2>
            <p className="max-w-[54ch] text-field text-secondary-foreground">{group.blurb}</p>
          </div>

          <div className="grid gap-x-10 gap-y-8 pt-8 sm:grid-cols-2">
            {group.items.map((item) => (
              <div key={item.name} className="flex flex-col gap-1.5">
                <h3 className="flex items-center gap-2 text-note-heading">
                  {item.name}
                  {item.building && (
                    <span className="rounded border border-offline-line bg-offline-bg px-1.5 py-0.5 text-micro font-normal text-offline-text">
                      being built
                    </span>
                  )}
                </h3>
                <p className="text-field text-secondary-foreground">{item.text}</p>
              </div>
            ))}
          </div>
        </section>
      ))}
    </main>
  )
}
