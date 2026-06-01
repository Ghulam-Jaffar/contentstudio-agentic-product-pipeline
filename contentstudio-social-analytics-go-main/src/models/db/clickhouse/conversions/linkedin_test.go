package conversions

import (
	"context"
	"testing"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestBulkInsertLinkedInPosts_EmptySlice(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	err := sink.BulkInsertLinkedInPosts(context.Background(), []*clickhousemodels.LinkedInPosts{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestBulkInsertLinkedInPosts_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertLinkedInPostsFunc: func(ctx context.Context, posts []*clickhousemodels.LinkedInPosts) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	posts := []*clickhousemodels.LinkedInPosts{{LinkedinID: "li123"}}
	err := sink.BulkInsertLinkedInPosts(context.Background(), posts)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertLinkedInPosts to be called")
	}
}

func TestBulkInsertLinkedInInsights_EmptySlice(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	err := sink.BulkInsertLinkedInInsights(context.Background(), []*clickhousemodels.LinkedInInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestBulkInsertLinkedInInsights_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertLinkedInInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.LinkedInInsights) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	insights := []*clickhousemodels.LinkedInInsights{{LinkedinID: "li123"}}
	err := sink.BulkInsertLinkedInInsights(context.Background(), insights)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertLinkedInInsights to be called")
	}
}

func TestBulkInsertLinkedInMediaAssets_EmptySlice(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	err := sink.BulkInsertLinkedInMediaAssets(context.Background(), []*clickhousemodels.LinkedInMediaAsset{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestBulkInsertLinkedInStats_EmptySlice(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	err := sink.BulkInsertLinkedInStats(context.Background(), []*clickhousemodels.LinkedInStat{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestConvertLinkedInPost_NilInput(t *testing.T) {
	result := ConvertLinkedInPost(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertLinkedInPost_ValidInput(t *testing.T) {
	now := time.Now()

	input := &kafkamodels.ParsedLinkedinPost{
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
		IsReshareDisabled:  false,
		FeedDistribution:   "MAIN_FEED",
		ThirdPartyChannels: []string{},
	}

	result := ConvertLinkedInPost(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.LinkedinID != "li123" {
		t.Fatalf("expected LinkedinID 'li123', got %s", result.LinkedinID)
	}
	if result.PostID != "post456" {
		t.Fatalf("expected PostID 'post456', got %s", result.PostID)
	}
	if result.TotalEngagement != 250.5 {
		t.Fatalf("expected TotalEngagement 250.5, got %f", result.TotalEngagement)
	}
	if result.HourOfDay != 10 {
		t.Fatalf("expected HourOfDay 10, got %d", result.HourOfDay)
	}
	if len(result.Hashtags) != 2 {
		t.Fatalf("expected 2 hashtags, got %d", len(result.Hashtags))
	}
	if result.SavingTime.IsZero() {
		t.Fatal("expected SavingTime to be set")
	}
}

func TestConvertLinkedInPost_MediaTypes(t *testing.T) {
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
		input := &kafkamodels.ParsedLinkedinPost{
			LinkedinID: "li123",
			MediaType:  tc.mediaType,
			Type:       tc.postType,
		}

		result := ConvertLinkedInPost(input)

		if result.MediaType != tc.mediaType {
			t.Fatalf("expected MediaType %s, got %s", tc.mediaType, result.MediaType)
		}
		if result.Type != tc.postType {
			t.Fatalf("expected Type %s, got %s", tc.postType, result.Type)
		}
	}
}

func TestConvertLinkedInMediaAsset_NilInput(t *testing.T) {
	result := ConvertLinkedInMediaAsset(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertLinkedInMediaAsset_ValidInput(t *testing.T) {
	input := &kafkamodels.ParsedLinkedinMediaAsset{
		ID:          "asset123",
		DownloadURL: "https://example.com/download/image.jpg",
		Thumbnail:   "https://example.com/thumb.jpg",
		Type:        "image",
	}

	result := ConvertLinkedInMediaAsset(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ID != "asset123" {
		t.Fatalf("expected ID 'asset123', got %s", result.ID)
	}
	if result.DownloadURL != "https://example.com/download/image.jpg" {
		t.Fatalf("expected DownloadURL, got %s", result.DownloadURL)
	}
	if result.Type != "image" {
		t.Fatalf("expected Type 'image', got %s", result.Type)
	}
	if result.SavingTime.IsZero() {
		t.Fatal("expected SavingTime to be set")
	}
}

func TestConvertLinkedInStat_NilInput(t *testing.T) {
	result := ConvertLinkedInStat(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertLinkedInStat_ValidInput(t *testing.T) {
	input := &kafkamodels.ParsedLinkedinStat{
		ActivityID:             "urn:li:activity:123456789",
		CommentCount:           50,
		LikeCount:              200,
		UniqueImpressionsCount: 5000,
		ShareCount:             25,
		ClickCount:             150,
		ImpressionCount:        10000,
	}

	result := ConvertLinkedInStat(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ActivityID != "urn:li:activity:123456789" {
		t.Fatalf("expected ActivityID, got %s", result.ActivityID)
	}
	if result.LikeCount != 200 {
		t.Fatalf("expected LikeCount 200, got %d", result.LikeCount)
	}
	if result.ImpressionCount != 10000 {
		t.Fatalf("expected ImpressionCount 10000, got %d", result.ImpressionCount)
	}
	if result.SavingTime.IsZero() {
		t.Fatal("expected SavingTime to be set")
	}
}

func TestConvertLinkedInInsights_NilInput(t *testing.T) {
	result := ConvertLinkedInInsights(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertLinkedInInsights_ValidInput(t *testing.T) {
	now := time.Now()

	input := &kafkamodels.ParsedLinkedinInsights{
		LinkedinID:           "li123",
		OrganizationName:     "Test Company",
		RecordID:             "rec456",
		ImpressionCount:      50000,
		OrganicFollowerCount: 8000,
		TotalFollowerCount:   10000,
		PaidFollowerCount:    2000,
		DailyFollowerCount:   50,
		Reach:                40000,
		Repost:               100,
		Comments:             500,
		PostClicks:           2000,
		Reactions:            3000,
		Engagement:           150.5,
		FollowersBySeniority: `{"Entry":1000,"Senior":2000}`,
		FollowersByIndustry:  `{"Technology":5000}`,
		FollowersByCountry:   `{"US":6000,"UK":2000}`,
		FollowersByCity:      `{"NYC":1000}`,
		CreatedAt:            now,
		PageViews:            25000,
		UniqueVisitors:       15000,
		DesktopPageViews:     20000,
		MobilePageViews:      5000,
		OverviewPageViews:    10000,
		AboutPageViews:       5000,
		JobsPageViews:        3000,
		PeoplePageViews:      2000,
		CareersPageViews:     1500,
		LifeAtPageViews:      1000,
		InsightsPageViews:    500,
		ProductsPageViews:    1000,
		PageViewsByCountry:   `{"US":15000}`,
		PageViewsByRegion:    `{"CA":5000}`,
		PageViewsByIndustry:  `{"Technology":10000}`,
		PageViewsBySeniority: `{"Senior":8000}`,
		PageViewsByFunction:  `{"Engineering":6000}`,
		PageViewsByStaffCount: `{"10001+":5000}`,
	}

	result := ConvertLinkedInInsights(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.LinkedinID != "li123" {
		t.Fatalf("expected LinkedinID 'li123', got %s", result.LinkedinID)
	}
	if result.OrganizationName != "Test Company" {
		t.Fatalf("expected OrganizationName 'Test Company', got %s", result.OrganizationName)
	}
	if result.TotalFollowerCount != 10000 {
		t.Fatalf("expected TotalFollowerCount 10000, got %d", result.TotalFollowerCount)
	}
	if result.PageViews != 25000 {
		t.Fatalf("expected PageViews 25000, got %d", result.PageViews)
	}
	if result.InsertedAt.IsZero() {
		t.Fatal("expected InsertedAt to be set")
	}
}

func TestConvertLinkedInInsights_PageViewBreakdowns(t *testing.T) {
	input := &kafkamodels.ParsedLinkedinInsights{
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

	result := ConvertLinkedInInsights(input)

	totalSectionViews := result.OverviewPageViews + result.AboutPageViews +
		result.JobsPageViews + result.PeoplePageViews + result.CareersPageViews +
		result.LifeAtPageViews + result.InsightsPageViews + result.ProductsPageViews

	if totalSectionViews != 24000 {
		t.Fatalf("expected total section views 24000, got %d", totalSectionViews)
	}
}
