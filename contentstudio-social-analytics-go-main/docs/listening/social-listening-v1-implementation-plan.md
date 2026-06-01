# Social Listening V1 Implementation Plan

## Purpose

This plan translates the clarified v1 social listening requirements into a repo-fit implementation sequence for `contentstudio-social-analytics-go`.

Primary source docs:
- `social-listening-implementation.md` — current v1 source of truth
- `social-listing-data-pipeline-prd.md` — future roadmap only, not current execution scope

## Locked Scope

- Build v1 for the 6 Data365-backed platforms only: TikTok, Instagram, Facebook, Twitter/X, Reddit, and Threads.
- Treat the 18-platform PRD as roadmap input, not current implementation scope.
- Store duplicate mention rows when the same post matches multiple topics.
- Support v1 matching features (aligned with Laravel topic schema):
  - `include_keywords` (OR logic — match any keyword)
  - `exclude_keywords`
  - `include_authors`
  - `exclude_authors`
  - `language` (ISO 639-1 codes, passed to Data365 search)
  - `regions` (ISO 3166-1 codes, passed to Data365 search)
- Trust Laravel for authz and billing source-of-truth decisions.
- Laravel enforces a per-workspace topic limit of 5. The Go service trusts that topics in MongoDB are valid and does not enforce billing limits independently.

## Non-Goals

- No 18-platform implementation in this phase.
- No independent Go-side billing authority.
- No redesign of the existing analytics API stack.
- No attempt to force the listening feature into a new architecture that diverges from current repo patterns.

## Repo-Fit Architecture

### API Layer

Follow the existing split already used by:
- `src/api/immediate_work_apis.go`
- `src/api/analytics/...`
- `src/cmd/api-server/main.go`

Add:
- `src/api/handle_listening_work.go`
- `src/api/handle_listening_work_test.go`
- `src/api/analytics/listening/handler.go`
- `src/api/analytics/listening/handler_test.go`

Keep `src/cmd/api-server/main.go` limited to wiring routes and dependencies.

### Service Layer

Follow the existing query-service pattern under `src/services/analytics/...` and the runnable service pattern under `src/services/<domain>/...`.

Add:
- `src/services/analytics/listening/service.go`
- `src/services/listening/listening-fetcher/`
- `src/services/listening/listening-parser/`
- `src/services/listening/listening-sentiment/`
- `src/services/listening/listening-clickhouse-sink/`

### Data Layer

MongoDB:
- `src/models/db/mongo/listening_topic.go` (includes `custom_type_id` field for custom topic types)
- `src/db/mongodb/listening.go`
- Note: `listening_settings` collection is deferred — v1 skips global excludes. Will be added when Laravel implements the settings CRUD.

ClickHouse:
- `src/models/db/clickhouse/listening.go`
- `src/db/clickhouse/listening_write.go`
- `src/db/clickhouse/analytics-get-queries/listening/repository.go`
- `src/deployments/clickhouse/schema/listening_schema.sql`

Kafka models:
- `src/models/kafka/listening.go`

API request/response models:
- `src/models/api/listening.go`

External client:
- `src/clients/social/data365.go`

## Key Implementation Constraints

- Keep configuration under the existing `APP_*` + Viper structure in `src/config/config.go`.
- Reuse the existing Redis client/config instead of introducing a new one.
- Reuse `src/api/httputil/response.go` conventions for query APIs.
- Reuse the existing logger and structured logging field patterns.
- Preserve the existing route registration style in `src/api/analytics/router.go`.

## Important Technical Risk (Resolved)

The current Kafka consumer commits offsets even after handler errors (`src/kafka/consumer.go`). This is acceptable for some current services, but risky for the new listening pipeline.

**Decision: Use option 1** — Add listening-specific handler behavior that writes failures to a `listening-dlq` Kafka topic before returning. This provides minimal repo disruption while giving us DLQ semantics for the listening pipeline. The DLQ message includes the original topic, stage, error, payload, and attempt count (see PRD Section 19 for schema).

## Execution Phases

### Phase 1: Foundations

Goal:
- Land schema, models, config, repositories, and route skeletons without turning on the full pipeline.

Work:
- Add Mongo model for `listening_topics` (include `custom_type_id` for custom topic types)
- Add Mongo repository for topic read operations and operational counters (CRUD stays in Laravel)
- Add Kafka message models for work orders and stage payloads
- Add ClickHouse schema for `listening_mentions` table and `listening_daily_stats` materialized view
- Add ClickHouse write/read models
- Add config bindings for Data365 and listening pipeline settings
- Add empty API handlers and service interfaces
- Add Kafka topic creation entries (including `listening-dlq`)
- Add Redis distributed lock utility for topic processing
- Verify Redis is in local infra compose

Verification:
- Unit tests for config binding
- Unit tests for Mongo repository behavior
- ClickHouse schema applies cleanly in local infra (both `listening_mentions` and `listening_daily_stats`)

### Phase 2: Triggering and Scheduling

Goal:
- Allow Laravel and scheduled jobs to queue listening work in the same style as the repo’s existing work-order flows.

Work:
- Add `POST /api/v1/listening-work`
- Validate topic existence and `mentions_limit_reached` before queueing
- Extend `src/cmd/jobs/main.go` with a listening branch that scans active topics and queues daily work
- Define the listening Kafka work-order topic

Verification:
- Handler tests
- Scheduler unit tests
- Kafka produce tests/mocks

### Phase 3: Fetcher

Goal:
- Fetch raw results from Data365 for all enabled platforms in a topic work order.

Work:
- Implement `src/clients/social/data365.go` with async POST→poll→GET flow:
  - POST search request to Data365
  - Poll GET status until "finished" (backoff: 5s→30s, max 5 min)
  - GET paginated results (100 items/page, cursor-based)
- Build a fetcher service that:
  - acquires Redis distributed lock per topic before starting
  - fans out by enabled platform (6 platform goroutines per topic, parallel)
  - handles Data365 async search/status/results flow per keyword per platform
  - auto-prepends `#` for Instagram keywords missing the prefix (IG only supports hashtag search)
  - applies rate limiting (per-token + global via RateManager) and bounded retries (5 attempts, exponential backoff 300ms→8s)
  - skips expected errors (404, 403) with logging
  - emits raw payloads to Kafka

Verification:
- Client tests for request/response mapping (all 6 platforms)
- Fetcher tests with mocked Data365 responses
- Instagram hashtag auto-prepend test

### Phase 4: Parser and Matching

Goal:
- Normalize raw Data365 responses into a single listening mention model and apply the v1 matching rules.

Work:
- Implement parser service
- Normalize per-platform fields into one parsed `ListeningMention` shape (see PRD Section 4 for field mapping)
- Generate `mention_id = "{platform}:{native_id}"` and `content_hash` (SHA256 dedup key)
- Check Redis dedup set — skip if duplicate within same topic (48h TTL)
- Apply keyword filtering:
  - `exclude_keywords` — case-insensitive substring match, skip mention if any match
  - `include_authors` — if non-empty, only keep mentions from listed authors
  - `exclude_authors` — skip mentions from listed authors
- Compute `matched_keywords` — detect which `include_keywords` appear in post text (case-insensitive substring match)
- Compute `total_engagement` (sum of likes, comments, shares, views, etc.)
- Normalize `content_type` and `media_type` per platform
- Truncate text to 10KB
- Note: `language` and `regions` are passed to Data365 at fetch time, not filtered at parser stage
- Note: workspace-level global excludes deferred to v2 (no `listening_settings` dependency)
- Emit parsed mentions to Kafka for each matching topic (duplicate rows per topic are expected)

Verification:
- Unit tests for all 6 platform response format normalization
- Unit tests for exclude_keywords filtering
- Unit tests for include_authors / exclude_authors filtering
- Unit tests for matched_keywords detection
- Unit tests for Redis dedup behavior
- Unit tests for duplicate-row-per-topic behavior

### Phase 5: Sentiment and Enrichment

Goal:
- Enrich parsed mentions via the existing AI-agents integration style.

Work:
- Implement batch sentiment service against the configured AI agents base URL
- Emit enriched mentions for sink consumption
- Choose and document fallback behavior for AI failures

Verification:
- Batch request/response tests
- Failure-path tests

### Phase 6: Sink and Operational Counters

Goal:
- Persist enriched mentions to ClickHouse and update topic operational counters.

Work:
- Implement bulk insert into `listening_mentions` (batch size: 1000)
- Group inserts by topic_id, atomically increment `usage.mentions_count` in MongoDB
- If `mentions_count >= mentions_limit`, set `mentions_limit_reached = true` and send limit-reached notification
- Update `last_fetched_at` and `last_fetched_cursors` on topic document
- Add alert detection hooks: volume spike (2x 7-day avg), sentiment shift (avg < -0.3), limit reached
- Release per-topic Redis distributed lock when pipeline run completes
- On insert failure: retry 3x with backoff, produce to `listening-dlq` if all retries fail

Verification:
- ClickHouse insert tests
- Counter update / mentions_limit_reached flag tests
- Redis lock release tests
- DLQ production tests
- Idempotency/dedup tests

### Phase 7: Query APIs

Goal:
- Serve the listening mention and analytics endpoints directly from Go + ClickHouse.

Work:
- Add listening analytics repository
- Add listening analytics service
- Add listening handler package
- Register endpoints in `src/api/analytics/router.go`

Endpoints:
- `GET /api/v1/listening/mentions`
- `GET /api/v1/listening/analytics/summary`
- `GET /api/v1/listening/analytics/timeline`
- `GET /api/v1/listening/analytics/top-authors`
- `GET /api/v1/listening/analytics/top-posts`
- `POST /api/v1/listening/mentions/bookmark`
- `POST /api/v1/listening/mentions/tag`
- Optional: `GET /api/v1/listening/bookmarks`

Verification:
- Handler tests
- Repository tests
- Query correctness tests against fixture data

### Phase 8: Operational Jobs and Hardening

Goal:
- Add cron jobs and operational features not covered by the core pipeline phases.

Work:
- Data retention: monthly cron to drop ClickHouse partitions older than 12 months (both `listening_mentions` and `listening_daily_stats`)
- Sentiment retry: hourly cron to re-process mentions with `sentiment_label = "pending"` (query ClickHouse, re-call AI Agents, re-insert with updated `updated_at`)
- Monthly usage reset: daily cron to reset `usage.mentions_count` and `mentions_limit_reached` for topics whose billing period has elapsed (30+ days since `current_period_start`)
- DLQ monitoring: log/alert on `listening-dlq` message count

Verification:
- Partition drop tests against test ClickHouse
- Sentiment retry tests with mocked AI Agents
- Usage reset tests

---

## Suggested PR Breakdown

1. Foundations: models, config, schema (including materialized view), repositories, Kafka topic creation scripts
2. Trigger path: listening work API + scheduler branch + Kafka topics + Redis topic lock
3. Ingestion path: Data365 client (async POST→poll→GET) + fetcher (6 platform goroutines, IG hashtag handling)
4. Matching path: parser + keyword filters + author filters + dedup + content hash
5. Enrichment and sink: sentiment + ClickHouse write path + usage tracking + DLQ
6. Read APIs: handlers, services, ClickHouse query repository (mentions, summary, timeline, top-authors, top-posts, bookmark, tag)
7. Operational jobs: data retention, sentiment retry, usage reset, DLQ monitoring
8. Hardening: alerts, docs, operational polish

## Remaining Product Decisions

- Alert delivery channel (Pusher, Redis pub/sub, email — see PRD Section 9)
- Whether webhook alerts ship in v1
- Export handling location
- Initial sync depth if 30 days becomes too expensive
- JWT authentication strategy for Go endpoints (validate in Go or gateway-level?)

## Resolved Product Decisions (from cross-reference with Laravel)

- AI sentiment fallback: insert with `sentiment_label = "pending"`, retry hourly (PRD Section 19)
- Matching features: aligned with Laravel (`include_keywords` OR logic, no `exact_match`/`case_sensitive`)
- Workspace global excludes: deferred to v2 (no `listening_settings` dependency)
- Blocked authors / excluded subreddits: removed from v1 scope (per-topic filtering only)
- Topic billing: Laravel enforces per-workspace limit of 5, Go trusts MongoDB state
- Custom topic types: Go model includes `custom_type_id` for completeness, no CRUD needed in Go
- DLQ strategy: use option 1 — listening-specific handler writes failures to `listening-dlq` before returning

## Recommended First Build Target

Start with an end-to-end thin slice:

1. Topic exists in MongoDB
2. `POST /api/v1/listening-work` queues one topic
3. Fetcher pulls Data365 results for one platform
4. Parser applies matching for one topic
5. Sink writes one row into ClickHouse
6. `GET /api/v1/listening/mentions` returns that row

That slice will validate the repo fit before we scale out to all six platforms and the full matching surface.
