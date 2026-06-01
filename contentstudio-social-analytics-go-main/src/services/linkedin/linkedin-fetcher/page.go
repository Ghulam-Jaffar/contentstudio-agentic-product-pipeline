package main

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/errgroup"
)

// pageWorkOrderProcessor is the main worker loop for processing LinkedIn page work orders.
// It continuously reads from the work order channel and processes each page/organization account.
// On successful processing, sends timestamp update request to update last_analytics_updated_at.
func pageWorkOrderProcessor(
	ctx context.Context,
	workerID int,
	workOrderChan <-chan WorkOrderMessage,
	liClient *social.LinkedInClient,
	geoResolver *social.GeoResolver,
	producer kafka2.Producer,
	mongoRepo mongodb.UnifiedSocialRepository,
	decryptionKey string,
	log *logger.Logger,
	timestampUpdateChan chan<- TimestampUpdateRequest,
) {
	workerLog := log.With().Int("worker_id", workerID).Str("pool", "page").Logger()
	workerLog.Info().Msg("Page work order processor started")

	for {
		select {
		case msg, ok := <-workOrderChan:
			if !ok {
				workerLog.Info().Msg("Page work order channel closed, stopping processor")
				return
			}
			if err := processPageWorkOrder(ctx, msg, liClient, geoResolver, producer, mongoRepo, decryptionKey, timestampUpdateChan); err != nil {
				workerLog.Error().Err(err).Str("linkedin_id", msg.LinkedinID).Msg("Failed to process page work order")
			}
		case <-ctx.Done():
			workerLog.Info().Msg("Context cancelled, stopping page processor")
			return
		}
	}
}

// pageWorkOrderProcessorWithTracking is the page worker with active job tracking for graceful shutdown.
func pageWorkOrderProcessorWithTracking(
	ctx context.Context,
	workerID int,
	workOrderChan <-chan WorkOrderMessage,
	liClient *social.LinkedInClient,
	geoResolver *social.GeoResolver,
	producer kafka2.Producer,
	mongoRepo mongodb.UnifiedSocialRepository,
	decryptionKey string,
	log *logger.Logger,
	timestampUpdateChan chan<- TimestampUpdateRequest,
	activeJobs *int64,
	lastMessageTime *int64,
) {
	workerLog := log.With().Int("worker_id", workerID).Str("pool", "page").Logger()
	workerLog.Info().Msg("Page work order processor started")

	for {
		select {
		case msg, ok := <-workOrderChan:
			if !ok {
				workerLog.Info().Msg("Page work order channel closed, stopping processor")
				return
			}

			// Track active job
			atomic.AddInt64(activeJobs, 1)

			if err := processPageWorkOrder(ctx, msg, liClient, geoResolver, producer, mongoRepo, decryptionKey, timestampUpdateChan); err != nil {
				workerLog.Error().Err(err).Str("linkedin_id", msg.LinkedinID).Msg("Failed to process page work order")
			}

			// Job completed - update tracking
			atomic.AddInt64(activeJobs, -1)
			atomic.StoreInt64(lastMessageTime, time.Now().UnixNano())

		case <-ctx.Done():
			workerLog.Info().Msg("Context cancelled, stopping page processor")
			return
		}
	}
}

// processPageWorkOrder processes a single LinkedIn page/organization work order.
// Page accounts fetch posts, insights, and organization details - all in parallel.
// On success, sends timestamp update request to the channel.
func processPageWorkOrder(
	ctx context.Context,
	msg WorkOrderMessage,
	li *social.LinkedInClient,
	geoResolver *social.GeoResolver,
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
	log := createLoggerWithContext(order, "page", "processPageWorkOrder")

	// Setup operation tracking
	op := createOperation(log, order, "page")
	op.Start("processing linkedin page work order")
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
	cutoffTime, startDate, endDate := calculateDateRanges(order.SyncType)

	// For pages, fetch posts, insights, and org details in parallel
	// If any returns a token error, the errgroup context is cancelled
	var tokenErrorDetected int32 // Atomic flag for token error
	g, gctx := errgroup.WithContext(ctx)

	// Fetch and publish posts with stats and assets
	g.Go(func() error {
		err := fetchAndPublishPagePosts(gctx, li, producer, &order, token, cutoffTime, pagePostsTopic, log)
		if err != nil && isTokenError(err) {
			atomic.StoreInt32(&tokenErrorDetected, 1)
		}
		return err
	})

	// Fetch and publish page insights
	g.Go(func() error {
		err := fetchAndPublishPageInsights(gctx, li, geoResolver, producer, order.LinkedinID, token, startDate, endDate, pageInsightsTopic, log)
		if err != nil && isTokenError(err) {
			atomic.StoreInt32(&tokenErrorDetected, 1)
		}
		return err
	})

	// Fetch and publish organization details
	g.Go(func() error {
		err := fetchAndPublishOrgDetails(gctx, li, producer, order.LinkedinID, token, pageOrganizationTopic, log)
		if err != nil && isTokenError(err) {
			atomic.StoreInt32(&tokenErrorDetected, 1)
		}
		return err
	})

	if err := g.Wait(); err != nil {
		// If token error, don't update timestamp - account needs re-auth
		if atomic.LoadInt32(&tokenErrorDetected) == 1 || isTokenError(err) {
			log.Warn().Err(err).Msg("Token invalid/expired - skipping timestamp update, account needs re-auth")
			if accountID, parseErr := primitive.ObjectIDFromHex(order.ID); parseErr == nil {
				mongoRepo.RecordProcessingError(context.Background(), accountID, err.Error())
			}
			return nil // Don't treat as fatal, just skip this account
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

// fetchAndPublishPageInsights orchestrates fetching all page analytics data
// and publishing the compiled results to Kafka.
// Returns ErrTokenInvalid if the token is invalid/expired.
func fetchAndPublishPageInsights(
	ctx context.Context,
	li *social.LinkedInClient,
	geoResolver *social.GeoResolver,
	producer kafka2.Producer,
	linkedinID string,
	token string,
	startDate, endDate time.Time,
	outputTopic string,
	log *logger.Logger,
) error {
	results, err := fetchPageAnalytics(ctx, li, geoResolver, linkedinID, token, startDate, endDate, log)
	if err != nil {
		if isTokenError(err) {
			return wrapTokenError(err)
		}
		return err
	}
	return publishPageInsights(ctx, producer, linkedinID, outputTopic, results, log)
}

// fetchPageAnalytics fetches all page/organization analytics data in parallel.
// Uses LinkedIn's Organization APIs for follower data, page stats, and share stats.
// If a token error is detected, returns immediately and cancels other requests.
// Uses GeoResolver to resolve geo IDs to human-readable names via ClickHouse cache + LinkedIn API.
func fetchPageAnalytics(
	ctx context.Context,
	li *social.LinkedInClient,
	geoResolver *social.GeoResolver,
	linkedinID string,
	token string,
	startDate, endDate time.Time,
	log *logger.Logger,
) (*pageAnalyticsResults, error) {
	results := &pageAnalyticsResults{}
	startMs := startDate.UnixMilli()
	endMs := endDate.UnixMilli()
	var tokenErrorCaptured int32

	eg, egctx := errgroup.WithContext(ctx)

	// Fetch follower demographics with geo ID resolution (total count, by industry, by seniority, etc.)
	// Uses GeoResolver to resolve geo IDs to names via ClickHouse cache + LinkedIn API fallback
	// Optimized: fetches follower stats once, extracts geo IDs, resolves, then builds response
	eg.Go(func() error {
		log.Info().Str("linkedin_id", linkedinID).Msg("Fetching follower data with geo resolution")

		// Step 1: Fetch follower stats and extract geo IDs in single API call
		followerStats, err := li.FetchFollowerStatsWithGeoIDs(egctx, linkedinID, token)
		if err != nil {
			if isExpectedError(err) {
				// Expected error - log as warning
				log.Warn().Err(err).Msg("Expected LinkedIn API error fetching follower stats")
				return err
			}
			if isTokenError(err) {
				if atomic.CompareAndSwapInt32(&tokenErrorCaptured, 0, 1) {
					log.Warn().Err(err).Msg("Token error detected - stopping all API calls")
				}
				return err
			}
			log.Error().Err(err).Msg("failed to fetch follower stats")
			return nil
		}

		// Step 2: Resolve geo IDs using cache + API fallback (includes geo type for caching)
		// Acquire geo semaphore to limit concurrent LinkedIn Geo API calls
		var geoNames map[string]string
		if len(followerStats.GeoIDs) > 0 {
			if err := geoConc.Acquire(egctx, 1); err != nil {
				log.Warn().Err(err).Str("linkedin_id", linkedinID).Msg("Failed to acquire geo semaphore")
				return nil
			}
			log.Info().Str("linkedin_id", linkedinID).Int("geo_ids_count", len(followerStats.GeoIDs)).Msg("Resolving geo IDs")
			geoNames, err = geoResolver.ResolveGeoIDsWithType(egctx, followerStats.GeoIDs, token)
			geoConc.Release(1)
			if err != nil {
				log.Warn().Err(err).Str("linkedin_id", linkedinID).Msg("Failed to resolve geo IDs, will use numeric IDs")
			} else {
				log.Info().Str("linkedin_id", linkedinID).Int("resolved_count", len(geoNames)).Msg("Resolved geo IDs to names")
			}
		}

		// Step 3: Build final follower data response with resolved geo names (no additional API call)
		data, err := li.BuildFollowerDataWithGeoNames(followerStats, geoNames)
		if err != nil {
			log.Error().Err(err).Msg("failed to build follower data response")
			return nil
		}
		results.FollowerData = data
		log.Info().Str("linkedin_id", linkedinID).Msg("Fetched follower data with geo names")
		return nil
	})

	// Fetch page view statistics (daily page views, unique visitors)
	eg.Go(func() error {
		data, err := li.FetchPageStatisticsRaw(egctx, linkedinID, token, startMs, endMs)
		if err != nil {
			if isExpectedError(err) {
				// Expected error - log as warning
				log.Warn().Err(err).Msg("Expected LinkedIn API error fetching page statistics")
				return err
			}
			if isTokenError(err) {
				if atomic.CompareAndSwapInt32(&tokenErrorCaptured, 0, 1) {
					log.Warn().Err(err).Msg("Token error detected - stopping all API calls")
				}
				return err
			}
			log.Error().Err(err).Msg("failed to fetch linkedin page statistics")
			return nil
		}
		results.PageStats = data
		return nil
	})

	// Fetch share statistics (daily engagement: likes, comments, shares, clicks)
	eg.Go(func() error {
		data, err := li.FetchShareStatisticsRaw(egctx, linkedinID, token, startMs, endMs)
		if err != nil {
			if isExpectedError(err) {
				// Expected error - log as warning
				log.Warn().Err(err).Msg("Expected LinkedIn API error fetching share statistics")
				return err
			}
			if isTokenError(err) {
				if atomic.CompareAndSwapInt32(&tokenErrorCaptured, 0, 1) {
					log.Warn().Err(err).Msg("Token error detected - stopping all API calls")
				}
				return err
			}
			log.Error().Err(err).Msg("failed to fetch linkedin share statistics")
			return nil
		}
		results.ShareStats = data
		return nil
	})

	if err := eg.Wait(); err != nil {
		return results, err
	}
	return results, nil
}

// publishPageInsights merges page analytics data and publishes to Kafka.
func publishPageInsights(
	ctx context.Context,
	producer kafka2.Producer,
	linkedinID string,
	outputTopic string,
	results *pageAnalyticsResults,
	log *logger.Logger,
) error {
	if results.FollowerData == nil && results.PageStats == nil && results.ShareStats == nil {
		return nil
	}

	merged := map[string]json.RawMessage{}
	if results.FollowerData != nil {
		merged["followerData"] = results.FollowerData
	}
	if results.PageStats != nil {
		merged["pageStatistics"] = results.PageStats
	}
	if results.ShareStats != nil {
		merged["shareStatistics"] = results.ShareStats
	}

	body, _ := json.Marshal(merged)
	_ = producer.Produce(ctx, outputTopic, []byte(linkedinID), body)

	log.Info().
		Str("linkedin_id", linkedinID).
		Str("topic", outputTopic).
		Bool("has_follower_data", results.FollowerData != nil).
		Bool("has_page_stats", results.PageStats != nil).
		Bool("has_share_stats", results.ShareStats != nil).
		Msg("produced linkedin page insights to kafka")

	return nil
}

// fetchAndPublishPagePosts fetches posts with their stats and assets, then publishes to Kafka.
// Steps: 1) Fetch posts, 2) Parse & collect IDs, 3) Fetch assets, 4) Merge & publish
func fetchAndPublishPagePosts(
	ctx context.Context,
	li *social.LinkedInClient,
	producer kafka2.Producer,
	order *LinkedInAccountWorkOrder,
	token string,
	cutoffTime time.Time,
	outputTopic string,
	log *logger.Logger,
) error {
	log.Info().Time("cutoff_time", cutoffTime).Msg("Step 1: Fetching posts from LinkedIn API")

	// Step 1: Fetch posts from LinkedIn API
	posts, err := li.FetchPostsPaginated(ctx, order.LinkedinID, entityTypeOrganization, token, cutoffTime)
	if err != nil {
		// Check for token error first
		if isTokenError(err) {
			log.Error().Err(err).Msg("Token error detected while fetching posts")
			return wrapTokenError(err)
		}
		log.Error().Err(err).Time("cutoff_time", cutoffTime).Msg("failed to fetch linkedin posts")
		return nil // Non-token errors don't stop processing
	}
	log.Info().Int("posts_fetched", len(posts)).Msg("Step 1 complete: fetched linkedin posts")

	// Step 2: Parse posts and collect IDs for stats and assets
	byActivity, ugcIDs, shareIDs, imageIDs, videoIDs, documentIDs := parsePostsAndCollectIDs(posts)
	log.Info().
		Int("unique_posts", len(byActivity)).
		Int("ugc_ids", len(ugcIDs)).
		Int("share_ids", len(shareIDs)).
		Int("image_ids", len(imageIDs)).
		Int("video_ids", len(videoIDs)).
		Int("document_ids", len(documentIDs)).
		Msg("Step 2 complete: parsed posts and collected asset IDs")

	// Step 3: Fetch all stats and assets in parallel
	log.Info().Msg("Step 3: Fetching stats and assets in parallel")
	statsMap, imgMap, vidMap, docMap := fetchPostAssets(ctx, li, order.LinkedinID, token, ugcIDs, shareIDs, imageIDs, videoIDs, documentIDs, log)
	log.Info().
		Int("stats_fetched", len(statsMap)).
		Int("images_fetched", len(imgMap)).
		Int("videos_fetched", len(vidMap)).
		Int("documents_fetched", len(docMap)).
		Msg("Step 3 complete: fetched assets")

	// Step 4: Merge and publish
	log.Info().Msg("Step 4: Merging and publishing posts")
	return mergeAndPublishPosts(ctx, producer, order.LinkedinID, outputTopic, byActivity, statsMap, imgMap, vidMap, docMap, log)
}

// parsePostsAndCollectIDs parses raw post JSON and extracts all asset IDs.
func parsePostsAndCollectIDs(posts []json.RawMessage) (
	map[string]*enrichedPost,
	[]string, []string, []string, []string, []string,
) {
	byActivity := make(map[string]*enrichedPost)
	ugcSet, shareSet := map[string]struct{}{}, map[string]struct{}{}
	imgSet, vidSet, docSet := map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}

	for _, raw := range posts {
		var post map[string]any
		if json.Unmarshal(raw, &post) != nil {
			continue
		}

		id, _ := post["id"].(string)
		if id == "" {
			continue
		}

		ep, ok := byActivity[id]
		if !ok {
			ep = &enrichedPost{Post: post, ActivityID: id}
			byActivity[id] = ep
		}

		classifyPostForStats(id, ugcSet, shareSet)
		collectAssetIDs(post, ep, imgSet, vidSet, docSet)
	}

	return byActivity,
		mapKeysToSlice(ugcSet), mapKeysToSlice(shareSet),
		mapKeysToSlice(imgSet), mapKeysToSlice(vidSet), mapKeysToSlice(docSet)
}

// classifyPostForStats determines which stats API to use based on post ID format.
func classifyPostForStats(id string, ugcSet, shareSet map[string]struct{}) {
	if strings.Contains(id, "ugcPost") {
		ugcSet[id] = struct{}{}
	}
	if strings.Contains(id, "share") {
		shareSet[id] = struct{}{}
	}
}

// collectAssetIDs extracts all media asset IDs from a post's content.
func collectAssetIDs(post map[string]any, ep *enrichedPost, imgSet, vidSet, docSet map[string]struct{}) {
	content, _ := post["content"].(map[string]any)
	if content == nil {
		return
	}

	for k, v := range content {
		switch k {
		case "multiImage":
			collectMultiImageIDs(v, ep, imgSet)
		case "article":
			collectArticleThumbnail(v, ep, imgSet)
		case "media":
			collectMediaID(v, ep, imgSet, vidSet, docSet)
		}
	}
}

// collectMultiImageIDs extracts image IDs from a multiImage (carousel) post.
func collectMultiImageIDs(v any, ep *enrichedPost, imgSet map[string]struct{}) {
	m, ok := v.(map[string]any)
	if !ok {
		return
	}
	imgs, ok := m["images"].([]any)
	if !ok {
		return
	}
	for _, it := range imgs {
		if mm, ok := it.(map[string]any); ok {
			if iid, ok := mm["id"].(string); ok && iid != "" {
				imgSet[iid] = struct{}{}
				ep.ImageIDs = append(ep.ImageIDs, iid)
			}
		}
	}
}

// collectArticleThumbnail extracts the thumbnail image ID from an article post.
func collectArticleThumbnail(v any, ep *enrichedPost, imgSet map[string]struct{}) {
	m, ok := v.(map[string]any)
	if !ok {
		return
	}
	if thumb, ok := m["thumbnail"].(string); ok && thumb != "" {
		imgSet[thumb] = struct{}{}
		ep.ImageIDs = append(ep.ImageIDs, thumb)
	}
}

// collectMediaID extracts and classifies a single media asset ID.
func collectMediaID(v any, ep *enrichedPost, imgSet, vidSet, docSet map[string]struct{}) {
	m, ok := v.(map[string]any)
	if !ok {
		return
	}
	idStr, ok := m["id"].(string)
	if !ok || idStr == "" {
		return
	}

	switch {
	case strings.Contains(idStr, "video"):
		vidSet[idStr] = struct{}{}
		ep.VideoIDs = append(ep.VideoIDs, idStr)
	case strings.Contains(idStr, "document"):
		docSet[idStr] = struct{}{}
		ep.DocumentIDs = append(ep.DocumentIDs, idStr)
	default:
		imgSet[idStr] = struct{}{}
		ep.ImageIDs = append(ep.ImageIDs, idStr)
	}
}

// fetchPostAssets fetches all post stats and media assets in parallel.
func fetchPostAssets(
	ctx context.Context,
	li *social.LinkedInClient,
	linkedinID string,
	token string,
	ugcIDs, shareIDs, imageIDs, videoIDs, documentIDs []string,
	log *logger.Logger,
) (statsMap, imgMap, vidMap, docMap assetMap) {
	var statsMu, imgMu, vidMu, docMu sync.Mutex
	statsMap = make(assetMap)
	imgMap = make(assetMap)
	vidMap = make(assetMap)
	docMap = make(assetMap)

	eg, egctx := errgroup.WithContext(ctx)

	// Fetch UGC and share stats
	fetchStatsForType(egctx, eg, li, linkedinID, token, ugcIDs, true, &statsMu, statsMap, log)
	fetchStatsForType(egctx, eg, li, linkedinID, token, shareIDs, false, &statsMu, statsMap, log)

	// Fetch images
	fetchAssetsInChunks(egctx, eg, &assetFetchConfig{
		name:      "images",
		ids:       imageIDs,
		chunkSize: 80,
		semaphore: mediaConc,
		fetchFunc: func(ctx context.Context, ids []string) ([]byte, error) {
			return li.FetchImagesRaw(ctx, ids, token)
		},
		parseFunc:  parseAssetBatch,
		mu:         &imgMu,
		resultMap:  imgMap,
		linkedinID: linkedinID,
		log:        log,
	})

	// Fetch videos
	fetchAssetsInChunks(egctx, eg, &assetFetchConfig{
		name:      "videos",
		ids:       videoIDs,
		chunkSize: 80,
		semaphore: mediaConc,
		fetchFunc: func(ctx context.Context, ids []string) ([]byte, error) {
			return li.FetchVideosRaw(ctx, ids, token)
		},
		parseFunc:  parseAssetBatch,
		mu:         &vidMu,
		resultMap:  vidMap,
		linkedinID: linkedinID,
		log:        log,
	})

	// Fetch documents
	if len(documentIDs) > 0 {
		log.Info().Str("linkedin_id", linkedinID).Int("document_count", len(documentIDs)).Msg("fetching linkedin documents")
	}
	fetchAssetsInChunks(egctx, eg, &assetFetchConfig{
		name:      "documents",
		ids:       documentIDs,
		chunkSize: 80,
		semaphore: mediaConc,
		fetchFunc: func(ctx context.Context, ids []string) ([]byte, error) {
			return li.FetchDocumentsRaw(ctx, ids, token)
		},
		parseFunc:  parseAssetBatch,
		mu:         &docMu,
		resultMap:  docMap,
		linkedinID: linkedinID,
		log:        log,
	})

	_ = eg.Wait()
	return
}

// fetchStatsForType fetches share statistics for either UGC posts or Share posts.
func fetchStatsForType(
	ctx context.Context,
	eg *errgroup.Group,
	li *social.LinkedInClient,
	linkedinID, token string,
	ids []string,
	isUGC bool,
	mu *sync.Mutex,
	statsMap assetMap,
	log *logger.Logger,
) {
	typeName := "share"
	if isUGC {
		typeName = "ugc"
	}

	for _, chunkIDs := range chunk(ids, 100) {
		chunkIDs := chunkIDs
		eg.Go(func() error {
			if err := statsConc.Acquire(ctx, 1); err != nil {
				return nil
			}
			defer statsConc.Release(1)

			var body []byte
			var err error
			if isUGC {
				body, err = li.FetchStatsRaw(ctx, linkedinID, chunkIDs, nil, token)
			} else {
				body, err = li.FetchStatsRaw(ctx, linkedinID, nil, chunkIDs, token)
			}

			if err != nil {
				log.Error().Err(err).Str("linkedin_id", linkedinID).Int(typeName+"_count", len(chunkIDs)).Msgf("failed to fetch linkedin %s stats", typeName)
				return nil
			}

			if body != nil {
				parsed := parseStatsBatch(body)
				mu.Lock()
				for k, v := range parsed {
					statsMap[k] = v
				}
				mu.Unlock()
			}
			return nil
		})
	}
}

// mergeAndPublishPosts combines posts with their stats and assets, then publishes to Kafka.
func mergeAndPublishPosts(
	ctx context.Context,
	producer kafka2.Producer,
	linkedinID string,
	outputTopic string,
	byActivity map[string]*enrichedPost,
	statsMap, imgMap, vidMap, docMap assetMap,
	log *logger.Logger,
) error {
	postsPublished := 0
	for _, ep := range byActivity {
		meta := buildPostMeta(ep, statsMap, imgMap, vidMap, docMap)
		if len(meta) > 0 {
			ep.Post["meta"] = meta
		}

		b, _ := json.Marshal(ep.Post)
		_ = producer.Produce(ctx, outputTopic, []byte(linkedinID), b)
		postsPublished++
	}

	log.Info().
		Str("linkedin_id", linkedinID).
		Str("topic", outputTopic).
		Int("posts_published", postsPublished).
		Msg("produced linkedin posts to kafka")

	return nil
}

// buildPostMeta constructs the meta object containing stats and assets for a post.
func buildPostMeta(ep *enrichedPost, statsMap, imgMap, vidMap, docMap assetMap) map[string]any {
	meta := map[string]any{}

	if st, ok := statsMap[ep.ActivityID]; ok {
		meta["stats"] = st
	}

	assets := map[string]any{}
	if imgs := collectAssetData(ep.ImageIDs, imgMap); len(imgs) > 0 {
		assets["images"] = imgs
	}
	if vids := collectAssetData(ep.VideoIDs, vidMap); len(vids) > 0 {
		assets["videos"] = vids
	}
	if docs := collectAssetData(ep.DocumentIDs, docMap); len(docs) > 0 {
		assets["documents"] = docs
	}
	if len(assets) > 0 {
		meta["assets"] = assets
	}

	return meta
}

// collectAssetData retrieves asset data for a list of IDs from the asset map.
func collectAssetData(ids []string, assetMap assetMap) []any {
	if len(ids) == 0 {
		return nil
	}
	result := make([]any, 0, len(ids))
	for _, id := range ids {
		if data, ok := assetMap[id]; ok {
			result = append(result, data)
		}
	}
	return result
}
