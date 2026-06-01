package clickhouse

import (
	"context"
	"errors"
	"testing"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

func Test_BulkInsertPinterestUsers_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertPinterestUsers(context.Background(), []clickhousemodels.PinterestUser{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertPinterestUsers_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		users     []clickhousemodels.PinterestUser
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty users",
			users:     []clickhousemodels.PinterestUser{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single user success",
			users: []clickhousemodels.PinterestUser{
				{
					UserID:     "user_1",
					Username:   "testuser",
					InsertedAt: now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "multiple users success",
			users: []clickhousemodels.PinterestUser{
				{UserID: "user_1", Username: "user1", InsertedAt: now},
				{UserID: "user_2", Username: "user2", InsertedAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			users: []clickhousemodels.PinterestUser{
				{UserID: "user_1", InsertedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			users: []clickhousemodels.PinterestUser{
				{UserID: "user_1", InsertedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			users: []clickhousemodels.PinterestUser{
				{UserID: "user_1", InsertedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertPinterestUsers(context.Background(), tc.users)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertPinterestUsers_WithAllFields(t *testing.T) {
	now := time.Now()
	users := []clickhousemodels.PinterestUser{
		{
			UserID:        "user_123",
			ProfileImage:  "https://example.com/profile.jpg",
			WebsiteURL:    "https://example.com",
			Username:      "testuser",
			About:         "Test user bio",
			BusinessName:  "Test Business",
			BoardCount:    10,
			PinCount:      500,
			AccountType:   "business",
			FollowerCount: 5000,
			FollowingCount: 200,
			MonthlyViews:  100000,
			InsertedAt:    now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertPinterestUsers(context.Background(), users)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertPinterestBoards_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertPinterestBoards(context.Background(), []clickhousemodels.PinterestBoard{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertPinterestBoards_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		boards    []clickhousemodels.PinterestBoard
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty boards",
			boards:    []clickhousemodels.PinterestBoard{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single board success",
			boards: []clickhousemodels.PinterestBoard{
				{
					BoardID:    "board_1",
					UserID:     "user_1",
					Name:       "Test Board",
					InsertedAt: now,
					CreatedAt:  now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			boards: []clickhousemodels.PinterestBoard{
				{BoardID: "board_1", UserID: "user_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			boards: []clickhousemodels.PinterestBoard{
				{BoardID: "board_1", UserID: "user_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			boards: []clickhousemodels.PinterestBoard{
				{BoardID: "board_1", UserID: "user_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertPinterestBoards(context.Background(), tc.boards)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertPinterestBoards_WithAllFields(t *testing.T) {
	now := time.Now()
	boards := []clickhousemodels.PinterestBoard{
		{
			RecordID:         "rec_1",
			BoardID:          "board_123",
			UserID:           "user_123",
			Name:             "Test Board",
			Owner:            "testuser",
			Description:      "A test board",
			Privacy:          "public",
			ImageCoverURL:    "https://example.com/cover.jpg",
			PinThumbnailURLs: []string{"https://example.com/thumb1.jpg", "https://example.com/thumb2.jpg"},
			CollaboratorCount: "5",
			PinCount:          "100",
			FollowerCount:     "500",
			CreatedAt:         now.AddDate(-1, 0, 0),
			InsertedAt:        now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertPinterestBoards(context.Background(), boards)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertPinterestPins_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertPinterestPins(context.Background(), []clickhousemodels.PinterestPin{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertPinterestPins_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		pins      []clickhousemodels.PinterestPin
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty pins",
			pins:      []clickhousemodels.PinterestPin{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single pin success",
			pins: []clickhousemodels.PinterestPin{
				{
					PinID:      "pin_1",
					UserID:     "user_1",
					BoardID:    "board_1",
					InsertedAt: now,
					CreatedAt:  now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			pins: []clickhousemodels.PinterestPin{
				{PinID: "pin_1", UserID: "user_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			pins: []clickhousemodels.PinterestPin{
				{PinID: "pin_1", UserID: "user_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			pins: []clickhousemodels.PinterestPin{
				{PinID: "pin_1", UserID: "user_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertPinterestPins(context.Background(), tc.pins)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertPinterestPins_WithAllFields(t *testing.T) {
	now := time.Now()
	pins := []clickhousemodels.PinterestPin{
		{
			PinID:           "pin_123",
			UserID:          "user_123",
			BoardID:         "board_123",
			Title:           "Test Pin",
			Note:            "A test pin note",
			ParentPinID:     "parent_pin_1",
			BoardSectionID:  "section_1",
			Description:     "Test pin description",
			BoardOwner:      "owner_user",
			MediaType:       "image",
			CoverImageURL:   "https://example.com/pin.jpg",
			VideoURL:        "",
			Duration:        "0",
			Height:          "600",
			Width:           "400",
			DominantColor:   "#FF5733",
			ProductTags:     []string{"tag1", "tag2"},
			CreativeType:    "regular",
			IsStandard:      "1",
			IsOwner:         "1",
			HasBeenPromoted: "0",
			HourOfDay:       14,
			DayOfWeek:       "3",
			CreatedAt:       now.AddDate(0, -1, 0),
			InsertedAt:      now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertPinterestPins(context.Background(), pins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertPinterestPinInsights_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertPinterestPinInsights(context.Background(), []clickhousemodels.PinterestPinInsight{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertPinterestPinInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []clickhousemodels.PinterestPinInsight
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []clickhousemodels.PinterestPinInsight{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []clickhousemodels.PinterestPinInsight{
				{
					RecordID:   "rec_1",
					PinID:      "pin_1",
					UserID:     "user_1",
					Date:       now,
					InsertedAt: now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []clickhousemodels.PinterestPinInsight{
				{RecordID: "rec_1", PinID: "pin_1", Date: now, InsertedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []clickhousemodels.PinterestPinInsight{
				{RecordID: "rec_1", PinID: "pin_1", Date: now, InsertedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []clickhousemodels.PinterestPinInsight{
				{RecordID: "rec_1", PinID: "pin_1", Date: now, InsertedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertPinterestPinInsights(context.Background(), tc.insights)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertPinterestPinInsights_WithAllFields(t *testing.T) {
	now := time.Now()
	insights := []clickhousemodels.PinterestPinInsight{
		{
			RecordID:            "rec_1",
			PinID:               "pin_123",
			UserID:              "user_123",
			BoardID:             "board_123",
			PinClicks:           500,
			VideoMRCView:        200,
			FullScreenPlay:      150,
			OutboundClicks:      100,
			VideoV50WatchTime:   3000,
			Clickthrough:        300,
			ClickthroughRate:    2.5,
			Engagement:          800,
			EngagementRate:      5.0,
			VideoStart:          250,
			ProfileVisit:        50,
			Closeup:             400,
			FullScreenPlaytime:  6000,
			VideoAvgWatchTime:   15,
			Video10sView:        180,
			Quartile95sPercent:  120,
			UserFollow:          25,
			Impression:          10000,
			Saves:               350,
			SaveRate:            3.5,
			DataStatus:          "READY",
			DayOfWeek:           "3",
			HourOfDay:           14,
			Date:                now,
			InsertedAt:          now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertPinterestPinInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertPinterestUserInsights_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertPinterestUserInsights(context.Background(), []clickhousemodels.PinterestUserInsight{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertPinterestUserInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []clickhousemodels.PinterestUserInsight
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []clickhousemodels.PinterestUserInsight{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []clickhousemodels.PinterestUserInsight{
				{
					RecordID:   "rec_1",
					UserID:     "user_1",
					Date:       now,
					InsertedAt: now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []clickhousemodels.PinterestUserInsight{
				{RecordID: "rec_1", UserID: "user_1", Date: now, InsertedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []clickhousemodels.PinterestUserInsight{
				{RecordID: "rec_1", UserID: "user_1", Date: now, InsertedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []clickhousemodels.PinterestUserInsight{
				{RecordID: "rec_1", UserID: "user_1", Date: now, InsertedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertPinterestUserInsights(context.Background(), tc.insights)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertPinterestUserInsights_WithAllFields(t *testing.T) {
	now := time.Now()
	insights := []clickhousemodels.PinterestUserInsight{
		{
			RecordID:            "rec_1",
			UserID:              "user_123",
			PinClicks:           1000,
			PinClickRate:        3.5,
			VideoMRCView:        500,
			FullScreenPlay:      300,
			OutboundClicks:      200,
			VideoV50WatchTime:   5000,
			Clickthrough:        600,
			ClickthroughRate:    4.0,
			Engagement:          1500,
			EngagementRate:      8.0,
			VideoStart:          400,
			ProfileVisit:        100,
			Closeup:             700,
			FullScreenPlaytime:  10000,
			VideoAvgWatchTime:   20,
			Video10sView:        350,
			Quartile95sPercent:  250,
			Impression:          20000,
			Saves:               800,
			SaveRate:            4.0,
			DataStatus:          "READY",
			Date:                now,
			InsertedAt:          now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertPinterestUserInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertPinterestUsers_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	users := []clickhousemodels.PinterestUser{{UserID: "user_1"}}
	_ = client.BulkInsertPinterestUsers(ctx, users)
}

func Test_BulkInsertPinterestPins_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	pins := []clickhousemodels.PinterestPin{{PinID: "pin_1", UserID: "user_1"}}
	_ = client.BulkInsertPinterestPins(ctx, pins)
}

func Test_BulkInsertPinterestPinInsights_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	insights := []clickhousemodels.PinterestPinInsight{{RecordID: "rec_1", PinID: "pin_1"}}
	_ = client.BulkInsertPinterestPinInsights(ctx, insights)
}

func Test_BulkInsertPinterestUsers_MultipleItems(t *testing.T) {
	now := time.Now()
	users := []clickhousemodels.PinterestUser{
		{UserID: "user_1", Username: "user1", InsertedAt: now},
		{UserID: "user_2", Username: "user2", InsertedAt: now},
		{UserID: "user_3", Username: "user3", InsertedAt: now},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertPinterestUsers(context.Background(), users)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertPinterestPins_MultipleItems(t *testing.T) {
	now := time.Now()
	pins := []clickhousemodels.PinterestPin{
		{PinID: "pin_1", UserID: "user_1", BoardID: "board_1", InsertedAt: now, CreatedAt: now},
		{PinID: "pin_2", UserID: "user_1", BoardID: "board_1", InsertedAt: now, CreatedAt: now},
		{PinID: "pin_3", UserID: "user_1", BoardID: "board_2", InsertedAt: now, CreatedAt: now},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertPinterestPins(context.Background(), pins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
