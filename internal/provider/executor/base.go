package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/omniroute/omniroute/internal/config"
)

// BaseExecutor provides common executor functionality — the retry loop only.
type BaseExecutor struct {
	ProviderID string
	Config     *config.Config
}

// NewBaseExecutor creates a BaseExecutor with configuration.
func NewBaseExecutor(providerID string, cfg *config.Config) *BaseExecutor {
	return &BaseExecutor{
		ProviderID: providerID,
		Config:     cfg,
	}
}

// DoRequest sends a single HTTP request to the upstream provider with retry logic.
func (b *BaseExecutor) DoRequest(ctx context.Context, method, url string, headers map[string]string, bodyJSON []byte, maxRetries int, skipUpstreamRetry bool) (*http.Response, error) {
	for attempt := 0; attempt < maxRetries; attempt++ {
		reqCtx, cancel := context.WithTimeout(ctx, b.Config.FetchTimeout())
		req, err := http.NewRequestWithContext(reqCtx, method, url, bytes.NewReader(bodyJSON))
		cancel()
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}

		log.Printf("[%s] %s %s (attempt %d)", b.ProviderID, method, url, attempt+1)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return nil, fmt.Errorf("fetch failed after %d retries: %w", maxRetries, err)
		}

		// Retry on 429 or 5xx (unless skipUpstreamRetry)
		if !skipUpstreamRetry && (resp.StatusCode == 429 || (resp.StatusCode >= 500 && resp.StatusCode < 600)) {
			resp.Body.Close()
			backoff := time.Duration(attempt+1) * 2 * time.Second
			log.Printf("[%s] retry after %v: status %d", b.ProviderID, backoff, resp.StatusCode)
			time.Sleep(backoff)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded")
}

// ReadErrorBody reads the full response body and closes it, returning the bytes.
func (b *BaseExecutor) ReadErrorBody(resp *http.Response) []byte {
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return body
}
