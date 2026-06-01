package kafka

import (
	"encoding/json"
	"testing"
	"time"
)

// TestTwitterAccountWorkOrder_JSON tests JSON marshaling/unmarshaling
func TestTwitterAccountWorkOrder_JSON(t *testing.T) {
	order := TwitterAccountWorkOrder{
		ID:               "507f1f77bcf86cd799439011",
		WorkspaceID:      "workspace_123",
		TwitterID:        "twitter_456",
		OAuthToken:       "oauth_token_value",
		OAuthTokenSecret: "oauth_secret_value",
		NTweets:          100,
		APIKey:           "api_key_value",
		APISecret:        "api_secret_value",
		AppName:          "MyApp",
		AppID:            "app_id_123",
		ExecutedBy:       "internal",
		SyncType:         "incremental",
	}

	// Marshal to JSON
	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var unmarshaled TwitterAccountWorkOrder
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != order.ID {
		t.Errorf("ID mismatch: expected %s, got %s", order.ID, unmarshaled.ID)
	}
	if unmarshaled.TwitterID != order.TwitterID {
		t.Errorf("TwitterID mismatch: expected %s, got %s", order.TwitterID, unmarshaled.TwitterID)
	}
	if unmarshaled.OAuthToken != order.OAuthToken {
		t.Errorf("OAuthToken mismatch: expected %s, got %s", order.OAuthToken, unmarshaled.OAuthToken)
	}
	if unmarshaled.SyncType != order.SyncType {
		t.Errorf("SyncType mismatch: expected %s, got %s", order.SyncType, unmarshaled.SyncType)
	}
}

// TestTwitterAccountWorkOrder_Fields tests that all fields are populated correctly
func TestTwitterAccountWorkOrder_Fields(t *testing.T) {
	order := TwitterAccountWorkOrder{
		ID:               "507f1f77bcf86cd799439011",
		WorkspaceID:      "workspace_123",
		TwitterID:        "twitter_456",
		OAuthToken:       "token123",
		OAuthTokenSecret: "secret456",
		NTweets:          50,
		APIKey:           "key789",
		APISecret:        "secret789",
		AppName:          "TestApp",
		AppID:            "app_789",
		ExecutedBy:       "internal",
		SyncType:         "full_sync",
	}

	if order.ID != "507f1f77bcf86cd799439011" {
		t.Errorf("ID not set correctly")
	}
	if order.WorkspaceID != "workspace_123" {
		t.Errorf("WorkspaceID not set correctly")
	}
	if order.TwitterID != "twitter_456" {
		t.Errorf("TwitterID not set correctly")
	}
	if order.NTweets != 50 {
		t.Errorf("NTweets not set correctly: expected 50, got %d", order.NTweets)
	}
	if order.SyncType != "full_sync" {
		t.Errorf("SyncType not set correctly")
	}
}

// TestTwitterBatchWorkOrder_JSON tests JSON marshaling/unmarshaling
func TestTwitterBatchWorkOrder_JSON(t *testing.T) {
	now := time.Now()
	batch := TwitterBatchWorkOrder{
		BatchID:  "batch_123",
		SyncType: "incremental",
		Accounts: []TwitterAccountWorkOrder{
			{
				ID:        "account_1",
				TwitterID: "twitter_001",
				SyncType:  "incremental",
			},
			{
				ID:        "account_2",
				TwitterID: "twitter_002",
				SyncType:  "incremental",
			},
		},
		CreatedAt: now,
	}

	// Marshal to JSON
	data, err := json.Marshal(batch)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var unmarshaled TwitterBatchWorkOrder
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify fields
	if unmarshaled.BatchID != batch.BatchID {
		t.Errorf("BatchID mismatch: expected %s, got %s", batch.BatchID, unmarshaled.BatchID)
	}
	if unmarshaled.SyncType != batch.SyncType {
		t.Errorf("SyncType mismatch: expected %s, got %s", batch.SyncType, unmarshaled.SyncType)
	}
	if len(unmarshaled.Accounts) != len(batch.Accounts) {
		t.Errorf("Accounts count mismatch: expected %d, got %d", len(batch.Accounts), len(unmarshaled.Accounts))
	}
}

// TestTwitterBatchWorkOrder_EmptyAccounts tests batch with no accounts
func TestTwitterBatchWorkOrder_EmptyAccounts(t *testing.T) {
	batch := TwitterBatchWorkOrder{
		BatchID:   "batch_empty",
		SyncType:  "incremental",
		Accounts:  []TwitterAccountWorkOrder{},
		CreatedAt: time.Now(),
	}

	if len(batch.Accounts) != 0 {
		t.Errorf("expected empty accounts, got %d", len(batch.Accounts))
	}
}

// TestTwitterBatchWorkOrder_MultipleAccounts tests batch with multiple accounts
func TestTwitterBatchWorkOrder_MultipleAccounts(t *testing.T) {
	accounts := make([]TwitterAccountWorkOrder, 0, 200)
	for i := 0; i < 200; i++ {
		accounts = append(accounts, TwitterAccountWorkOrder{
			ID:        "account_" + string(rune(i)),
			TwitterID: "twitter_" + string(rune(i)),
		})
	}

	batch := TwitterBatchWorkOrder{
		BatchID:   "batch_large",
		SyncType:  "incremental",
		Accounts:  accounts,
		CreatedAt: time.Now(),
	}

	if len(batch.Accounts) != 200 {
		t.Errorf("expected 200 accounts, got %d", len(batch.Accounts))
	}
}

// TestRawTwitterPost_JSON tests JSON marshaling/unmarshaling
func TestRawTwitterPost_JSON(t *testing.T) {
	rawData := json.RawMessage(`{"id":"tweet_123","text":"Hello Twitter"}`)
	post := RawTwitterPost{
		WorkspaceID: "workspace_123",
		TwitterID:   "twitter_456",
		Data:        rawData,
	}

	// Marshal to JSON
	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var unmarshaled RawTwitterPost
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify fields
	if unmarshaled.WorkspaceID != post.WorkspaceID {
		t.Errorf("WorkspaceID mismatch: expected %s, got %s", post.WorkspaceID, unmarshaled.WorkspaceID)
	}
	if unmarshaled.TwitterID != post.TwitterID {
		t.Errorf("TwitterID mismatch: expected %s, got %s", post.TwitterID, unmarshaled.TwitterID)
	}
	if string(unmarshaled.Data) != string(post.Data) {
		t.Errorf("Data mismatch: expected %s, got %s", string(post.Data), string(unmarshaled.Data))
	}
}

// TestParsedTwitterPost_JSON tests JSON marshaling/unmarshaling
func TestParsedTwitterPost_JSON(t *testing.T) {
	post := ParsedTwitterPost{
		TwitterID:           "twitter_123",
		Name:                "John Doe",
		Username:            "johndoe",
		ProfileImageURL:     "https://example.com/profile.jpg",
		FollowersCount:      1000,
		FollowingCount:      500,
		TweetCount:          250,
		ListedCount:         10,
		TweetID:             "tweet_001",
		EditHistoryTweetIDs: []string{"tweet_001", "tweet_001_v1"},
		AuthorID:            "author_123",
		AuthorUsername:      "author_johndoe",
		IDCreatedAt:         "2020-01-01T00:00:00Z",
		AuthorIDCreated:     "2020-01-01T00:00:00Z",
		TweetedAt:           "2025-01-15T14:30:00Z",
		Hashtags:            []string{"test", "golang"},
		Permalink:           "https://twitter.com/johndoe/status/tweet_001",
		TweetType:           "original",
		URLs:                []string{"https://example.com"},
		MediaURL:            []string{"https://example.com/image.jpg"},
		UsernameMentioned:   []string{"user1", "user2"},
		UseridMentioned:     []string{"id1", "id2"},
		Lang:                "en",
		TweetText:           "Hello Twitter",
		ImpressionCount:     10000,
		RetweetCount:        100,
		ReplyCount:          50,
		LikeCount:           500,
		BookmarkCount:       75,
		QuoteCount:          10,
		TotalEngagement:     735,
	}

	// Marshal to JSON
	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var unmarshaled ParsedTwitterPost
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify critical fields
	if unmarshaled.TwitterID != post.TwitterID {
		t.Errorf("TwitterID mismatch: expected %s, got %s", post.TwitterID, unmarshaled.TwitterID)
	}
	if unmarshaled.TweetID != post.TweetID {
		t.Errorf("TweetID mismatch: expected %s, got %s", post.TweetID, unmarshaled.TweetID)
	}
	if unmarshaled.FollowersCount != post.FollowersCount {
		t.Errorf("FollowersCount mismatch: expected %d, got %d", post.FollowersCount, unmarshaled.FollowersCount)
	}
	if unmarshaled.LikeCount != post.LikeCount {
		t.Errorf("LikeCount mismatch: expected %d, got %d", post.LikeCount, unmarshaled.LikeCount)
	}
}

// TestParsedTwitterPost_AllFields tests that all fields are properly populated
func TestParsedTwitterPost_AllFields(t *testing.T) {
	post := ParsedTwitterPost{
		TwitterID:           "twitter_123",
		Name:                "John Doe",
		Username:            "johndoe",
		ProfileImageURL:     "https://example.com/profile.jpg",
		FollowersCount:      5000,
		FollowingCount:      500,
		TweetCount:          250,
		ListedCount:         10,
		TweetID:             "tweet_001",
		EditHistoryTweetIDs: []string{"tweet_001"},
		AuthorID:            "author_123",
		AuthorUsername:      "author_johndoe",
		IDCreatedAt:         "2020-01-01T00:00:00Z",
		AuthorIDCreated:     "2020-01-01T00:00:00Z",
		TweetedAt:           "2025-01-15T14:30:00Z",
		Hashtags:            []string{"golang"},
		Permalink:           "https://twitter.com/johndoe/status/tweet_001",
		TweetType:           "original",
		URLs:                []string{"https://example.com"},
		MediaURL:            []string{"https://example.com/image.jpg"},
		UsernameMentioned:   []string{"user1"},
		UseridMentioned:     []string{"id1"},
		Lang:                "en",
		TweetText:           "Hello Twitter",
		ImpressionCount:     10000,
		RetweetCount:        100,
		ReplyCount:          50,
		LikeCount:           500,
		BookmarkCount:       75,
		QuoteCount:          10,
		TotalEngagement:     735,
	}

	// Verify all fields
	if post.TwitterID != "twitter_123" {
		t.Errorf("TwitterID not set correctly")
	}
	if post.FollowersCount != 5000 {
		t.Errorf("FollowersCount not set correctly: expected 5000, got %d", post.FollowersCount)
	}
	if post.TotalEngagement != 735 {
		t.Errorf("TotalEngagement not set correctly: expected 735, got %d", post.TotalEngagement)
	}
	if len(post.Hashtags) != 1 {
		t.Errorf("Hashtags not set correctly: expected 1, got %d", len(post.Hashtags))
	}
}

// TestParsedTwitterInsights_JSON tests JSON marshaling/unmarshaling
func TestParsedTwitterInsights_JSON(t *testing.T) {
	insights := ParsedTwitterInsights{
		TwitterID:          "twitter_123",
		RecordID:           "record_001",
		Name:               "John Doe",
		Username:           "johndoe",
		ProfileImageURL:    "https://example.com/profile.jpg",
		Description:        "Software developer",
		Verified:           true,
		AccountCreatedDate: "2020-01-01T00:00:00Z",
		FollowersCount:     50000,
		FollowingCount:     1000,
		TweetCount:         5000,
		ListedCount:        150,
		LikeCount:          100000,
		InsertedAt:         time.Now().Unix(),
	}

	// Marshal to JSON
	data, err := json.Marshal(insights)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var unmarshaled ParsedTwitterInsights
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify fields
	if unmarshaled.TwitterID != insights.TwitterID {
		t.Errorf("TwitterID mismatch: expected %s, got %s", insights.TwitterID, unmarshaled.TwitterID)
	}
	if unmarshaled.Verified != insights.Verified {
		t.Errorf("Verified mismatch: expected %v, got %v", insights.Verified, unmarshaled.Verified)
	}
	if unmarshaled.FollowersCount != insights.FollowersCount {
		t.Errorf("FollowersCount mismatch: expected %d, got %d", insights.FollowersCount, unmarshaled.FollowersCount)
	}
}

// TestParsedTwitterInsights_AllFields tests that all fields are properly populated
func TestParsedTwitterInsights_AllFields(t *testing.T) {
	now := time.Now().Unix()
	insights := ParsedTwitterInsights{
		TwitterID:          "twitter_123",
		RecordID:           "record_001",
		Name:               "John Doe",
		Username:           "johndoe",
		ProfileImageURL:    "https://example.com/profile.jpg",
		Description:        "Tech enthusiast",
		Verified:           true,
		AccountCreatedDate: "2020-01-01T00:00:00Z",
		FollowersCount:     50000,
		FollowingCount:     1000,
		TweetCount:         5000,
		ListedCount:        150,
		LikeCount:          100000,
		InsertedAt:         now,
	}

	if insights.TwitterID != "twitter_123" {
		t.Errorf("TwitterID not set correctly")
	}
	if insights.FollowersCount != 50000 {
		t.Errorf("FollowersCount not set correctly: expected 50000, got %d", insights.FollowersCount)
	}
	if insights.Verified != true {
		t.Errorf("Verified not set correctly")
	}
	if insights.InsertedAt != now {
		t.Errorf("InsertedAt not set correctly")
	}
}

// TestParsedTwitterPost_EmptyOptionalFields tests post with empty optional fields
func TestParsedTwitterPost_EmptyOptionalFields(t *testing.T) {
	post := ParsedTwitterPost{
		TwitterID:       "twitter_123",
		Name:            "John Doe",
		Username:        "johndoe",
		TweetID:         "tweet_001",
		AuthorID:        "author_123",
		Permalink:       "https://twitter.com/johndoe/status/tweet_001",
		TweetType:       "original",
		TweetText:       "Hello",
		TotalEngagement: 0,
		// Optional fields left empty
		EditHistoryTweetIDs: []string{},
		Hashtags:            []string{},
		URLs:                []string{},
		MediaURL:            []string{},
		UsernameMentioned:   []string{},
		UseridMentioned:     []string{},
	}

	if post.TwitterID != "twitter_123" {
		t.Errorf("TwitterID not set correctly")
	}
	if len(post.Hashtags) != 0 {
		t.Errorf("Hashtags should be empty")
	}
}

// TestTwitterBatchWorkOrder_SyncTypes tests different sync types
func TestTwitterBatchWorkOrder_SyncTypes(t *testing.T) {
	syncTypes := []string{"incremental", "full_sync"}

	for _, syncType := range syncTypes {
		batch := TwitterBatchWorkOrder{
			BatchID:   "batch_test",
			SyncType:  syncType,
			Accounts:  []TwitterAccountWorkOrder{},
			CreatedAt: time.Now(),
		}

		if batch.SyncType != syncType {
			t.Errorf("SyncType not set correctly: expected %s, got %s", syncType, batch.SyncType)
		}
	}
}

// TestTwitterAccountWorkOrder_SyncTypes tests different sync types for account work order
func TestTwitterAccountWorkOrder_SyncTypes(t *testing.T) {
	syncTypes := []string{"incremental", "full_sync"}

	for _, syncType := range syncTypes {
		order := TwitterAccountWorkOrder{
			ID:        "account_1",
			TwitterID: "twitter_123",
			SyncType:  syncType,
		}

		if order.SyncType != syncType {
			t.Errorf("SyncType not set correctly: expected %s, got %s", syncType, order.SyncType)
		}
	}
}

// TestRawTwitterPost_WithLargeData tests raw post with large JSON data
func TestRawTwitterPost_WithLargeData(t *testing.T) {
	largeData := json.RawMessage(`{
		"id": "tweet_123",
		"text": "This is a very long tweet with lots of information and details about various topics that might be relevant to the Twitter API consumer.",
		"author_id": "author_123",
		"created_at": "2025-01-15T14:30:00Z",
		"public_metrics": {
			"retweet_count": 100,
			"reply_count": 50,
			"like_count": 500,
			"impression_count": 10000
		},
		"entities": {
			"hashtags": [{"start": 0, "end": 5, "tag": "test"}],
			"urls": [{"url": "https://example.com", "display_url": "example.com"}]
		}
	}`)

	post := RawTwitterPost{
		WorkspaceID: "workspace_123",
		TwitterID:   "twitter_456",
		Data:        largeData,
	}

	if post.WorkspaceID != "workspace_123" {
		t.Errorf("WorkspaceID not set correctly")
	}
	if len(post.Data) == 0 {
		t.Errorf("Data is empty")
	}
}
