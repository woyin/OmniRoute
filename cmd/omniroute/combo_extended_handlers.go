package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/db"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

func modelComboMappingDetailHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			jsonError(w, 404, "Mapping not found")
			return
		}
		switch r.Method {
		case http.MethodGet:
			var model, comboID, created string
			var priority int
			err = dbConn.QueryRow(`SELECT model,combo_id,priority,created_at FROM model_combo_mappings WHERE id=?`, id).Scan(&model, &comboID, &priority, &created)
			if err == sql.ErrNoRows {
				jsonError(w, 404, "Mapping not found")
				return
			}
			if err != nil {
				jsonError(w, 500, "Failed to get mapping")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"mapping": map[string]interface{}{"id": id, "pattern": model, "comboId": comboID, "priority": priority, "enabled": true, "createdAt": created}})
		case http.MethodPut:
			var body struct {
				Pattern  *string `json:"pattern"`
				ComboID  *string `json:"comboId"`
				Priority *int    `json:"priority"`
			}
			if json.NewDecoder(r.Body).Decode(&body) != nil {
				jsonError(w, 400, "Invalid JSON body")
				return
			}
			res, err := dbConn.Exec(`UPDATE model_combo_mappings SET model=COALESCE(?,model),combo_id=COALESCE(?,combo_id),priority=COALESCE(?,priority) WHERE id=?`, body.Pattern, body.ComboID, body.Priority, id)
			if err != nil {
				jsonError(w, 500, "Failed to update mapping")
				return
			}
			n, _ := res.RowsAffected()
			if n == 0 {
				jsonError(w, 404, "Mapping not found")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"mapping": map[string]interface{}{"id": id}})
		case http.MethodDelete:
			res, err := dbConn.Exec(`DELETE FROM model_combo_mappings WHERE id=?`, id)
			if err != nil {
				jsonError(w, 500, "Failed to delete mapping")
				return
			}
			n, _ := res.RowsAffected()
			if n == 0 {
				jsonError(w, 404, "Mapping not found")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"success": true})
		}
	}
}

func comboBuilderOptionsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connections, err := db.ListProviderConnections(dbConn, "")
		if err != nil {
			jsonError(w, 500, "Failed to fetch combo builder options")
			return
		}
		providers := make([]map[string]interface{}, 0)
		for _, c := range connections {
			if c.IsActive {
				providers = append(providers, map[string]interface{}{"id": c.Provider, "connectionId": c.ID, "name": c.Name})
			}
		}
		models := make([]map[string]string, 0)
		for _, entry := range registry.List() {
			for _, m := range entry.Models {
				models = append(models, map[string]string{"id": m.ID, "name": m.Name, "provider": entry.ID})
			}
		}
		writeJSONResponse(w, map[string]interface{}{"providers": providers, "models": models, "strategies": []string{"priority", "weighted", "round-robin", "random", "fill-first"}})
	}
}

func comboReorderHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ComboIDs []string `json:"comboIds"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || len(body.ComboIDs) == 0 {
			jsonError(w, 400, "Invalid JSON body")
			return
		}
		tx, err := dbConn.Begin()
		if err != nil {
			jsonError(w, 500, "Failed to reorder combos")
			return
		}
		defer tx.Rollback()
		for i, id := range body.ComboIDs {
			if _, err := tx.Exec(`UPDATE combos SET priority=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`, len(body.ComboIDs)-i, id); err != nil {
				jsonError(w, 500, "Failed to reorder combos")
				return
			}
		}
		if err := tx.Commit(); err != nil {
			jsonError(w, 500, "Failed to reorder combos")
			return
		}
		combos, err := db.ListCombos(dbConn)
		if err != nil {
			jsonError(w, 500, "Failed to reorder combos")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"combos": combos})
	}
}
