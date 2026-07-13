package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type chaosProviderOverride struct {
	ProviderID string `json:"providerId"`
	ModelID    string `json:"modelId,omitempty"`
	Enabled    bool   `json:"enabled"`
}

type chaosConfig struct {
	Enabled           bool                    `json:"enabled"`
	DefaultMode       string                  `json:"defaultMode"`
	ProviderOverrides []chaosProviderOverride `json:"providerOverrides"`
	SystemPrompt      *string                 `json:"systemPrompt,omitempty"`
	TimeoutMS         int                     `json:"timeoutMs"`
	MaxTokens         int                     `json:"maxTokens"`
}

func defaultChaosConfig() chaosConfig {
	return chaosConfig{DefaultMode: "parallel", ProviderOverrides: []chaosProviderOverride{}, TimeoutMS: 120000, MaxTokens: 4096}
}

func chaosConfigHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config := defaultChaosConfig()
		switch r.Method {
		case http.MethodGet:
			stored, err := loadChaosConfig(dbConn)
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to load chaos config")
				return
			}
			config = stored
		case http.MethodPut:
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&config); err != nil || !validChaosConfig(config) {
				jsonError(w, http.StatusBadRequest, "Invalid chaos config")
				return
			}
			raw, _ := json.Marshal(config)
			if _, err := dbConn.Exec(`INSERT INTO key_value(namespace,key,value) VALUES('settings','chaosModeConfig',?) ON CONFLICT(namespace,key) DO UPDATE SET value=excluded.value`, string(raw)); err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to update chaos config")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"config": config, "message": "Chaos config updated"})
			return
		case http.MethodDelete:
			if _, err := dbConn.Exec("DELETE FROM key_value WHERE namespace='settings' AND key='chaosModeConfig'"); err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to reset chaos config")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"config": config, "message": "Chaos config reset to defaults"})
			return
		}
		writeJSONResponse(w, map[string]interface{}{"config": config})
	}
}

func loadChaosConfig(dbConn *sql.DB) (chaosConfig, error) {
	config := defaultChaosConfig()
	var raw string
	err := dbConn.QueryRow("SELECT value FROM key_value WHERE namespace='settings' AND key='chaosModeConfig'").Scan(&raw)
	if err == sql.ErrNoRows {
		return config, nil
	}
	if err != nil {
		return config, err
	}
	if json.Unmarshal([]byte(raw), &config) != nil || !validChaosConfig(config) {
		return defaultChaosConfig(), nil
	}
	return config, nil
}

func validChaosConfig(config chaosConfig) bool {
	if config.DefaultMode != "parallel" && config.DefaultMode != "collaborative" {
		return false
	}
	if config.TimeoutMS < 5000 || config.TimeoutMS > 600000 || config.MaxTokens < 256 || config.MaxTokens > 128000 || len(config.ProviderOverrides) > 200 {
		return false
	}
	if config.SystemPrompt != nil && len(*config.SystemPrompt) > 10000 {
		return false
	}
	for _, item := range config.ProviderOverrides {
		if item.ProviderID == "" {
			return false
		}
	}
	return true
}
