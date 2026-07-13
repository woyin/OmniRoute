package db

import (
	"path/filepath"
	"testing"

	"github.com/omniroute/omniroute/internal/config"
)

func TestFilesPersistence(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.Exec(`INSERT INTO files(id,bytes,created_at,filename,purpose,content) VALUES('file-one',3,1,'a','batch',?)`, []byte("abc")); err != nil {
		t.Fatal(err)
	}
	files, err := ListFiles(conn, 10)
	if err != nil || len(files) != 1 {
		t.Fatalf("files=%+v err=%v", files, err)
	}
	_, content, err := GetFileContent(conn, "file-one")
	if err != nil || string(content) != "abc" {
		t.Fatalf("content=%q err=%v", content, err)
	}
}
