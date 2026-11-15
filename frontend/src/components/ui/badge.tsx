import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium tracking-wide",
  {
    variants: {
      variant: {
        default: "border-border bg-muted text-foreground",
        cyan: "border-cyan/40 bg-cyan-muted text-cyan",
        quantum: "border-quantum/40 bg-quantum-muted text-quantum",
        elevated: "border-elevated/40 bg-elevated/15 text-elevated",
        lower: "border-lower/40 bg-lower/15 text-lower",
        unknown: "border-unknown/40 bg-unknown/15 text-unknown",
        review: "border-review/40 bg-review/15 text-review",
        outline: "border-border text-muted-foreground",
      },
    },
    defaultVariants: { variant: "default" },
  },
);

export function Badge({
  className,
  variant,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & VariantProps<typeof badgeVariants>) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />;
}
