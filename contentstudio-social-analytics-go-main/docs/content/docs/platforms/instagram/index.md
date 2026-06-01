---
title: Instagram Integration
description: Instagram data pipeline and analytics
---

Documentation for Instagram social media analytics integration.

## Overview

The Instagram integration pipeline follows the standard five-stage architecture:

1. **Scheduler**: Unified Account Fetcher
2. **Fetcher**: Instagram Fetcher Service
3. **Parser**: Instagram Posts Parser
4. **Processor**: Instagram Immediate Processor
5. **Sink**: Instagram ClickHouse Sink

## Data Fetch Periods

| Sync Type | Data Period | Description |
|-----------|-------------|-------------|
| **Incremental Sync** | 14 days | Regular scheduled updates fetch posts from last 14 days |
| **Immediate Sync** | 30 days | On-demand API requests fetch posts from last 30 days |
| **Full Sync** | 89 days | Complete data refresh fetches posts from last 89 days |

## Scheduler Configuration

| Parameter | Value | Description |
|-----------|-------|-------------|
| Update Interval | 6 hours | Time between scheduled analytics updates |
| Batch Size | 200 | Number of accounts processed per batch |

## URL Refresher Job

Instagram media URLs can be refreshed independently of the analytics fetch pipeline.

**Job**: `url_refresher`
**Platform Flag**: `-platform instagram`

**Purpose**:
- Refreshes stale Instagram media URLs in `instagram_posts`
- Prefers decrypted long-lived tokens when available

**Notes**:
- The job reads accounts from MongoDB and updates ClickHouse rows older than the refresh threshold
- Use `-platform all` to run every refresher sequentially

## Services

- [Instagram Fetcher](/services/fetcher/instagram-fetcher)
- [Instagram Parser](/services/parser/instagram-posts-parser)
- [Instagram Processor](/services/processor/instagram-immediate-processor)
- [Instagram Sink](/services/sink/instagram-clickhouse-sink)

## Data Flow

The Instagram pipeline processes:
- Posts and their engagement metrics
- Stories analytics
- Reels performance data
- Account insights

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-instagram` | Scheduled work orders |
| `immediate-work-order-instagram` | On-demand API requests |
| `raw-instagram-posts` | Raw post data from API |
| `parsed-instagram-posts` | Processed post analytics |

## Configuration

Key environment variables for Instagram integration:
- `APP_INSTAGRAM_CLIENT_ID`
- `APP_INSTAGRAM_CLIENT_SECRET`
- `APP_INSTAGRAM_ACCESS_TOKEN`
