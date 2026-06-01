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

// Constants for post fetching limits and default fetch window
const (
	PostsLimitIncremental  = 999 // Max posts to fetch during incremental sync
	PostsLimitFull         = 999 // Max posts to fetch during full sync
	DaysToFetchIncremental = 14  // Number of days of posts to fetch
	DaysToFetchFull        = 90  // Number of days of posts to fetch
	SafeFetchLimit         = 25
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

// FetchResult contains the raw data fetched from Instagram API
type FetchResult struct {
	Payload           *apiModels.InstagramCompetitorPayload
	BusinessDiscovery *apiModels.BusinessDiscovery
	MediaBatches      [][]apiModels.InstagramMedia
	CurrentState      string
	TotalFetched      int
	TotalFiltered     int
	Error             error
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
	Payload       *apiModels.InstagramCompetitorPayload
	Posts         []*clickhouseModels.InstagramCompetitorPosts
	Insights      *clickhouseModels.InstagramCompetitorInsights
	ReportID      string // MongoDB report ID for notifications
	CurrentState  string
	ProfileImage  string
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

// CompetitorAnalysisService orchestrates fetching, processing, and storing Instagram competitor data
type CompetitorAnalysisService struct {
	igClient  *social.InstagramClient
	mongoRepo *mongoRepo.CompetitorRepository
	chRepo    *clickhouseRepo.Client
	log       *logger.Logger
}

// NewCompetitorAnalysisService creates a new Instagram competitor analysis service
func NewCompetitorAnalysisService(
	igClient *social.InstagramClient,
	mongoRepo *mongoRepo.CompetitorRepository,
	chRepo *clickhouseRepo.Client,
	log *logger.Logger,
) *CompetitorAnalysisService {
	return &CompetitorAnalysisService{
		igClient:  igClient,
		mongoRepo: mongoRepo,
		chRepo:    chRepo,
		log:       log,
	}
}

// FetchCompetitorData performs the fetch stage - retrieves raw data from Instagram API
func (s *CompetitorAnalysisService) FetchCompetitorData(ctx context.Context, payload *apiModels.InstagramCompetitorPayload) *FetchResult {
	op := s.log.Operation("FetchCompetitorData").
		WithField("page_id", payload.PageID).
		WithField("page_name", payload.PageName).
		WithSentryTags(map[string]string{
			"page_id":     payload.PageID,
			"page_name":   payload.PageName,
			"sync_status": string(payload.SyncStatus),
			"report_id":   payload.ReportID,
		})

	op.Start("Starting Instagram data fetch")
	result := &FetchResult{
		Payload:      payload,
		MediaBatches: [][]apiModels.InstagramMedia{},
	}

	// Step 1: Fetch competitor metadata from MongoDB
	s.log.Info().Str("step", "GetCompetitorMongo").Str("page_id", payload.PageID).Msg("Fetching competitor details")
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

	// Step 2: Determine fetch range and limit
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

	s.log.Info().
		Str("step", "DeterminePostRange").
		Time("since", since).
		Time("until", until).
		Int("limit", limit).
		Msg("Post fetch range and limit set")

	// Step 3: Fetch business discovery and iterate posts
	var cursor string
	igTimeLayout := "2006-01-02T15:04:05-0700"
	emptyCount := 0
	seenPostIDs := make(map[string]bool)
	totalProcessed := 0

	for {
		response, err := s.igClient.GetBusinessDiscovery(ctx, payload.PageName, SafeFetchLimit, cursor, payload.AccessToken, payload.BusinessID)
		if err != nil {
			// Check if this is an expected competitor error (permissions/auth 4xx errors)
			if social.IsExpectedCompetitorError(err) {
				s.log.Warn().
					Err(err).
					Str("step", "FetchBusinessDiscovery").
					Str("page_id", payload.PageID).
					Str("page_name", payload.PageName).
					Msg("Expected Instagram API error (permissions/auth) - skipping workload")

				result.Error = fmt.Errorf("CompetitorAnalysisService.FetchCompetitorData: expected api error: %w", err)
				s.updateMongoOnError(ctx, payload.PageID, result.Error, true)
				op.Complete(result.Error, "Expected Instagram API error")
				return result
			}

			// Unexpected error - log as error (will be sent to Sentry)
			s.log.Error().Err(err).Str("step", "FetchBusinessDiscovery").Msg("Failed to fetch business discovery")
			if totalProcessed > 0 {
				s.log.Info().Int("total_processed", totalProcessed).Msg("Partial completion allowed due to API error")
				break
			}
			result.Error = fmt.Errorf("CompetitorAnalysisService.FetchCompetitorData: failed to fetch business discovery: %w", err)
			op.Complete(result.Error, "Failed to fetch business discovery")
			return result
		}

		if result.BusinessDiscovery == nil {
			result.BusinessDiscovery = &response.BusinessDiscovery
		}

		mediaBatch := response.BusinessDiscovery.Media.Data
		if len(mediaBatch) == 0 {
			emptyCount++
			if emptyCount >= 3 {
				s.log.Warn().Str("step", "FetchPosts").Msg("3 consecutive empty Instagram media pages, stopping pagination")
				break
			}
			if response.BusinessDiscovery.Media.Paging != nil &&
				response.BusinessDiscovery.Media.Paging.Cursors != nil &&
				response.BusinessDiscovery.Media.Paging.Cursors.After != "" {
				cursor = response.BusinessDiscovery.Media.Paging.Cursors.After
				continue
			}
			break
		}
		emptyCount = 0
		result.TotalFetched += len(mediaBatch)

		// Filter and deduplicate
		validPosts := []apiModels.InstagramMedia{}
		postsOutOfRange := 0
		foundOldPost := false

		for _, post := range mediaBatch {
			postTime, parseErr := time.Parse(igTimeLayout, post.Timestamp)
			if parseErr != nil {
				s.log.Warn().Str("raw_timestamp", post.Timestamp).Str("post_id", post.ID).Msg("Failed to parse timestamp")
				result.TotalFiltered++
				continue
			}

			if postTime.Before(since) {
				postsOutOfRange++
				foundOldPost = true
				break
			}
			if postTime.After(until) {
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
			result.MediaBatches = append(result.MediaBatches, validPosts)
			totalProcessed += len(validPosts)
		}

		if foundOldPost || totalProcessed >= limit {
			break
		}

		if response.BusinessDiscovery.Media.Paging != nil &&
			response.BusinessDiscovery.Media.Paging.Cursors != nil &&
			response.BusinessDiscovery.Media.Paging.Cursors.After != "" {
			cursor = response.BusinessDiscovery.Media.Paging.Cursors.After
		} else {
			break
		}

		if result.BusinessDiscovery.MediaCount > 0 && len(seenPostIDs) >= int(result.BusinessDiscovery.MediaCount) {
			s.log.Info().
				Int("total_seen", len(seenPostIDs)).
				Int("media_count", int(result.BusinessDiscovery.MediaCount)).
				Msg("Reached total media count")
			break
		}
	}

	if result.BusinessDiscovery == nil {
		result.Error = fmt.Errorf("CompetitorAnalysisService.FetchCompetitorData: no business discovery data received")
		op.Complete(result.Error, "Business discovery data missing")
		return result
	}

	op.Complete(nil, "Instagram data fetch completed successfully")
	op.WithSentryExtras(map[string]interface{}{
		"total_fetched":  result.TotalFetched,
		"total_filtered": result.TotalFiltered,
		"batches":        len(result.MediaBatches),
	})
	s.log.Info().
		Str("step", "FetchComplete").
		Str("page_id", payload.PageID).
		Int("total_batches", len(result.MediaBatches)).
		Int("total_fetched", result.TotalFetched).
		Int("total_filtered", result.TotalFiltered).
		Msg("Completed Instagram data fetch")

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

	op.Start("Starting Instagram data parsing")

	result := &ParseResult{
		Payload:       fetchResult.Payload,
		ReportID:      fetchResult.Payload.ReportID,
		CurrentState:  fetchResult.CurrentState,
		Posts:         []*clickhouseModels.InstagramCompetitorPosts{},
		TotalFetched:  fetchResult.TotalFetched,
		TotalFiltered: fetchResult.TotalFiltered,
	}

	// Check for fetch errors
	if fetchResult.Error != nil {
		result.Error = fetchResult.Error
		op.Complete(result.Error, "Fetch stage had errors")
		return result
	}

	if fetchResult.BusinessDiscovery == nil {
		result.Error = fmt.Errorf("CompetitorAnalysisService.ParseCompetitorData: no business discovery data to parse")
		op.Complete(result.Error, "No data to parse")
		return result
	}

	// Parse all media batches
	p := parser.NewInstagramCompetitorParser(
		fetchResult.Payload.PageID,
		fetchResult.Payload.PageName,
		fetchResult.Payload.DisplayName,
	)

	for _, batch := range fetchResult.MediaBatches {
		posts := p.ParsePosts(batch, fetchResult.BusinessDiscovery, fetchResult.BusinessDiscovery.ProfilePictureURL)
		result.Posts = append(result.Posts, posts...)
	}

	// Parse page insights
	result.Insights = p.ParsePageInsights(fetchResult.BusinessDiscovery)
	result.ProfileImage = fetchResult.BusinessDiscovery.ProfilePictureURL
	result.DataFlag = len(result.Posts) > 0 || fetchResult.BusinessDiscovery.MediaCount == 0

	op.Complete(nil, "Instagram data parsing completed successfully")
	op.WithSentryExtras(map[string]interface{}{
		"posts_parsed": len(result.Posts),
	})
	s.log.Info().
		Str("step", "ParseComplete").
		Str("page_id", fetchResult.Payload.PageID).
		Int("total_posts_parsed", len(result.Posts)).
		Msg("Completed Instagram data parsing")

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

	op.Start("Starting Instagram data storage")
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

	// Store posts in ClickHouse
	if len(parseResult.Posts) > 0 {
		if err := s.chRepo.InsertInstagramCompetitorPosts(ctx, parseResult.Posts); err != nil {
			result.Error = fmt.Errorf("CompetitorAnalysisService.StoreCompetitorData: failed to insert posts: %w", err)
			s.log.Error().Err(err).Str("step", "InsertPosts").Msg("Failed to store posts")
			s.updateMongoOnError(ctx, parseResult.Payload.PageID, result.Error, false)
			op.Complete(result.Error, "Failed to insert posts")
			return result
		}
		result.TotalProcessed = len(parseResult.Posts)
	}

	// Store insights in ClickHouse
	if parseResult.Insights != nil {
		if err := s.chRepo.InsertInstagramCompetitorInsights(ctx, []*clickhouseModels.InstagramCompetitorInsights{parseResult.Insights}); err != nil {
			result.Error = fmt.Errorf("CompetitorAnalysisService.StoreCompetitorData: failed to insert insights: %w", err)
			s.log.Error().Err(err).Str("step", "InsertInsights").Msg("Failed to store insights")
			s.updateMongoOnError(ctx, parseResult.Payload.PageID, result.Error, false)
			op.Complete(result.Error, "Failed to insert insights")
			return result
		}
	}

	// Update MongoDB with success status
	s.updateMongoObject(
		ctx,
		parseResult.Payload.PageID,
		parseResult.DataFlag,
		time.Now(),
		parseResult.CurrentState,
		parseResult.ProfileImage,
	)

	result.Success = true
	op.WithSentryExtras(map[string]interface{}{
		"posts_stored": result.TotalProcessed,
		"duration_ms":  time.Since(startTime).Milliseconds(),
	})

	op.Complete(nil, "Instagram data storage completed successfully")
	s.log.Info().
		Str("step", "StoreComplete").
		Str("page_id", parseResult.Payload.PageID).
		Str("page_name", parseResult.Payload.PageName).
		Int("total_posts_stored", result.TotalProcessed).
		Int("total_posts_fetched", parseResult.TotalFetched).
		Int("total_filtered", parseResult.TotalFiltered).
		Dur("duration", time.Since(startTime)).
		Msg("Completed Instagram competitor analysis")

	return result
}

// updateMongoObject updates the MongoDB competitor object after successful processing
func (s *CompetitorAnalysisService) updateMongoObject(ctx context.Context, pageID string, dataFlag bool, savingTime time.Time, currentState string, profileImage string) {
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

		// Update profile image if provided
		if profileImage != "" {
			if err := s.mongoRepo.UpdateImage(ctx, comp.ID, profileImage); err != nil {
				s.log.Error().Err(err).Str("competitor_id", comp.ID.Hex()).Msg("Failed to update image")
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

// updateMongoOnError updates the MongoDB competitor object on error.
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
