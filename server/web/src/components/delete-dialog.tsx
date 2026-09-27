import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Loading03Icon } from "@hugeicons/core-free-icons"
import { HugeiconsIcon } from "@hugeicons/react"
import { useState } from "react"

type DeleteDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: React.ReactNode
  description?: React.ReactNode
  confirmText?: string
  onConfirm: () => void | Promise<unknown>
  pending?: boolean
  /** The destructive button's text; "Delete" unless given. */
  confirmLabel?: string
}

export function DeleteDialog({
  open,
  onOpenChange,
  ...props
}: DeleteDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent showCloseButton={false} className="gap-4 p-4 sm:max-w-sm">
        <DeleteForm onCancel={() => onOpenChange(false)} {...props} />
      </DialogContent>
    </Dialog>
  )
}

function DeleteForm({
  title,
  description,
  confirmText,
  onConfirm,
  pending,
  confirmLabel = "Delete",
  onCancel,
}: Omit<DeleteDialogProps, "open" | "onOpenChange"> & {
  onCancel: () => void
}) {
  const [typed, setTyped] = useState("")
  const confirmed = confirmText === undefined || typed === confirmText

  return (
    <form
      className="flex flex-col gap-4"
      onSubmit={(e) => {
        e.preventDefault()
        if (confirmed && !pending) void onConfirm()
      }}
    >
      <div className="flex flex-col gap-1.5">
        <DialogTitle className="text-sm">{title}</DialogTitle>
        {description && (
          <DialogDescription className="text-sm">
            {description}
          </DialogDescription>
        )}
      </div>

      {confirmText !== undefined && (
        <div className="flex flex-col gap-2">
          <label
            htmlFor="delete-confirm"
            className="text-sm text-muted-foreground"
          >
            Type{" "}
            <span className="rounded-full font-medium text-foreground select-all">
              {confirmText}
            </span>{" "}
            to confirm.
          </label>
          <Input
            id="delete-confirm"
            autoFocus
            autoComplete="off"
            spellCheck={false}
            className="h-9"
            value={typed}
            onChange={(e) => setTyped(e.target.value)}
          />
        </div>
      )}

      <div className="flex justify-end gap-2">
        <Button type="button" variant="ghost" size="sm" onClick={onCancel}>
          Cancel
        </Button>
        <Button
          type="submit"
          variant="destructive"
          size="sm"
          disabled={!confirmed || pending}
        >
          {pending && (
            <HugeiconsIcon icon={Loading03Icon} className="animate-spin" />
          )}
          {confirmLabel}
        </Button>
      </div>
    </form>
  )
}
