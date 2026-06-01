package processor

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/notification"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
)

const fullSyncDays = 86

type WorkOrder struct {
	ID          string `json:"id"`
	AccountID   string `json:"account_id"`
	AccessToken string `json:"access_token"`
	AccountType string `json:"account_type"`
	BoardID     string `json:"board_id,omitempty"`
	WorkspaceID string `json:"workspace_id"`
	SyncType    string `json:"sync_type"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
}

type Processor struct {
	mongoRepo       mongodb.UnifiedSocialRepository
	sink            *conversions.ClickHouseSink
	pinterestClient social.PinterestAPI
	notifier        *notification.Service
	pusher          *notification.PusherClient
	log             *logger.Logger
	cfg             *config.Config
}

// New creates a new Pinterest Processor with all dependencies.
// The Pinterest API client is created internally.
func New(
	mongoRepo mongodb.UnifiedSocialRepository,
	sink *conversions.ClickHouseSink,
	notifier *notification.Service,
	pusher *notification.PusherClient,
	log *logger.Logger,
	cfg *config.Config,
) *Processor {
	return &Processor{
		mongoRepo:       mongoRepo,
		sink:            sink,
		pinterestClient: social.NewPinterestClient(),
		notifier:        notifier,
		pusher:          pusher,
		log:             log,
		cfg:             cfg,
	}
}

func generateRecordID(id string, date time.Time) string {
	hash := md5.Sum([]byte(id + "_" + date.Format("20060102")))
	return hex.EncodeToString(hash[:])
}

func parsePinterestTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func parsePinterestDateRange(startDateStr, endDateStr string) (time.Time, time.Time, bool, error) {
	startDateStr = strings.TrimSpace(startDateStr)
	endDateStr = strings.TrimSpace(endDateStr)
	if startDateStr == "" && endDateStr == "" {
		return time.Time{}, time.Time{}, false, nil
	}
	if startDateStr == "" || endDateStr == "" {
		return time.Time{}, time.Time{}, false, fmt.Errorf("both start_date and end_date are required")
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
		return time.Time{}, time.Time{}, false, fmt.Errorf("end_date must be on or after start_date")
	}

	return startDate.UTC(), endDate.UTC(), true, nil
}

func (p *Processor) ProcessAccount(ctx context.Context, wo WorkOrder) (err error) {
	var accountID primitive.ObjectID
	hasAccountID := false

	if wo.ID != "" {
		accountID, err = primitive.ObjectIDFromHex(wo.ID)
		if err != nil {
			p.log.Warn().Err(err).Str("account_id", wo.ID).Msg("Invalid account ID, skipping retry metadata cleanup")
		} else {
			hasAccountID = true
			account, findErr := p.mongoRepo.FindByID(ctx, accountID)
			if findErr != nil {
				p.log.Warn().Err(findErr).Str("account_id", wo.ID).Msg("Failed to fetch account for retry metadata cleanup")
			} else if account != nil && mongodb.HasProcessingErrorMeta(account.MetaData) {
				if clearErr := p.mongoRepo.ClearProcessingError(ctx, accountID); clearErr != nil {
					p.log.Warn().Err(clearErr).Str("account_id", wo.ID).Msg("Failed to clear stale processing error before retry")
				}
			}
		}
	}

	defer func() {
		if !hasAccountID {
			return
		}
		if err != nil {
			if recordErr := p.mongoRepo.RecordProcessingError(ctx, accountID, err.Error()); recordErr != nil {
				p.log.Warn().Err(recordErr).Str("account_id", wo.ID).Msg("Failed to record processing error")
			}
			return
		}
		if clearErr := p.mongoRepo.ClearProcessingError(ctx, accountID); clearErr != nil {
			p.log.Warn().Err(clearErr).Str("account_id", wo.ID).Msg("Failed to clear processing error")
		}
	}()

	accessToken := wo.AccessToken
	if decrypted, err := crypto.DecryptToken(accessToken, p.cfg.DecryptionKey); err == nil {
		accessToken = decrypted
	}

	startDate, endDate, hasRequestedRange, err := parsePinterestDateRange(wo.StartDate, wo.EndDate)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: invalid date range: %w", err)
	}

	now := time.Now().UTC()
	if !hasRequestedRange {
		endDate = now
		startDate = endDate.AddDate(0, 0, -fullSyncDays)
	}

	userAccount, err := p.pinterestClient.GetUserAccount(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: failed to fetch user account: %w", err)
	}

	chUser := clickhouse.PinterestUser{
		RecordID:       generateRecordID(userAccount.ID, now),
		UserID:         userAccount.ID,
		Username:       userAccount.Username,
		About:          userAccount.About,
		ProfileImage:   userAccount.ProfileImage,
		WebsiteURL:     userAccount.WebsiteURL,
		BusinessName:   userAccount.BusinessName,
		BoardCount:     userAccount.BoardCount,
		PinCount:       userAccount.PinCount,
		AccountType:    userAccount.AccountType,
		FollowerCount:  userAccount.FollowerCount,
		FollowingCount: userAccount.FollowingCount,
		MonthlyViews:   userAccount.MonthlyViews,
		InsertedAt:     now,
	}
	if err := p.sink.BulkInsertPinterestUsers(ctx, []clickhouse.PinterestUser{chUser}); err != nil {
		p.log.Warn().Err(err).Str("error_message", err.Error()).Str("user_id", userAccount.ID).Str("function", "ProcessAccount").Str("stage", "insert_user").Msg("Failed to insert user (continuing)")
		logger.CaptureException(err, map[string]string{"platform": "pinterest", "component": "immediate-processor", "stage": "insert_user", "account_id": wo.AccountID}, nil)
	}

	userAnalytics, err := p.pinterestClient.GetUserAccountAnalytics(ctx, accessToken, startDate, endDate)
	if err != nil {
		p.log.Warn().Err(err).Str("error_message", err.Error()).Str("account_id", wo.AccountID).Str("function", "ProcessAccount").Str("stage", "fetch_user_analytics").Msg("Failed to fetch user analytics (continuing)")
		logger.CaptureException(err, map[string]string{"platform": "pinterest", "component": "immediate-processor", "stage": "fetch_user_analytics", "account_id": wo.AccountID}, nil)
	} else if userAnalytics != nil {
		var userInsights []clickhouse.PinterestUserInsight
		for _, metric := range userAnalytics.All.DailyMetrics {
			if metric.DataStatus == kafkamodels.PinterestDataStatusProcessing ||
				metric.DataStatus == kafkamodels.PinterestDataStatusBeforePinCreated ||
				metric.DataStatus == kafkamodels.PinterestDataStatusBeforeBusinessCreated {
				continue
			}

			metricDate, _ := time.Parse("2006-01-02", metric.Date)
			userInsights = append(userInsights, clickhouse.PinterestUserInsight{
				RecordID:           generateRecordID(userAccount.ID, metricDate),
				UserID:             userAccount.ID,
				Date:               metricDate,
				DataStatus:         metric.DataStatus,
				Impression:         social.GetInt64FromMetrics(metric.Metrics, "IMPRESSION"),
				PinClicks:          social.GetInt64FromMetrics(metric.Metrics, "PIN_CLICK"),
				PinClickRate:       social.GetFloat64FromMetrics(metric.Metrics, "PIN_CLICK_RATE"),
				OutboundClicks:     social.GetInt64FromMetrics(metric.Metrics, "OUTBOUND_CLICK"),
				Saves:              social.GetInt64FromMetrics(metric.Metrics, "SAVE"),
				SaveRate:           social.GetFloat64FromMetrics(metric.Metrics, "SAVE_RATE"),
				Clickthrough:       social.GetInt64FromMetrics(metric.Metrics, "CLICKTHROUGH"),
				ClickthroughRate:   social.GetFloat64FromMetrics(metric.Metrics, "CLICKTHROUGH_RATE"),
				Engagement:         social.GetInt64FromMetrics(metric.Metrics, "ENGAGEMENT"),
				EngagementRate:     social.GetFloat64FromMetrics(metric.Metrics, "ENGAGEMENT_RATE"),
				VideoMRCView:       social.GetInt64FromMetrics(metric.Metrics, "VIDEO_MRC_VIEW"),
				VideoStart:         social.GetInt64FromMetrics(metric.Metrics, "VIDEO_START"),
				Video10sView:       social.GetInt64FromMetrics(metric.Metrics, "VIDEO_10S_VIEW"),
				VideoAvgWatchTime:  social.GetInt64FromMetrics(metric.Metrics, "VIDEO_AVG_WATCH_TIME"),
				VideoV50WatchTime:  social.GetInt64FromMetrics(metric.Metrics, "VIDEO_V50_WATCH_TIME"),
				FullScreenPlay:     social.GetInt64FromMetrics(metric.Metrics, "FULL_SCREEN_PLAY"),
				FullScreenPlaytime: social.GetInt64FromMetrics(metric.Metrics, "FULL_SCREEN_PLAYTIME"),
				ProfileVisit:       social.GetInt64FromMetrics(metric.Metrics, "PROFILE_VISIT"),
				Closeup:            social.GetInt64FromMetrics(metric.Metrics, "CLOSEUP"),
				Quartile95sPercent: social.GetInt64FromMetrics(metric.Metrics, "QUARTILE_95S_PERCENT_VIEW"),
				InsertedAt:         now,
			})
		}
		if len(userInsights) > 0 {
			if err := p.sink.BulkInsertPinterestUserInsights(ctx, userInsights); err != nil {
				p.log.Warn().Err(err).Str("error_message", err.Error()).Str("user_id", userAccount.ID).Str("function", "ProcessAccount").Str("stage", "insert_user_insights").Msg("Failed to insert user insights (continuing)")
				logger.CaptureException(err, map[string]string{"platform": "pinterest", "component": "immediate-processor", "stage": "insert_user_insights", "account_id": wo.AccountID}, nil)
			}
		}
	}

	if wo.AccountType == kafkamodels.PinterestAccountTypeBoard && wo.BoardID != "" {
		if err := p.processBoardData(ctx, accessToken, wo, userAccount.ID, startDate, endDate, hasRequestedRange, now); err != nil {
			return err
		}
	} else {
		if err := p.processAllBoardsData(ctx, accessToken, wo, userAccount.ID, startDate, endDate, hasRequestedRange, now); err != nil {
			return err
		}
	}

	if hasAccountID {
		if err := p.mongoRepo.UpdateAnalyticsTimestamp(ctx, accountID, "analytics", now); err != nil {
			p.log.Warn().Err(err).Str("error_message", err.Error()).Str("account_id", wo.AccountID).Str("function", "ProcessAccount").Str("stage", "update_analytics_timestamp").Msg("Failed to update analytics timestamp")
		}
	}

	return nil
}

func (p *Processor) processBoardData(
	ctx context.Context,
	accessToken string,
	wo WorkOrder,
	userID string,
	startDate, endDate time.Time,
	hasRequestedRange bool,
	now time.Time,
) error {
	board, err := p.pinterestClient.GetBoard(ctx, accessToken, wo.BoardID)
	if err != nil {
		return fmt.Errorf("Processor.processBoardData: failed to fetch board: %w", err)
	}

	boardCreatedAt := parsePinterestTime(board.CreatedAt)
	owner := social.GetStringFromMap(board.Owner, "username")
	imageCoverURL := social.GetStringFromMap(board.Media, "image_cover_url")

	chBoard := clickhouse.PinterestBoard{
		RecordID:          generateRecordID(board.ID, now),
		BoardID:           board.ID,
		UserID:            userID,
		Name:              board.Name,
		Description:       board.Description,
		Privacy:           board.Privacy,
		PinCount:          fmt.Sprintf("%d", board.PinCount),
		FollowerCount:     fmt.Sprintf("%d", board.FollowerCount),
		CollaboratorCount: fmt.Sprintf("%d", board.CollaboratorCount),
		Owner:             owner,
		ImageCoverURL:     imageCoverURL,
		PinThumbnailURLs:  board.PinThumbnailURLs,
		CreatedAt:         boardCreatedAt,
		InsertedAt:        now,
	}
	if err := p.sink.BulkInsertPinterestBoards(ctx, []clickhouse.PinterestBoard{chBoard}); err != nil {
		p.log.Warn().Err(err).Str("error_message", err.Error()).Str("board_id", board.ID).Str("function", "processBoardData").Str("stage", "insert_board").Msg("Failed to insert board (continuing)")
		logger.CaptureException(err, map[string]string{"platform": "pinterest", "component": "immediate-processor", "stage": "insert_board", "board_id": board.ID}, nil)
	}

	if err := p.fetchPinsForBoard(ctx, accessToken, wo, userID, board.ID, startDate, endDate, hasRequestedRange, now); err != nil {
		return err
	}
	p.sendPusherNotification(wo.WorkspaceID, board.ID)
	return nil
}

// sendPusherNotification fires a real-time Pusher event for a single Pinterest board.
// Frontend subscribes to pt-analytics-channel-{workspaceID}-{boardID} and listens for
// event syncing-{workspaceID}-{boardID}.
func (p *Processor) sendPusherNotification(workspaceID, boardID string) {
	if p.pusher == nil {
		return
	}
	channel := fmt.Sprintf("pt-analytics-channel-%s-%s", workspaceID, boardID)
	event := fmt.Sprintf("syncing-%s-%s", workspaceID, boardID)
	data := map[string]interface{}{
		"state":                     "Processed",
		"account":                   boardID,
		"last_analytics_updated_at": time.Now().UTC().Format("2006-01-02"),
	}
	if err := p.pusher.Trigger(channel, event, data); err != nil {
		p.log.Warn().
			Err(err).
			Str("error_message", err.Error()).
			Str("channel", channel).
			Str("event", event).
			Str("function", "sendPusherNotification").
			Msg("Failed to send Pinterest Pusher notification")
	} else {
		p.log.Debug().
			Str("channel", channel).
			Str("event", event).
			Msg("Sent Pinterest Pusher notification")
	}
}

func (p *Processor) processAllBoardsData(
	ctx context.Context,
	accessToken string,
	wo WorkOrder,
	userID string,
	startDate, endDate time.Time,
	hasRequestedRange bool,
	now time.Time,
) error {
	boards, err := p.pinterestClient.GetBoards(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("Processor.processAllBoardsData: failed to fetch boards: %w", err)
	}

	for _, board := range boards.Items {
		boardCreatedAt := parsePinterestTime(board.CreatedAt)
		owner := social.GetStringFromMap(board.Owner, "username")
		imageCoverURL := social.GetStringFromMap(board.Media, "image_cover_url")

		chBoard := clickhouse.PinterestBoard{
			RecordID:          generateRecordID(board.ID, now),
			BoardID:           board.ID,
			UserID:            userID,
			Name:              board.Name,
			Description:       board.Description,
			Privacy:           board.Privacy,
			PinCount:          fmt.Sprintf("%d", board.PinCount),
			FollowerCount:     fmt.Sprintf("%d", board.FollowerCount),
			CollaboratorCount: fmt.Sprintf("%d", board.CollaboratorCount),
			Owner:             owner,
			ImageCoverURL:     imageCoverURL,
			PinThumbnailURLs:  board.PinThumbnailURLs,
			CreatedAt:         boardCreatedAt,
			InsertedAt:        now,
		}
		if err := p.sink.BulkInsertPinterestBoards(ctx, []clickhouse.PinterestBoard{chBoard}); err != nil {
			p.log.Warn().Err(err).Str("error_message", err.Error()).Str("board_id", board.ID).Str("function", "processAllBoardsData").Str("stage", "insert_board").Msg("Failed to insert board (continuing)")
			logger.CaptureException(err, map[string]string{"platform": "pinterest", "component": "immediate-processor", "stage": "insert_board", "board_id": board.ID}, nil)
		}

		if board.Privacy != "PUBLIC" {
			continue
		}

		if err := p.fetchPinsForBoard(ctx, accessToken, wo, userID, board.ID, startDate, endDate, hasRequestedRange, now); err != nil {
			p.log.Warn().Err(err).Str("error_message", err.Error()).Str("board_id", board.ID).Str("function", "processAllBoardsData").Str("stage", "fetch_pins_for_board").Msg("Failed to process pins for board (continuing)")
			logger.CaptureException(err, map[string]string{"platform": "pinterest", "component": "immediate-processor", "stage": "fetch_pins_for_board", "board_id": board.ID}, nil)
		} else {
			p.sendPusherNotification(wo.WorkspaceID, board.ID)
		}
	}

	return nil
}

func (p *Processor) fetchPinsForBoard(
	ctx context.Context,
	accessToken string,
	wo WorkOrder,
	userID, boardID string,
	startDate, endDate time.Time,
	hasRequestedRange bool,
	now time.Time,
) error {
	bookmark := ""
	pageSize := 250
	endExclusive := endDate.AddDate(0, 0, 1)
	reachedCutoff := false

	for {
		pinsResp, err := p.pinterestClient.GetBoardPins(ctx, accessToken, boardID, pageSize, bookmark)
		if err != nil {
			return fmt.Errorf("Processor.fetchPinsForBoard: failed to fetch pins: %w", err)
		}

		if len(pinsResp.Items) == 0 {
			break
		}

		var filteredPins []social.PinterestPin
		for _, pin := range pinsResp.Items {
			pinCreatedAt := parsePinterestTime(pin.CreatedAt)
			if hasRequestedRange {
				if pinCreatedAt.Before(startDate) {
					reachedCutoff = true
					break
				}
				if !pinCreatedAt.Before(endExclusive) {
					continue
				}
			}
			filteredPins = append(filteredPins, pin)
		}

		pinIDs := make([]string, 0, len(filteredPins))
		for _, pin := range filteredPins {
			if pin.ID != "" {
				pinIDs = append(pinIDs, pin.ID)
			}
		}

		analyticsMap := map[string]*social.PinterestPinAnalyticsResponse{}
		if len(pinIDs) > 0 {
			analyticsMap, err = p.pinterestClient.GetMultiPinAnalytics(ctx, accessToken, pinIDs, startDate, endDate)
			if err != nil {
				p.log.Warn().Err(err).Str("error_message", err.Error()).Str("board_id", boardID).Str("function", "fetchPinsForBoard").Str("stage", "fetch_multi_pin_analytics").Msg("Failed to fetch multi-pin analytics (continuing)")
				logger.CaptureException(err, map[string]string{"platform": "pinterest", "component": "immediate-processor", "stage": "fetch_multi_pin_analytics", "board_id": boardID}, nil)
			}
		}

		var pins []clickhouse.PinterestPin
		var pinInsights []clickhouse.PinterestPinInsight

		for _, pin := range filteredPins {
			pinCreatedAt := parsePinterestTime(pin.CreatedAt)
			dayOfWeek := pinCreatedAt.Weekday().String()
			hourOfDay := pinCreatedAt.Hour()

			mediaType := social.GetMediaField(pin.Media, "media_type")
			coverImageURL := social.GetPinCoverImageURL(pin)
			videoURL := social.GetMediaField(pin.Media, "video_url")
			duration := social.GetMediaField(pin.Media, "duration")
			height := social.GetMediaField(pin.Media, "height")
			width := social.GetMediaField(pin.Media, "width")
			boardOwner := social.GetStringFromMap(pin.BoardOwner, "username")

			isStandard := "0"
			if pin.IsStandard {
				isStandard = "1"
			}
			isOwner := "0"
			if pin.IsOwner {
				isOwner = "1"
			}
			hasBeenPromoted := "0"
			if pin.HasBeenPromoted {
				hasBeenPromoted = "1"
			}

			pins = append(pins, clickhouse.PinterestPin{
				RecordID:        generateRecordID(pin.ID, now),
				PinID:           pin.ID,
				UserID:          userID,
				BoardID:         pin.BoardID,
				BoardSectionID:  pin.BoardSectionID,
				ParentPinID:     pin.ParentPinID,
				Title:           pin.Title,
				Note:            pin.Note,
				Description:     pin.Description,
				Link:            pin.Link,
				DominantColor:   pin.DominantColor,
				CreativeType:    pin.CreativeType,
				MediaType:       mediaType,
				CoverImageURL:   coverImageURL,
				VideoURL:        videoURL,
				Duration:        duration,
				Height:          height,
				Width:           width,
				IsStandard:      isStandard,
				IsOwner:         isOwner,
				HasBeenPromoted: hasBeenPromoted,
				BoardOwner:      boardOwner,
				CreatedAt:       pinCreatedAt,
				DayOfWeek:       dayOfWeek,
				HourOfDay:       hourOfDay,
				InsertedAt:      now,
			})

			if analytics, ok := analyticsMap[pin.ID]; ok && analytics != nil {
				for _, metric := range analytics.All.DailyMetrics {
					if metric.DataStatus == kafkamodels.PinterestDataStatusProcessing ||
						metric.DataStatus == kafkamodels.PinterestDataStatusBeforePinCreated ||
						metric.DataStatus == kafkamodels.PinterestDataStatusBeforeBusinessCreated {
						continue
					}

					metricDate, _ := time.Parse("2006-01-02", metric.Date)
					engagement := social.GetInt64FromMetrics(metric.Metrics, "PIN_CLICK") +
						social.GetInt64FromMetrics(metric.Metrics, "CLICKTHROUGH") +
						social.GetInt64FromMetrics(metric.Metrics, "SAVE") +
						social.GetInt64FromMetrics(metric.Metrics, "OUTBOUND_CLICK")

					impression := social.GetInt64FromMetrics(metric.Metrics, "IMPRESSION")
					var engagementRate float64
					if impression > 0 {
						engagementRate = float64(engagement) / float64(impression)
					}

					pinInsights = append(pinInsights, clickhouse.PinterestPinInsight{
						RecordID:           generateRecordID(pin.ID, metricDate),
						PinID:              pin.ID,
						UserID:             userID,
						BoardID:            pin.BoardID,
						Date:               metricDate,
						DataStatus:         metric.DataStatus,
						Impression:         impression,
						PinClicks:          social.GetInt64FromMetrics(metric.Metrics, "PIN_CLICK"),
						OutboundClicks:     social.GetInt64FromMetrics(metric.Metrics, "OUTBOUND_CLICK"),
						Saves:              social.GetInt64FromMetrics(metric.Metrics, "SAVE"),
						SaveRate:           social.GetFloat64FromMetrics(metric.Metrics, "SAVE_RATE"),
						Clickthrough:       social.GetInt64FromMetrics(metric.Metrics, "CLICKTHROUGH"),
						ClickthroughRate:   social.GetFloat64FromMetrics(metric.Metrics, "CLICKTHROUGH_RATE"),
						Engagement:         engagement,
						EngagementRate:     engagementRate,
						VideoMRCView:       social.GetInt64FromMetrics(metric.Metrics, "VIDEO_MRC_VIEW"),
						VideoStart:         social.GetInt64FromMetrics(metric.Metrics, "VIDEO_START"),
						Video10sView:       social.GetInt64FromMetrics(metric.Metrics, "VIDEO_10S_VIEW"),
						VideoAvgWatchTime:  social.GetInt64FromMetrics(metric.Metrics, "VIDEO_AVG_WATCH_TIME"),
						VideoV50WatchTime:  social.GetInt64FromMetrics(metric.Metrics, "VIDEO_V50_WATCH_TIME"),
						FullScreenPlay:     social.GetInt64FromMetrics(metric.Metrics, "FULL_SCREEN_PLAY"),
						FullScreenPlaytime: social.GetInt64FromMetrics(metric.Metrics, "FULL_SCREEN_PLAYTIME"),
						ProfileVisit:       social.GetInt64FromMetrics(metric.Metrics, "PROFILE_VISIT"),
						Closeup:            social.GetInt64FromMetrics(metric.Metrics, "CLOSEUP"),
						Quartile95sPercent: social.GetInt64FromMetrics(metric.Metrics, "QUARTILE_95S_PERCENT_VIEW"),
						UserFollow:         social.GetInt64FromMetrics(metric.Metrics, "USER_FOLLOW"),
						DayOfWeek:          metricDate.Weekday().String(),
						HourOfDay:          metricDate.Hour(),
						InsertedAt:         now,
					})
				}
			}
		}

		if len(pins) > 0 {
			if err := p.sink.BulkInsertPinterestPins(ctx, pins); err != nil {
				p.log.Warn().Err(err).Str("error_message", err.Error()).Str("board_id", boardID).Str("function", "fetchPinsForBoard").Str("stage", "insert_pins").Msg("Failed to insert pins (continuing)")
				logger.CaptureException(err, map[string]string{"platform": "pinterest", "component": "immediate-processor", "stage": "insert_pins", "board_id": boardID}, nil)
			}
		}

		if len(pinInsights) > 0 {
			if err := p.sink.BulkInsertPinterestPinInsights(ctx, pinInsights); err != nil {
				p.log.Warn().Err(err).Str("error_message", err.Error()).Str("board_id", boardID).Str("function", "fetchPinsForBoard").Str("stage", "insert_pin_insights").Msg("Failed to insert pin insights (continuing)")
				logger.CaptureException(err, map[string]string{"platform": "pinterest", "component": "immediate-processor", "stage": "insert_pin_insights", "board_id": boardID}, nil)
			}
		}

		bookmark = pinsResp.Bookmark
		if bookmark == "" || reachedCutoff {
			break
		}
	}

	return nil
}
