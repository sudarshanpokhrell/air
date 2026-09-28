import {
  EmptyState,
  ErrorBanner,
  TableCard,
  Th,
  theadClass,
} from "@/components/page"
import { Tag } from "@/components/tags"
import { Skeleton } from "@/components/ui/skeleton"
import { updatesQuery } from "@/hooks/use-apps"
import { isApiError } from "@/lib/api"
import { useQuery } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { format } from "date-fns"

export const Route = createFileRoute("/_authed/apps/$slug/")({
  component: UpdatesPage,
})

function UpdatesPage() {
  const { slug } = Route.useParams()
  const { data: updates, isPending, error } = useQuery(updatesQuery(slug))

  if (error) {
    // Publishing isn't implemented on the server yet (501).
    return isApiError(error, 501) ? (
      <EmptyState>
        Update history will appear here once publishing is available.
      </EmptyState>
    ) : (
      <ErrorBanner error={error} />
    )
  }

  if (!isPending && updates.length === 0) {
    return (
      <EmptyState>
        No updates yet. Publish one with the CLI and it will show up here.
      </EmptyState>
    )
  }

  return (
    <TableCard>
      <thead className={theadClass}>
        <tr>
          <Th>Update</Th>
          <Th>Platform</Th>
          <Th>Channel</Th>
          <Th>Runtime</Th>
          <Th>Published</Th>
        </tr>
      </thead>
      <tbody className="divide-y">
        {isPending &&
          Array.from({ length: 3 }).map((_, i) => (
            <tr key={i}>
              <td className="px-4 py-3" colSpan={5}>
                <Skeleton className="h-5 w-full" />
              </td>
            </tr>
          ))}
        {updates?.map((u) => (
          <tr key={u.id}>
            <td className="px-4 py-3">
              {u.kind === "rollback_to_embedded" ? (
                <Tag tone="danger">Rollback to embedded</Tag>
              ) : (
                <span className="font-medium">
                  {u.message || "Untitled update"}
                </span>
              )}
              <div className="font-mono text-caption text-muted-foreground">
                {u.id.slice(0, 8)}
                {u.git_commit && ` · ${u.git_commit.slice(0, 7)}`}
              </div>
            </td>
            <td className="px-4 py-3">
              {u.platform === "ios" ? "iOS" : "Android"}
            </td>
            <td className="px-4 py-3">{u.channel}</td>
            <td className="px-4 py-3 font-mono text-caption">
              {u.runtime_version}
            </td>
            <td className="px-4 py-3 text-muted-foreground">
              {format(new Date(u.created_at), "MMM d, HH:mm")}
            </td>
          </tr>
        ))}
      </tbody>
    </TableCard>
  )
}
