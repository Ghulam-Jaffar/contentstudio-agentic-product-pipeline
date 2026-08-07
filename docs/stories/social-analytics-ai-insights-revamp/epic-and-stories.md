# Epic C — Social analytics AI insights revamp

## Problem

Social analytics currently attaches an AI insight to each individual widget/card (the per-widget `AiInsightsCard`). This is scattered and inconsistent with the newer, cleaner pattern used for Meta Ads and Google analytics, which have a single unified **Insights** tab that summarizes performance with a hero summary, priorities, and a breakdown.

## Goal

Remove the per-widget AI insights from social analytics and replace them with one unified **Insights** tab per platform, matching the Meta/Google analytics pattern (hero summary, priorities, breakdown/plan). Apply across the social platforms.

## Platforms in scope

Facebook, Instagram, LinkedIn, TikTok, Pinterest, YouTube, and Google Business Profile (all social analytics that currently use per-widget insights).

## Stories (see `02-stories.md`)

1. `[FE]` Remove per-widget AI insights from social analytics
2. `[FE]` Add a unified Insights tab to social analytics (reuse the Meta/Google pattern)
3. `[BE]` Serve platform-level unified insights per social platform (if per-widget insight generation must become a single platform-level insight)

## Sequencing

Confirm the backend can return a single platform-level insight payload (story 3) or that the existing per-platform `ai-insights` endpoints suffice. Build the unified tab (story 2) against that, then remove the per-widget cards (story 1) so users are never left without insights mid-migration.

## Notes

- The unified pattern lives today in the legacy analytics module (`meta_ads` AI insights tab and its `ai_insights/*` components). Social analytics lives in `analytics_v3`. The revamp brings the unified-tab pattern into `analytics_v3` for the social platforms.

## Out of scope

- Meta Ads / Google analytics (already have the unified tab).
- Changing what insights say (this is a placement/structure revamp, not a prompt rewrite), unless a platform-level summary needs a new prompt.



---

---


# Epic C stories — Social analytics AI insights revamp

## [FE] Remove per-widget AI insights from social analytics

### Description
Remove the AI insight attached to each individual social analytics widget/card, so insights are no longer scattered across the dashboard. (The unified Insights tab replaces them.)

### Acceptance criteria
- [ ] The per-widget AI insight card no longer appears on social analytics widgets across all in-scope platforms.
- [ ] Removing it does not break the widget layout or the surrounding metrics.
- [ ] No dead code or empty containers remain where the per-widget insight used to render.
- [ ] This is coordinated with the unified Insights tab so users always have an insights surface (the tab) available.

### Impact on existing data
None (presentation). Per-widget insight requests are removed; confirm no orphaned calls remain.

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
- Per-widget insight components: `contentstudio-frontend/src/modules/analytics_v3/components/AiInsightsCard.vue` and its usage in `AnalyticsCardWrapper.vue` / `AccountStatistics.vue`. Remove the per-widget rendering.

---

## [FE] Add a unified Insights tab to social analytics

### Description
Add a single Insights tab to each social platform's analytics, matching the Meta/Google analytics pattern (hero summary, priorities, breakdown/plan), driven by the platform-level AI insights.

### Workflow
1. User opens a social platform's analytics and selects the Insights tab.
2. The user sees a unified AI summary for the selected account and date range: a hero summary, key priorities, and a breakdown, like Meta Ads and Google analytics.

### Acceptance criteria
- [ ] Each in-scope social platform's analytics has an Insights tab consistent with the Meta/Google unified insights layout (hero, priorities, breakdown/plan).
- [ ] The tab reflects the selected account and date range.
- [ ] Loading, empty ("not enough data to generate insights"), and error states are handled.
- [ ] The tab reuses the shared insights components/pattern rather than a new one-off design.
- [ ] Styling is theme-aware (works under white-label).

### UI copy
- Tab label: "Insights"
- Empty state: "Not enough data to generate insights for this period."
- Error state: "We could not generate insights right now. Please try again."

### Impact on existing data
None (presentation), assuming the insights payload exists.

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
- Reuse the unified insights pattern from the Meta Ads analytics: `contentstudio-frontend/src/modules/analytics/views/meta_ads/components/AIInsightsTab.vue` and `ai_insights/*` (`AIHero.vue`, `AIPriorities.vue`, `AIBreakdown.vue`, `AIPlan.vue`). Bring this pattern into `analytics_v3` for the social platforms.
- Data source: the per-platform `ai-insights` endpoints already present in the v1 analytics controllers; confirm they return a platform-level payload suitable for the unified tab.

---

## [BE] Serve platform-level unified insights per social platform

### Description
Ensure each social platform can return a single platform-level AI insights payload suitable for the unified tab (hero summary, priorities, breakdown), rather than only per-widget snippets. Only needed if the current insight generation is per-widget and cannot feed the unified tab directly.

### Acceptance criteria
- [ ] Each in-scope social platform can return a platform-level insights payload matching the structure the unified tab needs (summary, priorities, breakdown).
- [ ] The payload respects the account and date range.
- [ ] Insufficient-data cases return a clean empty result, not an error.
- [ ] The response structure matches what Meta/Google insights already return so the frontend can reuse the same components.

### Global quality and compliance checklist
- [ ] Multilingual support verified (insight text localized or handled)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research, not a contract.*
- v1 analytics controllers already expose `ai-insights` per platform (`contentstudio-backend/app/Http/Controllers/Api/V1/Analytics/*`). Align their payloads with the Meta/Google insights shape so the unified frontend components work unchanged. Reuse the Meta/Google insight generation approach where possible.
