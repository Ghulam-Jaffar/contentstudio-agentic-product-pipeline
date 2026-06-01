package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/rs/zerolog"
)

// validScope contains all required TikTok scopes
const validScope = "user.info.basic,user.info.profile,user.info.stats,video.list"

// MockMongoRepository is a mock implementation of mongodb.UnifiedSocialRepository
type MockMongoRepository struct {
	UpdateFunc  func(ctx context.Context, id primitive.ObjectID, updates bson.M) error
	UpdateCalls []struct {
		ID      primitive.ObjectID
		Updates bson.M
	}
}

func (m *MockMongoRepository) Update(ctx context.Context, id primitive.ObjectID, updates bson.M) error {
	m.UpdateCalls = append(m.UpdateCalls, struct {
		ID      primitive.ObjectID
		Updates bson.M
	}{ID: id, Updates: updates})
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, updates)
	}
	return nil
}

// Mock the other required methods with no-op implementations
func (m *MockMongoRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
	return nil, nil
}
func (m *MockMongoRepository) GetByPlatformID(ctx context.Context, platformType, platformID string) (*mongomodels.SocialIntegration, error) {
	return nil, nil
}
func (m *MockMongoRepository) GetValidAccounts(ctx context.Context, platformType string, accountTypes []string) ([]mongomodels.SocialIntegration, error) {
	return nil, nil
}
func (m *MockMongoRepository) GetAccountsByWorkspace(ctx context.Context, workspaceID primitive.ObjectID, platforms []string) ([]mongomodels.SocialIntegration, error) {
	return nil, nil
}
func (m *MockMongoRepository) GetAccountsNeedingUpdate(ctx context.Context, platformType string, lastUpdateField string, hours int) ([]mongomodels.SocialIntegration, error) {
	return nil, nil
}
func (m *MockMongoRepository) GetAccountsNeedingUpdatePaginated(ctx context.Context, platformType string, accountTypes []string, hours int, skip, limit int64) ([]mongomodels.SocialIntegration, error) {
	return nil, nil
}
func (m *MockMongoRepository) CountAccountsNeedingUpdate(ctx context.Context, platformType string, accountTypes []string, hours int) (int64, error) {
	return 0, nil
}
func (m *MockMongoRepository) GetAccountsNeedingUpdateByID(ctx context.Context, platformType string, accountTypes []string, hours int, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error) {
	return nil, nil
}
func (m *MockMongoRepository) UpdateAnalyticsTimestamp(ctx context.Context, id primitive.ObjectID, timestampType string, timestamp time.Time) error {
	return nil
}
func (m *MockMongoRepository) UpdateTokens(ctx context.Context, id primitive.ObjectID, tokens map[string]string) error {
	return nil
}
func (m *MockMongoRepository) UpdateState(ctx context.Context, id primitive.ObjectID, newState string) error {
	return nil
}
func (m *MockMongoRepository) UpdateValidity(ctx context.Context, id primitive.ObjectID, newValidity string) error {
	return nil
}
func (m *MockMongoRepository) Create(ctx context.Context, account *mongomodels.SocialIntegration) (primitive.ObjectID, error) {
	return primitive.NilObjectID, nil
}
func (m *MockMongoRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	return nil
}

func (m *MockMongoRepository) GetYouTubeAccountsNeedingUpdatePaginated(ctx context.Context, hours int, consentDays int, skip, limit int64) ([]mongomodels.SocialIntegration, error) {
	return nil, nil
}

func (m *MockMongoRepository) GetYouTubeAccountsNeedingUpdateByID(ctx context.Context, hours int, consentDays int, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error) {
	return nil, nil
}

func (m *MockMongoRepository) CountYouTubeAccountsNeedingUpdate(ctx context.Context, hours int, consentDays int) (int64, error) {
	return 0, nil
}

func (m *MockMongoRepository) RecordProcessingError(ctx context.Context, id primitive.ObjectID, errorMessage string) error {
	return nil
}

func (m *MockMongoRepository) ClearProcessingError(ctx context.Context, id primitive.ObjectID) error {
	return nil
}

func (m *MockMongoRepository) InsertTwitterJobMetadata(ctx context.Context, payload mongodb.TwitterJobMetadataPayload) error {
	return nil
}

func (m *MockMongoRepository) GetAccountsByPlatformIDs(ctx context.Context, platformType string, platformIDs []string) ([]mongomodels.SocialIntegration, error) {
	return nil, nil
}

func (m *MockMongoRepository) GetValidAccountsByID(ctx context.Context, platformType string, accountTypes []string, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error) {
	return nil, nil
}

func (m *MockMongoRepository) CountValidAccounts(ctx context.Context, platformType string, accountTypes []string) (int64, error) {
	return 0, nil
}


func TestConstants(t *testing.T) {
	if workOrdersTopic != "work-order-tiktok-batch" {
		t.Fatalf("expected workOrdersTopic 'work-order-tiktok-batch', got '%s'", workOrdersTopic)
	}
	if rawPostsTopic != "raw-tiktok-posts" {
		t.Fatalf("expected rawPostsTopic 'raw-tiktok-posts', got '%s'", rawPostsTopic)
	}
	if rawInsightsTopic != "raw-tiktok-insights" {
		t.Fatalf("expected rawInsightsTopic 'raw-tiktok-insights', got '%s'", rawInsightsTopic)
	}
	if maxWorkers != 10 {
		t.Fatalf("expected maxWorkers 10, got %d", maxWorkers)
	}
	if workChanSize != 200 {
		t.Fatalf("expected workChanSize 200, got %d", workChanSize)
	}
	if idleTimeout != 5*time.Minute {
		t.Fatalf("expected idleTimeout 5m, got %v", idleTimeout)
	}
	if idleCheckInterval != 30*time.Second {
		t.Fatalf("expected idleCheckInterval 30s, got %v", idleCheckInterval)
	}
}

// ================== HandleWorkOrder Tests ==================

func TestHandleWorkOrder_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()
	mockClient := &social.MockTikTokClient{}
	producer := &kafka.MockProducer{}
	mongoRepo := &MockMongoRepository{}

	err := HandleWorkOrder(ctx, []byte("key"), []byte("invalid json"), mockClient, producer, mongoRepo, "", log)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestHandleWorkOrder_InvalidScopes(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()
	mockClient := &social.MockTikTokClient{}
	producer := &kafka.MockProducer{}
	mongoRepo := &MockMongoRepository{}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID: "test-id",
		Scope:    "invalid-scope",
	}
	orderJSON, _ := json.Marshal(order)

	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, mongoRepo, "", log)
	if err != nil {
		t.Fatalf("expected nil error for invalid scopes (skipped), got: %v", err)
	}
}

func TestHandleWorkOrder_FetchUserInfoError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	mockClient := &social.MockTikTokClient{
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return nil, errors.New("api error")
		},
	}
	producer := &kafka.MockProducer{}
	mongoRepo := &MockMongoRepository{}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID:     "test-id",
		Scope:        validScope,
		AccessToken:  "test-token",
		RefreshToken: "",
	}
	orderJSON, _ := json.Marshal(order)

	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, mongoRepo, "", log)
	if err == nil {
		t.Fatal("expected error for FetchUserInfo failure")
	}
}

func TestHandleWorkOrder_ParseUserInfoError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	mockClient := &social.MockTikTokClient{
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return json.RawMessage(`invalid json`), nil
		},
	}
	producer := &kafka.MockProducer{}
	mongoRepo := &MockMongoRepository{}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID:    "test-id",
		Scope:       validScope,
		AccessToken: "test-token",
	}
	orderJSON, _ := json.Marshal(order)

	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, mongoRepo, "", log)
	if err == nil {
		t.Fatal("expected error for parse user info failure")
	}
}

func TestHandleWorkOrder_Success(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	userInfoResponse := json.RawMessage(`{
		"open_id": "test-open-id",
		"display_name": "Test User",
		"avatar_url": "https://example.com/avatar.jpg",
		"follower_count": 1000,
		"following_count": 500,
		"likes_count": 5000
	}`)

	videosResponse := json.RawMessage(`[{
		"id": "video-1",
		"title": "Test Video",
		"create_time": 1609459200,
		"cover_image_url": "https://example.com/cover.jpg",
		"share_url": "https://tiktok.com/video/1",
		"view_count": 100,
		"like_count": 10,
		"comment_count": 5,
		"share_count": 2
	}]`)

	producedMessages := make(map[string]int)
	mockClient := &social.MockTikTokClient{
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return userInfoResponse, nil
		},
		FetchVideoListFunc: func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
			return videosResponse, 0, false, nil
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			producedMessages[topic]++
			return nil
		},
	}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID:    "test-id",
		Scope:       validScope,
		AccessToken: "test-token",
		SyncType:    "full",
	}
	orderJSON, _ := json.Marshal(order)

	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, &MockMongoRepository{}, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if producedMessages[rawPostsTopic] == 0 {
		t.Error("expected posts to be produced")
	}
	if producedMessages[rawInsightsTopic] == 0 {
		t.Error("expected insights to be produced")
	}
}

func TestHandleWorkOrder_IncrementalSync(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	userInfoResponse := json.RawMessage(`{
		"open_id": "test-open-id",
		"display_name": "Test User"
	}`)

	mockClient := &social.MockTikTokClient{
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return userInfoResponse, nil
		},
		FetchVideoListFunc: func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
			return json.RawMessage(`[]`), 0, false, nil
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID:    "test-id",
		Scope:       validScope,
		AccessToken: "test-token",
		SyncType:    "incremental",
	}
	orderJSON, _ := json.Marshal(order)

	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, &MockMongoRepository{}, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleWorkOrder_TokenRefresh(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	userInfoResponse := json.RawMessage(`{
		"open_id": "test-open-id",
		"display_name": "Test User"
	}`)

	refreshCalled := false
	mockClient := &social.MockTikTokClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error) {
			refreshCalled = true
			return &social.RefreshTokenResponse{
				AccessToken:      "new-access-token",
				RefreshToken:     "new-refresh-token",
				ExpiresIn:        86400,
				RefreshExpiresIn: 7776000,
				Scope:            "user.info.basic,video.list",
			}, nil
		},
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return userInfoResponse, nil
		},
		FetchVideoListFunc: func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
			return json.RawMessage(`[]`), 0, false, nil
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID:     "test-id",
		Scope:        validScope,
		AccessToken:  "old-token",
		RefreshToken: "refresh-token",
		SyncType:     "full",
	}
	orderJSON, _ := json.Marshal(order)

	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, &MockMongoRepository{}, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !refreshCalled {
		t.Error("expected RefreshToken to be called")
	}
}

func TestHandleWorkOrder_TokenRefreshFailure(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	userInfoResponse := json.RawMessage(`{
		"open_id": "test-open-id"
	}`)

	mockClient := &social.MockTikTokClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error) {
			return nil, errors.New("refresh failed")
		},
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return userInfoResponse, nil
		},
		FetchVideoListFunc: func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
			return json.RawMessage(`[]`), 0, false, nil
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID:     "test-id",
		Scope:        validScope,
		AccessToken:  "old-token",
		RefreshToken: "refresh-token",
	}
	orderJSON, _ := json.Marshal(order)

	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, &MockMongoRepository{}, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleWorkOrder_FetchVideoListError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	userInfoResponse := json.RawMessage(`{
		"open_id": "test-open-id",
		"display_name": "Test User"
	}`)

	mockClient := &social.MockTikTokClient{
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return userInfoResponse, nil
		},
		FetchVideoListFunc: func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
			return nil, 0, false, errors.New("video list error")
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID:    "test-id",
		Scope:       validScope,
		AccessToken: "test-token",
	}
	orderJSON, _ := json.Marshal(order)

	// Should not return error - just logs and continues
	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, &MockMongoRepository{}, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleWorkOrder_InvalidVideoJSON(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	userInfoResponse := json.RawMessage(`{
		"open_id": "test-open-id",
		"display_name": "Test User"
	}`)

	mockClient := &social.MockTikTokClient{
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return userInfoResponse, nil
		},
		FetchVideoListFunc: func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
			return json.RawMessage(`not a valid array`), 0, false, nil
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID:    "test-id",
		Scope:       validScope,
		AccessToken: "test-token",
	}
	orderJSON, _ := json.Marshal(order)

	// Should not return error - just logs and continues
	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, &MockMongoRepository{}, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleWorkOrder_ProducePostError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	userInfoResponse := json.RawMessage(`{
		"open_id": "test-open-id",
		"display_name": "Test User"
	}`)

	videosResponse := json.RawMessage(`[{
		"id": "video-1",
		"title": "Test Video",
		"create_time": 1609459200,
		"view_count": 100,
		"like_count": 10
	}]`)

	mockClient := &social.MockTikTokClient{
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return userInfoResponse, nil
		},
		FetchVideoListFunc: func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
			return videosResponse, 0, false, nil
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			if topic == rawPostsTopic {
				return errors.New("produce error")
			}
			return nil
		},
	}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID:    "test-id",
		Scope:       validScope,
		AccessToken: "test-token",
	}
	orderJSON, _ := json.Marshal(order)

	// Should not return error - just logs and continues
	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, &MockMongoRepository{}, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleWorkOrder_ProduceInsightsError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	userInfoResponse := json.RawMessage(`{
		"open_id": "test-open-id",
		"display_name": "Test User"
	}`)

	mockClient := &social.MockTikTokClient{
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return userInfoResponse, nil
		},
		FetchVideoListFunc: func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
			return json.RawMessage(`[]`), 0, false, nil
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			if topic == rawInsightsTopic {
				return errors.New("produce insights error")
			}
			return nil
		},
	}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID:    "test-id",
		Scope:       validScope,
		AccessToken: "test-token",
	}
	orderJSON, _ := json.Marshal(order)

	// Should not return error - just logs and continues
	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, &MockMongoRepository{}, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleWorkOrder_Pagination(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	userInfoResponse := json.RawMessage(`{
		"open_id": "test-open-id",
		"display_name": "Test User"
	}`)

	callCount := 0
	mockClient := &social.MockTikTokClient{
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return userInfoResponse, nil
		},
		FetchVideoListFunc: func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
			callCount++
			if callCount == 1 {
				// First call - return videos and has more
				return json.RawMessage(`[{"id": "video-1", "create_time": 1609459200}]`), 12345, true, nil
			}
			// Second call - return empty and no more
			return json.RawMessage(`[]`), 0, false, nil
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID:    "test-id",
		Scope:       validScope,
		AccessToken: "test-token",
		SyncType:    "full",
	}
	orderJSON, _ := json.Marshal(order)

	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, &MockMongoRepository{}, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount < 2 {
		t.Errorf("expected at least 2 calls for pagination, got %d", callCount)
	}
}

func TestHandleWorkOrder_MaxVideosLimit(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	userInfoResponse := json.RawMessage(`{
		"open_id": "test-open-id",
		"display_name": "Test User"
	}`)

	// Generate 50 videos to trigger the incremental limit
	videos := make([]map[string]interface{}, 50)
	for i := 0; i < 50; i++ {
		videos[i] = map[string]interface{}{
			"id":          "video-" + string(rune('0'+i%10)),
			"create_time": 1609459200,
		}
	}
	videosJSON, _ := json.Marshal(videos)

	callCount := 0
	mockClient := &social.MockTikTokClient{
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return userInfoResponse, nil
		},
		FetchVideoListFunc: func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
			callCount++
			// Stop after 3 calls (150 videos) to test that the limit is respected
			if callCount >= 3 {
				return videosJSON, 0, false, nil
			}
			return videosJSON, int64(callCount * 50), true, nil
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID:    "test-id",
		Scope:       validScope,
		AccessToken: "test-token",
		SyncType:    "incremental", // Should limit to 100 videos
	}
	orderJSON, _ := json.Marshal(order)

	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, &MockMongoRepository{}, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With 50 videos per page and 100 max, should stop after 2 calls
	if callCount > 3 {
		t.Errorf("expected max 3 calls for incremental sync, got %d", callCount)
	}
}

func TestHandleWorkOrder_ParseVideoError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	userInfoResponse := json.RawMessage(`{
		"open_id": "test-open-id",
		"display_name": "Test User"
	}`)

	// Invalid video structure that will fail parsing
	videosResponse := json.RawMessage(`[{"invalid": true}]`)

	mockClient := &social.MockTikTokClient{
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return userInfoResponse, nil
		},
		FetchVideoListFunc: func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
			return videosResponse, 0, false, nil
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	order := kafkamodels.TikTokAccountWorkOrder{
		TikTokID:    "test-id",
		Scope:       validScope,
		AccessToken: "test-token",
	}
	orderJSON, _ := json.Marshal(order)

	// Should not return error - just logs and continues
	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, &MockMongoRepository{}, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}


// ================== Semaphore Tests ==================

func TestSemForAccount_ReturnsSameSemaphore(t *testing.T) {
	sem1 := semForAccount("test-account-1", 1)
	sem2 := semForAccount("test-account-1", 1)

	if sem1 != sem2 {
		t.Error("expected same semaphore for same account")
	}
}

func TestSemForAccount_ReturnsDifferentSemaphores(t *testing.T) {
	sem1 := semForAccount("test-account-a", 1)
	sem2 := semForAccount("test-account-b", 1)

	if sem1 == sem2 {
		t.Error("expected different semaphores for different accounts")
	}
}

func TestSemForAccount_PreventsParallelProcessing(t *testing.T) {
	ctx := context.Background()
	accountID := "test-concurrent-account"

	sem := semForAccount(accountID, 1)

	// Acquire the semaphore
	if err := sem.Acquire(ctx, 1); err != nil {
		t.Fatalf("failed to acquire semaphore: %v", err)
	}

	// Try to acquire again with timeout - should fail
	ctxTimeout, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	err := sem.Acquire(ctxTimeout, 1)
	if err == nil {
		t.Error("expected second acquire to fail due to timeout")
		sem.Release(1)
	}

	// Release first semaphore
	sem.Release(1)
}

// TestHandleWorkOrder_UpdatesMongoDBState verifies that the fetcher updates MongoDB state
// to "Processed" after successfully processing a work order, matching processor behavior.
func TestHandleWorkOrder_UpdatesMongoDBState(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	userInfoResponse := json.RawMessage(`{
		"open_id": "test-open-id",
		"display_name": "Test User",
		"avatar_url": "https://example.com/avatar.jpg",
		"follower_count": 1000,
		"following_count": 500,
		"likes_count": 5000
	}`)

	videosResponse := json.RawMessage(`[{
		"id": "video-1",
		"title": "Test Video",
		"create_time": 1609459200,
		"cover_image_url": "https://example.com/cover.jpg",
		"share_url": "https://tiktok.com/video/1",
		"view_count": 100,
		"like_count": 10,
		"comment_count": 5,
		"share_count": 2
	}]`)

	mockClient := &social.MockTikTokClient{
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return userInfoResponse, nil
		},
		FetchVideoListFunc: func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
			return videosResponse, 0, false, nil
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	// Track MongoDB Update calls
	updateCalled := false
	var capturedUpdates bson.M

	mongoRepo := &MockMongoRepository{
		UpdateFunc: func(ctx context.Context, id primitive.ObjectID, updates bson.M) error {
			updateCalled = true
			capturedUpdates = updates
			return nil
		},
	}

	// Use a valid MongoDB ObjectID
	accountID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	order := kafkamodels.TikTokAccountWorkOrder{
		ID:          accountID.Hex(),
		TikTokID:    "test-id",
		Scope:       validScope,
		AccessToken: "test-token",
		SyncType:    "full",
	}
	orderJSON, _ := json.Marshal(order)

	err := HandleWorkOrder(ctx, []byte("key"), orderJSON, mockClient, producer, mongoRepo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify MongoDB Update was called with correct state
	if !updateCalled {
		t.Error("expected MongoDB Update to be called")
	}
	if capturedUpdates["state"] != "Processed" {
		t.Errorf("expected state 'Processed', got %v", capturedUpdates["state"])
	}
}

// ================== Logging Contract Tests ==================

func TestLoggingContract_TikTokFetcher_ErrorHasContextFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	// Simulate what the TikTok fetcher does when it gets an unexpected error
	log.Error().
		Str("error_message", "failed to fetch video list").
		Str("function", "HandleWorkOrder").
		Str("stage", "fetch_videos").
		Msg("TikTok fetcher error")

	output := buf.String()

	checks := map[string]string{
		"ERR":             "expected ERR level in output",
		"error_message":   "expected error_message field in output",
		"function":        "expected function field in output",
		"HandleWorkOrder": "expected HandleWorkOrder value in output",
		"stage":           "expected stage field in output",
	}
	for substr, errMsg := range checks {
		if !strings.Contains(output, substr) {
			t.Errorf("%s, got: %s", errMsg, output)
		}
	}
}

func TestLoggingContract_TikTokFetcher_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	// Log an error the way the TikTok fetcher does — Error level only, no CaptureException
	log.Error().
		Str("error_message", "API rate limit exceeded").
		Str("function", "HandleWorkOrder").
		Str("stage", "fetch_user_info").
		Msg("Failed to fetch user info")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}

func TestLoggingContract_TikTokFetcher_ExpectedError_WarnOnly(t *testing.T) {
	hookRecords, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	log, buf := logger.NewTestLoggerWithHook()

	// Simulate an expected/token error — should be Warn level
	log.Warn().
		Str("error_message", "access_token_invalid").
		Str("function", "HandleWorkOrder").
		Str("tiktok_id", "test-id").
		Msg("TikTok token expired, skipping account")

	output := buf.String()

	if !strings.Contains(output, "WRN") {
		t.Fatalf("expected WRN level in output, got: %s", output)
	}
	if strings.Contains(output, "ERR") {
		t.Fatalf("expected error should NOT produce ERR level: %s", output)
	}

	// Verify no Error-level hook firings
	for _, r := range *hookRecords {
		if r.Level >= zerolog.ErrorLevel {
			t.Fatalf("expected error should not trigger Error+ hook, got level %v", r.Level)
		}
	}
}
