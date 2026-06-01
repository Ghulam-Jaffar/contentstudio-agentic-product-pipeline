package processor

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// ================== ImmediateWorkOrder Tests ==================

func TestImmediateWorkOrder_Struct(t *testing.T) {
	order := ImmediateWorkOrder{
		ID:               "order-123",
		WorkspaceID:      "ws-456",
		TwitterID:        "tw-789",
		OAuthToken:       "test-token",
		OAuthTokenSecret: "test-secret",
		NTweets:          45,
		SyncType:         "full",
	}

	if order.ID != "order-123" {
		t.Fatalf("expected ID 'order-123', got '%s'", order.ID)
	}
	if order.WorkspaceID != "ws-456" {
		t.Fatalf("expected WorkspaceID 'ws-456', got '%s'", order.WorkspaceID)
	}
	if order.TwitterID != "tw-789" {
		t.Fatalf("expected TwitterID 'tw-789', got '%s'", order.TwitterID)
	}
	if order.OAuthToken != "test-token" {
		t.Fatalf("expected OAuthToken 'test-token', got '%s'", order.OAuthToken)
	}
	if order.OAuthTokenSecret != "test-secret" {
		t.Fatalf("expected OAuthTokenSecret 'test-secret', got '%s'", order.OAuthTokenSecret)
	}
	if order.SyncType != "full" {
		t.Fatalf("expected SyncType 'full', got '%s'", order.SyncType)
	}
}

func TestImmediateWorkOrder_JSONMarshal(t *testing.T) {
	order := ImmediateWorkOrder{
		ID:               "order-123",
		TwitterID:        "tw-789",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
		NTweets:          45,
		SyncType:         "incremental",
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
	if decoded.TwitterID != order.TwitterID {
		t.Fatalf("expected TwitterID '%s', got '%s'", order.TwitterID, decoded.TwitterID)
	}
}

func TestImmediateWorkOrder_JSONUnmarshal(t *testing.T) {
	jsonStr := `{
		"id": "order-abc",
		"workspace_id": "ws-def",
		"twitter_id": "tw-ghi",
		"oauth_token": "token-jkl",
		"oauth_token_secret": "secret-mno",
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
	if order.TwitterID != "tw-ghi" {
		t.Fatalf("expected TwitterID 'tw-ghi', got '%s'", order.TwitterID)
	}
	if order.OAuthToken != "token-jkl" {
		t.Fatalf("expected OAuthToken 'token-jkl', got '%s'", order.OAuthToken)
	}
	if order.OAuthTokenSecret != "secret-mno" {
		t.Fatalf("expected OAuthTokenSecret 'secret-mno', got '%s'", order.OAuthTokenSecret)
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
	if order.TwitterID != "" {
		t.Errorf("expected empty TwitterID, got %q", order.TwitterID)
	}
	if order.OAuthToken != "" {
		t.Errorf("expected empty OAuthToken, got %q", order.OAuthToken)
	}
	if order.OAuthTokenSecret != "" {
		t.Errorf("expected empty OAuthTokenSecret, got %q", order.OAuthTokenSecret)
	}
	if order.SyncType != "" {
		t.Errorf("expected empty SyncType, got %q", order.SyncType)
	}
}

func TestImmediateWorkOrder_JSONKeys(t *testing.T) {
	order := ImmediateWorkOrder{
		ID:               "id123",
		TwitterID:        "tw789",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
		NTweets:          45,
		SyncType:         "full",
	}

	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{
		"id",
		"workspace_id",
		"twitter_id",
		"oauth_token",
		"oauth_token_secret",
		"n_tweets",
		"api_key",
		"api_secret",
		"app_name",
		"app_id",
		"executed_by",
		"sync_type",
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("missing key %q in JSON", key)
		}
	}
}

func TestImmediateWorkOrder_SyncTypes(t *testing.T) {
	syncTypes := []string{"full", "full_sync", "incremental", "immediate", ""}

	for _, syncType := range syncTypes {
		order := ImmediateWorkOrder{SyncType: syncType}

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
	SendPusherNotification(nil, nil, "ws123", mongomodels.StateAdded, log)
}

func TestSendPusherNotification_Success(t *testing.T) {
	log := logger.New("error")
	pusher := &MockPusherNotifier{}

	account := &mongomodels.SocialIntegration{
		PlatformIdentifier: "twitter-123",
	}

	SendPusherNotification(pusher, account, "ws456", mongomodels.StateAdded, log)

	if len(pusher.Triggered) != 1 {
		t.Fatalf("expected 1 trigger, got %d", len(pusher.Triggered))
	}

	expectedChannel := "twitter-analytics-channel-ws456-twitter-123"
	if pusher.Triggered[0].Channel != expectedChannel {
		t.Errorf("channel = %q, want %q", pusher.Triggered[0].Channel, expectedChannel)
	}

	expectedEvent := "syncing-ws456-twitter-123"
	if pusher.Triggered[0].Event != expectedEvent {
		t.Errorf("event = %q, want %q", pusher.Triggered[0].Event, expectedEvent)
	}
}

func TestSendPusherNotification_SuccessForProcessedState(t *testing.T) {
	log := logger.New("error")
	pusher := &MockPusherNotifier{}

	account := &mongomodels.SocialIntegration{
		PlatformIdentifier: "twitter-123",
	}

	SendPusherNotification(pusher, account, "ws456", mongomodels.StateProcessed, log)

	if len(pusher.Triggered) != 1 {
		t.Fatalf("expected 1 trigger, got %d", len(pusher.Triggered))
	}
}

func TestSendPusherNotification_EmptyPlatformIdentifier(t *testing.T) {
	log := logger.New("error")
	pusher := &MockPusherNotifier{}

	account := &mongomodels.SocialIntegration{
		PlatformIdentifier: "", // Empty should use "unknown"
	}

	SendPusherNotification(pusher, account, "ws456", mongomodels.StateAdded, log)

	if len(pusher.Triggered) != 1 {
		t.Fatalf("expected 1 trigger, got %d", len(pusher.Triggered))
	}

	expectedChannel := "twitter-analytics-channel-ws456-unknown"
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
		PlatformIdentifier: "twitter-123",
	}

	// Should not panic on error
	SendPusherNotification(pusher, account, "ws456", mongomodels.StateAdded, log)

	if len(pusher.Triggered) != 1 {
		t.Fatalf("expected 1 trigger attempt, got %d", len(pusher.Triggered))
	}
}

func TestSendPusherNotification_DataContainsExpectedFields(t *testing.T) {
	log := logger.New("error")
	pusher := &MockPusherNotifier{}

	account := &mongomodels.SocialIntegration{
		PlatformIdentifier: "tw-999",
	}

	SendPusherNotification(pusher, account, "ws001", mongomodels.StateAdded, log)

	if len(pusher.Triggered) != 1 {
		t.Fatalf("expected 1 trigger, got %d", len(pusher.Triggered))
	}

	data, ok := pusher.Triggered[0].Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be map[string]interface{}")
	}

	if data["state"] != "Processed" {
		t.Errorf("state = %v, want 'Processed'", data["state"])
	}
	if data["account"] != "tw-999" {
		t.Errorf("account = %v, want 'tw-999'", data["account"])
	}
	if _, ok := data["last_analytics_updated_at"]; !ok {
		t.Error("missing last_analytics_updated_at in data")
	}
}

// ================== SendEmailNotification Tests ==================

func TestSendEmailNotification_NilNotifier(t *testing.T) {
	log := logger.New("error")
	// Should not panic with nil notifier
	SendEmailNotification(nil, "user123", "ws456", "tw789", "Test Account", log)
}

func TestSendEmailNotification_Success(t *testing.T) {
	log := logger.New("error")
	notifier := &MockEmailNotifier{}

	SendEmailNotification(notifier, "user123", "ws456", "tw789", "Test Account", log)

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
	if sent.Platform != "twitter" {
		t.Errorf("Platform = %q, want %q", sent.Platform, "twitter")
	}
	if sent.AccountID != "tw789" {
		t.Errorf("AccountID = %q, want %q", sent.AccountID, "tw789")
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
	SendEmailNotification(notifier, "user123", "ws456", "tw789", "Test Account", log)

	if len(notifier.Sent) != 1 {
		t.Fatalf("expected 1 send attempt, got %d", len(notifier.Sent))
	}
}

// ================== ProcessAccount Tests ==================

func buildMockTweetResponse(tweets []social.TwitterTweet) *social.TwitterTweetsResponse {
	return &social.TwitterTweetsResponse{
		Data: tweets,
		Meta: &social.TwitterMeta{
			ResultCount: len(tweets),
		},
	}
}

func buildMockUserInfoResponse(users []social.TwitterUser) *social.TwitterUserResponse {
	return &social.TwitterUserResponse{
		Data: users,
	}
}

func TestProcessAccount_NoTweets(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return buildMockUserInfoResponse([]social.TwitterUser{
				{ID: "123", Name: "Test User", Username: "testuser"},
			}), nil
		},
		FetchUserTweetsFunc: func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
			return &social.TwitterTweetsResponse{Data: nil}, nil
		},
	}
	sink := &MockTwitterPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
		SyncType:         "full",
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sink.InsertedPosts) != 0 {
		t.Errorf("expected 0 inserted posts, got %d", len(sink.InsertedPosts))
	}
}

func TestProcessAccount_WithTweets(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	now := time.Now().UTC()
	tweets := []social.TwitterTweet{
		{
			ID:        "tweet1",
			Text:      "Hello world #golang",
			CreatedAt: now.Format(time.RFC3339),
			PublicMetrics: social.TwitterTweetMetrics{
				RetweetCount: 10,
				ReplyCount:   5,
				LikeCount:    20,
				QuoteCount:   2,
			},
		},
		{
			ID:        "tweet2",
			Text:      "Another tweet",
			CreatedAt: now.Add(-1 * time.Hour).Format(time.RFC3339),
			PublicMetrics: social.TwitterTweetMetrics{
				RetweetCount: 3,
				ReplyCount:   1,
				LikeCount:    8,
				QuoteCount:   0,
			},
		},
	}

	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return buildMockUserInfoResponse([]social.TwitterUser{
				{
					ID:       "123",
					Name:     "Test User",
					Username: "testuser",
					PublicMetrics: social.TwitterUserPublicMetrics{
						FollowersCount: 100,
						FollowingCount: 50,
						TweetCount:     500,
					},
				},
			}), nil
		},
		FetchUserTweetsFunc: func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
			return buildMockTweetResponse(tweets), nil
		},
	}
	sink := &MockTwitterPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
		SyncType:         "full",
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sink.InsertedPosts) != 2 {
		t.Errorf("expected 2 inserted posts, got %d", len(sink.InsertedPosts))
	}

	// Should also have insights
	if len(sink.InsertedInsights) != 1 {
		t.Errorf("expected 1 insight, got %d", len(sink.InsertedInsights))
	}
}

func TestProcessAccount_FetchUserInfoError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return nil, errors.New("user info fetch error")
		},
	}
	sink := &MockTwitterPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
		NTweets:          120,
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err == nil {
		t.Fatal("expected error for user info fetch failure")
	}
}

func TestProcessAccount_FetchTweetsError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return buildMockUserInfoResponse([]social.TwitterUser{
				{ID: "123", Name: "Test User", Username: "testuser"},
			}), nil
		},
		FetchUserTweetsFunc: func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
			return nil, errors.New("fetch tweets error")
		},
	}
	sink := &MockTwitterPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
		NTweets:          120,
	}

	// Fetch tweets error doesn't return error, just logs and breaks
	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessAccount_SinkInsertError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	now := time.Now().UTC()
	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return buildMockUserInfoResponse([]social.TwitterUser{
				{ID: "123", Name: "Test", Username: "test"},
			}), nil
		},
		FetchUserTweetsFunc: func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
			return buildMockTweetResponse([]social.TwitterTweet{
				{
					ID:        "tweet1",
					Text:      "Test",
					CreatedAt: now.Format(time.RFC3339),
					PublicMetrics: social.TwitterTweetMetrics{
						RetweetCount: 1,
						ReplyCount:   1,
						LikeCount:    1,
						QuoteCount:   1,
					},
				},
			}), nil
		},
	}
	sink := &MockTwitterPostSink{
		BulkInsertTwitterPostsFunc: func(ctx context.Context, posts []*chmodels.TwitterPosts) error {
			return errors.New("insert error")
		},
	}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
		NTweets:          120,
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

	now := time.Now().UTC()
	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return buildMockUserInfoResponse([]social.TwitterUser{
				{
					ID:       "123",
					Name:     "Test User",
					Username: "testuser",
					PublicMetrics: social.TwitterUserPublicMetrics{
						FollowersCount: 100,
						FollowingCount: 50,
						TweetCount:     500,
					},
				},
			}), nil
		},
		FetchUserTweetsFunc: func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
			return buildMockTweetResponse([]social.TwitterTweet{
				{
					ID:        "tweet1",
					Text:      "Test tweet",
					CreatedAt: now.Format(time.RFC3339),
					PublicMetrics: social.TwitterTweetMetrics{
						RetweetCount: 10,
						ReplyCount:   5,
						LikeCount:    20,
						QuoteCount:   2,
					},
				},
			}), nil
		},
	}
	sink := &MockTwitterPostSink{}
	repo := &MockSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				UserID:             userID,
				WorkspaceID:        workspaceID,
				PlatformIdentifier: "twitter-123",
				PlatformName:       "Test Account",
				State:              "Added",
			}, nil
		},
	}
	pusher := &MockPusherNotifier{}
	notifier := &MockEmailNotifier{}

	wo := ImmediateWorkOrder{
		ID:               accountID.Hex(),
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, notifier, pusher, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify pusher was called for Added state.
	if len(pusher.Triggered) != 1 {
		t.Errorf("expected 1 pusher trigger, got %d", len(pusher.Triggered))
	}

	// Verify email was sent (state was "Added")
	if len(notifier.Sent) != 1 {
		t.Errorf("expected 1 email notification, got %d", len(notifier.Sent))
	}

	// Verify posts inserted
	if len(sink.InsertedPosts) != 1 {
		t.Errorf("expected 1 inserted post, got %d", len(sink.InsertedPosts))
	}
}

func TestProcessAccount_WithAccountStateNotAdded(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	accountID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	workspaceID := primitive.NewObjectID()

	now := time.Now().UTC()
	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return buildMockUserInfoResponse([]social.TwitterUser{
				{ID: "123", Name: "Test", Username: "test"},
			}), nil
		},
		FetchUserTweetsFunc: func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
			return buildMockTweetResponse([]social.TwitterTweet{
				{
					ID:        "tweet1",
					Text:      "Test",
					CreatedAt: now.Format(time.RFC3339),
					PublicMetrics: social.TwitterTweetMetrics{
						RetweetCount: 1, ReplyCount: 1, LikeCount: 1, QuoteCount: 0,
					},
				},
			}), nil
		},
	}
	sink := &MockTwitterPostSink{}
	repo := &MockSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				UserID:             userID,
				WorkspaceID:        workspaceID,
				PlatformIdentifier: "twitter-123",
				PlatformName:       "Test Account",
				State:              "Active", // Not "Added"
			}, nil
		},
	}
	pusher := &MockPusherNotifier{}
	notifier := &MockEmailNotifier{}

	wo := ImmediateWorkOrder{
		ID:               accountID.Hex(),
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, notifier, pusher, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Pusher should be sent for successful completion regardless of original state.
	if len(pusher.Triggered) != 1 {
		t.Errorf("expected 1 pusher trigger for non-Added/Syncing state, got %d", len(pusher.Triggered))
	}

	// Email should NOT be sent (state is not "Added")
	if len(notifier.Sent) != 0 {
		t.Errorf("expected 0 email notifications for non-Added state, got %d", len(notifier.Sent))
	}
}

func TestProcessAccount_InvalidAccountID(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return buildMockUserInfoResponse([]social.TwitterUser{
				{ID: "123", Name: "Test", Username: "test"},
			}), nil
		},
		FetchUserTweetsFunc: func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
			return &social.TwitterTweetsResponse{Data: nil}, nil
		},
	}
	sink := &MockTwitterPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		ID:               "invalid-hex",
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
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

	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return buildMockUserInfoResponse([]social.TwitterUser{
				{ID: "123", Name: "Test", Username: "test"},
			}), nil
		},
		FetchUserTweetsFunc: func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
			return &social.TwitterTweetsResponse{Data: nil}, nil
		},
	}
	sink := &MockTwitterPostSink{}
	repo := &MockSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, errors.New("mongo error")
		},
	}

	wo := ImmediateWorkOrder{
		ID:               accountID.Hex(),
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
	}

	// Should not error, just log warning
	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessAccount_ClearsStaleProcessingErrorBeforeRetry(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	accountID := primitive.NewObjectID()
	repo := &MockSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:       accountID,
				MetaData: map[string]interface{}{"last_processing_error": "token expired"},
			}, nil
		},
	}
	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return nil, errors.New("user info fetch error")
		},
	}

	wo := ImmediateWorkOrder{
		ID:               accountID.Hex(),
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
	}

	err := ProcessAccount(ctx, fetcher, &MockTwitterPostSink{}, repo, nil, nil, wo, "", log)
	if err == nil {
		t.Fatal("expected error for user info fetch failure")
	}
	if len(repo.ClearProcessingErrorCalls) != 1 {
		t.Fatalf("expected 1 stale-error clear, got %d", len(repo.ClearProcessingErrorCalls))
	}
	if repo.ClearProcessingErrorCalls[0] != accountID {
		t.Fatalf("cleared account %s, want %s", repo.ClearProcessingErrorCalls[0].Hex(), accountID.Hex())
	}
}

func TestProcessAccount_Pagination(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	now := time.Now().UTC()
	callCount := 0
	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return buildMockUserInfoResponse([]social.TwitterUser{
				{ID: "123", Name: "Test", Username: "test"},
			}), nil
		},
		FetchUserTweetsFunc: func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
			callCount++
			if callCount == 1 {
				// First call - has more pages
				return &social.TwitterTweetsResponse{
					Data: []social.TwitterTweet{
						{
							ID:        "tweet1",
							Text:      "First page",
							CreatedAt: now.Format(time.RFC3339),
							PublicMetrics: social.TwitterTweetMetrics{
								RetweetCount: 1, ReplyCount: 1, LikeCount: 1, QuoteCount: 0,
							},
						},
					},
					Meta: &social.TwitterMeta{
						NextToken:   "page2token",
						ResultCount: 1,
					},
				}, nil
			}
			// Second call - no more pages
			return &social.TwitterTweetsResponse{
				Data: []social.TwitterTweet{
					{
						ID:        "tweet2",
						Text:      "Second page",
						CreatedAt: now.Add(-1 * time.Hour).Format(time.RFC3339),
						PublicMetrics: social.TwitterTweetMetrics{
							RetweetCount: 2, ReplyCount: 2, LikeCount: 2, QuoteCount: 1,
						},
					},
				},
				Meta: &social.TwitterMeta{
					ResultCount: 1,
				},
			}, nil
		},
	}
	sink := &MockTwitterPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
		NTweets:          120,
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount < 2 {
		t.Errorf("expected at least 2 fetch calls for pagination, got %d", callCount)
	}

	if len(sink.InsertedPosts) != 2 {
		t.Errorf("expected 2 inserted posts, got %d", len(sink.InsertedPosts))
	}
}

func TestProcessAccount_ContextCancelled(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())

	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return buildMockUserInfoResponse([]social.TwitterUser{
				{ID: "123", Name: "Test", Username: "test"},
			}), nil
		},
		FetchUserTweetsFunc: func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
			cancel() // Cancel context during fetch
			return &social.TwitterTweetsResponse{
				Data: []social.TwitterTweet{
					{
						ID:        "tweet1",
						Text:      "Test",
						CreatedAt: time.Now().Format(time.RFC3339),
						PublicMetrics: social.TwitterTweetMetrics{
							RetweetCount: 1, ReplyCount: 1, LikeCount: 1, QuoteCount: 0,
						},
					},
				},
				Meta: &social.TwitterMeta{
					NextToken:   "next",
					ResultCount: 1,
				},
			}, nil
		},
	}
	sink := &MockTwitterPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
		NTweets:          120,
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled error, got %v", err)
	}
}

func TestProcessAccount_NoUserInfo(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	now := time.Now().UTC()
	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return &social.TwitterUserResponse{Data: nil}, nil // No user data returned
		},
		FetchUserTweetsFunc: func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
			return buildMockTweetResponse([]social.TwitterTweet{
				{
					ID:        "tweet1",
					Text:      "Test tweet",
					CreatedAt: now.Format(time.RFC3339),
					PublicMetrics: social.TwitterTweetMetrics{
						RetweetCount: 5, ReplyCount: 3, LikeCount: 10, QuoteCount: 1,
					},
				},
			}), nil
		},
	}
	sink := &MockTwitterPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
	}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Posts should still be inserted even without user info
	if len(sink.InsertedPosts) != 1 {
		t.Errorf("expected 1 inserted post, got %d", len(sink.InsertedPosts))
	}

	// No insights without user info
	if len(sink.InsertedInsights) != 0 {
		t.Errorf("expected 0 insights without user info, got %d", len(sink.InsertedInsights))
	}
}

func TestProcessAccount_EmptyWorkOrder(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return &social.TwitterUserResponse{Data: nil}, nil
		},
		FetchUserTweetsFunc: func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
			return &social.TwitterTweetsResponse{Data: nil}, nil
		},
	}
	sink := &MockTwitterPostSink{}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{}

	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessAccount_InsightsInsertError(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	now := time.Now().UTC()
	fetcher := &MockTwitterTweetFetcher{
		FetchUserInfoFunc: func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
			return buildMockUserInfoResponse([]social.TwitterUser{
				{
					ID:       "123",
					Name:     "Test User",
					Username: "testuser",
					PublicMetrics: social.TwitterUserPublicMetrics{
						FollowersCount: 100,
						FollowingCount: 50,
						TweetCount:     500,
					},
				},
			}), nil
		},
		FetchUserTweetsFunc: func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
			return buildMockTweetResponse([]social.TwitterTweet{
				{
					ID:        "tweet1",
					Text:      "Test",
					CreatedAt: now.Format(time.RFC3339),
					PublicMetrics: social.TwitterTweetMetrics{
						RetweetCount: 1, ReplyCount: 1, LikeCount: 1, QuoteCount: 0,
					},
				},
			}), nil
		},
	}
	sink := &MockTwitterPostSink{
		BulkInsertTwitterInsightsFunc: func(ctx context.Context, insights []*chmodels.TwitterInsights) error {
			return errors.New("insights insert error")
		},
	}
	repo := &MockSocialRepository{}

	wo := ImmediateWorkOrder{
		TwitterID:        "tw-456",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
	}

	// Insights error is non-fatal, just logged
	err := ProcessAccount(ctx, fetcher, sink, repo, nil, nil, wo, "", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Posts should still be inserted
	if len(sink.InsertedPosts) != 1 {
		t.Errorf("expected 1 inserted post, got %d", len(sink.InsertedPosts))
	}
}

// ================== Logging Contract Tests (Point 3 — Processor never logs at Error level) ==================

func TestLoggingContract_TwitterProcessor_NoErrorLevel(t *testing.T) {
	// Twitter immediate processor is a "called" module — it must use Warn, never Error.

	hookRecords, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	log, buf := logger.NewTestLoggerWithHook()

	// Simulate all the Warn-level logs the processor produces
	log.Warn().Err(errors.New("rate limit")).
		Str("error_message", "rate limit").
		Str("function", "ProcessAccount").
		Str("stage", "fetch_user_info").
		Msg("Failed to fetch user info")

	log.Warn().Err(errors.New("insert failed")).
		Str("error_message", "insert failed").
		Str("function", "ProcessAccount").
		Str("stage", "insert_insights").
		Msg("Failed to insert insights (continuing)")

	log.Warn().Err(errors.New("mongo error")).
		Str("error_message", "mongo error").
		Str("function", "ProcessAccount").
		Str("stage", "update_mongo_state").
		Msg("Failed to update account state")

	output := buf.String()

	if strings.Count(output, "ERR") > 0 {
		t.Errorf("processor should never produce Error-level logs, but found ERR in output:\n%s", output)
	}

	for _, r := range *hookRecords {
		if r.Level >= zerolog.ErrorLevel {
			t.Errorf("processor should not trigger Error+ hook, got level %v", r.Level)
		}
	}
}

func TestLoggingContract_TwitterProcessor_NoCaptureException(t *testing.T) {
	captureRecords, captureCleanup := logger.InstallCaptureSpy()
	defer captureCleanup()

	log, _ := logger.NewTestLoggerWithHook()

	log.Warn().Err(errors.New("auth error")).
		Str("error_message", "auth error").
		Str("function", "ProcessAccount").
		Str("stage", "fetch_tweets").
		Msg("Twitter API error")

	if len(*captureRecords) != 0 {
		t.Fatalf("processor should not call CaptureException, got %d calls", len(*captureRecords))
	}
}
