import { apiFetch, mocksEnabled } from "@/lib/api/client";
import type {
  HealthResponse,
  InfoResponse,
  RiskAssessment,
  ScanJob,
  SecurityReport,
  WalletExposureReport,
} from "@/lib/models/types";
import {
  mockHealth,
  mockInfo,
  mockReport,
  mockScanCompleted,
  mockScanQueued,
} from "@/lib/api/mocks";

export async function getHealth(): Promise<HealthResponse> {
  if (mocksEnabled()) return mockHealth;
  return apiFetch<HealthResponse>("/api/v1/health");
}

export async function getInfo(): Promise<InfoResponse> {
  if (mocksEnabled()) return mockInfo;
  return apiFetch<InfoResponse>("/api/v1/info");
}

export async function startScan(address: string): Promise<ScanJob> {
  if (mocksEnabled()) return { ...mockScanQueued, address, created_at: new Date().toISOString() };
  return apiFetch<ScanJob>("/api/v1/scan", {
    method: "POST",
    body: JSON.stringify({ address }),
  });
}

export async function getScan(id: string): Promise<ScanJob> {
  if (mocksEnabled()) {
    return id.endsWith("-done") ? mockScanCompleted : { ...mockScanQueued, id, status: "running", stage: "fetch_history" };
  }
  return apiFetch<ScanJob>(`/api/v1/scan/${encodeURIComponent(id)}`);
}

export async function getReport(id: string): Promise<SecurityReport> {
  if (mocksEnabled()) return { ...mockReport, id };
  return apiFetch<SecurityReport>(`/api/v1/reports/${encodeURIComponent(id)}`);
}

export async function getWallet(address: string) {
  return apiFetch(`/api/v1/wallet/${encodeURIComponent(address)}`);
}

export async function getWalletExposure(address: string): Promise<WalletExposureReport> {
  return apiFetch<WalletExposureReport>(`/api/v1/wallet/${encodeURIComponent(address)}/exposure`);
}

export async function getWalletRisk(address: string): Promise<RiskAssessment> {
  return apiFetch<RiskAssessment>(`/api/v1/wallet/${encodeURIComponent(address)}/risk`);
}

/** Poll until completed or failed. Default interval 2s; default timeout 180s. Cancellation of jobs is not supported by the API. */
export async function pollScan(
  id: string,
  opts?: {
    intervalMs?: number;
    timeoutMs?: number;
    signal?: AbortSignal;
    onUpdate?: (job: ScanJob) => void;
  },
): Promise<ScanJob> {
  const intervalMs = opts?.intervalMs ?? 2000;
  const timeoutMs = opts?.timeoutMs ?? 180_000;
  const started = Date.now();
  for (;;) {
    if (opts?.signal?.aborted) {
      throw new Error("Scan polling aborted");
    }
    if (Date.now() - started > timeoutMs) {
      throw new Error(
        `Scan timed out after ${Math.round(timeoutMs / 1000)}s while still pending/running. The backend may still finish — check GET /api/v1/scan/${id}.`,
      );
    }
    const job = await getScan(id);
    opts?.onUpdate?.(job);
    if (job.status === "completed" || job.status === "failed") {
      return job;
    }
    // statuses: queued | running (API has no cancelled status)
    await new Promise((r) => setTimeout(r, intervalMs));
  }
}
