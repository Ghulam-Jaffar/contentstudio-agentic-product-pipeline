# Research — Custom day picker for scheduled analytics reports

**Date:** 2026-05-18
**Pipeline:** `/story` (local docs only — not pushed to Shortcut)
**Type:** Feature (BE + FE)
**Surfaces:** Analytics → "Schedule PDF as Email" modal

**Epic (notional, not in Shortcut):** Custom day picker for scheduled analytics reports

---

## What we're solving

Today the Analytics "Schedule PDF as Email" modal only lets a user pick **Weekly** (fixed to Monday) or **Monthly** (fixed to the 2nd). The user cannot change which day of the week or which day of the month the report goes out. We want to:

1. Let the user pick **which day** when they choose Weekly (Mon → Sun pills).
2. Let the user pick **which day-of-month** when they choose Monthly (1–28 numbered grid; capped at 28 to skip Feb edge cases).
3. Persist that choice **per-scheduled-report** (each scheduled report has its own day).
4. Also persist a **per-user default**, with an opt-in "save as my default" toggle on the modal, so future "Schedule" actions prefill the user's preferred day.

UI sketch (agreed):

```
○ Weekly
  Send every  [M][T][W][T][F][S][S]    ← single-select, default Monday
  Report of the previous week.

● Monthly
  Send on day  [1][2][3][4][5][6][7]
               [8][9][10][11][12][13][14]
               [15][16][17][18][19][20][21]
               [22][23][24][25][26][27][28]
  Report of the previous month.

[ ] Save as my default for future reports
```

Pills/grid render only under the selected cadence. The unselected cadence stays compact (title + subtitle only).

---

## Current State

### Frontend

**Modal:** [analytics/components/reports/modals/ScheduleReportModal.vue](contentstudio-frontend/src/modules/analytics/components/reports/modals/ScheduleReportModal.vue) (1280 lines). Lives in the **legacy `analytics` module**, but `analytics_v3` reuses it via shared imports and events — there is no `analytics_v3`-native replacement. Touching this file affects both legacy and v3 surfaces.

**Open via event:** `EventBus.$emit('schedule-report', { ... })` triggered from places like `analytics/components/reports/DownloadPdfButton.vue` and `analytics/views/common/ExportButton.vue`.

**Schedule state today:** [stores/analytics/useAnalyticsStore.ts](contentstudio-frontend/src/stores/analytics/useAnalyticsStore.ts) at L40–L74:

```ts
interface ScheduledReportItem {
  ...
  interval: string           // 'weekly' | 'monthly' — that's the entire knob
  email_list: string[]
  email: string | null
  copy_email_to_myself: boolean
  ...
}
// default: interval: 'weekly'
```

**Submit:** `analyticsStore.scheduleReportsService(payload)` POSTs `{ ..., interval, language }`. Currently sends nothing about which day.

**Disclaimer copy** (i18n key `analytics.common.schedule_report_modal.intervals.*.description`, all 8 locales):
- weekly → "Every Monday you'll be sent a report of the previous week."
- monthly → "On 2nd of every month you'll be sent a report of the previous month."

**Locale files affected** (all 8 — en/de/el/es/fr/it/pl/zh):
- `contentstudio-frontend/src/locales/<locale>/analytics.json` — the `intervals.*.description` keys become dynamic; new keys for day pills (Mon–Sun short labels), day numbers (or just numeric), "Save as my default" toggle copy, validation errors.

### Backend

**Surprise win:** the BE already supports `day_of_week` and `day_of_month` — only the FE is missing the UI to set them.

**Model:** [Models/Analytics/ScheduleReportsModel.php](contentstudio-backend/app/Models/Analytics/ScheduleReportsModel.php) — `fillable` at L13–L30 already includes:

```php
'frequency',        // daily, weekly, monthly, custom
'day_of_week',      // 0-6 for weekly (0=Sunday)
'day_of_month',     // 1-31 for monthly
```

**Next-run computation:** [Services/Analytics/ReportScheduleCalculator.php](contentstudio-backend/app/Services/Analytics/ReportScheduleCalculator.php) — already branches on `frequency`, reads `$schedule->day_of_week ?? 1` (default Monday) and `$schedule->day_of_month ?? 1` (default 1st), handles month-day overflow via `daysInMonth`.

**Scheduler cron:** `app/Console/Commands/Analytics/RunDueReportsCommand.php` runs every minute (per `app/Console/Kernel.php` L17 "Process due scheduled reports every minute") and picks up reports whose `next_run` ≤ now.

**Controller:** [Http/Controllers/Analytics/Analytics/ScheduleReports.php](contentstudio-backend/app/Http/Controllers/Analytics/Analytics/ScheduleReports.php) — `send()` reads `$request->get('interval')` and routes through `ReportsHelper::scheduleReports($interval)`. **The controller is currently throwing away any day-of-week / day-of-month value the FE sends** because the FE doesn't send them and the controller doesn't pass them through. The write path (probably `ScheduleReportsRepo::add/update`) needs to accept the new fields too.

**Default mismatch flagged:** FE copy says "On 2nd of every month"; BE calculator defaults to `day_of_month ?? 1` (the 1st). Either:
- The current default actually fires on the 1st and the copy is wrong, or
- Existing records were seeded with `day_of_month = 2` somewhere else.
This needs to be verified during BE story implementation. Either way, the new behavior is "user picks the day, default Monday / day 1 (or 2 — confirm with QA against production behavior)."

---

## What Needs to Change

### Backend
- Accept `day_of_week` (0–6) and `day_of_month` (1–28) on the **scheduled-report write endpoint** (`ScheduleReports::send` / `ScheduleReportsRepo::add|update`). Validate ranges. Map `interval` → `frequency` as today, but persist the new day fields alongside.
- Clamp `day_of_month` to **1–28** (no 29/30/31) to skip Feb edge cases. Per agreed UI scope.
- Add user-profile preference fields: `preferred_schedule_day_of_week` (0–6, default 1) and `preferred_schedule_day_of_month` (1–28, default 2 to match the historical UI copy). Stored on the user profile.
- Expose those prefs on `GET /profile` (or wherever the FE already reads the user profile) and accept them on the profile-update endpoint.
- Trigger `ReportScheduleCalculator::calculate()` to recompute `next_run` whenever a report's `day_of_week` or `day_of_month` is changed (and ensure existing recompute hook still fires on `interval`/`frequency` change).

### Frontend
- Add the **day-pills** picker (M T W T F S S) under the Weekly radio when Weekly is selected. Default to the user's `preferred_schedule_day_of_week`, falling back to Monday.
- Add the **1–28 numbered grid** picker under the Monthly radio when Monthly is selected. Default to the user's `preferred_schedule_day_of_month`, falling back to 2 (current historical default).
- Add a **"Save as my default for future reports"** checkbox below the schedule block. When checked + Schedule clicked, the FE writes the chosen day back to the user profile.
- Update `ScheduledReportItem` type to include `day_of_week?: number` and `day_of_month?: number`, plus a transient `save_as_default: boolean` flag for the toggle.
- Send `day_of_week` / `day_of_month` in the schedule payload to the BE.
- Update i18n disclaimer text in **all 8 locales** to be dynamic: `"Every {day} you'll be sent a report of the previous week."` / `"On {day} of every month you'll be sent a report of the previous month."` — with `{day}` filled in from the user's selection. Add the new pill labels (M/T/W/T/F/S/S — short forms, week starts on Monday for ContentStudio's region defaults) and "Save as my default" copy.
- Open-modal flow: when `EventBus.$emit('schedule-report', ...)` or `'edit-schedule-report'` fires, prefill day-of-week/day-of-month from the existing report (edit case) or from the user profile (new case).
- Validation: if Weekly is selected without a pill, show inline error; if Monthly without a day, show inline error. (In practice we always have a default, so this is defensive.)

### Out of scope (deferred)
- Daily / custom frequencies — the BE supports them, the FE doesn't expose them today, and we're not adding them in this epic.
- Time-of-day picker — defer; the BE stores a time, the FE currently has no time picker, and adding one is a separate scope.
- A general scheduling-rules refactor of the legacy `analytics` module to `analytics_v3` — out of scope.
- Hour/timezone selection — out of scope; the existing workspace timezone behavior is reused.

---

## UX Reference

Day-pills for weekly recurrence is a standard pattern (Apple Calendar, Google Calendar, Slack reminders, every cron-style UI). For monthly day-of-month, both a numbered grid (Apple Reminders) and a dropdown ("On the [2nd] of every month") are common — we're going with the grid for visual parity with the weekly pills.

---

## Mobile Context

Not in scope. Scheduling analytics reports is a web-only flow today — neither the iOS nor Android apps surface this modal. If we later expose it on mobile, a follow-up story is needed.

---

## Files Involved

### Backend
- `contentstudio-backend/app/Http/Controllers/Analytics/Analytics/ScheduleReports.php` — accept new fields on send
- `contentstudio-backend/app/Services/Analytics/ReportsHelper.php` — pass new fields through; check method exists for `scheduleReports($interval)` (controller calls it but grep didn't find the implementation — verify)
- `contentstudio-backend/app/Models/Analytics/ScheduleReportsModel.php` — `fillable` already covers it; verify validation rules
- `contentstudio-backend/app/Services/Analytics/ReportScheduleCalculator.php` — already supports day_of_week / day_of_month; just verify recompute hooks fire on update
- `contentstudio-backend/app/Repository/Analytics/ReportsRepo.php` / a `ScheduleReportsRepo.php` (TBD by engineering) — write-path needs to persist the new fields
- User profile model + profile controller — add `preferred_schedule_day_of_week`, `preferred_schedule_day_of_month` and expose on the profile API

### Frontend
- `contentstudio-frontend/src/modules/analytics/components/reports/modals/ScheduleReportModal.vue` — the UI
- `contentstudio-frontend/src/stores/analytics/useAnalyticsStore.ts` — `ScheduledReportItem` type + defaults + payload composition
- `contentstudio-frontend/src/locales/<locale>/analytics.json` — all 8 locales: pill labels, dynamic disclaimer, toggle copy, validation strings
- `contentstudio-frontend/src/locales/<locale>/settings.json` (or wherever profile prefs are read/written) — only if new profile-pref strings are needed
- `contentstudio-frontend/src/stores/core/useProfileStore.ts` — read/write the new profile prefs

### Out of scope (flagged)
- The legacy `analytics` module — this modal lives there even though the active analytics surface is `analytics_v3`. Per `frontend/CLAUDE.md` we don't add features to legacy modules. Either (a) modernize this modal into `analytics_v3` as part of the FE story, or (b) accept that we're touching legacy because there's no v3 replacement yet. Recommend (b) and flag a future migration story.
