# Research — Supported-account-type tooltips on the analytics account selector

## Goal

Add an info tooltip on the **right side of the account-selector dropdown** on each
platform's analytics page, stating which account/profile types that platform's
analytics supports (e.g. Facebook → Pages only, Instagram → Business/Creator only).
In scope: **Facebook, Instagram, LinkedIn, Pinterest, TikTok, YouTube, GMB**.
**Out of scope: X (Twitter)** and the cross-platform **Overview** page.

## Current State

- The analytics account selector is one shared component:
  `contentstudio-frontend/src/modules/analytics/views/common/PlatformAccountSelect.vue`.
  It receives a `type` prop (`facebook | instagram | linkedin | tiktok | youtube | pinterest | gmb | twitter`)
  and renders a single-select `Dropdown` for all platforms — except **Pinterest**,
  which renders **two** dropdowns (profile → board).
- Each platform mounts it inside its filter bar via a `#left` slot of
  `AnalyticsFilterBarWrapper`:
  - `views/facebook/components/FilterBar.vue`, and the same shape for
    `instagram`, `linkedin`, `pinterest`, `youtube`, `tiktok`.
  - GMB mounts it directly in `views/gmb/MainComponent.vue`.
- There is **no** account-type support hint anywhere near the selector today. The
  only per-account hint is a reconnect warning icon inside dropdown rows
  (`v-tooltip="…reconnect_tooltip"`).
- An existing per-platform info-popover component already exists —
  `components/common/PlatformTooltip.vue` (info `Icon` + `v-menu` popover with
  numbered lines + a "Learn More" HelpScout link). **But** it lives in the tabs
  header (`views/common/AnalyticsTabsHeader.vue`, `TabsComponent.vue`), and its
  content is only **data caveats** — timezone, 24-hour refresh, API delays,
  "private posts not available". It says nothing about which account *types* are
  supported. Keys live under `analytics.platform_tooltip.<platform>.lineN`.

Key implication: the account-type facts are the true gate that decides which
accounts even appear in the selector, and those rules are enforced in
`components/common/composables/useAnalyticsUtils.ts` (`getPlatformAccounts`):
- **Facebook** — only `type === 'Page'` accounts show (personal profiles + Groups excluded).
- **Instagram** — every IG account is listed; personal-vs-business is only *labeled*
  (`getProfileType` → Personal/Business). The Business/Creator requirement is
  enforced upstream (Meta Graph API + connection flow), not in this selector.
- **LinkedIn** — only accounts with a **numeric** `linkedin_id` show (company/organization
  pages); personal profiles (string ids) are filtered out.
- **Pinterest / TikTok / GMB** — no account-type filter.
- **YouTube** — all channels; a missing analytics OAuth scope flags the row as reconnect-required.

(Cross-checked against the Go analytics pipeline: same picture. The one nuance —
the backend *does* fetch LinkedIn personal Member Profiles too, with limited data;
the frontend selector currently surfaces company pages only. See open question below.)

## What Needs to Change

- Add a small info icon (ℹ) immediately to the **right** of the account-selector
  dropdown, showing the platform's supported-account-type message on hover/focus.
- Because `PlatformAccountSelect.vue` is shared and already knows its `type`, the
  cleanest place to add this is **inside that one component** (to the right of the
  dropdown), so all in-scope platforms get it from a single change — and the icon
  is simply **not rendered when `type === 'twitter'`**.
- For **Pinterest**, place the icon to the right of the *board* dropdown (i.e. at
  the end of the profile → board group).
- Add new i18n copy under `analytics.common.account_support.<platform>` to **all 8
  locale files** (`en fr de it es el zh pl`).

## UX Reference

Matches the existing analytics `PlatformTooltip` pattern (info `Icon` + popover) and
the `v-tooltip` directive already used throughout `PlatformAccountSelect.vue` — so no
new interaction pattern is introduced. Tooltip is **not** a standalone
`@contentstudio/ui` component (per `docs/ui-components.md`); the established approaches
are the `v-tooltip` directive (floating-vue) or the `v-menu` popover used by
`PlatformTooltip.vue`.

## Mobile Context

N/A — analytics dashboards are a web surface. No iOS/Android stories.

## Open Question (carried from prior analysis)

**LinkedIn wording.** The selector currently surfaces **Company Pages only**, so the
draft copy says that. The Go backend, however, also fetches personal Member Profiles
(with limited data). If a quick check in the live app shows personal LinkedIn profiles
*do* appear in the analytics account dropdown, the LinkedIn line should change to note
"Company Pages + personal profiles (limited analytics)". Defaulting to "Company Pages
only" until confirmed.

## Files Involved

- `contentstudio-frontend/src/modules/analytics/views/common/PlatformAccountSelect.vue` — add the info icon/tooltip (shared, single change)
- `contentstudio-frontend/src/modules/analytics/components/common/PlatformTooltip.vue` — reference pattern (info icon + popover)
- `contentstudio-frontend/src/locales/{en,fr,de,it,es,el,zh,pl}/analytics.json` — new `analytics.common.account_support.*` keys
- (context only, not necessarily edited) each platform filter bar:
  `views/{facebook,instagram,linkedin,pinterest,youtube,tiktok}/components/FilterBar.vue`,
  `views/gmb/MainComponent.vue`
