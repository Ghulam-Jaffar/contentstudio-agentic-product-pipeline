# Schedule Time Manual Input — Research
*Story Pipeline · Step 1 · Generated 2026-04-13*

---

## Current State

ContentStudio has **two separate time picker components** used across the product. Both render scrollable dropdown lists for hour and minute selection — requiring users to scroll through up to 60 items to pick a minute value.

### Component 1: `TimePickerDropdown.vue`
**File:** `contentstudio-frontend/src/modules/integration/components/platforms/social_v2/components/TimePickerDropdown.vue`

A popover-based time picker used in **Settings → Social Accounts** and **Automation**. It renders two side-by-side scrollable columns:
- Hour column: 24 rows (0–23 in 24h mode) or 12 rows (1–12 in 12h mode) using `DropdownItem`
- Minute column: 60 rows (00–59) using `DropdownItem`
- AM/PM column in 12h mode
- Cancel and Submit buttons at the bottom
- Scroll position auto-snaps to selected value on open via `scrollToSelectedTime()`

**Used in:**
- `SocialAccountDetailsModal.vue` — Settings → Social Accounts → Account Details → Queue time slots (add and edit per-day publishing time slots)
- `ScheduleOptions.vue` — Automation → Evergreen Automation → Custom Schedule → per-day time slots

### Component 2: `SelectTime.vue`
**File:** `contentstudio-frontend/src/modules/composer_v2/components/PostingSchedule/SelectTime.vue`

A compact inline time picker used in the Composer's date picker footer. Renders two legacy `CstDropdown` components (hours + minutes) and optionally an AM/PM dropdown.

**Used in:**
- `SingleAccountSchedulerComponent.vue` — Composer → Post Scheduler → per-account date/time picker
- `SocialMediaPostingScheduler.vue` — Composer → Post Scheduler → same-time-for-all-accounts view
- `PostingSchedule.vue` — Composer → Posting Schedule section

---

## What Needs to Change

Both pickers should support **two input methods simultaneously** — scrolling/clicking (existing) AND direct typing:

- **`TimePickerDropdown.vue`:** Add a manual-entry header above the scrollable columns. Render two small `TextInput` fields (hour + minute) showing the currently selected time. User can either:
  - Type directly into the hour/minute inputs to set a value quickly, OR
  - Continue scrolling and clicking the `DropdownItem` list rows as today
  - Typing in an input updates the highlighted/selected row in the list below in real time
  - Clicking a list row updates the input field value
  - Inputs validate: hour 0–23 (24h) or 1–12 (12h); minute 0–59
  - Pad to 2 digits on blur (`9` → `09`)
  - Invalid value: input border turns red, Submit button disabled
  - Keep AM/PM column for 12h format
  - Keep Cancel + Submit button layout unchanged

- **`SelectTime.vue`:** Add editable input fields to the existing `CstDropdown` trigger areas (the "selected" slot currently shows the value as plain text). The displayed hour/minute value becomes a `TextInput` the user can click and type into directly, while the dropdown list still appears on open and remains scrollable/clickable. Same validation rules.

---

## UX Reference

Tools like Google Calendar, Buffer, and Metricool use plain `<input type="text">` or `<input type="number">` for time fields — click the field, type the value, done. No scrolling. This is the standard "spinbox" pattern for time entry in schedulers.

---

## Files Involved

| File | Change |
|---|---|
| `contentstudio-frontend/src/modules/integration/components/platforms/social_v2/components/TimePickerDropdown.vue` | Replace scrollable DropdownItem lists with TextInput fields |
| `contentstudio-frontend/src/modules/composer_v2/components/PostingSchedule/SelectTime.vue` | Replace CstDropdown hour/minute with TextInput fields |
| `contentstudio-frontend/src/locales/*/composer.json` (all locale dirs) | Add i18n keys for any new validation messages |
| `contentstudio-frontend/src/locales/*/settings.json` (all locale dirs) | Add i18n keys if any copy changes in the queue slot picker |
