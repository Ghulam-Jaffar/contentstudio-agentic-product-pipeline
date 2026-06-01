package main

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"golang.org/x/sync/errgroup"
)

// ProfileProcessor handles profile work order processing with injectable dependencies
type ProfileProcessor struct {
	client   LinkedInAPI
	producer kafka2.Producer
	log      *logger.Logger
}

// NewProfileProcessor creates a new profile processor with injected dependencies
func NewProfileProcessor(client LinkedInAPI, producer kafka2.Producer, log *logger.Logger) *ProfileProcessor {
	return &ProfileProcessor{
		client:   client,
		producer: producer,
		log:      log,
	}
}

// ProcessProfileWorkOrderTestable processes a profile work order with interface-based dependencies
func (p *ProfileProcessor) ProcessProfileWorkOrderTestable(
	ctx context.Context,
	order LinkedInAccountWorkOrder,
	token string,
	timestampUpdateChan chan<- TimestampUpdateRequest,
) error {
	log := p.log

	// Calculate date ranges based on sync type
	_, startDate, endDate := calculateDateRanges(order.SyncType)

	// Fetch and publish profile insights
	if err := p.fetchAndPublishProfileInsightsTestable(ctx, order.LinkedinID, token, startDate, endDate); err != nil {
		if isTokenError(err) {
			log.Warn().Err(err).Msg("Token invalid/expired - skipping timestamp update")
			return nil
		}
		return err
	}

	// Success - send timestamp update
	select {
	case timestampUpdateChan <- TimestampUpdateRequest{AccountID: order.ID, LinkedinID: order.LinkedinID}:
		log.Debug().Msg("Queued timestamp update")
	default:
		log.Warn().Msg("Timestamp update channel full")
	}

	return nil
}

// fetchAndPublishProfileInsightsTestable fetches and publishes profile insights
func (p *ProfileProcessor) fetchAndPublishProfileInsightsTestable(
	ctx context.Context,
	linkedinID string,
	token string,
	startDate, endDate time.Time,
) error {
	p.log.Info().
		Time("start_date", startDate).
		Time("end_date", endDate).
		Msg("Starting profile insights fetch")

	results, err := p.fetchProfileAnalyticsTestable(ctx, linkedinID, token, startDate, endDate)
	if err != nil {
		if isTokenError(err) {
			return wrapTokenError(err)
		}
		return err
	}

	return p.publishProfileInsightsTestable(ctx, linkedinID, results)
}

// fetchProfileAnalyticsTestable fetches all profile analytics in parallel
func (p *ProfileProcessor) fetchProfileAnalyticsTestable(
	ctx context.Context,
	linkedinID string,
	token string,
	startDate, endDate time.Time,
) (*profileAnalyticsResults, error) {
	results := &profileAnalyticsResults{}
	var tokenErrorCaptured int32

	eg, egctx := errgroup.WithContext(ctx)

	// Define all analytics queries
	analyticsQueries := []struct {
		queryType string
		setter    func(data []byte)
	}{
		{"IMPRESSION", func(data []byte) { results.ImpressionData = data }},
		{"MEMBERS_REACHED", func(data []byte) { results.MembersReachedData = data }},
		{"RESHARE", func(data []byte) { results.ReshareData = data }},
		{"REACTION", func(data []byte) { results.ReactionData = data }},
		{"COMMENT", func(data []byte) { results.CommentData = data }},
	}

	// Fetch all analytics query types in parallel
	for _, q := range analyticsQueries {
		q := q
		eg.Go(func() error {
			data, err := p.client.FetchMemberCreatorPostAnalyticsRaw(egctx, token, q.queryType, &startDate, &endDate)
			if err != nil {
				if isTokenError(err) {
					atomic.CompareAndSwapInt32(&tokenErrorCaptured, 0, 1)
					return err
				}
				p.log.Error().Err(err).Str("query_type", q.queryType).Msg("Failed to fetch analytics")
				return nil
			}
			q.setter(data)
			return nil
		})
	}

	// Fetch daily follower count
	eg.Go(func() error {
		data, err := p.client.FetchMemberFollowersCountRaw(egctx, token, &startDate, &endDate)
		if err != nil {
			if isTokenError(err) {
				atomic.CompareAndSwapInt32(&tokenErrorCaptured, 0, 1)
				return err
			}
			p.log.Error().Err(err).Msg("Failed to fetch daily follower count")
			return nil
		}
		results.FollowerData = data
		return nil
	})

	// Fetch total follower count
	eg.Go(func() error {
		data, err := p.client.FetchMemberFollowersCountRaw(egctx, token, nil, nil)
		if err != nil {
			if isTokenError(err) {
				atomic.CompareAndSwapInt32(&tokenErrorCaptured, 0, 1)
				return err
			}
			p.log.Error().Err(err).Msg("Failed to fetch total follower count")
			return nil
		}
		results.TotalFollowerData = data
		return nil
	})

	if err := eg.Wait(); err != nil {
		return results, err
	}
	return results, nil
}

// publishProfileInsightsTestable publishes profile insights to Kafka
func (p *ProfileProcessor) publishProfileInsightsTestable(
	ctx context.Context,
	linkedinID string,
	results *profileAnalyticsResults,
) error {
	if !hasProfileData(results) {
		return nil
	}

	merged := buildProfilePayload(results)
	body, _ := json.Marshal(merged)
	_ = p.producer.Produce(ctx, profileInsightsTopic, []byte(linkedinID), body)

	p.log.Info().
		Str("linkedin_id", linkedinID).
		Bool("has_impression_data", results.ImpressionData != nil).
		Bool("has_members_reached_data", results.MembersReachedData != nil).
		Bool("has_reshare_data", results.ReshareData != nil).
		Bool("has_reaction_data", results.ReactionData != nil).
		Bool("has_comment_data", results.CommentData != nil).
		Bool("has_follower_data", results.FollowerData != nil).
		Bool("has_total_follower_data", results.TotalFollowerData != nil).
		Msg("Published profile insights to Kafka")

	return nil
}
