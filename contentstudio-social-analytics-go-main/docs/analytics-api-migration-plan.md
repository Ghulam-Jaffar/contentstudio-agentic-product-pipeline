# Analytics API Migration Plan: PHP Laravel → Go

## Context

ContentStudio's analytics APIs currently live in the PHP Laravel backend (`contentstudio-backend`). These APIs serve analytics dashboards by querying ClickHouse. We need to migrate the **ClickHouse-hitting endpoints only** (~96 endpoints) to the existing Go codebase (`contentstudio-social-analytics-go`) for better performance and consolidation with the existing Go data pipeline.

---

## Scope

### IN SCOPE (ClickHouse queries) — 96 endpoints

| # | Platform | Endpoints | Builder (PHP source) |
|---|----------|-----------|---------------------|
| 1 | Facebook Analytics | 13 | `FacebookBuilder.php` (2,399 lines) |
| 2 | Facebook Competitor | 12 | `FacebookCompetitorBuilder.php` |
| 3 | Instagram Analytics | 13 | `InstagramBuilder.php` |
| 4 | Instagram Competitor | 6 | `InstagramCompetitorBuilder.php` |
| 5 | LinkedIn Analytics | 9 | `LinkedinBuilder.php` |
| 6 | YouTube Analytics | 11 | `YoutubeBuilder.php` |
| 7 | TikTok Analytics | 6 | `TiktokBuilder.php` |
| 8 | Pinterest Analytics | 10 | `PinterestBuilder.php` |
| 9 | Twitter/X Analytics | 6 | `TwitterBuilder.php` |
| 10 | Overview V2 | 6 | `OverviewV2Builder.php` |
| 11 | Campaign/Label | 4 | `CampaignLabelAnalyticsBuilder.php` |
| | **Total** | **96** | |

### OUT OF SCOPE (not ClickHouse)

- Twitter/X Settings (5) — MongoDB CRUD
- Overview V1 (6) — Elasticsearch legacy
- Dashboard Analytics (3) — MongoDB
- Share Link Management (7) — MongoDB CRUD
- Reports & Scheduled Reports (8) — MongoDB + job queue
- Job Triggers (2) — queue dispatch
- AI Insights (7) — external AI service
- Competitor Search — Graph API
- Competitor CRUD (Instagram add/update/get/delete) — MongoDB

---

## Architecture

### Decision 1: Extend existing API server

Add analytics routes to the existing `src/cmd/api-server/main.go` which already has JWT middleware, config, MongoDB, graceful shutdown. No new service binary needed.

### Decision 2: Add `go-chi/chi` v5 router

Replace `http.NewServeMux()` with `chi` for route grouping, middleware chaining per group, and clean URL parameter support. `chi` is 100% `net/http` compatible — existing handlers and JWT middleware work unchanged.

### Decision 3: Layered architecture (mirrors PHP exactly)

```
HTTP Request
    ↓
[chi Router + JWT Middleware]
    ↓
[Handler]  (src/api/analytics/{platform}_handler.go)
    - Parse request, validate, compute previous period
    - Call builder for current + previous period
    - Format response matching PHP JSON structure
    ↓
[Builder]  (src/api/analytics/builders/{platform}_builder.go)
    - Construct raw SQL strings (identical to PHP)
    - Stateless functions, each returns a SQL string
    ↓
[Query Executor]  (src/db/clickhouse/query.go)
    - Execute SQL via clickhouse.Conn.Query()
    - Return []map[string]interface{} (flexible, matches PHP's dynamic row access)
```

---

## Directory Structure

```
src/
  api/
    analytics/                              # NEW
      request.go                            # Request parsing, validation, date utilities
      response.go                           # JSON response helpers, growth formula
      router.go                             # chi route registration for all platforms
      facebook_handler.go                   # 13 handlers
      instagram_handler.go                  # 13 handlers
      linkedin_handler.go                   # 9 handlers
      youtube_handler.go                    # 11 handlers
      tiktok_handler.go                     # 6 handlers
      pinterest_handler.go                  # 10 handlers
      twitter_handler.go                    # 6 handlers
      overview_handler.go                   # 6 handlers
      campaign_label_handler.go             # 4 handlers
      facebook_competitor_handler.go        # 12 handlers
      instagram_competitor_handler.go       # 6 handlers
      builders/
        base.go                             # Shared SQL utilities (date filters, dedup CTEs)
        facebook_builder.go                 # ~25 query methods
        instagram_builder.go                # ~30 query methods
        linkedin_builder.go                 # ~20 query methods
        youtube_builder.go                  # ~20 query methods
        tiktok_builder.go                   # ~8 query methods
        pinterest_builder.go                # ~12 query methods
        twitter_builder.go                  # ~8 query methods
        overview_builder.go                 # ~8 query methods
        campaign_label_builder.go           # ~5 query methods
        facebook_competitor_builder.go      # ~14 query methods
        instagram_competitor_builder.go     # ~8 query methods
  db/
    clickhouse/
      query.go                              # NEW: QueryRows, QueryRow methods
  cmd/
    api-server/
      main.go                               # MODIFY: Add chi router, ClickHouse init, route registration
  api/
    immediate_work_apis.go                  # MODIFY: Add ClickHouseClient field to APIServer struct
```

---

## Implementation Phases

### Phase 0: Foundation (prerequisite for all platforms)

**0.1 — Add chi router dependency**
- `go get github.com/go-chi/chi/v5`
- Refactor `cmd/api-server/main.go` to use `chi.NewRouter()` instead of `http.NewServeMux()`
- Re-register existing endpoints (`/api/v1/immediate-work`, `/api/v1/competitor-work`, `/health`)

**0.2 — Add ClickHouse SELECT capability**
- File: `src/db/clickhouse/query.go`
- Add `QueryRows(ctx, sql) ([]map[string]interface{}, error)` — returns all rows
- Add `QueryRow(ctx, sql) (map[string]interface{}, error)` — returns first row
- Uses `c.Conn.Query()` + `rows.ColumnTypes()` + `rows.Scan()` for dynamic column mapping

**0.3 — Common request/response utilities**
- File: `src/api/analytics/request.go`
  - `ParseDateRange(date string) (start, end time.Time, err error)` — splits `"YYYY-MM-DD - YYYY-MM-DD"`
  - `ComputePreviousPeriod(start, end time.Time) (prevStart, prevEnd time.Time)` — `prevStart = start - (end - start)`
  - `FormatIDsForSQL(ids []string) string` — produces `('id1','id2')`
  - `FormatMediaTypesForSQL(types []string) string` — produces `'text','link','images'`
  - `ParseAnalyticsRequest(r *http.Request) (*AnalyticsRequest, error)` — JSON decode + defaults
- File: `src/api/analytics/response.go`
  - `SendJSON(w, data)`, `SendError(w, code, message)`
  - `GrowthPercentage(current, previous) interface{}` — `round((c-p)/max(p,1)*100, 2)`, returns `"N/A"` when previous is 0
  - `ComputeDifferences(current, previous map) map` — computes difference + percentage for all matching keys

**0.4 — Base builder utilities**
- File: `src/api/analytics/builders/base.go`
  - `PostsDateFilter(col, tz, start, end string) string` — `toDateTime({col}, 0, '{tz}') BETWEEN ...`
  - `InsightsDateFilter(col, start, end string) string` — `toDate({col}) BETWEEN ...`
  - `DeduplicationCTE(table, idCol, timeCol, platformIDs, dateFilter string) string` — standard `WITH posts as (SELECT id, max(saving_time) ...)`

**0.5 — Extend APIServer + route registration**
- Add `ClickHouseClient *clickhouse.Client` to `APIServer` struct in `immediate_work_apis.go`
- Create `src/api/analytics/router.go` with `RegisterAnalyticsRoutes(r chi.Router, server *APIServer)`
- Initialize ClickHouse client in `cmd/api-server/main.go` and pass to server

---

### Phase 1: Facebook Analytics (13 endpoints) — Reference Implementation

This is the most complex platform and establishes all patterns for subsequent phases.

**Handler file:** `src/api/analytics/facebook_handler.go`
**Builder file:** `src/api/analytics/builders/facebook_builder.go`

| # | Route | Handler | Builder Methods |
|---|-------|---------|----------------|
| 1 | `POST /overview/facebook/summary` | `FacebookSummary` | `GetSummaryQuery()` |
| 2 | `POST /overview/facebook/overviewAudienceGrowth` | `FacebookAudienceGrowth` | `GetAudienceGrowthQuery()` + `GetAudienceRollupQuery()` |
| 3 | `POST /overview/facebook/overviewPublishingBehaviour` | `FacebookPublishingBehaviour` | `GetPublishingBehaviourQuery()` + `GetPublishingRollupQuery()` |
| 4 | `POST /overview/facebook/overviewTopPosts` | `FacebookOverviewTopPosts` | `GetTop15PostsQuery()` |
| 5 | `POST /overview/facebook/getTopPosts` | `FacebookGetTopPosts` | `GetTop15PostsQuery()` |
| 6 | `POST /overview/facebook/overviewActiveUsers` | `FacebookActiveUsers` | `GetActiveUsersQuery()` + `GetActiveUsersPerDayQuery()` |
| 7 | `POST /overview/facebook/overviewImpressions` | `FacebookImpressions` | `GetImpressionsQuery()` + `GetImpressionsRollupQuery()` |
| 8 | `POST /overview/facebook/overviewEngagement` | `FacebookEngagement` | `GetEngagementQuery()` + `GetEngagementRollupQuery()` |
| 9 | `POST /overview/facebook/overviewReelsAnalytics` | `FacebookReelsAnalytics` | `GetReelsAnalyticsQuery()` + `GetReelsRollupQuery()` |
| 10 | `POST /overview/facebook/overviewVideoInsights` | `FacebookVideoInsights` | `GetVideoInsightsQuery()` + `GetVideoRollupQuery()` |
| 11 | `POST /overview/facebook/demographics` | `FacebookDemographics` | `GetAudienceGenderQuery()` + `GetAudienceAgeQuery()` + country/city/language |
| 12 | `POST /overview/facebook/overviewDemographics` | `FacebookOverviewDemographics` | `GetAudienceGenderQuery()` + `GetMaxGenderAgeQuery()` |
| 13 | `POST /overview/facebook/overviewAudienceLocation` | `FacebookAudienceLocation` | `GetAudienceCountryQuery()` + `GetAudienceCityQuery()` |

**Critical business logic to preserve:**
- Active users timezone offset: `offset = round(Carbon::now(tz)->offsetHours) + 8` (Facebook stores page_fans_online in PST/UTC-8)
- Audience growth zero-fill: `arrayFill(x -> not x==0, fan_count_temp)`
- 2-year fallback when current fan_count starts at 0
- Media asset grouping in top posts (multiple rows per post → grouped by post_id)
- PHP `getDateFilters` mutates `currentEndDate` via `addDay()` — Go builders must NOT mutate; compute date bounds per query method independently

**Source reference:**
- PHP Controller: `contentstudio-backend/app/Http/Controllers/Analytics/Analyze/FacebookController.php`
- PHP Builder: `contentstudio-backend/app/Builders/Analytics/Analyze/FacebookBuilder.php`
- SQL Reference: `docs/analytics-api-reference.md` (Sections 2-3)
- API Reference: `docs/analytics-api-endpoints.md` (Section 2)

---

### Phase 2: Instagram Analytics (13 endpoints)

**Key differences from Facebook:**
- Uses `instagram_id`/`media_id` instead of `page_id`/`post_id`
- Uses `stored_event_at` for saving time, `post_created_at` for creation time
- Engagement = likes + comments + saved (not reactions-based)
- Has story metrics (exits, replies, taps_forward, taps_back)
- Has reel metrics (avg_watch_time, total_watch_time)
- Has hashtag analytics

---

### Phase 3: LinkedIn Analytics (9 endpoints)

**Key differences:**
- Page view metrics (desktop, mobile, by country, by seniority, by industry)
- Followers stored as JSON strings (`followers_by_seniority`, `followers_by_industry`)
- Uses `created_at`/`inserted_at` column names

---

### Phase 4: YouTube Analytics (11 endpoints)

**Key differences:**
- Queries span 5 tables: `youtube_channels`, `youtube_videos`, `youtube_activity_insights`, `youtube_traffic_insights`, `youtube_shared_insights`
- Uses `channel_id` as account identifier
- Video sharing insights have 31 platform-specific columns
- Date filter uses `>=` and `<` instead of `BETWEEN`

---

### Phase 5: TikTok Analytics (6 endpoints)

Simplest platform. Straightforward queries on `tiktok_posts` and `tiktok_insights`.

---

### Phase 6: Pinterest Analytics (10 endpoints)

**Key differences:**
- Uses 5 tables: `pinterest_pins`, `pinterest_pin_insights`, `pinterest_users`, `pinterest_user_insights`, `pinterest_boards`
- Engagement = saves + pin_clicks + outbound_clicks

---

### Phase 7: Twitter/X Analytics (6 endpoints)

**Key differences:**
- Uses `last_value()` for deduplication instead of `max(saving_time)`
- Has credits/usage tracking
- Pre-computed `total_engagement` column

---

### Phase 8: Overview V2 (6 endpoints)

**Key differences:**
- Reads from `mv_social_daily_metrics` materialized view
- Uses `uniqMerge()`, `sumMerge()` aggregate functions
- Cross-platform: UNION ALL across all platform subqueries
- Request accepts arrays of platform-specific IDs

---

### Phase 9: Campaign/Label Analytics (4 endpoints)

**Hybrid flow:**
1. Query MongoDB for campaign/label → post ID mappings
2. Pass post IDs into ClickHouse `WHERE post_id IN (...)` filter
3. Run UNION ALL across all `{platform}_posts` tables

Requires new MongoDB repositories for campaigns/labels (or extend existing ones).

---

### Phase 10: Facebook Competitor (12 endpoints)

Queries `facebook_competitor_posts` table instead of `facebook_posts`. Separate builder with competitor-specific metrics.

---

### Phase 11: Instagram Competitor (6 ClickHouse endpoints)

Queries `instagram_competitor_posts` table. Only the ClickHouse-hitting endpoints (excluding search/CRUD).

---

## Key Files to Modify

| File | Change |
|------|--------|
| `src/cmd/api-server/main.go` | Replace `http.NewServeMux()` with `chi.NewRouter()`, init ClickHouse client, register analytics routes |
| `src/api/immediate_work_apis.go` | Add `ClickHouseClient *clickhouse.Client` field to `APIServer` |
| `src/db/clickhouse/client.go` | No changes (connection setup stays) |
| `src/go.mod` | Add `github.com/go-chi/chi/v5` |

## Key Files to Create

| File | Purpose |
|------|---------|
| `src/db/clickhouse/query.go` | `QueryRows()` and `QueryRow()` SELECT methods |
| `src/api/analytics/request.go` | Request parsing, date utilities |
| `src/api/analytics/response.go` | Response formatting, growth calculations |
| `src/api/analytics/router.go` | Route registration for all 96 endpoints |
| `src/api/analytics/builders/base.go` | Shared SQL utilities |
| `src/api/analytics/builders/{platform}_builder.go` | 11 builder files, one per platform |
| `src/api/analytics/{platform}_handler.go` | 11 handler files, one per platform |

---

## Testing Strategy

1. **Builder unit tests**: Verify generated SQL strings match PHP output for identical inputs
2. **Integration tests**: Run same request against both PHP and Go, compare JSON responses field-by-field
3. **Load tests**: Verify <100ms p95 for simple queries, <500ms for complex aggregations at 100 concurrent requests

---

## Deployment Strategy

1. **Shadow mode**: Deploy Go alongside PHP, route analytics traffic to both, compare responses (1-2 weeks per platform)
2. **Feature flag cutover**: Per-platform flags (e.g., `ANALYTICS_GO_FACEBOOK=true`) for Nginx routing
3. **Full migration**: After all platforms validated, make Go primary, keep PHP fallback for 30 days

---

## Verification

After each phase:
1. Run builder unit tests: `make test-analytics-{platform}`
2. Start the API server: `make run-api-server`
3. Send test requests matching the payloads documented in `docs/analytics-api-endpoints.md`
4. Compare responses against the PHP backend running on staging
5. Verify exact JSON key names and structure match
