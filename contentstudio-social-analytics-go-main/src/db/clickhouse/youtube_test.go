package clickhouse

import (
	"context"
	"errors"
	"testing"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

func Test_BulkInsertYouTubeChannels_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertYouTubeChannels(context.Background(), []*clickhousemodels.YouTubeChannel{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertYouTubeChannels_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		channels  []*clickhousemodels.YouTubeChannel
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty channels",
			channels:  []*clickhousemodels.YouTubeChannel{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single channel success",
			channels: []*clickhousemodels.YouTubeChannel{
				{
					RecordID:   "rec_1",
					ChannelID:  "ch_123",
					Title:      "Test Channel",
					InsertedAt: now,
					CreatedAt:  now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "multiple channels success",
			channels: []*clickhousemodels.YouTubeChannel{
				{RecordID: "rec_1", ChannelID: "ch_1", InsertedAt: now, CreatedAt: now},
				{RecordID: "rec_2", ChannelID: "ch_2", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			channels: []*clickhousemodels.YouTubeChannel{
				{RecordID: "rec_1", ChannelID: "ch_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			channels: []*clickhousemodels.YouTubeChannel{
				{RecordID: "rec_1", ChannelID: "ch_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			channels: []*clickhousemodels.YouTubeChannel{
				{RecordID: "rec_1", ChannelID: "ch_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertYouTubeChannels(context.Background(), tc.channels)
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

func Test_BulkInsertYouTubeChannels_WithAllFields(t *testing.T) {
	now := time.Now()
	channels := []*clickhousemodels.YouTubeChannel{
		{
			RecordID:        "rec_1",
			ChannelID:       "ch_123",
			Title:           "Test Channel",
			Description:     "A test YouTube channel",
			CustomURL:       "@testchannel",
			ThumbnailURL:    "https://example.com/thumb.jpg",
			ExternalBanner:  "https://example.com/banner.jpg",
			Country:         "US",
			SubscriberCount: 100000,
			VideoCount:      500,
			ViewCount:       5000000,
			PublishedAt:     now.AddDate(-5, 0, 0),
			CreatedAt:       now,
			InsertedAt:      now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertYouTubeChannels(context.Background(), channels)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertYouTubeVideos_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertYouTubeVideos(context.Background(), []*clickhousemodels.YouTubeVideo{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertYouTubeVideos_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		videos    []*clickhousemodels.YouTubeVideo
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty videos",
			videos:    []*clickhousemodels.YouTubeVideo{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single video success",
			videos: []*clickhousemodels.YouTubeVideo{
				{
					VideoID:    "vid_1",
					ChannelID:  "ch_1",
					Title:      "Test Video",
					InsertedAt: now,
					CreatedAt:  now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "multiple videos success",
			videos: []*clickhousemodels.YouTubeVideo{
				{VideoID: "vid_1", ChannelID: "ch_1", InsertedAt: now, CreatedAt: now},
				{VideoID: "vid_2", ChannelID: "ch_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			videos: []*clickhousemodels.YouTubeVideo{
				{VideoID: "vid_1", ChannelID: "ch_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			videos: []*clickhousemodels.YouTubeVideo{
				{VideoID: "vid_1", ChannelID: "ch_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			videos: []*clickhousemodels.YouTubeVideo{
				{VideoID: "vid_1", ChannelID: "ch_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertYouTubeVideos(context.Background(), tc.videos)
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

func Test_BulkInsertYouTubeVideos_WithAllFields(t *testing.T) {
	now := time.Now()
	videos := []*clickhousemodels.YouTubeVideo{
		{
			VideoID:                     "vid_123",
			ChannelID:                   "ch_123",
			Title:                       "Test Video",
			Description:                 "A test video description",
			Duration:                    "PT10M30S",
			ThumbnailURL:                "https://example.com/thumb.jpg",
			IframeEmbedHTML:             "<iframe></iframe>",
			Likes:                       1000,
			Dislikes:                    50,
			Views:                       50000,
			Comments:                    200,
			Shares:                      100,
			Favorites:                   300,
			Saved:                       150,
			SubscribersGained:           500,
			RedViews:                    1000,
			MinutesWatched:              250000,
			RedMinutesWatched:           5000,
			AvgViewDuration:             630,
			AvgViewPercentage:           75.5,
			Impressions:                 100000,
			ImpressionsClickThroughRate: 5.2,
			PublishedAt:                 now.AddDate(0, -1, 0),
			CreatedAt:                   now,
			InsertedAt:                  now,
			MediaType:                   "video",
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertYouTubeVideos(context.Background(), videos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertYouTubeActivityInsights_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertYouTubeActivityInsights(context.Background(), []*clickhousemodels.YouTubeActivityInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertYouTubeActivityInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []*clickhousemodels.YouTubeActivityInsights
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []*clickhousemodels.YouTubeActivityInsights{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []*clickhousemodels.YouTubeActivityInsights{
				{
					RecordID:   "rec_1",
					ChannelID:  "ch_1",
					Views:      1000,
					InsertedAt: now,
					CreatedAt:  now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []*clickhousemodels.YouTubeActivityInsights{
				{RecordID: "rec_1", ChannelID: "ch_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []*clickhousemodels.YouTubeActivityInsights{
				{RecordID: "rec_1", ChannelID: "ch_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []*clickhousemodels.YouTubeActivityInsights{
				{RecordID: "rec_1", ChannelID: "ch_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertYouTubeActivityInsights(context.Background(), tc.insights)
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

func Test_BulkInsertYouTubeActivityInsights_WithAllFields(t *testing.T) {
	now := time.Now()
	insights := []*clickhousemodels.YouTubeActivityInsights{
		{
			RecordID:                    "rec_1",
			ChannelID:                   "ch_123",
			RedViews:                    500,
			Views:                       10000,
			Likes:                       800,
			Dislikes:                    20,
			Comments:                    150,
			Shares:                      50,
			SubscribersGained:           200,
			EstimatedMinutesWatched:     50000,
			EstimatedRedMinutesWatched:  2000,
			AvgViewDuration:             420,
			AvgViewPercentage:           65.5,
			CreatedAt:                   now,
			InsertedAt:                  now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertYouTubeActivityInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertYouTubeTrafficInsights_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertYouTubeTrafficInsights(context.Background(), []*clickhousemodels.YouTubeTrafficInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertYouTubeTrafficInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []*clickhousemodels.YouTubeTrafficInsights
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []*clickhousemodels.YouTubeTrafficInsights{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []*clickhousemodels.YouTubeTrafficInsights{
				{
					RecordID:  "rec_1",
					ChannelID: "ch_1",
					CreatedAt: now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []*clickhousemodels.YouTubeTrafficInsights{
				{RecordID: "rec_1", ChannelID: "ch_1", CreatedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []*clickhousemodels.YouTubeTrafficInsights{
				{RecordID: "rec_1", ChannelID: "ch_1", CreatedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []*clickhousemodels.YouTubeTrafficInsights{
				{RecordID: "rec_1", ChannelID: "ch_1", CreatedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertYouTubeTrafficInsights(context.Background(), tc.insights)
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

func Test_BulkInsertYouTubeTrafficInsights_WithAllFields(t *testing.T) {
	now := time.Now()
	insights := []*clickhousemodels.YouTubeTrafficInsights{
		{
			RecordID:               "rec_1",
			ChannelID:              "ch_123",
			PaidViews:              500,
			AnnotationViews:        100,
			EndScreenViews:         200,
			CampaignCardViews:      50,
			SubscriberViews:        3000,
			NoLinkOtherViews:       1000,
			YTChannelViews:         5000,
			YTSearchViews:          8000,
			RelatedVideoViews:      6000,
			YTOtherPageViews:       2000,
			ExtURLViews:            1500,
			PlaylistViews:          700,
			NotificationViews:      300,
			SubscriberWatchTime:    15000,
			NonSubscriberWatchTime: 25000,
			CreatedAt:              now,
			ShortsViews:            4000,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertYouTubeTrafficInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertYouTubeSharedInsights_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertYouTubeSharedInsights(context.Background(), []*clickhousemodels.YouTubeSharedInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertYouTubeSharedInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []*clickhousemodels.YouTubeSharedInsights
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []*clickhousemodels.YouTubeSharedInsights{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []*clickhousemodels.YouTubeSharedInsights{
				{
					RecordID:   "rec_1",
					ChannelID:  "ch_1",
					InsertedAt: now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []*clickhousemodels.YouTubeSharedInsights{
				{RecordID: "rec_1", ChannelID: "ch_1", InsertedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []*clickhousemodels.YouTubeSharedInsights{
				{RecordID: "rec_1", ChannelID: "ch_1", InsertedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []*clickhousemodels.YouTubeSharedInsights{
				{RecordID: "rec_1", ChannelID: "ch_1", InsertedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertYouTubeSharedInsights(context.Background(), tc.insights)
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

func Test_BulkInsertYouTubeSharedInsights_WithAllFields(t *testing.T) {
	now := time.Now()
	insights := []*clickhousemodels.YouTubeSharedInsights{
		{
			RecordID:      "rec_1",
			ChannelID:     "ch_123",
			Ameba:         10,
			Blogger:       20,
			CopyPaste:     30,
			Cyworld:       5,
			Digg:          15,
			Dropbox:       25,
			Embed:         100,
			Mail:          50,
			WhatsApp:      200,
			Other:         75,
			FacebookMsgr:  150,
			FacebookPages: 80,
			Facebook:      300,
			Fotka:         5,
			VKontakte:     40,
			Discord:       60,
			GooglePlus:    10,
			Goo:           5,
			Hangouts:      20,
			LinkedIn:      90,
			Pinterest:     45,
			Myspace:       2,
			Reddit:        120,
			Skype:         30,
			Telegram:      85,
			Twitter:       250,
			Tumblr:        15,
			Viber:         40,
			Weibo:         25,
			WeChat:        35,
			YouTube:       500,
			InsertedAt:    now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertYouTubeSharedInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertYouTubeChannels_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	channels := []*clickhousemodels.YouTubeChannel{
		{RecordID: "rec_1", ChannelID: "ch_1"},
	}
	_ = client.BulkInsertYouTubeChannels(ctx, channels)
}

func Test_BulkInsertYouTubeVideos_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	videos := []*clickhousemodels.YouTubeVideo{
		{VideoID: "vid_1", ChannelID: "ch_1"},
	}
	_ = client.BulkInsertYouTubeVideos(ctx, videos)
}

func Test_BulkInsertYouTubeActivityInsights_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	insights := []*clickhousemodels.YouTubeActivityInsights{
		{RecordID: "rec_1", ChannelID: "ch_1"},
	}
	_ = client.BulkInsertYouTubeActivityInsights(ctx, insights)
}

func Test_BulkInsertYouTubeTrafficInsights_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	insights := []*clickhousemodels.YouTubeTrafficInsights{
		{RecordID: "rec_1", ChannelID: "ch_1"},
	}
	_ = client.BulkInsertYouTubeTrafficInsights(ctx, insights)
}

func Test_BulkInsertYouTubeSharedInsights_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	insights := []*clickhousemodels.YouTubeSharedInsights{
		{RecordID: "rec_1", ChannelID: "ch_1"},
	}
	_ = client.BulkInsertYouTubeSharedInsights(ctx, insights)
}

func Test_BulkInsertYouTubeVideos_MultipleItems(t *testing.T) {
	now := time.Now()
	videos := []*clickhousemodels.YouTubeVideo{
		{VideoID: "vid_1", ChannelID: "ch_1", Title: "Video 1", InsertedAt: now, CreatedAt: now},
		{VideoID: "vid_2", ChannelID: "ch_1", Title: "Video 2", InsertedAt: now, CreatedAt: now},
		{VideoID: "vid_3", ChannelID: "ch_1", Title: "Video 3", InsertedAt: now, CreatedAt: now},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertYouTubeVideos(context.Background(), videos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertYouTubeChannels_MultipleItems(t *testing.T) {
	now := time.Now()
	channels := []*clickhousemodels.YouTubeChannel{
		{RecordID: "rec_1", ChannelID: "ch_1", Title: "Channel 1", InsertedAt: now, CreatedAt: now},
		{RecordID: "rec_2", ChannelID: "ch_2", Title: "Channel 2", InsertedAt: now, CreatedAt: now},
		{RecordID: "rec_3", ChannelID: "ch_3", Title: "Channel 3", InsertedAt: now, CreatedAt: now},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertYouTubeChannels(context.Background(), channels)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
