package main

import (
	"context"
	"flag"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	facebookrefresher "github.com/d4interactive/contentstudio-social-analytics-go/src/cmd/jobs/url-refresher/facebook"
	facebookcompetitorrefresher "github.com/d4interactive/contentstudio-social-analytics-go/src/cmd/jobs/url-refresher/facebook_competitor"
	instagramrefresher "github.com/d4interactive/contentstudio-social-analytics-go/src/cmd/jobs/url-refresher/instagram"
	instagramcompetitorrefresher "github.com/d4interactive/contentstudio-social-analytics-go/src/cmd/jobs/url-refresher/instagram_competitor"
	linkedinrefresher "github.com/d4interactive/contentstudio-social-analytics-go/src/cmd/jobs/url-refresher/linkedin"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
)

func main() {
	startApp := time.Now()

	cfg := mustLoadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := initLogger(cfg)
	log.Info().Msg("Starting URL refresher job")

	platform := flag.String("platform", "facebook", "Platform to refresh: facebook, instagram, linkedin, facebook-competitor, instagram-competitor, competitors, all")
	accountType := flag.String("accountType", "", "Type of account to fetch (e.g., page, group)")
	flag.Parse()

	db, mongoCleanup := mustConnectMongo(ctx, cfg, *log)
	defer mongoCleanup()

	zClick := log.With().Str("component", "clickhouse").Logger()
	clickLog := &zClick
	chSink := conversions.NewClickHouseSink(clickLog, cfg)
	clickLog.Info().Msg("ClickHouse sink initialized")

	rm := buildRateManager(cfg, *log)
	fbClient := social.NewFacebookClientWithRates(cfg.Facebook.AppSecret, rm)
	igClientFB := social.NewInstagramClientWithRates(cfg.Facebook.AppSecret, rm)
	igClientIG := social.NewInstagramClientWithRates(cfg.Facebook.AppSecret, rm).WithBaseURL("https://graph.instagram.com/")
	liClient := social.NewLinkedInClient()

	zRepo := log.With().Str("component", "repository").Str("repository", "unified_social").Logger()
	repo := mongodb.NewUnifiedSocialRepository(db, zRepo)
	zCompRepo := log.With().Str("component", "repository").Str("repository", "competitor").Logger()
	competitorRepo := mongodb.NewCompetitorRepository(db, log)
	_ = zCompRepo
	redisClient := mustConnectRedis(cfg)
	defer redisClient.Close()

	switch strings.ToLower(strings.TrimSpace(*platform)) {
	case "", "facebook":
		facebookrefresher.Run(ctx, cfg, *log, repo, fbClient, chSink, 30, *accountType)
	case "instagram":
		instagramrefresher.Run(ctx, cfg, *log, repo, igClientFB, igClientIG, chSink, 30, *accountType)
	case "linkedin":
		linkedinrefresher.Run(ctx, cfg, *log, repo, liClient, chSink, 15, *accountType)
	case "facebook-competitor":
		facebookcompetitorrefresher.Run(ctx, cfg, *log, competitorRepo, redisClient, fbClient, chSink, 20)
	case "instagram-competitor":
		instagramcompetitorrefresher.Run(ctx, cfg, *log, competitorRepo, redisClient, igClientFB, chSink, 20)
	case "competitors":
		facebookcompetitorrefresher.Run(ctx, cfg, *log, competitorRepo, redisClient, fbClient, chSink, 20)
		instagramcompetitorrefresher.Run(ctx, cfg, *log, competitorRepo, redisClient, igClientFB, chSink, 20)
	case "all":
		facebookrefresher.Run(ctx, cfg, *log, repo, fbClient, chSink, 30, *accountType)
		instagramrefresher.Run(ctx, cfg, *log, repo, igClientFB, igClientIG, chSink, 30, *accountType)
		linkedinrefresher.Run(ctx, cfg, *log, repo, liClient, chSink, 15, *accountType)
		facebookcompetitorrefresher.Run(ctx, cfg, *log, competitorRepo, redisClient, fbClient, chSink, 20)
		instagramcompetitorrefresher.Run(ctx, cfg, *log, competitorRepo, redisClient, igClientFB, chSink, 20)
	default:
		log.Fatal().Str("platform", *platform).Msg("Unsupported platform; expected facebook, instagram, linkedin, facebook-competitor, instagram-competitor, competitors, or all")
	}

	log.Info().
		Dur("uptime", time.Since(startApp)).
		Msg("URL refresher job finished")
}
