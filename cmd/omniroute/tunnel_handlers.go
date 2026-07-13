package main

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
)

func tailscaleCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path, err := exec.LookPath("tailscale")
		installed := err == nil
		running := false
		authenticated := false
		if installed {
			if out, err := exec.Command(path, "status", "--json").Output(); err == nil {
				running = true
				authenticated = !strings.Contains(string(out), `"BackendState":"NeedsLogin"`)
			}
		}
		writeJSONResponse(w, map[string]interface{}{"installed": installed, "running": running, "authenticated": authenticated, "commandPath": path})
	}
}
func tailscaleCommandHandler(args ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path, err := exec.LookPath("tailscale")
		if err != nil {
			jsonError(w, 503, "Tailscale is not installed")
			return
		}
		cmd := exec.Command(path, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			jsonError(w, 500, strings.TrimSpace(string(output)))
			return
		}
		writeJSONResponse(w, map[string]interface{}{"success": true, "output": strings.TrimSpace(string(output))})
	}
}
func tailscaleLoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			AuthKey     string `json:"authKey"`
			LoginServer string `json:"loginServer"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		args := []string{"up"}
		if body.AuthKey != "" {
			args = append(args, "--authkey", body.AuthKey)
		}
		if body.LoginServer != "" {
			args = append(args, "--login-server", body.LoginServer)
		}
		tailscaleCommandHandler(args...)(w, r)
	}
}
func tailscaleInstallHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := exec.LookPath("tailscale"); err == nil {
			writeJSONResponse(w, map[string]interface{}{"success": true, "message": "Tailscale already installed"})
			return
		}
		jsonError(w, 501, "Automatic Tailscale installation is unavailable on this platform")
	}
}
