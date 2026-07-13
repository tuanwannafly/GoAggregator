import * as React from "react";
import { cn } from "@/lib/cn";

type BadgeVariant = "default" | "success" | "warning" | "error" | "primary";

const variantStyles: Record<BadgeVariant, string> = {
  default:
    "bg-surface-muted text-on-surface border border-transparent",
  primary:
    "bg-primary text-on-primary border border-transparent",
  success:
    "bg-[#D8E5DA] text-[#2F5D3A] border border-transparent",
  warning:
    "bg-[#EAD9B8] text-[#6B5320] border border-transparent",
  error:
    "bg-[#F4D8D6] text-[#7A2A24] border border-transparent",
};

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant;
}

export function Badge({
  className,
  variant = "default",
  ...props
}: BadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[11px] font-semibold leading-[1.1] tracking-[0.1em] uppercase font-sans",
        variantStyles[variant],
        className,
      )}
      {...props}
    />
  );
}
