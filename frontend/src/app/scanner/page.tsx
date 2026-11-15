"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { Header } from "@/components/layout/header";
import { ErrorState } from "@/components/states/error-state";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ScanForm } from "@/components/wallet/scan-form";
import { ScanProgressPanel } from "@/components/wallet/scan-progress";
import { ApiError } from "@/lib/api/client";
import { pollScan, startScan } from "@/lib/api/qguard";
import { recordScanStarted, updateSessionScan } from "@/lib/session/scan-history";
import type { ScanJob } from "@/lib/models/types";

export default function ScannerPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [job, setJob] = useState<ScanJob | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function onScan(values: { address: string }) {
    setLoading(true);
    setError(null);
    setJob(null);
    try {
      const started = await startScan(values.address);
      recordScanStarted(started.id, started.address, started.status);
      setJob(started);
      const done = await pollScan(started.id, {
        onUpdate: (u) => {
          setJob(u);
          updateSessionScan(u.id, { status: u.status });
        },
      });
      updateSessionScan(done.id, {
        status: done.status,
        classification: done.risk?.classification,
        reportId: done.report_id,
        completedAt: done.completed_at,
      });
      setJob(done);
      if (done.status === "completed") {
        router.push(`/wallet/${done.address}?scan=${done.id}${done.report_id ? `&report=${done.report_id}` : ""}`);
      } else {
        setError(done.error || "Scan failed.");
      }
    } catch (err) {
      setError(err instanceof ApiError ? `${err.code}: ${err.message}` : "Unable to reach the Q-Guard API.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div>
      <Header
        title="Wallet scanner"
        subtitle="Submit a public Ethereum address. No wallet connection. No private keys."
      />
      <div className="mx-auto grid max-w-3xl gap-6 p-6">
        <Card>
          <CardHeader>
            <CardTitle>Scan an address</CardTitle>
            <CardDescription>
              Starts <code className="text-cyan">POST /api/v1/scan</code> then polls until completed or failed.
              Analysis covers Mainnet history within provider caps.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ScanForm onSubmit={onScan} loading={loading} />
          </CardContent>
        </Card>

        {job ? <ScanProgressPanel job={job} /> : null}
        {error ? <ErrorState title="Scan error" message={error} /> : null}
      </div>
    </div>
  );
}
