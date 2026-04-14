# Stories: Onboarding Flow Simplification

---

## Story 1: [FE] Simplify onboarding flow: default plan selection, remove phone field, skip social connect step

### Description:
As a new user signing up for ContentStudio, I want a faster onboarding experience so that I can start using the product sooner. This story makes three changes to the onboarding flow:

1. **Plan selection page**: The "Full Suite" plan card is pre-selected by default (user can still switch to API plan)
2. **Personal details step (Welcome)**: The phone number field is removed — only name and timezone remain
3. **Social account connection step**: Removed entirely from the flow — after selecting a business type, Full Suite users go directly to the Brand Website step, and API plan users go directly to the dashboard

**New onboarding flow:**
- Full Suite: Select Plan (pre-selected) → Welcome (name + timezone) → Business Type → Brand Website → Dashboard
- API plan: Select Plan → Welcome (name + timezone) → Business Type → Dashboard

---

### Workflow:
1. User signs up and arrives at the plan selection page (`/select-plan`)
2. User sees the "Full Suite" card is already selected with a highlighted border — user can click "API Plan" to switch if desired
3. User clicks "Start Your Free Trial" to confirm and proceed
4. User arrives at the Welcome step — sees "Full Name" and "Workspace Timezone" fields (no phone number field)
5. User fills in their name and timezone, clicks "Next"
6. User arrives at Business Type step — selects their business type, clicks "Next"
7. **If Full Suite plan:** User arrives at Brand Website step — enters their website URL and completes onboarding
8. **If API plan:** User is taken directly to the dashboard — onboarding is marked as complete

---

### Acceptance criteria:

**Plan selection defaults:**
- [ ] On `/select-plan`, the "Full Suite" card renders with the selected state (highlighted border, primary-colored icon background) on page load
- [ ] The "Start Your Free Trial" button is enabled on page load (not disabled waiting for selection)
- [ ] User can still click "API Plan" card to switch selection before confirming

**Phone number removal:**
- [ ] The phone number field (`VueTelInput` component) no longer appears on the Welcome step
- [ ] The Welcome step shows only: Full Name input, Workspace Timezone dropdown, and Next button
- [ ] The `phone_no` field is sent as an empty string in the profile update (backend still accepts it, no error)

**Social connect step removal:**
- [ ] The social account connection step (`/onboarding/setup/connect`) is no longer part of the onboarding flow
- [ ] Navigating directly to `/onboarding/setup/connect` redirects to the appropriate active step
- [ ] The step counter in the onboarding layout reflects the reduced number of steps (3 for Full Suite, 2 for API plan)
- [ ] After completing Business Type, Full Suite users navigate to Brand Website step (not social connect)
- [ ] After completing Business Type, API plan users are taken directly to the dashboard with onboarding marked complete

**API plan completion handling:**
- [ ] When API plan users complete the Business Type step, the `watch_video` and `accounts_connection_modal_closed` onboarding steps are auto-marked as done (so the getting-started video modal doesn't appear on the dashboard)
- [ ] The `onboarding_completed` analytics event fires when API plan users finish the Business Type step

**Backward compatibility:**
- [ ] Existing users who already completed onboarding (with `social_connect` in their profile) are not affected — `is_completed: true` takes precedence
- [ ] Existing users mid-onboarding who had completed `social_connect` are not regressed — they resume at their next incomplete step

---

### Mock-ups:
N/A — no new UI elements. Changes are removal of existing elements and default state changes.

---

### Impact on existing data:
- No schema changes
- The `phone_no` field on user profiles will be empty for new signups going forward (existing phone numbers remain)
- The `signup_on_boarding.steps` object sent during onboarding updates will no longer include `social_connect`
- The backend still has `social_connect` in the initial `signup_on_boarding` state for new users — the frontend simply ignores it and does not check or submit that step

---

### Impact on other products:
- Mobile apps: Not affected — mobile apps do not use the web onboarding flow
- Chrome extension: Not affected
- White-label: The onboarding flow changes apply to all white-label domains equally — the default plan selection, phone removal, and step removal are global

---

### Dependencies:
None — this is a frontend-only change. The backend's `social_connect` field in the initial state is harmless; the frontend simply stops checking for it.

---

### Files to modify:

| File | Change |
|---|---|
| `contentstudio-frontend/src/modules/account/views/SelectTrialPlan.vue` | Change `selectedPlan` default from `null` to `'advanced'` |
| `contentstudio-frontend/src/modules/onboarding/views/OnboardingWelcomePage.vue` | Remove VueTelInput, phone field template block, phone-related imports/logic, SCSS styles |
| `contentstudio-frontend/src/modules/onboarding/config/flow.ts` | Remove `social_connect` from `ONBOARDING_STEPS`, update `computeOnboardingResumeRoute()` and `buildSignupOnboardingState()` |
| `contentstudio-frontend/src/modules/onboarding/config/routes/onboardingFlow.ts` | Remove the `onboardingConnect` route child |
| `contentstudio-frontend/src/modules/onboarding/composables/useOnboardingFlow.ts` | Move API-centric completion logic from `social_connect` check to `business_type` check |
| `contentstudio-frontend/src/router.js` | Remove `'onboardingConnect'` from `ONBOARDING_FLOW_ROUTE_NAMES` set |
| `contentstudio-frontend/src/modules/account/composables/useUserOnboarding.ts` | Remove `social_connect` tracking case, remove phone from `personal_details` tracking |
| `contentstudio-frontend/src/modules/account/types/index.ts` | Remove `'social_connect'` from `OnboardingStepKey` union type |

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
