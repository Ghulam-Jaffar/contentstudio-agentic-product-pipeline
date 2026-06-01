# LinkedIn Analytics Platform Documentation

## Overview

LinkedIn supports both Organization Pages and Member Profiles. It uses LinkedIn Marketing API with URN-based identifiers and supports batch work orders for efficient processing of multiple accounts.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Scheduler     │────▶│  work-order-     │────▶│   Fetcher       │
│  (Account       │     │  linkedin        │     │  (Marketing API)│
│   Fetcher)      │     │  (Batch: 200)    │     └────────┬────────┘
└─────────────────┘     └──────────────────┘              │
                                                          ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  ClickHouse     │◀────│  parsed-linkedin │◀────│   Parser        │
│  Sink           │     │  -posts/-insights│     │  (URN handling) │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

## Services (5 Microservices)

| Service | Purpose | Consumer Group |
|---------|---------|----------------|
| `linkedin-fetcher` | Work order consumption and API calls | `linkedin-fetcher-group` |
| `linkedin-parser` | Post and insights parsing | `linkedin-parser-group` |
| `linkedin-immediate-processor` | Real-time processing | `linkedin-immediate-processor-group` |
| `linkedin-clickhouse-sink` | Batch storage | `linkedin-clickhouse-sink-group` |
| `linkedin-analytics-sink` | Analytics aggregation | `linkedin-analytics-sink-group` |

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-linkedin` | Batch work orders (max 200 accounts) |
| `immediate-work-order-linkedin` | Real-time requests |
| `raw-linkedin-posts` | Raw post JSON from API |
| `raw-linkedin-followers` | Follower data |
| `raw-linkedin-stats` | Post statistics |
| `raw-linkedin-geo-data` | Geographic breakdowns |
| `parsed-linkedin-posts` | Normalized post data |
| `parsed-linkedin-insights` | Daily insights |

## ClickHouse Tables

### `linkedin_posts`
Post analytics including:
- Engagement: likes, comments, shares, clicks
- Reach and impressions
- Hashtags array
- Media assets (images, videos, documents)
- Poll data
- Article URLs

### `linkedin_insights`
Daily organization/profile insights:
- Follower counts by type (organic, paid)
- Page views by section
- Audience demographics:
  - Seniority levels
  - Industries
  - Company sizes
  - Job functions
  - Countries/regions

### `linkedin_media_assets`
Media file storage for posts:
- Images
- Videos
- Documents (PDFs, etc.)

## API Client Methods

```go
// Post fetching
FetchShares(organisationID, accessToken)
FetchPostsPaginated(linkedinID, entityType, accessToken, cutoffTime)

// Media fetching (bulk)
FetchImagesRaw(imageIDs, accessToken)
FetchVideosRaw(videoIDs, accessToken)
FetchDocumentsRaw(docIDs, accessToken)

// Statistics
FetchStatsRaw(linkedinID, ugcPosts, shares, accessToken)

// Follower data
FetchFollowerData(linkedinID, accessToken)
FetchFollowerDataWithGeoNames(linkedinID, accessToken, geoNames)
FetchFollowerStatsWithGeoIDs(linkedinID, accessToken)

// Geo resolution
ResolveGeoIDs(geoIDs, accessToken)

// Organization details
FetchOrganizationDetails(linkedinID, accessToken)
FetchPageStatistics(linkedinID, accessToken, startMs, endMs)
```

## Rate Limiting

| Limit Type | Configuration |
|------------|---------------|
| HTTP Timeout | 30 seconds |
| Pagination Delay | 100ms between pages |
| Batch Size | 200 accounts max |

## Custom Parsing Logic

### URN Format Handling
LinkedIn uses URN format for all identifiers:
```go
// Organization URN
"urn:li:organization:12345678"

// Share URN
"urn:li:share:7123456789012345678"

// UGC Post URN
"urn:li:ugcPost:7123456789012345678"

// Person URN (for member profiles)
"urn:li:person:abc123def456"
```

### Entity Type Classification
```go
const (
    EntityTypeOrganization = "organization"
    EntityTypePerson       = "person"
)
```

### Post Activity Types
```go
// Post content types
"ARTICLE"     // Link to article
"IMAGE"       // Single image
"VIDEO"       // Video content
"DOCUMENT"    // PDF or document
"MULTI_IMAGE" // Multiple images
"POLL"        // Poll post
"NONE"        // Text only
```

### Engagement Calculation
```go
engagement = likes + comments + shares + clicks
engagement_rate = (engagement / follower_count) * 100
```

### Poll Data Extraction
```go
type PollOption struct {
    Text       string `json:"text"`
    VoteCount  int64  `json:"vote_count"`
    Percentage float64 `json:"percentage"`
}
```

### Geo ID Resolution
LinkedIn returns geo IDs that need resolution:
```go
// Raw: "urn:li:geo:103644278"
// Resolved: "United States"

geoNames, err := client.ResolveGeoIDs(geoIDs, accessToken)
```

## Special Handling

### Two API Versions
LinkedIn has deprecated v1 endpoints, now using v2:
```go
// v2 endpoint (current)
baseURL = "https://api.linkedin.com/v2"

// API version header
"LinkedIn-Version": "202509"
```

### Batch Work Orders
LinkedIn uses batch processing for efficiency:
```json
{
  "batch_id": "uuid-v4",
  "sync_type": "incremental",
  "account_type": "Page",
  "accounts": [
    {"linkedin_id": "123", "access_token": "..."},
    {"linkedin_id": "456", "access_token": "..."}
  ],
  "created_at": "2024-01-15T00:00:00Z"
}
```

### Member vs Organization
Different API endpoints for different entity types:
```go
if entityType == "organization" {
    url = "/organizationShares"
} else {
    url = "/shares" // For member profiles
}
```

### Lifecycle States
Posts have lifecycle states:
```go
"PUBLISHED"   // Live post
"DRAFT"       // Unpublished draft
"DELETED"     // Removed
"PROCESSING"  // Being processed
```

### Visibility Levels
```go
"PUBLIC"      // Anyone can see
"CONNECTIONS" // First-degree only
"LOGGED_IN"   // LinkedIn members only
```

## Date Range Configuration

| Sync Type | Posts | Insights |
|-----------|-------|----------|
| Incremental | 7 days | 7 days |
| Immediate | 30 days | 30 days |
| Full Sync | 90 days | 90 days |

## MongoDB Model

```go
type LinkedInAccount struct {
    ID                  primitive.ObjectID `bson:"_id"`
    PlatformType        string            `bson:"platform_type"` // "linkedin"
    PlatformIdentifier  string            `bson:"platform_identifier"` // linkedin_id
    LinkedInProfileID   string            `bson:"linkedin_profile_id"`
    Headline            string            `bson:"headline"`
    Type                string            `bson:"type"` // "Page" or "Profile"
    AccessToken         string            `bson:"access_token"`
    RefreshToken        string            `bson:"refresh_token"`
    LastAnalyticsUpdate *time.Time        `bson:"last_analytics_update"`
}
```

## Work Order Format

### Single Work Order
```json
{
  "id": "mongodb_object_id",
  "linkedin_id": "12345678",
  "type": "Page",
  "access_token": "encrypted_token",
  "workspace_id": "workspace_id",
  "sync_type": "incremental"
}
```

### Batch Work Order
```json
{
  "batch_id": "550e8400-e29b-41d4-a716-446655440000",
  "sync_type": "incremental",
  "account_type": "Page",
  "accounts": [
    {
      "id": "mongodb_id_1",
      "linkedin_id": "123",
      "access_token": "token_1"
    },
    {
      "id": "mongodb_id_2",
      "linkedin_id": "456",
      "access_token": "token_2"
    }
  ],
  "created_at": "2024-01-15T00:00:00Z"
}
```

## Performance Characteristics

- **Batch Size**: 200 accounts per batch work order
- **Posts per Page**: 100 posts
- **ClickHouse Batch**: 1000 items per insert
- **Workers**: 10 concurrent processors

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

- Token expiration: OAuth refresh flow
- Rate limiting: Backoff with delay
- Geo resolution failure: Fallback to raw URN
- Invalid URN: Skip with warning

## Audience Demographics

LinkedIn provides rich demographic data:

### Seniority Levels
- Entry, Senior, Manager, Director, VP, CXO, Owner, Partner

### Industries
- Technology, Finance, Healthcare, Education, etc.

### Company Sizes
- 1-10, 11-50, 51-200, 201-500, 501-1000, 1001-5000, 5001-10000, 10001+

### Job Functions
- Engineering, Sales, Marketing, Operations, HR, Finance, etc.

## Key Files

| File | Purpose |
|------|---------|
| `src/clients/social/linkedin.go` | API client (26KB) |
| `src/utils/parsing/linkedin_parser.go` | Data transformation |
| `src/services/linkedin/linkedin-fetcher/main.go` | Fetcher service |
| `src/models/kafka/linkedin.go` | Kafka message models |
| `src/models/db/clickhouse/linkedin.go` | ClickHouse schemas |
