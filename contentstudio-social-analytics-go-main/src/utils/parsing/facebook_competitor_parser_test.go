package parsing

import (
	"context"
	"testing"
	"time"

	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	models "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

type mockFacebookClient struct {
	GetCompetitorSharedPostDetailsFunc func(ctx context.Context, parentID string, accessToken string) (*apiModels.Post, error)
	GetCompetitorPagePictureFunc       func(ctx context.Context, pageID string, accessToken string) (*apiModels.Picture, error)
}

func (m *mockFacebookClient) GetCompetitorSharedPostDetails(ctx context.Context, parentID string, accessToken string) (*apiModels.Post, error) {
	if m.GetCompetitorSharedPostDetailsFunc != nil {
		return m.GetCompetitorSharedPostDetailsFunc(ctx, parentID, accessToken)
	}
	return nil, nil
}

func (m *mockFacebookClient) GetCompetitorPagePicture(ctx context.Context, pageID string, accessToken string) (*apiModels.Picture, error) {
	if m.GetCompetitorPagePictureFunc != nil {
		return m.GetCompetitorPagePictureFunc(ctx, pageID, accessToken)
	}
	return nil, nil
}

func TestNewFacebookCompetitorParser(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "token123")

	if parser == nil {
		t.Fatal("expected non-nil parser")
	}

	if parser.pageID != "page123" {
		t.Fatalf("expected pageID 'page123', got %q", parser.pageID)
	}

	if parser.pageName != "Test Page" {
		t.Fatalf("expected pageName 'Test Page', got %q", parser.pageName)
	}

	if parser.accessToken != "token123" {
		t.Fatalf("expected accessToken 'token123', got %q", parser.accessToken)
	}
}

func TestFacebookCompetitorParser_ParsePageInsights(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	pageDetails := &apiModels.FacebookPageDetails{
		FanCount:          10000,
		TalkingAboutCount: 500,
		About:             "This is our page",
		Category:          "Technology",
		FollowersCount:    12000,
		Emails:            []string{"contact@example.com"},
		Birthday:          "01/01/2020",
		WereHereCount:     1000,
		Link:              "https://facebook.com/testpage",
		Cover: &apiModels.Cover{
			Source: "https://example.com/cover.jpg",
		},
	}

	picture := &apiModels.Picture{
		Data: &apiModels.PictureData{
			URL: "https://example.com/profile.jpg",
		},
	}

	result := parser.ParsePageInsights(pageDetails, picture)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got %q", result.PageID)
	}

	if result.PageName != "Test Page" {
		t.Fatalf("expected PageName 'Test Page', got %q", result.PageName)
	}

	if result.TotalFanCount != 10000 {
		t.Fatalf("expected TotalFanCount 10000, got %d", result.TotalFanCount)
	}

	if result.TalkingAboutThis != 500 {
		t.Fatalf("expected TalkingAboutThis 500, got %d", result.TalkingAboutThis)
	}

	if result.Biography != "This is our page" {
		t.Fatalf("expected Biography, got %q", result.Biography)
	}

	if result.PageCategory != "Technology" {
		t.Fatalf("expected PageCategory 'Technology', got %q", result.PageCategory)
	}

	if result.FollowersCount != 12000 {
		t.Fatalf("expected FollowersCount 12000, got %d", result.FollowersCount)
	}

	if result.ProfilePictureURL != "https://example.com/profile.jpg" {
		t.Fatalf("expected ProfilePictureURL, got %q", result.ProfilePictureURL)
	}

	if result.CoverPhotoURL != "https://example.com/cover.jpg" {
		t.Fatalf("expected CoverPhotoURL, got %q", result.CoverPhotoURL)
	}

	if result.RecordID == "" {
		t.Fatal("expected RecordID to be generated")
	}
}

func TestFacebookCompetitorParser_ParsePageInsights_NilPicture(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	pageDetails := &apiModels.FacebookPageDetails{
		FanCount: 10000,
	}

	result := parser.ParsePageInsights(pageDetails, nil)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.ProfilePictureURL != "" {
		t.Fatalf("expected empty ProfilePictureURL, got %q", result.ProfilePictureURL)
	}
}

func TestFacebookCompetitorParser_parseEngagements(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	post := &apiModels.Post{
		Like: &apiModels.ReactionSummary{
			Summary: &apiModels.Summary{
				TotalCount: 100,
			},
		},
		Love: &apiModels.ReactionSummary{
			Summary: &apiModels.Summary{
				TotalCount: 50,
			},
		},
		Haha: &apiModels.ReactionSummary{
			Summary: &apiModels.Summary{
				TotalCount: 25,
			},
		},
		Wow: &apiModels.ReactionSummary{
			Summary: &apiModels.Summary{
				TotalCount: 10,
			},
		},
		Sad: &apiModels.ReactionSummary{
			Summary: &apiModels.Summary{
				TotalCount: 5,
			},
		},
		Angry: &apiModels.ReactionSummary{
			Summary: &apiModels.Summary{
				TotalCount: 3,
			},
		},
		Comments: &apiModels.CommentSummary{
			Summary: &apiModels.Summary{
				TotalCount: 75,
			},
		},
		Shares: &apiModels.ShareCount{
			Count: 30,
		},
	}

	engagements := parser.parseEngagements(post)

	if engagements["like"] != 100 {
		t.Fatalf("expected like 100, got %d", engagements["like"])
	}

	if engagements["love"] != 50 {
		t.Fatalf("expected love 50, got %d", engagements["love"])
	}

	if engagements["haha"] != 25 {
		t.Fatalf("expected haha 25, got %d", engagements["haha"])
	}

	if engagements["wow"] != 10 {
		t.Fatalf("expected wow 10, got %d", engagements["wow"])
	}

	if engagements["sad"] != 5 {
		t.Fatalf("expected sad 5, got %d", engagements["sad"])
	}

	if engagements["angry"] != 3 {
		t.Fatalf("expected angry 3, got %d", engagements["angry"])
	}

	if engagements["comments"] != 75 {
		t.Fatalf("expected comments 75, got %d", engagements["comments"])
	}

	if engagements["shares"] != 30 {
		t.Fatalf("expected shares 30, got %d", engagements["shares"])
	}

	expectedTotalReactions := int64(100 + 50 + 25 + 10 + 5 + 3)
	if engagements["total_reactions"] != expectedTotalReactions {
		t.Fatalf("expected total_reactions %d, got %d", expectedTotalReactions, engagements["total_reactions"])
	}

	expectedTotalEngagement := expectedTotalReactions + 75 + 30
	if engagements["post_engagement"] != expectedTotalEngagement {
		t.Fatalf("expected post_engagement %d, got %d", expectedTotalEngagement, engagements["post_engagement"])
	}
}

func TestFacebookCompetitorParser_determineMediaType(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	cases := []struct {
		name               string
		post               *apiModels.Post
		expectedMediaType  string
		expectedStatusType string
	}{
		{
			name: "carousel with child attachments",
			post: &apiModels.Post{
				ChildAttachments: []*apiModels.ChildAttachment{{ID: "child1"}},
			},
			expectedMediaType:  "carousel",
			expectedStatusType: "",
		},
		{
			name: "share with child attachments and parent ID",
			post: &apiModels.Post{
				ChildAttachments: []*apiModels.ChildAttachment{{ID: "child1"}},
				ParentID:         "parent123",
			},
			expectedMediaType:  "share",
			expectedStatusType: "",
		},
		{
			name: "text post without attachments",
			post: &apiModels.Post{
				Attachments: nil,
			},
			expectedMediaType:  "text",
			expectedStatusType: "",
		},
		{
			name: "added_video status type",
			post: &apiModels.Post{
				StatusType: "added_video",
				Attachments: &apiModels.Attachments{
					Data: []*apiModels.AttachmentData{{Type: "video"}},
				},
			},
			expectedMediaType:  "videos",
			expectedStatusType: "",
		},
		{
			name: "added_photos status type",
			post: &apiModels.Post{
				StatusType: "added_photos",
				Attachments: &apiModels.Attachments{
					Data: []*apiModels.AttachmentData{{Type: "photo"}},
				},
			},
			expectedMediaType:  "image",
			expectedStatusType: "",
		},
		{
			name: "shared_story status type",
			post: &apiModels.Post{
				StatusType: "shared_story",
				Attachments: &apiModels.Attachments{
					Data: []*apiModels.AttachmentData{{Type: "link"}},
				},
			},
			expectedMediaType:  "link",
			expectedStatusType: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mediaType, statusType := parser.determineMediaType(tc.post)

			if mediaType != tc.expectedMediaType {
				t.Fatalf("expected mediaType %q, got %q", tc.expectedMediaType, mediaType)
			}

			if statusType != tc.expectedStatusType {
				t.Fatalf("expected statusType %q, got %q", tc.expectedStatusType, statusType)
			}
		})
	}
}

func TestFacebookCompetitorParser_mapMediaType(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	cases := []struct {
		name           string
		attachmentType string
		expected       string
	}{
		{
			name:           "multi_share_no_end_card is carousel",
			attachmentType: "multi_share_no_end_card",
			expected:       "carousel",
		},
		{
			name:           "album is carousel",
			attachmentType: "album",
			expected:       "carousel",
		},
		{
			name:           "photo is image",
			attachmentType: "photo",
			expected:       "image",
		},
		{
			name:           "video is videos",
			attachmentType: "video",
			expected:       "videos",
		},
		{
			name:           "video_inline is videos",
			attachmentType: "video_inline",
			expected:       "videos",
		},
		{
			name:           "link is link",
			attachmentType: "link",
			expected:       "link",
		},
		{
			name:           "share is link",
			attachmentType: "share",
			expected:       "link",
		},
		{
			name:           "unknown type is others",
			attachmentType: "unknown",
			expected:       "others",
		},
		{
			name:           "empty type is others",
			attachmentType: "",
			expected:       "others",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.mapMediaType(tc.attachmentType)

			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestFacebookCompetitorParser_ParsePosts(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	posts := []*apiModels.Post{
		{
			ID:          "post123",
			Message:     "Test post #hashtag",
			CreatedTime: "2024-01-15T10:30:00+0000",
			StatusType:  "added_photos",
		},
		{
			ID:          "post456",
			Message:     "Another post",
			CreatedTime: "2024-01-16T10:30:00+0000",
			StatusType:  "added_video",
		},
	}

	pageDetails := &apiModels.FacebookPageDetails{
		FanCount:       10000,
		FollowersCount: 12000,
		Category:       "Technology",
	}

	competitorPosts, mediaAssets := parser.ParsePosts(context.Background(), posts, pageDetails)

	if len(competitorPosts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(competitorPosts))
	}

	if competitorPosts[0].PostID != "post123" {
		t.Fatalf("expected PostID 'post123', got %q", competitorPosts[0].PostID)
	}

	if competitorPosts[0].Caption != "Test post #hashtag" {
		t.Fatalf("expected Caption, got %q", competitorPosts[0].Caption)
	}

	if competitorPosts[0].FanCount != 10000 {
		t.Fatalf("expected FanCount 10000, got %d", competitorPosts[0].FanCount)
	}

	if competitorPosts[0].FollowersCount != 12000 {
		t.Fatalf("expected FollowersCount 12000, got %d", competitorPosts[0].FollowersCount)
	}

	if len(competitorPosts[0].Hashtags) != 1 || competitorPosts[0].Hashtags[0] != "hashtag" {
		t.Fatalf("expected hashtags [hashtag], got %v", competitorPosts[0].Hashtags)
	}

	_ = mediaAssets // mediaAssets can be nil or empty depending on post structure
}

func TestFacebookCompetitorParser_parseMediaAssets_ChildAttachments(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	post := &apiModels.Post{
		ID:          "post123",
		CreatedTime: "2024-01-15T10:30:00+0000",
		ChildAttachments: []*apiModels.ChildAttachment{
			{
				ID:          "child1",
				Caption:     "First image",
				Description: "Description 1",
				Picture:     "https://example.com/pic1.jpg",
				Link:        "https://example.com/link1",
			},
			{
				ID:          "child2",
				Caption:     "Second image",
				Description: "Description 2",
				Picture:     "https://example.com/pic2.jpg",
			},
		},
	}

	assets := parser.parseMediaAssets(post)

	if len(assets) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(assets))
	}

	if assets[0].MediaID != "child1" {
		t.Fatalf("expected MediaID 'child1', got %q", assets[0].MediaID)
	}

	if assets[0].PostID != "post123" {
		t.Fatalf("expected PostID 'post123', got %q", assets[0].PostID)
	}

	if assets[0].PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got %q", assets[0].PageID)
	}

	if assets[0].Caption != "First image" {
		t.Fatalf("expected Caption 'First image', got %q", assets[0].Caption)
	}

	if assets[0].Link != "https://example.com/pic1.jpg" {
		t.Fatalf("expected Link, got %q", assets[0].Link)
	}
}

func TestFacebookCompetitorParser_parseMediaAssets_NoAttachments(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	post := &apiModels.Post{
		ID:          "post123",
		CreatedTime: "2024-01-15T10:30:00+0000",
		Attachments: nil,
	}

	assets := parser.parseMediaAssets(post)

	if len(assets) != 0 {
		t.Fatalf("expected 0 assets, got %d", len(assets))
	}
}

func TestFacebookCompetitorParser_extractPageIDFromURL(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	cases := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "extracts from URL with ID parameter",
			url:      "https://www.facebook.com/photo.php?id=123456",
			expected: "123456",
		},
		{
			name:     "extracts from story_fbid URL (contains id=)",
			url:      "https://www.facebook.com/permalink.php?story_fbid=123",
			expected: "123",
		},
		{
			name:     "handles empty URL",
			url:      "",
			expected: "",
		},
		{
			name:     "returns empty for URL without facebook.com segment or id parameter",
			url:      "https://example.com/page",
			expected: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.extractPageIDFromURL(tc.url)

			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestFacebookCompetitorParser_generateMediaID(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	result1 := parser.generateMediaID("post123", 0)
	result2 := parser.generateMediaID("post123", 1)
	result3 := parser.generateMediaID("post456", 0)

	if len(result1) != 32 {
		t.Fatalf("expected 32 character hash, got %d", len(result1))
	}

	if result1 == result2 {
		t.Fatal("expected different hashes for different indices")
	}

	if result1 == result3 {
		t.Fatal("expected different hashes for different post IDs")
	}

	result4 := parser.generateMediaID("post123", 0)
	if result1 != result4 {
		t.Fatal("expected same hash for same inputs")
	}
}

func TestExtractHashtags(t *testing.T) {
	cases := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "extracts single hashtag",
			text:     "Check out #summer",
			expected: []string{"summer"},
		},
		{
			name:     "extracts multiple hashtags",
			text:     "#hello #world #test",
			expected: []string{"hello", "world", "test"},
		},
		{
			name:     "handles empty text",
			text:     "",
			expected: []string{},
		},
		{
			name:     "handles text without hashtags",
			text:     "No hashtags here",
			expected: []string{},
		},
		{
			name:     "extracts hashtags with underscores",
			text:     "#hello_world #test_123",
			expected: []string{"hello_world", "test_123"},
		},
		{
			name:     "handles mixed case hashtags",
			text:     "#Hello #WORLD #TeSt",
			expected: []string{"Hello", "WORLD", "TeSt"},
		},
		{
			name:     "extracts hashtags from middle of text",
			text:     "Check out this #awesome post about #coding",
			expected: []string{"awesome", "coding"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := extractHashtags(tc.text)

			if len(result) != len(tc.expected) {
				t.Fatalf("expected %d hashtags, got %d", len(tc.expected), len(result))
			}

			for i, expected := range tc.expected {
				if result[i] != expected {
					t.Fatalf("expected hashtag[%d] = %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

func TestGenerateRecordID(t *testing.T) {
	now := time.Now()

	result1 := generateRecordID("page123", now)
	result2 := generateRecordID("page456", now)
	result3 := generateRecordID("page123", now.Add(24*time.Hour))

	if len(result1) != 32 {
		t.Fatalf("expected 32 character hash, got %d", len(result1))
	}

	if result1 == result2 {
		t.Fatal("expected different hashes for different page IDs")
	}

	if result1 == result3 {
		t.Fatal("expected different hashes for different dates")
	}

	result4 := generateRecordID("page123", now)
	if result1 != result4 {
		t.Fatal("expected same hash for same inputs")
	}
}

func TestFacebookCompetitorParser_determineAssetMediaType(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	cases := []struct {
		name           string
		post           *apiModels.Post
		attachmentData *apiModels.AttachmentData
		subAttachment  *apiModels.AttachmentData
		expected       string
	}{
		{
			name: "video status type",
			post: &apiModels.Post{
				StatusType: "added_video",
			},
			expected: "video",
		},
		{
			name: "photo media type",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				MediaType: "photo",
			},
			expected: "image",
		},
		{
			name: "video media type",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				MediaType: "video",
			},
			expected: "video",
		},
		{
			name: "link media type",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				MediaType: "link",
			},
			expected: "link",
		},
		{
			name: "subattachment type photo",
			post: &apiModels.Post{},
			subAttachment: &apiModels.AttachmentData{
				Type: "photo",
			},
			expected: "image",
		},
		{
			name: "subattachment type video",
			post: &apiModels.Post{},
			subAttachment: &apiModels.AttachmentData{
				Type: "video",
			},
			expected: "video",
		},
		{
			name: "default to image with media",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				Media: &apiModels.FacebookMedia{
					Image: &apiModels.MediaImage{
						Src: "https://example.com/image.jpg",
					},
				},
			},
			expected: "image",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.determineAssetMediaType(tc.post, tc.attachmentData, tc.subAttachment)

			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestFacebookCompetitorParser_parseSharedPostInfo_WithClient(t *testing.T) {
	mockClient := &mockFacebookClient{
		GetCompetitorSharedPostDetailsFunc: func(ctx context.Context, parentID string, accessToken string) (*apiModels.Post, error) {
			return &apiModels.Post{
				From: &apiModels.From{
					ID:   "sharer123",
					Name: "Sharer Page",
				},
				CreatedTime: "2024-01-10T10:30:00+0000",
			}, nil
		},
		GetCompetitorPagePictureFunc: func(ctx context.Context, pageID string, accessToken string) (*apiModels.Picture, error) {
			return &apiModels.Picture{
				Data: &apiModels.PictureData{
					URL: "https://example.com/sharer.jpg",
				},
			}, nil
		},
	}

	parser := NewFacebookCompetitorParser("page123", "Test Page", mockClient, "token123")

	post := &apiModels.Post{
		ID:       "post123",
		ParentID: "parent456",
	}

	competitorPost := &models.FacebookCompetitorPosts{}

	parser.parseSharedPostInfo(context.Background(), post, competitorPost)

	if competitorPost.SharedFromName != "Sharer Page" {
		t.Fatalf("expected SharedFromName 'Sharer Page', got %q", competitorPost.SharedFromName)
	}

	if competitorPost.SharedFromID != "sharer123" {
		t.Fatalf("expected SharedFromID 'sharer123', got %q", competitorPost.SharedFromID)
	}

	if competitorPost.SharedFromPic != "https://example.com/sharer.jpg" {
		t.Fatalf("expected SharedFromPic, got %q", competitorPost.SharedFromPic)
	}
}

func TestFacebookCompetitorParser_parseSharedPostInfo_NoParentID(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	post := &apiModels.Post{
		ID:       "post123",
		ParentID: "",
	}

	competitorPost := &models.FacebookCompetitorPosts{}

	parser.parseSharedPostInfo(context.Background(), post, competitorPost)

	if competitorPost.SharedFromName != "" {
		t.Fatalf("expected empty SharedFromName, got %q", competitorPost.SharedFromName)
	}
}

func TestFacebookCompetitorParser_determineChildAttachmentMediaType(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	cases := []struct {
		name           string
		post           *apiModels.Post
		attachmentData *apiModels.AttachmentData
		expected       string
	}{
		{
			name: "added_video status returns video",
			post: &apiModels.Post{
				StatusType: "added_video",
			},
			attachmentData: nil,
			expected:       "video",
		},
		{
			name: "photo media_type returns image",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				MediaType: "photo",
			},
			expected: "image",
		},
		{
			name: "album media_type returns image",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				MediaType: "album",
			},
			expected: "image",
		},
		{
			name: "video media_type returns video",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				MediaType: "video",
			},
			expected: "video",
		},
		{
			name: "video_inline media_type returns video",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				MediaType: "video_inline",
			},
			expected: "video",
		},
		{
			name: "link media_type returns link",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				MediaType: "link",
			},
			expected: "link",
		},
		{
			name: "photo type returns image",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				Type: "photo",
			},
			expected: "image",
		},
		{
			name: "album type returns image",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				Type: "album",
			},
			expected: "image",
		},
		{
			name: "video type returns video",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				Type: "video",
			},
			expected: "video",
		},
		{
			name: "share type returns link",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				Type: "share",
			},
			expected: "link",
		},
		{
			name: "link type returns link",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				Type: "link",
			},
			expected: "link",
		},
		{
			name: "media with Image returns image",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				Media: &apiModels.FacebookMedia{
					Image: &apiModels.MediaImage{
						Src: "https://example.com/image.jpg",
					},
				},
			},
			expected: "image",
		},
		{
			name: "media with Source returns video",
			post: &apiModels.Post{},
			attachmentData: &apiModels.AttachmentData{
				Media: &apiModels.FacebookMedia{
					Source: "https://example.com/video.mp4",
				},
			},
			expected: "video",
		},
		{
			name: "nil attachment returns image",
			post: &apiModels.Post{},
			attachmentData: nil,
			expected:       "image",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.determineChildAttachmentMediaType(tc.post, tc.attachmentData)
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestFacebookCompetitorParser_parseMediaAssets_ChildAttachmentsPointers(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	post := &apiModels.Post{
		ID:          "post123",
		CreatedTime: "2024-01-15T10:30:00+0000",
		ChildAttachments: []*apiModels.ChildAttachment{
			{
				ID:          "child1",
				Caption:     "First image",
				Description: "Description 1",
				Picture:     "https://example.com/image1.jpg",
			},
			{
				ID:          "child2",
				Caption:     "Second image",
				Description: "Description 2",
				Picture:     "https://example.com/image2.jpg",
			},
		},
		Attachments: &apiModels.Attachments{
			Data: []*apiModels.AttachmentData{
				{
					MediaType: "photo",
				},
			},
		},
	}

	assets := parser.parseMediaAssets(post)

	if len(assets) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(assets))
	}

	if assets[0].MediaID != "child1" {
		t.Fatalf("expected MediaID 'child1', got %q", assets[0].MediaID)
	}

	if assets[0].Caption != "First image" {
		t.Fatalf("expected Caption 'First image', got %q", assets[0].Caption)
	}

	if assets[0].AssetType != "image" {
		t.Fatalf("expected AssetType 'image', got %q", assets[0].AssetType)
	}

	if assets[0].PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got %q", assets[0].PageID)
	}
}

func TestFacebookCompetitorParser_parseMediaAssets_SingleAttachment(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	post := &apiModels.Post{
		ID:          "post123",
		Message:     "Test message",
		CreatedTime: "2024-01-15T10:30:00+0000",
		StatusType:  "added_photos",
		Attachments: &apiModels.Attachments{
			Data: []*apiModels.AttachmentData{
				{
					Type:        "photo",
					MediaType:   "photo",
					Description: "Single photo",
					Target: &apiModels.Target{
						ID: "target123",
					},
					Media: &apiModels.FacebookMedia{
						Image: &apiModels.MediaImage{
							Src: "https://example.com/photo.jpg",
						},
					},
				},
			},
		},
	}

	assets := parser.parseMediaAssets(post)

	if len(assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(assets))
	}

	// MediaID may be generated hash if target ID is not used directly
	if assets[0].MediaID == "" {
		t.Fatal("expected non-empty MediaID")
	}

	if assets[0].AssetType != "image" {
		t.Fatalf("expected AssetType 'image', got %q", assets[0].AssetType)
	}

	if assets[0].PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got %q", assets[0].PageID)
	}
}

func TestFacebookCompetitorParser_parseMediaAssets_Empty(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	post := &apiModels.Post{
		ID:          "post123",
		CreatedTime: "2024-01-15T10:30:00+0000",
	}

	assets := parser.parseMediaAssets(post)

	if len(assets) != 0 {
		t.Fatalf("expected 0 assets, got %d", len(assets))
	}
}

func TestFacebookCompetitorParser_parseMediaAssets_Subattachments(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	post := &apiModels.Post{
		ID:          "post123",
		CreatedTime: "2024-01-15T10:30:00+0000",
		StatusType:  "added_photos",
		Attachments: &apiModels.Attachments{
			Data: []*apiModels.AttachmentData{
				{
					Type: "album",
					Subattachments: &apiModels.Subattachments{
						Data: []*apiModels.AttachmentData{
							{
								Type: "photo",
								Media: &apiModels.FacebookMedia{
									Image: &apiModels.MediaImage{
										Src: "https://example.com/sub1.jpg",
									},
								},
							},
							{
								Type: "photo",
								Media: &apiModels.FacebookMedia{
									Image: &apiModels.MediaImage{
										Src: "https://example.com/sub2.jpg",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	assets := parser.parseMediaAssets(post)

	if len(assets) < 2 {
		t.Fatalf("expected at least 2 assets from subattachments, got %d", len(assets))
	}
}

func TestFacebookCompetitorParser_determineMediaType_AllCases(t *testing.T) {
	parser := NewFacebookCompetitorParser("page123", "Test Page", nil, "")

	cases := []struct {
		name     string
		post     *apiModels.Post
		expected string
	}{
		{
			name: "added_video status with attachment",
			post: &apiModels.Post{
				StatusType: "added_video",
				Attachments: &apiModels.Attachments{
					Data: []*apiModels.AttachmentData{
						{Type: "video"},
					},
				},
			},
			expected: "videos",
		},
		{
			name: "added_photos status with attachment",
			post: &apiModels.Post{
				StatusType: "added_photos",
				Attachments: &apiModels.Attachments{
					Data: []*apiModels.AttachmentData{
						{Type: "photo"},
					},
				},
			},
			expected: "image",
		},
		{
			name: "shared_story status with attachment",
			post: &apiModels.Post{
				StatusType: "shared_story",
				Attachments: &apiModels.Attachments{
					Data: []*apiModels.AttachmentData{
						{Type: "link"},
					},
				},
			},
			expected: "link",
		},
		{
			name: "attachment video type",
			post: &apiModels.Post{
				Attachments: &apiModels.Attachments{
					Data: []*apiModels.AttachmentData{
						{Type: "video"},
					},
				},
			},
			expected: "videos",
		},
		{
			name: "attachment photo type",
			post: &apiModels.Post{
				Attachments: &apiModels.Attachments{
					Data: []*apiModels.AttachmentData{
						{Type: "photo"},
					},
				},
			},
			expected: "image",
		},
		{
			name: "attachment album type",
			post: &apiModels.Post{
				Attachments: &apiModels.Attachments{
					Data: []*apiModels.AttachmentData{
						{Type: "album"},
					},
				},
			},
			expected: "carousel",
		},
		{
			name: "attachment share type",
			post: &apiModels.Post{
				Attachments: &apiModels.Attachments{
					Data: []*apiModels.AttachmentData{
						{Type: "share"},
					},
				},
			},
			expected: "link",
		},
		{
			name: "has child attachments",
			post: &apiModels.Post{
				ChildAttachments: []*apiModels.ChildAttachment{
					{ID: "child1"},
				},
			},
			expected: "carousel",
		},
		{
			name: "text only post",
			post: &apiModels.Post{
				Message: "Just text",
			},
			expected: "text",
		},
		{
			name: "multi_share_no_end_card type",
			post: &apiModels.Post{
				Attachments: &apiModels.Attachments{
					Data: []*apiModels.AttachmentData{
						{Type: "multi_share_no_end_card"},
					},
				},
			},
			expected: "carousel",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, _ := parser.determineMediaType(tc.post)
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}
