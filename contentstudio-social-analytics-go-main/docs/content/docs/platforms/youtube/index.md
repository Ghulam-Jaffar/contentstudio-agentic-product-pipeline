---
title: YouTube Integration
description: YouTube data pipeline and analytics
---

Documentation for YouTube social media analytics integration.

## Overview

The YouTube integration pipeline follows the standard five-stage architecture:

1. **Scheduler**: Unified Account Fetcher
2. **Fetcher**: YouTube Fetcher Service
3. **Parser**: YouTube Parser
4. **Processor**: YouTube Immediate Processor
5. **Sink**: YouTube ClickHouse Sink

## Services

- YouTube Fetcher (`cmd/jobs/fetcher/youtube.go`)
- Unified Account Fetcher (scheduler)
- Unified Immediate Processor (API-based immediate sync)

## Data Flow

The YouTube pipeline processes:
- Channel videos and their metrics
- Engagement data (views, likes, comments)
- Video analytics and performance data
- Channel insights

## Data Fetch Periods

| Sync Type | Data Period | Description |
|-----------|-------------|-------------|
| **Incremental Sync** | 14 days | Regular scheduled updates fetch videos from last 14 days |
| **Immediate Sync** | 90 days | On-demand API requests fetch videos from last 90 days |
| **Full Sync** | 365 days | Complete data refresh fetches videos from last 365 days |

## Scheduler Configuration

| Parameter | Value | Description |
|-----------|-------|-------------|
| Update Interval | 6 hours | Time between scheduled analytics updates |
| Batch Size | 200 | Number of accounts processed per batch |
| Consent Window | 30 days | Maximum days since last YouTube consent |

## YouTube Consent Validation

YouTube integration requires valid user consent. The system validates that:
- `preferences.last_youtube_consent_time` exists and is not null
- Consent was given within the last 30 days

Accounts with expired or missing consent are excluded from processing.

## Token Format

YouTube supports two token storage formats:

### String Format (Legacy)
```json
{
  "access_token": "ya29.xxx...",
  "refresh_token": "1//xxx..."
}
```

### Embedded Document Format (Current)
```json
{
  "access_token": {
    "access_token": "ya29.xxx...",
    "refresh_token": "1//xxx...",
    "expires_in": 3600,
    "token_type": "Bearer"
  }
}
```

The system automatically detects and handles both formats.

## Configuration

Key environment variables for YouTube integration:
- `APP_YOUTUBE_API_KEY`
- `APP_YOUTUBE_CLIENT_ID`
- `APP_YOUTUBE_CLIENT_SECRET`

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-youtube` | Scheduled work orders |
| `immediate-work-order-youtube` | On-demand API requests |
| `raw-youtube-posts` | Raw video data from API |
| `parsed-youtube-posts` | Processed video analytics |

## Recent Changes

### YouTube Consent Time Filter (2024)
- Added 30-day consent window validation for scheduler
- Added consent validation for immediate sync API
- Invalid consent returns JSON error: `{"code": "CONSENT_EXPIRED", "message": "YouTube consent expired"}`

### Embedded Document Token Support (2024)
- Added support for embedded document `access_token` format
- Backward compatible with string token format
- Automatic format detection in MongoDB BSON decoding
