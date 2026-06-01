# YouTube Analytics Platform Documentation

## Overview

YouTube integration uses two APIs: YouTube Data API v3 for channel/video metadata and YouTube Analytics API v2 for detailed analytics. It supports channel data, video statistics, activity insights, traffic sources, and sharing metrics.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Scheduler     │────▶│  work-order-     │────▶│   Fetcher       │
│  (Account       │     │  youtube         │     │  (Data API +    │
│   Fetcher)      │     │  (Batch)         │     │   Analytics API)│
└─────────────────┘     └──────────────────┘     └────────┬────────┘
                                                          │
                        ┌─────────────────────────────────┼─────────────────────────────────┐
                        ▼                                 ▼                                 ▼
              ┌──────────────────┐             ┌──────────────────┐             ┌──────────────────┐
              │ raw-youtube-     │             │ raw-youtube-     │             │ raw-youtube-     │
              │ channels/videos  │             │ activity-insights│             │ traffic-insights │
              └────────┬─────────┘             └────────┬─────────┘             └────────┬─────────┘
                       │                                │                                │
                       └────────────────────────────────┼────────────────────────────────┘
                                                        ▼
                                              ┌─────────────────┐
                                              │  Analytics      │
                                              │  Sink           │
                                              │  (Deduplication)│
                                              └─────────────────┘
```

## Services (3 Microservices)

| Service | Purpose | Consumer Group |
|---------|---------|----------------|
| `youtube-fetcher` | Work order consumption, parallel API calls | `youtube-fetcher-group` |
| `youtube-immediate-processor` | Real-time processing | `youtube-immediate-processor-group` |
| `youtube-analytics-sink` | Batch storage with deduplication | `youtube-analytics-sink-group` |

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-youtube` | Batch work orders |
| `immediate-work-order-youtube` | Real-time requests |
| `raw-youtube-channels` | Raw channel data |
| `raw-youtube-videos` | Raw video metadata |
| `raw-youtube-activity-insights` | Daily activity metrics |
| `raw-youtube-traffic-insights` | Traffic source breakdowns |
| `raw-youtube-shared-insights` | Sharing service metrics |

## ClickHouse Tables

### `youtube_channels`
Channel snapshots (daily deduplication via ReplacingMergeTree):
- Channel metadata: title, description, custom URL
- Statistics: subscribers, video count, view count
- Branding: thumbnail, banner URL
- Timestamps: published_at, created_at, inserted_at

### `youtube_videos`
Video analytics:
- Metadata: title, description, duration, thumbnail
- Lifetime stats: views, likes, dislikes, comments, shares
- Analytics: watch time, avg view duration, avg view percentage
- Impressions and CTR
- Media type: VIDEO or SHORT
- Embed HTML

### `youtube_activity_insights`
Daily activity metrics (deduplicated by channel_id + date):
- Views (regular and Red/Premium)
- Engagement: likes, dislikes, comments, shares
- Subscribers gained
- Watch time: estimated minutes watched
- Average view duration and percentage

### `youtube_traffic_insights`
Traffic source breakdowns (deduplicated by channel_id + date):
- Search views (YouTube search)
- Browse views (homepage, subscriptions)
- External URL views
- Suggested video views
- Playlist views
- Notification views
- Shorts views
- Channel page views
- End screen views
- Cards views

### `youtube_shared_insights`
Sharing service metrics:
- Social networks: Facebook, Twitter, LinkedIn, Reddit, Pinterest, etc.
- Messaging: WhatsApp, Telegram, Facebook Messenger, etc.
- Other: Email, Copy/Paste, Embed, etc.

## API Client Methods

```go
// Token management
RefreshToken(refreshToken) (*TokenResponse, error)

// Data API v3
FetchChannels(accessToken) (*ChannelListResponse, error)
FetchVideos(accessToken, since) ([]ActivityItem, error)
FetchVideoDetails(accessToken, videoIDs) ([]VideoItem, error)

// Analytics API v2
FetchActivityInsights(accessToken, startDate, endDate) (*AnalyticsResponse, error)
FetchTrafficInsights(accessToken, startDate, endDate) (*AnalyticsResponse, error)
FetchSharedInsights(accessToken, startDate, endDate) (*AnalyticsResponse, error)
FetchVideoInsights(accessToken, videoID, startDate, endDate) (*AnalyticsResponse, error)
FetchAllVideosAnalytics(accessToken, startDate, endDate) (*AnalyticsResponse, error)

// Media type detection
DetectMediaTypes(videos []VideoItem) map[string]string
IsYouTubeShort(videoID string) bool
```

## Rate Limiting

| Limit Type | Rate | Burst |
|------------|------|-------|
| Per-client | 5 RPS | 10 |
| Backoff Base | 500ms | - |
| Backoff Max | 8 seconds | - |

## Custom Parsing Logic

### YouTube Shorts Detection
Shorts are videos ≤ 60 seconds:
```go
func (c *YouTubeClient) DetectMediaTypes(videos []VideoItem) map[string]string {
    results := make(map[string]string)
    for _, v := range videos {
        duration := parseDuration(v.ContentDetails.Duration)
        if duration <= 60 {
            results[v.ID] = "SHORT"
        } else {
            results[v.ID] = "VIDEO"
        }
    }
    return results
}
```

### ISO 8601 Duration Parsing
YouTube returns durations in ISO 8601 format:
```go
// Input: "PT1H2M3S" (1 hour, 2 minutes, 3 seconds)
// Output: 3723 seconds

func parseDuration(iso string) int64 {
    // Parse PT[hours]H[minutes]M[seconds]S format
}
```

### Daily Snapshot Deduplication
Record IDs are MD5 hashes of channel_id + date:
```go
func GenerateYouTubeRecordID(channelID string, date time.Time) string {
    dateStr := date.Format("2006-01-02")
    return GenerateRecordID(channelID + "_" + dateStr)
}
```

### Activity Insights Aggregation
Multiple rows for same date are aggregated:
```go
// Deduplication via dayData map
dayData := make(map[string]*YouTubeActivityInsights)
for _, row := range response.Rows {
    dateStr := row["day"].(string)
    if dayData[dateStr] == nil {
        dayData[dateStr] = &YouTubeActivityInsights{...}
    }
    // Accumulate metrics
    dayData[dateStr].Views += row["views"]
    dayData[dateStr].Likes += row["likes"]
    // ...
}
```

### Traffic Source Mapping
```go
const (
    TrafficSourcePaid         = "ADVERTISING"
    TrafficSourceAnnotation   = "ANNOTATION"
    TrafficSourceEndScreen    = "END_SCREEN"
    TrafficSourceCampaignCard = "CAMPAIGN_CARD"
    TrafficSourceSubscriber   = "SUBSCRIBER"
    TrafficSourceYTChannel    = "YT_CHANNEL"
    TrafficSourceYTSearch     = "YT_SEARCH"
    TrafficSourceRelatedVideo = "RELATED_VIDEO"
    TrafficSourceExtURL       = "EXT_URL"
    TrafficSourcePlaylist     = "PLAYLIST"
    TrafficSourceNotification = "NOTIFICATION"
    TrafficSourceShorts       = "SHORTS"
)
```

## Special Handling

### 3-Day Data Delay
YouTube Analytics API has a 2-3 day data delay:
```go
// End date is 3 days ago to ensure complete data
endDate := now.AddDate(0, 0, -3)
```

### Dual API Usage
- **Data API v3**: Channel info, video metadata, activities
- **Analytics API v2**: Views, engagement, traffic sources, sharing

### Parallel Data Fetching
All data types fetched in parallel using errgroup:
```go
eg, egCtx := errgroup.WithContext(ctx)

eg.Go(func() error { return fetchChannels(egCtx) })
eg.Go(func() error { return fetchVideos(egCtx) })
eg.Go(func() error { return fetchActivityInsights(egCtx) })
eg.Go(func() error { return fetchTrafficInsights(egCtx) })
eg.Go(func() error { return fetchSharedInsights(egCtx) })

if err := eg.Wait(); err != nil { ... }
```

### Per-Account Concurrency Control
Prevents multiple simultaneous fetches for same channel:
```go
sem := semForAccount(channelID, 1)
if err := sem.Acquire(ctx, 1); err != nil { return }
defer sem.Release(1)
```

### Red/Premium Metrics
YouTube Partner accounts have additional metrics:
- `redViews` - YouTube Premium views
- `estimatedRedMinutesWatched` - Premium watch time

### ReplacingMergeTree Deduplication
ClickHouse tables use ReplacingMergeTree for deduplication:
```sql
ENGINE = ReplacingMergeTree(inserted_at)
ORDER BY (record_id, created_at)
```

## Date Range Configuration

| Sync Type | Videos | Insights | End Date |
|-----------|--------|----------|----------|
| Incremental | 14 days | 14 days | now - 3 days |
| Immediate | 90 days | 90 days | now - 3 days |
| Full Sync | 365 days | 365 days | now - 3 days |

**Example (today = Jan 30, 2026):**
- End Date: Jan 27, 2026
- Incremental: Jan 14 → Jan 27 (14 days)
- Immediate: Oct 29, 2025 → Jan 27 (90 days)
- Full Sync: Jan 27, 2025 → Jan 27, 2026 (365 days)

## MongoDB Model

```go
type YouTubeAccount struct {
    ID                  primitive.ObjectID `bson:"_id"`
    PlatformType        string            `bson:"platform_type"` // "youtube"
    PlatformIdentifier  string            `bson:"platform_identifier"` // channel_id
    // Tokens stored as embedded document
    AccessToken         map[string]any    `bson:"access_token"`
    RefreshToken        string            `bson:"refresh_token"`
    LastAnalyticsUpdate *time.Time        `bson:"last_analytics_update"`
}
```

## Work Order Format

### Single Account
```json
{
  "id": "mongodb_object_id",
  "channel_id": "UC_channel_id",
  "access_token": "encrypted_token",
  "refresh_token": "encrypted_refresh_token",
  "workspace_id": "workspace_id",
  "sync_type": "incremental"
}
```

### Batch Work Order
```json
{
  "batch_id": "uuid-v4",
  "sync_type": "incremental",
  "accounts": [
    {
      "id": "mongodb_id",
      "channel_id": "UC_channel_1",
      "access_token": "token_1",
      "refresh_token": "refresh_1"
    }
  ],
  "created_at": "2024-01-15T00:00:00Z"
}
```

## Performance Characteristics

- **Workers**: 10 concurrent fetcher workers
- **Shorts Detection**: 15 parallel workers
- **Batch Size**: 5000 items per ClickHouse insert
- **Channel Buffer**: 50K messages
- **Idle Timeout**: 5 minutes
- **Idle Check Interval**: 30 seconds

## Graceful Shutdown

### Fetcher Shutdown
- Idle timeout: 5 minutes
- Check interval: 30 seconds
- Tracks active jobs with `atomic.AddInt64(&activeJobs, ±1)`
- Uses `processorWithTracking` wrapper for job tracking
- Updates last message time on each completed work order
- Conditions for shutdown:
  - Work channel empty (`len(workChan) == 0`)
  - No active jobs (`activeJobs == 0`)
  - Idle duration >= 5 minutes

### Analytics Sink Shutdown
- Idle timeout: 5 minutes
- Check interval: 30 seconds
- Waits for all batch channels to drain (channels, videos, activity, traffic, shared)
- Flushes pending batches before exit
- Ensures all ClickHouse INSERTs complete

## Error Handling

- **401 Unauthorized**: Stops all parallel requests, triggers token refresh
- **Rate limiting**: Exponential backoff (500ms base, 8s max)
- **Token refresh failure**: Log and skip account
- **API errors**: Logged to Sentry with context

## Deduplication Strategy

1. **Record ID**: MD5(channel_id + date) ensures unique daily records
2. **ReplacingMergeTree**: ClickHouse deduplicates on merge
3. **inserted_at Version**: Keeps latest row during merge
4. **Parser Deduplication**: Aggregates same-date rows before insert
5. **OPTIMIZE TABLE FINAL**: Forces immediate deduplication

## Key Files

| File | Purpose |
|------|---------|
| `src/clients/social/youtube.go` | API client (31KB) |
| `src/services/youtube/youtube-fetcher/main.go` | Fetcher service |
| `src/services/youtube/youtube-analytics-sink/main.go` | Sink with parsing |
| `src/services/youtube/youtube-immediate-processor/` | Immediate processing |
| `src/models/kafka/youtube.go` | Kafka message models |
| `src/models/db/clickhouse/youtube.go` | ClickHouse schemas |
| `src/deployments/clickhouse/schema/youtube_schema.sql` | DDL statements |
