---
title: Google My Business Integration
description: GMB data pipeline and analytics for location-level business performance
---

Documentation for Google My Business (Google Business Profile) analytics integration.

## Overview

The GMB integration uses the Google Business Profile Performance API (v1) and My Business API (v4) with OAuth2 Bearer token authentication. It provides location-level analytics covering 5 data types: performance metrics, search keywords, local posts, reviews, and media assets.

A key prerequisite is **Voice of Merchant (VoM)** verification — performance metrics and search keywords are only fetched for locations with verified VoM status.

Unlike other platforms, GMB uses a **2-service architecture** (Fetcher + merged Analytics Sink) instead of the standard 5-stage pipeline, plus integration with the Unified Immediate Processor for on-demand processing.

## Data Fetch Periods

| Sync Type | Data Period | Description |
|-----------|-------------|-------------|
| **Incremental Sync** | 3 months | Rolling 3-month window for performance metrics and keywords |
| **Immediate Sync** | 3 months | Same window, triggered on-demand via API |
| **Local Posts** | All time | All posts fetched regardless of sync type |
| **Reviews** | 2 pages | Up to 2 pages ordered by update time |
| **Media Assets** | All time | Current snapshot of all media |

## Scheduler Configuration

| Parameter | Value | Description |
|-----------|-------|-------------|
| Update Interval | 6 hours | Time between scheduled analytics updates |
| Batch Size | 200 | Number of accounts processed per batch |
| Scale Factor | 1000 | MongoDB fetch cursor size multiplier |

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Unified       │    │  GMB             │    │  GMB Analytics  │    │                 │
│   Account       │───▶│  Fetcher         │───▶│  Sink           │───▶│  ClickHouse     │
│   Fetcher       │    │  (10 workers)    │    │  (Parser+Sink)  │    │  Database        │
└─────────────────┘    └──────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │                       │
         ▼                       ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ work-order-     │    │ raw-gmb-data     │    │ 5 parsers +     │    │ 5 ClickHouse    │
│ gmb-batch       │    │ (5 data types)   │    │ 15 batch procs  │    │ tables          │
└─────────────────┘    └──────────────────┘    └─────────────────┘    └─────────────────┘
```

**Immediate Processing Path:**

```
API Request → immediate-work-order-gmb → Unified Immediate Processor
                                              → Direct ClickHouse Insert
                                              → Pusher + Email Notifications
```

## Complete Data Pipeline Flow

```
SCHEDULING
   Unified Account Fetcher → work-order-gmb-batch (200 accounts per batch)

EXTRACTION
   work-order-gmb-batch → GMB Fetcher → {
     Token Refresh → VoM Check → Fetch 5 data types → raw-gmb-data
   }

PARSE + STORE (Merged)
   raw-gmb-data → GMB Analytics Sink → {
     5 parser workers route by data_type,
     15 batch processors (3 per type),
     10,000-item batches or 10-second timeout
   } → ClickHouse

IMMEDIATE PROCESSING
   immediate-work-order-gmb → Unified Immediate Processor → {
     Token Refresh → VoM Check → Fetch all → Direct ClickHouse Insert
     → Pusher notification → Email notification (new accounts only)
   }
```

## Components

### 1. Account Fetcher (Scheduler)

**Service**: Unified Account Fetcher
**Location**: `src/cmd/jobs/fetcher/gmb.go`

**Purpose**: Queries MongoDB for GMB accounts needing analytics updates and publishes batch work orders.

**Responsibilities**:
- Queries `social_integrations` where `platform_type = "gmb"`, `validity = "valid"`, `state in ["Added","Syncing","Processed"]`
- Parses platform identifier (`accounts/{id}/locations/{id}`) to extract account and location IDs
- Batches up to 200 accounts per Kafka message
- Publishes to `work-order-gmb-batch`

### 2. GMB Fetcher

**Service**: `gmb-fetcher`
**Location**: `src/services/gmb/gmb-fetcher/`

**Purpose**: Consumes batch work orders and calls Google APIs to fetch raw data.

**Key Features**:
- 10 concurrent worker goroutines
- Per-location processing semaphore
- Token refresh with MongoDB update
- VoM status check gates performance metrics and search keywords
- 5-minute idle timeout for graceful shutdown
- Produces raw data to `raw-gmb-data` topic

### 3. GMB Analytics Sink (Merged Parser + Sink)

**Service**: `gmb-analytics-sink`
**Location**: `src/services/gmb/gmb-analytics-sink/`

**Purpose**: Consumes raw GMB data, parses inline by data type, and batch-inserts into ClickHouse.

**Architecture**:
```
Kafka Consumer → 5 Parser Workers → Batch Collectors → 15 Batch Processors → ClickHouse
                                          ↓                    ↓
                                  Channel Buffers        Parallel Processing
                                   (50K capacity)        (3 per data type)
```

**Batch Configuration**:
- Max batch size: 10,000 items
- Batch timeout: 10 seconds
- Channel buffer: 50,000 messages
- 5 parser workers route messages by `data_type`
- 15 batch processors (3 per type: daily_metrics, media_assets, search_keywords, local_posts, reviews)
- 5-minute idle timeout for graceful shutdown

### 4. Unified Immediate Processor (Library)

**Location**: `src/services/gmb/gmb-immediate-processor/processor/`

**Purpose**: On-demand processing triggered by API requests. Used as a library by the unified immediate processor.

**Processing per account**:
1. Decrypt and refresh OAuth2 tokens
2. Check VoM status
3. Fetch performance metrics (3-month rolling, if VoM)
4. Fetch search keywords (3-month rolling, if VoM)
5. Fetch local posts (all)
6. Fetch reviews (2 pages)
7. Fetch media assets (all)
8. Direct ClickHouse insert (no Kafka)
9. Update MongoDB state
10. Send Pusher + email notifications

## Voice of Merchant (VoM)

| VoM Status | Available Data Types |
|------------|---------------------|
| `hasVoiceOfMerchant = true` | Performance metrics, search keywords, local posts, reviews, media assets |
| `hasVoiceOfMerchant = false` | Local posts, reviews, media assets only |
| VoM check fails | Local posts, reviews, media assets only (warning logged) |

VoM status is persisted to MongoDB (`has_voice_of_merchant` field) after each check.

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

## Kafka Topics

| Topic | Purpose |
|-------|---------|
| `work-order-gmb-batch` | Batch work order messages (200 accounts max) |
| `immediate-work-order-gmb` | On-demand processing requests |
| `raw-gmb-data` | Raw GMB data (5 types) for sink processing |

## Configuration

Key environment variables for GMB integration:

```bash
# Google OAuth2 Configuration
APP_GMB_CLIENT_ID="xxx.apps.googleusercontent.com"
APP_GMB_CLIENT_SECRET="GOCSPX-xxx"

# Token Encryption
APP_DECRYPTION_KEY="base64_encoded_key"

# ClickHouse Configuration
APP_CLICKHOUSE_HOST="localhost:9000"
APP_CLICKHOUSE_DATABASE="analytics"
APP_CLICKHOUSE_USERNAME="default"
APP_CLICKHOUSE_PASSWORD="password"

# Kafka Configuration
APP_KAFKA_BROKERS="localhost:9092"
APP_KAFKA_SASL_USERNAME="username"
APP_KAFKA_SASL_PASSWORD="password"
APP_KAFKA_SASL_MECHANISM="SCRAM-SHA-256"
```

## Error Handling

- **Token refresh failure**: Skip account entirely (critical error)
- **VoM check failure**: Log warning, proceed with non-VoM data (posts, reviews, media)
- **Individual data type failure**: Log warning, continue to next data type
- **ClickHouse insert failure**: Log warning, continue
- **Key Principle**: A single failing data type never blocks other data types
