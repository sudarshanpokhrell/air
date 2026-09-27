import { ChangePasswordForm } from "@/components/change-password-form"
import { Page } from "@/components/page"
import { Tag } from "@/components/tags"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useUpdateMe, useUser } from "@/hooks/use-auth"
import { getErrorMessage } from "@/lib/api"
import { createFileRoute } from "@tanstack/react-router"
import { useForm } from "react-hook-form"
import { toast } from "sonner"

export const Route = createFileRoute("/_authed/profile")({
  component: ProfilePage,
})

function Section({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) {
  return (
    <section className="space-y-4 rounded-lg border p-5">
      <h2 className="font-medium">{title}</h2>
      {children}
    </section>
  )
}

function ProfilePage() {
  const user = useUser()!
  const updateMe = useUpdateMe()

  const {
    register,
    handleSubmit,
    setError,
    reset,
    formState: { errors, isSubmitting, isDirty },
  } = useForm<{ name: string }>({ values: { name: user.name } })

  const onSubmit = async (values: { name: string }) => {
    try {
      const { user: updated } = await updateMe.mutateAsync({
        name: values.name.trim(),
      })
      reset({ name: updated.name })
      toast.success("Profile updated.")
    } catch (error) {
      setError("name", { message: getErrorMessage(error) })
    }
  }

  return (
    <Page title="Profile" className="max-w-2xl">
      <Section title="Account">
        <dl className="grid grid-cols-[6rem_1fr] items-center gap-y-2 text-body-sm">
          <dt className="text-muted-foreground">Email</dt>
          <dd>{user.email}</dd>
          <dt className="text-muted-foreground">Role</dt>
          <dd>
            {user.is_admin ? (
              <Tag tone="strong">Global admin</Tag>
            ) : (
              <Tag>Member</Tag>
            )}
          </dd>
        </dl>

        <form
          className="flex items-end gap-2"
          noValidate
          onSubmit={handleSubmit(onSubmit)}
        >
          <div className="flex-1 space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              aria-invalid={errors.name ? true : undefined}
              autoComplete="name"
              className="h-9"
              id="name"
              {...register("name", {
                required: "Name is required.",
                validate: (v) => v.trim() !== "" || "Name is required.",
              })}
            />
          </div>
          <Button
            className="h-9"
            disabled={!isDirty || isSubmitting}
            type="submit"
          >
            {isSubmitting ? "Saving…" : "Save"}
          </Button>
        </form>
        {errors.name && (
          <p className="text-body-sm text-destructive">{errors.name.message}</p>
        )}
      </Section>

      <Section title="Password">
        <p className="text-body-sm text-muted-foreground">
          Changing it signs you out on every other device.
        </p>
        <ChangePasswordForm
          className="max-w-sm space-y-4"
          onSuccess={() => {
            toast.success("Password changed.")
          }}
        />
      </Section>
    </Page>
  )
}
