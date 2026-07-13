package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/db"
)

func managementFilesListHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		files, err := db.ListFiles(dbConn, limit)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch files")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"files": files})
	}
}

func managementFileContentHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, content, err := db.GetFileContent(dbConn, chi.URLParam(r, "id"))
		if err == sql.ErrNoRows {
			jsonError(w, http.StatusNotFound, "File not found")
			return
		}
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch file")
			return
		}
		if content == nil {
			jsonError(w, http.StatusNotFound, "File content not found")
			return
		}
		contentType := file.MimeType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		filename := strings.ReplaceAll(file.Filename, `"`, "")
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		_, _ = w.Write(content)
	}
}
