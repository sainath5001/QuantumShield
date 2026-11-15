"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Header } from "@/components/layout/header";
import { EmptyState } from "@/components/states/empty-state";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { RiskBadge } from "@/components/wallet/risk-badge";
import { listSessionScans } from "@/lib/session/scan-history";
import type { SessionScanRecord } from "@/lib/models/types";
import { formatIso, shortenAddress } from "@/lib/utils";

export default function ReportsIndexPage() {
  const [records, setRecords] = useState<SessionScanRecord[]>([]);

  useEffect(() => {
    setRecords(listSessionScans().filter((r) => r.reportId || r.status === "completed"));
  }, []);

  return (
    <div>
      <Header
        title="Reports"
        subtitle="Security reports from scans in this browser session. Backend storage is in-memory."
      />
      <div className="space-y-6 p-6">
        <Card>
          <CardHeader>
            <CardTitle>Availability notice</CardTitle>
            <CardDescription>
              The Go API stores scans and reports in memory. Restarting the backend clears them. This page
              indexes report IDs captured in your current browser session only — it does not invent historical data.
            </CardDescription>
          </CardHeader>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Available reports</CardTitle>
          </CardHeader>
          <CardContent>
            {!records.length ? (
              <EmptyState
                title="No reports in this session"
                description="Complete a wallet scan to generate a report ID, then open it here."
                action={
                  <Link
                    href="/scanner"
                    className="inline-flex h-8 items-center rounded-md bg-cyan px-3 text-xs font-semibold text-background"
                  >
                    Run a scan
                  </Link>
                }
              />
            ) : (
              <div className="overflow-x-auto rounded-lg border border-border">
                <table className="w-full min-w-[700px] text-left text-sm">
                  <thead className="bg-muted/40 text-xs uppercase text-muted-foreground">
                    <tr>
                      <th className="px-3 py-2">Wallet</th>
                      <th className="px-3 py-2">Scan</th>
                      <th className="px-3 py-2">Report</th>
                      <th className="px-3 py-2">Classification</th>
                      <th className="px-3 py-2">Timestamp</th>
                      <th className="px-3 py-2">Open</th>
                    </tr>
                  </thead>
                  <tbody>
                    {records.map((r) => (
                      <tr key={r.scanId} className="border-t border-border">
                        <td className="px-3 py-2 font-mono text-xs text-cyan">{shortenAddress(r.address, 6)}</td>
                        <td className="px-3 py-2 font-mono text-xs">{shortenAddress(r.scanId, 4)}</td>
                        <td className="px-3 py-2 font-mono text-xs">
                          {r.reportId ? shortenAddress(r.reportId, 4) : "—"}
                        </td>
                        <td className="px-3 py-2">
                          <RiskBadge classification={r.classification} />
                        </td>
                        <td className="px-3 py-2 text-muted-foreground">
                          {formatIso(r.completedAt || r.createdAt)}
                        </td>
                        <td className="px-3 py-2">
                          {r.reportId ? (
                            <Link className="text-cyan hover:underline" href={`/reports/${r.reportId}`}>
                              Details
                            </Link>
                          ) : (
                            <Link className="text-cyan hover:underline" href={`/wallet/${r.address}?scan=${r.scanId}`}>
                              Analysis
                            </Link>
                          )}
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
