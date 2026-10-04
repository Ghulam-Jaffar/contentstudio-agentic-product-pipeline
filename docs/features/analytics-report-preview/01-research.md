# Research: Analytics Report Preview

**Date:** 2026-10-01
**Feature:** A live preview of the analytics PDF report inside the Export, Email and Schedule report modals, so users see what they are about to generate or send.

**Prototype (canvas):** https://claude.ai/artifact/6qpafL7Usef92RCxFVRw2z

---

## 1. Problem in one paragraph

When a user exports, emails or schedules an analytics report, the modal asks for a name, language, report type, accounts and sections, but shows nothing of the result. The report type names alone ("Individual overview reports (separate PDFs)" vs "Detailed insights report (single PDF)") don't tell a user how many files they will get or what is in them. They find out by generating the report, opening it, and going back if it's wrong. For an emailed or scheduled report the mistake reaches a client or manager before the user sees it.

---

## 2. Competitor and industry research

### What the feature is

A preview shows what the exported, emailed or scheduled PDF will look like before the user commits. It matters most to agencies and teams who send reports to clients or managers: once a PDF is emailed or a schedule fires, a bad layout, a missing section or the wrong branding can't be taken back. A preview builds confidence, cuts the "export, open, check, fix, re-export" loop, and explains what each option changes.

### Competitor analysis

| Tool | Has report preview? | How | Real or sample data | Notes |
|---|---|---|---|---|
| Agorapulse | Yes | Live preview while configuring. "PDF live preview" updates in real time as sections change. Report Studio opens a full-screen preview with export options | Real | The live preview leads to both Download PDF and Send PDF. Report Studio options (AI summary, orientation, author) don't update live. Sources: https://support.agorapulse.com/en/articles/11114496-how-to-export-agorapulse-reports , https://support.agorapulse.com/en/articles/16043113-how-to-export-reports-in-report-studio |
| Sprout Social | Partial | WYSIWYG builder: the Report Builder page is the report. A PDF-specific preview is unverified | Real | Offers live vs snapshot scheduled delivery. https://support.sproutsocial.com/hc/en-us/articles/115005750066-Report-Builder |
| Hootsuite | No | The export/schedule modal covers format, frequency and recipients only | n/a | https://help.hootsuite.com/hc/en-us/articles/1260804306709-Export-an-Analytics-report |
| Buffer | Partial | The Reports page is the preview, with no separate step before Export as PDF | Real | https://support.buffer.com/article/534-creating-custom-analytics-reports |
| Metricool | Partial | Template editor previews branding. Scheduled reports have "Send a test report now" | Real for test sends | https://help.metricool.com/en/article/how-to-generate-reports-17aytqg/ |
| Sendible | Partial | Drag-and-drop builder with an in-browser preview that can be shared as a link | Real | Help article blocked. Details unverified |
| Publer | Unverified | Exports to PDF straight from each tab. No preview documented | n/a | |
| Later | No | Report saved through the browser's Print to PDF | n/a | https://help.later.com/hc/en-us/articles/1500004246721-Share-Your-Performance-Report |
| Loomly | Unverified | One Export Analytics button. No preview documented | n/a | |
| SocialBee | Unverified | Export tab picks metrics, logo and description. No preview documented | n/a | |
| AgencyAnalytics | Yes | WYSIWYG editor with a sample vs live data toggle that remembers the choice. Scheduled reports can require approval before sending | Both | The closest analog to this plan. https://help.agencyanalytics.com/en/articles/2392859-toggle-between-sample-data-and-live-data-in-the-report-editor |
| Whatagraph | Yes | WYSIWYG builder. PDF export chooses merged vs per-tab pages | Real | https://help.whatagraph.com/en/articles/6254357-how-to-download-reports-as-a-pdf |

### Common patterns

- **The dashboard doubles as the preview.** Most tools treat the on-screen report as the preview and have no PDF-faithful preview in the export step.
- **Export modals are thin.** They cover format, dates, recipients and frequency, and the user configures them without seeing a result.
- **Only Agorapulse puts a live preview in the export flow itself.**
- **Scheduled reports get trust features instead.** Examples are test sends, approval holds, and live vs snapshot delivery.

### Worth borrowing

- **Agorapulse:** a live preview that updates as sections are toggled, on the same surface as both Download and Email.
- **AgencyAnalytics:** a clearly labelled sample-data mode and a remembered per-user preference. This validates both halves of our plan.
- **Metricool:** "Send a test report now" for schedules. A possible later follow-up.

### Table stakes vs delighters

- **Table stakes:** PDF export with branding, dates and section choice. Email and scheduled delivery.
- **Delighters:** a live preview inside the modal that reacts to every option; a preview in the real PDF layout; sample data so it loads instantly and works for new or empty accounts; a remembered on/off preference.

### Recommendation

Build the Agorapulse pattern with AgencyAnalytics' remembered preference. Among the 10 social media tools we track, only Agorapulse has this, so it is a real differentiator against Hootsuite, Buffer, Later, Loomly, Publer and SocialBee. Keep the "sample data" label permanently visible so no one mistakes example numbers for a client's real figures.

---

## 3. Codebase analysis

### The modals

All three modals live in `contentstudio-frontend/src/modules/analytics/components/reports/modals/`:

- `ScheduleReportModal.vue` (1,661 lines, marked `TODO(analytics-v3)` for migration). Despite the name, this is the modal behind **Export, Email and Schedule** for Overview, every social platform dashboard (Facebook, Instagram, LinkedIn, TikTok, YouTube, Pinterest, X, Threads, Bluesky, GMB), and Meta Ads and Google Ads. It holds export name, language, report type (5 options, hidden for Campaign & Label), social accounts, sections, and email and schedule fields. Its title and button change with the action ("Export Report" / "Send PDF as Email" / "Schedule PDF as Email").
- `ExportReportModal.vue` (243 lines) handles competitor exports (`network: '<platform>_competitor'`, `competitorReportId`, `dateRange`) and Campaign & Label exports. It holds language and sections.
- `SendReportByEmailModal.vue` (232 lines) handles Campaign & Label email.

Routing is in `src/modules/analytics/views/common/ExportButton.vue` (through `AnalyticsFilterBarWrapper.vue`) and, for competitor reports, `src/modules/analytics/components/competitor/MainAnalyticsHeader.vue:368`, which has export only (no email or schedule). All modals are mounted in `components/AnalyticsMain.vue` and `views/competitor/MainAnalytics.vue`, and are opened through `$cstuModal.show(...)` plus an EventBus event.

### Sections and report types

- `src/modules/analytics/utils/reportSections.ts` defines `REPORT_SECTIONS` per platform key (`overview`, `facebook`, `instagram`, `linkedin`, `twitter`, `pinterest`, `bluesky`, `threads`, `tiktok`, `youtube`, `gmb`, `meta_ads`, `google_ads`, `campaign_label`, `facebook_competitor`, `instagram_competitor`, `youtube_competitor`). Each section has an `i18nKey` that reuses the dashboard heading. `widgetIdsForSections()` maps sections to Go report-engine widget ids.
- `views/common/ReportSectionPicker.vue` is the existing section picker.
- Report type options (`grouped_overview`, `overview_single_pdf`, `overview_multiple_pdf`, `platform_single_pdf`, `platform_multiple_pdf`) are defined in `ScheduleReportModal.vue` around line 145, with copy under `analytics.common.schedule_report_modal.report_options` in `src/locales/<lang>/analytics.json`.
- Report languages: `getSupportedLanguages()` in `src/i18n/index.ts` (en, fr, de, it, es, pl, sv, cs, zh, el).

### How the PDF is produced today (two engines)

1. **Every render first saves a report record.** `contentstudio-backend/app/Http/Controllers/Analytics/Analytics/AnalyticsReports.php:73` calls `ReportsRepo::createOrUpdate()` before acting on `action: render|save`. Scheduled runs do the same (`ReportsHelper::runScheduledReport`, around line 283).
2. **Go report engine.** `ReportsHelper::createPDFFile` tries `GoReportsClient::supports($report)` first. This path is flag- and platform-gated and honours the `widgets` list.
3. **Vue page through Gotenberg (fallback).** `ReportsHelper::generatePDF` (line 392) sends Gotenberg to `{LUMOTIVE_APP_URL}/{workspace}/analytics/reports/{report_id}` and waits for `window.pdfReportReady === true`.
   - Route: `contentstudio-frontend/src/modules/analytics/config/routes/analytics.ts:148`, path `reports/:reportId?`, name `analytics_pdf_report`, `meta.guest: true`.
   - Component: `src/modules/analytics/components/PDFReports.vue`.
   - Data: `src/modules/analytics/composables/usePDFReports.ts:1753` loads the saved report through `getReportsServiceLocal` → `fetchReportsApi` (`src/api/analytics-store.ts:369`, POST `analytics/reports/show`). Then `getAnalyticsForReports()` (around line 745) dispatches per type: `fetchFacebookReport`, `fetchOverviewReport`, `fetchMetaAdsReport`, `fetchGoogleAdsReport`, `fetchLabelCampaignReport`, and so on.
   - Readiness: `markPdfReady` / `markPdfFailed` at `usePDFReports.ts:224-250`, after `waitForAiInsights` and `waitForPdfCaptureReady`. There is a 360s hard timeout.
   - Templates: `src/modules/analytics/views/reports/`: `FacebookReport_v2`, `InstagramReport_v2`, `LinkedinReport_v2`, `TiktokReport`, `TwitterReport`, `YoutubeReport`, `GmbReport`, `MetaAdsReport`, `GoogleAdsReport`, `OverviewReport`, `FbCompetitorReport`, `IgCompetitorReport`. Label & Campaign and Pinterest templates are imported from elsewhere.
   - **The Vue templates do not filter by the selected widgets.** Each type renders its fixed set. Section selection is honoured on the Go path.

**Implication for the preview.** "Render the real report page with sample data" is not one answer:

- For areas printed by the Go engine, the Vue templates may not match the printed PDF.
- For areas printed through Gotenberg, the templates match but don't respect section selection.
- The preview also must not create report records or call analytics APIs on every option change.

This is why the epic has a `[Research]` ticket for the technical approach.

### No sample-data mode exists

The only nearby concept is the sample workspace (`src/composables/useSampleWorkspaceAnalytics.ts`, `isSampleWorkspace` in `src/utils/demoWorkspace.ts`). It hides Export, Share and Sync for read-only showcase workspaces and supplies no mock data.

### Per-user preference storage (for the remembered on/off choice)

- **Frontend:** `setUserPreferencesApi({ key, value })` in `src/api/profile.ts:186` and `src/api/analytics-composables.ts:216` posts to `preferences/setPreferences` (`src/config/api-utils.ts:210`). Values are read back from `profileStore.getProfile.preferences.<key>`.
- **Backend:** `contentstudio-backend/routes/web/settings.php:78` routes to `UserPreferencesController::setPreferences` (line 222), then `UsersRepository::setPreferences` (line 896), which writes `preferences.<key>` on the user document. Any string key is accepted.
- **Existing precedent:** `src/stores/analytics/useAnalyticsSidebarSectionsStore.ts`, key `analytics_sidebar_sections`. It updates locally first and fails silently, with the comment "Stored as a user preference rather than in localStorage so the choice follows the person". The preview toggle should do the same, for example with key `analytics_report_preview`.
- **Conclusion:** no backend work is needed for the preference.

### Existing Usermaven events

- `analytics_report_sections_customized`: `{ platform, action, sections_selected, sections_total, language? }`. Fired from all three modals, only when fewer than all sections are picked.
- `analytics_report_created`: `{ source, report_type, entry_point }`. Fired from `ExportReportModal`, `SendReportByEmailModal` and `useReportSourceFlow.ts:189`.

### Mobile

- The Flutter app is out of scope (PO, 2026-10-01).
- The **web app on small screens** is in scope: tablet and phone layouts are covered by their own story.

### Related existing work

- **Analytics Report Customization and Overview Card Cleanup** (Helpin epic, completed 2026-08-14) added the section picker this preview builds on.
- **Report names for exported and emailed reports** (Helpin epic, in progress) added the Export Name field.

---

## 4. Prototype decisions carried into the stories

These were explored on the canvas and confirmed by the PO on 2026-10-01:

- **Option B:** a preview in the real report layout, filled with sample data.
- **Option A, the labelled wireframe,** is kept as the loading state that shows while B renders.
- **The preview is optional.** It is on by default at rollout. Turning it off is remembered per user and can be turned back on.
- **Layout:**
  - Preview off: the current modal layout, with options in two columns.
  - Preview on: options move to a left column and the preview sits on the right.
- **Separate PDFs:**
  - Desktop: a strip of file tabs with previous/next arrows and a count.
  - Phone: a file picker dropdown.
- **One PDF covering several accounts:** "Jump to" chips that go to each account's chapter.
- **Small screens:**
  - Tablet: the preview opens as a side panel.
  - Phone: the modal is a full-screen sheet with Settings and Preview tabs.
- **Scope:** every report area (social overview and platform dashboards, competitor, Campaign & Label, Meta Ads and Google Ads). No Flutter.
