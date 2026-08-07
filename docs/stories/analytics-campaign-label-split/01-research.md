# Research — Split Campaign & Label in Performance Analytics

## Current State

The **Performance Analytics → Campaign & Label report** (Analytics module, "Labels & Campaigns" report) currently treats campaigns and labels as **one combined dimension** in two places:

**1. Top header filter — one combined dropdown.**
`FilterBar.vue` renders a single `LabelAndCampaignSelect.vue` dropdown. That one dropdown contains **both** a "Campaigns" collapsible section and a "Labels" collapsible section stacked inside it, with a shared count badge ("X campaigns & labels selected") and a single **Apply** button. Selections are stored together in `selectedCampaignsAndLabels = { campaigns: [], labels: [] }` and persisted as one preference (`PREFERNCES_TYPES.campaigns_and_labels`).

**2. Section breakdown tables — combined rows.**
`OverviewSection.vue` renders three tables (`LableAndCampaignTable.vue`), each titled from a single string:
- **"Post breakdown"** (`type: breakdown`)
- **"Post impressions"** (`type: impressions`, paired with the Impressions graph)
- **"Post engagements"** (`type: engagements`, paired with the Engagements graph)

Each table's rows are built in the composable by merging **both** dimensions: `keys = [...selectedLabels, ...selectedCampaigns]` (see `transformBreakdownData`, `transformBreakdownDatabyImpression`, `transformBreakdownDatabyEngagements`). Every row already carries a `type` field of either `'labels'` or `'campaigns'`, so campaign rows and label rows are interleaved in one table with no way to view a single dimension.

## What Needs to Change

- **Header:** Replace the single combined `LabelAndCampaignSelect` with **two dedicated dropdowns** side by side — one **Campaigns** selector, one **Labels** selector — each with its own multi-select list, count badge, and Apply action.
- **Each section heading** ("Post breakdown", "Post impressions", "Post engagements"): add a small **dimension selector** (Campaigns / Labels) next to the title so the user views **one dimension at a time**. The selected dimension filters that section's table rows (and its paired graph, where present) to only that dimension's rows.
- No new data is needed — rows already carry `type`, and the two dropdowns map cleanly onto the existing `campaigns[]` / `labels[]` preference arrays. This is **client-side filtering + UI restructuring only**.

### Scope decisions (confirmed with PO)

- **Section selector is per-section and independent**, defaulting to **Campaigns**. Each of the three sections has its own Campaigns/Labels selector; changing one does not change the others. (Not persisted across visits.)
- **Empty dimension is disabled.** If the user has no labels selected in the header, the "Labels" option in each section selector is disabled (greyed, with a tooltip); likewise for campaigns. The selector auto-selects whichever dimension actually has data, so a section never lands on an empty dimension.

## Data / API Context

- Data comes from `useCampaignsAndLabelsQuery` and `useLabelCampaignSectionQuery` (TanStack Query). The API is keyed per campaign/label id and returns both dimensions together — **no backend change required**.
- Preferences already store `campaigns` and `labels` as **separate arrays**; splitting the header does not change the persisted shape.

## UX Reference (brief)

Competitors (e.g., Sprout Social, Metricool) keep tag/campaign report filters as **separate, clearly-labeled selectors** rather than one merged control, and let a report section scope to a single grouping dimension at a time — matches the requested direction.

## Mobile Context

Not applicable — this detailed Campaign & Label performance report is a **Web App** view only. No iOS/Android screens are affected.

## Files Involved

- `contentstudio-frontend/src/modules/analytics/views/performance-report/label-and-campaign/components/FilterBar.vue` — header layout; swap one selector for two
- `contentstudio-frontend/src/modules/analytics/views/performance-report/label-and-campaign/components/LabelAndCampaignSelect.vue` — split into a reusable single-dimension selector (Campaigns / Labels)
- `contentstudio-frontend/src/modules/analytics/views/performance-report/label-and-campaign/components/OverviewSection.vue` — add per-section dimension selector; pass selected dimension to each table/graph
- `contentstudio-frontend/src/modules/analytics/views/performance-report/label-and-campaign/components/LableAndCampaignTable.vue` — accept a `dimension` and filter rows by `type`; render the selector in the header row
- `contentstudio-frontend/src/modules/analytics/views/performance-report/label-and-campaign/composables/useLabelAndCampaign.js` — data already tagged with `type`; expose per-dimension filtered getters if helpful
- Graphs: `graphs/ImpressionsGraph.vue`, `graphs/EngagementsGraph.vue` — honor the section's selected dimension
- i18n: `src/locales/*/analytics.json` — new labels for the two dropdowns + dimension selector
