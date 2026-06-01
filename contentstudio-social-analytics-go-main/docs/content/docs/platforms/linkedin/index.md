---
title: LinkedIn Integration
description: LinkedIn data pipeline and analytics
---

Documentation for LinkedIn social media analytics integration.

## Overview

The LinkedIn integration pipeline follows the standard five-stage architecture:

1. **Scheduler**: Unified Account Fetcher
2. **Fetcher**: LinkedIn Fetcher Service
3. **Parser**: LinkedIn Parser
4. **Processor**: LinkedIn Immediate Processor
5. **Sink**: LinkedIn ClickHouse Sink

## Data Fetch Periods

| Sync Type | Data Period | Description |
|-----------|-------------|-------------|
| **Incremental Sync** | 10 days | Regular scheduled updates fetch posts from last 10 days |
| **Full Sync** | 365 days | Complete data refresh fetches posts from last 365 days (1 year) |

## Scheduler Configuration

| Parameter | Value | Description |
|-----------|-------|-------------|
| Update Interval | 6 hours | Time between scheduled analytics updates |
| Batch Size | 200 | Number of accounts processed per batch |

## URL Refresher Job

LinkedIn post URLs can be refreshed independently of the analytics fetch pipeline.

**Job**: `url_refresher`
**Platform Flag**: `-platform linkedin`

**Purpose**:
- Refreshes stale LinkedIn post URLs in `linkedin_posts`
- Uses decrypted access tokens when available

**Notes**:
- Account type filters are normalized to stored LinkedIn entity types
- Use `-platform all` to run every refresher sequentially

## Services

- [LinkedIn Fetcher](/services/fetcher/linkedin-fetcher)
- [LinkedIn Parser](/services/parser/linkedin-parser)
- [LinkedIn Processor](/services/processor/linkedin-immediate-processor)
- [LinkedIn Sink](/services/sink/linkedin-clickhouse-sink)

## Data Flow

The LinkedIn pipeline processes:
- Company page posts
- Engagement metrics (reactions, comments, shares)
- Follower demographics
- Page analytics

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-linkedin` | Scheduled work orders |
| `immediate-work-order-linkedin` | On-demand API requests |
| `raw-linkedin-posts` | Raw post data from API |
| `parsed-linkedin-posts` | Processed post analytics |

## Configuration

Key environment variables for LinkedIn integration:
- `APP_LINKEDIN_CLIENT_ID`
- `APP_LINKEDIN_CLIENT_SECRET`
- `APP_LINKEDIN_ACCESS_TOKEN`
