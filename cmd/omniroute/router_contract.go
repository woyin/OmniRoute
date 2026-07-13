package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
)

type RouteContract struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Source string `json:"source"`
	Auth   string `json:"auth"`
	Stream string `json:"stream"`
}

type RouteInventory struct {
	SchemaVersion int             `json:"schemaVersion"`
	Routes        []RouteContract `json:"routes"`
}

var (
	chiParameter        = regexp.MustCompile(`\{([^}]*)\}`)
	optionalChiCatchAll = regexp.MustCompile(`\{([^}]*(?:\.\.|:\.\*)[^}]*)\}\?`)
)

func normalizeGoRoutePath(path string) string {
	path = strings.ReplaceAll(path, "/*", "/{...}")
	path = optionalChiCatchAll.ReplaceAllString(path, "{...?}")
	return chiParameter.ReplaceAllStringFunc(path, func(parameter string) string {
		value := strings.TrimSuffix(strings.TrimPrefix(parameter, "{"), "}")
		optional := strings.HasSuffix(value, "?")
		value = strings.TrimSuffix(value, "?")
		catchAll := strings.HasSuffix(value, "...") || strings.Contains(value, ":.*")
		if catchAll && optional {
			return "{...?}"
		}
		if catchAll {
			return "{...}"
		}
		return "{}"
	})
}

func inventoryGoRoutes(routes chi.Routes, requireAPIKey bool) (RouteInventory, error) {
	inventory := RouteInventory{SchemaVersion: 1, Routes: []RouteContract{}}
	err := chi.Walk(routes, func(method, path string, _ http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if method == http.MethodOptions {
			return nil
		}
		inventory.Routes = append(inventory.Routes, RouteContract{
			Method: method,
			Path:   normalizeGoRoutePath(path),
			Source: "go:chi.Routes",
			Auth:   routeAuth(middlewares, requireAPIKey),
			Stream: routeStream(method, path),
		})
		return nil
	})
	if err != nil {
		return RouteInventory{}, err
	}
	sort.Slice(inventory.Routes, func(i, j int) bool {
		if inventory.Routes[i].Path == inventory.Routes[j].Path {
			return inventory.Routes[i].Method < inventory.Routes[j].Method
		}
		return inventory.Routes[i].Path < inventory.Routes[j].Path
	})
	if err := validateRouteContracts(inventory.Routes); err != nil {
		return RouteInventory{}, err
	}
	return inventory, nil
}

func routeAuth(middlewares []func(http.Handler) http.Handler, requireAPIKey bool) string {
	for _, middleware := range middlewares {
		name := runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name()
		switch {
		case strings.Contains(name, "RequireAPIKey"):
			return "required"
		case strings.Contains(name, "OptionalAPIKey"):
			if requireAPIKey {
				return "required"
			}
			return "optional"
		case strings.Contains(name, "LoginMiddleware"):
			return "required"
		}
	}
	return "none"
}

func routeStream(method, path string) string {
	switch {
	case path == "/api/mcp/sse", path == "/api/gamification/stream", path == "/api/gamification/notifications", method == http.MethodGet && path == "/api/mcp/stream":
		return "sse"
	case path == "/api/v1/ws", path == "/api/internal/codex-responses-ws", path == "/api/tools/traffic-inspector/ws":
		return "websocket"
	case path == "/api/v1/chat/completions", path == "/api/v1/responses", path == "/api/v1/responses/{path}", path == "/api/v1/messages":
		return "json+sse"
	default:
		return "json"
	}
}

func validateRouteContracts(routes []RouteContract) error {
	seen := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		if route.Method == "" || route.Path == "" || route.Auth == "" || route.Stream == "" {
			return fmt.Errorf("route contract missing metadata: %+v", route)
		}
		key := route.Method + " " + route.Path
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate route contract: %s", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func marshalGoRouteInventory(routes chi.Routes, requireAPIKey bool) ([]byte, error) {
	inventory, err := inventoryGoRoutes(routes, requireAPIKey)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(inventory, "", "  ")
}
