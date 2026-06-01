---
title: Engineer Onboarding
description: Documentation for engineer onboarding
---

This guide helps you get productive quickly on the Go-based social analytics pipeline. It covers local setup, how the services fit together, common workflows, and where to look in the code.

## 1) Mental Model

Pipeline per platform (Facebook, Instagram, LinkedIn, TikTok):

MongoDB (accounts/tokens) → Kafka work orders → Fetcher (API calls) → Kafka raw → Parser (normalize) → Kafka parsed → ClickHouse Sink (batch insert)

Services are independent binaries that communicate only via Kafka topics. Each platform follows the same pattern; Facebook is the most complete reference implementation.

## 2) Repo Structure (high‑value locations)

- `src/cmd/services/*`: Service entry points (fetchers, parsers, sinks, immediate processors)
- `src/cmd/scheduler/account_fetcher`: Emits work orders from MongoDB accounts
- `src/internal/clients/`:
  - `social/`: API clients (e.g., `facebook.go`)
  - `clickhouse/`: batched insert helpers
- `src/internal/kafka/`: franz-go producer/consumer wrappers
- `src/internal/parsing/`: per‑platform parsers to normalize payloads
- `src/pkg/models/`: typed models for Kafka, ClickHouse, Mongo
- `src/internal/config/config.go`: env + defaults via Viper
- `src/sql/clickhouse_schemas/*`: ClickHouse DDL
- `../reference/kafka-topics.md`: canonical topic names per stage/platform
- `scripts/*`, `src/scripts/*`: convenience scripts (topic creation, service runners)

Good docs for deeper context: `CLAUDE.md` (repo root), `../architecture/social-media-integration-workflow.md`, `../platforms/facebook/analytics-pipeline-review.md`.

## 3) Prerequisites

- Go 1.21+
- Docker + Docker Compose
- Access to API credentials per platform
- Kafka, ClickHouse, MongoDB reachable locally (use docker-compose)

## 4) Configuration

- Copy `src/.env.example` to `src/.env` (if present); otherwise configure via env vars. Viper reads `APP_`‑prefixed env vars with dot keys flattened by `_`.
  - `APP_MONGO_URI`, `APP_MONGO_DATABASE`
  - `APP_KAFKA_BROKERS` (comma‑separated), `APP_KAFKA_TOPIC_PREFIX` (optional)
  - `APP_FACEBOOK_APP_ID`, `APP_FACEBOOK_APP_SECRET`
  - `APP_CLICKHOUSE_HOST`, `APP_CLICKHOUSE_PORT`, `APP_CLICKHOUSE_DATABASE`, `APP_CLICKHOUSE_USERNAME`, `APP_CLICKHOUSE_PASSWORD`
  - `APP_DECRYPTION_KEY` (base64 AES‑256 key for token decryption)

Important: `APP_KAFKA_TOPIC_PREFIX` is prepended verbatim. If you want `stg_...`, include the underscore in the prefix.

## 5) Local Stack

1. `cd src/deployments && docker-compose up -d` (Kafka, ClickHouse, Mongo)
2. Create topics: `src/scripts/create-topics.sh` (edit brokers if needed). See `../reference/kafka-topics.md` for full list.
3. Apply ClickHouse schemas: run all files in `src/sql/clickhouse_schemas/` against your DB.

## 6) Build & Run

- Build all services: `cd src && make build`
- Binaries output to `src/../bin/`

Typical Facebook end‑to‑end (same pattern for other platforms):

1) Start sink(s):
- `bin/facebook_clickhouse_sink`

2) Start parser:
- `bin/facebook_posts_videos_parser`

3) Start fetcher:
- `bin/facebook_fetcher`

4) Seed work orders (scheduler):
- `bin/account_fetcher -socialNetwork facebook -accountType page -syncType incremental`

You should see messages flow: work orders → raw_* topics → parsed_* topics → ClickHouse tables.

## 7) Topics You’ll Touch Most

- Work orders: `work-order-facebook`, `work-order-instagram`, `work-order-linkedin`, `immediate-work-order-*`
- Raw: `raw-facebook-posts`, `raw-facebook-videos`, `raw-facebook-insights`, etc.
- Parsed: `parsed-facebook-posts`, `parsed-facebook-media-assets`, `parsed-facebook-video-insights`, `parsed-facebook-insights`, etc.
See `../reference/kafka-topics.md` for the full matrix per platform.

## 8) Key Code Paths (Facebook reference)

- Fetcher: `src/cmd/services/facebook-fetcher/*`
  - Uses `src/internal/clients/social/facebook.go` (Graph API v19; appsecret_proof; pagination)
  - Publishes raw posts/videos/insights
- Parser: `src/cmd/services/facebook-posts-parser/main.go`
  - Parses to typed models via `src/internal/parsing/facebook_parser.go`
  - Publishes parsed topics
- Sink: `src/pkg/sinks/facebook_clickhouse.go` → `src/internal/clients/clickhouse/facebook.go`
  - Batched inserts for posts/media/video/insights
- Scheduler: `src/cmd/scheduler/account_fetcher/*`
  - Reads Mongo, emits `work-order-*` messages

## 9) Dev Tips

- Logging: Zerolog; look for Info/Warn/Error for state and failures. Avoid logging secrets.
- Kafka: franz-go; producers are sync now for simplicity; consumers poll and then commit; see review notes below.
- Parsing: Follow existing mapping for metrics and types exactly; parity with Python is important.
- ClickHouse: Always batch; keep arrays sanitized (no empty strings).

## 10) Troubleshooting

- No data in ClickHouse: ensure sink is running and schemas exist; confirm parsed topics receiving messages; verify consumer group IDs are unique per service.
- Rate limits/API errors: check fetcher logs; Graph API errors include `fbtrace_id`.
- Token issues: ensure `APP_DECRYPTION_KEY` is correct (base64 32‑byte key). Fetcher will fall back to plaintext token if decrypt fails.

## 11) Next Steps When You Add/Change Code

- Replicate the Facebook pattern for new platforms (models → client → fetcher → parser → sink).
- Keep models typed; update ClickHouse schemas if fields change.
- Run end‑to‑end locally before merging.
