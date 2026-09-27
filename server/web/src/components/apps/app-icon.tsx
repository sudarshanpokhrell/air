import { cn } from "@/lib/utils"

/** Gray letter tile standing in for an app's icon. */
export function AppIcon({
  name,
  className,
}: {
  name: string
  className?: string
}) {
  return (
    <span
      aria-hidden
      className={cn(
        "flex shrink-0 items-center justify-center rounded-md border bg-surface-2 text-sm font-semibold text-ink-muted",
        className
      )}
    >
      {name.trim().charAt(0).toUpperCase()}
    </span>
  )
}
