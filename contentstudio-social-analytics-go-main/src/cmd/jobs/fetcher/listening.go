package fetcher

import (
	"context"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	applogger "github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/scheduler"
)

// ProcessListeningTopics runs a single recurring listening scheduling pass.
func ProcessListeningTopics(ctx context.Context, db *mongo.Database, producer kafka.Producer, logger zerolog.Logger) {
	log := logger.With().Str("module", "listening").Logger()
	appLog := &applogger.Logger{Logger: log}
	repo := mongodb.NewListeningRepository(db, appLog)
	workspaceRepo := mongodb.NewListeningWorkspaceRepository(db, appLog)

	sched := scheduler.NewRecurringScheduler(repo, producer, nil, appLog, 0).
		WithSuperAdminResolver(workspaceRepo).
		WithOwnerQuotaChecker(workspaceRepo)
	if _, err := sched.RunOnce(ctx); err != nil {
		log.Error().Err(err).Msg("Recurring listening scheduling pass failed")
	}
}
