# Research: Customer.io `subscribed_at` person attribute

## Request

Send a Customer.io **person attribute** (not an event) holding the time the account first subscribed to a paid plan. It is set when the super admin or an admin upgrades. After that, every user on that account gets the same value whenever they log in: the super admin, other admins, and any team member whose account has a plan.

## Current State

**Backend: events only, no attributes.**
- `contentstudio-backend/app/Helpers/Integrations/CustomEventHelper.php` → `SendEventToCustomerIo()` POSTs to `{event_api}{userId}/events` (Track API v1, basic auth `site_id:api_key`). It returns early when either credential is blank, because the Track API creates the person as a side effect.
- There is **no** helper for updating a person's attributes. In Track API v1 that is `PUT https://track.customer.io/api/v1/customers/{id}` with the attributes as the JSON body. The config key `integrations.other_integrations.customer_io.event_api` already points at `.../api/v1/customers/`, so the same base URL works.
- `contentstudio-backend/app/Http/Controllers/Billing/Webhooks/PaddleBillingController.php`
  - subscription created handler (~L340-440) sends the `subscription_created` event with email, name, plan_name, plan_id, plan_slug, cc_plan_state, and **no timestamp**
  - subscription updated handler (~L534) sends `subscription_updated`
  - Both send to `$user->_id`. That is the **super admin**, because the subscription belongs to the account owner even when an admin completed the checkout.

**Where a "first subscribed" time already exists:**
- `PaddleBillingSubscription.first_billed_at` (Paddle Billing). It is already mapped to `subscription_created_at` in `app/Helpers/Billing/PlanHelper.php:262`
- `BasePlanSubscription.subscription_created_at`, keyed by `super_admin_id` (`app/Repository/Billing/Subscriptions/BasePlanSubscriptionRepo.php`). It covers legacy Paddle Classic, and the repo already has `->orderBy('subscription_created_at')` queries per super admin
- FastSpring (`app/Models/Billing/FastSpring/FSSubscriptionsModel.php`) and lifetime/redeem codes (`app/Models/Billing/Lifetime/RedeemCodesModel.php`) are older sources
- There is **no** stored `first_subscribed_at` on the user. The earliest subscription record across providers is the source of truth. Storing it once on the account owner is worth considering so the login path doesn't have to query every provider

**Frontend: Customer.io identify exists but does not run on login**
- `contentstudio-frontend/src/modules/common/composables/useIdentify.ts` → `customerIOIdentify()` calls `_cio.identify()` with id, name, email, `created_at` (unix), state, team_role, current_plan, counts, and workspace info. For team members the id is the member's own user id. The plan fields come from the account's plan, which is the super admin's plan.
- It is skipped when the `admin_logged_user` cookie is set (ContentStudio staff logged in as the user)
- `identity()` is only called from SignUp, VerifyEmail and onboarding. It is **not** called on a regular login or app load.
- **Latent bug:** `UpgradePlanConfirmation.vue`, `UpgradePlanComponent.vue` and `useWorkspaceSwitcher.ts` destructure `identify` from `useIdentify()`, but the composable returns `identity`, not `identify`. Their post-upgrade `identify()` calls are therefore `undefined()` and throw. Either the call never worked or it fails inside a caught path. Worth a separate fix, and it's another reason not to rely on the FE identify for this attribute.

**Who can upgrade:** the super admin, and admins with the `can_see_subscription` permission (`TopHeaderBar.vue:614`, `getAdminsWithBillingPermissions` in `FeatureAddOnModal.vue`). Whoever checks out, the subscription lands on the super admin's account.

## Why BE, not FE

- The FE identify doesn't run on login, and the post-upgrade identify is broken. Fixing both only covers web.
- Setting the attribute server-side covers the Paddle webhook, which fires whether or not anyone is in the app, as well as web and Flutter logins, from one place.
- It matches how `subscription_created` and `subscription_updated` are already sent.

## What Needs to Change

- Add a Customer.io "update person attributes" call next to `SendEventToCustomerIo`, with the same blank-credential guard
- Resolve the account's **first** paid subscription time (earliest across Paddle Billing `first_billed_at`, `BasePlanSubscription.subscription_created_at`, and the older providers), optionally persisted on the super admin
- On subscription created: set `subscribed_at` on the super admin. The checkout user is in Paddle `custom_data` if we want to set it on the admin who upgraded too. Otherwise that admin gets it on next login
- On login (email/password, Google/social, SSO/SAML, web and mobile): if the user's account has a paid plan, set `subscribed_at` on that user. Skip it for staff "login as user" sessions
- Value format: **Unix timestamp in seconds**, matching the existing `created_at` attribute (Customer.io treats `*_at` unix values as dates)

## Open Questions for the PO

1. **CC trial accounts** (`cc_plan_state = trialing`): they have a subscription but have not paid yet. Proposed: no attribute until the first charge, so `subscribed_at` means "first paid". Confirm.
2. **Lifetime / AppSumo / redeem-code accounts:** should they count as subscribed, with the redeem time as the value?
3. **Cancel then resubscribe:** proposed to keep the original first time and never overwrite it. Confirm.
4. **"Logs in" vs "opens the app":** sessions are long-lived, so some users rarely log in again. Should it also send on app load or session refresh?

## Files Involved

- `contentstudio-backend/app/Helpers/Integrations/CustomEventHelper.php`
- `contentstudio-backend/app/Http/Controllers/Billing/Webhooks/PaddleBillingController.php`
- `contentstudio-backend/app/Repository/Billing/Subscriptions/BasePlanSubscriptionRepo.php`
- `contentstudio-backend/app/Models/Billing/PaddleBilling/PaddleBillingSubscription.php`
- Backend login handlers: email/password, social, `app/Http/Controllers/SSO/SamlController.php`
- `contentstudio-backend/config/integrations.php` (existing `customer_io` block, no change expected)
