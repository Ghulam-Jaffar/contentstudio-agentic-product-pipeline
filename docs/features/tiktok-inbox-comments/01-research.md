# TikTok Inbox — Research & Competitor Analysis

**Feature:** TikTok Business Account support in ContentStudio Inbox — view and reply to TikTok comments, connect TikTok business accounts (standard OAuth + easy connect), and auto-reply to TikTok comments.
**Date:** 2026-04-30

---

## What is this feature?

TikTok Business Account Inbox gives social media management teams the ability to monitor, reply to, hide, and moderate comments left on their TikTok videos — all from within the same unified inbox they already use for Facebook, Instagram, LinkedIn, and other channels. It eliminates the need to switch to the native TikTok app for comment management.

For brands, the motivation is threefold: TikTok's algorithm actively rewards fast comment engagement (comments improve For You page placement), comment sections are a primary customer service touchpoint for younger audiences, and managing TikTok alongside other channels without a unified tool creates operational overhead.

The feature set covers three distinct capabilities:
1. **View & reply to TikTok comments** (manual engagement)
2. **Connect TikTok Business Accounts** (standard OAuth + easy connect)
3. **Auto-reply to TikTok comments** (rule-based automation)

---

## TikTok Business API Capabilities

| Capability | API Support |
|---|---|
| Read comments on owned videos | Yes |
| Reply to a comment | Yes |
| Reply to a reply (nested) | Yes (via `comment_id` + `parent_comment_id`) |
| Like a comment | Yes |
| Hide a comment (soft moderation) | Yes — toggle public/hidden status |
| Unhide a comment | Yes |
| Delete a comment | Yes |
| Bulk actions (hide/delete multiple) | Yes (batch endpoints) |
| Auto-reply via API | Yes — same Reply endpoint, triggered by automation logic |
| OAuth account connect (standard) | Yes — OAuth 2.0, ~2hr access token + 1-year refresh token |
| TikTok Business Messaging (DMs) | Yes — separate Business Messaging API; geo-restricted |

Key constraint: The standard TikTok Login Kit (OAuth) gives access to organic post comments. Business API expands to ad comment management. ContentStudio needs Business API scope for full coverage.

---

## Competitor Analysis

### Hootsuite

**Has TikTok Comment Inbox:** Yes — full inbox + streams

**Key Capabilities:**
- View comments via Streams; reply to comments
- TikTok Business Messaging (DMs) in Inbox 2.0
- Saved replies; auto-responders; mentions filter
- Post-context visible before replying (shows which video triggered the comment)

**UX Approach:** Dual-surface: Streams (monitoring view) + Inbox 2.0 (conversation view). Comments visible per-post in Streams with inline reply. Inbox shows DMs as threaded conversations.

**Unique Differentiator:** TikTok Business Messaging (DMs) natively in inbox — one of the first to support this. Shows full post context (which video the comment is on) before the agent hits Reply.

---

### Sprout Social

**Has TikTok Comment Inbox:** Yes — Smart Inbox

**Key Capabilities:**
- Reply to comments; hide comments; filter by profile and message type
- Saved replies; emoji support; assign to team member; create Cases
- Automated Rules (auto-categorize, auto-assign, keyword triggers)
- Salesforce + help desk integrations
- TikTok comment senders added to Contact Lists with profile history

**UX Approach:** Smart Inbox with powerful filtering (source, message type). Workflow-centric: comments can become Cases assigned to team members.

**Unique Differentiator:** Deepest CRM + help desk integration. Strongest for enterprise support workflows. Inbox Activity Reporting (first-response time, reply volume, resolution rate) per channel including TikTok.

---

## Common Patterns (across both competitors)

- **Unified inbox model** — TikTok comments appear alongside other channels; no separate TikTok-only view
- **Filter/source selector** — Users can isolate TikTok-only messages via channel/profile filters
- **Inline reply** — Reply composed directly within the comment thread
- **Saved/canned replies** — Reusable response templates to speed up high-volume comment handling
- **Emoji support in replies**
- **Hide action** — Both support the TikTok API hide/unhide moderation action
- **Automated rules** — Trigger-based automation (keyword → action), though Sprout's is more configurable

---

## User Expectations

### Table stakes (must-have for v1)
- View all comments on owned TikTok videos in the inbox
- Inline text reply to a comment
- Hide/unhide a comment
- Filter inbox by TikTok channel
- Saved/canned reply support
- Standard OAuth connect for TikTok Business accounts
- Show which post/video the comment belongs to

### Delighters (differentiators to consider for v1/v2)
- Auto-reply rules (keyword triggers → automated response) — **significant differentiator at ContentStudio's price tier**
- Auto-hide rules for spam/hate keywords
- Easy connect (QR/short-link) to reduce OAuth friction on mobile-first platforms like TikTok
- Bulk comment actions (hide multiple)
- TikTok Business Messaging DMs (Phase 2 — geo-restricted)

---

## Recommended Approach for ContentStudio (v1)

**Connect (prerequisite):**
- Implement standard TikTok OAuth 2.0 for Business Accounts with proper refresh token handling (1-year refresh)
- Implement easy connect flow — a shareable link/QR that lets users authenticate on mobile without going through full desktop OAuth flow (ContentStudio already has this pattern via `ExternalLinkIntegrationController`)

**Inbox (core v1):**
- Surface TikTok comments in ContentStudio's existing social inbox
- Comments filterable by TikTok profile
- Each comment card shows: commenter username + avatar, comment text, timestamp, and the video thumbnail + title it was left on
- Support inline text reply and hide/unhide action
- Add saved replies support for TikTok
- Defer delete (destructive, low demand vs. hide)

**Auto-reply (differentiator v1):**
- Build rule engine for TikTok comments: trigger conditions (keywords, all comments, first comment from user) → actions (reply with template, hide, or both)
- Support multiple response variants (randomized rotation) to avoid repetitive replies flagged by TikTok
- This is a meaningful differentiator — Hootsuite locks advanced automation behind enterprise plans; Sprout Social requires Standard plan+

**Phase 2 (defer):** TikTok DM/Business Messaging (geo-restricted, separate API), CRM contact profiles, ad comment management

---

## Codebase Analysis

### Existing Inbox Infrastructure

**Backend (`contentstudio-backend/`):**

The inbox already has a well-established multi-platform pattern:

- **Platform helper classes:** `FacebookHelper.php`, `InstagramHelper.php`, `LinkedinHelper.php`, `GmbHelper.php` in `/app/Libraries/Inbox/HelperClasses/` — TikTok needs a new `TiktokInboxHelper.php` following this pattern
- **InboxDetailsRepository** (`/app/Repository/Inbox/InboxDetailsRepository.php`): Core MongoDB-backed storage — already platform-agnostic, will work for TikTok
- **InboxDetails Model** (`/app/Models/Inbox/InboxDetails.php`): Two inbox item types: `post` (comments on posts) and `conversation` (DMs)
- **Queue system:** `InboxQueueMasterJob.php` processes `get_conversations` and `get_posts` via Redis — TikTok needs to be added here
- **Saved replies:** `InboxSavedReplyController.php`, `InboxSavedReplyRepository` — already platform-agnostic, saved replies can be reused for TikTok

**TikTok already exists in backend:**
- `TiktokController.php` (`/app/Http/Controllers/Integrations/Platforms/Social/TiktokController.php`): webhook endpoint at `/tiktok/webhook` — currently only handles `post.publish` / `INBOX_SHARE` events; needs to be extended for comment events
- `TiktokHelper.php` (`/app/Helpers/Integrations/TiktokHelper.php`): Token validation + refresh (`validateTiktokAccessToken()`) — already implemented and working
- Routes in `/routes/web/integrations.php` (lines 65–70): `connectLocal`, `webhook` already exist
- TikTok registered in `/config/social_platforms.php` and `/config/socialPost.php`

**Easy Connect:**
- `ExternalLinkIntegrationController.php` — manages shareable external connection links
- `ExternalCloudConnect.vue` — TikTok already listed in bulk reconnect modal (line 85)

**Webhook pattern (follow Instagram's):**
- `InstagramWebhooksController.php` — `hub_verify_token` challenge verification, async job enqueue via `InstagramWebhookExecuteJob`, `LogsBuilder` for audit trail

**Auto-reply:** No automation/rule engine exists yet. Needs new models: `InboxAutoReplyRule`, `InboxAutoReplyTrigger`.

**Real-time:** `InboxPusherBroadcast.php` — Pusher integration for real-time updates already in place.

---

### Frontend (`contentstudio-frontend/`)

**Inbox module:** `/src/modules/inbox-revamp/`
- `InboxView.vue` — main inbox page
- `ConversationView.vue` — single conversation detail
- `MessageComposer.vue` — reply composition (has character limit validation — needs TikTok-specific limit of 150 chars)
- `InboxListing.vue` — conversation/post list with filters
- `PlatformPill.vue` — platform badge (will need TikTok variant)

**Saved replies UI:** `SavedReplyModal.vue`, `SavedReplyListing.vue`, `SavedReplyItem.vue` — all present

**Account connection UI:** `Social.vue`, `SocialConnectModal.vue`, `ExternalCloudConnect.vue` (TikTok already listed)

**Store:** `inbox-revamp.js` (1053 lines, Vuex-based) — needs TikTok-specific state additions

**i18n:** `inbox.*` namespace, `settings.integrations.social.easy_connect.*` for easy connect

---

### Integration Points for TikTok Support

| Layer | Action | File |
|---|---|---|
| BE | Create TikTok inbox worker | New: `/app/Libraries/Inbox/HelperClasses/TiktokInboxHelper.php` |
| BE | Extend webhook for comment events | Modify: `TiktokController.php` |
| BE | Add TikTok to inbox queue | Modify: `InboxQueueMasterJob.php` |
| BE | Add TikTok reply endpoint | Modify: Inbox reply controller |
| BE | Extend OAuth for Business Account scopes | Modify: `TiktokController.php` + `TiktokHelper.php` |
| BE | Build auto-reply rule engine | New: `InboxAutoReplyRule`, `InboxAutoReplyTrigger` models |
| FE | Adapt MessageComposer for TikTok char limits | Modify: `MessageComposer.vue` |
| FE | Add TikTok filter in inbox listing | Modify: `InboxListing.vue` |
| FE | Add auto-reply rules UI | New components in inbox-revamp module |
| FE | TikTok Business Account OAuth connect UI | Modify: `SocialConnectModal.vue` |

---

### Technical Considerations & Gotchas

- **TikTok access token TTL:** ~2 hours access token, 1-year refresh token. Background renewal is critical — pattern already in `TiktokHelper::validateTiktokAccessToken()`
- **Comment character limit:** TikTok comment replies are capped at 150 characters — `MessageComposer.vue` currently uses Instagram/Facebook limits and needs TikTok-specific validation
- **Nested comment threading:** TikTok comments support nested replies (`parent_comment_id`) — InboxDetails data model needs to handle this
- **Webhook signature verification:** TikTok uses HMAC-SHA256, header `X-TikTok-Signature` — pattern exists in Instagram webhook controller
- **Rate limits:** TikTok Business API — 300 requests/minute standard tier; implement exponential backoff in queue job
- **Auto-reply variants:** Send randomized response variants to avoid TikTok flagging repeated identical replies as spam
- **TikTok Business API rate:** `GET /v1/video/comments/search/` requires `video_id` — comment ingestion must iterate per video, not fetch all comments at once
- **No auto-reply infrastructure today:** Needs new BE models and FE rule builder from scratch
