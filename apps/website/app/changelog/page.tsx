import type { Metadata } from "next"

import { releases, type Release } from "@/lib/changelog"
import { SOURCE } from "@nooks/shared"

export const metadata: Metadata = {
  title: "Changelog — nooks",
  description: "Every released version of nooks, and what changed in it.",
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

/**
 * How many releases the page carries.
 *
 * What somebody arrives here to answer is "what changed since the version I am on",
 * which is near the top, and a page nobody scrolls is a page nobody reads. Everything
 * further back is on the releases page, which is already paginated and already the
 * canonical record.
 *
 * Ten rather than six because LINES below bounds each entry: a release contributes at
 * most five lines however many commits went into it, so the length of this page is now
 * something this file decides rather than something the size of a release decides. It
 * was six when one release could be a hundred and twenty-two lines on its own.
 */
const SHOWN = 10

/**
 * How many lines of one release the page draws.
 *
 * A release gathers however many commits went into it, and a big one can carry sixty.
 * Drawn in full they push every release under them off the screen, so the page stops
 * answering the question it exists for: what changed since the version I am on. Five is
 * enough to see the shape of a release and decide whether to read the rest, which is on
 * GitHub with the binaries beside it.
 */
const LINES = 5

export default async function Changelog() {
  const published = await releases()
  const recent = published.slice(0, SHOWN)
  const older = published.length - recent.length

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
            nooks has not had a release yet. The first one will appear here.
          </p>
        ) : (
          recent.map((release) => <Entry key={release.version} release={release} />)
        )}
      </div>

      {older > 0 && (
        <section className="mt-10 flex flex-wrap items-center gap-6 rounded-xl border border-border bg-sidebar px-8 py-9">
          <div className="flex min-w-0 flex-1 flex-col gap-1.5">
            <h2 className="text-section">
              {older} earlier {older === 1 ? "release" : "releases"}
            </h2>
            <p className="text-field text-secondary-foreground">
              This page keeps the most recent {SHOWN}. Every release nooks has ever made is
              on GitHub, with the binaries and checksums attached to each one.
            </p>
          </div>
          <a
            href={`${SOURCE}/releases`}
            className="flex h-8.5 items-center rounded-md bg-primary px-4 text-chrome font-medium text-primary-foreground"
          >
            All releases
          </a>
        </section>
      )}
    </main>
  )
}

/**
 * One release: what it was called and when, then what it did.
 *
 * The version sits in a column of its own so a reader scanning for the one they are on
 * reads down a single line rather than across every entry.
 */
interface EntryProps {
  release: Release
}

function Entry({ release }: EntryProps) {
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
        {release.changes.slice(0, LINES).map((change, index) => (
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

        {/* Sits in the same column as the lines it continues, so it reads as the end of
            the list rather than as a second thing. The release's own tag page, not the
            index: somebody following this wants the rest of this release. */}
        {release.changes.length > LINES && (
          <a
            href={`${SOURCE}/releases/tag/${release.version}`}
            className="self-start text-chrome text-muted-foreground underline underline-offset-3 hover:text-foreground"
          >
            {release.changes.length - LINES} more in {release.version}
          </a>
        )}
      </div>
    </div>
  )
}
