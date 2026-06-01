package main

import (
	"context"
	"encoding/json"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// MessageHandler handles parsed messages from Kafka
type MessageHandler struct {
	log *logger.Logger
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(log *logger.Logger) *MessageHandler {
	return &MessageHandler{log: log}
}

// HandleParsedPostTestable handles a parsed post message
func (h *MessageHandler) HandleParsedPostTestable(
	ctx context.Context,
	key, value []byte,
	postsChan chan<- *kafkamodels.ParsedLinkedinPost,
) error {
	var parsed kafkamodels.ParsedLinkedinPost
	if err := json.Unmarshal(value, &parsed); err != nil {
		h.log.Error().Err(err).Str("key", string(key)).Msg("unmarshal post failed")
		return err
	}

	select {
	case postsChan <- &parsed:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// HandleParsedInsightsTestable handles a parsed insights message
func (h *MessageHandler) HandleParsedInsightsTestable(
	ctx context.Context,
	key, value []byte,
	insightsChan chan<- *kafkamodels.ParsedLinkedinInsights,
) error {
	var parsed kafkamodels.ParsedLinkedinInsights
	if err := json.Unmarshal(value, &parsed); err != nil {
		h.log.Error().Err(err).Str("key", string(key)).Msg("unmarshal insights failed")
		return err
	}

	select {
	case insightsChan <- &parsed:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// PostsWorkerTestable processes posts messages from a channel
func (h *MessageHandler) PostsWorkerTestable(
	ctx context.Context,
	msgChan <-chan Message,
	postsChan chan<- *kafkamodels.ParsedLinkedinPost,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-msgChan:
			if !ok {
				return
			}
			_ = h.HandleParsedPostTestable(ctx, m.Key, m.Value, postsChan)
		}
	}
}

// InsightsWorkerTestable processes insights messages from a channel
func (h *MessageHandler) InsightsWorkerTestable(
	ctx context.Context,
	msgChan <-chan Message,
	insightsChan chan<- *kafkamodels.ParsedLinkedinInsights,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-msgChan:
			if !ok {
				return
			}
			_ = h.HandleParsedInsightsTestable(ctx, m.Key, m.Value, insightsChan)
		}
	}
}
