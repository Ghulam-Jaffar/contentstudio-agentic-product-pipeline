package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/notification"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

const (
	StatsConcsPerWorker = 4
	MediaConcPerWorker  = 4
	GeoConcPerWorker    = 2
	DefaultWorkerCount  = 10
)

type WorkOrder struct {
	ID          string `json:"id"`
	AccessToken string `json:"access_token"`
	WorkspaceID string `json:"workspace_id"`
	SyncType    string `json:"sync_type"`
	AccountID   string `json:"account_id"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
}

type enrichedPost struct {
	Post        map[string]any
	ActivityID  string
	ImageIDs    []string
	VideoIDs    []string
	DocumentIDs []string
}

type FetchedData struct {
	Posts        map[string]*enrichedPost
	StatsMap     map[string]map[string]any
	ImageMap     map[string]map[string]any
	VideoMap     map[string]map[string]any
	DocumentMap  map[string]map[string]any
	FollowerData []byte
	PageStats    []byte
	ShareStats   []byte
}

type ParsedData struct {
	Posts       []kafkamodels.ParsedLinkedinPost
	MediaAssets []kafkamodels.ParsedLinkedinMediaAsset
	Stats       []kafkamodels.ParsedLinkedinStat
	Insights    []kafkamodels.ParsedLinkedinInsights
}

type Processor struct {
	MongoRepo    mongodb.UnifiedSocialRepository
	LiClient     *social.LinkedInClient
	Sink         *conversions.ClickHouseSink
	GeoResolver  *social.GeoResolver
	Producer     kafka2.Producer
	Notifier     *notification.Service
	PusherClient *notification.PusherClient
	Logger       *logger.Logger
	Cfg          *config.Config
	StatsConc    *semaphore.Weighted
	MediaConc    *semaphore.Weighted
	GeoConc      *semaphore.Weighted
}

func New(
	mongoRepo mongodb.UnifiedSocialRepository,
	sink *conversions.ClickHouseSink,
	producer kafka2.Producer,
	notifier *notification.Service,
	pusherClient *notification.PusherClient,
	log *logger.Logger,
	cfg *config.Config,
) *Processor {
	liClient := social.NewLinkedInClient()
	return &Processor{
		MongoRepo:    mongoRepo,
		LiClient:     liClient,
		Sink:         sink,
		GeoResolver:  social.NewGeoResolver(liClient, sink.RawClient),
		Producer:     producer,
		Notifier:     notifier,
		PusherClient: pusherClient,
		Logger:       log,
		Cfg:          cfg,
		StatsConc:    semaphore.NewWeighted(StatsConcsPerWorker * DefaultWorkerCount),
		MediaConc:    semaphore.NewWeighted(MediaConcPerWorker * DefaultWorkerCount),
		GeoConc:      semaphore.NewWeighted(GeoConcPerWorker * DefaultWorkerCount),
	}
}

type captureFunc func(stage string, e error, extra map[string]interface{})

func newCaptureFunc(baseTags map[string]string, baseExtras map[string]interface{}) captureFunc {
	return func(stage string, e error, extra map[string]interface{}) {
		if e == nil {
			return
		}
		tags := make(map[string]string, len(baseTags)+1)
		for k, v := range baseTags {
			tags[k] = v
		}
		tags["stage"] = stage

		extras := make(map[string]interface{}, len(baseExtras)+len(extra))
		for k, v := range baseExtras {
			extras[k] = v
		}
		for k, v := range extra {
			extras[k] = v
		}
		logger.CaptureException(e, tags, extras)
	}
}

func isTokenError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "invalid_token") ||
		strings.Contains(errStr, "expired") ||
		strings.Contains(errStr, "access_token") ||
		strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "permission") ||
		strings.Contains(errStr, "not authorized")
}

// isExpectedError checks if an error is expected (permissions/auth) and should not be sent to Sentry
func isExpectedError(err error) bool {
	if err == nil {
		return false
	}
	return social.IsExpectedCompetitorErrorLI(err)
}

func (p *Processor) ProcessAccount(ctx context.Context, wo WorkOrder) (err error) {
	baseTags := map[string]string{
		"platform":     "linkedin",
		"component":    "immediate-processor",
		"linkedin_id":  wo.AccountID,
		"workspace_id": wo.WorkspaceID,
		"sync_type":    wo.SyncType,
	}
	baseExtras := map[string]interface{}{
		"linkedin_id":  wo.AccountID,
		"workspace_id": wo.WorkspaceID,
		"sync_type":    wo.SyncType,
	}
	capture := newCaptureFunc(baseTags, baseExtras)

	if wo.AccountID == "" || wo.AccessToken == "" {
		p.Logger.Warn().
			Str("linkedin_id", wo.AccountID).
			Bool("has_token", wo.AccessToken != "").
			Msg("Skipping work order with missing linkedin_id or access_token")
		return nil
	}

	if decrypted, decErr := crypto.DecryptToken(wo.AccessToken, p.Cfg.DecryptionKey); decErr == nil {
		wo.AccessToken = decrypted
	}

	accountID, err := primitive.ObjectIDFromHex(wo.ID)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: invalid account ID: %w", err)
	}
	defer func() {
		if err == nil {
			return
		}
		if recordErr := p.MongoRepo.RecordProcessingError(ctx, accountID, err.Error()); recordErr != nil {
			p.Logger.Warn().Err(recordErr).Str("account_id", wo.ID).Msg("Failed to record processing error")
		}
	}()

	account, err := p.MongoRepo.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: failed to fetch account from MongoDB: %w", err)
	}
	if account == nil {
		return fmt.Errorf("Processor.ProcessAccount: account not found: %s", wo.ID)
	}
	if mongodb.HasProcessingErrorMeta(account.MetaData) {
		if err := p.MongoRepo.ClearProcessingError(ctx, accountID); err != nil {
			p.Logger.Warn().Err(err).Str("account_id", wo.ID).Msg("Failed to clear stale processing error before retry")
		}
	}

	originalState := account.State

	p.Logger.Info().Str("linkedin_id", wo.AccountID).Msg("Starting data fetch")
	fetchedData, err := p.fetchAllData(ctx, wo, capture)
	if err != nil {
		return err
	}

	p.Logger.Info().Str("linkedin_id", wo.AccountID).Msg("Starting data parsing")
	parsedData, err := p.parseAllData(wo, fetchedData, capture)
	if err != nil {
		return err
	}

	p.Logger.Info().
		Str("linkedin_id", wo.AccountID).
		Int("posts", len(parsedData.Posts)).
		Int("insights", len(parsedData.Insights)).
		Msg("Storing data in ClickHouse")

	if err := p.storeInClickHouse(ctx, wo, parsedData, capture); err != nil {
		return err
	}

	// Persist successful completion state and analytics timestamp in MongoDB.
	now := time.Now().UTC()
	if err := p.MongoRepo.UpdateState(ctx, accountID, mongomodels.StateProcessed); err != nil {
		p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("account_id", wo.ID).Str("function", "ProcessAccount").Msg("Failed to update account state to Processed")
	}
	if err := p.MongoRepo.ClearProcessingError(ctx, accountID); err != nil {
		p.Logger.Warn().Err(err).Str("account_id", wo.ID).Msg("Failed to clear processing error")
	}
	if err := p.MongoRepo.UpdateAnalyticsTimestamp(ctx, accountID, "analytics", now); err != nil {
		p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("account_id", wo.ID).Str("function", "ProcessAccount").Msg("Failed to update last_analytics_updated_at")
	}

	userID := account.GetUserIDHex()
	workspaceID := account.GetWorkspaceIDHex()
	if workspaceID == "" {
		workspaceID = wo.WorkspaceID
	}

	p.sendPusherNotification(account, workspaceID, originalState)

	if originalState == "Added" {
		p.sendEmailNotification(userID, workspaceID, wo.AccountID, account.PlatformName)
	}

	return nil
}

func (p *Processor) fetchAllData(ctx context.Context, wo WorkOrder, capture captureFunc) (*FetchedData, error) {
	startDate, endDate, err := parseLinkedInDateRange(wo.StartDate, wo.EndDate)
	if err != nil {
		return nil, err
	}
	cutoffTime := startDate

	p.Logger.Info().
		Str("linkedin_id", wo.AccountID).
		Time("start_date", startDate).
		Time("end_date", endDate).
		Time("cutoff_time", cutoffTime).
		Msg("Fetching LinkedIn posts")
	posts, err := p.LiClient.FetchPostsPaginated(ctx, wo.AccountID, "organization", wo.AccessToken, cutoffTime)
	if err != nil {
		p.Logger.Warn().
			Err(err).
			Str("error_message", err.Error()).
			Str("linkedin_id", wo.AccountID).
			Str("function", "fetchAllData").
			Str("stage", "fetch_posts").
			Msg("Failed to fetch posts")
		return nil, err
	}
	p.Logger.Info().Str("linkedin_id", wo.AccountID).Int("posts_fetched", len(posts)).Msg("Fetched LinkedIn posts")
	posts = filterLinkedInPostsByDate(posts, startDate, endDate)
	p.Logger.Info().Str("linkedin_id", wo.AccountID).Int("posts_in_range", len(posts)).Msg("Filtered LinkedIn posts by requested date range")

	byActivity, ugcIDs, shareIDs, imageIDs, videoIDs, documentIDs := p.extractAssetIDs(posts, wo.AccountID)

	result := &FetchedData{
		Posts:       byActivity,
		StatsMap:    make(map[string]map[string]any),
		ImageMap:    make(map[string]map[string]any),
		VideoMap:    make(map[string]map[string]any),
		DocumentMap: make(map[string]map[string]any),
	}

	var (
		statsMu sync.Mutex
		imgMu   sync.Mutex
		vidMu   sync.Mutex
		docMu   sync.Mutex
	)

	eg, egctx := errgroup.WithContext(ctx)

	startDateInsights := startDate
	endDateInsights := endDate
	startMsInsights := startDateInsights.UnixMilli()
	endMsInsights := endDateInsights.UnixMilli()

	// Fetch follower data
	eg.Go(func() error {
		followerStats, err := p.LiClient.FetchFollowerStatsWithGeoIDs(egctx, wo.AccountID, wo.AccessToken)
		if err != nil {
			if isExpectedError(err) || isTokenError(err) {
				return err
			}
			capture("fetch_follower_data", err, nil)
			return nil
		}

		var geoNames map[string]string
		if len(followerStats.GeoIDs) > 0 {
			if err := p.GeoConc.Acquire(egctx, 1); err != nil {
				return nil
			}
			geoNames, _ = p.GeoResolver.ResolveGeoIDsWithType(egctx, followerStats.GeoIDs, wo.AccessToken)
			p.GeoConc.Release(1)
		}

		data, err := p.LiClient.BuildFollowerDataWithGeoNames(followerStats, geoNames)
		if err == nil {
			result.FollowerData = data
		} else {
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("linkedin_id", wo.AccountID).Str("function", "fetchAllData").Str("stage", "build_follower_data").Msg("Failed to build follower data with geo names (continuing)")
			capture("build_follower_data", err, nil)
		}
		return nil
	})

	// Fetch page statistics
	eg.Go(func() error {
		data, err := p.LiClient.FetchPageStatisticsRaw(egctx, wo.AccountID, wo.AccessToken, startMsInsights, endMsInsights)
		if err != nil {
			if isExpectedError(err) || isTokenError(err) {
				return err
			}
			capture("fetch_page_statistics", err, nil)
		} else {
			result.PageStats = data
		}
		return nil
	})

	// Fetch share statistics
	eg.Go(func() error {
		data, err := p.LiClient.FetchShareStatisticsRaw(egctx, wo.AccountID, wo.AccessToken, startMsInsights, endMsInsights)
		if err != nil {
			if isExpectedError(err) || isTokenError(err) {
				return err
			}
			capture("fetch_share_statistics", err, nil)
		} else {
			result.ShareStats = data
		}
		return nil
	})

	// Fetch UGC stats in chunks
	for _, ids := range chunk(ugcIDs, 100) {
		ids := ids
		eg.Go(func() error {
			if err := p.StatsConc.Acquire(egctx, 1); err != nil {
				return nil
			}
			defer p.StatsConc.Release(1)

			body, err := p.LiClient.FetchStatsRaw(egctx, wo.AccountID, ids, nil, wo.AccessToken)
			if err != nil {
				if isExpectedError(err) || isTokenError(err) {
					return err
				}
				p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("linkedin_id", wo.AccountID).Str("function", "fetchAllData").Str("stage", "fetch_ugc_stats").Int("chunk_size", len(ids)).Msg("Failed to fetch UGC stats chunk (continuing)")
				capture("fetch_ugc_stats", err, map[string]interface{}{"method": "FetchStatsRaw", "chunk_size": len(ids)})
				return nil
			}
			if body != nil {
				m := parseStatsBatch(body)
				statsMu.Lock()
				for k, v := range m {
					result.StatsMap[k] = v
				}
				statsMu.Unlock()
			}
			return nil
		})
	}

	// Fetch share stats in chunks
	for _, ids := range chunk(shareIDs, 100) {
		ids := ids
		eg.Go(func() error {
			if err := p.StatsConc.Acquire(egctx, 1); err != nil {
				return nil
			}
			defer p.StatsConc.Release(1)

			body, err := p.LiClient.FetchStatsRaw(egctx, wo.AccountID, nil, ids, wo.AccessToken)
			if err != nil {
				if isExpectedError(err) || isTokenError(err) {
					return err
				}
				p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("linkedin_id", wo.AccountID).Str("function", "fetchAllData").Str("stage", "fetch_share_stats").Int("chunk_size", len(ids)).Msg("Failed to fetch share stats chunk (continuing)")
				capture("fetch_share_stats", err, map[string]interface{}{"method": "FetchStatsRaw", "chunk_size": len(ids)})
				return nil
			}
			if body != nil {
				m := parseStatsBatch(body)
				statsMu.Lock()
				for k, v := range m {
					result.StatsMap[k] = v
				}
				statsMu.Unlock()
			}
			return nil
		})
	}

	// Fetch images in chunks
	for _, ids := range chunk(imageIDs, 80) {
		ids := ids
		eg.Go(func() error {
			if err := p.MediaConc.Acquire(egctx, 1); err != nil {
				return nil
			}
			defer p.MediaConc.Release(1)

			body, err := p.LiClient.FetchImagesRaw(egctx, ids, wo.AccessToken)
			if err != nil {
				if isExpectedError(err) || isTokenError(err) {
					return err
				}
				p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("linkedin_id", wo.AccountID).Str("function", "fetchAllData").Str("stage", "fetch_images").Int("chunk_size", len(ids)).Msg("Failed to fetch images chunk (continuing)")
				capture("fetch_images", err, map[string]interface{}{"method": "FetchImagesRaw", "chunk_size": len(ids)})
				return nil
			}
			if body != nil {
				m := parseAssetBatch(body)
				imgMu.Lock()
				for k, v := range m {
					result.ImageMap[k] = v
				}
				imgMu.Unlock()
			}
			return nil
		})
	}

	// Fetch videos in chunks
	for _, ids := range chunk(videoIDs, 80) {
		ids := ids
		eg.Go(func() error {
			if err := p.MediaConc.Acquire(egctx, 1); err != nil {
				return nil
			}
			defer p.MediaConc.Release(1)

			body, err := p.LiClient.FetchVideosRaw(egctx, ids, wo.AccessToken)
			if err != nil {
				if isExpectedError(err) || isTokenError(err) {
					return err
				}
				p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("linkedin_id", wo.AccountID).Str("function", "fetchAllData").Str("stage", "fetch_videos").Int("chunk_size", len(ids)).Msg("Failed to fetch videos chunk (continuing)")
				capture("fetch_videos", err, map[string]interface{}{"method": "FetchVideosRaw", "chunk_size": len(ids)})
				return nil
			}
			if body != nil {
				m := parseAssetBatch(body)
				vidMu.Lock()
				for k, v := range m {
					result.VideoMap[k] = v
				}
				vidMu.Unlock()
			}
			return nil
		})
	}

	// Fetch documents in chunks
	for _, ids := range chunk(documentIDs, 80) {
		ids := ids
		eg.Go(func() error {
			if err := p.MediaConc.Acquire(egctx, 1); err != nil {
				return nil
			}
			defer p.MediaConc.Release(1)

			body, err := p.LiClient.FetchDocumentsRaw(egctx, ids, wo.AccessToken)
			if err != nil {
				if isExpectedError(err) || isTokenError(err) {
					return err
				}
				p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("linkedin_id", wo.AccountID).Str("function", "fetchAllData").Str("stage", "fetch_documents").Int("chunk_size", len(ids)).Msg("Failed to fetch documents chunk (continuing)")
				capture("fetch_documents", err, map[string]interface{}{"method": "FetchDocumentsRaw", "chunk_size": len(ids)})
				return nil
			}
			if body != nil {
				m := parseAssetBatch(body)
				docMu.Lock()
				for k, v := range m {
					result.DocumentMap[k] = v
				}
				docMu.Unlock()
			}
			return nil
		})
	}

	_ = eg.Wait()

	p.Logger.Info().
		Str("linkedin_id", wo.AccountID).
		Int("posts_count", len(result.Posts)).
		Int("stats_fetched", len(result.StatsMap)).
		Int("images_fetched", len(result.ImageMap)).
		Int("videos_fetched", len(result.VideoMap)).
		Int("documents_fetched", len(result.DocumentMap)).
		Msg("Completed fetching all data")

	return result, nil
}

func parseLinkedInDateRange(startDateStr, endDateStr string) (time.Time, time.Time, error) {
	if strings.TrimSpace(startDateStr) == "" && strings.TrimSpace(endDateStr) == "" {
		startDate, endDate := defaultLinkedInDateRange(time.Now().UTC())
		return startDate, endDate, nil
	}

	if strings.TrimSpace(startDateStr) == "" || strings.TrimSpace(endDateStr) == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("LinkedIn immediate work order requires both start_date and end_date")
	}

	startDate, err := time.Parse("2006-01-02", strings.TrimSpace(startDateStr))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("LinkedIn immediate work order: invalid start_date %q: %w", startDateStr, err)
	}
	endDate, err := time.Parse("2006-01-02", strings.TrimSpace(endDateStr))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("LinkedIn immediate work order: invalid end_date %q: %w", endDateStr, err)
	}
	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("LinkedIn immediate work order: end_date cannot be before start_date")
	}

	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	return startDate, endDate, nil
}

func defaultLinkedInDateRange(now time.Time) (time.Time, time.Time) {
	lastAvailableDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -2)
	startDate := time.Date(lastAvailableDate.Year(), lastAvailableDate.Month(), lastAvailableDate.Day(), 0, 0, 0, 0, time.UTC).AddDate(-1, 0, 0)
	endDate := time.Date(lastAvailableDate.Year(), lastAvailableDate.Month(), lastAvailableDate.Day(), 23, 59, 59, 0, time.UTC).AddDate(0, 0, 1)
	return startDate, endDate
}

func filterLinkedInPostsByDate(posts []json.RawMessage, startDate, endDate time.Time) []json.RawMessage {
	if len(posts) == 0 {
		return posts
	}

	filtered := make([]json.RawMessage, 0, len(posts))
	for _, raw := range posts {
		var post struct {
			CreatedAt int64 `json:"createdAt"`
		}
		if err := json.Unmarshal(raw, &post); err != nil || post.CreatedAt <= 0 {
			continue
		}

		postTime := time.UnixMilli(post.CreatedAt).UTC()
		if postTime.Before(startDate) {
			break
		}
		if postTime.After(endDate) {
			continue
		}
		filtered = append(filtered, raw)
	}

	return filtered
}

func (p *Processor) extractAssetIDs(posts []json.RawMessage, linkedinID string) (
	byActivity map[string]*enrichedPost,
	ugcIDs, shareIDs, imageIDs, videoIDs, documentIDs []string,
) {
	byActivity = make(map[string]*enrichedPost)
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

		if strings.Contains(id, "ugcPost") {
			ugcSet[id] = struct{}{}
		}
		if strings.Contains(id, "share") {
			shareSet[id] = struct{}{}
		}

		if content, _ := post["content"].(map[string]any); content != nil {
			for k, v := range content {
				switch k {
				case "multiImage":
					if m, ok := v.(map[string]any); ok {
						if imgs, ok := m["images"].([]any); ok {
							for _, it := range imgs {
								if mm, ok := it.(map[string]any); ok {
									if iid, ok := mm["id"].(string); ok && iid != "" {
										imgSet[iid] = struct{}{}
										ep.ImageIDs = append(ep.ImageIDs, iid)
									}
								}
							}
						}
					}
				case "article":
					if m, ok := v.(map[string]any); ok {
						if thumb, ok := m["thumbnail"].(string); ok && thumb != "" {
							imgSet[thumb] = struct{}{}
							ep.ImageIDs = append(ep.ImageIDs, thumb)
						}
					}
				case "media":
					if m, ok := v.(map[string]any); ok {
						if idStr, ok := m["id"].(string); ok && idStr != "" {
							if strings.Contains(idStr, "video") {
								vidSet[idStr] = struct{}{}
								ep.VideoIDs = append(ep.VideoIDs, idStr)
							} else if strings.Contains(idStr, "document") {
								docSet[idStr] = struct{}{}
								ep.DocumentIDs = append(ep.DocumentIDs, idStr)
							} else {
								imgSet[idStr] = struct{}{}
								ep.ImageIDs = append(ep.ImageIDs, idStr)
							}
						}
					}
				}
			}
		}
	}

	toSlice := func(s map[string]struct{}) []string {
		out := make([]string, 0, len(s))
		for k := range s {
			out = append(out, k)
		}
		return out
	}
	ugcIDs = toSlice(ugcSet)
	shareIDs = toSlice(shareSet)
	imageIDs = toSlice(imgSet)
	videoIDs = toSlice(vidSet)
	documentIDs = toSlice(docSet)

	return
}

func (p *Processor) parseAllData(wo WorkOrder, data *FetchedData, capture captureFunc) (*ParsedData, error) {
	parsedPosts := make([]kafkamodels.ParsedLinkedinPost, 0, len(data.Posts))

	for _, ep := range data.Posts {
		meta := map[string]any{}

		if st, ok := data.StatsMap[ep.ActivityID]; ok {
			meta["stats"] = st
		}

		type anySlice []any
		if len(ep.ImageIDs) > 0 {
			imgs := make(anySlice, 0, len(ep.ImageIDs))
			for _, iid := range ep.ImageIDs {
				if im, ok := data.ImageMap[iid]; ok {
					imgs = append(imgs, im)
				}
			}
			if len(imgs) > 0 {
				if meta["assets"] == nil {
					meta["assets"] = map[string]any{}
				}
				meta["assets"].(map[string]any)["images"] = imgs
			}
		}

		if len(ep.VideoIDs) > 0 {
			vids := make(anySlice, 0, len(ep.VideoIDs))
			for _, vid := range ep.VideoIDs {
				if vm, ok := data.VideoMap[vid]; ok {
					vids = append(vids, vm)
				}
			}
			if len(vids) > 0 {
				if meta["assets"] == nil {
					meta["assets"] = map[string]any{}
				}
				meta["assets"].(map[string]any)["videos"] = vids
			}
		}

		if len(ep.DocumentIDs) > 0 {
			docs := make(anySlice, 0, len(ep.DocumentIDs))
			for _, did := range ep.DocumentIDs {
				if dm, ok := data.DocumentMap[did]; ok {
					docs = append(docs, dm)
				}
			}
			if len(docs) > 0 {
				if meta["assets"] == nil {
					meta["assets"] = map[string]any{}
				}
				meta["assets"].(map[string]any)["documents"] = docs
			}
		}

		if len(meta) > 0 {
			ep.Post["meta"] = meta
		}

		raw, _ := json.Marshal(ep.Post)
		post, err := parsing.ParsePost(raw)
		if err != nil || post == nil {
			continue
		}
		post.LinkedinID = wo.AccountID
		parsedPosts = append(parsedPosts, *post)
	}

	var parsedInsights []kafkamodels.ParsedLinkedinInsights
	if data.FollowerData != nil || data.PageStats != nil || data.ShareStats != nil {
		merged := map[string]json.RawMessage{}
		if data.FollowerData != nil {
			merged["followerData"] = data.FollowerData
		}
		if data.PageStats != nil {
			merged["pageStatistics"] = data.PageStats
		}
		if data.ShareStats != nil {
			merged["shareStatistics"] = data.ShareStats
		}

		mergedJSON, _ := json.Marshal(merged)

		dailyInsights, err := parsing.ParseInsightsDaily(mergedJSON)
		if err != nil {
			capture("parse_insights", err, nil)
		} else if len(dailyInsights) > 0 {
			for _, ins := range dailyInsights {
				ins.LinkedinID = wo.AccountID
				ins.RecordID = wo.AccountID + "_" + ins.CreatedAt.Format("2006-01-02")
				parsedInsights = append(parsedInsights, *ins)
			}
		}
	}

	return &ParsedData{
		Posts:    parsedPosts,
		Insights: parsedInsights,
	}, nil
}

func (p *Processor) storeInClickHouse(ctx context.Context, wo WorkOrder, data *ParsedData, capture captureFunc) error {
	if len(data.Posts) > 0 {
		var chPosts []*clickhousemodels.LinkedInPosts
		for _, post := range data.Posts {
			cp := conversions.ConvertLinkedInPost(&post)
			if cp != nil {
				chPosts = append(chPosts, cp)
			}
		}
		if err := p.Sink.BulkInsertLinkedInPosts(ctx, chPosts); err != nil {
			return fmt.Errorf("Processor.storeInClickHouse: failed to insert posts: %w", err)
		}
		p.Logger.Info().Str("linkedin_id", wo.AccountID).Int("count", len(chPosts)).Msg("Inserted posts into ClickHouse")
	}

	if len(data.MediaAssets) > 0 {
		var chAssets []*clickhousemodels.LinkedInMediaAsset
		for _, m := range data.MediaAssets {
			ca := conversions.ConvertLinkedInMediaAsset(&m)
			if ca != nil {
				chAssets = append(chAssets, ca)
			}
		}
		if err := p.Sink.BulkInsertLinkedInMediaAssets(ctx, chAssets); err != nil {
			return fmt.Errorf("Processor.storeInClickHouse: failed to insert media assets: %w", err)
		}
	}

	if len(data.Stats) > 0 {
		var chStats []*clickhousemodels.LinkedInStat
		for _, s := range data.Stats {
			cs := conversions.ConvertLinkedInStat(&s)
			if cs != nil {
				chStats = append(chStats, cs)
			}
		}
		if err := p.Sink.BulkInsertLinkedInStats(ctx, chStats); err != nil {
			return fmt.Errorf("Processor.storeInClickHouse: failed to insert stats: %w", err)
		}
	}

	if len(data.Insights) > 0 {
		var chInsights []*clickhousemodels.LinkedInInsights
		for _, ins := range data.Insights {
			ci := conversions.ConvertLinkedInInsights(&ins)
			if ci != nil {
				chInsights = append(chInsights, ci)
			}
		}
		if err := p.Sink.BulkInsertLinkedInInsights(ctx, chInsights); err != nil {
			return fmt.Errorf("Processor.storeInClickHouse: failed to insert insights: %w", err)
		}
		p.Logger.Info().Str("linkedin_id", wo.AccountID).Int("count", len(chInsights)).Msg("Inserted insights into ClickHouse")
	}

	return nil
}

func (p *Processor) sendPusherNotification(account *mongomodels.SocialIntegration, workspaceID, originalState string) {
	if p.PusherClient == nil {
		return
	}

	accountID := account.PlatformIdentifier
	if accountID == "" {
		accountID = account.LinkedinProfileID
	}
	if accountID == "" {
		accountID = account.GetPlatformID()
	}
	if accountID == "" {
		return
	}

	// Frontend uses li-analytics-channel-{workspace_id}-{linkedin_id}.
	channel := fmt.Sprintf("li-analytics-channel-%s-%s", workspaceID, accountID)
	event := fmt.Sprintf("syncing-%s-%s", workspaceID, accountID)

	data := map[string]interface{}{
		"state":                     "Processed",
		"account":                   accountID,
		"last_analytics_updated_at": time.Now().UTC().Format("2006-01-02"),
	}

	if err := p.PusherClient.Trigger(channel, event, data); err != nil {
		p.Logger.Warn().
			Err(err).
			Str("error_message", err.Error()).
			Str("channel", channel).
			Str("event", event).
			Str("function", "sendPusherNotification").
			Msg("Failed to send Pusher notification")
	}
}

func (p *Processor) sendEmailNotification(userID, workspaceID, accountID, accountName string) {
	if p.Notifier == nil {
		return
	}
	if err := p.Notifier.SendAnalyticsNotification(userID, workspaceID, "linkedin", accountID, accountName, false); err != nil {
		p.Logger.Warn().
			Err(err).
			Str("error_message", err.Error()).
			Str("user_id", userID).
			Str("workspace_id", workspaceID).
			Str("account_id", accountID).
			Str("function", "sendEmailNotification").
			Msg("Failed to send analytics notification to backend")
	}
}

func chunk[T any](in []T, n int) [][]T {
	if n <= 0 || len(in) == 0 {
		return nil
	}
	var out [][]T
	for len(in) > n {
		out = append(out, in[:n])
		in = in[n:]
	}
	return append(out, in)
}

func parseStatsBatch(body []byte) map[string]map[string]any {
	type elem struct {
		UGCPost string         `json:"ugcPost"`
		Share   string         `json:"share"`
		Total   map[string]any `json:"totalShareStatistics"`
	}
	var resp struct {
		Elements []elem `json:"elements"`
	}
	_ = json.Unmarshal(body, &resp)

	out := make(map[string]map[string]any, len(resp.Elements))
	for _, e := range resp.Elements {
		id := e.UGCPost
		if id == "" {
			id = e.Share
		}
		if id != "" {
			out[id] = e.Total
		}
	}
	return out
}

func parseAssetBatch(body []byte) map[string]map[string]any {
	var resp struct {
		Results map[string]map[string]any `json:"results"`
	}
	_ = json.Unmarshal(body, &resp)

	out := make(map[string]map[string]any, len(resp.Results))
	for _, m := range resp.Results {
		if id, ok := m["id"].(string); ok && id != "" {
			out[id] = m
			continue
		}
		if id, ok := m["asset"].(string); ok && id != "" {
			out[id] = m
		}
	}
	return out
}
