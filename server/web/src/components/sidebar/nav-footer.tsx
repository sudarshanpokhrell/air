import { Link, useRouterState } from "@tanstack/react-router"
import {
  SidebarFooter,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"
import { useLogout, useUser } from "@/hooks/use-auth"
import { navButtonClass } from "./nav-link"

export function NavFooter() {
  const user = useUser()
  const logout = useLogout()
  const pathname = useRouterState({ select: (s) => s.location.pathname })

  return (
    <SidebarFooter className="border-t p-2">
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton
            isActive={pathname === "/profile"}
            className={`${navButtonClass} h-auto py-1.5`}
            render={<Link to="/profile" />}
            tooltip="Profile"
          >
            <span className="flex min-w-0 flex-col">
              <span className="truncate text-foreground">{user?.name}</span>
              <span className="truncate text-caption text-muted-foreground">
                {user?.email}
              </span>
            </span>
          </SidebarMenuButton>
        </SidebarMenuItem>
        <SidebarMenuItem>
          <SidebarMenuButton
            className={navButtonClass}
            disabled={logout.isPending}
            onClick={() => logout.mutate()}
          >
            Log out
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>
  )
}
