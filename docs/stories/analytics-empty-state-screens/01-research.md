# Research — Analytics empty-state (onboarding) screens redesign

## Goal

Replace the plain "No [platform] Connected" analytics empty states with richer
onboarding screens: a centered white card with a platform logo, title, subtitle, a
row of 4 info cards (or, for Overview, a supported-platforms list + "Coming soon"
banner), a primary Connect CTA, and a "Learn More" link. 10 screens: Overview, Meta
Ads, Google Ads, Facebook, Instagram, LinkedIn, TikTok, YouTube, Pinterest, GBP.

## Current State

- Every platform's analytics page renders its empty state inline in its
  `MainComponent.vue` — a simple block: an image
  (`assets/imgs/no_data_images/no-*-account.png`), a "No X Connected" line, and a
  Connect button. Example (`views/facebook/MainComponent.vue`):
  - copy: `analytics.facebook.main.status.no_account_connected` ("No Facebook Page Connected")
  - CTA: `analytics.facebook.main.actions.connect_account` → opens the shared
    `social-connect-modal` via `$cstuModal.show('social-connect-modal')`.
- The same shape repeats in `views/{instagram,linkedin,tiktok,youtube,pinterest,gmb,overview,meta_ads,google_ads}/MainComponent.vue`.
- There is **no** shared empty-state component today — each page hand-rolls it.
- Existing empty-state copy already lives under each platform's namespace in
  `src/locales/*/analytics.json` (e.g. `pinterest.main_component.empty_states.*`,
  `youtube.main.status.*`, `tiktok.main.status.*`, `common.filter_bar.*`).
- An existing per-platform HelpScout Beacon article map already exists in
  `components/common/PlatformTooltip.vue` (`learnMoreBeacons`) — overview, facebook,
  instagram, linkedin, tiktok, youtube, pinterest, gmb, meta_ads, google_ads — opened
  via `window.Beacon('article', id, { type: 'modal' })`. This is what "Learn More"
  should reuse.
- Meta Ads and Google Ads tabs are feature-flag gated (see `config/routes/analytics.ts`);
  their empty states appear only within those already-gated tabs.

## What Needs to Change

- Build **one shared empty-state component** (suggested
  `views/common/AnalyticsEmptyState.vue`) with two layouts:
  - **Cards layout** (9 platform screens): logo + title + subtitle + 4 cards + CTA + Learn More.
  - **List layout** (Overview only): title + subtitle + supported-platform list +
    "Coming soon" banner + CTA + Learn More (no logo, no eyebrow pill).
- Cards 3 (AI Insights) and 4 (Reports) are identical across all platform screens;
  Card 1 (Connect) and Card 2 (unique) vary per platform. Drive all content from
  per-platform config + i18n.
- Swap each platform `MainComponent.vue`'s inline empty state to render the new
  component in the `!hasAccounts` branch. Keep the CTA wired to the existing connect
  flow (`social-connect-modal`; Meta Ads / Google Ads use their own OAuth connect).
- "Learn More" opens the platform's existing HelpScout Beacon article (reuse the
  `learnMoreBeacons` map).
- Add all new copy to `analytics.json` in **all 8 locales**.
- X (Twitter): **unchanged** — no new screen; per the spec X stays only in Overview's
  "Coming soon" row (Threads, Bluesky, X). The existing Twitter empty state is left as-is.

## Design / Component Notes (CS design system)

- Use `@contentstudio/ui`: `Button` (primary variant for the blue CTA — theme-token
  driven, **not** hardcoded blue, so white-label recolors it), `Badge` (green "Coming
  soon"), `Icon`. Card container + hover via Tailwind utilities + theme tokens
  (`bg-primary-cs-50`, `text-primary-cs-500`, `text-gray-*`).
- **Icon gap to flag:** the unique Card-2 icons (target rings, multicolor pie, colored
  bar chart, hashtag, briefcase, video camera, play triangle, eye, map pin) and the
  brand platform logos are **multi-color/brand SVGs** that likely aren't all in the
  `@contentstudio/ui` `Icon` set. Platform logos already exist as assets
  (`getSocialImageRounded()` / `no_data_images/`); the decorative card icons may need
  new SVG assets (a small [Design] asset drop or inline SVGs). Shared-card icons
  (link/chain, purple sparkle, document) map to standard icons.
- No standalone Tooltip needed here; "Learn More" is a text link + `Icon name="CircleHelp"`.

## Mock-ups

Provided by the user, one per screen, in this story folder's `mockups/`:
`overview.png`, `metaads.png`, `googleads.png`, `fb.png`, `IG.png`, `linkedin.png`,
`tiktok.png`, `YT.png`, `pinterest.png`, `GBP.png`. Verified readable; they match the
written spec. Embedded per-screen in `02-stories.md`.

## Mobile Context

N/A — analytics dashboards are web-only. No iOS/Android stories.

## Decisions (from user)

- **X/Twitter:** follow the spec literally — X stays in Overview "Coming soon", no X
  screen, existing X empty state untouched.
- **Learn More:** reuse the existing per-platform HelpScout help articles.

## Files Involved

- New: `contentstudio-frontend/src/modules/analytics/views/common/AnalyticsEmptyState.vue` (suggested shared component)
- Edit empty-state branch in:
  `views/{overview,meta_ads,google_ads,facebook,instagram,linkedin,tiktok,youtube,pinterest,gmb}/MainComponent.vue`
- `contentstudio-frontend/src/modules/analytics/components/common/PlatformTooltip.vue` — source of the `learnMoreBeacons` article-id map to reuse
- `contentstudio-frontend/src/locales/{en,fr,de,it,es,el,zh,pl}/analytics.json` — new empty-state copy
- Possibly new SVG assets for the unique Card-2 decorative icons
