import * as React from "react";
import { cn } from "@/lib/cn";

type CardVariant = "default" | "muted" | "primary-light" | "neutral";

const variantStyles: Record<CardVariant, string> = {
  default:
    "bg-surface-elevated text-on-surface border border-border",
  muted:
    "bg-surface-muted text-on-surface border border-transparent",
  "primary-light":
    "bg-primary-light text-on-surface border border-transparent",
  neutral:
    "bg-neutral text-on-surface border border-transparent",
};

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: CardVariant;
}

export const Card = React.forwardRef<HTMLDivElement, CardProps>(
  ({ className, variant = "default", ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          "rounded-lg p-8 shadow-[0_1px_2px_rgba(15,23,42,0.04),0_4px_12px_rgba(15,23,42,0.04)]",
          variantStyles[variant],
          className,
        )}
        {...props}
      />
    );
  },
);

Card.displayName = "Card";
