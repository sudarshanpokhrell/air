import { StrictMode } from "react"
import ReactDOM from "react-dom/client"
import { RouterProvider, createRouter } from "@tanstack/react-router"
import { QueryClientProvider } from "@tanstack/react-query"

import { routeTree } from "./routeTree.gen"
import { queryClient } from "./lib/query-client"
import { onUnauthorized } from "./lib/api"
import { meQuery } from "./hooks/use-auth"
import "./index.css"

const router = createRouter({
  routeTree,
  context: { queryClient },
  defaultPreload: "intent",
  defaultPreloadStaleTime: 0,
})

// Session expired or revoked (logout elsewhere, password change, deactivation):
// go to login and come back here afterwards.
onUnauthorized(async () => {
  if (queryClient.getQueryData(meQuery.queryKey) === null) return
  queryClient.setQueryData(meQuery.queryKey, null)
  await router.navigate({
    to: "/login",
    search: { redirect: router.state.location.href },
    replace: true,
  })
  queryClient.removeQueries({ predicate: (q) => q.queryKey[0] !== "me" })
})

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router
  }
}

const rootElement = document.getElementById("root")!
if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement)
  root.render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </StrictMode>
  )
}
