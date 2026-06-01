---
title: Twitter Integration
description: Twitter data pipeline and analytics
---

Documentation for Twitter social analytics integration.

## Overview

Twitter implementation uses the unified scheduler + immediate architecture with two runtime services:

1. **Twitter Fetcher** (`services/twitter/twitter-fetcher`) for scheduled batch jobs
2. **Twitter Analytics Sink** (`services/twitter/twitter-analytics-sink`) for parsing + ClickHouse inserts

Immediate sync is handled by:

- Immediate API producer (`api/immediate_work_apis.go`)
- Unified immediate processor (`services/unified/immediate-processor`)
- Twitter immediate processor (`services/twitter/twitter-immediate-processor/processor`)

## Scheduler (Daily/Weekly/Monthly/Never)

The unified account fetcher calls `cmd/jobs/fetcher/twitter.go`.

Eligibility and filtering:

- Social integration must be `platform_type=twitter` and `validity=valid`
- `developer_app_id` must be present
- Matching `developer_apps` document must exist with `analytics_enabled=true`
- Matching `twitter_jobs_setting` by `platform_id=platform_identifier` must exist
- `job_type` handling:
  - `daily`: always schedule
  - `weekly`: schedule when current weekday equals `trigger_day` (Mon=1..Sun=7)
  - `monthly`: schedule when day-of-month equals `trigger_day`
  - `never`: do not schedule

Batch work orders are produced to `work-order-twitter-batch`.

## Immediate Sync

The immediate endpoint is `POST /immediate-work` with:

```json
{
  "account_id": "<social_integration_id>",
  "channel": "twitter"
}
```

`ProcessTwitterWork` checks:

- `developer_app_id` is non-null
- Developer app exists and is analytics-enabled
- `twitter_jobs_setting` exists for `platform_identifier`

On success it produces `TwitterAccountWorkOrder` to:

- `immediate-work-order-twitter`

`sync_type` is `full_sync` for immediate requests.

## Twitter API Calls

Both scheduled fetcher and immediate processor call:

- `GET /2/users?ids={twitter_id}&user.fields=...&tweet.fields=...`
- `GET /2/users/{twitter_id}/tweets?max_results={post_count}&tweet.fields=...&user.fields=...&media.fields=...&expansions=...`

Rules:

- OAuth 1.0a user context
- Signature method: `HMAC-SHA1`
- Headers: `Content-Type: application/json`, `Accept: application/json`
- If `post_count > 100`: paginate with `pagination_token`, `max_results=50` per call
- Uses only `post_count` (no 14/90-day filtering in Twitter flow)

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-twitter-batch` | Scheduled batch work orders |
| `immediate-work-order-twitter` | Immediate work orders |
| `raw-twitter-posts` | Parsed tweet payloads from fetcher |
| `raw-twitter-insights` | Parsed account insights payloads from fetcher |

## ClickHouse Tables

| Table | Data |
|-------|------|
| `twitter_posts` | Tweet-level analytics rows |
| `twitter_insights` | Account-level snapshot metrics |

## Job Metadata Logging

After successful processing (scheduled and immediate), metadata is written to MongoDB collection `twitter_jobs_metadata` with:

- `platform_id`, `workspace_id`, `platform_type=twitter`
- `job_type=posts`
- `credits_used = tweets_fetched + 1 (if user info fetched)`
- `executed_by`, `app_id`, `app_name`
- `job_executed_at`, `day_of_week`, `hour_of_day`, `created_at`, `updated_at`

## Services

- Twitter Fetcher: `src/services/twitter/twitter-fetcher`
- Twitter Analytics Sink: `src/services/twitter/twitter-analytics-sink`
- Twitter Immediate Processor: `src/services/twitter/twitter-immediate-processor/processor`

