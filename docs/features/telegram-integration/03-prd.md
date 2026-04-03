# PRD: Telegram Integration

**Author:** Product Team  
**Last Updated:** April 1, 2026  
**Status:** Draft  
**Target Release:** Q2 2026

---

## 1. Overview

ContentStudio currently supports publishing and scheduling across Facebook, Instagram, LinkedIn, X (Twitter), TikTok, Pinterest, YouTube, Threads, Bluesky, and more — but not Telegram. With Telegram crossing 1 billion monthly active users in early 2025 and becoming a primary broadcast channel for brands, publishers, and creators worldwide, this gap is increasingly costing ContentStudio deals. This PRD defines the v1 Telegram integration: connecting Telegram Channels and Groups as publishing destinations, scheduling text, image, video, and album posts from the Composer, managing Telegram accounts in Settings, surfacing Telegram posts in the Planner calendar, and extending RSS automation to Telegram — all powered by ContentStudio's shared Telegram bot (@contentstudio_bot).

---

## 2. Problem Statement

**What problem are we solving?**

ContentStudio users who operate Telegram channels or groups cannot manage them from ContentStudio. They are forced to either: (a) publish to Telegram manually and separately from their other platforms, breaking their content calendar workflow; or (b) pay for a second tool (typically Publer or Postly) just for Telegram scheduling. Agencies managing clients with Telegram presences cannot offer ContentStudio as a complete solution. This creates churn risk, loses deals to competitors who support Telegram, and forces a fragmented workflow on users who otherwise love ContentStudio.

**Who has this problem?**

- **Social media managers** at brands and agencies managing Telegram channels/groups alongside other platforms — they experience this pain daily.
- **Content publishers and media companies** who use Telegram as their primary newsletter/broadcast channel — for them, ContentStudio is currently unusable as an all-in-one tool.
- **Agency owners** managing multiple client accounts — the inability to include Telegram in a client's social media plan means clients must use a second tool or the agency loses the account.
- **Creators** cross-posting content from Instagram, LinkedIn, or X to a private Telegram group for their paying community.

Based on the user feature request board (contentstudio.frill.co), Telegram integration is one of the most-upvoted platform requests. The competitive gap is significant: of ContentStudio's 10 direct competitors, only Publer has full native Telegram integration — and Publer is ContentStudio's closest price-point competitor.

**What happens if we don't solve it?**

- **Churn risk**: Users with active Telegram channels who switch to ContentStudio but find it missing will either churn back to Publer or not convert in the first place.
- **Competitive positioning**: ContentStudio cannot market itself as a "complete social media management platform" with a 1-billion-user platform absent.
- **Agency business loss**: Agencies managing clients with Telegram presences will not choose ContentStudio as their primary tool.
- **Revenue impact**: Every deal lost to Publer due to Telegram absence represents direct revenue loss. Publer's Telegram support is a key selling point they use against ContentStudio in direct comparisons.

---

## 3. Goals & Success Metrics

| Goal | Metric | Target | How We'll Measure |
|---|---|---|---|
| Enable Telegram publishing | # of workspaces with at least 1 Telegram account connected | 500 workspaces within 60 days of launch | Product analytics (social_integrations collection) |
| Reduce churn attributed to missing Telegram | Churn survey / support tickets citing Telegram absence | Reduce mentions by 80% within 90 days | Intercom / churn survey data |
| Establish publishing usage | # of posts published to Telegram per week | 2,000+ posts/week by week 8 | Publishing job logs / analytics |
| Guard: publishing reliability | Telegram post failure rate | < 3% failure rate | Job failure logs (Horizon) |
| Guard: no negative impact on existing platform publishing | Existing platform publishing failure rates | No increase | Platform-specific job failure rates |

---

## 4. Target Users

**Primary Persona:**  
**The Brand Social Media Manager** — Manages 2–5 social accounts for a single brand, including a Telegram channel with 1,000–50,000 subscribers. Publishes 1–3 times per day to Telegram. Currently uses ContentStudio for all other platforms but switches to Telegram's native app or a second tool for Telegram posts. Intermediate tech sophistication — comfortable with social media tools, not a developer.

**Secondary Persona:**  
**The Agency Account Manager** — Manages social media for 5–20 client accounts. Some clients have Telegram channels. Needs to include Telegram in multi-platform content plans, approval workflows, and Planner views. Frustrated that ContentStudio cannot be their single workspace for all client platforms.

**Tertiary Persona:**  
**The Content Publisher / Newsletter Owner** — Uses Telegram as their primary broadcast channel (1,000–500,000 subscribers). Publishes 1–5 posts per day. Currently uses Publer for Telegram scheduling. Would migrate to ContentStudio if Telegram is fully supported alongside their other channels.

**Non-Users (explicitly out of scope):**  
- Personal Telegram users (direct messaging, personal accounts) — Telegram Bot API cannot post to individual DMs without the user initiating contact; this is not a use case for a social media management tool.
- Users who need Telegram as a customer support inbox/engagement tool — This is a separate product area (Inbox) and out of scope for this integration, which is publishing-only.

---

## 5. User Stories / Jobs to Be Done

| ID | As a... | I want to... | So that... | Priority |
|---|---|---|---|---|
| US-1 | Social media manager | Connect my Telegram channel to ContentStudio | I can manage all my social accounts from one place | P0 |
| US-2 | Social media manager | Connect my Telegram group to ContentStudio | I can schedule community announcements without switching tools | P0 |
| US-3 | Social media manager | Schedule text posts to Telegram from the Composer | I can plan my Telegram content calendar in advance | P0 |
| US-4 | Social media manager | Schedule image posts to Telegram from the Composer | I can post visual content to my Telegram channel on schedule | P0 |
| US-5 | Social media manager | Schedule video posts to Telegram from the Composer | I can publish video content to Telegram without manual effort | P0 |
| US-6 | Social media manager | Schedule an album (up to 10 images/videos) to Telegram | I can post photo collections or product galleries in one message | P1 |
| US-7 | Social media manager | See my scheduled Telegram posts on the Planner calendar | I can see my full cross-platform content calendar including Telegram | P0 |
| US-8 | Social media manager | See a Telegram-style preview in the Composer | I know exactly how my post will look before sending it | P1 |
| US-9 | Social media manager | Toggle silent mode when scheduling a Telegram post | I can post at night or during off-hours without alerting subscribers | P1 |
| US-10 | Social media manager | Toggle "disable link preview" when my post includes a URL | I can keep the post clean when the link context is already clear | P1 |
| US-11 | Agency account manager | Connect multiple Telegram channels (one per client) to a workspace | I can manage all client Telegram channels from a single ContentStudio workspace | P0 |
| US-12 | Agency account manager | See when a Telegram account needs reconnecting (bot removed as admin) | I can proactively fix broken integrations before a client's scheduled post fails | P1 |
| US-13 | Content publisher | Use RSS automation to publish RSS feed items to my Telegram channel | I can auto-distribute content to Telegram without manual effort | P1 |
| US-14 | Social media manager | Disconnect a Telegram channel from ContentStudio | I can remove accounts I no longer need to manage | P0 |
| US-17 | Content creator | Prevent recipients from forwarding or saving a post | I can share exclusive content with my community without it spreading beyond my channel | P1 |
| US-15 | Mobile user (iOS/Android) | See my Telegram accounts in the mobile Composer account selector | I can include Telegram when scheduling from my phone | P1 |
| US-16 | Mobile user (iOS/Android) | See Telegram posts on the mobile Planner calendar | I can review my full content schedule from my phone | P1 |

---

## 6. Requirements

### 6.1 Must Have (P0)

- **Telegram account connection:** Users can connect both public and private Telegram Channels and Groups to ContentStudio via the shared @contentstudio_bot. Connection is initiated from Settings → Social Accounts. The user adds @contentstudio_bot as admin to their channel/group in Telegram, then enters the @username (public) or invite link (private) in ContentStudio and confirms. Public channels/groups are validated synchronously via `getChat` + `getChatMember`; private groups are resolved via a `my_chat_member` webhook event cached in Redis when the bot was added as admin.
- **Multi-account support:** Multiple Telegram channels/groups can be connected per workspace (same as multiple Facebook pages or LinkedIn company pages).
- **Text post publishing:** Users can publish text-only posts to Telegram from the Composer. Telegram's 4,096-character limit is enforced with a live counter.
- **Single image publishing:** Users can publish a single image with caption (up to 1,024 characters) to Telegram via `sendPhoto`.
- **Single video publishing:** Users can publish a single video with caption to Telegram via `sendVideo`. Files up to 50 MB are supported.
- **Scheduled publishing:** Telegram posts can be scheduled for a future date/time from the Composer (same as all other platforms). The ContentStudio backend handles scheduling server-side — no reliance on Telegram's native scheduling.
- **Planner calendar visibility:** Scheduled and published Telegram posts appear on the Planner calendar with the Telegram icon.
- **Account management in Settings:** The Telegram tile appears in Settings → Social Accounts. Users can connect, view connected accounts, and disconnect accounts.
- **Post status tracking:** Post status (Scheduled / Published / Failed) is tracked and visible in the Planner. Failed posts show the failure reason.
- **Account validity notifications:** If the ContentStudio bot is removed as admin from a Telegram channel/group, the account is flagged as invalid and the user receives an in-app notification to reconnect.
- **Disconnect:** Users can disconnect a Telegram account from Settings, stopping all future publishing to it.

### 6.2 Should Have (P1)

- **Album publishing:** Users can publish an album of 2–10 images and/or videos in a single Telegram post using `sendMediaGroup`.
- **Telegram post preview in Composer:** A Telegram-style preview renders in the Composer's right preview panel showing channel name, avatar, content, and media thumbnail.
- **Silent mode toggle:** A "Silent post" toggle in the Telegram options panel (sends post without push notification to subscribers). Default: off.
- **Disable link preview toggle:** A "Disable link preview" toggle in the Telegram options panel (suppresses Telegram's auto-generated link preview card). Default: off. Appears only when the post contains a URL.
- **Protect content toggle:** A "Prevent forwarding and saving" toggle in the Telegram options panel. When enabled, recipients cannot forward the post to other chats or save media from it (`protect_content: true` in the Bot API). Default: off.
- **Spoiler blur toggle:** A "Spoiler" toggle in the Telegram options panel, visible only when an image or video is attached. When enabled, media is blurred with a SPOILER overlay until the recipient taps to reveal it (`has_spoiler: true` in `sendPhoto`, `sendVideo`, and `sendMediaGroup` items). Default: off.
- **First comment (follow-up comment):** A "First Comment" section in the Telegram Composer (same UX pattern as LinkedIn first comment). Users can write a follow-up text message (up to 4,096 characters) that ContentStudio automatically posts as a reply to the original Telegram post immediately after it publishes. Uses `reply_to_message_id` in the Telegram Bot API.
- **RSS Automation support:** Telegram channels/groups are available as destinations in the RSS-to-post automation feature.
- **Mobile (iOS) — Account selector:** Connected Telegram accounts appear in the mobile iOS Composer account selector.
- **Mobile (iOS) — Planner calendar:** Telegram posts appear on the iOS mobile Planner calendar.
- **Mobile (Android) — Account selector:** Connected Telegram accounts appear in the Android Composer account selector.
- **Mobile (Android) — Planner calendar:** Telegram posts appear on the Android Planner calendar.
- **Reconnect flow:** When a Telegram account is invalid (bot removed as admin), the Settings tile shows a "Reconnect" button that re-initiates the connection flow.

### 6.3 Nice to Have (P2)

- **Character count warning:** A yellow warning at 90% of the character limit, and red at 100%, in the Composer text area when Telegram is selected.
- **Failed post retry:** A "Retry" button on failed Telegram posts in the Planner detail panel.
- **Bulk scheduling / CSV upload:** Telegram as a destination in the bulk CSV post upload feature.

### 6.4 Explicitly Out of Scope (v1)

- Inline CTA button keyboards (URL buttons below a post) — deferred to v2
- Poll creation and scheduling — deferred to v2
- Telegram analytics (post views, subscriber counts, channel growth) — deferred to v2
- Branded/custom bot token (user's own @bot) — deferred to v2
- AI caption generation optimized for Telegram — deferred to v2
- Telegram inbox/engagement (responding to comments in linked discussion groups) — different product area (Inbox), not in scope
- Personal Telegram account (DM) publishing — not technically feasible via Bot API
- Files/documents/audio/voice notes publishing — deferred to v2

---

## 7. User Flow (High Level)

### Account Connection
> **UI Design Prototype:** See [Connect Telegram Modal — UI Design Prototype](https://app.shortcut.com/contentstudio-team/doc/69cf496d-a54a-4c31-9015-ce29b35e9f72) for full state specifications, UI copy, color palette, and interaction rules. Interactive HTML prototype at `docs/features/telegram-integration/connect-telegram-modal.html`.

1. User opens **Settings → Social Accounts** and clicks **"Connect"** on the Telegram tile.
2. A modal opens: **"Connect Telegram"** — a single input field for the channel/group @username or invite link, a confirmation toggle ("I confirm to have added @contentstudio_bot as an admin"), and Add/Cancel buttons.
3. User adds @contentstudio_bot as an administrator to their channel or group in Telegram, then enters the identifier in the modal and turns on the confirmation toggle.
4. User clicks **"Add"**. ContentStudio resolves the account:
   - **Public (@username):** Calls `getChat` + `getChatMember` synchronously.
   - **Private (invite link):** Looks up the `chat_id` from the `my_chat_member` webhook event stored in Redis when the bot was added, then calls `getChatMember` to confirm admin status.
5. On success, the account is stored and the modal shows: channel name, avatar, and account type (Channel / Group).
6. Account now appears in connected social accounts and is available in the Composer.

### Publishing a Post
1. User opens the **Composer**, selects one or more Telegram accounts in the account selector.
2. User writes content (text, optionally attaches media). Character counter shows remaining characters for Telegram.
3. Telegram-style preview renders in the right panel.
4. Telegram options panel shows: **Silent post** toggle, **Disable link preview** toggle (when URL detected), **Prevent forwarding and saving** toggle, **Spoiler** toggle (when media is attached).
5. Optionally, user expands the **First Comment** section and types a follow-up message (up to 4,096 chars) to be auto-posted as a reply immediately after the main post publishes.
6. User clicks **Schedule**, sets date/time, confirms.
7. Post appears in the **Planner calendar** at the scheduled time.
8. At scheduled time, ContentStudio fires the appropriate Telegram Bot API method. If a first comment is set, a follow-up `sendMessage` with `reply_to_message_id` is posted immediately after.
9. Post status updates to **Published** (or **Failed** with error reason if it fails).

---

## 8. Business Rules & Constraints

| Rule ID | Rule | Rationale |
|---|---|---|
| BR-1 | The ContentStudio bot must be an administrator of the Telegram channel or group before posting is possible. If it loses admin status, all future posts to that account fail until it is reconnected. | Telegram Bot API requirement — bots must have `can_post_messages` (channels) or `can_send_messages` (groups) permission. |
| BR-2 | Public channels/groups (@username) are validated synchronously via `getChat` + `getChatMember`. Private groups (invite link) are resolved by matching the invite hash to a `my_chat_member` webhook event cached in Redis (30-minute TTL) when the bot was added as admin — then confirmed via `getChatMember`. If either check fails, the user sees an inline error and can retry without reopening the modal. | Provides immediate, seamless UX for both public and private groups; mirrors Publer's approach. Private group support requires the webhook to have fired (bot added as admin before submitting). |
| BR-3 | Media files must be ≤ 50 MB. Files above this limit are rejected at the Composer level with a clear error message. | Telegram Bot API hard limit for standard bot uploads. |
| BR-4 | Albums must contain between 2 and 10 media items. Single media items use `sendPhoto` / `sendVideo` instead of `sendMediaGroup`. | Telegram API requires at least 2 items for a media group; max 10 items per API spec. |
| BR-5 | Text post captions on media are capped at 1,024 characters (enforced in Composer). Text-only posts are capped at 4,096 characters. | Telegram Bot API character limits. |
| BR-6 | If Telegram returns HTTP 429 (rate limit), the posting job retries with exponential backoff using the `Retry-After` header value. After 3 retries, the post is marked Failed. | Prevents cascading rate limit failures and maintains data consistency. |
| BR-7 | A user cannot publish to a Telegram account connected to a different workspace. Accounts are workspace-scoped, same as all other social integrations. | Data isolation — same rule as Facebook pages, LinkedIn pages, etc. |
| BR-8 | Silent mode and disable link preview are post-level options, not account-level defaults. They must be set per post. | Post-level granularity is what users need (different posts warrant different settings). |
| BR-9 | Telegram integration uses the shared @contentstudio_bot in v1. Each workspace gets the same bot. The `chat_id` stored per workspace/user uniquely identifies the channel/group — no conflicts. | Simplifies auth model for v1; custom bot tokens (branded bot) deferred to v2. |
| BR-11 | The "Prevent forwarding and saving" toggle maps to `protect_content: true` in the Telegram Bot API and applies to all message types (`sendMessage`, `sendPhoto`, `sendVideo`, `sendMediaGroup`). It is a post-level option, not an account-level default. Default: off. | Post-level granularity — different posts may need different privacy settings (e.g., exclusive paid content vs. regular updates). |
| BR-10 | Telegram posts follow the same content approval workflow as all other platform posts. If a workspace has approval enabled, Telegram posts require approval before they are scheduled or published. | Consistency with existing workflow; no special handling needed. |

---

## 9. Open Questions

| Question | Options | Owner | Due Date | Decision |
|---|---|---|---|---|
| Should @contentstudio_bot name be `contentstudio_bot` or something more brand-friendly like `cs_publishing_bot`? | `@contentstudio_bot` / `@cs_publisher_bot` / `@contentstudio_publisher` | Product + Marketing | Before development kickoff | Pending |
| What happens when multiple ContentStudio workspaces connect the same Telegram channel? (Two different workspaces adding the bot to the same channel) | Allow (both publish independently) / Block (first workspace owns it) / Warn | Product + Backend | Sprint planning | Pending |
| Should the failed-post retry in the Planner be automatic (retry 1x automatically) or manual-only? | Auto-retry 1x / Manual only / Configurable | Product | Sprint planning | Pending |
| For the mobile apps (v1 scope: account selector + Planner visibility), should Telegram-specific options (silent mode, disable link preview) appear on mobile? | Mobile-only shows basic options / Full options on mobile / Defer all Telegram options to web | Product + Mobile | Before mobile story kickoff | Pending |

---

## 10. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Telegram changes Bot API permissions/structure, breaking the integration | Low | High | Follow Telegram Bot API changelog; use documented stable endpoints (`sendMessage`, `sendPhoto`, `sendVideo`, `sendMediaGroup`). Monitor for deprecation notices. |
| @contentstudio_bot gets blocked or rate-limited by Telegram due to misuse (spam) by a subset of users | Low | High | Implement per-workspace rate limiting in the queue system. Monitor for abuse patterns. Shared bot means one bad actor can affect all users — consider workspace isolation at the queue level. |
| Users confused by bot-based connection flow (more steps than OAuth-based platforms) | Medium | Medium | Invest in clear UX copy and step-by-step visual instructions in the connection modal. A/B test modal UX if connection completion rate is low. |
| Bot removed as admin from channel without user realizing it, causing silent publishing failures | Medium | High | Proactive validity checking (same pattern as expired tokens on Twitter/LinkedIn). In-app notification + email when account goes invalid. |
| File size limits (50 MB) cause frustration for video-heavy users | Medium | Low | Enforce limit at Composer upload time with a clear error message referencing the 50 MB limit. Suggest video compression. (Telegram's local Bot API server which allows 2 GB is a v2 option for high-tier plans.) |
| Greenfield backend implementation introduces regressions in the existing SocialPosting dispatcher (`SocialPosting.php`) | Low | High | Add Telegram behind a feature flag. Full integration test coverage for the posting dispatcher before rollout. Roll out to internal accounts first. |
| Telegram analytics are not available via Bot API — users will expect analytics parity with other platforms | High | Medium | Clear in-product messaging that Telegram analytics are not available in v1 (Telegram does not provide them via API). Set expectations at account connection and in the Analytics section. |

---

## 11. Dependencies

**Internal (Codebase):**
- `app/Libraries/Publish/Posting/SocialPosting.php` — Main posting dispatcher; Telegram call added here. Must not break existing platform dispatches.
- `app/Strategy/Integrations/Connector.php` — Integration connector router; `TelegramConnector` registered here.
- `app/Strategy/Planner/Posting.php` — Platform posting strategy router; `TelegramPosting` registered here.
- `config/integrations.php` — Bot token and API URL config added here.
- `config/social_platforms.php` — `telegram` added to `platforms` and `account_selection_fields` arrays.
- `app/Models/Publish/Planner/Plans.php` — `telegram_sharing_details` added to `$fillable`.
- `contentstudio-frontend/src/modules/integration/` — New `AddTelegram.vue` connection modal; Telegram tile in Integrations page.
- `contentstudio-frontend/src/modules/composer_v2/` — `EditorTelegramBox.vue` (options panel), `TelegramPreview.vue`.
- `contentstudio-frontend/src/modules/integration/store/states/platforms/social/telegram.js` — Vuex store state for Telegram accounts.
- RSS automation engine — Telegram added as a destination (same pattern as other platforms).

**External:**
- **Telegram Bot API** (`https://api.telegram.org/bot{token}/`) — All publishing, account metadata fetching, and webhook handling depends on this API. No SLA guaranteed by Telegram.
- **@BotFather** — ContentStudio's shared bot must be created and maintained via BotFather. Bot token must be stored securely as an env variable (`TELEGRAM_BOT_TOKEN`).
- **Telegram Bot API for validation** — The `POST /integrations/telegram/connect` endpoint calls `getChat` and `getChatMember` synchronously for public channels/groups. The bot token must be valid and the API must be reachable for connection to succeed.
- **Telegram webhook endpoint** — ContentStudio must expose and maintain a webhook URL (`TELEGRAM_WEBHOOK_URL`) to receive `my_chat_member` updates when the bot is added as admin to a private group. This is required for private group support.

**Blockers:**
- @contentstudio_bot must be created (or already exist) and the bot token must be provisioned before development can begin.
- Backend `.env` must include `TELEGRAM_BOT_TOKEN` and `TELEGRAM_WEBHOOK_URL` before integration tests can run.
- Webhook URL must be registered with Telegram's `setWebhook` API pointing to ContentStudio's production/staging server.

---

## 12. Appendix

- **Research & Competitor Analysis:** `docs/features/telegram-integration/01-research.md`
- **Workflow Design:** `docs/features/telegram-integration/02-workflow.md`
- **Connect Telegram Modal — UI Design Prototype:** [Shortcut Doc](https://app.shortcut.com/contentstudio-team/doc/69cf496d-a54a-4c31-9015-ce29b35e9f72) · `docs/features/telegram-integration/connect-telegram-modal.html`
- **Telegram Bot API Documentation:** https://core.telegram.org/bots/api
- **Telegram Bot API Limits Reference:** https://limits.tginfo.me/en
- **Reference Implementation (Bluesky):**
  - `app/Strategy/Integrations/BlueskyConnector.php`
  - `app/Strategy/Planner/BlueskyPosting.php`
  - `app/Http/Controllers/Integrations/Platforms/Social/BlueSkyController.php`
  - `contentstudio-frontend/src/modules/integration/components/dialogs/AddBluesky.vue`
  - `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/BlueskyPreview.vue`
  - `contentstudio-frontend/src/modules/composer_v2/components/EditorBox/EditorBlueskyBox.vue`

---

## Changelog

| Date | Author | Changes |
|---|---|---|
| April 1, 2026 | Product Team | Initial draft — generated via ContentStudio Product Pipeline |
| April 2, 2026 | Product Team | Updated connection flow: replaced code-based webhook approach with synchronous username/invite link input (Publer-style). Removed BR-2 (code expiry), updated BR-2 to synchronous validation rule, removed webhook URL blocker. |
| April 2, 2026 | Product Team | Added private group support: private groups resolved via my_chat_member webhook + Redis (30-min TTL). Re-added TELEGRAM_WEBHOOK_URL dependency. Updated BR-2, Section 6.1, Section 7, and dependencies accordingly. |
| April 3, 2026 | Product Team | Added UI design prototype for Connect Telegram modal (6 states, full copy, color spec). Referenced in Section 7 and Appendix. Shortcut Doc: Connect Telegram Modal — UI Design Prototype. |
| April 3, 2026 | Product Team | Expanded scope: moved Spoiler blur and Follow-up comment scheduling from v2 to v1. Added to Section 6.2 (P1), Section 7 Publishing flow. Updated Section 6.4 accordingly. New stories: [BE] + [FE] Telegram first comment, spoiler added to existing BE Publishing + FE Composer stories. |
