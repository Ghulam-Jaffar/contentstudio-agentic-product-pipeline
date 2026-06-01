# Facebook Analytics Platform Documentation

## Overview

Facebook is a production-ready platform with the most comprehensive implementation. It uses Facebook Graph API v20.0 with HMAC-SHA256 appsecret_proof for security.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Scheduler     │────▶│  work-order-     │────▶│   Fetcher       │
│  (Account       │     │  facebook        │     │  (Graph API)    │
│   Fetcher)      │     └──────────────────┘     └────────┬────────┘
└─────────────────┘                                       │
                                                          ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  ClickHouse     │◀────│  parsed-facebook │◀────│   Parser        │
│  Sink           │     │  -posts/-videos  │     │  (60+ metrics)  │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

## Services (6 Microservices)

| Service | Purpose | Consumer Group |
|---------|---------|----------------|
| `facebook-fetcher` | Consumes work orders, calls Graph API | `facebook-fetcher-group` |
| `facebook-posts-parser` | Parses raw posts to analytics-ready format | `facebook-posts-parser-group` |
| `facebook-immediate-processor` | Real-time processing pipeline | `facebook-immediate-processor-group` |
| `facebook-clickhouse-sink` | Batch inserts to ClickHouse | `facebook-clickhouse-sink-group` |
| `facebook-analytics-sink` | Aggregation/analytics processing | `facebook-analytics-sink-group` |
| `facebook-competitor-analysis` | Competitor tracking | `facebook-competitor-group` |

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-facebook` | Batch work order messages |
| `immediate-work-order-facebook` | Real-time processing requests |
| `raw-facebook-posts` | Raw post data from API |
| `raw-facebook-videos` | Raw video data from API |
| `raw-facebook-insights` | Account-level insights |
| `parsed-facebook-posts` | Parsed/normalized posts |
| `parsed-facebook-videos` | Parsed/normalized videos |
| `parsed-facebook-insights` | Processed insights |

## ClickHouse Tables

### `facebook_posts`
Individual post analytics with 60+ metrics including:
- Engagement: likes, loves, haha, wow, sad, angry, thankful, shares, comments
- Impressions: unique, viral, organic, paid
- Video metrics: views, 3s views, 10s views, complete views
- Demographics: reach by age, gender, location

### `facebook_insights`
Daily account-level insights:
- Follower growth
- Page views and impressions
- Audience demographics
- Sentiment analysis

### `facebook_competitors_posts` / `facebook_competitors_insights`
Competitor tracking tables with public metrics.

## API Client Methods

```go
// Post fetching
FetchPosts(pageID, accessToken)
FetchPostsWithLimit(pageID, accessToken, maxPages)
FetchPostsSince(pageID, accessToken, since, until)

// Video fetching (60+ metrics)
FetchVideos(pageID, accessToken, maxPages)
FetchVideosSince(pageID, accessToken, since, until)

// Account insights
FetchInsights(pageID, accessToken, since, until)

// Competitor analysis
GetCompetitorPageDetails(pageID, accessToken)
GetCompetitorPosts(pageID, accessToken, since, until, limit)
```

## Rate Limiting

| Limit Type | Rate | Burst |
|------------|------|-------|
| Per-token | 4 RPS | 4 |
| Global | 12 RPS | 12 |

## Custom Parsing Logic

### Timestamp Conversion
Facebook uses `+0000` timezone format which is converted to `+00:00` for Go parsing.

```go
// Input:  "2024-01-15T10:30:00+0000"
// Output: "2024-01-15T10:30:00+00:00"
```

### Reaction Aggregation
Reactions are aggregated by type:
- LIKE, LOVE, HAHA, WOW, SAD, ANGRY, THANKFUL (added in v20.0)

### Media Extraction
Multi-level attachment structure parsing:
- Full picture URL extraction
- Video thumbnail generation
- Carousel/album handling

### Engagement Calculation
```go
engagement = reactions + comments + shares
engagement_rate = (engagement / fan_count) * 100
```

## Special Handling

### HMAC-SHA256 Security
All API requests include `appsecret_proof`:
```go
proof := hmac.New(sha256.New, []byte(appSecret))
proof.Write([]byte(accessToken))
appsecret_proof := hex.EncodeToString(proof.Sum(nil))
```

### Post Types
- Regular posts (text, link, photo)
- Videos (with 60+ video-specific metrics)
- Live videos
- Reels
- Stories

### Admin Creator Detection
Tracks which admin created each post for multi-admin pages.

### Parent Post Tracking
Shared posts include reference to original post via `parent_id`.

## Date Range Configuration

| Sync Type | Posts | Videos | Insights |
|-----------|-------|--------|----------|
| Incremental | 7 days | 7 days | 7 days |
| Immediate | 30 days | 30 days | 30 days |
| Full Sync | 90 days | 90 days | 90 days |

## MongoDB Model

```go
type FacebookAccount struct {
    ID                  primitive.ObjectID `bson:"_id"`
    PlatformType        string            `bson:"platform_type"` // "facebook"
    PlatformIdentifier  string            `bson:"platform_identifier"` // facebook_id
    AccessToken         string            `bson:"access_token"`
    LongAccessToken     string            `bson:"long_access_token"`
    FanCount           int64              `bson:"fan_count"`
    PostedAs           string             `bson:"posted_as"`
    LastAnalyticsUpdate *time.Time        `bson:"last_analytics_update"`
}
```

## Work Order Format

```json
{
  "id": "mongodb_object_id",
  "facebook_id": "page_id",
  "type": "Page",
  "access_token": "encrypted_token",
  "workspace_id": "workspace_id",
  "sync_type": "incremental",
  "fan_count": 50000
}
```

## Performance Characteristics

- **Throughput**: ~500 posts + ~100 videos per work order
- **Batch Size**: 1000 items per ClickHouse insert
- **Workers**: 15 concurrent processors
- **Channel Buffer**: 50K messages

## Graceful Shutdown

### Fetcher Shutdown
- Idle timeout: 5 minutes
- Check interval: 30 seconds
- Tracks active jobs with `atomic.AddInt64(&activeJobs, ±1)`
- Updates last message time on each completed work order
- Conditions for shutdown:
  - Work channel empty (`len(workChan) == 0`)
  - No active jobs (`activeJobs == 0`)
  - Idle duration >= 5 minutes

### ClickHouse Sink Shutdown
- Idle timeout: 5 minutes
- Check interval: 30 seconds
- Waits for all batch channels to drain
- Flushes pending batches before exit
- Ensures all ClickHouse INSERTs complete

## Error Handling

- Token expiration: Automatic refresh via long-lived token
- Rate limiting: Exponential backoff with jitter
- API errors: Logged to Sentry with context
- Invalid data: Skipped with warning log

## Key Files

| File | Purpose |
|------|---------|
| `src/clients/social/facebook.go` | API client (59KB) |
| `src/utils/parsing/facebook_parser.go` | Data transformation |
| `src/services/facebook/facebook-fetcher/main.go` | Fetcher service |
| `src/models/kafka/facebook.go` | Kafka message models |
| `src/models/db/clickhouse/facebook.go` | ClickHouse schemas |
