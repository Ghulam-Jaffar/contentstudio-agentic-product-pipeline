# Research: Quarterly Billing Cycle

**Date:** 2026-09-25
**Scope:** Add a quarterly billing cycle next to monthly and annual for the **new Paddle Billing plans only** (Standard, Advanced, Agency Unlimited, API plan, and their card-trial copies). Legacy Paddle Classic, FastSpring, Apple, lifetime and deal plans are out of scope. Stories are written from the frontend point of view.

Pricing reference page (PO-approved figures): https://claude.ai/artifact/Mtqht713YfeT9QYqcNxA6b

---

## 1. Locked pricing decisions (PO, 2026-09-25)

### Plans

| Plan | Monthly | Quarterly | Annual (unchanged) |
|---|---|---|---|
| Standard | $29/mo | $25/mo, **$75 every 3 months**, 13.79% off, saves $48/yr | $19/mo, $228/yr, 34.48% off |
| Advanced | $69/mo | $59/mo, **$177 every 3 months**, 14.49% off, saves $120/yr | $49/mo, $588/yr, 28.99% off |
| Agency Unlimited | $139/mo | $119/mo, **$357 every 3 months**, 14.39% off, saves $240/yr | $99/mo, $1,188/yr, 28.78% off |
| API plan | $19/mo | $17/mo, **$51 every 3 months**, 10.53% off, saves $24/yr | $15/mo, $180/yr, 21.05% off |

Card-trial (7-day, card required) plans offer all three cycles at the same prices.

### On/off add-ons

Rule: **a quarterly add-on gets exactly its plan's quarterly discount %**, rounded to the nearest whole dollar. **Annual add-on prices stay exactly as they are today** (they keep their own discounts, they do not move to the plan rate).

| Add-on | Monthly | Quarterly (plan's quarterly %) | Annual (unchanged) |
|---|---|---|---|
| White Label | $50 | Standard $129 · Advanced $128 · Agency Unlimited $128 | $500 |
| White Label Reseller | $50 | Standard $129 · Advanced $128 · Agency Unlimited $128 | $500 |
| SAML SSO | $150 | Standard $388 · Advanced $385 · Agency Unlimited $385 | $1,440 |
| Social Listening | $99 (Advanced) / $150 (Agency Unlimited) | $254 (Advanced) / $385 (Agency Unlimited) | $843 / $1,282 |

- The API plan has **no on/off add-ons**. Its 10.53% quarterly discount applies only to quantity add-ons in Manage add-ons.
- **AI Auto-Reply is not an on/off add-on** (plans get default credits); it is excluded from this pricing work.

### Quantity add-ons (Manage add-ons modal)

Today annual multiplies each unit price by 12 and applies the plan's annual discount. **Quarterly multiplies by 3 and applies the plan's quarterly discount** (13.79% / 14.49% / 14.39% / 10.53%).

---

## 2. Competitor & industry research

### Why quarterly

- **Cash flow and fewer failed renewals.** The vendor collects 3 months up front, and failed payment attempts drop from 12 a year to 4 ([Baremetrics](https://baremetrics.com/blog/annual-vs-monthly-pricing-better-retention)).
- **A commitment step between monthly and annual.** It suits buyers who won't pay for a full year up front.
- **Agencies and SMBs budget by quarter.** Hootsuite offers "Quarterly, Net 45" payment terms to enterprise customers on request ([Hootsuite payment terms](https://www.hootsuite.com/legal/payment-terms)).
- **Counterpoint:** quarterly can "satisfy neither goal" if it sits too close to annual ([Salable](https://medium.com/salable-app/flexible-billing-beyond-monthly-subscriptions-8729f78d2487)). This is why we keep the gap to annual wide: ~14% vs ~29-34%.

### Competitor table

| Competitor | Quarterly? | Cycles offered | Annual discount | Toggle / savings display | Notes |
|---|---|---|---|---|---|
| Sendible | **Yes, not advertised** | Monthly, quarterly, biannual, annual | 15% (quarterly 5%, biannual 7%) | Page shows only Monthly / Yearly ("Save 15%") ([pricing](https://www.sendible.com/pricing)) | Quarterly rates appear only in a [support article](https://support.sendible.com/hc/en-us/articles/19771965502493-Subscription-discounts). The only confirmed self-serve quarterly in the category. |
| Hootsuite | Enterprise payment terms only | Monthly, annual | Not confirmed | Annual price is the headline ([plans](https://www.hootsuite.com/plans)) | "Quarterly, Net 45" through sales |
| Buffer | Not found | Monthly, annual | ~20% ("2 months free") | Toggle ([pricing](https://buffer.com/pricing)) | Per-channel pricing |
| Publer | Not found | Monthly, yearly | 20% | Toggle ([plans](https://publer.com/plans)) | |
| Later | Not found | Monthly, annual | 25% | Toggle ([pricing](https://later.com/pricing/)) | |
| Sprout Social | Not found | Monthly, annual, multi-year | ~20-25% | Annual per-seat price is the headline | |
| Loomly | Not found | Monthly, annual | 25% | Toggle ([pricing](https://www.loomly.com/pricing)) | |
| SocialBee | No | Monthly, annual | ~16-17% | Toggle ([pricing](https://socialbee.com/pricing/)) | |
| Agorapulse | Not found | Monthly, annual | Up to 20% | Annual per-user price is the headline | |
| Metricool | Not found | Monthly, annual | ~20-24% | Toggle, "2 months free" | |

**Takeaway:** no competitor shows quarterly on its public pricing toggle. A visible quarterly option is a small differentiator, especially for agencies.

### UI patterns

- **Where and how.** Put the toggle above the plan cards, and use a segmented control for three options.
- **Big number and savings.** Show the per-month equivalent as the large number, with the billed amount underneath ("billed $Z every 3 months"). Show savings as a dollar figure and/or a %.
- **Default.** Many pages default to annual.
- **No benchmark yet.** No benchmarked three-option toggle was found, so ContentStudio sets its own pattern here.

### Discount levels

- **Only category data point:** Sendible gives quarterly 5% and annual 15%.
- **Rule of thumb:** a quarterly discount is usually a third to a half of the annual one.
- **Ours:** ContentStudio's 10.5-14.5% sits inside that range, since annual is 21-34%.

### Add-ons on quarterly

- **Paddle requires one cycle per subscription.** Every recurring item must share the same billing cycle ([Paddle](https://developer.paddle.com/build/lifecycle/subscription-creation)). On a quarterly plan, every add-on must also be quarterly, which means a quarterly price ID for each add-on.
- **Showing only the current cycle is right.** Showing only the current cycle's option in add-on purchase modals matches this constraint.

---

## 3. Codebase analysis

### 3.1 Backend (context for dependencies; not the story focus)

**How it works today:**
- **Plans:** Mongo `subscription_plans`, model `contentstudio-backend/app/Models/Account/Subscription.php`. Each record has `slug`, `price`, `paddle_id`, `limits`, `features`. There is **no billing-period field**. The cycle is encoded only in the slug suffix (`-monthly` / `-annually`), as the comment at `app/Services/Billing/UsageLimitsService.php:211-221` notes.
- **Price ID to plan map:** `config/paddle.php` → `paddle.billing.products_ids.{env}`. Sandbox is L11-603, staging L605-1227, production L1229-1841. Each entry has `type`, `name`, `db_key`, `slug`, `billing_cycle` ('monthly'|'annually'), `plan_tier`, `is_discounted_price`.
- **Annual discount:** the annual record stores 12 × monthly. Paddle discount codes `StandardOFF`, `AdvancedOFF`, `AgencyUnlimitedOFF`, `ApiAnnualOFF` (`config/paddle.php:1835-1840`) are applied in `PaddleUserController.php:1362-1371` via `PlanHelper::getDiscountId` (`PlanHelper.php:272`).
- **Cycle detection is two-way everywhere:** `interval === 'year' ? 'annually' : 'monthly'`. Paddle sends quarterly as `{interval: 'month', frequency: 3}`, so it would be treated as monthly. Affected sites:
  - `PaddleUserController.php:932` (white label), `:999` (AI auto-reply), `:1195` (social listening), `:1264` (SAML), `:1355`
  - `SubscriptionLimitUpgradeService.php:64,146`
  - `SocialAccountAutoScaleService.php:171`
  - `AutoScaleController.php:357`
  - `SocialAccountPricing.php:28,50-57,125-135`
  - `PlanController.php:452`
  - `UsageLimitsService.php:205,218-221` (`is_annually`)
  - `PlanHelper.php:59-76,195-213,282-287,771-778,829-841` (annual add-on discount rates)
  - `CustomEventHelper.php:44-68` (MRR maths, month/year only)
- **Frequency is stored but unused.** `PaddleBillingSubscription` stores `interval` and `frequency` (`BillingCycleData`), but nothing reads `frequency`.
- **Webhooks** (`PaddleBillingController.php:327-545`) map `price_id` to a plan record. Every quarterly price needs its own `subscription_plans` record and config entry.
- **Credits reset on the 1st of every month regardless of cycle** (`ResetUsedCreditsCommand`, `config/billing.php credit_reset_cron`). Quarterly needs no credit changes.
- **Upgrade and cycle switch:** `POST /paddle/billing/upgradeSubscription` (`PaddleUserController.php:1297-1388`) with `prorated_immediately`. A cycle switch is just a different price ID.
- **Public API:** `is_annually` is in the OpenAPI schema (`app/Http/Controllers/Api/V1/LimitApiSchemas.php:49`).

**Backend work the FE stories depend on:**
1. Quarterly Paddle prices for every plan, card-trial copy and add-on, in all 3 environments.
2. `*-quarterly` plan records.
3. A cycle resolver that reads `interval` + `frequency`.
4. A real `billing_cycle` field in the plan and limits responses.
5. Quarterly add-on price resolution, including the per-plan Social Listening SKUs.
6. Quarterly discount handling for quantity add-ons.
7. MRR maths.

### 3.2 Frontend (story focus)

**Cycle model today.** The cycle is a boolean `isAnnually` everywhere, plus substring tests on the plan slug. There is no cycle enum.

#### Plan selection

- `contentstudio-frontend/src/modules/billing/components/BillingCycleToggle.vue`
  - Uses `SegmentedControl` (`size="md"`, `radius="full"`) with `defineModel<boolean>`.
  - Options `monthly` "Monthly" / `yearly` "Yearly". The yearly option carries an inline "Save up to 34%" span (`settings.billing.subscription_plans.save_up_to`).
- `SubscriptionPlansMain.vue`
  - Default cycle at L67: `isAnnually = ref(!isCurrentBillingCycleMonthly())`. It defaults to annual for new users and to monthly for any slug without "annual", which would include a quarterly slug.
  - Other lines: toggle at L152, cards at L171-179, confirm emit at L128. The toggle is hidden for `change-trial-plan` (L150-154).
- `SubscriptionPlanCard.vue`
  - Price math at L158-202: `basePrice` monthly, `annualPrice = basePrice*12*(1-discountPercentage)`, `formattedMonthlyPrice = annualPrice/12`.
  - Display at L575-588: "$X/month", then "${amount} billed annually" (`billed_annually`) and an emerald pill "Save ${amount} per year" (`save_per_year`).
- `modules/billing/constants/plansDetails.ts`
  - `planTypes` (L7-16) use `-monthly|-annually` suffixes.
  - `billing` shape: `{monthly, yearly:{price,annual,saveAmount,discountCode,discountPercentage}}` (L53-69).
  - Plan entries: Standard L86-96, Advanced L213-223, Agency L359-369, API L505-515. The API values are marked "Placeholder".
- `PlansComparisonTable.vue`
  - Reuses the toggle (L122-126). Price is `isAnnually ? yearly.price : monthly.price` + "/mo" (L136-137).
- `modules/onboarding/views/OnboardingTrialPage.vue`
  - Embeds `SubscriptionPlansMain` with `cc-trial-only` (L58-60), so it gets the same toggle.
- `useBilling.ts` (`modules/billing/composables/`)
  - `resolvePlanCtaState` (L93-127): the slug match `${planId}-${annually|monthly}` gives current or change-plan; the same plan on the other cycle gives change-plan.
  - `isCurrentPlan` (L448-466).
  - `isCurrentBillingCycleMonthly` (L858-865): `!slug.includes('annual')`.
  - `isNewBillingMonthlyPlan` (L191-203): a hardcoded monthly list.
  - `generatePlanItems` (L726-818): `billingCycle` and `cycleKey` (with `-cc`), `socialAccounts[billingCycle]`, `apiCredits[billingCycle]`, and a discount code only when annual.
  - `handlePlanChange` (L694-724).
- **CTA states:** `choose` "Choose Plan", `current` "Current Plan", `change-plan` "Change plan", `trial` "Start your 7-day free trial", `upgrade`, `downgrade`, `switch`.
- **Confirm popup:** "Confirm Purchase" / "Confirm Plan Change" with `purchase_warning` + `addon_reset_warning`.

#### Checkout

- `modules/billing/composables/usePaddle.ts`
  - `PlanCyclePrices {monthly, annually, 'monthly-cc', 'annually-cc'}` (L15-20).
  - `socialAccounts` and `apiCredits` keyed by monthly/annually (L30-38).
  - Hardcoded `priceIds` per environment (`qa-features`, `develop`, `staging`, `uat`, `production`) at L137-378.
  - Analytics mapping `interval === 'year' ? 'yearly' : 'monthly'` at L576-579.
- `PaddleCheckoutModal.vue`
  - Cycle badge (L128-141): trial "{days} days free", else `interval === 'year'` → "Annual Plan" : "Monthly Plan".
  - "Billing cycle:" row (L231-237): "Annual" / "Monthly".
  - The type at L14/L21 covers only `interval`, so **a quarterly price would render "Monthly"**.

#### On/off add-on unlock modals

**Shared behavior in all of them:**
- `disabledPlans = { monthly: slug.includes('month'), annual: slug.includes('annual') }`, applied only for Paddle Billing users. Naming is inverted: `disabledPlans.monthly` disables the **annual** tile.
- The other cycle's tile is **shown but disabled** (opacity-50) and the current cycle is pre-selected.
- A `*-quarterly` slug matches neither check, so both tiles would be enabled and nothing pre-selected.
- Legacy Classic users can pick either tile and go to Classic checkout.

**`modules/listening/components/ListeningUpgradeModal.vue`**
- Tiles at L328-383, logic at L96-167, `Modal` + `Radio` from `@contentstudio/ui`.
- Title "Unlock Social Listening", CTA "Unlock Social Listening".
- Tile labels "Monthly" / "Annual". Prices "$99/month" / "$843/year" (Advanced), "$150/month" / "$1,282/year" (Agency), keyed `listening.unlock_modal.${tier}_monthly_price` / `_yearly_price`.
- Disabled tooltip `plan_switch_disabled_tooltip`: "You currently have {currentPlan} billing enabled, so switching to {targetPlan} is not available right now."
- Confirm "Confirm purchase" / "The Social Listening addon will be added to your current subscription and billed immediately."
- Sends `billing_cycle: planType` (`'monthly'|'annual'`, `api/billing.ts:263`).
- Hardcoded analytics prices at L60-66.

**`modules/setting/components/sso/SamlUpgradeModal.vue`**
- Tiles at L222-269, logic at L130-168.
- Title "Unlock Organization Single Sign-On".
- Tiles "Monthly" "$150/month" / "Annual" "$1,440/year". CTA "Purchase Now".
- Sends `billing_cycle` (`api/sso.ts:85`). Analytics price at L44 (`annual ? 1440 : 150`).
- `sso_upgrade.monthly_subtext` / `yearly_subtext` ("Billed annually. Save $300.", wrong since the real saving is $360) are defined but **not rendered**.

**`modules/setting/components/white-label/WhiteLabelUpgradeModal.vue`**
- Tiles at L224-273, logic at L118-139. Embedded in a `CstuModal` in `EnrolledPlanView.vue` L8-16.
- Title "White Label". Tiles "Monthly" "$50/month" / **"Yearly"** "$500/year".
- Native radios, legacy `CstButton`, no disabled tooltip. The Paddle Billing confirm sends only `workspace_id`.
- Stale `select_plan_tooltip`: "$25/month OR $250/year".

**`components/common/FeatureAddOnModal.vue`**
- Used only by the Inbox add-on (`components/addons/InboxAddOnModal.vue`). All English is hardcoded; the Classic path uses white-label Paddle IDs.
- **Not used for White Label Reseller.** No White Label Reseller purchase modal exists in the app. Reseller is sold outside these modals (the plan details and pricing strings only list its price).

#### Manage add-ons (quantity add-ons)

**`modules/billing/components/AdjustLimitsModal.vue` + `composables/useAdjustLimits.ts` + `LimitItem.vue`**
- **Cycle input:** `isAnnually` comes from the API field `is_annually` (useAdjustLimits.ts L585; `types/generated/billing.ts:87`).
- **Header:**
  - Title "Manage add-ons", summary "Billing Summary".
  - Pill `isAnnually ? "Annual Plan" : "Monthly Plan"` (L96-102).
  - Tooltip `tooltips.billing_cycle`: "Add-ons follow the same billing cycle as your plan. / You'll be billed immediately (prorated) for the remaining time in your current cycle. / ..."
- **Summary rows:**
  - "Add-on Additions: +$X", "Add-on Removals: -$X" (L336-363, ×12 **without** discount).
  - "Discount: {n}%" (L126-133, annual only).
  - "Net Adjustment".
  - "New Total: ${total} / annually|monthly" (L153-157).
- **Discount maths:** `applyDiscount` (useAdjustLimits.ts L827-845) applies ×12 and then a slug switch: `standard-annually` 0.3448, `advanced-annually` 0.2899, `agency-unlimited-annually` 0.2878, none for API. It is duplicated in `AdjustLimitsModal.vue` L371-386.
- **Unit prices:** hardcoded monthly values at useAdjustLimits.ts L147-161.
- **LimitItem row cost:** "+ $X/month|year" (L156, L306), with `annualMultiplier = isAnnually ? 12 : 1` (L861-920) and no discount on the row.
- **Footer tooltip:** "Prorated charges will apply" says "remaining days of the current month", even on annual.
- **Analytics:** `billing_cycle: isAnnually ? 'annually' : 'monthly'` (L794-806).
- **`AutoScaleLimitsModal.vue`:** takes `price_unit` from the backend (default 'month', L105/L125) and interpolates it into "approx. {price}/account/{unit}".

#### Billing settings page

- **`modules/setting/components/billing/EnrolledPlanView.vue`:** plan name comes from backend `display_name` (L81-83).
- **`sections/PlanDetailsCard.vue`:**
  - Shows "Renewal Amount" and "Next Renewal:" (L122-145).
  - Buttons: "Upgrade Subscription", "Change Plan", "Upgrade Limits".
  - **Shows no cycle text.** The cycle appears only if `display_name` contains it.
- **`sections/SubscriptionsTable.vue`:** columns "Subscription Plan", "Created Date", "Renewal Amount", "Renewal Date", "Status", "Action". It has no cycle column.

#### Other slug lists and mappers that need quarterly slugs

- `modules/common/composables/usePermissions.ts:282-331` (`isPlanUpgradeable`, `isHighestPlan`)
- `modules/billing/composables/useBillingAddonPolicy.ts:23-28` (`STANDARD_PLAN_SLUGS`)
- `modules/listening/composables/useSocialListeningTier.ts:9-12` (`ADVANCED_PLAN_SLUGS`)
- `modules/common/composables/useWorkspace.ts:127-131` (`isMonthlyPlan`)
- `services/helpin.ts:47-48`, `composables/useFrillWidget.ts:107-109`, `composables/useFrillSurvey.ts:128-130` (cycle labels)

#### Existing analytics

- `usePaddle.ts:576-579` sends `billing_cycle` 'yearly'|'monthly' on checkout.
- `useAdjustLimits.ts:794-806` sends `billing_cycle` 'annually'|'monthly' on add-on purchase.
- The unlock modals send analytics with hardcoded prices.
- Before naming new events, reuse `addon_purchased` and the existing checkout events (search `userMaven.track(`).

#### Locales

10 locales: cs, de, el, en, es, fr, it, pl, sv, zh. Billing copy lives in `settings.json` (`settings.billing.*`, `settings.sso_upgrade.*`, `settings.white_label.upgrade.*`) and `listening.json`.

### 3.3 Mobile impact

**None.** `contentstudio-flutter/lib/features/billing/` sells one iOS StoreKit product (`com.contentstudio.standard.monthly.v3`, `domain/subscription_product.dart:11`). It never shows or changes the web plan cycle. No `[Flutter]` story.

### 3.4 Developer surfaces

- **Public API:** the limits schema exposes `is_annually` (`LimitApiSchemas.php:49`). A new `billing_cycle` field must be added alongside it, keeping `is_annually` for backward compatibility. That work belongs in the backend story. The CLI and MCP read the same field.

---

## 4. Gotchas and latent issues found

1. **`isCurrentBillingCycleMonthly()` treats every non-annual slug as monthly**, so quarterly users would see the toggle default to Monthly.
2. **The unlock modals' slug checks** (`includes('month')` / `includes('annual')`) match neither on a quarterly slug. The fix is to read a real cycle from the backend.
3. **The checkout modal only tests `interval === 'year'`.** Quarterly has to read `frequency: 3`.
4. **AdjustLimits rows are undiscounted by design.** Additions and Removals show ×12 without the discount, the Discount line shows the %, and New Total is discounted. PO decision (2026-09-25): keep this layout, and quarterly follows it with ×3 rows.
5. **The "Prorated charges will apply" tooltip says "current month"** even on annual. It should say "current billing period".
6. **The White Label modal's `select_plan_tooltip`** still says "$25/month OR $250/year".
7. **The SAML "Save $300" subtext is wrong** (the saving is $360) and never rendered.
8. **Price mismatches that predate this work:**
   - The Social Listening plan details show $950 / $1,440 annual against $843 / $1,282 at checkout.
   - The API plan is $15/mo on the website against $19/mo in the app.
9. **White Label Reseller has no in-app purchase modal.** Its quarterly price only needs to exist in Paddle and wherever it is displayed.
10. **Paddle cannot mix cycles.** An existing monthly or annual subscriber who switches to quarterly must have every add-on move to its quarterly price in the same change.
