# Stories: Analytics Sync — Date Range Selection Modal

Epic: [Analytics Sync — Date Range Selection](https://app.shortcut.com/contentstudio-team/epic/115221)

---

## Story 1: [FE] Add date range selection modal to the analytics Sync Data flow

### Description:

As an analytics user, I want to choose a specific date range before syncing analytics data, so that I can refresh only the period I care about instead of always triggering a full all-time sync.

---

### Workflow:

1. User is on any platform analytics page (Facebook, Instagram, LinkedIn, TikTok, etc.) and sees the **"Sync Data"** button in the analytics tabs header.
2. User clicks **"Sync Data"**.
3. A modal opens titled **"Refresh Analytics Data"** with a brief description and a date picker.
4. The date picker defaults to **Last 30 Days** and shows the same shortcut options available in the analytics filter bar:
   - All Time
   - Last 24 Hours
   - Last 48 Hours
   - Last 3 Days
   - Last 7 Days
   - Last 30 Days *(default)*
   - Last 3 Months
   - Last 6 Months
   - Last 1 Year
   - Last 2 Years
5. User selects a date range (either via a shortcut or custom dates using the calendar).
6. User clicks **"Refresh Data"** (primary CTA).
7. The modal closes, the button enters its loading/spinning state, and the sync is triggered for the selected date range.
8. On success: a success toast appears — **"Sync started. Your analytics data will update shortly."**
9. On error: an error toast appears — **"Couldn't start sync. Please try again."**
10. User can click **"Cancel"** at any time to dismiss the modal without triggering a sync.

---

### UI Copy

**Modal:**
- **Title:** `Refresh Analytics Data`
- **Description:** `Select the date range you'd like to re-fetch. Only data within this period will be updated — this won't affect data outside the selected range.`
- **Date picker label:** `Date Range`
- **Date picker placeholder:** `Select date range`
- **Primary CTA:** `Refresh Data`
- **Secondary CTA:** `Cancel`

**Sync button (already exists — no copy change needed):**
- Idle: `Sync Data`
- Loading: `Syncing...`

**Toast — success:** `Sync started. Your analytics data will update shortly.`
**Toast — error:** `Couldn't start sync. Please try again.`

**Tooltip on the Sync Data button** (add/update):
- Content: `Re-fetch analytics data for a specific date range. Choose how far back you'd like to refresh — useful if you notice missing or outdated data.`

---

### Component References

- Use the `Modal` component from `@contentstudio/ui` for the modal dialog.
- Use `CstInputFields` (type=`"date"`) with `dateOptions` for the date picker — same configuration as `AnalyticsFilterBarWrapper.vue` (`src/modules/analytics/views/common/AnalyticsFilterBarWrapper.vue:73-196`). Copy the shortcuts and `disabledDate` logic (disable future dates).
- Use `Button` from `@contentstudio/ui` for primary and secondary CTAs.
- Create a new component: `src/modules/analytics/components/common/SyncDateRangeModal.vue`

---

### Files to Create/Modify

- **New:** `src/modules/analytics/components/common/SyncDateRangeModal.vue` — the modal with date picker
- **Modify:** `src/modules/analytics/components/common/composables/useManualSync.js` — accept `startDate`/`endDate` and include in API payload
- **Modify:** `src/modules/analytics/views/common/TabsComponent.vue` — open `SyncDateRangeModal` on button click instead of directly calling sync
- **Modify:** `src/locales/*/analytics.json` (all locales) — add i18n keys for all modal copy above

---

### Acceptance Criteria:

- [ ] Clicking "Sync Data" opens the "Refresh Analytics Data" modal instead of immediately triggering a sync
- [ ] The modal date picker defaults to Last 30 Days on open
- [ ] All 10 date shortcuts are present and functional (All Time through Last 2 Years)
- [ ] Custom date selection via calendar is supported
- [ ] Future dates are disabled in the date picker
- [ ] Clicking "Refresh Data" closes the modal and triggers sync with the selected `start_date` and `end_date` in the API payload
- [ ] Clicking "Cancel" closes the modal with no sync triggered
- [ ] The Sync Data button shows a spinner and "Syncing..." text while the request is in flight
- [ ] Success toast appears on a successful API response
- [ ] Error toast appears when the API returns an error
- [ ] Modal is dismissible via the X button and by clicking outside
- [ ] All copy (title, description, CTAs, toasts, tooltip) matches the spec above
- [ ] All strings are translated via `$t()` with keys added to all locale files under `src/locales/`
- [ ] Modal renders correctly on all screen sizes (mobile responsive)
- [ ] Works on all platforms that show the Sync Data button: Facebook, Instagram, LinkedIn, TikTok, Twitter

---

### Mock-ups:

N/A — use the existing date picker pattern from `AnalyticsFilterBarWrapper.vue` as the visual reference.

---

### Impact on existing data:

No changes to stored data. Only changes the parameters passed to the existing sync API endpoint.

---

### Impact on other products:

- The sync modal is web-only. Mobile apps do not have an analytics sync button.
- Chrome extension is not affected.
- White-label workspaces are fully supported — all components use theme-aware `@contentstudio/ui` components and CSS variable-based colors (`text-primary-cs-500`, `bg-primary-cs-50`, etc.).

---

### Dependencies:

Depends on: **[BE] Accept date range parameters in analytics sync API and forward to pipeline**

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

## Story 2: [BE] Accept date range parameters in analytics sync API and forward to pipeline

### Description:

As a backend system, I need to accept optional `start_date` and `end_date` parameters on the analytics sync endpoint, so that the Go analytics pipeline can fetch data only for the user's selected date range instead of always running a full all-time sync.

---

### Workflow:

1. Frontend sends `POST api/analytics/triggerJob` with `{ workspace_id, account_id, platform, start_date, end_date }` — `start_date` and `end_date` are optional ISO date strings (e.g., `"2026-03-01"`, `"2026-03-31"`).
2. `TriggerAnalyticsJobRequest` validates the new optional fields.
3. `AnalyticsJobController::triggerAnalyticsJob()` extracts the dates and passes them to `AnalyticsHelper::triggerAnalyticsPipeline()`.
4. `AnalyticsHelper::triggerImmediateWorker()` includes `start_date` and `end_date` in the POST payload to the Go analytics pipeline (`ANALYTICS_PIPELINE_API_URL/immediate-work`).
5. The Go analytics pipeline (`ImmediateWorkOrder` struct in `src/services/unified/immediate-processor/main.go`) is updated to accept `StartDate`/`EndDate` fields and passes them to the platform-specific fetchers (Facebook, Instagram, LinkedIn, TikTok, Twitter).
6. Platform fetchers use the date range to limit the data they pull from external APIs.
7. If `start_date`/`end_date` are omitted, the pipeline falls back to its current all-time behavior — no breaking change.

---

### Acceptance Criteria:

- [ ] `POST api/analytics/triggerJob` accepts optional `start_date` and `end_date` fields (ISO date string format, e.g., `"2026-03-01"`)
- [ ] If `start_date` or `end_date` is provided but invalid (not a valid date), the API returns a `422` validation error
- [ ] When dates are provided, they are included in the payload sent to the Go analytics pipeline
- [ ] When dates are omitted, the sync behaves identically to today (full all-time sync) — no regression
- [ ] The Go `ImmediateWorkOrder` struct includes `StartDate` and `EndDate` fields (empty string or omitted = all time)
- [ ] The immediate-processor HTTP handler reads `start_date`/`end_date` from the request body and populates the work order
- [ ] Platform fetchers (Facebook, Instagram, LinkedIn, TikTok, Twitter) respect the date range when set — they fetch data from `start_date` to `end_date` only
- [ ] The 1-hour cache key (which prevents re-triggering for the same account within an hour) remains in place
- [ ] No changes to the competitor sync endpoint (`triggerCompetitorJob`) — out of scope

---

### Mock-ups:

N/A — backend-only story.

---

### Impact on existing data:

No schema changes. The date range parameters are passed through to the Go pipeline at request time only — no new fields stored in MongoDB or ClickHouse.

---

### Impact on other products:

- No impact on frontend beyond the new optional params it now sends.
- No impact on mobile apps or Chrome extension.
- The Python Argo job queue path (used for platforms not on the Go pipeline) is not in scope — if `start_date`/`end_date` are passed for those platforms, they can be accepted but ignored initially with a follow-up to add date range support.

---

### Dependencies:

None — this story can be implemented independently. The frontend story can be developed in parallel using the existing endpoint; the date params are additive.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
