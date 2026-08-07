# Research — X (Twitter) integration in the ContentStudio Social Inbox (pay-per-use)

> **Status:** Research complete (competitor + codebase + X pricing). Two findings materially
> change the feature's shape — see "⚠️ Critical findings" immediately below.

## Decisions locked with the PO (billing)

- **Two billing models run side by side** (PO confirmed, 2026-08): **long-standing workspaces bill in
  X plan credits** (the shipped count-based model) and **newer workspaces bill on the X pay-per-use
  wallet**. The inbox must support both and route per workspace.
- **The segmentation rule (old vs. new) and the X credit rate card are already handled** outside this
  epic. The inbox **adheres to** them — it does not define the cutoff and does not invent credit rates.
- **Credit-billed workspaces share one pool:** X inbox usage draws from the **same monthly
  `x_posting_credits` pool as X publishing** — one balance, one top-up path, one monthly reset. Heavy
  inbox use can therefore exhaust publishing credits, which the UI must warn about.
- **Wallet pricing:** **X cost + 20% markup.** Per action: post/mention/reply/quote **read** $0.006 ·
  DM **read** $0.012 · reply/quote **write** $0.018 · **link** post $0.24 · DM **send** $0.018.
- **Metering:** both syncing (reads) and outbound (writes) deduct, in whichever model applies.
- **No migration** of credit-billed workspaces onto the wallet is in scope here.
- **Quote posts:** outbound quote-posting **in scope for v1** (flag: verify against X's
  write-endpoint tier — reportedly Enterprise-gated as of ~Apr 2026; possible cut).

## ⚠️ Critical findings (read first)

1. **Read metering is the real design problem, not the wallet.** With billing reused, the open
   architecture question is that X charges per *read* and the inbox reads continuously, but
   **reads happen inside the Python inbox service (SIM) while the wallet lives in Laravel** —
   there is **no SIM→wallet connection today**. Metering X inbox reads (with the 24h-dedup and
   the ~7-day mention window in mind) is **net-new cross-service work** and the crux of the design.

2. **The `social-inbox-manager` (SIM) source is not mounted** in this workspace, so its
   internals (strategy registry, Mongo schemas, sync/poll cadence, webhook-vs-poll split) are
   reconstructed from in-repo docs, not verified reads. Verify against the SIM repo before build.

## Decisions locked with the PO (scope)

- **Interaction types (v1):** DMs, mentions (@you), replies to your posts, and quote posts — all four
  (outbound quote-posting included, pending X-tier verification).
- **Platforms:** web + iOS + Android.
- *(Billing decisions above.)*

---

## X API cost reality (why pay-per-use fits)

X moved to a **pay-per-use** model itself on **6 Feb 2026** (credits purchased upfront in the
Developer Console, deducted per request). This is the cost ContentStudio passes through.

**Reads (per resource):** Post read **$0.005** · User read **$0.010** · **DM read $0.010** ·
Followers/Following $0.010 · Likes $0.001.

**Writes (per request):** Post creation **$0.015** · Post **with a link $0.20** ·
**DM send $0.015** · User interaction (follow/like/retweet) $0.015 · Delete $0.010.

**Rules that shape the design:**
- **Monthly cap:** 2M post reads / month on pay-per-use; above that, Enterprise is required.
- **24h dedup:** the same resource requested again within a 24h UTC window is **not** re-charged.
- **No streaming, no webhooks** on pay-per-use; **mentions/search is recent-only**
  (`GET /2/tweets/search/recent`, ~7-day window) — no full archive.
- **Legacy tiers** (grandfathered, closed to new signups): Basic $200/mo, Pro $5,000/mo, Enterprise $42k+.

**Three implications for this feature:**
1. **X inbox must be poll-based, not webhook-based** (unlike FB/IG/LinkedIn). Every sync is
   metered reads → strongest justification for "syncing consumes credits." The 24h dedup rule
   softens repeat-poll cost.
2. **Mentions/replies/quote posts limited to ~7 days** on pay-per-use (recent search only) →
   a real v1 scope constraint (the inbox surfaces roughly the last week, not full history).
3. All four interaction types have concrete per-item costs to base credit consumption on
   (DM read $0.010 / send $0.015; mention/reply/quote reads $0.005).

Sources: docs.x.com/x-api/getting-started/pricing; postproxy.dev/blog/x-api-pricing-2026;
socialcrawl.dev/blog/x-twitter-api-2026.

---

## Competitor & industry research

**Market split since X's 2023 hike + Feb 2026 pay-per-use shift:** enterprise tools absorb
X's API cost into $200+/seat plans; mid-market tools either **dropped X inbox** or moved it
to a **paid add-on / surcharge**. This directly validates ContentStudio's pay-per-use approach.

| Competitor | X in inbox? | Interaction types | How they handle X API cost |
|---|---|---|---|
| **Sprout Social** | Yes (Smart Inbox) | DMs, mentions, replies, quote posts | Absorbed into $199–399/seat/mo |
| **Hootsuite** | Yes (Inbox) | DMs, mentions, replies | Absorbed into Team/Enterprise plans |
| **Agorapulse** | Yes — **paid add-on** | Real-time mentions + DMs (X Plus) | X Lite $29 (publish only) / **X Plus $69/profile/mo** (inbox). Base sync throttled 60min–4hr |
| **Metricool** | Yes | DMs, mentions, comments | **Paid add-on ~$10/account/mo** |
| **SocialBee** | Yes (Engage) | comments, mentions, DMs | Bundled |
| **Loomly** | Yes | mentions, comments, DMs | Bundled; limited filtering |
| **Buffer** | Partial — comments only, **no DM inbox** | replies/mentions | Bundled (Buffer Reply discontinued ~2019) |
| **Sendible** | **Dropped X inbox** | — | Cut X activity/alerts after 2023 to avoid cost |
| **Publer** | Partial — comment mgmt, no DM inbox | post comments | Bundled |
| **Later** | ⚠️ unconfirmed (IG/TikTok-centric) | — | — |
| *CoSchedule* | Yes (recent) | DMs, comments, quote posts, mentions, reposts | Bundled — good full-set reference |

**Takeaways:**
- **Nobody in the mid-market solved this cleanly.** Options seen: absorb cost in a $200+/seat plan,
  charge a flat per-profile add-on (Agorapulse $69, Metricool $10), or drop X entirely (Sendible).
- **The add-on precedent (Agorapulse/Metricool) is the closest analog** — and ContentStudio can do it
  *better* with **usage-fair pay-per-use** (light users pay pennies, heavy agencies pay proportionally),
  reusing the existing X wallet as the wedge vs. flat surcharges.
- **Sync cadence as a priced lever:** Agorapulse throttles base sync (60min/4hr) and restores real-time
  on the paid tier → ContentStudio can make cadence a **user-controlled, credit-priced** choice.
- **No historical backfill:** X doesn't import history retroactively; sync starts fresh on connect
  (Agorapulse states this explicitly). Must be set as an expectation in onboarding copy.

**Table stakes:** unified DM+mention+reply queue across accounts · reply/DM from the inbox · read/unread +
archive statuses · team assignment + notes/collision detection · filtering · thread grouping · search ·
near-real-time (or stated cadence).
**Delighters:** transparent per-action credit metering with live balance in the inbox · user-controlled
sync cadence tied to credits · saved replies · SLA/spike alerts · quote-post/repost read visibility ·
per-account/-workspace credit budgets · cost preview before high-volume actions.

**⚠️ Two API constraints that affect scope (surfaced by research — verify at docs.x.com before build):**
1. **Quote-post as an *outbound* action may be Enterprise-only.** Reports indicate X moved quote-post/
   like/follow **write** endpoints off self-serve tiers (~Apr 2026). **Reading/displaying** incoming quote
   posts is fine; **posting** a quote-post from the inbox is at-risk. Treat plain **reply** (thread post)
   as the reliable outbound primitive. → *This touches the PO's "all interaction types" choice — see gate.*
2. **DM read rate limit is tight (~15 requests / 15 min per user).** Caps how "real-time" DM polling can be;
   shapes the sync-cadence design and the credit model.

**Billing UX best practices to carry into the PRD:** live wallet balance always visible · staged
low-balance warnings (50/75/90% or projected run-out) · per-action cost shown *before* acting · one-click
top-up + optional auto-recharge with a cap · itemized usage log · **graceful degradation** (pause new
sync + outbound at zero, keep already-synced conversations readable — never drop incoming messages).

## Backend: social inbox service + X credit wallet

### Social inbox backend (`social-inbox-manager` / SIM + Laravel bridge)
- Inbox is a **three-tier split**: the web frontend talks **directly** to the Python **SIM**
  service (FastAPI, Kafka, MongoDB, Pusher) for most reads/writes; Laravel is a **proxy + auth**
  layer (mints a short-lived HS256 JWT to call SIM) plus auto-reply/brand-doc APIs.
- SIM uses a **strategy-per-platform** pattern (`app/social_sync/<platform>_strategy.py`;
  Facebook/Instagram confirmed by name, others implied). Data model in Mongo is **split**:
  `inbox_details` (conversations/posts/reviews) + `inbox_messages` + `inbox_comments`, stitched
  at read time (no denormalized list doc). Realtime via Pusher (some events via Kafka →
  `pusher_notification_worker`); `new_element` is a "notify-then-refetch" signal.
- **Outbound writes** proxy through Laravel `app/Http/Controllers/Api/V1/Inbox/` →
  `InboxServiceClient` → SIM endpoints `/inbox/send_text_message`, `/inbox/send_media_message`,
  etc. **This Laravel proxy is the cleanest hook point for metering outbound X credits.**
- **To add X:** new SIM `twitter_strategy.py` + X ingest/webhook handler + register in the
  sync dispatch; reuse the 3 Mongo repos with `platform: 'x'`; Laravel-side add X to
  `config/integrations.php` `inbox_channels` + social-platform configs; auto-reply + public-API
  proxy inherit X for free. The WhatsApp-inbox research doc is the closest "add a platform" template.
- **Webhooks vs polling:** SIM does both, per platform. For X, DMs/mentions can use the Account
  Activity API (webhook) but broader mention/search is **polled → every poll is a metered read**.
  The current SIM sync (dynamic list reconstruction, repeated refetch) would **multiply X read
  costs** if ported unchanged — confirm/optimize the X poll strategy before estimating cost.

### X credit wallet — what's actually shipped vs. planned
- **Shipped (verified in code):** credit-**count** model. Balance = `used_x_posting_credits` int
  on the `workspace` doc vs. `limits.x_posting_credits`. Rate card in
  `app/Helpers/Billing/PlanHelper.php`: `X_PLAIN_POST_CREDITS=1`, `X_URL_POST_CREDITS=15`;
  `calculateXPostingCreditCost()`, `checkAvailableXPostingCredits()`, `hasInsufficientXPostingCredits()`,
  `deductXPostingCredits()` (**read-modify-write, non-atomic**). Publishing deducts inside
  `TwitterPlatform.php` per helper (textPost/imagePost/…), only on success, **bypassed for
  custom-app accounts** (`$isContentStudioApp`). Top-up via Paddle addon `x_posting_credits`
  (1 item = 60 credits). Monthly reset cron (`usedCredits:reset`).
- **NOT built — the intended reuse target:** the prepaid **dollar wallet**
  (`docs/features/x-pay-per-use-credits/`, PRD "In Review", Q3 2026, **Paddle-blocked**):
  account-level USD balance, atomic per-action deduct, editable pricing config, **generic usage
  ledger designed to later cover inbox/analytics/listening**, top-up + auto-recharge + spending limit.
- **`developer_app` / `analytics_enabled`:** how X API *access* is provisioned per account (BYO X
  app creds in `developer_apps`); custom-app accounts currently bypass shared-app credit consumption
  (the dollar-wallet PRD says custom apps are being discontinued → assume all X inbox traffic is metered).
- **Ledgers today are narrow & separate:** `developer_apps_usage` (internal API-call counter) and
  `twitter_jobs_metadata.credits_used` (X analytics — **recorded/reported only, does NOT deduct**).
  Posting credits and analytics credits are **separate pools**; no unified wallet/ledger exists yet.

### Where inbox would hook the wallet
- **Outbound (writes):** wrap the Laravel inbox proxy calls (`send_text_message` / `send_media_message`)
  with check → deduct. Straightforward against either wallet.
- **Inbound (reads/sync):** metering lives **inside SIM**, which has no access to the Laravel wallet
  today → requires a SIM→wallet callback (or moving wallet logic). **Net-new, and the crux of the design.**
- **Clean path** = build X inbox billing on the **planned dollar wallet + generic ledger** (it's
  designed for exactly this), adding `inbox_read` / `inbox_send` action types — **but that depends on
  the dollar-wallet feature shipping first.** Extending the non-atomic count stopgap to high-frequency
  inbox reads is not advisable (race conditions, no read-metering home).

## Frontend (web) + mobile (iOS/Android)

### Web inbox (`contentstudio-frontend/src/modules/inbox-revamp/`, Vue 3)
- Inbox is served by the **separate FastAPI `social-inbox-manager` backend** (not the PHP monolith).
  Everything is an "element" with `inbox_type ∈ {conversation, post, review}` (DM / post+comments /
  GMB review). **There is no first-class "mention" or "quote post" element type today** — a gap for X.
- **X is deliberately excluded today:** `composables/useInboxUI.ts` `filteredChannels` has
  `if (item.name === 'twitter') return`, and `selectedChannels` omits `twitter`. The type layer
  (`InboxPlatform`) already includes `'twitter'`, and platform visuals (`getSocialImageRounded`,
  `bg-social-twitter`, `twitter-x-rounded.svg`) already handle X — so list rows would render an X
  badge once X items flow through.
- **Composer is platform-agnostic:** send is keyed on `platform_type`; X replies/DMs route through the
  same `useSendTextMessageMutation` / `useSendCommentMutation` mutations once the FastAPI backend accepts
  `platform_type: 'twitter'`. This is the natural place to enforce the X credit check + cost confirmation.
- **Real-time** via a generic Pusher/Centrifugo interface (`new_message` / `new_comment` / `new_element`
  on `inbox_events_${workspaceId}`) — X can ride the same events if the backend emits them.
- **To surface X on web:** remove the twitter guard + seed `selectedChannels.twitter`, decide how
  **mentions & quote posts** map onto element types (likely new types → FastAPI + generated-DTO change =
  an API-contract decision to settle with the backend), add credit gating on send, wire real-time.

### Existing X pay-per-use credit UI — **fully built and reusable (web)**
The X posting-credit system already exists for the composer and is directly reusable in the inbox:
- **Credit math:** `composer/utils/xCredits.ts` — `calcXCredits()`, plain post = **1 credit**, post
  **with a link = 15 credits**. (Mirrors X's own $0.005 read / $0.015 write / $0.20 link economics.)
- **Live credit widget + warnings:** `composer/components/TwitterPostUsageAlert.vue` (used/limit, "this
  will consume…", URL warning, zero-remaining, over-limit CTA). Balance via `fetchTwitterPostingLimitsApi`.
- **Top-up / purchase modal:** `components/common/TwitterPostingAddon.vue` (1 pack = 300 credits = $5/mo,
  Paddle checkout, `addon_purchased` / `addon_unlocked_x_posting` events). Addon key
  `ADDON_KEYS.X_POSTING_CREDITS`; balance surfaced in billing limits screens.
- **Lock/gate pattern:** `useTwitterLockGate.ts` — "X locked → open credit purchase modal" (mirror for inbox).
- **Custom-app exemption (billing-critical):** accounts with their own `developer_app_id`
  (`developer_app.analytics_enabled`) **bypass** shared-app credit consumption. So X inbox actions on
  shared-app accounts consume/preview credits; **BYO-app accounts don't**. Must carry into the inbox.

### Mobile (iOS + Android) — very uneven starting points
- **Android** (`contentstudio-android-v2/`, Java) **already has substantial X inbox support**: DMs via
  `Inbox/Chat/ChatActivity.java` (branches on `platform == "twitter"`), posts/mentions/replies via
  `Inbox/Posts/TwitterPostsActivity.java` + `Model/TwitterModel/*` (`Tweets`, `ParentTweet`, `MentionedUser`).
  The gap is parity/quote-posts, not greenfield. Inbox is plan-gated (`getSubscriptionSocialInbox()`).
- **iOS** (`contentstudio-ios-v2/`, UIKit/Swift): inbox exists (FB/IG/LinkedIn/GMB) but **X is not surfaced** —
  only an account-filter reference + dormant tweet model fields (`retweetCount`, `tweetAttachment`).
  Per-platform rendering is hardcoded `if platform == …` chains; adding X = new asset + branches + cell types.
- **⚠️ No mobile billing/credit/wallet/paywall UI exists on EITHER app.** No StoreKit IAP, no BillingClient,
  no Paddle. "Upgrade" strings are just plan-gate alerts. So the pay-per-use **wallet balance / top-up /
  low-balance / per-action cost UI is greenfield on mobile** — either build native, or webview-redirect to
  the existing web Paddle flow (how plan upgrades are effectively handled today). **This is the single
  biggest scope surprise for the "web + mobile" choice.**

### Biggest cross-cutting uncertainty
Whether X **mentions** and **quote posts** map onto existing element types (conversation/post/review) or
require **new element types + FastAPI/DTO changes**. Per repo rules this API-contract decision must be
settled with the `social-inbox-manager` backend before implementation.
