"use client";

import Link from "next/link";
import { useParams, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Header } from "@/components/layout/header";
import { EmptyState } from "@/components/states/empty-state";
import { ErrorState } from "@/components/states/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { AddressDisplay } from "@/components/wallet/address-display";
import { EvidenceTable } from "@/components/wallet/evidence-table";
import { FindingBadge, RiskBadge } from "@/components/wallet/risk-badge";
import { ApiError } from "@/lib/api/client";
import { getScan, getWalletExposure, getWalletRisk } from "@/lib/api/qguard";
import { isValidEthereumAddress } from "@/lib/validation/address";
import type { RiskAssessment, ScanJob, WalletExposureReport } from "@/lib/models/types";
import { formatIso } from "@/lib/utils";

export default function WalletAnalysisPage() {
  const params = useParams<{ address: string }>();
  const search = useSearchParams();
  const address = decodeURIComponent(params.address || "");
  const scanId = search.get("scan");
  const reportId = search.get("report");

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [job, setJob] = useState<ScanJob | null>(null);
  const [exposure, setExposure] = useState<WalletExposureReport | null>(null);
  const [risk, setRisk] = useState<RiskAssessment | null>(null);

  useEffect(() => {
    if (!isValidEthereumAddress(address)) {
      setError("Invalid Ethereum address in URL.");
      setLoading(false);
      return;
    }

    let cancelled = false;
    async function load() {
      setLoading(true);
      setError(null);
      try {
        if (scanId) {
          const j = await getScan(scanId);
          if (cancelled) return;
          setJob(j);
          setExposure(j.exposure ?? null);
          setRisk(j.risk ?? null);
        } else {
          const [exp, rk] = await Promise.all([
            getWalletExposure(address),
            getWalletRisk(address).catch((e) => {
              if (e instanceof ApiError && e.status === 404) return null;
              throw e;
            }),
          ]);
          if (cancelled) return;
          setExposure(exp);
          setRisk(rk);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof ApiError ? err.message : "Failed to load wallet analysis.");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [address, scanId]);

  const resolvedReportId = reportId || job?.report_id;

  return (
    <div>
      <Header title="Wallet analysis" subtitle="Confirmed findings vs unknown / out-of-scope information." />
      <div className="space-y-6 p-6">
        {!isValidEthereumAddress(address) ? (
          <ErrorState title="Invalid address" message="URL must contain a 0x-prefixed 20-byte hex address." />
        ) : loading ? (
          <div className="space-y-3">
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-48 w-full" />
          </div>
        ) : error ? (
          <ErrorState title="Analysis unavailable" message={error} />
        ) : (
          <>
            <Card>
              <CardHeader>
                <CardTitle>Wallet</CardTitle>
                <CardDescription>Public address under analysis — Mainnet only.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <AddressDisplay address={address} />
                <div className="flex flex-wrap gap-2">
                  <Badge variant="outline">
                    Account: {risk?.account_kind || exposure?.account_kind || job?.account?.kind || "unknown"}
                  </Badge>
                  <RiskBadge classification={risk?.classification} />
                  <FindingBadge finding={exposure?.finding} />
                </div>
                {resolvedReportId ? (
                  <p className="text-sm">
                    <Link className="text-cyan hover:underline" href={`/reports/${resolvedReportId}`}>
                      Open security report
                    </Link>
                  </p>
                ) : null}
              </CardContent>
            </Card>

            <div className="grid gap-6 lg:grid-cols-2">
              <Card>
                <CardHeader>
                  <CardTitle>Cryptography</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2 text-sm">
                  <Row label="Curve" value={risk?.curve || "secp256k1 (Ethereum EOA default)"} />
                  <Row label="Signature scheme" value={risk?.signature_scheme || "—"} />
                  <Row
                    label="Public-key exposure"
                    value={
                      risk?.public_key_exposed || exposure?.confirmed_exposure
                        ? "Confirmed in scope"
                        : "Not confirmed in scope"
                    }
                  />
                  <Row label="Evidence quality" value={risk?.evidence_quality || "—"} />
                  <Row label="Risk model" value={risk?.risk_model_version || "—"} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Coverage</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2 text-sm">
                  <Row label="History coverage" value={risk?.history_coverage || exposure?.history_coverage_status || "—"} />
                  <Row label="Analyzed txs" value={String(risk?.analyzed_tx_count ?? exposure?.analyzed_tx_count ?? "—")} />
                  <Row label="Unsupported txs" value={String(risk?.unsupported_tx_count ?? exposure?.unsupported_count ?? "—")} />
                  <Row label="Failed recoveries" value={String(risk?.failed_tx_count ?? exposure?.failed_count ?? "—")} />
                  <Row label="Verified exposures" value={String(risk?.verified_exposure_count ?? exposure?.confirmed_count ?? "—")} />
                  <Row label="Assessed at" value={formatIso(risk?.assessed_at || exposure?.analyzed_at)} />
                </CardContent>
              </Card>
            </div>

            <Card>
              <CardHeader>
                <CardTitle>Recovered public-key evidence</CardTitle>
                <CardDescription>
                  Confirmed rows require successful recovery and sender verification. Unconfirmed attempts are labeled explicitly.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <EvidenceTable evidence={exposure?.evidence ?? []} />
              </CardContent>
            </Card>

            <div className="grid gap-6 lg:grid-cols-2">
              <Card>
                <CardHeader>
                  <CardTitle>Reason & next steps</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3 text-sm text-muted-foreground">
                  <p>{risk?.reason || "No completed risk assessment available for this address yet."}</p>
                  <ul className="list-disc space-y-1 pl-5">
                    {(risk?.recommended_next_steps || []).map((s) => (
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
                  {(risk?.limitations || exposure?.limitations || []).map((l) => (
                    <p key={l}>• {l}</p>
                  ))}
                  {(risk?.disclaimers || []).slice(0, 4).map((d) => (
                    <p key={d}>• {d}</p>
                  ))}
                  {!risk && !exposure ? (
                    <EmptyState title="No analysis payload" description="Start a scan from the scanner page." />
                  ) : null}
                </CardContent>
              </Card>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-4">
      <span className="text-muted-foreground">{label}</span>
      <span className="max-w-[60%] text-right font-medium text-foreground">{value}</span>
    </div>
  );
}
