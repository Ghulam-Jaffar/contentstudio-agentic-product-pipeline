# ClickHouse Schema Migrations — ReplacingMergeTree Version Column Audit

**Last updated:** 2026-04-30  
**Cluster:** `chi-clickhouse-prod` (namespace: `clickhouse`)  
**Database:** `contentstudiobackend`  
**Primary replica:** `chi-clickhouse-prod-s1-replica1-0`

---

## Background

All analytics tables use `ReplicatedReplacingMergeTree`. Without a version column, ClickHouse keeps an arbitrary row when deduplicating on merge. Adding a version column (e.g. `inserted_at`, `updated_at`) ensures the latest row is always kept.

### Safe Migration Pattern

Every table below was migrated using this zero-data-loss approach:

1. `SHOW CREATE TABLE` — capture exact schema
2. `CREATE TABLE table_new` — same schema + version column, ZooKeeper path = `table_v2`
3. `INSERT INTO table_new SELECT * FROM table`
4. `OPTIMIZE TABLE table_new FINAL` — force deduplication
5. Verify: `count()`, `uniqExact(order_by_keys)`, `sum(metric)`, date range
6. `RENAME TABLE table → table_backup, table_new → table`
7. Confirm ZooKeeper path via `system.replicas`

**Backup tables are kept for at least 1 week before dropping.**

### ZooKeeper Path Rule

Always use `table_v2` (not `table_new`) in the ZooKeeper path. The path is fixed at table creation and does NOT change on RENAME — using `_v2` makes it unambiguous after rename.

---

## Completed Migrations

### Session 1 (prior to 2026-04-29)

| Table | Version Column | ZooKeeper Path |
|---|---|---|
| `facebook_posts` | `updated_at` | `facebook_posts_v2` |
| `facebook_insights` | `updated_at` | `facebook_insights_v2` |
| `facebook_media_assets` | `updated_at` | `facebook_media_assets_v2` |
| `facebook_video_insights` | `updated_at` | `facebook_video_insights_v2` |
| `facebook_reels_insights` | `updated_at` | `facebook_reels_insights_v2` |

### Session 2 (2026-04-29)

| Table | Rows Before | Duplicates Removed | Version Column | ZooKeeper Path |
|---|---|---|---|---|
| `pinterest_pins` | 8.9M | 1.6M | `inserted_at` | `pinterest_pins_v2` |
| `instagram_insights` | 9.6M | 875K | `updated_time` | `instagram_insights_v2` |
| `instagram_posts` | 18.8M | 2.5M | `updated_time` | `instagram_posts_v2` |
| `facebook_competitor_insights` | 34M | 17K | `inserted_at` | `facebook_competitor_insights_v2` |
| `instagram_competitor_posts` | 61.5M | 6.4M | `inserted_at` | `instagram_competitor_posts_v2` |
| `facebook_competitor_media_assets` | 315M | 17K | `inserted_at` | `facebook_competitor_media_assets_v2` |
| `facebook_competitor_posts` | 348M | 62M | `inserted_at` | `facebook_competitor_posts_v2` |
| `pinterest_pin_insights` | 397M | 43M | `inserted_at` | `pinterest_pin_insights_v2` |
| `twitter_insights` | 3.9K | — | `saving_time` | `twitter_insights_v2` |
| `pinterest_users` | 105K | — | `inserted_at` | `pinterest_users_v2` |
| `twitter_posts` | 9K | — | `saving_time` | `twitter_posts_v2` |
| `pinterest_user_insights` | 799K | — | `inserted_at` | `pinterest_user_insights_v2` |
| `youtube_shared_insights` | 2.1M | — | `inserted_at` | `youtube_shared_insights_v2` |
| `youtube_traffic_insights` | 3.9M | — | `created_at` | `youtube_traffic_insights_v2` |
| `youtube_activity_insights` | 3.9M | — | `inserted_at` | `youtube_activity_insights_v2` |
| `pinterest_boards` | 3.3M | — | `inserted_at` | `pinterest_boards_v2` |
| `tiktok_insights` | 2.8M | — | `inserted_at` | `tiktok_insights_v2` |
| `youtube_channels` | 2M | — | `inserted_at` | `youtube_channels_v2` |
| `instagram_competitor_insights` | 5.9M | — | `inserted_at` | `instagram_competitor_insights_v2` |
| `youtube_videos` | 28.4M | — | `inserted_at` | `youtube_videos_v2` |
| `tiktok_posts` | 2.6M | — | `inserted_at` | `tiktok_posts_v2` |

### Session 3 (2026-04-30) — Partition Key Redesign

Tables that were partitioned by write-time timestamps (`inserted_at`) had cross-partition duplicate issues: when the same logical record was re-fetched in a later month, it landed in a different partition and could never be deduplicated. These 5 tables were re-created with logical partition keys.

| Table | Old Partition Key | New Partition Key | Version Column | ZooKeeper Path | Rows Before | Dups Removed |
|---|---|---|---|---|---|---|
| `facebook_insights` | `(year * 100 + month)` (from columns) | Same — but fixed from write-time to consistent logical month | `updated_at` | `facebook_insights_v3` | 15.3M | 2.1M |
| `instagram_insights` | `toYYYYMM(stored_event_at)` | `if(position(record_id,'_')>0, toYYYYMM(toDate(substr(record_id,pos+1))), toYYYYMM(stored_event_at))` | `updated_time` | `instagram_insights_v4` | 8.75M | 305K |
| `linkedin_insights` | `toYYYYMM(inserted_at)` | `toYYYYMM(toDate(substr(record_id, pos+1)))` — logical date from `{linkedin_id}_{date}` record_id | `inserted_at` | `linkedin_insights_v3` | 5.83M | 1.1M |
| `youtube_videos` | `toYYYYMM(inserted_at)` | `toYYYYMM(published_at)` — video publication date | `inserted_at` | `youtube_videos_v4` | 1.41M | 600K |
| `pinterest_users` | `toYYYYMM(inserted_at)` | `toYear(inserted_at)` + ORDER BY `(user_id)` | `inserted_at` | `pinterest_users_v4` | 10.3K | 6.6K |

**Note on `instagram_insights` record_id formats:**
- Legacy rows (7M): MD5 hash `2e0a659162...` — no underscore, falls back to `stored_event_at` for partition
- New rows (1.7M): `{instagram_id}_{date}` format — extracts logical date for partition

### Already correct (no migration needed)

| Table | Version Column | Notes |
|---|---|---|
| `linkedin_posts` | `saving_time` | `linkedin_posts_v2` |
| `linkedin_insights` | `inserted_at` | `linkedin_insights_v2` |
| `gmb_daily_metrics` | `inserted_at` | Non-replicated, correct |
| `gmb_local_posts` | `inserted_at` | Non-replicated, correct |
| `gmb_media_assets` | `inserted_at` | Non-replicated, correct |
| `gmb_reviews` | `inserted_at` | Non-replicated, correct |
| `gmb_search_keywords_monthly` | `inserted_at` | Non-replicated, correct |
| `mv_social_daily_followers` | `updated_at` | Materialized view target |

---

## Backup Tables (pending drop)

These are the original tables renamed to `_backup` after migration. Drop after confirming live tables are stable (minimum 1 week).

| Backup Table | Size | Created From |
|---|---|---|
| `facebook_posts_backup` | 76 GiB | Nov 2023 migration |
| `facebook_media_assets_backup` | 90 GiB | Jan 2024 migration |
| `facebook_video_insights_backup` | 1.19 GiB | Nov 2023 migration |
| `facebook_reels_insights_backup` | 47 MiB | Dec 2023 migration |
| `facebook_insights_backup` | 1.35 GiB | Aug 2024 migration |
| `pinterest_pins_backup` | 1.07 GiB | 2026-04-29 |
| `instagram_insights_backup` | 1.83 GiB | 2026-04-29 |
| `instagram_posts_backup` | 5.23 GiB | 2026-04-29 |
| `facebook_competitor_insights_backup` | 8 GiB | 2026-04-29 |
| `instagram_competitor_posts_backup` | 38.6 GiB | 2026-04-29 |
| `facebook_competitor_media_assets_backup` | 90 GiB | 2026-04-29 |
| `facebook_competitor_posts_backup` | ~120 GiB | 2026-04-29 |
| `pinterest_pin_insights_backup` | ~40 GiB | 2026-04-29 |
| `twitter_insights_backup` | ~0 | 2026-04-29 |
| `pinterest_users_backup` | ~2 MiB | 2026-04-29 |
| `twitter_posts_backup` | ~2 MiB | 2026-04-29 |
| `pinterest_user_insights_backup` | ~8 MiB | 2026-04-29 |
| `youtube_shared_insights_backup` | ~60 MiB | 2026-04-29 |
| `youtube_traffic_insights_backup` | ~130 MiB | 2026-04-29 |
| `youtube_activity_insights_backup` | ~132 MiB | 2026-04-29 |
| `pinterest_boards_backup` | ~142 MiB | 2026-04-29 |
| `tiktok_insights_backup` | ~214 MiB | 2026-04-29 |
| `youtube_channels_backup` | ~407 MiB | 2026-04-29 |
| `instagram_competitor_insights_backup` | ~909 MiB | 2026-04-29 |
| `youtube_videos_backup` | ~1.3 GiB | 2026-04-29 |
| `tiktok_posts_backup` | ~1.71 GiB | 2026-04-29 |
| `facebook_insights_backup2` | ~1.5 GiB | 2026-04-30 |
| `instagram_insights_backup2` | ~1.6 GiB | 2026-04-30 |
| `linkedin_insights_backup2` | ~255 MiB | 2026-04-30 |
| `youtube_videos_backup2` | (from previous session ORDER BY fix) | 2026-04-29 |
| `youtube_videos_backup3` | ~563 MiB | 2026-04-30 |
| `pinterest_users_backup2` | (from previous session ORDER BY fix) | 2026-04-29 |
| `pinterest_users_backup3` | ~1.4 MiB | 2026-04-30 |

**Drop command template:**
```sql
DROP TABLE contentstudiobackend.table_name_backup;
```

---

## Known Remaining Issues

### 1. ~~Volatile ORDER BY timestamps~~ — FIXED (Session 2, 2026-04-29)

All 6 tables that had timestamps in ORDER BY were migrated:

| Table | Fixed ORDER BY | Version Column | ZooKeeper Path |
|---|---|---|---|
| `pinterest_users` | `(user_id)` | `inserted_at` | `pinterest_users_v3` → then `v4` after partition fix |
| `tiktok_insights` | `(tiktok_id, record_id)` | `inserted_at` | `tiktok_insights_v3` |
| `twitter_insights` | `(record_id)` | `saving_time` | `twitter_insights_v3` |
| `youtube_channels` | `(record_id, channel_id)` | `inserted_at` | `youtube_channels_v3` |
| `youtube_shared_insights` | `(record_id, channel_id)` | `inserted_at` | `youtube_shared_insights_v3` |
| `youtube_videos` | `(video_id, channel_id)` | `inserted_at` | `youtube_videos_v3` → then `v4` after partition fix |

### 2. ~~instagram_insights cross-partition duplicates~~ — FIXED (2026-04-30)

`instagram_insights` was partitioned by write time (`stored_event_at`). When historical data is re-fetched in a later month, the same `record_id` landed in a different partition — ReplacingMergeTree cannot deduplicate across partitions.

**Fix applied:** Conditional partition key handles both record_id formats:
- Old MD5 hash format (no underscore) → falls back to `toYYYYMM(stored_event_at)`
- New `{instagram_id}_{date}` format → extracts logical date from record_id

Result: 161K cross-partition duplicates removed. ZooKeeper path: `instagram_insights_v4`.

### 3. Unpartitioned tables (not assessed)

| Table | Notes |
|---|---|
| `posts_cluster_feeds` | No partition key — assess for duplicates |
| `tmp_pin_board_map` | Likely temporary — assess if still needed |
| `linkedin_geo_mapping` | Small reference table — likely fine |

### 4. Go pipeline bug fix (already deployed)

`facebook_video_insights`: `BulkInsertVideoInsights` in `db/clickhouse/facebook.go` was missing 3 columns in INSERT:
- `total_video_views_organic`
- `total_video_views_paid`  
- `total_video_view_total_time_paid`

Fixed and committed. Without this fix, Go pipeline rows were overwriting Python rows with zeroes for those 3 columns.

---

## How to Connect

```bash
# Via kubectl exec
kubectl exec -n clickhouse chi-clickhouse-prod-s1-replica1-0 -c clickhouse -- \
  clickhouse-client --query "SELECT ..."

# Check table health
kubectl exec -n clickhouse chi-clickhouse-prod-s1-replica1-0 -c clickhouse -- \
  clickhouse-client --query "
  SELECT table, zookeeper_path, replica_name, is_leader
  FROM system.replicas
  WHERE database = 'contentstudiobackend'
  ORDER BY table"
```

---

## Verification Queries

```sql
-- Check all live tables have version columns
SELECT name, engine_full
FROM system.tables
WHERE database = 'contentstudiobackend'
  AND engine LIKE '%ReplacingMergeTree%'
  AND name NOT LIKE '%_backup%'
  AND name NOT LIKE '%_new%'
  AND name NOT LIKE '%_old%'
  AND engine_full NOT LIKE '%, %replica%)'  -- has 3rd arg = version column
ORDER BY name;

-- Check backup tables still present
SELECT name, formatReadableSize(total_bytes) AS size
FROM system.tables
WHERE database = 'contentstudiobackend' AND name LIKE '%_backup'
ORDER BY total_bytes DESC;

-- Check part counts (high = needs OPTIMIZE)
SELECT table, count() AS parts
FROM system.parts
WHERE database = 'contentstudiobackend' AND active
GROUP BY table
HAVING parts > 100
ORDER BY parts DESC;
```
