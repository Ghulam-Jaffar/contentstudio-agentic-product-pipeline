package parsing

import (
	"testing"
	"time"

	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
)

func TestNewInstagramCompetitorParser(t *testing.T) {
	parser := NewInstagramCompetitorParser("page123", "testuser", "Test Display Name")

	if parser == nil {
		t.Fatal("expected non-nil parser")
	}

	if parser.pageID != "page123" {
		t.Fatalf("expected pageID 'page123', got %q", parser.pageID)
	}

	if parser.pageName != "testuser" {
		t.Fatalf("expected pageName 'testuser', got %q", parser.pageName)
	}

	if parser.displayName != "Test Display Name" {
		t.Fatalf("expected displayName 'Test Display Name', got %q", parser.displayName)
	}
}

func TestInstagramCompetitorParser_ParsePageInsights(t *testing.T) {
	parser := NewInstagramCompetitorParser("page123", "testuser", "Test Display Name")

	businessDiscovery := &apiModels.BusinessDiscovery{
		ID:                "17841400000000000",
		IgID:              123456789,
		FollowersCount:    10000,
		FollowsCount:      500,
		ProfilePictureURL: "https://example.com/profile.jpg",
		Username:          "testuser",
		Name:              "Test User",
		Biography:         "This is my bio",
		MediaCount:        100,
	}

	result := parser.ParsePageInsights(businessDiscovery)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.InstagramAccountID != "17841400000000000" {
		t.Fatalf("expected InstagramAccountID '17841400000000000', got %q", result.InstagramAccountID)
	}

	if result.TotalFollowedByCount != 10000 {
		t.Fatalf("expected TotalFollowedByCount 10000, got %d", result.TotalFollowedByCount)
	}

	if result.TotalFollowingCount != 500 {
		t.Fatalf("expected TotalFollowingCount 500, got %d", result.TotalFollowingCount)
	}

	if result.ProfilePictureURL != "https://example.com/profile.jpg" {
		t.Fatalf("expected ProfilePictureURL, got %q", result.ProfilePictureURL)
	}

	if result.PageName != "Test Display Name" {
		t.Fatalf("expected PageName 'Test Display Name', got %q", result.PageName)
	}

	if result.RecordID == "" {
		t.Fatal("expected RecordID to be generated")
	}

	if len(result.RecordID) != 32 {
		t.Fatalf("expected RecordID to be 32 character hash, got %d", len(result.RecordID))
	}

	if result.InsertedAt.IsZero() {
		t.Fatal("expected InsertedAt to be set")
	}
}

func TestInstagramCompetitorParser_ParsePosts(t *testing.T) {
	parser := NewInstagramCompetitorParser("page123", "testuser", "Test Display Name")

	businessDiscovery := &apiModels.BusinessDiscovery{
		ID:                "17841400000000000",
		IgID:              123456789,
		FollowersCount:    10000,
		FollowsCount:      500,
		Username:          "testuser",
		Name:              "Test User",
		Biography:         "This is my bio",
		MediaCount:        100,
	}

	media := []apiModels.InstagramMedia{
		{
			ID:               "media123",
			Timestamp:        "2024-01-15T10:30:00+0000",
			Caption:          "Check out this #amazing photo #instagram",
			LikeCount:        500,
			CommentsCount:    50,
			MediaType:        "IMAGE",
			MediaProductType: "FEED",
			MediaURL:         "https://example.com/image1.jpg",
			Permalink:        "https://instagram.com/p/abc123",
		},
		{
			ID:               "media456",
			Timestamp:        "2024-01-16T10:30:00+0000",
			Caption:          "Video post",
			LikeCount:        1000,
			CommentsCount:    100,
			MediaType:        "VIDEO",
			MediaProductType: "REELS",
			MediaURL:         "https://example.com/video1.mp4",
			Permalink:        "https://instagram.com/reel/def456",
		},
	}

	profileImage := "https://example.com/profile.jpg"

	posts := parser.ParsePosts(media, businessDiscovery, profileImage)

	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}

	post1 := posts[0]

	if post1.InstagramID != 123456789 {
		t.Fatalf("expected InstagramID 123456789, got %d", post1.InstagramID)
	}

	if post1.PostID != "media123" {
		t.Fatalf("expected PostID 'media123', got %q", post1.PostID)
	}

	if post1.BusinessAccountID != "17841400000000000" {
		t.Fatalf("expected BusinessAccountID '17841400000000000', got %q", post1.BusinessAccountID)
	}

	if post1.TotalFollowedByCount != 10000 {
		t.Fatalf("expected TotalFollowedByCount 10000, got %d", post1.TotalFollowedByCount)
	}

	if post1.TotalFollowingCount != 500 {
		t.Fatalf("expected TotalFollowingCount 500, got %d", post1.TotalFollowingCount)
	}

	if post1.Username != "testuser" {
		t.Fatalf("expected Username 'testuser', got %q", post1.Username)
	}

	if post1.Name != "Test User" {
		t.Fatalf("expected Name 'Test User', got %q", post1.Name)
	}

	if post1.Biography != "This is my bio" {
		t.Fatalf("expected Biography, got %q", post1.Biography)
	}

	if post1.ProfilePictureURL != profileImage {
		t.Fatalf("expected ProfilePictureURL, got %q", post1.ProfilePictureURL)
	}

	if post1.MediaCount != 100 {
		t.Fatalf("expected MediaCount 100, got %d", post1.MediaCount)
	}

	if post1.LikeCount != 500 {
		t.Fatalf("expected LikeCount 500, got %d", post1.LikeCount)
	}

	if post1.CommentsCount != 50 {
		t.Fatalf("expected CommentsCount 50, got %d", post1.CommentsCount)
	}

	expectedEngagement := int64(500 + 50)
	if post1.Engagement != expectedEngagement {
		t.Fatalf("expected Engagement %d, got %d", expectedEngagement, post1.Engagement)
	}

	if post1.MediaType != "IMAGE" {
		t.Fatalf("expected MediaType 'IMAGE', got %q", post1.MediaType)
	}

	if post1.MediaProductType != "FEED" {
		t.Fatalf("expected MediaProductType 'FEED', got %q", post1.MediaProductType)
	}

	if post1.MediaURL != "https://example.com/image1.jpg" {
		t.Fatalf("expected MediaURL, got %q", post1.MediaURL)
	}

	if post1.Permalink != "https://instagram.com/p/abc123" {
		t.Fatalf("expected Permalink, got %q", post1.Permalink)
	}

	if len(post1.Hashtags) != 2 {
		t.Fatalf("expected 2 hashtags, got %d", len(post1.Hashtags))
	}

	if post1.Hashtags[0] != "amazing" || post1.Hashtags[1] != "instagram" {
		t.Fatalf("expected hashtags [amazing, instagram], got %v", post1.Hashtags)
	}

	post2 := posts[1]

	if post2.MediaType != "VIDEO" {
		t.Fatalf("expected MediaType 'VIDEO', got %q", post2.MediaType)
	}

	if post2.MediaProductType != "REELS" {
		t.Fatalf("expected MediaProductType 'REELS', got %q", post2.MediaProductType)
	}
}

func TestInstagramCompetitorParser_ParsePosts_CarouselAlbum(t *testing.T) {
	parser := NewInstagramCompetitorParser("page123", "testuser", "Test Display Name")

	businessDiscovery := &apiModels.BusinessDiscovery{
		ID:   "17841400000000000",
		IgID: 123456789,
	}

	media := []apiModels.InstagramMedia{
		{
			ID:        "carousel123",
			Timestamp: "2024-01-15T10:30:00+0000",
			MediaType: "CAROUSEL_ALBUM",
			Children: &apiModels.Children{
				Data: []apiModels.ChildMedia{
					{ID: "child1", MediaURL: "https://example.com/child1.jpg"},
					{ID: "child2", MediaURL: "https://example.com/child2.jpg"},
					{ID: "child3", MediaURL: "https://example.com/child3.jpg"},
				},
			},
		},
	}

	posts := parser.ParsePosts(media, businessDiscovery, "")

	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}

	post := posts[0]

	if post.MediaType != "CAROUSEL_ALBUM" {
		t.Fatalf("expected MediaType 'CAROUSEL_ALBUM', got %q", post.MediaType)
	}

	expectedMediaURL := "https://example.com/child1.jpg,https://example.com/child2.jpg,https://example.com/child3.jpg"
	if post.MediaURL != expectedMediaURL {
		t.Fatalf("expected MediaURL %q, got %q", expectedMediaURL, post.MediaURL)
	}
}

func TestInstagramCompetitorParser_ParsePosts_EmptyMedia(t *testing.T) {
	parser := NewInstagramCompetitorParser("page123", "testuser", "Test Display Name")

	businessDiscovery := &apiModels.BusinessDiscovery{
		ID:   "17841400000000000",
		IgID: 123456789,
	}

	media := []apiModels.InstagramMedia{}

	posts := parser.ParsePosts(media, businessDiscovery, "")

	if len(posts) != 0 {
		t.Fatalf("expected 0 posts, got %d", len(posts))
	}
}

func TestInstagramCompetitorParser_ParsePosts_TimestampParsing(t *testing.T) {
	parser := NewInstagramCompetitorParser("page123", "testuser", "Test Display Name")

	businessDiscovery := &apiModels.BusinessDiscovery{
		ID:   "17841400000000000",
		IgID: 123456789,
	}

	cases := []struct {
		name            string
		timestamp       string
		expectValidTime bool
	}{
		{
			name:            "valid timestamp with timezone",
			timestamp:       "2024-01-15T10:30:00+0000",
			expectValidTime: true,
		},
		{
			name:            "RFC3339 with Z suffix",
			timestamp:       "2024-01-15T10:30:00Z",
			expectValidTime: true,
		},
		{
			name:            "invalid timestamp",
			timestamp:       "invalid",
			expectValidTime: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			media := []apiModels.InstagramMedia{
				{
					ID:        "media123",
					Timestamp: tc.timestamp,
				},
			}

			posts := parser.ParsePosts(media, businessDiscovery, "")

			if tc.expectValidTime {
				if len(posts) != 1 {
					t.Fatalf("expected 1 post, got %d", len(posts))
				}
				if posts[0].CreatedAt.IsZero() {
					t.Fatal("expected valid CreatedAt")
				}
			} else {
				if len(posts) != 0 {
					t.Fatalf("expected post to be skipped for invalid timestamp, got %d posts", len(posts))
				}
			}
		})
	}
}

func TestInstagramCompetitorParser_ParsePosts_HashtagExtraction(t *testing.T) {
	parser := NewInstagramCompetitorParser("page123", "testuser", "Test Display Name")

	businessDiscovery := &apiModels.BusinessDiscovery{
		ID:   "17841400000000000",
		IgID: 123456789,
	}

	cases := []struct {
		name             string
		caption          string
		expectedHashtags []string
	}{
		{
			name:             "multiple hashtags",
			caption:          "#travel #photography #nature",
			expectedHashtags: []string{"travel", "photography", "nature"},
		},
		{
			name:             "hashtags in sentence",
			caption:          "Check out this #amazing sunset in #bali",
			expectedHashtags: []string{"amazing", "bali"},
		},
		{
			name:             "no hashtags",
			caption:          "Just a regular caption",
			expectedHashtags: []string{},
		},
		{
			name:             "empty caption",
			caption:          "",
			expectedHashtags: []string{},
		},
		{
			name:             "hashtags with numbers",
			caption:          "#summer2024 #photo123",
			expectedHashtags: []string{"summer2024", "photo123"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			media := []apiModels.InstagramMedia{
				{
					ID:        "media123",
					Timestamp: "2024-01-15T10:30:00+0000",
					Caption:   tc.caption,
				},
			}

			posts := parser.ParsePosts(media, businessDiscovery, "")

			if len(posts) != 1 {
				t.Fatalf("expected 1 post, got %d", len(posts))
			}

			if len(posts[0].Hashtags) != len(tc.expectedHashtags) {
				t.Fatalf("expected %d hashtags, got %d", len(tc.expectedHashtags), len(posts[0].Hashtags))
			}

			for i, expected := range tc.expectedHashtags {
				if posts[0].Hashtags[i] != expected {
					t.Fatalf("expected hashtag[%d] = %q, got %q", i, expected, posts[0].Hashtags[i])
				}
			}
		})
	}
}

func TestInstagramCompetitorParser_ParsePosts_InsertedAt(t *testing.T) {
	parser := NewInstagramCompetitorParser("page123", "testuser", "Test Display Name")

	businessDiscovery := &apiModels.BusinessDiscovery{
		ID:   "17841400000000000",
		IgID: 123456789,
	}

	media := []apiModels.InstagramMedia{
		{
			ID:        "media123",
			Timestamp: "2024-01-15T10:30:00+0000",
		},
	}

	beforeParse := time.Now().UTC()
	posts := parser.ParsePosts(media, businessDiscovery, "")
	afterParse := time.Now().UTC()

	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}

	if posts[0].InsertedAt.Before(beforeParse) || posts[0].InsertedAt.After(afterParse) {
		t.Fatalf("InsertedAt %v should be between %v and %v", posts[0].InsertedAt, beforeParse, afterParse)
	}
}

func TestInstagramCompetitorParser_parsePost(t *testing.T) {
	parser := NewInstagramCompetitorParser("page123", "testuser", "Test Display Name")

	businessDiscovery := &apiModels.BusinessDiscovery{
		ID:             "17841400000000000",
		IgID:           123456789,
		FollowersCount: 10000,
		FollowsCount:   500,
		Username:       "testuser",
		Name:           "Test User",
	}

	media := apiModels.InstagramMedia{
		ID:            "media123",
		Timestamp:     "2024-01-15T10:30:00+0000",
		LikeCount:     100,
		CommentsCount: 25,
		Caption:       "#test post",
		MediaType:     "IMAGE",
		Permalink:     "https://instagram.com/p/abc",
	}

	profileImage := "https://example.com/profile.jpg"

	post := parser.parsePost(media, businessDiscovery, profileImage)

	if post == nil {
		t.Fatal("expected non-nil post")
	}

	if post.LikeCount != 100 {
		t.Fatalf("expected LikeCount 100, got %d", post.LikeCount)
	}

	if post.CommentsCount != 25 {
		t.Fatalf("expected CommentsCount 25, got %d", post.CommentsCount)
	}

	if post.Engagement != 125 {
		t.Fatalf("expected Engagement 125, got %d", post.Engagement)
	}

	if len(post.Hashtags) != 1 || post.Hashtags[0] != "test" {
		t.Fatalf("expected hashtags [test], got %v", post.Hashtags)
	}
}
