import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { ScanJob } from "@/lib/models/types";

const stageLabels: Record<string, string> = {
  queued: "Queued",
  classify_account: "Classifying account",
  fetch_history: "Fetching transaction history",
  analyze_exposure: "Analyzing public-key exposure",
  assess_risk: "Assessing risk classification",
  build_report: "Building security report",
  completed: "Completed",
  failed: "Failed",
};

export function ScanProgressPanel({ job }: { job: ScanJob }) {
  const label = stageLabels[job.stage] ?? job.stage;
  const pct =
    job.status === "completed"
      ? 100
      : job.status === "failed"
        ? 100
        : Math.min(95, Math.round(((job.progress.txs_processed || 0) / Math.max(job.progress.txs_fetched || 1, 1)) * 80) + 15);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Scan progress</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Status</span>
          <span className="font-medium capitalize">{job.status}</span>
        </div>
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Stage</span>
          <span className="font-medium">{label}</span>
        </div>
        <div
          className="h-2 w-full overflow-hidden rounded-full bg-muted"
          role="progressbar"
          aria-valuenow={pct}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-label="Scan progress"
        >
          <div
            className={`h-full transition-all ${job.status === "failed" ? "bg-destructive" : "bg-cyan"}`}
            style={{ width: `${pct}%` }}
          />
        </div>
        <p className="text-xs text-muted-foreground">
          Transactions fetched {job.progress.txs_fetched} · processed {job.progress.txs_processed}
          {job.progress.message ? ` · ${job.progress.message}` : ""}
        </p>
        {job.error ? <p className="text-sm text-destructive">{job.error}</p> : null}
      </CardContent>
    </Card>
  );
}
