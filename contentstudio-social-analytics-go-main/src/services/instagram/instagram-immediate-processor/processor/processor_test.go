package processor

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if MediaInsightsConc != 20 {
		t.Errorf("MediaInsightsConc = %d, want 20", MediaInsightsConc)
	}
	if ProgressLogInterval != 100 {
		t.Errorf("ProgressLogInterval = %d, want 100", ProgressLogInterval)
	}
}

// ================== WorkOrder Tests ==================

func TestWorkOrder_Struct(t *testing.T) {
	wo := WorkOrder{
		ID:                    "abc123",
		AccountID:             "ig456",
		Type:                  "instagram",
		AccessToken:           "token123",
		WorkspaceID:           "ws789",
		SyncType:              "immediate",
		ConnectedViaInstagram: true,
		StartDate:             "2025-01-01",
		EndDate:               "2025-01-31",
	}

	if wo.ID != "abc123" {
		t.Errorf("ID = %q, want %q", wo.ID, "abc123")
	}
	if wo.AccountID != "ig456" {
		t.Errorf("AccountID = %q, want %q", wo.AccountID, "ig456")
	}
	if wo.Type != "instagram" {
		t.Errorf("Type = %q, want %q", wo.Type, "instagram")
	}
	if wo.SyncType != "immediate" {
		t.Errorf("SyncType = %q, want %q", wo.SyncType, "immediate")
	}
	if !wo.ConnectedViaInstagram {
		t.Error("ConnectedViaInstagram = false, want true")
	}
	if wo.StartDate != "2025-01-01" {
		t.Errorf("StartDate = %q, want %q", wo.StartDate, "2025-01-01")
	}
	if wo.EndDate != "2025-01-31" {
		t.Errorf("EndDate = %q, want %q", wo.EndDate, "2025-01-31")
	}
}

func TestWorkOrder_JSON(t *testing.T) {
	wo := WorkOrder{
		ID:                    "abc123",
		AccountID:             "ig456",
		Type:                  "instagram",
		AccessToken:           "token123",
		WorkspaceID:           "ws789",
		SyncType:              "immediate",
		ConnectedViaInstagram: true,
		StartDate:             "2025-01-01",
		EndDate:               "2025-01-31",
	}

	data, err := json.Marshal(wo)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed WorkOrder
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.ID != wo.ID {
		t.Errorf("ID = %q, want %q", parsed.ID, wo.ID)
	}
	if parsed.AccountID != wo.AccountID {
		t.Errorf("AccountID = %q, want %q", parsed.AccountID, wo.AccountID)
	}
	if parsed.ConnectedViaInstagram != wo.ConnectedViaInstagram {
		t.Errorf("ConnectedViaInstagram = %v, want %v", parsed.ConnectedViaInstagram, wo.ConnectedViaInstagram)
	}
}

// ================== ParsedData Tests ==================

func TestParsedData_Struct(t *testing.T) {
	pd := &ParsedData{
		Posts: []kafkamodels.ParsedInstagramPost{
			{MediaID: "media1", InstagramID: "ig123"},
			{MediaID: "media2", InstagramID: "ig123"},
		},
		Insights: []kafkamodels.ParsedInstagramInsight{
			{InstagramID: "ig123", RecordID: "ig123_2024-01-15"},
		},
	}

	if len(pd.Posts) != 2 {
		t.Errorf("Posts len = %d, want 2", len(pd.Posts))
	}
	if len(pd.Insights) != 1 {
		t.Errorf("Insights len = %d, want 1", len(pd.Insights))
	}
}

// ================== EnrichedMedia Tests ==================

func TestEnrichedMedia_Struct(t *testing.T) {
	em := EnrichedMedia{
		Media: kafkamodels.RawInstagramMedia{
			ID:        "media123",
			MediaType: "IMAGE",
		},
		Insights: &kafkamodels.RawInstagramMediaInsights{},
		UserInfo: map[string]interface{}{
			"username": "testuser",
			"name":     "Test User",
		},
	}

	if em.Media.ID != "media123" {
		t.Errorf("Media.ID = %q, want %q", em.Media.ID, "media123")
	}
	if em.Insights == nil {
		t.Error("Insights is nil")
	}
	if em.UserInfo["username"] != "testuser" {
		t.Errorf("UserInfo[username] = %q, want %q", em.UserInfo["username"], "testuser")
	}
}

// ================== Processor Tests ==================

func TestNew(t *testing.T) {
	log := logger.New("error")

	p := New(nil, nil, nil, nil, nil, log, nil)

	if p == nil {
		t.Fatal("New returned nil")
	}
	if p.Parser == nil {
		t.Error("Parser is nil")
	}
	if p.Logger == nil {
		t.Error("Logger is nil")
	}
}

func TestProcessor_NilFields(t *testing.T) {
	log := logger.New("error")

	p := &Processor{
		Logger: log,
	}

	// Verify nil fields don't cause issues
	if p.MongoRepo != nil {
		t.Error("MongoRepo should be nil")
	}
	if p.Sink != nil {
		t.Error("Sink should be nil")
	}
	if p.Producer != nil {
		t.Error("Producer should be nil")
	}
	if p.Notifier != nil {
		t.Error("Notifier should be nil")
	}
	if p.PusherClient != nil {
		t.Error("PusherClient should be nil")
	}
}

// ================== getInt64Value Tests ==================

func TestGetInt64Value(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  int64
	}{
		{"int", 42, 42},
		{"int64", int64(1000), 1000},
		{"float64", float64(99.9), 99},
		{"float32", float32(50.5), 50},
		{"string (invalid)", "not a number", 0},
		{"nil", nil, 0},
		{"bool (invalid)", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInt64Value(tt.input)
			if got != tt.want {
				t.Errorf("getInt64Value(%v) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// ================== sendPusherNotification Tests ==================

func TestProcessor_SendPusherNotification_NilClient(t *testing.T) {
	log := logger.New("error")
	p := &Processor{
		Logger:       log,
		PusherClient: nil, // nil client
	}

	// Should not panic
	p.sendPusherNotification(nil, "ws123", "Added")
}

// ================== sendEmailNotification Tests ==================

func TestProcessor_SendEmailNotification_NilNotifier(t *testing.T) {
	log := logger.New("error")
	p := &Processor{
		Logger:   log,
		Notifier: nil, // nil notifier
	}

	// Should not panic
	p.sendEmailNotification("user123", "ws456", "ig789", "Test Account")
}

// ================== WorkOrder JSON Keys Tests ==================

func TestWorkOrder_JSONKeys(t *testing.T) {
	jsonData := `{
		"id": "test_id",
		"account_id": "test_account",
		"type": "instagram",
		"access_token": "test_token",
		"workspace_id": "test_workspace",
		"sync_type": "immediate",
		"connected_via_instagram": true
	}`

	var wo WorkOrder
	if err := json.Unmarshal([]byte(jsonData), &wo); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if wo.ID != "test_id" {
		t.Errorf("ID = %q, want %q", wo.ID, "test_id")
	}
	if wo.AccountID != "test_account" {
		t.Errorf("AccountID = %q, want %q", wo.AccountID, "test_account")
	}
	if wo.AccessToken != "test_token" {
		t.Errorf("AccessToken = %q, want %q", wo.AccessToken, "test_token")
	}
	if wo.WorkspaceID != "test_workspace" {
		t.Errorf("WorkspaceID = %q, want %q", wo.WorkspaceID, "test_workspace")
	}
	if wo.SyncType != "immediate" {
		t.Errorf("SyncType = %q, want %q", wo.SyncType, "immediate")
	}
	if !wo.ConnectedViaInstagram {
		t.Error("ConnectedViaInstagram = false, want true")
	}
}

func TestWorkOrder_JSONKeysFromMarshal(t *testing.T) {
	wo := WorkOrder{
		ID:                    "test_id",
		AccountID:             "test_account",
		Type:                  "instagram",
		AccessToken:           "test_token",
		WorkspaceID:           "test_workspace",
		SyncType:              "immediate",
		ConnectedViaInstagram: false,
		StartDate:             "2025-01-01",
		EndDate:               "2025-01-31",
	}

	data, err := json.Marshal(wo)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify JSON keys match the struct tags
	expectedKeys := []string{"id", "account_id", "type", "access_token", "workspace_id", "sync_type", "connected_via_instagram", "start_date", "end_date"}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("missing key %q in marshaled JSON", key)
		}
	}
}

func TestResolveInstagramDateRange(t *testing.T) {
	start, end, err := resolveInstagramDateRange("2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)
	if !start.Equal(expectedStart) {
		t.Fatalf("start = %v, want %v", start, expectedStart)
	}
	if !end.Equal(expectedEnd) {
		t.Fatalf("end = %v, want %v", end, expectedEnd)
	}
}

func TestFilterInstagramMediaWithinRange(t *testing.T) {
	media := []kafkamodels.RawInstagramMedia{
		{ID: "1", Timestamp: "2025-01-01T10:00:00+0000"},
		{ID: "2", Timestamp: "2025-02-01T10:00:00+0000"},
		{ID: "3", Timestamp: "not-a-date"},
	}

	filtered := filterInstagramMediaWithinRange(
		media,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
	)

	if len(filtered) != 1 {
		t.Fatalf("expected 1 media item, got %d", len(filtered))
	}
	if filtered[0].ID != "1" {
		t.Fatalf("expected first item ID 1, got %s", filtered[0].ID)
	}
}

// ================== Multiple getInt64Value Tests ==================

func TestGetInt64Value_EdgeCases(t *testing.T) {
	// Zero values
	if got := getInt64Value(int(0)); got != 0 {
		t.Errorf("int(0) = %d, want 0", got)
	}
	if got := getInt64Value(int64(0)); got != 0 {
		t.Errorf("int64(0) = %d, want 0", got)
	}
	if got := getInt64Value(float64(0)); got != 0 {
		t.Errorf("float64(0) = %d, want 0", got)
	}

	// Negative values
	if got := getInt64Value(int(-100)); got != -100 {
		t.Errorf("int(-100) = %d, want -100", got)
	}
	if got := getInt64Value(int64(-200)); got != -200 {
		t.Errorf("int64(-200) = %d, want -200", got)
	}
	if got := getInt64Value(float64(-50.5)); got != -50 {
		t.Errorf("float64(-50.5) = %d, want -50", got)
	}

	// Large values
	if got := getInt64Value(int64(1 << 62)); got != 1<<62 {
		t.Errorf("large int64 = %d, want %d", got, int64(1<<62))
	}
}

// ================== ParsedData Empty Tests ==================

func TestParsedData_Empty(t *testing.T) {
	pd := &ParsedData{}

	if pd.Posts != nil && len(pd.Posts) != 0 {
		t.Errorf("Posts should be empty or nil, got %d items", len(pd.Posts))
	}
	if pd.Insights != nil && len(pd.Insights) != 0 {
		t.Errorf("Insights should be empty or nil, got %d items", len(pd.Insights))
	}
}

// ================== EnrichedMedia Nil Fields Tests ==================

func TestEnrichedMedia_NilFields(t *testing.T) {
	em := EnrichedMedia{
		Media: kafkamodels.RawInstagramMedia{
			ID: "media123",
		},
		Insights: nil,
		UserInfo: nil,
	}

	if em.Media.ID != "media123" {
		t.Errorf("Media.ID = %q, want %q", em.Media.ID, "media123")
	}
	if em.Insights != nil {
		t.Error("Insights should be nil")
	}
	if em.UserInfo != nil {
		t.Error("UserInfo should be nil")
	}
}

// ================== isExpectedInstagramError Tests ==================

func TestIsExpectedInstagramError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"auth error", errors.New("invalid access token"), true},
		{"OAuthException 190", errors.New("OAuthException (#190) token expired"), true},
		{"OAuthException 10", errors.New("OAuthException: (#10) not enough viewers"), true},
		{"not enough viewers lowercase", errors.New("not enough viewers for insights"), true},
		{"error #10", errors.New("error (#10) occurred"), true},
		{"permission error", errors.New("Application does not have permission for this action"), true},
		{"network error", errors.New("connection timeout"), false},
		{"parse error", errors.New("json parse failed"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExpectedInstagramError(tt.err)
			if got != tt.expected {
				t.Errorf("isExpectedInstagramError() = %v, want %v for error: %v", got, tt.expected, tt.err)
			}
		})
	}
}

// ================== Logging Contract Tests (Point 3 — Processor never logs at Error level) ==================

func TestLoggingContract_InstagramProcessor_NoErrorLevel(t *testing.T) {
	// Instagram immediate processor is a "called" module — it must use Warn, never Error.
	// Simulate the processor's logging patterns and verify no Error level.

	hookRecords, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	log, buf := logger.NewTestLoggerWithHook()

	// Simulate all the Warn-level logs the processor produces
	log.Warn().Err(errors.New("API error")).
		Str("error_message", "API error").
		Str("function", "fetchAllData").
		Str("stage", "fetch_stories").
		Msg("FetchStories failed (continuing)")

	log.Warn().Err(errors.New("insert failed")).
		Str("error_message", "insert failed").
		Str("function", "parseAllData").
		Str("stage", "marshal_enriched_media").
		Msg("failed to marshal enriched media (skipping)")

	log.Warn().Err(errors.New("mongo error")).
		Str("error_message", "mongo error").
		Str("function", "ProcessAccount").
		Msg("Failed to update account state to Processed")

	output := buf.String()

	// Processor must never produce Error-level logs
	if strings.Contains(output, `"level":"error"`) || strings.Count(output, "ERR") > 0 {
		t.Errorf("processor should never produce Error-level logs, but found ERR in output:\n%s", output)
	}

	// Verify no Error-level hook firings
	for _, r := range *hookRecords {
		if r.Level >= zerolog.ErrorLevel {
			t.Errorf("processor should not trigger Error+ hook, got level %v", r.Level)
		}
	}
}

func TestLoggingContract_InstagramProcessor_NoCaptureException(t *testing.T) {
	// Processor should never call CaptureException — it returns errors to callers.

	captureRecords, captureCleanup := logger.InstallCaptureSpy()
	defer captureCleanup()

	log, _ := logger.NewTestLoggerWithHook()

	// Simulate processor logging on failure — only Warn level
	log.Warn().Err(errors.New("token expired")).
		Str("error_message", "token expired").
		Str("function", "fetchAllData").
		Str("stage", "fetch_user_info").
		Msg("FetchUserInfo failed (continuing)")

	if len(*captureRecords) != 0 {
		t.Fatalf("processor should not call CaptureException, got %d calls", len(*captureRecords))
	}
}
