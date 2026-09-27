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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useAddMember } from "@/hooks/use-apps"
import { getErrorMessage, isApiError } from "@/lib/api"
import type { AddMemberInput } from "@/types/apps"
import { Add01Icon, Loading03Icon } from "@hugeicons/core-free-icons"
import { HugeiconsIcon } from "@hugeicons/react"
import { useState } from "react"
import { Controller, useForm } from "react-hook-form"
import { toast } from "sonner"

export function AddMemberDialog({ slug }: { slug: string }) {
  const [open, setOpen] = useState(false)
  const addMember = useAddMember(slug)

  const {
    register,
    control,
    handleSubmit,
    reset,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<AddMemberInput>({
    defaultValues: { email: "", role: "developer" },
  })

  const onOpenChange = (next: boolean) => {
    setOpen(next)
    if (!next) reset()
  }

  const onSubmit = async (values: AddMemberInput) => {
    try {
      await addMember.mutateAsync({ ...values, email: values.email.trim() })
      toast.success(`${values.email} added as ${values.role}.`)
      onOpenChange(false)
    } catch (e) {
      if (isApiError(e, 422) && e.fields?.email) {
        setError("email", { message: `Email ${e.fields.email}.` })
        return
      }
      setError("root", { message: getErrorMessage(e) })
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger render={<Button />}>
        <HugeiconsIcon icon={Add01Icon} />
        Add member
      </DialogTrigger>
      <DialogContent className="mt-5 sm:max-w-md">
        <form
          className="space-y-5"
          noValidate
          onSubmit={handleSubmit(onSubmit)}
        >
          <DialogHeader>
            <DialogTitle>Add a member</DialogTitle>
            <DialogDescription>
              The user must already have an account. Ask a global admin to
              create one.
            </DialogDescription>
          </DialogHeader>

          <div>
            <Label htmlFor="member-email">Email</Label>
            <Input
              id="member-email"
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
            <Label htmlFor="member-role">Role</Label>
            <Controller
              control={control}
              name="role"
              render={({ field }) => (
                <Select
                  value={field.value}
                  onValueChange={(v) => v && field.onChange(v)}
                >
                  <SelectTrigger id="member-role" className="mt-2 h-9 w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="developer">Developer</SelectItem>
                    <SelectItem value="admin">Admin</SelectItem>
                  </SelectContent>
                </Select>
              )}
            />
            <p className="mt-1.5 text-xs text-muted-foreground">
              Developers publish and roll back. Admins also manage members,
              platforms, code signing and every API key.
            </p>
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
              Add member
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
