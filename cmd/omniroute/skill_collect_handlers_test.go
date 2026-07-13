package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSkillCollectInstallPlansTargets(t *testing.T) {
	r := httptest.NewRecorder()
	skillCollectInstallHandler().ServeHTTP(r, httptest.NewRequest(http.MethodPost, "/api/skills/collect/install", bytes.NewBufferString(`{"repoName":"owner/code-review","targets":["codex","bad"]}`)))
	if r.Code != http.StatusOK || !bytes.Contains(r.Body.Bytes(), []byte(`software-development`)) || !bytes.Contains(r.Body.Bytes(), []byte(`"ok":false`)) {
		t.Fatalf("status=%d body=%s", r.Code, r.Body.String())
	}
}
