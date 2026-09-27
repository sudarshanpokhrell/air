import { Loading03Icon } from "@hugeicons/core-free-icons"
import { HugeiconsIcon } from "@hugeicons/react"
import { useForm } from "react-hook-form"
import { Button } from "@/components/ui/button"
import { PasswordInput } from "@/components/ui/password-input"
import { Label } from "@/components/ui/label"
import { useChangePassword } from "@/hooks/use-auth"
import { getErrorMessage, isApiError } from "@/lib/api"

type ChangePasswordValues = {
  current_password: string
  new_password: string
  confirm_password: string
}

export function ChangePasswordForm({
  onSuccess,
  submitLabel = "Change password",
  className,
}: {
  onSuccess?: () => void | Promise<void>
  submitLabel?: string
  className?: string
}) {
  const changePassword = useChangePassword()

  const {
    register,
    handleSubmit,
    setError,
    reset,
    getValues,
    formState: { errors, isSubmitting },
  } = useForm<ChangePasswordValues>({
    defaultValues: {
      current_password: "",
      new_password: "",
      confirm_password: "",
    },
  })

  const onSubmit = async ({
    current_password,
    new_password,
  }: ChangePasswordValues) => {
    try {
      await changePassword.mutateAsync({ current_password, new_password })
      reset()
      await onSuccess?.()
    } catch (error) {
      if (isApiError(error, 422) && error.fields) {
        for (const [field, message] of Object.entries(error.fields)) {
          if (field === "current_password" || field === "new_password") {
            setError(field, {
              message: `${field === "current_password" ? "Current" : "New"} password ${message}`,
            })
          } else {
            setError("root", { message })
          }
        }
        return
      }
      setError("root", { message: getErrorMessage(error) })
    }
  }

  return (
    <form
      className={className ?? "space-y-5"}
      noValidate
      onSubmit={handleSubmit(onSubmit)}
    >
      <div className="space-y-4">
        <div>
          <Label htmlFor="current_password">Current password</Label>
          <div className="mt-2">
            <PasswordInput
              aria-invalid={errors.current_password ? true : undefined}
              autoComplete="current-password"
              className="h-10"
              id="current_password"
              {...register("current_password", {
                required: "Current password is required.",
              })}
            />
          </div>
          {errors.current_password && (
            <p className="mt-2 text-sm text-destructive">
              {errors.current_password.message}
            </p>
          )}
        </div>

        <div>
          <Label htmlFor="new_password">New password</Label>
          <div className="mt-2">
            <PasswordInput
              aria-invalid={errors.new_password ? true : undefined}
              autoComplete="new-password"
              className="h-10"
              id="new_password"
              {...register("new_password", {
                required: "New password is required.",
                minLength: {
                  value: 8,
                  message: "Password must be at least 8 characters.",
                },
                validate: (value) =>
                  value !== getValues("current_password") ||
                  "New password must differ from the current one.",
              })}
            />
          </div>
          {errors.new_password && (
            <p className="mt-2 text-sm text-destructive">
              {errors.new_password.message}
            </p>
          )}
        </div>

        <div>
          <Label htmlFor="confirm_password">Confirm new password</Label>
          <div className="mt-2">
            <PasswordInput
              aria-invalid={errors.confirm_password ? true : undefined}
              autoComplete="new-password"
              className="h-10"
              id="confirm_password"
              {...register("confirm_password", {
                validate: (value) =>
                  value === getValues("new_password") ||
                  "Passwords do not match.",
              })}
            />
          </div>
          {errors.confirm_password && (
            <p className="mt-2 text-sm text-destructive">
              {errors.confirm_password.message}
            </p>
          )}
        </div>
      </div>

      {errors.root && (
        <p
          className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive"
          role="alert"
        >
          {errors.root.message}
        </p>
      )}

      <Button className="h-10 w-full" disabled={isSubmitting} type="submit">
        {isSubmitting && (
          <HugeiconsIcon
            icon={Loading03Icon}
            className="h-4 w-4 animate-spin"
          />
        )}
        {submitLabel}
      </Button>
    </form>
  )
}
