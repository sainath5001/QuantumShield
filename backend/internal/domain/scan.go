package domain

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// ScanStatus is the lifecycle state of a scan job.
type ScanStatus string

const (
	ScanQueued    ScanStatus = "queued"
	ScanRunning   ScanStatus = "running"
	ScanCompleted ScanStatus = "completed"
	ScanFailed    ScanStatus = "failed"
)

// ScanStage names the current orchestration stage for progress reporting.
type ScanStage string

const (
	StageQueued          ScanStage = "queued"
	StageClassifyAccount ScanStage = "classify_account"
	StageFetchHistory    ScanStage = "fetch_history"
	StageAnalyzeExposure ScanStage = "analyze_exposure"
	StageAssessRisk      ScanStage = "assess_risk"
	StageBuildReport     ScanStage = "build_report"
	StageCompleted       ScanStage = "completed"
	StageFailed          ScanStage = "failed"
)

// ScanJob is a persisted (in-memory) wallet scan workflow.
type ScanJob struct {
	ID          string         `json:"id"`
	Address     common.Address `json:"address"`
	ChainID     int64          `json:"chain_id"`
	Status      ScanStatus     `json:"status"`
	Stage       ScanStage      `json:"stage"`
	Progress    ScanProgress   `json:"progress"`
	Error       string         `json:"error,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`

	Account  *AccountSnapshot      `json:"account,omitempty"`
	History  *TransactionHistory   `json:"history,omitempty"`
	Exposure *WalletExposureReport `json:"exposure,omitempty"`
	Risk     *RiskAssessment       `json:"risk,omitempty"`
	ReportID string                `json:"report_id,omitempty"`
}

// ScanProgress tracks coarse progress counters.
type ScanProgress struct {
	Stage        ScanStage `json:"stage"`
	TxsFetched   int       `json:"txs_fetched"`
	TxsProcessed int       `json:"txs_processed"`
	Message      string    `json:"message,omitempty"`
}
