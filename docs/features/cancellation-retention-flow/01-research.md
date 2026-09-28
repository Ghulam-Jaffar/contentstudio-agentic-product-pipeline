# Cancellation and retention flow - research

Holds every codebase pointer and gotcha so none of it leaks into the story bodies.

---

## 1. The headline finding: this is a rework, not a new build

**A two-step cancellation dialog already exists, and it already offers a pause.** `CancelRecurringPlanDialog.vue` is 488 lines, has a warning step and a reason step, captures `can_stay`, and its first step is labelled *"pause instead"*.

Everything the epic proposes is a change to a live flow, not a greenfield build. That changes the estimate, the risk, and the regression surface.

| Piece | Status today |
|---|---|
| Cancellation dialog, 2 steps | **Exists** - `CancelRecurringPlanDialog.vue` |
| A second, older dialog | **Exists** - `CancelPlanDialog.vue`, different reason list |
| Reason capture and storage | **Exists** - `CancellationFeedbackService` |
| Pause, end of paid period | **Exists** - both billing stacks |
| Resume | **Exists** - `PaddleBillingHelper::resumeSubscription` |
| Pause for a fixed 1 to 3 months | **Missing** - see 4.2 |
| Offer engine, tiers, guard rails | **Missing** |
| Discount application from the flow | **Missing** |
| Downgrade offer inside the flow | **Missing** |
| Outcome recorded against the account | **Missing** |

---

## 2. Where the detail lives

| Area | Path |
|---|---|
| Live cancellation dialog | `contentstudio-frontend/src/modules/setting/components/billing/dialogs/CancelRecurringPlanDialog.vue` |
| Older cancellation dialog | `contentstudio-frontend/src/modules/setting/components/billing/dialogs/CancelPlanDialog.vue` |
| Plan queries | `contentstudio-frontend/src/modules/setting/queries/usePlanQueries.ts` |
| Billing API client | `contentstudio-frontend/src/api/billing.ts`, `src/api/plan.ts` |
| Cancel and pause endpoints | `contentstudio-backend/app/Http/Controllers/Billing/PaddleUserController.php` |
| Paddle Billing wrapper | `contentstudio-backend/app/Helpers/Billing/PaddleBillingHelper.php` |
| Reason capture | `contentstudio-backend/app/Services/Billing/CancellationFeedbackService.php` |
| Plan and price catalogue | `contentstudio-backend/config/paddle.php` |
| Plan limits | `contentstudio-backend/app/Libraries/Settings/SubscriptionLimits.php` |
| Subscription model | `contentstudio-backend/app/Models/Billing/PaddleBilling/PaddleBillingSubscription.php` |

**Locked prototype:** https://claude.ai/artifact/2RuyPekdyeQheZCdknWfQ8 (version 43)
**Locked spec:** https://claude.ai/code/artifact/0eff1388-eb6c-4b8c-b541-28d889edc0e1

---

## 3. Scope, locked 2026-09-25

Offers are capped and nothing sits outside the cap: **a 20% discount, a pause, a downgrade, a demo, or support.** Verified by enumerating all 5,120 combinations of plan, cycle, usage shape, add-on set, existing-discount state, offer-used state and reason. No combination produces anything outside that set, none produces two discounts, and none comes back empty.

Locked decisions:

- 20% on both cycles, applied to the account total, **kept for the life of the subscription**
- Never offered to an account that already carries a discount on its plan
- One save offer per account per 12 months
- No annual upgrade prompt
- Pause is 1 to 3 months, once per 12 months, and reachable directly from Billing
- Seven cancellation reasons. The platform-change reason was removed
- Enterprise and any account at $400+ counted MRR are offered a demo or support before any automated offer
- **No phone call is offered anywhere in the flow.** Where a conversation is the right answer the user gets **Book a demo** or **Contact support**, and nothing else

---

## 4. The eight things that change how this is built

### 4.1 Two billing stacks, already handled

`PaddleUserController` branches on `$user->paddle_billing` for every operation: Paddle Billing via `PaddleBillingHelper`, and classic Paddle on the v2 API. Cancel, pause and resume each already have both implementations, and `paddle_billing` is set by the Paddle Billing webhook so every new subscription lands on the current stack.

**This is solved and is not a decision for this epic.** New actions follow the same branch as the existing ones.

### 4.2 Pause exists, but it has no duration

`PaddleBillingHelper::pauseSubscription` sends one field:

```php
"effective_from" => "next_billing_period"
```

That is an **indefinite** pause starting at the end of the paid period. Two consequences.

The good one: *effective_from next_billing_period* is exactly the rule the spec settled on, a pause never starts mid-period. That behaviour is already correct and needs no change.

The gap: **there is no `resume_at`**, so 1, 2 and 3 month durations do not exist. Paddle's pause API accepts a scheduled resume, and `resumeSubscription($subscriptionId, $isPauseScheduled)` already has a flag for that case, so the change is passing a date rather than building a mechanism.

The user model already carries `subscription_paused` and `subscription_paused_at`, and `scheduled_change` is stored on the subscription, so a scheduled pause is already representable.

### 4.3 The support ticket already exists, and it is the whole mechanism

`CancellationFeedbackService::submit($request, $user, $type, $flow)` stores `cancellation_reason`, `can_stay`, `question`, `question_details` and `permission_granted`, deduplicated by a sha256 of the payload behind a 60 second lock. It then **emails a Helpin-routed support inbox** through Postmark, subject *Account cancellation!*, body carrying the user id, the contact and the feedback id, with reply-to set to the customer. It refuses to run if the configured inbox is not Helpin-routed, and tracks `delivery_status` so a timeout cannot double-send.

**That is the entire follow-up mechanism and it needs nothing new.** There is no queue, no Slack alert, no service level and no owner. A ticket lands in Helpin and is worked like any other.

But the live reasons are not the seven in the spec. `CancelPlanDialog.vue` offers, verbatim:

- Team not adapting
- Missing features
- I don't see enough value for the price
- Price is too high for us to afford it
- Closing Company/Project/Downsizing
- Already paying for another account
- Could not upgrade/downgrade due to credit card issue

**Two of these are worth keeping and are not in the spec.** *Already paying for another account* and *Could not upgrade or downgrade due to a credit card issue* are both real and both actionable, and the second one is an involuntary-churn signal hiding in a voluntary-churn form.

Changing the list breaks continuity with whatever historical reporting exists on these strings. Map old to new explicitly rather than swapping the list.

### 4.4 There are two cancellation dialogs

`CancelPlanDialog.vue` and `CancelRecurringPlanDialog.vue` both exist, with different reason lists. Establish which is live for which cohort before touching either. Shipping the new flow into one and leaving the other is how two cancellation experiences end up in production.

### 4.5 The offer engine has to be backend

Every rung is decided from tiers, usage and tenure. If that logic sits in the client, a user can set their own value tier or force the discount branch from devtools. The server computes the permitted offers, the client renders what it is given, and the server re-checks on apply.

### 4.6 There is no retention window, and no cancellation email to the customer

An earlier draft of the spec and prototype promised a 90 day retention window, a deletion date on the confirmation screen, a one-click reactivate and two emails. **None of it exists.** Checked against the code on 2026-09-28.

| Claim that was made | What the code does |
|---|---|
| "We will email you the reactivate link now" | **No customer email is sent on cancellation at all.** `cancelUserSubscription` calls `CancellationFeedbackService`, which emails the internal Helpin inbox and nothing else |
| "and once more before your data is deleted" | No such email, and no job that would trigger one |
| "Data kept until <date>, 90 days after your access ends" | No retention window anywhere in the repo. Nothing counts down |
| "Everything is permanently deleted after that date" | **Nothing is ever deleted automatically.** Deletion is user-initiated only |
| "Reactivate: one click, any time before <date>" | There is no deadline, and it is not one click. The customer subscribes again |

What actually happens:

1. `paddle:cancel:accounts` (`PaddleCancellationCommand`) runs on a schedule. Once `cancellation_effective_date` passes and no active subscription remains, it calls `UsersRepository::updateAccountState($userId, 'deactivated')`. That is the whole of it.
2. The customer then lands on `DeletedPlan.vue`, which reads *"Your account is suspended."* and offers three buttons: **Subscribe**, **Contact support**, and **Permanently delete data**.
3. The data sits there indefinitely. The only path to deletion is the customer pressing that third button and confirming.

**So the strongest honest line is the opposite of the one that was written.** Not "your data is kept for 90 days and then destroyed", but "nothing is deleted, and you can come back whenever you like". That is a better retention line anyway, and it has the advantage of being true.

If a win-back email is wanted, it is a new build, not a change. There is nothing to modify.

### 4.7 Pause starts at the end of the paid period. The defect is on the way back, not the way in

**The timing is exactly what you would expect, and nothing changes when the customer clicks pause.**

`PaddleBillingHelper::pauseSubscription` sends `effective_from: next_billing_period`, so Paddle schedules the pause and leaves the subscription `active` with a `scheduled_change` until the paid period runs out. `handleSubscriptionUpdated` mirrors Paddle's status onto `user.state`, so the state stays `active` too. The one thing that changes immediately is `subscription_paused = true` on the user, and that flag is read in exactly one place, to block a plan upgrade.

So between clicking pause and the period ending: **posts publish normally, automations run, analytics work, nothing is locked.** Exactly as the flow describes it.

When the pause does take effect, Paddle fires `subscription.paused`, `user.state` becomes `paused`, and `WorkspacesHelper` locks the workspace: `Account::pauseUserPostings` reverts pending scheduled posts to drafts with `on_hold: true`, and `lockUserWorkspaceAutomations` switches the RSS and Evergreen automations off. That part is defensible. A paused subscription should not be publishing.

**The defect is that coming back does not undo it.** `handleSubscriptionResumed` sets `user.state = 'active'` and stops there. The held posts stay drafts and the automations stay off. The only un-hold path is a manual workspace-level action in `WorkspaceController`, and its automation half is commented out.

| Moment | What happens |
|---|---|
| Customer clicks Pause | Paddle schedules it. Nothing changes. Publishing, automations and analytics all carry on |
| Paid period ends | Workspace locks. Pending posts become drafts, automations switch off |
| Pause ends, billing resumes | State returns to active. **The posts stay drafts and the automations stay off** |

The practical damage is smaller than a lockout but it is still real: a customer who paused for two months comes back to automations silently switched off and anything scheduled past the pause window sitting in drafts. They will not know to go looking, and the automations are the part they will not notice for weeks.

**So the pause rung needs one backend story: restore the held posts and re-enable the automations on resume.** With it, the offer copy is true as written. Without it, the copy cannot say anything is waiting.

One deliberate exception in that story: a post whose slot passed during the pause comes back as a draft, not a late publish. Posting a three week old update to someone's audience on their behalf is its own incident.

### 4.8 Apple IAP subscribers cannot be cancelled here

`docs/features/apple-iap-billing-ui/` exists, so some subscriptions are billed by Apple, and Apple requires those to be cancelled in the App Store. No Paddle discount, pause or downgrade applies. **The flow needs a guard at the top** that detects an IAP subscription and shows Apple's instructions instead. Not in the spec, not in the prototype.

---

## 5. What fits with no change

- Pause starting at the end of the paid period, both stacks
- Resume, including the scheduled-resume flag
- Reason storage, deduplication and the feedback email
- `previewSubscriptionUpdate` and `updateSubscription`, which the downgrade rung needs
- `calculateAddonsPrice`, for showing what a downgrade costs
- Helpin chat, opened app-wide by the `open-help-widget` EventBus event

---

## 6. Plan and price facts

| Plan | Monthly | Annual | Accounts | Workspaces | Users | Tier below |
|---|---|---|---|---|---|---|
| Standard | $29 | $228 | 5 | 1 | 1 | none |
| Advanced | $69 | $588 | 10 | 2 | 2 | Standard |
| Agency Unlimited | $139 | $1,188 | 25 | unlimited | unlimited | Advanced |
| API | $15 | $180 | 1 to 50 | 1 | 1 | none |
| Enterprise | custom | custom | custom | custom | custom | Agency |

The API plan costs $180 a year on either cycle, so there is no annual saving to offer it.

Eighteen add-ons exist in `config/paddle.php`, all recurring. **Only `social_accounts_addon`, `saml_sso_addon`, `white_label_addon` and `social_listening_addon` count toward the value tier.** Workspace and team member add-ons are ignored because both are unlimited on Agency. Credit packs are ignored because consumption is not commitment. `blog_addon` is excluded, blog publishing is sunset.

Four discount codes already exist: `AgencyUnlimitedOFF`, `AdvancedOFF`, `StandardOFF`, `ApiAnnualOFF`. Their depth is configured in Paddle and cannot be read from the repo.

---

## 7. Open questions to answer before estimating

1. Which of the two cancellation dialogs is live, and for whom?
2. Are the four Paddle discount codes set to 20%, and should the flow reuse them or get its own for reporting?
3. Does pausing suspend the jobs that refresh social tokens? Meta long-lived tokens run about 60 days, so a 3 month pause could return a customer to disconnected accounts.
4. Should the two live reasons the spec dropped be kept: *already paying for another account*, and *credit card issue*?
5. Can a team member without billing permission reach Cancel plan at all?
6. Which booking page does **Book a demo** open? Nothing in the product owns one today, so the link, the calendar behind it and who picks the booking up all have to be settled before the demo rung can ship.
