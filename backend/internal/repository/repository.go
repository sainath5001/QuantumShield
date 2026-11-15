// Package repository defines persistence interfaces and an in-memory MVP implementation.
// Data is lost on process restart. Interfaces allow a later durable backend without
// changing scan/risk packages.
package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/quantumshield/qguard/backend/internal/domain"
)

var (
	// ErrNotFound indicates a missing scan or report.
	ErrNotFound = errors.New("not found")
	// ErrCapacity indicates the in-memory store rejected a write due to capacity.
	ErrCapacity = errors.New("repository capacity exceeded")
)

// ScanStore persists scan jobs and generated reports.
type ScanStore interface {
	CreateScan(ctx context.Context, job *domain.ScanJob) error
	UpdateScan(ctx context.Context, job *domain.ScanJob) error
	GetScan(ctx context.Context, id string) (*domain.ScanJob, error)
	LatestScanByAddress(ctx context.Context, addr common.Address) (*domain.ScanJob, error)
	SaveReport(ctx context.Context, report *domain.SecurityReport) error
	GetReport(ctx context.Context, id string) (*domain.SecurityReport, error)
	GetReportByScanID(ctx context.Context, scanID string) (*domain.SecurityReport, error)
}

// MemoryConfig configures the ephemeral store.
type MemoryConfig struct {
	MaxScans   int
	MaxReports int
	TTL        time.Duration // 0 disables TTL eviction
	Now        func() time.Time
}

// Memory is a thread-safe in-memory ScanStore.
// WARNING: all data is lost when the application restarts.
type Memory struct {
	mu       sync.RWMutex
	scans    map[string]*domain.ScanJob
	reports  map[string]*domain.SecurityReport
	byAddr   map[common.Address]string // latest scan id
	byScan   map[string]string         // scan id -> report id
	maxScans int
	maxReps  int
	ttl      time.Duration
	now      func() time.Time
}

// NewMemory constructs an in-memory repository with sensible defaults.
func NewMemory(cfg MemoryConfig) *Memory {
	if cfg.MaxScans <= 0 {
		cfg.MaxScans = 256
	}
	if cfg.MaxReports <= 0 {
		cfg.MaxReports = 256
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Memory{
		scans:    make(map[string]*domain.ScanJob),
		reports:  make(map[string]*domain.SecurityReport),
		byAddr:   make(map[common.Address]string),
		byScan:   make(map[string]string),
		maxScans: cfg.MaxScans,
		maxReps:  cfg.MaxReports,
		ttl:      cfg.TTL,
		now:      cfg.Now,
	}
}

func (m *Memory) CreateScan(_ context.Context, job *domain.ScanJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.evictLocked()
	if len(m.scans) >= m.maxScans {
		return ErrCapacity
	}
	cp := cloneScan(job)
	m.scans[cp.ID] = cp
	m.byAddr[cp.Address] = cp.ID
	return nil
}

func (m *Memory) UpdateScan(_ context.Context, job *domain.ScanJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.scans[job.ID]; !ok {
		return ErrNotFound
	}
	cp := cloneScan(job)
	m.scans[cp.ID] = cp
	m.byAddr[cp.Address] = cp.ID
	return nil
}

func (m *Memory) GetScan(_ context.Context, id string) (*domain.ScanJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, ok := m.scans[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneScan(job), nil
}

func (m *Memory) LatestScanByAddress(_ context.Context, addr common.Address) (*domain.ScanJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.byAddr[addr]
	if !ok {
		return nil, ErrNotFound
	}
	job, ok := m.scans[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneScan(job), nil
}

func (m *Memory) SaveReport(_ context.Context, report *domain.SecurityReport) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.evictLocked()
	if len(m.reports) >= m.maxReps {
		return ErrCapacity
	}
	cp := cloneReport(report)
	m.reports[cp.ID] = cp
	m.byScan[cp.ScanID] = cp.ID
	return nil
}

func (m *Memory) GetReport(_ context.Context, id string) (*domain.SecurityReport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.reports[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneReport(r), nil
}

func (m *Memory) GetReportByScanID(_ context.Context, scanID string) (*domain.SecurityReport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.byScan[scanID]
	if !ok {
		return nil, ErrNotFound
	}
	r, ok := m.reports[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneReport(r), nil
}

func (m *Memory) evictLocked() {
	if m.ttl <= 0 {
		return
	}
	cutoff := m.now().Add(-m.ttl)
	for id, job := range m.scans {
		if job.CreatedAt.Before(cutoff) {
			delete(m.scans, id)
			if m.byAddr[job.Address] == id {
				delete(m.byAddr, job.Address)
			}
			if rid, ok := m.byScan[id]; ok {
				delete(m.reports, rid)
				delete(m.byScan, id)
			}
		}
	}
}

func cloneScan(j *domain.ScanJob) *domain.ScanJob {
	if j == nil {
		return nil
	}
	cp := *j
	if j.StartedAt != nil {
		t := *j.StartedAt
		cp.StartedAt = &t
	}
	if j.CompletedAt != nil {
		t := *j.CompletedAt
		cp.CompletedAt = &t
	}
	if j.Account != nil {
		a := *j.Account
		cp.Account = &a
	}
	if j.History != nil {
		h := *j.History
		h.Transactions = append([]domain.NormalizedTransaction{}, j.History.Transactions...)
		cp.History = &h
	}
	if j.Exposure != nil {
		e := *j.Exposure
		e.Evidence = append([]domain.ExposureEvidence{}, j.Exposure.Evidence...)
		e.Limitations = append([]string{}, j.Exposure.Limitations...)
		cp.Exposure = &e
	}
	if j.Risk != nil {
		r := *j.Risk
		r.Limitations = append([]string{}, j.Risk.Limitations...)
		r.Disclaimers = append([]string{}, j.Risk.Disclaimers...)
		r.QuantumThreatNotes = append([]string{}, j.Risk.QuantumThreatNotes...)
		r.RecommendedNextSteps = append([]string{}, j.Risk.RecommendedNextSteps...)
		cp.Risk = &r
	}
	return &cp
}

func cloneReport(r *domain.SecurityReport) *domain.SecurityReport {
	if r == nil {
		return nil
	}
	cp := *r
	cp.Evidence = append([]domain.ExposureEvidence{}, r.Evidence...)
	cp.RecommendedSteps = append([]string{}, r.RecommendedSteps...)
	cp.Limitations = append([]string{}, r.Limitations...)
	cp.Disclaimers = append([]string{}, r.Disclaimers...)
	cp.QuantumThreat.Bullets = append([]string{}, r.QuantumThreat.Bullets...)
	cp.Risk.Limitations = append([]string{}, r.Risk.Limitations...)
	cp.Risk.Disclaimers = append([]string{}, r.Risk.Disclaimers...)
	cp.Risk.QuantumThreatNotes = append([]string{}, r.Risk.QuantumThreatNotes...)
	cp.Risk.RecommendedNextSteps = append([]string{}, r.Risk.RecommendedNextSteps...)
	return &cp
}
