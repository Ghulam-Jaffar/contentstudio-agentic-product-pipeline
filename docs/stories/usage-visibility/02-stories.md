# Epic + stories — Usage visibility

**Scope of this doc:** the epic and **2 stories**, both backend: one for AI usage, one for X usage. Both report the full breakdown to Usermaven and to internal admin. Nothing is pushed to any tracker. The PO creates this by hand.

---

## Epic: Usage visibility

ContentStudio meters a growing list of features. AI text, image and video generation, video clips, AI auto-replies, API calls, social listening mentions, and X posting are all consumed against an allowance or a balance, and every one of them costs us money per use. X posting is already pay-per-use, and X inbox, X analytics and competitor analytics are heading the same way.

We cannot see any of it properly.

The reason is how consumption is recorded. For AI and most allowances it is a running counter on the workspace: a number that goes up until it hits the plan limit, then gets reset. That is enough to decide whether someone may generate one more image, and useless for anything else. There is no record of what was generated, by whom, with which model, or when, and the history is destroyed on every reset. So nobody can say which plans consume more than they pay for, which accounts are loss-making, which models are eating the budget, or which surface the usage came from.

Usermaven has the same blind spot, for a different reason. We do already send usage-related events from the backend, but every one of them is an exception or a money-in event: a user blocked for insufficient balance, a spending limit hit, a top-up purchased, auto-recharge fired. Not one event fires when someone actually spends something. So our product analytics can tell us when a customer was stopped, and never what they consumed.

There is also a piece of luck worth taking. The AI platform already calculates what each generation costs us and what we charge for it, and already sends both numbers on every usage event. The main product receives those events and drops both. The single most valuable number for understanding our unit economics is being produced today and thrown away.

This epic makes usage visible, generically. Every metered action produces a full, permanent record, that record goes to internal admin with the complete breakdown including what it cost us, and a compact version goes to Usermaven so consumption can be sliced alongside every other product signal we already track. It is built as one mechanism, not one per feature, so X inbox, X analytics and competitor analytics plug into it rather than each needing their own instrumentation.

Two stories start it: AI usage and X usage. The rest of the metered features follow the same pattern once these two prove it.

### Why two destinations, and what goes to each

They answer different questions and have different constraints.

- **Internal admin** gets the complete record: every dimension, plus what we paid the provider, what we charged, and the plan in force at the time. This is where margin, loss-making accounts and model mix are answered.
- **Usermaven** gets one compact event per action, carrying the dimensions needed to slice consumption against everything else we already track there. Payloads stay within our event conventions, and provider cost and margin are deliberately not sent to a third-party analytics processor.

### Out of scope

- Changing prices, plan limits, or what any plan includes. This epic measures. Pricing decisions come from what it shows.
- Changing how credits or balances are enforced. Existing limits and blocks keep working exactly as they do now.
- Building customer-facing usage screens. That comes after these two stories, from what they record.
- Rebuilding the X wallet ledger. It feeds this rather than being replaced.

### Stories

1. `[BE] Report full AI usage and credit consumption to Usermaven and internal admin`
2. `[BE] Report full X usage and spend to Usermaven and internal admin`

### Sequencing

Either can go first. The AI story is the larger of the two because nothing exists today, and it is where the discarded cost data is recovered. The X story is smaller because a transaction ledger already exists and the gap is the missing spend event and the cost side. Whichever goes first establishes the shared record shape and the internal admin ingestion path, and the second reuses both.

---

## [BE] Report full AI usage and credit consumption to Usermaven and internal admin

### Description

As the business, we want every AI action a customer takes recorded in full and reported to both internal admin and Usermaven, so that we can see who consumed what, on which plan, with which model, from which surface, and what it cost us, instead of reading a counter that tells us only how much of an allowance is left.

### Workflow

1. A user performs an AI action: generating a caption, an image, a video, a video clip, an inbox auto-reply, or an AI call through the public API.
2. When the action completes, a full usage record is written: who did it, in which workspace, which feature, which model, how much was consumed, what we charged, what it cost us, which plan was in force, and which surface the request came from.
3. That record is delivered to internal admin, where the team can break usage and cost down by user, workspace, account, plan, feature, model, surface and period.
4. A compact event is sent to Usermaven for the same action, so AI consumption can be sliced alongside the product signals already tracked there.
5. Neither delivery slows the user's action or can fail it. If a destination is unavailable, the record is still kept and delivered later.
6. Existing credit counters, plan limits and blocking behavior continue exactly as before.

### Acceptance criteria

**Recording**

- [ ] Every completed AI action writes one usage record, covering AI caption and text generation, AI image generation, AI video generation, video clips, AI inbox auto-replies, and AI usage through the public API.
- [ ] Each record contains at minimum: user, workspace, account where relevant, feature, model where relevant, quantity consumed, credits charged, provider cost, retail price, credit type, the plan in force at that moment, the surface the request came from, the outcome, and the time.
- [ ] The provider cost and retail price already produced upstream by the AI platform are stored rather than discarded, verified against a real generation of each kind. Where an action's cost is genuinely unknown, it is recorded as unknown, never as zero.
- [ ] The plan in force is stored on the record. Changing a customer's plan later does not alter any historical record.
- [ ] The surface is recorded and distinguishes web, public API, mobile app, and an AI assistant acting on the user's behalf.
- [ ] Records survive the periodic credit reset. Resetting a customer's credits does not remove or alter history.
- [ ] A failed AI action that still cost us money is recorded, and is distinguishable from one that succeeded and from one that cost nothing.
- [ ] One record is written per action. A retry, a partial failure, or a duplicate upstream event does not produce a second record, verified with a test.
- [ ] Records use a shared shape that a new metered feature can adopt without a new record type.

**Internal admin**

- [ ] The full record, including provider cost, retail price and the plan in force, is delivered to internal admin.
- [ ] Internal admin can break AI usage and cost down by user, workspace, account, plan, feature, model, surface and time period, and can identify loss-making accounts and plans.
- [ ] Where internal admin lives, who owns it, and how it ingests this data is confirmed and documented as part of this story. This must be resolved, not deferred.
- [ ] A recurring reconciliation compares the delivered records against the existing credit counters and reports any disagreement to the team.

**Usermaven**

- [ ] A Usermaven event fires for every successful metered AI action: `ai_credits_consumed`, with `{ feature, model, credit_type, credits, plan, workspace_id }`.
- [ ] A Usermaven event fires when a user is blocked by their AI allowance: `ai_credits_exhausted`, with `{ feature, credit_type, plan, workspace_id }`.
- [ ] Event and property names are `snake_case` and follow the existing server-side event conventions, and the ticket confirms which dimensions Usermaven already holds from identify so they are not duplicated in the payload.
- [ ] No provider cost, retail price or margin figure is sent to Usermaven.
- [ ] No personally identifying data and no generated content, prompt text or caption text is sent to Usermaven.
- [ ] Both events are verified arriving in Usermaven with the correct payload for every covered feature.

**Safety**

- [ ] Reporting to either destination is dispatched asynchronously and never blocks or slows the user's AI action. The user-perceived time to complete a generation does not regress, and this is measured rather than assumed.
- [ ] A failure or timeout at either destination never fails the user's action and never double-charges credits.
- [ ] If a destination is unavailable, records are retained and delivered once it recovers, with no silent loss.
- [ ] Delivery failures are visible to the team as an alert, not only in logs.
- [ ] Existing credit counters, plan limits, blocking and all existing usage indicators in the product behave exactly as before and show the same figures.
- [ ] The agreed retention behavior is implemented, including whatever aggregate is kept after per-event data expires, following the approach already agreed for analytics data.

### Mock-ups

None. Backend-only story. Internal admin views are built on this data separately.

### Impact on existing data

Introduces a new usage record. No existing data is changed or deleted, and the existing credit counters stay exactly as they are. Where existing records allow a partial backfill of history, that is done and the coverage achieved is stated, so the first internal views are not empty. Retention is applied as agreed.

### Impact on other products

AI usage originating from the mobile app, the public API and AI assistants must all be recorded with the surface identified. No behavior changes for any of those clients, and the story must verify that.

### Dependencies

- Shares the record shape and the internal admin ingestion path with `[BE] Report full X usage and spend to Usermaven and internal admin`. Whichever ships first establishes both.
- Should follow the retention approach already agreed for analytics data.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support — N/A, no user-facing copy in this story
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review (white-label usage must be attributed correctly, and the story must state whether it is attributed to the reseller or the end customer)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## [BE] Report full X usage and spend to Usermaven and internal admin

### Description

As the business, we want every X action that consumes a customer's balance recorded in full and reported to both internal admin and Usermaven, so that we can see what our customers actually spend on X, across both the wallet and legacy credit cohorts, and what our margin on it is, instead of only seeing the moments a customer was blocked or topped up.

### Workflow

1. A user publishes to X, or uses another metered X feature as they come online.
2. When the spend is taken, a full usage record is written: who spent it, in which workspace, on which account, which X feature, what kind of post, how much was charged, what X charged us, which cohort the customer is in, which plan was in force, and the balance left afterwards.
3. That record is delivered to internal admin, where the team can break X spend and margin down by user, workspace, account, plan, feature, cohort and period.
4. A compact event is sent to Usermaven for the same spend, so X consumption can be sliced alongside the existing X events for blocks, top-ups and auto-recharge.
5. Neither delivery slows publishing or can fail a post.
6. Existing wallet behavior is untouched: balance checks, spending limits, blocks, auto-recharge and the existing transaction ledger all work exactly as before.

### Acceptance criteria

**Recording**

- [ ] Every X action that consumes a customer's balance writes one usage record, for both the wallet cohort and the legacy credit cohort.
- [ ] Each record contains at minimum: user, workspace, the X account, the X feature, the post type where relevant, the amount charged in its native unit, the provider cost, the cohort, the plan in force at that moment, the surface the request came from, the balance remaining afterwards, the outcome, and the time.
- [ ] The record carries what X charged us as well as what we charged the customer, so margin on X is derivable from the record alone.
- [ ] Cohort is recorded on every row, and a dollar balance is never presented or aggregated as if it were the same unit as a legacy credit count.
- [ ] For cross-cohort analysis, a normalized dollar amount is recorded alongside the native unit, and the conversion rate used is stored on the record so the figure is auditable.
- [ ] The plan in force is stored on the record. Changing a customer's plan later does not alter any historical record.
- [ ] The record shape covers X features beyond posting, so X inbox, X analytics and competitor analytics can be reported through it without a new record type.
- [ ] A failed publish that still cost us money is recorded, and is distinguishable from one that succeeded and from one that was blocked before any spend.
- [ ] One record is written per spend. A retry or a duplicate event does not produce a second record, verified with a test.
- [ ] Existing X wallet transactions are represented in the same view, so history is not split across two places.

**Internal admin**

- [ ] The full record, including provider cost and the plan in force, is delivered to internal admin.
- [ ] Internal admin can break X spend and margin down by user, workspace, account, plan, X feature, post type, cohort, surface and time period, and can identify loss-making accounts and plans.
- [ ] Where internal admin lives, who owns it, and how it ingests this data is confirmed and documented as part of this story, reusing whatever the AI usage story established if that shipped first.
- [ ] A recurring reconciliation compares the delivered records against the existing wallet ledger and the legacy credit counters, and reports any disagreement to the team.

**Usermaven**

- [ ] A Usermaven event fires on every successful X spend: `x_credits_spent`, with `{ feature, post_type, cost_cents, cohort, plan, workspace_id }`, where `cost_cents` is the normalized dollar amount for both cohorts.
- [ ] The event name pairs with the existing `x_credits_purchased` event, and the existing X events for blocks, spending limits, auto-recharge and purchases keep firing unchanged with their current names and payloads.
- [ ] Event and property names are `snake_case` and follow the existing server-side event conventions, and the ticket confirms which dimensions Usermaven already holds from identify so they are not duplicated in the payload.
- [ ] No provider cost or margin figure is sent to Usermaven.
- [ ] No personally identifying data and no post content is sent to Usermaven.
- [ ] The event is verified arriving in Usermaven with the correct payload for a plain post, a post with a link, a wallet-cohort customer and a legacy-cohort customer.

**Safety**

- [ ] Reporting to either destination is dispatched asynchronously and never blocks or slows publishing. Time to publish does not regress, and this is measured rather than assumed.
- [ ] A failure or timeout at either destination never fails a post, never blocks a publish, and never double-charges a balance.
- [ ] If a destination is unavailable, records are retained and delivered once it recovers, with no silent loss.
- [ ] Delivery failures are visible to the team as an alert, not only in logs.
- [ ] Existing wallet behavior is unchanged: balance checks, monthly spending limits, blocks, auto-recharge, the transaction ledger, the customer-facing usage list and its CSV export all behave exactly as before and show the same figures.
- [ ] The agreed retention behavior is implemented, following the approach already agreed for analytics data.

### Mock-ups

None. Backend-only story. Internal admin views are built on this data separately.

### Impact on existing data

Introduces a new usage record and adds provider cost to what is captured per spend. The existing wallet ledger, balances and legacy credit counters are not changed or migrated. Where the existing ledger allows a partial backfill of history, that is done and the coverage achieved is stated. Provider cost cannot be backfilled for past spends unless a rate for that period is known, and the story must state what was and was not recoverable rather than inferring it.

### Impact on other products

X publishing happens from the web app, the mobile app, the public API and AI assistants. All of them must be recorded with the surface identified, and none of their behavior changes.

### Dependencies

- Shares the record shape and the internal admin ingestion path with `[BE] Report full AI usage and credit consumption to Usermaven and internal admin`. Whichever ships first establishes both.
- Builds on the existing X pay-per-use wallet and X metering work rather than replacing it.

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support — N/A, no user-facing copy in this story
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review (white-label X spend must be attributed correctly, and the story must state whether it is attributed to the reseller or the end customer)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
