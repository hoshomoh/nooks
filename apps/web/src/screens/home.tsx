import { getRouteApi } from "@tanstack/react-router"

// getRouteApi rather than importing the route object: the route imports this component,
// so importing it back would be a cycle.
const route = getRouteApi("/")

export function Home() {
  const { instance, member } = route.useLoaderData()

  return (
    <div className="bg-desk min-h-dvh">
      <main className="mx-auto max-w-[660px] px-6 pt-14">
        <h1 className="text-[33px] leading-[1.1] font-semibold tracking-[-0.025em]">
          {instance.name}
        </h1>
        <div className="border-hair mt-3.5 flex items-center gap-3 border-b pb-4 text-[13.5px]">
          <span className="text-muted-foreground">Signed in as {member.name}</span>
        </div>
        <p className="text-secondary-foreground mt-8 max-w-[46ch] text-[15px] leading-[1.65]">
          Lists arrive in M3. Until then this is proof that first run, sign-in and the
          session all work end to end.
        </p>
      </main>
    </div>
  )
}
