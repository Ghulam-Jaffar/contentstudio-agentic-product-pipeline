package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/sync/semaphore"
)

var ErrUnauthorized = errors.New("unauthorized: invalid or expired token")

var accountSemaphores sync.Map

func semForAccount(id string, capacity int64) *semaphore.Weighted {
	if v, ok := accountSemaphores.Load(id); ok {
		return v.(*semaphore.Weighted)
	}
	sem := semaphore.NewWeighted(capacity)
	if old, loaded := accountSemaphores.LoadOrStore(id, sem); loaded {
		return old.(*semaphore.Weighted)
	}
	return sem
}

func isUnauthorizedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "status 401") || strings.Contains(errStr, "unauthorized")
}

// parsePinterestDate tries multiple date formats used by Pinterest API
func parsePinterestDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	return time.Time{}
}

// parsePinterestDateWithLog tries multiple date formats and logs failures
func parsePinterestDateWithLog(dateStr string, itemType string, itemID string, log *logger.Logger) time.Time {
	if dateStr == "" {
		log.Debug().Str("item_type", itemType).Str("item_id", itemID).Msg("Empty created_at date")
		return time.Time{}
	}

	t := parsePinterestDate(dateStr)
	if t.IsZero() {
		log.Warn().Str("item_type", itemType).Str("item_id", itemID).Str("date_str", dateStr).Msg("Failed to parse created_at date")
	}
	return t
}

const (
	maxWorkers        = 5
	workOrderChanSize = 100
	timestampChanSize = 200
	consumerGroup     = "pinterest-fetcher-group"

	// maxConcurrentAccounts is the max number of Pinterest accounts processed simultaneously.
	maxConcurrentAccounts = 50

	topicWorkOrderBatch = "work-order-pinterest"

	topicRawUsers        = "raw-pinterest-users"
	topicRawBoards       = "raw-pinterest-boards"
	topicRawPins         = "raw-pinterest-pins"
	topicRawPinInsights  = "raw-pinterest-pin-insights"
	topicRawUserInsights = "raw-pinterest-user-insights"

	idleTimeout = 15 * time.Minute

	fullSyncDays        = 86
	incrementalSyncDays = 3
	immediateSyncDays   = 90
	// Skip today and yesterday (data may be incomplete/processing)
	analyticsEndDateOffset = 2
	fullPageSize           = 250
	incrementalPageSize    = 25
	immediatePageSize      = 250
	maxIncrementalPages    = 2
)

type WorkOrderMessage struct {
	AccountID   string
	Value       []byte
	AccessToken string
}

type TimestampUpdateRequest struct {
	AccountID string
	UserID    string
}

type Service struct {
	totalProcessed        int64
	totalFailed           int64
	PinterestClient       social.PinterestAPI
	Producer              kafka2.Producer
	Consumer              kafka2.Consumer
	MongoRepo             mongodb.UnifiedSocialRepository
	Logger                *logger.Logger
	DecryptionKey         string
	MaxWorkers            int
	MaxConcurrentAccounts int
	IdleTimeout           time.Duration
	IdleCheckPeriod       time.Duration
}

func NewService(
	pinterestClient social.PinterestAPI,
	producer kafka2.Producer,
	consumer kafka2.Consumer,
	mongoRepo mongodb.UnifiedSocialRepository,
	log *logger.Logger,
	decryptionKey string,
) *Service {
	return &Service{
		PinterestClient:       pinterestClient,
		Producer:              producer,
		Consumer:              consumer,
		MongoRepo:             mongoRepo,
		Logger:                log,
		DecryptionKey:         decryptionKey,
		MaxWorkers:            maxWorkers,
		MaxConcurrentAccounts: maxConcurrentAccounts,
		IdleTimeout:           idleTimeout,
		IdleCheckPeriod:       30 * time.Second,
	}
}

func (s *Service) Run(ctx context.Context) error {
	timestampUpdateChan := make(chan TimestampUpdateRequest, timestampChanSize)

	var lastMessageTime int64 = time.Now().UnixNano()

	maxConc := s.MaxConcurrentAccounts
	if maxConc <= 0 {
		maxConc = maxConcurrentAccounts
	}
	accountSem := semaphore.NewWeighted(int64(maxConc))
	var dispatchWg sync.WaitGroup

	var timestampWg sync.WaitGroup
	s.startTimestampUpdater(ctx, &timestampWg, timestampUpdateChan)
	s.startBatchConsumer(ctx, accountSem, &dispatchWg, timestampUpdateChan, &lastMessageTime)

	s.Logger.Info().Int("max_concurrent_accounts", maxConc).Msg("Pinterest Fetcher service is running")

	<-ctx.Done()
	s.Logger.Info().Msg("Context cancelled, stopping service...")

	dispatchWg.Wait()
	s.Logger.Info().
		Int64("total_processed", atomic.LoadInt64(&s.totalProcessed)).
		Int64("total_failed", atomic.LoadInt64(&s.totalFailed)).
		Msg("Pinterest Fetcher service stopped")
	s.Logger.Info().Msg("All dispatch goroutines finished")

	close(timestampUpdateChan)
	timestampWg.Wait()
	s.Logger.Info().Msg("Timestamp updater finished")

	return nil
}

func (s *Service) startWorkerPool(
	ctx context.Context,
	wg *sync.WaitGroup,
	workOrderChan chan WorkOrderMessage,
	timestampUpdateChan chan TimestampUpdateRequest,
	lastActivityTime *int64,
) {
	for i := 0; i < s.MaxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			s.workOrderProcessor(ctx, workerID, workOrderChan, timestampUpdateChan, lastActivityTime)
		}(i)
	}
	s.Logger.Info().Int("workers", s.MaxWorkers).Msg("Started Pinterest fetcher workers")
}

func (s *Service) startBatchConsumer(
	ctx context.Context,
	accountSem *semaphore.Weighted,
	dispatchWg *sync.WaitGroup,
	timestampUpdateChan chan TimestampUpdateRequest,
	lastMessageTime *int64,
) {
	go func() {
		s.Logger.Info().Str("topic", topicWorkOrderBatch).Msg("Starting batch Kafka consumer")
		err := s.Consumer.Consume(ctx, []string{topicWorkOrderBatch}, func(ctx context.Context, topic string, key, value []byte) error {
			if lastMessageTime != nil {
				atomic.StoreInt64(lastMessageTime, time.Now().UnixNano())
			}

			var batch kafkamodels.PinterestBatchWorkOrder
			if err := json.Unmarshal(value, &batch); err != nil {
				s.Logger.Error().Err(err).Str("function", "startBatchConsumer").Str("stage", "unmarshal_batch_work_order").Msg("Failed to unmarshal batch work order")
				return nil
			}

			total := len(batch.Accounts)
			s.Logger.Info().
				Str("batch_id", batch.BatchID).
				Int("accounts", total).
				Msg("Received batch work order, dispatching goroutines")

			var batchWg sync.WaitGroup
			var batchProcessed, batchFailed int64

			for _, account := range batch.Accounts {
				acc := account
				accountPayload, err := json.Marshal(acc)
				if err != nil {
					s.Logger.Error().Err(err).Str("account_id", acc.AccountID).Str("function", "startBatchConsumer").Str("stage", "marshal_account_work_order").Msg("Failed to marshal account work order")
					atomic.AddInt64(&batchFailed, 1)
					continue
				}
				dispatchWg.Add(1)
				batchWg.Add(1)
				go func() {
					defer dispatchWg.Done()
					defer batchWg.Done()
					if err := accountSem.Acquire(ctx, 1); err != nil {
						atomic.AddInt64(&batchFailed, 1)
						return
					}
					defer accountSem.Release(1)
					msg := WorkOrderMessage{
						AccountID:   acc.ID,
						Value:       accountPayload,
						AccessToken: acc.AccessToken,
					}
					s.processWorkOrder(ctx, msg, s.Logger, timestampUpdateChan, nil)
					atomic.AddInt64(&batchProcessed, 1)
				}()
			}

			batchID := batch.BatchID
			go func() {
				batchWg.Wait()
				p := atomic.LoadInt64(&batchProcessed)
				f := atomic.LoadInt64(&batchFailed)
				atomic.AddInt64(&s.totalProcessed, p)
				atomic.AddInt64(&s.totalFailed, f)
				s.Logger.Info().
					Str("batch_id", batchID).
					Int("total", total).
					Int64("processed", p).
					Int64("failed", f).
					Msg("Batch processing complete")
			}()

			return nil
		})

		if err != nil && err != context.Canceled {
			s.Logger.Error().Err(err).Str("function", "startBatchConsumer").Str("stage", "consume_batch").Msg("Batch consumer error")
		}
	}()
}

func (s *Service) startTimestampUpdater(
	ctx context.Context,
	wg *sync.WaitGroup,
	timestampUpdateChan <-chan TimestampUpdateRequest,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Logger.Info().Msg("Timestamp updater started")

		for {
			select {
			case req, ok := <-timestampUpdateChan:
				if !ok {
					s.Logger.Info().Msg("Timestamp update channel closed")
					return
				}

				objectID, err := primitive.ObjectIDFromHex(req.AccountID)
				if err != nil {
					s.Logger.Error().Err(err).Str("error_message", err.Error()).Str("user_id", req.UserID).Str("function", "startTimestampUpdater").Str("stage", "parse_object_id").Msg("Invalid ObjectID for timestamp update")
					continue
				}

				if err := s.MongoRepo.UpdateAnalyticsTimestamp(context.Background(), objectID, "analytics", time.Now().UTC()); err != nil {
					s.Logger.Error().Err(err).Str("error_message", err.Error()).Str("user_id", req.UserID).Str("function", "startTimestampUpdater").Str("stage", "update_analytics_timestamp").Msg("Failed to update analytics timestamp")
				} else {
					s.MongoRepo.ClearProcessingError(context.Background(), objectID)
					s.Logger.Debug().Str("user_id", req.UserID).Msg("Updated analytics timestamp")
				}

			case <-ctx.Done():
				s.Logger.Info().Msg("Timestamp updater stopping")
				return
			}
		}
	}()
}

func (s *Service) workOrderProcessor(
	ctx context.Context,
	workerID int,
	workOrderChan <-chan WorkOrderMessage,
	timestampUpdateChan chan<- TimestampUpdateRequest,
	lastActivityTime *int64,
) {
	workerLog := s.Logger.With().Int("worker_id", workerID).Logger()
	workerLog.Info().Msg("Worker started")

	for {
		select {
		case <-ctx.Done():
			workerLog.Info().Msg("Worker stopped (context cancelled)")
			return
		case msg, ok := <-workOrderChan:
			if !ok {
				workerLog.Info().Msg("Worker stopped (channel closed)")
				return
			}
			// Update activity time when starting work
			if lastActivityTime != nil {
				atomic.StoreInt64(lastActivityTime, time.Now().UnixNano())
			}
			s.processWorkOrder(ctx, msg, &logger.Logger{Logger: workerLog}, timestampUpdateChan, lastActivityTime)
		}
	}
}

func (s *Service) processWorkOrder(
	ctx context.Context,
	msg WorkOrderMessage,
	log *logger.Logger,
	timestampUpdateChan chan<- TimestampUpdateRequest,
	lastActivityTime *int64,
) {
	var wo kafkamodels.PinterestAccountWorkOrder
	if err := json.Unmarshal(msg.Value, &wo); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processWorkOrder").Str("stage", "unmarshal_work_order").Msg("Failed to unmarshal work order")
		return
	}

	if wo.AccountID == "" || wo.AccessToken == "" {
		log.Warn().Str("account_id", wo.AccountID).Msg("Skipping work order with missing account_id or access_token")
		if wo.ID != "" {
			if accountID, parseErr := primitive.ObjectIDFromHex(wo.ID); parseErr == nil {
				s.MongoRepo.RecordProcessingError(ctx, accountID, "Access token is empty or account ID is missing")
			}
		}
		return
	}

	sem := semForAccount(wo.AccountID, 1)
	if err := sem.Acquire(ctx, 1); err != nil {
		log.Warn().Err(err).Str("account_id", wo.AccountID).Msg("Failed to acquire semaphore")
		return
	}
	defer sem.Release(1)

	// Log token info for debugging (only prefix for security)
	tokenPrefix := ""
	if len(wo.AccessToken) > 10 {
		tokenPrefix = wo.AccessToken[:10] + "..."
	}
	log.Info().
		Str("account_id", wo.AccountID).
		Int("received_token_length", len(wo.AccessToken)).
		Str("token_prefix", tokenPrefix).
		Msg("Pinterest fetcher received work order")

	accessToken := wo.AccessToken
	if decrypted, err := crypto.DecryptToken(accessToken, s.DecryptionKey); err == nil {
		decryptedPrefix := ""
		if len(decrypted) > 10 {
			decryptedPrefix = decrypted[:10] + "..."
		}
		log.Info().
			Int("decrypted_token_length", len(decrypted)).
			Str("decrypted_prefix", decryptedPrefix).
			Msg("Token processed in fetcher")
		accessToken = decrypted
	} else {
		log.Info().
			Err(err).
			Int("using_original_token_length", len(accessToken)).
			Msg("Token decryption failed in fetcher, using original")
	}

	log.Info().
		Str("account_id", wo.AccountID).
		Str("account_type", wo.AccountType).
		Str("sync_type", wo.SyncType).
		Msg("Processing Pinterest account")

	now := time.Now().UTC()
	// End date excludes today and yesterday (data may be incomplete)
	endDate := now.AddDate(0, 0, -analyticsEndDateOffset)
	var startDate time.Time
	var pageSize int
	var maxPages int

	switch wo.SyncType {
	case kafkamodels.PinterestSyncTypeFullSync:
		startDate = endDate.AddDate(0, 0, -fullSyncDays)
		pageSize = fullPageSize
		maxPages = 0
	case kafkamodels.PinterestSyncTypeImmediate:
		// Get last 90 days (ending 2 days ago)
		startDate = endDate.AddDate(0, 0, -immediateSyncDays+1)
		pageSize = immediatePageSize
		maxPages = 0
	default:
		// Get last 7 days (ending 2 days ago)
		startDate = endDate.AddDate(0, 0, -incrementalSyncDays+1)
		pageSize = incrementalPageSize
		maxPages = maxIncrementalPages
	}

	log.Info().
		Str("account_id", wo.AccountID).
		Str("sync_type", wo.SyncType).
		Time("start_date", startDate).
		Time("end_date", endDate).
		Msg("Date range for Pinterest analytics")

	userAccount, err := s.PinterestClient.GetUserAccount(ctx, accessToken)
	if err != nil {
		if isUnauthorizedError(err) {
			log.Warn().Err(err).Str("error_message", err.Error()).Str("account_id", wo.AccountID).Str("function", "processWorkOrder").Str("stage", "fetch_user_account").Msg("Unauthorized - cannot fetch user account")
			if accountID, parseErr := primitive.ObjectIDFromHex(wo.ID); parseErr == nil {
				s.MongoRepo.RecordProcessingError(context.Background(), accountID, err.Error())
			}
			return
		}
		log.Error().Err(err).Str("error_message", err.Error()).Str("account_id", wo.AccountID).Str("function", "processWorkOrder").Str("stage", "fetch_user_account").Msg("Failed to fetch user account")
		if accountID, parseErr := primitive.ObjectIDFromHex(wo.ID); parseErr == nil {
			s.MongoRepo.RecordProcessingError(context.Background(), accountID, err.Error())
		}
		return
	}

	rawUser := kafkamodels.RawPinterestUser{
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
		WorkspaceID:    wo.WorkspaceID,
		SavingTime:     now,
	}
	produceMessage(ctx, s.Producer, topicRawUsers, userAccount.ID, rawUser, log)

	userAnalytics, err := s.PinterestClient.GetUserAccountAnalytics(ctx, accessToken, startDate, endDate)
	if err != nil {
		log.Warn().Err(err).Str("account_id", wo.AccountID).Msg("Failed to fetch user analytics")
		if social.IsAuthError(err) {
			if accountID, parseErr := primitive.ObjectIDFromHex(wo.ID); parseErr == nil {
				s.MongoRepo.RecordProcessingError(context.Background(), accountID, err.Error())
			}
			return
		}
	} else if userAnalytics != nil {
		for _, metric := range userAnalytics.All.DailyMetrics {
			if metric.DataStatus == kafkamodels.PinterestDataStatusProcessing ||
				metric.DataStatus == kafkamodels.PinterestDataStatusBeforePinCreated ||
				metric.DataStatus == kafkamodels.PinterestDataStatusBeforeBusinessCreated {
				continue
			}

			metricDate, _ := time.Parse("2006-01-02", metric.Date)
			rawInsight := kafkamodels.RawPinterestUserInsight{
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
				WorkspaceID:        wo.WorkspaceID,
				SavingTime:         now,
			}
			produceMessage(ctx, s.Producer, topicRawUserInsights, userAccount.ID, rawInsight, log)
		}
		log.Info().Str("account_id", wo.AccountID).Int("metrics_count", len(userAnalytics.All.DailyMetrics)).Msg("Produced user insights")
	}

	if wo.AccountType == kafkamodels.PinterestAccountTypeBoard && wo.BoardID != "" {
		s.fetchBoardData(ctx, accessToken, wo, userAccount.ID, startDate, endDate, pageSize, maxPages, now, log, lastActivityTime)
	} else {
		s.fetchAllBoardsData(ctx, accessToken, wo, userAccount.ID, startDate, endDate, pageSize, maxPages, now, log, lastActivityTime)
	}

	// Check context before sending to avoid panic on closed channel during shutdown
	select {
	case <-ctx.Done():
		log.Info().Str("account_id", wo.AccountID).Msg("Context cancelled, skipping timestamp update")
	case timestampUpdateChan <- TimestampUpdateRequest{AccountID: wo.ID, UserID: userAccount.ID}:
	default:
		log.Warn().Str("account_id", wo.AccountID).Msg("Timestamp update channel full, skipping update")
	}

	log.Info().Str("account_id", wo.AccountID).Msg("Completed Pinterest account fetch")
}

func (s *Service) fetchBoardData(
	ctx context.Context,
	accessToken string,
	wo kafkamodels.PinterestAccountWorkOrder,
	userID string,
	startDate, endDate time.Time,
	pageSize, maxPages int,
	now time.Time,
	log *logger.Logger,
	lastActivityTime *int64,
) {
	board, err := s.PinterestClient.GetBoard(ctx, accessToken, wo.BoardID)
	if err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("board_id", wo.BoardID).Str("function", "fetchBoardData").Str("stage", "fetch_board").Msg("Failed to fetch board")
		return
	}

	boardCreatedAt := parsePinterestDate(board.CreatedAt)
	owner := social.GetStringFromMap(board.Owner, "username")
	imageCoverURL := social.GetStringFromMap(board.Media, "image_cover_url")

	rawBoard := kafkamodels.RawPinterestBoard{
		BoardID:           board.ID,
		UserID:            userID,
		Name:              board.Name,
		Description:       board.Description,
		Privacy:           board.Privacy,
		PinCount:          board.PinCount,
		FollowerCount:     board.FollowerCount,
		CollaboratorCount: board.CollaboratorCount,
		CreatedAt:         boardCreatedAt,
		Owner:             owner,
		ImageCoverURL:     imageCoverURL,
		PinThumbnailURLs:  board.PinThumbnailURLs,
		WorkspaceID:       wo.WorkspaceID,
		SavingTime:        now,
	}
	produceMessage(ctx, s.Producer, topicRawBoards, board.ID, rawBoard, log)

	s.fetchPinsForBoard(ctx, accessToken, wo, userID, board.ID, startDate, endDate, pageSize, maxPages, now, log, lastActivityTime)
}

func (s *Service) fetchAllBoardsData(
	ctx context.Context,
	accessToken string,
	wo kafkamodels.PinterestAccountWorkOrder,
	userID string,
	startDate, endDate time.Time,
	pageSize, maxPages int,
	now time.Time,
	log *logger.Logger,
	lastActivityTime *int64,
) {
	boards, err := s.PinterestClient.GetBoards(ctx, accessToken)
	if err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("account_id", wo.AccountID).Str("function", "fetchAllBoardsData").Str("stage", "fetch_boards").Msg("Failed to fetch boards")
		return
	}

	log.Info().Str("account_id", wo.AccountID).Int("boards_count", len(boards.Items)).Msg("Fetched boards")

	// Publish all board metadata first (fast, no API calls)
	for _, board := range boards.Items {
		boardCreatedAt := parsePinterestDate(board.CreatedAt)
		owner := social.GetStringFromMap(board.Owner, "username")
		imageCoverURL := social.GetStringFromMap(board.Media, "image_cover_url")
		rawBoard := kafkamodels.RawPinterestBoard{
			BoardID:           board.ID,
			UserID:            userID,
			Name:              board.Name,
			Description:       board.Description,
			Privacy:           board.Privacy,
			PinCount:          board.PinCount,
			FollowerCount:     board.FollowerCount,
			CollaboratorCount: board.CollaboratorCount,
			CreatedAt:         boardCreatedAt,
			Owner:             owner,
			ImageCoverURL:     imageCoverURL,
			PinThumbnailURLs:  board.PinThumbnailURLs,
			WorkspaceID:       wo.WorkspaceID,
			SavingTime:        now,
		}
		produceMessage(ctx, s.Producer, topicRawBoards, board.ID, rawBoard, log)
	}

	// Fetch pins + analytics for all public boards in parallel (max 5 concurrent per account
	// to stay within Pinterest's per-user rate limit of 10 req/sec)
	boardSem := semaphore.NewWeighted(5)
	var boardWg sync.WaitGroup
	for _, board := range boards.Items {
		if board.Privacy != "PUBLIC" {
			log.Debug().Str("board_id", board.ID).Str("privacy", board.Privacy).Msg("Skipping non-public board")
			continue
		}
		b := board
		boardWg.Add(1)
		go func() {
			defer boardWg.Done()
			if err := boardSem.Acquire(ctx, 1); err != nil {
				return
			}
			defer boardSem.Release(1)
			s.fetchPinsForBoard(ctx, accessToken, wo, userID, b.ID, startDate, endDate, pageSize, maxPages, now, log, lastActivityTime)
		}()
	}
	boardWg.Wait()

	log.Info().Str("account_id", wo.AccountID).Int("boards_count", len(boards.Items)).Msg("Completed fetching all boards")
}

func (s *Service) fetchPinsForBoard(
	ctx context.Context,
	accessToken string,
	wo kafkamodels.PinterestAccountWorkOrder,
	userID, boardID string,
	startDate, endDate time.Time,
	pageSize, maxPages int,
	now time.Time,
	log *logger.Logger,
	lastActivityTime *int64,
) {
	bookmark := ""
	pageCount := 0
	totalPins := 0

	for {
		// Update activity time to prevent idle shutdown during long-running board processing
		if lastActivityTime != nil {
			atomic.StoreInt64(lastActivityTime, time.Now().UnixNano())
		}
		pinsResp, err := s.PinterestClient.GetBoardPins(ctx, accessToken, boardID, pageSize, bookmark)
		if err != nil {
			log.Error().Err(err).Str("error_message", err.Error()).Str("board_id", boardID).Str("function", "fetchPinsForBoard").Str("stage", "fetch_pins").Msg("Failed to fetch pins")
			break
		}

		if len(pinsResp.Items) == 0 {
			break
		}

		pinIDs := make([]string, 0, len(pinsResp.Items))
		for _, pin := range pinsResp.Items {
			if pin.ID != "" {
				pinIDs = append(pinIDs, pin.ID)
			}
		}

		log.Debug().Str("board_id", boardID).Int("pin_count", len(pinIDs)).Msg("Fetching multi-pin analytics")
		analyticsStartTime := time.Now()
		analyticsMap, err := s.PinterestClient.GetMultiPinAnalytics(ctx, accessToken, pinIDs, startDate, endDate)
		analyticsDuration := time.Since(analyticsStartTime)
		if err != nil {
			log.Warn().Err(err).Str("board_id", boardID).Int("pin_count", len(pinIDs)).Dur("duration", analyticsDuration).Msg("Failed to fetch multi-pin analytics")
		} else {
			log.Info().Str("board_id", boardID).Int("pin_count", len(pinIDs)).Int("analytics_count", len(analyticsMap)).Dur("duration", analyticsDuration).Msg("Fetched pin analytics")
		}

		for _, pin := range pinsResp.Items {
			pinCreatedAt := parsePinterestDateWithLog(pin.CreatedAt, "pin", pin.ID, log)
			dayOfWeek := pinCreatedAt.Weekday().String()
			hourOfDay := pinCreatedAt.Hour()

			mediaType := social.GetMediaField(pin.Media, "media_type")
			coverImageURL := social.GetPinCoverImageURL(pin)
			videoURL := social.GetMediaField(pin.Media, "video_url")
			duration := social.GetMediaField(pin.Media, "duration")
			height := social.GetMediaField(pin.Media, "height")
			width := social.GetMediaField(pin.Media, "width")
			boardOwner := social.GetStringFromMap(pin.BoardOwner, "username")

			rawPin := kafkamodels.RawPinterestPin{
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
				IsStandard:      pin.IsStandard,
				IsOwner:         pin.IsOwner,
				HasBeenPromoted: pin.HasBeenPromoted,
				BoardOwner:      boardOwner,
				CreatedAt:       pinCreatedAt,
				DayOfWeek:       dayOfWeek,
				HourOfDay:       hourOfDay,
				WorkspaceID:     wo.WorkspaceID,
				SavingTime:      now,
			}
			produceMessage(ctx, s.Producer, topicRawPins, pin.ID, rawPin, log)

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

					rawInsight := kafkamodels.RawPinterestPinInsight{
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
						WorkspaceID:        wo.WorkspaceID,
						SavingTime:         now,
					}
					produceMessage(ctx, s.Producer, topicRawPinInsights, pin.ID, rawInsight, log)
				}
			}

			totalPins++
		}

		bookmark = pinsResp.Bookmark
		pageCount++

		if bookmark == "" {
			break
		}

		if maxPages > 0 && pageCount >= maxPages {
			log.Info().Str("board_id", boardID).Int("pages", pageCount).Msg("Reached max pages limit for incremental sync")
			break
		}
	}

	log.Info().Str("board_id", boardID).Int("total_pins", totalPins).Int("pages", pageCount).Msg("Fetched pins for board")
}

func produceMessage(ctx context.Context, producer kafka2.Producer, topic, key string, value interface{}, log *logger.Logger) {
	payload, err := json.Marshal(value)
	if err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("topic", topic).Str("function", "produceMessage").Str("stage", "marshal_message").Msg("Failed to marshal message")
		return
	}

	if err := producer.Produce(ctx, topic, []byte(key), payload); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("topic", topic).Str("key", key).Str("function", "produceMessage").Str("stage", "produce_kafka").Msg("Failed to produce message")
	} else {
		log.Debug().Str("topic", topic).Str("key", key).Int("payload_size", len(payload)).Msg("Produced message to Kafka")
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().
		Int("workers", maxWorkers).
		Str("consumer_group", consumerGroup).
		Msg("Starting Pinterest Fetcher service")

	pinterestClient := social.NewPinterestClient()

	mongoClient, mongoRepo := initMongoDB(cfg, log)
	defer mongoClient.Disconnect(context.Background())

	consumer, producer := initKafkaClients(cfg, log)
	defer consumer.Close()
	defer producer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	svc := NewService(pinterestClient, producer, consumer, mongoRepo, log, cfg.DecryptionKey)

	done := make(chan struct{})
	go func() {
		defer close(done)
		svc.Run(ctx)
	}()

	select {
	case <-sigChan:
		log.Info().Msg("Shutdown signal received, stopping service...")
	case <-done:
		log.Info().Msg("Service stopped")
	}

	cancel()
	<-done

	log.Info().Msg("Pinterest Fetcher service stopped")
}

func initMongoDB(cfg *config.Config, log *logger.Logger) (*mongo.Client, mongodb.UnifiedSocialRepository) {
	credential := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}
	clientOpts := options.Client().ApplyURI(cfg.Mongo.URI).SetAuth(credential)

	mongoClient, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}

	db := mongoClient.Database(cfg.Mongo.Database)
	repo := mongodb.NewUnifiedSocialRepository(db, log.Logger)

	log.Info().Msg("MongoDB connected for timestamp updates")
	return mongoClient, repo
}

func initKafkaClients(cfg *config.Config, log *logger.Logger) (kafka2.Consumer, kafka2.Producer) {
	consumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka consumer")
	}

	producer, err := kafka2.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}

	return consumer, producer
}
