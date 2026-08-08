# Research — Sync the in-app Upgrade Plan UI with the website pricing page

Item 15 of the 7 Aug 2026 backlog batch.

## Current state

### Where the in-app plan UI lives

`contentstudio-frontend/src/modules/billing/`

- `components/SubscriptionPlansModal.vue` — the modal shell (`CstuModal`, id `subscription-plans-modal`). Reads the current plan's limits from the plan store on show to set its context.
- `components/SubscriptionPlansMain.vue` — the plan selection body.
- `components/SubscriptionPlanCard.vue` — one plan card: price, savings, feature list, call to action. Computes `isAnnually` pricing, `getSaveAmount`, `discountPercentage`, `isPopular`, `isAgencyPlan`, `isApiCentricPlan`, and the CTA text with branches for trial, change-trial-plan and credit-card-trial-only cases.
- `components/PlansComparisonTable.vue` + `constants/plansComparison.ts` — the feature comparison table.
- `components/FeatureValue.vue`, `LimitItem.vue`, `SocialNetworks.vue` — cell renderers.
- `components/PaddleCheckoutModal.vue` — checkout.
- Copy lives in `src/locales/*/settings.json` under `settings.billing.*`.

### Plan set in the app

Plan ids used across the billing constants: `standard`, `advanced`, `agency-unlimited`, `api-centric`, each with `-monthly` and `-annually` variants.

`plansComparison.ts` types its comparison rows with exactly four plan columns: `standard`, `advanced`, `agency-unlimited`, `api-centric`.

Comparison categories in the app, in order: Publishing and scheduling, Planning and collaboration, Analytics and reporting, Social inbox, RSS feed management, Essentials and support.

### The API plan is already in-app, and it is the most bespoke card

Confirmed by the PO: **Enterprise is deliberately not coming into the in-app UI.** The API plan occupies the slot the website gives Enterprise. It is already rendered today, but its presentation needs updating alongside everything else.

What the API-Centric card already does, all in `SubscriptionPlanCard.vue`:

- **Its own visual treatment.** `planIdentifiers.isApiCentricPlan` drives `border-2! border-[#16a34a]!` (line 326), sitting beside `isPopular`'s `border-2! border-[#007bff]!`. **Both are hardcoded hex**, so neither follows a white-label primary colour. This is a concrete defect the build story has to fix regardless of the redesign.
- **A social-account stepper.** `apiCentricSocialHandler` (line 68) with `increment`, `decrement`, `validate` and `finalize`, rendered at lines 456 to 480. The user dials how many social accounts they want.
- **Dynamic pricing driven by that stepper.** `hasDynamicPricing` (line 127) is true for agency plans and `api-centric`, which changes how savings are computed: `getSaveAmount` calculates `basePrice * 12 - annualPrice` rather than reading the plan's static `billing.yearly.saveAmount`.
- **A parallel-track feature heading.** `getFeatureListHeading` (line 155) has an explicit comment: *"api-centric is a parallel track, not part of the standard → advanced → agency ladder"*. It gets "You get" rather than "Everything in {previous plan}, plus".
- **A Media Storage row that only it shows** (line 530).

Visibility rules live in `SubscriptionPlansMain.vue` (lines 41 to 45), with the comment *"Hide Standard when on api-centric; hide API Centric on white-label domains"*:

- A customer already on the API plan does not see Standard.
- **On a white-label domain the API plan is hidden entirely.**

That last rule matters for the redesign: whatever the API plan's new slot looks like, the layout has to hold with that column absent, because white-label customers never see it.

The billing toggle already has a `save_up_to` string (`SubscriptionPlansMain.vue` line 159), which is the app's existing counterpart to the website's "Save up to 34%".

### What the website pricing page now shows

Fetched from `contentstudio.io/pricing` on 7 Aug 2026:

| Plan | Monthly | Annual | Notes |
|---|---|---|---|
| Standard | $29/mo | $19/mo, $228/yr | saves $120/yr |
| Advanced | $69/mo | $49/mo, $588/yr | saves $240/yr |
| Agency Unlimited | $139/mo | $99/mo, $1,188/yr | saves $480/yr, badged **Best Value** |
| Enterprise | custom | billed annually | CTA **Contact Sales** |
| API-Centric | $15/mo | $180/yr | 21% saving |

Limits shown on the website:

| | Standard | Advanced | Agency Unlimited |
|---|---|---|---|
| Social accounts | 5 | 10 | 25 |
| Workspaces | 1 | 2 | Unlimited |
| Users | 1 | 2 | Unlimited |
| AI text credits | 25,000 | 50,000 | 125,000 |
| AI image credits | 25 | 50 | 125 |
| AI video credits | 100 | 200 | 500 |

Other website specifics:

- Billing toggle labels: **Monthly** and **Yearly**, with **Save up to 34%** alongside.
- CTAs in use: **Start your free trial**, **Start your 7-day free trial**, **Book a demo**, **Contact Sales**.
- Trial framing: **7-day free trial**, with **No credit card required • Cancel anytime**.
- An **Add Ons** section, priced monthly or annually: Social listening, Extra social account, Extra user, Extra workspace, White-label, SSO (SAML), White-label reseller.

## The deltas

These are the differences worth designing against. The designer should re-diff against the live page at the time they pick this up, since the website can move again.

1. **Enterprise is on the website and stays off the app. Decided.** The PO has confirmed Enterprise is not coming into the in-app upgrade UI. The API plan occupies that slot instead, and it is already rendered today. So `plansComparison.ts` keeps its four columns and needs no structural change, which removes what would otherwise have been the largest piece of work here. What remains is updating how the API plan presents, alongside the other three.
2. **Billing-cycle wording.** The website says Monthly and Yearly with "Save up to 34%". The app's card computes a per-plan `save_per_year` amount and a per-plan discount percentage. Both can be right, but the toggle label and the headline saving claim should read the same in both places.
3. **Best Value versus popular.** The website badges Agency Unlimited as Best Value. The app card has an `isPopular` flag. Same intent, different word.
4. **AI credit presentation.** The website separates text, image and video credits with specific numbers per plan. Worth confirming the app's limit rows present the same three credit types with the same numbers.
5. **Add-ons.** The website presents a named Add Ons section with seven add-ons and monthly or annual pricing. The app surfaces add-ons through separate flows (auto-scale limits, white-label, SAML, X wallet, listening). The upgrade UI should at least acknowledge the same add-on set with the same names.
6. **Comparison table categories.** The app has six categories. Whether they match the website's grouping and ordering needs checking, so a user comparing the two pages does not have to re-orient.
7. **Trial framing.** The website is explicit about 7 days, no credit card and cancel anytime. The app card has trial branches (`changeTrialPlan`, `ccTrialOnly`) whose copy may or may not say the same thing.

## What needs to change

- A design pass that diffs the in-app upgrade UI against the live pricing page and specifies the reconciled result: plan presentation, prices, badges, limits, feature comparison, add-ons, CTAs, terminology and visual hierarchy.
- A frontend pass that implements it, keeping the existing checkout, trial and plan-change logic intact. This is a presentation and copy change, not a billing-logic change.

## Explicitly not in scope

Changing prices, plan entitlements or what any plan actually includes. If a number differs between the app and the website, the correct fix is to display the true entitlement, and any change to the entitlement itself is separate billing work.

## Related existing work

- `docs/stories/billing-change-plan-on-max-plan/` — shows the Change plan call to action on the highest plan. Touches the same cards, so worth sequencing with this.
- `docs/stories/apple-iap-billing-ui/` — mentions the upgrade modal for the in-app-purchase path.
- `docs/features/api-centric-plan/` — where the API-Centric plan's in-app presentation was defined.
- `docs/stories/social-account-auto-scale-limits/`, `saml-addon-modal`, `whitelabel-reseller-promo` — the add-on flows the website's Add Ons section names.

## Files involved

- `contentstudio-frontend/src/modules/billing/components/SubscriptionPlansModal.vue`
- `contentstudio-frontend/src/modules/billing/components/SubscriptionPlansMain.vue`
- `contentstudio-frontend/src/modules/billing/components/SubscriptionPlanCard.vue`
- `contentstudio-frontend/src/modules/billing/components/PlansComparisonTable.vue`
- `contentstudio-frontend/src/modules/billing/components/{FeatureValue,LimitItem,SocialNetworks}.vue`
- `contentstudio-frontend/src/modules/billing/constants/plansComparison.ts`
- `contentstudio-frontend/src/locales/*/settings.json` under `settings.billing.*`

## Mobile

The mobile apps have their own in-app-purchase billing surface and do not render this modal. No mobile story here. If the plan presentation in the apps should follow, that is separate work on the in-app-purchase screens.
