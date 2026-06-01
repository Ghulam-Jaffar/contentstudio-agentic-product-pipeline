package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

// PostParserFunc defines the interface for post parsing
type PostParserFunc func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error)

// InsightsParserFunc defines the interface for insights parsing
type InsightsParserFunc func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error)

// ParserService handles parsing with injectable dependencies
type ParserService struct {
	parsePost     PostParserFunc
	parseInsights InsightsParserFunc
	producer      kafka2.Producer
	log           *logger.Logger
}

// NewParserService creates a new parser service with default parsers
func NewParserService(producer kafka2.Producer, log *logger.Logger) *ParserService {
	return &ParserService{
		parsePost:     parsing.ParsePost,
		parseInsights: parsing.ParseInsightsDaily,
		producer:      producer,
		log:           log,
	}
}

// NewParserServiceWithFuncs creates a parser service with custom parser functions
func NewParserServiceWithFuncs(
	parsePost PostParserFunc,
	parseInsights InsightsParserFunc,
	producer kafka2.Producer,
	log *logger.Logger,
) *ParserService {
	return &ParserService{
		parsePost:     parsePost,
		parseInsights: parseInsights,
		producer:      producer,
		log:           log,
	}
}

// ParseAndPublishPostTestable parses a post and publishes to Kafka
func (s *ParserService) ParseAndPublishPostTestable(
	ctx context.Context,
	linkedinID string,
	value []byte,
	outputTopic string,
) (*kafkamodels.ParsedLinkedinPost, error) {
	startTime := time.Now()

	parsed, err := s.parsePost(json.RawMessage(value))
	if err != nil {
		s.log.Error().
			Err(err).
			Str("linkedin_id", linkedinID).
			Dur("duration", time.Since(startTime)).
			Msg("Failed to parse linkedin post")
		return nil, err
	}
	if parsed == nil {
		return nil, nil
	}

	if parsed.LinkedinID == "" {
		parsed.LinkedinID = linkedinID
	}

	data, _ := json.Marshal(parsed)
	if err := s.producer.Produce(ctx, outputTopic, []byte(parsed.PostID), data); err != nil {
		s.log.Error().
			Err(err).
			Str("linkedin_id", linkedinID).
			Str("post_id", parsed.PostID).
			Str("topic", outputTopic).
			Msg("Failed to produce parsed post")
		return nil, err
	}

	s.log.Debug().
		Str("linkedin_id", linkedinID).
		Str("post_id", parsed.PostID).
		Dur("duration", time.Since(startTime)).
		Msg("Published parsed linkedin post")

	return parsed, nil
}

// ParseAndPublishInsightsTestable parses insights and publishes to Kafka
func (s *ParserService) ParseAndPublishInsightsTestable(
	ctx context.Context,
	linkedinID string,
	value []byte,
	outputTopic string,
) ([]*kafkamodels.ParsedLinkedinInsights, error) {
	startTime := time.Now()

	parsedList, err := s.parseInsights(json.RawMessage(value))
	if err != nil {
		s.log.Error().
			Err(err).
			Str("linkedin_id", linkedinID).
			Dur("duration", time.Since(startTime)).
			Msg("Failed to parse linkedin insights")
		return nil, err
	}
	if len(parsedList) == 0 {
		return nil, nil
	}

	s.log.Info().
		Str("linkedin_id", linkedinID).
		Str("topic", outputTopic).
		Int("daily_buckets", len(parsedList)).
		Msg("Parsed linkedin insights into daily buckets")

	for _, parsed := range parsedList {
		if parsed.LinkedinID == "" {
			parsed.LinkedinID = linkedinID
		}

		parsed.RecordID = fmt.Sprintf("%s_%s", parsed.LinkedinID, parsed.CreatedAt.Format("2006-01-02"))

		data, _ := json.Marshal(parsed)
		if err := s.producer.Produce(ctx, outputTopic, []byte(parsed.RecordID), data); err != nil {
			s.log.Error().
				Err(err).
				Str("linkedin_id", linkedinID).
				Str("record_id", parsed.RecordID).
				Msg("Failed to produce parsed insights")
			continue
		}
	}

	s.log.Debug().
		Str("linkedin_id", linkedinID).
		Int("daily_buckets", len(parsedList)).
		Dur("duration", time.Since(startTime)).
		Msg("Published parsed linkedin insights")

	return parsedList, nil
}

// PostParserWorkerTestable is a testable version of the post parser worker
func (s *ParserService) PostParserWorkerTestable(
	ctx context.Context,
	in <-chan ParseJob,
	counter *uint64,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-in:
			if !ok {
				return
			}
			outputTopic := job.OutputTopic
			if outputTopic == "" {
				outputTopic = topicParsedPagePosts
			}
			parsed, err := s.ParseAndPublishPostTestable(ctx, string(job.Key), job.Value, outputTopic)
			if err == nil && parsed != nil && counter != nil {
				atomic.AddUint64(counter, 1)
			}
		}
	}
}

// InsightsParserWorkerTestable is a testable version of the insights parser worker
func (s *ParserService) InsightsParserWorkerTestable(
	ctx context.Context,
	in <-chan ParseJob,
	counter *uint64,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-in:
			if !ok {
				return
			}
			outputTopic := job.OutputTopic
			if outputTopic == "" {
				outputTopic = topicParsedPageInsights
			}
			parsedList, err := s.ParseAndPublishInsightsTestable(ctx, string(job.Key), job.Value, outputTopic)
			if err == nil && len(parsedList) > 0 && counter != nil {
				atomic.AddUint64(counter, uint64(len(parsedList)))
			}
		}
	}
}
