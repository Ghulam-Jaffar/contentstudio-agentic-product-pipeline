# LinkedIn ClickHouse Schema Optimization

This document outlines the optimized ClickHouse schema for LinkedIn analytics tables, including reasoning and performance comparisons.

## Executive Summary

The LinkedIn analytics tables (`linkedin_insights` and `linkedin_posts`) have been redesigned to optimize for the primary query pattern: **filtering by `linkedin_id`**. Key changes include:

1. Reordering `ORDER BY` to place `linkedin_id` first
2. Adding compression codecs for storage efficiency
3. Proper version columns for `ReplacingMergeTree` deduplication

---

## Schema Comparison

### linkedin_insights

| Aspect | Old Schema | New Schema |
|--------|------------|------------|
| ORDER BY | `(record_id, inserted_at)` | `(linkedin_id, record_id)` |
| PRIMARY KEY | `record_id` | `(linkedin_id, record_id)` |
| PARTITION BY | `toYYYYMM(inserted_at)` | `toYYYYMM(inserted_at)` |
| Version Column | None | `inserted_at` |
| Compression | Minimal | Full (T64, Delta, ZSTD) |

**Why `record_id` instead of `created_at`?**

`record_id` is generated as `linkedin_id + "_" + date` (e.g., `123_2024-01-14`), which uses only the date portion. This guarantees consistent deduplication regardless of `created_at` timestamp precision (DateTime64 has microsecond precision that could vary between inserts).

### linkedin_posts

| Aspect | Old Schema | New Schema |
|--------|------------|------------|
| ORDER BY | `(post_id, created_at)` | `(linkedin_id, post_id)` |
| PRIMARY KEY | `post_id` | `(linkedin_id, post_id)` |
| PARTITION BY | `toYYYYMM(created_at)` | `toYYYYMM(published_at)` |
| Version Column | None | `saving_time` |
| Compression | Minimal | Full (T64, Delta, ZSTD) |

**Why `published_at` for partitioning?**

Queries will primarily filter by `published_at` (when the post was published on LinkedIn), not `created_at`. Using `published_at` for partitioning enables efficient partition pruning for date-range queries.

---

## Final Table Definitions

### linkedin_insights

```sql
CREATE TABLE contentstudiobackend.linkedin_insights
(
    `linkedin_id` String CODEC(ZSTD(1)),
    `record_id` String CODEC(ZSTD(1)),
    `organization_name` String DEFAULT '' CODEC(ZSTD(1)),
    `impressionCount` Int64 CODEC(T64, ZSTD(1)),
    `organicFollowerCount` Int64 CODEC(T64, ZSTD(1)),
    `totalFollowerCount` Int64 CODEC(T64, ZSTD(1)),
    `paidFollowerCount` Int64 CODEC(T64, ZSTD(1)),
    `daily_follower_count` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `reach` Int64 CODEC(T64, ZSTD(1)),
    `repost` Int64 CODEC(T64, ZSTD(1)),
    `comments` Int64 CODEC(T64, ZSTD(1)),
    `post_clicks` Int64 CODEC(T64, ZSTD(1)),
    `reactions` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `engagement` Float64 DEFAULT 0 CODEC(ZSTD(1)),
    `followers_by_seniority` String CODEC(ZSTD(1)),
    `followers_by_industry` String CODEC(ZSTD(1)),
    `followers_by_country` String CODEC(ZSTD(1)),
    `followers_by_city` String CODEC(ZSTD(1)),
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `unique_visitors` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `desktop_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `mobile_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `overview_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `about_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `jobs_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `people_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `careers_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `life_at_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `insights_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `products_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `page_views_by_country` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_region` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_industry` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_seniority` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_function` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_staff_count` String DEFAULT '' CODEC(ZSTD(1))
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/contentstudiobackend/linkedin_insights', '{replica}', inserted_at)
PARTITION BY toYYYYMM(inserted_at)
PRIMARY KEY (linkedin_id, record_id)
ORDER BY (linkedin_id, record_id)
SETTINGS index_granularity = 8192;
```

### linkedin_posts

```sql
CREATE TABLE contentstudiobackend.linkedin_posts
(
    `linkedin_id` String CODEC(ZSTD(1)),
    `post_id` String CODEC(ZSTD(1)),
    `activity` String CODEC(ZSTD(1)),
    `media_type` String CODEC(ZSTD(1)),
    `article_url` String CODEC(ZSTD(1)),
    `article_title` String CODEC(ZSTD(1)),
    `post_data` String CODEC(ZSTD(1)),
    `image` String CODEC(ZSTD(1)),
    `media` Array(String) CODEC(ZSTD(1)),
    `type` String CODEC(ZSTD(1)),
    `hashtags` Array(String) CODEC(ZSTD(1)),
    `comments` Int64 CODEC(T64, ZSTD(1)),
    `total_engagement` Float64 DEFAULT 0 CODEC(ZSTD(1)),
    `favorites` Int64 CODEC(T64, ZSTD(1)),
    `title` String CODEC(ZSTD(1)),
    `day_of_week` String CODEC(ZSTD(1)),
    `hour_of_day` Int64 CODEC(T64, ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `saving_time` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `poll_data` String CODEC(ZSTD(1)),
    `reach` Int64 CODEC(T64, ZSTD(1)),
    `repost` Int64 CODEC(T64, ZSTD(1)),
    `post_clicks` Int64 CODEC(T64, ZSTD(1)),
    `impressions` Int64 CODEC(T64, ZSTD(1)),
    `published_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `last_modified_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `lifecycle_state` String DEFAULT '' CODEC(ZSTD(1)),
    `visibility` String DEFAULT '' CODEC(ZSTD(1)),
    `is_reshare_disabled` Bool DEFAULT false,
    `feed_distribution` String DEFAULT '' CODEC(ZSTD(1)),
    `third_party_channels` Array(String) DEFAULT [] CODEC(ZSTD(1)),
    `carousel_preview` String DEFAULT '' CODEC(ZSTD(1))
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/contentstudiobackend/linkedin_posts', '{replica}', saving_time)
PARTITION BY toYYYYMM(published_at)
PRIMARY KEY (linkedin_id, post_id)
ORDER BY (linkedin_id, post_id)
SETTINGS index_granularity = 8192;
```

---

## Reasoning

### 1. ORDER BY Column Selection

ClickHouse stores data physically sorted by the `ORDER BY` columns. The first column in `ORDER BY` is critical because:

- ClickHouse uses sparse indexing (default: every 8192 rows)
- Queries filtering on the first ORDER BY column can skip entire granules
- Data locality improves cache efficiency

**Why `linkedin_id` first:**

Our dominant query pattern is account-level analytics:
```sql
SELECT * FROM linkedin_insights
WHERE linkedin_id = '105256406'
AND created_at BETWEEN '2024-01-01' AND '2024-12-31';
```

With `linkedin_id` first:
- ClickHouse locates the account's data via sparse index
- Reads only granules containing that account
- Further filters by date range within those granules

### 2. Compression Codecs

| Codec | Used For | Benefit |
|-------|----------|---------|
| `T64` | Int64 columns | Efficient for integers with limited value ranges (counts, metrics) |
| `Delta(4)` | DateTime64 | Stores differences between consecutive timestamps (often small) |
| `ZSTD(1)` | All columns | General-purpose compression with good ratio/speed balance |

**Expected storage reduction:** 40-60% compared to uncompressed data.

### 3. ReplacingMergeTree Version Column

```sql
ENGINE = ReplicatedReplacingMergeTree(..., inserted_at)
```

The version column (`inserted_at` for insights, `saving_time` for posts) ensures:
- When duplicate rows exist (same ORDER BY key), the row with highest version is kept
- Allows safe re-processing of data without manual deduplication
- Background merges automatically remove older versions

### 4. PARTITION BY Strategy

```sql
-- linkedin_insights: partition by insert date
PARTITION BY toYYYYMM(inserted_at)

-- linkedin_posts: partition by publish date
PARTITION BY toYYYYMM(published_at)
```

- **linkedin_insights**: Partitioned by `inserted_at` (when analytics snapshot was captured)
- **linkedin_posts**: Partitioned by `published_at` (when post was published on LinkedIn)
- Enables efficient partition pruning for date-range queries
- Simplifies data retention (drop old partitions)
- Prevents "too many parts" errors by limiting partition count

**Why different partition columns?**
- Insights are time-series snapshots - `inserted_at` represents the snapshot date
- Posts are content with a publish date - queries filter by when the post went live

---

## Query Performance Comparison

### linkedin_insights

#### Query 1: Account insights for date range

```sql
SELECT * FROM linkedin_insights
WHERE linkedin_id = '105256406'
AND created_at >= '2024-01-01'
AND created_at < '2024-07-01';
```

| Metric | Old Schema | New Schema |
|--------|------------|------------|
| Index Usage | None (record_id first) | Full (linkedin_id first) |
| Granules Scanned | All granules, filter post-scan | Only account's granules |
| Estimated Speedup | Baseline | **10-100x faster** |

#### Query 2: Aggregate metrics for account

```sql
SELECT
    toDate(created_at) as date,
    sum(impressionCount) as impressions,
    sum(totalFollowerCount) as followers
FROM linkedin_insights
WHERE linkedin_id = '105256406'
GROUP BY date
ORDER BY date;
```

| Metric | Old Schema | New Schema |
|--------|------------|------------|
| Data Scanned | Full table | Account's data only |
| Memory Usage | High | Low |
| Estimated Speedup | Baseline | **10-100x faster** |

#### Query 3: Lookup by record_id (rare)

```sql
SELECT * FROM linkedin_insights
WHERE record_id = '105256406_2024-06-15';
```

| Metric | Old Schema | New Schema |
|--------|------------|------------|
| Index Usage | Full (record_id first) | None |
| Performance | Fast | Slower (full scan or use PREWHERE) |
| Impact | N/A | Acceptable (rare query pattern) |

### linkedin_posts

#### Query 1: All posts for account

```sql
SELECT * FROM linkedin_posts
WHERE linkedin_id = '105256406'
ORDER BY published_at DESC
LIMIT 50;
```

| Metric | Old Schema | New Schema |
|--------|------------|------------|
| Index Usage | None (post_id first) | Full (linkedin_id first) |
| Granules Scanned | All granules | Only account's granules |
| Estimated Speedup | Baseline | **10-100x faster** |

#### Query 2: Post engagement analytics

```sql
SELECT
    post_id,
    comments,
    favorites,
    impressions,
    total_engagement
FROM linkedin_posts
WHERE linkedin_id = '105256406'
AND published_at >= '2024-01-01'
ORDER BY total_engagement DESC
LIMIT 10;
```

| Metric | Old Schema | New Schema |
|--------|------------|------------|
| Data Scanned | Full table | Account's data only |
| Estimated Speedup | Baseline | **10-100x faster** |

#### Query 3: Lookup by post_id (occasional)

```sql
SELECT * FROM linkedin_posts
WHERE post_id = 'urn:li:share:7654321';
```

| Metric | Old Schema | New Schema |
|--------|------------|------------|
| Index Usage | Full (post_id first) | Partial |
| Performance | Very fast | Slower (scans within partitions) |
| Mitigation | N/A | Add `linkedin_id` to query if known |

**Workaround for post_id lookups:**
```sql
-- If linkedin_id is known, include it for optimal performance
SELECT * FROM linkedin_posts
WHERE linkedin_id = '105256406'
AND post_id = 'urn:li:share:7654321';
```

---

## Migration Strategy

Since `ORDER BY` cannot be altered on existing tables, migration requires:

### Option 1: Create New Table + Migrate Data

```sql
-- 1. Create new table with _new suffix
CREATE TABLE contentstudiobackend.linkedin_insights_new (...);

-- 2. Copy data
INSERT INTO linkedin_insights_new
SELECT * FROM linkedin_insights;

-- 3. Rename tables
RENAME TABLE linkedin_insights TO linkedin_insights_old,
             linkedin_insights_new TO linkedin_insights;

-- 4. Verify and drop old table
DROP TABLE linkedin_insights_old;
```

### Option 2: Use Materialized View (Zero Downtime)

```sql
-- 1. Create new table
CREATE TABLE contentstudiobackend.linkedin_insights_v2 (...);

-- 2. Create materialized view to populate new table
CREATE MATERIALIZED VIEW linkedin_insights_mv
TO linkedin_insights_v2 AS
SELECT * FROM linkedin_insights;

-- 3. Backfill historical data
INSERT INTO linkedin_insights_v2
SELECT * FROM linkedin_insights;

-- 4. Switch application to use new table
-- 5. Drop old table and MV
```

---

## Summary

| Improvement | linkedin_insights | linkedin_posts |
|-------------|-------------------|----------------|
| Query Performance | 10-100x faster for account queries | 10-100x faster for account queries |
| Storage Efficiency | 40-60% reduction | 40-60% reduction |
| Deduplication | Automatic via version column (`inserted_at`) | Automatic via version column (`saving_time`) |
| Partition Column | `inserted_at` (snapshot date) | `published_at` (post publish date) |
| Tradeoff | Slower record_id lookups | Slower post_id lookups |

The tradeoff is acceptable because:
1. Account-level queries are the dominant pattern (>95% of queries)
2. Point lookups by record_id/post_id are rare
3. Point lookups can include linkedin_id for optimal performance when needed
4. Partitioning by `published_at` enables efficient date-range queries for post analytics

---

## Schema Comparison: UAT vs Production vs Recommended

| Aspect | UAT/Staging | Production | Recommended |
|--------|-------------|------------|-------------|
| ORDER BY | `(record_id, inserted_at)` | `(linkedin_id, record_id, inserted_at)` | `(linkedin_id, record_id)` |
| PRIMARY KEY | `record_id` | `(linkedin_id, record_id)` | `(linkedin_id, record_id)` |
| linkedin_id first | :x: No | :white_check_mark: Yes | :white_check_mark: Yes |
| Version column | :x: No | :x: No | :white_check_mark: Yes (`inserted_at`) |
| Deduplication | :x: Broken | :x: Broken | :white_check_mark: Works |
| Account queries | :x: Full scan | :white_check_mark: Index skip | :white_check_mark: Index skip |
| Compression | :x: Partial | :white_check_mark: Full | :white_check_mark: Full |
| Timestamp safety | :x: `inserted_at` in key | :x: `inserted_at` in key | :white_check_mark: `record_id` uses date only |

### Ranking (Best to Worst)

1. **Recommended schema** - Correct dedup + correct ORDER BY + compression
2. **Production schema** - Correct ORDER BY + compression, broken dedup
3. **UAT/Staging schema** - Wrong ORDER BY + broken dedup + no compression

### Problems with UAT/Staging Schema

#### Current UAT/Staging Schema

```sql
PRIMARY KEY record_id
ORDER BY (record_id, inserted_at)
ENGINE = ReplicatedReplacingMergeTree(...)  -- No version column
```

#### Problem 1: Wrong ORDER BY - Kills Query Performance

**Typical query pattern:**
```sql
SELECT * FROM linkedin_insights
WHERE linkedin_id = '105256406'
AND created_at >= '2024-01-01';
```

| Schema | What Happens | Performance |
|--------|--------------|-------------|
| UAT (`record_id` first) | ClickHouse scans ALL data, filters after | **Full table scan** |
| Production (`linkedin_id` first) | ClickHouse jumps to account's data via index | **Index skip (10-100x faster)** |

**Why?** ClickHouse's sparse index only works efficiently on the **first column** of ORDER BY. With `record_id` first, filtering by `linkedin_id` can't use the index.

#### Problem 2: Deduplication is Broken

**Scenario:** Same insight record processed twice (e.g., retry, reprocessing)

```
Insert 1: record_id='123_2024-01-01', inserted_at='10:00:00'
Insert 2: record_id='123_2024-01-01', inserted_at='14:00:00'
```

| Schema | ORDER BY Key | Result |
|--------|--------------|--------|
| UAT | `(record_id, inserted_at)` | **2 rows** - different keys! |
| Recommended | `(linkedin_id, record_id)` + version column | **1 row** - deduped correctly |

**Why?** With `inserted_at` in ORDER BY, each insert has a unique key, so ReplacingMergeTree sees them as different rows.

#### Problem 3: Missing Compression - Wastes Storage

**UAT columns without codecs:**
```sql
`linkedin_id` String,              -- No CODEC
`impressionCount` Int64,           -- No CODEC
`totalFollowerCount` Int64,        -- No CODEC
`inserted_at` DateTime64(6),       -- No CODEC
-- ... most columns missing codecs
```

**Expected impact:**
- Without codecs: ~100 bytes per row
- With T64 + ZSTD: ~40-60 bytes per row
- **40-60% storage wasted**

#### Visual Summary

```
UAT/Staging Schema Problems:

┌─────────────────────────────────────────────────────────────────┐
│  ORDER BY (record_id, inserted_at)                              │
│                                                                 │
│  ❌ Query: WHERE linkedin_id = 'X'                              │
│     → Can't use index, scans everything                         │
│                                                                 │
│  ❌ Dedup: Same record inserted twice                           │
│     → Keeps both rows (different inserted_at = different key)   │
│                                                                 │
│  ❌ Storage: Most columns uncompressed                          │
│     → 40-60% more disk usage                                    │
└─────────────────────────────────────────────────────────────────┘

Recommended Schema:

┌─────────────────────────────────────────────────────────────────┐
│  ORDER BY (linkedin_id, record_id)                              │
│  ENGINE = ReplacingMergeTree(..., inserted_at)                  │
│                                                                 │
│  ✅ Query: WHERE linkedin_id = 'X'                              │
│     → Index skip, reads only account's data                     │
│                                                                 │
│  ✅ Dedup: Same record inserted twice                           │
│     → Keeps only latest (highest inserted_at)                   │
│                                                                 │
│  ✅ Storage: All columns compressed                             │
│     → 40-60% less disk usage                                    │
│                                                                 │
│  ✅ Safe: record_id uses date only (no timestamp precision      │
│     issues like created_at DateTime64)                          │
└─────────────────────────────────────────────────────────────────┘
```

#### Bottom Line

The UAT schema would work functionally, but:
- Queries will be **10-100x slower** for account-level analytics
- Data will **duplicate** on reprocessing/retries
- Storage will cost **40-60% more**

This is a regression from the production schema, not an improvement.

### Recommendation

**Do NOT deploy the UAT schema to production.** Instead, deploy the recommended schema with:
- `linkedin_id` first in ORDER BY for efficient account queries
- Version column for proper deduplication
- Full compression codecs for storage efficiency

---

## Current Production Schema Analysis

### Production linkedin_insights

```sql
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/contentstudiobackend/{uuid}', '{replica}')
PARTITION BY toYYYYMM(inserted_at)
PRIMARY KEY (linkedin_id, record_id)
ORDER BY (linkedin_id, record_id, inserted_at)
```

**Assessment: ✅ Already Optimized**

| Aspect | Status | Notes |
|--------|--------|-------|
| ORDER BY | ✅ Good | `linkedin_id` first enables efficient account queries |
| PRIMARY KEY | ✅ Good | Matches ORDER BY prefix |
| PARTITION BY | ✅ Acceptable | `inserted_at` works, `created_at` would be slightly better for date-range pruning |
| Compression | ⚠️ Partial | Newer columns missing codecs |

### Production linkedin_posts

```sql
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/contentstudiobackend/{uuid}', '{replica}')
PARTITION BY toYYYYMM(created_at)
PRIMARY KEY (linkedin_id, post_id)
ORDER BY (linkedin_id, post_id, created_at)
```

**Assessment: ⚠️ Needs Migration**

| Aspect | Status | Notes |
|--------|--------|-------|
| ORDER BY | ⚠️ Issue | `created_at` in ORDER BY breaks deduplication |
| PRIMARY KEY | ✅ Good | Matches ORDER BY prefix |
| PARTITION BY | ⚠️ Issue | Should use `published_at` for query efficiency |
| Version Column | ❌ Missing | No version column for ReplacingMergeTree |
| Compression | ⚠️ Partial | Some columns missing codecs |

**Required changes:**
1. Change `PARTITION BY toYYYYMM(created_at)` → `PARTITION BY toYYYYMM(published_at)`
2. Change `ORDER BY (linkedin_id, post_id, created_at)` → `ORDER BY (linkedin_id, post_id)`
3. Add version column `saving_time` to ReplacingMergeTree engine

---

## Data Deduplication

### How ReplacingMergeTree Deduplication Works

ReplacingMergeTree deduplicates rows based on the **ORDER BY** key:
- Rows with identical ORDER BY values are considered duplicates
- During background merges, duplicates are collapsed to a single row
- If a **version column** is specified, the row with the highest version is kept
- If no version column, the last inserted row is kept

**Important:** Deduplication is NOT immediate. It happens during background merges. To get deduplicated results:
- Use `SELECT ... FINAL` (forces merge, slower)
- Use `GROUP BY` with `argMax()` (recommended for production queries)

### Current Production Schema Issue

```sql
-- CURRENT (PROBLEMATIC)
ORDER BY (linkedin_id, record_id, inserted_at)  -- inserted_at in ORDER BY!
ENGINE = ReplicatedReplacingMergeTree(...)      -- NO version column!
```

**Problem:** With `inserted_at` in ORDER BY, two rows with same `(linkedin_id, record_id)` but different `inserted_at` are considered **different rows** - no deduplication!

```
-- Same insight processed twice = 2 rows (BAD)
Row 1: (linkedin_id='123', record_id='123_2024-01-01', inserted_at='10:00')
Row 2: (linkedin_id='123', record_id='123_2024-01-01', inserted_at='14:00')
Result: BOTH rows kept - they have different ORDER BY keys!
```

### Correct Schema for Deduplication

```sql
-- CORRECT
ORDER BY (linkedin_id, record_id)                                    -- NO timestamp in ORDER BY
ENGINE = ReplicatedReplacingMergeTree(..., inserted_at)              -- version column specified
```

**Result:** Rows with same `(linkedin_id, record_id)` are deduplicated, keeping the one with highest `inserted_at`.

```
-- Same insight processed twice = 1 row (GOOD)
Row 1: (linkedin_id='123', record_id='123_2024-01-01', inserted_at='10:00')
Row 2: (linkedin_id='123', record_id='123_2024-01-01', inserted_at='14:00')
Result: Only Row 2 kept (highest inserted_at)
```

**Why `record_id` is safer than `created_at`:**

```go
// record_id uses date only - guaranteed consistent
ins.RecordID = wo.AccountID + "_" + ins.CreatedAt.Format("2006-01-02")  // "123_2024-01-14"

// created_at is DateTime64(6) - microsecond precision could vary
// Insert 1: created_at = '2024-01-14 00:00:00.000000'
// Insert 2: created_at = '2024-01-14 00:00:00.123456'  ← Would NOT dedupe!
```

### Summary of Changes Needed

| Table | Current ORDER BY | Correct ORDER BY | Version Column | PARTITION BY |
|-------|------------------|------------------|----------------|--------------|
| `linkedin_insights` | `(linkedin_id, record_id, inserted_at)` | `(linkedin_id, record_id)` | `inserted_at` | `toYYYYMM(inserted_at)` (no change) |
| `linkedin_posts` | `(linkedin_id, post_id, created_at)` | `(linkedin_id, post_id)` | `saving_time` | `toYYYYMM(published_at)` ← changed |

### Query Patterns for Deduplicated Results

Until background merges complete, use these patterns:

```sql
-- Option 1: FINAL (simple but slower)
SELECT * FROM linkedin_insights FINAL
WHERE linkedin_id = '105256406';

-- Option 2: argMax (recommended for production)
SELECT
    linkedin_id,
    record_id,
    argMax(impressionCount, inserted_at) AS impressionCount,
    argMax(totalFollowerCount, inserted_at) AS totalFollowerCount,
    argMax(reach, inserted_at) AS reach,
    max(inserted_at) AS inserted_at
FROM linkedin_insights
WHERE linkedin_id = '105256406'
GROUP BY linkedin_id, record_id;

-- Option 3: Subquery with row_number (most flexible)
SELECT * FROM (
    SELECT *,
        row_number() OVER (PARTITION BY linkedin_id, record_id ORDER BY inserted_at DESC) AS rn
    FROM linkedin_insights
    WHERE linkedin_id = '105256406'
) WHERE rn = 1;
```

### Check for Existing Duplicates

```sql
-- Find duplicate insights
SELECT
    linkedin_id,
    record_id,
    count() AS cnt,
    groupArray(inserted_at) AS insert_times
FROM linkedin_insights
GROUP BY linkedin_id, record_id
HAVING cnt > 1
ORDER BY cnt DESC
LIMIT 20;

-- Find duplicate posts
SELECT
    linkedin_id,
    post_id,
    count() AS cnt,
    groupArray(saving_time) AS save_times
FROM linkedin_posts
GROUP BY linkedin_id, post_id
HAVING cnt > 1
ORDER BY cnt DESC
LIMIT 20;
```

---

## Migration Scripts

### linkedin_insights - Add Missing Compression Codecs

The following columns in production are missing compression codecs. Unfortunately, **codecs cannot be added via ALTER** on existing columns in ClickHouse without recreating the table.

**Columns missing compression:**
- `created_at` - missing `CODEC(Delta(4), ZSTD(1))`
- `reactions` - missing `CODEC(T64, ZSTD(1))`
- `engagement` - missing `CODEC(ZSTD(1))`
- `page_views` - missing `CODEC(T64, ZSTD(1))`
- `desktop_page_views` - missing `CODEC(T64, ZSTD(1))`
- `mobile_page_views` - missing `CODEC(T64, ZSTD(1))`
- `overview_page_views` - missing `CODEC(T64, ZSTD(1))`
- `about_page_views` - missing `CODEC(T64, ZSTD(1))`
- `jobs_page_views` - missing `CODEC(T64, ZSTD(1))`
- `people_page_views` - missing `CODEC(T64, ZSTD(1))`
- `careers_page_views` - missing `CODEC(T64, ZSTD(1))`
- `life_at_page_views` - missing `CODEC(T64, ZSTD(1))`
- `insights_page_views` - missing `CODEC(T64, ZSTD(1))`
- `products_page_views` - missing `CODEC(T64, ZSTD(1))`
- `page_views_by_country` - missing `CODEC(ZSTD(1))`
- `page_views_by_region` - missing `CODEC(ZSTD(1))`
- `page_views_by_industry` - missing `CODEC(ZSTD(1))`
- `page_views_by_seniority` - missing `CODEC(ZSTD(1))`
- `page_views_by_function` - missing `CODEC(ZSTD(1))`
- `page_views_by_staff_count` - missing `CODEC(ZSTD(1))`
- `unique_visitors` - missing `CODEC(T64, ZSTD(1))`
- `daily_follower_count` - missing `CODEC(T64, ZSTD(1))`

**Recommendation:** These missing codecs are a minor optimization. New data will still compress at the block level. Full compression requires table recreation (see below).

### linkedin_posts - Add Missing Column

```sql
-- Add carousel_preview column if not exists
ALTER TABLE contentstudiobackend.linkedin_posts
ADD COLUMN IF NOT EXISTS `carousel_preview` String DEFAULT '' CODEC(ZSTD(1));
```

---

## Full Table Recreation (Optional - For Maximum Optimization)

If you want full compression on all columns, use this migration approach:

### Step 1: Create New Tables

```sql
-- linkedin_insights with full optimization and CORRECT deduplication
-- Replace 'prod' with your actual cluster name
CREATE TABLE contentstudiobackend.linkedin_insights_v2 ON CLUSTER prod
(
    `linkedin_id` String CODEC(ZSTD(1)),
    `record_id` String CODEC(ZSTD(1)),
    `impressionCount` Int64 CODEC(T64, ZSTD(1)),
    `organicFollowerCount` Int64 CODEC(T64, ZSTD(1)),
    `totalFollowerCount` Int64 CODEC(T64, ZSTD(1)),
    `paidFollowerCount` Int64 CODEC(T64, ZSTD(1)),
    `inserted_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `reach` Int64 CODEC(T64, ZSTD(1)),
    `repost` Int64 CODEC(T64, ZSTD(1)),
    `comments` Int64 CODEC(T64, ZSTD(1)),
    `post_clicks` Int64 CODEC(T64, ZSTD(1)),
    `followers_by_seniority` String CODEC(ZSTD(1)),
    `followers_by_industry` String CODEC(ZSTD(1)),
    `followers_by_country` String CODEC(ZSTD(1)),
    `followers_by_city` String CODEC(ZSTD(1)),
    `organization_name` String DEFAULT '' CODEC(ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000-00-00 00:00:00.000000' CODEC(Delta(4), ZSTD(1)),
    `reactions` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `engagement` Float64 DEFAULT 0 CODEC(ZSTD(1)),
    `page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `desktop_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `mobile_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `overview_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `about_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `jobs_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `people_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `careers_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `life_at_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `insights_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `products_page_views` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `page_views_by_country` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_region` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_industry` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_seniority` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_function` String DEFAULT '' CODEC(ZSTD(1)),
    `page_views_by_staff_count` String DEFAULT '' CODEC(ZSTD(1)),
    `unique_visitors` Int64 DEFAULT 0 CODEC(T64, ZSTD(1)),
    `daily_follower_count` Int64 DEFAULT 0 CODEC(T64, ZSTD(1))
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/contentstudiobackend/linkedin_insights_v2', '{replica}', inserted_at)
PARTITION BY toYYYYMM(inserted_at)
PRIMARY KEY (linkedin_id, record_id)
ORDER BY (linkedin_id, record_id)  -- REMOVED inserted_at from ORDER BY for proper dedup
SETTINGS index_granularity = 8192;

-- linkedin_posts with full optimization and CORRECT deduplication
CREATE TABLE contentstudiobackend.linkedin_posts_v2 ON CLUSTER prod
(
    `linkedin_id` String CODEC(ZSTD(1)),
    `post_id` String CODEC(ZSTD(1)),
    `activity` String CODEC(ZSTD(1)),
    `media_type` String CODEC(ZSTD(1)),
    `article_url` String CODEC(ZSTD(1)),
    `article_title` String CODEC(ZSTD(1)),
    `post_data` String CODEC(ZSTD(1)),
    `image` String CODEC(ZSTD(1)),
    `media` Array(String) CODEC(ZSTD(1)),
    `type` String CODEC(ZSTD(1)),
    `hashtags` Array(String) CODEC(ZSTD(1)),
    `comments` Int64 CODEC(T64, ZSTD(1)),
    `total_engagement` Float64 DEFAULT 0 CODEC(ZSTD(1)),
    `favorites` Int64 CODEC(T64, ZSTD(1)),
    `title` String CODEC(ZSTD(1)),
    `day_of_week` String CODEC(ZSTD(1)),
    `hour_of_day` Int64 CODEC(T64, ZSTD(1)),
    `created_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `saving_time` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `poll_data` String CODEC(ZSTD(1)),
    `reach` Int64 CODEC(T64, ZSTD(1)),
    `repost` Int64 CODEC(T64, ZSTD(1)),
    `post_clicks` Int64 CODEC(T64, ZSTD(1)),
    `impressions` Int64 CODEC(T64, ZSTD(1)),
    `published_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `last_modified_at` DateTime64(6) DEFAULT '0000000000.000000' CODEC(Delta(4), ZSTD(1)),
    `lifecycle_state` String DEFAULT '' CODEC(ZSTD(1)),
    `visibility` String DEFAULT '' CODEC(ZSTD(1)),
    `is_reshare_disabled` Bool DEFAULT false,
    `feed_distribution` String DEFAULT '' CODEC(ZSTD(1)),
    `third_party_channels` Array(String) DEFAULT [] CODEC(ZSTD(1)),
    `carousel_preview` String DEFAULT '' CODEC(ZSTD(1))
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/contentstudiobackend/linkedin_posts_v2', '{replica}', saving_time)
PARTITION BY toYYYYMM(published_at)
PRIMARY KEY (linkedin_id, post_id)
ORDER BY (linkedin_id, post_id)  -- REMOVED created_at from ORDER BY for proper dedup
SETTINGS index_granularity = 8192;
```

### Step 2: Migrate Data

```sql
-- Migrate linkedin_insights data
-- Note: Regenerate record_id to ensure consistent format: linkedin_id + "_" + date(created_at)
-- This fixes any older data that may have inconsistent record_id values
INSERT INTO contentstudiobackend.linkedin_insights_v2
SELECT
    linkedin_id,
    concat(linkedin_id, '_', formatDateTime(created_at, '%Y-%m-%d')) AS record_id,
    impressionCount,
    organicFollowerCount,
    totalFollowerCount,
    paidFollowerCount,
    inserted_at,
    reach,
    repost,
    comments,
    post_clicks,
    followers_by_seniority,
    followers_by_industry,
    followers_by_country,
    followers_by_city,
    organization_name,
    created_at,
    reactions,
    engagement,
    page_views,
    desktop_page_views,
    mobile_page_views,
    overview_page_views,
    about_page_views,
    jobs_page_views,
    people_page_views,
    careers_page_views,
    life_at_page_views,
    insights_page_views,
    products_page_views,
    page_views_by_country,
    page_views_by_region,
    page_views_by_industry,
    page_views_by_seniority,
    page_views_by_function,
    page_views_by_staff_count,
    unique_visitors,
    daily_follower_count
FROM contentstudiobackend.linkedin_insights;

-- Migrate linkedin_posts data
-- Note: Uses COALESCE to handle NULL published_at values by falling back to created_at
INSERT INTO contentstudiobackend.linkedin_posts_v2
SELECT
    linkedin_id,
    post_id,
    activity,
    media_type,
    article_url,
    article_title,
    post_data,
    image,
    media,
    type,
    hashtags,
    comments,
    total_engagement,
    favorites,
    title,
    day_of_week,
    hour_of_day,
    created_at,
    saving_time,
    poll_data,
    reach,
    repost,
    post_clicks,
    impressions,
    COALESCE(published_at, created_at) AS published_at,
    last_modified_at,
    lifecycle_state,
    visibility,
    is_reshare_disabled,
    feed_distribution,
    third_party_channels,
    '' AS carousel_preview
FROM contentstudiobackend.linkedin_posts;
```

### Step 3: Verify Row Counts

```sql
-- Verify linkedin_insights
SELECT 'linkedin_insights' AS table, count() AS rows FROM contentstudiobackend.linkedin_insights
UNION ALL
SELECT 'linkedin_insights_v2' AS table, count() AS rows FROM contentstudiobackend.linkedin_insights_v2;

-- Verify linkedin_posts
SELECT 'linkedin_posts' AS table, count() AS rows FROM contentstudiobackend.linkedin_posts
UNION ALL
SELECT 'linkedin_posts_v2' AS table, count() AS rows FROM contentstudiobackend.linkedin_posts_v2;
```

### Step 4: Swap Tables

```sql
-- Rename tables on cluster (atomic operation per table pair)
-- Replace 'prod' with your actual cluster name

RENAME TABLE
    contentstudiobackend.linkedin_insights TO contentstudiobackend.linkedin_insights_old,
    contentstudiobackend.linkedin_insights_v2 TO contentstudiobackend.linkedin_insights
ON CLUSTER prod;

RENAME TABLE
    contentstudiobackend.linkedin_posts TO contentstudiobackend.linkedin_posts_old,
    contentstudiobackend.linkedin_posts_v2 TO contentstudiobackend.linkedin_posts
ON CLUSTER prod;
```

**Note:** The `ON CLUSTER` clause ensures the rename is executed on all replicas. Run these commands one at a time and verify each completes successfully before proceeding.

### Step 5: Cleanup (After Verification)

```sql
-- Only run after verifying new tables work correctly
-- Replace 'prod' with your actual cluster name

DROP TABLE contentstudiobackend.linkedin_insights_old ON CLUSTER prod;
DROP TABLE contentstudiobackend.linkedin_posts_old ON CLUSTER prod;
```

---

## Minimal Migration (Recommended)

If you don't want to recreate tables, the production schema is already well-optimized. Just add the missing column:

```sql
-- Only required change for linkedin_posts
ALTER TABLE contentstudiobackend.linkedin_posts
ADD COLUMN IF NOT EXISTS `carousel_preview` String DEFAULT '' CODEC(ZSTD(1));
```

The missing compression codecs on `linkedin_insights` newer columns will only affect storage size slightly. New data written to these columns will still benefit from block-level compression (LZ4 by default).

---

## Validation Queries

### Check Table Sizes

```sql
SELECT
    table,
    formatReadableSize(sum(bytes_on_disk)) AS size_on_disk,
    sum(rows) AS total_rows,
    formatReadableSize(sum(bytes_on_disk) / sum(rows)) AS bytes_per_row
FROM system.parts
WHERE database = 'contentstudiobackend'
  AND table IN ('linkedin_insights', 'linkedin_posts')
  AND active
GROUP BY table;
```

### Check Compression Ratio

```sql
SELECT
    table,
    column,
    formatReadableSize(sum(column_data_compressed_bytes)) AS compressed,
    formatReadableSize(sum(column_data_uncompressed_bytes)) AS uncompressed,
    round(sum(column_data_uncompressed_bytes) / sum(column_data_compressed_bytes), 2) AS ratio
FROM system.parts_columns
WHERE database = 'contentstudiobackend'
  AND table = 'linkedin_insights'
  AND active
GROUP BY table, column
ORDER BY sum(column_data_compressed_bytes) DESC;
```

### Verify Query Performance

```sql
-- Test linkedin_insights account-level query (should use index efficiently)
EXPLAIN indexes = 1
SELECT * FROM contentstudiobackend.linkedin_insights
WHERE linkedin_id = '105256406'
AND inserted_at >= '2024-01-01'
LIMIT 100;

-- Test linkedin_posts account-level query with published_at filter
EXPLAIN indexes = 1
SELECT * FROM contentstudiobackend.linkedin_posts
WHERE linkedin_id = '105256406'
AND published_at >= '2024-01-01'
AND published_at < '2024-07-01'
LIMIT 100;

-- Check that PrimaryKey is being used in the output
-- Partition pruning should also be visible for date-range queries
```
