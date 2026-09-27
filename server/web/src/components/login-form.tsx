import { useRouter, useSearch } from "@tanstack/react-router"
import { useForm } from "react-hook-form"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { PasswordInput } from "@/components/ui/password-input"
import { useLogin } from "@/hooks/use-auth"
import { getErrorMessage } from "@/lib/api"

export function safeRedirect(redirect: string | undefined) {
  return redirect?.startsWith("/") && !redirect.startsWith("//")
    ? redirect
    : "/"
}

type LoginValues = {
  email: string
  password: string
}

export default function LoginForm() {
  const login = useLogin()
  const router = useRouter()
  const { redirect } = useSearch({ from: "/_auth/login" })

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginValues>({ defaultValues: { email: "", password: "" } })

  const onSubmit = async (values: LoginValues) => {
    try {
      await login.mutateAsync(values)
      // redirect is a full href (path + search), so navigate by href.
      await router.navigate({ href: safeRedirect(redirect), replace: true })
    } catch (error) {
      setError("root", { message: getErrorMessage(error) })
    }
  }

  return (
    <div className="flex min-h-dvh items-center justify-center bg-surface-2 p-4">
      <form
        className="w-full max-w-sm space-y-5 rounded-lg border bg-card p-6"
        noValidate
        onSubmit={handleSubmit(onSubmit)}
      >
        <div className="space-y-1">
          <h1 className="text-lg font-semibold">Sign in to AIR</h1>
          <p className="text-body-sm text-muted-foreground">
            Manage your apps and OTA updates.
          </p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="email">Email</Label>
          <Input
            aria-invalid={errors.email ? true : undefined}
            autoComplete="email"
            autoFocus
            className="h-9"
            id="email"
            type="email"
            {...register("email", {
              required: "Email is required.",
              pattern: {
                value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
                message: "Enter a valid email address.",
              },
            })}
          />
          {errors.email && (
            <p className="text-body-sm text-destructive">
              {errors.email.message}
            </p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="password">Password</Label>
          <PasswordInput
            aria-invalid={errors.password ? true : undefined}
            autoComplete="current-password"
            className="h-9"
            id="password"
            {...register("password", { required: "Password is required." })}
          />
          {errors.password && (
            <p className="text-body-sm text-destructive">
              {errors.password.message}
            </p>
          )}
        </div>

        {errors.root && (
          <p className="text-body-sm text-destructive" role="alert">
            {errors.root.message}
          </p>
        )}

        <Button className="h-9 w-full" disabled={isSubmitting} type="submit">
          {isSubmitting ? "Signing in…" : "Sign in"}
        </Button>

        <p className="text-caption text-muted-foreground">
          Forgot your password? Ask your administrator.
        </p>
      </form>
    </div>
  )
}
