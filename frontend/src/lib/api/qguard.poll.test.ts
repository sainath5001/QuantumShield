import { afterEach, describe, expect, it, vi } from "vitest";
import { pollScan } from "./qguard";

function scanPayload(status: "running" | "completed" | "failed", extra: Record<string, unknown> = {}) {
  return {
    id: "1",
    address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    chain_id: 1,
    status,
    stage: status === "running" ? "fetch_history" : status,
    progress: { stage: status === "running" ? "fetch_history" : status, txs_fetched: 1, txs_processed: status === "completed" ? 1 : 0 },
    created_at: "2026-11-15T00:00:00Z",
    ...extra,
  };
}

describe("pollScan", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it("returns when status becomes completed", async () => {
    vi.stubEnv("NEXT_PUBLIC_QGUARD_USE_MOCKS", "false");
    vi.stubEnv("NEXT_PUBLIC_QGUARD_API_BASE_URL", "http://example.test");
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => scanPayload("running"),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => scanPayload("completed", { report_id: "r1" }),
      });
    vi.stubGlobal("fetch", fetchMock);

    const updates: string[] = [];
    const job = await pollScan("1", {
      intervalMs: 10,
      onUpdate: (j) => updates.push(j.status),
    });
    expect(job.status).toBe("completed");
    expect(job.report_id).toBe("r1");
    expect(updates).toEqual(["running", "completed"]);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("surfaces failed scans without inventing success", async () => {
    vi.stubEnv("NEXT_PUBLIC_QGUARD_USE_MOCKS", "false");
    vi.stubEnv("NEXT_PUBLIC_QGUARD_API_BASE_URL", "http://example.test");
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => scanPayload("failed", { error: "RPC unavailable" }),
      }),
    );
    const job = await pollScan("1", { intervalMs: 5 });
    expect(job.status).toBe("failed");
    expect(job.error).toMatch(/RPC/);
  });
});
