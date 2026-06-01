package fetcher

import (
	"testing"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"

	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Constants Tests ==================

func TestPinterestConstants(t *testing.T) {
	if pinterestMongoFetchSize != 50 {
		t.Errorf("pinterestMongoFetchSize = %d, want 50", pinterestMongoFetchSize)
	}
	if pinterestKafkaBatchSize != 200 {
		t.Errorf("pinterestKafkaBatchSize = %d, want 200", pinterestKafkaBatchSize)
	}
	if pinterestUpdateIntervalHours != 6 {
		t.Errorf("pinterestUpdateIntervalHours = %d, want 6", pinterestUpdateIntervalHours)
	}
	if topicPinterestBatch != "work-order-pinterest" {
		t.Errorf("topicPinterestBatch = %q, want %q", topicPinterestBatch, "work-order-pinterest")
	}
}

// ================== buildPinterestAccountBatch Tests ==================

func TestBuildPinterestAccountBatch_ValidAccounts(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestPinterestAccount("acc1", "pinterest123", kafkamodels.PinterestAccountTypeProfile, "", true, true),
		createTestPinterestAccount("acc2", "pinterest456", kafkamodels.PinterestAccountTypeProfile, "", true, true),
	}

	batch, skipped := buildPinterestAccountBatch(accounts, "incremental", log)

	if len(batch) != 2 {
		t.Errorf("expected 2 accounts in batch, got %d", len(batch))
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}

	// Verify first account
	if batch[0].AccountID != "pinterest123" {
		t.Errorf("expected AccountID 'pinterest123', got %q", batch[0].AccountID)
	}
	if batch[0].SyncType != "incremental" {
		t.Errorf("expected SyncType 'incremental', got %q", batch[0].SyncType)
	}
	if batch[0].AccountType != kafkamodels.PinterestAccountTypeProfile {
		t.Errorf("expected AccountType 'profile', got %q", batch[0].AccountType)
	}
}

func TestBuildPinterestAccountBatch_BoardAccount(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestPinterestAccount("acc1", "pinterest123", kafkamodels.PinterestAccountTypeBoard, "board123", true, true),
	}

	batch, skipped := buildPinterestAccountBatch(accounts, "incremental", log)

	if len(batch) != 1 {
		t.Errorf("expected 1 account in batch, got %d", len(batch))
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}

	if batch[0].AccountType != kafkamodels.PinterestAccountTypeBoard {
		t.Errorf("expected AccountType 'board', got %q", batch[0].AccountType)
	}
	if batch[0].BoardID != "board123" {
		t.Errorf("expected BoardID 'board123', got %q", batch[0].BoardID)
	}
}

func TestBuildPinterestAccountBatch_BoardAccountMissingBoardID(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestPinterestAccount("acc1", "pinterest123", kafkamodels.PinterestAccountTypeBoard, "", true, true), // board account without board_id
	}

	batch, skipped := buildPinterestAccountBatch(accounts, "full_sync", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts in batch, got %d", len(batch))
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestBuildPinterestAccountBatch_MissingAccessToken(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestPinterestAccount("acc1", "pinterest123", kafkamodels.PinterestAccountTypeProfile, "", false, true), // no access token
	}

	batch, skipped := buildPinterestAccountBatch(accounts, "full_sync", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts in batch, got %d", len(batch))
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestBuildPinterestAccountBatch_MissingWorkspaceID(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestPinterestAccount("acc1", "pinterest123", kafkamodels.PinterestAccountTypeProfile, "", true, false), // no workspace
	}

	batch, skipped := buildPinterestAccountBatch(accounts, "full_sync", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts in batch, got %d", len(batch))
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestBuildPinterestAccountBatch_MixedValidity(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestPinterestAccount("acc1", "pinterest1", kafkamodels.PinterestAccountTypeProfile, "", true, true),             // valid profile
		createTestPinterestAccount("acc2", "pinterest2", kafkamodels.PinterestAccountTypeProfile, "", false, true),            // missing access token
		createTestPinterestAccount("acc3", "pinterest3", kafkamodels.PinterestAccountTypeBoard, "board1", true, true),         // valid board
		createTestPinterestAccount("acc4", "pinterest4", kafkamodels.PinterestAccountTypeBoard, "", true, true),               // board missing board_id
		createTestPinterestAccount("acc5", "pinterest5", kafkamodels.PinterestAccountTypeProfile, "", true, false),            // missing workspace
	}

	batch, skipped := buildPinterestAccountBatch(accounts, "incremental", log)

	if len(batch) != 2 {
		t.Errorf("expected 2 valid accounts in batch, got %d", len(batch))
	}
	if skipped != 3 {
		t.Errorf("expected 3 skipped, got %d", skipped)
	}
}

func TestBuildPinterestAccountBatch_EmptyInput(t *testing.T) {
	log := zerolog.Nop()

	batch, skipped := buildPinterestAccountBatch([]mongomodels.SocialIntegration{}, "incremental", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts in batch, got %d", len(batch))
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}
}

func TestBuildPinterestAccountBatch_SyncTypePreserved(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestPinterestAccount("acc1", "pinterest123", kafkamodels.PinterestAccountTypeProfile, "", true, true),
	}

	testCases := []string{"incremental", "full_sync", "immediate"}

	for _, syncType := range testCases {
		batch, _ := buildPinterestAccountBatch(accounts, syncType, log)
		if len(batch) > 0 && batch[0].SyncType != syncType {
			t.Errorf("expected SyncType %q, got %q", syncType, batch[0].SyncType)
		}
	}
}

func TestBuildPinterestAccountBatch_DefaultAccountType(t *testing.T) {
	log := zerolog.Nop()

	// Create account without explicit type
	acc := mongomodels.SocialIntegration{
		ID:                 primitive.NewObjectID(),
		PlatformType:       mongomodels.PlatformPinterest,
		PlatformIdentifier: "pinterest123",
		Type:               "", // empty type
		State:              mongomodels.StateAdded,
		Validity:           mongomodels.ValidityValid,
		AccessToken:        "test-token",
		WorkspaceID:        primitive.NewObjectID(),
	}

	batch, skipped := buildPinterestAccountBatch([]mongomodels.SocialIntegration{acc}, "incremental", log)

	if len(batch) != 1 {
		t.Errorf("expected 1 account in batch, got %d", len(batch))
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}

	// Should default to profile
	if batch[0].AccountType != kafkamodels.PinterestAccountTypeProfile {
		t.Errorf("expected AccountType to default to 'profile', got %q", batch[0].AccountType)
	}
}

// ================== Helper Functions ==================

func createTestPinterestAccount(id, pinterestID, accountType, boardID string, hasAccessToken, hasWorkspace bool) mongomodels.SocialIntegration {
	acc := mongomodels.SocialIntegration{
		ID:                 primitive.NewObjectID(),
		PlatformType:       mongomodels.PlatformPinterest,
		PlatformIdentifier: pinterestID,
		Type:               accountType,
		State:              mongomodels.StateAdded,
		Validity:           mongomodels.ValidityValid,
	}

	if hasWorkspace {
		acc.WorkspaceID = primitive.NewObjectID()
	}
	if hasAccessToken {
		acc.AccessToken = "test-access-token-" + pinterestID
	}
	if boardID != "" {
		acc.ExtraData = map[string]interface{}{
			"board_id": boardID,
		}
	}

	return acc
}
