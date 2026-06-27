# Research — Q2 Misc batch: Inbox data deletion, Notification panel redesign, Stuck scheduled posts, Horizon jobs improvement

Four independent stories requested together. All land under the **Q2 - 2026: Miscellaneous** epic.

| # | Story | Type | Codebase |
|---|---|---|---|
| 1 | Stop fetching + purge inbox data for inactive accounts | `[BE]` | `social-inbox-manager/` (Python) |
| 2 | Redesign + unify the two desktop-rail notification panels | `[Design]` | `contentstudio-frontend/` (Vue 3) |
| 3 | Sweep stuck `scheduled` posts and silently mark failed | `[BE]` | `contentstudio-backend/` (Laravel) |
| 4 | Horizon / queued-jobs reliability + observability | `[BE]`/chore | `contentstudio-backend/` (Laravel) |

---

## Story 1 — Inbox data deletion for inactive accounts

### Current State

**The analytics pattern we're mirroring** (`contentstudio-social-analytics-go/`):
- Account filtering happens **at scheduler time**, before any fetch is queued.
- The scheduler query (`db/mongodb/client.go` → `buildNeedingUpdateFilter()`) only returns accounts where:
  - `validity = "valid"`
  - `state ∈ {added, syncing, processed, failed}` — **never `deleted`**
  - `super_admin_state ∈ {active, past_due}` — excludes `cancelled`
- `super_admin_state` lives on the `social_integrations` collection and is driven by the backend Workspace/subscription state.
- **Analytics never purges historical data** — it simply stops ingesting new rows. (This is the key difference from what we want for inbox — see "What Needs to Change".)

**The inbox side** (`social-inbox-manager/`):
- Per-platform sync jobs (`app/jobs/<platform>_inbox_job.py`, e.g. `facebook_inbox_job.py`) currently query accounts with `{"validity": "valid", "type": "Page"}` — **no account/workspace status check**. Stopped/deleted accounts still get sync jobs queued.
- Account repositories (`app/database/mongo/repository/<platform>_account_repository.py`) have `find_by_filters()` with no status filtering. One repo per platform: Facebook, Instagram, YouTube, LinkedIn, GMB.
- Inbox data lives in MongoDB collections `inbox_details`, `inbox_messages`, `inbox_comments`, keyed by `workspace_id` + `account_id`.
- Webhook ingestion (`app/workers/<platform>_inbox_worker.py`, Kafka fan-out) also has **no account-status pre-filter** before saving.
- The inbox service learns account/workspace status by **directly querying the shared `social_integrations` collection** — there's no Kafka event stream for status changes today.
- **No existing cleanup/purge/retention logic** for inbox data.

### What Needs to Change

- **Stop fetching (parity with analytics):** add an account/workspace status filter to the per-platform sync jobs (and their account repositories) so inactive accounts never get a sync job queued. Mirror the analytics rule: keep fetching only for `valid` + `super_admin_state ∈ {active, past_due}` + `state ≠ deleted`.
- **Filter webhooks (recommended):** skip/short-circuit webhook processing for accounts that are inactive/deleted.
- **Purge existing data (this is the part analytics did NOT do):** a cleanup job that deletes existing `inbox_details` / `inbox_messages` / `inbox_comments` for accounts/workspaces that are deleted or expired.

### Open decisions (see review gate)
- Which states stop fetching? (Proposed: deleted accounts + workspaces that are cancelled/trial-expired — i.e., anything not `active`/`past_due`.)
- "Handle existing data" = hard delete vs soft delete/archive vs retention window?

### Gotchas
- **Workspace vs account status** — inbox must consider both the account's `super_admin_state` and the workspace subscription state (analytics only checks account level).
- **One account can belong to multiple workspaces** — stopping/purging for one workspace must not affect the account in another. Key purge on `workspace_id` + `account_id`, not `account_id` alone.
- **Webhook timing** — webhooks may still arrive for a just-deleted account; pre-filtering in the worker handles the race.

### Files Involved
- `social-inbox-manager/app/jobs/facebook_inbox_job.py` (+ instagram/youtube/linkedin/gmb siblings)
- `social-inbox-manager/app/database/mongo/repository/facebook_account_repository.py` (+ siblings)
- `social-inbox-manager/app/workers/facebook_inbox_worker.py` (+ siblings)
- Shared `social_integrations` collection (status source of truth)
- New: an inbox-data cleanup/purge job

---

## Story 2 — Redesign + unify the two desktop-rail notification panels (Design)

### Current State
Two notification panels both anchor at the bottom of the desktop left rail (`contentstudio-frontend/src/components/layout/DesktopNavigationRail.vue` — two icon buttons at the bottom: a Stamp icon for approvals, a BellRing icon for general):

1. **Approval notifications panel (newer)** — `src/modules/approval-workflows/components/ApprovalNotificationsPanel.vue`
   - Modern: fully `@contentstudio/ui` (Icon, Button, ActionIcon) + Tailwind.
   - Gradient header with stacked icon/title/subtitle; "Unread" / "All" pill filters with counts.
   - Tone-aware item icons (green = approved, red = rejected, yellow = pending); unread row tint; blue read-dot; mark-read via hover Check icon; "Open Post" footer link.
   - Rich empty state (icon + title + body + CTA to planner).
   - Data via TanStack Query (`useApprovalNotifications`) with optimistic mark-read; Pusher hook for live additions.

2. **General notifications panel (older)** — `src/components/common/TopNotificationDropdown.vue`
   - Legacy: mixes `@contentstudio/ui` Icon with old `icon-*-cs` icon-font classes and `_dropdown.css`.
   - Plain header, simple h2 + generic settings link; four tabs (All, System, Team, Inbox).
   - All item backgrounds default to blue (`--cstu-info-500`) regardless of type; read state via `bg-gray-300`; whole-item click to navigate; mark-all-read text link.
   - Minimal empty state (image + message).
   - Data via Pinia (`useWorkspaceNotificationStore`) with `vue-infinite-loading` pagination.

### Key visual/UX inconsistencies to reconcile
- **Icon + color system** — approval uses semantic tone colors; general defaults everything to blue.
- **Component library** — approval is fully modern `@contentstudio/ui`; general still uses legacy icon-font classes.
- **Header/layout** — approval has a gradient header and more whitespace; general is condensed/legacy.
- **Filters** — approval = "Unread/All" pills with counts; general = 4 icon tabs.
- **Empty state** — approval = rich CTA; general = minimal image.
- **Read/unread + mark-read interaction** — different indicators and interaction models.

### What this story is
A **design deliverable** (Figma) defining one consistent notification-panel pattern that both panels adopt: shared header, list-item, read/unread indicator, filter, empty/loading/error states, icon + color system, and `@contentstudio/ui` component usage. Implementation (FE) is a follow-up, not part of this story.

### Files Involved (for the designer's reference)
- `contentstudio-frontend/src/components/layout/DesktopNavigationRail.vue`
- `contentstudio-frontend/src/modules/approval-workflows/components/ApprovalNotificationsPanel.vue`
- `contentstudio-frontend/src/components/common/TopNotificationDropdown.vue`

---

## Story 3 — Sweep stuck `scheduled` posts and silently mark failed

### Current State
- **Model:** `contentstudio-backend/app/Models/Publish/Planner/Plans.php` (collection `plans`).
- **Status field:** `status`; values in `app/Data/Enums/PostStatus.php`: `draft, scheduled, published, failed, review, rejected, processing, queued, on_hold, deleted, missed_review`.
- **Scheduled time:** `execution_time` (object with `.date`).
- **Publishing flow:** the `plan:posting` command (`app/Console/Commands/Planner/PlanPostingCommand.php`, run via Argo CronWorkflows) fetches overdue scheduled plans, adds each plan id to the Redis set `plan_posting`, flips status to `queued`, and dispatches `PlanPostingJob`. `PlanFinalizerJob` later writes the final status.
- **The stuck scenario:** a plan's job never executes (stuck delayed job / Redis), its `execution_time` passes, status stays `scheduled` (or `queued`), no processing, no status change.
- **Normal failure path** (`PlanFinalizerJob`): calls `PlansRepository::updatePlanDetails(...,['status'=>'failed'])` (an Eloquent `.update()` → **bumps `updated_at`**) and then `broadcastEvent()` → fires in-app + email notifications and writes a PlanActivity. This is exactly what we must NOT do here.
- **"Updated" sort:** `PlansRepository` grid queries sort on `updated_at` — any normal update floats the post to the top.

### What Needs to Change
- A scheduled sweep (new artisan command, sibling to `MissedPlansCommand.php`) that finds plans **stuck past their `execution_time`** in `scheduled` (and likely `queued`) status, beyond a safe grace window, and marks them `failed`:
  - **Silently** — no `broadcastEvent()`, no in-app/email notification.
  - **Without bumping `updated_at`** — use a raw MongoDB `$set` (`Plans::raw(fn($c)=>$c->updateOne(['_id'=>...],['$set'=>['status'=>'failed']]))`), not `PlansRepository::updatePlanDetails()` / Eloquent `.update()`. Precedent exists (`database/migrations/2026_02_26_000000_backfill_plan_comment_counts.php` uses raw bulkWrite).
  - **Clean up Redis** — `srem('plan_posting', $planId)` for swept plans.

### Open decisions (see review gate)
- **Grace window:** how far past `execution_time` before a post is considered stuck (must exceed the normal posting window — the posting supervisors allow up to ~4h timeouts). Proposed default: `execution_time` older than **6 hours** (safely past any in-flight posting). Tunable.
- **Statuses swept:** `scheduled` only, or also `queued`? (Proposed: both, since a stuck job can leave a plan in either.)
- **Audit trail:** write a silent PlanActivity / log entry for traceability (no user-facing notification)? Proposed: yes, log only.

### Gotchas
- Must not race the live posting flow — hence the grace window.
- Use the exact enum string `failed` (case-sensitive).
- Don't route through `PlanFinalizerJob` / `updatePlanDetails` — both notify and bump `updated_at`.

### Files Involved
- `contentstudio-backend/app/Models/Publish/Planner/Plans.php`
- `contentstudio-backend/app/Data/Enums/PostStatus.php`
- `contentstudio-backend/app/Repository/Publish/Planner/PlansRepository.php` (read fetch patterns; do NOT use its update helper here)
- `contentstudio-backend/app/Console/Commands/Planner/PlanPostingCommand.php` (Redis-set + queue precedent)
- `contentstudio-backend/app/Console/Commands/Planner/MissedPlansCommand.php` (closest command pattern to follow)
- `contentstudio-backend/app/Jobs/PlanFinalizerJob.php` (the notify path to avoid)
- New: stuck-plans sweep command + matching Argo CronWorkflow manifest

---

## Story 4 — Horizon / queued-jobs reliability + observability

### Current State
- Config: `config/horizon.php` (~35 dedicated supervisors + a default fallback), `config/queue.php` (Redis, `retry_after: 1500s`).
- All supervisors use `balance: auto`. Long-wait detection exists only on `redis:default` (60s).
- **Most of the ~44 job classes have no per-job `$tries` / `$timeout` / `$backoff` / `$maxExceptions`** — they inherit supervisor defaults, which are usually `tries: 1`. A single transient failure = permanent loss. Only the approval-cascade jobs set explicit retries/backoff.
- **Failed-job alerting is disabled** — `routeMailNotificationsTo()` / `routeSlackNotificationsTo()` are commented out in `HorizonServiceProvider.php`. `failed_jobs` accumulates with no retry/alert mechanism.
- **`horizon:snapshot` metrics scheduling is commented out** in `Console/Kernel.php` — no queue-depth history / trend visibility.
- **Orphaned supervisors/queues** still configured (e.g. `OnboardingPlans2018Job`, `OnboardingPlans2019Job`, `validate_accounts`) — never dispatched; waste worker slots and clutter monitoring.
- **`GenerateReportJob` uses `reports-*` queues with no matching supervisor** — only the default fallback (or unprocessed).
- No queue prioritization — long posting jobs can starve fast user-facing queues (emails, notifications).

### Candidate improvements (scope TBD — see review gate)
1. Enable **failed-job + long-wait alerting** (Slack/email) via Horizon notifications.
2. Schedule **`horizon:snapshot`** for metrics/trend history.
3. Add **per-job reliability config** (`$tries`, `$timeout`, `$backoff`, `$maxExceptions`) to high-traffic jobs (posting, finalizer, email, notifications, retry).
4. **Exponential backoff** on jobs calling external APIs (social platforms, inbox, analytics).
5. **Queue isolation/prioritization** — dedicate/raise capacity for critical user-facing queues so heavy posting jobs can't starve them.
6. **Remove orphaned supervisors/queues**.
7. **Add a supervisor for `reports-*` queues** (or consolidate).
8. **Failed-job retry command** (`jobs:retry-failed`).

### Files Involved
- `contentstudio-backend/config/horizon.php`
- `contentstudio-backend/config/queue.php`
- `contentstudio-backend/app/Console/Kernel.php`
- `contentstudio-backend/app/Providers/HorizonServiceProvider.php`
- `contentstudio-backend/app/Jobs/*`
