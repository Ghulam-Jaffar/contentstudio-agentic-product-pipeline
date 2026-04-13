# PRD: Telegram Integration

**Author:** Product Team  
**Last Updated:** April 6, 2026  
**Status:** Updated — Bot Token Model  
**Target Release:** Q2 2026

---

## 1. Overview

ContentStudio currently supports publishing and scheduling across Facebook, Instagram, LinkedIn, X (Twitter), TikTok, Pinterest, YouTube, Threads, Bluesky, and more — but not Telegram. With Telegram crossing 1 billion monthly active users in early 2025 and becoming a primary broadcast channel for brands, publishers, and creators worldwide, this gap is increasingly costing ContentStudio deals. This PRD defines the v1 Telegram integration: connecting Telegram Channels and Groups as publishing destinations via the user's own bot token (obtained from @BotFather), scheduling text, image, video, album, and PDF document posts from the Composer, managing Telegram accounts in Settings, surfacing Telegram posts in the Planner calendar, and extending RSS automation to Telegram.

The connection model uses per-user bot tokens — each user/workspace creates their own Telegram bot, adds it as admin to their channel(s)/group(s), and provides the token to ContentStudio. This eliminates shared-bot risks (rate limiting, bot name in admin list, one workspace affecting others) and is the enterprise-grade approach.

---

## 2. Problem Statement

**What problem are we solving?**

ContentStudio users who operate Telegram channels or groups cannot manage them from ContentStudio. They are forced to either: (a) publish to Telegram manually and separately from their other platforms, breaking their content calendar workflow; or (b) pay for a second tool just for Telegram scheduling. Agencies managing clients with Telegram presences cannot offer ContentStudio as a complete solution. This creates churn risk, loses deals to competitors who support Telegram, and forces a fragmented workflow on users who otherwise love ContentStudio.

**Who has this problem?**

- **Social media managers** at brands and agencies managing Telegram channels/groups alongside other platforms — they experience this pain daily.
- **Content publishers and media companies** who use Telegram as their primary newsletter/broadcast channel — for them, ContentStudio is currently unusable as an all-in-one tool.
- **Agency owners** managing multiple client accounts — the inability to include Telegram in a client's social media plan means clients must use a second tool or the agency loses the account.
- **Creators** cross-posting content from Instagram, LinkedIn, or X to a private Telegram group for their paying community.

Telegram integration is one of the most-upvoted platform requests on contentstudio.frill.co. Of ContentStudio's 10 direct competitors, only Publer has full native Telegram integration.

---

## 3. Goals & Success Metrics

| Goal | Metric | Target | How We'll Measure |
|---|---|---|---|
| Enable Telegram publishing | # of workspaces with at least 1 Telegram account connected | 500 workspaces within 60 days of launch | Product analytics (social_integrations collection) |
| Reduce churn attributed to missing Telegram | Churn survey / support tickets citing Telegram absence | Reduce mentions by 80% within 90 days | Intercom / churn survey data |
| Establish publishing usage | # of posts published to Telegram per week | 2,000+ posts/week by week 8 | Publishing job logs |
| Guard: publishing reliability | Telegram post failure rate | < 3% failure rate | Job failure logs (Horizon) |
| Guard: no negative impact on existing platforms | Existing platform publishing failure rates | No increase | Platform-specific job failure rates |

---

## 4. Target Users

**Primary Persona:**  
**The Brand Social Media Manager** — Manages 2–5 social accounts for a single brand, including a Telegram channel with 1,000–50,000 subscribers. Publishes 1–3 times per day to Telegram. Currently uses ContentStudio for all other platforms but switches to Telegram's native app or a second tool for Telegram posts. Intermediate tech sophistication — comfortable with social media tools, can follow a @BotFather token creation guide.

**Secondary Persona:**  
**The Agency Account Manager** — Manages social media for 5–20 client accounts. Some clients have Telegram channels. Needs to include Telegram in multi-platform content plans, approval workflows, and Planner views.

**Tertiary Persona:**  
**The Content Publisher / Newsletter Owner** — Uses Telegram as their primary broadcast channel. Publishes 1–5 posts per day. Would migrate to ContentStudio if Telegram is fully supported.

---

## 5. User Stories / Jobs to Be Done

| ID | As a... | I want to... | So that... | Priority |
|---|---|---|---|---|
| US-1 | Social media manager | Connect my Telegram channel to ContentStudio using my bot token | I can manage all my social accounts from one place | P0 |
| US-2 | Social media manager | Connect my Telegram group to ContentStudio | I can schedule community announcements without switching tools | P0 |
| US-3 | Social media manager | Discover and select chats my bot is already in during setup | I don't have to manually enter every channel/group username | P0 |
| US-4 | Social media manager | Schedule text posts to Telegram from the Composer | I can plan my Telegram content calendar in advance | P0 |
| US-5 | Social media manager | Schedule image posts to Telegram from the Composer | I can post visual content on schedule | P0 |
| US-6 | Social media manager | Schedule video posts to Telegram from the Composer | I can publish video content to Telegram without manual effort | P0 |
| US-7 | Social media manager | Schedule a PDF document to Telegram | I can share reports, guides, and catalogs to my channel | P1 |
| US-8 | Social media manager | Schedule an album (up to 10 images/videos) to Telegram | I can post photo collections or product galleries in one message | P1 |
| US-9 | Social media manager | See my scheduled Telegram posts on the Planner calendar | I have a complete cross-platform content calendar | P0 |
| US-10 | Social media manager | See a Telegram-style preview in the Composer | I know exactly how my post will look before sending | P1 |
| US-11 | Social media manager | Post silently (without push notification) | I can post at night without alerting subscribers | P1 |
| US-12 | Social media manager | Disable link preview for a post | I can keep the post clean when the link is already clear | P1 |
| US-13 | Social media manager | Pin a post to the top of my channel after publishing | I can highlight important announcements automatically | P1 |
| US-14 | Agency account manager | Connect multiple Telegram channels (one per client) to a workspace | I can manage all client Telegram channels from ContentStudio | P0 |
| US-15 | Agency account manager | See when a Telegram account needs reconnecting | I can proactively fix broken integrations before failures | P1 |
| US-16 | Content publisher | Use RSS automation to publish RSS feed items to my Telegram channel | I can auto-distribute content to Telegram without manual effort | P1 |
| US-17 | Social media manager | Disconnect a Telegram channel from ContentStudio | I can remove accounts I no longer need | P0 |
| US-18 | Mobile user (iOS/Android) | See my Telegram accounts in the mobile Composer account selector | I can include Telegram when scheduling from my phone | P1 |
| US-19 | Mobile user (iOS/Android) | See Telegram posts on the mobile Planner calendar | I can review my full content schedule from my phone | P1 |
| US-20 | Agency account manager | Share an EasyConnect link with a client so they can connect their own Telegram channel | I can onboard client Telegram channels without handling their bot tokens myself | P1 |

---

## 6. Requirements

### 6.1 Must Have (P0)

- **Telegram account connection via bot token:** Users connect their Telegram Channels and Groups to ContentStudio by providing a bot token from @BotFather. A 3-step modal guides them through: (1) validate token, (2) discover and select chats, (3) confirm and save. API endpoints: `POST /telegram/validate-bot`, `POST /telegram/discover-chats`, `POST /telegram/validate-chat` (manual add), `POST /telegram/add-chats`.
- **Multi-account support:** Multiple Telegram channels/groups can be connected per workspace using the same or different bot tokens.
- **Text post publishing:** Users can publish text-only posts to Telegram from the Composer. Telegram's 4,096-character limit is enforced with a live counter.
- **Single image publishing:** Users can publish a single image with caption (up to 1,024 characters) to Telegram via `sendPhoto`.
- **Single video publishing:** Users can publish a single video with caption to Telegram via `sendVideo`. Files up to 50 MB.
- **Scheduled publishing:** Telegram posts can be scheduled for a future date/time from the Composer.
- **Planner calendar visibility:** Scheduled and published Telegram posts appear on the Planner calendar with the Telegram icon.
- **Account management in Settings:** The Telegram card appears in the integrations settings page. Users can connect, manage connected chats (add more, remove individual chats), and view connected accounts.
- **Post status tracking:** Post status (Scheduled / Published / Failed) is tracked and visible in the Planner. Failed posts show the failure reason.
- **Account validity notifications:** If the user's bot is removed as admin from a Telegram channel/group, the account is flagged as invalid and the user receives an in-app notification to reconnect.
- **Disconnect:** Users can remove individual Telegram chats (`DELETE /telegram/remove-chat/{id}`) from Settings.

### 6.2 Should Have (P1)

- **EasyConnect support:** Agencies using ContentStudio's EasyConnect feature can include Telegram in their client-facing account connection links. When a client visits an EasyConnect link and selects Telegram, the 3-step bot token connection modal opens inline (same flow as Settings → Social Accounts). Connected chats are saved with `connection_via_link: true` and `connection_link_id`. Unlike all other EasyConnect platforms which use OAuth redirects, Telegram's modal-based flow is handled as a special case (same pattern as Bluesky's non-OAuth EasyConnect handling). Backend: `telegramAccounts()` relationship added to `ExternalLinkIntegration` model; eager loading updated in `ExternalLinkIntegrationRepo`; EasyConnect-scoped routes added for validate-bot, discover-chats, and add-chats under `EasyConnectMiddleware`. Frontend: Telegram added to the `useSocialAccounts.js` PLATFORMS array and reactive state; `ExternalCloudConnect.vue` and `ExternalCloudConnectLink.vue` updated to display Telegram accounts and trigger the bot-token modal instead of an OAuth redirect.
- **Album publishing:** Users can publish an album of 2–10 images and/or videos in a single Telegram post using `sendMediaGroup`.
- **PDF document publishing:** Users can publish a single PDF document (up to 50 MB) to Telegram via `sendDocument`. This is mutually exclusive with images/videos in the same post.
- **Telegram post preview in Composer:** `TelegramPreview.vue` renders in the Composer's right preview panel showing channel name, avatar, content, media thumbnail, and character counter.
- **Silent message toggle:** A "Silent message" toggle in the Telegram options panel (`TelegramOptions.vue`). Maps to `disable_notification: true` in the Bot API. Default: off.
- **Disable link preview toggle:** A "Disable link preview" toggle in `TelegramOptions.vue`. Maps to `disable_web_page_preview: true`. Default: off.
- **Pin message toggle:** A "Pin message" toggle in `TelegramOptions.vue`. After successful posting, ContentStudio calls `pinChatMessage`. Default: off.
- **Dynamic character counter:** Counter shows 4096 (text-only) or 1024 (has media) and updates dynamically as media is added/removed.
- **Media type exclusivity:** PDF mode and images/videos mode are mutually exclusive in the Telegram-specific editor. When a PDF is attached, image/video upload is disabled, and vice versa.
- **Common box PDF handling:** PDFs are not available in the common editor (`common_box_status: true`). A hint directs users to switch to per-platform mode to attach a PDF for Telegram.
- **RSS Automation support:** Telegram channels/groups are available as destinations in the RSS-to-post automation feature.
- **Mobile (iOS/Android) — Account selector:** Connected Telegram accounts appear in the mobile Composer account selector.
- **Mobile (iOS/Android) — Planner calendar:** Telegram posts appear on the mobile Planner calendar.
- **Reconnect flow:** When a Telegram account is invalid (bot removed as admin), the Settings card shows a "Reconnect" button. Clicking re-initiates the connection flow (same 3-step modal, bot token pre-populated if available).

### 6.3 Nice to Have (P2)

- **Failed post retry:** A "Retry" button on failed Telegram posts in the Planner detail panel.
- **Bulk scheduling / CSV upload:** Telegram as a destination in the bulk CSV post upload feature.
- **Character count warning indicator:** Yellow at 90% of limit, red at 100%.

### 6.4 Explicitly Out of Scope (v1)

- Protect content toggle (`protect_content: true`) — deferred to v2
- Spoiler blur for media (`has_spoiler: true`) — deferred to v2
- First comment / follow-up reply scheduling — deferred to v2
- Inline CTA button keyboards (URL buttons below post) — deferred to v2
- Poll creation and scheduling — deferred to v2
- Telegram analytics (post views, subscriber counts, channel growth) — deferred to v2; Bot API does not expose them
- AI caption generation optimized for Telegram — deferred to v2
- Telegram inbox/engagement (responding to comments in linked discussion groups) — different product area (Inbox)
- Personal Telegram account (DM) publishing — not technically feasible via Bot API
- Audio/voice notes publishing — deferred to v2
- Shared/single ContentStudio bot approach — opted for user's own bot token model instead

---

## 7. User Flow (High Level)

### Account Connection (3 Steps)

1. User opens **Settings → Social Accounts** and clicks **"Connect"** on the Telegram card.
2. **Step 1:** Modal "Connect Telegram" — user enters their bot token (obtained from @BotFather), clicks "Validate". Backend calls `POST /telegram/validate-bot`. On success, bot info is shown and Step 2 loads automatically.
3. **Step 2:** Backend calls `POST /telegram/discover-chats` to list chats the bot is already admin in. User selects desired channels/groups. For unlisted chats: user enters @username or chat ID, clicks "Add" → backend calls `POST /telegram/validate-chat`. If no chats found, an empty state with instructions is shown.
4. **Step 3:** Summary of selected chats + plan note. User clicks "Connect" → `POST /telegram/add-chats`. On success: toast, modal closes, accounts list refreshes.

### Publishing a Post

1. User opens the **Composer**, selects one or more Telegram accounts.
2. User writes content. Character counter: `X / 4096` for text-only, switches to `X / 1024` when media is attached.
3. `TelegramPreview.vue` renders in the right panel.
4. `TelegramOptions.vue` panel shows: **Silent message** toggle, **Disable link preview** toggle, **Pin message** toggle.
5. For media: user chooses "Images / Videos" mode (up to 10 items, mixed OK) or "PDF Document" mode (single .pdf, up to 50 MB). These modes are mutually exclusive.
6. User clicks **Schedule**, sets date/time, confirms.
7. Post appears in the **Planner calendar**.
8. At scheduled time, ContentStudio fires the appropriate API method using the stored bot token. If `pin_message: true`, `pinChatMessage` is called after a successful post.
9. Post status updates to **Published** or **Failed** with error reason.

---

## 8. Business Rules & Constraints

| Rule ID | Rule | Rationale |
|---|---|---|
| BR-1 | The user's own bot must be an administrator of the Telegram channel or group before posting is possible. If it loses admin status, all future posts to that account fail until reconnected. | Telegram Bot API requirement — bots must have `can_post_messages` (channels) or `can_send_messages` (groups) permission. |
| BR-2 | Channel discovery at Step 2 uses `POST /telegram/discover-chats`. If a channel/group is not discovered (bot was recently added), it can be added manually via `POST /telegram/validate-chat` using @username or chat ID. | The Bot API requires at least one message to have been sent after the bot was added as admin for the chat to appear in the bot's updates; manual add covers edge cases. |
| BR-3 | Media files must be ≤ 50 MB. Files above this limit are rejected at the Composer level with a clear error message. | Telegram Bot API hard limit for standard bot uploads via multipart. |
| BR-4 | Albums must contain 2–10 media items. Single media items use `sendPhoto` / `sendVideo` instead of `sendMediaGroup`. | Telegram API requires at least 2 items for a media group; max 10 per API spec. |
| BR-5 | Text-only posts: max 4,096 characters. Posts with media (image, video, PDF): max 1,024 characters (caption limit). Both enforced in the Composer with a live counter that updates when media is added/removed. | Telegram Bot API character limits per message type. |
| BR-6 | HTTP 429 responses from Telegram trigger retry with exponential backoff using the `Retry-After` header. After 3 retries, post is marked Failed. | Prevents cascading failures. |
| BR-7 | Accounts are workspace-scoped. A chat connected to one workspace cannot be published to from another workspace with a different bot token; each workspace stores its own bot_token + chat_id pair. | Data isolation — same rule as all other social integrations. |
| BR-8 | Silent message and disable link preview are post-level options set per post. Default: off. | Post-level granularity is what users need. |
| BR-9 | PDF document mode and images/videos mode are mutually exclusive for Telegram. When a PDF is attached, image/video upload buttons are disabled. When images/videos are present, PDF upload is disabled. | Telegram's sendDocument cannot include images/videos; sendPhoto/sendMediaGroup cannot include PDFs. |
| BR-10 | PDFs are not available in the common editor (`common_box_status: true`). PDF is a Telegram-specific feature. To attach a PDF for Telegram, user must switch to per-platform mode. | PDFs don't exist as a media type on other platforms (Facebook, Instagram, etc.). |
| BR-11 | If `pin_message: true` and `pinChatMessage` fails after a successful post, the post remains Published. A non-blocking warning is recorded and shown in the post detail panel. | The pin is a bonus action; the post content is what matters. Failing to pin should not mark the post as failed. |
| BR-12 | Telegram posts follow the same content approval workflow as all other platform posts. | Consistency; no special handling needed. |
| BR-13 | The bot token is stored encrypted at rest (`access_token` field, same encryption as other platform credentials). | Security requirement. |

---

## 9. Open Questions

| Question | Options | Owner | Due Date | Decision |
|---|---|---|---|---|
| Should the bot token be validated again at publish time (fresh API call), or is storing + using it enough? | Re-validate at publish / Trust stored token | Backend | Sprint planning | Pending |
| When multiple chats from the same bot token fail at once (token revoked), should we invalidate all accounts under that token or just the one that failed? | Invalidate all / Invalidate per-chat | Backend | Sprint planning | Pending |
| Should the failed-post retry in the Planner be automatic (retry 1x automatically) or manual-only? | Auto-retry 1x / Manual only | Product | Sprint planning | Pending |
| For mobile apps (iOS/Android): should Telegram-specific options (silent message, disable link preview, pin message) appear on mobile in v1? | Show all options / Show simplified options / Defer to web only | Product + Mobile | Before mobile story kickoff | Pending |
| Should "Add more" in the manage-chats flow allow connecting chats from a *different* bot token, or only the same token already on file? | Same token only / Allow different token | Product | Sprint planning | Pending |

---

## 10. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Telegram changes Bot API permissions/structure, breaking the integration | Low | High | Follow Telegram Bot API changelog; use documented stable endpoints. Monitor for deprecation notices. |
| Users confused by bot token creation steps (more steps than OAuth) | Medium | Medium | Clear UX copy and collapsible step-by-step instructions in the connection modal. Track Step 1 → Step 2 funnel completion rates; A/B test UX if conversion is low. |
| Users accidentally revoke or delete their bot between scheduling and publishing, causing silent failures | Medium | High | Proactive validity checking. In-app notification + email when account goes invalid. Clear error messages on Failed posts pointing to the fix. |
| File size limits (50 MB) cause frustration for video-heavy users | Medium | Low | Enforce at Composer upload time with a clear error message. Suggest video compression. |
| Greenfield backend implementation introduces regressions in `SocialPosting.php` | Low | High | Add Telegram behind a feature flag. Full integration test coverage before rollout. Roll out to internal accounts first. |
| Telegram analytics not available via Bot API — users expect analytics parity | High | Medium | Clear in-product messaging that Telegram analytics are not available in v1. Set expectations at account connection and in the Analytics section. |
| `pinChatMessage` may fail if the bot doesn't have the "Pin messages" admin permission (separate from post permission) | Medium | Low | Non-blocking failure (see BR-11). Show a specific warning message pointing to the missing permission. |

---

## 11. Dependencies

**Internal (Codebase):**
- `app/Libraries/Publish/Posting/SocialPosting.php` — Main posting dispatcher; Telegram call added here.
- `app/Strategy/Integrations/Connector.php` — Integration connector router; `TelegramConnector` registered here.
- `app/Strategy/Planner/Posting.php` — Platform posting strategy router; `TelegramPosting` registered here.
- `config/integrations.php` — Telegram API base URL config.
- `config/social_platforms.php` — `telegram` added to `platforms` and `account_selection_fields`.
- `app/Models/Publish/Planner/Plans.php` — `telegram_sharing_details` added to `$fillable`.
- `src/modules/composer_v2/components/ChannelOptions/TelegramOptions.vue` — New file: 3 toggles.
- `src/modules/composer_v2/components/SocialPreviews/TelegramPreview.vue` — New file: post preview.
- `src/modules/composer_v2/views/composerInitialState.js` — Add `defaultTelegramSharingDetails`, `defaultTelegramOptions`, `telegram: []` to `defaultAccountSelection`.
- `src/modules/composer_v2/views/SocialModal.vue` — Validation, payload generation.
- `src/modules/composer_v2/components/AccountSelectionAside.vue` — Telegram platform section.
- `src/modules/common/constants/composer.js` — Add `TELEGRAM` to enums, add to `MULTIMEDIA_ALLOWED_PLATFORMS`.
- `src/modules/integration/config/api-utils.js` — Telegram validation config block.
- `src/modules/publish/config/api-utils.js` — Telegram API endpoint URL constants.
- `src/locales/en/composer.json` + 10 other locale files — Telegram i18n keys.

**External:**
- **Telegram Bot API** (`https://api.telegram.org/bot{token}/`) — All publishing and account discovery. No SLA guaranteed by Telegram.
- **@BotFather** — Users create their own bots here; ContentStudio provides UX guidance but does not control this step.

**No global bot token required.** There is no `TELEGRAM_BOT_TOKEN` env variable — each workspace stores its own bot token encrypted in `social_integrations.access_token`. No Telegram webhook endpoint needed (user-bot model doesn't require webhook for connection).

---

## 12. File Changes Summary

### New Files

| File | Purpose |
|---|---|
| `src/modules/composer_v2/components/ChannelOptions/TelegramOptions.vue` | 3 toggles: silent_message, disable_link_preview, pin_message |
| `src/modules/composer_v2/components/SocialPreviews/TelegramPreview.vue` | Live post preview in composer sidebar |
| Connection modal component (in integrations module) | 3-step bot connection flow |
| Telegram SVG icon file | Platform icon |

### Modified Files

| File | Change |
|---|---|
| `src/modules/composer_v2/views/composerInitialState.js` | Add `defaultTelegramSharingDetails`, `defaultTelegramOptions`, add `telegram: []` to `defaultAccountSelection` |
| `src/modules/composer_v2/composables/useComposerHelper.js` | Add Telegram to SUPPORTED_PLATFORMS, add icon mapping |
| `src/modules/common/constants/composer.js` | Add `TELEGRAM: 'telegram'` to enum, add `'telegram'` to `MULTIMEDIA_ALLOWED_PLATFORMS` |
| `src/modules/integration/config/api-utils.js` | Add telegram validation config block |
| `src/modules/composer_v2/views/SocialModal.vue` | Validation (telegramErrors), payload generation, error aggregation |
| `src/modules/composer_v2/components/AccountSelectionAside.vue` | Add Telegram platform section |
| `src/locales/en/composer.json` (+ 10 other locales) | Add Telegram i18n keys |
| Integrations settings page | Add Telegram card to platform grid |
| `src/modules/publish/config/api-utils.js` | Add Telegram API endpoint URL constants |
| `app/Models/Integrations/Platforms/ExternalLinkIntegration.php` | Add `telegramAccounts()` relationship method |
| `app/Repository/Integrations/Platforms/ExternalLinkIntegrationRepo.php` | Add `'telegramAccounts'` to all `.with()` eager load calls |
| `src/modules/integration/components/platforms/social_v2/composables/useSocialAccounts.js` | Add `'telegram'` to PLATFORMS array; add `telegramAccounts: []` to reactive state |
| `src/modules/integration/components/platforms/social_v2/ExternalCloudConnect.vue` | Add Telegram accounts to account list; handle Telegram modal (bot token) instead of OAuth redirect |
| `src/modules/integration/components/platforms/social/ExternalCloudConnectLink.vue` | Add Telegram `AccountListing` entry (legacy EasyConnect page) |

---

## 13. Appendix

- **Research & Competitor Analysis:** `docs/features/telegram-integration/01-research.md`
- **Workflow Design:** `docs/features/telegram-integration/02-workflow.md`
- **Interactive Design Prototype:** `docs/features/telegram-integration/telegram-integration-mockup.html`
- **Telegram Bot API Documentation:** https://core.telegram.org/bots/api
- **Telegram Bot API Limits Reference:** https://limits.tginfo.me/en
- **Reference Implementation (Bluesky):**
  - `app/Strategy/Integrations/BlueskyConnector.php`
  - `app/Strategy/Planner/BlueskyPosting.php`
  - `contentstudio-frontend/src/modules/integration/components/dialogs/AddBluesky.vue`
  - `contentstudio-frontend/src/modules/composer_v2/components/SocialPreviews/BlueskyPreview.vue`
  - `contentstudio-frontend/src/modules/composer_v2/components/EditorBox/EditorBlueskyBox.vue`

---

## Changelog

| Date | Author | Changes |
|---|---|---|
| April 1, 2026 | Product Team | Initial draft — generated via ContentStudio Product Pipeline |
| April 2, 2026 | Product Team | Updated connection flow: replaced code-based webhook approach with synchronous username/invite link input (Publer-style). |
| April 2, 2026 | Product Team | Added private group support: private groups resolved via my_chat_member webhook + Redis (30-min TTL). |
| April 3, 2026 | Product Team | Added UI design prototype for Connect Telegram modal. Expanded scope: added spoiler blur, first comment to v1. |
| April 6, 2026 | Product Team | **Major workflow change:** Switched connection model from shared @contentstudio_bot to user's own bot token (3-step: validate-bot → discover-chats → add-chats). Removed protect_content, spoiler blur, first comment from v1 scope (deferred to v2). Added pin_message and PDF document support to v1. Removed global TELEGRAM_BOT_TOKEN and webhook URL dependencies. Updated all requirements, business rules, risks, dependencies accordingly. |
