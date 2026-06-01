package kafka

import (
	"encoding/json"
	"testing"
	"time"
)

func TestLinkedinAccountType_Constants(t *testing.T) {
	if LinkedinAccountTypePage != "Page" {
		t.Fatalf("expected LinkedinAccountTypePage 'Page', got %s", LinkedinAccountTypePage)
	}
	if LinkedinAccountTypeProfile != "Profile" {
		t.Fatalf("expected LinkedinAccountTypeProfile 'Profile', got %s", LinkedinAccountTypeProfile)
	}
}

func TestLinkedinAccountWorkOrder_Struct(t *testing.T) {
	wo := LinkedinAccountWorkOrder{
		ID:          "acc123",
		WorkspaceID: "ws456",
		LinkedinID:  "li789",
		AccessToken: "token012",
		SyncType:    "incremental",
		AccountType: LinkedinAccountTypePage,
	}

	if wo.ID != "acc123" {
		t.Fatalf("expected ID 'acc123', got %s", wo.ID)
	}
	if wo.LinkedinID != "li789" {
		t.Fatalf("expected LinkedinID 'li789', got %s", wo.LinkedinID)
	}
	if wo.AccountType != LinkedinAccountTypePage {
		t.Fatalf("expected AccountType 'Page', got %s", wo.AccountType)
	}
}

func TestLinkedinBatchWorkOrder_Struct(t *testing.T) {
	now := time.Now()
	batch := LinkedinBatchWorkOrder{
		BatchID:     "batch123",
		SyncType:    "full_sync",
		AccountType: LinkedinAccountTypeProfile,
		CreatedAt:   now,
		Accounts: []LinkedinAccountWorkOrder{
			{ID: "acc1", LinkedinID: "li1", AccountType: LinkedinAccountTypeProfile},
			{ID: "acc2", LinkedinID: "li2", AccountType: LinkedinAccountTypeProfile},
		},
	}

	if batch.BatchID != "batch123" {
		t.Fatalf("expected BatchID 'batch123', got %s", batch.BatchID)
	}
	if batch.AccountType != LinkedinAccountTypeProfile {
		t.Fatalf("expected AccountType 'Profile', got %s", batch.AccountType)
	}
	if len(batch.Accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(batch.Accounts))
	}
}

func TestParsedLinkedinPost_Struct(t *testing.T) {
	now := time.Now()
	post := ParsedLinkedinPost{
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

func TestParsedLinkedinInsights_Struct(t *testing.T) {
	now := time.Now()
	insights := ParsedLinkedinInsights{
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
		FollowersByIndustry:  `{"Technology":5000,"Finance":3000}`,
		FollowersByCountry:   `{"US":6000,"UK":2000}`,
		FollowersByCity:      `{"New York":1000,"London":800}`,
		InsertedAt:           now,
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
		PageViewsByCountry:   `{"US":15000,"UK":5000}`,
		PageViewsByRegion:    `{"California":5000,"Texas":3000}`,
		PageViewsByIndustry:  `{"Technology":10000}`,
		PageViewsBySeniority: `{"Senior":8000}`,
		PageViewsByFunction:  `{"Engineering":6000}`,
		PageViewsByStaffCount: `{"10001+":5000}`,
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

func TestParsedLinkedinMediaAsset_Struct(t *testing.T) {
	asset := ParsedLinkedinMediaAsset{
		ID:          "asset123",
		DownloadURL: "https://example.com/download/image.jpg",
		Thumbnail:   "https://example.com/thumb.jpg",
		Type:        "image",
	}

	if asset.ID != "asset123" {
		t.Fatalf("expected ID 'asset123', got %s", asset.ID)
	}
	if asset.Type != "image" {
		t.Fatalf("expected Type 'image', got %s", asset.Type)
	}
}

func TestParsedLinkedinMediaAsset_VideoType(t *testing.T) {
	asset := ParsedLinkedinMediaAsset{
		ID:          "video123",
		DownloadURL: "https://example.com/download/video.mp4",
		Thumbnail:   "https://example.com/video_thumb.jpg",
		Type:        "video",
	}

	if asset.Type != "video" {
		t.Fatalf("expected Type 'video', got %s", asset.Type)
	}
}

func TestParsedLinkedinStat_Struct(t *testing.T) {
	stat := ParsedLinkedinStat{
		ActivityID:             "urn:li:activity:123456789",
		CommentCount:           50,
		LikeCount:              200,
		UniqueImpressionsCount: 5000,
		ShareCount:             25,
		ClickCount:             150,
		ImpressionCount:        10000,
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

func TestLinkedinAccountWorkOrder_JSON_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "acc123",
		"workspace_id": "ws456",
		"linkedin_id": "li789",
		"access_token": "token012",
		"sync_type": "incremental",
		"account_type": "Page"
	}`

	var wo LinkedinAccountWorkOrder
	err := json.Unmarshal([]byte(jsonData), &wo)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if wo.ID != "acc123" {
		t.Fatalf("expected ID 'acc123', got %s", wo.ID)
	}
	if wo.AccountType != LinkedinAccountTypePage {
		t.Fatalf("expected AccountType 'Page', got %s", wo.AccountType)
	}
}

func TestLinkedinBatchWorkOrder_JSON_Unmarshal(t *testing.T) {
	jsonData := `{
		"batch_id": "batch123",
		"sync_type": "full_sync",
		"account_type": "Profile",
		"created_at": "2025-01-15T12:00:00Z",
		"accounts": [
			{"id": "acc1", "linkedin_id": "li1", "access_token": "token1", "sync_type": "full_sync", "account_type": "Profile"},
			{"id": "acc2", "linkedin_id": "li2", "access_token": "token2", "sync_type": "full_sync", "account_type": "Profile"}
		]
	}`

	var batch LinkedinBatchWorkOrder
	err := json.Unmarshal([]byte(jsonData), &batch)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if batch.BatchID != "batch123" {
		t.Fatalf("expected BatchID 'batch123', got %s", batch.BatchID)
	}
	if len(batch.Accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(batch.Accounts))
	}
}

func TestParsedLinkedinPost_JSON_Marshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	post := ParsedLinkedinPost{
		LinkedinID:      "li123",
		PostID:          "post456",
		Comments:        50,
		TotalEngagement: 250.5,
		Hashtags:        []string{"linkedin", "business"},
		CreatedAt:       now,
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result ParsedLinkedinPost
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.LinkedinID != post.LinkedinID {
		t.Fatalf("expected LinkedinID %s, got %s", post.LinkedinID, result.LinkedinID)
	}
	if result.TotalEngagement != post.TotalEngagement {
		t.Fatalf("expected TotalEngagement %f, got %f", post.TotalEngagement, result.TotalEngagement)
	}
}

func TestParsedLinkedinInsights_JSON_Marshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	insights := ParsedLinkedinInsights{
		LinkedinID:         "li123",
		OrganizationName:   "Test Company",
		TotalFollowerCount: 10000,
		PageViews:          25000,
		CreatedAt:          now,
	}

	data, err := json.Marshal(insights)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result ParsedLinkedinInsights
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.TotalFollowerCount != insights.TotalFollowerCount {
		t.Fatalf("expected TotalFollowerCount %d, got %d", insights.TotalFollowerCount, result.TotalFollowerCount)
	}
}

func TestLinkedinAccountType_Values(t *testing.T) {
	cases := []struct {
		accountType LinkedinAccountType
		expected    string
	}{
		{LinkedinAccountTypePage, "Page"},
		{LinkedinAccountTypeProfile, "Profile"},
	}

	for _, tc := range cases {
		if string(tc.accountType) != tc.expected {
			t.Fatalf("expected %s, got %s", tc.expected, string(tc.accountType))
		}
	}
}
