package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// FetcherConfig holds service configuration
type FetcherConfig struct {
	MaxPageWorkers          int
	MaxProfileWorkers       int
	WorkOrderChanSize       int
	TimestampUpdateChanSize int
	MaxConcurrentAccounts   int
	DecryptionKey           string
}

// DefaultFetcherConfig returns default configuration
func DefaultFetcherConfig() FetcherConfig {
	return FetcherConfig{
		MaxPageWorkers:          maxPageWorkers,
		MaxProfileWorkers:       maxProfileWorkers,
		WorkOrderChanSize:       workOrderChanSize,
		TimestampUpdateChanSize: timestampUpdateChanSize,
		MaxConcurrentAccounts:   maxConcurrentAccounts,
	}
}

// FetcherDependencies holds external dependencies
type FetcherDependencies struct {
	PageConsumer    kafka2.Consumer
	ProfileConsumer kafka2.Consumer
	Producer        kafka2.Producer
	MongoRepo       mongodb.UnifiedSocialRepository
	LinkedInClient  *social.LinkedInClient
	GeoResolver     *social.GeoResolver
	Log             *logger.Logger
}

// FetcherMetrics holds service metrics
type FetcherMetrics struct {
	PageBatchesReceived      uint64
	ProfileBatchesReceived   uint64
	PageAccountsProcessed    uint64
	ProfileAccountsProcessed uint64
	TimestampUpdates         uint64
}

// FetcherService represents the LinkedIn fetcher service
type FetcherService struct {
	config  FetcherConfig
	deps    FetcherDependencies
	metrics FetcherMetrics
}

// NewFetcherService creates a new fetcher service
func NewFetcherService(cfg FetcherConfig, deps FetcherDependencies) *FetcherService {
	return &FetcherService{
		config: cfg,
		deps:   deps,
	}
}

// Run starts the fetcher service
func (s *FetcherService) Run(ctx context.Context) error {
	log := s.deps.Log

	timestampUpdateChan := make(chan TimestampUpdateRequest, s.config.TimestampUpdateChanSize)
	var updaterWg sync.WaitGroup
	if s.deps.MongoRepo != nil {
		startTimestampUpdater(ctx, &updaterWg, s.deps.MongoRepo, timestampUpdateChan, log)
	}

	maxConc := s.config.MaxConcurrentAccounts
	if maxConc <= 0 {
		maxConc = maxConcurrentAccounts
	}
	pageSem := semaphore.NewWeighted(int64(maxConc))
	profileSem := semaphore.NewWeighted(int64(maxConc))

	var dispatchWg sync.WaitGroup
	var consumerWg sync.WaitGroup

	if s.deps.PageConsumer != nil {
		consumerWg.Add(1)
		go func() {
			defer consumerWg.Done()
			s.consumePageBatches(ctx, pageSem, &dispatchWg)
		}()
	}
	if s.deps.ProfileConsumer != nil {
		consumerWg.Add(1)
		go func() {
			defer consumerWg.Done()
			s.consumeProfileBatches(ctx, profileSem, &dispatchWg)
		}()
	}

	<-ctx.Done()
	consumerWg.Wait()
	dispatchWg.Wait()
	close(timestampUpdateChan)
	updaterWg.Wait()

	return nil
}

func (s *FetcherService) consumePageBatches(ctx context.Context, sem *semaphore.Weighted, dispatchWg *sync.WaitGroup) {
	log := s.deps.Log

	s.deps.PageConsumer.ConsumeWithAck(ctx, []string{topicWorkOrderPageBatch}, func(ctx context.Context, topic string, key, value []byte, ack func()) error {
		var batch LinkedInBatchWorkOrder
		if err := json.Unmarshal(value, &batch); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal page batch work order")
			ack()
			return nil
		}

		atomic.AddUint64(&s.metrics.PageBatchesReceived, 1)
		log.Info().Str("batch_id", batch.BatchID).Int("accounts", len(batch.Accounts)).Msg("Received page batch, dispatching goroutines")

		for _, account := range batch.Accounts {
			acc := account
			dispatchWg.Add(1)
			go func() {
				defer dispatchWg.Done()
				if err := sem.Acquire(ctx, 1); err != nil {
					return
				}
				defer sem.Release(1)

				perAccSem := semForAccount(acc.LinkedinID)
				if err := perAccSem.Acquire(ctx, 1); err != nil {
					return
				}
				defer perAccSem.Release(1)

				atomic.AddUint64(&s.metrics.PageAccountsProcessed, 1)
			}()
		}
		ack()
		return nil
	})
}

func (s *FetcherService) consumeProfileBatches(ctx context.Context, sem *semaphore.Weighted, dispatchWg *sync.WaitGroup) {
	log := s.deps.Log

	s.deps.ProfileConsumer.ConsumeWithAck(ctx, []string{topicWorkOrderProfileBatch}, func(ctx context.Context, topic string, key, value []byte, ack func()) error {
		var batch LinkedInBatchWorkOrder
		if err := json.Unmarshal(value, &batch); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal profile batch work order")
			ack()
			return nil
		}

		atomic.AddUint64(&s.metrics.ProfileBatchesReceived, 1)
		log.Info().Str("batch_id", batch.BatchID).Int("accounts", len(batch.Accounts)).Msg("Received profile batch, dispatching goroutines")

		for _, account := range batch.Accounts {
			acc := account
			dispatchWg.Add(1)
			go func() {
				defer dispatchWg.Done()
				if err := sem.Acquire(ctx, 1); err != nil {
					return
				}
				defer sem.Release(1)

				perAccSem := semForAccount(acc.LinkedinID)
				if err := perAccSem.Acquire(ctx, 1); err != nil {
					return
				}
				defer perAccSem.Release(1)

				atomic.AddUint64(&s.metrics.ProfileAccountsProcessed, 1)
			}()
		}
		ack()
		return nil
	})
}

func (s *FetcherService) pageWorkerLoop(ctx context.Context, workerID int, jobs <-chan WorkOrderMessage, timestampUpdateChan chan<- TimestampUpdateRequest) {
	log := s.deps.Log
	workerLog := log.With().Int("worker_id", workerID).Str("pool", "page").Logger()
	workerLog.Info().Msg("Page worker started")

	for {
		select {
		case <-ctx.Done():
			workerLog.Info().Msg("Context canceled; page worker stopping")
			return
		case wo, ok := <-jobs:
			if !ok {
				workerLog.Info().Msg("Page work order channel closed; worker stopping")
				return
			}

			sem := semForAccount(wo.LinkedinID)
			if err := sem.Acquire(ctx, 1); err != nil {
				workerLog.Error().Err(err).Str("linkedin_id", wo.LinkedinID).Msg("Failed to acquire per-account semaphore; skipping page work order")
				continue
			}

			// Process work order (actual processing would call LinkedIn API)
			workerLog.Debug().Str("linkedin_id", wo.LinkedinID).Msg("Processing page work order")

			sem.Release(1)
			if wo.Ack != nil {
				wo.Ack()
			}
		}
	}
}

func (s *FetcherService) profileWorkerLoop(ctx context.Context, workerID int, jobs <-chan WorkOrderMessage, timestampUpdateChan chan<- TimestampUpdateRequest) {
	log := s.deps.Log
	workerLog := log.With().Int("worker_id", workerID).Str("pool", "profile").Logger()
	workerLog.Info().Msg("Profile worker started")

	for {
		select {
		case <-ctx.Done():
			workerLog.Info().Msg("Context canceled; profile worker stopping")
			return
		case wo, ok := <-jobs:
			if !ok {
				workerLog.Info().Msg("Profile work order channel closed; worker stopping")
				return
			}

			sem := semForAccount(wo.LinkedinID)
			if err := sem.Acquire(ctx, 1); err != nil {
				workerLog.Error().Err(err).Str("linkedin_id", wo.LinkedinID).Msg("Failed to acquire per-account semaphore; skipping profile work order")
				if wo.Ack != nil {
					wo.Ack()
				}
				continue
			}

			// Process work order (actual processing would call LinkedIn API)
			workerLog.Debug().Str("linkedin_id", wo.LinkedinID).Msg("Processing profile work order")

			sem.Release(1)
			if wo.Ack != nil {
				wo.Ack()
			}
		}
	}
}

// GetMetrics returns current service metrics
func (s *FetcherService) GetMetrics() FetcherMetrics {
	return FetcherMetrics{
		PageBatchesReceived:      atomic.LoadUint64(&s.metrics.PageBatchesReceived),
		ProfileBatchesReceived:   atomic.LoadUint64(&s.metrics.ProfileBatchesReceived),
		PageAccountsProcessed:    atomic.LoadUint64(&s.metrics.PageAccountsProcessed),
		ProfileAccountsProcessed: atomic.LoadUint64(&s.metrics.ProfileAccountsProcessed),
		TimestampUpdates:         atomic.LoadUint64(&s.metrics.TimestampUpdates),
	}
}

// FetchPostsWithClient fetches posts using the provided LinkedIn API client
func FetchPostsWithClient(
	ctx context.Context,
	client LinkedInAPI,
	linkedinID string,
	entityType string,
	accessToken string,
	cutoffTime time.Time,
) ([]json.RawMessage, error) {
	return client.FetchPostsPaginated(ctx, linkedinID, entityType, accessToken, cutoffTime)
}

// FetchInsightsWithClient fetches follower insights using the provided LinkedIn API client
func FetchInsightsWithClient(
	ctx context.Context,
	client LinkedInAPI,
	linkedinID string,
	accessToken string,
) ([]byte, error) {
	return client.FetchFollowerData(ctx, linkedinID, accessToken)
}

// FetchOrganizationWithClient fetches organization details using the provided LinkedIn API client
func FetchOrganizationWithClient(
	ctx context.Context,
	client LinkedInAPI,
	linkedinID string,
	accessToken string,
) ([]byte, error) {
	return client.FetchOrganizationDetailsRaw(ctx, linkedinID, accessToken)
}

// FetchPageStatisticsWithClient fetches page statistics using the provided LinkedIn API client
func FetchPageStatisticsWithClient(
	ctx context.Context,
	client LinkedInAPI,
	linkedinID string,
	accessToken string,
	startMs, endMs int64,
) ([]byte, error) {
	return client.FetchPageStatisticsRaw(ctx, linkedinID, accessToken, startMs, endMs)
}

// FetchShareStatisticsWithClient fetches share statistics using the provided LinkedIn API client
func FetchShareStatisticsWithClient(
	ctx context.Context,
	client LinkedInAPI,
	linkedinID string,
	accessToken string,
	startMs, endMs int64,
) ([]byte, error) {
	return client.FetchShareStatisticsRaw(ctx, linkedinID, accessToken, startMs, endMs)
}

// ProcessPageWorkOrderWithClient processes a page work order using mocked dependencies
func ProcessPageWorkOrderWithClient(
	ctx context.Context,
	client LinkedInAPI,
	geoResolver GeoResolverAPI,
	producer kafka2.Producer,
	order LinkedInAccountWorkOrder,
	accessToken string,
) (*PageFetchResult, error) {
	result := &PageFetchResult{}

	// Fetch posts
	cutoffTime := time.Now().AddDate(0, -12, 0)
	posts, err := client.FetchPostsPaginated(ctx, order.LinkedinID, "organization", accessToken, cutoffTime)
	if err != nil {
		return nil, err
	}
	result.PostsCount = len(posts)

	// Fetch insights
	insightsData, err := client.FetchFollowerData(ctx, order.LinkedinID, accessToken)
	if err != nil {
		return nil, err
	}
	result.HasInsights = len(insightsData) > 0

	// Fetch organization details
	orgData, err := client.FetchOrganizationDetailsRaw(ctx, order.LinkedinID, accessToken)
	if err != nil {
		return nil, err
	}
	result.HasOrgDetails = len(orgData) > 0

	return result, nil
}

// PageFetchResult represents the result of fetching page data
type PageFetchResult struct {
	PostsCount    int
	HasInsights   bool
	HasOrgDetails bool
}

// ProcessProfileWorkOrderWithClient processes a profile work order using mocked dependencies
func ProcessProfileWorkOrderWithClient(
	ctx context.Context,
	client LinkedInAPI,
	producer kafka2.Producer,
	order LinkedInAccountWorkOrder,
	accessToken string,
) (*ProfileFetchResult, error) {
	result := &ProfileFetchResult{}

	// Fetch member analytics
	startDate := time.Now().AddDate(0, -12, 0)
	endDate := time.Now()

	// Fetch post analytics
	for _, queryType := range []string{"POST_LEVEL", "AGGREGATE"} {
		data, err := client.FetchMemberCreatorPostAnalyticsRaw(ctx, accessToken, queryType, &startDate, &endDate)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			result.AnalyticsCount++
		}
	}

	// Fetch follower count
	followerData, err := client.FetchMemberFollowersCountRaw(ctx, accessToken, &startDate, &endDate)
	if err != nil {
		return nil, err
	}
	result.HasFollowerData = len(followerData) > 0

	return result, nil
}

// ProfileFetchResult represents the result of fetching profile data
type ProfileFetchResult struct {
	AnalyticsCount  int
	HasFollowerData bool
}

// RawPagePostsData represents raw post data for Kafka
type RawPagePostsData struct {
	AccountID   string            `json:"account_id"`
	WorkspaceID string            `json:"workspace_id"`
	LinkedinID  string            `json:"linkedin_id"`
	Posts       []json.RawMessage `json:"posts"`
	Stats       []byte            `json:"stats"`
	Images      []byte            `json:"images"`
	Videos      []byte            `json:"videos"`
	Documents   []byte            `json:"documents"`
}

// RawPageInsightsData represents raw insights data for Kafka
type RawPageInsightsData struct {
	AccountID    string `json:"account_id"`
	WorkspaceID  string `json:"workspace_id"`
	LinkedinID   string `json:"linkedin_id"`
	FollowerData []byte `json:"follower_data"`
	PageStats    []byte `json:"page_stats"`
	ShareStats   []byte `json:"share_stats"`
	OrgDetails   []byte `json:"org_details"`
}

// BuildRawPagePosts builds a RawPagePostsData message for Kafka
func BuildRawPagePosts(
	accountID string,
	workspaceID string,
	linkedinID string,
	posts []json.RawMessage,
	statsData []byte,
	imagesData []byte,
	videosData []byte,
	documentsData []byte,
) *RawPagePostsData {
	return &RawPagePostsData{
		AccountID:   accountID,
		WorkspaceID: workspaceID,
		LinkedinID:  linkedinID,
		Posts:       posts,
		Stats:       statsData,
		Images:      imagesData,
		Videos:      videosData,
		Documents:   documentsData,
	}
}

// BuildRawPageInsights builds a RawPageInsightsData message for Kafka
func BuildRawPageInsights(
	accountID string,
	workspaceID string,
	linkedinID string,
	followerData []byte,
	pageStatsData []byte,
	shareStatsData []byte,
	orgDetailsData []byte,
) *RawPageInsightsData {
	return &RawPageInsightsData{
		AccountID:    accountID,
		WorkspaceID:  workspaceID,
		LinkedinID:   linkedinID,
		FollowerData: followerData,
		PageStats:    pageStatsData,
		ShareStats:   shareStatsData,
		OrgDetails:   orgDetailsData,
	}
}
