package management

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DBBackupsHandler provides real DB backup management endpoints.
type DBBackupsHandler struct {
	DB      *sql.DB
	DataDir string
	DBPath  string
}

// resolveDataDir returns the configured or default data directory.
func (h *DBBackupsHandler) resolveDataDir() string {
	if h.DataDir != "" {
		return h.DataDir
	}
	if d := os.Getenv("DATA_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return home + "/.omniroute"
}

// resolveDBPath returns the path to the SQLite database file.
func (h *DBBackupsHandler) resolveDBPath() string {
	if h.DBPath != "" {
		return h.DBPath
	}
	return h.resolveDataDir() + "/storage.sqlite"
}

// List returns all backup files found in the backups directory.
func (h *DBBackupsHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	backupDir := filepath.Join(h.resolveDataDir(), "backups")

	type backupFile struct {
		Name      string `json:"name"`
		Size      int64  `json:"size"`
		SizeMB    string `json:"sizeMB"`
		CreatedAt string `json:"createdAt"`
		Path      string `json:"path"`
	}

	var backups []backupFile

	if entries, err := os.ReadDir(backupDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			bf := backupFile{
				Name:      entry.Name(),
				Size:      info.Size(),
				SizeMB:    fmt.Sprintf("%.2f", float64(info.Size())/1024/1024),
				CreatedAt: info.ModTime().UTC().Format(time.RFC3339),
				Path:      filepath.Join(backupDir, entry.Name()),
			}
			backups = append(backups, bf)
		}
	}

	if backups == nil {
		backups = []backupFile{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"backups": backups,
		"total":   len(backups),
	})
}

// Export creates a backup of the SQLite database using the .backup command.
func (h *DBBackupsHandler) Export(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbPath := h.resolveDBPath()
	backupDir := filepath.Join(h.resolveDataDir(), "backups")

	// Ensure backups directory exists
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create backups directory")
		return
	}

	// Check source DB exists
	if _, err := os.Stat(dbPath); err != nil {
		writeJSONError(w, http.StatusNotFound, "database file not found: "+dbPath)
		return
	}

	// Generate backup filename
	timestamp := time.Now().UTC().Format("20060102-150405")
	backupName := fmt.Sprintf("backup-%s.sqlite", timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	// Use sqlite3 .backup command for a safe online backup
	cmd := exec.Command("sqlite3", dbPath, fmt.Sprintf(".backup '%s'", backupPath))
	if output, err := cmd.CombinedOutput(); err != nil {
		// Fallback: try file copy if sqlite3 CLI not available
		if copyErr := copyFile(dbPath, backupPath); copyErr != nil {
			writeJSONError(w, http.StatusInternalServerError,
				fmt.Sprintf("backup failed: %s, copy fallback: %s", string(output), copyErr.Error()))
			return
		}
	}

	// Get backup file info
	info, err := os.Stat(backupPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "backup created but unable to stat file")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"name":      backupName,
		"path":      backupPath,
		"size":      info.Size(),
		"sizeMB":    fmt.Sprintf("%.2f", float64(info.Size())/1024/1024),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ExportAll creates a full export of all database files in the data directory.
func (h *DBBackupsHandler) ExportAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dataDir := h.resolveDataDir()
	backupDir := filepath.Join(dataDir, "backups")
	timestamp := time.Now().UTC().Format("20060102-150405")
	exportDir := filepath.Join(backupDir, "export-"+timestamp)

	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create export directory")
		return
	}

	// Find all .sqlite files in data dir
	var exported []string
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to read data directory")
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".sqlite") || strings.HasSuffix(entry.Name(), ".db") {
			src := filepath.Join(dataDir, entry.Name())
			dst := filepath.Join(exportDir, entry.Name())
			if err := copyFile(src, dst); err == nil {
				exported = append(exported, entry.Name())
			}
		}
	}

	if exported == nil {
		exported = []string{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"exported":  exported,
		"count":     len(exported),
		"directory": exportDir,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Import restores a database from a backup file.
func (h *DBBackupsHandler) Import(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		BackupName string `json:"backupName"`
		BackupPath string `json:"backupPath"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	backupDir := filepath.Join(h.resolveDataDir(), "backups")
	backupPath := body.BackupPath
	if backupPath == "" && body.BackupName != "" {
		backupPath = filepath.Join(backupDir, body.BackupName)
	}

	if backupPath == "" {
		writeJSONError(w, http.StatusBadRequest, "backupName or backupPath is required")
		return
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); err != nil {
		writeJSONError(w, http.StatusNotFound, "backup file not found")
		return
	}

	dbPath := h.resolveDBPath()

	// Use sqlite3 to restore: backup from the import file into the live DB
	cmd := exec.Command("sqlite3", dbPath, fmt.Sprintf(".restore '%s'", backupPath))
	if output, err := cmd.CombinedOutput(); err != nil {
		// Fallback: file copy
		if copyErr := copyFile(backupPath, dbPath); copyErr != nil {
			writeJSONError(w, http.StatusInternalServerError,
				fmt.Sprintf("import failed: %s, copy fallback: %s", string(output), copyErr.Error()))
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "Database import completed",
		"source":    backupPath,
		"target":    dbPath,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}
	return os.WriteFile(dst, data, 0o644)
}
