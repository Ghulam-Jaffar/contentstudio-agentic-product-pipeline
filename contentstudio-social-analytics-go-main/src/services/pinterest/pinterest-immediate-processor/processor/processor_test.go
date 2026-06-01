package processor

import (
	"context"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
)

func TestParsePinterestDateRange(t *testing.T) {
	start, end, ok, err := parsePinterestDateRange("2025-01-10", "2025-01-20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected requested range to be detected")
	}
	wantStart, _ := time.Parse("2006-01-02", "2025-01-10")
	wantEnd, _ := time.Parse("2006-01-02", "2025-01-20")
	if !start.Equal(wantStart.UTC()) {
		t.Fatalf("unexpected start: %v", start)
	}
	if !end.Equal(wantEnd.UTC()) {
		t.Fatalf("unexpected end: %v", end)
	}

	_, _, _, err = parsePinterestDateRange("2025-01-20", "2025-01-10")
	if err == nil {
		t.Fatal("expected error for inverted range")
	}
}

func TestProcessAccount_UsesRequestedDateRange(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	startDate := "2025-01-10"
	endDate := "2025-01-20"
	wantStart, _ := time.Parse("2006-01-02", startDate)
	wantEnd, _ := time.Parse("2006-01-02", endDate)

	pinNew := social.PinterestPin{
		ID:         "pin-new",
		Title:      "New",
		CreatedAt:  "2025-01-21T00:00:00Z",
		BoardID:    "board-1",
		IsStandard: true,
	}
	pinKeep := social.PinterestPin{
		ID:         "pin-keep",
		Title:      "Keep",
		CreatedAt:  "2025-01-15T12:00:00Z",
		BoardID:    "board-1",
		IsStandard: true,
	}
	pinOld := social.PinterestPin{
		ID:         "pin-old",
		Title:      "Old",
		CreatedAt:  "2025-01-09T23:59:00Z",
		BoardID:    "board-1",
		IsStandard: true,
	}

	var gotUserAnalyticsStart time.Time
	var gotUserAnalyticsEnd time.Time
	var gotMultiPinIDs []string
	var insertedPins []string

	mockClient := &social.MockPinterestClient{
		GetUserAccountFunc: func(ctx context.Context, accessToken string) (*social.PinterestUserAccount, error) {
			return &social.PinterestUserAccount{
				ID:            "user-1",
				Username:      "test-user",
				AccountType:   "BUSINESS",
				FollowerCount: 10,
			}, nil
		},
		GetUserAccountAnalyticsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.PinterestUserAnalyticsResponse, error) {
			gotUserAnalyticsStart = startDate
			gotUserAnalyticsEnd = endDate
			return &social.PinterestUserAnalyticsResponse{}, nil
		},
		GetBoardsFunc: func(ctx context.Context, accessToken string) (*social.PinterestBoardsResponse, error) {
			return &social.PinterestBoardsResponse{
				Items: []social.PinterestBoard{
					{
						ID:        "board-1",
						Name:      "Board 1",
						Privacy:   "PUBLIC",
						CreatedAt: "2025-01-01T00:00:00Z",
						Owner:     map[string]interface{}{"username": "test-user"},
					},
				},
			}, nil
		},
		GetBoardPinsFunc: func(ctx context.Context, accessToken, boardID string, pageSize int, bookmark string) (*social.PinterestPinsResponse, error) {
			return &social.PinterestPinsResponse{
				Items: []social.PinterestPin{pinNew, pinKeep, pinOld},
			}, nil
		},
		GetMultiPinAnalyticsFunc: func(ctx context.Context, accessToken string, pinIDs []string, startDate, endDate time.Time) (map[string]*social.PinterestPinAnalyticsResponse, error) {
			gotMultiPinIDs = append([]string{}, pinIDs...)
			return map[string]*social.PinterestPinAnalyticsResponse{}, nil
		},
	}

	mockCH := &conversions.MockClickHouseClient{
		BulkInsertPinterestPinsFunc: func(ctx context.Context, pins []clickhousemodels.PinterestPin) error {
			for _, pin := range pins {
				insertedPins = append(insertedPins, pin.PinID)
			}
			return nil
		},
	}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCH)

	proc := &Processor{
		mongoRepo:       &mongodb.MockUnifiedSocialRepository{},
		sink:            sink,
		pinterestClient: mockClient,
		log:             log,
		cfg:             &config.Config{DecryptionKey: "test"},
	}

	err := proc.ProcessAccount(ctx, WorkOrder{
		AccountID:   "acct-1",
		AccessToken: "token",
		AccountType: "profile",
		WorkspaceID: "ws-1",
		SyncType:    "immediate",
		StartDate:   startDate,
		EndDate:     endDate,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !gotUserAnalyticsStart.Equal(wantStart.UTC()) {
		t.Fatalf("unexpected analytics start: %v", gotUserAnalyticsStart)
	}
	if !gotUserAnalyticsEnd.Equal(wantEnd.UTC()) {
		t.Fatalf("unexpected analytics end: %v", gotUserAnalyticsEnd)
	}
	if len(gotMultiPinIDs) != 1 || gotMultiPinIDs[0] != "pin-keep" {
		t.Fatalf("expected only in-range pin analytics, got %v", gotMultiPinIDs)
	}
	if len(insertedPins) != 1 || insertedPins[0] != "pin-keep" {
		t.Fatalf("expected only in-range pin inserted, got %v", insertedPins)
	}
}
