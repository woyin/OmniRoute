package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGamificationNotificationsSSE(t *testing.T) {
	handler := gamificationNotificationsHandler()
	missing := httptest.NewRecorder()
	handler.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/api/gamification/notifications", nil))
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing status=%d", missing.Code)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/gamification/notifications?apiKeyId=key-1", nil).WithContext(ctx))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "text/event-stream" || response.Body.String() != ": connected\n\n" {
		t.Fatalf("status=%d headers=%v body=%q", response.Code, response.Header(), response.Body.String())
	}
}
