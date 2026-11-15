import { ApiError, mocksEnabled } from "@/lib/api/client";

export function getQuantumApiBaseUrl(): string {
  const base = process.env.NEXT_PUBLIC_QGUARD_QUANTUM_API_BASE_URL?.trim();
  return (base && base.length > 0 ? base : "http://127.0.0.1:8090").replace(/\/$/, "");
}

async function quantumFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 120_000);
  try {
    const res = await fetch(`${getQuantumApiBaseUrl()}${path}`, {
      ...init,
      signal: controller.signal,
      headers: {
        Accept: "application/json",
        ...(init?.body ? { "Content-Type": "application/json" } : {}),
        ...init?.headers,
      },
    });
    if (!res.ok) {
      let message = `Quantum API HTTP ${res.status}`;
      try {
        const body = (await res.json()) as { detail?: string | { msg?: string }[] };
        if (typeof body.detail === "string") message = body.detail;
        else if (Array.isArray(body.detail) && body.detail[0]?.msg) message = body.detail[0].msg!;
      } catch {
        /* ignore */
      }
      throw new ApiError(res.status, "QUANTUM_API_ERROR", message);
    }
    return (await res.json()) as T;
  } catch (err) {
    if (err instanceof ApiError) throw err;
    if (err instanceof DOMException && err.name === "AbortError") {
      throw new ApiError(408, "TIMEOUT", "Quantum experiment timed out.");
    }
    throw new ApiError(
      0,
      "NETWORK_ERROR",
      err instanceof Error ? err.message : "Network error contacting the quantum API.",
    );
  } finally {
    clearTimeout(timer);
  }
}

export interface QuantumDemoMeta {
  id: string;
  name: string;
  path: string;
  notes?: string;
}

export async function getQuantumHealth(): Promise<{ status: string; version?: string }> {
  if (mocksEnabled()) {
    return { status: "ok", version: "mock" };
  }
  return quantumFetch("/health");
}

export async function listQuantumDemos(): Promise<{ demos: QuantumDemoMeta[]; limitations: string }> {
  if (mocksEnabled()) {
    return {
      limitations: "Mock quantum fixtures — not live simulation.",
      demos: [
        { id: "shor", name: "Shor toy", path: "POST /demos/shor" },
        { id: "grover", name: "Grover toy", path: "POST /demos/grover" },
      ],
    };
  }
  return quantumFetch("/demos");
}

export async function runShorDemo(body?: {
  n?: number;
  a?: number;
  shots?: number;
  seed?: number;
}): Promise<Record<string, unknown>> {
  if (mocksEnabled()) {
    throw new ApiError(503, "MOCKS_ENABLED", "Disable NEXT_PUBLIC_QGUARD_USE_MOCKS to run real quantum demos.");
  }
  return quantumFetch("/demos/shor", {
    method: "POST",
    body: JSON.stringify({ n: 15, a: 7, shots: 256, seed: 42, ...body }),
  });
}

export async function runGroverDemo(body?: {
  n_qubits?: number;
  target?: string;
  shots?: number;
  seed?: number;
}): Promise<Record<string, unknown>> {
  if (mocksEnabled()) {
    throw new ApiError(503, "MOCKS_ENABLED", "Disable NEXT_PUBLIC_QGUARD_USE_MOCKS to run real quantum demos.");
  }
  return quantumFetch("/demos/grover", {
    method: "POST",
    body: JSON.stringify({ n_qubits: 2, target: "11", shots: 256, seed: 42, ...body }),
  });
}

export async function getPqcExplainer(): Promise<Record<string, unknown>> {
  return quantumFetch("/explainers/post-quantum");
}
