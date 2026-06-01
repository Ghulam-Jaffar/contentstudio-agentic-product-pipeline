package main

import (
	"context"
	"errors"
	"sync"
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
)

func TestConstants(t *testing.T) {
	if rawDataTopic != "raw-gmb-data" {
		t.Fatalf("expected rawDataTopic 'raw-gmb-data', got '%s'", rawDataTopic)
	}
	if maxWorkers != 10 {
		t.Fatalf("expected maxWorkers 10, got %d", maxWorkers)
	}
	if workChanSize != 200 {
		t.Fatalf("expected workChanSize 200, got %d", workChanSize)
	}
}

// ================== HandleWorkOrder Tests ==================

func TestHandleWorkOrder_InvalidObjectID(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()
	mockClient := &social.MockGMBClient{}
	producer := &kafka.MockProducer{}
	mongoRepo := &mongodb.MockUnifiedSocialRepository{}

	order := kafkamodels.GMBAccountWorkOrder{
		ID:         "not-a-valid-hex-id",
		AccountID:  "12345",
		LocationID: "67890",
	}

	err := HandleWorkOrder(ctx, order, mockClient, producer, mongoRepo, "", log)
	if err == nil {
		t.Fatal("expected error for invalid object ID")
	}
}

func TestHandleWorkOrder_AccountNotFound(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()
	mockClient := &social.MockGMBClient{}
	producer := &kafka.MockProducer{}

	objID := primitive.NewObjectID()
	mongoRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, nil
		},
	}

	order := kafkamodels.GMBAccountWorkOrder{
		ID:         objID.Hex(),
		AccountID:  "12345",
		LocationID: "67890",
	}

	err := HandleWorkOrder(ctx, order, mockClient, producer, mongoRepo, "", log)
	if err != nil {
		t.Fatalf("expected nil error for account not found (skipped), got: %v", err)
	}
}

func TestHandleWorkOrder_FindByIDError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()
	mockClient := &social.MockGMBClient{}
	producer := &kafka.MockProducer{}

	objID := primitive.NewObjectID()
	mongoRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, errors.New("database error")
		},
	}

	order := kafkamodels.GMBAccountWorkOrder{
		ID:         objID.Hex(),
		AccountID:  "12345",
		LocationID: "67890",
	}

	err := HandleWorkOrder(ctx, order, mockClient, producer, mongoRepo, "", log)
	if err == nil {
		t.Fatal("expected error for FindByID failure")
	}
}

func TestHandleWorkOrder_TokenRefreshError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	objID := primitive.NewObjectID()
	mockClient := &social.MockGMBClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error) {
			return nil, errors.New("token refresh failed")
		},
	}
	producer := &kafka.MockProducer{}
	mongoRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:           objID,
				AccessToken:  "old-token",
				RefreshToken: "refresh-token",
			}, nil
		},
	}

	order := kafkamodels.GMBAccountWorkOrder{
		ID:           objID.Hex(),
		AccountID:    "12345",
		LocationID:   "67890",
		RefreshToken: "refresh-token",
	}

	err := HandleWorkOrder(ctx, order, mockClient, producer, mongoRepo, "", log)
	if err == nil {
		t.Fatal("expected error for token refresh failure")
	}
}

func TestHandleWorkOrder_Success_WithVoM(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	objID := primitive.NewObjectID()
	var updateCalls []bson.M

	mockClient := &social.MockGMBClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error) {
			return &social.RefreshTokenResponse{AccessToken: "new-token"}, nil
		},
		FetchVoiceOfMerchantFunc: func(ctx context.Context, locationID, accessToken string) (*social.VoiceOfMerchantResponse, error) {
			return &social.VoiceOfMerchantResponse{HasVoiceOfMerchant: true}, nil
		},
		FetchPerformanceMetricsFunc: func(ctx context.Context, locationID, accessToken string, startDate, endDate time.Time) (*social.GMBPerformanceResponse, error) {
			return &social.GMBPerformanceResponse{}, nil
		},
		FetchSearchKeywordsFunc: func(ctx context.Context, locationID, accessToken string, startMonth, endMonth time.Time) (*social.GMBSearchKeywordsResponse, error) {
			return &social.GMBSearchKeywordsResponse{}, nil
		},
		FetchLocalPostsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBLocalPostsResponse, error) {
			return &social.GMBLocalPostsResponse{}, nil
		},
		FetchReviewsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBReviewsResponse, error) {
			return &social.GMBReviewsResponse{}, nil
		},
		FetchMediaAssetsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBMediaAssetsResponse, error) {
			return &social.GMBMediaAssetsResponse{}, nil
		},
	}

	var producedMessages []struct {
		Topic string
		Key   []byte
		Value []byte
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			producedMessages = append(producedMessages, struct {
				Topic string
				Key   []byte
				Value []byte
			}{topic, key, value})
			return nil
		},
	}

	mongoRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:           objID,
				AccessToken:  "old-token",
				RefreshToken: "refresh-token",
			}, nil
		},
		UpdateFunc: func(ctx context.Context, id primitive.ObjectID, updates bson.M) error {
			updateCalls = append(updateCalls, updates)
			return nil
		},
	}

	order := kafkamodels.GMBAccountWorkOrder{
		ID:           objID.Hex(),
		WorkspaceID:  "ws-123",
		AccountID:    "12345",
		LocationID:   "67890",
		RefreshToken: "refresh-token",
	}

	err := HandleWorkOrder(ctx, order, mockClient, producer, mongoRepo, "", log)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(producedMessages) < 5 {
		t.Fatalf("expected at least 5 produced messages (with VoM), got %d", len(producedMessages))
	}

	if len(updateCalls) < 1 {
		t.Fatalf("expected at least 1 MongoDB update call, got %d", len(updateCalls))
	}
}

func TestHandleWorkOrder_Success_WithoutVoM(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	objID := primitive.NewObjectID()
	perfMetricsCalled := false
	searchKeywordsCalled := false

	mockClient := &social.MockGMBClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error) {
			return &social.RefreshTokenResponse{AccessToken: "new-token"}, nil
		},
		FetchVoiceOfMerchantFunc: func(ctx context.Context, locationID, accessToken string) (*social.VoiceOfMerchantResponse, error) {
			return &social.VoiceOfMerchantResponse{HasVoiceOfMerchant: false}, nil
		},
		FetchPerformanceMetricsFunc: func(ctx context.Context, locationID, accessToken string, startDate, endDate time.Time) (*social.GMBPerformanceResponse, error) {
			perfMetricsCalled = true
			return &social.GMBPerformanceResponse{}, nil
		},
		FetchSearchKeywordsFunc: func(ctx context.Context, locationID, accessToken string, startMonth, endMonth time.Time) (*social.GMBSearchKeywordsResponse, error) {
			searchKeywordsCalled = true
			return &social.GMBSearchKeywordsResponse{}, nil
		},
		FetchLocalPostsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBLocalPostsResponse, error) {
			return &social.GMBLocalPostsResponse{}, nil
		},
		FetchReviewsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBReviewsResponse, error) {
			return &social.GMBReviewsResponse{}, nil
		},
		FetchMediaAssetsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBMediaAssetsResponse, error) {
			return &social.GMBMediaAssetsResponse{}, nil
		},
	}

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	mongoRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:           objID,
				AccessToken:  "old-token",
				RefreshToken: "refresh-token",
			}, nil
		},
		UpdateFunc: func(ctx context.Context, id primitive.ObjectID, updates bson.M) error {
			return nil
		},
	}

	order := kafkamodels.GMBAccountWorkOrder{
		ID:           objID.Hex(),
		WorkspaceID:  "ws-123",
		AccountID:    "12345",
		LocationID:   "67890",
		RefreshToken: "refresh-token",
	}

	err := HandleWorkOrder(ctx, order, mockClient, producer, mongoRepo, "", log)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if perfMetricsCalled {
		t.Fatal("expected performance metrics NOT to be called when VoM=false")
	}
	if searchKeywordsCalled {
		t.Fatal("expected search keywords NOT to be called when VoM=false")
	}
}

func TestHandleWorkOrder_VoMCheckError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	objID := primitive.NewObjectID()
	perfMetricsCalled := false

	mockClient := &social.MockGMBClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error) {
			return &social.RefreshTokenResponse{AccessToken: "new-token"}, nil
		},
		FetchVoiceOfMerchantFunc: func(ctx context.Context, locationID, accessToken string) (*social.VoiceOfMerchantResponse, error) {
			return nil, errors.New("VoM API error")
		},
		FetchPerformanceMetricsFunc: func(ctx context.Context, locationID, accessToken string, startDate, endDate time.Time) (*social.GMBPerformanceResponse, error) {
			perfMetricsCalled = true
			return &social.GMBPerformanceResponse{}, nil
		},
		FetchLocalPostsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBLocalPostsResponse, error) {
			return &social.GMBLocalPostsResponse{}, nil
		},
		FetchReviewsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBReviewsResponse, error) {
			return &social.GMBReviewsResponse{}, nil
		},
		FetchMediaAssetsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBMediaAssetsResponse, error) {
			return &social.GMBMediaAssetsResponse{}, nil
		},
	}

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	mongoRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:           objID,
				AccessToken:  "old-token",
				RefreshToken: "refresh-token",
			}, nil
		},
		UpdateFunc: func(ctx context.Context, id primitive.ObjectID, updates bson.M) error {
			return nil
		},
	}

	order := kafkamodels.GMBAccountWorkOrder{
		ID:           objID.Hex(),
		WorkspaceID:  "ws-123",
		AccountID:    "12345",
		LocationID:   "67890",
		RefreshToken: "refresh-token",
	}

	err := HandleWorkOrder(ctx, order, mockClient, producer, mongoRepo, "", log)
	if err != nil {
		t.Fatalf("expected no error (VoM failure is a warning), got: %v", err)
	}

	if perfMetricsCalled {
		t.Fatal("expected performance metrics NOT to be called when VoM fails")
	}
}

func TestSemForAccount(t *testing.T) {
	accountSemaphores = syncMapNew()

	sem1 := semForAccount("location-1", 1)
	sem2 := semForAccount("location-1", 1)

	if sem1 != sem2 {
		t.Fatal("expected same semaphore for same location ID")
	}

	sem3 := semForAccount("location-2", 1)
	if sem1 == sem3 {
		t.Fatal("expected different semaphore for different location ID")
	}
}

func syncMapNew() syncMap {
	return syncMap{}
}

type syncMap = sync.Map
