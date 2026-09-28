# Social Listening 7-Day Trial — Research

**Date:** 2026-09-18
**Scope:** Web only. Social Listening does not exist in `contentstudio-flutter/` (no `lib/features/listening*` module), so there is no `[Flutter]` story.
**Design canvas:** https://claude.ai/artifact/HJd9qq3A6yAZBvCHfpjhV3

This doc stays local. Nothing here goes into the story bodies.

---

## 1. What we are building

Paid-plan workspaces that can have Social Listening but have not bought it currently see a sample-data preview plus an "Unlock Social Listening" modal. We add a second path: a 7-day free trial that auto-charges on day 7.

Decisions taken with the CTO on 2026-09-18:

| Question | Decision |
|---|---|
| One CTA or two? | **Two buttons** in the unlock modal: `Unlock now` (secondary) and `Start 7-day free trial` (primary) |
| Trial confirmation | Yes. Clicking the trial button opens a confirmation dialog stating the end date and the exact amount before anything is scheduled |
| New plans (Paddle Billing) | **One** cycle box only, showing the cycle the subscription is already on and that cycle's price. No second option, no disabled tile |
| Old plans (Paddle Classic) | **Both** cycle boxes, nothing pre-selected, both CTAs disabled until one is picked |
| Day 7 success | No modal, no toast interruption. The trial banner just disappears |
| Day 7 failure | Blocking modal on every Listening visit. Admins get Retry payment + Update card details. Members get the admin list |
| Backdrop | Both modals sit over the existing sample-data Listening page with the background **blurred**, not just dimmed |

---

## 2. Frontend entry points

| Concern | File |
|---|---|
| The unlock modal being changed | `contentstudio-frontend/src/modules/listening/components/ListeningUpgradeModal.vue` |
| Access flags, single source of truth | `contentstudio-frontend/src/modules/listening/composables/useListeningAccess.ts` |
| Pure nav/route gating rules | `contentstudio-frontend/src/modules/listening/utils/access.ts` |
| Tier and price resolution | `contentstudio-frontend/src/modules/listening/composables/useSocialListeningTier.ts` |
| Preview banner the trial banner replaces | `contentstudio-frontend/src/modules/listening/components/preview/ListeningPreviewBanner.vue` |
| Purchase mutation | `usePurchaseSocialListeningAddonMutation` in `contentstudio-frontend/src/modules/setting/queries/usePlanQueries.ts` |
| Copy | `contentstudio-frontend/src/locales/en/listening.json` under `listening.unlock_modal` and `listening.preview.banner` |
| Card-update pattern to copy | `contentstudio-frontend/src/modules/common/components/header-notifications/HeaderBillingNotifications.vue` |
| Past-due status helpers | `contentstudio-frontend/src/modules/billing/utils/subscriptionStatus.ts` |
| Billing add-on table | `contentstudio-frontend/src/modules/setting/components/billing/sections/SubscriptionsTable.vue`, `EnrolledPlanView.vue` |

### Useful details

- `useListeningAccess` derives everything from two plan-payload flags: `social_listening_supported` and `social_listening_lock`. A trial needs a third signal alongside them (status + end date) rather than a new access composable.
- `getListeningNavState` already returns `enabled` / `preview` / `locked_upgrade` / `hidden`. A trial is functionally `enabled`, so the cleanest shape is to keep returning `enabled` and surface the trial separately for the banner.
- `isSampleWorkspace` short-circuits to `enabled` for demo workspaces. A trial must not be startable from a sample workspace.
- The modal's `disabledPlans` computed and `plan_switch_disabled_tooltip` string exist only to grey out the wrong cycle for Paddle Billing users. With the one-box decision, **both become dead code** and should be removed along with the tooltip string.
- The card-update URL comes from `profileStore.getPaymentFailedDetails?.[0]?.update_url` with a fallback to `getBillingDetails.update_url`, each sanitized separately. The failed-trial modal must reuse that exact resolution, including the fallback, or the CTA will silently disappear for users whose failed-transaction list is empty.

---

## 3. Backend entry points

| Concern | File |
|---|---|
| Paddle Billing unlock | `PaddleUserController::unlockSocialListening` (approx. line 1180) in `contentstudio-backend/app/Http/Controllers/Billing/PaddleUserController.php` |
| SKU resolution | `PlanHelper::getSocialListeningAddonPriceId` and `getAllSocialListeningUnlockPriceIds` in `app/Helpers/Billing/PlanHelper.php` |
| Tier rules | `PlanHelper::resolveSocialListeningPlanTier`, `ADVANCED_PLAN_SLUGS` |
| Flag writes | `app/Libraries/Settings/SubscriptionLimits.php` (line ~362), `app/Http/Controllers/Billing/PlanController.php` (line ~198) |
| Webhook handling | `app/Http/Controllers/Billing/Webhooks/PaddleBillingController.php` (`social_listening_addon` at lines ~68, ~785, ~833) |

### Gotchas that shape the implementation

1. **The frontend `billing_cycle` is already ignored for Paddle Billing.** `unlockSocialListening` derives the cycle from `$subscriptionDetails['billing_cycle']['interval']`, not from the request. The one-box decision brings the UI in line with what the backend already does. For Paddle Classic the FE genuinely picks the SKU, which is why two boxes stay there.
2. **Duplicate-unlock guard exists.** `getAllSocialListeningUnlockPriceIds` blocks stacking an Advanced SKU on an Agency SKU. Trial start must run through the same guard, otherwise a trial can be started on a subscription that already holds the add-on.
3. **Current unlock uses `proration_billing_mode: prorated_immediately` and `on_payment_failure: prevent_change`.** A trial cannot use that combination, since it must not bill today. Two workable routes for the dev to evaluate:
   - Attach the add-on item now with `do_not_bill`, and have a scheduled job on day 7 switch it to `prorated_immediately`.
   - Use a Paddle price that carries a 7-day trial period, and let Paddle fire the charge.
   The second keeps the retry and dunning behaviour inside Paddle, which is why the state machine assumes automatic retries. Confirm which one the Paddle account supports before estimating.
4. **Paddle Classic path is a different animal.** The FE opens a Paddle Classic checkout with `social_listening_monthly` / `social_listening_annual` product IDs from `@common/constants/pricing`. Classic trials are configured per product, so a trial likely needs its own Classic product IDs rather than a checkout flag. Confirm with the Paddle account before committing the Classic half of the story.
5. **Promo pricing is live.** `social_listening_promo_enabled` swaps in discounted SKUs. Whatever price the trial confirmation dialog shows must come from the same resolution path, or the quoted amount and the charged amount will diverge during a promo.
6. **Trial eligibility must be stored per subscription, not per workspace.** A workspace-scoped flag lets someone create a second workspace and trial again.

---

## 4. Analytics

Existing events fired by the modal today (see `sendUserMavenEvent` / `sendCustomerIoEvent` in `ListeningUpgradeModal.vue`):

- `social_listening_purchased` with `{ billingCycle, addonName, price }`

New events proposed, staying in the same family:

| Event | Fired by | Payload |
|---|---|---|
| `social_listening_trial_started` | FE, on confirmation dialog confirm | `{ billing_cycle, price, plan_tier }` |
| `social_listening_trial_cancelled` | FE, on cancel-trial confirm | `{ days_remaining, plan_tier }` |
| `social_listening_purchased` | BE, on day-7 success | existing shape plus `source: 'trial'` |
| `social_listening_trial_payment_failed` | BE, on day-7 failure | `{ billing_cycle, price, plan_tier }` |

---

## 5. Open questions for the PO

1. **Data retention after a failed charge.** The design copy says 30 days. Confirm the real number with support and billing.
2. **Retry window.** How long does Paddle keep retrying before the trial is written off? The modal copy says "a few days" deliberately, so the exact figure is not baked into the UI.
3. **Trial length as a config value or a constant?** 7 days is in the button label and several strings. If marketing will ever run a 14-day promo, the length should be a single config value and the copy should interpolate it.
4. **Does the Paddle Classic half ship in the same release?** If Classic trial products are slow to set up, the trial button can ship for Paddle Billing first and Classic users keep seeing only `Unlock now`.
5. **Blurred backdrop support.** The `Modal` component in `@contentstudio/ui` v0.2.21 uses a flat scrim. Confirm whether a blurred-backdrop variant can be added to the library or whether the two Listening modals wrap their own overlay.

---

## 6. Story split note

The guidelines normally split a feature of this size into separate `[BE]` and `[FE]` stories. The CTO asked for a single implementation story plus a single design story for this epic, so the implementation story carries both halves under a `[FE][BE]` prefix. If the team wants to parallelise, the natural cut line is the billing lifecycle (trial start, day-7 charge, failure state) versus the UI (modals, banners, billing row).
