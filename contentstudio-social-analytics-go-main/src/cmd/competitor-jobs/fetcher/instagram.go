package fetcher

import (
	"context"
	"encoding/json"

	"go.mongodb.org/mongo-driver/mongo"

	repository "github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkaModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

//
// -------------------- Constants --------------------
//

const (
	// Kafka topic for Instagram batch work orders
	InstagramBatchTopic = "competitor-work-order-instagram-batch"

	// Platform identifier
	PlatformInstagram = "instagram"
)

//
// -------------------- Functions --------------------
//

// ProcessInstagramAccounts fetches competitor accounts for Instagram and produces Kafka work orders.
// syncType determines incremental vs full_refresh processing.
func ProcessInstagramAccounts(
	ctx context.Context,
	db *mongo.Database,
	producer kafka.Producer,
	log *logger.Logger,
	syncType string,
) {

	op := log.
		Operation("process_instagram_accounts").
		WithSentryTags(map[string]string{
			"platform":  PlatformInstagram,
			"sync_type": syncType,
		})

	defer func() {
		op.Complete(nil, "")
	}()

	// --------------------
	// Initialize repository
	// --------------------
	repo := repository.NewCompetitorRepository(db, log)

	// --------------------
	// Fetch all competitor accounts
	// --------------------
	competitors, err := repo.GetAccounts(ctx, PlatformInstagram)
	if err != nil {
		op.Complete(err, "fetch_accounts_failed")
		log.Error().
			Err(err).
			Str("platform", PlatformInstagram).
			Msg("Failed to fetch Instagram competitor accounts")
		return
	}

	log.Info().
		Int("count", len(competitors)).
		Str("sync_type", syncType).
		Msg("Fetched Instagram competitor accounts, producing work orders")

	// --------------------
	// Produce Kafka work orders
	// --------------------
	for _, comp := range competitors {
		compOp := log.
			Operation("produce_instagram_work_order").
			WithSentryTags(map[string]string{
				"platform": PlatformInstagram,
				"comp_id":  comp.ID.Hex(),
			})

		workOrder := kafkaModels.CompetitorWorkOrder{
			PageID:   comp.ID.Hex(),
			Channel:  PlatformInstagram,
			ReportID: "",
			Mode:     syncType,
		}

		payload, err := json.Marshal(workOrder)
		if err != nil {
			compOp.Complete(err, "marshal_failed")
			log.Error().
				Err(err).
				Str("comp_id", comp.ID.Hex()).
				Msg("Failed to marshal Kafka work order")
			continue
		}

		// Use competitor ID as key to maintain ordering per competitor
		err = producer.Produce(ctx, InstagramBatchTopic, []byte(comp.ID.Hex()), payload)
		if err != nil {
			compOp.Complete(err, "produce_failed")
			log.Error().
				Err(err).
				Str("comp_id", comp.ID.Hex()).
				Str("topic", InstagramBatchTopic).
				Msg("Failed to produce Kafka work order")
			continue
		}

		compOp.Complete(nil, "")
		log.Debug().
			Str("comp_id", comp.ID.Hex()).
			Str("topic", InstagramBatchTopic).
			Msg("Produced Kafka work order successfully")
	}

	log.Info().
		Int("total_produced", len(competitors)).
		Msg("Completed producing Instagram work orders")
}
