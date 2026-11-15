package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/quantumshield/qguard/backend/internal/domain"
)

func TestMemoryScanLifecycle(t *testing.T) {
	store := NewMemory(MemoryConfig{MaxScans: 10, MaxReports: 10})
	addr := common.HexToAddress("0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78")
	job := &domain.ScanJob{
		ID:        "scan-1",
		Address:   addr,
		ChainID:   1,
		Status:    domain.ScanQueued,
		CreatedAt: time.Now().UTC(),
	}
	if err := store.CreateScan(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	job.Status = domain.ScanCompleted
	risk := domain.RiskAssessment{Classification: domain.RiskLowerExposure}
	job.Risk = &risk
	if err := store.UpdateScan(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetScan(context.Background(), "scan-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.ScanCompleted {
		t.Fatalf("status = %s", got.Status)
	}
	latest, err := store.LatestScanByAddress(context.Background(), addr)
	if err != nil || latest.ID != "scan-1" {
		t.Fatalf("latest = %+v err=%v", latest, err)
	}

	rep := &domain.SecurityReport{ID: "rep-1", ScanID: "scan-1", ExecutiveSummary: "x"}
	if err := store.SaveReport(context.Background(), rep); err != nil {
		t.Fatal(err)
	}
	gotRep, err := store.GetReport(context.Background(), "rep-1")
	if err != nil || gotRep.ScanID != "scan-1" {
		t.Fatalf("report = %+v err=%v", gotRep, err)
	}
	byScan, err := store.GetReportByScanID(context.Background(), "scan-1")
	if err != nil || byScan.ID != "rep-1" {
		t.Fatalf("byScan = %+v err=%v", byScan, err)
	}
}

func TestMemoryCapacity(t *testing.T) {
	store := NewMemory(MemoryConfig{MaxScans: 1, MaxReports: 1})
	j1 := &domain.ScanJob{ID: "a", Address: common.HexToAddress("0x1"), CreatedAt: time.Now().UTC()}
	j2 := &domain.ScanJob{ID: "b", Address: common.HexToAddress("0x2"), CreatedAt: time.Now().UTC()}
	if err := store.CreateScan(context.Background(), j1); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateScan(context.Background(), j2); err != ErrCapacity {
		t.Fatalf("err = %v", err)
	}
}

func TestMemoryConcurrency(t *testing.T) {
	store := NewMemory(MemoryConfig{MaxScans: 500, MaxReports: 500})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			job := &domain.ScanJob{
				ID:        fmt.Sprintf("scan-%d", i),
				Address:   common.HexToAddress(fmt.Sprintf("0x%040x", i+1)),
				Status:    domain.ScanQueued,
				CreatedAt: time.Now().UTC(),
			}
			if err := store.CreateScan(context.Background(), job); err != nil {
				t.Errorf("create: %v", err)
				return
			}
			job.Status = domain.ScanCompleted
			if err := store.UpdateScan(context.Background(), job); err != nil {
				t.Errorf("update: %v", err)
				return
			}
			if _, err := store.GetScan(context.Background(), job.ID); err != nil {
				t.Errorf("get: %v", err)
			}
		}(i)
	}
	wg.Wait()
}

func TestMemoryTTLEviction(t *testing.T) {
	now := time.Unix(1000, 0).UTC()
	store := NewMemory(MemoryConfig{
		MaxScans:   10,
		MaxReports: 10,
		TTL:        10 * time.Second,
		Now:        func() time.Time { return now },
	})
	old := &domain.ScanJob{
		ID: "old", Address: common.HexToAddress("0x1"),
		CreatedAt: now.Add(-1 * time.Hour),
	}
	if err := store.CreateScan(context.Background(), old); err != nil {
		t.Fatal(err)
	}
	now = now.Add(1 * time.Hour)
	store.now = func() time.Time { return now }
	fresh := &domain.ScanJob{
		ID: "fresh", Address: common.HexToAddress("0x2"),
		CreatedAt: now,
	}
	if err := store.CreateScan(context.Background(), fresh); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetScan(context.Background(), "old"); err != ErrNotFound {
		t.Fatalf("expected old scan evicted, err=%v", err)
	}
}
