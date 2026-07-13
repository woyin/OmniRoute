package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type lettaAuthFile struct {
	Version   int                               `json:"version"`
	Providers map[string]map[string]interface{} `json:"providers"`
}

func lettaSettingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		home, err := os.UserHomeDir()
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to resolve home directory")
			return
		}
		lettaDir := filepath.Join(home, ".letta")
		settingsPath := filepath.Join(lettaDir, "settings.json")
		authPath := filepath.Join(lettaDir, "lc-local-backend", "providers", "auth.json")
		backupPath := authPath + ".omniroute-backup"
		switch r.Method {
		case http.MethodGet:
			_, pathErr := exec.LookPath("letta")
			_, dirErr := os.Stat(lettaDir)
			installed := pathErr == nil || dirErr == nil
			if !installed {
				writeJSONResponse(w, map[string]interface{}{"installed": false, "config": nil, "message": "Letta CLI is not installed"})
				return
			}
			authFile, err := readLettaAuth(authPath)
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to read Letta config")
				return
			}
			settings, _ := readJSONObject(settingsPath)
			provider := authFile.Providers["lmstudio"]
			baseURL, _ := provider["base_url"].(string)
			writeJSONResponse(w, map[string]interface{}{"installed": true, "config": authFile, "hasOmniRoute": isOmniRouteURL(baseURL), "lmstudioConflict": baseURL != "" && !isOmniRouteURL(baseURL), "configPath": authPath, "letta": map[string]interface{}{"baseURL": nullable(baseURL)}, "backendMode": stringValue(settings["preferredBackendMode"], "api")})
		case http.MethodPost:
			var body struct {
				BaseURL   string `json:"baseUrl"`
				APIKey    string `json:"apiKey"`
				Overwrite bool   `json:"overwrite"`
			}
			if json.NewDecoder(r.Body).Decode(&body) != nil || strings.TrimSpace(body.BaseURL) == "" {
				jsonError(w, http.StatusBadRequest, "Invalid request")
				return
			}
			authFile, err := readLettaAuth(authPath)
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to read Letta config")
				return
			}
			existing := authFile.Providers["lmstudio"]
			existingURL, _ := existing["base_url"].(string)
			if existingURL != "" && !isOmniRouteURL(existingURL) && !body.Overwrite {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "lmstudio provider is already configured for " + existingURL + ". Overwriting will break your existing LM Studio connection. Apply again to overwrite.", "conflict": true, "existingBaseUrl": existingURL})
				return
			}
			if existingURL != "" && !isOmniRouteURL(existingURL) {
				raw, _ := json.Marshal(existing)
				if err := atomicWriteFile(backupPath, raw, 0600); err != nil {
					jsonError(w, http.StatusInternalServerError, "Failed to back up Letta config")
					return
				}
			}
			settings, _ := readJSONObject(settingsPath)
			settings["preferredBackendMode"] = "local"
			if raw, err := json.MarshalIndent(settings, "", "  "); err != nil || atomicWriteFile(settingsPath, raw, 0600) != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to write Letta settings")
				return
			}
			delete(authFile.Providers, "lc-omniroute")
			baseURL := strings.TrimRight(body.BaseURL, "/")
			if !strings.HasSuffix(baseURL, "/v1") {
				baseURL += "/v1"
			}
			now := time.Now().UTC().Format(time.RFC3339)
			created := now
			if value, ok := existing["created_at"].(string); ok {
				created = value
			}
			authFile.Providers["lmstudio"] = map[string]interface{}{"id": "local-provider-lmstudio", "name": "lmstudio", "provider_type": "lmstudio_openai", "provider_category": "byok", "auth": map[string]interface{}{"type": "api", "key": body.APIKey}, "base_url": baseURL, "created_at": created, "updated_at": now}
			if err := writeLettaAuth(authPath, authFile); err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to write Letta config")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"success": true, "message": "Settings applied. Restart Letta CLI, then use /model to select a OmniRoute model.", "needsRestart": true})
		case http.MethodDelete:
			authFile, err := readLettaAuth(authPath)
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to read Letta config")
				return
			}
			restored := false
			if _, exists := authFile.Providers["lmstudio"]; exists {
				if raw, err := os.ReadFile(backupPath); err == nil {
					var provider map[string]interface{}
					if json.Unmarshal(raw, &provider) == nil {
						authFile.Providers["lmstudio"] = provider
						restored = true
						_ = os.Remove(backupPath)
					}
				} else {
					delete(authFile.Providers, "lmstudio")
				}
			}
			delete(authFile.Providers, "lc-omniroute")
			if err := writeLettaAuth(authPath, authFile); err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to write Letta config")
				return
			}
			settings, _ := readJSONObject(settingsPath)
			if settings["preferredBackendMode"] == "local" {
				settings["preferredBackendMode"] = "api"
				raw, _ := json.MarshalIndent(settings, "", "  ")
				_ = atomicWriteFile(settingsPath, raw, 0600)
			}
			message := "OmniRoute config removed. Restart Letta CLI to take effect."
			if restored {
				message = "OmniRoute config removed. Your original LM Studio provider has been restored. Restart Letta CLI to take effect."
			}
			writeJSONResponse(w, map[string]interface{}{"success": true, "message": message, "needsRestart": true})
		}
	}
}

func readLettaAuth(path string) (lettaAuthFile, error) {
	result := lettaAuthFile{Version: 1, Providers: map[string]map[string]interface{}{}}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, err
	}
	if result.Providers == nil {
		result.Providers = map[string]map[string]interface{}{}
	}
	return result, nil
}
func writeLettaAuth(path string, value lettaAuthFile) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(path, raw, 0600)
}
func readJSONObject(path string) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(raw, &result)
	return result, err
}
func atomicWriteFile(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".omniroute-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(mode); err == nil {
		_, err = tmp.Write(data)
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(name, path)
}
func isOmniRouteURL(value string) bool {
	value = strings.ToLower(value)
	return strings.Contains(value, ":20128") || strings.Contains(value, ":3000") || strings.Contains(value, "omniroute")
}
func stringValue(value interface{}, fallback string) string {
	if text, ok := value.(string); ok && text != "" {
		return text
	}
	return fallback
}
