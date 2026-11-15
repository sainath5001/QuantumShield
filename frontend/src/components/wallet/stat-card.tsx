import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import type { LucideIcon } from "lucide-react";

export function StatCard({
  title,
  value,
  hint,
  icon: Icon,
  accent,
}: {
  title: string;
  value: string | number;
  hint?: string;
  icon?: LucideIcon;
  accent?: "cyan" | "elevated" | "lower" | "unknown" | "quantum";
}) {
  const accentClass =
    accent === "cyan"
      ? "text-cyan"
      : accent === "elevated"
        ? "text-elevated"
        : accent === "lower"
          ? "text-lower"
          : accent === "unknown"
            ? "text-unknown"
            : accent === "quantum"
              ? "text-quantum"
              : "text-foreground";

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
        {Icon ? <Icon className={cn("h-4 w-4", accentClass)} aria-hidden /> : null}
      </CardHeader>
      <CardContent>
        <div className={cn("text-2xl font-semibold tabular-nums", accentClass)}>{value}</div>
        {hint ? <p className="mt-1 text-xs text-muted-foreground">{hint}</p> : null}
      </CardContent>
    </Card>
  );
}
