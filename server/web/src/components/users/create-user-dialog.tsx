import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { useCreateUser } from "@/hooks/use-users"
import { getErrorMessage, isApiError } from "@/lib/api"
import type { CreateUserInput } from "@/types/auth"
import { Add01Icon, Loading03Icon } from "@hugeicons/core-free-icons"
import { HugeiconsIcon } from "@hugeicons/react"
import { useState } from "react"
import { Controller, useForm } from "react-hook-form"
import { toast } from "sonner"

export function CreateUserDialog() {
  const [open, setOpen] = useState(false)
  const createUser = useCreateUser()

  const {
    register,
    control,
    handleSubmit,
    reset,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<CreateUserInput>({
    defaultValues: { name: "", email: "", password: "", is_admin: false },
  })

  const onOpenChange = (next: boolean) => {
    setOpen(next)
    if (!next) reset()
  }

  const onSubmit = async (values: CreateUserInput) => {
    try {
      const { user } = await createUser.mutateAsync({
        ...values,
        name: values.name.trim(),
        email: values.email.trim(),
      })
      toast.success(`${user.name} added.`)
      onOpenChange(false)
    } catch (e) {
      if (isApiError(e, 422) && e.fields) {
        for (const [field, message] of Object.entries(e.fields)) {
          if (field === "name" || field === "email" || field === "password") {
            setError(field, { message })
          }
        }
        return
      }
      setError("root", { message: getErrorMessage(e) })
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger render={<Button />}>
        <HugeiconsIcon icon={Add01Icon} />
        Add user
      </DialogTrigger>
      <DialogContent className="mt-5 sm:max-w-md">
        <form
          className="space-y-5"
          noValidate
          onSubmit={handleSubmit(onSubmit)}
        >
          <DialogHeader>
            <DialogTitle>Add a user</DialogTitle>
          </DialogHeader>

          <div>
            <Label htmlFor="user-name">Name</Label>
            <Input
              id="user-name"
              className="mt-2 h-9"
              placeholder="Sujan Thapa"
              {...register("name", { required: "Name is required." })}
            />
            {errors.name && (
              <p className="mt-1.5 text-sm text-destructive">
                {errors.name.message}
              </p>
            )}
          </div>

          <div>
            <Label htmlFor="user-email">Email</Label>
            <Input
              id="user-email"
              type="email"
              className="mt-2 h-9"
              placeholder="dev@example.com"
              {...register("email", { required: "Email is required." })}
            />
            {errors.email && (
              <p className="mt-1.5 text-sm text-destructive">
                {errors.email.message}
              </p>
            )}
          </div>

          <div>
            <Label htmlFor="user-password">Temporary password</Label>
            <Input
              id="user-password"
              type="text"
              autoComplete="off"
              className="mt-2 h-9 font-mono"
              placeholder="at least 8 characters"
              {...register("password", {
                required: "Password is required.",
                minLength: {
                  value: 8,
                  message: "Password must be at least 8 characters.",
                },
              })}
            />
            {errors.password && (
              <p className="mt-1.5 text-sm text-destructive">
                {errors.password.message}
              </p>
            )}
          </div>

          <div className="flex items-center justify-between gap-4 rounded-lg border p-3">
            <div>
              <Label htmlFor="user-admin">Global admin</Label>
              <p className="text-xs text-muted-foreground">
                Manages users, creates apps and has full access to every app.
              </p>
            </div>
            <Controller
              control={control}
              name="is_admin"
              render={({ field }) => (
                <Switch
                  id="user-admin"
                  checked={field.value}
                  onCheckedChange={field.onChange}
                />
              )}
            />
          </div>

          {errors.root && (
            <p className="text-sm text-destructive" role="alert">
              {errors.root.message}
            </p>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting && (
                <HugeiconsIcon icon={Loading03Icon} className="animate-spin" />
              )}
              Add user
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
