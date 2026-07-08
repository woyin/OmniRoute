package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

// OpencodeExecutor handles requests to OpenCode providers (zen, go).
type OpencodeExecutor struct {
	providerID string
	config     *config.Config
	base       BaseExecutor

	mu            sync.Mutex
	accounts      []opencodeAccount
	nextAccountIdx int
	requestFormat string
}

type opencodeAccount struct {
	Fingerprint      string
	CooldownUntil    time.Time
	ConsecutiveFails int
	Proxy            *opencodeProxy
}

type opencodeProxy struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

const (
	opencodeCooldownBaseMs = 5000
	opencodeCooldownMaxMs  = 60000
)

// NewOpencodeExecutor creates an OpencodeExecutor for the given provider.
func NewOpencodeExecutor(providerID string, cfg *config.Config) *OpencodeExecutor {
	return &OpencodeExecutor{
		providerID: providerID,
		config:     cfg,
		base:       BaseExecutor{ProviderID: providerID, Config: cfg},
		accounts:   []opencodeAccount{{Fingerprint: ""}},
	}
}

func (e *OpencodeExecutor) syncAccountsFromCredentials(credentials Credentials) {
	e.mu.Lock()
	defer e.mu.Unlock()

	psd := credentials.ProviderSpecificData
	if psd == nil {
		e.accounts = []opencodeAccount{{Fingerprint: ""}}
		e.nextAccountIdx = 0
		return
	}

	fingerprints, _ := psd["fingerprints"].([]interface{})
	if len(fingerprints) == 0 {
		e.accounts = []opencodeAccount{{Fingerprint: ""}}
		e.nextAccountIdx = 0
		return
	}

	previous := make(map[string]opencodeAccount)
	for _, a := range e.accounts {
		previous[a.Fingerprint] = a
	}

	var newAccounts []opencodeAccount
	for _, fpRaw := range fingerprints {
		fp, ok := fpRaw.(string)
		if !ok {
			continue
		}
		prior, existed := previous[fp]
		if existed {
			newAccounts = append(newAccounts, prior)
		} else {
			newAccounts = append(newAccounts, opencodeAccount{Fingerprint: fp})
		}
	}
	e.accounts = newAccounts
	if e.nextAccountIdx >= len(e.accounts) {
		e.nextAccountIdx = 0
	}
}

func (e *OpencodeExecutor) isAccountReady(account *opencodeAccount) bool {
	return time.Now().After(account.CooldownUntil)
}

func (e *OpencodeExecutor) pickAccount() *opencodeAccount {
	for i := 0; i < len(e.accounts); i++ {
		idx := (e.nextAccountIdx + i) % len(e.accounts)
		acct := &e.accounts[idx]
		if e.isAccountReady(acct) {
			e.nextAccountIdx = (idx + 1) % len(e.accounts)
			return acct
		}
	}
	idx := e.nextAccountIdx % len(e.accounts)
	e.nextAccountIdx = (e.nextAccountIdx + 1) % len(e.accounts)
	return &e.accounts[idx]
}

func (e *OpencodeExecutor) markCooldown(account *opencodeAccount) {
	account.ConsecutiveFails++
	backoff := opencodeCooldownBaseMs
	for i := 1; i < account.ConsecutiveFails; i++ {
		backoff *= 2
		if backoff > opencodeCooldownMaxMs {
			backoff = opencodeCooldownMaxMs
			break
		}
	}
	account.CooldownUntil = time.Now().Add(time.Duration(backoff) * time.Millisecond)
}

func (e *OpencodeExecutor) markSuccess(account *opencodeAccount) {
	account.ConsecutiveFails = 0
}

// Execute dispatches through account rotation with fallback.
func (e *OpencodeExecutor) Execute(ctx context.Context, input ExecuteInput) (*ExecuteResult, error) {
	e.syncAccountsFromCredentials(input.Credentials)
	e.requestFormat = resolveTargetFormat(e.providerID, input.Model)

	transformedBody := e.TransformRequest(input.Model, input.Body, input.Stream, input.Credentials)
	bodyJSON, err := json.Marshal(transformedBody)
	if err != nil {
		return nil, fmt.Errorf("marshal opencode body: %w", err)
	}

	// Fast path: single account, no proxies
	if len(e.accounts) == 1 && e.accounts[0].Proxy == nil {
		url := e.BuildURL(input.Model, input.Stream, input.Credentials)
		headers := e.BuildHeaders(input.Credentials, input.Stream, input.ClientHeaders)
		for k, v := range input.UpstreamExtraHeaders {
			headers[k] = v
		}

		resp, err := e.base.DoRequest(ctx, "POST", url, headers, bodyJSON, 3, input.SkipUpstreamRetry)
		if err != nil {
			return nil, err
		}
		return &ExecuteResult{Response: resp, URL: url, Headers: headers, TransformedBody: transformedBody}, nil
	}

	for attempt := 0; attempt < len(e.accounts); attempt++ {
		account := e.pickAccount()
		masked := "direct"
		if account.Fingerprint != "" && len(account.Fingerprint) > 8 {
			masked = account.Fingerprint[:8] + "..."
		}

		log.Printf("[OPENCODE] dispatch via account %s (idx %d/%d)", masked, attempt+1, len(e.accounts))

		url := e.BuildURL(input.Model, input.Stream, input.Credentials)
		headers := e.BuildHeaders(input.Credentials, input.Stream, input.ClientHeaders)

		resp, err := e.base.DoRequest(ctx, "POST", url, headers, bodyJSON, 1, true)
		if err != nil {
			e.markCooldown(account)
			continue
		}

		if resp.StatusCode == 429 {
			e.markCooldown(account)
			resp.Body.Close()
			log.Printf("[OPENCODE] Rate limited (429) on account %s, rotating...", masked)
			continue
		}

		e.markSuccess(account)
		return &ExecuteResult{Response: resp, URL: url, Headers: headers, TransformedBody: transformedBody}, nil
	}

	// All accounts failed — try direct
	url := e.BuildURL(input.Model, input.Stream, input.Credentials)
	headers := e.BuildHeaders(input.Credentials, input.Stream, input.ClientHeaders)
	resp, err := e.base.DoRequest(ctx, "POST", url, headers, bodyJSON, 3, false)
	if err != nil {
		return nil, err
	}
	return &ExecuteResult{Response: resp, URL: url, Headers: headers, TransformedBody: transformedBody}, nil
}

// BuildURL constructs the URL based on the request format.
func (e *OpencodeExecutor) BuildURL(model string, stream bool, credentials Credentials) string {
	entry := registry.Get(e.providerID)
	baseURL := "https://opencode.ai/zen/v1"
	if entry != nil && entry.BaseURL != "" {
		baseURL = entry.BaseURL
	}

	switch e.requestFormat {
	case "claude":
		return baseURL + "/messages"
	case "openai-responses":
		return baseURL + "/responses"
	case "gemini":
		suffix := "generateContent"
		if stream {
			suffix = "streamGenerateContent?alt=sse"
		}
		return fmt.Sprintf("%s/models/%s:%s", baseURL, model, suffix)
	default:
		return baseURL + "/chat/completions"
	}
}

// BuildHeaders constructs headers for the OpenCode provider.
func (e *OpencodeExecutor) BuildHeaders(credentials Credentials, stream bool, clientHeaders map[string]string) map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	key := credentials.APIKey
	if key == "" {
		key = credentials.AccessToken
	}
	if key != "" {
		if e.requestFormat == "claude" {
			headers["x-api-key"] = key
		} else {
			headers["Authorization"] = "Bearer " + key
		}
	}

	if e.requestFormat == "claude" {
		headers["anthropic-version"] = "2023-06-01"
	}

	if stream {
		headers["Accept"] = "text/event-stream"
	}

	if e.config.OpenCodeSynthesizeCliHeaders {
		headers["User-Agent"] = "opencode-cli/1.0.0"
		headers["x-opencode-client"] = "cli"
	}

	return headers
}

// TransformRequest modifies the request body for OpenCode-specific requirements.
func (e *OpencodeExecutor) TransformRequest(model string, body interface{}, stream bool, credentials Credentials) interface{} {
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		return body
	}
	delete(bodyMap, "client_metadata")
	if tools, ok := bodyMap["tools"].([]interface{}); ok && len(tools) > 128 {
		bodyMap["tools"] = tools[:128]
	}
	if parsed := parseDeepSeekEffortLevel(model); parsed != nil {
		bodyMap["model"] = parsed.BaseModel
		if _, hasEffort := bodyMap["reasoning_effort"]; !hasEffort {
			bodyMap["reasoning_effort"] = parsed.Effort
		}
	}
	return bodyMap
}

// RefreshCredentials returns nil for OpenCode (no OAuth).
func (e *OpencodeExecutor) RefreshCredentials(ctx context.Context, credentials Credentials) (*Credentials, error) {
	return nil, nil
}

func resolveTargetFormat(providerID, model string) string {
	entry := registry.Get(providerID)
	if entry != nil {
		m := entry.GetModel(model)
		if m != nil && m.TargetFormat != "" {
			return string(m.TargetFormat)
		}
	}
	return "openai"
}

type deepSeekEffort struct {
	BaseModel string
	Effort    string
}

func parseDeepSeekEffortLevel(model string) *deepSeekEffort {
	levels := []string{"low", "medium", "high", "max"}
	for _, level := range levels {
		suffix := "-" + level
		if strings.HasSuffix(model, suffix) {
			base := model[:len(model)-len(suffix)]
			if strings.EqualFold(base, "deepseek-v4-pro") {
				return &deepSeekEffort{BaseModel: base, Effort: level}
			}
		}
	}
	return nil
}
