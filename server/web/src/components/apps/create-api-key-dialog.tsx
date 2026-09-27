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
import { useCreateApiKey } from "@/hooks/use-apps"
import { getErrorMessage } from "@/lib/api"
import {
  Add01Icon,
  Copy01Icon,
  Loading03Icon,
  Tick02Icon,
} from "@hugeicons/core-free-icons"
import { HugeiconsIcon } from "@hugeicons/react"
import { useState } from "react"
import { useForm } from "react-hook-form"

/**
 * Two steps: name the key, then show it once. The server only stores its
 * hash, so this is the only time the plain key is visible.
 */
export function CreateApiKeyDialog({ slug }: { slug: string }) {
  const [open, setOpen] = useState(false)
  const [plainKey, setPlainKey] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const createKey = useCreateApiKey(slug)

  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<{ name: string }>({ defaultValues: { name: "" } })

  const onOpenChange = (next: boolean) => {
    setOpen(next)
    if (!next) {
      reset()
      setPlainKey(null)
      setCopied(false)
    }
  }

  const onSubmit = async (values: { name: string }) => {
    try {
      const { key } = await createKey.mutateAsync({ name: values.name.trim() })
      setPlainKey(key)
    } catch (e) {
      setError("root", { message: getErrorMessage(e) })
    }
  }

  const copy = async () => {
    if (!plainKey) return
    await navigator.clipboard.writeText(plainKey)
    setCopied(true)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger render={<Button />}>
        <HugeiconsIcon icon={Add01Icon} />
        New API key
      </DialogTrigger>
      <DialogContent className="mt-5 sm:max-w-md">
        {plainKey ? (
          <div className="space-y-5">
            <DialogHeader>
              <DialogTitle>Copy your key</DialogTitle>
              <DialogDescription>
                This is the only time it's shown. Store it as a secret, e.g.{" "}
                <code>OTA_API_KEY</code> in CI.
              </DialogDescription>
            </DialogHeader>

            <div className="flex items-center gap-2">
              <Input
                readOnly
                value={plainKey}
                className="h-9 font-mono text-xs"
                onFocus={(e) => e.target.select()}
              />
              <Button
                type="button"
                variant="outline"
                size="icon"
                onClick={copy}
                aria-label="Copy key"
              >
                <HugeiconsIcon icon={copied ? Tick02Icon : Copy01Icon} />
              </Button>
            </div>

            <DialogFooter>
              <Button type="button" onClick={() => onOpenChange(false)}>
                Done
              </Button>
            </DialogFooter>
          </div>
        ) : (
          <form
            className="space-y-5"
            noValidate
            onSubmit={handleSubmit(onSubmit)}
          >
            <DialogHeader>
              <DialogTitle>New API key</DialogTitle>
              <DialogDescription>
                Used by the CLI to publish updates for this app. It stops
                working if you lose access to the app.
              </DialogDescription>
            </DialogHeader>

            <div>
              <Label htmlFor="key-name">Name</Label>
              <Input
                id="key-name"
                className="mt-2 h-9"
                placeholder="github-ci"
                {...register("name", { required: "Name is required." })}
              />
              {errors.name && (
                <p className="mt-1.5 text-sm text-destructive">
                  {errors.name.message}
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
                  <HugeiconsIcon
                    icon={Loading03Icon}
                    className="animate-spin"
                  />
                )}
                Create key
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
