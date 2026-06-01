package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
)

type AgentClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     zerolog.Logger
}

func NewAgentClient(cfg *config.AIAgentsConfig, logger zerolog.Logger) *AgentClient {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 300
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/") + "/"

	return &AgentClient{
		baseURL: baseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		logger: logger.With().Str("component", "ai-agent-client").Logger(),
	}
}

func (c *AgentClient) Request(ctx context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	url := c.baseURL + "api/v1/" + strings.TrimLeft(endpoint, "/")

	httputil.SanitizeFloats(payload)
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	c.logger.Debug().Str("url", url).Msg("AI agent request")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ai agent request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		c.logger.Warn().Int("status", resp.StatusCode).Str("url", url).Msg("AI agent request failed")
		return nil, fmt.Errorf("ai agent returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return result, nil
}
