import type { Metadata } from "next"

import { releases, type Release } from "@/lib/changelog"

export const metadata: Metadata = {
  title: "Changelog — nooks",
  description: "Every released version of Nooks, and what changed in it.",
}

/**
 * What each kind of change is drawn in.
 *
 * The three headings release-please writes, and nothing else: a changelog that sorts
 * itself into eight categories is one nobody scans.
 */
const TONES: Record<string, string> = {
  New: "text-done",
  // The amber the design calls --warn. In the foundations it is --offline, named after
  // the first thing that needed it rather than after what it means; it is the same
  // value, and a second token holding the same colour is how two of them drift apart.
  Changed: "text-offline",
  Fixed: "text-muted-foreground",
}

export default async function Changelog() {
  const published = await releases()

  return (
    <main className="mx-auto max-w-band px-7 pt-19">
      <div className="flex max-w-[620px] flex-col gap-4">
        <span className="text-label text-muted-foreground uppercase">Changelog</span>
        <h1 className="text-page text-pretty">What changed, and when.</h1>
        <p className="text-lede text-pretty text-secondary-foreground">
          Written from the commits themselves, so nothing ships without appearing here.
          Anything that changes the shape of stored data says so at the top of the entry.
        </p>
      </div>

      <div className="pt-11">
        {published.length === 0 ? (
          <p className="border-t border-border pt-7 text-field text-secondary-foreground">
            Nooks has not had a release yet. The first one will appear here.
          </p>
        ) : (
          published.map((release) => <Entry key={release.version} release={release} />)
        )}
      </div>
    </main>
  )
}

/**
 * One release: what it was called and when, then what it did.
 *
 * The version sits in a column of its own so a reader scanning for the one they are on
 * reads down a single line rather than across every entry.
 */
function Entry({ release }: { release: Release }) {
  return (
    <div className="flex flex-wrap gap-x-10 gap-y-2 border-t border-border py-7">
      <div className="flex w-50 flex-none flex-col gap-1.5">
        <span className="font-mono text-field font-medium">{release.version}</span>
        <span className="text-meta text-muted-foreground">{release.date}</span>
        {release.migration && (
          <span className="self-start rounded-md border border-offline-line px-1.5 py-0.5 text-badge text-offline">
            Migration
          </span>
        )}
      </div>

      <div className="flex min-w-0 flex-1 basis-95 flex-col gap-3.5">
        {release.changes.map((change, index) => (
          <div key={index} className="flex items-baseline gap-3 text-chrome leading-[1.55]">
            <span
              className={`w-13.5 flex-none font-mono text-micro tracking-wide uppercase ${
                TONES[change.kind] ?? "text-muted-foreground"
              }`}
            >
              {change.kind}
            </span>
            <span className="text-secondary-foreground">{change.text}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
