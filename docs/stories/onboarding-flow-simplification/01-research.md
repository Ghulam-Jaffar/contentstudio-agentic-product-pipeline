# Research: Onboarding Flow Simplification

## Current State

The onboarding flow has two phases:

### Phase 1: Plan Selection (`/select-plan`)
- `contentstudio-frontend/src/modules/account/views/SelectTrialPlan.vue`
- Shows two cards: **API Centric** and **Full Suite (Advanced)**
- No default selection — user must click to select
- After selection, calls `handleTrialPlanChange()` then redirects to home (router guard catches and sends to onboarding)

### Phase 2: Onboarding Steps (`/onboarding/setup/*`)
Current step order defined in `contentstudio-frontend/src/modules/onboarding/config/flow.ts`:

1. **Welcome / Credentials** (`onboardingWelcome`) — Full name, phone number, timezone
2. **Business Type** (`onboardingBusiness`) — Select business type
3. **Social Connect** (`onboardingConnect`) — Connect social media accounts (skippable)
4. **Brand Website** (`onboardingBrand`) — Enter website URL (skipped for API-centric plans)

Flow logic in `useOnboardingFlow.ts` composable handles step navigation, and `computeOnboardingResumeRoute()` determines which step to resume from based on `signup_on_boarding.steps` in the user profile.

### Backend Initial State
`contentstudio-backend/app/Http/Controllers/Authentication/AuthController.php` initializes new users with:
```php
'signup_on_boarding' => [
    'is_completed' => false,
    'steps' => [
        'credentials' => false,
        'business_type' => false,
        'social_connect' => false
    ]
]
```

## What Needs to Change

1. **Default Full Suite selected** — Set `selectedPlan = ref<PlanChoice>('advanced')` instead of `null` in `SelectTrialPlan.vue`
2. **Remove phone number field** — Strip the VueTelInput component, phone-related logic, and SCSS from `OnboardingWelcomePage.vue`
3. **Remove social connect step** — Remove from `ONBOARDING_STEPS` array in `flow.ts`, remove route from `onboardingFlow.ts`, remove from router guard set in `router.js`
4. **Update flow after business type** — After business type: Full Suite users go to Brand Website step, API-centric users complete onboarding and go to home
5. **Update backend initial state** — Remove `social_connect` from initial `signup_on_boarding.steps` in `AuthController.php`

## New Flow (After Changes)

**Full Suite plan:**
Select Plan (Full Suite default) → Welcome (name + timezone) → Business Type → Brand Website → Home

**API-centric plan:**
Select Plan → Welcome (name + timezone) → Business Type → Home

## Files Involved

### Frontend
- `contentstudio-frontend/src/modules/account/views/SelectTrialPlan.vue` — default plan selection
- `contentstudio-frontend/src/modules/onboarding/views/OnboardingWelcomePage.vue` — remove phone field
- `contentstudio-frontend/src/modules/onboarding/config/flow.ts` — remove social_connect step
- `contentstudio-frontend/src/modules/onboarding/config/routes/onboardingFlow.ts` — remove connect route
- `contentstudio-frontend/src/modules/onboarding/composables/useOnboardingFlow.ts` — update API-centric completion logic (currently special-cases social_connect step)
- `contentstudio-frontend/src/router.js` — remove `onboardingConnect` from `ONBOARDING_FLOW_ROUTE_NAMES`
- `contentstudio-frontend/src/modules/account/composables/useUserOnboarding.ts` — remove phone tracking from `personal_details` event, clean up phone-related draft logic
- `contentstudio-frontend/src/modules/account/types/index.ts` — remove `social_connect` from `OnboardingStepKey` type

### Backend
- `contentstudio-backend/app/Http/Controllers/Authentication/AuthController.php` — remove `social_connect` from initial `signup_on_boarding.steps`
