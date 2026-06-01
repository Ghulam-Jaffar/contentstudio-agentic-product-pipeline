package clickhouse

import (
	"encoding/json"
	"testing"
	"time"
)

func TestLinkedInPosts_Struct(t *testing.T) {
	now := time.Now()
	post := LinkedInPosts{
		LinkedinID:         "li123",
		PostID:             "post456",
		Activity:           "urn:li:activity:123456789",
		Comments:           50,
		TotalEngagement:    250.5,
		Favorites:          100,
		PollData:           "",
		Reach:              5000,
		Repost:             25,
		PostClicks:         150,
		Impressions:        10000,
		Title:              "Post Title",
		Image:              "https://example.com/image.jpg",
		ArticleURL:         "https://example.com/article",
		ArticleTitle:       "Article Title",
		Media:              []string{"img1.jpg", "img2.jpg"},
		MediaType:          "IMAGE",
		Type:               "SHARE",
		Hashtags:           []string{"linkedin", "business"},
		DayOfWeek:          "Tuesday",
		HourOfDay:          10,
		CreatedAt:          now,
		PublishedAt:        now,
		LastModifiedAt:     now,
		LifecycleState:     "PUBLISHED",
		Visibility:         "PUBLIC",
		SavingTime:         now,
		IsReshareDisabled:  false,
		FeedDistribution:   "MAIN_FEED",
		ThirdPartyChannels: []string{},
	}

	if post.LinkedinID != "li123" {
		t.Fatalf("expected LinkedinID 'li123', got %s", post.LinkedinID)
	}
	if post.TotalEngagement != 250.5 {
		t.Fatalf("expected TotalEngagement 250.5, got %f", post.TotalEngagement)
	}
	if len(post.Hashtags) != 2 {
		t.Fatalf("expected 2 hashtags, got %d", len(post.Hashtags))
	}
}

func TestLinkedInPosts_JSON_Marshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	post := LinkedInPosts{
		LinkedinID:      "li123",
		PostID:          "post456",
		TotalEngagement: 250.5,
		CreatedAt:       now,
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result LinkedInPosts
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.TotalEngagement != post.TotalEngagement {
		t.Fatalf("expected TotalEngagement %f, got %f", post.TotalEngagement, result.TotalEngagement)
	}
}

func TestLinkedInPosts_MediaTypes(t *testing.T) {
	cases := []struct {
		mediaType string
		postType  string
	}{
		{"IMAGE", "SHARE"},
		{"VIDEO", "SHARE"},
		{"ARTICLE", "SHARE"},
		{"DOCUMENT", "SHARE"},
		{"", "TEXT"},
	}

	for _, tc := range cases {
		post := LinkedInPosts{
			LinkedinID: "li123",
			MediaType:  tc.mediaType,
			Type:       tc.postType,
		}

		if post.MediaType != tc.mediaType {
			t.Fatalf("expected MediaType %s, got %s", tc.mediaType, post.MediaType)
		}
	}
}

func TestLinkedInMinimalPost_Struct(t *testing.T) {
	post := LinkedInMinimalPost{
		LinkedinID: "li123",
		PostID:     "post456",
		Activity:   "urn:li:ugcPost:post456",
		Image:      "https://example.com/image.jpg",
		Media:      []string{"https://example.com/image.jpg", "https://example.com/doc.pdf"},
	}

	if post.LinkedinID != "li123" {
		t.Fatalf("expected LinkedinID li123, got %s", post.LinkedinID)
	}
	if post.PostID != "post456" {
		t.Fatalf("expected PostID post456, got %s", post.PostID)
	}
	if len(post.Media) != 2 {
		t.Fatalf("expected 2 media items, got %d", len(post.Media))
	}
}

func TestLinkedInPosts_LifecycleStates(t *testing.T) {
	states := []string{"PUBLISHED", "DRAFT", "PROCESSING", "DELETED"}

	for _, state := range states {
		post := LinkedInPosts{
			LinkedinID:     "li123",
			LifecycleState: state,
		}

		if post.LifecycleState != state {
			t.Fatalf("expected LifecycleState %s, got %s", state, post.LifecycleState)
		}
	}
}

func TestLinkedInInsights_Struct(t *testing.T) {
	now := time.Now()
	insights := LinkedInInsights{
		LinkedinID:            "li123",
		OrganizationName:      "Test Company",
		RecordID:              "rec456",
		ImpressionCount:       50000,
		OrganicFollowerCount:  8000,
		TotalFollowerCount:    10000,
		PaidFollowerCount:     2000,
		DailyFollowerCount:    50,
		Reach:                 40000,
		Repost:                100,
		Comments:              500,
		PostClicks:            2000,
		Reactions:             3000,
		Engagement:            150.5,
		FollowersBySeniority:  `{"Entry":1000,"Senior":2000,"Manager":3000}`,
		FollowersByIndustry:   `{"Technology":5000,"Finance":3000,"Healthcare":2000}`,
		FollowersByCountry:    `{"US":6000,"UK":2000,"CA":1000,"DE":500}`,
		FollowersByCity:       `{"New York":1000,"San Francisco":800,"London":600}`,
		InsertedAt:            now,
		CreatedAt:             now,
		PageViews:             25000,
		UniqueVisitors:        15000,
		DesktopPageViews:      20000,
		MobilePageViews:       5000,
		OverviewPageViews:     10000,
		AboutPageViews:        5000,
		JobsPageViews:         3000,
		PeoplePageViews:       2000,
		CareersPageViews:      1500,
		LifeAtPageViews:       1000,
		InsightsPageViews:     500,
		ProductsPageViews:     1000,
		PageViewsByCountry:    `{"US":15000,"UK":5000,"CA":3000}`,
		PageViewsByRegion:     `{"California":5000,"New York":4000,"Texas":3000}`,
		PageViewsByIndustry:   `{"Technology":10000,"Finance":5000}`,
		PageViewsBySeniority:  `{"Senior":8000,"Entry":5000}`,
		PageViewsByFunction:   `{"Engineering":6000,"Sales":4000}`,
		PageViewsByStaffCount: `{"10001+":5000,"1001-5000":3000}`,
	}

	if insights.LinkedinID != "li123" {
		t.Fatalf("expected LinkedinID 'li123', got %s", insights.LinkedinID)
	}
	if insights.TotalFollowerCount != 10000 {
		t.Fatalf("expected TotalFollowerCount 10000, got %d", insights.TotalFollowerCount)
	}
	if insights.PageViews != 25000 {
		t.Fatalf("expected PageViews 25000, got %d", insights.PageViews)
	}
}

func TestLinkedInInsights_JSON_Marshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	insights := LinkedInInsights{
		LinkedinID:         "li123",
		TotalFollowerCount: 10000,
		PageViews:          25000,
		CreatedAt:          now,
	}

	data, err := json.Marshal(insights)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result LinkedInInsights
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.TotalFollowerCount != insights.TotalFollowerCount {
		t.Fatalf("expected TotalFollowerCount %d, got %d", insights.TotalFollowerCount, result.TotalFollowerCount)
	}
}

func TestLinkedInInsights_PageViewBreakdowns(t *testing.T) {
	insights := LinkedInInsights{
		LinkedinID:        "li123",
		OverviewPageViews: 10000,
		AboutPageViews:    5000,
		JobsPageViews:     3000,
		PeoplePageViews:   2000,
		CareersPageViews:  1500,
		LifeAtPageViews:   1000,
		InsightsPageViews: 500,
		ProductsPageViews: 1000,
	}

	totalSectionViews := insights.OverviewPageViews + insights.AboutPageViews +
		insights.JobsPageViews + insights.PeoplePageViews + insights.CareersPageViews +
		insights.LifeAtPageViews + insights.InsightsPageViews + insights.ProductsPageViews

	if totalSectionViews != 24000 {
		t.Fatalf("expected total section views 24000, got %d", totalSectionViews)
	}
}

func TestLinkedInMediaAsset_Struct(t *testing.T) {
	now := time.Now()
	asset := LinkedInMediaAsset{
		ID:          "asset123",
		DownloadURL: "https://example.com/download/image.jpg",
		Thumbnail:   "https://example.com/thumb.jpg",
		Type:        "image",
		SavingTime:  now,
	}

	if asset.ID != "asset123" {
		t.Fatalf("expected ID 'asset123', got %s", asset.ID)
	}
	if asset.Type != "image" {
		t.Fatalf("expected Type 'image', got %s", asset.Type)
	}
}

func TestLinkedInMediaAsset_Types(t *testing.T) {
	types := []string{"image", "video", "document", "article"}

	for _, assetType := range types {
		asset := LinkedInMediaAsset{
			ID:   "asset123",
			Type: assetType,
		}

		if asset.Type != assetType {
			t.Fatalf("expected Type %s, got %s", assetType, asset.Type)
		}
	}
}

func TestLinkedInStat_Struct(t *testing.T) {
	now := time.Now()
	stat := LinkedInStat{
		ActivityID:             "urn:li:activity:123456789",
		CommentCount:           50,
		LikeCount:              200,
		UniqueImpressionsCount: 5000,
		ShareCount:             25,
		ClickCount:             150,
		ImpressionCount:        10000,
		SavingTime:             now,
	}

	if stat.ActivityID != "urn:li:activity:123456789" {
		t.Fatalf("expected ActivityID, got %s", stat.ActivityID)
	}
	if stat.LikeCount != 200 {
		t.Fatalf("expected LikeCount 200, got %d", stat.LikeCount)
	}
	if stat.ImpressionCount != 10000 {
		t.Fatalf("expected ImpressionCount 10000, got %d", stat.ImpressionCount)
	}
}

func TestLinkedInStat_JSON_Marshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	stat := LinkedInStat{
		ActivityID:      "urn:li:activity:123",
		LikeCount:       200,
		ImpressionCount: 10000,
		SavingTime:      now,
	}

	data, err := json.Marshal(stat)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result LinkedInStat
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.LikeCount != stat.LikeCount {
		t.Fatalf("expected LikeCount %d, got %d", stat.LikeCount, result.LikeCount)
	}
}

func TestLinkedInStat_EngagementCalculation(t *testing.T) {
	stat := LinkedInStat{
		ActivityID:      "urn:li:activity:123",
		CommentCount:    50,
		LikeCount:       200,
		ShareCount:      25,
		ClickCount:      150,
		ImpressionCount: 10000,
	}

	totalEngagement := stat.CommentCount + stat.LikeCount + stat.ShareCount + stat.ClickCount
	if totalEngagement != 425 {
		t.Fatalf("expected total engagement 425, got %d", totalEngagement)
	}

	engagementRate := float64(totalEngagement) / float64(stat.ImpressionCount) * 100
	expectedRate := 4.25
	if engagementRate != expectedRate {
		t.Fatalf("expected engagement rate %f, got %f", expectedRate, engagementRate)
	}
}
