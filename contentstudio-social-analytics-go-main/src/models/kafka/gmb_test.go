package kafka

import (
	"encoding/json"
	"testing"
)

func TestGMBAccountWorkOrder_JSONRoundTrip(t *testing.T) {
	wo := GMBAccountWorkOrder{
		ID:           "64a1b2c3d4e5f60001",
		WorkspaceID:  "ws123",
		AccountID:    "accounts/123456",
		LocationID:   "locations/789",
		AccessToken:  "ya29.test-token",
		RefreshToken: "1//test-refresh",
		AccountName:  "Test Business",
		LocationName: "Main Street Location",
		LanguageCode: "en",
		SyncType:     "incremental",
	}

	data, err := json.Marshal(wo)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded GMBAccountWorkOrder
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.LocationID != wo.LocationID {
		t.Errorf("LocationID mismatch: got %q, want %q", decoded.LocationID, wo.LocationID)
	}
	if decoded.LanguageCode != wo.LanguageCode {
		t.Errorf("LanguageCode mismatch: got %q, want %q", decoded.LanguageCode, wo.LanguageCode)
	}
}

func TestParsedGMBDailyMetrics_JSONRoundTrip(t *testing.T) {
	m := ParsedGMBDailyMetrics{
		AccountID:                        "accounts/123",
		LocationID:                       "locations/456",
		AccountName:                      "Test Business",
		LocationName:                     "Downtown",
		Date:                             "2026-01-15",
		BusinessImpressionsDesktopMaps:   100,
		BusinessImpressionsDesktopSearch: 200,
		BusinessImpressionsMobileMaps:    300,
		BusinessImpressionsMobileSearch:  400,
		CallClicks:                       50,
		WebsiteClicks:                    60,
		BusinessDirectionRequests:        70,
		BusinessConversations:            80,
		BusinessBookings:                 90,
		BusinessFoodOrders:               10,
		BusinessFoodMenuClicks:           20,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ParsedGMBDailyMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.CallClicks != m.CallClicks {
		t.Errorf("CallClicks mismatch: got %d, want %d", decoded.CallClicks, m.CallClicks)
	}
}

func TestParsedGMBReview_JSONRoundTrip(t *testing.T) {
	r := ParsedGMBReview{
		AccountID:               "accounts/123",
		LocationID:              "locations/456",
		ReviewID:                "review1",
		ReviewName:              "accounts/123/locations/456/reviews/review1",
		ReviewerDisplayName:     "John Doe",
		ReviewerProfilePhotoURL: "https://example.com/photo.jpg",
		StarRating:              "FOUR",
		Comment:                 "Great place!",
		CreateTime:              "2025-06-24T10:38:25Z",
		UpdateTime:              "2025-06-24T10:38:25Z",
		ReplyComment:    "Thank you!",
		ReplyUpdateTime: "2025-06-25T08:00:00Z",
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ParsedGMBReview
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ReviewID != r.ReviewID {
		t.Errorf("ReviewID mismatch: got %q, want %q", decoded.ReviewID, r.ReviewID)
	}
	if decoded.ReplyComment != "Thank you!" {
		t.Errorf("ReplyComment mismatch: got %q, want %q", decoded.ReplyComment, "Thank you!")
	}
}

func TestParsedGMBLocalPost_JSONRoundTrip(t *testing.T) {
	p := ParsedGMBLocalPost{
		AccountID:       "accounts/123",
		LocationID:      "locations/456",
		PostName:        "accounts/123/locations/456/localPosts/post123",
		State:           "LIVE",
		TopicType:       "STANDARD",
		SearchURL:       "https://search.example.com",
		CreateTime:      "2025-11-11T11:07:51Z",
		UpdateTime:      "2025-11-11T11:07:51Z",
		MediaNames:      []string{"media1"},
		MediaFormats:    []string{"PHOTO"},
		MediaGoogleURLs: []string{"https://lh3.googleusercontent.com/1"},
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ParsedGMBLocalPost
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.PostName != p.PostName {
		t.Errorf("PostName mismatch: got %q, want %q", decoded.PostName, p.PostName)
	}
}

func TestRawGMBData_JSONRoundTrip(t *testing.T) {
	raw := RawGMBData{
		WorkspaceID: "ws123",
		AccountID:   "accounts/123",
		LocationID:  "locations/456",
		DataType:    "daily_metrics",
		Data:        []byte(`{"test": "data"}`),
	}

	data, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RawGMBData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.DataType != raw.DataType {
		t.Errorf("DataType mismatch: got %q, want %q", decoded.DataType, raw.DataType)
	}
}
