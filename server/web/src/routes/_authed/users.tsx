import { ErrorBanner, Page, TableCard, Th, theadClass } from "@/components/page"
import { StatusDot, Tag } from "@/components/tags"
import { CreateUserDialog } from "@/components/users/create-user-dialog"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Skeleton } from "@/components/ui/skeleton"
import { useUser } from "@/hooks/use-auth"
import { usersQuery, useUpdateUser } from "@/hooks/use-users"
import { getErrorMessage } from "@/lib/api"
import { cn } from "@/lib/utils"
import type { UpdateUserInput, User } from "@/types/auth"
import { MoreHorizontalIcon } from "@hugeicons/core-free-icons"
import { HugeiconsIcon } from "@hugeicons/react"
import { useQuery } from "@tanstack/react-query"
import { createFileRoute, redirect } from "@tanstack/react-router"
import { format } from "date-fns"
import { toast } from "sonner"

export const Route = createFileRoute("/_authed/users")({
  beforeLoad: ({ context }) => {
    if (!context.user.is_admin) throw redirect({ to: "/", replace: true })
  },
  component: UsersPage,
})

function UsersPage() {
  const me = useUser()!
  const { data: users, isPending, error } = useQuery(usersQuery)
  const updateUser = useUpdateUser()

  const update = async (input: UpdateUserInput, done: string) => {
    try {
      await updateUser.mutateAsync(input)
      toast.success(done)
    } catch (e) {
      toast.error(getErrorMessage(e))
    }
  }

  return (
    <Page
      title="Users"
      description="Everyone who can sign in. Give them access to an app from that app's Members tab."
      actions={<CreateUserDialog />}
    >
      {error ? (
        <ErrorBanner error={error} />
      ) : (
        <TableCard>
          <thead className={theadClass}>
            <tr>
              <Th>User</Th>
              <Th>Role</Th>
              <Th>Status</Th>
              <Th>Joined</Th>
              <Th>
                <span className="sr-only">Actions</span>
              </Th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {isPending &&
              Array.from({ length: 3 }).map((_, i) => (
                <tr key={i}>
                  <td className="px-4 py-3" colSpan={5}>
                    <Skeleton className="h-8 w-full" />
                  </td>
                </tr>
              ))}
            {users?.map((user) => (
              <UserRow
                key={user.id}
                user={user}
                isSelf={user.id === me.id}
                onUpdate={update}
              />
            ))}
          </tbody>
        </TableCard>
      )}
    </Page>
  )
}

function UserRow({
  user,
  isSelf,
  onUpdate,
}: {
  user: User
  isSelf: boolean
  onUpdate: (input: UpdateUserInput, done: string) => void
}) {
  return (
    <tr
      className={cn(
        "hover:bg-surface-2",
        !user.is_active && "text-muted-foreground"
      )}
    >
      <td className="px-4 py-3">
        <div className="flex items-center gap-3">
          <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-surface-3 text-sm font-semibold">
            {user.name.charAt(0).toUpperCase()}
          </span>
          <div>
            <div className="font-medium">
              {user.name}
              {isSelf && (
                <span className="ml-1.5 font-normal text-muted-foreground">
                  (you)
                </span>
              )}
            </div>
            <div className="text-caption text-muted-foreground">
              {user.email}
            </div>
          </div>
        </div>
      </td>
      <td className="px-4 py-3">
        {user.is_admin ? (
          <Tag tone="strong">Global admin</Tag>
        ) : (
          <Tag>Member</Tag>
        )}
      </td>
      <td className="px-4 py-3">
        {user.is_active ? (
          <StatusDot tone="ok">Active</StatusDot>
        ) : (
          <StatusDot tone="off">Deactivated</StatusDot>
        )}
      </td>
      <td className="px-4 py-3 text-muted-foreground">
        {format(new Date(user.created_at), "MMM d, yyyy")}
      </td>
      <td className="px-4 py-3 text-right">
        {/* You can't demote or deactivate yourself (the API refuses too). */}
        {!isSelf && (
          <DropdownMenu>
            <DropdownMenuTrigger
              render={<Button size="icon-sm" variant="ghost" />}
              aria-label="User actions"
            >
              <HugeiconsIcon icon={MoreHorizontalIcon} />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onClick={() =>
                  onUpdate(
                    { id: user.id, is_admin: !user.is_admin },
                    user.is_admin
                      ? `${user.name} is no longer a global admin.`
                      : `${user.name} is now a global admin.`
                  )
                }
              >
                {user.is_admin ? "Remove global admin" : "Make global admin"}
              </DropdownMenuItem>
              <DropdownMenuItem
                variant={user.is_active ? "destructive" : "default"}
                onClick={() =>
                  onUpdate(
                    { id: user.id, is_active: !user.is_active },
                    user.is_active
                      ? `${user.name} deactivated and signed out.`
                      : `${user.name} reactivated.`
                  )
                }
              >
                {user.is_active ? "Deactivate" : "Reactivate"}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </td>
    </tr>
  )
}
