package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"golang.org/x/sync/errgroup"
)

// PageProcessor handles page work order processing with injectable dependencies
type PageProcessor struct {
	client      LinkedInAPI
	geoResolver GeoResolverAPI
	producer    kafka2.Producer
	log         *logger.Logger
}

// NewPageProcessor creates a new page processor with injected dependencies
func NewPageProcessor(client LinkedInAPI, geoResolver GeoResolverAPI, producer kafka2.Producer, log *logger.Logger) *PageProcessor {
	return &PageProcessor{
		client:      client,
		geoResolver: geoResolver,
		producer:    producer,
		log:         log,
	}
}

// ProcessPageWorkOrderTestable processes a page work order with interface-based dependencies
func (p *PageProcessor) ProcessPageWorkOrderTestable(
	ctx context.Context,
	order LinkedInAccountWorkOrder,
	token string,
	timestampUpdateChan chan<- TimestampUpdateRequest,
) error {
	log := p.log

	// Calculate date ranges based on sync type
	cutoffTime, startDate, endDate := calculateDateRanges(order.SyncType)

	var tokenErrorDetected int32
	g, gctx := errgroup.WithContext(ctx)

	// Fetch and publish posts
	g.Go(func() error {
		err := p.fetchAndPublishPagePostsTestable(gctx, &order, token, cutoffTime)
		if err != nil && isTokenError(err) {
			atomic.StoreInt32(&tokenErrorDetected, 1)
		}
		return err
	})

	// Fetch and publish insights
	g.Go(func() error {
		err := p.fetchAndPublishPageInsightsTestable(gctx, order.LinkedinID, token, startDate, endDate)
		if err != nil && isTokenError(err) {
			atomic.StoreInt32(&tokenErrorDetected, 1)
		}
		return err
	})

	// Fetch and publish org details
	g.Go(func() error {
		err := p.fetchAndPublishOrgDetailsTestable(gctx, order.LinkedinID, token)
		if err != nil && isTokenError(err) {
			atomic.StoreInt32(&tokenErrorDetected, 1)
		}
		return err
	})

	if err := g.Wait(); err != nil {
		if atomic.LoadInt32(&tokenErrorDetected) == 1 || isTokenError(err) {
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

// fetchAndPublishPagePostsTestable fetches posts and publishes to Kafka
func (p *PageProcessor) fetchAndPublishPagePostsTestable(
	ctx context.Context,
	order *LinkedInAccountWorkOrder,
	token string,
	cutoffTime time.Time,
) error {
	log := p.log

	// Fetch posts
	posts, err := p.client.FetchPostsPaginated(ctx, order.LinkedinID, entityTypeOrganization, token, cutoffTime)
	if err != nil {
		if isTokenError(err) {
			return wrapTokenError(err)
		}
		log.Error().Err(err).Msg("Failed to fetch posts")
		return nil
	}
	log.Info().Int("posts_fetched", len(posts)).Msg("Fetched posts")

	// Parse posts and collect IDs
	byActivity, ugcIDs, shareIDs, imageIDs, videoIDs, documentIDs := parsePostsAndCollectIDs(posts)

	// Fetch stats and assets
	statsMap, imgMap, vidMap, docMap := p.fetchPostAssetsTestable(ctx, order.LinkedinID, token, ugcIDs, shareIDs, imageIDs, videoIDs, documentIDs)

	// Merge and publish
	return p.mergeAndPublishPostsTestable(ctx, order.LinkedinID, byActivity, statsMap, imgMap, vidMap, docMap)
}

// fetchPostAssetsTestable fetches all post assets in parallel
func (p *PageProcessor) fetchPostAssetsTestable(
	ctx context.Context,
	linkedinID string,
	token string,
	ugcIDs, shareIDs, imageIDs, videoIDs, documentIDs []string,
) (statsMap, imgMap, vidMap, docMap assetMap) {
	var statsMu, imgMu, vidMu, docMu sync.Mutex
	statsMap = make(assetMap)
	imgMap = make(assetMap)
	vidMap = make(assetMap)
	docMap = make(assetMap)

	eg, egctx := errgroup.WithContext(ctx)

	// Fetch stats for UGC posts
	for _, chunkIDs := range chunk(ugcIDs, 100) {
		chunkIDs := chunkIDs
		eg.Go(func() error {
			body, err := p.client.FetchStatsRaw(egctx, linkedinID, chunkIDs, nil, token)
			if err != nil {
				return nil
			}
			if body != nil {
				parsed := parseStatsBatch(body)
				statsMu.Lock()
				for k, v := range parsed {
					statsMap[k] = v
				}
				statsMu.Unlock()
			}
			return nil
		})
	}

	// Fetch stats for share posts
	for _, chunkIDs := range chunk(shareIDs, 100) {
		chunkIDs := chunkIDs
		eg.Go(func() error {
			body, err := p.client.FetchStatsRaw(egctx, linkedinID, nil, chunkIDs, token)
			if err != nil {
				return nil
			}
			if body != nil {
				parsed := parseStatsBatch(body)
				statsMu.Lock()
				for k, v := range parsed {
					statsMap[k] = v
				}
				statsMu.Unlock()
			}
			return nil
		})
	}

	// Fetch images
	for _, chunkIDs := range chunk(imageIDs, 80) {
		chunkIDs := chunkIDs
		eg.Go(func() error {
			body, err := p.client.FetchImagesRaw(egctx, chunkIDs, token)
			if err != nil {
				return nil
			}
			if body != nil {
				parsed := parseAssetBatch(body)
				imgMu.Lock()
				for k, v := range parsed {
					imgMap[k] = v
				}
				imgMu.Unlock()
			}
			return nil
		})
	}

	// Fetch videos
	for _, chunkIDs := range chunk(videoIDs, 80) {
		chunkIDs := chunkIDs
		eg.Go(func() error {
			body, err := p.client.FetchVideosRaw(egctx, chunkIDs, token)
			if err != nil {
				return nil
			}
			if body != nil {
				parsed := parseAssetBatch(body)
				vidMu.Lock()
				for k, v := range parsed {
					vidMap[k] = v
				}
				vidMu.Unlock()
			}
			return nil
		})
	}

	// Fetch documents
	for _, chunkIDs := range chunk(documentIDs, 80) {
		chunkIDs := chunkIDs
		eg.Go(func() error {
			body, err := p.client.FetchDocumentsRaw(egctx, chunkIDs, token)
			if err != nil {
				return nil
			}
			if body != nil {
				parsed := parseAssetBatch(body)
				docMu.Lock()
				for k, v := range parsed {
					docMap[k] = v
				}
				docMu.Unlock()
			}
			return nil
		})
	}

	_ = eg.Wait()
	return
}

// mergeAndPublishPostsTestable merges posts with assets and publishes
func (p *PageProcessor) mergeAndPublishPostsTestable(
	ctx context.Context,
	linkedinID string,
	byActivity map[string]*enrichedPost,
	statsMap, imgMap, vidMap, docMap assetMap,
) error {
	for _, ep := range byActivity {
		meta := buildPostMeta(ep, statsMap, imgMap, vidMap, docMap)
		if len(meta) > 0 {
			ep.Post["meta"] = meta
		}

		b, _ := json.Marshal(ep.Post)
		_ = p.producer.Produce(ctx, pagePostsTopic, []byte(linkedinID), b)
	}

	p.log.Info().
		Str("linkedin_id", linkedinID).
		Int("posts_published", len(byActivity)).
		Msg("Published posts to Kafka")

	return nil
}

// fetchAndPublishPageInsightsTestable fetches and publishes page insights
func (p *PageProcessor) fetchAndPublishPageInsightsTestable(
	ctx context.Context,
	linkedinID string,
	token string,
	startDate, endDate time.Time,
) error {
	results, err := p.fetchPageAnalyticsTestable(ctx, linkedinID, token, startDate, endDate)
	if err != nil {
		if isTokenError(err) {
			return wrapTokenError(err)
		}
		return err
	}
	return p.publishPageInsightsTestable(ctx, linkedinID, results)
}

// fetchPageAnalyticsTestable fetches all page analytics in parallel
func (p *PageProcessor) fetchPageAnalyticsTestable(
	ctx context.Context,
	linkedinID string,
	token string,
	startDate, endDate time.Time,
) (*pageAnalyticsResults, error) {
	results := &pageAnalyticsResults{}
	startMs := startDate.UnixMilli()
	endMs := endDate.UnixMilli()
	var tokenErrorCaptured int32

	eg, egctx := errgroup.WithContext(ctx)

	// Fetch follower data with geo resolution
	eg.Go(func() error {
		followerStats, err := p.client.FetchFollowerStatsWithGeoIDs(egctx, linkedinID, token)
		if err != nil {
			if isTokenError(err) {
				atomic.CompareAndSwapInt32(&tokenErrorCaptured, 0, 1)
				return err
			}
			p.log.Error().Err(err).Msg("Failed to fetch follower stats")
			return nil
		}

		var geoNames map[string]string
		if len(followerStats.GeoIDs) > 0 {
			geoNames, _ = p.geoResolver.ResolveGeoIDsWithType(egctx, followerStats.GeoIDs, token)
		}

		data, err := p.client.BuildFollowerDataWithGeoNames(followerStats, geoNames)
		if err != nil {
			p.log.Error().Err(err).Msg("Failed to build follower data")
			return nil
		}
		results.FollowerData = data
		return nil
	})

	// Fetch page statistics
	eg.Go(func() error {
		data, err := p.client.FetchPageStatisticsRaw(egctx, linkedinID, token, startMs, endMs)
		if err != nil {
			if isTokenError(err) {
				atomic.CompareAndSwapInt32(&tokenErrorCaptured, 0, 1)
				return err
			}
			p.log.Error().Err(err).Msg("Failed to fetch page stats")
			return nil
		}
		results.PageStats = data
		return nil
	})

	// Fetch share statistics
	eg.Go(func() error {
		data, err := p.client.FetchShareStatisticsRaw(egctx, linkedinID, token, startMs, endMs)
		if err != nil {
			if isTokenError(err) {
				atomic.CompareAndSwapInt32(&tokenErrorCaptured, 0, 1)
				return err
			}
			p.log.Error().Err(err).Msg("Failed to fetch share stats")
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

// publishPageInsightsTestable publishes page insights to Kafka
func (p *PageProcessor) publishPageInsightsTestable(
	ctx context.Context,
	linkedinID string,
	results *pageAnalyticsResults,
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
	_ = p.producer.Produce(ctx, pageInsightsTopic, []byte(linkedinID), body)

	p.log.Info().
		Str("linkedin_id", linkedinID).
		Bool("has_follower_data", results.FollowerData != nil).
		Bool("has_page_stats", results.PageStats != nil).
		Bool("has_share_stats", results.ShareStats != nil).
		Msg("Published page insights to Kafka")

	return nil
}

// fetchAndPublishOrgDetailsTestable fetches and publishes organization details
func (p *PageProcessor) fetchAndPublishOrgDetailsTestable(
	ctx context.Context,
	linkedinID string,
	token string,
) error {
	data, err := p.client.FetchOrganizationDetailsRaw(ctx, linkedinID, token)
	if err != nil {
		if isTokenError(err) {
			return wrapTokenError(err)
		}
		p.log.Error().Err(err).Msg("Failed to fetch org details")
		return nil
	}

	if data != nil {
		_ = p.producer.Produce(ctx, pageOrganizationTopic, []byte(linkedinID), data)
		p.log.Info().Str("linkedin_id", linkedinID).Msg("Published org details to Kafka")
	}

	return nil
}
