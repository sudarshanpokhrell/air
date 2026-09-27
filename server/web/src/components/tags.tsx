import { cn } from "@/lib/utils"

const tagTones = {
  default: "border-border text-ink-muted",
  strong: "border-transparent bg-surface-3 text-foreground",
  danger: "border-destructive/30 text-destructive",
} as const

/** Small bordered label, e.g. a role. */
export function Tag({
  children,
  tone = "default",
  className,
}: {
  children: React.ReactNode
  tone?: keyof typeof tagTones
  className?: string
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-md border px-2 py-0.5 text-caption font-medium",
        tagTones[tone],
        className
      )}
    >
      {children}
    </span>
  )
}

const dotTones = {
  ok: "bg-success",
  warn: "bg-warning",
  off: "bg-ink-tertiary",
  danger: "bg-destructive",
} as const

/** Colored dot + label for a status. */
export function StatusDot({
  tone,
  children,
}: {
  tone: keyof typeof dotTones
  children: React.ReactNode
}) {
  return (
    <span className="inline-flex items-center gap-2">
      <span className={cn("size-1.5 shrink-0 rounded-full", dotTones[tone])} />
      {children}
    </span>
  )
}
