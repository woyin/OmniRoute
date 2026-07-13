package db

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/omniroute/omniroute/internal/config"
)

func TestGetDBReturnsInitialErrorAfterOnce(t *testing.T) {
	once = sync.Once{}
	instance = nil
	instanceErr = nil
	t.Cleanup(func() {
		once = sync.Once{}
		instance = nil
		instanceErr = nil
	})

	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "missing", "storage.sqlite")

	firstDB, firstErr := GetDB(cfg)
	secondDB, secondErr := GetDB(cfg)
	if firstDB != nil || secondDB != nil {
		t.Fatalf("GetDB() connections = (%v, %v), want nil", firstDB, secondDB)
	}
	if firstErr == nil {
		t.Fatal("first GetDB() error = nil, want error")
	}
	if secondErr != firstErr {
		t.Fatalf("second GetDB() error = %v, want same error %v", secondErr, firstErr)
	}
}
