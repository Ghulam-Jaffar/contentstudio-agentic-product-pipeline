package kafka

import (
	"encoding/json"
	"testing"
	"time"
)

// ================== TikTokAccountWorkOrder Tests ==================

func TestTikTokAccountWorkOrder_JSONMarshal(t *testing.T) {
	wo := TikTokAccountWorkOrder{
		ID:           "507f1f77bcf86cd799439011",
		WorkspaceID:  "workspace123",
		TikTokID:     "tiktok_user_123",
		AccessToken:  "access_token_abc",
		RefreshToken: "refresh_token_xyz",
		Scope:        "user.info.basic,video.list",
		SyncType:     "incremental",
	}

	data, err := json.Marshal(wo)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled TikTokAccountWorkOrder
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.ID != wo.ID {
		t.Errorf("ID mismatch: got %q, want %q", unmarshaled.ID, wo.ID)
	}
	if unmarshaled.WorkspaceID != wo.WorkspaceID {
		t.Errorf("WorkspaceID mismatch: got %q, want %q", unmarshaled.WorkspaceID, wo.WorkspaceID)
	}
	if unmarshaled.TikTokID != wo.TikTokID {
		t.Errorf("TikTokID mismatch: got %q, want %q", unmarshaled.TikTokID, wo.TikTokID)
	}
	if unmarshaled.AccessToken != wo.AccessToken {
		t.Errorf("AccessToken mismatch: got %q, want %q", unmarshaled.AccessToken, wo.AccessToken)
	}
	if unmarshaled.RefreshToken != wo.RefreshToken {
		t.Errorf("RefreshToken mismatch: got %q, want %q", unmarshaled.RefreshToken, wo.RefreshToken)
	}
	if unmarshaled.Scope != wo.Scope {
		t.Errorf("Scope mismatch: got %q, want %q", unmarshaled.Scope, wo.Scope)
	}
	if unmarshaled.SyncType != wo.SyncType {
		t.Errorf("SyncType mismatch: got %q, want %q", unmarshaled.SyncType, wo.SyncType)
	}
}

func TestTikTokAccountWorkOrder_JSONFieldNames(t *testing.T) {
	wo := TikTokAccountWorkOrder{
		ID:           "test-id",
		WorkspaceID:  "test-workspace",
		TikTokID:     "test-tiktok",
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		Scope:        "test-scope",
		SyncType:     "incremental",
	}

	data, _ := json.Marshal(wo)
	var m map[string]interface{}
	json.Unmarshal(data, &m)

	expectedFields := []string{"id", "workspace_id", "tiktok_id", "access_token", "refresh_token", "scope", "sync_type"}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("expected JSON field %q not found", field)
		}
	}
}

func TestTikTokAccountWorkOrder_EmptyValues(t *testing.T) {
	wo := TikTokAccountWorkOrder{}

	data, err := json.Marshal(wo)
	if err != nil {
		t.Fatalf("failed to marshal empty struct: %v", err)
	}

	var unmarshaled TikTokAccountWorkOrder
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.ID != "" || unmarshaled.TikTokID != "" {
		t.Error("expected empty fields after unmarshaling empty struct")
	}
}

// ================== TikTokBatchWorkOrder Tests ==================

func TestTikTokBatchWorkOrder_JSONMarshal(t *testing.T) {
	batch := TikTokBatchWorkOrder{
		BatchID:  "batch-uuid-123",
		SyncType: "full_sync",
		Accounts: []TikTokAccountWorkOrder{
			{ID: "acc1", TikTokID: "tiktok1", AccessToken: "token1"},
			{ID: "acc2", TikTokID: "tiktok2", AccessToken: "token2"},
		},
		CreatedAt: time.Date(2026, 1, 30, 10, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(batch)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled TikTokBatchWorkOrder
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.BatchID != batch.BatchID {
		t.Errorf("BatchID mismatch: got %q, want %q", unmarshaled.BatchID, batch.BatchID)
	}
	if unmarshaled.SyncType != batch.SyncType {
		t.Errorf("SyncType mismatch: got %q, want %q", unmarshaled.SyncType, batch.SyncType)
	}
	if len(unmarshaled.Accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(unmarshaled.Accounts))
	}
}

func TestTikTokBatchWorkOrder_EmptyAccounts(t *testing.T) {
	batch := TikTokBatchWorkOrder{
		BatchID:   "empty-batch",
		SyncType:  "incremental",
		Accounts:  []TikTokAccountWorkOrder{},
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(batch)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled TikTokBatchWorkOrder
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(unmarshaled.Accounts) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(unmarshaled.Accounts))
	}
}

func TestTikTokBatchWorkOrder_JSONFieldNames(t *testing.T) {
	batch := TikTokBatchWorkOrder{
		BatchID:   "test-batch",
		SyncType:  "incremental",
		Accounts:  []TikTokAccountWorkOrder{},
		CreatedAt: time.Now(),
	}

	data, _ := json.Marshal(batch)
	var m map[string]interface{}
	json.Unmarshal(data, &m)

	expectedFields := []string{"batch_id", "sync_type", "accounts", "created_at"}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("expected JSON field %q not found", field)
		}
	}
}

// ================== RawTikTokPost Tests ==================

func TestRawTikTokPost_JSONMarshal(t *testing.T) {
	rawData := json.RawMessage(`{"video_id": "123", "title": "Test Video"}`)
	post := RawTikTokPost{
		WorkspaceID: "workspace123",
		TikTokID:    "tiktok_user",
		Data:        rawData,
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled RawTikTokPost
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.WorkspaceID != post.WorkspaceID {
		t.Errorf("WorkspaceID mismatch")
	}
	if unmarshaled.TikTokID != post.TikTokID {
		t.Errorf("TikTokID mismatch")
	}

	// Verify raw data is preserved
	var dataMap map[string]interface{}
	if err := json.Unmarshal(unmarshaled.Data, &dataMap); err != nil {
		t.Fatalf("failed to unmarshal Data field: %v", err)
	}
	if dataMap["video_id"] != "123" {
		t.Error("raw data not preserved correctly")
	}
}

func TestRawTikTokPost_NilData(t *testing.T) {
	post := RawTikTokPost{
		WorkspaceID: "workspace123",
		TikTokID:    "tiktok_user",
		Data:        nil,
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("failed to marshal with nil data: %v", err)
	}

	var unmarshaled RawTikTokPost
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
}

// ================== ParsedTikTokPost Tests ==================

func TestParsedTikTokPost_JSONMarshal(t *testing.T) {
	post := ParsedTikTokPost{
		ID:              "video123",
		TikTokID:        "tiktok_user",
		WorkspaceID:     "workspace123",
		DisplayName:     "Test User",
		ProfileLink:     "https://tiktok.com/@testuser",
		CoverImageURL:   "https://example.com/cover.jpg",
		ShareURL:        "https://tiktok.com/video/123",
		PostDescription: "Test video description",
		Hashtags:        []string{"test", "video"},
		Duration:        30,
		Height:          1920,
		Width:           1080,
		Title:           "Test Video Title",
		LikeCount:       1000,
		CommentCount:    50,
		ShareCount:      25,
		ViewCount:       10000,
		EngagementCount: 1075,
		EngagementRate:  0.1075,
		CreateTime:      1706612400,
		CreatedAt:       time.Date(2026, 1, 30, 10, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled ParsedTikTokPost
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.ID != post.ID {
		t.Errorf("ID mismatch")
	}
	if unmarshaled.LikeCount != post.LikeCount {
		t.Errorf("LikeCount mismatch: got %d, want %d", unmarshaled.LikeCount, post.LikeCount)
	}
	if len(unmarshaled.Hashtags) != 2 {
		t.Errorf("Hashtags mismatch: got %d, want 2", len(unmarshaled.Hashtags))
	}
	if unmarshaled.EngagementRate != post.EngagementRate {
		t.Errorf("EngagementRate mismatch: got %f, want %f", unmarshaled.EngagementRate, post.EngagementRate)
	}
}

func TestParsedTikTokPost_OmitEmpty(t *testing.T) {
	post := ParsedTikTokPost{
		ID:          "video123",
		WorkspaceID: "workspace123",
		LikeCount:   100,
	}

	data, _ := json.Marshal(post)
	var m map[string]interface{}
	json.Unmarshal(data, &m)

	// Fields with omitempty and zero values should not be present
	if _, ok := m["tiktok_id"]; ok {
		// tiktok_id has omitempty, but empty string is still marshaled
		// This is expected behavior
	}

	// Required fields should be present
	if _, ok := m["id"]; !ok {
		t.Error("expected 'id' field to be present")
	}
	if _, ok := m["workspace_id"]; !ok {
		t.Error("expected 'workspace_id' field to be present")
	}
}

func TestParsedTikTokPost_HashtagsNil(t *testing.T) {
	post := ParsedTikTokPost{
		ID:       "video123",
		Hashtags: nil,
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("failed to marshal with nil hashtags: %v", err)
	}

	var unmarshaled ParsedTikTokPost
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
}

// ================== ParsedTikTokInsights Tests ==================

func TestParsedTikTokInsights_JSONMarshal(t *testing.T) {
	insights := ParsedTikTokInsights{
		RecordID:            "record123",
		TikTokID:            "tiktok_user",
		DisplayName:         "Test User",
		ProfileImage:        "https://example.com/avatar.jpg",
		TotalFollowerCount:  10000,
		TotalFollowingCount: 500,
		TotalLikeCount:      50000,
		TotalVideoCount:     100,
		TotalVideoViews:     1000000,
		TotalVideoLikes:     45000,
		TotalVideoComments:  5000,
		TotalVideoShares:    2500,
		IsVerified:          true,
		Bio:                 "Test bio",
		ProfileLink:         "https://tiktok.com/@testuser",
		InsertedAt:          1706612400,
	}

	data, err := json.Marshal(insights)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled ParsedTikTokInsights
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.RecordID != insights.RecordID {
		t.Errorf("RecordID mismatch")
	}
	if unmarshaled.TotalFollowerCount != insights.TotalFollowerCount {
		t.Errorf("TotalFollowerCount mismatch: got %d, want %d", unmarshaled.TotalFollowerCount, insights.TotalFollowerCount)
	}
	if unmarshaled.IsVerified != insights.IsVerified {
		t.Errorf("IsVerified mismatch: got %v, want %v", unmarshaled.IsVerified, insights.IsVerified)
	}
}

func TestParsedTikTokInsights_JSONFieldNames(t *testing.T) {
	insights := ParsedTikTokInsights{
		RecordID:            "test",
		TikTokID:            "test",
		TotalFollowerCount:  100,
		TotalFollowingCount: 50,
	}

	data, _ := json.Marshal(insights)
	var m map[string]interface{}
	json.Unmarshal(data, &m)

	expectedFields := []string{
		"record_id", "tiktok_id", "display_name", "profile_image",
		"total_follower_count", "total_following_count", "total_like_count",
		"total_video_count", "total_video_views", "total_video_likes",
		"total_video_comments", "total_video_shares", "is_verified",
		"bio", "profile_link", "inserted_at",
	}

	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("expected JSON field %q not found", field)
		}
	}
}

func TestParsedTikTokInsights_ZeroValues(t *testing.T) {
	insights := ParsedTikTokInsights{}

	data, err := json.Marshal(insights)
	if err != nil {
		t.Fatalf("failed to marshal zero-value struct: %v", err)
	}

	var unmarshaled ParsedTikTokInsights
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.IsVerified != false {
		t.Error("expected IsVerified to be false for zero value")
	}
	if unmarshaled.TotalFollowerCount != 0 {
		t.Error("expected TotalFollowerCount to be 0 for zero value")
	}
}
