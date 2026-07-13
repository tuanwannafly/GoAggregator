import * as React from "react";
import { cn } from "@/lib/cn";

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {
  invalid?: boolean;
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, invalid, ...props }, ref) => {
    return (
      <input
        ref={ref}
        className={cn(
          "w-full rounded-md bg-surface-elevated text-on-surface text-[16px] leading-[1.6] font-sans px-3.5 py-3 border transition-colors duration-150 placeholder:text-on-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-light focus-visible:border-primary disabled:opacity-50 disabled:cursor-not-allowed",
          invalid
            ? "border-error focus-visible:ring-error"
            : "border-border",
          className,
        )}
        {...props}
      />
    );
  },
);

Input.displayName = "Input";
