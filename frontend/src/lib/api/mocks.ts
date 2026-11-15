/**
 * UI-only fixtures for isolated layout work when NEXT_PUBLIC_QGUARD_USE_MOCKS=true.
 * Never used in production API flows unless explicitly enabled.
 */
import type {
  HealthResponse,
  InfoResponse,
  ScanJob,
  SecurityReport,
} from "@/lib/models/types";

export const mockHealth: HealthResponse = {
  status: "ok",
  app: "Q-Guard",
  version: "0.1.0-mock",
  risk_model_version: "qguard-risk-v1",
  chain_id: 1,
  environment: "development",
};

export const mockInfo: InfoResponse = {
  app: "Q-Guard",
  version: "0.1.0-mock",
  risk_model_version: "qguard-risk-v1",
  chain: { chain_id: 1, name: "Ethereum Mainnet", native_asset: "ETH" },
  capabilities: {
    address_scan_implemented: true,
    account_lookup: true,
    history_retrieval: true,
    public_key_recovery: true,
    risk_assessment: true,
    wallet_connect_required: false,
    private_keys_accepted: false,
  },
  limitations: ["Mock fixture — not live chain data."],
};

export const mockScanQueued: ScanJob = {
  id: "mock-scan-1",
  address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
  chain_id: 1,
  status: "queued",
  stage: "queued",
  progress: { stage: "queued", txs_fetched: 0, txs_processed: 0 },
  created_at: "2026-11-15T00:00:00Z",
};

export const mockScanCompleted: ScanJob = {
  id: "mock-scan-1-done",
  address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
  chain_id: 1,
  status: "completed",
  stage: "completed",
  progress: { stage: "completed", txs_fetched: 12, txs_processed: 12 },
  created_at: "2026-11-15T00:00:00Z",
  completed_at: "2026-11-15T00:00:08Z",
  report_id: "mock-report-1",
  account: {
    address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    kind: "likely_eoa",
    has_code: false,
    balance_wei: "0",
    chain_id: 1,
    block_number: 21000000,
    limitations: ["Mock fixture."],
  },
  risk: {
    classification: "ELEVATED_EXPOSURE",
    reason: "Mock elevated finding for UI development.",
    evidence_quality: "HIGH",
    public_key_exposed: true,
    signature_scheme: "secp256k1-ecdsa",
    curve: "secp256k1",
    account_kind: "likely_eoa",
    history_coverage: "complete_within_scope",
    analyzed_tx_count: 12,
    unsupported_tx_count: 0,
    failed_tx_count: 0,
    verified_exposure_count: 2,
    sender_mismatch_count: 0,
    sender_tx_count: 12,
    distinct_public_keys: 1,
    limitations: ["Mock fixture."],
    disclaimers: ["Mock fixture — not live evidence."],
    quantum_threat_notes: ["Educational mock only."],
    recommended_next_steps: ["Disable mocks and scan against the live API."],
    risk_model_version: "qguard-risk-v1",
    assessed_at: "2026-11-15T00:00:08Z",
  },
};

export const mockReport: SecurityReport = {
  id: "mock-report-1",
  scan_id: "mock-scan-1-done",
  generated_at: "2026-11-15T00:00:08Z",
  risk_model_version: "qguard-risk-v1",
  executive_summary: "Mock report for UI development only.",
  wallet: {
    address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    account_kind: "likely_eoa",
    has_code: false,
    chain_id: 1,
  },
  cryptography: {
    curve: "secp256k1",
    signature_scheme: "secp256k1-ecdsa",
    notes: "Mock cryptography section.",
  },
  exposure_findings: {
    confirmed: true,
    finding: "public_key_exposed",
    verified_exposure_count: 2,
    distinct_public_keys: 1,
  },
  evidence: [],
  quantum_threat: {
    shor_relevance: "Mock educational note.",
    grover_relevance: "Mock educational note.",
    practical_note: "Not live evidence.",
    bullets: ["Mock fixture."],
  },
  risk: mockScanCompleted.risk!,
  coverage: {
    history_coverage: "complete_within_scope",
    evidence_quality: "HIGH",
    analyzed_tx_count: 12,
    unsupported_tx_count: 0,
    failed_tx_count: 0,
    sender_tx_count: 12,
  },
  recommended_next_steps: ["Use the live API."],
  limitations: ["Mock fixture."],
  disclaimers: ["Mock fixture — not production scan evidence."],
};
