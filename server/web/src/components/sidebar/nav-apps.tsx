import { Link, useRouterState } from "@tanstack/react-router"
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"
import { navButtonClass } from "./nav-link"
import type { App } from "@/types/apps"

/** Shortcut list of the apps the user can open. */
export function NavApps({ apps }: { apps: App[] }) {
  const pathname = useRouterState({ select: (s) => s.location.pathname })

  if (apps.length === 0) return null

  return (
    <SidebarGroup>
      <SidebarGroupLabel className="px-2.5 text-caption">
        Apps
      </SidebarGroupLabel>
      <SidebarMenu>
        {apps.map((app) => (
          <SidebarMenuItem key={app.id}>
            <SidebarMenuButton
              tooltip={app.name}
              isActive={pathname.startsWith(`/apps/${app.slug}`)}
              className={navButtonClass}
              render={<Link to="/apps/$slug" params={{ slug: app.slug }} />}
            >
              <span className="truncate">{app.name}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  )
}
