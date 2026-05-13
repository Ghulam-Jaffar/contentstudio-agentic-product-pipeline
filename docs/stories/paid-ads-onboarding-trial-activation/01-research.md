# Research: Paid-Ads Onboarding Plan Selection & Trial Activation

## Current State

### Onboarding Flow
- **Main shell:** `contentstudio-frontend/src/modules/account/views/onboarding/GetStartedNewModal.vue`
- Onboarding steps tracked via `signup_on_boarding.steps` on the user profile: `{ credentials, brand_website, business_type, social_connect }`
- Users land on `get-started-modal` after sign-up; the first step currently shown is the full plan-type selection (Full Suite vs API Plan)
- Exit points that take a user to the home page:
  1. "Skip & Continue" / "Let's Go" after the social connect step → `handleButtonClick()` → `$cstuModal.hide('get-started-modal')`
  2. VideoIntro "Go to ContentStudio" CTA (after `watch_video` step is marked complete)
  3. "I don't have a brand" — triggers brand generation and hides the modal; handled via `BrandGenerationWidget.vue` + `useOnboardingBrandGeneration` composable
- **No current mechanism to distinguish paid-ads signup users** from organic users. The backend User model has no `signup_source` field in any frontend-visible API.

### Paddle Billing
- **Composable:** `contentstudio-frontend/src/modules/billing/composables/usePaddle.js`
- Supports Paddle Classic + Paddle v1 (new billing)
- `openCheckout(payload)` → shows `paddle-checkout-modal` inline
- `handleCheckoutEvent()` handles `checkout.completed`, `checkout.loaded`, `checkout.error`
- On `checkout.completed`: fires `gtag`, Facebook Pixel, and Usermaven `plan_upgraded` event; then reloads the page
- **Price IDs for new billing plans** (production):
  - standard monthly: `pri_01jcz75fa5t730rnje8apb13f7`
  - advanced monthly: `pri_01jcz7c0fm1eka9gmn3pp607m2`
  - agency-unlimited monthly: `pri_01jcz7egvpr9gjn6016nad3zz1`
  - (annual variants also present)
  - **API Plan price ID: NOT present in current `usePaddle.js`** — needs to be added (coordinate with BE)

### Plans & Pricing
- Current new-billing plans (shown in screenshot): Standard $29, Advanced $69, Agency Unlimited $139, API Plan $19
- These map to Paddle v1 price IDs in `usePaddle.js` (API Plan is missing — gap to address)
- Existing plan display components: `SubscriptionPlansMain.vue`, `SubscriptionPlanCard.vue`, `plansDetails` constant, `pricing.js` (legacy Paddle Classic IDs)

### Existing Banners
- **`StickyBanner.vue`** — shows trial countdown ("Your trial expires in X days") driven by `profile.trial_overs_in`; triggers `showUpgradeModal()` from `useBilling`
- **`DashboardNotificationBanner.vue`** — seasonal promo banner (Black Friday / API launch); uses `isTrialPlan()` from `usePermissions`
- **No existing "trial just started" banner** for paid-ads users

### Limit-Hit Modals
- `WorkspaceLimitsDialog.vue` — shown when workspace limits are hit
- `MediaStorageLimitsExceededModal.vue` — for storage limits
- `FeatureAddOnModal.vue` — for feature-gated upsell
- These currently point to the standard upgrade flow (`showUpgradeModal()` from `useBilling`)

### Backend
- `app/Http/Controllers/Billing/PlanController.php` — returns subscription details including `is_trial`
- `app/Http/Controllers/Accounts/AccountController.php` — handles registration
- `app/Models/Account/User.php` — user model
- Paddle webhook handling exists but exact file for v1 webhook not located in this research scope
- No `signup_source` or `is_from_paid_ads` field found in the profile or plan API

## What Needs to Change

### Backend
- Store `signup_source` (value: `paid_ads`) on the User model at registration when the parameter is present in the signup request
- Expose `signup_source` in the profile/plan API response
- Store `trial_card_collected: true` on the User (or subscription) when Paddle v1 `subscription.activated` / `checkout.completed` webhook fires for a paid-ads trial user
- Expose `trial_card_collected` flag in the plan API response

### Frontend
1. Read `signup_source === 'paid_ads'` from profile API; use it to route the onboarding variation
2. Skip the plan-type selection step (Full Suite / API) at the start of onboarding for paid-ads users
3. Add a new `PaidAdsPlanSelection` onboarding step/screen that intercepts all exit paths before the user reaches the home page
4. Render the 4 plans (Standard, Advanced, Agency Unlimited, API Plan) with Monthly/Yearly toggle and Paddle checkout
5. "Talk to Support" opens the Beacon/Help Scout chat widget via `EventBus.$emit('open-help-widget')`
6. After Paddle `checkout.completed` (card collected) → navigate to home page
7. On home page: show a new "Trial Started" banner when `signup_source === 'paid_ads'` AND `trial_card_collected === true` AND trial is still active
8. Banner CTA "Activate" → confirmation modal → Paddle checkout (charge immediately)
9. Limit-hit modals: add Activate/Buy option alongside existing upgrade flow when user is a paid-ads trial user with card on file

## UX Reference
The screenshots show the exact plan layout: 4-column card grid, "MOST POPULAR" badge on Advanced, "API FIRST" badge on API Plan, Monthly/Yearly toggle with "SAVE UP TO 34%" pill, "Start Free Trial" primary button + "Talk to Support" text link per card. The Paddle popup appears inline over the pricing page.

## Files Involved

**Frontend (to create/modify):**
- `contentstudio-frontend/src/modules/account/views/onboarding/GetStartedNewModal.vue` — intercept exit points
- `contentstudio-frontend/src/modules/account/composables/useUserOnboarding.ts` — onboarding step logic
- `contentstudio-frontend/src/modules/account/views/onboarding/PaidAdsPlanSelection.vue` — new component
- `contentstudio-frontend/src/modules/billing/composables/usePaddle.js` — add API Plan price IDs
- `contentstudio-frontend/src/components/common/StickyBanner.vue` — extend or replace for trial-started state
- `contentstudio-frontend/src/modules/dashboard/components/DashboardNotificationBanner.vue` — or new trial banner component
- `contentstudio-frontend/src/modules/common/components/subscription-limits-exceeded/*.vue` — add activation flow to limit modals

**Backend (to create/modify):**
- `contentstudio-backend/app/Http/Controllers/Accounts/AccountController.php` — accept + store `signup_source`
- `contentstudio-backend/app/Models/Account/User.php` — add `signup_source` field
- `contentstudio-backend/app/Http/Controllers/Billing/PlanController.php` — expose new fields
- Paddle v1 webhook handler — update to set `trial_card_collected` flag
