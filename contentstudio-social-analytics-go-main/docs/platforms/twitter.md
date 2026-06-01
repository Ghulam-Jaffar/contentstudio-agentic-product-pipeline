# Twitter Analytics Platform Documentation

## Overview

Twitter integration is production-ready and uses Twitter API v2 with OAuth 1.0a user context authentication.  
The pipeline is split into scheduled batch flow and immediate sync flow.

## Architecture

```
Scheduled Flow:
Unified Account Fetcher
  -> work-order-twitter-batch
  -> twitter-fetcher
  -> raw-twitter-posts / raw-twitter-insights
  -> twitter-analytics-sink
  -> ClickHouse (twitter_posts, twitter_insights)

Immediate Flow:
PHP/Frontend -> /immediate-work (channel=twitter)
  -> immediate-work-order-twitter
  -> unified-immediate-processor
  -> twitter-immediate-processor
  -> ClickHouse (twitter_posts, twitter_insights)
```

## Services

| Service | Purpose | Consumer Group |
|---------|---------|----------------|
| `unified_account_fetcher` | Scheduled account discovery and batch work-order production | N/A |
| `twitter-fetcher` | Consumes batch orders, calls Twitter API, publishes raw topics | `twitter-fetcher-group` |
| `twitter-analytics-sink` | Consumes raw topics and inserts batches to ClickHouse | `twitter-analytics-sink-group` |
| `unified-immediate-processor` | Consumes immediate orders and routes by platform | `twitter-immediate-processor-group` (Twitter topic) |
| `twitter-immediate-processor` | Immediate Twitter API fetch + direct ClickHouse insert | Internal processor in unified immediate worker |

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-twitter-batch` | Batch of Twitter account work orders for scheduled sync |
| `immediate-work-order-twitter` | Immediate one-account work order |
| `raw-twitter-posts` | Parsed tweet payloads for sink insertion |
| `raw-twitter-insights` | Parsed account insights payloads for sink insertion |

## Work Order Contract

`TwitterAccountWorkOrder`:

```json
{
  "id": "mongo_account_id_hex",
  "workspace_id": "workspace_id",
  "twitter_id": "platform_identifier",
  "oauth_token": "encrypted_token",
  "oauth_token_secret": "encrypted_token_secret",
  "post_count": 50,
  "api_key": "developer_app_key",
  "api_secret": "developer_app_secret",
  "app_name": "developer_app_name",
  "app_id": "developer_app_object_id_hex",
  "executed_by": "internal",
  "sync_type": "incremental|full_sync"
}
```

## Scheduling and Eligibility Rules

### Scheduled Flow (`cmd/jobs/fetcher/twitter.go`)

- Source: `social_integrations`
- Base filter: `platform_type=twitter`, `validity=valid`
- Additional checks per account:
  - `developer_app_id` present
  - Matching `developer_apps` entry exists and `analytics_enabled=true`
  - Matching `twitter_jobs_setting` exists by `platform_id=platform_identifier`
  - `job_type` scheduling:
    - `daily`: include
    - `weekly`: include when weekday matches `trigger_day` (Mon=1..Sun=7)
    - `monthly`: include when day-of-month matches `trigger_day`
    - `never`: exclude

### Immediate Flow (`api/immediate_work_apis.go`)

- `developer_app_id` must be present
- developer app must be analytics-enabled
- `twitter_jobs_setting` must exist
- Produces immediate work order with `sync_type=full_sync`

## Twitter API Integration

Shared client: `src/clients/social/twitter.go`

Authentication:

- OAuth 1.0a
- Signature method: HMAC-SHA1
- Consumer key/secret from developer app
- Access token/secret decrypted from work order

Headers:

- `Content-Type: application/json`
- `Accept: application/json`

Endpoints:

- User info:
  - `GET /2/users?ids={twitter_id}&user.fields=...&tweet.fields=...`
- Tweets:
  - `GET /2/users/{twitter_id}/tweets?max_results={post_count}&tweet.fields=...&user.fields=...&media.fields=...&expansions=...`

Pagination behavior:

- If `post_count <= 100`, single-page style fetch up to requested count
- If `post_count > 100`, uses `pagination_token` with `max_results=50` per request until count is met or next token ends

Important:

- Twitter implementation uses `post_count` only
- No 14-day / 90-day publication-date cutoffs in Twitter fetch loop

## ClickHouse Storage

Tables:

- `twitter_posts`
- `twitter_insights`

Notable type alignment with production:

- `twitter_insights.verified` stored as `String`
- `twitter_posts.author_id_created` stored as `DateTime64(6)`

## MongoDB Job Metadata

Collection: `twitter_jobs_metadata`

Inserted after processing via `InsertTwitterJobMetadata`:

- `platform_id`
- `workspace_id`
- `platform_type=twitter`
- `job_type=posts`
- `credits_used` (`tweets_fetched + 1` when user info was fetched)
- `executed_by`
- `app_id`, `app_name`
- `job_executed_at`, `day_of_week`, `hour_of_day`
- `created_at`, `updated_at`

## Key Files

| File | Purpose |
|------|---------|
| `src/cmd/jobs/fetcher/twitter.go` | Scheduled Twitter account filtering + batch work-order production |
| `src/api/immediate_work_apis.go` | Immediate Twitter work-order production |
| `src/services/twitter/twitter-fetcher/run.go` | Batch consumer + Twitter API fetch + raw topic production |
| `src/services/twitter/twitter-analytics-sink/run.go` | Raw topic consume + ClickHouse insertion |
| `src/services/twitter/twitter-immediate-processor/processor/processor.go` | Immediate processing path |
| `src/clients/social/twitter.go` | OAuth1 signing and Twitter API v2 client |
| `src/models/db/clickhouse/twitter.go` | ClickHouse model contracts |
| `src/models/kafka/twitter.go` | Kafka work-order and payload contracts |

