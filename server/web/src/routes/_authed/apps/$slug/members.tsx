import { AddMemberDialog } from "@/components/apps/add-member-dialog"
import { DeleteDialog } from "@/components/delete-dialog"
import { ErrorBanner, TableCard, Th, theadClass } from "@/components/page"
import { Tag } from "@/components/tags"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { useUser } from "@/hooks/use-auth"
import {
  appQuery,
  membersQuery,
  useRemoveMember,
  useUpdateMember,
} from "@/hooks/use-apps"
import { getErrorMessage } from "@/lib/api"
import type { AppMember, AppRole } from "@/types/apps"
import { useQuery, useSuspenseQuery } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { format } from "date-fns"
import { useState } from "react"
import { toast } from "sonner"

export const Route = createFileRoute("/_authed/apps/$slug/members")({
  component: MembersPage,
})

const roleLabel = { admin: "Admin", developer: "Developer" } as const

function MembersPage() {
  const { slug } = Route.useParams()
  const me = useUser()!
  const { data: appData } = useSuspenseQuery(appQuery(slug))
  const { data: members, isPending, error } = useQuery(membersQuery(slug))
  const updateMember = useUpdateMember(slug)
  const removeMember = useRemoveMember(slug)
  const [removing, setRemoving] = useState<AppMember | null>(null)

  const canManage = appData.my_role === "admin"

  const changeRole = async (member: AppMember, role: AppRole) => {
    try {
      await updateMember.mutateAsync({ userId: member.user_id, role })
      toast.success(`${member.name} is now ${roleLabel[role].toLowerCase()}.`)
    } catch (e) {
      toast.error(getErrorMessage(e))
    }
  }

  const remove = async () => {
    if (!removing) return
    try {
      await removeMember.mutateAsync(removing.user_id)
      toast.success(`${removing.name} removed.`)
      setRemoving(null)
    } catch (e) {
      toast.error(getErrorMessage(e))
    }
  }

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <p className="max-w-lg text-body-sm text-muted-foreground">
          Developers publish and roll back. Admins also manage members,
          platforms and keys. Global admins have access without being listed.
        </p>
        {canManage && <AddMemberDialog slug={slug} />}
      </div>

      {error ? (
        <ErrorBanner error={error} />
      ) : (
        <TableCard>
          <thead className={theadClass}>
            <tr>
              <Th>Member</Th>
              <Th>Role</Th>
              <Th>Added</Th>
              {canManage && (
                <Th>
                  <span className="sr-only">Actions</span>
                </Th>
              )}
            </tr>
          </thead>
          <tbody className="divide-y">
            {isPending &&
              Array.from({ length: 2 }).map((_, i) => (
                <tr key={i}>
                  <td className="px-4 py-3" colSpan={4}>
                    <Skeleton className="h-8 w-full" />
                  </td>
                </tr>
              ))}
            {members?.map((m) => (
              <tr key={m.user_id} className="hover:bg-surface-2">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-3">
                    <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-surface-3 text-sm font-semibold">
                      {m.name.charAt(0).toUpperCase()}
                    </span>
                    <div>
                      <div className="font-medium">
                        {m.name}
                        {m.user_id === me.id && (
                          <span className="ml-1.5 font-normal text-muted-foreground">
                            (you)
                          </span>
                        )}
                      </div>
                      <div className="text-caption text-muted-foreground">
                        {m.email}
                      </div>
                    </div>
                  </div>
                </td>
                <td className="px-4 py-3">
                  {canManage ? (
                    <Select
                      value={m.role}
                      onValueChange={(v) =>
                        v && v !== m.role && changeRole(m, v as AppRole)
                      }
                    >
                      <SelectTrigger className="h-8 w-32">
                        <SelectValue>
                          {(v: AppRole) => roleLabel[v]}
                        </SelectValue>
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="developer">Developer</SelectItem>
                        <SelectItem value="admin">Admin</SelectItem>
                      </SelectContent>
                    </Select>
                  ) : (
                    <Tag tone={m.role === "admin" ? "strong" : "default"}>
                      {roleLabel[m.role]}
                    </Tag>
                  )}
                </td>
                <td className="px-4 py-3 text-muted-foreground">
                  {format(new Date(m.created_at), "MMM d, yyyy")}
                </td>
                {canManage && (
                  <td className="px-4 py-3 text-right">
                    <Button
                      size="sm"
                      variant="ghost"
                      className="text-muted-foreground hover:text-destructive"
                      onClick={() => setRemoving(m)}
                    >
                      Remove
                    </Button>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </TableCard>
      )}

      <DeleteDialog
        open={removing !== null}
        onOpenChange={(open) => !open && setRemoving(null)}
        title={`Remove ${removing?.name} from this app?`}
        description="They lose access immediately, and their API keys for this app stop working."
        confirmLabel="Remove"
        pending={removeMember.isPending}
        onConfirm={remove}
      />
    </div>
  )
}
