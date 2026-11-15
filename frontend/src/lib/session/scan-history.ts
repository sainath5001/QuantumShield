"use client";

import type { RiskClassification, ScanStatus, SessionScanRecord } from "@/lib/models/types";

const STORAGE_KEY = "qguard.session.scans.v1";

function readAll(): SessionScanRecord[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = window.sessionStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw) as SessionScanRecord[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function writeAll(records: SessionScanRecord[]) {
  if (typeof window === "undefined") return;
  window.sessionStorage.setItem(STORAGE_KEY, JSON.stringify(records.slice(0, 100)));
}

export function listSessionScans(): SessionScanRecord[] {
  return readAll().sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1));
}

export function upsertSessionScan(record: SessionScanRecord) {
  const all = readAll().filter((r) => r.scanId !== record.scanId);
  all.unshift(record);
  writeAll(all);
}

export function updateSessionScan(
  scanId: string,
  patch: Partial<Pick<SessionScanRecord, "status" | "classification" | "reportId" | "completedAt">>,
) {
  const all = readAll();
  const idx = all.findIndex((r) => r.scanId === scanId);
  if (idx < 0) return;
  all[idx] = { ...all[idx], ...patch };
  writeAll(all);
}

export interface SessionStats {
  total: number;
  elevated: number;
  lower: number;
  unknown: number;
  review: number;
  completed: number;
  failed: number;
}

export function computeSessionStats(records: SessionScanRecord[] = listSessionScans()): SessionStats {
  const stats: SessionStats = {
    total: records.length,
    elevated: 0,
    lower: 0,
    unknown: 0,
    review: 0,
    completed: 0,
    failed: 0,
  };
  for (const r of records) {
    if (r.status === "completed") stats.completed += 1;
    if (r.status === "failed") stats.failed += 1;
    switch (r.classification as RiskClassification | undefined) {
      case "ELEVATED_EXPOSURE":
        stats.elevated += 1;
        break;
      case "LOWER_EXPOSURE":
        stats.lower += 1;
        break;
      case "UNKNOWN":
        stats.unknown += 1;
        break;
      case "REQUIRES_REVIEW":
        stats.review += 1;
        break;
      default:
        break;
    }
  }
  return stats;
}

export function recordScanStarted(scanId: string, address: string, status: ScanStatus = "queued") {
  upsertSessionScan({
    scanId,
    address,
    status,
    createdAt: new Date().toISOString(),
  });
}
