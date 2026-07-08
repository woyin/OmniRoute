package oauth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type OAuthHandler struct{ DB *sql.DB }

func NewOAuthHandler(db *sql.DB) *OAuthHandler { return &OAuthHandler{DB: db} }

func (h *OAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/oauth/"), "/")
	if len(parts) < 2 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": map[string]interface{}{"message": "Expected /api/oauth/{provider}/{action}"}})
		return
	}
	provider, action := parts[0], parts[1]
	switch action {
	case "start":
		h.handleStart(w, r, provider)
	case "callback":
		h.handleCallback(w, r, provider)
	case "refresh":
		h.handleRefresh(w, r, provider)
	case "status":
		h.handleStatus(w, r, provider)
	case "disconnect":
		h.handleDisconnect(w, r, provider)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": map[string]interface{}{"message": fmt.Sprintf("Unknown action: %s", action)}})
	}
}

func (h *OAuthHandler) handleStart(w http.ResponseWriter, r *http.Request, provider string) {
	config := getOAuthConfig(provider)
	if config == nil {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{"error": map[string]interface{}{"message": fmt.Sprintf("No OAuth config for: %s", provider)}})
		return
	}
	state := uuid.New().String()
	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&state=%s&response_type=code&scope=%s",
		config.AuthURL, url.QueryEscape(config.ClientID), url.QueryEscape(config.RedirectURI), state, url.QueryEscape(config.Scope))
	if h.DB != nil {
		h.DB.Exec("INSERT INTO key_value (namespace, key, value) VALUES ('oauth_state', ?, ?) ON CONFLICT(namespace, key) DO UPDATE SET value=excluded.value", state, provider)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"authUrl": authURL, "state": state, "provider": provider})
}

func (h *OAuthHandler) handleCallback(w http.ResponseWriter, r *http.Request, provider string) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": map[string]interface{}{"message": fmt.Sprintf("OAuth error: %s", errParam)}})
		return
	}
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": map[string]interface{}{"message": "Missing authorization code"}})
		return
	}
	if h.DB != nil {
		var storedProvider string
		if err := h.DB.QueryRow("SELECT value FROM key_value WHERE namespace = 'oauth_state' AND key = ?", state).Scan(&storedProvider); err != nil || storedProvider != provider {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": map[string]interface{}{"message": "Invalid OAuth state"}})
			return
		}
		h.DB.Exec("DELETE FROM key_value WHERE namespace = 'oauth_state' AND key = ?", state)
	}
	config := getOAuthConfig(provider)
	if config == nil {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{"error": map[string]interface{}{"message": "No OAuth config"}})
		return
	}
	tokens, err := exchangeCode(config, code)
	if err != nil {
		log.Printf("[OAUTH] Token exchange failed for %s: %v", provider, err)
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"error": map[string]interface{}{"message": err.Error()}})
		return
	}
	if h.DB != nil {
		connID := uuid.New().String()
		h.DB.Exec("INSERT INTO provider_connections (id, provider, name, access_token, refresh_token, is_active, updated_at) VALUES (?, ?, ?, ?, ?, 1, ?) ON CONFLICT(id) DO UPDATE SET access_token=excluded.access_token, refresh_token=excluded.refresh_token, updated_at=excluded.updated_at",
			connID, provider, provider+" OAuth", tokens.AccessToken, tokens.RefreshToken, time.Now().UTC().Format(time.RFC3339))
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "provider": provider, "hasAccessToken": tokens.AccessToken != ""})
}

func (h *OAuthHandler) handleRefresh(w http.ResponseWriter, r *http.Request, provider string) {
	config := getOAuthConfig(provider)
	if config == nil {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{"error": map[string]interface{}{"message": "No OAuth config"}})
		return
	}
	var refreshToken string
	if h.DB != nil {
		h.DB.QueryRow("SELECT refresh_token FROM provider_connections WHERE provider = ? AND is_active = 1 AND refresh_token != '' LIMIT 1", provider).Scan(&refreshToken)
	}
	if refreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": map[string]interface{}{"message": "No refresh token"}})
		return
	}
	tokens, err := refreshTokenReq(config, refreshToken)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"error": map[string]interface{}{"message": err.Error()}})
		return
	}
	if h.DB != nil {
		h.DB.Exec("UPDATE provider_connections SET access_token = ?, refresh_token = COALESCE(?, refresh_token), updated_at = ? WHERE provider = ? AND is_active = 1",
			tokens.AccessToken, tokens.RefreshToken, time.Now().UTC().Format(time.RFC3339), provider)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "provider": provider})
}

func (h *OAuthHandler) handleStatus(w http.ResponseWriter, r *http.Request, provider string) {
	status := map[string]interface{}{"provider": provider, "hasOAuth": getOAuthConfig(provider) != nil, "connected": false, "hasToken": false}
	if h.DB != nil {
		var count int
		h.DB.QueryRow("SELECT COUNT(*) FROM provider_connections WHERE provider = ? AND is_active = 1 AND access_token != ''", provider).Scan(&count)
		if count > 0 { status["connected"] = true; status["hasToken"] = true }
	}
	writeJSON(w, http.StatusOK, status)
}

func (h *OAuthHandler) handleDisconnect(w http.ResponseWriter, r *http.Request, provider string) {
	if h.DB != nil {
		h.DB.Exec("UPDATE provider_connections SET access_token = '', refresh_token = '', is_active = 0, updated_at = ? WHERE provider = ?", time.Now().UTC().Format(time.RFC3339), provider)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "provider": provider})
}

func (h *OAuthHandler) HandlePasteCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed); return }
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/oauth/"), "/")
	provider := ""
	for i, p := range parts { if p == "paste-credentials" && i > 0 { provider = parts[i-1]; break } }
	if provider == "" { writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": map[string]interface{}{"message": "Provider required"}}); return }
	var body struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		APIKey       string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": map[string]interface{}{"message": "Invalid JSON"}}); return }
	if body.AccessToken == "" && body.APIKey == "" { writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": map[string]interface{}{"message": "accessToken or apiKey required"}}); return }
	if h.DB != nil {
		connID := uuid.New().String()
		h.DB.Exec("INSERT INTO provider_connections (id, provider, name, api_key, access_token, refresh_token, is_active, updated_at) VALUES (?, ?, ?, ?, ?, ?, 1, ?) ON CONFLICT(id) DO UPDATE SET api_key=excluded.api_key, access_token=excluded.access_token, refresh_token=excluded.refresh_token, updated_at=excluded.updated_at",
			connID, provider, provider+" (manual)", body.APIKey, body.AccessToken, body.RefreshToken, time.Now().UTC().Format(time.RFC3339))
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "provider": provider})
}

func (h *OAuthHandler) HandleCLIProxyImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed); return }
	var body struct {
		Provider    string `json:"provider"`
		AccessToken string `json:"accessToken"`
		APIKey      string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": map[string]interface{}{"message": "Invalid JSON"}}); return }
	if body.Provider == "" { writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": map[string]interface{}{"message": "provider required"}}); return }
	if h.DB != nil {
		connID := uuid.New().String()
		h.DB.Exec("INSERT INTO provider_connections (id, provider, name, api_key, access_token, is_active, updated_at) VALUES (?, ?, ?, ?, ?, 1, ?) ON CONFLICT(id) DO UPDATE SET api_key=excluded.api_key, access_token=excluded.access_token, updated_at=excluded.updated_at",
			connID, body.Provider, body.Provider+" (CLI)", body.APIKey, body.AccessToken, time.Now().UTC().Format(time.RFC3339))
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "provider": body.Provider})
}

// --- Internal types ---

type oauthConfig struct {
	ClientID, ClientSecret, AuthURL, TokenURL, RedirectURI, Scope string
}
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func getOAuthConfig(provider string) *oauthConfig {
	configs := map[string]*oauthConfig{
		// --- Existing providers ---
		"cursor":        {AuthURL: "https://cursor.com/oauth/authorize", TokenURL: "https://cursor.com/oauth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"kiro":          {AuthURL: "https://auth.kiro.dev/oauth/authorize", TokenURL: "https://auth.kiro.dev/oauth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"github-copilot": {AuthURL: "https://github.com/login/oauth/authorize", TokenURL: "https://github.com/login/oauth/access_token", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"windsurf":      {AuthURL: "https://codeium.com/oauth/authorize", TokenURL: "https://codeium.com/oauth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"claude-code":   {AuthURL: "https://console.anthropic.com/oauth/authorize", TokenURL: "https://console.anthropic.com/oauth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"antigravity":   {AuthURL: "https://api.antigravity.com/oauth/authorize", TokenURL: "https://api.antigravity.com/oauth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"codex":         {AuthURL: "https://auth.openai.com/authorize", TokenURL: "https://auth.openai.com/oauth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		// --- Ported from main branch ---
		"claude":        {AuthURL: "https://claude.ai/oauth/authorize", TokenURL: "https://api.anthropic.com/v1/oauth/token", Scope: "org:create_api_key user:profile user:inference user:sessions:claude_code user:mcp_servers", RedirectURI: "https://platform.claude.com/oauth/code/callback"},
		"openai":        {AuthURL: "https://auth.openai.com/oauth/authorize", TokenURL: "https://auth.openai.com/oauth/token", Scope: "openid profile email offline_access", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"github":        {AuthURL: "https://github.com/login/oauth/authorize", TokenURL: "https://github.com/login/oauth/access_token", Scope: "copilot", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"qwen":          {AuthURL: "https://qwen.ai/api/v1/oauth2/device/code", TokenURL: "https://qwen.ai/api/v1/oauth2/token", Scope: "openid profile email model.completion", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"grok-cli":      {AuthURL: "https://auth.x.ai/oauth2/authorize", TokenURL: "https://auth.x.ai/oauth2/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"kimi-coding":   {AuthURL: "https://auth.kimi.com/api/oauth/device_authorization", TokenURL: "https://auth.kimi.com/api/oauth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"codebuddy-cn":  {AuthURL: "https://copilot.tencent.com/v2/plugin/auth/authorize", TokenURL: "https://copilot.tencent.com/v2/plugin/auth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"cline":         {AuthURL: "https://api.cline.bot/api/v1/auth/authorize", TokenURL: "https://api.cline.bot/api/v1/auth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"agy":           {AuthURL: "https://accounts.google.com/o/oauth2/v2/auth", TokenURL: "https://oauth2.googleapis.com/token", Scope: "https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"gitlab-duo":    {AuthURL: "https://gitlab.com/oauth/authorize", TokenURL: "https://gitlab.com/oauth/token", Scope: "ai_features read_user", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"trae":          {AuthURL: "https://api.trae.ai/oauth/authorize", TokenURL: "https://api.trae.ai/oauth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"zed":           {AuthURL: "https://oauth.zed.dev/authorize", TokenURL: "https://oauth.zed.dev/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"zed-hosted":    {AuthURL: "https://cloud.zed.dev/oauth/authorize", TokenURL: "https://cloud.zed.dev/oauth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"kilocode":      {AuthURL: "https://auth.kilocode.ai/oauth/authorize", TokenURL: "https://auth.kilocode.ai/oauth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"devin-cli":     {AuthURL: "https://app.devin.ai/oauth/authorize", TokenURL: "https://app.devin.ai/oauth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
		"amazon-q":      {AuthURL: "https://auth.amazon.com/oauth/authorize", TokenURL: "https://auth.amazon.com/oauth/token", Scope: "openid profile email", RedirectURI: "http://localhost:3456/api/oauth/callback"},
	}
	cfg, ok := configs[provider]
	if !ok { return nil }
	envPrefix := strings.ToUpper(strings.ReplaceAll(provider, "-", "_"))
	if v := os.Getenv("OAUTH_" + envPrefix + "_CLIENT_ID"); v != "" { cfg.ClientID = v }
	if v := os.Getenv("OAUTH_" + envPrefix + "_CLIENT_SECRET"); v != "" { cfg.ClientSecret = v }
	return cfg
}

func exchangeCode(config *oauthConfig, code string) (*tokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code"); data.Set("code", code)
	data.Set("redirect_uri", config.RedirectURI); data.Set("client_id", config.ClientID); data.Set("client_secret", config.ClientSecret)
	resp, err := http.Post(config.TokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil { return nil, fmt.Errorf("request failed: %w", err) }
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("token exchange returned %d: %s", resp.StatusCode, body) }
	var tokens tokenResponse
	if err := json.Unmarshal(body, &tokens); err != nil { return nil, fmt.Errorf("parse response: %w", err) }
	return &tokens, nil
}

func refreshTokenReq(config *oauthConfig, refreshToken string) (*tokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token"); data.Set("refresh_token", refreshToken)
	data.Set("client_id", config.ClientID); data.Set("client_secret", config.ClientSecret)
	resp, err := http.Post(config.TokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil { return nil, fmt.Errorf("request failed: %w", err) }
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("refresh returned %d: %s", resp.StatusCode, body) }
	var tokens tokenResponse
	if err := json.Unmarshal(body, &tokens); err != nil { return nil, fmt.Errorf("parse response: %w", err) }
	return &tokens, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(status); json.NewEncoder(w).Encode(data)
}
