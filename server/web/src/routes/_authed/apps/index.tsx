import { AppIcon } from "@/components/apps/app-icon"
import { CreateAppDialog } from "@/components/apps/create-app-dialog"
import { EmptyState, ErrorBanner, Page } from "@/components/page"
import { Skeleton } from "@/components/ui/skeleton"
import { useUser } from "@/hooks/use-auth"
import { appsQuery } from "@/hooks/use-apps"
import { useQuery } from "@tanstack/react-query"
import { createFileRoute, Link } from "@tanstack/react-router"
import { format } from "date-fns"

export const Route = createFileRoute("/_authed/apps/")({
  component: AppsPage,
})

function AppsPage() {
  const user = useUser()!
  const { data: apps, isPending, error } = useQuery(appsQuery)

  return (
    <Page
      title="Apps"
      description={
        user.is_admin
          ? "Every app on this server."
          : "The apps you're a member of."
      }
      actions={user.is_admin && <CreateAppDialog />}
    >
      {error ? (
        <ErrorBanner error={error} />
      ) : isPending ? (
        <div className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-14" />
          ))}
        </div>
      ) : apps.length === 0 ? (
        <EmptyState>
          {user.is_admin
            ? "No apps yet. Create one to start publishing updates."
            : "You're not a member of any app yet. Ask an admin to add you."}
        </EmptyState>
      ) : (
        <ul className="divide-y rounded-lg border">
          {apps.map((app) => (
            <li key={app.id}>
              <Link
                to="/apps/$slug"
                params={{ slug: app.slug }}
                className="flex items-center gap-3 px-4 py-3 hover:bg-surface-2"
              >
                <AppIcon name={app.name} className="size-8" />
                <div className="min-w-0 flex-1">
                  <div className="truncate font-medium">{app.name}</div>
                  <div className="truncate font-mono text-caption text-muted-foreground">
                    {app.slug}
                  </div>
                </div>
                <div className="text-caption text-muted-foreground">
                  {format(new Date(app.created_at), "MMM d, yyyy")}
                </div>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </Page>
  )
}
