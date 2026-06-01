package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api"
	listeningAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/listening"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/middleware"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mentionsSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/mentions"
	viewsSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/views"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Listening API Server")

	mongoClient, err := connectMongo(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Error().Err(err).Msg("MongoDB disconnect error")
		}
	}()

	chClient, err := clickhouse.NewClient(cfg.ClickHouse, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to ClickHouse")
	}
	defer chClient.Close()

	producer, err := kafka.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()

	db := mongoClient.Database(cfg.Mongo.Database)
	listeningRepo := mongodb.NewListeningRepository(db, log)
	listeningViewsRepo := mongodb.NewListeningViewsRepository(db, log)
	listeningWorkspaceRepo := mongodb.NewListeningWorkspaceRepository(db, log)

	workDispatch := &api.APIServer{
		ListeningRepo:          listeningRepo,
		ListeningWorkspaceRepo: listeningWorkspaceRepo,
		Producer:               producer,
		Logger:                 log,
		Config:                 cfg,
	}

	listeningReadRepo := clickhouse.NewListeningReadRepository(chClient, log.Logger)
	mentionsService := mentionsSvc.NewService(listeningReadRepo, log.Logger)
	filterResolver := listeningAPI.NewMentionFilterResolver(listeningViewsRepo, listeningRepo)
	mentionsH := listeningAPI.NewMentionsHandler(mentionsService, filterResolver, log.Logger)
	analyticsH := listeningAPI.NewAnalyticsHandler(mentionsService, filterResolver, log.Logger)
	viewsService := viewsSvc.NewService(listeningViewsRepo, listeningReadRepo, listeningRepo, log.Logger)
	viewsH := listeningAPI.NewViewsHandler(viewsService, log.Logger)

	var jwtMiddleware *middleware.JWTMiddleware
	if cfg.JWT.Secret != "" {
		jwtMiddleware = middleware.NewJWTMiddleware(&cfg.JWT, log)
	} else {
		log.Warn().Msg("JWT secret not configured — authentication disabled")
	}
	apiKeyRepo := mongodb.NewApiKeyRepository(db, log)
	shareableLinkRepo := mongodb.NewShareableLinkRepository(db, log)
	authMiddleware := middleware.NewAuthMiddleware(jwtMiddleware, apiKeyRepo, shareableLinkRepo, log)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", workDispatch.HandleHealth)
	mux.HandleFunc("/api/v1/listening-work", workDispatch.HandleListeningWork)
	listeningAPI.RegisterRoutes(mux, mentionsH, viewsH, analyticsH)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8081"
		log.Info().Str("port", port).Msg("API_PORT not set, using default")
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      loggingMiddleware(log, corsMiddleware(methodOverrideMiddleware(authMiddleware.Authenticate(mux)))),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("Listening API Server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	<-sigChan
	log.Info().Msg("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server shutdown error")
		return
	}
	log.Info().Msg("Listening API Server stopped")
}

func connectMongo(cfg *config.Config) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(cfg.Mongo.URI)
	if cfg.Mongo.Username != "" {
		opts.SetAuth(options.Credential{
			Username:   cfg.Mongo.Username,
			Password:   cfg.Mongo.Password,
			AuthSource: cfg.Mongo.Database,
		})
	}
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("MongoDB ping failed: %w", err)
	}
	return client, nil
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
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
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Int("status", rw.status).
			Dur("duration", time.Since(start)).
			Msg("HTTP request")
	})
}

var validOverrideMethods = map[string]bool{
	http.MethodGet:    true,
	http.MethodPut:    true,
	http.MethodPatch:  true,
	http.MethodDelete: true,
}

func methodOverrideMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if override := r.Header.Get("X-Http-Method-Override"); override != "" {
				upper := strings.ToUpper(override)
				if validOverrideMethods[upper] {
					r.Method = upper
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
