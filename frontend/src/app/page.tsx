"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { AlertTriangle, FileSearch, ShieldAlert, ShieldCheck } from "lucide-react";
import { ExposureDistributionChart } from "@/components/charts/exposure-distribution";
import { Header } from "@/components/layout/header";
import { EmptyState } from "@/components/states/empty-state";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { RiskBadge } from "@/components/wallet/risk-badge";
import { ScanForm } from "@/components/wallet/scan-form";
import { StatCard } from "@/components/wallet/stat-card";
import { ApiError } from "@/lib/api/client";
import { pollScan, startScan } from "@/lib/api/qguard";
import {
  computeSessionStats,
  listSessionScans,
  recordScanStarted,
  updateSessionScan,
  type SessionStats,
} from "@/lib/session/scan-history";
import type { SessionScanRecord } from "@/lib/models/types";
import { formatIso, shortenAddress } from "@/lib/utils";
import { useRouter } from "next/navigation";

export default function OverviewPage() {
  const router = useRouter();
  const [records, setRecords] = useState<SessionScanRecord[]>([]);
  const [stats, setStats] = useState<SessionStats>(computeSessionStats([]));
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const list = listSessionScans();
    setRecords(list);
    setStats(computeSessionStats(list));
  }, []);

  async function onScan(values: { address: string }) {
    setLoading(true);
    setError(null);
    try {
      const job = await startScan(values.address);
      recordScanStarted(job.id, job.address, job.status);
      const done = await pollScan(job.id, {
        onUpdate: (u) => {
          updateSessionScan(u.id, { status: u.status });
        },
      });
      updateSessionScan(done.id, {
        status: done.status,
        classification: done.risk?.classification,
        reportId: done.report_id,
        completedAt: done.completed_at,
      });
      const list = listSessionScans();
      setRecords(list);
      setStats(computeSessionStats(list));
      if (done.status === "completed") {
        router.push(`/wallet/${done.address}?scan=${done.id}`);
      } else {
        setError(done.error || "Scan failed.");
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Unable to start scan.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div>
      <Header
        title="Overview"
        subtitle="Classical secp256k1 public-key exposure analysis for Ethereum Mainnet addresses."
      />
      <div className="space-y-6 p-6">
        <Card>
          <CardHeader>
            <CardTitle>Q-Guard</CardTitle>
            <CardDescription>
              Assess whether a public Ethereum address has verifiable secp256k1 public-key exposure
              within scoped on-chain history. Educational quantum context is separate from scan evidence.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm text-muted-foreground">
            <p>
              Exposure is not present-day compromise. Absence of exposure in scope is not a quantum-safe claim.
              Private keys are never requested.
            </p>
            <div className="rounded-md border border-border bg-muted/30 p-3 text-xs leading-relaxed">
              <strong className="text-foreground">Limitations:</strong> Mainnet only; normal transactions within
              provider caps; in-memory scan storage (lost on backend restart); session metrics below are
              browser-local because the API does not list historical scans.
            </div>
          </CardContent>
        </Card>

        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          <StatCard title="Session scans" value={stats.total} icon={FileSearch} accent="cyan" hint="This browser session" />
          <StatCard title="Elevated" value={stats.elevated} icon={ShieldAlert} accent="elevated" />
          <StatCard title="Lower" value={stats.lower} icon={ShieldCheck} accent="lower" />
          <StatCard title="Unknown / review" value={stats.unknown + stats.review} icon={AlertTriangle} accent="unknown" />
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Quick wallet scanner</CardTitle>
              <CardDescription>Public address only — no wallet connection.</CardDescription>
            </CardHeader>
            <CardContent>
              <ScanForm onSubmit={onScan} loading={loading} />
              {error ? <p className="mt-3 text-sm text-destructive" role="alert">{error}</p> : null}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Exposure distribution</CardTitle>
              <CardDescription>Classifications from completed scans in this session.</CardDescription>
            </CardHeader>
            <CardContent>
              <ExposureDistributionChart stats={stats} />
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <div>
              <CardTitle>Recent scans</CardTitle>
              <CardDescription>Session-local index. Backend storage is ephemeral.</CardDescription>
            </div>
            <Link href="/scanner" className="inline-flex h-8 items-center rounded-md border border-border px-3 text-xs hover:bg-muted/60">
              Open scanner
            </Link>
          </CardHeader>
          <CardContent>
            {!records.length ? (
              <EmptyState
                title="No scans in this session"
                description="Run a wallet scan to populate recent activity. Nothing is invented."
                action={
                  <Link href="/scanner" className="inline-flex h-8 items-center rounded-md bg-cyan px-3 text-xs font-semibold text-background hover:bg-cyan/90">
                    Scan an address
                  </Link>
                }
              />
            ) : (
              <div className="overflow-x-auto rounded-lg border border-border">
                <table className="w-full min-w-[640px] text-left text-sm">
                  <thead className="bg-muted/40 text-xs uppercase text-muted-foreground">
                    <tr>
                      <th className="px-3 py-2">Address</th>
                      <th className="px-3 py-2">Status</th>
                      <th className="px-3 py-2">Classification</th>
                      <th className="px-3 py-2">Started</th>
                      <th className="px-3 py-2">Open</th>
                    </tr>
                  </thead>
                  <tbody>
                    {records.slice(0, 8).map((r) => (
                      <tr key={r.scanId} className="border-t border-border">
                        <td className="px-3 py-2 font-mono text-xs text-cyan">{shortenAddress(r.address, 6)}</td>
                        <td className="px-3 py-2 capitalize">{r.status}</td>
                        <td className="px-3 py-2">
                          <RiskBadge classification={r.classification} />
                        </td>
                        <td className="px-3 py-2 text-muted-foreground">{formatIso(r.createdAt)}</td>
                        <td className="px-3 py-2">
                          <Link className="text-cyan hover:underline" href={`/wallet/${r.address}?scan=${r.scanId}`}>
                            Analysis
                          </Link>
                          {r.reportId ? (
                            <>
                              {" · "}
                              <Link className="text-cyan hover:underline" href={`/reports/${r.reportId}`}>
                                Report
                              </Link>
                            </>
                          ) : null}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
