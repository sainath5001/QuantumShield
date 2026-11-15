import type { ApiErrorBody } from "@/lib/models/types";

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly details?: Record<string, unknown>;

  constructor(status: number, code: string, message: string, details?: Record<string, unknown>) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

export function getApiBaseUrl(): string {
  const base = process.env.NEXT_PUBLIC_QGUARD_API_BASE_URL?.trim();
  return (base && base.length > 0 ? base : "http://127.0.0.1:8080").replace(/\/$/, "");
}

export function mocksEnabled(): boolean {
  return process.env.NEXT_PUBLIC_QGUARD_USE_MOCKS === "true";
}

async function parseError(res: Response): Promise<ApiError> {
  try {
    const body = (await res.json()) as ApiErrorBody;
    if (body?.error?.message) {
      return new ApiError(res.status, body.error.code || "UNKNOWN", body.error.message, body.error.details);
    }
  } catch {
    /* fall through */
  }
  return new ApiError(res.status, "HTTP_ERROR", `Request failed with status ${res.status}`);
}

export async function apiFetch<T>(
  path: string,
  init?: RequestInit & { timeoutMs?: number },
): Promise<T> {
  const { timeoutMs = 30_000, ...rest } = init ?? {};
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  try {
    const res = await fetch(`${getApiBaseUrl()}${path}`, {
      ...rest,
      signal: controller.signal,
      headers: {
        Accept: "application/json",
        ...(rest.body ? { "Content-Type": "application/json" } : {}),
        ...rest.headers,
      },
    });
    if (!res.ok) {
      throw await parseError(res);
    }
    if (res.status === 204) {
      return undefined as T;
    }
    return (await res.json()) as T;
  } catch (err) {
    if (err instanceof ApiError) throw err;
    if (err instanceof DOMException && err.name === "AbortError") {
      throw new ApiError(408, "TIMEOUT", "Request timed out contacting the Q-Guard API.");
    }
    throw new ApiError(
      0,
      "NETWORK_ERROR",
      err instanceof Error ? err.message : "Network error contacting the Q-Guard API.",
    );
  } finally {
    clearTimeout(timer);
  }
}
