# PRD: X (Twitter) Analytics Pay-Per-Use Metering + Wallet Gating

**Author:** Ghulam Jaffar (Product Owner)
**Last Updated:** 2026-08-03
**Status:** In Review
**Target Release:** Q4 2026 (after the X Pay-Per-Use Credit Wallet ships)

---

## 1. Overview

X (Twitter) moved its API to **pay-per-use pricing**, so every X analytics refresh now costs ContentStudio real money: each sync reads one user profile (~$0.010) plus the tweets the user asked for at **$0.005 per tweet read**. This feature meters X analytics against the **existing prepaid X Credit Wallet** (the same dollar balance used for X publishing): X analytics is gated behind a one-time **per-workspace unlock card**, every sync shows its **dollar cost before it runs** and updates live as the user changes the tweet count, and the **actual cost is deducted from the wallet on a successful fetch** and recorded in the shared usage log. Billing-capable admins turn the capability on with an **Enable X analytics** toggle on the X Wallet card (off by default). The whole feature is built around one principle: **the user always sees what a sync costs before they spend, and can never be surprise-billed.**

---

## 2. Problem Statement

**What problem are we solving?**
X's per-read API pricing means X analytics is no longer free to serve: a single 30-tweet sync costs ContentStudio about $0.16 in raw X fees, and a heavy 150-tweet sync on a schedule several times a week multiplies fast across thousands of connected accounts. Today X analytics syncs run with no cost ceiling and no user-visible price. There is no way to recover the cost, no way for a user to see what a refresh will cost, and no guard against runaway spend. The analytics module already fetches an internal `credits_used` count per sync but has no prepaid balance, no pre-sync price, and no gate.

**Who has this problem?**
Two parties. **ContentStudio** carries uncapped, growing X API cost on every analytics sync. **Users** who rely on X analytics need it to keep working, but (post X-pricing news) now expect that X data may cost extra and want to see and control that cost rather than have it hidden or the feature quietly removed.

**What happens if we don't solve it?**
ContentStudio either eats unbounded X analytics cost (margin erosion that scales with usage) or is forced to throttle or drop X analytics (as Later and Loomly did), frustrating paying customers. A transparent, wallet-metered model recovers cost with margin while keeping X analytics available and, uniquely in this market, showing users exactly what each sync costs before they spend.

---

## 3. Goals & Success Metrics

| Goal | Metric | Target | How We'll Measure |
|---|---|---|---|
| Recover X analytics API cost with margin | Gross margin on X analytics syncs | ≥ 20% markup (per wallet pricing config) | Usage ledger + pricing config |
| Keep X analytics adopted after gating | % of X-analytics workspaces that unlock and run ≥ 1 paid sync | ≥ 60% of previously-active X-analytics workspaces in 90 days | Usermaven (`x_analytics_unlocked`, `x_analytics_sync_charged`) |
| No bill shock | X-analytics-related billing support tickets | < 1% of X-analytics workspaces | Support data |
| Cost visibility works | % of blocked syncs that convert to a top-up within 24h | ≥ 40% | Usermaven (`x_analytics_sync_blocked_insufficient_balance` → `x_credits_purchased`) |
| Guard rail: scheduled spend stays intentional | Auto-paused scheduled syncs that resume after top-up | ≥ 50% resume within 7 days | Usermaven (`x_analytics_schedule_paused_insufficient_balance`) |

### 3.1 Analytics Events (Usermaven)

New events for this feature. Top-up itself already emits `x_credits_purchased` from the wallet feature and is **not** duplicated here. Before implementing, search `contentstudio-frontend/src/` for `userMaven.track(` to confirm none of these names already exist.

| Event Name | Trigger | Payload | What we measure with it |
|---|---|---|---|
| `x_analytics_enabled` | Billing user turns on the Enable X analytics toggle | `{}` | Account-level adoption of the capability |
| `x_analytics_unlocked` | User confirms the in-page unlock/consent card in a workspace (FE) | `{}` | Per-workspace activation; unlock funnel |
| `x_analytics_sync_charged` | A successful sync deducts from the wallet (BE, server-side on fetch completion) | `{ amount_usd, tweet_count, source: 'manual' \| 'scheduled' }` | Sync volume, spend, manual vs scheduled split |
| `x_analytics_sync_blocked_insufficient_balance` | A manual sync is attempted with too-low balance (FE) | `{}` | Friction; top-up conversion trigger |
| `x_analytics_schedule_paused_insufficient_balance` | A scheduled run is skipped and its schedule paused (BE) | `{}` | Scheduled-spend friction; resume behavior |

Naming follows guidelines §19 (snake_case, past tense, no PII). These map 1:1 to acceptance criteria in the FE/BE stories.

---

## 4. Target Users

**Primary Persona — Analyst / social media manager who views X analytics.** Opens the X analytics dashboard, runs manual syncs, and sets up scheduled refreshes. Cares about fresh, complete X data and about knowing what a refresh costs before running it. May or may not have billing access.

**Secondary Persona — Workspace super admin / billing owner.** Owns the X wallet, turns X analytics metering on, tops up, sets auto-recharge and the monthly spending limit, and watches spend. Cares about predictable cost and not overpaying.

**Non-Users (out of scope):**
- Users on platforms other than X (this is X-analytics-only; other platforms' analytics stay unmetered).
- Mobile app users (X metering is web-first per the wallet PRD; no Flutter/iOS/Android work in this epic).
- X **publishing**, **inbox**, and **listening** metering (publishing is already handled by the wallet feature; inbox and listening are future epics on the same wallet).

---

## 5. User Stories / Jobs to Be Done

| ID | As a… | I want to… | So that… | Priority |
|---|---|---|---|---|
| US-1 | Analyst | see what an X analytics sync will cost before I run it | I am never surprised by the charge | Must |
| US-2 | Analyst | have the cost update as I change the tweet count | I can choose a cheaper or fuller sync knowingly | Must |
| US-3 | Analyst | be told clearly when the wallet cannot cover a sync, and who to ask | I know why it did not run and how to fix it | Must |
| US-4 | Super admin | turn X analytics metering on or off for the account | I control whether we spend on X analytics at all | Must |
| US-5 | Super admin | see X analytics syncs in the same wallet usage log as publishing | I have one place to audit all X spend | Must |
| US-6 | Analyst | know that a scheduled X sync will pause (not silently fail forever) if the wallet runs dry, and be notified | my automated refreshes stay under control | Must |
| US-7 | Analyst without billing access | understand X analytics is metered and who can fund it | I am not stuck when the wallet is empty | Must |
| US-8 | Super admin | see the projected recurring cost when I set a scheduled X sync | I can budget the automation | Should |

---

## 6. Requirements

### 6.1 Must Have (P0)

- **Per-workspace unlock / consent card** on the X analytics tab, shown over a blurred demo dashboard, explaining that refreshing X analytics draws from the shared X wallet, the per-sync dollar cost, and a learn-more link. Includes the "enable required" and "ask your super admin" states.
- **Account master enable**: an **Enable X analytics** toggle on the billing X Wallet card, actionable only by billing-capable users (`can_see_subscription` / super admin), **off by default**.
- **Live per-sync cost preview in dollars** on the Sync Data button and inside the sync settings modal, **recalculating as the tweet count changes** (`10-150`, default 30).
- **Balance gate on manual syncs**: when the wallet (and auto-recharge) cannot cover the sync, block it, deduct nothing, and show a top-up CTA (billing users) or an "ask your super admin" message (non-billing users).
- **Deduct on successful fetch, charged on the actual tweets read** (not the estimate, and nothing on failure or empty result), recorded in the shared X wallet usage ledger as an **analytics-sync** consumption type.
- **Scheduled sync metering**: show the projected per-run cost in the scheduled sync settings modal; when a scheduled run cannot be funded, **skip it, pause the schedule, and notify** the user; resume on top-up or successful auto-recharge.
- **Always-visible X wallet balance** in the X analytics surface, with a top-up shortcut for billing users.
- **Pricing from the wallet's config-driven rate** (upstream cost + markup); **never expose X's raw cost or the markup** (BR-12 from the wallet PRD).
- **Usermaven events** per §3.1.

### 6.2 Should Have (P1)

- **Estimated-vs-actual reconciliation**: after a sync, show the actual amount charged (from tweets actually returned) alongside the estimate, so the number the user saw is visibly honored.
- **Projected recurring spend summary** for a scheduled sync (for example, "about $0.16 per run, roughly $4.80 per month at daily").
- **Low-balance surfacing inside the analytics view** (not only in billing), so an analyst sees the wallet is running low before a sync is blocked.

### 6.3 Nice to Have (P2)

- **Threshold-based confirm dialog** for unusually large syncs (only above a configurable dollar amount).
- **Free reuse of very recently synced data** (offer cached data at no charge when it is still fresh, instead of re-charging).
- **Cheaper-path nudge** (for example, suggesting a smaller tweet count when the balance is low).

### 6.4 Explicitly Out of Scope

- X **publishing** metering (already covered by the X Pay-Per-Use Credit Wallet epic).
- X **inbox** and **listening** metering (future epics on the same wallet).
- **Date-range-based** X analytics (X analytics is count-based; the API takes no start/end time, so there is no date range to price).
- **Mobile** app changes (web-first for v1).
- A **second currency** or separate analytics credit balance (analytics spends from the one X dollar wallet).
- Building the wallet, top-up modal, auto-recharge, or usage ledger (reused from the wallet epic).

---

## 7. User Flow (High Level)

1. A user opens **Analytics** and selects the **X (Twitter)** tab. If the workspace has not unlocked X analytics, the **unlock card** appears over a blurred demo dashboard, explaining the pay-per-use model, the per-sync cost, and a learn-more link.
2. The user clicks **Unlock and fetch data**. If the account has not enabled X analytics, a billing user can enable it here (non-billing users are told to ask their super admin).
3. If the wallet has enough balance, ContentStudio runs the first sync, renders the dashboard, **deducts the actual cost**, and updates the balance. If not, the top-up state appears.
4. For later refreshes, the user clicks **Sync data**; the sync modal shows the tweet-count dropdown and a **live dollar estimate**, and the button reads the cost. On success the cost is deducted and logged.
5. For hands-off refreshes, the user sets a **scheduled sync** (frequency + tweet count) and sees the **projected per-run cost**. Scheduled runs consume automatically while funded; an unfunded run pauses the schedule and notifies the user.

```mermaid
flowchart TD
    Start([User opens Analytics then the X tab]) --> Unlocked{X analytics unlocked for this workspace}
    Unlocked -->|No| Card[Show unlock card with per-sync cost explainer and help link]
    Card --> Consent[User clicks Unlock and fetch data]
    Unlocked -->|Yes| Enough{X wallet balance covers this sync}
    Consent --> Enough
    Enough -->|Yes| Fetch[Render analytics and run the sync]
    Enough -->|No| TopUp[Show top up needed state]
    TopUp -->|Billing access| WalletModal[Open Manage X Wallet top up modal]
    TopUp -->|No billing access| Ask[Show ask your super admin message]
    WalletModal --> Enough
    Fetch --> Deduct[Deduct actual sync cost from X wallet and log usage]
    Deduct --> Done([Dashboard shown with updated balance])
```

A second diagram covering the scheduled-sync auto-pause behavior is in `02-workflow.md` §4.

---

## 8. Business Rules & Constraints

| Rule ID | Rule | Rationale |
|---|---|---|
| BR-1 | Cost per sync = X cost × the wallet's configured markup, where X cost = (tweets read × per-post-read rate) + one user-read rate. Read from the wallet pricing config, not hardcoded. | Recover cost with margin; no deploy when X changes prices |
| BR-2 | Deduct only on a **successful** fetch, charged on the **actual** tweets returned; nothing on failure or empty result | Never charge for a failed or empty sync |
| BR-3 | Spend comes from the **account/super-admin-level X wallet**, shared with publishing; analytics posts a distinct **analytics-sync** entry to the same usage ledger | One balance, one history, clearer audit |
| BR-4 | Deduction is **atomic** (atomic decrement + ledger insert, idempotent per sync) | Concurrent manual + scheduled syncs must not double-charge or lose a deduction |
| BR-5 | Insufficient balance → the sync **does not run** (manual) or the scheduled run is **skipped and the schedule paused** (scheduled); not held or silently retried forever | Predictable behavior, no silent repeated failures |
| BR-6 | X analytics metering is **off by default** per account; a billing-capable user must enable it (via the billing toggle or by unlocking from the card) | Explicit opt-in; no spend before an admin decides |
| BR-7 | Only **billing-capable users** (`can_see_subscription` / super admin) can enable metering or top up; others see "ask your super admin" | Mirrors the wallet and white-label permission pattern |
| BR-8 | Unlock/consent is recorded **per workspace**; spend always draws from the single account-level wallet | Matches per-workspace analytics navigation with centralized billing |
| BR-9 | The pre-sync figure is an **estimate**; the charge is the **actual**. X analytics is **count-based** (no date range) | X API takes tweet count, not a time window |
| BR-10 | **Never** expose X's raw per-read cost or ContentStudio's markup; show only the price the user pays | Protects margin; avoids "why the fee?" friction (wallet BR-12) |

---

## 9. Open Questions

| Question | Options | Owner | Decision |
|---|---|---|---|
| Does the same markup % apply to analytics reads as to publishing? | Same flat markup / analytics-specific | PM + Billing | Working: same config-driven markup |
| Estimated-vs-actual: single "about $X" or a range when requested tweets may exceed available? | Single estimate + reconcile / range | PM | Leaning single estimate + reconcile (P1) |
| Should the existing internal `credits_used` value be shown to users at all, or fully replaced by the dollar cost? | Replace fully / keep as secondary detail | PM | Leaning replace fully with dollars |
| Does an auto-paused scheduled sync require a manual "resume," or does a successful top-up/auto-recharge resume it automatically? | Manual resume / auto-resume | PM + BE | **Decided: auto-resume** on the next due run once the wallet is funded (top-up or auto-recharge) |
| Threshold-confirm dialog and free fresh-data reuse — v1 or v2? | v1 / v2 | PM | Deferred to v2 (§6.3) |

---

## 10. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Users perceive X analytics as "newly paywalled" and churn | Medium | High | Unlock card frames the X pay-per-use context (not a ContentStudio choice); transparent per-sync cost; publishing wallet already normalizes the model |
| Scheduled syncs silently drain the wallet | Medium | High | Projected per-run cost shown at setup; auto-pause + notify when unfunded; auto-recharge with monthly spending limit |
| Estimate and actual charge diverge and erode trust | Medium | Medium | Charge on actual tweets read, reconcile visibly (P1), never charge more than shown per tweet rate |
| Two credit concepts (posting credits vs analytics) confuse users | Medium | Medium | Retire the analytics "credits used" spend unit; one dollar wallet; consistent copy |
| Concurrency double-charge across manual + scheduled | Medium | Medium | Atomic decrement + idempotent per-sync deduction + ledger |
| Wallet epic slips, blocking this feature | Medium | High | Sequence explicitly after the wallet; build against its ledger/pricing/permission contracts |

---

## 11. Dependencies

- **Internal:**
  - **X Pay-Per-Use Credit Wallet** (`docs/features/x-pay-per-use-credits/`) — the balance, atomic deduction + usage ledger, config-driven pricing, Manage X Wallet top-up modal, auto-recharge, monthly spending limit, and `can_see_subscription` permission gating. **Hard dependency; must land first.**
  - **Analytics module** (`contentstudio-frontend/src/modules/analytics/**`) — the X analytics view, the Sync Data button and `SyncDateRangeModal.vue` (twitter mode), the `TwitterJobSettingsModal.vue`, and the existing `credits_used` readout being replaced.
  - **Billing UI** (`contentstudio-frontend/src/modules/setting/components/billing/**`) — where the Enable X analytics toggle lives on the X Wallet card.
  - **Analytics data pipeline** (`contentstudio-social-analytics-go`) — the `twitter-fetcher` and manual `triggerJob` paths, which must deduct on a successful fetch and report the actual tweet count read (`src/services/twitter/twitter-fetcher/run.go`, `src/api/immediate_work_apis.go`, `src/cmd/jobs/fetcher/twitter.go`).
  - **Feature-unlock precedent** — `CompetitorAnalyticsLanding.vue`, `CompetitorUpgradeModal.vue`, `CompetitorDummyGraphs.vue` for the in-page gate.
- **External:** X (Twitter) API pay-per-use pricing ($0.005 per post read, ~$0.010 per user read); Paddle top-up mechanism (via the wallet).
- **Blockers:** The X wallet's balance getter and deduct/consume endpoint must exist (backend foundation on the unmerged `feature/cont-2623-story-1-credit-consumption-engine-foundation` branch) before this feature can deduct.

---

## 12. Appendix

- **Workflow & full UX detail:** `02-workflow.md` (this feature folder).
- **Research (competitor + codebase):** `01-research.md` (this feature folder).
- **Wallet feature it extends:** `docs/features/x-pay-per-use-credits/` (PRD, workflow, epic + stories).
- **Codebase anchors:** frontend `src/modules/analytics/views/twitter/**`, `src/modules/analytics/components/common/SyncDateRangeModal.vue`, `src/modules/analytics/views/twitter/components/TwitterJobSettingsModal.vue`, `src/modules/setting/components/billing/**`; Go `src/clients/social/twitter.go`, `src/services/twitter/twitter-fetcher/run.go`, `src/cmd/jobs/fetcher/twitter.go`.

---

## Changelog

| Date | Author | Changes |
|---|---|---|
| 2026-08-03 | Ghulam Jaffar | Initial PRD from approved research + workflow; cost in dollars, per-workspace unlock, scheduled auto-pause, master toggle off by default |
