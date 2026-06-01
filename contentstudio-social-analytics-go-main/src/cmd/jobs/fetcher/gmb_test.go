package fetcher

import (
	"testing"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"

	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Constants Tests ==================

func TestGMBConstants(t *testing.T) {
	if gmbMongoFetchSize != 50 {
		t.Errorf("gmbMongoFetchSize = %d, want 50", gmbMongoFetchSize)
	}
	if gmbKafkaBatchSize != 50 {
		t.Errorf("gmbKafkaBatchSize = %d, want 50", gmbKafkaBatchSize)
	}
	if gmbUpdateIntervalHours != 6 {
		t.Errorf("gmbUpdateIntervalHours = %d, want 6", gmbUpdateIntervalHours)
	}
	if topicGMBBatch != "work-order-gmb" {
		t.Errorf("topicGMBBatch = %q, want %q", topicGMBBatch, "work-order-gmb")
	}
}

// ================== parseGMBPlatformIdentifier Tests ==================

func TestParseGMBPlatformIdentifier_Valid(t *testing.T) {
	accountID, locationID, ok := parseGMBPlatformIdentifier("accounts/111098760901606453992/locations/2941480710306834283")
	if !ok {
		t.Fatal("expected ok=true for valid platform_identifier")
	}
	if accountID != "111098760901606453992" {
		t.Errorf("accountID = %q, want %q", accountID, "111098760901606453992")
	}
	if locationID != "2941480710306834283" {
		t.Errorf("locationID = %q, want %q", locationID, "2941480710306834283")
	}
}

func TestParseGMBPlatformIdentifier_Invalid(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"no_slashes", "foobar"},
		{"missing_locations", "accounts/123"},
		{"wrong_prefix", "foo/123/locations/456"},
		{"wrong_middle", "accounts/123/bar/456"},
		{"empty_account_id", "accounts//locations/456"},
		{"empty_location_id", "accounts/123/locations/"},
		{"too_many_parts", "accounts/123/locations/456/extra"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, ok := parseGMBPlatformIdentifier(tc.input)
			if ok {
				t.Errorf("expected ok=false for %q", tc.input)
			}
		})
	}
}

// ================== buildGMBAccountBatch Tests ==================

func TestBuildGMBAccountBatch_ValidAccounts(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestGMBAccount("acc1", "accounts/111/locations/222", true, true, true),
		createTestGMBAccount("acc2", "accounts/333/locations/444", true, true, true),
	}

	batch, skipped := buildGMBAccountBatch(accounts, "incremental", log)

	if len(batch) != 2 {
		t.Errorf("expected 2 accounts in batch, got %d", len(batch))
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}

	if batch[0].AccountID != "111" {
		t.Errorf("expected AccountID '111', got %q", batch[0].AccountID)
	}
	if batch[0].LocationID != "222" {
		t.Errorf("expected LocationID '222', got %q", batch[0].LocationID)
	}
	if batch[0].SyncType != "incremental" {
		t.Errorf("expected SyncType 'incremental', got %q", batch[0].SyncType)
	}
	if batch[0].LanguageCode != "en" {
		t.Errorf("expected default LanguageCode 'en', got %q", batch[0].LanguageCode)
	}
}

func TestBuildGMBAccountBatch_MissingAccessToken(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestGMBAccount("acc1", "accounts/111/locations/222", false, true, true),
	}

	batch, skipped := buildGMBAccountBatch(accounts, "incremental", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts in batch, got %d", len(batch))
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestBuildGMBAccountBatch_MissingRefreshToken(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestGMBAccount("acc1", "accounts/111/locations/222", true, false, true),
	}

	batch, skipped := buildGMBAccountBatch(accounts, "full_sync", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts in batch, got %d", len(batch))
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestBuildGMBAccountBatch_MissingWorkspaceID(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestGMBAccount("acc1", "accounts/111/locations/222", true, true, false),
	}

	batch, skipped := buildGMBAccountBatch(accounts, "incremental", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts in batch, got %d", len(batch))
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestBuildGMBAccountBatch_InvalidPlatformIdentifier(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestGMBAccount("acc1", "invalid-identifier", true, true, true),
	}

	batch, skipped := buildGMBAccountBatch(accounts, "incremental", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts in batch, got %d", len(batch))
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestBuildGMBAccountBatch_MixedValidity(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestGMBAccount("acc1", "accounts/111/locations/222", true, true, true),  // valid
		createTestGMBAccount("acc2", "accounts/333/locations/444", false, true, true), // no access token
		createTestGMBAccount("acc3", "accounts/555/locations/666", true, true, true),  // valid
	}

	batch, skipped := buildGMBAccountBatch(accounts, "incremental", log)

	if len(batch) != 2 {
		t.Errorf("expected 2 valid, got %d", len(batch))
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestBuildGMBAccountBatch_SyncTypePreserved(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestGMBAccount("acc1", "accounts/111/locations/222", true, true, true),
	}

	for _, syncType := range []string{"incremental", "full_sync"} {
		batch, _ := buildGMBAccountBatch(accounts, syncType, log)
		if len(batch) != 1 {
			t.Errorf("sync_type=%s: expected 1 account, got %d", syncType, len(batch))
			continue
		}
		if batch[0].SyncType != syncType {
			t.Errorf("sync_type=%s: expected SyncType %q, got %q", syncType, syncType, batch[0].SyncType)
		}
	}
}

func TestBuildGMBAccountBatch_WorkspaceIDSet(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestGMBAccount("acc1", "accounts/111/locations/222", true, true, true),
	}

	batch, _ := buildGMBAccountBatch(accounts, "incremental", log)

	if len(batch) != 1 {
		t.Fatalf("expected 1 account, got %d", len(batch))
	}
	if batch[0].WorkspaceID == "" {
		t.Error("expected WorkspaceID to be set")
	}
}

func TestBuildGMBAccountBatch_IDPreserved(t *testing.T) {
	log := zerolog.Nop()

	accounts := []mongomodels.SocialIntegration{
		createTestGMBAccount("acc1", "accounts/111/locations/222", true, true, true),
	}

	batch, _ := buildGMBAccountBatch(accounts, "incremental", log)

	if len(batch) != 1 {
		t.Fatalf("expected 1 account, got %d", len(batch))
	}
	if batch[0].ID != accounts[0].ID.Hex() {
		t.Errorf("expected ID %q, got %q", accounts[0].ID.Hex(), batch[0].ID)
	}
}

func TestBuildGMBAccountBatch_LanguageCodeFromAccount(t *testing.T) {
	log := zerolog.Nop()

	account := createTestGMBAccount("acc1", "accounts/111/locations/222", true, true, true)
	account.LanguageCode = "fr"

	batch, _ := buildGMBAccountBatch([]mongomodels.SocialIntegration{account}, "incremental", log)

	if len(batch) != 1 {
		t.Fatalf("expected 1 account, got %d", len(batch))
	}
	if batch[0].LanguageCode != "fr" {
		t.Errorf("expected LanguageCode 'fr', got %q", batch[0].LanguageCode)
	}
}

func TestBuildGMBAccountBatch_EmptyBatch(t *testing.T) {
	log := zerolog.Nop()

	batch, skipped := buildGMBAccountBatch(nil, "incremental", log)

	if len(batch) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(batch))
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}
}

// ================== Model round-trip Test ==================

func TestGMBBatchWorkOrderFields(t *testing.T) {
	wo := kafkamodels.GMBAccountWorkOrder{
		ID:           primitive.NewObjectID().Hex(),
		WorkspaceID:  primitive.NewObjectID().Hex(),
		AccountID:    "111",
		LocationID:   "222",
		AccessToken:  "at",
		RefreshToken: "rt",
		AccountName:  "My Biz",
		LocationName: "Main Branch",
		LanguageCode: "en",
		SyncType:     "incremental",
	}

	if wo.AccountID != "111" {
		t.Errorf("AccountID = %q, want %q", wo.AccountID, "111")
	}
	if wo.LocationID != "222" {
		t.Errorf("LocationID = %q, want %q", wo.LocationID, "222")
	}
}

// ================== Helpers ==================

// createTestGMBAccount creates a test SocialIntegration for GMB testing.
func createTestGMBAccount(idHex, platformIdentifier string, hasAccessToken, hasRefreshToken, hasWorkspaceID bool) mongomodels.SocialIntegration {
	objID := primitive.NewObjectID()

	account := mongomodels.SocialIntegration{
		ID:                 objID,
		PlatformType:       mongomodels.PlatformGMB,
		PlatformIdentifier: platformIdentifier,
		PlatformName:       "Test Business",
	}

	if hasAccessToken {
		account.AccessToken = "test-access-token"
	}
	if hasRefreshToken {
		account.RefreshToken = "test-refresh-token"
	}
	if hasWorkspaceID {
		account.WorkspaceID = primitive.NewObjectID()
	}

	return account
}
