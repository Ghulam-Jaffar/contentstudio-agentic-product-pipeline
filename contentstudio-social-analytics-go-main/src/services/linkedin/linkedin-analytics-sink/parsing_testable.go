package main

import (
	"context"
	"encoding/json"
	"sync/atomic"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

// PostParserFunc defines the interface for post parsing
type PostParserFunc func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error)

// InsightsParserFunc defines the interface for insights parsing
type InsightsParserFunc func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error)

// ParsingService handles parsing with injectable dependencies
type ParsingService struct {
	parsePost     PostParserFunc
	parseInsights InsightsParserFunc
	log           *logger.Logger
}

// NewParsingService creates a new parsing service with default parsers
func NewParsingService(log *logger.Logger) *ParsingService {
	return &ParsingService{
		parsePost:     parsing.ParsePost,
		parseInsights: parsing.ParseInsightsDaily,
		log:           log,
	}
}

// NewParsingServiceWithFuncs creates a parsing service with custom parser functions
func NewParsingServiceWithFuncs(parsePost PostParserFunc, parseInsights InsightsParserFunc, log *logger.Logger) *ParsingService {
	return &ParsingService{
		parsePost:     parsePost,
		parseInsights: parseInsights,
		log:           log,
	}
}

// ParseAndQueuePostTestable parses a post and queues it for batch processing
func (s *ParsingService) ParseAndQueuePostTestable(
	ctx context.Context,
	linkedinID string,
	value []byte,
	postsChan chan<- *kafkamodels.ParsedLinkedinPost,
	counter *uint64,
) error {
	parsed, err := s.parsePost(json.RawMessage(value))
	if err != nil {
		s.log.Error().Err(err).Str("linkedin_id", linkedinID).Msg("Failed to parse post")
		return err
	}
	if parsed == nil {
		return nil
	}

	if parsed.LinkedinID == "" {
		parsed.LinkedinID = linkedinID
	}

	select {
	case postsChan <- parsed:
		if counter != nil {
			atomic.AddUint64(counter, 1)
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// ParseAndQueueInsightsTestable parses insights and queues them for batch processing
func (s *ParsingService) ParseAndQueueInsightsTestable(
	ctx context.Context,
	linkedinID string,
	value []byte,
	insightsChan chan<- *kafkamodels.ParsedLinkedinInsights,
	counter *uint64,
) (int, error) {
	parsedList, err := s.parseInsights(json.RawMessage(value))
	if err != nil {
		s.log.Error().Err(err).Str("linkedin_id", linkedinID).Msg("Failed to parse insights")
		return 0, err
	}
	if len(parsedList) == 0 {
		return 0, nil
	}

	queued := 0
	for _, parsed := range parsedList {
		if parsed.LinkedinID == "" {
			parsed.LinkedinID = linkedinID
		}
		parsed.RecordID = parsed.LinkedinID + "_" + parsed.CreatedAt.Format("2006-01-02")

		select {
		case insightsChan <- parsed:
			if counter != nil {
				atomic.AddUint64(counter, 1)
			}
			queued++
		case <-ctx.Done():
			return queued, ctx.Err()
		}
	}

	return queued, nil
}

// PostsParserWorkerTestable is a testable version of the posts parser worker
func (s *ParsingService) PostsParserWorkerTestable(
	ctx context.Context,
	msgChan <-chan RawMessage,
	postsChan chan<- *kafkamodels.ParsedLinkedinPost,
	counter *uint64,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-msgChan:
			if !ok {
				return
			}
			_ = s.ParseAndQueuePostTestable(ctx, string(m.Key), m.Value, postsChan, counter)
		}
	}
}

// InsightsParserWorkerTestable is a testable version of the insights parser worker
func (s *ParsingService) InsightsParserWorkerTestable(
	ctx context.Context,
	msgChan <-chan RawMessage,
	insightsChan chan<- *kafkamodels.ParsedLinkedinInsights,
	counter *uint64,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-msgChan:
			if !ok {
				return
			}
			_, _ = s.ParseAndQueueInsightsTestable(ctx, string(m.Key), m.Value, insightsChan, counter)
		}
	}
}
