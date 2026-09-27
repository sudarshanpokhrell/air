import { meQuery } from "@/hooks/use-auth"
import { createFileRoute, Outlet, redirect } from "@tanstack/react-router"

// Pages for signed-out users; signed-in users are sent to the app.
export const Route = createFileRoute("/_auth")({
  beforeLoad: async ({ context }) => {
    const user = await context.queryClient.ensureQueryData(meQuery)
    if (user) throw redirect({ to: "/", replace: true })
  },
  component: () => <Outlet />,
})
