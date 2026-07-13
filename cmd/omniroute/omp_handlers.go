package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func ompSettingsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		home, err := os.UserHomeDir()
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to resolve home directory")
			return
		}
		dir := filepath.Join(home, ".omp", "agent")
		dbPath := filepath.Join(dir, "agent.db")
		modelsPath := filepath.Join(dir, "models.yml")
		switch r.Method {
		case http.MethodGet:
			_, pathErr := exec.LookPath("omp")
			_, dbErr := os.Stat(dbPath)
			installed := pathErr == nil || dbErr == nil
			if !installed {
				writeJSONResponse(w, map[string]interface{}{"installed": false, "config": nil, "message": "Oh My Pi is not installed"})
				return
			}
			baseURL, apiKey, discovery, _ := readOMPProvider(modelsPath)
			if baseURL == "" {
				baseURL, apiKey = readOMPCredentials(dbConn)
			}
			writeJSONResponse(w, map[string]interface{}{"installed": true, "config": map[string]interface{}{"providers": map[string]interface{}{"omniroute": map[string]interface{}{"baseUrl": baseURL, "apiKey": apiKey, "discovery": nullable(discovery)}}}, "hasOmniRoute": baseURL != "" || apiKey != "", "configPath": modelsPath})
		case http.MethodPost:
			var body struct {
				BaseURL string `json:"baseUrl"`
				APIKey  string `json:"apiKey"`
			}
			if json.NewDecoder(r.Body).Decode(&body) != nil || strings.TrimSpace(body.BaseURL) == "" {
				jsonError(w, http.StatusBadRequest, "Invalid request")
				return
			}
			baseURL := strings.TrimRight(body.BaseURL, "/")
			if !strings.HasSuffix(baseURL, "/v1") {
				baseURL += "/v1"
			}
			key := body.APIKey
			if key == "" {
				key = "sk_omniroute"
			}
			if err := writeOMPProvider(modelsPath, baseURL, key); err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to write Oh My Pi settings")
				return
			}
			raw, _ := json.Marshal(map[string]string{"apiKey": key, "baseUrl": baseURL})
			if _, err := dbConn.Exec(`INSERT INTO key_value(namespace,key,value) VALUES('omp_credentials','omniroute',?) ON CONFLICT(namespace,key) DO UPDATE SET value=excluded.value`, string(raw)); err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to save Oh My Pi credentials")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"success": true, "message": "Oh My Pi settings applied! Run omp and all OmniRoute models appear under omniroute in /model.", "configPath": modelsPath})
		case http.MethodDelete:
			if err := removeOMPProvider(modelsPath); err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to update Oh My Pi settings")
				return
			}
			_, _ = dbConn.Exec("DELETE FROM key_value WHERE namespace='omp_credentials' AND key='omniroute'")
			writeJSONResponse(w, map[string]interface{}{"success": true, "message": "OmniRoute removed from Oh My Pi"})
		}
	}
}

func readOMPCredentials(dbConn *sql.DB) (string, string) {
	var raw string
	if dbConn.QueryRow("SELECT value FROM key_value WHERE namespace='omp_credentials' AND key='omniroute'").Scan(&raw) != nil {
		return "", ""
	}
	var value map[string]string
	_ = json.Unmarshal([]byte(raw), &value)
	return value["baseUrl"], value["apiKey"]
}

func readOMPProvider(path string) (baseURL, apiKey, discovery string, err error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", "", "", nil
	}
	if err != nil {
		return "", "", "", err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	in := false
	for scanner.Scan() {
		line := scanner.Text()
		trim := strings.TrimSpace(line)
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if trim == "omniroute:" && indent == 2 {
			in = true
			continue
		}
		if in && indent <= 2 && trim != "" {
			break
		}
		if !in {
			continue
		}
		parts := strings.SplitN(trim, ":", 2)
		if len(parts) != 2 {
			continue
		}
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		switch parts[0] {
		case "baseUrl":
			baseURL = value
		case "apiKey":
			apiKey = value
		case "type":
			discovery = value
		}
	}
	return baseURL, apiKey, discovery, scanner.Err()
}

func writeOMPProvider(path, baseURL, apiKey string) error {
	raw, _ := os.ReadFile(path)
	content := stripOMPProvider(string(raw))
	block := "  omniroute:\n    baseUrl: " + strconv.Quote(baseURL) + "\n    apiKey: " + strconv.Quote(apiKey) + "\n    api: openai-completions\n    authHeader: true\n    disableStrictTools: true\n    discovery:\n      type: proxy\n"
	lines := strings.Split(content, "\n")
	inserted := false
	for i, line := range lines {
		if strings.TrimSpace(line) == "providers:" && len(line)-len(strings.TrimLeft(line, " ")) == 0 {
			lines = append(lines[:i+1], append(strings.Split(strings.TrimSuffix(block, "\n"), "\n"), lines[i+1:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		content = strings.TrimRight(content, "\n")
		if content != "" {
			content += "\n"
		}
		content += "providers:\n" + block
		return atomicWriteFile(path, []byte(content), 0600)
	}
	return atomicWriteFile(path, []byte(strings.Join(lines, "\n")), 0600)
}
func removeOMPProvider(path string) error {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return atomicWriteFile(path, []byte(stripOMPProvider(string(raw))), 0600)
}
func stripOMPProvider(content string) string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	skip := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if trim == "omniroute:" && indent == 2 {
			skip = true
			continue
		}
		if skip && trim != "" && indent <= 2 {
			skip = false
		}
		if !skip {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}
