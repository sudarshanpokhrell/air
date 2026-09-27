import * as React from "react"
import { ViewIcon, ViewOffSlashIcon } from "@hugeicons/core-free-icons"
import { HugeiconsIcon } from "@hugeicons/react"
import { cn } from "cn"
import { Input } from "@/components/ui/input"

// Password field with a show/hide toggle. Takes the same props as Input, so
// react-hook-form's register() spreads onto it directly.
function PasswordInput({
  className,
  id,
  ...props
}: Omit<React.ComponentProps<"input">, "type">) {
  const [isVisible, setIsVisible] = React.useState(false)

  return (
    <div className="relative">
      <Input
        className={cn("pe-9", className)}
        id={id}
        type={isVisible ? "text" : "password"}
        {...props}
      />
      <button
        aria-controls={id}
        aria-label={isVisible ? "Hide password" : "Show password"}
        aria-pressed={isVisible}
        className="absolute inset-y-0 inset-e-0 flex h-full w-9 items-center justify-center rounded-e-md text-muted-foreground/80 transition-[color,box-shadow] outline-none hover:text-foreground focus:z-10 focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50"
        disabled={props.disabled}
        onClick={() => setIsVisible((prev) => !prev)}
        type="button"
      >
        <HugeiconsIcon
          icon={isVisible ? ViewOffSlashIcon : ViewIcon}
          aria-hidden="true"
          size={16}
        />
      </button>
    </div>
  )
}

export { PasswordInput }
