package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/quantumshield/qguard/backend/internal/domain"
	"github.com/quantumshield/qguard/backend/internal/repository"
)

func TestLatestScanByAddressIgnoresQueuedAndRunning(t *testing.T) {
	ctx := context.Background()
	store := repository.NewMemory(repository.MemoryConfig{MaxScans: 16, MaxReports: 16})
	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")

	completed := &domain.ScanJob{
		ID:        "scan-complete",
		Address:   addr,
		Status:    domain.ScanCompleted,
		CreatedAt: time.Now().UTC(),
	}
	if err := store.CreateScan(ctx, completed); err != nil {
		t.Fatal(err)
	}
	got, err := store.LatestScanByAddress(ctx, addr)
	if err != nil || got.ID != "scan-complete" {
		t.Fatalf("expected completed scan, got %+v err=%v", got, err)
	}

	queued := &domain.ScanJob{
		ID:        "scan-queued",
		Address:   addr,
		Status:    domain.ScanQueued,
		CreatedAt: time.Now().UTC(),
	}
	if err := store.CreateScan(ctx, queued); err != nil {
		t.Fatal(err)
	}
	got, err = store.LatestScanByAddress(ctx, addr)
	if err != nil || got.ID != "scan-complete" {
		t.Fatalf("queued must not hide completed; got %+v err=%v", got, err)
	}

	running := &domain.ScanJob{
		ID:        "scan-running",
		Address:   addr,
		Status:    domain.ScanRunning,
		CreatedAt: time.Now().UTC(),
	}
	if err := store.CreateScan(ctx, running); err != nil {
		t.Fatal(err)
	}
	_ = store.UpdateScan(ctx, running)
	got, err = store.LatestScanByAddress(ctx, addr)
	if err != nil || got.ID != "scan-complete" {
		t.Fatalf("running must not hide completed; got %+v err=%v", got, err)
	}

	newer := &domain.ScanJob{
		ID:        "scan-newer",
		Address:   addr,
		Status:    domain.ScanQueued,
		CreatedAt: time.Now().UTC(),
	}
	if err := store.CreateScan(ctx, newer); err != nil {
		t.Fatal(err)
	}
	newer.Status = domain.ScanCompleted
	if err := store.UpdateScan(ctx, newer); err != nil {
		t.Fatal(err)
	}
	got, err = store.LatestScanByAddress(ctx, addr)
	if err != nil || got.ID != "scan-newer" {
		t.Fatalf("expected newer completed scan; got %+v err=%v", got, err)
	}
}

func TestCapacity(t *testing.T) {
	ctx := context.Background()
	store := repository.NewMemory(repository.MemoryConfig{MaxScans: 1, MaxReports: 1, TTL: 0})
	addr := common.HexToAddress("0x2222222222222222222222222222222222222222")
	if err := store.CreateScan(ctx, &domain.ScanJob{ID: "a", Address: addr, Status: domain.ScanQueued, CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateScan(ctx, &domain.ScanJob{ID: "b", Address: addr, Status: domain.ScanQueued, CreatedAt: time.Now().UTC()}); err != repository.ErrCapacity {
		t.Fatalf("expected ErrCapacity, got %v", err)
	}
}
