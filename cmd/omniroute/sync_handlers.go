package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/db"
)

func syncBundleHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			token = r.Header.Get("x-sync-token")
		}
		hash := sha256.Sum256([]byte(token))
		var id string
		err := dbConn.QueryRow(`SELECT id FROM sync_tokens WHERE token_hash=? AND revoked_at IS NULL`, hex.EncodeToString(hash[:])).Scan(&id)
		if err != nil {
			jsonError(w, 401, "Invalid sync token")
			return
		}
		settings, err := db.ListSettingsByNamespace(dbConn, "settings")
		if err != nil {
			jsonError(w, 500, "Failed to build sync bundle")
			return
		}
		providers, err := db.ListProviderConnections(dbConn, "")
		if err != nil {
			jsonError(w, 500, "Failed to build sync bundle")
			return
		}
		combos, err := db.ListCombos(dbConn)
		if err != nil {
			jsonError(w, 500, "Failed to build sync bundle")
			return
		}
		raw, _ := json.Marshal(map[string]interface{}{"settings": settings, "providers": providers, "combos": combos})
		versionRaw := sha256.Sum256(raw)
		version := hex.EncodeToString(versionRaw[:])
		_, _ = dbConn.Exec(`UPDATE sync_tokens SET last_used_at=CURRENT_TIMESTAMP WHERE id=?`, id)
		w.Header().Set("ETag", `"`+version+`"`)
		w.Header().Set("X-Config-Version", version)
		w.Header().Set("Cache-Control", "private, no-store")
		if strings.Contains(r.Header.Get("If-None-Match"), version) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		writeJSONResponse(w, map[string]interface{}{"version": version, "bundle": map[string]interface{}{"settings": settings, "providers": providers, "combos": combos}})
	}
}
func syncTokenDetailHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var name, apiKey, created string
		err := dbConn.QueryRow(`SELECT name,sync_api_key_id,created_at FROM sync_tokens WHERE id=? AND revoked_at IS NULL`, id).Scan(&name, &apiKey, &created)
		if err == sql.ErrNoRows {
			jsonError(w, 404, "Sync token not found")
			return
		}
		if err != nil {
			jsonError(w, 500, "Failed to revoke sync token")
			return
		}
		now := time.Now().UTC().Format(time.RFC3339)
		_, err = dbConn.Exec(`UPDATE sync_tokens SET revoked_at=? WHERE id=?`, now, id)
		if err != nil {
			jsonError(w, 500, "Failed to revoke sync token")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"message": "Sync token revoked successfully", "syncToken": map[string]interface{}{"id": id, "name": name, "syncApiKeyId": apiKey, "createdAt": created, "revokedAt": now}})
	}
}
