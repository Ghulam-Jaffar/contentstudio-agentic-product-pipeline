# Epic B — YouTube demographics analytics

## Problem

YouTube analytics in ContentStudio does not surface audience demographics. YouTube Studio exposes rich audience data (age & gender, top geographies, device type, subtitle/CC languages, subscribers gained/lost, monthly/unique audience) that we do not fetch or show. Users want a Demographics tab for YouTube like other platforms.

## Goal

Add a Demographics tab for YouTube that surfaces whatever the YouTube Analytics API provides, and thread it through the whole analytics stack: data pipeline, in-app analytics, public analytics API, PDF reports, and the Looker (Data) Studio connector. Where YouTube returns no data for a metric, show the platform's usual "not enough data" empty state.

## Metrics in scope (map to what the YouTube Analytics API returns)

- Age & gender (viewer percentage by age group and gender)
- Top geographies (by country)
- Device type (watch time by device)
- Subtitle / CC languages
- Subscribers gained/lost and net change over the range
- Monthly / unique audience where available

Metrics the API does not provide, or provides sparsely (for example "channels your audience watches", "what your audience watches"), are out of scope or shown only if the API exposes them.

## Stories (see `02-stories.md`)

1. `[BE]` Fetch and store YouTube demographics in the analytics pipeline (Go)
2. `[FE]` Add a Demographics tab to YouTube analytics
3. `[BE]` Expose YouTube demographics via the public Analytics API (v1)
4. `[FE]`/`[BE]` Add YouTube demographics to Analytics Reports (PDF)
5. `[Data Studio]` Add YouTube demographics to the Looker Studio connector

## Sequencing

Story 1 (pipeline) is the foundation. 2, 3, 4, 5 consume it and can run in parallel once the data shape is set. Freeze the data shape in story 1 so the four consumers agree on field names.

## Out of scope

- Non-YouTube platforms (Facebook/Instagram already have demographics; not changing them here).
- Metrics YouTube does not expose through its Analytics API.



---

---


# Epic B stories — YouTube demographics analytics

## [BE] Fetch and store YouTube demographics in the analytics pipeline

### Description
Extend the YouTube analytics fetcher to pull demographics from the YouTube Analytics API (age & gender, top geographies, device type, subtitle/CC languages, subscribers gained/lost, monthly/unique audience) and store them so the rest of the stack can serve them. Define a stable data shape the API, reports, and Looker connector all consume.

### Acceptance criteria
- [ ] The YouTube fetcher requests the demographic reports the YouTube Analytics API supports (age group + gender, country, device type, subtitle language, subscriber gained/lost, unique viewers where available).
- [ ] Fetched demographics are stored with the account and date range in the analytics store.
- [ ] A documented, stable data shape/field names are defined for downstream consumers.
- [ ] When YouTube returns no data for a metric, that is represented as empty (not an error) so the UI can show a "not enough data" state.
- [ ] Fetch failures and quota/permission errors are handled and do not break the rest of the YouTube analytics sync.

### Impact on existing data
Adds demographic datasets for YouTube accounts. Confirm storage/retention matches other platforms.

### Impact on other products
Feeds in-app analytics, public API, reports, and Looker. No direct user surface in this story.

### Global quality and compliance checklist
- [ ] Mobile responsiveness (N/A, backend-only story)
- [ ] Multilingual support (N/A, data pipeline)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research, not a contract.*
- YouTube fetcher: `contentstudio-social-analytics-go/src/services/youtube/youtube-fetcher/main.go` and `src/cmd/jobs/fetcher/youtube.go`. Extend to request the YouTube Analytics API demographic dimensions.
- Follow how Facebook/Instagram demographics are fetched and stored for parity.

---

## [FE] Add a Demographics tab to YouTube analytics

### Description
Add a Demographics tab to YouTube analytics, matching the pattern used by other platforms, showing age & gender, top geographies, device type, subtitle/CC languages, and the subscriber change, with "not enough data" empty states where YouTube returns none.

### Workflow
1. User opens YouTube analytics and selects the Demographics tab.
2. The user sees age & gender, top geographies, device type, subtitle/CC languages, and subscriber change for the selected date range.
3. Where YouTube has no data for a section, the user sees a clear "not enough data" message instead of an empty chart.

### Acceptance criteria
- [ ] YouTube analytics has a Demographics tab consistent with the other platforms' analytics tabs.
- [ ] The tab shows: age & gender, top geographies, device type, subtitle/CC languages, and subscriber gained/lost/net.
- [ ] Each section respects the selected date range and account.
- [ ] Sections with no data show the standard "not enough data" empty state, for example: "Not enough data to show this report."
- [ ] Loading and error states are handled.
- [ ] Charts use the shared analytics chart components and theme-aware colors.

### UI copy
- Tab label: "Demographics"
- Empty state (per section): "Not enough data to show this report."
- Section titles: "Age and gender", "Top geographies", "Device type", "Subtitle/CC languages", "Subscribers".

### Impact on existing data
None (presentation).

### Impact on other products
Web analytics. Mobile/Chrome: N/A.

### Global quality and compliance checklist
- [ ] Mobile responsiveness tested (frontend only)
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default and white-label) — no dark mode
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research, not a contract.*
- Analytics module: `contentstudio-frontend/src/modules/analytics_v3`. Follow the existing per-platform tab + chart structure and reuse the shared chart/empty-state components. Confirm the current YouTube analytics view location and add the tab there.

---

## [BE] Expose YouTube demographics via the public Analytics API (v1)

### Description
Add public API v1 endpoints for YouTube demographics, consistent with the existing public analytics endpoints for other platforms.

### Acceptance criteria
- [ ] Public API v1 exposes YouTube demographics (age & gender, geographies, device type, subtitle languages, subscribers) under the YouTube analytics path.
- [ ] Endpoints require the API key and workspace scoping like the other v1 analytics endpoints.
- [ ] Response shapes are consistent with the existing analytics API and the pipeline data shape.
- [ ] Empty/no-data cases return cleanly (empty datasets, not errors).
- [ ] OpenAPI/guide updated.

### Global quality and compliance checklist
- [ ] Multilingual support verified (messages localized or fallback handled)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research, not a contract.*
- v1 YouTube analytics controller: `contentstudio-backend/app/Http/Controllers/Api/V1/Analytics/YouTubeAnalyticsController.php` (proxies to the Go analytics pipeline). Add demographics methods mirroring the Instagram/Facebook demographics endpoints; wire the routes in `routes/api/v1.php`.
- Go analytics API handler for YouTube in `contentstudio-social-analytics-go/src/api/analytics/` (add a youtube handler/section if not present).

---

## [FE]/[BE] Add YouTube demographics to Analytics Reports (PDF)

### Description
Include the YouTube demographics in the analytics PDF reports so scheduled/exported reports show the same demographic breakdowns.

### Acceptance criteria
- [ ] YouTube analytics reports include a demographics section (age & gender, geographies, device type, subtitle languages, subscribers).
- [ ] The report section respects the report's date range and account.
- [ ] Sections with no data render a "not enough data" note rather than a broken chart.
- [ ] The report layout stays consistent with existing report sections.

### Global quality and compliance checklist
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research, not a contract.*
- Reuse the report rendering used for other platforms' demographic sections. Confirm the report generation path (analytics reports are rendered server-side; see the reports/gotenberg flow referenced in the analytics work).

---

## [Data Studio] Add YouTube demographics to the Looker Studio connector

### Description
Add YouTube demographics fields to the Looker (Data) Studio connector so users can build YouTube demographic reports in Looker Studio.

### Acceptance criteria
- [ ] The Looker Studio connector exposes YouTube demographics fields (age & gender, geographies, device type, subtitle languages, subscribers).
- [ ] Fields map to the same pipeline data shape used elsewhere.
- [ ] A YouTube connector entry exists (today the connector has Instagram, LinkedIn, Facebook, but not YouTube).
- [ ] No-data cases are handled gracefully in the connector output.

### Global quality and compliance checklist
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research, not a contract.*
- Looker Studio connectors: `contentstudio-social-analytics-go/src/looker-studio/*.gs` (Apps Script). There is `instagram.gs`, `linkedin.gs`, `facebook.gs` but no `youtube.gs`. Add YouTube (or extend the connector) with the demographics fields, following the existing connectors' schema pattern.
