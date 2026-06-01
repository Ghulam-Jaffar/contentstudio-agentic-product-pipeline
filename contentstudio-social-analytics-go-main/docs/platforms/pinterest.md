# Pinterest Analytics Platform Documentation

## Overview

Pinterest integration uses the Pinterest API v5 with OAuth2 Bearer token authentication. It supports two account types — Profile and Board — with comprehensive pin-level and user-level daily analytics including engagement, video, and interaction metrics.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Scheduler     │────▶│  work-order-     │────▶│   Fetcher       │
│  (Account       │     │  pinterest       │     │  (Pinterest     │
│   Fetcher)      │     │  (Batch: 200)    │     │   API v5)       │
└─────────────────┘     └──────────────────┘     └────────┬────────┘
                                                          │
                                                          ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Analytics      │◀────│  raw-pinterest-  │◀────│   (Direct to    │
│  Sink           │     │  */parsed-*      │     │    Sink)        │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

## Services (4 Microservices)

| Service | Purpose | Consumer Group |
|---------|---------|----------------|
| `pinterest-fetcher` | Work order consumption and API calls | `pinterest-fetcher-group` |
| `pinterest-parser` | Raw data transformation to parsed format | `pinterest-parser-group` |
| `pinterest-immediate-processor` | Real-time on-demand processing | `pinterest-immediate-processor-group` |
| `pinterest-analytics-sink` | Merged parser+sink for batch storage | `pinterest-analytics-sink-group` |

**Note**: The analytics sink is a merged parser+sink service that consumes directly from `raw-*` topics, parses inline, and batch-inserts to ClickHouse.

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-pinterest` | Batch work order messages (200 accounts max) |
| `immediate-work-order-pinterest` | Real-time processing requests |
| `raw-pinterest-users` | Raw user profile data |
| `raw-pinterest-boards` | Raw board metadata |
| `raw-pinterest-pins` | Raw pin metadata with media |
| `raw-pinterest-pin-insights` | Raw daily pin analytics |
| `raw-pinterest-user-insights` | Raw daily user analytics |
| `parsed-pinterest-users` | Parsed user data |
| `parsed-pinterest-boards` | Parsed board data |
| `parsed-pinterest-pins` | Parsed pin data |
| `parsed-pinterest-pin-insights` | Parsed pin insights |
| `parsed-pinterest-user-insights` | Parsed user insights |

## ClickHouse Tables

### `pinterest_pins`
Pin metadata and content:
- Identity: pin_id, user_id, board_id, board_section_id, parent_pin_id
- Content: title, note, description, link, dominant_color
- Media: media_type, cover_image_url, video_url, duration, height, width
- Classification: creative_type, is_standard, is_owner, has_been_promoted
- Ownership: board_owner, product_tags (Array)
- Temporal: created_at, day_of_week, hour_of_day, inserted_at

### `pinterest_pin_insights`
Daily pin-level analytics (28 metrics):
- Engagement: pin_clicks, clickthrough, saves, outbound_click, engagement
- Rates: clickthrough_rate, engagement_rate, save_rate
- Video: video_mrc_view, video_start, video_10s_view, video_avg_watch_time, video_v50_watch_time, full_screen_play, full_screen_playtime, quartile_95s_percent_view
- Interaction: impression, profile_visit, closeup, user_follow
- Status: data_status (READY, ESTIMATE, PROCESSING, etc.)

### `pinterest_boards`
Board metadata:
- Identity: board_id, user_id, owner
- Content: name, description, privacy
- Media: image_cover_url, pin_thumbnail_urls (Array)
- Counts: pin_count, follower_count, collaborator_count
- Temporal: created_at, inserted_at

### `pinterest_users`
User profile snapshots:
- Identity: user_id, username, account_type
- Profile: about, profile_image, website_url, business_name
- Counts: board_count, pin_count, follower_count, following_count, monthly_views
- Temporal: inserted_at

### `pinterest_user_insights`
Daily user-level analytics (24 metrics):
- Same as pin_insights but at account level
- Includes pin_click_rate (not in pin insights)
- Excludes user_follow (only in pin insights)

## API Client Methods

```go
// User account
GetUserAccount(accessToken)
GetUserAccountAnalytics(accessToken, startDate, endDate)

// Boards
GetBoards(accessToken)
GetBoard(accessToken, boardID)

// Pins
GetBoardPins(accessToken, boardID, pageSize, bookmark)
GetUserPins(accessToken, pageSize, bookmark)

// Pin analytics
GetPinAnalytics(accessToken, pinID, startDate, endDate)
GetMultiPinAnalytics(accessToken, pinIDs, startDate, endDate)
```

## Rate Limiting

| Limit Type | Configuration |
|------------|---------------|
| Requests per second | 1 RPS |
| Burst | 1 request |
| HTTP Timeout | 30 seconds |
| Max Retries | 3 |
| Base Backoff | 1 second |
| Max Backoff | 10 seconds |
| 429 Handling | Waits x-ratelimit-reset-seconds + 1 |

## Custom Parsing Logic

### Multi-Pin Analytics Batching
Analytics are fetched in batches of 25 pins per API call, with automatic fallback to single-pin requests on failure:
```go
const PinterestMultiPinBatchSize = 25
```

### Record ID Generation
MD5 hash for ClickHouse deduplication:
```go
recordID = MD5(entity_id + "_" + date.Format("20060102"))
```

### Engagement Rate Calculation
```go
engagement = pin_clicks + clickthrough + saves + outbound_clicks
engagement_rate = engagement / impression  // (if impression > 0)
```

### Date Parsing
Pinterest API returns timestamps in multiple formats. The parser handles:
```go
formats := []string{
    time.RFC3339,           // "2024-01-15T10:30:00+00:00"
    time.RFC3339Nano,       // "2024-01-15T10:30:00.000000+00:00"
    "2006-01-02T15:04:05.000Z",
    "2006-01-02T15:04:05Z",
    "2006-01-02T15:04:05", // Most common from API
    "2006-01-02",
}
```

### Boolean to String Conversion
ClickHouse stores booleans as strings:
```go
isStandard = "1"  // true
isOwner    = "0"  // false
```

### Data Status Filtering
Insights with these statuses are skipped:
- `PROCESSING` — Data not ready yet
- `BEFORE_PIN_CREATED` — Pin didn't exist on this date
- `BEFORE_BUSINESS_CREATED` — Business account didn't exist

## Special Handling

### Board vs Profile Accounts
- **Board Account**: Fetches only the specific board and its pins
- **Profile Account**: Fetches ALL boards, but only processes PUBLIC boards

### Analytics Date Lag
Fetcher excludes last 2 days of analytics (incomplete data):
```go
endDate = time.Now().AddDate(0, 0, -2)
```

### Bookmark-Based Pagination
Pinterest uses bookmark-based pagination for pins:
```go
type PinterestPinsResponse struct {
    Items    []PinterestPin `json:"items"`
    Bookmark string         `json:"bookmark,omitempty"`
}
```

### Token Encryption
- Tokens stored encrypted in MongoDB
- Decrypted at runtime using `APP_DECRYPTION_KEY`
- Never logged (only prefix and length for debugging)

### Account Semaphore (Fetcher)
One concurrent request per account prevents duplicate processing.

## Date Range Configuration

| Sync Type | Data Period | Page Size | Max Pages | Description |
|-----------|-------------|-----------|-----------|-------------|
| Incremental | 7 days | 100 | 2 | Regular scheduled updates |
| Full Sync | 86 days | 250 | Unlimited | Complete data refresh |
| Immediate | 86 days | 250 | Unlimited | On-demand API processing |

## MongoDB Model

```go
type SocialIntegration struct {
    ID                  primitive.ObjectID     `bson:"_id"`
    PlatformType        string                `bson:"platform_type"`  // "pinterest"
    PlatformIdentifier  string                `bson:"platform_identifier"`
    AccessToken         string                `bson:"access_token"`
    BoardID             string                `bson:"board_id,omitempty"`
    ExtraData           map[string]interface{} `bson:",inline"`
    LastAnalyticsUpdate *time.Time            `bson:"last_analytics_update"`
}
// Board ID retrieved via: GetStringFromExtraData(account.ExtraData, "board_id")
```

## Work Order Format

### Batch Work Order
```json
{
  "batch_id": "uuid-v4",
  "sync_type": "incremental",
  "accounts": [
    {
      "id": "mongodb_object_id",
      "account_id": "pinterest_user_id",
      "access_token": "encrypted_token",
      "account_type": "profile",
      "workspace_id": "workspace_id",
      "sync_type": "incremental"
    },
    {
      "id": "mongodb_object_id",
      "account_id": "pinterest_user_id",
      "access_token": "encrypted_token",
      "account_type": "board",
      "board_id": "board_id",
      "workspace_id": "workspace_id",
      "sync_type": "incremental"
    }
  ],
  "created_at": "2026-01-15T00:00:00Z"
}
```

### Immediate Work Order
```json
{
  "id": "mongodb_object_id",
  "platform": "pinterest",
  "account_id": "pinterest_user_id",
  "type": "board",
  "access_token": "encrypted_token",
  "workspace_id": "workspace_id",
  "sync_type": "immediate"
}
```

## Performance Characteristics

### Fetcher
- **Workers**: 5 concurrent processors
- **Work Channel**: 100 buffer
- **Idle Timeout**: 15 minutes
- **Batch API**: 25 pins per multi-pin analytics request (25x reduction in API calls)

### Parser
- **Parser Workers**: 6 concurrent
- **Publisher Workers**: 6 concurrent
- **Parse Channel**: 500 buffer
- **Publish Channel**: 1000 buffer

### Analytics Sink (Merged Parser+Sink)
- **Parser Workers**: 19 total (3 users, 3 boards, 5 pins, 5 pin insights, 3 user insights)
- **Batch Processors**: 15 total (3 per data type)
- **Batch Size**: 10,000 records per batch
- **Batch Timeout**: 10 seconds
- **Message Channel**: 50,000 buffer
- **Idle Timeout**: 5 minutes

### Immediate Processor
- **Workers**: 5 concurrent (standalone), 15 (unified)
- **Work Channel**: 50 buffer (standalone), 200 (unified)

### Scheduler (Account Fetcher)
- **MongoDB Fetch Size**: 50 accounts per batch
- **Kafka Batch Size**: 200 accounts per message
- **Update Interval**: 6 hours

## Graceful Shutdown

### Fetcher Shutdown
- Idle timeout: 15 minutes
- Check interval: 30 seconds
- Tracks active jobs with atomic counters
- Account-level semaphore for deduplication

### Analytics Sink Shutdown
- Idle timeout: 5 minutes
- Flushes all pending batches before exit
- Waits for batch channels to drain
- Ensures all ClickHouse INSERTs complete

## Error Handling

- Token expiration: 401 response fails immediately (no retry)
- Rate limiting: Waits for `x-ratelimit-reset-seconds` header duration
- API errors: Exponential backoff (1s → 2s → 4s, max 10s)
- Multi-pin analytics failure: Falls back to individual pin requests
- Invalid data: Skipped with warning log

## Key Files

| File | Purpose |
|------|---------|
| `src/clients/social/pinterest.go` | API client |
| `src/services/pinterest/pinterest-fetcher/main.go` | Fetcher service |
| `src/services/pinterest/pinterest-parser/main.go` | Parser service |
| `src/services/pinterest/pinterest-immediate-processor/main.go` | Immediate processor |
| `src/services/pinterest/pinterest-analytics-sink/main.go` | Merged analytics sink |
| `src/models/kafka/pinterest.go` | Kafka message models |
| `src/models/db/clickhouse/pinterest.go` | ClickHouse schemas |
| `src/models/db/clickhouse/conversions/pinterest_clickhouse.go` | ClickHouse conversions |
| `src/db/clickhouse/pinterest.go` | ClickHouse insert queries |
| `src/cmd/jobs/fetcher/pinterest.go` | Batch fetcher job |
| `src/api/immediate_work_apis.go` | Immediate work API endpoint |
