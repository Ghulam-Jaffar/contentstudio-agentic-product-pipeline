# Stories: Analytics Date Picker — "Last Month" + Sticky Date Range

---

## [FE] Add "Last Month" date range and persist selected date range across analytics platforms

---

### Description:

As a ContentStudio user, I want the analytics date picker to include a "Last Month" option and remember my last-used date range across all platform dashboards, so that I don't have to manually reapply the same filter every time I switch between social channels.

**Two improvements in this story:**

1. **"Last Month" preset** — Add a "Last Month" quick-select option to the analytics date picker. This should represent the previous full calendar month (e.g. on April 5, selecting "Last Month" gives March 1 – March 31), not a rolling 30-day window. This is different from "Last 30 Days" which already exists.

2. **Sticky date range** — Whenever a user selects a date range in analytics (any platform), save it to their user preferences using the existing `setUserPreferences` API (`POST preferences/setPreferences`). When the user returns to any analytics platform dashboard — even after navigating away or logging back in — the last-used range is restored automatically. The range applies globally across all analytics platform tabs (Facebook, Instagram, LinkedIn, TikTok, Twitter, YouTube, Pinterest, Overview).

**Key files:**
- `src/modules/analytics/components/common/helper.js` — add "Last Month" to the `ranges` object in `analyticsDatePickerValues()`
- `src/stores/core/useProfileStore.ts` — add getter and setter for `preferences.analytics_date_range`
- All platform analytics composables — on init, read the saved preference and use it as `dateRange` default; on date range change, persist via `setUserPreferences`

---

### Workflow:

1. User navigates to Analytics and opens any platform dashboard (e.g. Facebook).
2. User clicks the date picker — they see a new **"Last Month"** option in the quick-select list, alongside the existing options.
3. User selects "Last Month" — the dashboard filters to the previous full calendar month (e.g. March 1 – March 31).
4. User switches to another platform tab (e.g. Instagram) — the date picker automatically shows "Last Month" already applied. No need to reselect.
5. User selects a different range (e.g. "Last 7 Days") — the new range is saved.
6. User navigates away from Analytics and comes back — the date picker restores "Last 7 Days" automatically.
7. User logs out and logs back in — the date picker still remembers their last-used range on first load of Analytics.

---

### Acceptance criteria:

- [ ] "Last Month" appears in the analytics date picker quick-select options, positioned between "Last 30 Days" and "Last 3 Months"
- [ ] "Last Month" selects the previous full calendar month (first day to last day of last month), not a rolling 30-day window
- [ ] Selecting any date range (quick-select or custom) saves it to `user.preferences.analytics_date_range` via the `setUserPreferences` API
- [ ] On page load, if `user.preferences.analytics_date_range` is set, it is used as the initial date range instead of the hardcoded default
- [ ] The saved range persists when switching between platform analytics tabs (Facebook → Instagram → LinkedIn, etc.) — the date picker shows the same range across all platforms
- [ ] The saved range persists after navigating away from Analytics and returning
- [ ] The saved range persists across sessions (user logs out and logs back in)
- [ ] If no saved preference exists (new user or first time using Analytics), the existing default (Last 30 Days) is applied as before
- [ ] Custom date ranges (manually selected start/end dates) are also persisted

---

### UI Copy:

**Date picker quick-select label:**
- `"Last Month"` — positioned between "Last 30 Days" and "Last 3 Months" in the list

No new tooltips, modals, or error states are introduced by this change.

---

### Mock-ups:

N/A

---

### Impact on existing data:

A new key `analytics_date_range` is written to `user.preferences` in MongoDB when a user selects a date range in analytics. No existing data is modified. Users who have never visited Analytics after this change is deployed will simply fall back to the existing default (Last 30 Days).

---

### Impact on other products:

Analytics is web-only. No impact on mobile apps or Chrome extension.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, Analytics is web-only
- [ ] Multilingual support — N/A, "Last Month" label uses the existing date picker's i18n mechanism; no new translation keys needed beyond the range label
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
