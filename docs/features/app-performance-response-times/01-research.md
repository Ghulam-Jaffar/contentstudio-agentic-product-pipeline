# Research: application response and query time

**Date:** 2026-09-07
**Trigger:** Response times and query times are slow throughout the app. The ask is a research story first, inside its own epic, with follow-up stories created from the findings.

This document is the pre-work for that spike, not the spike itself. It records what measurement infrastructure already exists, what is missing, and the hypotheses the spike should test. Nobody has to rediscover any of this.

---

## 1. Backlog check

No existing performance epic or story directory in `docs/features/` or `docs/stories/` (196 story directories checked). Nearest neighbours are data-pipeline stories (`analytics-data-fetch-job-split`, `analytics-observability-and-data-retention`, `analytics-php-to-golang-migration`), which are about analytics ingestion, not app-wide latency. This is new ground, not a duplicate.

---

## 2. The blocking problem: we are half blind

### Frontend tracing is on, backend tracing is off by default

`contentstudio-frontend/src/main.ts:446-467`

```
Sentry.browserTracingIntegration({ router })
tracePropagationTargets: ['localhost', 'staging.contentstudio.io', 'app.contentstudio.io', /^https:\/\/(?!um\.).*\.contentstudio\.io/]
tracesSampleRate: 0.1
```

The browser samples 10 percent of transactions and, because `tracePropagationTargets` covers the API domains, it **already sends `sentry-trace` and `baggage` headers to the backend on every API call**.

`contentstudio-backend/config/sentry.php:53`

```
'traces_sample_rate' => (float)(env('SENTRY_TRACES_SAMPLE_RATE', 0.0))
```

Default is `0.0`. Queue tracing (`SENTRY_TRACE_QUEUE_ENABLED`) also defaults to `false`.

**Consequence:** unless production explicitly sets these, every distributed trace dies at the API boundary. We can see in Sentry that a browser request took four seconds, and nothing at all about what the server did during those four seconds. Confirming the production values of `SENTRY_TRACES_SAMPLE_RATE` and `SENTRY_TRACE_QUEUE_ENABLED` in Doppler is the very first action of the spike. If they are zero, the spike cannot start until they are raised.

Sentry is **self-hosted** (`sentry-onpremise.contentstudio.io`). Raising the backend sample rate increases ingest and storage on infrastructure we own, so the rate needs to be chosen deliberately rather than set to 1.0.

### No server-side query visibility at all

Searched `app/`, `bootstrap/` and `config/` in `contentstudio-backend` for `DB::listen`, `enableQueryLog`, `whenQueryingForLongerThan` and any slow-query threshold. The only hit is a commented-out `enableQueryLog()` in `app/Repository/Notification/NotificationLumotiveRepository.php:13`.

There is no slow-query log, no query-count-per-request metric, no N+1 detection, and no Telescope, Clockwork or Debugbar in `composer.json`. Nothing in the application tells us how many database round trips a request makes or how long they take.

### What we do have

| Tool | Where | State |
|---|---|---|
| Sentry Laravel `^4.0` | backend `composer.json` | Installed. Error reporting working. Performance tracing gated on an env value that defaults to off. |
| `@sentry/vue` `10.48.0` | frontend `package.json` | Installed with browser tracing at 10 percent and trace propagation to the API. |
| Laravel Horizon `^5.0` | backend | Queue dashboard. Gives job throughput and wait times, useful for the async half of the picture. |
| Laravel Octane `^2.4` + Swoole | `docker/supervisord.conf:8` | The API runs on Octane with Swoole. Relevant because Octane keeps the framework in memory between requests, so per-request bootstrap cost is not the usual Laravel story, but leaked state and per-worker memory are. |
| MongoDB Atlas | primary datastore | Atlas has its own Profiler and Performance Advisor. This is very likely the highest-value data source we are not looking at, and it needs no code change to start reading. |
| ClickHouse | analytics reads | Has `system.query_log` with per-query timing, also free to read. |
| TanStack Query `5.99.0` | frontend | Client-side caching and deduplication already available. |

---

## 3. Hypotheses for the spike to test

These are hypotheses, deliberately not conclusions. The spike exists to confirm or kill them with data.

### H1. Database index coverage is unmanaged

`database/migrations` holds 49 migrations and only 5 mention an index. The primary store is MongoDB, where indexes are commonly created out of band in Atlas rather than in migrations, so index state is effectively undocumented and unreviewed. Atlas Performance Advisor will answer this immediately. Expect collection scans on high-traffic collections.

Highest-suspicion collections, by how often they are read: `social_integrations` (every page load), plans and posts (planner, calendar, approvals), inbox conversations and messages, workspaces and workspace team.

### H2. A fat bootstrap payload sits in front of every page

The SPA needs profile, workspaces, permissions and connected social accounts before it can render anything useful. `social_integrations` is a single collection holding every platform, with a per-platform global scope applied on top. If the bootstrap fetch pulls full account documents including tokens and queue slot arrays for every account in the workspace, large workspaces pay that cost on every cold load. Measure payload size and server time for the boot sequence specifically, separately from the page the user asked for.

### H3. N+1 query patterns in list views

Planner, approvals, inbox and analytics all render lists of entities that each carry a related account, author or workspace. With no query logging in place, an N+1 in any of these is currently invisible. Query-count-per-request instrumentation will find these in an afternoon.

### H4. Route-level code splitting is largely absent on the frontend

Across `contentstudio-frontend/src/` there are only 17 lazy route imports (`() => import(...)`) in module route configs. Vendor chunking is configured in `vite.config` with `manualChunks` for heavy libraries such as echarts, and `chunkSizeWarningLimit` is raised to 1000. That combination suggests large route components are being pulled into shared bundles rather than split per route. If the initial JS payload is multi-megabyte, time-to-interactive suffers everywhere and will be misattributed to "the API is slow".

### H5. Mixed data-fetching patterns cause avoidable refetching

120 files use TanStack Query, 48 still call axios directly. The direct callers get no deduplication, no caching and no request cancellation, so the same data can be fetched several times per navigation. Worth quantifying as request-count-per-navigation before deciding whether it matters.

### H6. Slowness is not uniform, and the average hides it

"Slow throughout the app" is the user-perceived symptom. It is usually a handful of endpoints on the hot path, plus one or two pathological large-workspace cases, rather than everything being uniformly slow. The spike should report p50, p75, p95 and p99 with call volume, never averages, and should check whether latency correlates with workspace size, account count or date range.

---

## 4. Surfaces to cover

The spike should measure the actual hot path, in rough order of traffic:

- Application boot: profile, workspaces, permissions, connected social accounts
- Dashboard and home
- Planner: calendar view, list view, feed view
- Composer: open, account selection, preview generation
- Analytics: overview report and the per-platform reports
- Social Inbox: conversation list and single conversation thread
- Approvals
- Discovery and content feeds
- Settings: team, social accounts

Cross-service note: analytics reads go Laravel to the Go analytics service to ClickHouse (`contentstudio-backend/app/Services/Analytics/AnalyticsPipelineClient.php` proxies to `contentstudio-social-analytics-go`), and inbox reads involve `social-inbox-manager`. A slow analytics report could be Laravel, the Go service, ClickHouse, or the hop between them. Without server-side tracing we cannot tell which, which is another reason phase one of the spike is turning tracing on.

---

## 5. Suggested measurement approach for the spike

1. Confirm and raise the backend trace sample rate. Start at 0.2 for a week, watch self-hosted Sentry ingest, adjust.
2. Turn on queue job tracing so async work is attributable too.
3. Add a temporary per-request query count and total query time, attached to the trace, so N+1s show up without reading code.
4. Pull the Atlas Profiler slow-query list and the Performance Advisor index suggestions for the same window. No code change needed, can happen in parallel with 1 to 3.
5. Pull ClickHouse `system.query_log` for the analytics read path over the same window.
6. Capture a frontend baseline for the same surfaces: initial bundle size, Largest Contentful Paint, Interaction to Next Paint, and requests per navigation.
7. Split each surface's total wall-clock into browser time and server time so follow-up work goes to the right team.

---

## 6. Notes on scope

- The spike must not fix anything beyond what is needed to measure. Every fix it identifies becomes its own story so it can be sized, prioritised and verified against the baseline.
- Enabling production tracing is a config and cost decision on self-hosted infrastructure. It could reasonably be split into its own small story if a different person owns infrastructure changes.
- Baseline numbers are the point. Without them the follow-up stories have no way to prove they worked, and the epic has no exit criteria.

---

## 7. Open questions for the PO

1. What is the target? Suggest agreeing a p95 target per surface, for example under 800ms server time for list views and under 2s for analytics reports, so the epic can be closed on evidence rather than vibes.
2. Is there a known worst case to reproduce? A specific large workspace, or a specific report that is reliably slow, shortens the spike considerably.
3. Do we have permission to raise self-hosted Sentry ingest volume, and who owns that cost?
4. Is frontend performance in scope for this epic, or is it a separate track? The research story measures both so time is attributed correctly, but the follow-up fix stories could split.
