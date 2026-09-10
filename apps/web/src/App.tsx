import { useQuery } from "@tanstack/react-query"
import { createClient } from "@connectrpc/connect"
import { InstanceService } from "@nooks/api"

import { Mark } from "./components/mark"
import { transport } from "./lib/transport"
import { useTheme } from "./lib/use-theme"
import type { Theme } from "./lib/theme"

const instanceClient = createClient(InstanceService, transport)

const THEMES: Theme[] = ["light", "dark", "system"]

/**
 * The shell, for now: it proves the whole stack end to end — proto to Go to Connect to
 * TanStack Query to a page rendered from Nooks' tokens. The real screens land in M2.
 */
export function App() {
  const { theme, setTheme } = useTheme()

  const instance = useQuery({
    queryKey: ["instance"],
    queryFn: () => instanceClient.getInstance({}),
  })

  return (
    <div className="min-h-dvh bg-desk">
      <header className="flex h-[52px] items-center gap-3 border-b border-hair bg-background px-6">
        <Mark size={20} items="var(--control)" />
        <span className="text-[14px] font-medium tracking-[-0.02em]">nooks</span>
        <span className="text-muted-foreground text-[12.5px]">first run</span>
        <div className="flex-1" />
        <ThemeSegment value={theme} onChange={setTheme} />
      </header>

      <main className="mx-auto max-w-[660px] px-6 pt-14">
        <h1 className="text-[33px] leading-[1.1] font-semibold tracking-[-0.025em]">
          {instance.data?.name || "Nooks is running."}
        </h1>

        <div className="border-hair mt-3.5 flex items-center gap-3 border-b pb-4 text-[13.5px]">
          <StatusLine query={instance} />
        </div>

        <p className="text-secondary-foreground mt-8 max-w-[46ch] text-[15px] leading-[1.65]">
          Make yourself an account and name the instance. Everything else can wait until someone asks
          for it.
        </p>
      </main>
    </div>
  )
}

type StatusLineProps = {
  query: ReturnType<typeof useQuery<Awaited<ReturnType<typeof instanceClient.getInstance>>>>
}

/** One line of orientation under the title, in whichever state the request is in. */
function StatusLine({ query }: StatusLineProps) {
  if (query.isPending) {
    return <span className="bg-muted h-[13px] w-[240px] rounded-sm" />
  }
  if (query.isError) {
    return (
      <span className="text-destructive">
        Can&rsquo;t reach the server. {query.error.message}
      </span>
    )
  }
  return (
    <>
      <span className="text-muted-foreground">
        {query.data.needsSetup ? "No account yet" : "Ready"}
      </span>
      <span className="bg-border h-3 w-px" />
      <span className="text-muted-foreground font-mono text-[12.5px]">{query.data.version}</span>
      {query.data.publicSignup ? (
        <>
          <span className="bg-border h-3 w-px" />
          <span className="bg-shared-bg text-shared rounded-sm px-1.5 py-0.5 text-[11.5px]">
            public signup
          </span>
        </>
      ) : null}
    </>
  )
}

type ThemeSegmentProps = {
  value: Theme
  onChange: (theme: Theme) => void
}

/** The three-segment control from DESIGN.md §2, in its smallest useful form. */
function ThemeSegment({ value, onChange }: ThemeSegmentProps) {
  return (
    <div className="border-border flex overflow-hidden rounded-lg border">
      {THEMES.map((option, index) => (
        <button
          key={option}
          type="button"
          onClick={() => onChange(option)}
          className={[
            "h-8 px-3.5 text-[13.5px] capitalize",
            index === 0 ? "" : "border-border border-l",
            option === value
              ? "bg-secondary font-medium"
              : "text-secondary-foreground hover:bg-accent",
          ].join(" ")}
        >
          {option}
        </button>
      ))}
    </div>
  )
}
