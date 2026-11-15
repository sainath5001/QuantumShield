"use client";

import {
  Cell,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  Legend,
} from "recharts";
import { EmptyState } from "@/components/states/empty-state";
import type { SessionStats } from "@/lib/session/scan-history";

const COLORS = {
  elevated: "#F87171",
  lower: "#34D399",
  unknown: "#FBBF24",
  review: "#A78BFA",
};

export function ExposureDistributionChart({ stats }: { stats: SessionStats }) {
  const data = [
    { name: "Elevated", value: stats.elevated, key: "elevated" as const },
    { name: "Lower", value: stats.lower, key: "lower" as const },
    { name: "Unknown", value: stats.unknown, key: "unknown" as const },
    { name: "Review", value: stats.review, key: "review" as const },
  ].filter((d) => d.value > 0);

  if (!data.length) {
    return (
      <EmptyState
        title="No classification data yet"
        description="Complete a wallet scan in this browser session to populate the distribution chart. Counts are session-local — the API does not list historical scans."
      />
    );
  }

  return (
    <div className="h-64 w-full" role="img" aria-label="Exposure classification distribution">
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie data={data} dataKey="value" nameKey="name" innerRadius={55} outerRadius={90} paddingAngle={2}>
            {data.map((entry) => (
              <Cell key={entry.key} fill={COLORS[entry.key]} stroke="transparent" />
            ))}
          </Pie>
          <Tooltip
            contentStyle={{
              background: "#151A21",
              border: "1px solid #2A3340",
              borderRadius: 8,
              color: "#E8EEF5",
            }}
          />
          <Legend />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}
