package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	clickhouseRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	mongoRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	clickhouseModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	parser "github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

// Constants defining post fetch limits and lookback period
const (
	PostsLimitIncremental  = 999 // Max posts to fetch during incremental sync
	PostsLimitFull         = 999 // Max posts to fetch during full sync
	DaysToFetchIncremental = 14  // Number of days of posts to fetch
	DaysToFetchFull        = 90  // Number of days of posts to fetch
	SafeFetchLimit         = 50
)

func parseCompetitorDateRange(startDateStr, endDateStr string) (time.Time, time.Time, bool, error) {
	startDateStr = strings.TrimSpace(startDateStr)
	endDateStr = strings.TrimSpace(endDateStr)
	if startDateStr == "" && endDateStr == "" {
		return time.Time{}, time.Time{}, false, nil
	}
	if startDateStr == "" || endDateStr == "" {
		return time.Time{}, time.Time{}, false, fmt.Errorf("start_date and end_date are both required")
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid start_date %q: %w", startDateStr, err)
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid end_date %q: %w", endDateStr, err)
	}
	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, false, fmt.Errorf("end_date must not be before start_date")
	}

	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	return startDate, endDate, true, nil
}

// FetchResult contains the raw data fetched from Facebook API
type FetchResult struct {
	Payload       *apiModels.FacebookCompetitorPayload
	PageDetails   *apiModels.FacebookPageDetails
	AccessToken   string
	Picture       *apiModels.Picture
	PostBatches   [][]*apiModels.Post
	CurrentState  string
	TotalFetched  int
	TotalFiltered int
	Error         error
}

// IsExpectedError returns true if the error is an expected API error (permissions/auth) that should not be sent to Sentry
func (r *FetchResult) IsExpectedError() bool {
	if r.Error == nil {
		return false
	}
	return strings.Contains(r.Error.Error(), "expected api error:")
}

// ParseResult contains the parsed data ready for storage
type ParseResult struct {
	Payload       *apiModels.FacebookCompetitorPayload
	Posts         []*clickhouseModels.FacebookCompetitorPosts
	MediaAssets   []*clickhouseModels.FacebookCompetitorMediaAssets
	Insights      *clickhouseModels.FacebookCompetitorInsights
	ReportID      string // MongoDB report ID for notifications
	CurrentState  string
	DataFlag      bool
	TotalFetched  int
	TotalFiltered int
	Error         error
}

// StoreResult contains the final result after storage
type StoreResult struct {
	PageID         string
	PageName       string
	Success        bool
	TotalProcessed int
	Error          error
}

// FacebookClientInterface defines the interface for Facebook API operations
type FacebookClientInterface interface {
	GetCompetitorPageDetails(ctx context.Context, pageID, accessToken string) (*apiModels.FacebookPageDetails, *apiModels.Picture, error)
	GetCompetitorPosts(ctx context.Context, pageID, accessToken string, since, until time.Time, limit int) ([]*apiModels.Post, string, error)
	GetCompetitorPostsFromURL(ctx context.Context, url, pageID, accessToken string) ([]*apiModels.Post, string, error)
}

// CompetitorRepositoryInterface defines the interface for MongoDB competitor operations
type CompetitorRepositoryInterface interface {
	GetByCompetitorID(ctx context.Context, competitorID string) ([]*mongoModels.Competitor, error)
	UpdateState(ctx context.Context, id interface{}, state string) error
	UpdateField(ctx context.Context, id interface{}, timestamp time.Time) error
	AddError(ctx context.Context, id interface{}, errMsg string) error
}

// ClickHouseRepositoryInterface defines the interface for ClickHouse operations
type ClickHouseRepositoryInterface interface {
	InsertCompetitorInsights(ctx context.Context, insights []*clickhouseModels.FacebookCompetitorInsights) error
	InsertCompetitorPosts(ctx context.Context, posts []*clickhouseModels.FacebookCompetitorPosts) error
	InsertCompetitorMediaAssets(ctx context.Context, assets []*clickhouseModels.FacebookCompetitorMediaAssets) error
}

// CompetitorAnalysisService orchestrates the competitor analysis process
type CompetitorAnalysisService struct {
	fbClient  *social.FacebookClient
	mongoRepo *mongoRepo.CompetitorRepository
	chRepo    *clickhouseRepo.Client
	log       *logger.Logger
}

// NewCompetitorAnalysisService creates a new service instance
func NewCompetitorAnalysisService(
	fbClient *social.FacebookClient,
	mongoRepo *mongoRepo.CompetitorRepository,
	chRepo *clickhouseRepo.Client,
	log *logger.Logger,
) *CompetitorAnalysisService {
	return &CompetitorAnalysisService{
		fbClient:  fbClient,
		mongoRepo: mongoRepo,
		chRepo:    chRepo,
		log:       log,
	}
}

// FetchCompetitorData performs the fetch stage - retrieves raw data from Facebook API
func (s *CompetitorAnalysisService) FetchCompetitorData(ctx context.Context, payload *apiModels.FacebookCompetitorPayload) *FetchResult {
	op := s.log.Operation("FetchCompetitorData").
		WithField("page_id", payload.PageID).
		WithField("page_name", payload.PageName).
		WithSentryTags(map[string]string{
			"page_id":     payload.PageID,
			"page_name":   payload.PageName,
			"sync_status": string(payload.SyncStatus),
			"report_id":   payload.ReportID,
		})

	op.Start("Starting Facebook data fetch")
	result := &FetchResult{
		Payload:     payload,
		PostBatches: [][]*apiModels.Post{},
		AccessToken: payload.AccessToken,
	}

	// Step 1: Fetch competitor metadata from MongoDB
	s.log.Info().Str("step", "GetCompetitorMongo").Str("page_id", payload.PageID).Msg("Fetching competitor details from MongoDB")
	competitors, err := s.mongoRepo.GetByCompetitorID(ctx, payload.PageID)
	if err != nil {
		result.Error = fmt.Errorf("CompetitorAnalysisService.FetchCompetitorData: failed to get competitor from MongoDB: %w", err)
		op.Complete(result.Error, "Failed to get competitor from MongoDB")
		return result
	}

	for _, comp := range competitors {
		result.CurrentState = comp.State
	}

	startDate, endDate, hasRequestedRange, err := parseCompetitorDateRange(payload.StartDate, payload.EndDate)
	if err != nil {
		result.Error = fmt.Errorf("CompetitorAnalysisService.FetchCompetitorData: %w", err)
		op.Complete(result.Error, "Invalid competitor date range")
		return result
	}

	// Step 2: Fetch page details from Facebook API
	s.log.Info().Str("step", "FetchPageDetails").Str("page_id", payload.PageID).Msg("Fetching page details from Facebook API")
	pageDetails, picture, err := s.fbClient.GetCompetitorPageDetails(ctx, payload.PageID, payload.AccessToken)
	if err != nil {
		// Check if this is an expected competitor error (permissions/auth 4xx errors)
		if social.IsExpectedCompetitorErrorFB(err) {
			s.log.Warn().
				Err(err).
				Str("step", "FetchPageDetails").
				Str("page_id", payload.PageID).
				Msg("Expected Facebook API error (permissions/auth) - skipping workload")

			result.Error = fmt.Errorf("CompetitorAnalysisService.FetchCompetitorData: expected api error: %w", err)
			s.updateMongoOnError(ctx, payload.PageID, result.Error, true)
			op.Complete(result.Error, "Expected Facebook API error")
			return result
		}

		// Unexpected error - log as error (will be sent to Sentry)
		result.Error = fmt.Errorf("CompetitorAnalysisService.FetchCompetitorData: failed to fetch page details: %w", err)
		op.Complete(result.Error, "Failed to fetch page details from Facebook")
		return result
	}

	result.PageDetails = pageDetails
	result.Picture = picture

	// Step 3: Determine post fetch range
	since := time.Now().UTC().AddDate(0, 0, -DaysToFetchFull)
	until := time.Now().UTC()
	limit := PostsLimitFull
	if payload.SyncStatus == apiModels.SyncModeIncremental {
		limit = PostsLimitIncremental
		since = time.Now().UTC().AddDate(0, 0, -DaysToFetchIncremental)
	}
	if hasRequestedRange {
		since = startDate
		until = endDate
	}
	fetchLimit := SafeFetchLimit

	s.log.Info().
		Str("step", "DeterminePostRange").
		Time("since", since).
		Time("until", until).
		Int("limit", limit).
		Msg("Post fetch range and limit set")

	// Step 4: Fetch and filter posts
	nextURL := ""
	emptyCount := 0
	fbTimeLayout := "2006-01-02T15:04:05-0700"
	seenPostIDs := make(map[string]bool)
	totalProcessed := 0
	partialCompletion := false
	var lastError error

	for {
		var posts []*apiModels.Post
		var err error

		if nextURL == "" {
			posts, nextURL, err = s.fbClient.GetCompetitorPosts(ctx, payload.PageID, payload.AccessToken, since, until, fetchLimit)
		} else {
			posts, nextURL, err = s.fbClient.GetCompetitorPostsFromURL(ctx, nextURL, payload.PageID, payload.AccessToken)
		}

		if err != nil {
			lastError = err

			// Check if this is an expected competitor error (permissions/auth 4xx errors)
			if social.IsExpectedCompetitorErrorFB(err) {
				s.log.Warn().
					Err(err).
					Str("step", "FetchPosts").
					Str("page_id", payload.PageID).
					Int("posts_processed_so_far", totalProcessed).
					Msg("Expected Facebook API error (permissions/auth) - skipping workload")

				// Return error to mark workload as processed, but it's logged as warning
				result.Error = fmt.Errorf("CompetitorAnalysisService.FetchCompetitorData: expected api error: %w", err)
				op.Complete(result.Error, "Expected Facebook API error")
				return result
			}

			// Unexpected error
			s.log.Error().
				Err(err).
				Str("step", "FetchPosts").
				Str("page_id", payload.PageID).
				Int("posts_processed_so_far", totalProcessed).
				Msg("Failed to fetch posts from Facebook API")

			if totalProcessed > 0 {
				// We have some data, continue with partial result
				s.log.Warn().
					Int("total_processed", totalProcessed).
					Int("total_batches", len(result.PostBatches)).
					Err(err).
					Msg("Partial completion due to API error - continuing with partial data")
				partialCompletion = true
				break
			}

			// No data at all, this is a failure
			result.Error = fmt.Errorf("CompetitorAnalysisService.FetchCompetitorData: failed to fetch posts: %w", err)
			op.Complete(result.Error, "Failed to fetch posts with no data retrieved")
			return result
		}

		result.TotalFetched += len(posts)

		// Handle empty responses
		if len(posts) == 0 {
			emptyCount++
			if emptyCount >= 3 {
				s.log.Warn().
					Str("step", "FetchPosts").
					Str("page_id", payload.PageID).
					Int("total_processed", totalProcessed).
					Msg("Three consecutive empty Facebook post pages, stopping pagination")
				break
			}
			// No more pages to fetch
			if nextURL == "" {
				break
			}
			continue
		}

		// Reset empty counter when we get posts
		emptyCount = 0

		// Filter and validate posts
		var validPosts []*apiModels.Post
		postsOutOfRange := 0
		foundOldPost := false

		for _, post := range posts {
			parsedTime, parseErr := time.Parse(fbTimeLayout, post.CreatedTime)
			if parseErr != nil {
				s.log.Warn().
					Str("raw_created_time", post.CreatedTime).
					Str("post_id", post.ID).
					Msg("Failed to parse CreatedTime, skipping post")
				result.TotalFiltered++
				continue
			}

			if parsedTime.Before(since) {
				postsOutOfRange++
				foundOldPost = true
				break
			}
			if parsedTime.After(until) {
				postsOutOfRange++
				continue
			}

			if seenPostIDs[post.ID] {
				result.TotalFiltered++
				continue
			}

			seenPostIDs[post.ID] = true
			validPosts = append(validPosts, post)
		}

		result.TotalFiltered += postsOutOfRange

		if len(validPosts) > 0 {
			result.PostBatches = append(result.PostBatches, validPosts)
			totalProcessed += len(validPosts)
		}

		// Check exit conditions
		if foundOldPost {
			s.log.Debug().
				Str("page_id", payload.PageID).
				Int("total_processed", totalProcessed).
				Msg("Found post older than date range, stopping pagination")
			break
		}

		if totalProcessed >= limit {
			s.log.Info().
				Str("page_id", payload.PageID).
				Int("total_processed", totalProcessed).
				Int("limit", limit).
				Msg("Reached post limit, stopping pagination")
			break
		}

		if nextURL == "" {
			s.log.Debug().
				Str("page_id", payload.PageID).
				Int("total_processed", totalProcessed).
				Msg("No more pages to fetch, pagination complete")
			break
		}
	}

	// Log completion with appropriate level based on whether it was partial
	if partialCompletion {
		op.WithSentryExtras(map[string]interface{}{
			"total_fetched":  result.TotalFetched,
			"total_filtered": result.TotalFiltered,
			"batches":        len(result.PostBatches),
		})
		op.Complete(nil, "Facebook data fetch completed with partial data due to API error")
		s.log.Warn().
			Str("step", "FetchComplete").
			Str("page_id", payload.PageID).
			Int("total_batches", len(result.PostBatches)).
			Int("total_fetched", result.TotalFetched).
			Int("total_filtered", result.TotalFiltered).
			Bool("partial_completion", true).
			Err(lastError).
			Msg("Completed Facebook data fetch with partial data")
	} else {
		op.WithSentryExtras(map[string]interface{}{
			"total_fetched":  result.TotalFetched,
			"total_filtered": result.TotalFiltered,
			"batches":        len(result.PostBatches),
		})
		op.Complete(nil, "Facebook data fetch completed successfully")
		s.log.Info().
			Str("step", "FetchComplete").
			Str("page_id", payload.PageID).
			Int("total_batches", len(result.PostBatches)).
			Int("total_fetched", result.TotalFetched).
			Int("total_filtered", result.TotalFiltered).
			Msg("Completed Facebook data fetch")
	}

	return result
}

// ParseCompetitorData performs the parse stage - transforms raw data into structured models
func (s *CompetitorAnalysisService) ParseCompetitorData(ctx context.Context, fetchResult *FetchResult) *ParseResult {
	op := s.log.Operation("ParseCompetitorData").
		WithField("page_id", fetchResult.Payload.PageID).
		WithField("page_name", fetchResult.Payload.PageName).
		WithSentryTags(map[string]string{
			"page_id":   fetchResult.Payload.PageID,
			"page_name": fetchResult.Payload.PageName,
			"report_id": fetchResult.Payload.ReportID,
		})

	op.Start("Starting Facebook data parsing")

	result := &ParseResult{
		Payload:       fetchResult.Payload,
		ReportID:      fetchResult.Payload.ReportID,
		CurrentState:  fetchResult.CurrentState,
		Posts:         []*clickhouseModels.FacebookCompetitorPosts{},
		MediaAssets:   []*clickhouseModels.FacebookCompetitorMediaAssets{},
		TotalFetched:  fetchResult.TotalFetched,
		TotalFiltered: fetchResult.TotalFiltered,
	}

	// Check for fetch errors
	if fetchResult.Error != nil {
		result.Error = fetchResult.Error
		op.Complete(result.Error, "Fetch stage had errors")
		return result
	}

	if fetchResult.PageDetails == nil {
		result.Error = fmt.Errorf("CompetitorAnalysisService.ParseCompetitorData: no page details data to parse")
		op.Complete(result.Error, "No data to parse")
		return result
	}

	// Parse page insights
	p := parser.NewFacebookCompetitorParser(fetchResult.Payload.PageID, fetchResult.Payload.PageName, s.fbClient, fetchResult.AccessToken)
	result.Insights = p.ParsePageInsights(fetchResult.PageDetails, fetchResult.Picture)

	// Parse all post batches
	for _, batch := range fetchResult.PostBatches {
		posts, media := p.ParsePosts(ctx, batch, fetchResult.PageDetails)
		result.Posts = append(result.Posts, posts...)
		result.MediaAssets = append(result.MediaAssets, media...)
	}

	result.DataFlag = len(result.Posts) > 0

	op.WithSentryExtras(map[string]interface{}{
		"posts_parsed": len(result.Posts),
	})
	op.Complete(nil, "Facebook data parsing completed successfully")
	s.log.Info().
		Str("step", "ParseComplete").
		Str("page_id", fetchResult.Payload.PageID).
		Int("total_posts_parsed", len(result.Posts)).
		Int("total_media_assets", len(result.MediaAssets)).
		Msg("Completed Facebook data parsing")

	return result
}

// StoreCompetitorData performs the store stage - persists parsed data to databases
func (s *CompetitorAnalysisService) StoreCompetitorData(ctx context.Context, parseResult *ParseResult) *StoreResult {
	op := s.log.Operation("StoreCompetitorData").
		WithField("page_id", parseResult.Payload.PageID).
		WithField("page_name", parseResult.Payload.PageName).
		WithSentryTags(map[string]string{
			"page_id":   parseResult.Payload.PageID,
			"page_name": parseResult.Payload.PageName,
			"report_id": parseResult.ReportID,
		})

	op.Start("Starting Facebook data storage")
	startTime := time.Now()

	result := &StoreResult{
		PageID:   parseResult.Payload.PageID,
		PageName: parseResult.Payload.PageName,
	}

	// Check for parse errors
	if parseResult.Error != nil {
		result.Error = parseResult.Error
		s.updateMongoOnError(ctx, parseResult.Payload.PageID, parseResult.Error, false)
		op.Complete(result.Error, "Parse stage had errors")
		return result
	}

	// Store page insights in ClickHouse
	if parseResult.Insights != nil {
		if err := s.chRepo.InsertCompetitorInsights(ctx, []*clickhouseModels.FacebookCompetitorInsights{parseResult.Insights}); err != nil {
			result.Error = fmt.Errorf("CompetitorAnalysisService.StoreCompetitorData: failed to store insights: %w", err)
			s.log.Error().Err(err).Str("step", "InsertInsights").Msg("Failed to store insights in ClickHouse")
			s.updateMongoOnError(ctx, parseResult.Payload.PageID, result.Error, false)
			op.Complete(result.Error, "Failed to store insights in ClickHouse")
			return result
		}
	}

	// Store posts in ClickHouse
	if len(parseResult.Posts) > 0 {
		if err := s.chRepo.InsertCompetitorPosts(ctx, parseResult.Posts); err != nil {
			result.Error = fmt.Errorf("CompetitorAnalysisService.StoreCompetitorData: failed to store posts: %w", err)
			s.log.Error().Err(err).Str("step", "InsertPosts").Msg("Failed to store posts")
			s.updateMongoOnError(ctx, parseResult.Payload.PageID, result.Error, false)
			op.Complete(result.Error, "Failed to store posts")
			return result
		}
		result.TotalProcessed = len(parseResult.Posts)
	}

	// Store media assets
	if len(parseResult.MediaAssets) > 0 {
		if err := s.chRepo.InsertCompetitorMediaAssets(ctx, parseResult.MediaAssets); err != nil {
			s.log.Error().Err(err).Str("step", "InsertMedia").Msg("Failed to store media assets")
			// Non-fatal error - continue with MongoDB update
		}
	}

	// Update MongoDB with success status
	s.updateMongoObject(
		ctx,
		parseResult.Payload.PageID,
		parseResult.DataFlag,
		time.Now(),
		parseResult.CurrentState,
	)

	result.Success = true
	op.WithSentryExtras(map[string]interface{}{
		"posts_stored": result.TotalProcessed,
		"duration_ms":  time.Since(startTime).Milliseconds(),
	})
	op.Complete(nil, "Facebook data storage completed successfully")
	s.log.Info().
		Str("step", "StoreComplete").
		Str("page_id", parseResult.Payload.PageID).
		Str("page_name", parseResult.Payload.PageName).
		Int("total_posts_stored", result.TotalProcessed).
		Int("total_posts_fetched", parseResult.TotalFetched).
		Int("total_filtered", parseResult.TotalFiltered).
		Dur("duration", time.Since(startTime)).
		Msg("Completed Facebook competitor analysis")

	return result
}

// updateMongoObject updates MongoDB competitor object after successful processing
func (s *CompetitorAnalysisService) updateMongoObject(ctx context.Context, pageID string, dataFlag bool, savingTime time.Time, currentState string) {
	competitors, err := s.mongoRepo.GetByCompetitorID(ctx, pageID)
	if err != nil {
		s.log.Error().Err(err).Str("page_id", pageID).Msg("Failed to get competitor for update")
		return
	}

	for _, comp := range competitors {
		// Update state
		if currentState != mongoModels.StateProcessed && dataFlag {
			if err := s.mongoRepo.UpdateState(ctx, comp.ID, mongoModels.StateProcessed); err != nil {
				s.log.Error().Err(err).Str("competitor_id", comp.ID.Hex()).Msg("Failed to update state")
			}
		} else if currentState == mongoModels.StateAdded && !dataFlag {
			if err := s.mongoRepo.UpdateState(ctx, comp.ID, mongoModels.StateNotFound); err != nil {
				s.log.Error().Err(err).Str("competitor_id", comp.ID.Hex()).Msg("Failed to update state to NotFound")
			}
		}

		// Update timestamp
		if err := s.mongoRepo.UpdateField(ctx, comp.ID, savingTime); err != nil {
			s.log.Error().Err(err).Str("competitor_id", comp.ID.Hex()).Msg("Failed to update timestamp")
		}

		// Clear error
		if err := s.mongoRepo.AddError(ctx, comp.ID, "NoError"); err != nil {
			s.log.Error().Err(err).Str("competitor_id", comp.ID.Hex()).Msg("Failed to clear error")
		}
	}
}

// updateMongoOnError updates MongoDB competitor object on error.
// Set permanent=true for API errors that will never succeed (page not found, permission denied) —
// this marks the competitor state as Failed so the scheduler stops re-dispatching it.
func (s *CompetitorAnalysisService) updateMongoOnError(ctx context.Context, pageID string, processingErr error, permanent bool) {
	competitors, err := s.mongoRepo.GetByCompetitorID(ctx, pageID)
	if err != nil {
		s.log.Error().Err(err).Str("page_id", pageID).Msg("Failed to get competitor for error update")
		return
	}

	errorMsg := processingErr.Error()
	for _, comp := range competitors {
		if err := s.mongoRepo.AddError(ctx, comp.ID, errorMsg); err != nil {
			s.log.Error().Err(err).Str("competitor_id", comp.ID.Hex()).Msg("Failed to add error")
		}

		if err := s.mongoRepo.UpdateField(ctx, comp.ID, time.Now()); err != nil {
			s.log.Error().Err(err).Str("competitor_id", comp.ID.Hex()).Msg("Failed to update timestamp on error")
		}

		if permanent {
			if err := s.mongoRepo.UpdateState(ctx, comp.ID, mongoModels.StateFailed); err != nil {
				s.log.Error().Err(err).Str("competitor_id", comp.ID.Hex()).Msg("Failed to update state to Failed")
			}
		}
	}
}
