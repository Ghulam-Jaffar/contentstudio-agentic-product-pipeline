package clickhouse

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGMBDailyMetrics_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	m := GMBDailyMetrics{
		AccountID:                        "acc123",
		LocationID:                       "loc123",
		AccountName:                      "Test Account",
		LocationName:                     "Test Location",
		PlatformName:                     "gmb",
		InsertedAt:                       now,
		CreatedAt:                        now,
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

	var decoded GMBDailyMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.BusinessImpressionsDesktopMaps != m.BusinessImpressionsDesktopMaps {
		t.Errorf("BusinessImpressionsDesktopMaps mismatch: got %d, want %d", decoded.BusinessImpressionsDesktopMaps, m.BusinessImpressionsDesktopMaps)
	}
}

func TestGMBMediaAssets_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	m := GMBMediaAssets{
		AccountID:                   "acc123",
		LocationID:                  "loc123",
		AccountName:                 "Test Account",
		LocationName:                "Test Location",
		PlatformName:                "gmb",
		LanguageCode:                "en",
		InsertedAt:                  now,
		CreatedAt:                   now,
		MediaName:                   "accounts/acc123/locations/loc123/media/photo1",
		SourceURL:                   "https://example.com/photo.jpg",
		MediaFormat:                 "PHOTO",
		LocationAssociationCategory: "EXTERIOR",
		GoogleURL:                   "https://lh3.googleusercontent.com/photo1",
		ThumbnailURL:                "https://lh3.googleusercontent.com/photo1_thumb",
		WidthPixels:                 1920,
		HeightPixels:                1080,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded GMBMediaAssets
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.MediaName != m.MediaName {
		t.Errorf("MediaName mismatch: got %q, want %q", decoded.MediaName, m.MediaName)
	}
	if decoded.WidthPixels != m.WidthPixels {
		t.Errorf("WidthPixels mismatch: got %d, want %d", decoded.WidthPixels, m.WidthPixels)
	}
}

func TestGMBSearchKeywordsMonthly_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	m := GMBSearchKeywordsMonthly{
		AccountID:            "acc123",
		LocationID:           "loc123",
		AccountName:          "Test Account",
		LocationName:         "Test Location",
		PlatformName:         "gmb",
		InsertedAt:           now,
		KeywordMonth:         now,
		Keyword:              "coffee shop",
		ImpressionsValue:     150,
		ImpressionsThreshold: 15,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded GMBSearchKeywordsMonthly
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Keyword != m.Keyword {
		t.Errorf("Keyword mismatch: got %q, want %q", decoded.Keyword, m.Keyword)
	}
}

func TestGMBLocalPosts_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	m := GMBLocalPosts{
		AccountID:       "acc123",
		LocationID:      "loc123",
		AccountName:     "Test Account",
		LocationName:    "Test Location",
		PlatformName:    "gmb",
		LanguageCode:    "en",
		InsertedAt:      now,
		CreatedAt:       now,
		UpdatedAt:       now,
		PostName:        "accounts/acc123/locations/loc123/localPosts/post123",
		State:           "LIVE",
		TopicType:       "STANDARD",
		SearchURL:       "https://search.example.com/post123",
		MediaNames:      []string{"media1", "media2"},
		MediaFormats:    []string{"PHOTO", "VIDEO"},
		MediaGoogleURLs: []string{"https://lh3.googleusercontent.com/1", "https://lh3.googleusercontent.com/2"},
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded GMBLocalPosts
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.PostName != m.PostName {
		t.Errorf("PostName mismatch: got %q, want %q", decoded.PostName, m.PostName)
	}
	if len(decoded.MediaNames) != len(m.MediaNames) {
		t.Errorf("MediaNames length mismatch: got %d, want %d", len(decoded.MediaNames), len(m.MediaNames))
	}
}

func TestGMBReviews_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	m := GMBReviews{
		AccountID:               "acc123",
		LocationID:              "loc123",
		AccountName:             "Test Account",
		LocationName:            "Test Location",
		PlatformName:            "gmb",
		InsertedAt:              now,
		CreatedAt:               now,
		UpdatedAt:               now,
		ReviewID:                "review123",
		ReviewName:              "accounts/acc123/locations/loc123/reviews/review123",
		ReviewerDisplayName:     "John Doe",
		ReviewerProfilePhotoURL: "https://example.com/photo.jpg",
		StarRating:              4,
		Comment:                 "Great place!",
		ReplyComment:    "Thank you!",
		ReplyUpdateTime: now,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded GMBReviews
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ReviewID != m.ReviewID {
		t.Errorf("ReviewID mismatch: got %q, want %q", decoded.ReviewID, m.ReviewID)
	}
	if decoded.StarRating != m.StarRating {
		t.Errorf("StarRating mismatch: got %d, want %d", decoded.StarRating, m.StarRating)
	}
	if decoded.ReplyComment != m.ReplyComment {
		t.Errorf("ReplyComment mismatch: got %q, want %q", decoded.ReplyComment, m.ReplyComment)
	}
}

func TestGMBDailyMetrics_StructFields(t *testing.T) {
	m := GMBDailyMetrics{}
	if m.BusinessImpressionsDesktopMaps != 0 {
		t.Error("Expected zero value for BusinessImpressionsDesktopMaps")
	}
}
