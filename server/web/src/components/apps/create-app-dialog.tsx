import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useCreateApp } from "@/hooks/use-apps"
import { getErrorMessage, isApiError } from "@/lib/api"
import type { CreateAppInput } from "@/types/apps"
import { Add01Icon, Loading03Icon } from "@hugeicons/core-free-icons"
import { HugeiconsIcon } from "@hugeicons/react"
import { useNavigate } from "@tanstack/react-router"
import { useState } from "react"
import { useForm } from "react-hook-form"
import { toast } from "sonner"

/** "My App" → "my-app" */
function slugify(name: string) {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 63)
}

export function CreateAppDialog() {
  const [open, setOpen] = useState(false)
  const createApp = useCreateApp()
  const navigate = useNavigate()

  const {
    register,
    handleSubmit,
    reset,
    setError,
    setValue,
    getFieldState,
    formState: { errors, isSubmitting },
  } = useForm<CreateAppInput>({ defaultValues: { name: "", slug: "" } })

  const onOpenChange = (next: boolean) => {
    setOpen(next)
    if (!next) reset()
  }

  const onSubmit = async (values: CreateAppInput) => {
    try {
      const { app } = await createApp.mutateAsync({
        name: values.name.trim(),
        slug: values.slug.trim(),
      })
      toast.success(`${app.name} created.`)
      onOpenChange(false)
      await navigate({ to: "/apps/$slug", params: { slug: app.slug } })
    } catch (e) {
      if (isApiError(e, 422) && e.fields) {
        for (const [field, message] of Object.entries(e.fields)) {
          if (field === "name" || field === "slug") setError(field, { message })
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
        New app
      </DialogTrigger>
      <DialogContent className="mt-5 sm:max-w-md">
        <form
          className="space-y-5"
          noValidate
          onSubmit={handleSubmit(onSubmit)}
        >
          <DialogHeader>
            <DialogTitle>New app</DialogTitle>
            <DialogDescription>
              You'll be its admin. The slug goes into the app's update URL (
              <code className="whitespace-nowrap">?app=slug</code>) and can't be
              changed later.
            </DialogDescription>
          </DialogHeader>

          <div>
            <Label htmlFor="app-name">Name</Label>
            <Input
              id="app-name"
              className="mt-2 h-9"
              placeholder="Example App"
              {...register("name", {
                required: "Name is required.",
                // Suggest a slug until the user edits it themselves.
                onChange: (e) => {
                  if (!getFieldState("slug").isDirty) {
                    setValue("slug", slugify(e.target.value))
                  }
                },
              })}
            />
            {errors.name && (
              <p className="mt-1.5 text-sm text-destructive">
                {errors.name.message}
              </p>
            )}
          </div>

          <div>
            <Label htmlFor="app-slug">Slug</Label>
            <Input
              id="app-slug"
              className="mt-2 h-9 font-mono"
              placeholder="example-app"
              {...register("slug", {
                required: "Slug is required.",
                pattern: {
                  value: /^[a-z0-9][a-z0-9-]{1,62}$/,
                  message:
                    "2-63 characters: lowercase letters, digits and dashes.",
                },
              })}
            />
            {errors.slug && (
              <p className="mt-1.5 text-sm text-destructive">
                {errors.slug.message}
              </p>
            )}
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
              Create app
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
