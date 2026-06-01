package main

import (
	"testing"
	"time"

	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if maxBatchSize <= 0 {
		t.Errorf("maxBatchSize should be > 0, got %d", maxBatchSize)
	}
	if maxBatchSize != 10000 {
		t.Errorf("maxBatchSize = %d, want 10000", maxBatchSize)
	}
	if batchTimeout <= 0 {
		t.Errorf("batchTimeout should be > 0, got %v", batchTimeout)
	}
	if batchTimeout != 10*time.Second {
		t.Errorf("batchTimeout = %v, want 10s", batchTimeout)
	}
	if batchProcessorsPerType <= 0 {
		t.Errorf("batchProcessorsPerType should be > 0, got %d", batchProcessorsPerType)
	}
	if batchProcessorsPerType != 3 {
		t.Errorf("batchProcessorsPerType = %d, want 3", batchProcessorsPerType)
	}
	if messageChanSize <= 0 {
		t.Errorf("messageChanSize should be > 0, got %d", messageChanSize)
	}
	if messageChanSize != 50000 {
		t.Errorf("messageChanSize = %d, want 50000", messageChanSize)
	}
}

func TestConsumerGroupConstant(t *testing.T) {
	if consumerGroup == "" {
		t.Error("consumerGroup should not be empty")
	}
	if consumerGroup != "pinterest-analytics-sink-group" {
		t.Errorf("consumerGroup = %q, want %q", consumerGroup, "pinterest-analytics-sink-group")
	}
}

func TestIdleTimeoutConstants(t *testing.T) {
	if idleTimeout <= 0 {
		t.Errorf("idleTimeout should be > 0, got %v", idleTimeout)
	}
	if idleTimeout != 5*time.Minute {
		t.Errorf("idleTimeout = %v, want 5m", idleTimeout)
	}
	if idleCheckInterval <= 0 {
		t.Errorf("idleCheckInterval should be > 0, got %v", idleCheckInterval)
	}
	if idleCheckInterval != 30*time.Second {
		t.Errorf("idleCheckInterval = %v, want 30s", idleCheckInterval)
	}
}

// ================== RawMessage Tests ==================

func TestRawMessage_Fields(t *testing.T) {
	msg := RawMessage{
		Topic: "test-topic",
		Key:   []byte("test-key"),
		Value: []byte(`{"id": "123"}`),
	}

	if msg.Topic != "test-topic" {
		t.Errorf("Topic = %q, want %q", msg.Topic, "test-topic")
	}
	if string(msg.Key) != "test-key" {
		t.Errorf("Key = %q, want %q", string(msg.Key), "test-key")
	}
	if string(msg.Value) != `{"id": "123"}` {
		t.Errorf("Value = %q, want %q", string(msg.Value), `{"id": "123"}`)
	}
}

// ================== BatchCollectors Tests ==================

func TestBatchCollectors_Initialization(t *testing.T) {
	batches := &BatchCollectors{
		users:        make(chan *chmodels.PinterestUser, 100),
		boards:       make(chan *chmodels.PinterestBoard, 100),
		pins:         make(chan *chmodels.PinterestPin, 100),
		pinInsights:  make(chan *chmodels.PinterestPinInsight, 100),
		userInsights: make(chan *chmodels.PinterestUserInsight, 100),
	}

	if batches.users == nil {
		t.Error("users channel should not be nil")
	}
	if batches.boards == nil {
		t.Error("boards channel should not be nil")
	}
	if batches.pins == nil {
		t.Error("pins channel should not be nil")
	}
	if batches.pinInsights == nil {
		t.Error("pinInsights channel should not be nil")
	}
	if batches.userInsights == nil {
		t.Error("userInsights channel should not be nil")
	}
}

func TestBatchCollectors_ChannelCapacity(t *testing.T) {
	expectedCapacity := maxBatchSize * 5

	batches := &BatchCollectors{
		users:        make(chan *chmodels.PinterestUser, expectedCapacity),
		boards:       make(chan *chmodels.PinterestBoard, expectedCapacity),
		pins:         make(chan *chmodels.PinterestPin, expectedCapacity),
		pinInsights:  make(chan *chmodels.PinterestPinInsight, expectedCapacity),
		userInsights: make(chan *chmodels.PinterestUserInsight, expectedCapacity),
	}

	if cap(batches.users) != expectedCapacity {
		t.Errorf("users channel capacity = %d, want %d", cap(batches.users), expectedCapacity)
	}
	if cap(batches.boards) != expectedCapacity {
		t.Errorf("boards channel capacity = %d, want %d", cap(batches.boards), expectedCapacity)
	}
	if cap(batches.pins) != expectedCapacity {
		t.Errorf("pins channel capacity = %d, want %d", cap(batches.pins), expectedCapacity)
	}
	if cap(batches.pinInsights) != expectedCapacity {
		t.Errorf("pinInsights channel capacity = %d, want %d", cap(batches.pinInsights), expectedCapacity)
	}
	if cap(batches.userInsights) != expectedCapacity {
		t.Errorf("userInsights channel capacity = %d, want %d", cap(batches.userInsights), expectedCapacity)
	}
}

// ================== Helper Function Tests ==================

func TestGenerateRecordID(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	id1 := generateRecordID("pin_123", date)
	id2 := generateRecordID("pin_123", date)
	id3 := generateRecordID("pin_456", date)
	id4 := generateRecordID("pin_123", date.AddDate(0, 0, 1))

	if id1 == "" {
		t.Error("generateRecordID should return non-empty string")
	}

	if id1 != id2 {
		t.Error("same id and date should produce same record ID")
	}

	if id1 == id3 {
		t.Error("different id should produce different record ID")
	}

	if id1 == id4 {
		t.Error("different date should produce different record ID")
	}

	// MD5 produces 32 hex characters
	if len(id1) != 32 {
		t.Errorf("record ID length = %d, want 32 (MD5 hex)", len(id1))
	}
}

func TestGenerateRecordID_Format(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	id := generateRecordID("test_id", date)

	// Verify it's valid hexadecimal
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("record ID contains invalid hex character: %c", c)
		}
	}
}

// ================== Parsing Function Tests ==================

func TestParseRawUser(t *testing.T) {
	raw := &kafkamodels.RawPinterestUser{
		UserID:         "user_123",
		Username:       "testuser",
		About:          "Test bio",
		ProfileImage:   "https://example.com/image.jpg",
		WebsiteURL:     "https://example.com",
		BusinessName:   "Test Business",
		BoardCount:     10,
		PinCount:       100,
		AccountType:    "BUSINESS",
		FollowerCount:  5000,
		FollowingCount: 100,
		MonthlyViews:   10000,
		WorkspaceID:    "workspace_1",
		SavingTime:     time.Now(),
	}

	parsed := parseRawUser(raw)

	if parsed == nil {
		t.Fatal("parseRawUser returned nil")
	}
	if parsed.UserID != raw.UserID {
		t.Errorf("UserID = %q, want %q", parsed.UserID, raw.UserID)
	}
	if parsed.Username != raw.Username {
		t.Errorf("Username = %q, want %q", parsed.Username, raw.Username)
	}
	if parsed.FollowerCount != raw.FollowerCount {
		t.Errorf("FollowerCount = %d, want %d", parsed.FollowerCount, raw.FollowerCount)
	}
	if parsed.RecordID == "" {
		t.Error("RecordID should not be empty")
	}
}

func TestParseRawBoard(t *testing.T) {
	raw := &kafkamodels.RawPinterestBoard{
		BoardID:           "board_123",
		UserID:            "user_456",
		Name:              "Test Board",
		Description:       "Test description",
		Privacy:           "PUBLIC",
		PinCount:          50,
		FollowerCount:     100,
		CollaboratorCount: 2,
		CreatedAt:         time.Now(),
		Owner:             "owner_user",
		ImageCoverURL:     "https://example.com/cover.jpg",
		PinThumbnailURLs:  []string{"https://example.com/thumb1.jpg"},
		WorkspaceID:       "workspace_1",
		SavingTime:        time.Now(),
	}

	parsed := parseRawBoard(raw)

	if parsed == nil {
		t.Fatal("parseRawBoard returned nil")
	}
	if parsed.BoardID != raw.BoardID {
		t.Errorf("BoardID = %q, want %q", parsed.BoardID, raw.BoardID)
	}
	if parsed.Name != raw.Name {
		t.Errorf("Name = %q, want %q", parsed.Name, raw.Name)
	}
	if parsed.PinCount != "50" {
		t.Errorf("PinCount = %q, want %q", parsed.PinCount, "50")
	}
	if parsed.RecordID == "" {
		t.Error("RecordID should not be empty")
	}
}

func TestParseRawPin(t *testing.T) {
	raw := &kafkamodels.RawPinterestPin{
		PinID:         "pin_123",
		UserID:        "user_456",
		BoardID:       "board_789",
		Title:         "Test Pin",
		Description:   "Test description",
		Link:          "https://example.com",
		DominantColor: "#FF0000",
		CreativeType:  "IMAGE",
		MediaType:     "image",
		CoverImageURL: "https://example.com/image.jpg",
		IsStandard:    true,
		IsOwner:       true,
		CreatedAt:     time.Now(),
		DayOfWeek:     "Monday",
		HourOfDay:     10,
		WorkspaceID:   "workspace_1",
		SavingTime:    time.Now(),
	}

	parsed := parseRawPin(raw)

	if parsed == nil {
		t.Fatal("parseRawPin returned nil")
	}
	if parsed.PinID != raw.PinID {
		t.Errorf("PinID = %q, want %q", parsed.PinID, raw.PinID)
	}
	if parsed.BoardID != raw.BoardID {
		t.Errorf("BoardID = %q, want %q", parsed.BoardID, raw.BoardID)
	}
	if parsed.Title != raw.Title {
		t.Errorf("Title = %q, want %q", parsed.Title, raw.Title)
	}
	if parsed.RecordID == "" {
		t.Error("RecordID should not be empty")
	}
}

func TestParseRawPinInsight(t *testing.T) {
	raw := &kafkamodels.RawPinterestPinInsight{
		PinID:          "pin_123",
		UserID:         "user_456",
		BoardID:        "board_789",
		Date:           time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		DataStatus:     "READY",
		Impression:     1000,
		PinClicks:      100,
		OutboundClicks: 50,
		Saves:          25,
		Engagement:     175,
		WorkspaceID:    "workspace_1",
		SavingTime:     time.Now(),
	}

	parsed := parseRawPinInsight(raw)

	if parsed == nil {
		t.Fatal("parseRawPinInsight returned nil")
	}
	if parsed.PinID != raw.PinID {
		t.Errorf("PinID = %q, want %q", parsed.PinID, raw.PinID)
	}
	if parsed.Impression != raw.Impression {
		t.Errorf("Impression = %d, want %d", parsed.Impression, raw.Impression)
	}
	if parsed.RecordID == "" {
		t.Error("RecordID should not be empty")
	}
}

func TestParseRawUserInsight(t *testing.T) {
	raw := &kafkamodels.RawPinterestUserInsight{
		UserID:         "user_123",
		Date:           time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		DataStatus:     "READY",
		Impression:     10000,
		PinClicks:      1000,
		OutboundClicks: 500,
		Saves:          250,
		Engagement:     1750,
		WorkspaceID:    "workspace_1",
		SavingTime:     time.Now(),
	}

	parsed := parseRawUserInsight(raw)

	if parsed == nil {
		t.Fatal("parseRawUserInsight returned nil")
	}
	if parsed.UserID != raw.UserID {
		t.Errorf("UserID = %q, want %q", parsed.UserID, raw.UserID)
	}
	if parsed.Impression != raw.Impression {
		t.Errorf("Impression = %d, want %d", parsed.Impression, raw.Impression)
	}
	if parsed.RecordID == "" {
		t.Error("RecordID should not be empty")
	}
}
