import { ErrorBanner } from "@/components/page"
import { Tag } from "@/components/tags"
import { Button } from "@/components/ui/button"
import { appQuery } from "@/hooks/use-apps"
import { isApiError } from "@/lib/api"
import { cn } from "@/lib/utils"
import { useSuspenseQuery } from "@tanstack/react-query"
import {
  createFileRoute,
  Link,
  notFound,
  Outlet,
  useRouterState,
} from "@tanstack/react-router"

// Layout for one app: title + tabs. Non-members get 404 from the API, so
// they see "not found" rather than learning the app exists.
export const Route = createFileRoute("/_authed/apps/$slug")({
  loader: async ({ context, params }) => {
    try {
      await context.queryClient.ensureQueryData(appQuery(params.slug))
    } catch (e) {
      if (isApiError(e, 404)) throw notFound()
      throw e
    }
  },
  component: AppLayout,
  notFoundComponent: AppNotFound,
  errorComponent: ({ error }) => (
    <div className="p-8">
      <ErrorBanner error={error} />
    </div>
  ),
})

const tabs = [
  { label: "Updates", to: "/apps/$slug" },
  { label: "Members", to: "/apps/$slug/members" },
  { label: "API keys", to: "/apps/$slug/api-keys" },
] as const

function AppLayout() {
  const { slug } = Route.useParams()
  const { data } = useSuspenseQuery(appQuery(slug))
  const pathname = useRouterState({ select: (s) => s.location.pathname })

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-6 px-8 py-8">
      <header className="space-y-4">
        <Link
          to="/apps"
          className="text-body-sm text-muted-foreground hover:text-foreground"
        >
          ← Apps
        </Link>
        <div className="flex items-center gap-2">
          <h1 className="text-xl font-semibold">{data.app.name}</h1>
          <Tag tone={data.my_role === "admin" ? "strong" : "default"}>
            {data.my_role === "admin" ? "Admin" : "Developer"}
          </Tag>
        </div>
        <p className="font-mono text-caption text-muted-foreground">
          ?app={data.app.slug}
        </p>

        <nav className="flex gap-5 border-b">
          {tabs.map((tab) => {
            const href = tab.to.replace("$slug", slug)
            const active = pathname === href || pathname === `${href}/`
            return (
              <Link
                key={tab.to}
                to={tab.to}
                params={{ slug }}
                className={cn(
                  "-mb-px border-b-2 pb-2 text-body-sm",
                  active
                    ? "border-foreground font-medium text-foreground"
                    : "border-transparent text-muted-foreground hover:text-foreground"
                )}
              >
                {tab.label}
              </Link>
            )
          })}
        </nav>
      </header>

      <Outlet />
    </div>
  )
}

function AppNotFound() {
  return (
    <div className="flex flex-col items-center gap-3 px-8 py-24 text-center">
      <h1 className="text-xl font-semibold">App not found</h1>
      <p className="text-body-sm text-muted-foreground">
        It doesn't exist, or you're not a member of it.
      </p>
      <Button render={<Link to="/apps" />} variant="outline">
        Back to apps
      </Button>
    </div>
  )
}
