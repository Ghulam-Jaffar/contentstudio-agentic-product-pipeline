# Research — Low X Wallet balance email

## Current State

**The email already exists and ships today.** This is a copy and conditional-content change to a live email, not a new notification.

- **Mail class:** `contentstudio-backend/app/Mail/Billing/LowXWalletBalanceMail.php` — takes `first_name`, `balance`, `threshold`, `manage_url`, `whitelabel`
- **Template:** `contentstudio-backend/resources/views/emails/billing/x-wallet/low-balance.blade.php` — standard email layout (shared head/header/footer partials) with a centred blue CTA button
- **Copy keys:** `lang/en/emails.php` → `x_wallet.low_balance.*`, present in **all 8 locales** (`en`, `de`, `el`, `es`, `fr`, `it`, `pl`, `zh`)

### Current copy (en)

| Key | Value |
|---|---|
| `subject` | "Your X Wallet balance is low" |
| `intro` | "Your X Wallet balance is **:balance**, below your set threshold of **:threshold**." |
| `explainer` | "To keep publishing to X without interruption, top up your wallet or enable auto-recharge." |
| CTA button | `x_wallet.cta.top_up_now` → "Top up now", links to `manage_url` |

The greeting is a shared `common.hi` key plus the bolded first name.

### When it sends

Dispatched from `app/Jobs/Billing/WalletPostDebitJob.php` at two points:

1. **After a post debit** drops the balance below the wallet's auto-recharge threshold — but **only when auto-recharge is disabled**, since a successful recharge in the same tick would make "your balance is low" misleading. Debounced by `low_balance_notified_at` against `x_pay_per_use.low_balance_notify_debounce_hours` (currently **24 h**). The marker resets once the balance climbs back above the threshold.
2. **Payment-setup-required path** — auto-recharge is on but no payment method exists to charge. Uses the same email as a "please top up" prompt, debounced separately via `payment_setup_notified_at`.

`manage_url` is workspace-scoped and deep-links to the billing page's wallet top-up tab, falling back to the workspace picker when no workspace resolves.

## What Needs to Change

### 1. Rewrite the body into four conditional variants

The body becomes a fixed opening and closing with **one swappable middle sentence** driven by two independent flags — whether X Analytics is active and whether X Inbox is active for the account:

| X Analytics | X Inbox | Middle sentence |
|---|---|---|
| off | off | "This may interrupt your scheduled posts to X. To keep publishing without disruption, head over to ContentStudio and top up your wallet from Billings." |
| on | off | "This may interrupt your scheduled posts to X. Your X Analytics will also be paused until your balance is topped up." |
| off | on | "This may interrupt your scheduled posts to X. Your X Inbox will also be paused until your balance is topped up." |
| on | on | "This may interrupt your scheduled posts to X. Your X Analytics and X Inbox will also be paused until your balance is topped up." |

The three variants that name a paused surface are followed by a separate line: "Head over to ContentStudio and top up your wallet from Billings." The publishing-only variant folds that instruction into its single sentence instead.

All four end with a help link: **"Need help? How to top up your X Wallet →"**

### 2. Drop the threshold from the body

New copy reads "Your X Wallet balance is {balance}." with no "below your set threshold of {threshold}" clause. The mail class still receives `threshold` from both dispatch sites, so it becomes an unused field rather than a signature change — worth leaving in place for now since the subject line and future variants may want it.

### 3. Add a help-doc link

No such link exists in this email today. The destination URL is **supplied by the PO to the developer at implementation time** — the story specifies the link text and placement only, and does not invent a URL.

## Risks & Gaps

### The requested copy hardcodes "ContentStudio" — breaks white-label

"head over to **ContentStudio** and top up your wallet from Billings" names the brand literally. Every other email in this codebase resolves the brand dynamically: the footer renders `$whitelabel['name']` linked to `$whitelabel['url']`, and `buildMailData()` already pulls white-label settings for every wallet email. A white-label customer's user would receive an email naming a competitor's product.

**Recommendation:** substitute the app name as a placeholder so it renders the white-label brand, exactly as the footer does. Worth raising with the PO since it is a copy change to their supplied text.

### Analytics and Inbox become wallet consumers (PO-confirmed 2026-08-21)

Today the ledger's `source` values in `XWalletDebitService` are only `publish` and `publish_thread`. **Both X Analytics and X Inbox will also draw on the X Wallet** — both are in development alongside the wallet itself, so all four variants are real and in scope, not speculative.

- **X analytics metering** — `docs/features/x-analytics-metering/`. Adds an account-level "Enable X analytics" toggle and a new ledger consumption type; scheduled syncs pause on low balance.
- **X social inbox** — `docs/features/x-social-inbox-integration/`. Same wallet, same pause-on-low-balance behaviour.

The email must read the two live states at send time rather than hardcoding them, and an absent or unresolvable state must fall back to "not active" so the recipient is never told a surface is paused when it isn't.

### Low-balance notification was a P1 "Should Have"

`docs/features/x-pay-per-use-credits/03-prd.md` lists "low-balance / spending-limit-reached notifications surfaced outside the composer" under **Should Have (P1)** — the email exists, but in-app surfacing does not. This story covers the email only.

## Sibling emails (same family, unchanged by this story)

`spending-limit-reached`, `auto-recharge-failed`, `refund`, `wallet-available` — all under `resources/views/emails/billing/x-wallet/`. The spending-limit email has the same "keep posting" framing and would eventually want the same analytics/inbox treatment, but it is out of scope here.

## Files Involved

- `contentstudio-backend/lang/{en,de,el,es,fr,it,pl,zh}/emails.php` — `x_wallet.low_balance.*` keys
- `contentstudio-backend/resources/views/emails/billing/x-wallet/low-balance.blade.php` — body structure, new help link
- `contentstudio-backend/app/Mail/Billing/LowXWalletBalanceMail.php` — carry the two feature-state flags through to the view
- `contentstudio-backend/app/Services/Billing/XWallet/WalletService.php` — `buildMailData()`, where the flags would be resolved
- `contentstudio-backend/app/Jobs/Billing/WalletPostDebitJob.php` — both dispatch sites

## Mobile

No impact. This is a transactional email; the Flutter app does not render it.
