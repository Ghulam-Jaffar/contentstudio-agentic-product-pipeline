// Package sentiment enriches parsed mentions with a sentiment label/score via
// the internal AI agents API. It is a separate stage from tag enrichment so
// AI outages in one classifier do not block the other from progressing.
package sentiment

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// SentimentAnalyzer abstracts the AI agent sentiment API for testability.
type SentimentAnalyzer interface {
	Request(ctx context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error)
}

// SentimentService consumes parsed mentions, enriches them with sentiment
// labels/scores via the AI agents endpoint, and emits enriched mentions.
type SentimentService struct {
	agent    SentimentAnalyzer
	producer kafka.Producer
	log      *logger.Logger
}

// NewSentimentService creates a new SentimentService.
func NewSentimentService(
	agent SentimentAnalyzer,
	producer kafka.Producer,
	log *logger.Logger,
) *SentimentService {
	return &SentimentService{
		agent:    agent,
		producer: producer,
		log:      log,
	}
}

// HandleParsedMention is a kafka.MessageHandler that enriches a single mention.
func (s *SentimentService) HandleParsedMention(ctx context.Context, _ string, _ []byte, value []byte) error {
	var mention kafkamodels.ListeningMention
	if err := json.Unmarshal(value, &mention); err != nil {
		return fmt.Errorf("SentimentService.HandleParsedMention: unmarshal: %w", err)
	}

	log := s.log.With().
		Str("mention_id", mention.MentionID).
		Str("topic_id", mention.TopicID).
		Logger()

	if mention.PostText != "" && s.agent != nil {
		label, score, err := s.analyzeSentiment(ctx, mention.PostText)
		if err != nil {
			log.Warn().Err(err).Msg("Sentiment analysis failed, using fallback")
		} else {
			mention.SentimentLabel = label
			mention.SentimentScore = score
		}
	}

	data, err := json.Marshal(mention)
	if err != nil {
		return fmt.Errorf("SentimentService.HandleParsedMention: marshal: %w", err)
	}

	if err := s.producer.Produce(ctx, kafkamodels.TopicListeningEnriched, []byte(mention.MentionID), data); err != nil {
		return fmt.Errorf("SentimentService.HandleParsedMention: produce: %w", err)
	}

	log.Debug().
		Str("sentiment", mention.SentimentLabel).
		Float64("score", mention.SentimentScore).
		Msg("Enriched mention")

	return nil
}

// analyzeSentiment calls the AI agent for sentiment analysis.
func (s *SentimentService) analyzeSentiment(ctx context.Context, text string) (string, float64, error) {
	payload := map[string]interface{}{
		"text": text,
		"task": "sentiment_analysis",
	}

	resp, err := s.agent.Request(ctx, "sentiment/analyze", payload)
	if err != nil {
		return "", 0, fmt.Errorf("analyzeSentiment: agent request: %w", err)
	}

	label, _ := resp["label"].(string)
	score, _ := resp["score"].(float64)

	return label, score, nil
}
