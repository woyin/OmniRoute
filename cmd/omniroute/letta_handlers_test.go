package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLettaSettingsApplyAndRemove(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".letta"), 0700); err != nil {
		t.Fatal(err)
	}
	handler := lettaSettingsHandler()
	apply := httptest.NewRecorder()
	handler.ServeHTTP(apply, httptest.NewRequest(http.MethodPost, "/api/cli-tools/letta-settings", bytes.NewBufferString(`{"baseUrl":"http://omniroute.local:3456","apiKey":"secret"}`)))
	if apply.Code != http.StatusOK || !bytes.Contains(apply.Body.Bytes(), []byte(`"success":true`)) {
		t.Fatalf("apply status=%d body=%s", apply.Code, apply.Body.String())
	}
	raw, err := os.ReadFile(filepath.Join(home, ".letta", "lc-local-backend", "providers", "auth.json"))
	if err != nil || !bytes.Contains(raw, []byte(`http://omniroute.local:3456/v1`)) {
		t.Fatalf("auth=%s err=%v", raw, err)
	}
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/cli-tools/letta-settings", nil))
	if get.Code != http.StatusOK || !bytes.Contains(get.Body.Bytes(), []byte(`"hasOmniRoute":true`)) {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	remove := httptest.NewRecorder()
	handler.ServeHTTP(remove, httptest.NewRequest(http.MethodDelete, "/api/cli-tools/letta-settings", nil))
	if remove.Code != http.StatusOK || !bytes.Contains(remove.Body.Bytes(), []byte(`"success":true`)) {
		t.Fatalf("remove status=%d body=%s", remove.Code, remove.Body.String())
	}
}
