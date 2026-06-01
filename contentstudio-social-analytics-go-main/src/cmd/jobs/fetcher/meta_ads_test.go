package fetcher

import (
	"io"
	"testing"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"

	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

func TestMetaAdsConstants(t *testing.T) {
	if metaAdsMongoFetchSize != 50 {
		t.Fatalf("expected metaAdsMongoFetchSize 50, got %d", metaAdsMongoFetchSize)
	}
	if metaAdsKafkaBatchSize != 200 {
		t.Fatalf("expected metaAdsKafkaBatchSize 200, got %d", metaAdsKafkaBatchSize)
	}
	if metaAdsUpdateIntervalHours != 6 {
		t.Fatalf("expected metaAdsUpdateIntervalHours 6, got %d", metaAdsUpdateIntervalHours)
	}
	if topicMetaAdsBatch != "work-order-meta-ads" {
		t.Fatalf("unexpected topic: %s", topicMetaAdsBatch)
	}
}

func TestBuildMetaAdsAccountBatch(t *testing.T) {
	log := zerolog.New(io.Discard)
	workspaceID := primitive.NewObjectID()

	accounts := []mongomodels.SocialIntegration{
		{
			ID:                 primitive.NewObjectID(),
			PlatformIdentifier: "act_12345",
			WorkspaceID:        workspaceID,
			AccessToken:        "access-token",
			ExtraData: map[string]interface{}{
				"access_token": "access-token",
			},
		},
		{
			ID:                 primitive.NewObjectID(),
			PlatformIdentifier: "act_67890",
			WorkspaceID:        workspaceID,
			AccessToken:        "access-token-2",
			ExtraData: map[string]interface{}{
				"access_token": "access-token-2",
			},
		},
	}

	batch, skipped := buildMetaAdsAccountBatch(accounts, "scheduled", log)
	if skipped != 0 {
		t.Fatalf("expected 0 skipped, got %d", skipped)
	}
	if len(batch) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(batch))
	}
	if batch[0].AccountID != "12345" {
		t.Fatalf("expected account id stripped of act_ prefix, got %q", batch[0].AccountID)
	}
	if batch[0].WorkspaceID != workspaceID.Hex() {
		t.Fatalf("expected workspace id %q, got %q", workspaceID.Hex(), batch[0].WorkspaceID)
	}
	if batch[0].SyncType != "scheduled" {
		t.Fatalf("unexpected sync type: %s", batch[0].SyncType)
	}
}

func TestBuildMetaAdsAccountBatch_SkipsInvalid(t *testing.T) {
	log := zerolog.New(io.Discard)
	accounts := []mongomodels.SocialIntegration{
		{ID: primitive.NewObjectID(), PlatformIdentifier: "act_1"},
	}

	batch, skipped := buildMetaAdsAccountBatch(accounts, "scheduled", log)
	if len(batch) != 0 {
		t.Fatalf("expected empty batch, got %d", len(batch))
	}
	if skipped != 1 {
		t.Fatalf("expected 1 skipped, got %d", skipped)
	}
}
