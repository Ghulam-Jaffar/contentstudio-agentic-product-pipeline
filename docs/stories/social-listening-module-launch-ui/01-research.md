# Research: Social Listening Module Launch UI

## Current State

The Social Listening module exists in the codebase (`src/modules/listening/`) with a working feed, upgrade modal, billing logic, routes, and i18n. The nav item is already wired up in `useHeaderNavigation.ts` and `TopHeaderBar.vue`. What's **missing** are the five launch-surface changes requested.

### Key files:

| File | Relevant to |
|---|---|
| `contentstudio-frontend/src/components/layout/DesktopNavigationRail.vue` | Renders nav items — needs "New" badge slot per item |
| `contentstudio-frontend/src/components/layout/useHeaderNavigation.ts` | `HeaderNavigationItem` interface + listening item definition |
| `contentstudio-frontend/src/modules/listening/components/ListeningUpgradeModal.vue` | Billing modal — exists, copy needs updating |
| `contentstudio-frontend/src/modules/dashboard/components/DashboardNotificationBanner.vue` | Announcement banner — currently shows Black Friday promo |
| `contentstudio-frontend/src/components/authentication/LoginSideComponent.vue` | Login feature carousel — Social Listening not present |
| `contentstudio-frontend/src/locales/en/listening.json` | Copy for modal, tooltips, feature bullets |
| `contentstudio-frontend/src/locales/en/dashboard.json` | Copy for announcement banner |
| `contentstudio-frontend/src/locales/en/header.json` | Header i18n (has `listening: "Listening"`) |

## What Needs to Change

1. **"New" badge on listening nav icon** — `HeaderNavigationItem` interface lacks a `showNewBadge` field. Need to add it to the interface, set it `true` on the listening item in `useHeaderNavigation.ts`, and render a "New" pill in `DesktopNavigationRail.vue` (and the mobile nav equivalent).

2. **Login page carousel** — `LoginSideComponent.vue` fetches features from `fetchLoginFeaturesURL` API. Social Listening is not in the response. Either the API is updated (BE dep) or a static entry is added. FE needs to accommodate the Social Listening entry once a feature image URL is confirmed.

3. **Billing modal copy** — `ListeningUpgradeModal.vue` exists and works, but the i18n keys in `listening.json` need updating:
   - `unlock_modal.locked.description` — new subtitle copy
   - `feature1–4` — new bullet text

4. **Announcement banner** — `DashboardNotificationBanner.vue` currently shows Black Friday content. Needs a new Social Listening announcement section (with "Try it now" link to the listening route) and new i18n keys in `dashboard.json`.

5. **Hover tooltips** — `useHeaderNavigation.ts` already differentiates the two plan states. Just the i18n key values need updating:
   - `listening.landing.not_supported.heading` (Starter/Standard — no plan support) 
   - `listening.landing.locked.heading` (Pro/Advanced/Agency — supported but add-on not purchased)

## Files Involved

- `contentstudio-frontend/src/components/layout/DesktopNavigationRail.vue`
- `contentstudio-frontend/src/components/layout/useHeaderNavigation.ts`
- `contentstudio-frontend/src/modules/listening/components/ListeningUpgradeModal.vue` (no code change — only i18n)
- `contentstudio-frontend/src/modules/dashboard/components/DashboardNotificationBanner.vue`
- `contentstudio-frontend/src/components/authentication/LoginSideComponent.vue`
- `contentstudio-frontend/src/locales/*/listening.json` (all locales)
- `contentstudio-frontend/src/locales/*/dashboard.json` (all locales)
