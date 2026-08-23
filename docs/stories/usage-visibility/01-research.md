# 01 Research — Usage visibility

Date: 2026-08-23
Scope: how consumption of every metered feature is recorded today, what reaches Usermaven, what reaches an internal admin surface, and what is missing before anyone can answer "who used what, how much, on which plan, and what did it cost us".

This doc is the local grounding for the epic. Nothing here goes into the story body.

---

## 1. The metered surface today

Every metered currency found across `contentstudio-backend`:

| Currency | Feature |
|---|---|
| `caption_generation_credit` | AI text / caption generation |
| `image_generation_credit` | AI image generation |
| `video_generation_credits` | AI video generation |
| `video_clip_credits` | Video clips |
| `ai_auto_reply_credits` | AI inbox auto-replies |
| `api_credit` | Public API usage |
| `x_posting_credits` | X posting, legacy cohort |
| X dollar wallet | X posting, wallet cohort |
| `used_mention_credits` | Social listening mentions |

Planned next, per the brief: X inbox, X analytics, competitor analytics. All pay-per-use, all needing the same visibility.

## 2. How AI consumption is recorded: counters, not events

From `app/Helpers/Billing/PlanHelper.php`, consumption is a **running counter on the workspace document**, clamped to the plan limit:

- `caption_generation_credit` → exposed as `used_text_credits`
- `image_generation_credit` → exposed as `used_image_credits`
- `used_video_credits`, `used_video_clip_credits`, `used_ai_auto_reply_credits`
- availability computed as `limit - used`, limits from `SubscriptionLimits::getSubscriptionLimits($workspace_id)`
- deduction via `deductImageCredits()`, `deductTextAndImageCredits()`, clamping to the plan total
- reset by `app/Console/Commands/Billing/ResetUsedCreditsCommand.php`

A counter answers "may this user generate one more image". It cannot answer anything else. There is no record of what was generated, by whom, with which model, or when, and the history is destroyed on reset.

The one partial exception is `app/Models/AiChatUsage.php` writing to `ai_chat_usage`: `run_id`, `session_id`, `user_id`, `workspace_id`, `outcome`, `items`, `credits_charged`, `text_credits`, `image_credits`, `billed_at`. That is a **billing claim row for exactly-once charging of chat turns**, not a usage ledger. Image generation outside chat, video jobs, video clips, auto-replies and API credits have no equivalent.

## 3. Cost data is produced upstream and discarded

`contentstudio-ai-agents` already knows what everything costs:

- `src/utils/model_registry.py` carries per-model `api_cost`, keyed by resolution and by audio on or off for the video models, and per-second or per-minute rates for others. Costs vary by more than an order of magnitude between models.
- `src/events/schemas.py` defines, on the usage events themselves:
  - `api_cost` — "Provider API cost in USD (for internal tracking)"
  - `retail_cost` — "Retail cost in USD (api_cost * markup, for internal tracking)"
  - `credits_consumed`
- Topics: `ai-agents.chat.usage`, `ai-agents.video.job.lifecycle`, `ai-agents.video.batch.lifecycle`, `ai-agents.video.transform.lifecycle`, `ai-agents.reels.lifecycle`.

`app/Kafka/Handlers/ChatUsageEventHandler.php` and `VideoJobEventHandler.php` consume these. **A grep for `api_cost` and `retail_cost` across `contentstudio-backend/app` returns zero hits.** The producer computes the two numbers that make margin measurable, sends them, and the consumer drops them both.

## 4. What reaches Usermaven today

There **is** a server-side path: `app/Helpers/Integrations/CustomEventHelper.php`

```php
public static function SendEventToUsermaven(string $eventType, ?string $userId, array $eventAttributes = [], array $additionalData = []): bool
```

It does a plain `Http::post()` to the Usermaven event API, returns a bool, and swallows failures to Sentry. Also `sendTransactionEventToUsermaven()`, and `SendEventToCustomerIo()` alongside it.

Every server-side call site:

| Event | Where |
|---|---|
| `x_post_blocked_insufficient_balance` | `XWalletDebitService.php:57` |
| `x_post_blocked_spending_limit` | `XWalletDebitService.php:57` |
| `x_credits_purchased` | `WalletWebhookHandler.php:99` |
| `x_auto_recharge_failed` | `WalletWebhookHandler.php:340` |
| `x_spending_limit_reached` | `WalletPostDebitJob.php:142` |
| `x_auto_recharge_triggered` | `WalletPostDebitJob.php:187` |
| `social_account_auto_scaled` | `SocialAccountAutoScaleService.php:216` |
| `auto_reply_executed` | `Jobs/Inbox/TrackAutoReplyExecutionJob.php:41` |
| `transaction` | `CustomEventHelper.php:69` |
| `api_call_used` (Customer.io) | `ApiKeyMiddleware.php:125` |

**The critical gap: every one of the six X events is an exception or a money-in event. Not one fires on a successful spend.** Usermaven can tell you when a user was blocked, when they topped up, and when auto-recharge fired. It cannot tell you what anyone actually consumed. The same is true of AI: the ~51 `userMaven.track()` calls in the frontend include `ai_posts_generated`, `ai_post_regenerated`, `ai_post_compose`, `customize_ai_generate_success` and similar, but none carries credits consumed, model used, or plan.

Two further constraints this creates:

1. **The helper is a synchronous blocking HTTP call** on whatever path invokes it. Today that is acceptable because it fires on rare exception paths. Putting it on every AI generation and every X publish would add a third-party round trip to two hot paths, so the events have to be dispatched asynchronously.
2. **Usermaven payload discipline.** The story guidelines cap payloads at roughly six properties for dashboard usability, `snake_case` names, no PII. So Usermaven gets the slicing dimensions, not the full record. The full record belongs in internal admin. Provider cost and margin should not go to a third-party analytics processor at all.

## 5. What reaches internal admin today

Nothing was found. No internal admin surface exists in the mounted repositories.

There is indirect evidence one exists outside them: `contentstudio-social-analytics-go/src/api/middleware/jwt.go` accepts a token signed with `InternalAdminSecret` as an admin caller, alongside `AdminSecret`, and both skip the issuer check.

**Open question the stories must resolve: where the internal admin tool lives, who owns it, and how it ingests data.** Until that is answered, "push to internal admin" cannot be estimated. It is an acceptance criterion on both stories.

## 6. The one thing done properly: the X wallet ledger

`x_wallet_transactions` is a real append-only ledger, one row per charge:

- `app/Models/Billing/XWalletTransaction.php`, `app/Models/Billing/XWallet.php`
- `app/Http/Controllers/Billing/XWallet/UsageController.php` with `index` and `exportCsv`, at `billing/.../usage` and `usage/export.csv`
- `app/Data/Billing/XWallet/XWalletUsageRowData.php`, `XWalletTweetCostData.php`
- reconciliation: `XWalletReconcileLedgerCommand`, `WalletsReconcilePaddleCommand`
- spend controls on the wallet: `balance_usd_cents`, `spent_this_month_cents`, `monthly_spending_limit_cents`, `unlimited_spending`, auto-recharge fields

This is the pattern both stories should follow. Note it is per-wallet and per-transaction but **not** cost-aware in the margin sense: it records what we charged the customer, not what X charged us, so margin is still not derivable from it alone.

## 7. Dual currency, and normalizing it

X metering is dual-currency: a prepaid dollar wallet for the new cohort, legacy X credits for the old one, with the cohort decided by a wallet-enabled check on the user. Known constants from the existing X work: plain post $0.018 and link post $0.24 charged to the customer, against X's $0.015 and $0.20, and 1 legacy credit = $0.0166.

For visibility this matters twice:

- A dashboard must never present a dollar balance and a legacy credit count as the same unit.
- But cross-cohort slicing is exactly what the business wants. So the usable answer is to normalize to a dollar amount for analytics slicing, keep the native unit and the conversion rate on the internal record, and label the cohort on every row.

## 8. Where the plan comes from, and why it must be stored

Limits come from `SubscriptionLimits::getSubscriptionLimits($workspace_id)`, read at the time of the check. If the plan is only referenced rather than stored on the usage record, then a customer changing plan silently rewrites history and every past-period margin figure moves. The plan in force has to be written onto the record.

## 9. Questions currently unanswerable

1. Per user, workspace or account: what was consumed, of which kind, over an arbitrary period.
2. Per plan: consumption against what the plan includes, so underpriced plans are visible.
3. Per generation: what we paid the provider versus what we charged, so margin is measured rather than estimated.
4. Which accounts are loss-making and by how much.
5. Which models consume the budget, so model choice becomes a commercial decision.
6. Which surface consumed it: web, public API, mobile, or an AI assistant.
7. All of the above across every metered feature in one place, rather than one screen per feature.

## 10. Adjacent existing deliverables (do not duplicate)

- `x-pay-per-use-credits`, `x-analytics-metering`
- `analytics-observability-and-data-retention` — the retention precedent to follow
- `ai-surfaces-architecture`, `ai-tools-and-skills-platform`
- `api-centric-plan`, `public-api-ai-image-generation`, `public-api-ai-video-generation`
- Story guidelines section 17 governs every Usermaven event named in these stories.
