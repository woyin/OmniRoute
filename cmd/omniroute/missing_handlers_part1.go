package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/omniroute/omniroute/internal/db"
)

// --- Keys (API key management) handlers ---

func keysListHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keys, err := db.ListAPIKeys(dbConn)
		if err != nil {
			http.Error(w, `{"error":{"message":"Failed to list keys"}}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"object": "list", "data": keys})
	}
}

func keysDetailHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.Header().Set("Content-Type", "application/json")
		if dbConn != nil {
			var name, key, createdAt string
			err := dbConn.QueryRow("SELECT name, key, created_at FROM api_keys WHERE id = ?", id).Scan(&name, &key, &createdAt)
			if err == nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "name": name, "key": key, "createdAt": createdAt})
				return
			}
		}
		http.Error(w, `{"error":{"message":"Key not found"}}`, http.StatusNotFound)
	}
}

func keysUpdateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":{"message":"Invalid JSON"}}`, http.StatusBadRequest)
			return
		}
		if name, ok := body["name"].(string); ok && dbConn != nil {
			dbConn.Exec("UPDATE api_keys SET name = ? WHERE id = ?", name, id)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "updated": true})
	}
}

func keysDeleteHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if dbConn != nil {
			dbConn.Exec("DELETE FROM api_keys WHERE id = ?", id)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "deleted": true})
	}
}

func keysRevealHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.Header().Set("Content-Type", "application/json")
		if dbConn != nil {
			var key string
			err := dbConn.QueryRow("SELECT key FROM api_keys WHERE id = ?", id).Scan(&key)
			if err == nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "key": key})
				return
			}
		}
		http.Error(w, `{"error":{"message":"Key not found"}}`, http.StatusNotFound)
	}
}

func keysRegenerateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		newKey := "sk-or-" + uuid.New().String()
		if dbConn != nil {
			dbConn.Exec("UPDATE api_keys SET key = ? WHERE id = ?", newKey, id)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "key": newKey})
	}
}

func keysUsageLimitsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "limits": map[string]interface{}{}})
	}
}

func keysDevicesHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "devices": []interface{}{}})
	}
}

// --- Key Groups handlers ---

func keyGroupsListHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"object": "list", "data": []interface{}{}})
	}
}

func keyGroupsCreateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": uuid.New().String(), "created": true})
	}
}

func keyGroupsDetailHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "name": "default"})
	}
}

func keyGroupsKeysHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"object": "list", "data": []interface{}{}})
	}
}

func keyGroupsPermissionsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"permissions": []interface{}{}})
	}
}

// --- Relay tokens handlers ---

func relayTokensListHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type relayToken struct {
			ID       string `json:"id"`
			Token    string `json:"token"`
			Provider string `json:"provider"`
			Model    string `json:"model"`
			IsActive bool   `json:"isActive"`
		}
		var tokens []relayToken
		if dbConn != nil {
			rows, err := dbConn.Query("SELECT id, token, provider, model, is_active FROM relay_tokens ORDER BY created_at DESC")
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var t relayToken
					var active int
					if rows.Scan(&t.ID, &t.Token, &t.Provider, &t.Model, &active) == nil {
						t.IsActive = active == 1
						tokens = append(tokens, t)
					}
				}
			}
		}
		if tokens == nil {
			tokens = []relayToken{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"object": "list", "data": tokens})
	}
}

func relayTokenDetailHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.Header().Set("Content-Type", "application/json")
		if dbConn != nil {
			var token, provider, model string
			var isActive int
			err := dbConn.QueryRow("SELECT token, provider, model, is_active FROM relay_tokens WHERE id = ?", id).
				Scan(&token, &provider, &model, &isActive)
			if err == nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id": id, "token": token, "provider": provider, "model": model, "isActive": isActive == 1,
				})
				return
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	}
}

// --- Routing decision handlers ---

func routingDecisionDetailHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	}
}

// --- ACP (Agent Communication Protocol) handlers ---

func acpAgentsListHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"object": "list", "data": []interface{}{}})
	}
}

// --- Agent Skills handlers ---

func agentSkillsListHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"object": "list", "data": []interface{}{}})
	}
}

func agentSkillsCreateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": uuid.New().String(), "created": true})
	}
}

func agentSkillsDetailHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	}
}

func agentSkillsDeleteHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "deleted": true})
	}
}

func agentSkillsCoverageHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"coverage": 0.0})
	}
}

func agentSkillsGenerateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"generated": true, "skill": nil})
	}
}

func agentSkillsRawHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "raw": ""})
	}
}
