# Research: Analytics Sync — Date Range Selection Modal

## Current State

The "Sync Data" button in analytics always triggers a full "all time" sync — there's no way to select a specific date range.

**Two sync entry points:**
1. **Regular analytics** — `TabsComponent.vue` (`src/modules/analytics/views/common/TabsComponent.vue`) renders a "Sync Data" button for non-Twitter platforms. On click it calls `handleSyncData()` → `triggerManualSync({ account, platform })`.
2. **Competitor analytics** — `MainAnalyticsHeader.vue` (`src/modules/analytics_v3/components/MainAnalyticsHeader.vue`) has its own "Sync Data" button that calls `handleSyncData()` → `triggerCompetitorSync(...)` (different endpoint, out of scope for this story).

**Frontend sync chain (regular analytics):**
- `TabsComponent.vue` → `useManualSync` composable (`src/modules/analytics/components/common/composables/useManualSync.js`)
- Composable POSTs to `api/analytics/triggerJob` (defined in `src/modules/analytics/config/api-utils.js:175`)
- Payload: `{ workspace_id, account_id, platform }` — no date parameters

**Backend sync chain:**
- Route: `POST analytics/triggerJob` → `AnalyticsJobController::triggerAnalyticsJob()` (`app/Http/Controllers/Analytics/Analyze/AnalyticsJobController.php`)
- Validation: `TriggerAnalyticsJobRequest` — requires `workspace_id`, `account_id`, `platform` only
- Calls `AnalyticsHelper::triggerAnalyticsPipeline($channel, $account)` (`app/Libraries/Analytics/AnalyticsHelper.php`)
- For Facebook/Instagram/LinkedIn/Twitter/TikTok: calls `triggerImmediateWorker()` → POSTs to Go analytics pipeline (`ANALYTICS_PIPELINE_API_URL/immediate-work`) with `{ channel, account_id }`

**Go analytics pipeline (immediate processor):**
- `ImmediateWorkOrder` struct in `src/services/unified/immediate-processor/main.go:142`
- Current fields: `id`, `platform`, `account_id`, `type`, `access_token`, `workspace_id`, `sync_type`, `connected_via_instagram`
- **No `start_date`/`end_date` fields** — syncs all available history

**Analytics date picker (used in filter bar):**
- `CstInputFields` component with `type="date"` and `dateOptions` shortcuts
- Shortcuts available: All Time, Last 24 Hours, Last 48 Hours, Last 3 Days, Last 7 Days, Last 30 Days, Last 3 Months, Last 6 Months, Last 1 Year, Last 2 Years
- Fully defined in `AnalyticsFilterBarWrapper.vue` (`src/modules/analytics/views/common/AnalyticsFilterBarWrapper.vue:73-196`)
- Default range: Last 30 days

---

## What Needs to Change

**Frontend:**
- When user clicks "Sync Data", open a `Modal` instead of directly triggering sync
- Modal contains a `CstInputFields` date picker (same options as the analytics header) plus a confirm CTA
- On confirm, call `triggerManualSync` with `start_date` and `end_date` added to the payload
- Update `useManualSync` composable to accept and forward date range parameters
- Add new `SyncDateRangeModal.vue` component

**Backend (PHP):**
- Add optional `start_date` / `end_date` to `TriggerAnalyticsJobRequest` validation
- Forward the date range in `AnalyticsJobController` → `AnalyticsHelper::triggerImmediateWorker()`
- Include `start_date` / `end_date` in the payload POSTed to the Go analytics pipeline

**Go analytics pipeline:**
- Add `StartDate`/`EndDate` fields to `ImmediateWorkOrder` struct
- Pass date range from the work order to the platform fetchers (Facebook, Instagram, LinkedIn, Twitter, TikTok) so they fetch data only for the selected range

---

## UX Reference

Standard pattern for confirming a sync with date selection — similar to how analytics platforms (Hootsuite, Sprout Social) let you select a historical backfill range before triggering a re-sync.

---

## Files Involved

**Frontend:**
- `src/modules/analytics/views/common/TabsComponent.vue` — add modal trigger
- `src/modules/analytics/components/common/composables/useManualSync.js` — add date params
- `src/modules/analytics/config/api-utils.js` — no change needed (URL stays same)
- New: `src/modules/analytics/components/common/SyncDateRangeModal.vue`
- `src/locales/*/analytics.json` — new i18n keys for modal copy

**Backend (PHP):**
- `app/Http/Requests/Analytics/TriggerAnalyticsJobRequest.php` — add `start_date`, `end_date` rules
- `app/Http/Controllers/Analytics/Analyze/AnalyticsJobController.php` — forward dates
- `app/Libraries/Analytics/AnalyticsHelper.php` — include dates in Go pipeline payload

**Go analytics pipeline:**
- `src/services/unified/immediate-processor/main.go` — update `ImmediateWorkOrder` struct + HTTP handler
- Platform fetcher services — use date range when fetching
