---
title: TikTok Integration
description: TikTok data pipeline and analytics
---

Documentation for TikTok social media analytics integration.

## Overview

The TikTok integration pipeline follows the standard five-stage architecture:

1. **Scheduler**: Unified Account Fetcher
2. **Fetcher**: TikTok Fetcher Service
3. **Parser**: TikTok Parser
4. **Processor**: TikTok Immediate Processor
5. **Sink**: TikTok ClickHouse Sink

## Data Fetch Periods

| Sync Type | Data Period | Max Videos | Description |
|-----------|-------------|------------|-------------|
| **Incremental Sync** | 14 days | 999 | Regular scheduled updates fetch videos from last 14 days (max 999 videos) |
| **Full Sync** | Unlimited | Unlimited | Complete data refresh fetches all available videos |

## Scheduler Configuration

| Parameter | Value | Description |
|-----------|-------|-------------|
| Update Interval | 6 hours | Time between scheduled analytics updates |
| Batch Size | 200 | Number of accounts processed per batch |

## Services

- [TikTok Fetcher](/services/fetcher/tiktok-fetcher)
- [TikTok Parser](/services/parser/tiktok-parser)
- [TikTok Processor](/services/processor/tiktok-immediate-processor)
- [TikTok Sink](/services/sink/tiktok-clickhouse-sink)

## Data Flow

The TikTok pipeline processes:
- Video posts and their metrics
- Engagement data (likes, comments, shares, views)
- Hashtag performance
- Account analytics

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-tiktok` | Scheduled work orders |
| `immediate-work-order-tiktok` | On-demand API requests |
| `raw-tiktok-posts` | Raw video data from API |
| `parsed-tiktok-posts` | Processed video analytics |

## Configuration

Key environment variables for TikTok integration:
- `APP_TIKTOK_CLIENT_KEY`
- `APP_TIKTOK_CLIENT_SECRET`
- `APP_TIKTOK_ACCESS_TOKEN`
