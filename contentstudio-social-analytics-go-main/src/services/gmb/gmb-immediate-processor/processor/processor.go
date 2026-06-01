package processor

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/notification"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const gmbImmediateFallbackDays = 90

// ImmediateWorkOrder mirrors GMBAccountWorkOrder for immediate processing.
type ImmediateWorkOrder struct {
	ID           string `json:"id"`            // MongoDB _id (hex)
	WorkspaceID  string `json:"workspace_id"`  // Workspace identifier
	AccountID    string `json:"account_id"`    // GMB account ID
	LocationID   string `json:"location_id"`   // GMB location ID
	AccessToken  string `json:"access_token"`  // OAuth access token (encrypted)
	RefreshToken string `json:"refresh_token"` // OAuth refresh token (encrypted)
	AccountName  string `json:"account_name"`  // Display name of the account
	LocationName string `json:"location_name"` // Display name of the location
	LanguageCode string `json:"language_code"` // Language code (e.g., "en")
	SyncType     string `json:"sync_type"`     // "incremental" | "full_sync"
	StartDate    string `json:"start_date,omitempty"`
	EndDate      string `json:"end_date,omitempty"`
}

// Processor handles GMB immediate account processing.
type Processor struct {
	gmbClient    GMBClientInterface
	sink         ClickHouseSinkInterface
	mongoRepo    mongodb.UnifiedSocialRepository
	notifier     *notification.Service
	pusherClient *notification.PusherClient
	log          *logger.Logger
	cfg          *config.Config
}

// New creates a new GMB Processor with all dependencies.
// The GMB API client is created internally using config credentials with default rate limits.
func New(
	mongoRepo mongodb.UnifiedSocialRepository,
	sink ClickHouseSinkInterface,
	notifier *notification.Service,
	pusherClient *notification.PusherClient,
	log *logger.Logger,
	cfg *config.Config,
) *Processor {
	return &Processor{
		gmbClient:    social.NewGMBClient(cfg.GMB.ClientID, cfg.GMB.ClientSecret),
		sink:         sink,
		mongoRepo:    mongoRepo,
		notifier:     notifier,
		pusherClient: pusherClient,
		log:          log,
		cfg:          cfg,
	}
}

// ProcessAccount implements the GMBProcessor interface using the concrete dependencies.
func (p *Processor) ProcessAccount(ctx context.Context, wo ImmediateWorkOrder) error {
	return ProcessAccount(ctx, p.gmbClient, p.sink, p.mongoRepo, p.notifier, p.pusherClient, wo, p.cfg.DecryptionKey, time.Now().UTC(), p.log)
}

// ProcessAccount processes a GMB account: refreshes the token, fetches Voice of Merchant status
// to gate access to performance metrics and search keywords APIs, then fetches remaining data types
// (local posts, reviews, media assets), converts them to ClickHouse models, inserts into ClickHouse,
// updates MongoDB state (including has_voice_of_merchant), and sends notifications.
// All fetch errors are logged as warnings only and do not halt processing.
// Retry logic is handled by the GMB HTTP client.
func ProcessAccount(
	ctx context.Context,
	gmbClient GMBClientInterface,
	sink ClickHouseSinkInterface,
	mongoRepo interface{}, // UnifiedSocialRepository
	notifier *notification.Service,
	pusherClient *notification.PusherClient,
	wo ImmediateWorkOrder,
	decryptionKey string,
	now time.Time,
	log *logger.Logger,
) (err error) {
	op := log.Operation("ProcessGMBAccount").
		WithField("workspace_id", wo.WorkspaceID).
		WithField("account_id", wo.AccountID).
		WithField("platform_identifier", fmt.Sprintf("accounts/%s/locations/%s", wo.AccountID, wo.LocationID)).
		WithField("location_id", wo.LocationID).
		WithField("sync_type", wo.SyncType).
		WithField("start_date", wo.StartDate).
		WithField("end_date", wo.EndDate).
		WithSentryTags(map[string]string{
			"workspace_id": wo.WorkspaceID,
			"account_id":   wo.AccountID,
			"location_id":  wo.LocationID,
			"sync_type":    wo.SyncType,
		})
	op.Start("processing gmb work order")

	var totalInserted int
	defer func() {
		op.WithField("total_inserted", totalInserted).
			Complete(nil, "")
	}()

	// 1. Fetch account from MongoDB if ID is provided
	var account *mongomodels.SocialIntegration
	var originalState string
	var accountID primitive.ObjectID
	hasAccountID := false
	if wo.ID != "" {
		accountID, err = primitive.ObjectIDFromHex(wo.ID)
		if err != nil {
			log.Warn().Err(err).Str("account_id", wo.ID).Msg("Invalid account ID, continuing without MongoDB data")
		} else {
			hasAccountID = true
			repo, ok := mongoRepo.(interface {
				FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error)
			})
			if ok {
				acct, err := repo.FindByID(ctx, accountID)
				if err != nil {
					log.Warn().Err(err).Str("account_id", wo.ID).Msg("Failed to fetch account from MongoDB, continuing without")
				} else if acct != nil {
					if mongodb.HasProcessingErrorMeta(acct.MetaData) {
						if clearRepo, ok := mongoRepo.(interface {
							ClearProcessingError(ctx context.Context, id primitive.ObjectID) error
						}); ok {
							if clearErr := clearRepo.ClearProcessingError(ctx, accountID); clearErr != nil {
								log.Warn().Err(clearErr).Str("account_id", wo.ID).Msg("Failed to clear stale processing error before retry")
							}
						}
					}
					account = acct
					originalState = acct.State
					log.Info().
						Str("account_id", wo.ID).
						Str("platform_identifier", acct.PlatformIdentifier).
						Str("state", acct.State).
						Msg("Fetched account from MongoDB")
				}
			}
		}
	}
	defer func() {
		if !hasAccountID || err == nil {
			return
		}
		if repo, ok := mongoRepo.(interface {
			RecordProcessingError(ctx context.Context, id primitive.ObjectID, errorMessage string) error
		}); ok {
			if recordErr := repo.RecordProcessingError(ctx, accountID, err.Error()); recordErr != nil {
				log.Warn().Err(recordErr).Str("account_id", wo.ID).Msg("Failed to record processing error")
			}
		}
	}()

	// 2. Decrypt tokens
	accessToken := wo.AccessToken
	if dec, err := crypto.DecryptToken(accessToken, decryptionKey); err == nil {
		accessToken = dec
	}
	refreshToken := wo.RefreshToken
	if dec, err := crypto.DecryptToken(refreshToken, decryptionKey); err == nil {
		refreshToken = dec
	}

	// 3. Refresh access token
	tokenResp, err := gmbClient.RefreshToken(ctx, refreshToken)
	if err != nil {
		if social.IsExpectedCompetitorErrorGMB(err) {
			log.Warn().Err(err).Str("location_id", wo.LocationID).Msg("Expected token error, skipping account")
			return nil
		}
		return fmt.Errorf("ProcessGMBAccount: refresh token: %w", err)
	}
	accessToken = tokenResp.AccessToken

	now = now.UTC()
	accountName := wo.AccountName
	locationName := wo.LocationName
	startDate, endDate, _, err := resolveGMBDateRange(wo.StartDate, wo.EndDate, now)
	if err != nil {
		return err
	}

	// 4. Fetch Voice of Merchant status — determines if perf metrics & search keywords are available
	hasVoiceOfMerchant := false
	voMResp, err := gmbClient.FetchVoiceOfMerchant(ctx, wo.LocationID, accessToken)
	if err != nil {
		log.Warn().Err(err).Str("location_id", wo.LocationID).Msg("Failed to fetch VoM, defaulting hasVoiceOfMerchant to false")
	}
	if voMResp != nil {
		hasVoiceOfMerchant = voMResp.HasVoiceOfMerchant
	}

	var metricsInserted, keywordsInserted int

	// 5. Fetch and insert performance metrics — only if hasVoiceOfMerchant
	if hasVoiceOfMerchant {
		metricsInserted = fetchAndInsertPerformanceMetrics(ctx, gmbClient, sink, wo, accessToken, accountName, locationName, startDate, endDate, log)
		totalInserted += metricsInserted

		// 6. Fetch and insert search keywords — only if hasVoiceOfMerchant
		keywordsInserted = fetchAndInsertSearchKeywords(ctx, gmbClient, sink, wo, accessToken, accountName, locationName, startDate, endDate, log)
		totalInserted += keywordsInserted
	} else {
		log.Info().Str("location_id", wo.LocationID).Msg("Skipping performance metrics and search keywords (hasVoiceOfMerchant=false)")
	}

	// 7. Fetch and insert local posts
	postsInserted := fetchAndInsertLocalPosts(ctx, gmbClient, sink, wo, accessToken, accountName, locationName, startDate, endDate, log)
	totalInserted += postsInserted

	// 8. Fetch and insert reviews
	reviewsInserted := fetchAndInsertReviews(ctx, gmbClient, sink, wo, accessToken, accountName, locationName, startDate, endDate, log)
	totalInserted += reviewsInserted

	// 9. Fetch and insert media assets
	assetsInserted := fetchAndInsertMediaAssets(ctx, gmbClient, sink, wo, accessToken, accountName, locationName, startDate, endDate, log)
	totalInserted += assetsInserted

	log.Info().
		Int("metrics_inserted", metricsInserted).
		Int("keywords_inserted", keywordsInserted).
		Int("posts_inserted", postsInserted).
		Int("reviews_inserted", reviewsInserted).
		Int("assets_inserted", assetsInserted).
		Int("total_inserted", totalInserted).
		Bool("has_voice_of_merchant", hasVoiceOfMerchant).
		Msg("GMB data fetch and insert complete")

	// 10. Update MongoDB state to "Processed" + has_voice_of_merchant
	if wo.ID != "" {
		if objID, err := primitive.ObjectIDFromHex(wo.ID); err == nil {
			if repo, ok := mongoRepo.(interface {
				Update(ctx context.Context, id primitive.ObjectID, updates bson.M) error
				ClearProcessingError(ctx context.Context, id primitive.ObjectID) error
			}); ok {
				updates := bson.M{
					"state":                     mongomodels.StateProcessed,
					"last_analytics_fetched_at": now,
					"last_analytics_updated_at": now.Format("2006-01-02 15:04:05"),
					"has_voice_of_merchant":     hasVoiceOfMerchant,
				}
				if err := repo.Update(ctx, objID, updates); err != nil {
					log.Warn().Err(err).Str("account_id", wo.ID).Msg("Failed to update account state to Processed")
				} else {
					log.Info().Str("account_id", wo.ID).Str("state", mongomodels.StateProcessed).Bool("has_voice_of_merchant", hasVoiceOfMerchant).Msg("Updated account state to Processed")
				}
				if err := repo.ClearProcessingError(ctx, objID); err != nil {
					log.Warn().Err(err).Str("account_id", wo.ID).Msg("Failed to clear processing error")
				}
			}
		}
	}

	// 11. Send notifications
	if pusherClient != nil {
		SendNotifications(ctx, mongoRepo, notifier, pusherClient, wo, account, originalState, log)
	}

	return nil
}

// computePerformanceDateRange returns the [startDate, endDate] window for fetching performance metrics.
// The window is at most 90 days back from now. If account.CreatedAt falls within that window,
// it is used as the start so we never request data before the account existed.
func computePerformanceDateRange(account *mongomodels.SocialIntegration, now time.Time) (startDate, endDate time.Time) {
	now = now.UTC()
	endDate = now.Truncate(24 * time.Hour).Add(24*time.Hour - time.Nanosecond)
	fallback := now.AddDate(0, 0, -90).Truncate(24 * time.Hour)
	startDate = fallback
	if account != nil && account.CreatedAt != nil && !account.CreatedAt.IsZero() {
		created := account.CreatedAt.UTC().Truncate(24 * time.Hour)
		if created.After(fallback) {
			startDate = created
		}
	}
	return
}

func parseGMBDateRange(startDateStr, endDateStr string) (time.Time, time.Time, bool, error) {
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

func resolveGMBDateRange(startDateStr, endDateStr string, now time.Time) (time.Time, time.Time, bool, error) {
	startDate, endDate, hasRequestedRange, err := parseGMBDateRange(startDateStr, endDateStr)
	if err != nil || hasRequestedRange {
		return startDate, endDate, hasRequestedRange, err
	}

	fallbackEnd := now.UTC()
	fallbackStart := fallbackEnd.AddDate(0, 0, -gmbImmediateFallbackDays)
	return time.Date(fallbackStart.Year(), fallbackStart.Month(), fallbackStart.Day(), 0, 0, 0, 0, time.UTC),
		time.Date(fallbackEnd.Year(), fallbackEnd.Month(), fallbackEnd.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC),
		false,
		nil
}

func monthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func monthWindowsForRange(startDate, endDate time.Time) []time.Time {
	startMonth := monthStart(startDate)
	endMonth := monthStart(endDate)
	var months []time.Time
	for current := startMonth; !current.After(endMonth); current = current.AddDate(0, 1, 0) {
		months = append(months, current)
	}
	return months
}

func parseGMBAPITime(value string) (time.Time, bool) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, false
	}
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed.UTC(), true
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC(), true
	}
	return time.Time{}, false
}

func withinGMBDateRange(ts, startDate, endDate time.Time) bool {
	if startDate.IsZero() || endDate.IsZero() {
		return true
	}
	return !ts.Before(startDate) && !ts.After(endDate)
}

// fetchAndInsertPerformanceMetrics fetches daily metrics for the effective date range and inserts into ClickHouse.
func fetchAndInsertPerformanceMetrics(
	ctx context.Context,
	gmbClient GMBClientInterface,
	sink ClickHouseSinkInterface,
	wo ImmediateWorkOrder,
	accessToken, accountName, locationName string,
	startDate, endDate time.Time,
	log *logger.Logger,
) int {
	var allMetrics []*chmodels.GMBDailyMetrics

	resp, err := gmbClient.FetchPerformanceMetrics(ctx, wo.LocationID, accessToken, startDate, endDate)
	if err != nil {
		log.Warn().Err(err).Str("location_id", wo.LocationID).Msg("Failed to fetch performance metrics")
	} else {
		allMetrics = append(allMetrics, buildDailyMetricsFromResponse(resp, wo.AccountID, wo.LocationID, accountName, locationName)...)
	}

	if len(allMetrics) == 0 {
		return 0
	}

	if err := sink.BulkInsertGMBDailyMetrics(ctx, allMetrics); err != nil {
		log.Warn().Err(err).Int("count", len(allMetrics)).Msg("Failed to insert GMB daily metrics")
		return 0
	}
	log.Info().Int("count", len(allMetrics)).Msg("Inserted GMB daily metrics")
	return len(allMetrics)
}

func buildDailyMetricsFromResponse(resp *social.GMBPerformanceResponse, accountID, locationID, accountName, locationName string) []*chmodels.GMBDailyMetrics {
	var allMetrics []*chmodels.GMBDailyMetrics
	if resp == nil {
		return nil
	}

	builders := make(map[string]*conversions.GMBDailyMetricsBuilder)
	for _, multi := range resp.MultiDailyMetricTimeSeries {
		for _, dts := range multi.DailyMetricTimeSeries {
			metric := dts.DailyMetric
			for _, dv := range dts.TimeSeries.DatedValues {
				dateStr := fmt.Sprintf("%04d-%02d-%02d", dv.Date.Year, dv.Date.Month, dv.Date.Day)
				b, ok := builders[dateStr]
				if !ok {
					b = &conversions.GMBDailyMetricsBuilder{
						AccountID:  accountID,
						LocationID: locationID,
						Date:       dateStr,
					}
					builders[dateStr] = b
				}
				b.SetMetric(metric, dv.Value)
			}
		}
	}

	for _, b := range builders {
		m := b.Build()
		m.AccountName = accountName
		m.LocationName = locationName
		allMetrics = append(allMetrics, m)
	}
	return allMetrics
}

// fetchAndInsertSearchKeywords fetches monthly search keywords for the effective date range and inserts into ClickHouse.
func fetchAndInsertSearchKeywords(
	ctx context.Context,
	gmbClient GMBClientInterface,
	sink ClickHouseSinkInterface,
	wo ImmediateWorkOrder,
	accessToken, accountName, locationName string,
	startDate, endDate time.Time,
	log *logger.Logger,
) int {
	var allKeywords []*chmodels.GMBSearchKeywordsMonthly

	for idx, month := range monthWindowsForRange(startDate, endDate) {
		resp, err := gmbClient.FetchSearchKeywords(ctx, wo.LocationID, accessToken, month, month)
		if err != nil {
			log.Warn().Err(err).Str("location_id", wo.LocationID).Time("month", month).Int("month_offset", idx).Msg("Failed to fetch search keywords")
			continue
		}

		allKeywords = append(allKeywords, buildSearchKeywords(accountName, locationName, wo, month.Format("2006-01"), resp)...)
	}

	if len(allKeywords) == 0 {
		return 0
	}

	if err := sink.BulkInsertGMBSearchKeywordsMonthly(ctx, allKeywords); err != nil {
		log.Warn().Err(err).Int("count", len(allKeywords)).Msg("Failed to insert GMB search keywords")
		return 0
	}
	log.Info().Int("count", len(allKeywords)).Msg("Inserted GMB search keywords")
	return len(allKeywords)
}

func buildSearchKeywords(accountName, locationName string, wo ImmediateWorkOrder, keywordMonth string, resp *social.GMBSearchKeywordsResponse) []*chmodels.GMBSearchKeywordsMonthly {
	var allKeywords []*chmodels.GMBSearchKeywordsMonthly
	if resp == nil {
		return nil
	}
	for _, sk := range resp.SearchKeywordsCounts {
		impVal, _ := strconv.ParseUint(sk.InsightsValue.Value, 10, 64)
		impThresh, _ := strconv.ParseUint(sk.InsightsValue.Threshold, 10, 64)

		parsed := &kafkamodels.ParsedGMBSearchKeyword{
			AccountID:            wo.AccountID,
			LocationID:           wo.LocationID,
			AccountName:          accountName,
			LocationName:         locationName,
			KeywordMonth:         keywordMonth,
			Keyword:              sk.SearchKeyword,
			ImpressionsValue:     impVal,
			ImpressionsThreshold: impThresh,
		}
		chKw := conversions.ConvertGMBSearchKeyword(parsed)
		if chKw != nil {
			allKeywords = append(allKeywords, chKw)
		}
	}
	return allKeywords
}

// fetchAndInsertLocalPosts fetches local posts and inserts into ClickHouse.
func fetchAndInsertLocalPosts(
	ctx context.Context,
	gmbClient GMBClientInterface,
	sink ClickHouseSinkInterface,
	wo ImmediateWorkOrder,
	accessToken, accountName, locationName string,
	startDate, endDate time.Time,
	log *logger.Logger,
) int {
	var allPosts []*chmodels.GMBLocalPosts
	pageToken := ""
	for page := 0; ; page++ {
		resp, err := gmbClient.FetchLocalPosts(ctx, wo.AccountID, wo.LocationID, accessToken, pageToken)
		if err != nil {
			log.Warn().Err(err).Str("location_id", wo.LocationID).Int("page", page).Msg("Failed to fetch local posts")
			break
		}

		reachedOlderThanStart := false
		pageHasInRange := false
		for _, lp := range resp.LocalPosts {
			createdAt, ok := parseGMBAPITime(lp.CreateTime)
			if !ok {
				continue
			}
			if createdAt.Before(startDate) {
				reachedOlderThanStart = true
				continue
			}
			if !withinGMBDateRange(createdAt, startDate, endDate) {
				continue
			}
			pageHasInRange = true

			var mediaNames, mediaFormats, mediaGoogleURLs []string
			for _, m := range lp.Media {
				mediaNames = append(mediaNames, m.Name)
				mediaFormats = append(mediaFormats, m.MediaFormat)
				mediaGoogleURLs = append(mediaGoogleURLs, m.GoogleURL)
			}

			parsed := &kafkamodels.ParsedGMBLocalPost{
				AccountID:       wo.AccountID,
				LocationID:      wo.LocationID,
				AccountName:     accountName,
				LocationName:    locationName,
				LanguageCode:    lp.LanguageCode,
				PostName:        lp.Name,
				Summary:         lp.Summary,
				State:           lp.State,
				TopicType:       lp.TopicType,
				SearchURL:       lp.SearchURL,
				CreateTime:      lp.CreateTime,
				UpdateTime:      lp.UpdateTime,
				MediaNames:      mediaNames,
				MediaFormats:    mediaFormats,
				MediaGoogleURLs: mediaGoogleURLs,
			}
			chPost := conversions.ConvertGMBLocalPost(parsed)
			if chPost != nil {
				allPosts = append(allPosts, chPost)
			}
		}

		if (reachedOlderThanStart && !pageHasInRange) || resp.NextPageToken == "" || resp.NextPageToken == pageToken {
			break
		}
		pageToken = resp.NextPageToken
	}

	if len(allPosts) == 0 {
		return 0
	}

	if err := sink.BulkInsertGMBLocalPosts(ctx, allPosts); err != nil {
		log.Warn().Err(err).Int("count", len(allPosts)).Msg("Failed to insert GMB local posts")
		return 0
	}
	log.Info().Int("count", len(allPosts)).Msg("Inserted GMB local posts")
	return len(allPosts)
}

// fetchAndInsertReviews fetches reviews (up to 2 pages) and inserts into ClickHouse.
func fetchAndInsertReviews(
	ctx context.Context,
	gmbClient GMBClientInterface,
	sink ClickHouseSinkInterface,
	wo ImmediateWorkOrder,
	accessToken, accountName, locationName string,
	startDate, endDate time.Time,
	log *logger.Logger,
) int {
	var allReviews []*chmodels.GMBReviews
	pageToken := ""

	for page := 0; ; page++ {
		resp, err := gmbClient.FetchReviews(ctx, wo.AccountID, wo.LocationID, accessToken, pageToken)
		if err != nil {
			log.Warn().Err(err).Str("location_id", wo.LocationID).Int("page", page).Msg("Failed to fetch reviews")
			break
		}

		reachedOlderThanStart := false
		pageHasInRange := false
		for _, rv := range resp.Reviews {
			createdAt, ok := parseGMBAPITime(rv.CreateTime)
			if !ok {
				continue
			}
			if createdAt.Before(startDate) {
				reachedOlderThanStart = true
				continue
			}
			if !withinGMBDateRange(createdAt, startDate, endDate) {
				continue
			}
			pageHasInRange = true

			var replyComment, replyUpdateTime string
			if rv.ReviewReply != nil {
				replyComment = rv.ReviewReply.Comment
				replyUpdateTime = rv.ReviewReply.UpdateTime
			}

			parsed := &kafkamodels.ParsedGMBReview{
				AccountID:               wo.AccountID,
				LocationID:              wo.LocationID,
				AccountName:             accountName,
				LocationName:            locationName,
				ReviewID:                rv.ReviewID,
				ReviewName:              rv.Name,
				ReviewerDisplayName:     rv.Reviewer.DisplayName,
				ReviewerProfilePhotoURL: rv.Reviewer.ProfilePhotoURL,
				StarRating:              rv.StarRating,
				Comment:                 rv.Comment,
				CreateTime:              rv.CreateTime,
				UpdateTime:              rv.UpdateTime,
				ReplyComment:            replyComment,
				ReplyUpdateTime:         replyUpdateTime,
			}
			chReview := conversions.ConvertGMBReview(parsed)
			if chReview != nil {
				allReviews = append(allReviews, chReview)
			}
		}

		if (reachedOlderThanStart && !pageHasInRange) || resp.NextPageToken == "" || resp.NextPageToken == pageToken {
			break
		}
		pageToken = resp.NextPageToken
	}

	if len(allReviews) == 0 {
		return 0
	}

	if err := sink.BulkInsertGMBReviews(ctx, allReviews); err != nil {
		log.Warn().Err(err).Int("count", len(allReviews)).Msg("Failed to insert GMB reviews")
		return 0
	}
	log.Info().Int("count", len(allReviews)).Msg("Inserted GMB reviews")
	return len(allReviews)
}

// fetchAndInsertMediaAssets fetches media assets and inserts into ClickHouse.
func fetchAndInsertMediaAssets(
	ctx context.Context,
	gmbClient GMBClientInterface,
	sink ClickHouseSinkInterface,
	wo ImmediateWorkOrder,
	accessToken, accountName, locationName string,
	startDate, endDate time.Time,
	log *logger.Logger,
) int {
	var allAssets []*chmodels.GMBMediaAssets
	pageToken := ""
	for page := 0; ; page++ {
		resp, err := gmbClient.FetchMediaAssets(ctx, wo.AccountID, wo.LocationID, accessToken, pageToken)
		if err != nil {
			log.Warn().Err(err).Str("location_id", wo.LocationID).Int("page", page).Msg("Failed to fetch media assets")
			break
		}

		reachedOlderThanStart := false
		pageHasInRange := false
		for _, mi := range resp.MediaItems {
			createdAt, ok := parseGMBAPITime(mi.CreateTime)
			if !ok {
				continue
			}
			if createdAt.Before(startDate) {
				reachedOlderThanStart = true
				continue
			}
			if !withinGMBDateRange(createdAt, startDate, endDate) {
				continue
			}
			pageHasInRange = true

			parsed := &kafkamodels.ParsedGMBMediaAsset{
				AccountID:                   wo.AccountID,
				LocationID:                  wo.LocationID,
				AccountName:                 accountName,
				LocationName:                locationName,
				LanguageCode:                wo.LanguageCode,
				MediaName:                   mi.Name,
				SourceURL:                   mi.SourceURL,
				MediaFormat:                 mi.MediaFormat,
				LocationAssociationCategory: mi.LocationAssociation.Category,
				GoogleURL:                   mi.GoogleURL,
				ThumbnailURL:                mi.ThumbnailURL,
				WidthPixels:                 uint64(mi.Dimensions.WidthPixels),
				HeightPixels:                uint64(mi.Dimensions.HeightPixels),
				CreateTime:                  mi.CreateTime,
			}
			chAsset := conversions.ConvertGMBMediaAsset(parsed)
			if chAsset != nil {
				allAssets = append(allAssets, chAsset)
			}
		}

		if (reachedOlderThanStart && !pageHasInRange) || resp.NextPageToken == "" || resp.NextPageToken == pageToken {
			break
		}
		pageToken = resp.NextPageToken
	}

	if len(allAssets) == 0 {
		return 0
	}

	if err := sink.BulkInsertGMBMediaAssets(ctx, allAssets); err != nil {
		log.Warn().Err(err).Int("count", len(allAssets)).Msg("Failed to insert GMB media assets")
		return 0
	}
	log.Info().Int("count", len(allAssets)).Msg("Inserted GMB media assets")
	return len(allAssets)
}

// SendNotifications sends Pusher and email notifications for GMB analytics completion.
func SendNotifications(
	ctx context.Context,
	mongoRepo interface{},
	notifier *notification.Service,
	pusherClient *notification.PusherClient,
	wo ImmediateWorkOrder,
	account *mongomodels.SocialIntegration,
	originalState string,
	log *logger.Logger,
) {
	locationID := wo.LocationID
	if account != nil && account.PlatformIdentifier != "" {
		locationID = account.PlatformIdentifier
	}
	// Pusher forbids '/' in channel/event names. GMB location IDs look like
	// "accounts/123/locations/456" so we replace '/' with '-'.
	safeLocationID := strings.ReplaceAll(locationID, "/", "-")
	channel := fmt.Sprintf("gmb-analytics-channel-%s-%s", wo.WorkspaceID, safeLocationID)
	event := fmt.Sprintf("syncing-%s-%s", wo.WorkspaceID, safeLocationID)

	data := map[string]interface{}{
		"state":                     "Processed",
		"account":                   locationID,
		"last_analytics_updated_at": time.Now().UTC().Format("2006-01-02"),
	}

	if err := pusherClient.Trigger(channel, event, data); err != nil {
		log.Warn().
			Err(err).
			Str("error_message", err.Error()).
			Str("channel", channel).
			Str("event", event).
			Str("function", "SendNotifications").
			Msg("Failed to send Pusher notification")
	} else {
		log.Debug().
			Str("channel", channel).
			Str("event", event).
			Msg("Sent Pusher notification")
	}

	// Send email only for newly added accounts
	if notifier == nil || originalState != mongomodels.StateAdded {
		return
	}

	// Fetch account from MongoDB if not already fetched
	if account == nil {
		if wo.ID == "" {
			log.Warn().Msg("No account ID available for email notification")
			return
		}

		accountObjID, err := primitive.ObjectIDFromHex(wo.ID)
		if err != nil {
			log.Warn().Err(err).Str("account_id", wo.ID).Msg("Invalid account ID for email notification")
			return
		}

		repo, ok := mongoRepo.(interface {
			FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error)
		})
		if !ok {
			log.Warn().Msg("MongoDB repository does not support FindByID")
			return
		}

		var fetchErr error
		account, fetchErr = repo.FindByID(ctx, accountObjID)
		if fetchErr != nil {
			log.Warn().Err(fetchErr).Str("account_id", wo.ID).Msg("Failed to fetch account for email notification")
			return
		}
	}

	if account == nil {
		log.Warn().Str("account_id", wo.ID).Msg("Account not found for email notification")
		return
	}

	accountProfileName := account.PlatformName
	if accountProfileName == "" {
		accountProfileName = account.PlatformIdentifier
	}

	workspaceID := account.GetWorkspaceIDHex()
	if workspaceID == "" {
		workspaceID = wo.WorkspaceID
	}

	if err := notifier.SendAnalyticsNotification(account.GetUserIDHex(), workspaceID, "gmb", account.ID.Hex(), accountProfileName, false); err != nil {
		log.Warn().
			Err(err).
			Str("account_id", wo.ID).
			Str("user_id", account.GetUserIDHex()).
			Msg("Failed to send email notification")
	} else {
		log.Debug().
			Str("account_id", wo.ID).
			Str("user_id", account.GetUserIDHex()).
			Msg("Sent email notification")
	}
}
