---
title: Project Overview and Code Review
description: Documentation for project overview and code review
---

This document summarizes the architecture, key components, and an in‑depth review of the Go implementation, with a migration comparison to the legacy Python code under `old-python-code/`. It highlights strengths, risks, and concrete recommendations.

## 1) High‑Level Architecture

- Pattern per platform: Scheduler → Fetcher → Parser → Processor (optional) → ClickHouse Sink
- Transport: Kafka topics (work_order → raw → parsed)
- Storage: ClickHouse for analytics, MongoDB for accounts/schedules; optional Redis in future for retries
- Observability: Zerolog logging; Prometheus metrics planned but not wired

Core services live in `src/cmd/services/` and `src/cmd/scheduler/`. Internal building blocks live under `src/internal/` (clients, kafka wrappers, config, parsing), and shared models under `src/pkg/models/`.

Reference documents: `README.md`, `CLAUDE.md`, `../reference/kafka-topics.md`, `./social-media-integration-workflow.md`.

## 2) Data Flow by Stage (Facebook)

1. Scheduler (`src/cmd/scheduler/account_fetcher`):
   - Reads `facebook_accounts` from Mongo via repo (`src/internal/repository/mongodb/facebook_account_repository.go`)
   - Emits `work-order-facebook` with decrypted/derived fields coming from ExtraData

2. Fetcher (`src/cmd/services/facebook-fetcher`):
   - Consumes work orders, decrypts long‑lived tokens using `internal/crypto/token_utils.go`
   - Calls Graph API via `internal/clients/social/facebook.go`:
     - Posts (fields in `postFields`), Videos (`videoFields`), Insights (daily + demographic + page info)
     - appsecret_proof, pagination, rate‑limit aware logging
   - Publishes raw messages: `raw-facebook-posts`, `raw-facebook-videos`, `raw-facebook-insights`

3. Parser (`src/cmd/services/facebook-posts-parser`):
   - Consumes raw topics; routes by topic suffix
   - Normalizes via `internal/parsing/facebook_parser.go` to typed `pkg/models/kafka/facebook.go`
   - Publishes parsed topics: `parsed-facebook-posts`, `parsed-facebook-media-assets`, `parsed-facebook-video-insights`, `parsed-facebook-insights`

4. ClickHouse Sink (`src/pkg/sinks/facebook_clickhouse.go` → `src/internal/clients/clickhouse/facebook.go`):
   - Converts parsed models to ClickHouse row structs in `pkg/models/clickhouse/facebook.go`
   - Performs batched inserts with `clickhouse-go` using `PrepareBatch`/`Send`

The same pattern exists for Instagram, LinkedIn, and TikTok (varying completeness). Kafka topics are centralized in `../reference/kafka-topics.md`.

## 3) Key Components and Quality Review

### 3.1 Config (`src/internal/config/config.go`)
- Uses `godotenv` (optional) + `viper` for env‑first configuration.
- Helpful defaults, SASL support for Kafka.
- Note: The current implementation logs the entire config (`loaded_config`). This risks leaking secrets (Kafka SASL, ClickHouse password, decryption key). Recommendation: remove/obfuscate sensitive fields before logging.

### 3.2 Kafka Wrappers (`src/internal/kafka/*.go`)
- franz‑go (`kgo`) based Producer/Consumer with SASL support and reasonable batching/lingering defaults on producer.
- Producer uses `ProduceSync` for simplicity; OK to start, but throughput can benefit from async `Produce` with a result handler if/when needed.
- Consumer: good structure and logging, but one critical issue (see “Risks” below on offset commits).

### 3.3 Social API Clients (Facebook example)
- `src/internal/clients/social/facebook.go`:
  - Preserves Python request shape: fields for posts/videos, daily + demographic insights.
  - Uses `appsecret_proof` and thorough error decoding (fbtrace_id, code, type).
  - Paginates with a stop cap (`maxPagesToFetch`).
  - Returns raw models defined in `pkg/models/kafka/facebook.go`.

### 3.4 Parsing
- `src/internal/parsing/facebook_parser.go`:
  - Comprehensive mapping of insights and reactions with type‑safe assignments.
  - Robust helpers for time and dynamic fields; explicit handling of carousel/media assets.
  - Aligns with ClickHouse schema expectations.

### 3.5 ClickHouse Client/Sinks
- `src/internal/clients/clickhouse/*` and `src/pkg/sinks/*`:
  - Proper use of `PrepareBatch` and `Send` for throughput.
  - Array field sanitization to drop empty strings.
  - Separate bulk methods per table (posts, media_assets, video_insights, insights).

### 3.6 Mongo Models + Repos
- `pkg/models/mongo` uses `MongoTime` to handle legacy string timestamps or BSON DateTime seamlessly.
- `internal/repository/mongodb/facebook_account_repository.go` exposes explicit update methods for last‑updated timestamps.

## 4) Python vs Go — Migration Notes

Legacy code lives in `old-python-code/analytics`. Highlights:

- Python uses Dramatiq workers, Redis sorted sets, and direct ClickHouse HTTP client; Go uses Kafka + microservices + batched inserts.
- Python covers more platforms (Twitter, Pinterest, YouTube) and competitor flows; Go currently focuses on Facebook/Instagram/LinkedIn/TikTok.
- Business logic parity goal is mostly met where Go exists (e.g., Facebook fields), with equivalents for metrics and aggregations; Facebook is the strongest reference.
- Token handling: Python’s `decrypt_token` is preserved conceptually via `internal/crypto/token_utils.go` (AES‑256‑CBC w/ base64 JSON payload). Backward-compatibility fallback if value isn’t base64‑looking.

Where to compare for parity:
- Facebook posts/videos: `old-python-code/analytics/social_channels/facebook/facebook_analytics.py` vs `internal/clients/social/facebook.go` + parser.
- ClickHouse models: Python ClickHouse models under `old-python-code/analytics/models/clickhouse_models/analytics/facebook/*` vs `pkg/models/clickhouse/facebook.go` and `src/sql/clickhouse_schemas/facebook_schema.sql`.
- Schedulers/queues: Python’s Redis sorted sets vs Go’s Kafka `work-order-*` topics.

## 5) Risks, Gaps, and Recommendations

### 5.1 Kafka Consumer Offset Semantics (Important)
- Current consumer (`internal/kafka/consumer.go`) polls, processes, and then calls `CommitUncommittedOffsets(ctx)` without marking records. In franz‑go, you typically call `MarkCommitRecords` (or `AllowAutoCommit`) to commit only after successfully processing a record. As written, the consumer risks committing offsets for messages that failed processing, leading to potential data loss.
- Recommendation:
  - For each successfully processed `record`, invoke `client.MarkCommitRecords(record)` (or `MarkOffsets` equivalent), and then periodically call `CommitMarkedOffsets`.
  - For failures, do not mark; add retry/parking logic and/or send to a dead-letter topic.

### 5.2 No DLQ/Retry Strategy
- Errors in handlers are logged and then discarded. There is no DLQ (`dead_letter_*`) or retry backoff.
- Recommendation: add a per-platform DLQ topic and a simple retry policy (with max attempts and exponential backoff). Track error counts with metrics.

### 5.3 Secrets in Logs
- `config.LoadConfig()` logs the full config struct. This can leak credentials/secrets.
- Recommendation: mask or remove sensitive fields (passwords, tokens, keys) before logging.

### 5.4 Observability
- No Prometheus metrics are wired (despite being in the tech stack). Lacking metrics for consumer lag, processed counts, error rates, and batch sizes.
- Recommendation: add `promhttp` exporter and standard counters/histograms in fetchers/parsers/sinks; optionally expose health endpoints.

### 5.5 Validation and Idempotency
- Sinks assume idempotency at the DB layer via primary keys/ordering. Ensure ClickHouse insert patterns and dedup expectations align with partitioning/order by keys for each table.
- Recommendation: document idempotency strategy and add lightweight dedup logic if necessary (e.g., hash keys) to avoid duplicate inserts with at‑least‑once delivery.

### 5.6 Configuration Ergonomics
- `APP_KAFKA_TOPIC_PREFIX` concatenation does not add separators; accidental `prod` vs `prod_` mismatch can create surprise topics.
- Recommendation: normalize prefix (e.g., auto‑append `_` when non‑empty and missing) or document clearly (already documented in onboarding).

### 5.7 Coverage Gaps vs Python
- Twitter, Pinterest, YouTube flows in Go are not present or are partial.
- Recommendation: migrate next platforms following Facebook pattern; validate parity against Python models and SQL.

### 5.8 Tests and CI
- No automated tests; manual validation only.
- Recommendation: introduce unit tests for parsers and token crypto, plus a small integration test for sink batching (local Docker services). CI can lint, build, and run unit tests.

## 6) Security Considerations

- Token security: Decrypt only in memory; never log tokens; ensure `APP_DECRYPTION_KEY` length (base64 32‑byte) is validated.
- Kafka SASL credentials and ClickHouse passwords must be stored in env/secrets; avoid printing in logs.
- PII in logs: none observed; continue to avoid user‑identifiable data in structured logs.

## 7) Performance and Throughput

- Batch inserts (ClickHouse) in place; parsers produce per item; fetcher uses parallel goroutines to publish.
- For higher throughput:
  - Use async producer with batching (franz‑go `Produce` + `FetchProduceResponses`).
  - Tune consumer fetch sizes and increase worker concurrency in parsers/sinks.
  - Add backpressure accounting with buffered channels + metrics.

## 8) Developer Pointers by Module

- Config: `internal/config/config.go` — Viper with env replacer; defaults present; remove logging of secrets.
- Kafka: `internal/kafka/` — add mark/commit semantics and DLQ.
- Clients:
  - Facebook: `internal/clients/social/facebook.go` — solid parity, add backoff on 429s.
  - ClickHouse: `internal/clients/clickhouse/*` — keep batches around 1k rows; consider retry on transient errors.
- Parsing: `internal/parsing/*` — maintain exact field mapping parity with Python; ensure string/array coercions match ClickHouse schema.
- Models: `pkg/models/*` — maintain “FacebookTime” and “MongoTime” helpers for edge timestamp formats.
- Scheduler: `cmd/scheduler/account_fetcher/*` — casing mapping for account types is currently manual; extract into a shared normalizer if more cases appear.

## 9) Old‑Python Code: What to Reference

Key files when you need business‑logic parity:

- `analytics/social_channels/facebook/facebook_analytics.py` and related `*_sql.py` under `analytics/models/clickhouse_models/`.
- Dramatiq tasks under `analytics/dramatiq_tasks/facebook/*` for operational sequencing and post‑fetch processing.
- Settings in `analytics/settings.py` for legacy envs and defaults.

General migration guidance also captured in `CLAUDE.md` and `./social-media-integration-workflow.md`.

## 10) Actionable Next Steps

1) Fix consumer commit semantics and add DLQ topics across services.
2) Remove/obfuscate secret fields from config logs.
3) Add basic Prometheus metrics (processed, errors, lag, batch sizes).
4) Migrate next platform(s) from Python (Twitter or YouTube), following Facebook as blueprint; align ClickHouse schemas.
5) Add parser unit tests and crypto tests; wire a minimal CI build/test workflow.
6) Document idempotency assumptions per table and validate ClickHouse primary keys/order by.

With these addressed, the Go pipeline will be safer under failure, easier to operate, and closer to full parity with production Python flows.
