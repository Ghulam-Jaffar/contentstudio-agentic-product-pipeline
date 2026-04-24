# Research: Add "Hour(s)" Option to Repeat Post Interval Dropdown

## Current State

The "Repeat Post" section in the Composer's Posting Schedule allows users to repeat a post at a configurable interval. The interval type dropdown currently offers: **Day(s), Week(s), Month(s)**. The `Hour(s)` option was intentionally commented out in the frontend but the full implementation (frontend logic + backend API) already exists.

**Key file:** `contentstudio-frontend/src/modules/composer_v2/components/PostingSchedule/PostingSchedule.vue`

- Line 632: `<!-- <option value="Hour">Hour(s)</option> -->` — commented out, was never exposed to users on the web
- Lines 1114–1119: `processPostTimings` computed already handles `repeat_type === 'Hour'` (adds hours via Day.js `.add(repeat_gap, 'h')`)
- Lines 1342–1358: Validation watcher — interval error only fires for `repeat_type === 'Day'` with `repeat_gap < 3`; for Hour type the error is automatically cleared

**i18n** — `src/locales/en/composer.json`, section `posting_schedule.intervals`:
```json
"intervals": {
  "days": "Day(s)",
  "weeks": "Week(s)",
  "months": "Month(s)"
}
```
→ No `hours` key exists yet across any of the 8 locale files (`en`, `de`, `el`, `es`, `fr`, `it`, `pl`, `zh`).

**Backend** — Already fully supports `Hour` repeat type in:
- `contentstudio-backend/app/Http/Controllers/Planner/PostingController.php` (lines 372–388): `addHours($repeat_gap)`
- `contentstudio-backend/app/Http/Controllers/Planner/SocialPostingController.php` (lines 1145–1148): `addHours($repeat_gap)` in a switch/case

No backend changes are required.

**Planner scope** — The Planner (v2) opens the same Composer modal (`SocialModal.vue`) which includes `PostingSchedule.vue`. Fixing the composer fixes the planner too. There is no separate interval dropdown in the planner.

**Automation** — `src/modules/automation/components/ScheduleOptions.vue` already has `Hour(s)` for Evergreen automation. That is a separate feature and unrelated to this change.

## What Needs to Change

- Uncomment `<option value="Hour">Hour(s)</option>` in `PostingSchedule.vue` (line 632)
- Reference the new i18n key `$t('composer.posting_schedule.intervals.hours')` on that option
- Add `"hours": "Hour(s)"` to the `intervals` object in all 8 locale `composer.json` files

## UX Reference

Not applicable — the pattern (Hours option in an interval dropdown) is already present in ContentStudio's own Automation/Evergreen feature. No new UX pattern needed.

## Files Involved

| File | Change |
|---|---|
| `contentstudio-frontend/src/modules/composer_v2/components/PostingSchedule/PostingSchedule.vue` | Uncomment `Hour(s)` option; use i18n key |
| `contentstudio-frontend/src/locales/en/composer.json` | Add `intervals.hours` |
| `contentstudio-frontend/src/locales/de/composer.json` | Add `intervals.hours` |
| `contentstudio-frontend/src/locales/el/composer.json` | Add `intervals.hours` |
| `contentstudio-frontend/src/locales/es/composer.json` | Add `intervals.hours` |
| `contentstudio-frontend/src/locales/fr/composer.json` | Add `intervals.hours` |
| `contentstudio-frontend/src/locales/it/composer.json` | Add `intervals.hours` |
| `contentstudio-frontend/src/locales/pl/composer.json` | Add `intervals.hours` |
| `contentstudio-frontend/src/locales/zh/composer.json` | Add `intervals.hours` |
