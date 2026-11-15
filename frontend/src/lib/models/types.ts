/** TypeScript models aligned with Go domain JSON (Prompt 5 runtime shapes). */

export type ScanStatus = "queued" | "running" | "completed" | "failed";

export type ScanStage =
  | "queued"
  | "classify_account"
  | "fetch_history"
  | "analyze_exposure"
  | "assess_risk"
  | "build_report"
  | "completed"
  | "failed";

export type AccountKind = "likely_eoa" | "contract_account" | "unknown";

export type RiskClassification =
  | "LOWER_EXPOSURE"
  | "ELEVATED_EXPOSURE"
  | "UNKNOWN"
  | "REQUIRES_REVIEW";

export type EvidenceQuality = "HIGH" | "MEDIUM" | "LOW" | "MANUAL";

export type WalletExposureFinding =
  | "public_key_exposed"
  | "no_exposure_observed_in_scope"
  | "insufficient_evidence"
  | "requires_manual_review";

export type HistoryCoverageStatus =
  | "complete_within_scope"
  | "partial"
  | "rate_limited"
  | "failed"
  | "unsupported"
  | "unknown";

export interface ApiErrorBody {
  error: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
}

export interface HealthResponse {
  status: string;
  app: string;
  version: string;
  risk_model_version: string;
  chain_id: number;
  environment: string;
}

export interface InfoResponse {
  app: string;
  version: string;
  risk_model_version: string;
  chain: { chain_id: number; name: string; native_asset: string };
  capabilities: {
    address_scan_implemented: boolean;
    account_lookup: boolean;
    history_retrieval: boolean;
    public_key_recovery: boolean;
    risk_assessment: boolean;
    wallet_connect_required: boolean;
    private_keys_accepted: boolean;
  };
  limitations: string[];
}

export interface ScanProgress {
  stage: ScanStage;
  txs_fetched: number;
  txs_processed: number;
  message?: string;
}

export interface AccountSnapshot {
  address: string;
  kind: AccountKind;
  has_code: boolean;
  code_hash?: string;
  balance_wei: string | number;
  chain_id: number;
  block_number: number;
  limitations: string[];
}

export interface ExposureEvidence {
  wallet_address: string;
  recovered_public_key_hex?: string;
  public_key_fingerprint?: string;
  public_key_encoding?: string;
  curve: string;
  signature_scheme: string;
  transaction_hash: string;
  block_number?: number;
  transaction_type?: number;
  signing_digest_hex?: string;
  recovery_status: string;
  sender_verification_status: string;
  expected_sender: string;
  derived_sender?: string;
  evidence_source: string;
  detail?: string;
  analyzed_at: string;
}

export interface WalletExposureReport {
  address: string;
  account_kind: AccountKind;
  finding: WalletExposureFinding;
  confirmed_exposure: boolean;
  distinct_public_keys: number;
  confirmed_count: number;
  failed_count: number;
  unsupported_count: number;
  sender_mismatch_count: number;
  analyzed_tx_count: number;
  sender_tx_count: number;
  history_coverage_status: HistoryCoverageStatus;
  evidence: ExposureEvidence[];
  limitations: string[];
  analyzed_at: string;
}

export interface RiskAssessment {
  classification: RiskClassification;
  reason: string;
  evidence_quality: EvidenceQuality;
  public_key_exposed: boolean;
  signature_scheme?: string;
  curve?: string;
  account_kind: AccountKind;
  history_coverage: HistoryCoverageStatus;
  analyzed_tx_count: number;
  unsupported_tx_count: number;
  failed_tx_count: number;
  verified_exposure_count: number;
  sender_mismatch_count: number;
  sender_tx_count: number;
  distinct_public_keys: number;
  limitations: string[];
  disclaimers: string[];
  quantum_threat_notes: string[];
  recommended_next_steps: string[];
  risk_model_version: string;
  assessed_at: string;
}

export interface ScanJob {
  id: string;
  address: string;
  chain_id: number;
  status: ScanStatus;
  stage: ScanStage;
  progress: ScanProgress;
  error?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  account?: AccountSnapshot;
  exposure?: WalletExposureReport;
  risk?: RiskAssessment;
  report_id?: string;
}

export interface SecurityReport {
  id: string;
  scan_id: string;
  generated_at: string;
  risk_model_version: string;
  executive_summary: string;
  wallet: {
    address: string;
    account_kind: AccountKind;
    has_code: boolean;
    balance_wei?: string;
    chain_id: number;
    block_number?: number;
  };
  cryptography: {
    curve?: string;
    signature_scheme?: string;
    notes: string;
  };
  exposure_findings: {
    confirmed: boolean;
    finding: WalletExposureFinding;
    verified_exposure_count: number;
    distinct_public_keys: number;
  };
  evidence: ExposureEvidence[];
  quantum_threat: {
    shor_relevance: string;
    grover_relevance: string;
    practical_note: string;
    bullets: string[];
  };
  risk: RiskAssessment;
  coverage: {
    history_coverage: HistoryCoverageStatus;
    evidence_quality: EvidenceQuality;
    analyzed_tx_count: number;
    unsupported_tx_count: number;
    failed_tx_count: number;
    sender_tx_count: number;
    pagination_complete?: boolean;
  };
  recommended_next_steps: string[];
  limitations: string[];
  disclaimers: string[];
}

/** Client-side session index (backend has no list-scans endpoint). */
export interface SessionScanRecord {
  scanId: string;
  reportId?: string;
  address: string;
  status: ScanStatus;
  classification?: RiskClassification;
  createdAt: string;
  completedAt?: string;
}
