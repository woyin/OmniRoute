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

	"github.com/omniroute/omniroute/internal/auth"
	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
	"github.com/omniroute/omniroute/internal/sse"
	"github.com/omniroute/omniroute/internal/provider/executor"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

// --- Embeddings Handler ---

// EmbeddingsHandler handles /api/v1/embeddings requests.
type EmbeddingsHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *EmbeddingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	providerID, _ := resolveProviderForEndpoint(model, "embeddings")
	credentials := resolveCredentialsForEndpoint(r, h.DB, providerID)
	entry := registry.Get(providerID)

	// Build the URL
	if entry != nil && entry.BaseURL != "" {
	}

	// Execute
	exec := executor.GetExecutor(providerID, h.Config)
	ctx := r.Context()
	result, err := exec.Execute(ctx, executor.ExecuteInput{
		Model:        model,
		Body:         body,
		Stream:       false,
		Credentials:  credentials,
		EndpointPath: "/embeddings",
	})
	if err != nil {
		log.Printf("[EMBED] model=%s error: %v", model, err)
		writeJSONError(w, "Upstream request failed", "upstream_error", http.StatusBadGateway)
		return
	}
	defer result.Response.Body.Close()

	respBody, _ := io.ReadAll(result.Response.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Response.StatusCode)
	w.Write(respBody)

	// Record usage
	go db.RecordUsage(h.DB, db.UsageEntry{
		Provider: providerID, Model: model, Success: result.Response.StatusCode < 400,
	})
}

// --- Images Handler ---

// ImageGenerationsHandler handles /api/v1/images/generations requests.
type ImageGenerationsHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *ImageGenerationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	providerID, _ := resolveProviderForEndpoint(model, "images")
	credentials := resolveCredentialsForEndpoint(r, h.DB, providerID)

	exec := executor.GetExecutor(providerID, h.Config)
	ctx := r.Context()
	result, err := exec.Execute(ctx, executor.ExecuteInput{
		Model:        model,
		Body:         body,
		Stream:       false,
		Credentials:  credentials,
		EndpointPath: "/images/generations",
	})
	if err != nil {
		writeJSONError(w, "Upstream request failed", "upstream_error", http.StatusBadGateway)
		return
	}
	defer result.Response.Body.Close()

	respBody, _ := io.ReadAll(result.Response.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Response.StatusCode)
	w.Write(respBody)

	go db.RecordUsage(h.DB, db.UsageEntry{Provider: providerID, Model: model, Success: result.Response.StatusCode < 400})
}

// --- Audio Handlers ---

// AudioSpeechHandler handles /api/v1/audio/speech requests (TTS).
type AudioSpeechHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *AudioSpeechHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	providerID, _ := resolveProviderForEndpoint(model, "audio")
	credentials := resolveCredentialsForEndpoint(r, h.DB, providerID)

	exec := executor.GetExecutor(providerID, h.Config)
	ctx := r.Context()
	result, err := exec.Execute(ctx, executor.ExecuteInput{
		Model:        model,
		Body:         body,
		Stream:       false,
		Credentials:  credentials,
		EndpointPath: "/audio/speech",
	})
	if err != nil {
		writeJSONError(w, "Upstream request failed", "upstream_error", http.StatusBadGateway)
		return
	}
	defer result.Response.Body.Close()

	contentType := result.Response.Header.Get("Content-Type")
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.WriteHeader(result.Response.StatusCode)
	io.Copy(w, result.Response.Body)

	go db.RecordUsage(h.DB, db.UsageEntry{Provider: providerID, Model: model, Success: result.Response.StatusCode < 400})
}

// AudioTranscriptionsHandler handles /api/v1/audio/transcriptions requests.
type AudioTranscriptionsHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *AudioTranscriptionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Multipart form data — needs special handling
	writeJSONError(w, "Audio transcriptions not yet implemented in Go rewrite", "not_implemented", http.StatusNotImplemented)
}

// --- Moderations Handler ---

// ModerationsHandler handles /api/v1/moderations requests.
type ModerationsHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *ModerationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, "Invalid JSON body", "invalid_request_error", http.StatusBadRequest)
		return
	}

	model, _ := body["model"].(string)
	if model == "" {
		model = "omni-moderation-latest"
		body["model"] = model
	}

	providerID, _ := resolveProviderForEndpoint(model, "moderations")
	credentials := resolveCredentialsForEndpoint(r, h.DB, providerID)
	entry := registry.Get(providerID)

	if entry != nil && entry.BaseURL != "" {
	}

	exec := executor.GetExecutor(providerID, h.Config)
	ctx := r.Context()
	result, err := exec.Execute(ctx, executor.ExecuteInput{
		Model:        model,
		Body:         body,
		Stream:       false,
		Credentials:  credentials,
		EndpointPath: "/moderations",
	})
	if err != nil {
		writeJSONError(w, "Upstream request failed", "upstream_error", http.StatusBadGateway)
		return
	}
	defer result.Response.Body.Close()

	respBody, _ := io.ReadAll(result.Response.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Response.StatusCode)
	w.Write(respBody)

	go db.RecordUsage(h.DB, db.UsageEntry{Provider: providerID, Model: model, Success: result.Response.StatusCode < 400})
}

// --- Rerank Handler ---

// RerankHandler handles /api/v1/rerank requests.
type RerankHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *RerankHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	providerID, _ := resolveProviderForEndpoint(model, "rerank")
	credentials := resolveCredentialsForEndpoint(r, h.DB, providerID)

	exec := executor.GetExecutor(providerID, h.Config)
	ctx := r.Context()
	result, err := exec.Execute(ctx, executor.ExecuteInput{
		Model:        model,
		Body:         body,
		Stream:       false,
		Credentials:  credentials,
		EndpointPath: "/rerank",
	})
	if err != nil {
		writeJSONError(w, "Upstream request failed", "upstream_error", http.StatusBadGateway)
		return
	}
	defer result.Response.Body.Close()

	respBody, _ := io.ReadAll(result.Response.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Response.StatusCode)
	w.Write(respBody)

	go db.RecordUsage(h.DB, db.UsageEntry{Provider: providerID, Model: model, Success: result.Response.StatusCode < 400})
}

// --- Search Handler ---

// SearchHandler handles /api/v1/search requests.
type SearchHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, "Invalid JSON body", "invalid_request_error", http.StatusBadRequest)
		return
	}

	query, _ := body["query"].(string)
	if query == "" {
		writeJSONError(w, "query is required", "invalid_request_error", http.StatusBadRequest)
		return
	}

	providerID := "perplexity"
	if p, ok := body["provider"].(string); ok && p != "" {
		providerID = p
	}
	credentials := resolveCredentialsForEndpoint(r, h.DB, providerID)

	exec := executor.GetExecutor(providerID, h.Config)
	ctx := r.Context()
	result, err := exec.Execute(ctx, executor.ExecuteInput{
		Model:       "sonar",
		Body:        body,
		Stream:      false,
		Credentials: credentials,
	})
	if err != nil {
		writeJSONError(w, "Upstream request failed", "upstream_error", http.StatusBadGateway)
		return
	}
	defer result.Response.Body.Close()

	respBody, _ := io.ReadAll(result.Response.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Response.StatusCode)
	w.Write(respBody)

	go db.RecordUsage(h.DB, db.UsageEntry{Provider: providerID, Model: "search", Success: result.Response.StatusCode < 400})
}

// --- Completions Handler (legacy) ---

// CompletionsHandler handles /api/v1/completions requests.
type CompletionsHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *CompletionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	// Route as chat completions — legacy completions is deprecated
	providerID, resolvedModel := resolveProviderForEndpoint(model, "chat")
	credentials := resolveCredentialsForEndpoint(r, h.DB, providerID)

	exec := executor.GetExecutor(providerID, h.Config)
	ctx := r.Context()
	result, err := exec.Execute(ctx, executor.ExecuteInput{
		Model:       resolvedModel,
		Body:        body,
		Stream:      false,
		Credentials: credentials,
	})
	if err != nil {
		writeJSONError(w, "Upstream request failed", "upstream_error", http.StatusBadGateway)
		return
	}
	defer result.Response.Body.Close()

	respBody, _ := io.ReadAll(result.Response.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Response.StatusCode)
	w.Write(respBody)

	go db.RecordUsage(h.DB, db.UsageEntry{Provider: providerID, Model: model, Success: result.Response.StatusCode < 400})
}

// --- Files Handler ---

// FilesHandler handles /api/v1/files requests.
type FilesHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *FilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Return empty list
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"object": "list",
			"data":   []interface{}{},
		})
		return
	}
	if r.Method == http.MethodPost {
		writeJSONError(w, "File uploads not yet implemented in Go rewrite", "not_implemented", http.StatusNotImplemented)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// --- Batches Handler ---

// BatchesHandler handles /api/v1/batches requests.
type BatchesHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *BatchesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Return empty list
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"object": "list",
			"data":   []interface{}{},
		})
		return
	}
	if r.Method == http.MethodPost {
		writeJSONError(w, "Batches not yet implemented in Go rewrite", "not_implemented", http.StatusNotImplemented)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// --- Helper ---

func writeJSONError(w http.ResponseWriter, message, errType string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    errType,
		},
	})
}

// resolveProviderForEndpoint determines the provider for a non-chat endpoint.
func resolveProviderForEndpoint(model, endpoint string) (providerID, resolvedModel string) {
	// Check provider/model prefix
	parts := strings.SplitN(model, "/", 2)
	if len(parts) == 2 {
		if registry.Get(parts[0]) != nil {
			return parts[0], parts[1]
		}
	}

	// Try to find which provider supports this model
	for _, entry := range registry.List() {
		if entry.GetModel(model) != nil {
			return entry.ID, model
		}
	}

	// Default to openai
	return "openai", model
}

// resolveCredentialsForEndpoint extracts credentials for a provider.
func resolveCredentialsForEndpoint(r *http.Request, dbConn *sql.DB, providerID string) executor.Credentials {
	creds := executor.Credentials{
		ProviderSpecificData: make(map[string]interface{}),
	}

	apiKey := auth.ExtractAPIKey(r)
	if apiKey != "" {
		creds.APIKey = apiKey
	}

	if dbConn != nil {
		connections, err := db.GetActiveProviderConnections(dbConn, providerID)
		if err == nil && len(connections) > 0 {
			conn := connections[0]
			if creds.APIKey == "" {
				creds.APIKey = conn.APIKey
			}
			creds.AccessToken = conn.AccessToken
			creds.RefreshToken = conn.RefreshToken
			creds.ConnectionID = conn.ID
			creds.ProviderSpecificData = conn.ProviderSpecificData
		}
	}

	return creds
}

// writeJSONErrorf is a convenience wrapper around writeJSONError with fmt.Sprintf.
func writeJSONErrorf(w http.ResponseWriter, statusCode int, format string, args ...interface{}) {
	writeJSONError(w, fmt.Sprintf(format, args...), "invalid_request_error", statusCode)
}

// ensure time import is used
var _ = time.Second

// --- Video Generations Handler ---

// VideoGenerationsHandler handles /api/v1/videos/generations requests.
type VideoGenerationsHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *VideoGenerationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, "Invalid JSON body", "invalid_request_error", http.StatusBadRequest)
		return
	}
	model, _ := body["model"].(string)
	if model == "" {
		model = "comfyui/default"
		body["model"] = model
	}
	providerID, _ := resolveProviderForEndpoint(model, "videos")
	credentials := resolveCredentialsForEndpoint(r, h.DB, providerID)
	exec := executor.GetExecutor(providerID, h.Config)
	ctx := r.Context()
	result, err := exec.Execute(ctx, executor.ExecuteInput{
		Model:        model,
		Body:         body,
		Stream:       false,
		Credentials:  credentials,
		EndpointPath: "/videos/generations",
	})
	if err != nil {
		writeJSONError(w, "Upstream request failed", "upstream_error", http.StatusBadGateway)
		return
	}
	defer result.Response.Body.Close()
	respBody, _ := io.ReadAll(result.Response.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Response.StatusCode)
	w.Write(respBody)
	go db.RecordUsage(h.DB, db.UsageEntry{Provider: providerID, Model: model, Success: result.Response.StatusCode < 400})
}

// --- Music Generations Handler ---

// MusicGenerationsHandler handles /api/v1/music/generations requests.
type MusicGenerationsHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *MusicGenerationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, "Invalid JSON body", "invalid_request_error", http.StatusBadRequest)
			return
		}
		model, _ := body["model"].(string)
		if model == "" {
			model = "comfyui/music"
			body["model"] = model
		}
		providerID, _ := resolveProviderForEndpoint(model, "music")
		credentials := resolveCredentialsForEndpoint(r, h.DB, providerID)
		exec := executor.GetExecutor(providerID, h.Config)
		ctx := r.Context()
		result, err := exec.Execute(ctx, executor.ExecuteInput{
			Model:        model,
			Body:         body,
			Stream:       false,
			Credentials:  credentials,
			EndpointPath: "/music/generations",
		})
		if err != nil {
			writeJSONError(w, "Upstream request failed", "upstream_error", http.StatusBadGateway)
			return
		}
		defer result.Response.Body.Close()
		respBody, _ := io.ReadAll(result.Response.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(result.Response.StatusCode)
		w.Write(respBody)
		go db.RecordUsage(h.DB, db.UsageEntry{Provider: providerID, Model: model, Success: result.Response.StatusCode < 400})
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"object": "list",
			"data":   []interface{}{},
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- Messages Handler (Anthropic format) ---

// MessagesHandler handles /api/v1/messages requests (Anthropic Messages API).
type MessagesHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *MessagesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
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
	// Route to anthropic provider for messages API
	providerID := "anthropic"
	if p, ok := body["provider"].(string); ok && p != "" {
		providerID = p
	} else {
		pid, _ := resolveProvider(model)
		if pid != "openai-compatible" {
			providerID = pid
		}
	}
	credentials := resolveCredentialsForEndpoint(r, h.DB, providerID)
	exec := executor.GetExecutor(providerID, h.Config)
	ctx := r.Context()
	stream := false
	if s, ok := body["stream"].(bool); ok {
		stream = s
	}
	result, err := exec.Execute(ctx, executor.ExecuteInput{
		Model:       model,
		Body:        body,
		Stream:      stream,
		Credentials: credentials,
	})
	if err != nil {
		writeJSONError(w, "Upstream request failed", "upstream_error", http.StatusBadGateway)
		return
	}
	defer result.Response.Body.Close()
	if stream {
		sseWriter := sse.NewWriter(w)
		sseWriter.WriteHeader()
		sse.StreamFromUpstream(sseWriter, result.Response, make(chan struct{}))
	} else {
		respBody, _ := io.ReadAll(result.Response.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(result.Response.StatusCode)
		w.Write(respBody)
	}
	go db.RecordUsage(h.DB, db.UsageEntry{Provider: providerID, Model: model, Success: result.Response.StatusCode < 400})
}

// --- OCR Handler ---

// OCRHandler handles /api/v1/ocr requests.
type OCRHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *OCRHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, "Invalid JSON body", "invalid_request_error", http.StatusBadRequest)
		return
	}
	model, _ := body["model"].(string)
	if model == "" {
		model = "gemini-2.5-flash"
		body["model"] = model
	}
	providerID, _ := resolveProviderForEndpoint(model, "chat")
	credentials := resolveCredentialsForEndpoint(r, h.DB, providerID)
	exec := executor.GetExecutor(providerID, h.Config)
	ctx := r.Context()
	result, err := exec.Execute(ctx, executor.ExecuteInput{
		Model:        model,
		Body:         body,
		Stream:       false,
		Credentials:  credentials,
		EndpointPath: "/chat/completions",
	})
	if err != nil {
		writeJSONError(w, "Upstream request failed", "upstream_error", http.StatusBadGateway)
		return
	}
	defer result.Response.Body.Close()
	respBody, _ := io.ReadAll(result.Response.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Response.StatusCode)
	w.Write(respBody)
	go db.RecordUsage(h.DB, db.UsageEntry{Provider: providerID, Model: model, Success: result.Response.StatusCode < 400})
}

// --- Image Edits Handler ---

// ImageEditsHandler handles /api/v1/images/edits requests.
type ImageEditsHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *ImageEditsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Multipart form data — needs special handling
	writeJSONError(w, "Image edits not yet implemented in Go rewrite", "not_implemented", http.StatusNotImplemented)
}

// --- Audio Translations Handler ---

// AudioTranslationsHandler handles /api/v1/audio/translations requests.
type AudioTranslationsHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *AudioTranslationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Multipart form data — needs special handling
	writeJSONError(w, "Audio translations not yet implemented in Go rewrite", "not_implemented", http.StatusNotImplemented)
}

// --- Messages Count Tokens Handler ---

// MessagesCountTokensHandler handles /api/v1/messages/count_tokens requests.
type MessagesCountTokensHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *MessagesCountTokensHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, "Invalid JSON body", "invalid_request_error", http.StatusBadRequest)
		return
	}
	// Estimate tokens — rough approximation
	messages, _ := body["messages"].([]interface{})
	totalTokens := 0
	for _, msg := range messages {
		if m, ok := msg.(map[string]interface{}); ok {
			if content, ok := m["content"].(string); ok {
				totalTokens += len(content) / 4
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_tokens": totalTokens,
	})
}

// --- Provider-Specific Chat Handler ---

// ProviderChatHandler handles /api/v1/providers/{provider}/chat/completions —
// routes to a specific provider explicitly.
type ProviderChatHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *ProviderChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract provider from path: /api/v1/providers/{provider}/chat/completions
	providerID := ""
	parts := splitPath(r.URL.Path)
	for i, p := range parts {
		if p == "providers" && i+1 < len(parts) {
			providerID = parts[i+1]
			break
		}
	}
	if providerID == "" {
		writeJSONError(w, "provider id required in path", "invalid_request_error", http.StatusBadRequest)
		return
	}

	// Verify provider exists
	entry := registry.Get(providerID)
	if entry == nil {
		writeJSONError(w, fmt.Sprintf("provider %q not found", providerID), "invalid_request_error", http.StatusNotFound)
		return
	}

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

	credentials := resolveCredentialsForEndpoint(r, h.DB, providerID)
	stream := false
	if s, ok := body["stream"].(bool); ok {
		stream = s
	}

	exec := executor.GetExecutor(providerID, h.Config)
	ctx := r.Context()
	result, err := exec.Execute(ctx, executor.ExecuteInput{
		Model:       model,
		Body:        body,
		Stream:      stream,
		Credentials: credentials,
	})
	if err != nil {
		writeJSONError(w, "Upstream request failed", "upstream_error", http.StatusBadGateway)
		return
	}
	defer result.Response.Body.Close()

	if stream {
		sseWriter := sse.NewWriter(w)
		sseWriter.WriteHeader()
		sse.StreamFromUpstream(sseWriter, result.Response, make(chan struct{}))
	} else {
		respBody, _ := io.ReadAll(result.Response.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(result.Response.StatusCode)
		w.Write(respBody)
	}

	go db.RecordUsage(h.DB, db.UsageEntry{Provider: providerID, Model: model, Success: result.Response.StatusCode < 400})
}

// --- Quota Check Handler ---

// QuotaCheckHandler handles POST /api/v1/quotas/check.
type QuotaCheckHandler struct {
	DB *sql.DB
}

func (h *QuotaCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, "Invalid JSON body", "invalid_request_error", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"provider":    body.Provider,
		"model":       body.Model,
		"quotaLeft":   -1,
		"unlimited":   true,
	})
}

// --- Registered Keys Handlers ---

// RegisteredKeysHandler handles GET/POST /api/v1/registered-keys.
type RegisteredKeysHandler struct {
	DB *sql.DB
}

func (h *RegisteredKeysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"object": "list",
			"data":   []interface{}{},
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
