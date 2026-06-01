package main

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/errgroup"
)

// profileWorkOrderProcessor is the main worker loop for processing LinkedIn profile work orders.
// It continuously reads from the work order channel and processes each profile account.
// On successful processing, sends timestamp update request to update last_analytics_updated_at.
func profileWorkOrderProcessor(
	ctx context.Context,
	workerID int,
	workOrderChan <-chan WorkOrderMessage,
	liClient *social.LinkedInClient,
	producer kafka2.Producer,
	mongoRepo mongodb.UnifiedSocialRepository,
	decryptionKey string,
	log *logger.Logger,
	timestampUpdateChan chan<- TimestampUpdateRequest,
) {
	workerLog := log.With().Int("worker_id", workerID).Str("pool", "profile").Logger()
	workerLog.Info().Msg("Profile work order processor started")

	for {
		select {
		case msg, ok := <-workOrderChan:
			if !ok {
				workerLog.Info().Msg("Profile work order channel closed, stopping processor")
				return
			}
			if err := processProfileWorkOrder(ctx, msg, liClient, producer, mongoRepo, decryptionKey, timestampUpdateChan); err != nil {
				workerLog.Error().Err(err).Str("linkedin_id", msg.LinkedinID).Msg("Failed to process profile work order")
			}
		case <-ctx.Done():
			workerLog.Info().Msg("Context cancelled, stopping profile processor")
			return
		}
	}
}

// profileWorkOrderProcessorWithTracking is the profile worker with active job tracking for graceful shutdown.
func profileWorkOrderProcessorWithTracking(
	ctx context.Context,
	workerID int,
	workOrderChan <-chan WorkOrderMessage,
	liClient *social.LinkedInClient,
	producer kafka2.Producer,
	mongoRepo mongodb.UnifiedSocialRepository,
	decryptionKey string,
	log *logger.Logger,
	timestampUpdateChan chan<- TimestampUpdateRequest,
	activeJobs *int64,
	lastMessageTime *int64,
) {
	workerLog := log.With().Int("worker_id", workerID).Str("pool", "profile").Logger()
	workerLog.Info().Msg("Profile work order processor started")

	for {
		select {
		case msg, ok := <-workOrderChan:
			if !ok {
				workerLog.Info().Msg("Profile work order channel closed, stopping processor")
				return
			}

			// Track active job
			atomic.AddInt64(activeJobs, 1)

			if err := processProfileWorkOrder(ctx, msg, liClient, producer, mongoRepo, decryptionKey, timestampUpdateChan); err != nil {
				workerLog.Error().Err(err).Str("linkedin_id", msg.LinkedinID).Msg("Failed to process profile work order")
			}

			// Job completed - update tracking
			atomic.AddInt64(activeJobs, -1)
			atomic.StoreInt64(lastMessageTime, time.Now().UnixNano())

		case <-ctx.Done():
			workerLog.Info().Msg("Context cancelled, stopping profile processor")
			return
		}
	}
}

// processProfileWorkOrder processes a single LinkedIn profile work order.
// Profile accounts only fetch insights data (no posts or org details).
// On success, sends timestamp update request to the channel.
func processProfileWorkOrder(
	ctx context.Context,
	msg WorkOrderMessage,
	li *social.LinkedInClient,
	producer kafka2.Producer,
	mongoRepo mongodb.UnifiedSocialRepository,
	decryptionKey string,
	timestampUpdateChan chan<- TimestampUpdateRequest,
) (err error) {
	// Parse work order from Kafka message first to get context for logging
	var order LinkedInAccountWorkOrder
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		return err
	}

	// Create logger with full account context
	log := createLoggerWithContext(order, "profile", "processProfileWorkOrder")

	// Setup operation tracking
	op := createOperation(log, order, "profile")
	op.Start("processing linkedin profile work order")
	defer func() {
		op.Complete(err, "")
	}()

	// Decrypt access token (may be encrypted at rest)
	token := decryptToken(order.AccessToken, decryptionKey)
	if token == "" {
		if accountID, parseErr := primitive.ObjectIDFromHex(order.ID); parseErr == nil {
			mongoRepo.RecordProcessingError(ctx, accountID, "Access token is empty or decryption failed")
		}
		return nil
	}

	// Calculate date ranges based on sync type (incremental vs full)
	_, startDate, endDate := calculateDateRanges(order.SyncType)

	// Fetch and publish profile insights
	if err := fetchAndPublishProfileInsights(ctx, li, producer, order.LinkedinID, token, startDate, endDate, profileInsightsTopic, log); err != nil {
		// If token error, don't update timestamp - account needs re-authentication
		if isTokenError(err) {
			log.Warn().Err(err).Msg("Token invalid/expired - skipping timestamp update, account needs re-auth")
			if accountID, parseErr := primitive.ObjectIDFromHex(order.ID); parseErr == nil {
				mongoRepo.RecordProcessingError(context.Background(), accountID, err.Error())
			}
			// Return the error but don't treat as fatal - just skip this account
			return nil
		}
		if accountID, parseErr := primitive.ObjectIDFromHex(order.ID); parseErr == nil {
			mongoRepo.RecordProcessingError(context.Background(), accountID, err.Error())
		}
		return err
	}

	// Success - send timestamp update request (only if no token error)
	select {
	case timestampUpdateChan <- TimestampUpdateRequest{AccountID: order.ID, LinkedinID: order.LinkedinID}:
		log.Debug().Msg("Queued timestamp update")
	default:
		log.Warn().Msg("Timestamp update channel full, skipping update")
	}

	return nil
}

// fetchAndPublishProfileInsights orchestrates fetching all profile analytics data
// and publishing the compiled results to Kafka.
// Returns ErrTokenInvalid if the token is invalid/expired, which stops further processing.
func fetchAndPublishProfileInsights(
	ctx context.Context,
	li *social.LinkedInClient,
	producer kafka2.Producer,
	linkedinID string,
	token string,
	startDate, endDate time.Time,
	outputTopic string,
	log *logger.Logger,
) error {
	log.Info().
		Time("start_date", startDate).
		Time("end_date", endDate).
		Msg("Starting profile insights fetch")

	results, err := fetchProfileAnalytics(ctx, li, linkedinID, token, startDate, endDate, log)
	if err != nil {
		// Check if it's a token error - don't capture again, already captured in fetchProfileAnalytics
		if isTokenError(err) {
			log.Warn().Err(err).Msg("Token error detected, stopping profile processing")
			return wrapTokenError(err)
		}
		return err
	}

	if err := publishProfileInsights(ctx, producer, linkedinID, outputTopic, results, log); err != nil {
		return err
	}

	return nil
}

// fetchProfileAnalytics fetches all profile analytics data in parallel.
// Uses LinkedIn's memberCreatorPostAnalytics API for 5 different query types,
// plus memberFollowersCount for both daily and total follower counts.
// If a token error is detected, returns immediately and cancels other requests.
func fetchProfileAnalytics(
	ctx context.Context,
	li *social.LinkedInClient,
	linkedinID string,
	token string,
	startDate, endDate time.Time,
	log *logger.Logger,
) (*profileAnalyticsResults, error) {
	results := &profileAnalyticsResults{}
	var tokenErrorCaptured int32 // Atomic flag - only capture token error once to Sentry
	eg, egctx := errgroup.WithContext(ctx)

	// Define all analytics queries to fetch in parallel
	analyticsQueries := []analyticsQueryConfig{
		{queryType: "IMPRESSION", stageName: "fetch_profile_impression", logMsg: "failed to fetch linkedin profile impression data"},
		{queryType: "MEMBERS_REACHED", stageName: "fetch_profile_members_reached", logMsg: "failed to fetch linkedin profile members reached data"},
		{queryType: "RESHARE", stageName: "fetch_profile_reshare", logMsg: "failed to fetch linkedin profile reshare data"},
		{queryType: "REACTION", stageName: "fetch_profile_reaction", logMsg: "failed to fetch linkedin profile reaction data"},
		{queryType: "COMMENT", stageName: "fetch_profile_comment", logMsg: "failed to fetch linkedin profile comment data"},
	}

	// Fetch all analytics query types in parallel
	for _, q := range analyticsQueries {
		q := q // Capture for goroutine
		eg.Go(func() error {
			data, err := li.FetchMemberCreatorPostAnalyticsRaw(egctx, token, q.queryType, &startDate, &endDate)
			if err != nil {
				// If expected error, log as warning
				if isExpectedError(err) {
					log.Warn().Err(err).Str("queryType", q.queryType).Msg("Expected LinkedIn API error - " + q.logMsg)
					return err
				}
				// If token error, return it to cancel other goroutines
				if isTokenError(err) {
					if atomic.CompareAndSwapInt32(&tokenErrorCaptured, 0, 1) {
						log.Warn().Err(err).Msg("Token error detected - stopping all API calls")
					}
					return err // This cancels the errgroup context
				}
				// Non-token error - log but continue
				log.Error().Err(err).Msg(q.logMsg)
				return nil // Don't fail the whole batch for non-token errors
			}
			setProfileAnalyticsResult(results, q.queryType, data)
			return nil
		})
	}

	// Fetch daily follower count changes (q=dateRange)
	eg.Go(func() error {
		data, err := li.FetchMemberFollowersCountRaw(egctx, token, &startDate, &endDate)
		if err != nil {
			if isExpectedError(err) {
				log.Warn().Err(err).Msg("Expected LinkedIn API error fetching profile follower count data")
				return err
			}
			if isTokenError(err) {
				if atomic.CompareAndSwapInt32(&tokenErrorCaptured, 0, 1) {
					log.Warn().Err(err).Msg("Token error detected - stopping all API calls")
				}
				return err
			}
			log.Error().Err(err).Msg("failed to fetch linkedin profile follower count data")
			return nil
		}
		results.FollowerData = data
		return nil
	})

	// Fetch total follower count (q=me) - current snapshot, not daily
	eg.Go(func() error {
		data, err := li.FetchMemberFollowersCountRaw(egctx, token, nil, nil)
		if err != nil {
			if isExpectedError(err) {
				log.Warn().Err(err).Msg("Expected LinkedIn API error fetching profile total follower count")
				return err
			}
			if isTokenError(err) {
				if atomic.CompareAndSwapInt32(&tokenErrorCaptured, 0, 1) {
					log.Warn().Err(err).Msg("Token error detected - stopping all API calls")
				}
				return err
			}
			log.Error().Err(err).Msg("failed to fetch linkedin profile total follower count")
			return nil
		}
		results.TotalFollowerData = data
		return nil
	})

	if err := eg.Wait(); err != nil {
		return results, err // Return partial results + error
	}
	return results, nil
}

// setProfileAnalyticsResult stores the API response in the appropriate results field.
func setProfileAnalyticsResult(results *profileAnalyticsResults, queryType string, data []byte) {
	switch queryType {
	case "IMPRESSION":
		results.ImpressionData = data
	case "MEMBERS_REACHED":
		results.MembersReachedData = data
	case "RESHARE":
		results.ReshareData = data
	case "REACTION":
		results.ReactionData = data
	case "COMMENT":
		results.CommentData = data
	}
}

// publishProfileInsights merges all profile analytics data and publishes to Kafka.
func publishProfileInsights(
	ctx context.Context,
	producer kafka2.Producer,
	linkedinID string,
	outputTopic string,
	results *profileAnalyticsResults,
	log *logger.Logger,
) error {
	if !hasProfileData(results) {
		return nil
	}

	merged := buildProfilePayload(results)
	body, _ := json.Marshal(merged)
	_ = producer.Produce(ctx, outputTopic, []byte(linkedinID), body)

	log.Info().
		Str("linkedin_id", linkedinID).
		Str("topic", outputTopic).
		Bool("has_impression_data", results.ImpressionData != nil).
		Bool("has_members_reached_data", results.MembersReachedData != nil).
		Bool("has_reshare_data", results.ReshareData != nil).
		Bool("has_reaction_data", results.ReactionData != nil).
		Bool("has_comment_data", results.CommentData != nil).
		Bool("has_follower_data", results.FollowerData != nil).
		Bool("has_total_follower_data", results.TotalFollowerData != nil).
		Msg("produced linkedin profile insights to kafka")

	return nil
}

// hasProfileData checks if any profile analytics data was fetched.
func hasProfileData(results *profileAnalyticsResults) bool {
	return results.ImpressionData != nil ||
		results.MembersReachedData != nil ||
		results.ReshareData != nil ||
		results.ReactionData != nil ||
		results.CommentData != nil ||
		results.FollowerData != nil ||
		results.TotalFollowerData != nil
}

// buildProfilePayload constructs the Kafka message payload for profile insights.
func buildProfilePayload(results *profileAnalyticsResults) map[string]json.RawMessage {
	merged := map[string]json.RawMessage{
		"entityType": json.RawMessage(`"profile"`),
	}

	if results.ImpressionData != nil {
		merged["impressionData"] = results.ImpressionData
	}
	if results.MembersReachedData != nil {
		merged["membersReachedData"] = results.MembersReachedData
	}
	if results.ReshareData != nil {
		merged["reshareData"] = results.ReshareData
	}
	if results.ReactionData != nil {
		merged["reactionData"] = results.ReactionData
	}
	if results.CommentData != nil {
		merged["commentData"] = results.CommentData
	}
	if results.FollowerData != nil {
		merged["followerData"] = results.FollowerData
	}
	if results.TotalFollowerData != nil {
		merged["totalFollowerData"] = results.TotalFollowerData
	}

	return merged
}
