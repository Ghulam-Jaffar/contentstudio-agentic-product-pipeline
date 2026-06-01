---
title: Migration Plan
description: Documentation for migration plan
---

## Overview

This document is the execution plan to migrate the legacy Python social analytics pipeline to the Go-based, Kafka + ClickHouse microservices architecture. It includes the very first step to stand up Docker-based local infrastructure and Kafka topic creation, detailed phases with deliverables and acceptance criteria, and a stretched milestone timeline. One engineer will lead the work, using Cursor/Codex to keep iteration fast and consistent.

## Objectives

- Functional parity with Python for all supported platforms
- Reliable, observable, and scalable pipeline with Kafka and ClickHouse
- Clear idempotency and retry/DLQ semantics to prevent data loss
- Reproducible local stack and automated CI/CD

## Success Criteria

- Parity: CH row counts and sampled field parity vs Python over agreed windows
- Reliability: no data loss on errors; DLQ and retry volumes within thresholds
- Observability: actionable metrics, dashboards, and alerts in place
- Operability: one-command local env; CI green; documented runbooks

## Assumptions

- Access to API credentials and MongoDB/ClickHouse/Kafka endpoints
- Ability to run Docker locally for development
- Python pipelines remain available for shadow comparison until cutover

## Phase Breakdown

### Phase 1 — Local Dev Stack + Topics (First)
- Docker Compose:
  - Add/finish `src/deployments/docker-compose.yml` for Kafka (KRaft), ClickHouse (server + client), MongoDB, and Kafka UI.
  - Healthchecks, volumes, single network, sane defaults.
- Topics bootstrap:
  - Finalize `src/scripts/create-topics.sh` to read from `../docs/reference/kafka-topics.md` and create topics idempotently.
  - Parameterize broker from `.env` (`APP_KAFKA_BROKERS`) or compose service name.
- Env templates:
  - Add/update `src/.env.example` with all `APP_*` variables required by `internal/config`.
- Acceptance:
  - `docker compose up -d` starts all services healthy.
  - `./src/scripts/create-topics.sh` creates all topics.
  - Kafka UI shows topics; ClickHouse/Mongo respond to pings.

### Phase 2 — Core Hardening (Before Feature Work)
- Kafka consumer semantics:
  - Mark records only on successful processing; commit marked offsets periodically.
  - Publish failures to DLQ topics (per platform/stage), include attempt count and error context.
- Secrets hygiene:
  - Mask/remove sensitive fields from config logs.
- Metrics foundation:
  - Add counters (processed, errors), histograms (latency, batch size), gauges (consumer lag if feasible).
  - Expose `/metrics` for Prometheus.
- Acceptance:
  - No data loss on simulated handler failures; DLQ receives failed messages.
  - Metrics visible locally; secrets not printed in logs.

### Phase 3 — Facebook Parity (Reference)
- API parity:
  - Verify endpoints/params/pagination/rate limits vs Python implementation.
- Parser parity:
  - Confirm field mappings in `internal/parsing/facebook_parser.go` and models match ClickHouse schema.
- CH parity:
  - Align `src/sql/clickhouse_schemas/facebook_schema.sql` and `pkg/models/clickhouse/facebook.go`.
- Tests:
  - Golden JSON fixtures → parsed structs; e2e smoke from raw → parsed → CH.
- Acceptance:
  - 7-day sample parity (row counts and selected fields); stable processing with low DLQ.

### Phase 4 — Instagram Migration Completion
- Implement/align client, fetcher, parser, sink; handle reels/media/insights.
- Ensure topics match `../docs/reference/kafka-topics.md`.
- Tests + e2e smoke.
- Acceptance: sample parity vs Python; CH tables populated and queryable.

### Phase 5 — LinkedIn Migration Completion
- Posts/images/videos/stats parity; sink batching; error handling.
- Tests + e2e smoke.
- Acceptance: parity on selected period; stable lag/throughput.

### Phase 6 — TikTok Migration Completion
- OAuth + backoff + fetcher/parser/sink coverage.
- Tests + e2e smoke.
- Acceptance: parity vs Python for selected accounts/days; stable operation.

### Phase 7 — Remaining Platforms (Twitter, Pinterest, YouTube)
- Parity spec from Python for each platform (endpoints, fields, limits).
- Implement 5-stage pipeline; update/add ClickHouse schemas as needed.
- Tests + e2e smoke.
- Acceptance: 7-day parity per platform within tolerances.

### Phase 8 — Scheduler + Rate Limiting
- Enhance `cmd/scheduler/account_fetcher` with cadence windows and priorities.
- Implement exponential backoff per platform; avoid rate-limit violations.
- Acceptance: work-order volumes match plan; no throttle events.

### Phase 9 — Reliability + Observability
- Retries:
  - Capped exponential retries before DLQ; add attempt metadata in headers.
- Idempotency:
  - Document and validate CH primary keys/order-by; add hashing where beneficial.
- Dashboards + Alerts:
  - Grafana dashboards for throughput, errors, lag, CH latencies; alerts on thresholds.
- Acceptance: SLOs/alerts in place; runbook drafted and reviewed.

### Phase 10 — CI/CD
- CI (GitHub Actions):
  - Build, lint, unit tests; light integration with service containers (Kafka/CH/Mongo).
- CD:
  - Build versioned Docker images; env-specific manifests (Compose/K8s); secrets from vault/SSM.
- Acceptance: green pipeline; tagged releases; one-command deploys per env.

### Phase 11 — Data Backfill + Validation
- Backfill:
  - Historical fetch in bounded windows; optionally use separate topics.
- Validation:
  - Daily CH counts vs Python; random sampled field diffs with tolerances.
- Acceptance: >99.5% row parity; 100% sampled field match or justified differences.

### Phase 12 — Rollout + Cutover
- Shadow:
  - Dual-run Python and Go; nightly diffs and reports.
- Canary:
  - Route a subset to Go-only; monitor lag/errors/CH parity.
- Full cutover:
  - Switch all to Go; keep Python on hot standby for stability window; decommission thereafter.
- Acceptance: clean diffs; incident-free window; successful decommission.

### Phase 13 — Documentation + Handover
- Update:
- `../docs/guides/engineer-onboarding.md`, per-service READMEs, parity docs, runbooks, dashboard links.
- Handover: ops + dev sessions; drill on incident runbooks and dashboards.

## Deliverables per Phase
- Phase 1: `src/deployments/docker-compose.yml`, `src/scripts/create-topics.sh`, `src/.env.example`.
- Phase 2: Updated `internal/kafka/consumer.go` (mark/commit), DLQ definitions + publisher helper, masked config logs, `/metrics` exposure.
- Phases 3–7: Parity specs, parser tests, CH DDL updates, e2e smoke.
- Phases 8–9: Scheduler cadence config, backoff, dashboards, alerts, runbooks.
- Phase 10: CI workflows, Docker images, deployment manifests.
- Phases 11–12: Backfill scripts/playbooks, validation reports, cutover plan.
- Phase 13: Final docs and training materials.

## Stretched Milestones (Single Engineer, assisted by Cursor/Codex)

- Weeks 1–2
  - Phase 1: Docker Compose + topics + .env template
  - Phase 2 (start): consumer commit semantics, DLQ foundations, secret masking, basic metrics
- Weeks 3–4
  - Phase 2 (finish): metrics endpoint and local dashboard scaffolding
  - Phase 3: Facebook parity (API, parser, CH alignment, tests, e2e smoke)
- Weeks 5–6
  - Phase 4: Instagram migration completion (reels/insights), tests and e2e
- Weeks 7–8
  - Phase 5: LinkedIn migration completion, tests and e2e
- Weeks 9–10
  - Phase 6: TikTok migration completion, tests and e2e
- Weeks 11–14
  - Phase 7: Twitter, Pinterest, YouTube migrations (stagger if API complexity requires)
- Weeks 15–16
  - Phase 8: Scheduler cadence + rate limiting
  - Phase 9: Reliability (retries) + Observability (dashboards/alerts) + runbooks
- Weeks 17–18
  - Phase 10: CI/CD pipelines, versioned images, env manifests
- Weeks 19–20
  - Phase 11: Backfills (if needed) + validation tooling/reports
- Weeks 21–22
  - Phase 12: Shadow, canary, cutover; stability window; Python decommission prep
- Week 23
  - Phase 13: Documentation refresh + handover sessions

Notes:
- Cursor/Codex can accelerate boilerplate (models, topic wiring, tests) and repetitive schema alignment; still plan for external API nuance and validation time.
- If parallel execution is possible (e.g., while long backfills run), some later activities can overlap.

## Immediate Next Actions (Day 1–3)
- Create `src/deployments/docker-compose.yml` with Kafka, Kafka UI, ClickHouse (server+client), MongoDB.
- Finalize `src/scripts/create-topics.sh` using `../docs/reference/kafka-topics.md` and compose service names.
- Add `src/.env.example` aligned to `internal/config/config.go` (Kafka, CH, Mongo, Facebook app secrets, decryption key, prefixes).
- Spin up the stack and verify topics creation + basic ping checks.

## References
- Topic map: `../reference/kafka-topics.md`
- Onboarding: `../guides/engineer-onboarding.md`
- Architecture & review: `../architecture/project-overview-and-code-review.md`
