# Research — Analytics observability, data retention, and trial data removal

Request: a story for observability and retention of analytics data, and removal of trial account data.

Three related but separable concerns over the same data. Each has a precedent inside the company, and one of those precedents has two defects worth not copying.

## Backlog check

Nothing covers any of the three. The nearest neighbours:

- `analytics-improvements` Story 6, *[BE] Investigate & fix the Sentry errors coming from the analytics Go service* — fixing specific errors, not adding observability.
- `inbox-stability-hardening` — the inbox equivalent, different service.
- `q2-misc-cleanup-and-ops` — mentions cleanup but not analytics data.

## Part 1: Observability

### What exists

`contentstudio-social-analytics-go` has **error reporting and a health check, and nothing else**:

- Sentry: `src/logger/sentry_hook.go`, `src/logger/sentry_writer.go`, `github.com/getsentry/sentry-go v0.39.0` in `go.mod`.
- `/health`, registered in `src/cmd/api-server/main.go` and allowlisted in `src/api/middleware/auth.go`.
- OpenTelemetry appears in `go.mod` **only as `// indirect`** dependencies, pulled in transitively by the GCP libraries. Nothing uses it directly.
- No `/metrics` endpoint, no Prometheus, no traces.

So we learn about analytics when something throws. We cannot answer the question support actually gets, which is *"why is this workspace's analytics stale?"* — that needs following one account's sync across the pipeline.

### The pipeline this has to cover

Per the platform docs the pipeline is five stages (Scheduler → Fetcher → Parser → Processor → Sink), and `src/services/` shows it is replicated per platform: `facebook`, `instagram`, `linkedin`, `twitter`, `tiktok`, `youtube`, `pinterest`, `gmb`, `meta-ads`, `google-ads`, plus `listening`, `thumbnails`, `reports`, `ai`, `unified`, `analytics`. Kafka moves work between stages.

That shape is why per-service error counts are not enough. A stalled sync needs to be traceable through stages and across platforms.

### The precedent to follow, including a decision already litigated

`social-inbox-manager/PRD-structured-logging-and-observability.md` is a **final** PRD for exactly this problem in a sibling service, and its 2026-06-05 amendment is the valuable part:

> **This supersedes the dedicated-table design described below.** … instead of a new `observability_logs` ClickHouse table + a SIM query layer + a new admin UI, the canonical records are **mapped onto the existing ContentStudio `logs.app_logs` table** and shown on the **existing admin `/logs` page**.

The architecture that survived:

- A **canonical log line** per unit of work, with a `trace_id` threaded through, including across Kafka.
- Written to Kafka, then into ClickHouse `logs.app_logs` via a Kafka engine table plus a materialized view (`app/observability/clickhouse/app_logs_ingest.sql`), alongside the existing feeder.
- **Read on the existing admin `/logs` page.** No new query layer, no new API, no new UI. The service is write-only and needs no ClickHouse connection.
- Identified by `log_name = <service>[:channel]`, and a flow followed by searching `trace_message = trace_id`.
- Reference architecture cited: `contentstudio-backend`'s own API request-log pipeline (Middleware → Kafka → ClickHouse).

The implementation lives in `social-inbox-manager/app/observability/`: `canonical.py`, `context.py`, `logging.py`, `trace.py`, `http_middleware.py`, `httpx_hooks.py`, `kafka_consume.py`, and the ClickHouse DDL.

**So the org already has a logging pipeline, a storage table and an admin UI for this.** The analytics story should map onto them rather than design a third thing. The inbox team built the dedicated-table version first and explicitly reversed it.

## Part 2: Retention

### There is no retention at all

**No `TTL` clause exists anywhere in `src/deployments/clickhouse/`.** Analytics data accumulates indefinitely.

What the tables do have is useful: they are `ReplacingMergeTree(updated_at)` **partitioned monthly**. From `schema/facebook_schema.sql`:

```
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(created_date)
ORDER BY (page_id, hash_id)
```

The same shape repeats for `facebook_posts` (`PARTITION BY toYYYYMM(created_time)`) and `facebook_media_assets` (`toYYYYMM(created_at)`), and across the per-platform schema files (`instagram_schema.sql`, `linkedin_schema.sql`, `gmb_schema.sql`, `pinterest_schema.sql`, `meta_ads_schema.sql`, `google_ads_schema.sql`, `listening_schema.sql`).

Monthly partitioning matters because **dropping a partition is near-free**, whereas a TTL expression or a row-level delete on a large table is a mutation. Any retention design should lean on the partitioning that already exists.

### The constraint that decides the number

Retention shorter than what the product lets a user ask for means the UI silently returns nothing. Analytics has date-range selection, custom ranges and scheduled reports comparing periods, so the retention window has to be reconciled against the longest range a user can select, not chosen in isolation. There is also a commercial dimension: agencies and enterprise contracts may assume multi-year history.

Retention may also differ by data class. Raw per-post rows, aggregated insights, and the data behind generated reports do not necessarily need the same lifetime.

## Part 3: Removal of trial account data

### The precedent exists, and is not running

`contentstudio-backend/app/Console/Commands/CleanupTrialInboxData.php` does exactly this shape of job for inbox:

```
protected $signature = 'trial:cleanup-inbox-data
                        {--dry-run       : Only show what would be deleted}';
protected $description = 'Delete inbox_comments / inbox_details / inbox_messages for '
                       . 'workspaces whose owner’s trial ended ≥ 30 days ago.';
```

Grace period is a 14-day trial plus 30 days. It resolves target workspaces, then issues one `deleteMany` per Mongo collection and logs the counts. Good bones: a dry run, a stated grace period, per-collection logging.

**Two defects to fix rather than inherit.**

**It is not scheduled.** Grepping the whole backend for `CleanupTrialInboxData` or `trial:cleanup-inbox-data` outside the command's own file returns nothing. It is not in `app/Console/Kernel.php` and not in `routes/console.php`. It is a manual command, so in practice trial inbox data is only removed when somebody remembers to run it.

**It infers the trial-end date from `updated_at`.** The code says so, in its own comment:

```
/* You store the trial state in `state = "trial_finished"`.
   We assume the change-time is `updated_at`.
   If you track it with a dedicated `trial_finished_at`, swap that in. */
$userQuery = User::where('state', 'trial_finished')
                 ->where('updated_at', '<=', $cutoff);
```

`updated_at` moves on any write to the user document. A user whose trial ended a year ago but whose record was touched last week looks recent and is **excluded from cleanup forever**. It fails safe (nothing is deleted early) but it fails silently, and the records most likely to be touched are the ones most likely to be skipped. A real `trial_finished_at` timestamp is the fix, and it benefits both the inbox job and the analytics one.

Also present: `app/Console/Commands/Authentication/TrialMonitorCommand.php` (trial state monitoring) and `app/Console/Commands/Storage/DeleteWorkspaceMediaCommand.php` (an existing workspace-scoped deletion command worth reading for its safety patterns).

### Why analytics is harder than inbox here

Inbox data is in MongoDB, where `deleteMany` is straightforward. Analytics data is in ClickHouse, where row-level deletes are mutations and expensive at scale. Deleting one workspace's rows out of monthly partitions shared with every other workspace is a different problem from dropping an old partition. That needs an explicit approach rather than an assumed one.

## Scope of the deliverable

Three backend stories. No UI, so no design story: the observability precedent explicitly reuses the existing admin `/logs` page rather than adding a screen.

1. Observability of the analytics pipeline
2. A defined and enforced retention policy
3. Removal of analytics data for accounts that never converted

## Files and systems involved

`contentstudio-social-analytics-go`:
- `src/logger/` — Sentry hook and writer, the current extent of observability
- `src/cmd/api-server/main.go`, `src/api/middleware/auth.go` — `/health`
- `src/services/*` — the per-platform pipeline stages
- `src/deployments/clickhouse/schema/*.sql` — table definitions, partitioning, no TTL
- `src/deployments/clickhouse/SCHEMA_CONVENTIONS.md`

`social-inbox-manager` (reference only):
- `PRD-structured-logging-and-observability.md`
- `app/observability/` and `app/observability/clickhouse/app_logs_ingest.sql`

`contentstudio-backend`:
- `app/Console/Commands/CleanupTrialInboxData.php` — the pattern and its defects
- `app/Console/Commands/Authentication/TrialMonitorCommand.php`
- `app/Console/Commands/Storage/DeleteWorkspaceMediaCommand.php`
- `app/Console/Kernel.php` — where scheduling belongs

## Mobile

None. Infrastructure and data lifecycle.
