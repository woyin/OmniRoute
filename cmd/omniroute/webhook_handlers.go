package main

import (
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/db"
)

func webhookDeliveriesHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if _, err := db.GetWebhook(dbConn, id); err == sql.ErrNoRows {
			jsonError(w, http.StatusNotFound, "Webhook not found")
			return
		} else if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch deliveries")
			return
		}
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit < 1 {
			limit = 20
		} else if limit > 100 {
			limit = 100
		}
		deliveries, err := db.ListWebhookDeliveries(dbConn, id, limit)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch deliveries")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"deliveries": deliveries})
	}
}

func webhookValidateURLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonError(w, http.StatusBadRequest, "Invalid JSON body")
			return
		}
		valid, reason := validateWebhookURL(body.URL)
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{"valid": valid}
		if reason != "" {
			response["reason"] = reason
		}
		_ = json.NewEncoder(w).Encode(response)
	}
}

func validateWebhookURL(raw string) (bool, string) {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return false, "invalid_url"
	}
	if u.User != nil {
		return false, "blocked_private"
	}
	host := strings.ToLower(strings.Trim(u.Hostname(), "[]"))
	if isMetadataHost(host) || (!privateWebhookURLsAllowed() && isPrivateHost(host)) {
		return false, "blocked_private"
	}
	return true, ""
}

func privateWebhookURLsAllowed() bool {
	return envTrue("OMNIROUTE_ALLOW_PRIVATE_PROVIDER_URLS") || envFalse("OUTBOUND_SSRF_GUARD_ENABLED")
}

func envTrue(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

func envFalse(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "0", "false", "no", "off":
		return true
	}
	return false
}

func isMetadataHost(host string) bool {
	return host == "metadata.google.internal" || host == "metadata.goog" || host == "169.254.169.254" ||
		host == "100.100.100.200" || host == "fd00:ec2::254" || strings.HasPrefix(host, "169.254.")
}

func isPrivateHost(host string) bool {
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && (ip.IsPrivate() || ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast())
}
