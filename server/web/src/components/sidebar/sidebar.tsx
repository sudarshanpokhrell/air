import { useQuery } from "@tanstack/react-query"
import { AppStoreIcon, UserMultipleIcon } from "@hugeicons/core-free-icons"
import { Sidebar, SidebarContent, SidebarHeader } from "@/components/ui/sidebar"
import { useUser } from "@/hooks/use-auth"
import { appsQuery } from "@/hooks/use-apps"
import { NavApps } from "./nav-apps"
import { NavFooter } from "./nav-footer"
import { NavMain } from "./nav-main"
import type { NavItem } from "./types"

const navMain: NavItem[] = [
  { id: "apps", title: "Apps", url: "/apps", icon: AppStoreIcon },
  { id: "users", title: "Users", url: "/users", icon: UserMultipleIcon },
]

// Only global admins manage users.
const ADMIN_NAV_IDS = new Set(["users"])

export function AppSidebar(props: React.ComponentProps<typeof Sidebar>) {
  const user = useUser()
  const { data: apps } = useQuery(appsQuery)

  const items = navMain.filter(
    (item) => user?.is_admin || !ADMIN_NAV_IDS.has(item.id)
  )

  return (
    <Sidebar {...props}>
      <SidebarHeader className="px-4 py-4">
        <span className="font-semibold">AIR</span>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={items} />
        <NavApps apps={apps ?? []} />
      </SidebarContent>
      <NavFooter />
    </Sidebar>
  )
}
