# Research (light) — Analytics improvements batch

Quick grounding only — these stories are intentionally high-level (the dev researches the approach and implements). Anchors below.

| Story | Anchor(s) | Note |
|---|---|---|
| Reports architecture (research) | `contentstudio-backend/app/Jobs/Analytics/GenerateReportJob.php` → `app/Services/Analytics/ReportsHelper.php` → `app/Services/Pdf/GotenbergService.php` | Reports are rendered by **Gotenberg** (headless Chromium, HTML report URL → PDF), not Puppeteer. The "loading component in the export" issue = the page is captured before its data finishes loading; long renders also hit timeouts. |
| Migrate analytics FE | `contentstudio-frontend/src/modules/analytics_v3/` (~71 files, mostly JS; ~4% TS; Pinia + some TanStack) and legacy `src/modules/analytics/` | v3 is the active module; already partly on TanStack via composables. Migration = finish TS conversion + standardize on TanStack. Pure refactor, no visual change. |
| Daily sync → current day only | `contentstudio-social-analytics-go/src/cmd/jobs/fetcher/*.go` | The scheduler cadence is per-account (~6h); the **data look-back window** (the "last 10–14 days" re-fetched each run) is set in the fetch/date-range params — dev to confirm exact spot. |
| Immediate sync token handling | `contentstudio-social-analytics-go/src/api/immediate_work_apis.go` + the platform fetch path it calls | The dispatcher maps some token errors, but the expired-token case appears to be swallowed **upstream in the fetch/API call**, returning success — dev to pinpoint. |
| Post thumbnails architecture | `contentstudio-frontend/src/modules/analytics_v3/composables/useCompetitorHelper.js` (`getMediaLink`), `contentstudio-backend/app/Repository/Utilities/MediaRepository.php` (`Media` model) | Thumbnails are platform-native CDN URLs (Meta/YouTube), stored on `Media`. Likely problem: expiring CDN URLs break thumbnails → wants a more durable architecture (cache/store/proxy). |
| Golang Sentry errors | `contentstudio-social-analytics-go/src/logger/sentry_hook.go` | Sentry auto-captures all zerolog `Error`+ events. The errors flagged "shouldn't be happening" are these captured events — investigate sources and fix/quiet them. |

**Scope:** 1 new epic (Analytics Reports Architecture Improvement) with a single research story for now, plus 5 standalone analytics stories. No design story needed (the one FE story is a no-visual-change refactor). No Shortcut/Helpin metadata blocks (per current direction).
