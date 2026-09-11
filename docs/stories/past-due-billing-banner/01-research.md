# Research: Past-due billing banner — copy rewrite + Retry billing / Update credit card CTAs

**Date:** 2026-09-10
**Scope:** Legacy ("old billing" / Paddle Classic) users whose subscription is `past_due`.

---

## Current State

### The banner

`contentstudio-frontend/src/modules/common/components/header-notifications/HeaderBillingNotifications.vue`

Rendered from `Home.vue` when `getWorkspaceSuperAdminDetails().state === 'past_due'`
(`contentstudio-frontend/src/Home.vue:352-355`), gated by the dismiss flag
`userDisplayStore.getBillingActionRequiredStatus`.

It has **two states**, both prefixed with a bold `Action Required:` label:

| State | Condition | Copy key | CTA |
|---|---|---|---|
| Default | `profileStore.getPaymentFailedDetails` is empty | `header.billing_notifications.default_owner_message` | **View Billing Details** → `router.push({ name: 'plan' })` |
| Detailed | one or more `subscription_payment_failed` records exist | `retry_before_message` (with attempt count + next retry date) or `cancellation_message` | **Update Details** → opens the Paddle `update_url` checkout via `window.open(url, 'self')` |

Non-owners get a read-only variant (`non_owner_message`) with no CTA.

Current default copy (`contentstudio-frontend/src/locales/en/header.json:152-159`):

```
"action_required": "Action Required",
"default_owner_message": "We've been trying to charge your card several times with no success. Please update your credit card to proceed with the charge.",
"non_owner_message": "You have outstanding dues, please ask your account owner to clear the pending dues.",
"retry_before_message": "We have tried charging your card {count} times now with no success. Please update your card details before {date} to avoid account suspension.",
"cancellation_message": "We have tried charging your card {count} times now with no success, account has been set for Cancellation.",
"view_billing_details": "View Billing Details",
"update_details": "Update Details"
```

Markup notes: the component is legacy — hardcoded inline styles (`background: #ff9300`),
legacy classes (`notificationCarousel pink`, `btn btn_white large_btn`), no
`@contentstudio/ui` components, no white-label-safe colors. Locale keys exist in all 8
directories under `src/locales/` (en, de, el, es, fr, it, pl, zh).

### The upgrade path that creates parallel subscriptions

`contentstudio-frontend/src/modules/billing/composables/useBilling.ts:597-627` —
`handlePlanChange()`:

```
const hasActiveSubscription =
  getSubscription()?.paddle_id && profileStore.getProfile?.state === 'active'

if (hasActiveSubscription) {
  return handleSubscriptionUpgrade(plan, isAnnually)   // modifies the existing subscription
} else {
  const payload = await getPaddleCheckoutPayload(...)
  return openCheckout(payload, 'v1')                    // NEW checkout → NEW subscription
}
```

**This is the root cause of the double-subscription mess.** A `past_due` user fails the
`state === 'active'` test, so any plan change from the upgrade modal drops into the `else`
branch and opens a **fresh Paddle checkout**. The original past-due subscription is never
cancelled, so the user ends up paying on two parallel subscriptions.

Entry points into that modal for a past-due user:
- Billing page → `PlanDetailsCard` `@upgrade="showUpgradeModal"`
  (`contentstudio-frontend/src/modules/setting/components/billing/EnrolledPlanView.vue:98`)
- `EnrolledPlanView.vue:45-56` also renders a separate "Billing Problem" card for
  `due_date` / `overdue` states with an `upgrade_link` anchor whose `href` is empty (`href=""`) — dead link, worth cleaning up while we are here.
- `showUpgradeModal()` (`useBilling.ts:439-473`) routes legacy users to the
  `upgrade-plan-dialog` modal and new-billing users to `subscription-plans-modal`.

### Old billing vs new billing

`paddle_billing: true` on the subscription marks a Paddle Billing (new) customer; legacy
Paddle Classic users have it unset/false. The flag already drives branching in
`useBilling.ts`, `usePaddle.ts` (`'v1'` vs `'classic'` checkout), `FeatureAddOnModal.vue`,
and `useEnrolledPlanView.ts:100-110` (`handleUpdateClick` — new billing opens `update_url`
in a tab, classic opens it in the Paddle overlay widget).

The failed-payment records the banner reads come from the classic webhook path:
`contentstudio-backend/app/Http/Controllers/Billing/PaddleController.php:91`
(`subscription_payment_failed`) → stored via
`app/Repository/Billing/Paddle/PaddleTransactionsRepository.php:45` with
`attempt_number`, `next_retry_date`, `update_url`, `cancel_url`.

### Is there anything to build "Retry billing" on today?

No. There is **no retry-payment endpoint** in the backend. The closest existing capability
is `PaddleBillingService::applyOneOffChargeForAddonSubscription()`
(`app/Services/PaddleBillingService.php:1334-1357`), which posts to Paddle Classic's
`/api/2.0/subscription/{id}/charge`. That creates a *one-off charge on an active
subscription* — it is not "pay the outstanding invoice", and Paddle rejects charges against
subscriptions that are not active.

**Paddle's own behavior** (checked against Paddle's dunning docs): Paddle already
auto-retries a failed payment up to seven times over ~30 days, and the documented way to
recover a `past_due` subscription is to **update the payment method**, which triggers an
immediate reattempt. Paddle also explicitly errors on `subscription_update_when_past_due` —
you cannot change a past-due subscription's plan through the API at all, which is why the
current code silently falls through to a new checkout.

➡️ **Open question for the review gate:** what should "Retry billing" actually do for a
legacy Paddle Classic user? Realistic options:
1. **Retry with the card on file** — needs a new BE endpoint; Paddle Classic exposes no
   documented "retry this failed payment now" call, so this may not be buildable for
   classic users without Paddle support confirming an approach.
2. **Retry = re-open the Paddle payment flow** (`update_url`), which reattempts the charge
   as soon as the card is confirmed — buildable today, but nearly identical to
   "Update credit card", so the two CTAs would need distinct copy to stay honest.
3. **Retry only for new-billing users** (Paddle Billing supports paying the outstanding
   transaction) and show one CTA for classic users.

---

## What Needs to Change

1. **Banner copy** — replace the `Action Required:` + long sentence with Paddle's wording:
   - Heading: `Your recent payment failed!`
   - Subtext: `We just tried to charge your credit card, but unfortunately the payment did not go through. To keep your account active, please update your billing information.`
2. **Two CTAs instead of one** — `Retry billing` and `Update credit card`, replacing the
   single `View Billing Details` / `Update Details` button.
3. **Remove the upgrade path from the past-due experience** — the banner must not route the
   user anywhere that opens the upgrade plan modal.
4. **Stop plan changes from creating a second subscription** while the user is `past_due`
   (the `handlePlanChange` fall-through). This is the actual defect behind "it's a mess".
5. Rebuild the banner markup with design-system components and theme-aware colors — it is
   currently hardcoded orange/pink with legacy button classes.
6. Add the new locale keys to all 8 locale directories; retire the keys that stop being
   used.

---

## Mobile Context

Not affected. `contentstudio-flutter/` has no dunning banner — `past_due` only maps to a
workspace lock label (`contentstudio-flutter/lib/features/workspaces/domain/workspace.dart:177`,
copy "Payment due" in `assets/i18n/en.json`) and a login gate
(`lib/features/auth/domain/login_account_gate.dart:185`). Recovering a failed payment is a
web-only flow today, so no `[Flutter]` story.

---

## Files Involved

| File | Change |
|---|---|
| `contentstudio-frontend/src/modules/common/components/header-notifications/HeaderBillingNotifications.vue` | New copy, two CTAs, design-system rebuild |
| `contentstudio-frontend/src/locales/{en,de,el,es,fr,it,pl,zh}/header.json` | New `billing_notifications` keys |
| `contentstudio-frontend/src/modules/billing/composables/useBilling.ts` | Block the past-due → new-checkout fall-through in `handlePlanChange` |
| `contentstudio-frontend/src/modules/setting/components/billing/EnrolledPlanView.vue` | Past-due state on the billing page; dead `upgrade_link` anchor |
| `contentstudio-frontend/src/modules/setting/composables/useEnrolledPlanView.ts` | `handleUpdateClick` reuse for the Update credit card CTA |
| `contentstudio-backend/app/Services/PaddleBillingService.php` + a billing controller | New retry-payment endpoint, if option 1 above is chosen |

---

## Gotchas

- The banner is dismissible (`setBillingActionRequiredStatus(false)`) and the dismissal is
  not persisted per-session in a way that survives navigation resets — worth confirming
  whether a payment-failure banner should be dismissible at all.
- `window.open(url, 'self')` in the current CTA is a typo-ish call (`'self'` is not
  `'_self'`), so it opens a named popup rather than navigating. Fix while touching it.
- `getPaymentFailedDetails` is populated from classic webhook records only; a new-billing
  user in `past_due` may land in the "default" banner state with no attempt count.
- Non-owner copy must stay — only the owner sees actionable CTAs.
