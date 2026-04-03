# White-Label Reseller Promo Banner — Research

## Current State

The white-label settings page lives at **Settings → White Label** (`/settings/white-label/`).

- **Route:** `contentstudio-frontend/src/modules/setting/config/routes/setting.js` (line 265, name: `white-label`)
- **Main component:** `contentstudio-frontend/src/modules/setting/components/white-label/WhiteLabelMain.vue`
- **Layout:** 4-step wizard (General Settings → Theme Settings → Domain Settings → Email Settings) inside `LayoutCard` components
- **Access control:** Requires `can_see_subscription` permission + `white_label_addon` feature access; if not, redirects to home
- **Upgrade modal:** `WhiteLabelUpgradeModal.vue` shown when white-label is not unlocked (overlay with gray backdrop)
- **Completion popup:** `CompletionPopup.vue` shown when setup is complete

The page currently has **no mention of the Reseller ContentStudio product** — it only covers the basic white-label addon (custom domain, theme, email settings).

**Reseller-related code already exists** but only for the invite/onboarding flow:
- `contentstudio-frontend/src/composables/useResellerInvite.js` — reseller customer invite flow
- `contentstudio-frontend/src/modules/account/views/ResellerCustomerInvite.vue` — invite page

## What Needs to Change

- Add a promotional banner/card to the `WhiteLabelMain.vue` page (below the wizard steps or as a standalone card)
- Banner should inform users that a **full reseller portal** (Reseller ContentStudio) is available for complete white-label control
- CTA button should direct users to contact support (open Intercom chat or link to support page)
- The banner should be visible to all white-label addon users (not gated behind reseller status)
- Should be dismissible or subtle enough to not interfere with the white-label setup flow

## Files Involved

- `contentstudio-frontend/src/modules/setting/components/white-label/WhiteLabelMain.vue` — add promo banner
- Localization files under `src/locales/*/settings.json` — add new i18n keys for banner copy
