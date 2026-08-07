# 01 — Research: X (Twitter) Analytics Pay-Per-Use Metering + Wallet Gating

**Feature slug:** `x-analytics-metering`
**Date:** 2026-08-03
**Pipeline:** `/feature` — Step 1 (Research)

---

## 0. What this feature is (and why now)

X's API is **pay-per-use by default** (since Feb 2026). For analytics we make exactly **two read calls per sync**, both confirmed in `contentstudio-social-analytics-go`:

| Call | Endpoint | X meters | Per-sync volume |
|---|---|---|---|
| **user read** | `GET /2/users` (`FetchUserInfo`) | ~$0.010 per user lookup | exactly **1** per sync |
| **post read** | `GET /2/users/:id/tweets` (`FetchUserTweets`) | **$0.005 per post read** (per returned tweet) | **N tweets** the user selected |

Because ContentStudio's app reads **other customers'** accounts (not its own), these are **standard reads ($0.005)**, not "owned reads" ($0.001). Metering is per **returned object**, so the raw X cost of one sync is:

> **X cost per sync ≈ (N × $0.005) + $0.010**, then ContentStudio applies its markup per the wallet's config-driven pricing.

The number of API *requests* (`1 + ceil(N / pageSize)`, one page when N ≤ 100) matters only for **rate limits**, not the bill — the bill scales with **N (tweet count)**.

**The feature:** pass this per-read cost through to users via the **existing prepaid X dollar wallet**, gate X analytics behind an **in-page unlock/consent card**, and show a **live projected cost before every sync** so there is no bill shock. Credit-consumption **visibility is the #1 priority**.

> ⚠️ **Reframe vs the original brief:** X analytics cost is driven by **tweet count**, not date range. There is no date-range option for X sync — the X sync modal collects only a tweet-count dropdown. All cost math and the live preview key off tweet count.

---

## 1. Competitor & Industry Research

### 1.1 The market context

X's pay-per-use rates: **$0.005/post read**, $0.010/user read, capped at ~**2M reads/month (~$10k)**, then a cliff to Enterprise (~$42k+/mo). The legacy $200 "Basic" tier was deprecated (May 2026) and the free tier is gone. Net: **every X analytics refresh now has a real, variable marginal cost.** Users are already primed by the X-pricing news, so transparent metering reads as *fairness*, not a ContentStudio downgrade.

### 1.2 Competitor analysis

| Competitor | X analytics still offered? | How they handle X API cost | Usage model | Cost-visibility UX |
|---|---|---|---|---|
| **Metricool** | Yes, capped at 30 days history | **Flat add-on: $10/mo per connected X account** ($5 grandfathered) | Per-account monthly add-on | Fixed line item, no per-sync preview |
| **Agorapulse** | Yes, via **"X Plus"** add-on | Splits cheap "X Lite" (publish) from **~$50/profile/mo** "X Plus" (analytics + inbox) | Per-profile add-on | Tier described at purchase |
| **Publer** | Paid plans only | Absorbs cost, **gates X off the free plan** (~$5/mo minimum) | Bundled in subscription | "X requires a paid plan" gate |
| **Sprout Social** | Currently limited/unavailable | Absorbed into high seat price ($399/seat); API gated to Advanced | None | None user-facing |
| **Hootsuite** | Yes (enterprise partner) | Blanket **50–250% price increase**, citing API fees | None | Cost buried in plan price |
| **Buffer** | Publishing yes; API analytics no | Long-standing enterprise X deal; "won't impact users" | None | N/A |
| **Later** | **No — dropped X entirely (2025)** | Exited rather than pay | N/A | N/A |
| **Loomly** | **No — dropped X (2025)** | Exited X support | N/A | N/A |
| **Sendible** | Limited | Narrowed X analytics scope | None | N/A |
| **SocialBee** | Yes, comprehensive | Bundled into paid plans | None | N/A |
| **Brandwatch** | Yes (listening/partner) | Enterprise contract | None | N/A |
| **Rival IQ** | Yes (benchmarking) | Not disclosed | None | N/A |

### 1.3 Key finding — market whitespace

**No competitor exposes X's per-read marginal cost to end users, and none shows a live per-sync cost preview.** The market splits three ways: **(a) flat monthly add-on**, **(b) hide cost in plan price**, or **(c) drop X**. A **prepaid dollar wallet with a live, itemized per-sync estimate is genuinely novel** in this category. ContentStudio can own *"the tool that shows you exactly what X analytics costs, before you spend it."*

### 1.4 Cross-industry cost-transparency patterns (AI-credit / usage SaaS)

- **Cost shown at the point of action** — Jasper shows an action's credit cost next to the Run button; the ShapeofAI "Cost estimates" pattern puts estimates inline next to the input and on the action button, never buried in settings.
- **Live recalculation as inputs change** — Firefly / ElevenLabs / Krea update the projected cost instantly as the user changes options. Most relevant pattern for us (recalc as tweet count changes).
- **Ranges, not false precision + estimated-vs-actual reconciliation** — show "~$0.40–$0.60" when the exact count is uncertain, then reconcile to the real charge after the run.
- **Cost-driver breakdown** — "142 posts × $0.005" builds trust and teaches users how to cut cost.
- **Prepaid balance + hard stop at zero** — OpenAI model: deposit upfront, draw down, zero balance = blocked until top-up.
- **Auto-recharge + user-set monthly cap + staged alerts (50/80/95%)** — table stakes to avoid mid-work interruption and bill shock.
- **Threshold-based confirm** — only interrupt with a confirm dialog when an action exceeds a user-set threshold; keep cheap actions frictionless.
- **Gate at the feature, after the "aha"** — trigger the unlock prompt at the locked feature, framed as unlocking value.

### 1.5 Table stakes vs delighters

- **Table stakes:** always-visible balance + one-tap top-up; cost shown before the sync, in dollars, at the action; never charge silently; hard stop at zero with a clear message; usage/transaction history; low-balance alert.
- **Delighters:** live recalculation as options change; cost-driver breakdown + estimated-vs-actual; auto-recharge + monthly cap; cheaper-path nudges ("sync 30 tweets vs 100"); threshold-based confirm; reuse recently-synced data when fresh.

### 1.6 Recommended approach (synthesis)

1. **Opt-in, metered capability behind an in-page unlock/consent card** at the X analytics view — explain *why* X analytics now costs money (X went pay-per-use), state the model plainly, require explicit consent before the first charge. Keep publishing separate (already its own credit path).
2. **Charge against the existing prepaid X dollar wallet** (do not invent a second currency) — always-visible balance, prominent top-up, hard stop at zero, entries in the shared usage ledger.
3. **Make per-sync cost impossible to miss** — projected cost on the sync button and in the sync modal, in dollars, recalculating live as tweet count changes; show it as a range if needed and reconcile to the actual charge after.
4. **Graduated friction** — cheap syncs run with the estimate shown; only syncs above a threshold get a confirm dialog.
5. **Spend controls up front** — reuse the wallet's auto-recharge + monthly spending limit; low-balance / cap alerts.
6. **Nudge cheaper choices** — default to a modest tweet count; reuse fresh cached data instead of re-syncing.
7. **Honor BR-12** — never expose X's raw cost or ContentStudio's markup; show only the dollar price the user pays.

---

## 2. Codebase Analysis

> Frontend paths are relative to `contentstudio-frontend/`; Go paths to `contentstudio-social-analytics-go/`. Current FE branch: `develop`.

### 2.1 Current X analytics UI

- **Routes:** `src/modules/analytics/config/routes/analytics.ts` mounts platform dashboards under `/:workspace/analytics/*`; shell is `src/modules/analytics/components/AnalyticsMain.vue`.
- **X view:** `src/modules/analytics/views/twitter/MainComponent.vue` composes `OverviewSection.vue` + `PostsSection.vue` inside `FilterBar.vue`, with `AnalyticsTabsHeader.vue` / `TabsComponent.vue` chrome.
- **Layout pieces** (`views/twitter/components/`): `CardsComponent.vue` (overview summary cards), `OverviewSection.vue`, `PostsSection.vue` (uses shared `views/common/AnalyticsPostsTable.vue`), `TwitterPostModal.vue`, `graphs/*`.
- **State/logic:** `views/twitter/composables/useTwitterAnalytics.ts`; types `views/twitter/... types/twitter.ts`; section queries `queries/useAnalyticsQueries.ts` → `src/api/analytics.ts`.

### 2.2 A per-sync credit readout ALREADY exists

- `useTwitterAnalytics.ts` defines `routes.CREDITS_USED_COUNT = 'creditsUsed'`, a `creditsUsedCount` ref, and `applyFetchedData` for it.
- `types/twitter.ts` → `TwitterCreditsUsedResponse { credits: { credits_used } }`.
- Rendered as a query in `OverviewSection.vue` (`creditsUsedQuery`) and surfaced in the auto-sync tooltip `components/common/PlatformTooltip.vue` (i18n `analytics.json`: *"X (Twitter) Analytics credits used: {count}"*).

**Implication:** consumption *tracking* per sync partly exists. Missing: prepaid **balance**, **pre-sync cost preview**, **balance gate**, **unlock card**.

### 2.3 Sync mechanism & options (two paths)

Both live in the shared header `views/common/AnalyticsTabsHeader.vue`:

1. **Manual / immediate sync** — the "Sync Data" button opens `components/common/SyncDateRangeModal.vue`; `@refresh` → `handleSyncData()` → `composables/useManualSync.ts` `triggerManualSync(...)` → `triggerManualSyncApi({ workspace_id, account_id, platform, start_date, end_date, n_tweets })` (`src/api/analytics-composables.ts`, URL `…api/analytics/triggerJob`).
   - `SyncDateRangeModal.vue` has two modes. **Date mode** (non-Twitter): inline range picker. **Twitter mode** (`mode="twitter"`): a **tweet-count dropdown only** — `[10, 20, 30, 50, 80, 100, 120, 150]`, default 30. Emits `{ nTweets }` for Twitter.
2. **Scheduled job settings** (X-only) — the header gear opens `$cstuModal.show('twitter_job_settings')` → `views/twitter/components/TwitterJobSettingsModal.vue`. Options: `job_type` (daily / weekly / monthly / never), `trigger_day`, `post_count` (`tweets_to_process`, same `[10..150]`). Persisted via `createTwitterJobApi` / `updateTwitterJobApi`; `triggerTwitterJobApi` fires an immediate run.

**Implication:** there are **two consumption triggers** — manual sync **and** recurring scheduled syncs. Cost visibility must cover **both**: the sync button + modal *and* the scheduled-job settings modal (project the recurring per-run cost).

### 2.4 Sync → X API mapping (confirmed end-to-end)

- `GET /2/users` once per sync — `src/clients/social/twitter.go:323`, called at `src/services/twitter/twitter-fetcher/run.go:299`.
- `GET /2/users/:id/tweets` paginated — `src/clients/social/twitter.go:256`; sets `max_results` + `pagination_token` only. **No `start_time`/`end_time`** — X analytics is **count-based**. Loop at `run.go:326-448`: `requestedPostCount := order.NTweets` (default 30 if ≤ 0); `pageSize = 50` when paginating else `min(NTweets, 100)`, hard cap 100.
- **Tweet count wiring:** scheduled path `src/cmd/jobs/fetcher/twitter.go:308` (`PostCount: jobSetting.PostCount`); manual path FE `n_tweets` → `src/models/api/immediate_work_request.go` → `src/api/immediate_work_apis.go` → kafka `src/models/kafka/twitter.go` `NTweets`. Scheduled `PostCount` and manual `NTweets` converge on the same `order.NTweets` the fetcher reads.
- **Conclusion:** the frontend fully controls tweet count (per-account for scheduled, per-sync for manual). It does **not** control a date range for X.

### 2.5 X wallet UI state — what exists vs what must be built

**The prepaid dollar wallet does NOT exist on `develop`.** Grep for `x-wallet|pay-per-use|top-up|auto-recharge|spending limit|prepaid|wallet balance` returns nothing in app code. What exists is an adjacent **older monthly-quota + one-off addon** posting-credits system, plus the analytics `credits_used` readout:

- **Composer posting credits (old model, complete on develop):** `src/modules/composer/utils/xCredits.ts` (`X_CREDIT_PLAIN = 1`, `X_CREDIT_WITH_LINK = 15`, `calcXCredits`), `…/posting-schedule/composables/useTwitterCreditPreview.ts`, `components/TwitterPostUsageAlert.vue`, lock gate `…/account-selection/aside/useTwitterLockGate.ts`.
- **Addon purchase / pricing (complete):** `src/components/common/TwitterPostingAddon.vue` (300 credits per $5/mo pack, Paddle checkout, POSTs `{ x_posting_credits }`); billing plumbing in `src/modules/billing/constants/addonKeys.ts`, `billingAddonCatalog.ts`, `LimitItem.vue`, `useAdjustLimits.ts`, `AdjustLimitsModal.vue`.
- **Backend wallet foundation (unmerged branch):** `origin/feature/cont-2623-story-1-credit-consumption-engine-foundation` — the "credit consumption engine" referenced by `xCredits.ts`. Heavily diverged from `develop`; treat as backend contract, not a clean UI base.
- **Must be built (none found on develop):** X Wallet billing card; "Manage X Wallet" modal (Tab A top-up + auto-recharge / Tab B usage); top-up calculator; wallet **balance** store/getter; **deduct/consume** endpoint; a per-sync cost preview + gate.

### 2.6 Billing page & permission gating

- **Billing UI lives in the SETTING module:** `src/modules/setting/components/billing/` — `EnrolledPlanView.vue` (main enrolled-plan screen), `LimitCard.vue`, `sections/`, `reusable/`, `dialogs/`. `modules/billing/` holds plan/subscription + addon machinery (`AdjustLimitsModal.vue`, `LimitItem.vue`, `PaddleCheckoutModal.vue`; composables `useBilling.ts`, `usePaddle.ts`, `useAdjustLimits.ts`, `useFeatures.ts`, `useAddonHighlight.ts`).
- **Billing-capable detection (standardized):** `hasPermission('can_see_subscription')`; super-admin via `role === 'super_admin'` / `isActiveUserSuperAdmin()`. Canonical gate (`TopHeaderBar.vue`): `if (!hasBillingAccess && !isSuperAdmin) return`. `FeatureAddOnModal.vue` shows the "contact your Super Admin" fallback when billing access is missing.
- **Where the "enable X analytics" toggle lives:** the X Wallet card in `src/modules/setting/components/billing/`, reusing the addon-row pattern, gated by `can_see_subscription` / super_admin. (Flipping a billing gate is "stop and ask" per repo rules.)

### 2.7 Feature-unlock / locked-state precedents (reusable)

- **Closest precedent — in-page locked/upsell:** `src/modules/analytics/views/competitor/common/CompetitorAnalyticsLanding.vue` + `CompetitorUpgradeModal.vue` + `CompetitorDummyGraphs.vue` (blurred demo dashboard behind an unlock CTA when the plan lacks access). `AnalyticsMain.vue` also carries upgrade/locked logic.
- **Add-on / purchase modals:** `src/components/common/FeatureAddOnModal.vue` (generic add-on + confirmation + billing-access fallback), `InboxAddOnModal.vue`, `TwitterPostingAddon.vue`.
- **Highlight helper:** `src/modules/billing/composables/useAddonHighlight.ts` (scroll-to + highlight an addon row from a "limit reached" CTA).
- **Banners:** `StickyBanner.vue`, composer `AiImageCreditsBanner.vue` / `AiCaptionCreditsBanner.vue`, `TwitterPostUsageAlert.vue`.

### 2.8 Reusable components & i18n

- **`@contentstudio/ui`:** `CstuModal`/`Modal`, `Button`, `Icon`, `Dialog` (modal via `inject('root')` → `$cstuModal`).
- **`@ui` (`src/components/UI`):** `CstButton`, `CstDropdown`/`CstDropdownItem`, `CstInputFields`.
- **Analytics-shared:** `AnalyticsCardWrapper.vue` (card + skeleton), `MiniStatsCard.vue`, `NewStatsCard.vue`, `AnalyticsPostsTable.vue`, `AnalyticsDropdown(Item).vue`.
- **Billing-shared:** `LimitItem.vue`, `AdjustLimitsModal.vue`.
- **i18n:** 8 locales `src/locales/{en,de,el,es,fr,it,pl,zh}/`; new keys must land in **all 8 in one commit**. Relevant existing namespaces: `analytics.json` (`analytics.twitter.job_settings_modal.*`, `analytics.common.sync_date_range_modal.*`, `analytics.common.manual_sync.*`, credits-used line), `settings.json` (`settings.billing.*`), `composer.json` (`composer.twitter_usage_alert.*`), `common.json` (`common.twitter_posting_addon.*`).

### 2.9 Integration points (where the feature plugs in)

- **Per-sync cost preview + balance gate:** `AnalyticsTabsHeader.vue` `handleSyncData` and `SyncDateRangeModal.vue` (twitter mode). Cost = pure function of `nTweets` (reuse `[10..150]`). Also `TwitterJobSettingsModal.vue` for recurring cost.
- **Wallet balance/consume data:** extend the existing `credits_used` per-account return; per repo LUM-STATE rules, wallet **balance = TanStack Query** (new `src/api/*` fn + query keys), **UI-only state = Pinia** (never store balance in Pinia).
- **X Wallet billing card + toggle:** `src/modules/setting/components/billing/`, reusing billing addon/Paddle plumbing, gated by `can_see_subscription` / super_admin.
- **Locked/consent state:** mirror `CompetitorAnalyticsLanding.vue` / `CompetitorUpgradeModal.vue` for the in-page unlock/consent gate.

### 2.10 Technical considerations / gaps

- **Two credit concepts risk confusion:** (a) **X posting credits** (composer, monthly quota + $5/300 addon, 1/15 per tweet) and (b) **X analytics credits** (per-sync `credits_used`). The new prepaid **dollar** wallet should be the single balance both draw from; name analytics consumption to avoid colliding with `x_posting_credits`.
- **Dollars vs credits:** the prepaid wallet PRD is **dollar-denominated** ($0.018/post etc.); the existing analytics readout is a **count**. Recommend expressing analytics cost in **dollars** for consistency with the wallet (decision to confirm in workflow).
- **Date range is irrelevant to X cost** — keep the X sync modal count-only.
- **No wallet balance store or deduct/consume endpoint on develop** — both must be added (BE). Backend foundation is the unmerged `cont-2623` branch.
- **Reuse the standardized permission gate** (`can_see_subscription` + super_admin) — do not invent a new flag.
- **Charts** use `AnalyticsCardWrapper` + `useEcharts` (never import ECharts directly).

---

## 3. Dependencies & sequencing

- **Hard dependency:** the **X Pay-Per-Use Credit Wallet** (`docs/features/x-pay-per-use-credits/`) must land first — this epic charges against that wallet's balance, ledger, pricing config, top-up modal, auto-recharge, and permission model. That PRD already lists "metering X … analytics …" as out-of-scope-but-architected-for; this epic is that extension.
- **Reuse, don't rebuild:** X Wallet card, Manage X Wallet modal, top-up calculator, auto-recharge, usage ledger, pricing config, `can_see_subscription` gating, Usermaven event conventions.
- **New in this epic:** in-page unlock/consent card for X analytics, per-sync live cost preview (manual + scheduled), balance gate + insufficient-balance state, analytics consumption type in the ledger, the billing-card toggle to enable X analytics metering, deduct-on-successful-fetch hook.

---

## 4. Open questions for the review gate

1. **Cost unit:** express per-sync cost in **dollars** (consistent with the wallet PRD) rather than an abstract "credits" count? (Recommended: dollars. The existing `credits_used` count can be shown as the driver, e.g. "N tweets × rate".)
2. **Scheduled syncs:** scheduled jobs (daily/weekly/monthly) consume automatically. Do we (a) show projected recurring cost + require the wallet to cover it, and (b) auto-pause scheduled X syncs when the balance/auto-recharge can't cover them (mirroring the publish "fails, not held" rule)? (Recommended: yes to both, with clear notification.)
3. **Unlock granularity:** is the in-page unlock/consent **per workspace** (each workspace unlocks X analytics once) or **per account** (per connected X profile)? The wallet itself is account-level. (Recommended: consent is per user/workspace surface; spend is from the one account-level wallet.)
4. **Threshold confirm dialog:** adopt a "confirm only above $X" speed bump for large syncs, or keep every sync one-click with the estimate shown? (Recommended: estimate always shown; confirm only above a threshold.)
5. **Free "fresh data" reuse:** if a sync ran very recently, offer to reuse cached data at no charge instead of re-syncing? (Nice-to-have; candidate for v2.)
6. **Markup on reads:** confirm the same config-driven markup applies to analytics reads as to posting (per the wallet's pricing config). Never surfaced to users (BR-12).

---

## 5. Sources

- **X API pricing:** twitterapi.io, Blotato, api.sorsa.io, Social Media Today.
- **Competitors:** Buffer, Metricool (howsociable, schedpilot), Agorapulse (socialpilot), Publer, Sprout, Hootsuite, Later/Loomly alternatives, Brandwatch.
- **UX patterns:** ShapeofAI (Cost estimates), Jasper credits, OpenAI prepaid billing + spend limit, Zapier usage analytics, Dodo/Kinde billing-UX (caps/alerts), Cursor pre-execution confirm, Freemius feature locking.
- **Codebase:** `contentstudio-frontend/src/modules/analytics/**`, `…/setting/components/billing/**`, `…/composer/**/xCredits.ts`; `contentstudio-social-analytics-go/src/{clients/social/twitter.go, services/twitter/twitter-fetcher/run.go, cmd/jobs/fetcher/twitter.go}`; `docs/features/x-pay-per-use-credits/**`.
