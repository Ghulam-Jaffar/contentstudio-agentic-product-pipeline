---
title: Pinterest Integration
description: Pinterest data pipeline and analytics
---

Documentation for Pinterest social media analytics integration.

## Overview

The Pinterest integration pipeline follows the standard five-stage architecture:

1. **Scheduler**: Unified Account Fetcher
2. **Fetcher**: Pinterest Fetcher Service
3. **Parser**: Pinterest Parser
4. **Processor**: Pinterest Immediate Processor
5. **Sink**: Pinterest Analytics Sink (merged parser+sink)

## Account Types

Pinterest supports two account types:

| Account Type | Description | Board Handling |
|-------------|-------------|----------------|
| **Profile** | Fetches all public boards and their pins | Iterates all boards, skips private |
| **Board** | Fetches a specific board and its pins | Single board via `board_id` |

## Data Fetch Periods

| Sync Type | Data Period | Page Size | Max Pages | Description |
|-----------|-------------|-----------|-----------|-------------|
| **Incremental Sync** | 7 days | 100 | 2 | Regular scheduled updates |
| **Full Sync** | 86 days | 250 | Unlimited | Complete data refresh |
| **Immediate Sync** | 86 days | 250 | Unlimited | On-demand via API |

## Scheduler Configuration

| Parameter | Value | Description |
|-----------|-------|-------------|
| Update Interval | 6 hours | Time between scheduled analytics updates |
| MongoDB Fetch Size | 50 | Accounts fetched per MongoDB query |
| Kafka Batch Size | 200 | Accounts per work order message |

## Services

- [Pinterest Fetcher](/services/fetcher/pinterest-fetcher)
- [Pinterest Parser](/services/parser/pinterest-parser)
- [Pinterest Immediate Processor](/services/processor/pinterest-immediate-processor)
- [Pinterest Analytics Sink](/services/sink/pinterest-analytics-sink)

## Data Flow

The Pinterest pipeline processes 5 data types:
- **Users** — Profile snapshots (followers, monthly views, board/pin counts)
- **Boards** — Board metadata (name, privacy, pin count, collaborators)
- **Pins** — Pin metadata with media details (images, videos, dimensions)
- **Pin Insights** — Daily pin-level analytics (28 metrics)
- **User Insights** — Daily user-level analytics (24 metrics)

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-pinterest` | Scheduled batch work orders |
| `immediate-work-order-pinterest` | On-demand API requests |
| `raw-pinterest-users` | Raw user profile data |
| `raw-pinterest-boards` | Raw board metadata |
| `raw-pinterest-pins` | Raw pin metadata |
| `raw-pinterest-pin-insights` | Raw daily pin analytics |
| `raw-pinterest-user-insights` | Raw daily user analytics |
| `parsed-pinterest-users` | Parsed user data |
| `parsed-pinterest-boards` | Parsed board data |
| `parsed-pinterest-pins` | Parsed pin data |
| `parsed-pinterest-pin-insights` | Parsed pin insights |
| `parsed-pinterest-user-insights` | Parsed user insights |

## ClickHouse Tables

| Table | Description |
|-------|-------------|
| `pinterest_users` | User profile snapshots |
| `pinterest_boards` | Board metadata |
| `pinterest_pins` | Pin metadata with media |
| `pinterest_pin_insights` | Daily pin analytics (28 metrics) |
| `pinterest_user_insights` | Daily user analytics (24 metrics) |

## Analytics Metrics

### Pin-Level Metrics
- **Engagement**: pin_clicks, clickthrough, saves, outbound_click, engagement
- **Rates**: clickthrough_rate, engagement_rate, save_rate
- **Video**: video_mrc_view, video_start, video_10s_view, video_avg_watch_time, video_v50_watch_time, full_screen_play, full_screen_playtime
- **Interaction**: impression, profile_visit, closeup, user_follow, quartile_95s_percent_view

### User-Level Metrics
- Same as pin-level with `pin_click_rate` (replaces `user_follow`)

## Configuration

Key environment variables for Pinterest integration:
- `APP_DECRYPTION_KEY` — Token decryption key
- `APP_MONGO_*` — MongoDB connection
- `APP_KAFKA_*` — Kafka brokers and SASL auth
- `APP_CLICKHOUSE_*` — ClickHouse connection

No Pinterest-specific API credentials are required (uses per-account OAuth2 tokens stored in MongoDB).
