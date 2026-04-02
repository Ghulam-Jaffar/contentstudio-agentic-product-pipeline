# Research: Analytics Date Picker — "Last Month" + Sticky Date Range

## Current State

The analytics date picker is configured in `contentstudio-frontend/src/modules/analytics/components/common/helper.js` via the `analyticsDatePickerValues()` function. The `ranges` object currently has:

```
All Time, 24 hours, 48 hours, Last 3 days, Last 7 Days, Last 30 Days,
Last 3 Months, Last 6 Months, Last 1 Year, Last 2 Years
```

"Last Month" (the previous calendar month, e.g. March 1–31) is missing. "Last 30 Days" is a rolling window, not the same thing.

Each platform composable (`useFacebookAnalytics.js`, `useInstagramAnalytics.js`, `useLinkedinAnalytics.js`, etc.) initialises its own `dateRange` ref, defaulting to the last 30 days. When a user switches between platform tabs, the date range resets to that default — it is not remembered.

## What Needs to Change

- **Add "Last Month"** to the `ranges` object in `analyticsDatePickerValues()` — this is the previous calendar month (start of last month → end of last month), not a rolling 30-day window.
- **Persist the selected date range** on the user object under `preferences.analytics_date_range` using the existing `setUserPreferences` API (`POST preferences/setPreferences`, key/value). This API already exists and is used for other preferences (calendar, inbox, composer).
- **Restore on load** — when any analytics platform page is opened, read `profile.preferences.analytics_date_range` from `useProfileStore` and initialise `dateRange` with it instead of the hardcoded default.
- The profile store (`src/stores/core/useProfileStore.ts`) already exposes the `preferences` object; a new computed getter is needed for `analytics_date_range`.

## Files Involved

- `contentstudio-frontend/src/modules/analytics/components/common/helper.js` — add "Last Month" range
- `contentstudio-frontend/src/stores/core/useProfileStore.ts` — add getter + setter for `analytics_date_range` preference
- `contentstudio-frontend/src/config/api-utils.js` — `setUserPreferences` URL already defined
- All platform analytics composables (and/or a new shared composable to centralise the persistence logic):
  - `src/modules/analytics/views/facebook_v2/composables/useFacebookAnalytics.js`
  - `src/modules/analytics/views/instagram_v2/composables/useInstagramAnalytics.js`
  - `src/modules/analytics/views/linkedin_v2/composables/useLinkedinAnalytics.js`
  - `src/modules/analytics/views/twitter/composables/useTwitterAnalytics.js`
  - `src/modules/analytics/views/tiktok/composables/useTiktokAnalytics.js`
  - `src/modules/analytics/views/youtube/composables/useYoutubeAnalytics.js`
  - `src/modules/analytics/views/pinterest/composables/usePinterestAnalytics.js`
  - `src/modules/analytics/views/overviewV2/composables/useOverviewAnalytics.js`
