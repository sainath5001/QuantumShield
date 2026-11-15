"use client";

import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Header } from "@/components/layout/header";
import { ErrorState } from "@/components/states/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { AddressDisplay } from "@/components/wallet/address-display";
import { EvidenceTable } from "@/components/wallet/evidence-table";
import { FindingBadge, RiskBadge } from "@/components/wallet/risk-badge";
import { ApiError } from "@/lib/api/client";
import { getReport } from "@/lib/api/qguard";
import type { SecurityReport } from "@/lib/models/types";
import { formatIso } from "@/lib/utils";

export default function ReportDetailPage() {
  const params = useParams<{ id: string }>();
  const id = decodeURIComponent(params.id || "");
  const [report, setReport] = useState<SecurityReport | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    getReport(id)
      .then((r) => {
        if (!cancelled) setReport(r);
      })
      .catch((err) => {
        if (!cancelled) {
          setError(
            err instanceof ApiError
              ? err.message
              : "Report not found. In-memory storage may have been cleared by a backend restart.",
          );
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [id]);

  function exportJson() {
    if (!report) return;
    const blob = new Blob([JSON.stringify(report, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `qguard-report-${report.id}.json`;
    a.click();
    URL.revokeObjectURL(url);
  }

  return (
    <div>
      <Header title="Security report" subtitle="Full assessment payload from GET /api/v1/reports/:id" />
      <div className="space-y-6 p-6">
        {loading ? (
          <Skeleton className="h-64 w-full" />
        ) : error ? (
          <ErrorState title="Report unavailable" message={error} />
        ) : report ? (
          <>
            <Card>
              <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div className="space-y-2">
                  <CardTitle>Executive summary</CardTitle>
                  <CardDescription>
                    Generated {formatIso(report.generated_at)} · model {report.risk_model_version}
                  </CardDescription>
                  <AddressDisplay address={report.wallet.address} compact />
                  <div className="flex flex-wrap gap-2">
                    <RiskBadge classification={report.risk.classification} />
                    <FindingBadge finding={report.exposure_findings.finding} />
                  </div>
                </div>
                <Button type="button" variant="outline" onClick={exportJson}>
                  Export JSON
                </Button>
              </CardHeader>
              <CardContent>
                <p className="text-sm leading-relaxed text-muted-foreground">{report.executive_summary}</p>
              </CardContent>
            </Card>

            <div className="grid gap-6 lg:grid-cols-2">
              <Card>
                <CardHeader>
                  <CardTitle>Evidence summary</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2 text-sm">
                  <p>Confirmed exposure: {report.exposure_findings.confirmed ? "yes" : "no"}</p>
                  <p>Verified count: {report.exposure_findings.verified_exposure_count}</p>
                  <p>Distinct public keys: {report.exposure_findings.distinct_public_keys}</p>
                  <p>Analyzed txs: {report.coverage.analyzed_tx_count}</p>
                  <p>Unsupported txs: {report.coverage.unsupported_tx_count}</p>
                  <p>Evidence quality: {report.coverage.evidence_quality}</p>
                  <p>History coverage: {report.coverage.history_coverage}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>Cryptography</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2 text-sm text-muted-foreground">
                  <p>Curve: {report.cryptography.curve || "—"}</p>
                  <p>Scheme: {report.cryptography.signature_scheme || "—"}</p>
                  <p>{report.cryptography.notes}</p>
                </CardContent>
              </Card>
            </div>

            <Card>
              <CardHeader>
                <CardTitle>Evidence table</CardTitle>
              </CardHeader>
              <CardContent>
                <EvidenceTable evidence={report.evidence} />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Quantum threat context (educational)</CardTitle>
                <CardDescription>Not a claim that a quantum computer compromised this wallet.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2 text-sm text-muted-foreground">
                <p>{report.quantum_threat.shor_relevance}</p>
                <p>{report.quantum_threat.grover_relevance}</p>
                <p>{report.quantum_threat.practical_note}</p>
                <ul className="list-disc pl-5">
                  {report.quantum_threat.bullets.map((b) => (
                    <li key={b}>{b}</li>
                  ))}
                </ul>
              </CardContent>
            </Card>

            <div className="grid gap-6 lg:grid-cols-2">
              <Card>
                <CardHeader>
                  <CardTitle>Recommended next steps</CardTitle>
                </CardHeader>
                <CardContent>
                  <ul className="list-disc space-y-1 pl-5 text-sm text-muted-foreground">
                    {report.recommended_next_steps.map((s) => (
                      <li key={s}>{s}</li>
                    ))}
                  </ul>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>Limitations & disclaimers</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2 text-sm text-muted-foreground">
                  {report.limitations.map((l) => (
                    <p key={l}>• {l}</p>
                  ))}
                  {report.disclaimers.map((d) => (
                    <p key={d}>• {d}</p>
                  ))}
                </CardContent>
              </Card>
            </div>
          </>
        ) : null}
      </div>
    </div>
  );
}
