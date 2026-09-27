import { Outlet, createRootRouteWithContext } from "@tanstack/react-router"
import { type QueryClient } from "@tanstack/react-query"
import { Toaster } from "sonner"
import { SidebarProvider } from "@/components/ui/sidebar"

export interface RouterContext {
  queryClient: QueryClient
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: RootComponent,
})

function RootComponent() {
  return (
    <SidebarProvider>
      <div className="flex min-h-screen w-full flex-col bg-background text-foreground antialiased">
        <Toaster position="bottom-right" theme="light" closeButton />
        <Outlet />
      </div>
    </SidebarProvider>
  )
}
