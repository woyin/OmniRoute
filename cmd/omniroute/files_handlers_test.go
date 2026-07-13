package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
)

func TestManagementFilesListAndContent(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.Exec(`INSERT INTO files(id,bytes,created_at,filename,purpose,content,mime_type) VALUES('file-one',5,1,'hello.txt','batch',?,'text/plain')`, []byte("hello")); err != nil {
		t.Fatal(err)
	}

	list := httptest.NewRecorder()
	managementFilesListHandler(conn).ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/files", nil))
	if list.Code != http.StatusOK || !bytes.Contains(list.Body.Bytes(), []byte(`"file-one"`)) {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}

	router := chi.NewRouter()
	router.Get("/api/files/{id}/content", managementFileContentHandler(conn))
	content := httptest.NewRecorder()
	router.ServeHTTP(content, httptest.NewRequest(http.MethodGet, "/api/files/file-one/content", nil))
	if content.Code != http.StatusOK || content.Body.String() != "hello" || content.Header().Get("Content-Type") != "text/plain" {
		t.Fatalf("content status=%d headers=%v body=%s", content.Code, content.Header(), content.Body.String())
	}
}
