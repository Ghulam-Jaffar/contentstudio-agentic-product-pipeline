# Google My Business (GMB) Analytics Platform Documentation

## Overview

GMB integration uses the Google My Business API (v1/v4) with OAuth2 Bearer token authentication. It provides location-level analytics covering performance metrics, search keywords, local posts, reviews, and media assets. A key prerequisite is Voice of Merchant (VoM) verification — performance metrics and search keywords are only fetched for locations with VoM status.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Scheduler     │────▶│  work-order-     │────▶│   Fetcher       │
│  (Account       │     │  gmb-batch       │     │  (GMB API)      │
│   Fetcher)      │     │  (Batch: 200)    │     │                 │
└─────────────────┘     └──────────────────┘     └────────┬────────┘
                                                          │
                                                          ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Analytics      │◀────│  raw-gmb-data    │◀────│   (Direct to    │
│  Sink           │     │                  │     │    Sink)        │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

## Services (2 Microservices + Unified Processor)

| Service | Purpose | Consumer Group |
|---------|---------|----------------|
| `gmb-fetcher` | Daily batch work order consumption and API calls | `gmb-fetcher-group` |
| `gmb-analytics-sink` | Merged parser+sink for batch storage to ClickHouse | `gmb-analytics-sink-group` |

**Note**: Immediate/on-demand processing is handled by the **unified immediate processor** (not a standalone service). The GMB processor package (`services/gmb/gmb-immediate-processor/processor/`) is used as a library by the unified processor. The analytics sink consumes from raw GMB topics, parses inline, and batch-inserts to ClickHouse.

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-gmb-batch` | Batch work order messages (200 accounts max) |
| `immediate-work-order-gmb` | Real-time processing requests |
| `raw-gmb-data` | Raw GMB data for sink processing |

## Platform Identifier

The GMB platform identifier follows the format:

```
accounts/{accountID}/locations/{locationID}
```

Example: `accounts/111098760901606453992/locations/2941480710306834283`

Both `accountID` and `locationID` are extracted by splitting on `/`.

## Voice of Merchant (VoM)

Before fetching performance metrics and search keywords, the processor checks VoM status:
- **hasVoiceOfMerchant = true**: Fetch all 5 data types
- **hasVoiceOfMerchant = false**: Skip performance metrics and search keywords; still fetch local posts, reviews, and media assets

VoM is checked immediately after token refresh, before any data fetching.

## Data Processing Flow

1. **Token Refresh** — Decrypt and refresh OAuth2 access token
2. **VoM Check** — Verify Voice of Merchant status for the location
3. **Performance Metrics** — Daily business impressions, clicks, actions (if VoM)
4. **Search Keywords** — Monthly keyword impressions (if VoM)
5. **Local Posts** — Business posts with media
6. **Reviews** — Customer reviews with replies
7. **Media Assets** — Photos and videos

## ClickHouse Tables

### `gmb_daily_metrics`
Daily location performance metrics (11 metrics from Business Profile Performance API).
- Identity: account_id, location_id, account_name, location_name, platform_name
- Impressions: business_impressions_desktop_maps, business_impressions_desktop_search, business_impressions_mobile_maps, business_impressions_mobile_search
- Actions: call_clicks, website_clicks, business_direction_requests
- Engagement: business_conversations, business_bookings, business_food_orders, business_food_menu_clicks
- Temporal: inserted_at, created_at

Engine: `ReplacingMergeTree(inserted_at)`, Partition: `toYYYYMM(created_at)`, Order: `(location_id, created_at)`

### `gmb_search_keywords_monthly`
Monthly keyword search impressions.
- Identity: account_id, location_id, account_name, location_name, platform_name
- Keywords: keyword, impressions_value, impressions_threshold
- Temporal: inserted_at, keyword_month

Engine: `ReplacingMergeTree(inserted_at)`, Partition: `toYYYYMM(keyword_month)`, Order: `(location_id, keyword_month, keyword)`

### `gmb_local_posts`
Business post metadata with media arrays.
- Identity: account_id, location_id, account_name, location_name, platform_name, language_code
- Content: post_name, state, topic_type, search_url
- Media: media_names (Array), media_formats (Array), media_google_urls (Array)
- Temporal: inserted_at, created_at, updated_at

Engine: `ReplacingMergeTree(inserted_at)`, Partition: `toYYYYMM(created_at)`, Order: `(location_id, post_name)`

### `gmb_reviews`
Customer review data with replies.
- Identity: account_id, location_id, account_name, location_name, platform_name
- Review: review_id, review_name, star_rating, comment
- Reviewer: reviewer_display_name, reviewer_profile_photo_url
- Reply: reply_comment, reply_update_time
- Temporal: inserted_at, created_at, updated_at

Engine: `ReplacingMergeTree(inserted_at)`, Partition: `toYYYYMM(created_at)`, Order: `(location_id, review_id)`

### `gmb_media_assets`
Location photos and videos.
- Identity: account_id, location_id, account_name, location_name, platform_name, language_code
- Media: media_name, source_url, media_format, google_url, thumbnail_url
- Dimensions: width_pixels, height_pixels
- Classification: location_association_category
- Temporal: inserted_at, created_at

Engine: `ReplacingMergeTree(inserted_at)`, Partition: `toYYYYMM(created_at)`, Order: `(location_id, media_name)`

## API Client Methods

```go
// Voice of Merchant verification
FetchVoiceOfMerchant(ctx, accountID, locationID, accessToken) (*VoMResponse, error)

// Performance metrics (requires VoM)
FetchPerformanceMetrics(ctx, accountID, locationID, accessToken, startDate, endDate) (*PerformanceMetricsResponse, error)

// Search keywords (requires VoM)
FetchSearchKeywords(ctx, accountID, locationID, accessToken, startDate, endDate) (*SearchKeywordsResponse, error)

// Local posts
FetchLocalPosts(ctx, accountID, locationID, accessToken) (*LocalPostsResponse, error)

// Reviews
FetchReviews(ctx, accountID, locationID, accessToken) (*ReviewsResponse, error)

// Media assets
FetchMediaAssets(ctx, accountID, locationID, accessToken, languageCode) (*MediaAssetsResponse, error)

// Token management
RefreshToken(ctx, refreshToken) (*TokenResponse, error)
```

## Error Handling

- All API errors are logged as **warnings** — a single failing data type does not block other data fetches
- Token refresh failure skips the account entirely (critical error)
- VoM check failure logs a warning and proceeds with non-VoM data (posts, reviews, media)
- HTTP client has built-in retry (1 retry for all GET requests, 1 retry for RefreshToken POST)

## MongoDB Updates

After VoM check, the processor updates the social account document in MongoDB:
- Sets `has_voice_of_merchant` field on the account
- Used for UI display and filtering

## Unified Service Integration

GMB is fully integrated into the unified services:
- **Unified Account Fetcher**: Includes "gmb" in platform list, scale factor 1000
- **Unified Immediate Processor**: Consumes `immediate-work-order-gmb`, converts to `GMBAccountWorkOrder`
- **API Layer**: `ProcessGMBWork` endpoint for on-demand processing
