/**
 * What Nooks looks like, drawn rather than captured.
 *
 * A screenshot goes stale the first time a colour changes and has to be retaken by
 * somebody with the app running. This is built from the same tokens the app is, so it
 * cannot drift from it, and it follows the reader's theme the way a PNG never could.
 */
export function AppPreview() {
  return (
    <div className="overflow-hidden rounded-xl border border-border bg-background shadow-sm">
      <div className="flex h-8.5 items-center gap-1.5 border-b border-hair bg-sidebar px-3">
        <Dot />
        <Dot />
        <Dot />
        <span className="ml-2 font-mono text-[11.5px] text-muted-foreground">
          nooks.brunnen.house
        </span>
      </div>

      <div className="grid grid-cols-[132px_minmax(0,1fr)]">
        <aside className="flex flex-col gap-3.5 border-r border-hair bg-sidebar p-2.5">
          <span className="px-1 text-small font-medium">Brunnen St</span>
          <div className="flex flex-col gap-px">
            <span className="px-1 pb-1 text-label text-muted-foreground uppercase">Lists</span>
            <SidebarRow name="Groceries" active />
            <SidebarRow name="Weekend" shared />
            <SidebarRow name="Hardware" />
            <SidebarRow name="Packing" />
          </div>
        </aside>

        <div className="min-w-0 p-5">
          <div className="text-section tracking-tight">Groceries</div>
          <div className="mt-3 border-t border-foreground/15" />
          <PreviewItem label="Rye flour" quantity="1 kg" />
          <PreviewItem label="Oat milk" quantity="2" />
          <PreviewItem label="Tomatoes" quantity="Sat" />
          <PreviewItem label="Coffee beans" />
          <div className="flex items-center gap-2.5 py-2 text-small text-muted-foreground">
            <span className="size-[13px] rounded-[3px] border border-dashed border-border" />
            Add an item
          </div>
        </div>
      </div>
    </div>
  )
}

function Dot() {
  return <span className="size-2 rounded-full bg-border" />
}

function SidebarRow({ name, active, shared }: { name: string; active?: boolean; shared?: boolean }) {
  return (
    <span
      className={`flex items-center gap-1.5 rounded-md px-1.5 py-1 text-small ${
        active ? "bg-secondary font-medium text-foreground" : "text-secondary-foreground"
      }`}
    >
      {name}
      {shared && <span className="size-[5px] rounded-full bg-shared" />}
    </span>
  )
}

/** One row, at the proportions DESIGN.md §6 gives a real one. */
function PreviewItem({ label, quantity }: { label: string; quantity?: string }) {
  return (
    <div className="flex items-center gap-2.5 border-b border-hair py-2 text-small">
      <span className="size-[13px] flex-none rounded-[3px] border border-border" />
      <span className="min-w-0 flex-1 truncate">{label}</span>
      {quantity && <span className="font-mono text-[11px] text-muted-foreground">{quantity}</span>}
    </div>
  )
}
