# TikTok Analytics Platform Documentation

## Overview

TikTok integration uses the TikTok Content Posting API with OAuth2 authentication. It supports video fetching and account-level insights but has limited per-video analytics compared to other platforms.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Scheduler     │────▶│  work-order-     │────▶│   Fetcher       │
│  (Account       │     │  tiktok          │     │  (Content API)  │
│   Fetcher)      │     └──────────────────┘     └────────┬────────┘
└─────────────────┘                                       │
                                                          ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Analytics      │◀────│  raw-tiktok-     │◀────│   (Direct to    │
│  Sink           │     │  posts/insights  │     │    Sink)        │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

## Services (2 Main Services)

| Service | Purpose | Consumer Group |
|---------|---------|----------------|
| `tiktok-fetcher` | Work order consumption and API calls | `tiktok-fetcher-group` |
| `unified-analytics-sink` | Batch storage (shared with other platforms) | `analytics-sink-group` |

**Note**: TikTok uses the unified analytics sink instead of a dedicated parser/sink.

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-tiktok` | Work order messages |
| `raw-tiktok-posts` | Raw video data from API |
| `raw-tiktok-insights` | Account-level insights |

## ClickHouse Tables

### `tiktok_posts`
Video analytics including:
- Engagement: likes, comments, shares, views
- Video metadata: duration, dimensions, title, description
- Hashtags array
- Engagement rate
- Embed HTML/link

### `tiktok_insights`
Account-level metrics:
- Follower count
- Following count
- Total like count
- Video count
- Total video views
- Verification status

## API Client Methods

```go
// Video fetching
FetchUserVideos(userID, accessToken, cursor, maxCount)
FetchVideoList(accessToken, cursor, maxCount)
FetchVideoDetails(accessToken, videoIDs)
FetchVideoPaginated(accessToken, cutoffTime, maxVideos)

// User info
FetchUserInfo(accessToken)

// Token management
RefreshToken(refreshToken)
```

## Rate Limiting

| Limit Type | Configuration |
|------------|---------------|
| HTTP Timeout | 30 seconds |
| Inter-page Delay | 100ms |
| Max Videos per Request | 20 |

## Custom Parsing Logic

### Video Metadata Extraction
```go
type ParsedTikTokPost struct {
    ID           string    `json:"id"`
    TikTokID     string    `json:"tiktok_id"`
    Title        string    `json:"title"`
    Description  string    `json:"description"`
    Duration     int64     `json:"duration"`      // seconds
    Width        int64     `json:"width"`
    Height       int64     `json:"height"`
    ViewCount    int64     `json:"view_count"`
    LikeCount    int64     `json:"like_count"`
    CommentCount int64     `json:"comment_count"`
    ShareCount   int64     `json:"share_count"`
    CreateTime   time.Time `json:"create_time"`
    Hashtags     []string  `json:"hashtags"`
}
```

### Hashtag Extraction
Hashtags are parsed from description:
```go
// Input: "Check out this #viral #trend video!"
// Output: ["viral", "trend"]
hashtagRegex := regexp.MustCompile(`#(\w+)`)
```

### Engagement Rate Calculation
```go
engagement = likes + comments + shares
engagement_rate = (engagement / view_count) * 100
```

### Unix Timestamp Conversion
TikTok returns Unix timestamps:
```go
createTime := time.Unix(video.CreateTime, 0).UTC()
```

## Special Handling

### OAuth2 Token Refresh
TikTok tokens expire and need refresh:
```go
func (c *TikTokClient) RefreshToken(refreshToken string) (string, string, error) {
    // POST to oauth/token with refresh_token grant
    // Returns new access_token and refresh_token
}
```

### Scope Validation
TikTok requires specific scopes for analytics:
```go
requiredScopes := []string{
    "user.info.basic",
    "video.list",
}
```

### Cursor-Based Pagination
```go
type VideoListResponse struct {
    Data struct {
        Videos  []Video `json:"videos"`
        Cursor  int64   `json:"cursor"`
        HasMore bool    `json:"has_more"`
    } `json:"data"`
}
```

### No Historical Per-Video Analytics
Unlike other platforms, TikTok doesn't provide historical analytics per video. Only current snapshot data is available.

## Date Range Configuration

| Sync Type | Videos |
|-----------|--------|
| Incremental | 100 videos (no date filter) |
| Full Sync | All videos (no limit) |

**Note**: TikTok API doesn't support date-based filtering. Uses cutoff time at application level.

## MongoDB Model

```go
type TikTokAccount struct {
    ID                  primitive.ObjectID `bson:"_id"`
    PlatformType        string            `bson:"platform_type"` // "tiktok"
    PlatformIdentifier  string            `bson:"platform_identifier"` // tiktok_id
    AccessToken         string            `bson:"access_token"`
    RefreshToken        string            `bson:"refresh_token"`
    Scope               string            `bson:"scope"`
    LastAnalyticsUpdate *time.Time        `bson:"last_analytics_update"`
}
```

## Work Order Format

```json
{
  "id": "mongodb_object_id",
  "tiktok_id": "user_id",
  "access_token": "encrypted_token",
  "refresh_token": "encrypted_refresh_token",
  "scope": "user.info.basic,video.list",
  "workspace_id": "workspace_id",
  "sync_type": "incremental"
}
```

## Performance Characteristics

- **Max Videos per Request**: 20
- **Workers**: 10 concurrent processors
- **Pagination Delay**: 100ms between pages
- **Channel Buffer**: 200 messages

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
- Waits for all batch channels to drain
- Flushes pending batches before exit
- Ensures all ClickHouse INSERTs complete

## Error Handling

- Token expiration: Automatic refresh via refresh_token
- Invalid scope: Log and skip account
- Rate limiting: 100ms delay between requests
- API errors: Logged with context

## Limitations

1. **No Per-Video Historical Analytics**: Only current snapshot data
2. **No Date-Based Filtering**: Must use cutoff time at application level
3. **Limited Metrics**: Fewer metrics compared to Facebook/Instagram
4. **API Rate Limits**: Strict limits on API calls

## Insights Generation

Since TikTok doesn't provide account insights API, insights are generated from video data:
```go
func GenerateInsights(userInfo *UserInfo, tiktokID string,
    totalViews, totalLikes, totalComments, totalShares int64) *ParsedTikTokInsights {
    return &ParsedTikTokInsights{
        TikTokID:      tiktokID,
        FollowerCount: userInfo.FollowerCount,
        FollowingCount: userInfo.FollowingCount,
        LikeCount:     userInfo.LikeCount,
        VideoCount:    userInfo.VideoCount,
        TotalViews:    totalViews,
        TotalLikes:    totalLikes,
        TotalComments: totalComments,
        TotalShares:   totalShares,
    }
}
```

## Key Files

| File | Purpose |
|------|---------|
| `src/clients/social/tiktok.go` | API client (7KB) |
| `src/utils/parsing/tiktok_parser.go` | Data transformation |
| `src/services/tiktok/tiktok-fetcher/main.go` | Fetcher service |
| `src/models/kafka/tiktok.go` | Kafka message models |
| `src/models/db/clickhouse/tiktok.go` | ClickHouse schemas |
