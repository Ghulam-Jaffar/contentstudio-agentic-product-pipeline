package clickhouse

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFacebookInsights_Struct(t *testing.T) {
	now := time.Now()
	insights := FacebookInsights{
		HashID:       "hash123",
		PageID:       "page456",
		PageCategory: "Business",
		DayOfWeek:    "Monday",
		Year:         2025,
		Month:        1,
		CreatedTime:  now,
		SavingTime:   now,
		PageFans:     10000,
		PageFansCity: []string{"New York:1000", "Los Angeles:800", "Chicago:600"},
		PageFansCountry: []string{"US:5000", "UK:2000", "CA:1000"},
		PageFansLocale:  []string{"en_US:6000", "en_GB:2000"},
		PageFansAge:     []string{"25-34:3000", "35-44:2500", "45-54:1500"},
		PageFansGender:  []string{"M:5500", "F:4500"},
		PageFansGenderAge: []string{"M.25-34:1500", "F.25-34:1500"},
		PageFollows:       9500,
		PageViews:         50000,
		PageFanAddsByPaidNonPaidUnique: []string{"paid:500", "unpaid:1000"},
		PageFanAddsUnique:    1500,
		PageFanRemovesUnique: 200,
		PageImpressions:      100000,
		PageImpressionsUnique: 80000,
		PageVideoViews:        25000,
		PagePostEngagements:   15000,
		ActiveUsers:           5000,
		PostsCount:            100,
		LikesCount:            50000,
	}

	if insights.HashID != "hash123" {
		t.Fatalf("expected HashID 'hash123', got %s", insights.HashID)
	}
	if insights.PageFans != 10000 {
		t.Fatalf("expected PageFans 10000, got %d", insights.PageFans)
	}
	if len(insights.PageFansCity) != 3 {
		t.Fatalf("expected 3 cities, got %d", len(insights.PageFansCity))
	}
}

func TestFacebookInsights_JSON_Marshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	insights := FacebookInsights{
		HashID:      "hash123",
		PageID:      "page456",
		PageFans:    10000,
		CreatedTime: now,
	}

	data, err := json.Marshal(insights)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result FacebookInsights
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.PageFans != insights.PageFans {
		t.Fatalf("expected PageFans %d, got %d", insights.PageFans, result.PageFans)
	}
}

func TestFacebookPosts_Struct(t *testing.T) {
	now := time.Now()
	post := FacebookPosts{
		PageName:         "Test Page",
		PageID:           "page123",
		MediaType:        "photo",
		PostID:           "post456",
		Permalink:        "https://facebook.com/post/456",
		StatusType:       "added_photos",
		VideoID:          "",
		Category:         "Business",
		PublishedBy:      "Admin",
		PublishedByURL:   "https://facebook.com/admin",
		Like:             100,
		Love:             50,
		Haha:             25,
		Wow:              10,
		Sad:              5,
		Angry:            2,
		Thankful:         1,
		Total:            193,
		Shares:           30,
		Comments:         45,
		PostClicks:       150,
		TotalEngagement:  268,
		PostEngagedUsers: 200,
		DayOfWeek:        "Monday",
		HourOfDay:        14,
		CreatedTime:      now,
		UpdatedTime:      now,
		SavingTime:       now,
		MessageTags:      []string{"#test", "#facebook"},
		Caption:          "Test caption",
		Description:      "Test description",
		FullPicture:      "https://example.com/picture.jpg",
		PostImpressions:  10000,
	}

	if post.PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got %s", post.PageID)
	}
	if post.Total != 193 {
		t.Fatalf("expected Total 193, got %d", post.Total)
	}
	if post.TotalEngagement != 268 {
		t.Fatalf("expected TotalEngagement 268, got %d", post.TotalEngagement)
	}
}

func TestFacebookPosts_JSON_Marshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	post := FacebookPosts{
		PageID:      "page123",
		PostID:      "post456",
		Like:        100,
		CreatedTime: now,
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result FacebookPosts
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.Like != post.Like {
		t.Fatalf("expected Like %d, got %d", post.Like, result.Like)
	}
}

func TestFacebookMediaAssets_Struct(t *testing.T) {
	now := time.Now()
	asset := FacebookMediaAssets{
		PageID:       "page123",
		MediaID:      "media456",
		PostID:       "post789",
		AssetType:    "image",
		Link:         "https://example.com/image.jpg",
		CallToAction: "Learn More",
		CTAType:      "LEARN_MORE",
		Caption:      "Test caption",
		Description:  "Test description",
		CreatedAt:    now,
		InsertedAt:   now,
	}

	if asset.PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got %s", asset.PageID)
	}
	if asset.AssetType != "image" {
		t.Fatalf("expected AssetType 'image', got %s", asset.AssetType)
	}
}

func TestFacebookReelsInsights_Struct(t *testing.T) {
	now := time.Now()
	reels := FacebookReelsInsights{
		PageID:               "page123",
		PostID:               "post456",
		AverageTimeWatched:   15000,
		TotalTimeWatchedInMs: 500000,
		PlayCount:            1000,
		ImpressionsUnique:    800,
		ReelFollowers:        50,
		CreatedAt:            now,
		SavingTime:           now,
	}

	if reels.PlayCount != 1000 {
		t.Fatalf("expected PlayCount 1000, got %d", reels.PlayCount)
	}
	if reels.AverageTimeWatched != 15000 {
		t.Fatalf("expected AverageTimeWatched 15000, got %d", reels.AverageTimeWatched)
	}
}

func TestFacebookVideoInsights_Struct(t *testing.T) {
	now := time.Now()
	video := FacebookVideoInsights{
		PostID:                    "post123",
		PageID:                    "page456",
		VideoID:                   "video789",
		CreatedTime:               now,
		UpdatedTime:               now,
		TotalVideoFollowers:       500,
		TotalVideoViews:           10000,
		TotalVideoViewsUnique:     8000,
		TotalVideoViewsAutoplayed: 6000,
		TotalVideoViewsOrganic:    7000,
		TotalVideoViewsPaid:       3000,
		TotalVideoViewsSoundOn:    5000,
		TotalVideoViewsByDistributionType: []string{"organic:7000", "paid:3000"},
		TotalVideoPlayCount:              10000,
		TotalVideoConsumptionRate:        0.75,
		TotalVideoCompleteViews:          5000,
		TotalVideoCompleteViewsUnique:    4500,
		TotalVideo30sViews:               6000,
		TotalVideo10sViews:               8000,
		TotalVideoAvgTimeWatched:         25000,
		TotalVideoViewTotalTime:          250000000,
		TotalVideoImpressions:            50000,
		TotalVideoImpressionsUnique:      40000,
		TotalEngagement:                  15000,
		TotalVideoAdBreakEarnings:        150.50,
		TotalVideoAdBreakAdImpressions:   10000,
	}

	if video.VideoID != "video789" {
		t.Fatalf("expected VideoID 'video789', got %s", video.VideoID)
	}
	if video.TotalVideoViews != 10000 {
		t.Fatalf("expected TotalVideoViews 10000, got %d", video.TotalVideoViews)
	}
	if video.TotalVideoConsumptionRate != 0.75 {
		t.Fatalf("expected TotalVideoConsumptionRate 0.75, got %f", video.TotalVideoConsumptionRate)
	}
}

func TestFacebookVideoInsights_RetentionGraphs(t *testing.T) {
	video := FacebookVideoInsights{
		VideoID: "video123",
		TotalVideoRetentionGraphAutoplayed:    []string{"0:100", "10:90", "20:80", "30:70"},
		TotalVideoRetentionGraphClickedToPlay: []string{"0:100", "10:95", "20:90", "30:85"},
		TotalVideoRetentionGraphGenderMale:    []string{"0:100", "10:88", "20:76"},
		TotalVideoRetentionGraphGenderFemale:  []string{"0:100", "10:92", "20:84"},
	}

	if len(video.TotalVideoRetentionGraphAutoplayed) != 4 {
		t.Fatalf("expected 4 retention points, got %d", len(video.TotalVideoRetentionGraphAutoplayed))
	}
}

func TestMinimalPost_Struct(t *testing.T) {
	post := MinimalPost{
		PageID:      "page123",
		PostID:      "post456",
		FullPicture: "https://example.com/picture.jpg",
	}

	if post.PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got %s", post.PageID)
	}
	if post.FullPicture != "https://example.com/picture.jpg" {
		t.Fatalf("expected FullPicture, got %s", post.FullPicture)
	}
}

func TestFbItem_Struct(t *testing.T) {
	item := FbItem{
		ID:          "item123",
		FullPicture: "https://example.com/picture.jpg",
	}

	if item.ID != "item123" {
		t.Fatalf("expected ID 'item123', got %s", item.ID)
	}
}

func TestFbItem_WithAttachments(t *testing.T) {
	item := FbItem{
		ID:          "post123",
		FullPicture: "https://example.com/main.jpg",
		Attachments: &struct {
			Data []struct {
				MediaType string `json:"media_type"`
				Type      string `json:"type"`
				Target    *struct {
					ID string `json:"id"`
				} `json:"target"`
				Media *struct {
					Image *struct {
						Src           string `json:"src"`
						Width, Height int
					} `json:"image"`
				} `json:"media"`
				Subattachments *struct {
					Data []struct {
						MediaType string `json:"media_type"`
						Type      string `json:"type"`
						Target    *struct {
							ID string `json:"id"`
						} `json:"target"`
						Media *struct {
							Image *struct {
								Src           string `json:"src"`
								Width, Height int
							} `json:"image"`
						} `json:"media"`
					} `json:"data"`
				} `json:"subattachments"`
			} `json:"data"`
		}{
			Data: []struct {
				MediaType string `json:"media_type"`
				Type      string `json:"type"`
				Target    *struct {
					ID string `json:"id"`
				} `json:"target"`
				Media *struct {
					Image *struct {
						Src           string `json:"src"`
						Width, Height int
					} `json:"image"`
				} `json:"media"`
				Subattachments *struct {
					Data []struct {
						MediaType string `json:"media_type"`
						Type      string `json:"type"`
						Target    *struct {
							ID string `json:"id"`
						} `json:"target"`
						Media *struct {
							Image *struct {
								Src           string `json:"src"`
								Width, Height int
							} `json:"image"`
						} `json:"media"`
					} `json:"data"`
				} `json:"subattachments"`
			}{
				{
					MediaType: "photo",
					Type:      "photo",
					Target: &struct {
						ID string `json:"id"`
					}{ID: "target123"},
					Media: &struct {
						Image *struct {
							Src           string `json:"src"`
							Width, Height int
						} `json:"image"`
					}{
						Image: &struct {
							Src           string `json:"src"`
							Width, Height int
						}{Src: "https://example.com/image.jpg", Width: 800, Height: 600},
					},
				},
			},
		},
	}

	if len(item.Attachments.Data) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(item.Attachments.Data))
	}
	if item.Attachments.Data[0].MediaType != "photo" {
		t.Fatalf("expected MediaType 'photo', got %s", item.Attachments.Data[0].MediaType)
	}
}

func TestFacebookInsights_Demographics(t *testing.T) {
	insights := FacebookInsights{
		PageID:          "page123",
		PageFansCity:    []string{"NYC:1000", "LA:800", "CHI:600", "HOU:400", "PHX:300"},
		PageFansCountry: []string{"US:5000", "UK:2000", "CA:1000", "AU:500"},
		PageFansAge:     []string{"13-17:500", "18-24:2000", "25-34:3500", "35-44:2500", "45-54:1000", "55-64:400", "65+:100"},
		PageFansGender:  []string{"male:5500", "female:4300", "unknown:200"},
	}

	if len(insights.PageFansCity) != 5 {
		t.Fatalf("expected 5 cities, got %d", len(insights.PageFansCity))
	}
	if len(insights.PageFansAge) != 7 {
		t.Fatalf("expected 7 age groups, got %d", len(insights.PageFansAge))
	}
}

func TestFacebookCompetitorInsights_TableName(t *testing.T) {
	insights := FacebookCompetitorInsights{}
	expected := "facebook_competitor_insights"
	if got := insights.TableName(); got != expected {
		t.Errorf("TableName() = %v, want %v", got, expected)
	}
}

func TestFacebookCompetitorPosts_TableName(t *testing.T) {
	posts := FacebookCompetitorPosts{}
	expected := "facebook_competitor_posts"
	if got := posts.TableName(); got != expected {
		t.Errorf("TableName() = %v, want %v", got, expected)
	}
}

func TestFacebookCompetitorMediaAssets_TableName(t *testing.T) {
	assets := FacebookCompetitorMediaAssets{}
	expected := "facebook_competitor_media_assets"
	if got := assets.TableName(); got != expected {
		t.Errorf("TableName() = %v, want %v", got, expected)
	}
}

func TestFacebookCompetitorInsights_Struct(t *testing.T) {
	now := time.Now()
	insights := FacebookCompetitorInsights{
		RecordID:          "record123",
		PageID:            "page456",
		FollowersCount:    10000,
		TotalFanCount:     9500,
		TalkingAboutThis:  500,
		Biography:         "Test bio",
		ProfilePictureURL: "https://example.com/pic.jpg",
		PageName:          "Test Page",
		PageCategory:      "Business",
		Emails:            []string{"test@example.com"},
		Birthday:          "01/01/2000",
		WereHereCount:     100,
		CoverPhotoURL:     "https://example.com/cover.jpg",
		Permalink:         "https://facebook.com/testpage",
		Metadata:          map[string]string{"key": "value"},
		InsertedAt:        now,
	}

	if insights.RecordID != "record123" {
		t.Fatalf("expected RecordID 'record123', got %s", insights.RecordID)
	}
	if insights.FollowersCount != 10000 {
		t.Fatalf("expected FollowersCount 10000, got %d", insights.FollowersCount)
	}
}

func TestFacebookCompetitorPosts_Struct(t *testing.T) {
	now := time.Now()
	post := FacebookCompetitorPosts{
		FacebookID:         "fb123",
		PostID:             "post456",
		FollowersCount:     10000,
		FanCount:           9500,
		PageName:           "Test Page",
		PageCategory:       "Business",
		Biography:          "Test bio",
		PostEngagement:     500,
		Like:               100,
		Haha:               50,
		Angry:              10,
		Sad:                5,
		Thankful:           2,
		Love:               80,
		TotalPostReactions: 247,
		Comments:           30,
		Shares:             20,
		Caption:            "Test caption",
		MediaType:          "photo",
		StatusType:         "added_photos",
		SharedFromName:     "Shared Page",
		SharedFromID:       "shared123",
		SharedFromPic:      "https://example.com/shared.jpg",
		SharedCreatedAt:    now,
		Permalink:          "https://facebook.com/post/456",
		Hashtags:           []string{"#test"},
		DayOfWeek:          "Monday",
		HourOfDay:          14,
		CreatedAt:          now,
		InsertedAt:         now,
		Wow:                25,
	}

	if post.FacebookID != "fb123" {
		t.Fatalf("expected FacebookID 'fb123', got %s", post.FacebookID)
	}
	if post.TotalPostReactions != 247 {
		t.Fatalf("expected TotalPostReactions 247, got %d", post.TotalPostReactions)
	}
}

func TestFacebookCompetitorMediaAssets_Struct(t *testing.T) {
	now := time.Now()
	asset := FacebookCompetitorMediaAssets{
		MediaID:      "media123",
		PostID:       "post456",
		PageID:       "page789",
		Caption:      "Test caption",
		Description:  "Test description",
		Link:         "https://example.com/media.jpg",
		AssetType:    "image",
		CallToAction: "Learn More",
		CTAType:      "LEARN_MORE",
		CreatedAt:    now,
		InsertedAt:   now,
	}

	if asset.MediaID != "media123" {
		t.Fatalf("expected MediaID 'media123', got %s", asset.MediaID)
	}
	if asset.AssetType != "image" {
		t.Fatalf("expected AssetType 'image', got %s", asset.AssetType)
	}
}
