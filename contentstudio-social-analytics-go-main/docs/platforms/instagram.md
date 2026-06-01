# Instagram Analytics Platform Documentation

## Overview

Instagram is a production-ready platform supporting both Business and Creator accounts. It uses the Instagram Graph API via Facebook's platform with support for media, stories, reels, and account demographics.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Scheduler     │────▶│  work-order-     │────▶│   Fetcher       │
│  (Account       │     │  instagram       │     │  (Graph API)    │
│   Fetcher)      │     └──────────────────┘     └────────┬────────┘
└─────────────────┘                                       │
                                                          ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  ClickHouse     │◀────│  parsed-instagram│◀────│   Parser        │
│  Sink           │     │  -media/-insights│     │  (Media types)  │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

## Services (6 Microservices)

| Service | Purpose | Consumer Group |
|---------|---------|----------------|
| `instagram-fetcher` | Media and insights fetching | `instagram-fetcher-group` |
| `instagram-posts-parser` | Media data parsing | `instagram-posts-parser-group` |
| `instagram-immediate-processor` | Real-time processing | `instagram-immediate-processor-group` |
| `instagram-clickhouse-sink` | Batch storage | `instagram-clickhouse-sink-group` |
| `instagram-analytics-sink` | Analytics aggregation | `instagram-analytics-sink-group` |
| `instagram-competitor-analysis` | Competitor tracking | `instagram-competitor-group` |

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-instagram` | Batch work order messages |
| `immediate-work-order-instagram` | Real-time requests |
| `raw-instagram-media` | Raw media from API |
| `raw-instagram-insights` | Account and media insights |
| `raw-instagram-stories` | Story data |
| `parsed-instagram-media` | Normalized media data |
| `parsed-instagram-insights` | Processed insights |

## ClickHouse Tables

### `instagram_posts`
Media analytics including:
- Engagement: likes, comments, saves, shares
- Reach and impressions
- Video views and plays
- Hashtags array
- Media URLs (thumbnail, full)
- Carousel children

### `instagram_insights`
Daily account demographics:
- Follower count and growth
- Engagement rate
- Audience breakdown by:
  - Age groups
  - Gender
  - Top locations (cities, countries)
- Profile views
- Website clicks

### `instagram_competitors_posts` / `instagram_competitors_insights`
Competitor tracking with public metrics.

## API Client Methods

```go
// Media fetching
FetchMedia(instagramID, accessToken)
FetchMediaSince(instagramID, accessToken, since)
FetchMediaWithLimit(instagramID, accessToken, maxPages)
FetchAllMedia(instagramID, accessToken)

// Stories
FetchStories(instagramID, accessToken)

// Account insights
FetchInsights(instagramID, accessToken, since, until)
FetchInsightsDaily(instagramID, accessToken, days, concurrency)
FetchAccountDemographics(instagramID, accessToken)

// Per-media insights
FetchMediaInsights(mediaID, accessToken, mediaType, mediaProductType)
```

## Rate Limiting

| Limit Type | Rate | Burst |
|------------|------|-------|
| Per-token | 4 RPS | 4 |
| Global | 12 RPS | 12 |

## Custom Parsing Logic

### Media Type Detection
```go
const (
    MediaTypeImage    = "IMAGE"
    MediaTypeVideo    = "VIDEO"
    MediaTypeCarousel = "CAROUSEL_ALBUM"
    MediaTypeReel     = "REEL"
)
```

### Carousel/Child Assets
Carousels contain multiple media items:
```go
type CarouselChild struct {
    ID        string `json:"id"`
    MediaType string `json:"media_type"`
    MediaURL  string `json:"media_url"`
    Timestamp string `json:"timestamp"`
}
```

### Insights Breakdown Parsing
Demographics come with breakdown results:
```go
// Age breakdown
{"name": "18-24", "values": [{"value": 1500}]}
{"name": "25-34", "values": [{"value": 3000}]}

// Gender breakdown
{"name": "M", "values": [{"value": 4000}]}
{"name": "F", "values": [{"value": 5500}]}

// Location breakdown
{"name": "New York, US", "values": [{"value": 800}]}
```

### Engagement Calculation
```go
engagement = likes + comments + saves + shares
engagement_rate = (engagement / follower_count) * 100
```

## Special Handling

### Business vs Personal Accounts
- Business accounts have full insights access
- Creator accounts have limited metrics
- Personal accounts: media only, no insights

### Story-Specific Metrics
Stories have unique metrics:
- `exits` - Users who left the story
- `replies` - Direct message replies
- `taps_forward` - Skipped to next
- `taps_back` - Went to previous

### Reels Watch Time
Reels include additional video metrics:
- `plays` - Number of plays
- `total_interactions` - All engagement
- `ig_reels_aggregated_all_plays_count`

### Media Insights by Type
Different media types have different available insights:
```go
// IMAGE/CAROUSEL
metrics := "impressions,reach,likes,comments,saved,shares"

// VIDEO
metrics := "impressions,reach,likes,comments,saved,shares,video_views"

// REEL
metrics := "impressions,reach,likes,comments,saved,shares,plays,total_interactions"
```

### Connected via Instagram
Some accounts connect Instagram through Facebook:
```go
if account.ConnectedViaInstagram {
    // Use Facebook page token
    accessToken = account.FacebookPageToken
}
```

## Date Range Configuration

| Sync Type | Media | Insights |
|-----------|-------|----------|
| Incremental | 7 days | 7 days |
| Immediate | 30 days | 30 days |
| Full Sync | 90 days | 90 days |

## MongoDB Model

```go
type InstagramAccount struct {
    ID                  primitive.ObjectID `bson:"_id"`
    PlatformType        string            `bson:"platform_type"` // "instagram"
    PlatformIdentifier  string            `bson:"platform_identifier"` // instagram_id
    Username            string            `bson:"username"`
    IsBusiness          bool              `bson:"is_business"`
    AccessToken         string            `bson:"access_token"`
    // Token may also be in user_details.access_token
    UserDetails         map[string]any    `bson:"user_details"`
    LastAnalyticsUpdate *time.Time        `bson:"last_analytics_update"`
}
```

## Work Order Format

```json
{
  "id": "mongodb_object_id",
  "instagram_id": "instagram_business_id",
  "username": "account_username",
  "type": "Business",
  "access_token": "encrypted_token",
  "workspace_id": "workspace_id",
  "sync_type": "incremental",
  "connected_via_instagram": false
}
```

## Performance Characteristics

- **Throughput**: ~100 media items per request
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

- Token expiration: Re-authentication required
- Rate limiting: Automatic backoff
- Private accounts: Skip with warning
- Invalid media: Log and continue

## Insights API Limitations

Instagram Insights API has specific limitations:
- Data available for last 2 years only
- Daily insights: up to 30 days back
- Lifetime insights: all time
- Breakdown data: varies by metric

## Key Files

| File | Purpose |
|------|---------|
| `src/clients/social/instagram.go` | API client (42KB) |
| `src/utils/parsing/instagram_parser.go` | Data transformation |
| `src/services/instagram/instagram-fetcher/main.go` | Fetcher service |
| `src/models/kafka/instagram.go` | Kafka message models |
| `src/models/db/clickhouse/instagram.go` | ClickHouse schemas |
