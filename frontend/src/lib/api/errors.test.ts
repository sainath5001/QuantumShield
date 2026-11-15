import { describe, expect, it } from "vitest";
import { ApiError } from "./client";
import { humanizeApiError, humanizeQuantumError } from "./errors";

describe("humanizeApiError", () => {
  it("maps network failures", () => {
    expect(humanizeApiError(new ApiError(0, "NETWORK_ERROR", "fetch failed"))).toMatch(/Backend unavailable/i);
  });

  it("maps missing scan", () => {
    expect(humanizeApiError(new ApiError(404, "SCAN_NOT_FOUND", "gone"))).toMatch(/not found/i);
  });
});

describe("humanizeQuantumError", () => {
  it("maps unavailable quantum service", () => {
    expect(humanizeQuantumError(new ApiError(0, "NETWORK_ERROR", "x"))).toMatch(/Quantum research service unavailable/i);
  });
});
