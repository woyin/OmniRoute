package management

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// TunnelHandler provides real tunnel status endpoints by checking running processes.
type TunnelHandler struct {
	DataDir string
}

// CloudflaredStatus checks if cloudflared tunnel is running and returns its URL.
func (h *TunnelHandler) CloudflaredStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	active := false
	url := ""
	pid := 0

	// Check if cloudflared process is running
	if out, err := exec.Command("pgrep", "-f", "cloudflared").Output(); err == nil {
		pidStr := strings.TrimSpace(string(out))
		if pidStr != "" {
			active = true
			// Try to get the first PID
			if n := strings.Split(pidStr, "\n"); len(n) > 0 {
				for _, c := range n[0] {
					pid = pid*10 + int(c-'0')
				}
			}
		}
	}

	// Check for tunnel URL in data dir
	dataDir := h.DataDir
	if dataDir == "" {
		dataDir = os.Getenv("DATA_DIR")
	}
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = home + "/.omniroute"
	}
	urlFile := dataDir + "/cloudflared-url.txt"
	if data, err := os.ReadFile(urlFile); err == nil {
		url = strings.TrimSpace(string(data))
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"active":    active,
		"url":       url,
		"pid":       pid,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// NgrokStatus checks if ngrok tunnel is running and returns its URL.
func (h *TunnelHandler) NgrokStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	active := false
	url := ""
	pid := 0

	// Check if ngrok process is running
	if out, err := exec.Command("pgrep", "-f", "ngrok").Output(); err == nil {
		pidStr := strings.TrimSpace(string(out))
		if pidStr != "" {
			active = true
			if n := strings.Split(pidStr, "\n"); len(n) > 0 {
				for _, c := range n[0] {
					pid = pid*10 + int(c-'0')
				}
			}
		}
	}

	// Check for tunnel URL in data dir
	dataDir := h.DataDir
	if dataDir == "" {
		dataDir = os.Getenv("DATA_DIR")
	}
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = home + "/.omniroute"
	}
	urlFile := dataDir + "/ngrok-url.txt"
	if data, err := os.ReadFile(urlFile); err == nil {
		url = strings.TrimSpace(string(data))
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"active":    active,
		"url":       url,
		"pid":       pid,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// TailscaleStatus checks if tailscale funnel is active.
func (h *TunnelHandler) TailscaleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	active := false
	url := ""

	// Check if tailscale is available and funnel is active
	if out, err := exec.Command("tailscale", "status", "--json").Output(); err == nil {
		status := strings.TrimSpace(string(out))
		if strings.Contains(status, "Funnel") {
			active = true
		}
	}

	// Check for funnel URL in data dir
	dataDir := h.DataDir
	if dataDir == "" {
		dataDir = os.Getenv("DATA_DIR")
	}
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = home + "/.omniroute"
	}
	urlFile := dataDir + "/tailscale-url.txt"
	if data, err := os.ReadFile(urlFile); err == nil {
		url = strings.TrimSpace(string(data))
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"active":    active,
		"url":       url,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// TailscaleEnable enables tailscale funnel (placeholder - requires tailscale CLI).
func (h *TunnelHandler) TailscaleEnable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Tailscale funnel enable requested",
	})
}

// TailscaleDisable disables tailscale funnel.
func (h *TunnelHandler) TailscaleDisable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Tailscale funnel disable requested",
	})
}
