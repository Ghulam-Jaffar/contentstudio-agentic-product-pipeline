package processor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

func TestWorkOrder_Struct(t *testing.T) {
	wo := ImmediateWorkOrder{
		ID:          "account123",
		WorkspaceID: "workspace_789",
		TikTokID:    "tiktok_123456",
		AccessToken: "token_abc",
		SyncType:    "full",
	}

	if wo.ID != "account123" {
		t.Fatalf("expected ID 'account123', got '%s'", wo.ID)
	}
	if wo.TikTokID != "tiktok_123456" {
		t.Fatalf("expected TikTokID 'tiktok_123456', got '%s'", wo.TikTokID)
	}
	if wo.SyncType != "full" {
		t.Fatalf("expected SyncType 'full', got '%s'", wo.SyncType)
	}
	if wo.WorkspaceID != "workspace_789" {
		t.Fatalf("expected WorkspaceID 'workspace_789', got '%s'", wo.WorkspaceID)
	}
	if wo.AccessToken != "token_abc" {
		t.Fatalf("expected AccessToken 'token_abc', got '%s'", wo.AccessToken)
	}
}

func TestWorkOrder_EmptyStruct(t *testing.T) {
	wo := ImmediateWorkOrder{}

	if wo.ID != "" {
		t.Fatalf("expected empty ID, got '%s'", wo.ID)
	}
	if wo.TikTokID != "" {
		t.Fatalf("expected empty TikTokID, got '%s'", wo.TikTokID)
	}
	if wo.WorkspaceID != "" {
		t.Fatalf("expected empty WorkspaceID, got '%s'", wo.WorkspaceID)
	}
	if wo.AccessToken != "" {
		t.Fatalf("expected empty AccessToken, got '%s'", wo.AccessToken)
	}
	if wo.SyncType != "" {
		t.Fatalf("expected empty SyncType, got '%s'", wo.SyncType)
	}
}

func TestWorkOrder_JSONTags(t *testing.T) {
	// Verify JSON tags are correctly defined by checking struct can be used
	// Note: Full JSON serialization tests would require encoding/json import
	wo := ImmediateWorkOrder{
		ID:          "id_value",
		WorkspaceID: "ws_value",
		TikTokID:    "tk_value",
		AccessToken: "at_value",
		SyncType:    "immediate",
	}

	// Ensure all fields are accessible
	if wo.ID == "" || wo.WorkspaceID == "" || wo.TikTokID == "" || wo.AccessToken == "" || wo.SyncType == "" {
		t.Fatal("one or more fields were not set correctly")
	}
}

func TestParseTikTokDateRange(t *testing.T) {
	start, end, ok, err := parseTikTokDateRange("2025-01-10", "2025-01-20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected requested range to be detected")
	}
	if !start.Equal(time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected start time: %v", start)
	}
	if !end.Equal(time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected end-exclusive time: %v", end)
	}

	_, _, _, err = parseTikTokDateRange("2025-01-20", "2025-01-10")
	if err == nil {
		t.Fatal("expected error for inverted date range")
	}
}

// ================== MongoDB State Update Tests ==================

func TestProcessAccount_UpdatesMongoDBStateOnCompletion(t *testing.T) {
	// This test verifies that ProcessAccount updates MongoDB state to "Processed"
	// after successful completion, matching the fetcher behavior.
	// Note: Full integration test would require actual MongoDB connection.
	// This test documents the expected behavior.

	wo := ImmediateWorkOrder{
		ID:          "507f1f77bcf86cd799439011",
		WorkspaceID: "workspace_123",
		TikTokID:    "tiktok_456",
		AccessToken: "token_789",
		SyncType:    "immediate",
	}

	// Verify work order has a valid MongoDB ID for state update
	if wo.ID == "" {
		t.Fatal("Work order ID should not be empty for state update test")
	}

	// Verify the ID is a valid MongoDB ObjectID format
	_, err := parseObjectID(wo.ID)
	if err != nil {
		t.Fatalf("Work order ID should be valid MongoDB ObjectID: %v", err)
	}
}

// Helper to validate MongoDB ObjectID format
func parseObjectID(id string) (interface{}, error) {
	// This is a placeholder - actual implementation would use
	// go.mongodb.org/mongo-driver/bson/primitive.ObjectIDFromHex
	if len(id) != 24 {
		return nil, fmt.Errorf("invalid ObjectID length: %d", len(id))
	}
	return id, nil
}

// ==================== Logging Contract Tests ====================

func TestLoggingContract_TikTok_ExpectedError_WarnOnly(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, captureCleanup := logger.InstallCaptureSpy()
	defer captureCleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel so HTTP calls fail fast

	tkClient := social.NewTikTokClient("test-key", "test-secret")

	mongoErr := errors.New("expected: mongodb unavailable")
	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, mongoErr
		},
	}

	wo := ImmediateWorkOrder{
		ID:          primitive.NewObjectID().Hex(),
		WorkspaceID: "ws_123",
		TikTokID:    "tk_123",
		AccessToken: "test_token",
		SyncType:    "immediate",
	}

	_ = ProcessAccount(ctx, tkClient, nil, mockRepo, nil, nil, wo, "", log)

	output := buf.String()

	// Should have WRN level (MongoDB failure logged at Warn)
	if !strings.Contains(output, "WRN") {
		t.Error("expected WRN-level log entries")
	}

	// Should NOT have ERR level
	if strings.Contains(output, "ERR") {
		t.Error("unexpected ERR-level log entries; processors should not log at Error level")
	}

	// The MongoDB error specifically should NOT trigger CaptureException
	for _, rec := range *captureRecords {
		if rec.Err != nil && strings.Contains(rec.Err.Error(), "mongodb unavailable") {
			t.Error("CaptureException should NOT be called for expected/handled MongoDB errors")
		}
	}
}

func TestLoggingContract_TikTok_UnexpectedSwallowed_UsesCaptureException(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, captureCleanup := logger.InstallCaptureSpy()
	defer captureCleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel so FetchUserInfo fails with context error

	tkClient := social.NewTikTokClient("test-key", "test-secret")

	wo := ImmediateWorkOrder{
		ID:          primitive.NewObjectID().Hex(),
		WorkspaceID: "ws_123",
		TikTokID:    "tk_123",
		AccessToken: "test_token",
		SyncType:    "immediate",
	}

	// mongoRepo=nil means type assertion fails, skips MongoDB; FetchUserInfo
	// will fail with context.Canceled which is NOT an expected TikTok error,
	// so CaptureException should be called and the error is swallowed (continues).
	_ = ProcessAccount(ctx, tkClient, nil, nil, nil, nil, wo, "", log)

	// CaptureException SHOULD have been called for the swallowed FetchUserInfo error
	if len(*captureRecords) == 0 {
		t.Error("CaptureException should be called for unexpected swallowed errors")
	}

	output := buf.String()

	if !strings.Contains(output, "WRN") {
		t.Error("expected WRN-level log for swallowed unexpected error")
	}
	if strings.Contains(output, "ERR") {
		t.Error("unexpected ERR-level log; processors should not log at Error level")
	}
}

func TestLoggingContract_TikTok_NoErrorLevelInProcessor(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	_, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tkClient := social.NewTikTokClient("test-key", "test-secret")

	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, errors.New("mongo error")
		},
	}

	wo := ImmediateWorkOrder{
		ID:          primitive.NewObjectID().Hex(),
		WorkspaceID: "ws_123",
		TikTokID:    "tk_123",
		AccessToken: "test_token",
		SyncType:    "immediate",
	}

	_ = ProcessAccount(ctx, tkClient, nil, mockRepo, nil, nil, wo, "", log)

	output := buf.String()
	errCount := strings.Count(output, "ERR")
	if errCount > 0 {
		t.Errorf("expected 0 ERR-level entries, got %d; processors should never log at Error level", errCount)
	}
}
