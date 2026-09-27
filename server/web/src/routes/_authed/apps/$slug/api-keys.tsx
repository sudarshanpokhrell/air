import { CreateApiKeyDialog } from "@/components/apps/create-api-key-dialog"
import { DeleteDialog } from "@/components/delete-dialog"
import { ErrorBanner, TableCard, Th, theadClass } from "@/components/page"
import { StatusDot } from "@/components/tags"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useUser } from "@/hooks/use-auth"
import {
  apiKeysQuery,
  appQuery,
  membersQuery,
  useRevokeApiKey,
} from "@/hooks/use-apps"
import { getErrorMessage } from "@/lib/api"
import { cn } from "@/lib/utils"
import type { ApiKey } from "@/types/apps"
import { useQuery, useSuspenseQuery } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { format } from "date-fns"
import { useState } from "react"
import { toast } from "sonner"

export const Route = createFileRoute("/_authed/apps/$slug/api-keys")({
  component: ApiKeysPage,
})

function ApiKeysPage() {
  const { slug } = Route.useParams()
  const me = useUser()!
  const { data: appData } = useSuspenseQuery(appQuery(slug))
  const { data: keys, isPending, error } = useQuery(apiKeysQuery(slug))
  const { data: members } = useQuery(membersQuery(slug))
  const revokeKey = useRevokeApiKey(slug)
  const [revoking, setRevoking] = useState<ApiKey | null>(null)

  const isAdmin = appData.my_role === "admin"
  const creatorName = (userId: string) =>
    userId === me.id
      ? "You"
      : (members?.find((m) => m.user_id === userId)?.name ?? "—")

  const revoke = async () => {
    if (!revoking) return
    try {
      await revokeKey.mutateAsync(revoking.id)
      toast.success(`${revoking.name} revoked.`)
      setRevoking(null)
    } catch (e) {
      toast.error(getErrorMessage(e))
    }
  }

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <p className="max-w-lg text-body-sm text-muted-foreground">
          {isAdmin
            ? "Every API key for this app. A key publishes with its creator's access and stops working if they lose it."
            : "Your API keys for this app. The CLI uses them to publish updates."}
        </p>
        <CreateApiKeyDialog slug={slug} />
      </div>

      {error ? (
        <ErrorBanner error={error} />
      ) : (
        <TableCard>
          <thead className={theadClass}>
            <tr>
              <Th>Name</Th>
              {isAdmin && <Th>Created by</Th>}
              <Th>Created</Th>
              <Th>Status</Th>
              <Th>
                <span className="sr-only">Actions</span>
              </Th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {isPending &&
              Array.from({ length: 2 }).map((_, i) => (
                <tr key={i}>
                  <td className="px-4 py-3" colSpan={5}>
                    <Skeleton className="h-5 w-full" />
                  </td>
                </tr>
              ))}
            {!isPending && keys.length === 0 && (
              <tr>
                <td
                  className="px-4 py-10 text-center text-muted-foreground"
                  colSpan={5}
                >
                  No API keys yet. Create one to publish from the CLI or CI.
                </td>
              </tr>
            )}
            {keys?.map((k) => (
              <tr
                key={k.id}
                className={cn(
                  "hover:bg-surface-2",
                  k.revoked_at && "text-muted-foreground"
                )}
              >
                <td className="px-4 py-3 font-medium">{k.name}</td>
                {isAdmin && (
                  <td className="px-4 py-3">{creatorName(k.created_by)}</td>
                )}
                <td className="px-4 py-3 text-muted-foreground">
                  {format(new Date(k.created_at), "MMM d, yyyy")}
                </td>
                <td className="px-4 py-3">
                  {k.revoked_at ? (
                    <StatusDot tone="off">Revoked</StatusDot>
                  ) : (
                    <StatusDot tone="ok">Active</StatusDot>
                  )}
                </td>
                <td className="px-4 py-3 text-right">
                  {!k.revoked_at && (
                    <Button
                      size="sm"
                      variant="ghost"
                      className="text-muted-foreground hover:text-destructive"
                      onClick={() => setRevoking(k)}
                    >
                      Revoke
                    </Button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </TableCard>
      )}

      <DeleteDialog
        open={revoking !== null}
        onOpenChange={(open) => !open && setRevoking(null)}
        title={`Revoke ${revoking?.name}?`}
        description="Anything using this key (CI, scripts) can no longer publish. This can't be undone."
        confirmLabel="Revoke"
        pending={revokeKey.isPending}
        onConfirm={revoke}
      />
    </div>
  )
}
