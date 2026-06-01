package social

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	models "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ===================== MockFacebookClient Tests =====================

func TestMockFacebookClient_FetchPosts(t *testing.T) {
	mock := &MockFacebookClient{}

	// Test with nil function
	result, err := mock.FetchPosts(context.Background(), "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	expectedPosts := []kafkamodels.RawFacebookPost{{ID: "post1"}, {ID: "post2"}}
	mock.FetchPostsFunc = func(ctx context.Context, pageID, accessToken string) ([]kafkamodels.RawFacebookPost, error) {
		return expectedPosts, nil
	}
	result, err = mock.FetchPosts(context.Background(), "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(result))
	}
}

func TestMockFacebookClient_FetchPostsWithLimit(t *testing.T) {
	mock := &MockFacebookClient{}

	// Test with nil function
	result, err := mock.FetchPostsWithLimit(context.Background(), "page123", "token", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchPostsWithLimitFunc = func(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookPost, error) {
		return []kafkamodels.RawFacebookPost{{ID: "post1"}}, nil
	}
	result, err = mock.FetchPostsWithLimit(context.Background(), "page123", "token", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 post, got %d", len(result))
	}
}

func TestMockFacebookClient_FetchPostsSince(t *testing.T) {
	mock := &MockFacebookClient{}
	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()

	// Test with nil function
	result, err := mock.FetchPostsSince(context.Background(), "page123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchPostsSinceFunc = func(ctx context.Context, pageID, accessToken string, s, u time.Time) ([]kafkamodels.RawFacebookPost, error) {
		return []kafkamodels.RawFacebookPost{{ID: "post1"}}, nil
	}
	result, err = mock.FetchPostsSince(context.Background(), "page123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 post, got %d", len(result))
	}
}

func TestMockFacebookClient_FetchVideos(t *testing.T) {
	mock := &MockFacebookClient{}

	// Test with nil function
	result, err := mock.FetchVideos(context.Background(), "page123", "token", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchVideosFunc = func(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookVideo, error) {
		return []kafkamodels.RawFacebookVideo{{PostID: "video1"}}, nil
	}
	result, err = mock.FetchVideos(context.Background(), "page123", "token", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 video, got %d", len(result))
	}
}

func TestMockFacebookClient_FetchVideosWithLimit(t *testing.T) {
	mock := &MockFacebookClient{}

	// Test with nil function
	result, err := mock.FetchVideosWithLimit(context.Background(), "page123", "token", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchVideosWithLimitFunc = func(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookVideo, error) {
		return []kafkamodels.RawFacebookVideo{{PostID: "video1"}}, nil
	}
	result, err = mock.FetchVideosWithLimit(context.Background(), "page123", "token", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 video, got %d", len(result))
	}
}

func TestMockFacebookClient_FetchVideosSince(t *testing.T) {
	mock := &MockFacebookClient{}
	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()

	// Test with nil function
	result, err := mock.FetchVideosSince(context.Background(), "page123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchVideosSinceFunc = func(ctx context.Context, pageID, accessToken string, s, u time.Time) ([]kafkamodels.RawFacebookVideo, error) {
		return []kafkamodels.RawFacebookVideo{{PostID: "video1"}}, nil
	}
	result, err = mock.FetchVideosSince(context.Background(), "page123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 video, got %d", len(result))
	}
}

func TestMockFacebookClient_FetchInsights(t *testing.T) {
	mock := &MockFacebookClient{}
	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()

	// Test with nil function
	result, err := mock.FetchInsights(context.Background(), "page123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchInsightsFunc = func(ctx context.Context, pageID, accessToken string, s, u time.Time) (*kafkamodels.RawFacebookInsights, error) {
		return &kafkamodels.RawFacebookInsights{}, nil
	}
	result, err = mock.FetchInsights(context.Background(), "page123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestMockFacebookClient_GetPostThumbnails(t *testing.T) {
	mock := &MockFacebookClient{}
	posts := []clickhouse.MinimalPost{{PostID: "post1"}, {PostID: "post2"}}

	// Test with nil function (returns input posts)
	result, err := mock.GetPostThumbnails(context.Background(), "page123", "token", "longToken", "key", posts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(result))
	}

	// Test with custom function
	mock.GetPostThumbnailsFunc = func(ctx context.Context, pageID, accessToken, longAccessToken, decryptionKey string, p []clickhouse.MinimalPost) ([]clickhouse.MinimalPost, error) {
		for i := range p {
			p[i].FullPicture = "updated"
		}
		return p, nil
	}
	result, err = mock.GetPostThumbnails(context.Background(), "page123", "token", "longToken", "key", posts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].FullPicture != "updated" {
		t.Fatal("expected updated picture")
	}
}

func TestMockFacebookClient_GetCompetitorPageDetails(t *testing.T) {
	mock := &MockFacebookClient{}

	// Test with nil function
	details, picture, err := mock.GetCompetitorPageDetails(context.Background(), "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details != nil || picture != nil {
		t.Fatal("expected nil results")
	}

	// Test with custom function
	mock.GetCompetitorPageDetailsFunc = func(ctx context.Context, pageID, accessToken string) (*models.FacebookPageDetails, *models.Picture, error) {
		return &models.FacebookPageDetails{Name: "Test Page"}, &models.Picture{}, nil
	}
	details, picture, err = mock.GetCompetitorPageDetails(context.Background(), "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details.Name != "Test Page" {
		t.Fatalf("expected 'Test Page', got '%s'", details.Name)
	}
}

func TestMockFacebookClient_GetCompetitorPosts(t *testing.T) {
	mock := &MockFacebookClient{}
	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()

	// Test with nil function
	posts, next, err := mock.GetCompetitorPosts(context.Background(), "page123", "token", since, until, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if posts != nil || next != "" {
		t.Fatal("expected nil/empty results")
	}

	// Test with custom function
	mock.GetCompetitorPostsFunc = func(ctx context.Context, pageID, accessToken string, s, u time.Time, limit int) ([]*models.Post, string, error) {
		return []*models.Post{{ID: "post1"}}, "next_url", nil
	}
	posts, next, err = mock.GetCompetitorPosts(context.Background(), "page123", "token", since, until, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if next != "next_url" {
		t.Fatalf("expected 'next_url', got '%s'", next)
	}
}

func TestMockFacebookClient_GetCompetitorPostsFromURL(t *testing.T) {
	mock := &MockFacebookClient{}

	// Test with nil function
	posts, next, err := mock.GetCompetitorPostsFromURL(context.Background(), "url", "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if posts != nil || next != "" {
		t.Fatal("expected nil/empty results")
	}

	// Test with custom function
	mock.GetCompetitorPostsFromURLFunc = func(ctx context.Context, nextURL, pageID, accessToken string) ([]*models.Post, string, error) {
		return []*models.Post{{ID: "post1"}}, "", nil
	}
	posts, next, err = mock.GetCompetitorPostsFromURL(context.Background(), "url", "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
}

func TestMockFacebookClient_GetCompetitorSharedPostDetails(t *testing.T) {
	mock := &MockFacebookClient{}

	// Test with nil function
	post, err := mock.GetCompetitorSharedPostDetails(context.Background(), "parent123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.GetCompetitorSharedPostDetailsFunc = func(ctx context.Context, parentID, accessToken string) (*models.Post, error) {
		return &models.Post{ID: "post1"}, nil
	}
	post, err = mock.GetCompetitorSharedPostDetails(context.Background(), "parent123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post.ID != "post1" {
		t.Fatalf("expected 'post1', got '%s'", post.ID)
	}
}

func TestMockFacebookClient_GetCompetitorPagePicture(t *testing.T) {
	mock := &MockFacebookClient{}

	// Test with nil function
	picture, err := mock.GetCompetitorPagePicture(context.Background(), "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if picture != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.GetCompetitorPagePictureFunc = func(ctx context.Context, pageID, accessToken string) (*models.Picture, error) {
		return &models.Picture{}, nil
	}
	picture, err = mock.GetCompetitorPagePicture(context.Background(), "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if picture == nil {
		t.Fatal("expected non-nil result")
	}
}

// ===================== MockInstagramClient Tests =====================

func TestMockInstagramClient_FetchMedia(t *testing.T) {
	mock := &MockInstagramClient{}

	// Test with nil function
	result, err := mock.FetchMedia(context.Background(), "ig123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchMediaFunc = func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
		return []kafkamodels.RawInstagramMedia{{ID: "media1"}}, nil
	}
	result, err = mock.FetchMedia(context.Background(), "ig123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 media, got %d", len(result))
	}
}

func TestMockInstagramClient_FetchMediaSince(t *testing.T) {
	mock := &MockInstagramClient{}
	since := time.Now().Add(-24 * time.Hour)

	// Test with nil function
	result, err := mock.FetchMediaSince(context.Background(), "ig123", "token", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchMediaSinceFunc = func(ctx context.Context, instagramID, accessToken string, s time.Time) ([]kafkamodels.RawInstagramMedia, error) {
		return []kafkamodels.RawInstagramMedia{{ID: "media1"}}, nil
	}
	result, err = mock.FetchMediaSince(context.Background(), "ig123", "token", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 media, got %d", len(result))
	}
}

func TestMockInstagramClient_FetchAllMedia(t *testing.T) {
	mock := &MockInstagramClient{}

	// Test with nil function
	result, err := mock.FetchAllMedia(context.Background(), "ig123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchAllMediaFunc = func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
		return []kafkamodels.RawInstagramMedia{{ID: "media1"}, {ID: "media2"}}, nil
	}
	result, err = mock.FetchAllMedia(context.Background(), "ig123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 media, got %d", len(result))
	}
}

func TestMockInstagramClient_FetchStories(t *testing.T) {
	mock := &MockInstagramClient{}

	// Test with nil function
	result, err := mock.FetchStories(context.Background(), "ig123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchStoriesFunc = func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
		return []kafkamodels.RawInstagramMedia{{ID: "story1"}}, nil
	}
	result, err = mock.FetchStories(context.Background(), "ig123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 story, got %d", len(result))
	}
}

func TestMockInstagramClient_FetchInsightsDaily(t *testing.T) {
	mock := &MockInstagramClient{}

	// Test with nil function
	result, err := mock.FetchInsightsDaily(context.Background(), "ig123", "token", 7, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchInsightsDailyFunc = func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]DailyInsight, error) {
		return []DailyInsight{{Date: time.Now()}}, nil
	}
	result, err = mock.FetchInsightsDaily(context.Background(), "ig123", "token", 7, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(result))
	}
}

func TestMockInstagramClient_WithBaseURL(t *testing.T) {
	mock := &MockInstagramClient{}

	// Test with nil function
	result := mock.WithBaseURL("http://example.com")
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.WithBaseURLFunc = func(url string) *InstagramClient {
		return &InstagramClient{}
	}
	result = mock.WithBaseURL("http://example.com")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// ===================== MockLinkedInClient Tests =====================

func TestMockLinkedInClient_FetchShares(t *testing.T) {
	mock := &MockLinkedInClient{}

	// Test with nil function
	result, err := mock.FetchShares(context.Background(), "org123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchSharesFunc = func(ctx context.Context, organisationID, accessToken string) ([]json.RawMessage, error) {
		return []json.RawMessage{[]byte(`{"id":"share1"}`)}, nil
	}
	result, err = mock.FetchShares(context.Background(), "org123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 share, got %d", len(result))
	}
}

func TestMockLinkedInClient_FetchPostsPaginated(t *testing.T) {
	mock := &MockLinkedInClient{}
	cutoff := time.Now().Add(-24 * time.Hour)

	// Test with nil function
	result, err := mock.FetchPostsPaginated(context.Background(), "li123", "organization", "token", cutoff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchPostsPaginatedFunc = func(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error) {
		return []json.RawMessage{[]byte(`{"id":"post1"}`)}, nil
	}
	result, err = mock.FetchPostsPaginated(context.Background(), "li123", "organization", "token", cutoff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 post, got %d", len(result))
	}
}

func TestMockLinkedInClient_FetchFollowerStatsWithGeoIDs(t *testing.T) {
	mock := &MockLinkedInClient{}

	// Test with nil function
	result, err := mock.FetchFollowerStatsWithGeoIDs(context.Background(), "li123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchFollowerStatsWithGeoIDsFunc = func(ctx context.Context, linkedinID string, accessToken string) (*FollowerStatsWithGeoIDs, error) {
		return &FollowerStatsWithGeoIDs{}, nil
	}
	result, err = mock.FetchFollowerStatsWithGeoIDs(context.Background(), "li123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestMockLinkedInClient_ResolveGeoIDs(t *testing.T) {
	mock := &MockLinkedInClient{}

	// Test with nil function
	result, err := mock.ResolveGeoIDs(context.Background(), []string{"geo1", "geo2"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.ResolveGeoIDsFunc = func(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error) {
		return map[string]string{"geo1": "New York", "geo2": "Los Angeles"}, nil
	}
	result, err = mock.ResolveGeoIDs(context.Background(), []string{"geo1", "geo2"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
}

// ===================== MockTikTokClient Tests =====================

func TestMockTikTokClient_FetchUserVideos(t *testing.T) {
	mock := &MockTikTokClient{}

	// Test with nil function
	result, cursor, err := mock.FetchUserVideos(context.Background(), "user123", "token", 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}
	if cursor != 0 {
		t.Fatalf("expected cursor 0, got %d", cursor)
	}

	// Test with custom function
	mock.FetchUserVideosFunc = func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
		return []byte(`{"videos":[]}`), 100, nil
	}
	result, cursor, err = mock.FetchUserVideos(context.Background(), "user123", "token", 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if cursor != 100 {
		t.Fatalf("expected cursor 100, got %d", cursor)
	}
}

func TestMockTikTokClient_FetchUserInfo(t *testing.T) {
	mock := &MockTikTokClient{}

	// Test with nil function
	result, err := mock.FetchUserInfo(context.Background(), "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	mock.FetchUserInfoFunc = func(ctx context.Context, accessToken string) (json.RawMessage, error) {
		return []byte(`{"user":{"id":"123"}}`), nil
	}
	result, err = mock.FetchUserInfo(context.Background(), "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestMockTikTokClient_FetchVideoList(t *testing.T) {
	mock := &MockTikTokClient{}

	// Test with nil function
	result, cursor, hasMore, err := mock.FetchVideoList(context.Background(), "token", 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}
	if cursor != 0 || hasMore {
		t.Fatal("expected default values")
	}

	// Test with custom function
	mock.FetchVideoListFunc = func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
		return []byte(`{"videos":[]}`), 100, true, nil
	}
	result, cursor, hasMore, err = mock.FetchVideoList(context.Background(), "token", 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasMore {
		t.Fatal("expected hasMore to be true")
	}
}

func TestMockTikTokClient_RefreshToken(t *testing.T) {
	mock := &MockTikTokClient{}

	// Test with nil function
	resp, err := mock.RefreshToken(context.Background(), "refresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Fatal("expected nil response")
	}

	// Test with custom function
	mock.RefreshTokenFunc = func(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
		return &RefreshTokenResponse{
			AccessToken:      "new_access",
			RefreshToken:     "new_refresh",
			ExpiresIn:        86400,
			RefreshExpiresIn: 7776000,
			Scope:            "user.info.basic,video.list",
		}, nil
	}
	resp, err = mock.RefreshToken(context.Background(), "refresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "new_access" {
		t.Fatalf("expected 'new_access', got '%s'", resp.AccessToken)
	}
}

// ===================== Error Handling Tests =====================

func TestMockFacebookClient_Error(t *testing.T) {
	mock := &MockFacebookClient{
		FetchPostsFunc: func(ctx context.Context, pageID, accessToken string) ([]kafkamodels.RawFacebookPost, error) {
			return nil, errors.New("api error")
		},
	}

	_, err := mock.FetchPosts(context.Background(), "page123", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockInstagramClient_Error(t *testing.T) {
	mock := &MockInstagramClient{
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, errors.New("api error")
		},
	}

	_, err := mock.FetchMedia(context.Background(), "ig123", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockLinkedInClient_Error(t *testing.T) {
	mock := &MockLinkedInClient{
		FetchSharesFunc: func(ctx context.Context, organisationID, accessToken string) ([]json.RawMessage, error) {
			return nil, errors.New("api error")
		},
	}

	_, err := mock.FetchShares(context.Background(), "org123", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockTikTokClient_Error(t *testing.T) {
	mock := &MockTikTokClient{
		FetchUserInfoFunc: func(ctx context.Context, accessToken string) (json.RawMessage, error) {
			return nil, errors.New("api error")
		},
	}

	_, err := mock.FetchUserInfo(context.Background(), "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===================== Interface Implementation Tests =====================

func TestMockFacebookClient_ImplementsInterface(t *testing.T) {
	var _ FacebookAPI = (*MockFacebookClient)(nil)
}

func TestMockInstagramClient_ImplementsInterface(t *testing.T) {
	var _ InstagramAPI = (*MockInstagramClient)(nil)
}

func TestMockLinkedInClient_ImplementsInterface(t *testing.T) {
	var _ LinkedInAPI = (*MockLinkedInClient)(nil)
}

func TestMockTikTokClient_ImplementsInterface(t *testing.T) {
	var _ TikTokAPI = (*MockTikTokClient)(nil)
}

// ===================== MockGMBClient Tests =====================

func TestMockGMBClient_RefreshToken(t *testing.T) {
	mock := &MockGMBClient{}

	// Test with nil function (returns default)
	resp, err := mock.RefreshToken(context.Background(), "refresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "mock_access_token" {
		t.Fatalf("expected 'mock_access_token', got '%s'", resp.AccessToken)
	}

	// Test with custom function
	mock.RefreshTokenFunc = func(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
		return &RefreshTokenResponse{AccessToken: "custom_token"}, nil
	}
	resp, err = mock.RefreshToken(context.Background(), "refresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "custom_token" {
		t.Fatalf("expected 'custom_token', got '%s'", resp.AccessToken)
	}
}

func TestMockGMBClient_FetchVoiceOfMerchant(t *testing.T) {
	mock := &MockGMBClient{}

	// Test with nil function (returns default)
	resp, err := mock.FetchVoiceOfMerchant(context.Background(), "loc-1", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.HasVoiceOfMerchant {
		t.Fatal("expected HasVoiceOfMerchant to be true by default")
	}

	// Test with custom function
	mock.FetchVoiceOfMerchantFunc = func(ctx context.Context, locationID, accessToken string) (*VoiceOfMerchantResponse, error) {
		return &VoiceOfMerchantResponse{HasVoiceOfMerchant: false}, nil
	}
	resp, err = mock.FetchVoiceOfMerchant(context.Background(), "loc-1", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.HasVoiceOfMerchant {
		t.Fatal("expected HasVoiceOfMerchant to be false")
	}
}

func TestMockGMBClient_FetchPerformanceMetrics(t *testing.T) {
	mock := &MockGMBClient{}

	now := time.Now()
	resp, err := mock.FetchPerformanceMetrics(context.Background(), "loc-1", "token", now, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestMockGMBClient_FetchSearchKeywords(t *testing.T) {
	mock := &MockGMBClient{}

	now := time.Now()
	resp, err := mock.FetchSearchKeywords(context.Background(), "loc-1", "token", now, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestMockGMBClient_FetchLocalPosts(t *testing.T) {
	mock := &MockGMBClient{}

	resp, err := mock.FetchLocalPosts(context.Background(), "acc-1", "loc-1", "token", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestMockGMBClient_FetchReviews(t *testing.T) {
	mock := &MockGMBClient{}

	resp, err := mock.FetchReviews(context.Background(), "acc-1", "loc-1", "token", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestMockGMBClient_FetchMediaAssets(t *testing.T) {
	mock := &MockGMBClient{}

	resp, err := mock.FetchMediaAssets(context.Background(), "acc-1", "loc-1", "token", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestMockGMBClient_Error(t *testing.T) {
	mock := &MockGMBClient{
		FetchVoiceOfMerchantFunc: func(ctx context.Context, locationID, accessToken string) (*VoiceOfMerchantResponse, error) {
			return nil, errors.New("api error")
		},
	}

	_, err := mock.FetchVoiceOfMerchant(context.Background(), "loc-1", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockGMBClient_ImplementsInterface(t *testing.T) {
	var _ GMBAPI = (*MockGMBClient)(nil)
}

// ===================== MockGeoResolver Tests =====================

func TestMockGeoResolver_ResolveGeoIDs(t *testing.T) {
	mock := &MockGeoResolver{}

	// Test with nil function (returns empty map)
	result, err := mock.ResolveGeoIDs(context.Background(), []string{"geo1", "geo2"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(result))
	}

	// Test with custom function
	mock.ResolveGeoIDsFunc = func(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error) {
		return map[string]string{"geo1": "New York", "geo2": "Los Angeles"}, nil
	}
	result, err = mock.ResolveGeoIDs(context.Background(), []string{"geo1", "geo2"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result["geo1"] != "New York" {
		t.Fatalf("expected 'New York', got '%s'", result["geo1"])
	}
}

func TestMockGeoResolver_ResolveGeoIDsWithType(t *testing.T) {
	mock := &MockGeoResolver{}
	geoIDsWithType := []GeoIDWithType{
		{ID: "geo1", Type: "country"},
		{ID: "geo2", Type: "city"},
	}

	// Test with nil function (returns empty map)
	result, err := mock.ResolveGeoIDsWithType(context.Background(), geoIDsWithType, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(result))
	}

	// Test with custom function
	mock.ResolveGeoIDsWithTypeFunc = func(ctx context.Context, geoIDs []GeoIDWithType, accessToken string) (map[string]string, error) {
		return map[string]string{"geo1": "United States", "geo2": "San Francisco"}, nil
	}
	result, err = mock.ResolveGeoIDsWithType(context.Background(), geoIDsWithType, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result["geo2"] != "San Francisco" {
		t.Fatalf("expected 'San Francisco', got '%s'", result["geo2"])
	}
}

func TestMockGeoResolver_Error(t *testing.T) {
	mock := &MockGeoResolver{
		ResolveGeoIDsFunc: func(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error) {
			return nil, errors.New("api error")
		},
	}

	_, err := mock.ResolveGeoIDs(context.Background(), []string{"geo1"}, "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockGeoResolver_ImplementsInterface(t *testing.T) {
	var _ GeoResolverAPI = (*MockGeoResolver)(nil)
}
