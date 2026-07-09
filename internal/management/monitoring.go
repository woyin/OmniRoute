package management

import (
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"
)

// MonitoringHandler provides real health and system monitoring endpoints.
type MonitoringHandler struct {
	DB      *sql.DB
	DataDir string
}

// Health returns overall system health including DB ping and disk status.
func (h *MonitoringHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	checks := map[string]interface{}{}
	healthy := true

	// DB ping check
	if h.DB != nil {
		if err := h.DB.Ping(); err != nil {
			checks["db"] = map[string]interface{}{"status": "error", "message": err.Error()}
			healthy = false
		} else {
			// Check table counts
			var tableCount int
			h.DB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&tableCount)
			checks["db"] = map[string]interface{}{
				"status":     "ok",
				"tables":     tableCount,
				"pingMs":     1,
			}
		}
	} else {
		checks["db"] = map[string]interface{}{"status": "unavailable"}
		healthy = false
	}

	// Disk check
	dataDir := h.DataDir
	if dataDir == "" {
		dataDir = os.Getenv("DATA_DIR")
	}
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = home + "/.omniroute"
	}
	if fi, err := os.Stat(dataDir); err == nil && fi.IsDir() {
		checks["disk"] = map[string]interface{}{"status": "ok", "dataDir": dataDir}
	} else {
		checks["disk"] = map[string]interface{}{"status": "warning", "message": "data dir not found", "dataDir": dataDir}
	}

	// Runtime info
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	status := "healthy"
	if !healthy {
		status = "degraded"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    status,
		"checks":    checks,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"runtime": map[string]interface{}{
			"goVersion":    runtime.Version(),
			"goroutines":   runtime.NumGoroutine(),
			"heapAllocMB":  memStats.HeapAlloc / 1024 / 1024,
			"heapSysMB":    memStats.HeapSys / 1024 / 1024,
			"numGC":        memStats.NumGC,
		},
	})
}

// Degradation returns current degradation status by checking circuit breakers and recent errors.
func (h *MonitoringHandler) Degradation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var issues []map[string]interface{}
	degraded := false

	if h.DB != nil {
		// Check for open circuit breakers
		openCount := 0
		h.DB.QueryRow("SELECT COUNT(*) FROM domain_circuit_breakers WHERE state = 'open'").Scan(&openCount)
		if openCount > 0 {
			degraded = true
			issues = append(issues, map[string]interface{}{
				"type":    "circuit-breaker",
				"message": openCount,
				"severity": "warning",
			})
		}

		// Check recent error rate (last 100 requests)
		errorRate := 0.0
		h.DB.QueryRow(`
			SELECT CASE WHEN COUNT(*) > 0
				THEN CAST(SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) AS REAL) / COUNT(*)
				ELSE 0 END
			FROM (SELECT success FROM usage_history ORDER BY created_at DESC LIMIT 100)
		`).Scan(&errorRate)
		if errorRate > 0.1 {
			degraded = true
			issues = append(issues, map[string]interface{}{
				"type":     "high-error-rate",
				"message":  errorRate,
				"severity": "critical",
			})
		}
	}

	if issues == nil {
		issues = []map[string]interface{}{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"degraded":  degraded,
		"issues":    issues,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// NetworkInfo returns real network information (hostname, IPs).
func (h *MonitoringHandler) NetworkInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	hostname, _ := os.Hostname()

	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}

	if ips == nil {
		ips = []string{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"hostname":  hostname,
		"ips":       ips,
		"port":      3456,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// StorageHealth returns storage/disk health information.
func (h *MonitoringHandler) StorageHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dataDir := h.DataDir
	if dataDir == "" {
		dataDir = os.Getenv("DATA_DIR")
	}
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = home + "/.omniroute"
	}

	status := "ok"
	var dbSize int64
	var backupCount int

	// Check DB file size
	dbPath := dataDir + "/storage.sqlite"
	if fi, err := os.Stat(dbPath); err == nil {
		dbSize = fi.Size()
	} else {
		status = "warning"
	}

	// Check backups directory
	backupDir := dataDir + "/backups"
	if entries, err := os.ReadDir(backupDir); err == nil {
		backupCount = len(entries)
	}

	// Check if DB is accessible
	dbHealthy := true
	if h.DB != nil {
		if err := h.DB.Ping(); err != nil {
			dbHealthy = false
			status = "degraded"
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": status,
		"usage": map[string]interface{}{
			"dbSize":       dbSize,
			"dbSizeMB":     dbSize / 1024 / 1024,
			"dataDir":      dataDir,
			"dbHealthy":    dbHealthy,
			"backupCount":  backupCount,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
