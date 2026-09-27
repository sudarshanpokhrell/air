import LoginForm from "@/components/login-form"
import { createFileRoute } from "@tanstack/react-router"

type LoginSearch = {
  redirect?: string
}

export const Route = createFileRoute("/_auth/login")({
  validateSearch: (search: Record<string, unknown>): LoginSearch => ({
    redirect: typeof search.redirect === "string" ? search.redirect : undefined,
  }),
  component: LoginForm,
})
