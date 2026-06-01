package fetcher

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"

	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Constants Tests ==================

func TestTikTokConstants(t *testing.T) {
	if tiktokMongoFetchSize != 50 {
		t.Errorf("tiktokMongoFetchSize = %d, want 50", tiktokMongoFetchSize)
	}
	if tiktokKafkaBatchSize != 200 {
		t.Errorf("tiktokKafkaBatchSize = %d, want 200", tiktokKafkaBatchSize)
	}
	if tiktokUpdateIntervalHours != 6 {
		t.Errorf("tiktokUpdateIntervalHours = %d, want 6", tiktokUpdateIntervalHours)
	}
	if topicTikTokBatch != "work-order-tiktok-batch" {
		t.Errorf("topicTikTokBatch = %q, want %q", topicTikTokBatch, "work-order-tiktok-batch")
	}
}

// ================== buildTikTokAccountBatch Tests ==================

func TestBuildTikTokAccountBatch_ValidAccounts(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestTikTokAccount("acc1", "tiktok123", true, true, true, true),
		createTestTikTokAccount("acc2", "tiktok456", true, true, true, true),
	}

	batch, skipped := buildTikTokAccountBatch(accounts, "incremental", log)

	if len(batch) != 2 {
		t.Errorf("expected 2 accounts in batch, got %d", len(batch))
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}

	// Verify first account
	if batch[0].TikTokID != "tiktok123" {
		t.Errorf("expected TikTokID 'tiktok123', got %q", batch[0].TikTokID)
	}
	if batch[0].SyncType != "incremental" {
		t.Errorf("expected SyncType 'incremental', got %q", batch[0].SyncType)
	}
}

func TestBuildTikTokAccountBatch_MissingAccessToken(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestTikTokAccount("acc1", "tiktok123", false, true, true, true), // no access token
	}

	batch, skipped := buildTikTokAccountBatch(accounts, "full_sync", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts in batch, got %d", len(batch))
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestBuildTikTokAccountBatch_MissingRefreshToken(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestTikTokAccount("acc1", "tiktok123", true, false, true, true), // no refresh token
	}

	batch, skipped := buildTikTokAccountBatch(accounts, "full_sync", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts in batch, got %d", len(batch))
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestBuildTikTokAccountBatch_MissingScope(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestTikTokAccount("acc1", "tiktok123", true, true, false, true), // no scope
	}

	batch, skipped := buildTikTokAccountBatch(accounts, "full_sync", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts in batch, got %d", len(batch))
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestBuildTikTokAccountBatch_MissingWorkspaceID(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestTikTokAccount("acc1", "tiktok123", true, true, true, false), // no workspace
	}

	batch, skipped := buildTikTokAccountBatch(accounts, "full_sync", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts in batch, got %d", len(batch))
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestBuildTikTokAccountBatch_MixedValidity(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestTikTokAccount("acc1", "tiktok1", true, true, true, true),    // valid
		createTestTikTokAccount("acc2", "tiktok2", false, true, true, true),   // missing access token
		createTestTikTokAccount("acc3", "tiktok3", true, true, true, true),    // valid
		createTestTikTokAccount("acc4", "tiktok4", true, false, true, true),   // missing refresh token
		createTestTikTokAccount("acc5", "tiktok5", true, true, false, true),   // missing scope
		createTestTikTokAccount("acc6", "tiktok6", true, true, true, false),   // missing workspace
	}

	batch, skipped := buildTikTokAccountBatch(accounts, "incremental", log)

	if len(batch) != 2 {
		t.Errorf("expected 2 valid accounts in batch, got %d", len(batch))
	}
	if skipped != 4 {
		t.Errorf("expected 4 skipped, got %d", skipped)
	}
}

func TestBuildTikTokAccountBatch_EmptyInput(t *testing.T) {
	log := zerolog.Nop()

	batch, skipped := buildTikTokAccountBatch([]mongomodels.SocialIntegration{}, "incremental", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts in batch, got %d", len(batch))
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}
}

func TestBuildTikTokAccountBatch_SyncTypePreserved(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestTikTokAccount("acc1", "tiktok123", true, true, true, true),
	}

	testCases := []string{"incremental", "full_sync", "custom"}

	for _, syncType := range testCases {
		batch, _ := buildTikTokAccountBatch(accounts, syncType, log)
		if len(batch) > 0 && batch[0].SyncType != syncType {
			t.Errorf("expected SyncType %q, got %q", syncType, batch[0].SyncType)
		}
	}
}

// ================== processTikTokBatches Tests ==================

func TestProcessTikTokBatches_NoAccountsNeedingUpdate(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	mockRepo := &mockUnifiedSocialRepository{
		accounts:   []mongomodels.SocialIntegration{},
		totalCount: 0,
	}
	mockProd := &mockProducer{}

	processTikTokBatches(ctx, mockRepo, mockProd, log, "incremental")

	if len(mockProd.getMessages()) != 0 {
		t.Error("expected no messages produced when no accounts need update")
	}
}

func TestProcessTikTokBatches_CountError(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	mockRepo := &mockUnifiedSocialRepository{
		countErr: errors.New("count error"),
	}
	mockProd := &mockProducer{}

	processTikTokBatches(ctx, mockRepo, mockProd, log, "incremental")

	if len(mockProd.getMessages()) != 0 {
		t.Error("expected no messages produced when count fails")
	}
}

func TestProcessTikTokBatches_PaginationError(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	mockRepo := &mockUnifiedSocialRepository{
		totalCount:   100,
		paginatedErr: errors.New("pagination error"),
	}
	mockProd := &mockProducer{}

	processTikTokBatches(ctx, mockRepo, mockProd, log, "incremental")

	if len(mockProd.getMessages()) != 0 {
		t.Error("expected no messages produced when pagination fails")
	}
}

func TestProcessTikTokBatches_ProducesValidBatches(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	accounts := make([]mongomodels.SocialIntegration, 5)
	for i := 0; i < 5; i++ {
		accounts[i] = createTestTikTokAccount("acc", "tiktok"+string(rune('0'+i)), true, true, true, true)
	}

	mockRepo := &mockUnifiedSocialRepository{
		accounts:   accounts,
		totalCount: int64(len(accounts)),
	}
	mockProd := &mockProducer{}

	processTikTokBatches(ctx, mockRepo, mockProd, log, "incremental")

	messages := mockProd.getMessages()
	if len(messages) != 1 {
		t.Fatalf("expected 1 batch message, got %d", len(messages))
	}

	// Verify message structure
	var batch kafkamodels.TikTokBatchWorkOrder
	if err := json.Unmarshal(messages[0].value, &batch); err != nil {
		t.Fatalf("failed to unmarshal batch: %v", err)
	}

	if len(batch.Accounts) != 5 {
		t.Errorf("expected 5 accounts in batch, got %d", len(batch.Accounts))
	}
	if batch.SyncType != "incremental" {
		t.Errorf("expected SyncType 'incremental', got %q", batch.SyncType)
	}
	if batch.BatchID == "" {
		t.Error("expected non-empty BatchID")
	}
}

func TestProcessTikTokBatches_ProducerError(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestTikTokAccount("acc1", "tiktok123", true, true, true, true),
	}

	mockRepo := &mockUnifiedSocialRepository{
		accounts:   accounts,
		totalCount: 1,
	}
	mockProd := &mockProducer{
		err: errors.New("producer error"),
	}

	// Should not panic, just log error and continue
	processTikTokBatches(ctx, mockRepo, mockProd, log, "incremental")
}

func TestProcessTikTokBatches_MultipleBatches(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	// Create more accounts than batch size
	accounts := make([]mongomodels.SocialIntegration, 250)
	for i := 0; i < 250; i++ {
		accounts[i] = createTestTikTokAccount("acc", "tiktok"+string(rune('0'+i%10)), true, true, true, true)
	}

	mockRepo := &mockUnifiedSocialRepository{
		accounts:   accounts,
		totalCount: int64(len(accounts)),
	}
	mockProd := &mockProducer{}

	processTikTokBatches(ctx, mockRepo, mockProd, log, "full_sync")

	messages := mockProd.getMessages()
	// With batch size of 200, 250 accounts should produce 2 batches
	if len(messages) != 2 {
		t.Errorf("expected 2 batch messages for 250 accounts, got %d", len(messages))
	}
}

func TestProcessTikTokBatches_SkipsInvalidAccounts(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestTikTokAccount("acc1", "tiktok1", true, true, true, true),  // valid
		createTestTikTokAccount("acc2", "tiktok2", false, true, true, true), // invalid - no token
	}

	mockRepo := &mockUnifiedSocialRepository{
		accounts:   accounts,
		totalCount: int64(len(accounts)),
	}
	mockProd := &mockProducer{}

	processTikTokBatches(ctx, mockRepo, mockProd, log, "incremental")

	messages := mockProd.getMessages()
	if len(messages) != 1 {
		t.Fatalf("expected 1 batch message, got %d", len(messages))
	}

	var batch kafkamodels.TikTokBatchWorkOrder
	if err := json.Unmarshal(messages[0].value, &batch); err != nil {
		t.Fatalf("failed to unmarshal batch: %v", err)
	}

	if len(batch.Accounts) != 1 {
		t.Errorf("expected 1 valid account in batch, got %d", len(batch.Accounts))
	}
}

func TestProcessTikTokBatches_AllInvalidAccounts(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	// All accounts are invalid (missing tokens)
	accounts := []mongomodels.SocialIntegration{
		createTestTikTokAccount("acc1", "tiktok1", false, true, true, true),
		createTestTikTokAccount("acc2", "tiktok2", false, true, true, true),
	}

	mockRepo := &mockUnifiedSocialRepository{
		accounts:   accounts,
		totalCount: int64(len(accounts)),
	}
	mockProd := &mockProducer{}

	processTikTokBatches(ctx, mockRepo, mockProd, log, "incremental")

	// No messages should be produced since all accounts are invalid
	messages := mockProd.getMessages()
	if len(messages) != 0 {
		t.Errorf("expected 0 batch messages for all invalid accounts, got %d", len(messages))
	}
}

// ================== Helper Functions ==================

func createTestTikTokAccount(id, tiktokID string, hasAccessToken, hasRefreshToken, hasScope, hasWorkspace bool) mongomodels.SocialIntegration {
	acc := mongomodels.SocialIntegration{
		ID:                 primitive.NewObjectID(),
		PlatformType:       mongomodels.PlatformTikTok,
		PlatformIdentifier: tiktokID,
		Type:               mongomodels.TypeProfile,
		State:              mongomodels.StateAdded,
		Validity:           mongomodels.ValidityValid,
	}

	if hasWorkspace {
		acc.WorkspaceID = primitive.NewObjectID()
	}
	if hasAccessToken {
		acc.AccessToken = "test-access-token-" + tiktokID
	}
	if hasRefreshToken {
		acc.RefreshToken = "test-refresh-token-" + tiktokID
	}
	if hasScope {
		acc.Scope = "user.info.basic,user.info.profile,user.info.stats,video.list"
	}

	return acc
}
