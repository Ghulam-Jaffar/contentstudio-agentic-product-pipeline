package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

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
	MediaInsightsConc   = 20
	ProgressLogInterval = 100
)

type WorkOrder struct {
	ID                    string `json:"id"`
	AccountID             string `json:"account_id"`
	Type                  string `json:"type"`
	AccessToken           string `json:"access_token"`
	WorkspaceID           string `json:"workspace_id"`
	SyncType              string `json:"sync_type"`
	ConnectedViaInstagram bool   `json:"connected_via_instagram"`
	StartDate             string `json:"start_date,omitempty"`
	EndDate               string `json:"end_date,omitempty"`
}

// isExpectedInstagramError returns true for expected auth/permission/viewer errors that should not go to Sentry.
func isExpectedInstagramError(err error) bool {
	if err == nil {
		return false
	}
	if social.IsAuthError(err) || social.IsExpectedCompetitorError(err) {
		return true
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not enough viewers") || strings.Contains(errStr, "(#10)")
}

type ParsedData struct {
	Posts    []kafkamodels.ParsedInstagramPost
	Insights []kafkamodels.ParsedInstagramInsight
}

type EnrichedMedia struct {
	Media    kafkamodels.RawInstagramMedia
	Insights *kafkamodels.RawInstagramMediaInsights
	UserInfo map[string]interface{}
}

type Processor struct {
	MongoRepo    mongodb.UnifiedSocialRepository
	Parser       *parsing.InstagramParser
	Sink         *conversions.ClickHouseSink
	Producer     kafka2.Producer
	Notifier     *notification.Service
	PusherClient *notification.PusherClient
	Logger       *logger.Logger
	Cfg          *config.Config
	InFlight     sync.Map
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
	return &Processor{
		MongoRepo:    mongoRepo,
		Parser:       parsing.NewInstagramParser(),
		Sink:         sink,
		Producer:     producer,
		Notifier:     notifier,
		PusherClient: pusherClient,
		Logger:       log,
		Cfg:          cfg,
	}
}

func (p *Processor) ProcessAccount(ctx context.Context, wo WorkOrder) (err error) {
	oid, err := primitive.ObjectIDFromHex(wo.ID)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: invalid id: %w", err)
	}
	defer func() {
		if err == nil {
			return
		}
		if recordErr := p.MongoRepo.RecordProcessingError(ctx, oid, err.Error()); recordErr != nil {
			p.Logger.Warn().Err(recordErr).Str("account_id", wo.ID).Msg("Failed to record processing error")
		}
	}()
	account, err := p.MongoRepo.FindByID(ctx, oid)
	if err != nil || account == nil {
		return fmt.Errorf("Processor.ProcessAccount: account fetch error: %w", err)
	}
	if mongodb.HasProcessingErrorMeta(account.MetaData) {
		if err := p.MongoRepo.ClearProcessingError(ctx, oid); err != nil {
			p.Logger.Warn().Err(err).Str("account_id", wo.ID).Msg("Failed to clear stale processing error before retry")
		}
	}

	originalState := account.State

	token := wo.AccessToken
	if token == "" {
		return fmt.Errorf("Processor.ProcessAccount: missing access token in work order")
	}
	if decrypted, derr := crypto.DecryptToken(token, p.Cfg.DecryptionKey); derr == nil {
		token = decrypted
	} else {
		p.Logger.Warn().Err(derr).Str("instagram_id", account.InstagramID).Msg("Token appears to be plain, using as-is")
	}

	since, until, err := resolveInstagramDateRange(wo.StartDate, wo.EndDate)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: invalid date range: %w", err)
	}

	igClient := social.NewInstagramClient(p.Cfg.Facebook.AppSecret)
	apiDomain := "facebook"
	if wo.ConnectedViaInstagram {
		igClient = igClient.WithBaseURL("https://graph.instagram.com/")
		apiDomain = "instagram"
	}
	p.Logger.Debug().
		Str("instagram_id", account.InstagramID).
		Str("api_domain", apiDomain).
		Msg("Created Instagram client")

	enrichedMedia, insightsRaw, userInfo, err := p.fetchAllData(ctx, igClient, account.InstagramID, token, since, until)
	if err != nil {
		return err
	}

	if userInfo == nil || len(userInfo) == 0 {
		p.Logger.Warn().Str("instagram_id", account.InstagramID).Msg("No user info available, skipping")
		return nil
	}

	parsed, err := p.parseAllData(account.InstagramID, enrichedMedia, insightsRaw, userInfo)
	if err != nil {
		return err
	}

	if err := p.store(ctx, parsed); err != nil {
		return err
	}

	// Persist successful completion state and analytics timestamp in MongoDB.
	if wo.ID != "" {
		if accountID, parseErr := primitive.ObjectIDFromHex(wo.ID); parseErr != nil {
			p.Logger.Warn().Err(parseErr).Str("account_id", wo.ID).Msg("Invalid account ID, skipping MongoDB state update")
		} else {
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
		}
	}

	userID := account.GetUserIDHex()
	workspaceID := account.GetWorkspaceIDHex()
	if workspaceID == "" {
		workspaceID = wo.WorkspaceID
	}

	p.sendPusherNotification(account, wo.WorkspaceID, originalState)

	if originalState == "Added" {
		p.sendEmailNotification(userID, workspaceID, account.InstagramID, account.PlatformName)
	}

	return nil
}

func (p *Processor) fetchAllData(
	ctx context.Context,
	igClient *social.InstagramClient,
	igID, token string,
	since, until time.Time,
) ([]EnrichedMedia, []social.DailyInsight, map[string]interface{}, error) {
	var posts []kafkamodels.RawInstagramMedia
	var stories []kafkamodels.RawInstagramMedia
	var dailyInsights []social.DailyInsight
	var userInfo map[string]interface{}
	var fetchErr error
	var authErr error

	done := make(chan struct{}, 4)

	go func() {
		data, err := igClient.FetchMediaSince(ctx, igID, token, since)
		if err == nil {
			posts = filterInstagramMediaWithinRange(data, since, until)
		} else {
			if social.IsAuthError(err) {
				authErr = err
			}
			fetchErr = err
		}
		done <- struct{}{}
	}()

	go func() {
		data, err := igClient.FetchStories(ctx, igID, token)
		if err == nil {
			stories = data
		} else {
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("instagram_id", igID).Str("function", "fetchAllData").Str("stage", "fetch_stories").Msg("FetchStories failed (continuing)")
			if social.IsAuthError(err) {
				authErr = err
			} else {
				logger.CaptureException(err, map[string]string{"platform": "instagram", "component": "immediate-processor", "stage": "fetch_stories", "instagram_id": igID}, nil)
			}
		}
		done <- struct{}{}
	}()

	go func() {
		data, err := igClient.FetchInsightsDailyBetween(ctx, igID, token, since, until, 10)
		if err == nil {
			dailyInsights = data
		} else {
			if social.IsAuthError(err) {
				authErr = err
			}
			fetchErr = err
		}
		done <- struct{}{}
	}()

	go func() {
		data, err := igClient.FetchUserInfo(ctx, igID, token)
		if err == nil {
			userInfo = data
		} else {
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("instagram_id", igID).Str("function", "fetchAllData").Str("stage", "fetch_user_info").Msg("FetchUserInfo failed (continuing)")
			if social.IsAuthError(err) {
				authErr = err
			} else {
				logger.CaptureException(err, map[string]string{"platform": "instagram", "component": "immediate-processor", "stage": "fetch_user_info", "instagram_id": igID}, nil)
			}
		}
		done <- struct{}{}
	}()

	<-done
	<-done
	<-done
	<-done

	if authErr != nil {
		p.Logger.Warn().Err(authErr).Str("instagram_id", igID).Msg("Auth error detected, skipping media insights fetch")
		return nil, nil, nil, authErr
	}

	if len(stories) > 0 {
		p.Logger.Info().Str("instagram_id", igID).Int("stories_count", len(stories)).Msg("Fetched stories")
		posts = append(posts, stories...)
	}

	totalMedia := len(posts)
	enrichedMedia := make([]EnrichedMedia, totalMedia)
	var wg sync.WaitGroup
	var authErrorOccurred int32
	var completedCount int32
	sem := make(chan struct{}, MediaInsightsConc)

	p.Logger.Info().
		Str("instagram_id", igID).
		Int("total_media", totalMedia).
		Int("concurrency", MediaInsightsConc).
		Msg("Starting media insights fetch")

	startInsightsFetch := time.Now()

	for i, m := range posts {
		if atomic.LoadInt32(&authErrorOccurred) == 1 {
			break
		}

		wg.Add(1)
		go func(idx int, media kafkamodels.RawInstagramMedia) {
			defer wg.Done()

			if atomic.LoadInt32(&authErrorOccurred) == 1 {
				return
			}

			sem <- struct{}{}
			defer func() { <-sem }()

			enriched := EnrichedMedia{
				Media:    media,
				UserInfo: userInfo,
			}
			insightCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if mediaIns, err := igClient.FetchMediaInsights(insightCtx, media.ID, token, media.MediaType, media.MediaProductType); err == nil && mediaIns != nil {
				enriched.Insights = mediaIns
			} else if err != nil {
				p.Logger.Debug().Err(err).Str("media_id", media.ID).Msg("FetchMediaInsights failed")
				if social.IsAuthError(err) {
					atomic.StoreInt32(&authErrorOccurred, 1)
					p.Logger.Warn().Str("media_id", media.ID).Msg("Auth error on media insights, stopping further API calls")
				}
			}
			enrichedMedia[idx] = enriched

			completed := atomic.AddInt32(&completedCount, 1)
			if completed%ProgressLogInterval == 0 || int(completed) == totalMedia {
				elapsed := time.Since(startInsightsFetch)
				rate := float64(completed) / elapsed.Seconds()
				remaining := totalMedia - int(completed)
				eta := time.Duration(float64(remaining)/rate) * time.Second
				p.Logger.Info().
					Str("instagram_id", igID).
					Int32("completed", completed).
					Int("total", totalMedia).
					Float64("rate_per_sec", rate).
					Dur("elapsed", elapsed).
					Dur("eta", eta).
					Msg("Media insights fetch progress")
			}
		}(i, m)
	}
	wg.Wait()

	p.Logger.Info().
		Str("instagram_id", igID).
		Int("total_media", totalMedia).
		Int32("completed", atomic.LoadInt32(&completedCount)).
		Dur("elapsed", time.Since(startInsightsFetch)).
		Msg("Completed media insights fetch")

	if atomic.LoadInt32(&authErrorOccurred) == 1 {
		return nil, nil, nil, fmt.Errorf("Processor.fetchAllData: auth error occurred during media insights fetch")
	}

	return enrichedMedia, dailyInsights, userInfo, fetchErr
}

func resolveInstagramDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	defaultEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	defaultStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -90)

	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)
	if startDate == "" && endDate == "" {
		return defaultStart, defaultEnd, nil
	}
	if startDate == "" || endDate == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("start_date and end_date are both required")
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date %q: %w", startDate, err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date %q: %w", endDate, err)
	}

	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, time.UTC)
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("end_date must not be before start_date")
	}

	return start, end, nil
}

func filterInstagramMediaWithinRange(media []kafkamodels.RawInstagramMedia, since, until time.Time) []kafkamodels.RawInstagramMedia {
	if len(media) == 0 {
		return media
	}

	filtered := make([]kafkamodels.RawInstagramMedia, 0, len(media))
	for _, item := range media {
		mediaTime, ok := parseInstagramMediaTime(item.Timestamp)
		if !ok {
			continue
		}
		if mediaTime.Before(since) || mediaTime.After(until) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func parseInstagramMediaTime(timestamp string) (time.Time, bool) {
	if strings.TrimSpace(timestamp) == "" {
		return time.Time{}, false
	}

	layouts := []string{
		"2006-01-02T15:04:05-0700",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, timestamp); err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

func (p *Processor) parseAllData(igID string, media []EnrichedMedia, dailyInsights []social.DailyInsight, userInfo map[string]interface{}) (*ParsedData, error) {
	parsed := &ParsedData{}

	username, _ := userInfo["username"].(string)
	name, _ := userInfo["name"].(string)
	profilePic, _ := userInfo["profile_picture_url"].(string)

	for _, em := range media {
		type FetcherEnrichedMedia struct {
			*kafkamodels.RawInstagramMedia
			Insights *kafkamodels.RawInstagramMediaInsights `json:"insights,omitempty"`
			UserInfo map[string]interface{}                 `json:"user_info,omitempty"`
		}
		enrichedStruct := FetcherEnrichedMedia{
			RawInstagramMedia: &em.Media,
			Insights:          em.Insights,
			UserInfo: map[string]interface{}{
				"name":                name,
				"username":            username,
				"profile_picture_url": profilePic,
			},
		}

		jsonBytes, err := json.Marshal(enrichedStruct)
		if err != nil {
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("media_id", em.Media.ID).Str("function", "parseAllData").Str("stage", "marshal_enriched_media").Msg("failed to marshal enriched media (skipping)")
			logger.CaptureException(err, map[string]string{"platform": "instagram", "component": "immediate-processor", "stage": "marshal_enriched_media", "media_id": em.Media.ID}, nil)
			continue
		}
		var enrichedData map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &enrichedData); err != nil {
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("media_id", em.Media.ID).Str("function", "parseAllData").Str("stage", "unmarshal_enriched_media").Msg("failed to unmarshal enriched media (skipping)")
			logger.CaptureException(err, map[string]string{"platform": "instagram", "component": "immediate-processor", "stage": "unmarshal_enriched_media", "media_id": em.Media.ID}, nil)
			continue
		}

		post, err := p.Parser.ParseMediaWithInsights(enrichedData, igID)
		if err != nil || post == nil {
			p.Logger.Warn().Err(err).Str("media_id", em.Media.ID).Str("function", "parseAllData").Str("stage", "parse_media").Msg("parse media returned nil (skipping)")
			if err != nil {
				logger.CaptureException(err, map[string]string{"platform": "instagram", "component": "immediate-processor", "stage": "parse_media", "media_id": em.Media.ID}, nil)
			}
			continue
		}
		parsed.Posts = append(parsed.Posts, *post)
	}

	for _, daily := range dailyInsights {
		if daily.Data == nil {
			continue
		}

		dateStr := daily.Date.Format("2006-01-02")
		recordID := fmt.Sprintf("%s_%s", igID, dateStr)

		in, err := p.Parser.ParseInsights(daily.Data, igID, username, name, profilePic, recordID)
		if err != nil {
			p.Logger.Debug().Err(err).Str("date", dateStr).Msg("parse daily insights failed")
			continue
		}
		if in == nil {
			continue
		}

		if fc, ok := userInfo["followers_count"]; ok {
			in.FollowersCount = getInt64Value(fc)
			in.FollowerCount = in.FollowersCount
		}
		if fc, ok := userInfo["follows_count"]; ok {
			in.FollowsCount = getInt64Value(fc)
		}
		if mc, ok := userInfo["media_count"]; ok {
			in.MediaCount = getInt64Value(mc)
		}

		in.Metadata = map[string]string{"source": "live_fetch_daily"}
		in.CreatedTime = daily.Date
		in.UpdatedTime = time.Now().UTC()

		parsed.Insights = append(parsed.Insights, *in)
	}

	p.Logger.Info().
		Str("instagram_id", igID).
		Int("daily_insights_parsed", len(parsed.Insights)).
		Msg("Parsed daily insights")

	return parsed, nil
}

func (p *Processor) store(ctx context.Context, data *ParsedData) error {
	var chPosts []*clickhousemodels.InstagramPost
	for _, pp := range data.Posts {
		chPosts = append(chPosts, p.Sink.ConvertInstagramPost(&pp))
	}
	if len(chPosts) > 0 {
		if err := p.Sink.BulkInsertInstagramPosts(ctx, chPosts); err != nil {
			return err
		}
	}

	if len(data.Insights) > 0 {
		var chInsights []*clickhousemodels.InstagramInsight
		for i := range data.Insights {
			chInsights = append(chInsights, p.Sink.ConvertInstagramInsight(&data.Insights[i]))
		}
		if err := p.Sink.BulkInsertInstagramInsights(ctx, chInsights); err != nil {
			return err
		}
	}
	return nil
}

func (p *Processor) sendPusherNotification(account *mongomodels.SocialIntegration, workspaceID, originalState string) {
	if p.PusherClient == nil {
		return
	}

	accountID := account.InstagramID
	if accountID == "" {
		accountID = account.PlatformIdentifier
	}
	if accountID == "" {
		return
	}

	// Frontend subscribes to ig-analytics-channel-{workspace_id}-{instagram_id}.
	channel := fmt.Sprintf("ig-analytics-channel-%s-%s", workspaceID, accountID)
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
	} else {
		p.Logger.Debug().
			Str("channel", channel).
			Str("event", event).
			Msg("Sent Pusher notification")
	}
}

func (p *Processor) sendEmailNotification(userID, workspaceID, accountID, accountName string) {
	if p.Notifier == nil {
		return
	}

	err := p.Notifier.SendAnalyticsNotification(
		userID,
		workspaceID,
		"instagram",
		accountID,
		accountName,
		false,
	)

	if err != nil {
		p.Logger.Warn().
			Err(err).
			Str("error_message", err.Error()).
			Str("user_id", userID).
			Str("workspace_id", workspaceID).
			Str("account_id", accountID).
			Str("function", "sendEmailNotification").
			Msg("Failed to send analytics notification to backend")
	} else {
		p.Logger.Info().
			Str("user_id", userID).
			Str("workspace_id", workspaceID).
			Str("account_id", accountID).
			Msg("Analytics notification sent to backend successfully")
	}
}

func getInt64Value(v interface{}) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	default:
		return 0
	}
}
