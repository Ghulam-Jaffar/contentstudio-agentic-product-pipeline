package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	clickhouseRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	mongoRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	models "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkaModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/notification"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/facebook/facebook-competitor-analysis/service"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/competitortokens"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	PlatformTypeFacebook = "facebook"
	TokenQueueFacebook   = "facebook_valid_token_set"
	WorkersPerPool       = 10
)

// CompetitorJob represents a single competitor work unit
type CompetitorJob struct {
	ReportID  string // MongoDB report ID
	PageID    string // Facebook page ID
	CompID    string
	Mode      models.SyncMode
	StartDate string
	EndDate   string
}

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// Initialize Sentry
	telemetry.ConfigureSentry(cfg)

	// Initialize logging
	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Facebook Competitor Analysis Service")

	rootOp := log.Operation("facebook_competitor_service").
		WithSentryTags(map[string]string{
			"platform": PlatformTypeFacebook,
		})

	rootOp.Start("Service started")

	defer func() {
		rootOp.Complete(nil, "Service shutdown")
		logger.FlushSentry(5 * time.Second)
	}()

	// Context and graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handleShutdown(cancel, log)

	// Initialize external clients
	mongoClient, err := initMongo(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize MongoDB")
	}
	defer mongoClient.Disconnect(ctx)

	chConn, err := initClickHouse(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize ClickHouse")
	}
	defer chConn.Close()

	redisClient := initRedis(cfg)
	defer redisClient.Close()

	// Repositories
	mongoRepo := mongoRepo.NewCompetitorRepository(mongoClient.Database(cfg.Mongo.Database), log)
	chRepo, err := clickhouseRepo.NewClient(cfg.ClickHouse, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize ClickHouse client")
	}

	// Kafka consumers
	consumerRealtime, err := kafka2.NewConsumer(
		cfg.Kafka,
		"facebook-competitor-realtime",
		log.Logger,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Kafka realtime consumer")
	}
	defer consumerRealtime.Close()

	consumerBatch, err := kafka2.NewConsumer(
		cfg.Kafka,
		"facebook-competitor-batch",
		log.Logger,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Kafka batch consumer")
	}
	defer consumerBatch.Close()

	// Channels for realtime pipeline
	realtimeJobs := make(chan CompetitorJob, 100)
	realtimeFetchResults := make(chan *service.FetchResult, 100)
	realtimeParseResults := make(chan *service.ParseResult, 100)

	// Channels for batch pipeline
	batchJobs := make(chan CompetitorJob, 500)
	batchFetchResults := make(chan *service.FetchResult, 500)
	batchParseResults := make(chan *service.ParseResult, 500)

	var wg sync.WaitGroup

	perTokenRPS := cfg.Facebook.PerTokenRPS
	if perTokenRPS <= 0 {
		perTokenRPS = 4.0
	}
	perTokenBurst := cfg.Facebook.PerTokenBurst
	if perTokenBurst <= 0 {
		perTokenBurst = 4
	}
	globalRPS := cfg.Facebook.GlobalRPS
	if globalRPS <= 0 {
		globalRPS = 12.0
	}
	globalBurst := cfg.Facebook.GlobalBurst
	if globalBurst <= 0 {
		globalBurst = 12
	}
	// Optional: cap how many concurrent work orders for the same FB page (1-2 recommended)
	perAccountConcurrency := cfg.Facebook.PerAccountConcurrency
	if perAccountConcurrency <= 0 {
		perAccountConcurrency = 1
	}

	rateMgr := social.NewRateManager(social.RateLimits{
		PerTokenRPS:   perTokenRPS,
		PerTokenBurst: perTokenBurst,
		GlobalRPS:     globalRPS,
		GlobalBurst:   globalBurst,
	})

	emailClient := notification.NewService(cfg.Email, log.Logger, cfg.Email.BackendURL)
	pusherClient := notification.NewPusherClient(cfg.Pusher, log.Logger)

	// Start realtime worker pools (3 stages: fetch, parse, store)
	startFetchPool(ctx, &wg, "realtime-fetch", realtimeJobs, realtimeFetchResults, cfg, mongoRepo, redisClient, rateMgr, log)
	startParsePool(ctx, &wg, "realtime-parse", realtimeFetchResults, realtimeParseResults, cfg, rateMgr, log)
	startStorePool(ctx, &wg, "realtime-store", realtimeParseResults, mongoRepo, chRepo, log, emailClient, pusherClient, true)

	// Start batch worker pools (3 stages: fetch, parse, store)
	startFetchPool(ctx, &wg, "batch-fetch", batchJobs, batchFetchResults, cfg, mongoRepo, redisClient, rateMgr, log)
	startParsePool(ctx, &wg, "batch-parse", batchFetchResults, batchParseResults, cfg, rateMgr, log)
	startStorePool(ctx, &wg, "batch-store", batchParseResults, mongoRepo, chRepo, log, nil, nil, false)

	// Start Kafka consumers
	go consumeRealtime(ctx, consumerRealtime, realtimeJobs, log)
	go consumeBatch(ctx, consumerBatch, batchJobs, log)

	// Wait for shutdown
	<-ctx.Done()

	// Close job channels first
	close(realtimeJobs)
	close(batchJobs)

	// Wait for fetch pools to drain
	wg.Wait()

	// Close intermediate channels
	close(realtimeFetchResults)
	close(batchFetchResults)
	close(realtimeParseResults)
	close(batchParseResults)

	log.Info().Msg("Facebook competitor service stopped")
	logger.FlushSentry(5 * time.Second)
}

// startFetchPool launches workers for the fetch stage
func startFetchPool(
	ctx context.Context,
	wg *sync.WaitGroup,
	poolName string,
	jobs <-chan CompetitorJob,
	fetchResults chan<- *service.FetchResult,
	cfg *config.Config,
	mongoRepo *mongoRepo.CompetitorRepository,
	redis *redis.Client,
	rateMgr *social.RateManager,
	log *logger.Logger,
) {
	for i := 0; i < WorkersPerPool; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					log.Error().
						Interface("panic", r).
						Str("pool", poolName).
						Int("worker", workerID).
						Str("function", "startFetchPool").
						Str("stage", "panic_recovery").
						Msg("Fetch worker panicked")
				}
			}()

			svc := service.NewCompetitorAnalysisService(
				social.NewFacebookClientWithRates(cfg.Facebook.AppSecret, rateMgr),
				mongoRepo,
				nil, // ClickHouse not needed for fetch stage
				log,
			)

			log.Info().Str("pool", poolName).Int("worker", workerID).Msg("Fetch worker started")

			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobs:
					if !ok {
						return
					}

					payload := buildPayload(ctx, job, mongoRepo, redis, cfg, log)
					if payload == nil {
						continue
					}

					result := fetchWithTokenRetry(ctx, svc, payload, redis, cfg, log)
					// Skip sending to parse stage if error is expected (permissions/auth)
					if result.IsExpectedError() {
						log.Info().
							Str("page_id", payload.PageID).
							Str("page_name", payload.PageName).
							Msg("Skipping parse stage due to expected API error")
						continue
					}
					select {
					case fetchResults <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		}(i)
	}
}

// startParsePool launches workers for the parse stage
func startParsePool(
	ctx context.Context,
	wg *sync.WaitGroup,
	poolName string,
	fetchResults <-chan *service.FetchResult,
	parseResults chan<- *service.ParseResult,
	cfg *config.Config,
	rateMgr *social.RateManager,
	log *logger.Logger,
) {
	for i := 0; i < WorkersPerPool; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					log.Error().
						Interface("panic", r).
						Str("pool", poolName).
						Int("worker", workerID).
						Str("function", "startParsePool").
						Str("stage", "panic_recovery").
						Msg("Parse worker panicked")
				}
			}()

			svc := service.NewCompetitorAnalysisService(
				social.NewFacebookClientWithRates(cfg.Facebook.AppSecret, rateMgr),
				nil, // MongoDB not needed for parse stage
				nil, // ClickHouse not needed for parse stage
				log,
			)

			log.Info().Str("pool", poolName).Int("worker", workerID).Msg("Parse worker started")

			for {
				select {
				case <-ctx.Done():
					return
				case fetchResult, ok := <-fetchResults:
					if !ok {
						return
					}

					result := svc.ParseCompetitorData(ctx, fetchResult)

					select {
					case parseResults <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		}(i)
	}
}

// startStorePool launches workers for the store stage
func startStorePool(
	ctx context.Context,
	wg *sync.WaitGroup,
	poolName string,
	parseResults <-chan *service.ParseResult,
	mongoRepo *mongoRepo.CompetitorRepository,
	chRepo *clickhouseRepo.Client,
	log *logger.Logger,
	notifier *notification.Service,
	pusherClient *notification.PusherClient,
	isRealtime bool,
) {
	for i := 0; i < WorkersPerPool; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					log.Error().
						Interface("panic", r).
						Str("pool", poolName).
						Int("worker", workerID).
						Str("function", "startStorePool").
						Str("stage", "panic_recovery").
						Msg("Store worker panicked")
				}
			}()

			svc := service.NewCompetitorAnalysisService(
				nil, // Facebook client not needed for store stage
				mongoRepo,
				chRepo,
				log,
			)

			log.Info().Str("pool", poolName).Int("worker", workerID).Msg("Store worker started")

			for {
				select {
				case <-ctx.Done():
					return
				case parseResult, ok := <-parseResults:
					if !ok {
						return
					}

					result := svc.StoreCompetitorData(ctx, parseResult)

					if result.Error != nil {
						log.Error().
							Err(result.Error).
							Str("error_message", result.Error.Error()).
							Str("page_id", result.PageID).
							Str("page_name", result.PageName).
							Str("function", "startStorePool").
							Str("stage", "store_competitor_data").
							Msg("Store stage failed")
					} else {
						log.Info().
							Str("page_id", result.PageID).
							Str("page_name", result.PageName).
							Int("total_processed", result.TotalProcessed).
							Msg("Store stage completed successfully")

						// Send notifications for every successful competitor sync.
						if notifier != nil || pusherClient != nil {
							sendCompetitorNotifications(ctx, parseResult, mongoRepo, notifier, pusherClient, log)
						}
					}
				}
			}
		}(i)
	}
}

// buildPayload constructs a CompetitorPayload from a job
func buildPayload(
	ctx context.Context,
	job CompetitorJob,
	repo *mongoRepo.CompetitorRepository,
	redis *redis.Client,
	cfg *config.Config,
	log *logger.Logger,
) *models.FacebookCompetitorPayload {
	comp, err := repo.GetByID(ctx, job.CompID)
	if err != nil {
		// Competitor not found is expected - they may have been deleted or archived
		log.Warn().Err(err).Str("competitor_id", job.CompID).Msg("Competitor not found")
		return nil
	}

	token, err := getCompetitorToken(ctx, redis, cfg)
	if err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("competitor_id", job.CompID).Str("function", "buildPayload").Str("stage", "token_fetch").Msg("Token fetch failed")
		return nil
	}

	return &models.FacebookCompetitorPayload{
		ReportID:    job.ReportID,
		PageID:      toString(comp.CompetitorID, log),
		PageName:    comp.Name,
		SyncStatus:  job.Mode,
		AccessToken: token,
		StartDate:   job.StartDate,
		EndDate:     job.EndDate,
	}
}

// sendCompetitorNotifications sends Pusher notifications for realtime competitor analysis
func sendCompetitorNotifications(
	ctx context.Context,
	parseResult *service.ParseResult,
	mongoRepo *mongoRepo.CompetitorRepository,
	notifier *notification.Service,
	pusherClient *notification.PusherClient,
	log *logger.Logger,
) {
	if parseResult == nil || parseResult.Payload == nil || parseResult.ReportID == "" {
		return
	}

	pageID := parseResult.Payload.PageID
	reportID := parseResult.ReportID
	log.Info().Str("report_id", reportID).Str("page_id", pageID).Msg("Sending competitor notifications")

	// Get competitor by page ID
	competitors, err := mongoRepo.GetByCompetitorID(ctx, pageID)
	if err != nil || len(competitors) == 0 {
		log.Error().Err(err).Str("page_id", pageID).Str("function", "sendCompetitorNotifications").Str("stage", "fetch_competitor").Msg("Failed to fetch competitor for notifications")
		return
	}

	competitor := competitors[0]

	// Get the specific report by ID
	report, err := mongoRepo.GetReportByID(ctx, reportID)
	if err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("report_id", reportID).Str("function", "sendCompetitorNotifications").Str("stage", "fetch_report").Msg("Failed to fetch report for notifications")
		return
	}

	log.Info().Str("report_id", reportID).Str("comp_state", parseResult.CurrentState).Msg("Preparing to send notifications")

	// Send Pusher notification
	sendFacebookPusherNotification(pusherClient, competitor, report, pageID, parseResult, log)

	// Send email if competitor was in "Added" or "NotFound" state (use the original state from fetch)
	if (parseResult.CurrentState == "Added" || parseResult.CurrentState == "NotFound") && report != nil {
		sendFacebookEmailNotification(ctx, notifier, mongoRepo, report, competitor, parseResult, log)
	}
}

// sendFacebookPusherNotification sends real-time notification via Pusher
func sendFacebookPusherNotification(
	pusherClient *notification.PusherClient,
	competitor *mongoModels.Competitor,
	report *mongoModels.CompetitorReport,
	pageID string,
	parseResult *service.ParseResult,
	log *logger.Logger,
) {
	if pusherClient == nil || report == nil {
		return
	}

	channel := fmt.Sprintf("fb-competitor-analytics-%s", report.WorkspaceID.Hex())
	event := fmt.Sprintf("fb-competitor-analytics-%s", report.WorkspaceID.Hex())

	data := map[string]interface{}{
		"report_id":    report.ID.Hex(),
		"workspace_id": report.WorkspaceID.Hex(),
		"page_id":      pageID,
		"slug":         parseResult.Payload.PageName,
		"image":        competitor.Image,
		"report_name":  report.Name,
		"display_name": parseResult.Payload.PageName,
		"job_type":     parseResult.Payload.SyncStatus,
		"error":        competitor.Error,
		"state":        competitor.State,
	}

	if err := pusherClient.Trigger(channel, event, data); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("channel", channel).Str("function", "sendFacebookPusherNotification").Str("stage", "pusher_trigger").Msg("Failed to send Pusher notification")
	} else {
		log.Info().Str("channel", channel).Msg("Pusher notification sent successfully")
	}
}

// sendFacebookEmailNotification sends email notification to user via backend
func sendFacebookEmailNotification(
	ctx context.Context,
	notifier *notification.Service,
	mongoRepo *mongoRepo.CompetitorRepository,
	report *mongoModels.CompetitorReport,
	competitor *mongoModels.Competitor,
	parseResult *service.ParseResult,
	log *logger.Logger,
) {
	if notifier == nil || report == nil || mongoRepo == nil {
		return
	}

	// Get user from MongoDB using CreatedByUserID
	user, err := mongoRepo.GetUserByID(ctx, report.CreatedByUserID.Hex())
	if err != nil {
		log.Warn().Err(err).Str("userId", report.CreatedByUserID.Hex()).Msg("Could not fetch user for email notification")
		return
	}

	if user == nil || user.ID == primitive.NilObjectID {
		log.Warn().Str("userId", report.CreatedByUserID.Hex()).Msg("User not found for notification")
		return
	}

	// Send analytics notification to backend API
	accountID := ""
	accountName := ""
	if competitor != nil {
		accountID = competitor.GetCompetitorIDAsString()
		accountName = competitor.Name
	}
	if accountID == "" || accountID == "<nil>" {
		if parseResult != nil && parseResult.Payload != nil {
			accountID = parseResult.Payload.PageID
		}
	}
	if accountName == "" {
		if parseResult != nil && parseResult.Payload != nil {
			accountName = parseResult.Payload.PageName
		}
	}

	err = notifier.SendAnalyticsNotification(
		user.ID.Hex(),            // userID
		report.WorkspaceID.Hex(), // workspaceID
		"facebook",               // platform
		accountID,                // accountID (competitor ID)
		accountName,              // accountName
		true,                     // isCompetitor
	)

	if err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("userId", user.ID.Hex()).Str("competitorId", accountID).Str("function", "sendFacebookEmailNotification").Str("stage", "send_notification").Msg("Failed to send analytics notification to backend")
	} else {
		log.Info().Str("userId", user.ID.Hex()).Str("competitorId", accountID).Msg("Analytics notification sent to backend successfully")
	}
}

// consumeRealtime starts the Kafka consumer for realtime jobs
func consumeRealtime(ctx context.Context, consumer kafka2.Consumer, jobs chan<- CompetitorJob, log *logger.Logger) {
	log.Info().Msg("Starting Facebook realtime consumer")
	consumer.Consume(ctx, []string{"competitor-work-order-facebook"}, func(_ context.Context, _ string, _, value []byte) error {
		op := log.Operation("ConsumeRealtimeMessage")
		defer op.Complete(nil, "")

		var wo kafkaModels.CompetitorWorkOrder
		if err := json.Unmarshal(value, &wo); err != nil {
			log.Warn().Err(err).Str("raw_value", string(value)).Msg("Failed to unmarshal Kafka message")
			return nil
		}
		if wo.Channel != "facebook" {
			return nil
		}

		op.WithSentryTags(map[string]string{
			"report_id": wo.ReportID,
			"page_id":   wo.PageID,
			"mode":      wo.Mode,
		})

		reportID := wo.ReportID
		pageID := wo.PageID

		log.Info().Str("report_id", reportID).Str("page_id", pageID).Str("mode", wo.Mode).Msg("Kafka message received")

		select {
		case jobs <- CompetitorJob{ReportID: reportID, PageID: pageID, CompID: pageID, Mode: models.SyncMode(wo.Mode), StartDate: wo.StartDate, EndDate: wo.EndDate}:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	})
}

// consumeBatch starts the Kafka consumer for batch jobs
func consumeBatch(ctx context.Context, consumer kafka2.Consumer, jobs chan<- CompetitorJob, log *logger.Logger) {
	log.Info().Msg("Starting Facebook batch consumer")
	consumer.Consume(ctx, []string{"competitor-work-order-facebook-batch"}, func(_ context.Context, _ string, _, value []byte) error {
		op := log.Operation("ConsumeBatchMessage")
		defer op.Complete(nil, "")

		var wo kafkaModels.CompetitorWorkOrder
		if err := json.Unmarshal(value, &wo); err != nil {
			log.Warn().Err(err).Str("raw_value", string(value)).Msg("Failed to unmarshal Kafka message")
			return nil
		}
		if wo.Channel != "facebook" {
			return nil
		}

		op.WithSentryTags(map[string]string{
			"report_id": wo.ReportID,
			"page_id":   wo.PageID,
			"mode":      wo.Mode,
		})

		reportID := wo.ReportID
		pageID := wo.PageID

		select {
		case jobs <- CompetitorJob{ReportID: reportID, PageID: pageID, CompID: pageID, Mode: models.SyncMode(wo.Mode), StartDate: wo.StartDate, EndDate: wo.EndDate}:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	})
}

// ------------------------------
// Initialization
// ------------------------------

// initMongo initializes MongoDB connection
func initMongo(ctx context.Context, cfg *config.Config) (*mongo.Client, error) {
	credential := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}

	clientOpts := options.Client().
		ApplyURI(cfg.Mongo.URI).
		SetAuth(credential)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

// initClickHouse initializes ClickHouse connection
func initClickHouse(cfg *config.Config) (clickhouse.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.ClickHouse.Host, cfg.ClickHouse.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouse.Database,
			Username: cfg.ClickHouse.Username,
			Password: cfg.ClickHouse.Password,
		},
	})
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}
	return conn, nil
}

func initRedis(cfg *config.Config) *redis.Client {
	maxRetries := cfg.Redis.MaxRetries
	if maxRetries < 3 {
		maxRetries = 3
	}

	poolSize := cfg.Redis.PoolSize
	if poolSize < 50 {
		poolSize = 50
	}

	return redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		MaxRetries:   maxRetries,
		PoolSize:     poolSize,
		DialTimeout:  20 * time.Second,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		PoolTimeout:  20 * time.Second,
	})
}

// -----------------------------
// Helper functions
// -----------------------------

// getCompetitorToken retrieves and decrypts a Facebook token from Redis
func getCompetitorToken(ctx context.Context, redisClient *redis.Client, cfg *config.Config) (string, error) {
	candidate, err := competitortokens.FetchCandidate(ctx, redisClient, TokenQueueFacebook, cfg.DecryptionKey, false, nil)
	if err != nil {
		return "", err
	}
	return candidate.AccessToken, nil
}

func fetchWithTokenRetry(
	ctx context.Context,
	svc *service.CompetitorAnalysisService,
	payload *models.FacebookCompetitorPayload,
	redisClient *redis.Client,
	cfg *config.Config,
	log *logger.Logger,
) *service.FetchResult {
	result := svc.FetchCompetitorData(ctx, payload)
	if result.Error == nil || !competitortokens.IsFacebookTokenIssue(result.Error) {
		return result
	}

	exclude := map[string]struct{}{
		competitortokens.CandidateKey(payload.AccessToken, ""): {},
	}
	candidate, err := competitortokens.FetchCandidate(ctx, redisClient, TokenQueueFacebook, cfg.DecryptionKey, false, exclude)
	if err != nil {
		log.Warn().Err(err).Str("page_id", payload.PageID).Msg("No fresh Facebook competitor token available for retry")
		return result
	}

	retryPayload := *payload
	retryPayload.AccessToken = candidate.AccessToken

	log.Warn().Str("page_id", payload.PageID).Msg("Retrying Facebook competitor fetch with a fresh token")
	return svc.FetchCompetitorData(ctx, &retryPayload)
}

// handleShutdown sets up a signal handler for graceful shutdown
func handleShutdown(cancel context.CancelFunc, log *logger.Logger) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		log.Info().Msg("Shutdown signal received")
		cancel()
	}()
}

// toString converts various ID types to string
func toString(id interface{}, log *logger.Logger) string {
	switch v := id.(type) {
	case string:
		return v
	case int, int32, int64:
		return fmt.Sprintf("%v", v)
	default:
		log.Error().
			Str("type", fmt.Sprintf("%T", id)).
			Str("function", "toString").
			Str("stage", "type_conversion").
			Msg("Unsupported CompetitorID type")
		return ""

	}
}
