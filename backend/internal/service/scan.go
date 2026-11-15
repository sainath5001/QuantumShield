package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/quantumshield/qguard/backend/internal/blockchain"
	"github.com/quantumshield/qguard/backend/internal/domain"
	"github.com/quantumshield/qguard/backend/internal/exposure"
	"github.com/quantumshield/qguard/backend/internal/history"
	"github.com/quantumshield/qguard/backend/internal/report"
	"github.com/quantumshield/qguard/backend/internal/repository"
	"github.com/quantumshield/qguard/backend/internal/risk"
)

// ScanService orchestrates the read-only wallet scanning workflow.
type ScanService struct {
	Store    repository.ScanStore
	RPC      blockchain.RPCClient
	History  *history.Fetcher
	Exposure *exposure.Analyzer
	ChainID  int64
	MaxTxs   int
	Timeout  time.Duration
	Log      *slog.Logger
}

// StartScan validates the address and enqueues an asynchronous scan job.
func (s *ScanService) StartScan(ctx context.Context, rawAddress string) (*domain.ScanJob, error) {
	addr, err := blockchain.ParseAddress(rawAddress)
	if err != nil {
		return nil, err
	}
	if s.Store == nil {
		return nil, fmt.Errorf("scan store is not configured")
	}
	now := time.Now().UTC()
	job := &domain.ScanJob{
		ID:        uuid.NewString(),
		Address:   addr,
		ChainID:   s.ChainID,
		Status:    domain.ScanQueued,
		Stage:     domain.StageQueued,
		Progress:  domain.ScanProgress{Stage: domain.StageQueued, Message: "queued"},
		CreatedAt: now,
	}
	if err := s.Store.CreateScan(ctx, job); err != nil {
		return nil, err
	}
	go s.run(job.ID)
	return job, nil
}

// GetScan returns a scan job by ID.
func (s *ScanService) GetScan(ctx context.Context, id string) (*domain.ScanJob, error) {
	return s.Store.GetScan(ctx, id)
}

// GetReport returns a security report by ID.
func (s *ScanService) GetReport(ctx context.Context, id string) (*domain.SecurityReport, error) {
	return s.Store.GetReport(ctx, id)
}

// LatestRisk returns the latest completed risk assessment for an address, if any.
func (s *ScanService) LatestRisk(ctx context.Context, rawAddress string) (*domain.RiskAssessment, error) {
	addr, err := blockchain.ParseAddress(rawAddress)
	if err != nil {
		return nil, err
	}
	job, err := s.Store.LatestScanByAddress(ctx, addr)
	if err != nil {
		return nil, err
	}
	if job.Status != domain.ScanCompleted || job.Risk == nil {
		return nil, repository.ErrNotFound
	}
	cp := *job.Risk
	return &cp, nil
}

// LatestExposure returns exposure from the latest completed scan, or runs a live analysis via AccountService path.
func (s *ScanService) LatestExposure(ctx context.Context, rawAddress string) (*domain.WalletExposureReport, error) {
	addr, err := blockchain.ParseAddress(rawAddress)
	if err != nil {
		return nil, err
	}
	job, err := s.Store.LatestScanByAddress(ctx, addr)
	if err == nil && job.Status == domain.ScanCompleted && job.Exposure != nil {
		cp := *job.Exposure
		return &cp, nil
	}
	return nil, repository.ErrNotFound
}

func (s *ScanService) run(scanID string) {
	log := s.Log
	if log == nil {
		log = slog.Default()
	}
	timeout := s.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	job, err := s.Store.GetScan(ctx, scanID)
	if err != nil {
		log.Error("scan missing at start", "scan_id", scanID, "err", err)
		return
	}

	now := time.Now().UTC()
	job.Status = domain.ScanRunning
	job.StartedAt = &now
	job.Stage = domain.StageClassifyAccount
	job.Progress = domain.ScanProgress{Stage: job.Stage, Message: "classifying account"}
	_ = s.Store.UpdateScan(ctx, job)

	if err := s.execute(ctx, job); err != nil {
		fail := time.Now().UTC()
		job.Status = domain.ScanFailed
		job.Stage = domain.StageFailed
		job.CompletedAt = &fail
		job.Error = err.Error()
		job.Progress = domain.ScanProgress{Stage: domain.StageFailed, Message: err.Error()}
		_ = s.Store.UpdateScan(context.Background(), job)
		log.Warn("scan failed", "scan_id", scanID, "err", err)
		return
	}
}

func (s *ScanService) execute(ctx context.Context, job *domain.ScanJob) error {
	if s.RPC == nil {
		return fmt.Errorf("ethereum RPC client is not configured")
	}
	if s.History == nil {
		return fmt.Errorf("history client is not configured")
	}
	if s.Exposure == nil {
		return fmt.Errorf("exposure analyzer is not configured")
	}

	// 1. Classify account
	job.Stage = domain.StageClassifyAccount
	job.Progress = domain.ScanProgress{Stage: job.Stage, Message: "classifying account"}
	_ = s.Store.UpdateScan(ctx, job)

	snap, err := s.RPC.ClassifyAccount(ctx, job.Address)
	if err != nil {
		return fmt.Errorf("classify account: %w", err)
	}
	job.Account = &snap

	hist := domain.TransactionHistory{
		Address:  job.Address.Hex(),
		Coverage: domain.HistoryCoverage{Status: domain.CoverageUnknown, ChainID: job.ChainID, Provider: "etherscan"},
	}

	// 2. Fetch history (EOA path; contracts still may fetch for context but exposure skips recovery)
	job.Stage = domain.StageFetchHistory
	job.Progress = domain.ScanProgress{Stage: job.Stage, Message: "fetching transaction history"}
	_ = s.Store.UpdateScan(ctx, job)

	if snap.Kind != domain.AccountUnknown {
		fetched, ferr := s.History.Fetch(ctx, job.Address.Hex())
		if ferr != nil {
			// Soft-fail into unknown coverage if we got nothing usable.
			hist.Coverage.Status = domain.CoverageFailed
			hist.Coverage.APIErrors = append(hist.Coverage.APIErrors, domain.HistoryAPIError{
				Code: "HISTORY_FETCH", Message: ferr.Error(),
			})
		} else {
			hist = fetched
		}
	}
	job.History = &hist
	job.Progress.TxsFetched = len(hist.Transactions)
	_ = s.Store.UpdateScan(ctx, job)

	// 3. Exposure analysis
	job.Stage = domain.StageAnalyzeExposure
	job.Progress = domain.ScanProgress{
		Stage: job.Stage, Message: "analyzing public-key exposure",
		TxsFetched: len(hist.Transactions),
	}
	_ = s.Store.UpdateScan(ctx, job)

	expReport, err := s.Exposure.Analyze(ctx, exposure.AnalyzeInput{
		Address:         job.Address,
		AccountKind:     snap.Kind,
		History:         hist,
		MaxTransactions: s.MaxTxs,
	})
	if err != nil {
		return fmt.Errorf("exposure analysis: %w", err)
	}
	job.Exposure = &expReport
	job.Progress.TxsProcessed = expReport.AnalyzedTxCount
	_ = s.Store.UpdateScan(ctx, job)

	// 4. Risk assessment (pure)
	job.Stage = domain.StageAssessRisk
	job.Progress.Stage = job.Stage
	job.Progress.Message = "assessing risk"
	_ = s.Store.UpdateScan(ctx, job)

	assessment := risk.Assess(risk.Input{
		Account:  snap,
		History:  hist,
		Exposure: expReport,
		Now:      time.Now().UTC(),
	})
	job.Risk = &assessment

	// 5. Report
	job.Stage = domain.StageBuildReport
	job.Progress.Stage = job.Stage
	job.Progress.Message = "building report"
	_ = s.Store.UpdateScan(ctx, job)

	secReport := report.Build(report.BuildInput{
		ScanID:   job.ID,
		Account:  snap,
		History:  hist,
		Exposure: expReport,
		Risk:     assessment,
		Now:      time.Now().UTC(),
	})
	if err := s.Store.SaveReport(ctx, &secReport); err != nil {
		return fmt.Errorf("save report: %w", err)
	}
	job.ReportID = secReport.ID

	done := time.Now().UTC()
	job.Status = domain.ScanCompleted
	job.Stage = domain.StageCompleted
	job.CompletedAt = &done
	job.Progress = domain.ScanProgress{
		Stage:        domain.StageCompleted,
		TxsFetched:   len(hist.Transactions),
		TxsProcessed: expReport.AnalyzedTxCount,
		Message:      "completed",
	}
	if err := s.Store.UpdateScan(ctx, job); err != nil {
		return err
	}
	return nil
}
