import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, apiFetch, getApiBaseUrl } from "./client";

describe("getApiBaseUrl", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("defaults to local API", () => {
    vi.stubEnv("NEXT_PUBLIC_QGUARD_API_BASE_URL", "");
    expect(getApiBaseUrl()).toBe("http://127.0.0.1:8080");
  });

  it("strips trailing slash", () => {
    vi.stubEnv("NEXT_PUBLIC_QGUARD_API_BASE_URL", "http://localhost:8080/");
    expect(getApiBaseUrl()).toBe("http://localhost:8080");
  });
});

describe("apiFetch", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it("parses API error envelopes", async () => {
    vi.stubEnv("NEXT_PUBLIC_QGUARD_API_BASE_URL", "http://example.test");
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 400,
        json: async () => ({
          error: { code: "INVALID_ADDRESS", message: "bad address" },
        }),
      }),
    );
    await expect(apiFetch("/api/v1/scan")).rejects.toMatchObject({
      code: "INVALID_ADDRESS",
      message: "bad address",
      status: 400,
    } satisfies Partial<ApiError>);
  });
});
