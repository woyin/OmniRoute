package db

import (
	"path/filepath"
	"testing"

	"github.com/omniroute/omniroute/internal/config"
)

func TestPlaygroundPresetCRUD(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	created, err := CreatePlaygroundPreset(conn, PlaygroundPreset{Name: "Demo", Endpoint: "/api/v1/chat/completions", Model: "gpt-5", Params: map[string]interface{}{"temperature": .2}})
	if err != nil || created.ID == "" {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	updated, err := UpdatePlaygroundPreset(conn, created.ID, map[string]interface{}{"name": "Updated"})
	if err != nil || updated == nil || updated.Name != "Updated" {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	presets, err := ListPlaygroundPresets(conn)
	if err != nil || len(presets) != 1 {
		t.Fatalf("presets=%+v err=%v", presets, err)
	}
	deleted, err := DeletePlaygroundPreset(conn, created.ID)
	if err != nil || !deleted {
		t.Fatalf("deleted=%v err=%v", deleted, err)
	}
}
