package enrichment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"github.com/rs/zerolog"
)

// TopicContext bundles every signal the AI agents endpoint needs to classify
// mentions for a single topic. Built per-batch by the enrichment service.
type TopicContext struct {
	AIContext     mongoModels.AIContext
	TopicName     string
	TopicType     string
	TopicKeywords []string
	RelevanceHint string
}

type aiContextWire struct {
	BrandName     string           `json:"brand_name"`
	BrandKeywords []string         `json:"brand_keywords"`
	Industry      string           `json:"industry"`
	Competitors   []competitorWire `json:"competitors"`
	TopicName     string           `json:"topic_name,omitempty"`
	TopicType     string           `json:"topic_type,omitempty"`
	TopicKeywords []string         `json:"topic_keywords,omitempty"`
	RelevanceHint string           `json:"relevance_hint,omitempty"`
}

type competitorWire struct {
	Name     string   `json:"name"`
	Keywords []string `json:"keywords"`
}

type batchRequest struct {
	Mentions []MentionPayload `json:"mentions"`
	Context  aiContextWire    `json:"context"`
}

type batchResponse struct {
	Success bool            `json:"success"`
	Results []MentionResult `json:"results"`
}

// AgentAnalyzer calls the AI agents batch analysis endpoint over HTTP.
type AgentAnalyzer struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     zerolog.Logger
}

// NewAgentAnalyzer creates a new AgentAnalyzer.
func NewAgentAnalyzer(baseURL, apiKey string, timeout int, logger zerolog.Logger) *AgentAnalyzer {
	if timeout <= 0 {
		timeout = 300
	}
	return &AgentAnalyzer{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		logger: logger.With().Str("component", "listening-batch-analyzer").Logger(),
	}
}

// AnalyzeBatch sends a batch of mentions to the AI agents endpoint for enrichment.
func (a *AgentAnalyzer) AnalyzeBatch(
	ctx context.Context,
	mentions []MentionPayload,
	topicCtx TopicContext,
) ([]MentionResult, error) {
	reqBody := batchRequest{
		Mentions: mentions,
		Context:  buildWireContext(topicCtx),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("AgentAnalyzer.AnalyzeBatch: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.baseURL+"/api/v1/listening/analyze-batch",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("AgentAnalyzer.AnalyzeBatch: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if a.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.apiKey)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AgentAnalyzer.AnalyzeBatch: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("AgentAnalyzer.AnalyzeBatch: read response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf(
			"AgentAnalyzer.AnalyzeBatch: AI agents returned status %d: %s",
			resp.StatusCode, string(respBody),
		)
	}

	var result batchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("AgentAnalyzer.AnalyzeBatch: unmarshal response: %w", err)
	}
	if !result.Success {
		return nil, fmt.Errorf("AgentAnalyzer.AnalyzeBatch: AI agents reported failure")
	}
	return result.Results, nil
}

func buildWireContext(t TopicContext) aiContextWire {
	competitors := make([]competitorWire, 0, len(t.AIContext.Competitors))
	for _, c := range t.AIContext.Competitors {
		competitors = append(competitors, competitorWire{
			Name:     c.Name,
			Keywords: append([]string{}, c.Keywords...),
		})
	}
	keywords := []string{}
	if t.AIContext.BrandKeywords != nil {
		keywords = append(keywords, t.AIContext.BrandKeywords...)
	}
	return aiContextWire{
		BrandName:     t.AIContext.BrandName,
		BrandKeywords: keywords,
		Industry:      t.AIContext.Industry,
		Competitors:   competitors,
		TopicName:     t.TopicName,
		TopicType:     t.TopicType,
		TopicKeywords: append([]string{}, t.TopicKeywords...),
		RelevanceHint: t.RelevanceHint,
	}
}
