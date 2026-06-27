# Epic & Stories — X (Twitter) Pay-Per-Use Credit Wallet

**Scope of this doc:** the **epic** + **5 `[FE]` stories** and **1 `[BE]` umbrella story** (per PO scoping — backend is a single story the BE team will sub-task). No `[Design]` story for now. Web only. Nothing is pushed to Shortcut — this is the markdown the PO creates from.

**Mock-ups (all stories):** https://claude.ai/artifacts/latest/a2271a4d-5679-4ab4-ba96-4a43b900b177 — viewable by members in the organization.

**Pricing constants used throughout:** plain post = **$0.018** (X $0.015 + 20%), link post = **$0.24** (X $0.20 + 20%). Wallet is a prepaid, non-expiring dollar balance at the **account/super-admin level**.

---

## Epic: X (Twitter) Pay-Per-Use Credit Wallet

X (Twitter) moved its publishing API to pay-per-use pricing, charging ContentStudio per post ($0.015 plain, $0.20 with a link) against a prepaid balance. ContentStudio's old fixed daily-limit + recurring add-on model no longer fits. This epic replaces it with a **prepaid dollar wallet for X**: a non-expiring balance that deducts on each successful publish at X's cost + a 20% service fee, with transparent cost previews in the composer, a dedicated X Wallet card + modal in billing (top-up calculator, auto-recharge, spending limit, usage log), and a fair one-time migration for existing users.

The model is built generic so the same wallet later powers X inbox, analytics, and listening. Billing-capable users (super admins) manage and top up the wallet; other members see clear "ask your super admin" guidance. All X posting is metered (custom developer apps are no longer supported).

**Stories:** FE-1…FE-5 (frontend) + BE-1 (backend umbrella).

---

## FE-1 · [FE] Build the composer X wallet cost/balance widget (projection + over-balance states)
**Shortcut fields:** Template: New Feature Template · Type: feature · Project: Web App · Group: Frontend · Epic: X (Twitter) Pay-Per-Use Credit Wallet · Priority: High · Product area: Composer · Skill set: Frontend · Estimate: — · Labels: none

### Description
As a user composing an X post, I want to see what this post will cost and what my X wallet balance will be after it — including everything I've already scheduled — so I understand the cost up front and know whether my posts will actually publish.

### Workflow
1. The user selects an X (Twitter) account and starts composing. The X wallet widget appears under the post.
2. The widget shows the current balance, what **this** post will use, what **already-scheduled** X posts will use, and the **projected balance** after all of them — clearly labeled as an estimate.
3. If the projected spend exceeds the balance, the widget warns the user — with wording that depends on whether auto-recharge is on, and on whether the user can manage billing.

### Acceptance criteria
- [ ] The widget renders only when an X account is selected; it shows no monthly framing and **no progress bar**.
- [ ] Header shows "X (Twitter) Wallet" and, right-aligned, "Balance: $<balance>" with an info icon whose tooltip reads: **"Your prepaid X balance. Shared across all X posting, and only charged when a post actually publishes."**
- [ ] Row "This post will use:" shows **"$0.018 (plain post)"** or **"$0.24 (with a link)"**, switching live as a link is added/removed.
- [ ] When the user has queued X posts, a line shows **"You also have <N> scheduled X posts that will use ~$<amount> when they publish."** (hidden when N = 0).
- [ ] Row "Projected balance after all of these:" shows the balance minus (queued + this post); it's shown in red when it would go negative.
- [ ] A transparency footnote reads: **"This is an estimate — your balance is charged when each post actually publishes, not now."**
- [ ] For threads, "this post will use" reflects the per-delivered-tweet total.
- [ ] **Over-balance warning** appears when (queued + this) cost exceeds the balance, with copy by case:
  - Auto-recharge **OFF** (billing-capable user): **"Your $<balance> balance won't cover everything you've queued. Posts publish in order until it runs out (about <N> of <M>), and the rest will fail unless you top up."** + a **Manage X Wallet** link that opens the X Wallet modal on the Top-up tab.
  - Auto-recharge **ON, within spending limit:** **"Your balance is low, but auto-recharge is on — it'll top up automatically (up to your $<limit> spending limit) so your posts should keep publishing. They'd only pause if the spending limit is reached."**
  - Auto-recharge **ON, unlimited:** **"Your balance is low, but auto-recharge is on with no limit — all your posts will publish."**
- [ ] **No billing access** (member who can't manage billing): every "Manage X Wallet"/top-up CTA in the widget is replaced with the non-actionable message **"Ask your workspace's super admin to add X wallet credits."**
- [ ] "Post Now" is disabled only when this post's cost alone exceeds the balance and auto-recharge can't cover it; otherwise enabled.
- [ ] When a post is blocked at the composer for an empty/insufficient wallet, a `x_post_blocked_insufficient_balance` Usermaven event fires with `{}`.

### Mock-ups
https://claude.ai/artifacts/latest/a2271a4d-5679-4ab4-ba96-4a43b900b177 — viewable by members in the organization.

### Impact on existing data
None (read/display only). Reads wallet balance, the per-post rates, and the user's queued X posts.

### Impact on other products
Composer (web). No mobile/Chrome impact. White-label safe (theme tokens; copy uses no hardcoded brand name beyond "X (Twitter)").

### Dependencies
- **[BE] Implement the X pay-per-use wallet backend (deduction, billing, allocation, emails)** — provides balance, rates, queued-cost, and auto-recharge state.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*
- `contentstudio-frontend/src/modules/composer_v2/components/TwitterPostUsageAlert.vue` (in `PostingSchedule.vue`) — the existing X usage widget to convert from "posts used / daily limit" to this dollar projection.
- `contentstudio-frontend/src/modules/composer_v2/views/SocialModal.vue` (`initTwitterLimits`) + `composables/useComposerHelper.js` (`fetchTwitterLimits`, today hits `GET api/planner/getXPostsCount`) — extend to return balance, rates, queued-cost, auto-recharge state.
- Billing-access check: `hasPermission('can_see_subscription')` (mirrors the Social-Listening / white-label "contact super admin" pattern).
- `const { trackUserMaven } = useUserMaven()` for the event.

---

## FE-2 · [FE] Add the composer URL cost heads-up popup for X posts
**Shortcut fields:** Template: New Feature Template · Type: feature · Project: Web App · Group: Frontend · Epic: X (Twitter) Pay-Per-Use Credit Wallet · Priority: High · Product area: Composer · Skill set: Frontend · Estimate: — · Labels: none

### Description
As a user adding a link to an X post, I want a one-time heads-up that links cost much more, with my balance impact, so I can decide whether to keep the link before it quietly drains my wallet.

### Workflow
1. The user has an X account selected and types/pastes a URL into the post.
2. When focus leaves the text editor, a one-time popup explains the link's higher cost and the balance impact.
3. The user keeps the link ("Got it"), removes it ("Remove link"), or opts out of future popups.

### Acceptance criteria
- [ ] The popup appears when **(a)** an X account is selected, **(b)** the post contains a URL, and **(c)** focus leaves the text editor — at most once per distinct URL, and never if "Don't show again" was chosen.
- [ ] Title: **"Heads up — your link makes this post cost more"**.
- [ ] Body: **"Posts with a link use X's higher API rate. This post will cost $0.24, vs $0.018 for a plain-text post."**
- [ ] A balance line: **"Your balance: $<balance> → after this post: $<balance − 0.24>"**.
- [ ] A **"Don't show this again"** checkbox that, when checked, prevents the popup from showing again (persisted per user).
- [ ] Buttons: secondary **"Remove link"** (strips the URL from the post) and primary **"Got it"** (dismiss).
- [ ] The popup is **non-blocking** — the user can still publish either way.
- [ ] An amber banner above the wallet widget reads, only while the post contains a URL: **"This post contains a URL and therefore costs $0.24 to align with X's latest API pricing, compared to $0.018 for a plain-text post."**
- [ ] All copy is added to every locale directory under `src/locales/`, English first.

### Mock-ups
https://claude.ai/artifacts/latest/a2271a4d-5679-4ab4-ba96-4a43b900b177 — viewable by members in the organization.

### Impact on existing data
None — UI affordance + a per-user "don't show again" preference.

### Impact on other products
Composer (web) only.

### Dependencies
- **[FE] Build the composer X wallet cost/balance widget (projection + over-balance states)** (shares the cost/balance data and lives in the same composer area).

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*
- Same composer surface as FE-1 (`PostingSchedule.vue` / `SocialModal.vue`). URL detection can reuse the existing link helpers used for the composer's link handling.
- Persist "don't show again" via a per-user preference (`setPreferenceStatus` in `src/modules/common/composables/useHelper.js`).

---

## FE-3 · [FE] Add the X Wallet billing card and Manage X Wallet modal shell (remove X from add-ons)
**Shortcut fields:** Template: New Feature Template · Type: feature · Project: Web App · Group: Frontend · Epic: X (Twitter) Pay-Per-Use Credit Wallet · Priority: High · Product area: Billing · Skill set: Frontend · Estimate: — · Labels: none

### Description
As a super admin, I want a dedicated X Wallet card on the billing page that opens a clear Manage X Wallet modal, so I can manage my prepaid X balance in one place — separate from the recurring add-ons it no longer belongs with.

### Workflow
1. The super admin opens Billing and sees an **X (Twitter) Wallet** card (X is no longer listed among the recurring add-ons / usage limits).
2. They click **Manage X Wallet** (opens the modal on Top up & auto-recharge) or **View usage** (opens it on Usage).

### Acceptance criteria
- [ ] X is **removed** from the Manage Add-ons / Increase Limits modal and from the Usage Limits card; where the X row was, a small note reads: **"X posting moved to the prepaid X Wallet."**
- [ ] A new **X (Twitter) Wallet** card appears on the billing page (under/near Usage Limits) showing: current balance (prominent), an **auto-recharge ON/OFF** status, and a **low-balance** indicator when under the threshold.
- [ ] The card has two CTAs: **"Manage X Wallet"** (opens the modal on **Tab A — Top up & auto-recharge**) and **"View usage"** (opens the modal on **Tab B — Usage**), via an explicit initial-tab parameter.
- [ ] The card and its CTAs are shown only to **billing-capable users** (super admin / `can_see_subscription`); other members do not see the card (or see a read-only balance with "Ask your workspace's super admin to add X wallet credits.").
- [ ] The modal header shows the title **"Manage X (Twitter) Wallet"** with muted subtext: **"A prepaid balance — not a monthly or annual plan. It never expires and only drops as you post."** (No standalone balance line in the header.)
- [ ] The modal has two tabs: **"Top up & auto-recharge"** and **"Usage"**, defaulting to the tab requested by the entry point.
- [ ] The modal is capped at max-height 85vh.

### Mock-ups
https://claude.ai/artifacts/latest/a2271a4d-5679-4ab4-ba96-4a43b900b177 — viewable by members in the organization.

### Impact on existing data
None on the frontend. Removes the X add-on from the recurring add-on UI; reads the wallet balance + auto-recharge state.

### Impact on other products
Billing (web). The X add-on no longer appears in Manage Add-ons; confirm no other surface links to it.

### Dependencies
- **[BE] Implement the X pay-per-use wallet backend (deduction, billing, allocation, emails)** (balance + auto-recharge state).

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*
- `contentstudio-frontend/src/modules/setting/components/billing/sections/UsageLimitsCard.vue` (the "Manage Add-ons" entry + X row), `EnrolledPlanView.vue` (host the card + modal), `modules/billing/components/AdjustLimitsModal.vue` / `LimitItem.vue` / `constants/billingAddonCatalog.ts` (remove the X add-on row).
- Billing-access gating: `hasPermission('can_see_subscription')`.
- Use `@contentstudio/ui` `Modal`, `Tabs`/`SegmentedControl`, `Button`.

---

## FE-4 · [FE] Build the Top up & auto-recharge tab (calculator, spending limit, unlimited)
**Shortcut fields:** Template: New Feature Template · Type: feature · Project: Web App · Group: Frontend · Epic: X (Twitter) Pay-Per-Use Credit Wallet · Priority: High · Product area: Billing · Skill set: Frontend · Estimate: — · Labels: none

### Description
As a super admin, I want to top up my X wallet with a clear calculator and configure auto-recharge with a spending limit, so I can fund X posting predictably and never get a surprise bill.

### Workflow
1. On the Top up & auto-recharge tab, the super admin sees their current balance.
2. They pick a top-up amount; a card shows what the resulting balance buys; they Buy/Top up.
3. They optionally turn on auto-recharge and set a threshold, top-up amount, and a spending limit (or unlimited).

### Acceptance criteria

**Current balance + top-up**
- [ ] A "Current balance" block shows the label, the amount, and an auto-recharge ON/OFF pill (green when on).
- [ ] A top-up control in **$5 increments** (default **$10**) with − / + and a typeable amount, labeled "Top up your wallet".
- [ ] A **"WHAT YOUR $<resulting balance> GETS YOU"** card (resulting balance = current + top-up, updating live; never says "each month") with a responsive tile grid: **plain posts** = floor(balance / 0.018), "$0.018 each"; **posts with a link** = floor(balance / 0.24), "$0.24 each". With exactly 2 tiles show an **"OR"** divider; with 3+, drop it.
- [ ] Card footnote: **"It's one wallet — spend it on either type, in any mix. The numbers above are the max of each on its own. Rates include X's cost + a 20% service fee."**
- [ ] A single **"Buy / Top up $<amount>"** button (rendered once) that completes the top-up and raises the balance everywhere.
- [ ] When the user completes a top-up, a `x_credits_purchased` Usermaven event fires with `{ amount_usd, source: 'manual' }`.

**Auto-recharge (progressive disclosure)**
- [ ] When auto-recharge is **OFF**, only a toggle + one-line hint show: **"Automatically top up when your balance runs low."**
- [ ] When **ON**, three fields + helper appear: "Recharge when balance falls below $" (default **$1**), "Top-up amount $" (default **$10**), and **"Spending limit $"** (default **$30**) with helper: **"Set the most you want us to auto-spend on X. When it's reached, auto-recharge stops and posting pauses until you top up or raise the limit."** — a fixed total with **no monthly/cycle reset**.
- [ ] An **"Allow unlimited spending"** checkbox with copy: **"Turn off the spending limit. Auto-recharge keeps your wallet topped up so posting never pauses — your saved card is charged each time it recharges."** When checked, the Spending limit field is hidden/disabled.
- [ ] When the user saves auto-recharge settings, a `x_auto_recharge_configured` Usermaven event fires with `{ enabled, threshold_usd, topup_usd, spending_limit_usd, unlimited }`.
- [ ] The tab fits **without scrolling** in its default state (auto-recharge OFF).

### Mock-ups
https://claude.ai/artifacts/latest/a2271a4d-5679-4ab4-ba96-4a43b900b177 — viewable by members in the organization.

### Impact on existing data
Writes wallet top-ups and auto-recharge settings (via the BE story). No schema impact on the frontend.

### Impact on other products
Billing (web). The top-up purchase uses the new one-off top-up flow (BE).

### Dependencies
- **[FE] Add the X Wallet billing card and Manage X Wallet modal shell (remove X from add-ons)** (the modal + tab host).
- **[BE] Implement the X pay-per-use wallet backend (deduction, billing, allocation, emails)** (top-up purchase + auto-recharge persistence).

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*
- Calculator pattern to clone: `contentstudio-frontend/src/modules/billing/components/VideoCreditCalculatorModal.vue` + `useVideoCreditCalculator.js`.
- Live rate for the card: source from the pricing config (BE) rather than hardcoding $0.018 / $0.24.
- `@contentstudio/ui` `Switch` (auto-recharge / unlimited), `TextInput` (amounts), `Button`.
- `const { trackUserMaven } = useUserMaven()` for the two events.

---

## FE-5 · [FE] Build the Usage tab (per-post log, breakdown, CSV)
**Shortcut fields:** Template: New Feature Template · Type: feature · Project: Web App · Group: Frontend · Epic: X (Twitter) Pay-Per-Use Credit Wallet · Priority: Medium · Product area: Billing · Skill set: Frontend · Estimate: — · Labels: none

### Description
As a super admin, I want a transparent per-post usage log with a cost breakdown, so I can see exactly where my X wallet money goes and trust the pricing.

### Workflow
1. The super admin opens the Usage tab (directly via "View usage", or by switching tabs).
2. They see every X post's cost and the running balance, plus a summary splitting X's cost from ContentStudio's fee, and can export the log.

### Acceptance criteria
- [ ] The Usage tab shows a per-post log table with columns: **Date · Account · Type (Plain / With link) · Cost · Balance after**.
- [ ] Post type is visually distinguished (e.g., a Plain vs With-link tag).
- [ ] A summary shows **total spent**, split into **"X's cost"** and **"your 20% service fee"**.
- [ ] An **"Export CSV"** action downloads the log.
- [ ] Empty state (no usage yet): a headline + subtext (e.g., **"No X posts yet"** / **"Once you publish to X, every post and its cost will show here."**).
- [ ] Loading and error states are handled (skeleton while loading; a clear retry message on failure).
- [ ] The log paginates / is filterable by account and date range.

### Mock-ups
https://claude.ai/artifacts/latest/a2271a4d-5679-4ab4-ba96-4a43b900b177 — viewable by members in the organization.

### Impact on existing data
None — read-only over the usage ledger (BE).

### Impact on other products
Billing (web).

### Dependencies
- **[FE] Add the X Wallet billing card and Manage X Wallet modal shell (remove X from add-ons)** (the modal + tab host).
- **[BE] Implement the X pay-per-use wallet backend (deduction, billing, allocation, emails)** (the usage ledger + export data).

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*
- Reads the per-post usage ledger from BE; mirror existing billing/usage table patterns in `modules/setting/components/billing/`.

---

## BE-1 · [BE] Implement the X pay-per-use wallet backend (deduction, billing, allocation, emails)
**Shortcut fields:** Template: New Feature Template · Type: feature · Project: Web App · Group: Backend · Epic: X (Twitter) Pay-Per-Use Credit Wallet · Priority: High · Product area: Billing · Skill set: Backend · Estimate: — · Labels: none

> Single umbrella backend story per PO scoping — the backend team will sub-task it.

### Description
Build the backend for the X prepaid dollar wallet: the balance + usage ledger, per-post deduction on publish, configurable pricing, one-off top-up + auto-recharge + spending limit via billing, the rollout allocation/migration, the two transition emails, and the server-side analytics events — so the frontend surfaces have real data and money is collected correctly.

### Acceptance criteria

**Wallet, pricing & deduction**
- [ ] A prepaid USD wallet exists at the **account/super-admin level**, shared across the account's workspaces, **non-expiring**.
- [ ] Pricing is stored as **editable config** (per action type: upstream cost + markup → charged rate); defaults plain $0.015 +20% = $0.018, link $0.20 +20% = $0.24. Changing config requires no deploy.
- [ ] On **successful** X publish, the charged rate is **deducted atomically** from the wallet; **threads deduct per delivered tweet**; **failed publishes deduct nothing**.
- [ ] Link detection at publish time matches X's definition of a URL post.
- [ ] Deduction is race-safe (atomic decrement + ledger; idempotency so retries don't double-charge).
- [ ] When the wallet can't cover a post (and auto-recharge can't/again), the post **fails** with a clear "insufficient X balance" reason; the daily-limit gate is removed.
- [ ] A **usage ledger** records each event (date, account, type, charged amount, X cost, balance after) for the FE usage log + CSV; the ledger is generic enough to later cover inbox/analytics/listening.

**Top-up, auto-recharge & spending limit (billing)**
- [ ] Users can **top up** the wallet as a one-off purchase of any supported dollar amount; on success the balance increases.
- [ ] **Auto-recharge:** when balance falls below the threshold, automatically top up by the configured amount — only while total auto-spend is under the **spending limit** (a fixed total, no cycle reset) or when **unlimited** is set. When the spending limit is reached, auto-recharge stops.
- [ ] Trial users cannot purchase top-ups (must upgrade first).

**Initial allocation & migration (at rollout)**
- [ ] **Trials (from rollout):** grant **$0.50** to the wallet.
- [ ] **New subscribers:** no extra grant (they keep leftover trial balance).
- [ ] **Existing without the X add-on:** grant **$0.30 × number of connected X accounts** (super-admin level, across all workspaces); **0 accounts → no grant**.
- [ ] **Existing with the X add-on:** convert remaining value = **amount paid × fraction of the billing cycle still unused** into wallet balance; retire the recurring add-on.
- [ ] Migration is idempotent and safe to run against existing accounts (no double-grant, no double-charge).

**Transition emails (existing users with an X account only — not trials/new)**
- [ ] A **pre-rollout announcement** email: explains the move to pay-per-use, the per-post pricing, and the rollout date.
- [ ] A **rollout-day** email: states the user's starting wallet (granted amount or converted balance) and the per-post costs.
- [ ] Both honor recipient locale and follow the standard email template.

**Analytics (server-side)**
- [ ] When auto-recharge fires, a `x_auto_recharge_triggered` event fires with `{ amount_usd }`; the resulting purchase also emits `x_credits_purchased` with `{ amount_usd, source: 'auto_recharge' }`.
- [ ] When auto-recharge is blocked by the spending limit, a `x_spending_limit_reached` event fires with `{}`.

### Mock-ups
N/A — backend. (FE surfaces: https://claude.ai/artifacts/latest/a2271a4d-5679-4ab4-ba96-4a43b900b177 — viewable by members in the organization.)

### Impact on existing data
Introduces a wallet balance + usage ledger + pricing config; migrates existing X add-on holders' remaining value into wallet balance and retires the recurring X add-on. Removes the daily-limit enforcement for X posting.

### Impact on other products
Billing/Paddle (new one-off top-up + auto-recharge); the X publish pipeline (deduct hook). Web-first; the raised/charged balance applies wherever X posts are published.

### Dependencies
- **Billing-eng spike (blocking):** confirm the cleanest Paddle mechanism for a non-subscription one-off top-up purchase and how auto-recharge re-triggers it.
- The FE stories consume this story's APIs (balance, rates, queued-cost, auto-recharge state, usage ledger, top-up).

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (emails honor recipient locale)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review (emails use the dynamic app name)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*
- Deduct hook: `contentstudio-backend/app/Libraries/Integrations/Platforms/Social/TwitterPlatform.php` (`postingResponse` — success = `isset($response->id)`), `app/Jobs/PlatformPostingJob.php`.
- Replace the daily-limit gate `PlanHelper::isXPostingLimitReached` (`app/Helpers/Billing/PlanHelper.php`); deduction precedent (make it atomic, unlike) AI credits in `AIController` / `PlanHelper::deduct*Credits`.
- Billing: `app/Services/PaddleBillingService.php`, `config/paddle.php`; new wallet + `x_service_usage` ledger + pricing config; `XCreditWallet`-style service + settlement/auto-recharge logic.
- Emails: pattern in `app/Mail/Accounts/` (e.g. `UpgradeGrowthPlanAutomaticallyMail.php`); honor recipient locale per backend AGENTS.md §9.4.
- Server-side events via the Usermaven SDK / Customer.io pattern used elsewhere.
