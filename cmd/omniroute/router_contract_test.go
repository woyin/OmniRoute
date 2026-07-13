package main

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
)

func TestNormalizeGoRoutePath(t *testing.T) {
	cases := map[string]string{
		"/api/providers/{id}":   "/api/providers/{}",
		"/api/files/{rest...}":  "/api/files/{...}",
		"/api/files/{rest:.*}":  "/api/files/{...}",
		"/api/files/{rest:.*}?": "/api/files/{...?}",
	}
	for input, want := range cases {
		if got := normalizeGoRoutePath(input); got != want {
			t.Errorf("normalizeGoRoutePath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestValidateRouteContractsRejectsDuplicateAndMissingMetadata(t *testing.T) {
	valid := RouteContract{Method: "GET", Path: "/health", Auth: "none", Stream: "json"}
	for name, routes := range map[string][]RouteContract{
		"duplicate":      {valid, valid},
		"missing auth":   {{Method: "GET", Path: "/health", Stream: "json"}},
		"missing stream": {{Method: "GET", Path: "/health", Auth: "none"}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateRouteContracts(routes); err == nil {
				t.Fatal("validateRouteContracts() error = nil")
			}
		})
	}
}

func TestRouterInventoryStableAndComplete(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	dbConn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()

	router := buildRouter(cfg, dbConn)
	first, err := inventoryGoRoutes(router, cfg.RequireApiKey)
	if err != nil {
		t.Fatal(err)
	}
	second, err := inventoryGoRoutes(router, cfg.RequireApiKey)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatal("inventoryGoRoutes() output is unstable")
	}
	if len(first.Routes) < 680 {
		t.Fatalf("route count = %d, want at least 680", len(first.Routes))
	}
	wantMetadata := map[string][2]string{
		"GET /health":          {"none", "json"},
		"GET /api/providers":   {"required", "json"},
		"GET /api/mcp/sse":     {"none", "sse"},
		"GET /api/mcp/stream":  {"none", "sse"},
		"POST /api/mcp/stream": {"none", "json"},
		"GET /api/v1/ws":       {"optional", "websocket"},
	}
	for _, route := range first.Routes {
		key := route.Method + " " + route.Path
		if want, ok := wantMetadata[key]; ok {
			if route.Auth != want[0] || route.Stream != want[1] {
				t.Errorf("%s metadata = %s/%s, want %s/%s", key, route.Auth, route.Stream, want[0], want[1])
			}
			delete(wantMetadata, key)
		}
		if route.Method == "OPTIONS" {
			t.Fatalf("inventory includes OPTIONS route: %+v", route)
		}
		if route.Auth == "" || route.Stream == "" {
			t.Fatalf("route missing metadata: %+v", route)
		}
	}
	if len(wantMetadata) != 0 {
		t.Fatalf("routes missing from inventory: %v", wantMetadata)
	}
}
