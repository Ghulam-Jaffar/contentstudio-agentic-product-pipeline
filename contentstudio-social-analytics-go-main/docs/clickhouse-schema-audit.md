# ClickHouse Schema Audit — Partition Keys & Deduplication

**Audited against**: Production ClickHouse `136.243.18.44` (database `contentstudiobackend`)  
**Date**: 2026-04-29  
**ClickHouse version**: 24.8.14.39

---

## Overview

This document covers three categories of issues found during the audit:

1. **Epoch partition (`197001`)** — old Python-pipeline rows with zero timestamps stuck in an unqueryable partition, blocking `ReplacingMergeTree` deduplication.
2. **Wrong partition key on insights tables** — partitioned by fetch/insertion time instead of the data's own date, causing cross-month duplicate accumulation.
3. **Timestamp in ORDER BY** — `inserted_at` or `saving_time` included in the ORDER BY key, making `ReplacingMergeTree` deduplication impossible.
4. **No partition key + no dedup engine** — plain `MergeTree` tables with massive duplicate row counts.
5. **Go pipeline INSERT gap** — `facebook_video_insights` INSERT omits columns that GET queries read.

Tables confirmed as **correct** (partition key = actual data date, ORDER BY is stable) are listed at the end for completeness.

---

## Issue 1 — Epoch Partition `197001`

### Background

`toYYYYMM('1970-01-01') = 197001`. Any row where the partition key column contains a zero/epoch timestamp lands in the `197001` partition. `ReplacingMergeTree` only deduplicates within the same partition, so rows in `197001` can never be merged against rows in real monthly partitions.

All confirmed `197001` rows have **every** timestamp column set to `1970-01-01 00:00:00` — they are legacy data from the old Python pipeline with no recoverable date. Analytics GET queries exclude them via date filters, so they are invisible to users but waste storage and block deduplication.

### Affected Tables

| Table | Rows in `197001` |
|---|---|
| `linkedin_posts` | 302,485 |
| `linkedin_posts_old` | 274,661 |
| `mv_social_daily_metrics` | 22,608 |
| `facebook_posts` | 10,587 |
| `instagram_posts` | 9,449 |
| `facebook_insights` | 6,016 |
| `gmb_media_assets` | 5,404 |
| `tiktok_posts` | 2 |
| `facebook_media_assets` | 2 |

### Fix

Safe to drop — all rows have zero timestamps, they are unreachable by any analytics query.

```sql
ALTER TABLE linkedin_posts DROP PARTITION '197001';
ALTER TABLE linkedin_posts_old DROP PARTITION '197001';
ALTER TABLE facebook_posts DROP PARTITION '197001';
ALTER TABLE instagram_posts DROP PARTITION '197001';
ALTER TABLE facebook_insights DROP PARTITION '197001';
ALTER TABLE gmb_media_assets DROP PARTITION '197001';
ALTER TABLE tiktok_posts DROP PARTITION '197001';
ALTER TABLE facebook_media_assets DROP PARTITION '197001';
ALTER TABLE mv_social_daily_metrics DROP PARTITION '197001';
```

### Prevention

Add a guard in the Go pipeline before any INSERT to reject rows where the partition key column is zero:

```go
// Example for linkedin_posts (published_at is the partition key)
if post.PublishedAt.IsZero() {
    log.Warn().Str("post_id", post.PostID).Msg("skipping insert: published_at is zero")
    continue
}
```

---

## Issue 2 — Wrong Partition Key on Insights Tables

### Background

These tables are partitioned by the **fetch/insertion timestamp** (`saving_time`, `stored_event_at`, `inserted_at`) rather than the **date the insight data represents**. Because the partition key changes every time data is re-fetched in a new month, the same `record_id` can land in multiple monthly partitions. `ReplacingMergeTree` cannot deduplicate across partitions, so old copies are never removed. Storage grows indefinitely and queries may return stale aggregates.

### `facebook_insights`

| Property | Current | Correct |
|---|---|---|
| Partition key | `toYYYYMM(saving_time)` | `toYYYYMM(created_date)` |
| ORDER BY | `(page_id, hash_id)` | unchanged |
| Version | `updated_at` | unchanged |

`saving_time` = when we fetched the data. `created_date Date` already exists in the table and holds the actual insight date.

**Migration SQL:**

```sql
CREATE TABLE facebook_insights_new
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{shard}/contentstudiobackend/facebook_insights_v3',
    '{replica}',
    updated_at
)
PARTITION BY toYYYYMM(created_date)
ORDER BY (page_id, hash_id)
AS SELECT * FROM facebook_insights WHERE partition != '197001';

RENAME TABLE facebook_insights TO facebook_insights_backup2,
             facebook_insights_new TO facebook_insights;
```

---

### `instagram_insights`

| Property | Current | Correct |
|---|---|---|
| Partition key | `toYYYYMM(stored_event_at)` | `toYYYYMM(created_time)` |
| ORDER BY | `(instagram_id, record_id)` | unchanged |
| Version | none | add `updated_time` |

`stored_event_at` = when we stored it. `created_time DateTime64` already exists and holds the insight date.

**Migration SQL:**

```sql
CREATE TABLE instagram_insights_new
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{shard}/contentstudiobackend/instagram_insights_v2',
    '{replica}',
    updated_time
)
PARTITION BY toYYYYMM(created_time)
ORDER BY (instagram_id, record_id)
AS SELECT * FROM instagram_insights WHERE partition != '197001';

RENAME TABLE instagram_insights TO instagram_insights_backup,
             instagram_insights_new TO instagram_insights;
```

---

### `linkedin_insights`

| Property | Current | Correct |
|---|---|---|
| Partition key | `toYYYYMM(inserted_at)` | `toYYYYMM(created_at)` |
| ORDER BY | `(linkedin_id, record_id)` | unchanged |
| Version | `inserted_at` | unchanged |

`inserted_at` = insertion time. `created_at DateTime64` already exists and holds the actual LinkedIn insight date.

**Migration SQL:**

```sql
CREATE TABLE linkedin_insights_new
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{shard}/contentstudiobackend/linkedin_insights_v3',
    '{replica}',
    inserted_at
)
PARTITION BY toYYYYMM(created_at)
ORDER BY (linkedin_id, record_id)
AS SELECT * FROM linkedin_insights;

RENAME TABLE linkedin_insights TO linkedin_insights_backup2,
             linkedin_insights_new TO linkedin_insights;
```

---

### `facebook_competitor_insights`

| Property | Current | Correct |
|---|---|---|
| Partition key | `toYYYYMM(inserted_at)` | `toYYYYMM(<data_date_column>)` |
| ORDER BY | `(record_id, page_id)` | unchanged |
| Version | **none** | add appropriate version column |

No version column defined — newest insert does not reliably win. Confirm which column holds the insight's own date before migrating.

---

### `instagram_competitor_insights`

| Property | Current | Correct |
|---|---|---|
| Partition key | `toYYYYMM(inserted_at)` | `toYYYYMM(<data_date_column>)` |
| ORDER BY | `record_id` | unchanged |
| Version | **none** | add appropriate version column |

Same situation as `facebook_competitor_insights` — confirm the data date column first.

---

## Issue 3 — Timestamp in ORDER BY (No Deduplication Possible)

### Background

`ReplacingMergeTree` deduplicates rows that share the **same ORDER BY key**. When `inserted_at` or `saving_time` is included in the ORDER BY, every insert produces a unique key — deduplication never fires. These tables accumulate rows from every fetch run forever.

### `youtube_videos` — 28M rows / 1.3 GiB

`published_at` (stable YouTube publish date) already exists and is perfect for the partition key.

| Property | Current | Correct |
|---|---|---|
| Partition key | `toYYYYMM(inserted_at)` | `toYYYYMM(published_at)` |
| ORDER BY | `(video_id, channel_id, inserted_at)` | `(video_id, channel_id)` |
| Version | none | `inserted_at` |

**Migration SQL:**

```sql
CREATE TABLE youtube_videos_new
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{shard}/contentstudiobackend/youtube_videos_v2',
    '{replica}',
    inserted_at
)
PARTITION BY toYYYYMM(published_at)
ORDER BY (video_id, channel_id)
AS SELECT * FROM youtube_videos;

RENAME TABLE youtube_videos TO youtube_videos_backup,
             youtube_videos_new TO youtube_videos;
```

---

### `youtube_channels` — 2M rows / 406 MiB

Daily snapshot of channel stats. `created_at` is the snapshot date and is stable per record.

| Property | Current | Correct |
|---|---|---|
| Partition key | `toYYYYMM(inserted_at)` | `toYYYYMM(created_at)` |
| ORDER BY | `(record_id, channel_id, inserted_at)` | `(record_id, channel_id)` |
| Version | none | `inserted_at` |

**Migration SQL:**

```sql
CREATE TABLE youtube_channels_new
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{shard}/contentstudiobackend/youtube_channels_v2',
    '{replica}',
    inserted_at
)
PARTITION BY toYYYYMM(created_at)
ORDER BY (record_id, channel_id)
AS SELECT * FROM youtube_channels;

RENAME TABLE youtube_channels TO youtube_channels_backup,
             youtube_channels_new TO youtube_channels;
```

---

### `youtube_shared_insights` — 2M rows / 59 MiB

No date column other than `inserted_at` — keep `inserted_at` as partition but remove from ORDER BY.

| Property | Current | Correct |
|---|---|---|
| Partition key | `toYYYYMM(inserted_at)` | unchanged |
| ORDER BY | `(record_id, channel_id, inserted_at)` | `(record_id, channel_id)` |
| Version | none | `inserted_at` |

**Migration SQL:**

```sql
CREATE TABLE youtube_shared_insights_new
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{shard}/contentstudiobackend/youtube_shared_insights_v2',
    '{replica}',
    inserted_at
)
PARTITION BY toYYYYMM(inserted_at)
ORDER BY (record_id, channel_id)
AS SELECT * FROM youtube_shared_insights;

RENAME TABLE youtube_shared_insights TO youtube_shared_insights_backup,
             youtube_shared_insights_new TO youtube_shared_insights;
```

---

### `tiktok_insights` — 2.9M rows / 213 MiB

Account-level daily snapshot. Only `inserted_at` available as a date.

| Property | Current | Correct |
|---|---|---|
| Partition key | `toYYYYMM(inserted_at)` | unchanged |
| ORDER BY | `(tiktok_id, record_id, inserted_at)` | `(tiktok_id, record_id)` |
| Version | none | `inserted_at` |

**Migration SQL:**

```sql
CREATE TABLE tiktok_insights_new
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{shard}/contentstudiobackend/tiktok_insights_v2',
    '{replica}',
    inserted_at
)
PARTITION BY toYYYYMM(inserted_at)
ORDER BY (tiktok_id, record_id)
AS SELECT * FROM tiktok_insights;

RENAME TABLE tiktok_insights TO tiktok_insights_backup,
             tiktok_insights_new TO tiktok_insights;
```

---

### `twitter_insights` — 3.9K rows (small, low priority)

| Property | Current | Correct |
|---|---|---|
| Partition key | `toYYYYMM(saving_time)` | unchanged |
| ORDER BY | `(record_id, saving_time)` | `(twitter_id, record_id)` |
| Version | none | `saving_time` |

**Migration SQL:**

```sql
CREATE TABLE twitter_insights_new
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{shard}/contentstudiobackend/twitter_insights_v2',
    '{replica}',
    saving_time
)
PARTITION BY toYYYYMM(saving_time)
ORDER BY (twitter_id, record_id)
AS SELECT * FROM twitter_insights;

RENAME TABLE twitter_insights TO twitter_insights_backup,
             twitter_insights_new TO twitter_insights;
```

---

### `pinterest_users` — 105K rows / 1.9 MiB (low priority)

| Property | Current | Correct |
|---|---|---|
| Partition key | `toYYYYMM(inserted_at)` | unchanged |
| ORDER BY | `(user_id, inserted_at)` | `(user_id)` |
| Version | none | `inserted_at` |

**Migration SQL:**

```sql
CREATE TABLE pinterest_users_new
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{shard}/contentstudiobackend/pinterest_users_v2',
    '{replica}',
    inserted_at
)
PARTITION BY toYYYYMM(inserted_at)
ORDER BY (user_id)
AS SELECT * FROM pinterest_users;

RENAME TABLE pinterest_users TO pinterest_users_backup,
             pinterest_users_new TO pinterest_users;
```

---

## Issue 4 — No Partition Key + No Deduplication Engine

All four tables use plain `ReplicatedMergeTree` — no version-based deduplication exists at all. Without `ReplacingMergeTree`, ClickHouse never removes duplicate rows regardless of the ORDER BY definition.

| Table | Total Rows | Unique Keys | Duplicates | Duplicate % |
|---|---|---|---|---|
| `tmp_pin_board_map` | 19,911,933 | 6,636,720 | **13,275,213** | **66%** |
| `posts_cluster_feeds` | 6,421,695 | 4,017,778 | **2,403,917** | **37%** |
| `curated_topics` | 20,049 | 19,982 | 67 | <1% |
| `domain_rankings` | 0 | — | 0 | — |

---

### `tmp_pin_board_map` — 13.3M duplicates (66%)

Columns: `pin_id`, `board_id`, `user_id` — no date column exists.  
This is a lookup/staging table (`tmp` prefix). The right fix is a one-time dedup + engine change.

**One-time cleanup:**

```sql
CREATE TABLE tmp_pin_board_map_new
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{shard}/contentstudiobackend/tmp_pin_board_map_v2',
    '{replica}'
)
ORDER BY (pin_id, user_id)
AS SELECT DISTINCT pin_id, board_id, user_id FROM tmp_pin_board_map;

RENAME TABLE tmp_pin_board_map TO tmp_pin_board_map_old,
             tmp_pin_board_map_new TO tmp_pin_board_map;
DROP TABLE tmp_pin_board_map_old;
```

No partition key is acceptable here since the table has no date column. With `ReplacingMergeTree` and all rows in a single partition, deduplication will work correctly.

---

### `posts_cluster_feeds` — 2.4M duplicates (37%)

Columns: `id`, `rss_id`, `url`, `netloc`, `domain_rank`, `average_engagement_6_months`, `last_crawled_at Nullable(DateTime)`

`last_crawled_at` is nullable so requires `coalesce` for use as a partition key.

**Migration SQL:**

```sql
CREATE TABLE posts_cluster_feeds_new
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{shard}/contentstudiobackend/posts_cluster_feeds_v2',
    '{replica}',
    last_crawled_at
)
PARTITION BY toYYYYMM(coalesce(last_crawled_at, now()))
ORDER BY id
AS SELECT * FROM posts_cluster_feeds;

-- Run FINAL to deduplicate during the copy:
-- INSERT INTO posts_cluster_feeds_new SELECT * FROM posts_cluster_feeds FINAL;

RENAME TABLE posts_cluster_feeds TO posts_cluster_feeds_old,
             posts_cluster_feeds_new TO posts_cluster_feeds;
DROP TABLE posts_cluster_feeds_old;
```

---

### `curated_topics` — 67 duplicates

Minor. Switch engine to `ReplicatedReplacingMergeTree`. ORDER BY `(name, label)` is already the natural dedup key.

```sql
CREATE TABLE curated_topics_new
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{shard}/contentstudiobackend/curated_topics_v2',
    '{replica}'
)
ORDER BY (name, label)
AS SELECT * FROM curated_topics FINAL;

RENAME TABLE curated_topics TO curated_topics_old,
             curated_topics_new TO curated_topics;
DROP TABLE curated_topics_old;
```

---

## Issue 5 — Go Pipeline INSERT Gap: `facebook_video_insights`

### Background

`facebook_video_insights` uses `ReplacingMergeTree(updated_at)`. When the Go pipeline inserts a row for an existing video, it replaces the Python pipeline's row. However, the Go INSERT currently omits 3 columns that analytics GET queries actively read, so those metrics silently drop to 0 after replacement.

### Missing Columns

| Column | Used By |
|---|---|
| `total_video_views_organic` | `GetVideoInsights`, `GetVideoRollup` |
| `total_video_views_paid` | `GetVideoInsights`, `GetVideoRollup` |
| `total_video_view_total_time_paid` | `GetVideoInsights`, `GetVideoRollup` |

### Location

File: `src/db/clickhouse/facebook.go`, function `BulkInsertVideoInsights`

### Fix

Add the three columns to both the INSERT column list and the `batch.Append(...)` call. The values must be populated from the Facebook Graph API response in the parser (`src/internal/parsing/facebook_parser.go`).

---

## High-Part-Count Tables (Merge Backlog)

Background merges are falling behind on these tables. Run `OPTIMIZE TABLE ... FINAL` during low-traffic hours to force a full merge.

| Table | Active Parts | Recommendation |
|---|---|---|
| `facebook_competitor_media_assets` | 772 | `OPTIMIZE TABLE facebook_competitor_media_assets FINAL` |
| `facebook_media_assets` | 679 | `OPTIMIZE TABLE facebook_media_assets FINAL` |
| `facebook_posts` | 659 | `OPTIMIZE TABLE facebook_posts FINAL` |
| `mv_social_daily_metrics` | 549 | `OPTIMIZE TABLE mv_social_daily_metrics FINAL` |

> `mv_social_daily_metrics` already has a custom `parts_to_throw_insert = 8000` setting to avoid insert errors. The `OPTIMIZE FINAL` is still needed to reduce part count.

---

## Recommended Execution Order

| Step | Action | Risk | Impact |
|---|---|---|---|
| 1 | Drop `197001` partitions (Issue 1) | Low — data is invisible to queries | Unblocks dedup in affected tables |
| 2 | Migrate `facebook_insights` (Issue 2) | Medium — table recreation | Fixes largest insights table (15M rows) |
| 3 | Migrate `linkedin_insights` (Issue 2) | Medium | Fixes cross-month LinkedIn duplicates |
| 4 | Migrate `instagram_insights` (Issue 2) | Medium | Fixes cross-month Instagram duplicates |
| 5 | Migrate `youtube_videos` (Issue 3) | Medium — 28M row copy | Biggest storage win, fixes stable partition key |
| 6 | Migrate `youtube_channels`, `youtube_shared_insights`, `tiktok_insights` (Issue 3) | Medium | Enables dedup on snapshot tables |
| 7 | Deduplicate `tmp_pin_board_map` (Issue 4) | Low | Removes 13.3M wasted rows |
| 8 | Migrate `posts_cluster_feeds` (Issue 4) | Medium | Removes 2.4M wasted rows |
| 9 | Fix Go INSERT for `facebook_video_insights` (Issue 5) | Low — code change only | Restores organic/paid video metrics |
| 10 | `OPTIMIZE TABLE FINAL` on high-part-count tables | Low | Reduces merge pressure |
| 11 | Migrate `twitter_insights`, `pinterest_users` (Issue 3) | Low — small tables | Minor cleanup |
| 12 | Investigate and migrate `facebook/instagram_competitor_insights` (Issue 2) | Medium | Requires confirming data date column first |

---

## Tables Confirmed Correct

These tables have the partition key set to the actual content/data date and a stable ORDER BY — no changes needed.

| Table | Partition Key | ORDER BY |
|---|---|---|
| `facebook_posts` | `toYYYYMM(created_time)` | `(page_id, post_id)` |
| `facebook_video_insights` | `toYYYYMM(created_time)` | `(page_id, post_id, video_id)` |
| `facebook_reels_insights` | `toYYYYMM(created_at)` | `(page_id, post_id)` |
| `facebook_media_assets` | `toYYYYMM(created_at)` | `(page_id, post_id, media_id)` |
| `facebook_competitor_posts` | `toYYYYMM(created_at)` | `(facebook_id, post_id)` |
| `facebook_competitor_media_assets` | `toYYYYMM(created_at)` | `(post_id, media_id)` |
| `instagram_posts` | `toYYYYMM(post_created_at)` | `(instagram_id, media_id, post_created_at)` |
| `instagram_competitor_posts` | `toYYYYMM(created_at)` | `(post_id, created_at, ...)` |
| `linkedin_posts` | `toYYYYMM(published_at)` | `(linkedin_id, post_id)` |
| `tiktok_posts` | `toYYYYMM(created_at)` | `(post_id, created_at)` |
| `twitter_posts` | `toYYYYMM(tweeted_at)` | `(tweet_id, tweeted_at)` |
| `pinterest_pins` | `toYYYYMM(created_at)` | `(pin_id, created_at)` |
| `pinterest_pin_insights` | `toYYYYMM(created_at)` | `(record_id, pin_id, created_at)` |
| `pinterest_boards` | `toYYYYMM(created_at)` | `(record_id, created_at)` |
| `pinterest_user_insights` | `toYYYYMM(created_at)` | `(record_id, created_at)` |
| `youtube_activity_insights` | `toYYYYMM(created_at)` | `(record_id, channel_id, created_at)` |
| `youtube_traffic_insights` | `toYYYYMM(created_at)` | `(record_id, channel_id, created_at)` |
| `gmb_daily_metrics` | `toYYYYMM(created_at)` | `(gmb_id, created_at)` |
| `gmb_search_keywords_monthly` | `toYYYYMM(keyword_month)` | `(gmb_id, keyword_month, keyword)` |
| `gmb_local_posts` | `toYYYYMM(created_at)` | `(gmb_id, post_name)` |
| `gmb_media_assets` | `toYYYYMM(created_at)` | `(gmb_id, media_name)` |
| `gmb_reviews` | `toYYYYMM(created_at)` | `(gmb_id, review_id)` |
| `mv_social_daily_metrics` | `toYYYYMM(date)` | `(date, platform, account_id)` |
| `mv_social_daily_followers` | `toYYYYMM(date)` | `(date, platform, account_id)` |
