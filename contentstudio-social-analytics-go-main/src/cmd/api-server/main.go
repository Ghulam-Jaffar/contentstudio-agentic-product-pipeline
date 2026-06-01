package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics"
	campaignLabelHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/campaign_label"
	facebookHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/facebook"
	fbCompetitorHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/fb_competitor"
	gmbHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/gmb"
	igCompetitorHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/ig_competitor"
	instagramHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/instagram"
	linkedinHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/linkedin"
	lookerStudioHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/looker_studio"
	metaAdsHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/meta_ads"
	overviewHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/overview"
	pinterestHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/pinterest"
	tiktokHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/tiktok"
	twitterHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/twitter"
	youtubeHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/youtube"
	listeningAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/listening"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/middleware"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	campaignLabelRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/campaign_label"
	facebookRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/facebook"
	fbCompetitorRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/fb_competitor"
	gmbRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/gmb"
	igCompetitorRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/ig_competitor"
	instagramRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/instagram"
	linkedinRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/linkedin"
	metaAdsRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/meta_ads"
	overviewRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/overview"
	pinterestRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/pinterest"
	tiktokRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/tiktok"
	twitterRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/twitter"
	youtubeRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/youtube"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	goredis "github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/ai"
	campaignLabelService "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/campaign_label"
	facebookService "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/facebook"
	fbCompetitorService "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/fb_competitor"
	gmbService "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/gmb"
	igCompetitorService "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/ig_competitor"
	instagramService "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/instagram"
	linkedinService "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/linkedin"
	metaAdsService "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/meta_ads"
	overviewService "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/overview"
	pinterestService "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/pinterest"
	tiktokService "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/tiktok"
	twitterService "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/twitter"
	youtubeService "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/youtube"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting API Server")

	// MongoDB
	credential := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.Mongo.URI).SetAuth(credential))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoClient.Disconnect(context.Background())

	// Kafka producer
	producer, err := kafka.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()

	// ClickHouse
	chClient, err := clickhouse.NewClient(cfg.ClickHouse, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to ClickHouse")
	}
	defer chClient.Close()

	// Listening repository (decoupled from analytics)
	listeningRepo := mongodb.NewListeningRepository(mongoClient.Database(cfg.Mongo.Database), log)
	listeningViewsRepo := mongodb.NewListeningViewsRepository(mongoClient.Database(cfg.Mongo.Database), log)
	listeningWorkspaceRepo := mongodb.NewListeningWorkspaceRepository(mongoClient.Database(cfg.Mongo.Database), log)

	// Immediate work API server
	server := &api.APIServer{
		MongoClient:            mongoClient,
		UnifiedRepo:            mongodb.NewUnifiedSocialRepository(mongoClient.Database(cfg.Mongo.Database), log.Logger),
		ListeningRepo:          listeningRepo,
		ListeningWorkspaceRepo: listeningWorkspaceRepo,
		Producer:               producer,
		Logger:                 log,
		Config:                 cfg,
	}

	// LinkedIn analytics: repo → service → handler
	liRepo := linkedinRepo.NewRepository(chClient)
	linkedinSvc := linkedinService.NewLinkedInAnalyticsService(liRepo, log.Logger)
	linkedinH := linkedinHandler.NewLinkedInHandler(linkedinSvc, log.Logger)

	// GMB analytics: repo → service → handler
	gRepo := gmbRepo.NewRepository(chClient)
	gmbSvc := gmbService.NewGMBAnalyticsService(gRepo, log.Logger)
	gmbH := gmbHandler.NewGMBHandler(gmbSvc, log.Logger)

	// Facebook analytics: repo → service → handler
	fbRepo := facebookRepo.NewRepository(chClient)
	facebookSvc := facebookService.NewFacebookAnalyticsService(fbRepo, log.Logger)
	facebookH := facebookHandler.NewHandler(facebookSvc, log.Logger)

	// Instagram analytics: repo → service → handler
	igRepo := instagramRepo.NewRepository(chClient)
	instagramSvc := instagramService.NewInstagramAnalyticsService(igRepo, log.Logger)
	instagramH := instagramHandler.NewInstagramHandler(instagramSvc, log.Logger)

	// YouTube analytics: repo → service → handler
	ytRepo := youtubeRepo.NewRepository(chClient)
	youtubeSvc := youtubeService.NewYoutubeAnalyticsService(ytRepo, log.Logger)
	youtubeH := youtubeHandler.NewHandler(youtubeSvc, log.Logger)

	// Pinterest analytics: repo → service → handler
	ptRepo := pinterestRepo.NewRepository(chClient)
	pinterestSvc := pinterestService.NewPinterestAnalyticsService(ptRepo, log.Logger)
	pinterestH := pinterestHandler.NewHandler(pinterestSvc, log.Logger)
	// Twitter analytics: repo → service → handler
	twRepo := twitterRepo.NewRepository(chClient)
	twitterSvc := twitterService.NewTwitterAnalyticsService(twRepo, mongoClient.Database(cfg.Mongo.Database), log.Logger)
	twitterH := twitterHandler.NewHandler(twitterSvc, log.Logger)

	// TikTok analytics: repo → service → handler
	ttRepo := tiktokRepo.NewRepository(chClient)
	tiktokSvc := tiktokService.NewTiktokAnalyticsService(ttRepo, log.Logger)
	tiktokH := tiktokHandler.NewHandler(tiktokSvc, log.Logger)

	// Overview V2 analytics: repo → service → handler
	ovRepo := overviewRepo.NewRepository(chClient)
	overviewSvc := overviewService.NewOverviewAnalyticsService(ovRepo, log.Logger)
	overviewH := overviewHandler.NewHandler(overviewSvc, log.Logger)

	// Campaign & Label analytics: repo → service (with MongoDB) → handler
	clRepo := campaignLabelRepo.NewRepository(chClient)
	campaignLabelSvc := campaignLabelService.NewCampaignLabelAnalyticsService(clRepo, mongoClient.Database(cfg.Mongo.Database), log.Logger)
	campaignLabelH := campaignLabelHandler.NewHandler(campaignLabelSvc, log.Logger)

	mentionsH, analyticsH, viewsH := setupListeningAPI(chClient, listeningViewsRepo, listeningRepo, log.Logger)
	// Competitor analytics: shared MongoDB repo + per-platform ClickHouse repo → service → handler
	competitorRepo := mongodb.NewCompetitorRepository(mongoClient.Database(cfg.Mongo.Database), log)

	fbcRepo := fbCompetitorRepo.NewRepository(chClient)
	fbcSvc := fbCompetitorService.NewFacebookCompetitorService(fbcRepo, competitorRepo, log.Logger)
	fbcH := fbCompetitorHandler.NewHandler(fbcSvc, log.Logger)

	igcRepo := igCompetitorRepo.NewRepository(chClient)
	igcSvc := igCompetitorService.NewInstagramCompetitorService(igcRepo, competitorRepo, log.Logger)
	igcH := igCompetitorHandler.NewHandler(igcSvc, log.Logger)

	maRepo := metaAdsRepo.NewRepository(chClient)
	metaAdsSvc := metaAdsService.NewMetaAdsService(maRepo, log.Logger)
	metaAdsH := metaAdsHandler.NewHandler(metaAdsSvc, log.Logger)

	// AI Insights: Redis cache + AI agent client
	if cfg.AIAgents.BaseURL != "" {
		agentClient := ai.NewAgentClient(&cfg.AIAgents, log.Logger)

		var redisCache goredis.Client
		if cfg.Redis.Addr != "" {
			rc, err := goredis.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, log.Logger)
			if err != nil {
				log.Warn().Err(err).Msg("Redis unavailable — AI insights caching disabled")
			} else {
				redisCache = rc
				defer rc.Close()
			}
		}

		gmbAIInsightsSvc := gmbService.NewAIInsightsService(gmbSvc, agentClient, redisCache)
		facebookAIInsightsSvc := facebookService.NewAIInsightsService(facebookSvc, agentClient, redisCache)
		linkedinAIInsightsSvc := linkedinService.NewAIInsightsService(linkedinSvc, agentClient, redisCache)
		instagramAIInsightsSvc := instagramService.NewAIInsightsService(instagramSvc, agentClient, redisCache)
		youtubeAIInsightsSvc := youtubeService.NewAIInsightsService(youtubeSvc, agentClient, redisCache)
		pinterestAIInsightsSvc := pinterestService.NewAIInsightsService(pinterestSvc, agentClient, redisCache)
		tiktokAIInsightsSvc := tiktokService.NewAIInsightsService(tiktokSvc, agentClient, redisCache)
		overviewAIInsightsSvc := overviewService.NewAIInsightsService(overviewSvc, agentClient, redisCache)
		metaAdsAIInsightsSvc := metaAdsService.NewAIInsightsService(metaAdsSvc, agentClient, redisCache)

		gmbH.SetAIInsightsService(gmbAIInsightsSvc)
		facebookH.SetAIInsightsService(facebookAIInsightsSvc)
		linkedinH.SetAIInsightsService(linkedinAIInsightsSvc)
		instagramH.SetAIInsightsService(instagramAIInsightsSvc)
		tiktokH.SetAIInsightsService(tiktokAIInsightsSvc)
		youtubeH.SetAIInsightsService(youtubeAIInsightsSvc)
		pinterestH.SetAIInsightsService(pinterestAIInsightsSvc)
		overviewH.SetAIInsightsService(overviewAIInsightsSvc)
		metaAdsSvc.SetAIInsightsService(metaAdsAIInsightsSvc)
		log.Info().Msg("Analytics AI Insights enabled for GMB, Facebook, LinkedIn, Instagram, YouTube, Pinterest, TikTok, Overview, and Meta Ads")
	} else {
		log.Warn().Msg("AI_AGENTS_BASE_URL not set — analytics AI insights disabled")
	}

	// Authentication
	var jwtMiddleware *middleware.JWTMiddleware
	if cfg.JWT.Secret != "" {
		jwtMiddleware = middleware.NewJWTMiddleware(&cfg.JWT, log)
		log.Info().Msg("JWT authentication enabled")
		if cfg.JWT.AdminSecret != "" {
			log.Info().Msg("JWT admin secret configured (JWT_ADMIN_SECRET_KEY)")
		}
	}
	apiKeyRepo := mongodb.NewApiKeyRepository(mongoClient.Database(cfg.Mongo.Database), log)
	shareableLinkRepo := mongodb.NewShareableLinkRepository(mongoClient.Database(cfg.Mongo.Database), log)
	log.Info().Msg("API key authentication enabled (MongoDB-backed)")
	authMiddleware := middleware.NewAuthMiddleware(jwtMiddleware, apiKeyRepo, shareableLinkRepo, log)

	// Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.HandleHealth)
	mux.HandleFunc("/api/v1/immediate-work", server.HandleImmediateWork)
	mux.HandleFunc("/api/v1/competitor-work", server.HandleCompetitorWork)
	mux.HandleFunc("/api/v1/listening-work", server.HandleListeningWork)
	listeningAPI.RegisterRoutes(mux, mentionsH, viewsH, analyticsH)
	lookerStudioH := lookerStudioHandler.NewHandler(cfg.LookerStudio, apiKeyRepo, log.Logger)
	analytics.RegisterRoutes(mux, linkedinH, gmbH, facebookH, twitterH, tiktokH, instagramH, youtubeH, pinterestH, overviewH, campaignLabelH, lookerStudioH, fbcH, igcH, metaAdsH)

	appHandler := mux

	// Port
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	server.HttpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      loggingMiddleware(log, corsMiddleware(methodOverrideMiddleware(authMiddleware.Authenticate(appHandler)))),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("addr", server.HttpServer.Addr).Msg("API Server listening")
		if err := server.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	<-sigChan
	log.Info().Msg("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.HttpServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server shutdown error")
	}

	log.Info().Msg("API Server stopped")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-API-Key, X-Shareable-ID, X-LOCALE, X-FRONTEND-ORIGIN, baggage, sentry-trace, X-Http-Method-Override")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(log *logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Dur("duration", time.Since(start)).
			Msg("HTTP request")
	})
}

func methodOverrideMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			override := r.Header.Get("X-Http-Method-Override")
			if override != "" {
				r.Method = strings.ToUpper(override)
			}
		}
		next.ServeHTTP(w, r)
	})
}
