package management

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// CacheHandler provides real DB-backed cache management endpoints.
type CacheHandler struct {
	DB *sql.DB
}

// Status returns overall cache status with real entry counts.
func (h *CacheHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	entries := 0
	var oldestEntry, newestEntry string
	if h.DB != nil {
		h.DB.QueryRow("SELECT COUNT(*) FROM semantic_cache").Scan(&entries)
		h.DB.QueryRow("SELECT COALESCE(MIN(created_at), '') FROM semantic_cache").Scan(&oldestEntry)
		h.DB.QueryRow("SELECT COALESCE(MAX(created_at), '') FROM semantic_cache").Scan(&newestEntry)
	}

	// Calculate hit rate from cache entries' hit_count
	var totalHits int
	if h.DB != nil {
		h.DB.QueryRow("SELECT COALESCE(SUM(hit_count), 0) FROM semantic_cache").Scan(&totalHits)
	}

	hitRate := 0.0
	if entries > 0 {
		hitRate = float64(totalHits) / float64(entries)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled":     true,
		"entries":     entries,
		"hitRate":     hitRate,
		"totalHits":   totalHits,
		"mode":        "semantic",
		"oldestEntry": oldestEntry,
		"newestEntry": newestEntry,
	})
}

// Stats returns detailed cache statistics.
func (h *CacheHandler) Stats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	entries := 0
	totalHits := 0
	var avgHits float64
	if h.DB != nil {
		h.DB.QueryRow("SELECT COUNT(*), COALESCE(SUM(hit_count), 0), COALESCE(AVG(hit_count), 0) FROM semantic_cache").
			Scan(&entries, &totalHits, &avgHits)
	}

	// Count expired entries
	expired := 0
	if h.DB != nil {
		h.DB.QueryRow("SELECT COUNT(*) FROM semantic_cache WHERE expires_at IS NOT NULL AND expires_at < datetime('now')").
			Scan(&expired)
	}

	// Top models by cache usage
	type modelStat struct {
		Model   string `json:"model"`
		Entries int    `json:"entries"`
		Hits    int    `json:"hits"`
	}
	var topModels []modelStat
	if h.DB != nil {
		rows, err := h.DB.Query(`
			SELECT COALESCE(model, 'unknown') as m, COUNT(*) as cnt, COALESCE(SUM(hit_count), 0) as hits
			FROM semantic_cache
			GROUP BY m
			ORDER BY cnt DESC
			LIMIT 10
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var ms modelStat
				if err := rows.Scan(&ms.Model, &ms.Entries, &ms.Hits); err == nil {
					topModels = append(topModels, ms)
				}
			}
		}
	}

	if topModels == nil {
		topModels = []modelStat{}
	}

	missRate := 1.0
	if totalHits+entries > 0 {
		missRate = float64(entries) / float64(totalHits+entries)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries":      entries,
		"totalHits":    totalHits,
		"avgHits":      avgHits,
		"expired":      expired,
		"hitRate":      1.0 - missRate,
		"missRate":     missRate,
		"topModels":    topModels,
	})
}

// Entries returns paginated cache entries.
func (h *CacheHandler) Entries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	type entry struct {
		ID        int     `json:"id"`
		KeyHash   string  `json:"keyHash"`
		Model     string  `json:"model"`
		Provider  string  `json:"provider"`
		HitCount  int     `json:"hitCount"`
		CreatedAt string  `json:"createdAt"`
		ExpiresAt *string `json:"expiresAt"`
	}

	var entries []entry
	total := 0
	if h.DB != nil {
		h.DB.QueryRow("SELECT COUNT(*) FROM semantic_cache").Scan(&total)

		rows, err := h.DB.Query(`
			SELECT id, key_hash, COALESCE(model, ''), COALESCE(provider, ''),
			       hit_count, created_at, expires_at
			FROM semantic_cache
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`, limit, offset)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var e entry
				if err := rows.Scan(&e.ID, &e.KeyHash, &e.Model, &e.Provider,
					&e.HitCount, &e.CreatedAt, &e.ExpiresAt); err == nil {
					entries = append(entries, e)
				}
			}
		}
	}

	if entries == nil {
		entries = []entry{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   entries,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Flush deletes all semantic cache entries.
func (h *CacheHandler) Flush(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	flushed := 0
	if h.DB != nil {
		res, err := h.DB.Exec("DELETE FROM semantic_cache")
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to flush cache")
			return
		}
		n, _ := res.RowsAffected()
		flushed = int(n)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"flushed": flushed,
	})
}

// Reasoning returns reasoning cache entries.
func (h *CacheHandler) Reasoning(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	sessionFilter := r.URL.Query().Get("session")

	type rcEntry struct {
		ID               string `json:"id"`
		SessionID        string `json:"sessionId"`
		Model            string `json:"model"`
		TurnIndex        int    `json:"turnIndex"`
		ContentLength    int    `json:"contentLength"`
		CreatedAt        string `json:"createdAt"`
	}

	var entries []rcEntry
	total := 0
	if h.DB != nil {
		countQuery := "SELECT COUNT(*) FROM reasoning_cache"
		dataQuery := `SELECT id, session_id, model, turn_index,
			LENGTH(reasoning_content), created_at FROM reasoning_cache`
		var args []interface{}

		if sessionFilter != "" {
			countQuery += " WHERE session_id = ?"
			dataQuery += " WHERE session_id = ?"
			args = append(args, sessionFilter)
		}

		h.DB.QueryRow(countQuery, args...).Scan(&total)

		dataQuery += " ORDER BY created_at DESC LIMIT ?"
		dataArgs := append(args, limit)

		rows, err := h.DB.Query(dataQuery, dataArgs...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var e rcEntry
				if err := rows.Scan(&e.ID, &e.SessionID, &e.Model, &e.TurnIndex,
					&e.ContentLength, &e.CreatedAt); err == nil {
					entries = append(entries, e)
				}
			}
		}
	}

	if entries == nil {
		entries = []rcEntry{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   entries,
		"total":  total,
	})
}

// FlushReasoning clears the reasoning cache.
func (h *CacheHandler) FlushReasoning(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	flushed := 0
	if h.DB != nil {
		res, err := h.DB.Exec("DELETE FROM reasoning_cache")
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to flush reasoning cache")
			return
		}
		n, _ := res.RowsAffected()
		flushed = int(n)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"flushed": flushed,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
