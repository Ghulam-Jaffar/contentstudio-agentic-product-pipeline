package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"github.com/rs/zerolog"
)

// ================== Mock Implementations ==================

type MockTikTokVideoFetcher struct {
	FetchUserVideosFunc func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error)
	RefreshTokenFunc    func(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error)
}

func (m *MockTikTokVideoFetcher) FetchUserVideos(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
	if m.FetchUserVideosFunc != nil {
		return m.FetchUserVideosFunc(ctx, userID, accessToken, cursor, maxCount)
	}
	return nil, 0, nil
}

func (m *MockTikTokVideoFetcher) RefreshToken(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(ctx, refreshToken)
	}
	return nil, errors.New("refresh token not configured in mock")
}

type MockTikTokPostSink struct {
	BulkInsertTikTokPostsFunc func(ctx context.Context, posts []*chmodels.TikTokPosts) error
	InsertedPosts             []*chmodels.TikTokPosts
}

func (m *MockTikTokPostSink) BulkInsertTikTokPosts(ctx context.Context, posts []*chmodels.TikTokPosts) error {
	m.InsertedPosts = append(m.InsertedPosts, posts...)
	if m.BulkInsertTikTokPostsFunc != nil {
		return m.BulkInsertTikTokPostsFunc(ctx, posts)
	}
	return nil
}

type MockSocialRepository struct {
	FindByIDFunc             func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error)
	ClearProcessingErrorFunc func(ctx context.Context, id primitive.ObjectID) error
	ClearProcessingErrorIDs  []primitive.ObjectID
}

func (m *MockSocialRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSocialRepository) ClearProcessingError(ctx context.Context, id primitive.ObjectID) error {
	m.ClearProcessingErrorIDs = append(m.ClearProcessingErrorIDs, id)
	if m.ClearProcessingErrorFunc != nil {
		return m.ClearProcessingErrorFunc(ctx, id)
	}
	return nil
}

type MockPusherNotifier struct {
	TriggerFunc func(channel, event string, data interface{}) error
	Triggered   []struct {
		Channel string
		Event   string
		Data    interface{}
	}
}

func (m *MockPusherNotifier) Trigger(channel, event string, data interface{}) error {
	m.Triggered = append(m.Triggered, struct {
		Channel string
		Event   string
		Data    interface{}
	}{channel, event, data})
	if m.TriggerFunc != nil {
		return m.TriggerFunc(channel, event, data)
	}
	return nil
}

type MockEmailNotifier struct {
	SendFunc func(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error
	Sent     []struct {
		UserID, WorkspaceID, Platform, AccountID, AccountName string
		IsCompetitor                                          bool
	}
}

func (m *MockEmailNotifier) SendAnalyticsNotification(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error {
	m.Sent = append(m.Sent, struct {
		UserID, WorkspaceID, Platform, AccountID, AccountName string
		IsCompetitor                                          bool
	}{userID, workspaceID, platform, accountID, accountName, isCompetitor})
	if m.SendFunc != nil {
		return m.SendFunc(userID, workspaceID, platform, accountID, accountName, isCompetitor)
	}
	return nil
}

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if immediateTopic != "immediate-work-order-tiktok" {
		t.Fatalf("expected immediateTopic 'immediate-work-order-tiktok', got '%s'", immediateTopic)
	}
}

// ================== ImmediateWorkOrder Tests ==================

func TestImmediateWorkOrder_Struct(t *testing.T) {
	order := ImmediateWorkOrder{
		ID:          "order-123",
		WorkspaceID: "ws-456",
		TikTokID:    "tiktok-789",
		AccessToken: "test-token",
		SyncType:    "full",
	}

	if order.ID != "order-123" {
		t.Fatalf("expected ID 'order-123', got '%s'", order.ID)
	}
	if order.WorkspaceID != "ws-456" {
		t.Fatalf("expected WorkspaceID 'ws-456', got '%s'", order.WorkspaceID)
	}
	if order.TikTokID != "tiktok-789" {
		t.Fatalf("expected TikTokID 'tiktok-789', got '%s'", order.TikTokID)
	}
	if order.AccessToken != "test-token" {
		t.Fatalf("expected AccessToken 'test-token', got '%s'", order.AccessToken)
	}
	if order.SyncType != "full" {
		t.Fatalf("expected SyncType 'full', got '%s'", order.SyncType)
	}
}

func TestImmediateWorkOrder_JSONMarshal(t *testing.T) {
	order := ImmediateWorkOrder{
		ID:          "order-123",
		TikTokID:    "tiktok-789",
		AccessToken: "test-token",
		SyncType:    "incremental",
	}

	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ImmediateWorkOrder
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != order.ID {
		t.Fatalf("expected ID '%s', got '%s'", order.ID, decoded.ID)
	}
	if decoded.TikTokID != order.TikTokID {
		t.Fatalf("expected TikTokID '%s', got '%s'", order.TikTokID, decoded.TikTokID)
	}
}

func TestImmediateWorkOrder_JSONUnmarshal(t *testing.T) {
	jsonStr := `{
		"id": "order-abc",
		"workspace_id": "ws-def",
		"tiktok_id": "tiktok-ghi",
		"access_token": "token-jkl",
		"sync_type": "full"
	}`

	var order ImmediateWorkOrder
	if err := json.Unmarshal([]byte(jsonStr), &order); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if order.ID != "order-abc" {
		t.Fatalf("expected ID 'order-abc', got '%s'", order.ID)
	}
	if order.WorkspaceID != "ws-def" {
		t.Fatalf("expected WorkspaceID 'ws-def', got '%s'", order.WorkspaceID)
	}
	if order.TikTokID != "tiktok-ghi" {
		t.Fatalf("expected TikTokID 'tiktok-ghi', got '%s'", order.TikTokID)
	}
	if order.AccessToken != "token-jkl" {
		t.Fatalf("expected AccessToken 'token-jkl', got '%s'", order.AccessToken)
	}
	if order.SyncType != "full" {
		t.Fatalf("expected SyncType 'full', got '%s'", order.SyncType)
	}
}

func TestImmediateWorkOrder_EmptyFields(t *testing.T) {
	order := ImmediateWorkOrder{}

	if order.ID != "" {
		t.Errorf("expected empty ID, got %q", order.ID)
	}
	if order.WorkspaceID != "" {
		t.Errorf("expected empty WorkspaceID, got %q", order.WorkspaceID)
	}
	if order.TikTokID != "" {
		t.Errorf("expected empty TikTokID, got %q", order.TikTokID)
	}
	if order.AccessToken != "" {
		t.Errorf("expected empty AccessToken, got %q", order.AccessToken)
	}
	if order.SyncType != "" {
		t.Errorf("expected empty SyncType, got %q", order.SyncType)
	}
}

func TestImmediateWorkOrder_JSONKeys(t *testing.T) {
	order := ImmediateWorkOrder{
		ID:          "id123",
		TikTokID:    "tt789",
		AccessToken: "token",
		SyncType:    "full",
	}

	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"id", "workspace_id", "tiktok_id", "access_token", "sync_type"}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("missing key %q in JSON", key)
		}
	}
}

func TestImmediateWorkOrder_SyncTypes(t *testing.T) {
	syncTypes := []string{"full", "incremental", ""}

	for _, syncType := range syncTypes {
		order := ImmediateWorkOrder{
			SyncType: syncType,
		}

		data, err := json.Marshal(order)
		if err != nil {
			t.Fatalf("failed to marshal with sync_type %q: %v", syncType, err)
		}

		var decoded ImmediateWorkOrder
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal with sync_type %q: %v", syncType, err)
		}

		if decoded.SyncType != syncType {
			t.Errorf("sync_type = %q, want %q", decoded.SyncType, syncType)
		}
	}
}

// ================== SendPusherNotification Tests ==================

func TestSendPusherNotification_NilClient(t *testing.T) {
	log := logger.New("error")

	// Should not panic with nil client
	SendPusherNotification(nil, nil, "ws123", log)
}

func TestSendPusherNotification_Success(t *testing.T) {
	log := logger.New("error")
	pusher := &MockPusherNotifier{}

	account := &mongomodels.SocialIntegration{
		PlatformIdentifier: "tiktok-123",
	}

	SendPusherNotification(pusher, account, "ws456", log)

	if len(pusher.Triggered) != 1 {
		t.Fatalf("expected 1 trigger, got %d", len(pusher.Triggered))
	}

	expectedChannel := "tt-analytics-channel-ws456-tiktok-123"
	if pusher.Triggered[0].Channel != expectedChannel {
		t.Errorf("channel = %q, want %q", pusher.Triggered[0].Channel, expectedChannel)
	}

	expectedEvent := "syncing-ws456-tiktok-123"
	if pusher.Triggered[0].Event != expectedEvent {
		t.Errorf("event = %q, want %q", pusher.Triggered[0].Event, expectedEvent)
	}
}

func TestSendPusherNotification_EmptyPlatformIdentifier(t *testing.T) {
	log := logger.New("error")
	pusher := &MockPusherNotifier{}

	account := &mongomodels.SocialIntegration{
		PlatformIdentifier: "", // Empty should use "unknown"
	}

	SendPusherNotification(pusher, account, "ws456", log)

	if len(pusher.Triggered) != 1 {
		t.Fatalf("expected 1 trigger, got %d", len(pusher.Triggered))
	}

	expectedChannel := "tt-analytics-channel-ws456-unknown"
	if pusher.Triggered[0].Channel != expectedChannel {
		t.Errorf("channel = %q, want %q", pusher.Triggered[0].Channel, expectedChannel)
	}
}

func TestSendPusherNotification_TriggerError(t *testing.T) {
	log := logger.New("error")
	pusher := &MockPusherNotifier{
		TriggerFunc: func(channel, event string, data interface{}) error {
			return errors.New("pusher error")
		},
	}

	account := &mongomodels.SocialIntegration{
		PlatformIdentifier: "tiktok-123",
	}

	// Should not panic on error
	SendPusherNotification(pusher, account, "ws456", log)

	if len(pusher.Triggered) != 1 {
		t.Fatalf("expected 1 trigger attempt, got %d", len(pusher.Triggered))
	}
}

// ================== SendEmailNotification Tests ==================

func TestSendEmailNotification_NilNotifier(t *testing.T) {
	log := logger.New("error")

	// Should not panic with nil notifier
	SendEmailNotification(nil, "user123", "ws456", "tiktok789", "Test Account", log)
}

func TestSendEmailNotification_Success(t *testing.T) {
	log := logger.New("error")
	notifier := &MockEmailNotifier{}

	SendEmailNotification(notifier, "user123", "ws456", "tiktok789", "Test Account", log)

	if len(notifier.Sent) != 1 {
		t.Fatalf("expected 1 sent notification, got %d", len(notifier.Sent))
	}

	sent := notifier.Sent[0]
	if sent.UserID != "user123" {
		t.Errorf("UserID = %q, want %q", sent.UserID, "user123")
	}
	if sent.WorkspaceID != "ws456" {
		t.Errorf("WorkspaceID = %q, want %q", sent.WorkspaceID, "ws456")
	}
	if sent.Platform != "tiktok" {
		t.Errorf("Platform = %q, want %q", sent.Platform, "tiktok")
	}
	if sent.AccountID != "tiktok789" {
		t.Errorf("AccountID = %q, want %q", sent.AccountID, "tiktok789")
	}
	if sent.AccountName != "Test Account" {
		t.Errorf("AccountName = %q, want %q", sent.AccountName, "Test Account")
	}
	if sent.IsCompetitor != false {
		t.Errorf("IsCompetitor = %v, want %v", sent.IsCompetitor, false)
	}
}

func TestSendEmailNotification_Error(t *testing.T) {
	log := logger.New("error")
	notifier := &MockEmailNotifier{
		SendFunc: func(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error {
			return errors.New("email error")
		},
	}

	// Should not panic on error
	SendEmailNotification(notifier, "user123", "ws456", "tiktok789", "Test Account", log)

	if len(notifier.Sent) != 1 {
		t.Fatalf("expected 1 send attempt, got %d", len(notifier.Sent))
	}
}

// ================== ProcessAccount Tests ==================

func TestProcessAccount_NoVideos(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	fetcher := &MockTikTokVideoFetcher{
		FetchUserVideosFunc: func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
			return json.RawMessage(`[]`), 0, nil
		},
	}
	sink := &MockTikTokPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		ID:          "",
		TikTokID:    "tt-456",
		AccessToken: "token",
		SyncType:    "full",
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No posts to insert
	if len(sink.InsertedPosts) != 0 {
		t.Errorf("expected 0 inserted posts, got %d", len(sink.InsertedPosts))
	}
}

func TestProcessAccount_WithVideos(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	videosJSON := `[
		{"id": "video1", "desc": "Description 1", "create_time": 1609459200, "stats": {"digg_count": 10, "comment_count": 5, "share_count": 2, "play_count": 100}},
		{"id": "video2", "desc": "Description 2", "create_time": 1609545600, "stats": {"digg_count": 20, "comment_count": 10, "share_count": 5, "play_count": 200}}
	]`

	fetcher := &MockTikTokVideoFetcher{
		FetchUserVideosFunc: func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
			return json.RawMessage(videosJSON), 0, nil
		},
	}
	sink := &MockTikTokPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TikTokID:    "tt-456",
		AccessToken: "token",
		SyncType:    "full",
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sink.InsertedPosts) != 2 {
		t.Errorf("expected 2 inserted posts, got %d", len(sink.InsertedPosts))
	}
}

func TestProcessAccount_DateRangeFiltersAndStopsPaging(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	startDate := "2025-01-10"
	endDate := "2025-01-20"

	newVideoTime := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC).Unix()
	inRangeTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC).Unix()
	oldVideoTime := time.Date(2025, 1, 9, 23, 59, 0, 0, time.UTC).Unix()

	callCount := 0
	fetcher := &MockTikTokVideoFetcher{
		FetchUserVideosFunc: func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
			callCount++
			switch callCount {
			case 1:
				return json.RawMessage(fmt.Sprintf(`[
					{"id":"video-new","desc":"New video","create_time":%d,"stats":{"digg_count":1,"comment_count":1,"share_count":1,"play_count":1}},
					{"id":"video-keep","desc":"Keep video","create_time":%d,"stats":{"digg_count":2,"comment_count":2,"share_count":2,"play_count":2}}
				]`, newVideoTime, inRangeTime)), 123, nil
			case 2:
				return json.RawMessage(fmt.Sprintf(`[
					{"id":"video-old","desc":"Old video","create_time":%d,"stats":{"digg_count":3,"comment_count":3,"share_count":3,"play_count":3}}
				]`, oldVideoTime)), 0, nil
			default:
				return json.RawMessage(`[]`), 0, nil
			}
		},
	}
	sink := &MockTikTokPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TikTokID:    "tt-456",
		AccessToken: "token",
		SyncType:    "full",
		StartDate:   startDate,
		EndDate:     endDate,
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 2 {
		t.Fatalf("expected 2 fetch calls, got %d", callCount)
	}
	if len(sink.InsertedPosts) != 1 {
		t.Fatalf("expected 1 inserted post, got %d", len(sink.InsertedPosts))
	}
	if sink.InsertedPosts[0].PostID != "video-keep" {
		t.Fatalf("expected only the in-range video to be inserted, got %q", sink.InsertedPosts[0].PostID)
	}
}

func TestProcessAccount_FetchError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	fetcher := &MockTikTokVideoFetcher{
		FetchUserVideosFunc: func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
			return nil, 0, errors.New("fetch error")
		},
	}
	sink := &MockTikTokPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TikTokID:    "tt-456",
		AccessToken: "token",
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	// Error during fetch doesn't return error, just logs and continues
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessAccount_InvalidVideoJSON(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	fetcher := &MockTikTokVideoFetcher{
		FetchUserVideosFunc: func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
			return json.RawMessage(`invalid json`), 0, nil
		},
	}
	sink := &MockTikTokPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TikTokID:    "tt-456",
		AccessToken: "token",
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessAccount_SinkInsertError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	videosJSON := `[{"id": "video1", "desc": "Desc", "create_time": 1609459200, "stats": {"digg_count": 10, "comment_count": 5, "share_count": 2, "play_count": 100}}]`

	fetcher := &MockTikTokVideoFetcher{
		FetchUserVideosFunc: func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
			return json.RawMessage(videosJSON), 0, nil
		},
	}
	sink := &MockTikTokPostSink{
		BulkInsertTikTokPostsFunc: func(ctx context.Context, posts []*chmodels.TikTokPosts) error {
			return errors.New("insert error")
		},
	}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TikTokID:    "tt-456",
		AccessToken: "token",
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err == nil {
		t.Fatal("expected error for sink insert failure")
	}
}

func TestProcessAccount_WithAccountFromMongoDB(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	accountID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	workspaceID := primitive.NewObjectID()

	videosJSON := `[{"id": "video1", "desc": "Desc", "create_time": 1609459200, "stats": {"digg_count": 10, "comment_count": 5, "share_count": 2, "play_count": 100}}]`

	fetcher := &MockTikTokVideoFetcher{
		FetchUserVideosFunc: func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
			return json.RawMessage(videosJSON), 0, nil
		},
	}
	sink := &MockTikTokPostSink{}
	repo := &MockSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				UserID:             userID,
				WorkspaceID:        workspaceID,
				PlatformIdentifier: "tiktok-123",
				PlatformName:       "Test Account",
				State:              "Added",
			}, nil
		},
	}
	pusher := &MockPusherNotifier{}
	notifier := &MockEmailNotifier{}

	wo := ImmediateWorkOrder{
		ID:          accountID.Hex(),
		TikTokID:    "tt-456",
		AccessToken: "token",
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, notifier, pusher, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify pusher was called
	if len(pusher.Triggered) != 1 {
		t.Errorf("expected 1 pusher trigger, got %d", len(pusher.Triggered))
	}

	// Verify email was sent (state was "Added")
	if len(notifier.Sent) != 1 {
		t.Errorf("expected 1 email notification, got %d", len(notifier.Sent))
	}
}

func TestProcessAccount_WithAccountStateNotAdded(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	accountID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	workspaceID := primitive.NewObjectID()

	videosJSON := `[{"id": "video1", "desc": "Desc", "create_time": 1609459200, "stats": {"digg_count": 10, "comment_count": 5, "share_count": 2, "play_count": 100}}]`

	fetcher := &MockTikTokVideoFetcher{
		FetchUserVideosFunc: func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
			return json.RawMessage(videosJSON), 0, nil
		},
	}
	sink := &MockTikTokPostSink{}
	repo := &MockSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				UserID:             userID,
				WorkspaceID:        workspaceID,
				PlatformIdentifier: "tiktok-123",
				PlatformName:       "Test Account",
				State:              "Active", // Not "Added"
			}, nil
		},
	}
	pusher := &MockPusherNotifier{}
	notifier := &MockEmailNotifier{}

	wo := ImmediateWorkOrder{
		ID:          accountID.Hex(),
		TikTokID:    "tt-456",
		AccessToken: "token",
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, notifier, pusher, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify pusher was called
	if len(pusher.Triggered) != 1 {
		t.Errorf("expected 1 pusher trigger, got %d", len(pusher.Triggered))
	}

	// Email should NOT be sent (state is not "Added")
	if len(notifier.Sent) != 0 {
		t.Errorf("expected 0 email notifications for non-Added state, got %d", len(notifier.Sent))
	}
}

func TestProcessAccount_InvalidAccountID(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	fetcher := &MockTikTokVideoFetcher{
		FetchUserVideosFunc: func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
			return json.RawMessage(`[]`), 0, nil
		},
	}
	sink := &MockTikTokPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		ID:          "invalid-hex",
		TikTokID:    "tt-456",
		AccessToken: "token",
	}

	// Should not error, just log warning
	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessAccount_MongoFindError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	accountID := primitive.NewObjectID()

	fetcher := &MockTikTokVideoFetcher{
		FetchUserVideosFunc: func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
			return json.RawMessage(`[]`), 0, nil
		},
	}
	sink := &MockTikTokPostSink{}
	repo := &MockSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, errors.New("mongo error")
		},
	}

	wo := ImmediateWorkOrder{
		ID:          accountID.Hex(),
		TikTokID:    "tt-456",
		AccessToken: "token",
	}

	// Should not error, just log warning
	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessAccount_ClearsStaleProcessingErrorBeforeRetryAndOnSuccess(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	accountID := primitive.NewObjectID()
	repo := &MockSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:       accountID,
				MetaData: map[string]interface{}{"last_processing_error": "auth failed"},
			}, nil
		},
	}
	fetcher := &MockTikTokVideoFetcher{
		FetchUserVideosFunc: func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
			return nil, 0, errors.New("fetch failed")
		},
	}

	err := ProcessAccount(ctx, fetcher, &MockTikTokPostSink{}, repo, nil, nil, ImmediateWorkOrder{
		ID:          accountID.Hex(),
		TikTokID:    "tt-456",
		AccessToken: "token",
	}, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.ClearProcessingErrorIDs) != 2 {
		t.Fatalf("expected 2 processing-error clears, got %d", len(repo.ClearProcessingErrorIDs))
	}
	for _, clearedID := range repo.ClearProcessingErrorIDs {
		if clearedID != accountID {
			t.Fatalf("cleared account %s, want %s", clearedID.Hex(), accountID.Hex())
		}
	}
}

func TestProcessAccount_Pagination(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	callCount := 0
	fetcher := &MockTikTokVideoFetcher{
		FetchUserVideosFunc: func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
			callCount++
			if callCount == 1 {
				// First call - has more pages
				return json.RawMessage(`[{"id": "video1", "desc": "Desc", "create_time": 1609459200, "stats": {"digg_count": 10, "comment_count": 5, "share_count": 2, "play_count": 100}}]`), 12345, nil
			}
			// Second call - no more pages
			return json.RawMessage(`[]`), 0, nil
		},
	}
	sink := &MockTikTokPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TikTokID:    "tt-456",
		AccessToken: "token",
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount < 2 {
		t.Errorf("expected at least 2 fetch calls for pagination, got %d", callCount)
	}

	if len(sink.InsertedPosts) != 1 {
		t.Errorf("expected 1 inserted post, got %d", len(sink.InsertedPosts))
	}
}

// ================== Error-Flow Contract Tests (Caller logs Error with complete fields) ==================

func TestErrorFlowContract_ProcessAccountError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	err := errors.New("failed to fetch TikTok videos: API error")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("account_id", "acc123").
		Str("tiktok_id", "tt456").
		Str("workspace_id", "ws789").
		Str("function", "handler").
		Str("stage", "process_account").
		Msg("Failed to process TikTok account")

	output := buf.String()

	for _, field := range []string{"error_message", "function", "handler", "stage", "process_account", "account_id", "tiktok_id", "workspace_id"} {
		if !strings.Contains(output, field) {
			t.Errorf("missing %q in output: %s", field, output)
		}
	}
}

func TestErrorFlowContract_UnmarshalError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	err := errors.New("unexpected end of JSON input")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("function", "handler").
		Str("stage", "unmarshal_work_order").
		Msg("Failed to unmarshal work order")

	output := buf.String()
	for _, field := range []string{"error_message", "function", "handler", "stage", "unmarshal_work_order"} {
		if !strings.Contains(output, field) {
			t.Errorf("missing %q in output: %s", field, output)
		}
	}
}

func TestErrorFlowContract_ConsumerError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	err := errors.New("kafka broker not available")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("function", "main").
		Str("stage", "kafka_consume").
		Msg("Consumer error")

	output := buf.String()
	for _, field := range []string{"error_message", "function", "main", "stage", "kafka_consume"} {
		if !strings.Contains(output, field) {
			t.Errorf("missing %q in output: %s", field, output)
		}
	}
}

func TestErrorFlowContract_NotificationErrors_AreWarnNotError(t *testing.T) {
	hookRecords, cleanup := logger.InstallHookSpy()
	defer cleanup()

	log, buf := logger.NewTestLoggerWithHook()

	// Simulate notification error logged at Warn (not Error) — should NOT trigger Sentry hook
	err := errors.New("pusher connection refused")
	log.Warn().
		Err(err).
		Str("error_message", err.Error()).
		Str("channel", "tiktok-analytics-channel-ws1-tt1").
		Str("function", "SendPusherNotification").
		Msg("Failed to send Pusher notification")

	output := buf.String()
	if !strings.Contains(output, "WRN") {
		t.Errorf("expected WRN level, got: %s", output)
	}

	var errorCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.ErrorLevel {
			errorCount++
		}
	}
	if errorCount != 0 {
		t.Fatalf("notification Warn should not trigger Error-level hook, got %d", errorCount)
	}
}

func TestErrorFlowContract_ProcessAccountError_TriggersHook(t *testing.T) {
	hookRecords, cleanup := logger.InstallHookSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	err := errors.New("clickhouse insert failed")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("function", "handler").
		Str("stage", "process_account").
		Msg("Failed to process TikTok account")

	var errorCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.ErrorLevel {
			errorCount++
		}
	}
	if errorCount != 1 {
		t.Fatalf("expected exactly 1 Error-level hook firing, got %d", errorCount)
	}
}
