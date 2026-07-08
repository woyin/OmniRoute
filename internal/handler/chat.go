package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/omniroute/omniroute/internal/auth"
	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
	"github.com/omniroute/omniroute/internal/provider/executor"
	"github.com/omniroute/omniroute/internal/provider/registry"
	"github.com/omniroute/omniroute/internal/provider/translator"
	"github.com/omniroute/omniroute/internal/routing"
	"github.com/omniroute/omniroute/internal/sse"
)

// ChatHandler handles /api/v1/chat/completions requests.
type ChatHandler struct {
	DB     *sql.DB
	Config *config.Config
}

// ServeHTTP processes a chat completions request.
func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Content-Type guard
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(strings.Split(contentType, ";")[0]), "application/json") {
		writeJSONError(w, "Content-Type must be application/json", "invalid_request_error", http.StatusUnsupportedMediaType)
		return
	}

	// Parse request body
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, "Invalid JSON body", "invalid_request_error", http.StatusBadRequest)
		return
	}

	model, _ := body["model"].(string)
	if model == "" {
		writeJSONError(w, "model is required", "invalid_request_error", http.StatusBadRequest)
		return
	}

	stream := false
	if s, ok := body["stream"].(bool); ok {
		stream = s
	}

	requestID := uuid.New().String()
	log.Printf("[CHAT] request=%s model=%s stream=%v", requestID[:8], model, stream)

	// Prompt injection guard
	if result := checkPromptInjection(body); result != nil && result.Blocked {
		log.Printf("[GUARD] request=%s blocked: %s (pattern: %s)", requestID[:8], result.Reason, result.Pattern)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Request blocked: potential prompt injection detected",
				"type":    "invalid_request_error",
				"code":    "prompt_injection_detected",
			},
		})
		return
	}

	// Resolve provider and combo routing
	providerID, resolvedModel := resolveProvider(model)
	credentials := resolveCredentials(r, h.DB, providerID)

	// Check if we should use combo routing
	if comboTargets := resolveComboTargets(h.DB, model); len(comboTargets) > 0 {
		h.handleComboChat(w, r, body, comboTargets, stream, requestID)
		return
	}

	// Single provider request
	h.handleSingleProvider(w, r, body, providerID, resolvedModel, credentials, stream, requestID)
}

// isResponsesAPI checks if the request path indicates a Responses API call.
func isResponsesAPI(r *http.Request) bool {
	return strings.Contains(r.URL.Path, "/responses")
}

// handleSingleProvider processes a request to a single provider.
func (h *ChatHandler) handleSingleProvider(w http.ResponseWriter, r *http.Request, body map[string]interface{}, providerID, model string, credentials executor.Credentials, stream bool, requestID string) {
	entry := registry.Get(providerID)

	// Circuit breaker check
	if h.DB != nil && db.IsProviderCircuitOpen(h.DB, providerID) {
		log.Printf("[CHAT] request=%s circuit open for provider %s, skipping", requestID[:8], providerID)
		writeJSONErrorf(w, http.StatusServiceUnavailable, "Provider %s circuit breaker is open", providerID)
		return
	}

	// Determine source and target formats
	sourceFormat := translator.FormatOpenAI
	targetFormat := translator.FormatOpenAI
	if entry != nil {
		targetFormat = string(entry.Format)
		if m := entry.GetModel(model); m != nil && m.TargetFormat != "" {
			targetFormat = string(m.TargetFormat)
		}
	}

	// Translate request if needed
	translatedBody := body
	if translator.NeedsTranslation(sourceFormat, targetFormat) {
		translatedBody = translator.TranslateRequest(body, sourceFormat, targetFormat)
	}

	// Get the executor
	exec := executor.GetExecutor(providerID, h.Config)

	start := time.Now()

	// Execute the request
	ctx := r.Context()
	result, err := exec.Execute(ctx, executor.ExecuteInput{
		Model:        model,
		Body:         translatedBody,
		Stream:       stream,
		Credentials:  credentials,
		ClientHeaders: extractClientHeaders(r),
	})
	if err != nil {
		latencyMs := int(time.Since(start).Milliseconds())
		log.Printf("[CHAT] request=%s executor error: %v (%dms)", requestID[:8], err, latencyMs)
		h.recordCall(providerID, model, 0, latencyMs, requestID, err.Error())
		if h.DB != nil {
			db.RecordProviderFailure(h.DB, providerID)
		}
		writeJSONError(w, "Upstream request failed", "upstream_error", http.StatusBadGateway)
		return
	}

	latencyMs := int(time.Since(start).Milliseconds())
	statusCode := result.Response.StatusCode
	h.recordCall(providerID, model, statusCode, latencyMs, requestID, "")
	if h.DB != nil && statusCode < 400 {
		db.RecordProviderSuccess(h.DB, providerID)
	}

	// Handle the response — if Responses API route, transform output
	if isResponsesAPI(r) {
		h.handleResponsesAPIResponse(w, r, result, model, sourceFormat, targetFormat, stream, requestID)
	} else if stream {
		h.handleStreamingResponse(w, r, result, model, sourceFormat, targetFormat)
	} else {
		h.handleNonStreamingResponse(w, r, result, model, sourceFormat, targetFormat)
	}
}

// handleComboChat processes a combo-routed request with fallback.
func (h *ChatHandler) handleComboChat(w http.ResponseWriter, r *http.Request, body map[string]interface{}, targets []routing.ResolvedTarget, stream bool, requestID string) {
	var lastErr error
	for _, target := range targets {
		credentials := executor.Credentials{
			APIKey:      target.APIKey,
			AccessToken: target.AccessToken,
		}

		exec := executor.GetExecutor(target.Provider, h.Config)
		ctx := r.Context()

		start := time.Now()
		result, err := exec.Execute(ctx, executor.ExecuteInput{
			Model:       target.Model,
			Body:        body,
			Stream:      stream,
			Credentials: credentials,
		})
		if err != nil {
			lastErr = err
			latencyMs := int(time.Since(start).Milliseconds())
			h.recordCall(target.Provider, target.Model, 0, latencyMs, requestID, err.Error())
			log.Printf("[COMBO] request=%s target %s/%s failed: %v", requestID[:8], target.Provider, target.Model, err)
			continue
		}

		if result.Response.StatusCode >= 400 {
			lastErr = fmt.Errorf("upstream returned %d", result.Response.StatusCode)
			latencyMs := int(time.Since(start).Milliseconds())
			h.recordCall(target.Provider, target.Model, result.Response.StatusCode, latencyMs, requestID, "")
			log.Printf("[COMBO] request=%s target %s/%s returned %d", requestID[:8], target.Provider, target.Model, result.Response.StatusCode)
			result.Response.Body.Close()
			continue
		}

		// Success — record usage
		latencyMs := int(time.Since(start).Milliseconds())
		h.recordCall(target.Provider, target.Model, result.Response.StatusCode, latencyMs, requestID, "")

		if stream {
			h.handleStreamingResponse(w, r, result, target.Model, translator.FormatOpenAI, translator.FormatOpenAI)
		} else {
			h.handleNonStreamingResponse(w, r, result, target.Model, translator.FormatOpenAI, translator.FormatOpenAI)
		}
		return
	}

	// All targets failed
	log.Printf("[COMBO] request=%s all targets failed: %v", requestID[:8], lastErr)
	writeJSONError(w, "All combo targets failed", "upstream_error", http.StatusBadGateway)
}

// handleStreamingResponse forwards a streaming SSE response to the client.
func (h *ChatHandler) handleStreamingResponse(w http.ResponseWriter, r *http.Request, result *executor.ExecuteResult, model, sourceFormat, targetFormat string) {
	upstream := result.Response
	defer upstream.Body.Close()

	sseWriter := sse.NewWriter(w)
	sseWriter.WriteHeader()

	// Start heartbeat
	stopCh := make(chan struct{})
	defer close(stopCh)
	sse.StartHeartbeat(sseWriter, h.Config.SSEHeartbeatInterval(), stopCh)

	// Forward the stream
	if sourceFormat == targetFormat {
		sse.StreamFromUpstream(sseWriter, upstream, stopCh)
	} else {
		translateSSEStream(sseWriter, upstream, model, sourceFormat, targetFormat, stopCh)
	}
}

// handleNonStreamingResponse forwards a non-streaming response to the client.
func (h *ChatHandler) handleNonStreamingResponse(w http.ResponseWriter, r *http.Request, result *executor.ExecuteResult, model, sourceFormat, targetFormat string) {
	upstream := result.Response
	defer upstream.Body.Close()

	body, err := io.ReadAll(upstream.Body)
	if err != nil {
		writeJSONError(w, "Failed to read upstream response", "upstream_error", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(upstream.StatusCode)
	w.Write(body)
}

// recordCall logs a call to the call_logs and usage_history tables.
func (h *ChatHandler) recordCall(provider, model string, statusCode, latencyMs int, requestID, errMsg string) {
	if h.DB == nil {
		return
	}
	go db.RecordCallLog(h.DB, db.CallLog{
		Provider:     provider,
		Model:        model,
		StatusCode:   statusCode,
		LatencyMs:    latencyMs,
		RequestID:    requestID,
		ErrorMessage: errMsg,
	})
	if statusCode >= 200 && statusCode < 300 && errMsg == "" {
		go db.RecordUsage(h.DB, db.UsageEntry{
			Provider:  provider,
			Model:     model,
			LatencyMs: latencyMs,
			Success:   true,
		})
	}
}

// handleResponsesAPIResponse transforms Chat Completions output to Responses API format.
func (h *ChatHandler) handleResponsesAPIResponse(w http.ResponseWriter, r *http.Request, result *executor.ExecuteResult, model, sourceFormat, targetFormat string, stream bool, requestID string) {
	upstream := result.Response

	if stream {
		defer upstream.Body.Close()
		sseWriter := sse.NewWriter(w)
		sseWriter.WriteHeader()

		stopCh := make(chan struct{})
		defer close(stopCh)
		sse.StartHeartbeat(sseWriter, h.Config.SSEHeartbeatInterval(), stopCh)

		if sourceFormat != targetFormat {
			translateSSEStream(sseWriter, upstream, model, sourceFormat, targetFormat, stopCh)
		} else {
			sse.TransformChatToResponsesStream(sseWriter, upstream.Body, requestID, model, stopCh)
		}
	} else {
		defer upstream.Body.Close()
		respBody, err := io.ReadAll(upstream.Body)
		if err != nil {
			writeJSONError(w, "Failed to read upstream response", "upstream_error", http.StatusBadGateway)
			return
		}

		var chatResp map[string]interface{}
		if json.Unmarshal(respBody, &chatResp) == nil {
			responsesResp := sse.BuildResponsesAPINonStream(chatResp, requestID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(upstream.StatusCode)
			json.NewEncoder(w).Encode(responsesResp)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(upstream.StatusCode)
			w.Write(respBody)
		}
	}
}

// injectionCheckResult holds the result of a prompt injection check.
type injectionCheckResult struct {
	Blocked bool
	Reason  string
	Pattern string
}

// checkPromptInjection scans a request body for common prompt injection patterns.
func checkPromptInjection(body map[string]interface{}) *injectionCheckResult {
	patterns := []string{
		"ignore previous instructions",
		"ignore all previous",
		"disregard all previous",
		"forget your instructions",
		"new instructions:",
		"system override",
		"jailbreak",
		"DAN mode",
		"bypass",
		"override safety",
		"ignore the above",
	}

	messages, ok := body["messages"].([]interface{})
	if !ok {
		return nil
	}

	for _, msg := range messages {
		m, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}
		content, _ := m["content"].(string)
		if content == "" {
			continue
		}
		lower := strings.ToLower(content)
		for _, pattern := range patterns {
			if strings.Contains(lower, pattern) {
				return &injectionCheckResult{
					Blocked: true,
					Reason:  "Potential prompt injection detected",
					Pattern: pattern,
				}
			}
		}
	}

	return &injectionCheckResult{Blocked: false}
}

// resolveProvider determines the provider ID and resolved model from a model string.
// Priority order when multiple providers support the same model:
// 1. Explicit provider prefix (e.g. "openai/gpt-5.5")
// 2. Canonical provider for the model family (openai→gpt, anthropic→claude, gemini→gemini, deepseek→deepseek)
//    - If the canonical provider has PassthroughModels, accept any model matching the prefix
//    - Otherwise, verify the model ID exists in the provider'"'"'s model list
// 3. Any provider with the model in its registry (excluding command-code, codex, kiro, cursor, windsurf which proxy other providers'"'"' models)
// 4. Falls back to "openai-compatible"
func resolveProvider(model string) (providerID, resolvedModel string) {
	// Check if model has a provider prefix (provider/model)
	parts := strings.SplitN(model, "/", 2)
	if len(parts) == 2 {
		providerID = parts[0]
		resolvedModel = parts[1]
		return
	}

	// Canonical provider mapping based on model name patterns.
	// When multiple providers register the same model (e.g. both openai and command-code
	// register gpt-5.5), the canonical provider always wins.
	canonicalProviders := []struct{ prefix, providerID string }{
		{"gpt-", "openai"},
		{"o1", "openai"},
		{"o3", "openai"},
		{"o4", "openai"},
		{"chatgpt", "openai"},
		{"claude-", "anthropic"},
		{"claude_", "anthropic"},
		{"gemini-", "gemini"},
		{"gemini_", "gemini"},
		{"deepseek", "deepseek"},
		{"qwen", "alibaba"},
		{"llama", "meta-llama"},
		{"mixtral", "mistral"},
		{"codestral", "mistral"},
		{"mistral-", "mistral"},
		{"grok", "xai"},
		{"sonar", "perplexity"},
		{"kimi", "kimi"},
		{"minimax", "minimax"},
		{"glm", "glm"},
		{"moonshot", "moonshot"},
		{"doubao", "volcengine"},
		{"hunyuan", "volcengine"},
		{"cogview", "volcengine"},
		{"spark", "minimax"},
	}
	for _, cp := range canonicalProviders {
		if strings.HasPrefix(strings.ToLower(model), cp.prefix) {
			entry := registry.Get(cp.providerID)
			if entry == nil {
				continue
			}
			// If the canonical provider has PassthroughModels, accept any model matching the prefix
			if entry.PassthroughModels {
				return cp.providerID, model
			}
			// If the model is explicitly registered, use it
			if entry.GetModel(model) != nil {
				return cp.providerID, model
			}
			// Even without an exact model match, if the canonical prefix matches,
			// prefer the canonical provider over other providers that also register this model
			// (e.g. gpt-5.5 should go to openai, not command-code)
			return cp.providerID, model
		}
	}

	// Try to find which provider supports this model.
	// Exclude proxy providers (command-code, codex, kiro, cursor, windsurf, opencode, antigravity)
	// that register other providers'"'"' models — they should only be used when explicitly selected.
	proxyProviders := map[string]bool{
		"command-code": true, "codex": true, "kiro": true,
		"cursor": true, "windsurf": true, "opencode": true,
		"opencode-go": true, "antigravity": true,
	}
	for _, entry := range registry.List() {
		if proxyProviders[entry.ID] {
			continue
		}
		if entry.GetModel(model) != nil {
			return entry.ID, model
		}
	}

	// Default to openai-compatible
	return "openai-compatible", model
}

// resolveCredentials extracts credentials for a provider.
func resolveCredentials(r *http.Request, dbConn *sql.DB, providerID string) executor.Credentials {
	creds := executor.Credentials{
		ProviderSpecificData: make(map[string]interface{}),
	}

	// Extract from request headers
	apiKey := auth.ExtractAPIKey(r)
	if apiKey != "" {
		creds.APIKey = apiKey
	}

	// Look up stored connections for this provider
	if dbConn != nil {
		connections, err := db.GetActiveProviderConnections(dbConn, providerID)
		if err == nil && len(connections) > 0 {
			conn := connections[0]
			if creds.APIKey == "" {
				creds.APIKey = conn.APIKey
			}
			creds.AccessToken = conn.AccessToken
			creds.RefreshToken = conn.RefreshToken
			creds.ProjectID = conn.ProjectID
			creds.ConnectionID = conn.ID
			creds.ProviderSpecificData = conn.ProviderSpecificData
		}
	}

	return creds
}

// resolveComboTargets tries to find combo targets for a model.
func resolveComboTargets(dbConn *sql.DB, model string) []routing.ResolvedTarget {
	if dbConn == nil {
		return nil
	}

	combos, err := db.GetActiveCombos(dbConn)
	if err != nil {
		return nil
	}

	for _, combo := range combos {
		for _, target := range combo.Targets {
			if target.Model == model {
				var allConns []db.ProviderConnection
				for _, t := range combo.Targets {
					conns, err := db.GetActiveProviderConnections(dbConn, t.Provider)
					if err == nil {
						allConns = append(allConns, conns...)
					}
				}
				return routing.ResolveTargets(combo, allConns)
			}
		}
	}

	return nil
}

// extractClientHeaders extracts relevant client headers.
func extractClientHeaders(r *http.Request) map[string]string {
	headers := map[string]string{}
	for key, values := range r.Header {
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "x-opencode-") ||
			lower == "user-agent" ||
			lower == "anthropic-beta" {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
	}
	return headers
}

// translateSSEStream translates SSE events between formats.
func translateSSEStream(writer *sse.Writer, upstream *http.Response, model, sourceFormat, targetFormat string, stopCh <-chan struct{}) {
	// Simplified: for now, just passthrough
	sse.StreamFromUpstream(writer, upstream, stopCh)
}

// --- Health Handler ---

// HealthHandler returns the service health status.
type HealthHandler struct {
	DB *sql.DB
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "4.0.0-go",
	}

	if h.DB != nil {
		if err := h.DB.Ping(); err != nil {
			status["db"] = "error"
		} else {
			status["db"] = "ok"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// --- Models Handler ---

// ModelsHandler returns the list of available models.
type ModelsHandler struct {
	DB *sql.DB
}

func (h *ModelsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var models []map[string]interface{}
	for _, entry := range registry.List() {
		for _, m := range entry.Models {
			models = append(models, map[string]interface{}{
				"id":       m.ID,
				"name":     m.Name,
				"provider": entry.ID,
				"object":   "model",
				"created":  time.Now().Unix(),
				"owned_by": entry.ID,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   models,
	})
}

// --- Combos Handler ---

// CombosHandler handles combo CRUD operations.
type CombosHandler struct {
	DB *sql.DB
}

func (h *CombosHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		combos, err := db.ListCombos(h.DB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(combos)

	case http.MethodPost:
		var combo db.Combo
		if err := json.NewDecoder(r.Body).Decode(&combo); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if combo.ID == "" {
			combo.ID = uuid.New().String()
		}
		if err := db.SaveCombo(h.DB, combo); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(combo)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- Providers Handler ---

// ProvidersHandler handles provider connection CRUD operations.
type ProvidersHandler struct {
	DB *sql.DB
}

func (h *ProvidersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		provider := r.URL.Query().Get("provider")
		connections, err := db.ListProviderConnections(h.DB, provider)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(connections)

	case http.MethodPost:
		var conn db.ProviderConnection
		if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if conn.ID == "" {
			conn.ID = uuid.New().String()
		}
		if err := db.SaveProviderConnection(h.DB, conn); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(conn)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
