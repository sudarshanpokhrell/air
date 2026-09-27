import { getErrorMessage } from "@/lib/api"
import { cn } from "@/lib/utils"

/** Page shell: title, optional description, actions on the right. */
export function Page({
  title,
  description,
  actions,
  children,
  className,
}: {
  title: React.ReactNode
  description?: React.ReactNode
  actions?: React.ReactNode
  children: React.ReactNode
  className?: string
}) {
  return (
    <div
      className={cn(
        "mx-auto flex w-full max-w-5xl flex-col gap-6 px-8 py-8",
        className
      )}
    >
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="space-y-1">
          <h1 className="text-xl font-semibold">{title}</h1>
          {description && (
            <p className="text-body-sm text-muted-foreground">{description}</p>
          )}
        </div>
        {actions}
      </div>
      {children}
    </div>
  )
}

export function ErrorBanner({ error }: { error: unknown }) {
  return (
    <p
      className="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-body-sm text-destructive"
      role="alert"
    >
      {getErrorMessage(error)}
    </p>
  )
}

export function EmptyState({ children }: { children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-dashed px-6 py-12 text-center text-body-sm text-muted-foreground">
      {children}
    </div>
  )
}

export function TableCard({ children }: { children: React.ReactNode }) {
  return (
    <div className="overflow-x-auto rounded-lg border">
      <table className="w-full text-body-sm">{children}</table>
    </div>
  )
}

export function Th({
  children,
  className,
}: {
  children?: React.ReactNode
  className?: string
}) {
  return (
    <th className={cn("px-4 py-2.5 font-medium", className)}>{children}</th>
  )
}

export const theadClass =
  "border-b bg-surface-2 text-left text-caption text-muted-foreground"
