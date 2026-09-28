# Analytics reports access - research

Local research doc. Holds every codebase pointer, line number and gotcha so none of it leaks into the story bodies. Never recreated in the tracker.

---

## 1. Where the detail lives

| Area | Path |
|---|---|
| Analytics sidebar | `contentstudio-frontend/src/modules/analytics/components/competitor/AnalyticsSidebar.vue` |
| Analytics shell / layout | `contentstudio-frontend/src/modules/analytics/views/competitor/MainAnalytics.vue` |
| Sidebar route config | `contentstudio-frontend/src/modules/analytics/components/common/composables/useAnalyticsRoutes.ts` |
| Scheduled reports page | `contentstudio-frontend/src/modules/analytics/components/reports/MyReport.vue` |
| Download reports page | `contentstudio-frontend/src/modules/analytics/components/reports/DownloadReports.vue` |
| The modal | `contentstudio-frontend/src/modules/analytics/components/reports/modals/ScheduleReportModal.vue` |
| Entry point today | `contentstudio-frontend/src/modules/analytics/views/common/ExportButton.vue` |
| Report sections map | `contentstudio-frontend/src/modules/analytics/utils/reportSections.ts` |
| PDF renderer switch | `contentstudio-frontend/src/modules/analytics/components/PDFReports.vue` |
| Preference persistence | `contentstudio-frontend/src/modules/common/composables/useHelper.ts` → `setPreferenceStatus` |
| Profile store | `contentstudio-frontend/src/stores/core/useProfileStore.ts` |
| Preferences endpoint | `contentstudio-backend/app/Http/Controllers/Settings/UserPreferencesController.php` → `setPreferences` |

**Prototype:** https://claude.ai/artifact/4rw8PgVAEDK8SzURBQ42bc

---

## 2. Scope, locked 2026-09-17

Two problems raised from screenshots of the live app.

1. Manage Reports sits below the fold in the analytics sidebar and is not visible on ordinary screens.
2. Reports can be created from inside an analytics page but not from the Schedule Reports or Download Reports pages themselves.

Product decision on the sidebar: **Social Analytics and Manage Reports open by default, every open and close remembered at user level.** Not an accordion.

---

## 3. The five things that change how this is built

### 3.1 The sidebar overflows by design, and it gets worse every release

`AnalyticsSidebar.vue` renders seven `Collapsible` sections, each hardcoded `:default-open="true"` (lines 7, 58, 110, 151, 192, 224, 281, 324, 363). Nothing persists open state.

Measured content: 7 section headers + 20 nav rows ≈ **1,076px**. Available sidebar height on a 900px screen ≈ **818px**. Manage Reports is rendered last, so it is the section that drops off.

Social Analytics alone is 10 rows and grows with every network. Threads analytics already has a section map (see 3.4). **Reordering alone buys one release** - move Manage Reports up and Data Studio falls off instead. The default-collapsed state is the actual fix; the reorder is what makes the chosen default land above the fold.

Why the reorder is still needed: with only Social Analytics and Manage Reports expanded, Social Analytics alone runs to roughly 390px, and five collapsed headers plus gaps put a last-placed Manage Reports at roughly 700px - still clipped at 768px viewport height. Moving it to second position puts its three rows at roughly 390 to 535px, comfortably visible.

### 3.2 `Collapsible` already supports controlled state

`@contentstudio/ui` `CollapsibleProps` exposes `modelValue`, `defaultOpen`, `disabled`, `class`, `hideArrow`, `buttonClass`, `selectedClass`. So this is state plumbing in `AnalyticsSidebar.vue`, **not a UI library change**.

### 3.3 User preferences need no backend work

`UserPreferencesRequest` validates only:

```php
'key' => 'required|string',
'value' => 'required',
```

No whitelist. `UsersRepository::setPreferences($user_id, $key, $value)` writes any key generically. A new preference key is **frontend-only**.

Follow the established pattern, documented in `useAiToolFavorites.ts`: optimistic Pinia update via a profile store setter, then `void setPreferenceStatus(KEY, payload)`.

**The trap, already hit once on this codebase.** From the comment in `useAiToolFavorites.ts`:

> Persisted as `{ tools }` - the backend's `required` rule on `value` rejects a bare empty array, so wrapping keeps an emptied list saveable.

If a user collapses every section, a bare `[]` is rejected by `required`. Wrap it, e.g. `{ open: [...] }`, so the all-collapsed state saves. Add the key to `ProfilePreferences` in `src/types/common/auth.ts` and to `getDefaultProfile()` in `useProfileStore.ts`.

Suggested key: `analytics_sidebar_sections`.

### 3.4 Two platforms have report sections but no renderer

`reportSections.ts` `REPORT_SECTIONS` contains **`bluesky` (9 sections) and `threads` (12 sections)**. `PDFReports.vue`'s component switch covers Twitter, Pinterest, LinkedIn, Facebook, Instagram, TikTok, YouTube, GMB, Meta Ads, Google Ads, Campaign & Label, Overview, and the two competitor reports - **and nothing else**.

Build the source picker off the section map alone and you ship a blank PDF for Bluesky and Threads. Both must be excluded. Web Analytics (Usermaven) and Data Studio have no report path at all.

Worth a separate ticket: Threads analytics is being built now, and its PDF renderer is a known gap rather than a surprise.

### 3.5 The modal cannot open cold, and that is the whole story

Every creation path runs through `ExportButton.vue`, which emits an EventBus payload scraped from the current page:

```
accounts, network, startDate, endDate, topPosts, allAccountsSelected, labels, campaigns
```

`ScheduleReportModal.vue` sets `network` from that payload (lines 654, 688, 748, 789) and **never renders a control to change it**. Everything downstream keys off it:

- account list is filtered by network (see the comment at line 430)
- `reportSectionsFor(network)` drives the section list (line 280)
- `isDropdownOptionAllowed` (line 1266) filters the report type list
- the Report Type field is hidden entirely by `v-if="network !== 'campaign-and-label'"` (line 1371)

No network means no accounts, no report types and no sections. **This is a missing first step, not a forgotten button.**

---

## 4. The real option sets

### Report sources that actually produce a PDF

| Area | Sources |
|---|---|
| Social Analytics | Overview, Facebook, Instagram, X (Twitter), LinkedIn, TikTok, YouTube, Pinterest, Google Business |
| Competitor Analytics | Facebook Competitors, Instagram Competitors |
| Ad Analytics | Meta Ads, Google Ads |
| Performance Analytics | Campaign & Label |

### Report types - `optionsDropdown`, gated by `isDropdownOptionAllowed`

The three `type: 'overview'` entries are hidden unless `network` is `group` or `individual`.

| id | value | Shown for |
|---|---|---|
| `grouped_overview` | `group` | Overview only |
| `overview_single_pdf` | `single-pdf-overview` | Overview only |
| `overview_multiple_pdf` | `multiple-pdf-overview` | Overview only |
| `platform_single_pdf` | `single-pdf-detailed` | Everything except Campaign & Label |
| `platform_multiple_pdf` | `multiple-pdf-detailed` | Everything except Campaign & Label |

So the same control shows **five options, two options, or nothing at all** depending on step one.

### Section counts, from `reportSections.ts`

Google Ads 18 · Facebook 13 · YouTube 13 · Meta Ads 12 · Instagram 11 · Facebook Competitors 11 · LinkedIn 10 · Instagram Competitors 10 · GBP 8 · Pinterest 8 · TikTok 7 · Overview 6 · X 6 · Campaign & Label 6.

Facebook's 13 is why the live screenshot reads "13 of 13 sections selected".

`widgetIdsForSections` returns `[]` when everything is selected, which the engine reads as "render the full default report". That keeps a full-selection export byte-identical to today's and needs no backfill.

---

## 5. What fits with no change

- `ScheduleReportModal` is already mounted at shell level in `MainAnalytics.vue`, so it is in scope on the reports routes today. No mounting work.
- Edit mode already exists via the `edit-schedule-report` EventBus event, and already reconstructs network, accounts and sections from a saved report.
- `CstDateRangePicker` is already used in `AnalyticsFilterBarWrapper.vue` and can be reused for the new Export date range.
- `ReportSectionPicker.vue` already exists for the sections control.
- Existing validation copy covers most of the new flow: `select_report_type`, `select_at_least_one_account`, `add_email_address`.

---

## 6. Loose ends found while building the prototype

- **Broken Tailwind class.** `MainAnalytics.vue:5` has `max-h-[calc(100dvh - 60px)]`. Spaces inside an arbitrary value mean Tailwind never generates the class. Harmless today because the flex parent sets the height, but it reads as if it does something. One-line cleanup.
- **Reports Settings.** A page visited twice a year holding a permanent sidebar row. Moving it into workspace settings would cut Manage Reports from three rows to two. Out of scope here, worth a PO decision.
- **Send PDF as Email** is the third entry point in `reportTypeDetail` and has no list page of its own, so it stays reachable only from an analytics page. Deliberately not in scope.
- **Analytics has zero Usermaven tracking today.** `userMaven.track(` does not appear anywhere under `src/modules/analytics/`. The one event added in the FE story exists specifically to answer whether the new entry point gets used.

---

## 7. Open questions to answer before estimating

1. Does the sidebar preference need to survive a workspace switch? Assumed yes, it is a user preference not a workspace one, but confirm with the PO.
2. Should the Export date range cap at a maximum span? Large ranges on Google Ads with 18 sections are the slowest reports we generate.
3. Is `analytics_report_created` the right event name, or should report creation stay untracked in line with the rest of the module?
