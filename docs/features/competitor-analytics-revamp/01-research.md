# Research — Competitor Analytics Revamp (Facebook + Instagram)

**Date:** 2026-08-13
**Scope:** Codebase research only. Market/competitor-tool research was skipped — the PO locked the scope to five specific UI changes on existing screens, so the useful research is the actual implementation.

> ### ⚠️ Checkout freshness — read this first
>
> The mounted `contentstudio-frontend` is on a **`features` branch from 2026-08-04, 158 commits behind `origin/develop`**. The first pass of this research was written against that stale tree and **got the export findings wrong**. Section 6 has been corrected against `origin/develop`.
>
> The other four areas — empty state, Add Competitor modal, dropdowns/tables/logos, and top/least posts — were **re-verified against `origin/develop`** and are unchanged there, so those findings stand. Files confirmed untouched on develop: `ReportsList.vue`, `ManageCompetitorsModal.vue`, `TopAndLeastPerformingPosts.vue`, `CompetitorsTable.vue`, `CompetitorsTableComponent.vue`, `FacebookCompetitorReport.vue`, `InstagramCompetitorReport.vue`, `AnalyticsDropdown.vue`.
>
> **Lesson for future research in this workspace:** check `git log HEAD..origin/develop` before trusting the mounted frontend, and read disputed files with `git show origin/develop:<path>`.

---

## 1. Where competitor analytics lives

Competitor analytics is part of the unified **`analytics`** module (the old `analytics_v3` module was merged in — there is no `analytics_v3` anymore, and route names keep a historical `*_v3` suffix that is *not* a module signal).

| Concern | Path |
|---|---|
| Route tree | `contentstudio-frontend/src/modules/analytics/config/routes/competitor.ts` → mounted at `/:workspace/analyze/*` |
| Layout shell | `views/competitor/MainAnalytics.vue` (sidebar + router-view) |
| Report list | `views/competitor/common/ReportsList.vue` |
| Plan-gate / upsell | `views/competitor/common/CompetitorAnalyticsLanding.vue`, `CompetitorDummyGraphs.vue`, `CompetitorUpgradeModal.vue` |
| FB report screen | `views/competitor/facebook/FacebookCompetitorReport.vue` (1,381 lines) |
| IG report screen | `views/competitor/instagram/InstagramCompetitorReport.vue` (993 lines) |
| Header / actions | `components/competitor/MainAnalyticsHeader.vue` |
| Add/manage modal | `components/competitor/ManageCompetitorsModal.vue` (458 lines) |
| Section components | 34 files under `components/competitor/` |
| Section composables | 22 files under `composables/competitor/` |
| PDF report layouts | `views/reports/FbCompetitorReport.vue`, `views/reports/IgCompetitorReport.vue` |
| PDF report backend | `contentstudio-social-analytics-go/src/services/reports/{catalog,builder,localization}/` |

All competitor API calls funnel through `fetchCompetitorInfo()` (`composables/competitor/useCompetitorsFactory.ts` / `useSectionData.ts`) → `services.ts` → the Go analytics service. Endpoints come from `composables/competitor/constants.ts` — never hardcoded.

---

## 2. Empty state — current behavior

**Important distinction that shaped this research:** `CompetitorAnalyticsLanding.vue` is **not** the empty state. It is the plan gate — it renders `CompetitorDummyGraphs` (backed by ~1,600 lines of static demo data) for plans without competitor access, and redirects qualifying users to the Instagram competitor overview. Do not touch it for the empty-state work.

The real "no competitor report yet" surface is **`ReportsList.vue`**:

```
<ReportTile :is-empty="true" />          ← a single dashed "create" tile, always first
<ReportTile v-for="listItem in allReports" ... />
```

So a user with zero reports sees a grid containing **one dashed tile and nothing else** — no headline, no explanation of what competitor analytics does, no guidance. There is also a separate **"Create new"** button in `MainAnalyticsHeader.vue` (rendered when `!showReportOptions`), so the create action exists in two places with no primary empty-state CTA.

**Strong precedent available.** The story *"[FE] Redesign the Analytics empty-state screens with connect-account onboarding"* already established an onboarding empty-state pattern for the 10 platform analytics screens: centered white card on the grey page background, platform logo, title, subtitle, a row of 4 info cards ("what you can do"), a primary themed CTA, and a **Learn More** link with a `?` icon. Competitor empty states should follow that same shape with an "Add your first competitor" CTA, so the two sets of analytics empty states are visually consistent.

---

## 3. Add Competitor modal — current behavior

`ManageCompetitorsModal.vue`, a `CstuModal` with id `manage-competitors-modal`, opened via the EventBus event `show-manage-competitors-modal` (the same modal serves both create and edit — it receives existing report data on edit).

**Current fields:**

| Field | Component | Notes |
|---|---|---|
| Report name | `TextInput` (size `lg`, radius `lg`) | **Placeholder only — no visible label.** Error: `manage_modal.title_error` |
| Competitor search | `TextInput` with `Search` icon + `Loader` | Platform-specific placeholder (FB vs IG). Disabled at 5 competitors with a tooltip. Error: `manage_modal.competitor_error` |

**Search results** render in a hand-rolled `<div>` panel (not a design-system dropdown): avatar, name (or slug), a `CircleCheck` verified badge for `blue_verified`, and city/country when present. Selected competitors sit in `ListItem` rows with a danger `ActionIcon` to remove.

**Footer:** an orange info pill (max-competitors notice) on the left, `Cancel` + `Continue` buttons on the right.

**Validation** only fires on submit (`validateAndSaveReport`) — both fields are checked at once, nothing validates as you type.

**Concrete modernization gaps found in the markup:**
- Hardcoded hex colors that break white-label theming: `border-[#E9EFF6]`, `text-[#2f8ae0]`, `text-[#b57e00]`, `bg-white!`
- Fixed dimensions: `dialog-class="w-150!"`, body `h-116!`
- The results panel is custom markup rather than `Dropdown`/`DropdownItem`
- No field labels or helper text — placeholders carry all the meaning
- A `TextInput as unknown as Component` cast because CSTU's `.d.ts` doesn't declare its slots

Max competitors is **5**, enforced in the UI.

---

## 4. Dropdowns, tables, and competitor logos — current behavior

### Dropdowns

The Facebook competitor report uses exactly **three** `AnalyticsDropdown` instances (`views/common/AnalyticsDropdown.vue` + `AnalyticsDropdownItem.vue`):

| # | Section | Trigger copy |
|---|---|---|
| 1 | Posting Activity by Post Type | "Posting activity by competitors ({type})" |
| 2 | Post Engagement Over Time | "Posts per day by {name}" — **competitor selector** |
| 3 | Post Reactions Distribution | "{title} {selectedCompany.name}" — **competitor selector** |

`AnalyticsDropdown.vue`'s own header documents it as a **"content-width / dynamic"** dropdown. That is exactly the "dropdowns open half" problem — the panel sizes to its content rather than to the card header. The newer pattern in the same repo (`components/reports/modals/ScheduleReportModal.vue`) uses the `@contentstudio/ui` `Dropdown` with **`match-width`**, `placement="bottom-start"`, plus `dropdown-classes` / `button-classes`. That's the target.

### Competitor logos in dropdowns

The competitor dropdown items currently render **name text only**:

```
<AnalyticsDropdownItem :selected="selectedCompany?.name === company.name" ...>
  <span class="text-gray-900 text-sm truncate">{{ company.name }}</span>
</AnalyticsDropdownItem>
```

The competitor objects **already carry an `image` field** — it's used in `TopAndLeastPerformingPosts.vue`, `ManageCompetitorsModal.vue`, and `CompetitorTile.vue`, each with the same fallback (`.../default/profile_default.svg`) on image error. **So adding logos is a pure frontend change — no API work.** `Avatar` from `@contentstudio/ui` is the component to use.

### ⚠️ Instagram has no dropdowns at all

`InstagramCompetitorReport.vue` contains **zero** dropdowns — its post-type selector is passed down as a prop (`:selected-type="selectedPostType"`). And per the Go widget catalog, the two FB competitor-selector charts have **no Instagram equivalent**:

> *"reactions / post-engagement-comparison / engagement-over-time are Facebook-only (Instagram has no equivalents)."*

**Implication for "do the same for Instagram":** dropdown restyling and logos can only apply to Instagram where Instagram actually has a comparable control. This needs a scope decision (see Open Questions).

### Tables

Competitor screens use bespoke table components: `CompetitorsTable.vue` (564 lines), `CompetitorsTableComponent.vue` (291 lines), and helpers `ComparativeTableHelper.vue`, `BioAnalysisTableHelper.vue`, `HashtagTableHelper.vue`, `PostingActivityTableHelper.vue`.

The **newer** table stack is `views/common/data_table/` — `AnalyticsDataTable.vue`, `AnalyticsBaseTable.vue`, `AnalyticsTableToolbar.vue` — already adopted by Meta Ads, Google Ads (6 tabs), and the label/campaign performance report. That is the "latest formatting" reference.

---

## 5. Top / least performing posts — current behavior

`TopAndLeastPerformingPosts.vue` (301 lines for the child `CompetitorTopLeastPosts.vue`, 858 for the parent) renders, **per competitor**, a two-column grid:

```
grid grid-cols-2 gap-28     ← left column = top posts, right column = least posts
```

…stacked once for every competitor in the report. The "Top performing" / "Least performing" headings appear **only once in the card header**, and each row's own top/least labels are revealed on **hover** (`hidden group-hover:block`). With up to 5 competitors that's up to 5 stacked two-column blocks — the "one big clutter" described in the brief.

**The PDF report already does this correctly.** The Go widget catalog registers **separate per-competitor widgets**:

```go
competitor_top_posts_%d    Type: WidgetTopPostCards   Platforms: fbAndIg   DataDeps: top_least_posts
competitor_least_posts_%d  Type: WidgetTopPostCards   Platforms: fbAndIg   DataDeps: top_least_posts
```

So the exported PDF already presents top and least as distinct sections, while the on-screen dashboard does not. The revamp brings the screen in line with the PDF.

**Data is already available** — each competitor object carries `top_5_posts` and `least_5_posts`. Per the earlier research for *"Add the Top Posts sort dropdown + Least Posts to Facebook, Instagram, LinkedIn & TikTok analytics"*: *"Least lists exist today only for Pinterest/YouTube/TikTok/**competitors**."* **This story is therefore frontend-only.**

**Pattern to copy:** `views/pinterest/components/TopAndLeastEngagingPosts.vue` (and the YouTube equivalent) — a dropdown in the section heading, 5 posts per list, no "show more" and no full-list modal.

---

## 6. Export report — corrected against `origin/develop`

> **This section was wrong in the first pass.** The stale checkout had no export modal at all, so the research claimed one needed building. On current `origin/develop` the section picker is **already shipped**. The PO was right: sections exist, language doesn't.

### What already exists

A shared **`components/reports/modals/ExportReportModal.vue`** (157 lines, new on develop) serves both competitor reports and platform analytics exports. The competitor header opens it:

```ts
// MainAnalyticsHeader.vue — handleExportReport()
// "Opens the shared Export modal so the user can trim sections; the modal owns the
//  export call and its success/error toasts from here."
$cstuModal.show('exportReport')
EventBus.$emit('export-report-sections', {
  network: `${currentPlatform.value}_competitor`,
  competitorReportId: reportId.value,
  dateRange: dateRange.value,
})
```

`views/common/ExportButton.vue` emits the same event for platform exports — **one modal, two callers.**

Inside the modal:
- **`views/common/ReportSectionPicker.vue`** (253 lines, new) — a dropdown built from `Checkbox`, `ListItem`, and `SearchInput`, with a **Select all** checkbox, a **section search**, a `{selected} of {total} sections selected` counter, an "All sections selected" state, a help tooltip, and inline validation.
- **`utils/reportSections.ts`** (1,117 lines, new, with a 195-line spec) — `reportSectionsFor(platform)`, `widgetIdsForSections(platform, keys)`, and `sectionSelectionEvent(...)`. It already defines section lists keyed `facebook_competitor` and `instagram_competitor`, each section mapping to the Go widget IDs.
- On submit it branches: competitor reports go through `useReportExport().exportReport({ competitorReportId, dateRange, action: 'render', widgets })`; platform reports go through `useRenderReportMutation()` with `widgets` in the payload.
- It already fires a Usermaven event: **`analytics_report_sections_customized`**.

`useReportExport.ts` on develop gained `widgets?: string[]` and sends `widgets: params.widgets ?? []`.

**Competitor section keys already defined** (from `reportSections.ts`):

| Facebook (`facebook_competitor`) | Instagram (`instagram_competitor`) |
|---|---|
| `competitors`, `performance_comparison`, `comparative_table`, `followers_comparison`, `post_type_activity`, `hashtags`, `post_engagement`, `post_engagement_over_time`, `post_reactions`, `top_and_least_posts` | `competitors`, `performance_comparison`, `comparative_table`, `followers_comparison`, `posting_activity_by_post_type`, `post_type_activity_chart`, `post_type_activity_table`, `hashtags`, `top_and_least_posts` |

Note the platform asymmetry is already encoded correctly: Instagram's list has no `post_engagement`, `post_engagement_over_time`, or `post_reactions`.

### The actual gap: language

`ExportReportModal.vue` has **no language selector**. Consequences today:

- **Competitor exports** silently use the UI locale — `useReportExport.ts` hardcodes `language: i18n.global.locale.value` and the modal never passes a language, so it can't be overridden.
- **Platform exports** don't send a language at all — the `renderReportMutation` payload has no language field.

The language selector the PO is comparing against lives in `components/reports/modals/ScheduleReportModal.vue` (+276 lines on develop): `getSupportedLanguages()` from `src/i18n` paired with a country-flag component, defaulting to the current locale.

### Backend: already supports it

- `contentstudio-social-analytics-go/src/models/reports/definition.go` — `Widgets []string` and `Language string` are both on the report definition, and the per-platform adapters pass `Language: def.Language` through to widget rendering.
- Platform analytics handlers already read `language` from the request and fall back to an `X-LOCALE` header (`req.Language = r.Header.Get("X-LOCALE")`).
- `services/reports/localization/localization.go` — 8 supported report languages: **`en`, `de`, `es`, `fr`, `it`, `pl`, `el`, `zh`**, anything else falling back to English. Matches the frontend's 8 locales exactly.

**So the language work is frontend-only for the competitor path** — the value already travels on that request; the change is sending the user's choice instead of the UI locale. *Not verified:* whether the `analytics/reports/save` endpoint (which is not in the Go repo — it's reached via `analyticsBaseUrl`) honors a language on the **platform** render path. Engineering should confirm that before wiring language into the platform branch.

### Two scope consequences

1. **The modal is shared.** Adding a language selector to `ExportReportModal.vue` gives it to platform analytics exports too, not just competitor. That's very likely desirable and matches "the latest one has the language as well" — but it widens the change beyond competitor analytics, so the PO should know.
2. **Tracking already exists here.** The modal fires `analytics_report_sections_customized`. Rather than adding a new event for language, the chosen language can be added to that existing event's payload — consistent with the "no new tracking" decision.

---

## 7. Cross-cutting notes

- **Mobile:** no impact. Analytics reports are web-only, and competitor analytics has no mobile surface. No `[iOS]`/`[Android]` stories.
- **Theming:** the competitor code carries a lot of hardcoded hex (`#E9EFF6`, `#2f8ae0`, `#b57e00`, `#3a4557`, `#979CA0`). Any story touching these files should move them to `primary-cs-*` / `cstu-*` tokens for white-label correctness.
- **i18n:** namespace `analytics`, keys under `analytics.competitor.*`, and every new key must land in **all 8 locale dirs** in the same change or the other 7 locales render raw keys.
- **Usermaven:** corrected against `origin/develop` — the export flow **does** already emit `analytics_report_sections_customized` from `ExportReportModal.vue`. (The "no report events exist" claim in the first pass came from the stale checkout.) Everything else in this revamp is a view-only restyle, so per guidelines §19 the default is no new events; the language choice can ride on the existing section-customized event rather than adding one.
- **Charts:** all charts go through `useEcharts()` — don't import ECharts directly.
- **Feature gating:** competitor analytics is plan-gated; non-qualifying plans see `CompetitorDummyGraphs`. Empty-state work must not collide with the gate.

---

## 8. Effort read per story area

| Story area | Where the work lands | Notes |
|---|---|---|
| Empty state | Frontend only | Strong precedent from the analytics onboarding empty-state story |
| Add Competitor modal | Frontend only | Existing save/search API unchanged |
| Dropdowns + tables + logos | Frontend only | `image` already in the payload; needs an IG scope decision |
| Top/least posts split | Frontend only | `top_5_posts` / `least_5_posts` already returned; PDF already splits them |
| Export language selector | Frontend only | The shared export modal and its section picker already ship; only a language field is missing. The competitor export request already carries `language` |

Nothing here requires new Laravel work, and nothing requires new Go work for the competitor path. The single unverified item is whether the platform (non-competitor) render path honors a supplied language — worth confirming before wiring language into that branch of the shared modal.

---

## 9. Open questions for the PO

1. ~~**The export modal.**~~ **Resolved.** The section picker already ships in the shared `ExportReportModal.vue`; only the language selector is missing. See the corrected section 6. The open sub-question is whether adding language to that **shared** modal (which also serves platform analytics exports) is acceptable scope, or whether it should somehow be competitor-only.
2. **Instagram dropdown parity.** Instagram's competitor report has no dropdowns at all, and the two Facebook competitor-selector charts (post reactions distribution, engagement over time) have **no Instagram equivalent** in the backend. So for Instagram, "same as Facebook" can only mean restyling the controls IG actually has. Options: (a) restyle/add logos only where IG has an equivalent control, (b) additionally introduce a dropdown for IG's post-type selector to match FB, (c) treat IG dropdowns as out of scope for now.
3. **Artifacts.** Please drop the empty-state and Add-Competitor-modal designs into `docs/features/competitor-analytics-revamp/mockups/` (the analytics empty-state story did the same with per-screen PNGs). The stories will reference the attached mockups for visual layout and specify copy, behavior, and AC in text.
4. **Competitor report created event.** Want a `competitor_report_created` Usermaven event as part of the modal story, or keep this revamp instrumentation-free?
