package db

import (
	"path/filepath"
	"testing"

	"github.com/omniroute/omniroute/internal/config"
)

func TestBatchReads(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.Exec(`INSERT INTO batches(id,endpoint,completion_window,status,input_file_id,created_at,metadata) VALUES('batch-one','/v1/responses','24h','completed','file-one',1,'{"key":"value"}')`); err != nil {
		t.Fatal(err)
	}
	batch, err := GetBatch(conn, "batch-one")
	if err != nil || batch == nil || batch.ID != "batch-one" {
		t.Fatalf("batch=%+v err=%v", batch, err)
	}
	batches, err := ListBatches(conn, 10)
	if err != nil || len(batches) != 1 {
		t.Fatalf("batches=%+v err=%v", batches, err)
	}
}
