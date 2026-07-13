import * as React from "react";
import { cn } from "@/lib/cn";

type HeadingLevel = "display" | "lg" | "md" | "sm";

const sizeStyles: Record<HeadingLevel, string> = {
  display:
    "font-display text-[44px] sm:text-[56px] lg:text-[64px] font-semibold leading-[1.05] tracking-[-0.03em]",
  lg: "font-display text-[36px] sm:text-[42px] lg:text-[48px] font-semibold leading-[1.1] tracking-[-0.02em]",
  md: "font-display text-[28px] sm:text-[32px] lg:text-[36px] font-semibold leading-[1.15] tracking-[-0.01em]",
  sm: "font-display text-[22px] sm:text-[26px] lg:text-[28px] font-semibold leading-[1.2] tracking-[0em]",
};

export interface SectionHeadingProps
  extends React.HTMLAttributes<HTMLHeadingElement> {
  level?: HeadingLevel;
  as?: "h1" | "h2" | "h3" | "h4";
}

export function SectionHeading({
  className,
  level = "lg",
  as,
  ...props
}: SectionHeadingProps) {
  const Tag = (as ?? (level === "display" ? "h1" : "h2")) as React.ElementType;
  return (
    <Tag className={cn("text-on-surface", sizeStyles[level], className)} {...props} />
  );
}
