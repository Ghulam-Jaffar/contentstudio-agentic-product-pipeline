# Standard Plan: Usage Add-on Top-ups · Research

## Current State

Add-on purchasing is gated by a single plan-document boolean, `features.subscription_addon`. It is `false` on the current Standard plans and `true` from Advanced upward.

The frontend reads it once as `canAccess('subscription_addon')` and uses the result to **disable the entry-point button** — so a Standard user never reaches the add-on modal at all, and is only offered a plan upgrade. There is no per-add-on granularity anywhere in the system.

### Which plans are actually affected

Mapped every plan in `contentstudio-backend/database/seeders/SubscriptionsSeeder.php`:

| Plan | `subscription_addon` |
|---|---|
| `standard-monthly`, `standard-annually` | **false** |
| `standard-monthly-cc`, `standard-annually-cc` | **false** |
| `trial-standard` (and every other trial) | **false** |
| `apple-standard-monthly` | **false** |
| Advanced / Agency Unlimited / Professional / Business & Agency / Enterprise / Max | true |
| Lifetime and AppSumo plans | true |

**Important:** the legacy starter-tier plans that share Standard's limits — `contentstudio-starter-monthly`, `contentstudio-starter-annual`, `empowerers-monthly`, `empowerers-yearly`, `contentstudio-growth-yearly` — are already `true`. They can buy add-ons today. So the cohort this change unblocks is only the four current Standard slugs (plus the Apple IAP variant).

Trials are a separate axis: the frontend disables the button on `isTrialUser` / `isTrialPlan` independently of the flag, and CC trials get an extra override via `PlanHelper::applyCcTrialFeatureOverrides`. **Recommendation: leave trial behaviour exactly as-is** and scope this change to paid Standard subscriptions.

### Where the gate lives

**Backend**
- `contentstudio-backend/database/seeders/SubscriptionsSeeder.php` — `subscription_addon` per plan
- `contentstudio-backend/database/migrations/2026_05_05_120000_create_cc_trial_subscription_plans.php:283` — `standardFeatures()` sets it false for the `-cc` plans
- `contentstudio-backend/app/Http/Controllers/Billing/PlanController.php:214-240` — merges `user->addons` into `subscription.limits` **unconditionally**; no plan check anywhere in the read path
- `contentstudio-backend/app/Http/Controllers/Billing/PaddleUserController.php:1533` — `upgradePaddleBillingSubscriptionLimits`, the Paddle Billing purchase endpoint (`POST /billing/subscription/update_limits`)
- `contentstudio-backend/app/Http/Requests/Billing/UpgradeSubscriptionLimitsRequest.php` — validates `workspace_id`, `addons`, `preview` and checks `billing_access` permission. **No plan-level add-on check.**
- `contentstudio-backend/app/Http/Controllers/Billing/PaddleUserController.php:600` — `upgradeAddonLimits`, the legacy Paddle purchase endpoint (`POST /paddle/increaseLimits`)
- `contentstudio-backend/app/Services/Billing/SubscriptionLimitUpgradeService.php` — builds the Paddle line items from the `addons` payload

**Frontend**
- `contentstudio-frontend/src/modules/billing/constants/billingAddonCatalog.ts` — the 14 catalog entries; each has `addonKey`, `valueKey`, `limitKey`, `group`, `allowedForApiCentric`. There is no usage-vs-scalability concept yet.
- `contentstudio-frontend/src/modules/billing/constants/addonKeys.ts` — `ADDON_KEYS` stable identifiers
- `contentstudio-frontend/src/modules/billing/composables/useBillingAddonPolicy.ts:82` — `isAddonAllowed()`; currently only handles API-centric plans and the social listening double-gate
- `contentstudio-frontend/src/modules/billing/composables/useBilling.ts:772` — `handleIncreaseLimitsClick()`; routes to `adjust-limits-modal` (Paddle Billing) or `increase-limits-dialog` (legacy)
- `contentstudio-frontend/src/modules/billing/components/AdjustLimitsModal.vue:38-68` — renders `filteredLimitItems` as `LimitItem` rows
- `contentstudio-frontend/src/modules/billing/components/LimitItem.vue:457` — the per-row component; props have **no `disabled` or tooltip concept** today
- `contentstudio-frontend/src/modules/setting/components/billing/sections/UsageLimitsCard.vue:117-130` — "Manage add-ons" button, disabled on `!addonAccess.allowed || isTrialUser`
- `contentstudio-frontend/src/modules/setting/components/billing/sections/PlanDetailsCard.vue:173` — `v-if="addonAccess.allowed && hasAnyAllowedAddons"`
- `contentstudio-frontend/src/modules/setting/components/api/ApiModule.vue:136-160` — already has the exact pattern we want: `increaseLimitDisabled` + `increaseLimitTooltip` + an "Upgrade Plan" fallback CTA
- Other `canAccess('subscription_addon')` call sites that gate limit-exceeded upsells: `SubscriptionLimitsExceededModal.vue:172`, `MediaStorageLimitsExceededModal.vue:107`, `LimitExceeds.vue:81` (AI image), `useAiCaptionModal.ts:219`, `AIPostHeader.vue:321`, `ChatInput.vue:894`, `PlanInfoCard.vue:172`, `MessageComposer.vue:489`, `useSaveRssAutomation.ts:38`, `EnrolledPlanView.vue:226`

### Existing copy

`contentstudio-frontend/src/locales/en/settings.json:2143-2144`:
- `manage_addons_disabled`: "Managing add-ons is not available in your current plan. Upgrade your plan to manage your add-on limits"
- `manage_addons_disabled_trial`: "Add-ons are unavailable during your trial. You'll be able to manage them once your trial ends and your subscription is active."

Translations exist in `de`, `el`, `es`, `fr`, `it`, `pl`, `zh`.

---

## What Needs to Change

### 1. Split the add-on catalog into usage vs scalability

| Add-on | Limit | Bucket | Standard |
|---|---|---|---|
| AI text credits | `caption_generation_credit` | Usage | **Allow** |
| AI image credits | `image_generation_credit` | Usage | **Allow** |
| AI video credits | `video_generation_credits` | Usage | **Allow** |
| Video clip credits | `video_clip_credits` | Usage | **Allow** |
| AI auto-reply credits | `ai_auto_reply_credits` | Usage | **Allow** |
| API credits | `api_credit` | Usage | **Allow** |
| X posting credits | `x_posting_credits` | Usage | **Allow** |
| Workspaces | `workspaces` | Scalability | Block, row visible + disabled |
| Social accounts | `profiles` | Scalability | Block, row visible + disabled |
| Team members | `members` | Scalability | Block, row visible + disabled |
| Automation campaigns | `automations` | Scalability | Block, row visible + disabled |
| Media storage | `media_storage` | Scalability | Block, row visible + disabled |
| Listening topics | `listening_topics` | — | Stays hidden (feature locked) |
| Listening mentions | `listening_mentions` | — | Stays hidden (feature locked) |

### 2. Move the gate from the button to the row

- The "Manage add-ons" / "Increase Limit" button must **open** the modal for paid Standard users instead of being disabled.
- Inside the modal, scalability rows render but their steppers/inputs are disabled, with a hover tooltip and a click path into the upgrade flow.
- `PlanDetailsCard.vue`'s `v-if="addonAccess.allowed && ..."` needs the same treatment or the entry point stays hidden.

### 3. Add the server-side check that doesn't exist today

Both purchase endpoints accept an arbitrary `addons` payload with no plan-level validation:
- `POST /billing/subscription/update_limits` (Paddle Billing — the path Standard uses, since all Standard plans are `paddle_billing: true`)
- `POST /paddle/increaseLimits` (legacy Paddle)

A Standard user could craft a request for `socialAccounts` today and it would go through. The new rule needs enforcing here, not only in the UI.

### 4. Leave alone

- **X wallet** — `XWalletCard.vue` is gated on `useXWalletGate()` / `can_see_subscription`, **not** on `addonAccess`, and `WalletService::isEnabledForUser` has no plan-tier check. Standard can already top up X spend ($5–$500). No change needed; worth verifying in QA.
- **Auto-scale** — `useAutoScale.ts:78` bails on `!features?.subscription_addon`. It only covers social account slots (a scalability limit), so it stays off.
- **Social listening** — `useBillingAddonPolicy.ts:47-58` already hides both rows unless `social_listening_supported === true && social_listening_lock === false`. Standard fails that check regardless.
- **Trials** — out of scope, keep current behaviour.

---

## Gotchas

- **`LimitItem.vue` has no disabled state.** Props stop at `showAddAction` / `showRemoveAction`. A `disabled` + tooltip/upgrade-CTA affordance is genuinely new UI on that row, which is why this needs a `[Design]` story.
- **The API-centric allowlist overlaps confusingly.** `API_CENTRIC_ALLOWED_ADDONS` permits `WORKSPACES`, `SOCIAL_ACCOUNTS`, `TEAM_MEMBERS`, `MEDIA_STORAGE` — i.e. the exact set Standard must block. The two policies are orthogonal and both have to hold; `isAddonAllowed()` should AND them, not replace one with the other.
- **`api_credit` has a free-credits quirk.** `PaddleUserController.php:1545` subtracts `user.free_limits.api_credit` from the requested `apiCredits` quantity before pricing. Any new validation must run against the post-adjustment quantity or it will misfire for users with free API credits.
- **X posting credits legacy pricing is odd.** `upgradeAddonLimits` divides the requested quantity by 60 and forces price to 1. Only relevant if a Standard user ends up on the legacy path, which shouldn't happen since Standard is `paddle_billing: true`.
- **Existing analytics cover this.** `addon_purchased` and `addons_limits_updated` already fire from `useAdjustLimits.ts:771-779`. No new event is needed for the purchase itself. If the PO wants to measure the upsell, a new event on the locked-row upgrade click would be the place — flagging for their call rather than assuming it.
- **`subscription_addon` is doing double duty.** Whatever mechanism is chosen, it currently means both "can buy add-ons" and "is a plan that scales" across ~10 components. Changing its meaning in place will have blast radius; adding a second, narrower signal is safer.
- **7 locale directories** need the new copy (`de`, `el`, `es`, `fr`, `it`, `pl`, `zh` alongside `en`).

---

## Mobile Context

Not impacted. `contentstudio-flutter/lib/features/billing/` is IAP + paywall only (`paywall_screen.dart`, `store_iap_repository.dart`, `subscription_controller.dart`) and `lib/features/entitlements/` reads feature access without any add-on concept — grep for `addon` across both returns nothing. There is no add-on purchase surface in the app, so no `[Flutter]` story.

Worth noting for QA: `apple-standard-monthly` is also `subscription_addon: false`, and that plan is purchased through Apple IAP. Add-on purchases for IAP subscribers are a separate billing question — recommend keeping the Apple variant out of scope for this change.

---

## Files Involved

**Backend**
- `contentstudio-backend/database/seeders/SubscriptionsSeeder.php`
- `contentstudio-backend/database/migrations/2026_05_05_120000_create_cc_trial_subscription_plans.php`
- `contentstudio-backend/app/Http/Controllers/Billing/PaddleUserController.php` (both purchase handlers)
- `contentstudio-backend/app/Http/Requests/Billing/UpgradeSubscriptionLimitsRequest.php`
- `contentstudio-backend/app/Services/Billing/SubscriptionLimitUpgradeService.php`
- `contentstudio-backend/app/Http/Controllers/Billing/PlanController.php`
- New migration to set whatever new plan-level signal is agreed

**Frontend**
- `contentstudio-frontend/src/modules/billing/constants/billingAddonCatalog.ts`
- `contentstudio-frontend/src/modules/billing/composables/useBillingAddonPolicy.ts`
- `contentstudio-frontend/src/modules/billing/components/AdjustLimitsModal.vue`
- `contentstudio-frontend/src/modules/billing/components/LimitItem.vue`
- `contentstudio-frontend/src/modules/setting/components/billing/sections/UsageLimitsCard.vue`
- `contentstudio-frontend/src/modules/setting/components/billing/sections/PlanDetailsCard.vue`
- `contentstudio-frontend/src/modules/setting/components/api/ApiModule.vue`
- The limit-exceeded upsell modals listed above
- `contentstudio-frontend/src/locales/*/settings.json`
