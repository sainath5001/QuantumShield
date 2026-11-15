import { ApiError } from "@/lib/api/client";

/** Map API / network failures to user-visible copy. Never invent a successful assessment. */
export function humanizeApiError(err: unknown, fallback = "Request failed."): string {
  if (err instanceof ApiError) {
    switch (err.code) {
      case "NETWORK_ERROR":
        return `Backend unavailable: ${err.message}. Is the Go API running on the configured URL?`;
      case "TIMEOUT":
        return "Request timed out. The scan or provider may still be running — retry or check backend logs.";
      case "INVALID_ADDRESS":
      case "INVALID_REQUEST":
        return err.message;
      case "SCAN_NOT_FOUND":
        return "Scan not found. In-memory storage may have been cleared by a backend restart.";
      case "RPC_UNAVAILABLE":
        return `Ethereum RPC failure: ${err.message}`;
      case "HISTORY_PROVIDER_ERROR":
        return `Etherscan / history provider failure: ${err.message}`;
      case "RATE_LIMITED":
        return `Rate limited by a provider: ${err.message}`;
      case "CAPACITY_EXCEEDED":
        return err.message;
      default:
        return `${err.code}: ${err.message}`;
    }
  }
  if (err instanceof Error) {
    if (err.message.includes("Scan polling aborted")) {
      return "Scan polling stopped.";
    }
    if (err.message.includes("Scan timed out")) {
      return err.message;
    }
    return err.message;
  }
  return fallback;
}

export function humanizeQuantumError(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.code === "NETWORK_ERROR" || err.status === 0) {
      return "Quantum research service unavailable. Start it with scripts/run-quantum.sh (port 8090).";
    }
    return err.message;
  }
  if (err instanceof Error) return err.message;
  return "Quantum experiment failed.";
}
