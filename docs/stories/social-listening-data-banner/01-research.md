# Research: Social Listening Data Availability Banner

## Current State

The Social Listening module was designed and fully spec'd in the `/feature` pipeline (see `docs/features/social-listening/`). All stories were pushed to Shortcut (sc-113295 through sc-113367). The module shell and navigation story is `[FE] Listening module shell — navigation, routing, and user state management` (sc-113295).

**Key product constraint from the PRD (docs/features/social-listening/03-prd.md):**
> "Historical data tiers (30d / 90d / 1yr) — V1 ingests forward from activation."

This means Social Listening data is only collected from the day the user activates the feature — there is no historical backfill before that date. Users may not realise this and could be confused when they see limited data shortly after connecting.

**What exists in analytics for reference:**
The analytics module (`contentstudio-frontend/src/modules/analytics/views/`) uses a `CstAlert` component with `type="info"` inside a `v-slot:alert` slot provided by `TabsComponent`. This renders an inline info banner at the top of the analytics view. Example:
- `contentstudio-frontend/src/modules/analytics/views/linkedin_v2/MainComponent.vue` — uses `CstAlert` type="info" to show "We're syncing your latest data."
- `contentstudio-frontend/src/modules/analytics/views/facebook_v2/MainComponent.vue` — same pattern

There is currently **no equivalent banner** in the Social Listening module to communicate data availability scope to the user.

## What Needs to Change

- Add a persistent informational `CstAlert` banner (type="info") at the top of the Social Listening module main view (inside the module shell, sc-113295)
- Banner should appear for all users who have activated Social Listening
- Banner text should clearly state:
  - Data collection starts from the date the user activated Social Listening
  - Mentions and insights from before the activation date are not available
- The banner should be **dismissible** (user can close it) using the `CstAlert` close/dismiss mechanism if available, or remain persistent if not

## UX Reference

Analytics module banner pattern: `<CstAlert type="info" class="text-left mx-5">` — inline, sits at the top of the view before tab content, styled with left-aligned text and horizontal padding. No external UX research needed — the pattern is already established in the product.

## Files Involved

- Social Listening module shell component (created as part of sc-113295 — path to be determined by the frontend team, likely `contentstudio-frontend/src/modules/social-listening/views/ListeningMain.vue` or similar)
- `contentstudio-frontend/src/locales/en/` — new i18n key needed (likely a new `social_listening.json` namespace or added to `common.json`)
- All other locale directories under `contentstudio-frontend/src/locales/` — must mirror the new key
