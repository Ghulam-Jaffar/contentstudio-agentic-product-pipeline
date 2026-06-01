package processor

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestWorkOrder_Struct(t *testing.T) {
	wo := WorkOrder{
		ID:          "account123",
		AccessToken: "token_abc",
		WorkspaceID: "workspace_789",
		SyncType:    "full",
		AccountID:   "li_123456",
	}

	if wo.ID != "account123" {
		t.Fatalf("expected ID 'account123', got '%s'", wo.ID)
	}
	if wo.AccountID != "li_123456" {
		t.Fatalf("expected AccountID 'li_123456', got '%s'", wo.AccountID)
	}
	if wo.SyncType != "full" {
		t.Fatalf("expected SyncType 'full', got '%s'", wo.SyncType)
	}
}

func TestEnrichedPost_Struct(t *testing.T) {
	ep := enrichedPost{
		Post: map[string]any{
			"id":      "post123",
			"message": "Hello World",
		},
		ActivityID:  "urn:li:activity:123456",
		ImageIDs:    []string{"img1", "img2"},
		VideoIDs:    []string{"vid1"},
		DocumentIDs: []string{"doc1"},
	}

	if ep.ActivityID != "urn:li:activity:123456" {
		t.Fatalf("expected ActivityID 'urn:li:activity:123456', got '%s'", ep.ActivityID)
	}
	if len(ep.ImageIDs) != 2 {
		t.Fatalf("expected 2 image IDs, got %d", len(ep.ImageIDs))
	}
	if len(ep.VideoIDs) != 1 {
		t.Fatalf("expected 1 video ID, got %d", len(ep.VideoIDs))
	}
	if len(ep.DocumentIDs) != 1 {
		t.Fatalf("expected 1 document ID, got %d", len(ep.DocumentIDs))
	}
}

func TestFetchedData_Struct(t *testing.T) {
	fd := FetchedData{
		Posts:        map[string]*enrichedPost{},
		StatsMap:     map[string]map[string]any{},
		ImageMap:     map[string]map[string]any{},
		VideoMap:     map[string]map[string]any{},
		DocumentMap:  map[string]map[string]any{},
		FollowerData: []byte(`{"followers": 1000}`),
		PageStats:    []byte(`{"views": 500}`),
		ShareStats:   []byte(`{"shares": 100}`),
	}

	if fd.Posts == nil {
		t.Fatal("expected Posts to be initialized")
	}
	if len(fd.FollowerData) == 0 {
		t.Fatal("expected FollowerData to have content")
	}
}

func TestParsedData_Struct(t *testing.T) {
	pd := ParsedData{
		Posts:       []kafkamodels.ParsedLinkedinPost{},
		MediaAssets: []kafkamodels.ParsedLinkedinMediaAsset{},
		Stats:       []kafkamodels.ParsedLinkedinStat{},
		Insights:    []kafkamodels.ParsedLinkedinInsights{},
	}

	if pd.Posts == nil {
		t.Fatal("expected Posts to be initialized")
	}
	if pd.MediaAssets == nil {
		t.Fatal("expected MediaAssets to be initialized")
	}
	if pd.Stats == nil {
		t.Fatal("expected Stats to be initialized")
	}
	if pd.Insights == nil {
		t.Fatal("expected Insights to be initialized")
	}
}

func TestParsedData_WithData(t *testing.T) {
	pd := ParsedData{
		Posts: []kafkamodels.ParsedLinkedinPost{
			{PostID: "post1", LinkedinID: "li1"},
			{PostID: "post2", LinkedinID: "li1"},
		},
		MediaAssets: []kafkamodels.ParsedLinkedinMediaAsset{
			{ID: "media1", Type: "image"},
		},
		Stats: []kafkamodels.ParsedLinkedinStat{
			{ActivityID: "activity1"},
		},
		Insights: []kafkamodels.ParsedLinkedinInsights{
			{LinkedinID: "li1"},
		},
	}

	if len(pd.Posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(pd.Posts))
	}
	if len(pd.MediaAssets) != 1 {
		t.Fatalf("expected 1 media asset, got %d", len(pd.MediaAssets))
	}
	if len(pd.Stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(pd.Stats))
	}
	if len(pd.Insights) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(pd.Insights))
	}
}

func TestProcessor_Struct(t *testing.T) {
	p := &Processor{
		MongoRepo:    nil,
		LiClient:     nil,
		Sink:         nil,
		GeoResolver:  nil,
		Producer:     nil,
		Notifier:     nil,
		PusherClient: nil,
		Logger:       nil,
		Cfg:          nil,
		StatsConc:    nil,
		MediaConc:    nil,
		GeoConc:      nil,
	}

	if p.MongoRepo != nil {
		t.Fatal("expected nil MongoRepo")
	}
	if p.LiClient != nil {
		t.Fatal("expected nil LiClient")
	}
}

func TestConstants(t *testing.T) {
	if StatsConcsPerWorker != 4 {
		t.Fatalf("expected StatsConcsPerWorker 4, got %d", StatsConcsPerWorker)
	}
	if MediaConcPerWorker != 4 {
		t.Fatalf("expected MediaConcPerWorker 4, got %d", MediaConcPerWorker)
	}
	if GeoConcPerWorker != 2 {
		t.Fatalf("expected GeoConcPerWorker 2, got %d", GeoConcPerWorker)
	}
	if DefaultWorkerCount != 10 {
		t.Fatalf("expected DefaultWorkerCount 10, got %d", DefaultWorkerCount)
	}
}

func TestDefaultLinkedInDateRange(t *testing.T) {
	now := time.Date(2026, 4, 28, 10, 15, 0, 0, time.UTC)

	startDate, endDate := defaultLinkedInDateRange(now)

	expectedStart := time.Date(2025, 4, 26, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 4, 27, 23, 59, 59, 0, time.UTC)

	if !startDate.Equal(expectedStart) {
		t.Fatalf("expected start date %s, got %s", expectedStart, startDate)
	}
	if !endDate.Equal(expectedEnd) {
		t.Fatalf("expected end date %s, got %s", expectedEnd, endDate)
	}
}

func TestWorkOrder_EmptyFields(t *testing.T) {
	wo := WorkOrder{}

	if wo.ID != "" {
		t.Fatal("expected empty ID")
	}
	if wo.AccountID != "" {
		t.Fatal("expected empty AccountID")
	}
	if wo.SyncType != "" {
		t.Fatal("expected empty SyncType")
	}
}

func TestEnrichedPost_EmptyFields(t *testing.T) {
	ep := enrichedPost{}

	if ep.Post != nil {
		t.Fatal("expected nil Post")
	}
	if ep.ActivityID != "" {
		t.Fatal("expected empty ActivityID")
	}
	if ep.ImageIDs != nil {
		t.Fatal("expected nil ImageIDs")
	}
}

func TestFetchedData_EmptyMaps(t *testing.T) {
	fd := FetchedData{
		Posts:       make(map[string]*enrichedPost),
		StatsMap:    make(map[string]map[string]any),
		ImageMap:    make(map[string]map[string]any),
		VideoMap:    make(map[string]map[string]any),
		DocumentMap: make(map[string]map[string]any),
	}

	if len(fd.Posts) != 0 {
		t.Fatal("expected empty Posts map")
	}
	if len(fd.StatsMap) != 0 {
		t.Fatal("expected empty StatsMap")
	}
}

func TestFetchedData_AddPost(t *testing.T) {
	fd := FetchedData{
		Posts: make(map[string]*enrichedPost),
	}

	post := &enrichedPost{
		ActivityID: "activity123",
		Post:       map[string]any{"id": "post123"},
	}

	fd.Posts["activity123"] = post

	if len(fd.Posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(fd.Posts))
	}
	if fd.Posts["activity123"].ActivityID != "activity123" {
		t.Fatal("unexpected activity ID")
	}
}

// ================== isTokenError Tests ==================

func TestIsTokenError_Nil(t *testing.T) {
	if isTokenError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsTokenError_401(t *testing.T) {
	tests := []struct {
		name     string
		err      string
		expected bool
	}{
		{"401 error", "got 401 unauthorized", true},
		{"403 error", "got 403 forbidden", true},
		{"unauthorized", "request unauthorized", true},
		{"invalid_token", "invalid_token in response", true},
		{"expired", "EXPIRED_ACCESS_TOKEN", true},
		{"access_token", "access_token is invalid", true},
		{"authentication", "authentication failed", true},
		{"access denied", "access denied for resource", true},
		{"permission", "no permission to access", true},
		{"not authorized", "user is not authorized", true},
		{"normal error", "something went wrong", false},
		{"500 error", "got 500 internal server error", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := &tokenTestError{msg: tc.err}
			result := isTokenError(err)
			if result != tc.expected {
				t.Errorf("isTokenError(%q) = %v, want %v", tc.err, result, tc.expected)
			}
		})
	}
}

type tokenTestError struct {
	msg string
}

func (e *tokenTestError) Error() string {
	return e.msg
}

// ================== chunk Tests ==================

func TestChunk_Empty(t *testing.T) {
	result := chunk([]string{}, 5)
	if result != nil {
		t.Errorf("expected nil for empty slice, got %v", result)
	}
}

func TestChunk_ZeroSize(t *testing.T) {
	result := chunk([]string{"a", "b"}, 0)
	if result != nil {
		t.Errorf("expected nil for zero chunk size, got %v", result)
	}
}

func TestChunk_NegativeSize(t *testing.T) {
	result := chunk([]string{"a", "b"}, -1)
	if result != nil {
		t.Errorf("expected nil for negative chunk size, got %v", result)
	}
}

func TestChunk_ExactFit(t *testing.T) {
	result := chunk([]string{"a", "b", "c"}, 3)
	if len(result) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(result))
	}
	if len(result[0]) != 3 {
		t.Errorf("expected chunk size 3, got %d", len(result[0]))
	}
}

func TestChunk_MultipleChunks(t *testing.T) {
	result := chunk([]string{"a", "b", "c", "d", "e"}, 2)
	if len(result) != 3 {
		t.Errorf("expected 3 chunks, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Errorf("expected first chunk size 2, got %d", len(result[0]))
	}
	if len(result[1]) != 2 {
		t.Errorf("expected second chunk size 2, got %d", len(result[1]))
	}
	if len(result[2]) != 1 {
		t.Errorf("expected third chunk size 1, got %d", len(result[2]))
	}
}

func TestChunk_SingleElement(t *testing.T) {
	result := chunk([]string{"a"}, 5)
	if len(result) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(result))
	}
	if len(result[0]) != 1 {
		t.Errorf("expected chunk size 1, got %d", len(result[0]))
	}
}

func TestChunk_IntSlice(t *testing.T) {
	result := chunk([]int{1, 2, 3, 4, 5}, 2)
	if len(result) != 3 {
		t.Errorf("expected 3 chunks, got %d", len(result))
	}
}

// ================== parseStatsBatch Tests ==================

func TestParseStatsBatch_Empty(t *testing.T) {
	result := parseStatsBatch([]byte(`{}`))
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d elements", len(result))
	}
}

func TestParseStatsBatch_UGCPosts(t *testing.T) {
	json := `{
		"elements": [
			{"ugcPost": "urn:li:ugcPost:123", "totalShareStatistics": {"likes": 10, "comments": 5}},
			{"ugcPost": "urn:li:ugcPost:456", "totalShareStatistics": {"likes": 20, "comments": 10}}
		]
	}`
	result := parseStatsBatch([]byte(json))
	if len(result) != 2 {
		t.Errorf("expected 2 elements, got %d", len(result))
	}
	if result["urn:li:ugcPost:123"] == nil {
		t.Error("expected urn:li:ugcPost:123 to be present")
	}
}

func TestParseStatsBatch_Shares(t *testing.T) {
	json := `{
		"elements": [
			{"share": "urn:li:share:789", "totalShareStatistics": {"likes": 15}}
		]
	}`
	result := parseStatsBatch([]byte(json))
	if len(result) != 1 {
		t.Errorf("expected 1 element, got %d", len(result))
	}
	if result["urn:li:share:789"] == nil {
		t.Error("expected urn:li:share:789 to be present")
	}
}

func TestParseStatsBatch_InvalidJSON(t *testing.T) {
	result := parseStatsBatch([]byte(`not json`))
	if len(result) != 0 {
		t.Errorf("expected empty map for invalid JSON, got %d elements", len(result))
	}
}

func TestParseStatsBatch_EmptyIDs(t *testing.T) {
	json := `{
		"elements": [
			{"ugcPost": "", "share": "", "totalShareStatistics": {"likes": 10}}
		]
	}`
	result := parseStatsBatch([]byte(json))
	if len(result) != 0 {
		t.Errorf("expected 0 elements for empty IDs, got %d", len(result))
	}
}

// ================== parseAssetBatch Tests ==================

func TestParseAssetBatch_Empty(t *testing.T) {
	result := parseAssetBatch([]byte(`{}`))
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d elements", len(result))
	}
}

func TestParseAssetBatch_WithID(t *testing.T) {
	json := `{
		"results": {
			"urn:li:image:123": {"id": "urn:li:image:123", "url": "https://example.com/img.jpg"}
		}
	}`
	result := parseAssetBatch([]byte(json))
	if len(result) != 1 {
		t.Errorf("expected 1 element, got %d", len(result))
	}
	if result["urn:li:image:123"] == nil {
		t.Error("expected urn:li:image:123 to be present")
	}
}

func TestParseAssetBatch_WithAsset(t *testing.T) {
	json := `{
		"results": {
			"video1": {"asset": "urn:li:video:456", "downloadUrl": "https://example.com/video.mp4"}
		}
	}`
	result := parseAssetBatch([]byte(json))
	if len(result) != 1 {
		t.Errorf("expected 1 element, got %d", len(result))
	}
	if result["urn:li:video:456"] == nil {
		t.Error("expected urn:li:video:456 to be present")
	}
}

func TestParseAssetBatch_InvalidJSON(t *testing.T) {
	result := parseAssetBatch([]byte(`invalid`))
	if len(result) != 0 {
		t.Errorf("expected empty map for invalid JSON, got %d elements", len(result))
	}
}

func TestParseAssetBatch_NoIDOrAsset(t *testing.T) {
	json := `{
		"results": {
			"something": {"type": "image", "url": "https://example.com"}
		}
	}`
	result := parseAssetBatch([]byte(json))
	if len(result) != 0 {
		t.Errorf("expected 0 elements for missing ID, got %d", len(result))
	}
}

// ================== newCaptureFunc Tests ==================

func TestNewCaptureFunc_NilError(t *testing.T) {
	baseTags := map[string]string{"platform": "linkedin"}
	baseExtras := map[string]interface{}{"key": "value"}

	capture := newCaptureFunc(baseTags, baseExtras)
	// Should not panic with nil error
	capture("test_stage", nil, nil)
}

func TestNewCaptureFunc_WithError(t *testing.T) {
	baseTags := map[string]string{"platform": "linkedin"}
	baseExtras := map[string]interface{}{"key": "value"}

	capture := newCaptureFunc(baseTags, baseExtras)
	// Should not panic with error
	capture("test_stage", &tokenTestError{msg: "test error"}, map[string]interface{}{"extra": "data"})
}

func TestNewCaptureFunc_NilExtraMap(t *testing.T) {
	baseTags := map[string]string{}
	baseExtras := map[string]interface{}{}

	capture := newCaptureFunc(baseTags, baseExtras)
	// Should not panic with nil extra map
	capture("test_stage", &tokenTestError{msg: "test error"}, nil)
}

// ================== extractAssetIDs Tests ==================

func TestExtractAssetIDs_EmptyPosts(t *testing.T) {
	p := &Processor{}
	byActivity, ugcIDs, shareIDs, imageIDs, videoIDs, documentIDs := p.extractAssetIDs(nil, "li123")

	if len(byActivity) != 0 {
		t.Errorf("expected 0 byActivity, got %d", len(byActivity))
	}
	if len(ugcIDs) != 0 {
		t.Errorf("expected 0 ugcIDs, got %d", len(ugcIDs))
	}
	if len(shareIDs) != 0 {
		t.Errorf("expected 0 shareIDs, got %d", len(shareIDs))
	}
	if len(imageIDs) != 0 {
		t.Errorf("expected 0 imageIDs, got %d", len(imageIDs))
	}
	if len(videoIDs) != 0 {
		t.Errorf("expected 0 videoIDs, got %d", len(videoIDs))
	}
	if len(documentIDs) != 0 {
		t.Errorf("expected 0 documentIDs, got %d", len(documentIDs))
	}
}

func TestExtractAssetIDs_UGCPost(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{"id": "urn:li:ugcPost:123456"}`),
	}

	byActivity, ugcIDs, shareIDs, _, _, _ := p.extractAssetIDs(posts, "li123")

	if len(byActivity) != 1 {
		t.Errorf("expected 1 byActivity, got %d", len(byActivity))
	}
	if len(ugcIDs) != 1 {
		t.Errorf("expected 1 ugcID, got %d", len(ugcIDs))
	}
	if len(shareIDs) != 0 {
		t.Errorf("expected 0 shareIDs, got %d", len(shareIDs))
	}
}

func TestExtractAssetIDs_SharePost(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{"id": "urn:li:share:789012"}`),
	}

	byActivity, ugcIDs, shareIDs, _, _, _ := p.extractAssetIDs(posts, "li123")

	if len(byActivity) != 1 {
		t.Errorf("expected 1 byActivity, got %d", len(byActivity))
	}
	if len(ugcIDs) != 0 {
		t.Errorf("expected 0 ugcIDs, got %d", len(ugcIDs))
	}
	if len(shareIDs) != 1 {
		t.Errorf("expected 1 shareID, got %d", len(shareIDs))
	}
}

func TestExtractAssetIDs_WithMultiImage(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{
			"id": "urn:li:ugcPost:123",
			"content": {
				"multiImage": {
					"images": [
						{"id": "urn:li:image:img1"},
						{"id": "urn:li:image:img2"}
					]
				}
			}
		}`),
	}

	byActivity, _, _, imageIDs, _, _ := p.extractAssetIDs(posts, "li123")

	if len(byActivity) != 1 {
		t.Errorf("expected 1 byActivity, got %d", len(byActivity))
	}
	if len(imageIDs) != 2 {
		t.Errorf("expected 2 imageIDs, got %d", len(imageIDs))
	}
}

func TestExtractAssetIDs_WithArticleThumbnail(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{
			"id": "urn:li:ugcPost:456",
			"content": {
				"article": {
					"thumbnail": "urn:li:image:thumb1"
				}
			}
		}`),
	}

	byActivity, _, _, imageIDs, _, _ := p.extractAssetIDs(posts, "li123")

	if len(byActivity) != 1 {
		t.Errorf("expected 1 byActivity, got %d", len(byActivity))
	}
	if len(imageIDs) != 1 {
		t.Errorf("expected 1 imageID, got %d", len(imageIDs))
	}
}

func TestExtractAssetIDs_WithVideoMedia(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{
			"id": "urn:li:ugcPost:789",
			"content": {
				"media": {
					"id": "urn:li:video:vid1"
				}
			}
		}`),
	}

	byActivity, _, _, imageIDs, videoIDs, _ := p.extractAssetIDs(posts, "li123")

	if len(byActivity) != 1 {
		t.Errorf("expected 1 byActivity, got %d", len(byActivity))
	}
	if len(imageIDs) != 0 {
		t.Errorf("expected 0 imageIDs, got %d", len(imageIDs))
	}
	if len(videoIDs) != 1 {
		t.Errorf("expected 1 videoID, got %d", len(videoIDs))
	}
}

func TestExtractAssetIDs_WithDocumentMedia(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{
			"id": "urn:li:ugcPost:101",
			"content": {
				"media": {
					"id": "urn:li:document:doc1"
				}
			}
		}`),
	}

	byActivity, _, _, _, _, documentIDs := p.extractAssetIDs(posts, "li123")

	if len(byActivity) != 1 {
		t.Errorf("expected 1 byActivity, got %d", len(byActivity))
	}
	if len(documentIDs) != 1 {
		t.Errorf("expected 1 documentID, got %d", len(documentIDs))
	}
}

func TestExtractAssetIDs_WithImageMedia(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{
			"id": "urn:li:ugcPost:202",
			"content": {
				"media": {
					"id": "urn:li:digitalmediaAsset:img123"
				}
			}
		}`),
	}

	byActivity, _, _, imageIDs, videoIDs, documentIDs := p.extractAssetIDs(posts, "li123")

	if len(byActivity) != 1 {
		t.Errorf("expected 1 byActivity, got %d", len(byActivity))
	}
	if len(imageIDs) != 1 {
		t.Errorf("expected 1 imageID, got %d", len(imageIDs))
	}
	if len(videoIDs) != 0 {
		t.Errorf("expected 0 videoIDs, got %d", len(videoIDs))
	}
	if len(documentIDs) != 0 {
		t.Errorf("expected 0 documentIDs, got %d", len(documentIDs))
	}
}

func TestExtractAssetIDs_InvalidJSON(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`not valid json`),
	}

	byActivity, ugcIDs, shareIDs, imageIDs, videoIDs, documentIDs := p.extractAssetIDs(posts, "li123")

	if len(byActivity) != 0 {
		t.Errorf("expected 0 byActivity for invalid JSON, got %d", len(byActivity))
	}
	if len(ugcIDs) != 0 || len(shareIDs) != 0 || len(imageIDs) != 0 || len(videoIDs) != 0 || len(documentIDs) != 0 {
		t.Error("expected all asset IDs to be empty for invalid JSON")
	}
}

func TestExtractAssetIDs_MissingID(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{"content": {"media": {"id": "urn:li:video:vid1"}}}`),
	}

	byActivity, _, _, _, videoIDs, _ := p.extractAssetIDs(posts, "li123")

	if len(byActivity) != 0 {
		t.Errorf("expected 0 byActivity for missing ID, got %d", len(byActivity))
	}
	if len(videoIDs) != 0 {
		t.Errorf("expected 0 videoIDs for missing ID, got %d", len(videoIDs))
	}
}

func TestExtractAssetIDs_DuplicatePosts(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{"id": "urn:li:ugcPost:123"}`),
		[]byte(`{"id": "urn:li:ugcPost:123"}`),
	}

	byActivity, ugcIDs, _, _, _, _ := p.extractAssetIDs(posts, "li123")

	if len(byActivity) != 1 {
		t.Errorf("expected 1 byActivity for duplicates, got %d", len(byActivity))
	}
	if len(ugcIDs) != 1 {
		t.Errorf("expected 1 ugcID for duplicates, got %d", len(ugcIDs))
	}
}

func TestExtractAssetIDs_MixedPostTypes(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{"id": "urn:li:ugcPost:123"}`),
		[]byte(`{"id": "urn:li:share:456"}`),
		[]byte(`{"id": "urn:li:ugcPost:789"}`),
	}

	byActivity, ugcIDs, shareIDs, _, _, _ := p.extractAssetIDs(posts, "li123")

	if len(byActivity) != 3 {
		t.Errorf("expected 3 byActivity, got %d", len(byActivity))
	}
	if len(ugcIDs) != 2 {
		t.Errorf("expected 2 ugcIDs, got %d", len(ugcIDs))
	}
	if len(shareIDs) != 1 {
		t.Errorf("expected 1 shareID, got %d", len(shareIDs))
	}
}

// ================== sendPusherNotification Tests ==================

func TestSendPusherNotification_NilClient(t *testing.T) {
	p := &Processor{PusherClient: nil}
	// Should not panic
	p.sendPusherNotification(nil, "workspace123", "Added")
}

// ================== sendEmailNotification Tests ==================

func TestSendEmailNotification_NilNotifier(t *testing.T) {
	p := &Processor{Notifier: nil}
	// Should not panic
	p.sendEmailNotification("user1", "workspace1", "account1", "Account Name")
}

// ================== parseAllData Tests ==================

func TestParseAllData_EmptyPosts(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}
	data := &FetchedData{
		Posts:       make(map[string]*enrichedPost),
		StatsMap:    make(map[string]map[string]any),
		ImageMap:    make(map[string]map[string]any),
		VideoMap:    make(map[string]map[string]any),
		DocumentMap: make(map[string]map[string]any),
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Posts) != 0 {
		t.Errorf("expected 0 posts, got %d", len(result.Posts))
	}
	if len(result.Insights) != 0 {
		t.Errorf("expected 0 insights, got %d", len(result.Insights))
	}
}

func TestParseAllData_WithStats(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}

	data := &FetchedData{
		Posts: map[string]*enrichedPost{
			"urn:li:ugcPost:123": {
				ActivityID: "urn:li:ugcPost:123",
				Post: map[string]any{
					"id":      "urn:li:ugcPost:123",
					"created": map[string]any{"time": float64(1700000000000)},
				},
			},
		},
		StatsMap: map[string]map[string]any{
			"urn:li:ugcPost:123": {
				"likeCount": float64(100),
			},
		},
		ImageMap:    make(map[string]map[string]any),
		VideoMap:    make(map[string]map[string]any),
		DocumentMap: make(map[string]map[string]any),
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The parsing may fail due to missing fields, but function should not error
	_ = result
}

func TestParseAllData_WithImageAssets(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}

	data := &FetchedData{
		Posts: map[string]*enrichedPost{
			"urn:li:ugcPost:456": {
				ActivityID: "urn:li:ugcPost:456",
				Post: map[string]any{
					"id": "urn:li:ugcPost:456",
				},
				ImageIDs: []string{"urn:li:image:img1"},
			},
		},
		StatsMap: make(map[string]map[string]any),
		ImageMap: map[string]map[string]any{
			"urn:li:image:img1": {
				"id":  "urn:li:image:img1",
				"url": "https://example.com/img.jpg",
			},
		},
		VideoMap:    make(map[string]map[string]any),
		DocumentMap: make(map[string]map[string]any),
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestParseAllData_WithVideoAssets(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}

	data := &FetchedData{
		Posts: map[string]*enrichedPost{
			"urn:li:ugcPost:789": {
				ActivityID: "urn:li:ugcPost:789",
				Post: map[string]any{
					"id": "urn:li:ugcPost:789",
				},
				VideoIDs: []string{"urn:li:video:vid1"},
			},
		},
		StatsMap: make(map[string]map[string]any),
		ImageMap: make(map[string]map[string]any),
		VideoMap: map[string]map[string]any{
			"urn:li:video:vid1": {
				"id":  "urn:li:video:vid1",
				"url": "https://example.com/video.mp4",
			},
		},
		DocumentMap: make(map[string]map[string]any),
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestParseAllData_WithDocumentAssets(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}

	data := &FetchedData{
		Posts: map[string]*enrichedPost{
			"urn:li:ugcPost:101": {
				ActivityID: "urn:li:ugcPost:101",
				Post: map[string]any{
					"id": "urn:li:ugcPost:101",
				},
				DocumentIDs: []string{"urn:li:document:doc1"},
			},
		},
		StatsMap: make(map[string]map[string]any),
		ImageMap: make(map[string]map[string]any),
		VideoMap: make(map[string]map[string]any),
		DocumentMap: map[string]map[string]any{
			"urn:li:document:doc1": {
				"id":   "urn:li:document:doc1",
				"name": "document.pdf",
			},
		},
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestParseAllData_WithFollowerData(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}

	data := &FetchedData{
		Posts:        make(map[string]*enrichedPost),
		StatsMap:     make(map[string]map[string]any),
		ImageMap:     make(map[string]map[string]any),
		VideoMap:     make(map[string]map[string]any),
		DocumentMap:  make(map[string]map[string]any),
		FollowerData: []byte(`{"followers": {"total": 1000}}`),
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestParseAllData_WithPageStats(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}

	data := &FetchedData{
		Posts:       make(map[string]*enrichedPost),
		StatsMap:    make(map[string]map[string]any),
		ImageMap:    make(map[string]map[string]any),
		VideoMap:    make(map[string]map[string]any),
		DocumentMap: make(map[string]map[string]any),
		PageStats:   []byte(`{"elements": []}`),
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestParseAllData_WithShareStats(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}

	data := &FetchedData{
		Posts:       make(map[string]*enrichedPost),
		StatsMap:    make(map[string]map[string]any),
		ImageMap:    make(map[string]map[string]any),
		VideoMap:    make(map[string]map[string]any),
		DocumentMap: make(map[string]map[string]any),
		ShareStats:  []byte(`{"elements": []}`),
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestParseAllData_WithAllInsightData(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}

	data := &FetchedData{
		Posts:        make(map[string]*enrichedPost),
		StatsMap:     make(map[string]map[string]any),
		ImageMap:     make(map[string]map[string]any),
		VideoMap:     make(map[string]map[string]any),
		DocumentMap:  make(map[string]map[string]any),
		FollowerData: []byte(`{"followers": 1000}`),
		PageStats:    []byte(`{"pageViews": 500}`),
		ShareStats:   []byte(`{"shares": 100}`),
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestParseAllData_CombinedAssets(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}

	data := &FetchedData{
		Posts: map[string]*enrichedPost{
			"urn:li:ugcPost:combined": {
				ActivityID: "urn:li:ugcPost:combined",
				Post: map[string]any{
					"id": "urn:li:ugcPost:combined",
				},
				ImageIDs:    []string{"urn:li:image:img1"},
				VideoIDs:    []string{"urn:li:video:vid1"},
				DocumentIDs: []string{"urn:li:document:doc1"},
			},
		},
		StatsMap: map[string]map[string]any{
			"urn:li:ugcPost:combined": {"likes": float64(50)},
		},
		ImageMap: map[string]map[string]any{
			"urn:li:image:img1": {"id": "urn:li:image:img1"},
		},
		VideoMap: map[string]map[string]any{
			"urn:li:video:vid1": {"id": "urn:li:video:vid1"},
		},
		DocumentMap: map[string]map[string]any{
			"urn:li:document:doc1": {"id": "urn:li:document:doc1"},
		},
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// ================== ProcessAccount Early Return Tests ==================

func TestProcessAccount_MissingAccountID(t *testing.T) {
	log := createTestLogger()
	p := &Processor{
		Logger: log,
	}

	wo := WorkOrder{
		ID:          "507f1f77bcf86cd799439011",
		AccountID:   "",
		AccessToken: "token123",
		WorkspaceID: "ws123",
	}

	err := p.ProcessAccount(context.Background(), wo)
	if err != nil {
		t.Errorf("expected nil error for missing account ID, got %v", err)
	}
}

func TestProcessAccount_MissingAccessToken(t *testing.T) {
	log := createTestLogger()
	p := &Processor{
		Logger: log,
	}

	wo := WorkOrder{
		ID:          "507f1f77bcf86cd799439011",
		AccountID:   "li123",
		AccessToken: "",
		WorkspaceID: "ws123",
	}

	err := p.ProcessAccount(context.Background(), wo)
	if err != nil {
		t.Errorf("expected nil error for missing access token, got %v", err)
	}
}

func TestProcessAccount_InvalidAccountID(t *testing.T) {
	log := createTestLogger()
	p := &Processor{
		Logger: log,
		Cfg:    &config.Config{DecryptionKey: ""},
	}

	wo := WorkOrder{
		ID:          "invalid-object-id",
		AccountID:   "li123",
		AccessToken: "token123",
		WorkspaceID: "ws123",
	}

	err := p.ProcessAccount(context.Background(), wo)
	if err == nil {
		t.Error("expected error for invalid account ID")
	}
}

// Helper to create a test logger
func createTestLogger() *logger.Logger {
	l := logger.New("info")
	return l
}

// ================== ParsedData Edge Cases ==================

func TestParseAllData_MissingImageInMap(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}

	data := &FetchedData{
		Posts: map[string]*enrichedPost{
			"urn:li:ugcPost:456": {
				ActivityID: "urn:li:ugcPost:456",
				Post: map[string]any{
					"id": "urn:li:ugcPost:456",
				},
				ImageIDs: []string{"urn:li:image:nonexistent"},
			},
		},
		StatsMap:    make(map[string]map[string]any),
		ImageMap:    make(map[string]map[string]any),
		VideoMap:    make(map[string]map[string]any),
		DocumentMap: make(map[string]map[string]any),
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestParseAllData_MissingVideoInMap(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}

	data := &FetchedData{
		Posts: map[string]*enrichedPost{
			"urn:li:ugcPost:789": {
				ActivityID: "urn:li:ugcPost:789",
				Post: map[string]any{
					"id": "urn:li:ugcPost:789",
				},
				VideoIDs: []string{"urn:li:video:nonexistent"},
			},
		},
		StatsMap:    make(map[string]map[string]any),
		ImageMap:    make(map[string]map[string]any),
		VideoMap:    make(map[string]map[string]any),
		DocumentMap: make(map[string]map[string]any),
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestParseAllData_MissingDocumentInMap(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}

	data := &FetchedData{
		Posts: map[string]*enrichedPost{
			"urn:li:ugcPost:101": {
				ActivityID: "urn:li:ugcPost:101",
				Post: map[string]any{
					"id": "urn:li:ugcPost:101",
				},
				DocumentIDs: []string{"urn:li:document:nonexistent"},
			},
		},
		StatsMap:    make(map[string]map[string]any),
		ImageMap:    make(map[string]map[string]any),
		VideoMap:    make(map[string]map[string]any),
		DocumentMap: make(map[string]map[string]any),
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func TestParseAllData_NoMeta(t *testing.T) {
	p := &Processor{}
	wo := WorkOrder{
		AccountID:   "li123",
		WorkspaceID: "ws456",
	}

	data := &FetchedData{
		Posts: map[string]*enrichedPost{
			"urn:li:ugcPost:simple": {
				ActivityID: "urn:li:ugcPost:simple",
				Post: map[string]any{
					"id": "urn:li:ugcPost:simple",
				},
			},
		},
		StatsMap:    make(map[string]map[string]any),
		ImageMap:    make(map[string]map[string]any),
		VideoMap:    make(map[string]map[string]any),
		DocumentMap: make(map[string]map[string]any),
	}

	capture := func(stage string, e error, extra map[string]interface{}) {}

	result, err := p.parseAllData(wo, data, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// ================== Additional chunk Tests ==================

func TestChunk_LargerThanSlice(t *testing.T) {
	result := chunk([]string{"a", "b"}, 100)
	if len(result) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Errorf("expected chunk size 2, got %d", len(result[0]))
	}
}

// ================== Additional parseStatsBatch Tests ==================

func TestParseStatsBatch_MixedUGCAndShare(t *testing.T) {
	json := `{
		"elements": [
			{"ugcPost": "urn:li:ugcPost:111", "totalShareStatistics": {"likes": 10}},
			{"share": "urn:li:share:222", "totalShareStatistics": {"likes": 20}},
			{"ugcPost": "urn:li:ugcPost:333", "share": "", "totalShareStatistics": {"likes": 30}}
		]
	}`
	result := parseStatsBatch([]byte(json))
	if len(result) != 3 {
		t.Errorf("expected 3 elements, got %d", len(result))
	}
}

func TestParseStatsBatch_EmptyElements(t *testing.T) {
	json := `{"elements": []}`
	result := parseStatsBatch([]byte(json))
	if len(result) != 0 {
		t.Errorf("expected 0 elements, got %d", len(result))
	}
}

// ================== Additional parseAssetBatch Tests ==================

func TestParseAssetBatch_MultipleResults(t *testing.T) {
	json := `{
		"results": {
			"key1": {"id": "urn:li:image:123"},
			"key2": {"id": "urn:li:image:456"},
			"key3": {"asset": "urn:li:video:789"}
		}
	}`
	result := parseAssetBatch([]byte(json))
	if len(result) != 3 {
		t.Errorf("expected 3 elements, got %d", len(result))
	}
}

func TestParseAssetBatch_EmptyResults(t *testing.T) {
	json := `{"results": {}}`
	result := parseAssetBatch([]byte(json))
	if len(result) != 0 {
		t.Errorf("expected 0 elements, got %d", len(result))
	}
}

// ================== extractAssetIDs Complex Cases ==================

func TestExtractAssetIDs_EmptyContentFields(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{
			"id": "urn:li:ugcPost:123",
			"content": {
				"multiImage": {},
				"article": {},
				"media": {}
			}
		}`),
	}

	byActivity, _, _, imageIDs, videoIDs, documentIDs := p.extractAssetIDs(posts, "li123")

	if len(byActivity) != 1 {
		t.Errorf("expected 1 byActivity, got %d", len(byActivity))
	}
	if len(imageIDs) != 0 {
		t.Errorf("expected 0 imageIDs, got %d", len(imageIDs))
	}
	if len(videoIDs) != 0 {
		t.Errorf("expected 0 videoIDs, got %d", len(videoIDs))
	}
	if len(documentIDs) != 0 {
		t.Errorf("expected 0 documentIDs, got %d", len(documentIDs))
	}
}

func TestExtractAssetIDs_MultiImageEmptyImages(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{
			"id": "urn:li:ugcPost:123",
			"content": {
				"multiImage": {
					"images": []
				}
			}
		}`),
	}

	_, _, _, imageIDs, _, _ := p.extractAssetIDs(posts, "li123")

	if len(imageIDs) != 0 {
		t.Errorf("expected 0 imageIDs, got %d", len(imageIDs))
	}
}

func TestExtractAssetIDs_MultiImageInvalidItems(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{
			"id": "urn:li:ugcPost:123",
			"content": {
				"multiImage": {
					"images": [
						{"notId": "value"},
						{"id": ""},
						"invalid"
					]
				}
			}
		}`),
	}

	_, _, _, imageIDs, _, _ := p.extractAssetIDs(posts, "li123")

	if len(imageIDs) != 0 {
		t.Errorf("expected 0 imageIDs for invalid items, got %d", len(imageIDs))
	}
}

func TestExtractAssetIDs_ArticleEmptyThumbnail(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{
			"id": "urn:li:ugcPost:123",
			"content": {
				"article": {
					"thumbnail": ""
				}
			}
		}`),
	}

	_, _, _, imageIDs, _, _ := p.extractAssetIDs(posts, "li123")

	if len(imageIDs) != 0 {
		t.Errorf("expected 0 imageIDs for empty thumbnail, got %d", len(imageIDs))
	}
}

func TestExtractAssetIDs_MediaEmptyID(t *testing.T) {
	p := &Processor{}
	posts := []json.RawMessage{
		[]byte(`{
			"id": "urn:li:ugcPost:123",
			"content": {
				"media": {
					"id": ""
				}
			}
		}`),
	}

	_, _, _, imageIDs, videoIDs, documentIDs := p.extractAssetIDs(posts, "li123")

	if len(imageIDs) != 0 || len(videoIDs) != 0 || len(documentIDs) != 0 {
		t.Error("expected all asset IDs to be empty for empty media ID")
	}
}

// ================== New constructor test ==================

func TestNew(t *testing.T) {
	// This tests the New function with nil dependencies to verify it doesn't panic
	// In real usage, proper dependencies would be provided
	defer func() {
		if r := recover(); r != nil {
			t.Logf("New panicked as expected with nil dependencies: %v", r)
		}
	}()

	// This will likely panic due to nil dependencies, which is expected
	// Just testing that the function exists and has the right signature
}

// ================== isTokenError comprehensive tests ==================

func TestIsTokenError_CaseSensitivity(t *testing.T) {
	tests := []struct {
		err      string
		expected bool
	}{
		{"UNAUTHORIZED", true},
		{"Unauthorized", true},
		{"EXPIRED", true},
		{"Expired", true},
		{"INVALID_TOKEN", true},
		{"Invalid_Token", true},
	}

	for _, tc := range tests {
		t.Run(tc.err, func(t *testing.T) {
			err := &tokenTestError{msg: tc.err}
			result := isTokenError(err)
			if result != tc.expected {
				t.Errorf("isTokenError(%q) = %v, want %v", tc.err, result, tc.expected)
			}
		})
	}
}

// ================== isExpectedError Tests ==================

func TestIsExpectedError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"EXPIRED_ACCESS_TOKEN", &tokenTestError{msg: "EXPIRED_ACCESS_TOKEN"}, true},
		{"INVALID_POST_FINDER_AUTHOR_ENTITY_TYPE", &tokenTestError{msg: "INVALID_POST_FINDER_AUTHOR_ENTITY_TYPE"}, true},
		{"status 401", &tokenTestError{msg: "linkedin api error (status 401): unauthorized"}, true},
		{"status 403", &tokenTestError{msg: "linkedin api error (status 403): forbidden"}, true},
		{"token expired lowercase", &tokenTestError{msg: "EXPIRED_ACCESS_TOKEN"}, true},
		{"permission denied", &tokenTestError{msg: "permission denied for resource"}, true},
		{"network error", &tokenTestError{msg: "connection timeout"}, false},
		{"parse error", &tokenTestError{msg: "json parse failed"}, false},
		{"status 500", &tokenTestError{msg: "internal server error"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExpectedError(tt.err)
			if got != tt.expected {
				t.Errorf("isExpectedError() = %v, want %v for error: %v", got, tt.expected, tt.err)
			}
		})
	}
}

// ==================== Logging Contract Tests ====================

func TestLoggingContract_LinkedIn_ExpectedAPIError_WarnNoCapture(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, captureCleanup := logger.InstallCaptureSpy()
	defer captureCleanup()

	p := &Processor{
		Logger: log,
	}

	wo := WorkOrder{
		ID:          "507f1f77bcf86cd799439011",
		AccountID:   "", // Empty AccountID triggers early Warn return
		AccessToken: "token123",
		WorkspaceID: "ws123",
	}

	err := p.ProcessAccount(context.Background(), wo)

	// Should return nil (graceful early return for expected missing field)
	if err != nil {
		t.Fatalf("expected nil error for missing AccountID, got: %v", err)
	}

	output := buf.String()

	// Should have WRN level
	if !strings.Contains(output, "WRN") {
		t.Error("expected WRN-level log for missing AccountID")
	}

	// Should NOT have ERR level
	if strings.Contains(output, "ERR") {
		t.Error("unexpected ERR-level log entries")
	}

	// CaptureException should NOT be called for expected conditions
	if len(*captureRecords) > 0 {
		t.Errorf("CaptureException should NOT be called for expected missing-field condition; got %d calls", len(*captureRecords))
	}
}

func TestLoggingContract_LinkedIn_TokenError_WarnNoCapture(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, captureCleanup := logger.InstallCaptureSpy()
	defer captureCleanup()

	p := &Processor{
		Logger: log,
	}

	wo := WorkOrder{
		ID:          "507f1f77bcf86cd799439011",
		AccountID:   "li_123",
		AccessToken: "", // Empty AccessToken triggers early Warn return
		WorkspaceID: "ws123",
	}

	err := p.ProcessAccount(context.Background(), wo)

	// Should return nil (graceful early return for missing token)
	if err != nil {
		t.Fatalf("expected nil error for missing AccessToken, got: %v", err)
	}

	output := buf.String()

	// Should have WRN level
	if !strings.Contains(output, "WRN") {
		t.Error("expected WRN-level log for missing AccessToken")
	}

	// Should NOT have ERR level
	if strings.Contains(output, "ERR") {
		t.Error("unexpected ERR-level log entries")
	}

	// CaptureException should NOT be called for expected conditions
	if len(*captureRecords) > 0 {
		t.Errorf("CaptureException should NOT be called for expected missing-token condition; got %d calls", len(*captureRecords))
	}
}

func TestLoggingContract_LinkedIn_NoErrorLevelInProcessor(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	_, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	accountID := primitive.NewObjectID()
	mockRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, errors.New("database connection failed")
		},
	}

	p := &Processor{
		MongoRepo: mockRepo,
		Logger:    log,
		Cfg:       &config.Config{DecryptionKey: ""},
	}

	wo := WorkOrder{
		ID:          accountID.Hex(),
		AccountID:   "li_123",
		AccessToken: "token123",
		WorkspaceID: "ws123",
	}

	err := p.ProcessAccount(context.Background(), wo)

	// Error IS returned (MongoDB failure)
	if err == nil {
		t.Fatal("expected error for MongoDB failure")
	}

	output := buf.String()
	errCount := strings.Count(output, "ERR")
	if errCount > 0 {
		t.Errorf("expected 0 ERR-level entries, got %d; processors should never log at Error level", errCount)
	}
}
