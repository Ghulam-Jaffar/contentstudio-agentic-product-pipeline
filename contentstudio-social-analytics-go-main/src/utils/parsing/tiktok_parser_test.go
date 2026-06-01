package parsing

import (
	"encoding/json"
	"testing"
)

func TestNewTikTokParser(t *testing.T) {
	parser := NewTikTokParser()

	if parser == nil {
		t.Fatal("expected non-nil parser")
	}

	if parser.hashtagRegex == nil {
		t.Fatal("expected hashtagRegex to be initialized")
	}
}

func TestTikTokParser_ParseUserInfo(t *testing.T) {
	parser := NewTikTokParser()

	cases := []struct {
		name      string
		rawJSON   string
		checkFunc func(*TikTokUserInfo) bool
		expectErr bool
	}{
		{
			name: "parses complete user info",
			rawJSON: `{
				"open_id": "open123",
				"union_id": "union456",
				"avatar_url": "https://example.com/avatar.jpg",
				"avatar_url_100": "https://example.com/avatar100.jpg",
				"avatar_large_url": "https://example.com/avatar_large.jpg",
				"display_name": "Test User",
				"bio_description": "This is my bio",
				"profile_deep_link": "https://tiktok.com/@testuser",
				"is_verified": true,
				"follower_count": 10000,
				"following_count": 500,
				"likes_count": 50000,
				"video_count": 100
			}`,
			checkFunc: func(u *TikTokUserInfo) bool {
				return u.OpenID == "open123" &&
					u.UnionID == "union456" &&
					u.DisplayName == "Test User" &&
					u.IsVerified == true &&
					u.FollowerCount == 10000 &&
					u.FollowingCount == 500 &&
					u.LikesCount == 50000 &&
					u.VideoCount == 100
			},
			expectErr: false,
		},
		{
			name: "parses minimal user info",
			rawJSON: `{
				"open_id": "open123",
				"display_name": "Minimal User"
			}`,
			checkFunc: func(u *TikTokUserInfo) bool {
				return u.OpenID == "open123" &&
					u.DisplayName == "Minimal User" &&
					u.FollowerCount == 0
			},
			expectErr: false,
		},
		{
			name:      "invalid JSON returns error",
			rawJSON:   `{invalid}`,
			expectErr: true,
		},
		{
			name: "handles unverified user",
			rawJSON: `{
				"open_id": "open123",
				"is_verified": false
			}`,
			checkFunc: func(u *TikTokUserInfo) bool {
				return u.IsVerified == false
			},
			expectErr: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseUserInfo(json.RawMessage(tc.rawJSON))

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if !tc.checkFunc(result) {
				t.Fatal("check function failed")
			}
		})
	}
}

func TestTikTokParser_ParseVideo(t *testing.T) {
	parser := NewTikTokParser()

	userInfo := &TikTokUserInfo{
		DisplayName:     "Test User",
		ProfileDeepLink: "https://tiktok.com/@testuser",
	}

	cases := []struct {
		name      string
		rawJSON   string
		checkFunc func(*testing.T, interface{})
		expectErr bool
	}{
		{
			name: "parses complete video",
			rawJSON: `{
				"id": "video123",
				"create_time": 1704067200,
				"cover_image_url": "https://example.com/cover.jpg",
				"share_url": "https://tiktok.com/@user/video/123",
				"video_description": "Check out this #funny #video",
				"duration": 30,
				"height": 1920,
				"width": 1080,
				"title": "Funny Video",
				"embed_html": "<iframe></iframe>",
				"embed_link": "https://embed.tiktok.com/123",
				"like_count": 1000,
				"comment_count": 100,
				"share_count": 50,
				"view_count": 10000
			}`,
			checkFunc: func(t *testing.T, result interface{}) {
				post := result.(*interface{})
				if post == nil {
					t.Fatal("expected non-nil post")
				}
			},
			expectErr: false,
		},
		{
			name: "parses video with hashtags in title",
			rawJSON: `{
				"id": "video123",
				"create_time": 1704067200,
				"title": "#trending #viral content",
				"video_description": ""
			}`,
			expectErr: false,
		},
		{
			name:      "invalid JSON returns error",
			rawJSON:   `{invalid}`,
			expectErr: true,
		},
		{
			name: "parses video with zero counts",
			rawJSON: `{
				"id": "video123",
				"create_time": 1704067200,
				"like_count": 0,
				"comment_count": 0,
				"share_count": 0,
				"view_count": 0
			}`,
			expectErr: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseVideo(json.RawMessage(tc.rawJSON), userInfo, "tiktok123")

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}
		})
	}
}

func TestTikTokParser_ParseVideo_FieldValues(t *testing.T) {
	parser := NewTikTokParser()

	userInfo := &TikTokUserInfo{
		DisplayName:     "Test User",
		ProfileDeepLink: "https://tiktok.com/@testuser",
	}

	rawJSON := `{
		"id": "video123",
		"create_time": 1704067200,
		"cover_image_url": "https://example.com/cover.jpg",
		"share_url": "https://tiktok.com/@user/video/123",
		"video_description": "Check out this #funny #video content",
		"duration": 30,
		"height": 1920,
		"width": 1080,
		"title": "Funny Video",
		"embed_html": "<iframe></iframe>",
		"embed_link": "https://embed.tiktok.com/123",
		"like_count": 1000,
		"comment_count": 100,
		"share_count": 50,
		"view_count": 10000
	}`

	result, err := parser.ParseVideo(json.RawMessage(rawJSON), userInfo, "tiktok123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "video123" {
		t.Fatalf("expected ID 'video123', got %q", result.ID)
	}

	if result.DisplayName != "Test User" {
		t.Fatalf("expected DisplayName 'Test User', got %q", result.DisplayName)
	}

	if result.ProfileLink != "https://tiktok.com/@testuser" {
		t.Fatalf("expected ProfileLink, got %q", result.ProfileLink)
	}

	if result.CoverImageURL != "https://example.com/cover.jpg" {
		t.Fatalf("expected CoverImageURL, got %q", result.CoverImageURL)
	}

	if result.ShareURL != "https://tiktok.com/@user/video/123" {
		t.Fatalf("expected ShareURL, got %q", result.ShareURL)
	}

	if result.Duration != 30 {
		t.Fatalf("expected Duration 30, got %d", result.Duration)
	}

	if result.Height != 1920 {
		t.Fatalf("expected Height 1920, got %d", result.Height)
	}

	if result.Width != 1080 {
		t.Fatalf("expected Width 1080, got %d", result.Width)
	}

	if result.LikeCount != 1000 {
		t.Fatalf("expected LikeCount 1000, got %d", result.LikeCount)
	}

	if result.CommentCount != 100 {
		t.Fatalf("expected CommentCount 100, got %d", result.CommentCount)
	}

	if result.ShareCount != 50 {
		t.Fatalf("expected ShareCount 50, got %d", result.ShareCount)
	}

	if result.ViewCount != 10000 {
		t.Fatalf("expected ViewCount 10000, got %d", result.ViewCount)
	}

	expectedEngagement := int64(1000 + 100 + 50)
	if result.EngagementCount != expectedEngagement {
		t.Fatalf("expected EngagementCount %d, got %d", expectedEngagement, result.EngagementCount)
	}

	expectedRate := float64(expectedEngagement) / float64(10000)
	if result.EngagementRate != expectedRate {
		t.Fatalf("expected EngagementRate %f, got %f", expectedRate, result.EngagementRate)
	}

	if len(result.Hashtags) != 2 {
		t.Fatalf("expected 2 hashtags, got %d", len(result.Hashtags))
	}
}

func TestTikTokParser_ParseVideo_HashtagExtraction(t *testing.T) {
	parser := NewTikTokParser()

	userInfo := &TikTokUserInfo{
		DisplayName:     "Test User",
		ProfileDeepLink: "https://tiktok.com/@testuser",
	}

	cases := []struct {
		name             string
		rawJSON          string
		expectedHashtags []string
	}{
		{
			name: "extracts hashtags from description",
			rawJSON: `{
				"id": "video1",
				"create_time": 1704067200,
				"video_description": "#funny #viral #trending"
			}`,
			expectedHashtags: []string{"funny", "viral", "trending"},
		},
		{
			name: "extracts hashtags from title when description empty",
			rawJSON: `{
				"id": "video1",
				"create_time": 1704067200,
				"video_description": "",
				"title": "#comedy #lol"
			}`,
			expectedHashtags: []string{"comedy", "lol"},
		},
		{
			name: "handles no hashtags",
			rawJSON: `{
				"id": "video1",
				"create_time": 1704067200,
				"video_description": "Just a regular video"
			}`,
			expectedHashtags: []string{},
		},
		{
			name: "handles mixed content with hashtags",
			rawJSON: `{
				"id": "video1",
				"create_time": 1704067200,
				"video_description": "Check out #this amazing #content here"
			}`,
			expectedHashtags: []string{"this", "content"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseVideo(json.RawMessage(tc.rawJSON), userInfo, "tiktok123")

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Hashtags) != len(tc.expectedHashtags) {
				t.Fatalf("expected %d hashtags, got %d", len(tc.expectedHashtags), len(result.Hashtags))
			}

			for i, expected := range tc.expectedHashtags {
				if result.Hashtags[i] != expected {
					t.Fatalf("expected hashtag[%d] = %q, got %q", i, expected, result.Hashtags[i])
				}
			}
		})
	}
}

func TestTikTokParser_ParseVideo_EngagementRate(t *testing.T) {
	parser := NewTikTokParser()

	userInfo := &TikTokUserInfo{
		DisplayName: "Test User",
	}

	cases := []struct {
		name         string
		rawJSON      string
		expectedRate float64
	}{
		{
			name: "calculates engagement rate correctly",
			rawJSON: `{
				"id": "video1",
				"create_time": 1704067200,
				"like_count": 100,
				"comment_count": 50,
				"share_count": 25,
				"view_count": 1000
			}`,
			expectedRate: 0.175,
		},
		{
			name: "zero views results in zero rate",
			rawJSON: `{
				"id": "video1",
				"create_time": 1704067200,
				"like_count": 100,
				"comment_count": 50,
				"share_count": 25,
				"view_count": 0
			}`,
			expectedRate: 0,
		},
		{
			name: "high engagement rate",
			rawJSON: `{
				"id": "video1",
				"create_time": 1704067200,
				"like_count": 900,
				"comment_count": 50,
				"share_count": 50,
				"view_count": 1000
			}`,
			expectedRate: 1.0,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseVideo(json.RawMessage(tc.rawJSON), userInfo, "tiktok123")

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.EngagementRate != tc.expectedRate {
				t.Fatalf("expected EngagementRate %f, got %f", tc.expectedRate, result.EngagementRate)
			}
		})
	}
}

func TestTikTokParser_GenerateInsights(t *testing.T) {
	parser := NewTikTokParser()

	userInfo := &TikTokUserInfo{
		OpenID:          "open123",
		DisplayName:     "Test User",
		AvatarLargeURL:  "https://example.com/avatar_large.jpg",
		ProfileDeepLink: "https://tiktok.com/@testuser",
		BioDescription:  "This is my bio",
		IsVerified:      true,
		FollowerCount:   10000,
		FollowingCount:  500,
		LikesCount:      50000,
		VideoCount:      100,
	}

	insights := parser.GenerateInsights(
		userInfo,
		"tiktok123",
		"tiktok123",
		100000,
		25000,
		5000,
		2000,
	)

	if insights == nil {
		t.Fatal("expected non-nil insights")
	}

	if insights.TikTokID != "tiktok123" {
		t.Fatalf("expected TikTokID 'tiktok123', got %q", insights.TikTokID)
	}

	if insights.DisplayName != "Test User" {
		t.Fatalf("expected DisplayName 'Test User', got %q", insights.DisplayName)
	}

	if insights.ProfileImage != "https://example.com/avatar_large.jpg" {
		t.Fatalf("expected ProfileImage, got %q", insights.ProfileImage)
	}

	if insights.ProfileLink != "https://tiktok.com/@testuser" {
		t.Fatalf("expected ProfileLink, got %q", insights.ProfileLink)
	}

	if insights.Bio != "This is my bio" {
		t.Fatalf("expected Bio, got %q", insights.Bio)
	}

	if insights.IsVerified != true {
		t.Fatal("expected IsVerified to be true")
	}

	if insights.TotalFollowerCount != 10000 {
		t.Fatalf("expected TotalFollowerCount 10000, got %d", insights.TotalFollowerCount)
	}

	if insights.TotalFollowingCount != 500 {
		t.Fatalf("expected TotalFollowingCount 500, got %d", insights.TotalFollowingCount)
	}

	if insights.TotalLikeCount != 50000 {
		t.Fatalf("expected TotalLikeCount 50000, got %d", insights.TotalLikeCount)
	}

	if insights.TotalVideoCount != 100 {
		t.Fatalf("expected TotalVideoCount 100, got %d", insights.TotalVideoCount)
	}

	if insights.TotalVideoViews != 100000 {
		t.Fatalf("expected TotalVideoViews 100000, got %d", insights.TotalVideoViews)
	}

	if insights.TotalVideoLikes != 25000 {
		t.Fatalf("expected TotalVideoLikes 25000, got %d", insights.TotalVideoLikes)
	}

	if insights.TotalVideoComments != 5000 {
		t.Fatalf("expected TotalVideoComments 5000, got %d", insights.TotalVideoComments)
	}

	if insights.TotalVideoShares != 2000 {
		t.Fatalf("expected TotalVideoShares 2000, got %d", insights.TotalVideoShares)
	}

	if insights.RecordID == "" {
		t.Fatal("expected RecordID to be generated")
	}

	if len(insights.RecordID) != 32 {
		t.Fatalf("expected RecordID to be 32 character MD5 hash, got %d characters", len(insights.RecordID))
	}

	if insights.InsertedAt == 0 {
		t.Fatal("expected InsertedAt to be set")
	}
}

func TestValidateScopes(t *testing.T) {
	cases := []struct {
		name        string
		scopeString string
		expected    bool
	}{
		{
			name:        "all required scopes present",
			scopeString: "user.info.basic,user.info.profile,user.info.stats,video.list",
			expected:    true,
		},
		{
			name:        "all required scopes with extra scopes",
			scopeString: "user.info.basic,user.info.profile,user.info.stats,video.list,extra.scope",
			expected:    true,
		},
		{
			name:        "missing user.info.basic",
			scopeString: "user.info.profile,user.info.stats,video.list",
			expected:    false,
		},
		{
			name:        "missing user.info.profile",
			scopeString: "user.info.basic,user.info.stats,video.list",
			expected:    false,
		},
		{
			name:        "missing user.info.stats",
			scopeString: "user.info.basic,user.info.profile,video.list",
			expected:    false,
		},
		{
			name:        "missing video.list",
			scopeString: "user.info.basic,user.info.profile,user.info.stats",
			expected:    false,
		},
		{
			name:        "empty string returns false",
			scopeString: "",
			expected:    false,
		},
		{
			name:        "whitespace handling",
			scopeString: " user.info.basic , user.info.profile , user.info.stats , video.list ",
			expected:    true,
		},
		{
			name:        "single scope returns false",
			scopeString: "user.info.basic",
			expected:    false,
		},
		{
			name:        "unrelated scopes return false",
			scopeString: "other.scope,another.scope",
			expected:    false,
		},
		{
			name:        "duplicate scopes still work",
			scopeString: "user.info.basic,user.info.basic,user.info.profile,user.info.stats,video.list",
			expected:    true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateScopes(tc.scopeString)

			if result != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestTikTokUserInfo_Struct(t *testing.T) {
	userInfo := TikTokUserInfo{
		OpenID:          "open123",
		UnionID:         "union456",
		AvatarURL:       "https://example.com/avatar.jpg",
		AvatarURL100:    "https://example.com/avatar100.jpg",
		AvatarLargeURL:  "https://example.com/avatar_large.jpg",
		DisplayName:     "Test User",
		BioDescription:  "Bio text",
		ProfileDeepLink: "https://tiktok.com/@user",
		IsVerified:      true,
		FollowerCount:   10000,
		FollowingCount:  500,
		LikesCount:      50000,
		VideoCount:      100,
	}

	if userInfo.OpenID != "open123" {
		t.Fatalf("expected OpenID 'open123', got %q", userInfo.OpenID)
	}

	if userInfo.UnionID != "union456" {
		t.Fatalf("expected UnionID 'union456', got %q", userInfo.UnionID)
	}

	if userInfo.IsVerified != true {
		t.Fatal("expected IsVerified to be true")
	}

	if userInfo.FollowerCount != 10000 {
		t.Fatalf("expected FollowerCount 10000, got %d", userInfo.FollowerCount)
	}
}
