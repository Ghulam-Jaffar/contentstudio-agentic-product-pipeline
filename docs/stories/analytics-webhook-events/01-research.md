# Analytics Webhook Events — Research

Date: August 4, 2026

Scope: emitting outbound public webhooks for analytics data freshness and report generation, reusing the webhook delivery engine already built for publishing events.

---

## Current State

### The delivery engine is reused unchanged

Same position as the inbox webhook work. Everything on `contentstudio-backend` branch **`features`** is reused as-is: endpoint CRUD, secret rotation, test event, signing, retries with backoff, the API-credit-pool check, delivery logs, and CSV export. Nothing here rebuilds delivery.

The ingress topic `contentstudio.api.webhooks.events` is cross-service by design. Its envelope is `{ event_id, type, workspace_id, recipient_user_ids?, data }`, with `event_id` as the dedupe key, the type validated against the catalog enum, and routing always resolved server-side from current membership.

**The Go analytics pipeline already produces to Kafka** (`src/services/*/**-fetcher`, `src/cmd/jobs`, `src/cmd/api-server`), so publishing onto the ingress topic is plumbing it already has rather than a new capability.

### Sync events: a seam exists, but only on the path that matters least

Nine per-platform **immediate** processors already fire a Pusher notification for sync state — Facebook, Instagram, LinkedIn, YouTube, TikTok, Pinterest, Meta Ads, Google Ads, and Twitter:

| Function | Emits | Maps to |
|---|---|---|
| `sendPusherNotification` | `state: "Processed"` plus `last_analytics_updated_at` | `analytics.sync.completed` |
| `sendTokenExpiredNotification` | `state: "Failed"`, `error_type: "token_invalid"` | `analytics.sync.failed` |

Representative implementation: `src/services/facebook/facebook-immediate-processor/processor/processor.go:530-595`.

**The problem: these live only in the `*-immediate-processor` services**, which handle user-triggered syncs. The scheduled path runs fetcher → posts-parser → analytics-sink → clickhouse-sink, and **none of those emit any notification at all**, because historically nobody was watching an overnight run.

For webhooks that is exactly backwards. The unattended scheduled sync is the case a developer most needs told about. So `analytics.sync.completed` needs a **new emit point in the scheduled path**, not just a reuse of the immediate processor's existing call. This is the largest single piece of work in the epic.

### The token-expired gate inverts for webhooks

`sendTokenExpiredNotification` opens with an early return unless `syncType == "immediate"`, carrying the comment *"Gated on immediate syncs, since that is the only time a user is watching."*

Sound reasoning for a UI banner, wrong for a webhook. A scheduled sync failing silently at 3am is the highest-value failure event we could send. The webhook emit must sit outside that gate.

This is the same shape of trap as the notification-flag guard already flagged in the inbox stories, and worth calling out in the same way so it is not inherited by accident.

### Report events have a clean, idempotent choke point

`app/Services/Analytics/ReportCompletionService.php` is where every report outcome converges, for both scheduled and on-demand runs:

- `succeed($report, $url)` sets status `completed`, stores `export_url`, notifies, and records schedule history
- `fail($report, $reason)` sets status `failed` and records the reason

Both already carry `workspace_id`, `user_id`, report id, name, and type, which is everything a payload needs. `succeed()` already returns early when the report is finalized, so it is idempotent — a property that maps directly onto webhook deduplication.

`app/Models/Analytics/ReportRunHistory.php` tracks pending → running → success → failed → emailed. Report rendering itself lives in the Go `src/services/reports/` service, called through `GoReportsClient`, but completion is recorded in Laravel, so that is where the events belong.

### The frontend webhooks UI still does not exist

No webhook code in `contentstudio-frontend` on `develop`. The event picker this work adds to is specified but unbuilt, so the frontend story is blocked on the same two public-webhooks stories as the inbox one.

---

## What Needs to Change

**Laravel (small):**
- Add four cases to the webhook event catalog with labels
- Add an analytics payload builder alongside the existing post and inbox ones

**Laravel (report events):**
- Publish an envelope from `ReportCompletionService::succeed()` and `fail()`

**Go analytics pipeline (the real work):**
- Publish to the ingress topic from the immediate processors, where the Pusher seam already exists
- Add a **new** emit in the scheduled sink path, which has no seam today
- Emit sync failures regardless of sync type, outside the immediate-only gate

**Frontend (small, blocked):**
- Render the new event groups in the Webhooks tab once it exists

---

## Locked Event Catalog

| Event | Fires when | Routing |
|---|---|---|
| `analytics.sync.completed` | an account's analytics finish processing | broadcast |
| `analytics.sync.failed` | an analytics sync fails | broadcast |
| `report.completed` | a scheduled or on-demand report finishes | broadcast |
| `report.failed` | report generation fails | broadcast |

Naming follows the Standard Webhooks convention already locked for this programme: resource first, lowercase, dot-separated, past-tense action.

---

## Out of Scope

| Not building | Why |
|---|---|
| Metric threshold and anomaly alerts | Needs user-defined thresholds, comparison state, and dedup. That is an alerting product, and a webhook is the wrong mechanism for it |
| Per-post metric updates | Very high volume, values are recomputed continuously rather than changing at a moment, and there is no clear consumer |
| Competitor reports as their own event family | Reuse `report.completed` with a type field rather than splitting the catalog |
| `account.disconnected` | Surfaces during analytics sync but breaks publishing too, so it belongs in an accounts family rather than here. Worth raising as the natural next epic |

---

## Design

No `[Design]` story is needed. **[Design] Design the inbox event groups in the webhook event picker** already covers the grouped picker, including how it behaves as the list grows and whether groups collapse. Analytics adds two more groups to that established pattern. If that story has not landed when this work starts, it becomes a dependency rather than a duplicate.

---

## Files Involved

**contentstudio-backend** (branch `features`)
- `app/Data/Webhooks/Enums/WebhookEventType.php`
- `app/Services/Webhooks/PostWebhookPayload.php` (pattern to copy)
- `app/Services/Analytics/ReportCompletionService.php` (report emit point)
- `app/Models/Analytics/ReportRunHistory.php`

**contentstudio-social-analytics-go**
- `src/services/*/**-immediate-processor/processor/processor.go` across nine platforms (existing seam)
- `src/services/*/**-analytics-sink/` and `**-clickhouse-sink/` (new seam for scheduled syncs)
- Existing Kafka producer setup in `src/cmd/jobs` and the fetchers

**contentstudio-frontend**
- The Webhooks tab event picker, once it exists
