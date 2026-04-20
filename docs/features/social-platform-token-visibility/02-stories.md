# Epic & Stories — Social Platform Token Visibility & Expiry Gaps

## Epic

**Name:** Backend: Close token visibility & account expiry gaps across social platforms
**Shortcut:** [#116761](https://app.shortcut.com/contentstudio-team/epic/116761)
**Objective:** 2026 - Q2
**Timeline:** 2026-04-20 → 2026-05-01
**Group:** Backend
**State:** To Do

The epic body captures the four gaps surfaced in the research doc and ties them to four stories that, together, bring all 11 platforms to end-to-end parity on token lifecycle.

## Stories

All stories: `[BE]` prefix, skill set = backend, product area = integrations, priority = high, project = Web App, iteration = `20 April - 01 May - 2026`.

### 1. [BE] Wire platform observer to run on create/update for all social platforms
- **Scope:** Add `Redis::rpush('social_observers', ...)` in `boot()` for GBP, YouTube, TikTok, Threads, Bluesky, Tumblr models. Verify `PlatformSavedJob` has a branch (even if no-op) for each.
- **Why first:** This is the foundation — without the observer firing, the other downstream mechanisms (metadata refresh, initial validity) never run on connect/reconnect.

### 2. [BE] Extend AccountExpiryJob to send reconnect emails for all social platforms
- **Scope:** Add per-platform `check{Platform}Expiry()` methods in `AccountExpiryJob` for Instagram, Twitter, Pinterest, YouTube, TikTok, Threads, Bluesky, GBP, Tumblr. Reuse the existing `sendSocialExpiryNotification()` flow.
- **Why:** Users on 9 platforms silently lose publishing today when tokens expire.

### 3. [BE] Expose token validity fields in account API responses
- **Scope:** Add `validity`, `validity_error`, `validity_status`, `invalid_tries`, `sent_invalid_email`, `token_issued_at`, `token_expires_at` to `SocialRepo` default fields and `PlatformController::getSocialAccountsList()` required fields. Explicitly exclude encrypted credentials from responses (add a unit test).
- **Why:** Frontend has nothing to render "reconnect required" banners with today — fields exist in the DB but aren't surfaced.

### 4. [BE] Re-enable and extend periodic token validation to all social platforms
- **Blocked on sync with Bilal** (story comment added) before any code change — the paused job has no documented reason.
- **Scope:** Remove `return false;` in `SocialAccountsJob::handle()`, add per-platform validators for all 11 platforms, schedule via `Console/Kernel.php` at a cadence agreed with Bilal, add rate-limit and fail-open guards.
- **Why:** Actively refreshes `validity` / `expires_at` — the other three stories depend on these fields being maintained.

## Discussion points captured on Story 4 (for Bilal)

1. Why the job was paused (Facebook rate limits, API deprecation, cost, abandoned refactor?)
2. Safeguards required before re-enabling — circuit breaker, per-workspace throttling, kill-switch env var, staged rollout
3. Target cadence and spread — running all accounts at once will hit platform rate limits
4. Scope of first release — all 11 platforms at once, or Facebook + LinkedIn first then incremental
