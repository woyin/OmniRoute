package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/db"
)

func managementBatchesHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if id := chi.URLParam(r, "id"); id != "" {
			batch, err := db.GetBatch(dbConn, id)
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to fetch batch")
				return
			}
			if batch == nil {
				jsonError(w, http.StatusNotFound, "Batch not found")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"batch": batch})
			return
		}
		limit := 100
		if raw := r.URL.Query().Get("limit"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 {
				jsonError(w, http.StatusBadRequest, "Invalid limit")
				return
			}
			if parsed > 1000 {
				parsed = 1000
			}
			limit = parsed
		}
		batches, err := db.ListBatches(dbConn, limit)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch batches")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"batches": batches})
	}
}
