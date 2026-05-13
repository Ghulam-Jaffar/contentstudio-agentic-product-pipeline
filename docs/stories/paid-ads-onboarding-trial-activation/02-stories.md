# Stories: Paid-Ads Onboarding Plan Selection & Trial Activation

---

## Story 1: [BE] Store paid-ads signup source and expose trial card-collection state in user and plan APIs

### Description:
As a ContentStudio user who arrived via a paid advertisement, I want the system to recognize that I came from a paid ad campaign so that I am shown the appropriate onboarding experience and my trial card-collection status is tracked — enabling the product to present me with timely activation prompts without asking for my payment details again.

---

### Workflow:

1. A new user visits the ContentStudio signup page with a `signup_source=paid_ads` parameter in the URL (appended by the paid ad campaign link).
2. The user completes the registration form and submits.
3. The system stores the `signup_source` value (`paid_ads`) on the user record at the time of registration.
4. After registration, the profile API and plan API responses include `signup_source: "paid_ads"` so the frontend can apply the paid-ads onboarding flow.
5. During onboarding, the user selects a plan and completes the Paddle checkout (providing card details to start a free trial). Paddle fires a webhook (subscription activation event) to the ContentStudio backend.
6. The backend processes the Paddle webhook and sets a `trial_card_collected: true` flag on the user's subscription or plan record.
7. From this point on, the plan API returns `trial_card_collected: true` so the frontend knows card details have been captured and can show the trial activation banner.
8. If the user later clicks "Activate" and completes a Paddle checkout for a full charge, the Paddle `subscription.activated` or `transaction.completed` webhook fires and the backend updates the subscription status to active/paid — the `trial_card_collected` flag is no longer relevant once the subscription is paid.

---

### Acceptance criteria:

- [ ] A new registration request that includes `signup_source=paid_ads` results in the user's record having `signup_source` set to `"paid_ads"`.
- [ ] A registration request without a `signup_source` parameter does not set `signup_source` on the user record (defaults to `null` or is omitted).
- [ ] The profile API response includes `signup_source` when it has a value.
- [ ] The plan API response includes `signup_source` when it has a value.
- [ ] The plan API response includes a `trial_card_collected` boolean field (defaults to `false`).
- [ ] When Paddle fires a subscription activation or trial-start webhook for a user with `signup_source: "paid_ads"`, the backend sets `trial_card_collected: true` on that user's plan record.
- [ ] After `trial_card_collected` is set to `true`, the plan API returns `trial_card_collected: true` for that user.
- [ ] `signup_source` input is validated to accept only allowed string values (e.g., `paid_ads`); unknown values are rejected or ignored (not stored).
- [ ] No regression on existing registration flows — users without `signup_source` continue to register and receive the standard onboarding.
- [ ] If any confusion arises during implementation, consult the product team or your team lead before proceeding.

---

### Mock-ups:
N/A — backend-only story.

---

### Impact on existing data:
- New `signup_source` field added to the User model (nullable string; existing users have `null`).
- New `trial_card_collected` boolean field added to the plan/subscription record (defaults to `false`).
- No changes to existing subscription flow, Paddle webhooks, or plan limits for users without `signup_source`.

---

### Impact on other products:
- Mobile apps consume the plan API — the two new fields (`signup_source`, `trial_card_collected`) are additive and do not change existing response structure, so mobile apps are unaffected.
- Chrome extension is unaffected (does not handle billing or onboarding).

---

### Dependencies:
None — this story is a prerequisite for **[FE] Paid-ads onboarding plan selection screen and trial activation banner**.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story.
- [ ] Multilingual support — N/A, no user-facing strings in this story.
- [ ] UI theming support — N/A, backend-only story.
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-backend/app/Http/Controllers/Accounts/AccountController.php` — registration handler; `signup_source` should be accepted from the request and stored on the User here.
- `contentstudio-backend/app/Models/Account/User.php` — User model; add `signup_source` to the fillable fields.
- `contentstudio-backend/app/Http/Controllers/Billing/PlanController.php` — `plan()` method returns subscription details; add `trial_card_collected` and `signup_source` to the response here.
- Paddle v1 webhook handler (locate via routes or the `PaddleBillingService`) — the `subscription.activated` or `checkout.completed` event for a paid-ads user should set `trial_card_collected: true` on the subscription/user record.

**Existing patterns:**
- `is_trial` is already returned in the plan API response — `trial_card_collected` follows the same pattern.
- `SubscriptionRepository::getSubscriptionDetails()` is the data access layer used in `PlanController::plan()`.

**Suggested names:**
- Field: `signup_source` (on User), `trial_card_collected` (on subscription record or User)
- Allowed `signup_source` values: `paid_ads` (others can be added later as new ad channels come online)

---
---

## Story 2: [FE] Paid-ads onboarding plan selection screen and trial activation banner

### Description:
As a new ContentStudio user who arrived via a paid advertisement, I want to see a plan selection screen as part of my onboarding — right before I land on the home page — so that I can choose the plan I want and securely provide my card details to start my 7-day free trial. Once my trial has started, I want a clear banner on the home page that reminds me my trial is active and makes it easy to activate (pay for) my subscription at any time.

---

### Workflow:

#### Part 1 — Onboarding Plan Selection (Paid-Ads Users Only)

1. A user who signed up via a paid ad (identified by `signup_source === 'paid_ads'` in the profile API) opens the onboarding modal.
2. The user **does not see** the plan-type selection screen (Full Suite vs. API Plan) that is shown to regular users — they proceed directly into the standard onboarding steps (credentials → business type → social accounts).
3. When the user reaches any of the following exit points that would normally take them to the home page, the plan selection screen is shown **instead** of immediately navigating away:
   - Clicking "Skip & Continue" or "Let's Go" on the Social Accounts step
   - Clicking "Go to ContentStudio" on the video intro screen
   - Clicking "I don't have a brand" (or any action in the onboarding that exits to the home page)
4. The plan selection screen displays the following layout (matching the design screenshots):
   - **Page heading:** "Pick the plan that's right for you — Free for 7 days"
   - **Subtext:** "Choose your plan and add your card to continue. You won't be charged until your trial ends — cancel anytime during the trial period."
   - **Secure badge:** "🔒 Secure checkout powered by Paddle. Cancel anytime before your trial ends."
   - **Billing toggle:** Monthly / Yearly — use the `SegmentedControl` component. The Yearly tab shows a "SAVE UP TO 34%" pill badge (use `Badge` component, green variant).
   - **Plan cards (4 plans):**
     - **Standard** — $29/mo (monthly) | $20/mo (yearly)
       - Tagline: "For creators and small teams getting started."
       - "7 days free trial" (green text)
       - Features: 5 social accounts, 1 workspace, 1 user, 25,000 AI text credits, 25 AI image credits, 100 AI video credits, 100 video clipping credits
       - Also includes: AI Studio (text, image, video), AI brand knowledge & voice, Unlimited social posting, Multi-view content planner, Social media analytics, Media and assets library, X threads & LinkedIn carousels, Auto first comment, MCP integration, API access, Smart scheduling
     - **Advanced** — $69/mo (monthly) | $46/mo (yearly) — **"MOST POPULAR" badge** (primary variant, shown above the card header)
       - Tagline: "For growing brands needing deeper insights."
       - "7 days free trial" (green text)
       - Features: 10 social accounts, 2 workspaces, 2 users, 50,000 AI text credits, 50 AI image credits, 200 AI video credits, 200 video clipping credits
       - "Everything in 'Standard', plus:" Social inbox, Competitor analytics, Campaign and label analytics, Post recycling (evergreen), Bulk scheduling via CSV, RSS autoposting, Team collaboration, Approval workflow, Exports and scheduled reports
     - **Agency Unlimited** — $139/mo (monthly) | $93/mo (yearly)
       - Tagline: "For agencies managing multiple clients at scale."
       - "7 days free trial" (green text)
       - Features: 25 social accounts, Unlimited workspaces, Unlimited users, 125,000 AI text credits, 125 AI image credits, 500 AI video credits, 500 video clipping credits
       - "Everything in 'Advanced', plus:" White label, EasyConnect, Complete client management, Live training, Priority support, Guided onboarding, Dedicated account manager
     - **API Plan** — $19/mo (monthly) — **"API FIRST" badge** (secondary/outline variant, shown above the card header)
       - Tagline: "For developers integrating ContentStudio via API."
       - "7 days free trial" (green text)
       - Features: 10 social accounts, 1 workspace, 1 user, 3,000 API credits, 15 GB media storage
       - Also includes: API access, Media and assets library, Automate via 3rd party tools, MCP integration, Publish to all platforms, Content categories & campaigns, Schedule & queue posts, First comment support, Carousel, threads & reels posting, Scale easily with volume discounts
   - **Per card — two actions:**
     - Primary `Button` (full-width, primary variant): **"Start Free Trial"**
     - Text link below the button: **"Talk to Support"** (styled as a link/ghost — not a full button)
5. When the user clicks **"Start Free Trial"** on any plan card, the Paddle checkout popup appears (using the existing `openCheckout()` from `usePaddle`) pre-configured with that plan's Paddle price ID and a trial period.
6. When the user clicks **"Talk to Support"**, the Beacon/Help Scout chat widget opens from the bottom-right corner. (Use `EventBus.$emit('open-help-widget')` — the same event used in `DashboardNotificationBanner`.)
7. When the Paddle checkout fires `checkout.completed` (card details successfully submitted, trial started):
   - The plan selection screen closes.
   - The user is taken to the home page.
   - A Usermaven event fires: `trial_started` with `{ plan_slug, billing_period: 'monthly' | 'yearly', source: 'paid_ads_onboarding' }`.
8. The user can close/dismiss the plan selection screen (an "X" or "Skip for now" text link) — they land on the home page without starting a trial. In this case no Usermaven event fires and no trial banner is shown.

#### Part 2 — Trial Activation Banner on Home Page

9. When the user reaches the home page and the plan API returns `signup_source === 'paid_ads'` AND `trial_card_collected === true` AND the trial is still active (not expired), a **Trial Activation Banner** is shown at the top of the page (above existing banners, or as a replacement when the "trial expires in X days" banner is not yet active).
10. Banner copy:
    - **Left text:** "🎉 Your 7-day free trial has started!"
    - **CTA text (inline link or button):** "Activate your subscription now →"
    - The banner uses the existing `bg-primary-cs-50` / `text-primary-cs-500` theming (blue band, consistent with the product's primary color). Do not hardcode colors.
    - The banner is dismissible (an "×" button on the right). Once dismissed, it is hidden for the session (can reappear after page refresh until the subscription is activated or the trial expires).
11. When the user clicks **"Activate your subscription now →"**, a confirmation modal appears:
    - **Modal title:** "Activate your subscription"
    - **Description:** "You're about to activate your selected plan. You'll be charged the plan price starting today and your 7-day free trial will end. Are you sure you want to continue?"
    - **Primary CTA:** "Yes, Activate Now" (primary `Button`)
    - **Secondary CTA:** "Not Now" (ghost/outlined `Button` or text link)
    - Use the `Modal` component from `@contentstudio/ui`.
12. When the user clicks "Yes, Activate Now", the Paddle checkout popup opens to process the immediate charge.
13. On successful payment (`checkout.completed`):
    - The banner disappears.
    - A success toast appears: "Your subscription is now active. Welcome aboard! 🎉"
    - A Usermaven event fires: `subscription_activated` with `{ plan_slug, billing_period: 'monthly' | 'yearly', source: 'paid_ads_trial' }`.

#### Part 3 — Activation Flow at Limit-Hit Modals

14. When a paid-ads trial user with `trial_card_collected === true` hits a feature or usage limit (any existing limit modal: workspace limits, storage limits, feature-gated modals), an additional **"Activate to Unlock"** section is shown within the modal, above or alongside the existing upgrade options:
    - **Heading:** "You're on a free trial"
    - **Subtext:** "Activate your subscription now to remove this limit and unlock full access."
    - **Button:** "Activate Subscription" (primary `Button`) — triggers the same confirmation modal from step 11.
15. If the user dismisses the confirmation modal, the limit modal remains open with its existing content.

---

### Acceptance criteria:

**Onboarding — routing:**
- [ ] A user with `signup_source === 'paid_ads'` in the profile does not see the Full Suite / API Plan type-selection screen as their first onboarding step.
- [ ] A user without `signup_source === 'paid_ads'` continues to see the existing first onboarding step unchanged.

**Plan selection screen — layout:**
- [ ] The plan selection screen shows 4 plan cards: Standard, Advanced, Agency Unlimited, API Plan.
- [ ] Each card shows the plan name, tagline, price (monthly or yearly based on toggle), "7 days free trial" label in green, feature list, "Start Free Trial" primary button, and "Talk to Support" text link.
- [ ] Advanced plan card displays a "MOST POPULAR" badge above its header.
- [ ] API Plan card displays an "API FIRST" badge above its header.
- [ ] The Monthly/Yearly toggle (`SegmentedControl`) updates all plan card prices when switched.
- [ ] The Yearly toggle option shows a "SAVE UP TO 34%" `Badge` pill.
- [ ] "Start Free Trial" on each card is a primary `Button` (full-width within its card).
- [ ] "Talk to Support" is a text link (not a full button).

**Plan selection screen — behavior:**
- [ ] Clicking "Start Free Trial" on any plan card opens the Paddle checkout popup for that plan's price ID with a 7-day trial applied.
- [ ] Clicking "Talk to Support" opens the Beacon/Help Scout chat widget from the bottom right.
- [ ] On Paddle `checkout.completed`, the plan selection screen closes and the user is taken to the home page.
- [ ] On Paddle `checkout.completed`, a `trial_started` Usermaven event fires with `{ plan_slug, billing_period, source: 'paid_ads_onboarding' }`.
- [ ] The user can skip/close the plan selection screen and reach the home page without completing checkout; no Usermaven event fires in this case.
- [ ] The plan selection screen is shown for ALL onboarding exit paths (social connect, video intro "Go to ContentStudio", brand generation exit) before the user reaches the home page.

**Trial banner:**
- [ ] The trial activation banner is shown on the home page when `signup_source === 'paid_ads'` AND `trial_card_collected === true` AND the trial has not expired.
- [ ] The banner displays: "Your 7-day free trial has started! Activate your subscription now →"
- [ ] The banner is dismissible; clicking "×" hides it for the session.
- [ ] The banner does not appear if the user has already activated their subscription (paid plan).
- [ ] Banner uses `text-primary-cs-500` / `bg-primary-cs-50` theming — no hardcoded colors.

**Activation modal:**
- [ ] Clicking "Activate your subscription now →" opens the activation confirmation modal.
- [ ] Modal title: "Activate your subscription".
- [ ] Modal description text matches the copy in the Workflow section above.
- [ ] "Yes, Activate Now" button opens the Paddle checkout.
- [ ] "Not Now" / "Cancel" closes the modal without action.
- [ ] On successful Paddle payment, the banner disappears and a success toast appears: "Your subscription is now active. Welcome aboard! 🎉"
- [ ] On successful payment, a `subscription_activated` Usermaven event fires with `{ plan_slug, billing_period, source: 'paid_ads_trial' }`.

**Limit-hit flows:**
- [ ] Paid-ads trial users (`trial_card_collected === true`) see an "Activate to Unlock" section in the workspace limits modal, storage limits modal, and feature-gated addon modal.
- [ ] "Activate Subscription" button in limit modals triggers the same activation confirmation modal (step 11).
- [ ] Non-paid-ads users and non-trial users do not see the "Activate to Unlock" section in limit modals.

**Analytics:**
- [ ] When the user completes Paddle checkout on the plan selection screen, a `trial_started` Usermaven event fires with `{ plan_slug: string, billing_period: 'monthly' | 'yearly', source: 'paid_ads_onboarding' }`.
- [ ] When the user activates their subscription (immediate charge), a `subscription_activated` Usermaven event fires with `{ plan_slug: string, billing_period: 'monthly' | 'yearly', source: 'paid_ads_trial' }`.

**General:**
- [ ] The plan selection screen and trial banner are not shown to white-label domain users (`isWhiteLabelDomain === true`).
- [ ] No regression on existing onboarding flow for non-paid-ads users.
- [ ] If any confusion arises during implementation, consult the product team or your team lead before proceeding.

---

### Mock-ups:
Screenshots provided by product team show:
- 4-column plan card grid with Monthly/Yearly toggle and "SAVE UP TO 34%" pill
- Paddle checkout popup rendered inline over the plan selection screen
- Trial confirmation modal (order summary + payment form)

*(Confirm with product/design for final asset files before implementation)*

---

### Impact on existing data:
- No schema changes on the frontend. Reads two new fields from the plan API: `signup_source` and `trial_card_collected`.
- The existing `trial_overs_in` banner (`StickyBanner.vue`) remains unchanged and will still appear once the trial countdown begins.

---

### Impact on other products:
- Mobile apps: no impact. Paid-ads onboarding is web-only (the plan selection screen and trial activation banner are web features only).
- Chrome extension: no impact.
- White-label workspaces: this entire flow must be hidden for white-label domain users.

---

### Dependencies:
- Depends on: **[BE] Store paid-ads signup source and expose trial card-collection state in user and plan APIs** — the `signup_source` and `trial_card_collected` fields must be available in the plan API before FE can render the correct flow.
- API Plan Paddle v1 price IDs must be supplied by the billing/BE team and added to `usePaddle.js` before the API Plan card can trigger checkout.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (plan selection screen must be responsive — cards stack vertically on small screens)
- [ ] Multilingual support (all UI strings must use `$t()` keys; keys added to all locale directories under `src/locales/`)
- [ ] UI theming support (default + white-label — use `text-primary-cs-500`, `bg-primary-cs-50`; never hardcode colors; entire flow hidden for white-label domains)
- [ ] White-label domains impact review — flow must be skipped entirely for white-label users
- [ ] Cross-product impact assessed (web only; mobile and Chrome extension unaffected)

---

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-frontend/src/modules/account/views/onboarding/GetStartedNewModal.vue` — `handleButtonClick()` is the main exit gate for social connect; all three exit paths converge here or in adjacent logic. The paid-ads plan screen should be inserted at this gate.
- `contentstudio-frontend/src/modules/account/composables/useUserOnboarding.ts` — `useSteps()` manages step progression; the paid-ads routing logic (skip plan-type step) likely belongs here or in `GetStartedNewModal`.
- `contentstudio-frontend/src/modules/billing/composables/usePaddle.js` — `openCheckout(payload)` handles Paddle v1 checkout. `handleCheckoutEvent` → `checkout.completed` is where post-checkout navigation should be triggered.
- `contentstudio-frontend/src/components/common/StickyBanner.vue` — existing trial banner structure to reference when building the new "trial started" banner.
- `contentstudio-frontend/src/modules/dashboard/components/DashboardNotificationBanner.vue` — shows how `EventBus.$emit('open-help-widget')` is called for support chat.

**Existing patterns:**
- `isWhiteLabelDomain` from `useWhiteLabelApplication()` is already used in `GetStartedNewModal` to conditionally hide elements — use the same pattern to gate the paid-ads flow.
- `userMaven.track(...)` from `@src/composables/useUserMaven` for analytics events — see existing `onboarding_completed` call in `GetStartedNewModal` for reference.
- `EventBus.$emit('open-help-widget')` is already wired in `DashboardNotificationBanner.vue` to open the support chat widget.

**Suggested names:**
- New component: `PaidAdsPlanSelection.vue` in `contentstudio-frontend/src/modules/account/views/onboarding/`
- New composable (if needed): `usePaidAdsOnboarding.ts` in `contentstudio-frontend/src/modules/account/composables/`
- New trial banner component: `TrialActivationBanner.vue` in `contentstudio-frontend/src/modules/dashboard/components/` or `src/components/common/`
- i18n namespace suggestion: `onboarding.paid_ads.*` for plan selection strings; `billing.trial_activation.*` for banner and modal strings

**Gotcha — API Plan price IDs:**
- The existing `usePaddle.js` `priceIds` object has entries for `standard`, `advanced`, and `agency-unlimited` across all environments — but **no `api-plan` entry**. The API Plan card will not be able to trigger checkout until these IDs are added. Coordinate with the BE/billing team to obtain the Paddle v1 price IDs for all environments (develop, staging, uat, production) before building the API Plan card.

**Gotcha — checkout.completed page reload:**
- The current `handleCheckoutEvent` in `usePaddle.js` calls `window.location.reload()` after a 3-second delay on `checkout.completed`. For the onboarding plan selection flow, this behavior needs to be conditional — the paid-ads onboarding screen should instead navigate to the home page without a full reload (or the reload should be suppressed for this context). Engineering should evaluate whether to extend `openCheckout` with a callback or context flag.
