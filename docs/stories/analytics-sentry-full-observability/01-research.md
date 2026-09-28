# Research: Sentry full transaction tracing and observability for analytics

Request (2026-09-24): one ticket, requirements only, non-technical. Sentry for all transactions and observability across analytics, not just error logging. The PO asked that it stand alone and not depend on or reference the existing `analytics-observability-and-data-retention` or `analytics-improvements` stories.

## Current State

`contentstudio-social-analytics-go` already uses Sentry (`github.com/getsentry/sentry-go v0.39.0`):

- **Error reporting:** `src/logger/sentry_writer.go` and `src/logger/sentry_hook.go` forward Error+ log lines to Sentry, with a `sentry_captured` marker so the same error is not sent twice.
- **API transactions, sampled:** commit `1aa26713` (2026-09-24, "enable sampled Sentry API transactions") turned on tracing in `src/cmd/api-server/main.go` only, configured through `src/config/config.go`. `EnableTracing` / `TracesSampleRate` are read from env in `src/logger/sentry_hook.go`.
- **Environment / release:** env vars `APP_SENTRY_ENVIRONMENT`, `APP_SENTRY_RELEASE`, `APP_RELEASE_VERSION` (`src/observability/unit.go`).
- **Not covered:** the background workers, which are most of the service. Per-platform scheduler, fetcher, parser, processor and sink services under `src/services/` (facebook, instagram, linkedin, twitter, tiktok, youtube, pinterest, gmb, meta-ads, google-ads, listening, reports, …) and the `src/cmd/` binaries (`jobs`, `competitor-jobs`, `report-generator`, `competitor-insights-filler`, `validate-tokens`, …) produce no transactions. Kafka moves work between stages, so tracing one account's sync end to end requires passing trace context along with each Kafka message.

## What Needs to Change

- Extend tracing from the API server to every worker and scheduled binary
- Pass trace context through Kafka so a sync reads as one trace across stages
- Spans for outbound platform API calls, ClickHouse, MongoDB and Redis
- Tag every event with platform, stage, workspace, account and environment/release
- Sentry Cron Monitors for scheduled jobs
- Alerts and a shared dashboard
- Make sample rates configurable per environment and per service, and always capture errors
- Scrub tokens, secrets and customer content

## Files Involved

- `src/logger/sentry_hook.go`, `src/logger/sentry_writer.go`
- `src/config/config.go`
- `src/cmd/*/main.go`
- `src/services/*/*/main.go` (worker entry points)
- Kafka producer/consumer wrappers
- `src/observability/`
