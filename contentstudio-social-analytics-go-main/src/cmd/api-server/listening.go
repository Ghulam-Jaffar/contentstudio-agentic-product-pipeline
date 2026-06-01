package main

import (
	"github.com/rs/zerolog"

	listeningAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/listening"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	mentionsSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/mentions"
	viewsSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/views"
)

func setupListeningAPI(
	chClient *clickhouse.Client,
	listeningViewsRepo *mongodb.ListeningViewsRepository,
	listeningRepo *mongodb.ListeningRepository,
	log zerolog.Logger,
) (*listeningAPI.MentionsHandler, *listeningAPI.AnalyticsHandler, *listeningAPI.ViewsHandler) {
	listeningReadRepo := clickhouse.NewListeningReadRepository(chClient, log)
	mentionsService := mentionsSvc.NewService(listeningReadRepo, log)
	mentionFilterResolver := listeningAPI.NewMentionFilterResolver(listeningViewsRepo, listeningRepo)
	mentionsH := listeningAPI.NewMentionsHandler(mentionsService, mentionFilterResolver, log)
	analyticsH := listeningAPI.NewAnalyticsHandler(mentionsService, mentionFilterResolver, log)
	viewsService := viewsSvc.NewService(listeningViewsRepo, listeningReadRepo, listeningRepo, log)
	viewsH := listeningAPI.NewViewsHandler(viewsService, log)
	return mentionsH, analyticsH, viewsH
}
