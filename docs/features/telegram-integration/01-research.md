# Telegram Integration — Research Report

**Feature:** Telegram Integration for ContentStudio  
**Date:** April 1, 2026  
**Pipeline Step:** 01 — Research  

---

## Part A: Competitor & Industry Research

---

### 1. What Is This Feature?

Telegram Integration for a social media management tool means connecting Telegram Channels and/or Groups as publishing destinations — just like Facebook Pages, Instagram accounts, or LinkedIn pages — so users can schedule, publish, and analyze content on Telegram from within their existing social media workflow.

**Why users want it:**  
- Telegram crossed 1 billion monthly active users in early 2025, making it one of the world's largest social platforms. Brands and creators who already have Telegram channels cannot manage them from ContentStudio, forcing them to juggle a separate tool or publish manually.
- Telegram Channels are a broadcast medium perfectly suited for content marketing: newsletters, product updates, promotions, community announcements. They have no algorithm suppressing reach — every subscriber sees every message.
- Community managers running Telegram Groups (for customer support, niche communities, affiliate groups) want to schedule announcements and promotions into their groups without switching context.
- Agencies managing multiple clients increasingly have clients with active Telegram presences. Without native Telegram support, ContentStudio is not a complete solution for those clients.

**Core use cases:**
1. A media publisher scheduling daily news digest posts to a Telegram channel with thousands of subscribers.
2. An e-commerce brand pushing promotional posts (image + caption + link) to their Telegram channel on a planned cadence.
3. A SaaS company posting product updates, changelogs, and tips to a public Telegram channel.
4. An agency managing a client's Telegram channel alongside their Facebook, Instagram, and LinkedIn.
5. A creator cross-posting content from Instagram/Twitter to a Telegram group for their paying community.

---

### 2. Competitor Analysis Table

| Competitor | Has Telegram Integration? | Key Capabilities | Pricing Tier | UX Approach | Unique Differentiator |
|---|---|---|---|---|---|
| **Buffer** | No | No native Telegram support; only via Zapier automation (not publishing) | Starts ~$6/mo (Essentials) | N/A — indirect only | Does not support Telegram natively |
| **Hootsuite** | No | No native Telegram support; available only via Zapier/Integrately third-party automations | Starts ~$99/mo | N/A — indirect only | One of the most glaring gaps for an enterprise tool at this price |
| **Publer** | Yes — Full | Channels + Groups; text, image, video, GIF, multi-photo albums (up to 10); bulk CSV scheduling; recurring/evergreen posting; RSS auto-posting; follow-up comments (up to 2); UTM link tracking; PDF/CSV analytics export; mobile app support | Free (3 accounts); Pro from $5/mo; Business from $10/mo | Bot-based connection (Publer Bot added as channel admin); very clean UX with Telegram listed as a first-class platform alongside other networks | Most complete Telegram implementation of any scheduler; includes analytics on reach, views, ad performance, fake subscriber detection |
| **Later** | No | No Telegram support; focused on visual-first platforms (Instagram, Pinterest, TikTok) | Starts ~$25/mo | N/A | Telegram not a strategic priority given their visual-first audience |
| **Sprout Social** | No | No native Telegram integration; only through third-party automation services | Starts ~$199/mo | N/A — indirect only | Major competitive gap at enterprise pricing |
| **Loomly** | No | No native Telegram; only via Zapier | Starts ~$32/mo | N/A — indirect only | Focused on brand-building content; no current Telegram plans evident |
| **Sendible** | No | No native Telegram integration found; agency-focused but lacking Telegram | Starts ~$29/mo | N/A | Gap for agency customers who manage brands with Telegram channels |
| **SocialBee** | Partial (Semi-Automated) | "Universal Posting" feature: user creates/schedules post, receives a mobile push notification at the scheduled time, taps to manually copy-paste content into Telegram; no true auto-publish | Starts ~$29/mo | Mobile reminder workflow — not true scheduling; Telegram is in a list with WhatsApp, Facebook Groups, Quora as "unsupported platforms" | Clever workaround but not genuine automation; relies on user action |
| **Agorapulse** | No | No native Telegram; only through third-party platforms like ApiX-Drive | Starts ~$69/mo | N/A | Notable gap for mid-market/agency customers |
| **Metricool** | No | No native Telegram publishing; Zapier integration only | Starts ~$22/mo | N/A | Focused on analytics-led approach; Telegram publishing not yet supported |
| **Zoho Social** | Partial (Inbox only) | Added Telegram to unified inbox in 2024 for message management/responses, but no scheduling/publishing | Included in Zoho Social plans | Inbox/engagement focus, not publishing | First major platform to add Telegram inbox management, but still no publishing |
| **Nuelink** | Yes — Full | Channels + Groups; text, images, videos, multimedia carousels (up to 10 items); bulk CSV scheduling (up to 100 posts); polls; custom follow-up replies; AI writer | Starts ~$18/mo | Straightforward scheduling UI with Telegram as a first-class channel | AI tools + bulk CSV for Telegram; polls support |
| **Postly** | Yes — Full | Channels + Groups; text + media + albums; inline button keyboards (CTAs); auto-pin; silent post; web preview control; content protection (prevent forwarding); spoiler media; own branded bot option; approval workflows; analytics | Free tier; Basic from $3/mo | Most feature-rich Telegram publishing UX; brings Telegram-native features (silent mode, inline buttons, spoiler) to a scheduler | Telegram-native controls (silent post, content protection, inline CTA buttons) not found in other schedulers; branded bot option |
| **OnlySocial** | Yes — Full | Channels + Groups; scheduling down to the minute; polls/quizzes/surveys; real-time analytics; image watermarking; AI caption tools | Starts ~$24/mo | Clean all-in-one UI; AI-powered | Polls and quizzes built in; auto-watermarking on images |

**Summary:** Of the 10 standard ContentStudio competitors, **only Publer has full native Telegram integration**. Buffer, Hootsuite, Later, Sprout Social, Loomly, Sendible, and Metricool have no native support at all. Agorapulse has no native support. SocialBee has a semi-manual workaround. Zoho Social supports Telegram inbox but not publishing. The tools that do Telegram well (Publer, Postly, Nuelink, OnlySocial) are positioned as lower-cost, broader-platform schedulers — not ContentStudio's direct competitors, but increasingly eating into ContentStudio's potential market.

---

### 3. Telegram API Capabilities & Constraints

ContentStudio would integrate via the **Telegram Bot API** (not the MTProto API). Here is what this means in practice:

#### 3a. How It Works

A Telegram Bot is created via @BotFather and given an API token. The bot must be added as an **Administrator** of a Channel or Group to be able to post on behalf of that channel/group. ContentStudio would run one shared bot (like Publer's "publer_bot") or allow users to connect their own bot token.

#### 3b. Supported Content Types (What Can Be Published)

All of the following can be sent via Bot API:

| Method | Content Type | Notes |
|---|---|---|
| `sendMessage` | Text (with Markdown/HTML formatting) | Up to 4,096 characters |
| `sendPhoto` | Image + caption | Caption up to 1,024 characters (Premium: 4,096) |
| `sendVideo` | Video + caption | Up to 50 MB via Bot API (2 GB with local Bot API server) |
| `sendDocument` | File/document | Up to 50 MB; PDF, ZIP direct by URL |
| `sendAudio` | Audio file | MP3, M4A, etc. |
| `sendAnimation` | GIF or MP4 animation | |
| `sendVoice` | Voice message | OGG format, max 1 MB |
| `sendVideoNote` | Circular video (video note) | Max 1 min, max 12 MB |
| `sendPoll` | Interactive poll | Regular or quiz type |
| `sendMediaGroup` | Album (mixed photos + videos) | Up to 10 items per album |
| `sendLocation` | Geographic location pin | |
| `sendContact` | Contact card | |
| `sendSticker` | Sticker | |

**Inline Keyboards**: Any message type can include inline button keyboards (CTA buttons with URLs), enabling interactive posts with "Click here", "Learn more", "Buy now" buttons.

**Formatting**: Markdown v2 and HTML formatting supported — bold, italic, underline, strikethrough, inline code, hyperlinks.

#### 3c. Channel vs. Group vs. Personal Publishing

| Aspect | Channels | Groups (Supergroups) | Personal (DMs) |
|---|---|---|---|
| Publisher | Admins only (bot must be admin) | Any member can post; bot must be admin to post programmatically | Bot can only message users who started a chat with the bot |
| Audience size | Unlimited subscribers | Up to 200,000 members | 1-to-1 |
| Member interaction | Subscribers cannot reply in channel itself (can if linked group exists) | Members can reply, react, discuss | N/A |
| Bot rate limit | 1 message/sec per chat; 20 messages/min per chat | 20 messages/min per chat | 1 message/sec |
| Best for | Broadcasting content to large audiences; brand announcements; newsletters | Community management; customer engagement; announcements | Not applicable for SMM tools |
| Analytics available | Views per post via API | Limited | N/A |

**For ContentStudio's purposes: Channels and Groups are the two types to support. Personal messaging is not applicable.**

**Discussion Groups**: Telegram allows linking a discussion group to a channel. When a post is published to the channel, it automatically creates a thread in the linked group. This is a native Telegram behavior — ContentStudio does not need to do anything special to support it, but it's worth documenting in UX.

#### 3d. Scheduling Limitations

- Telegram's **native scheduled messages** (via the Telegram app itself) support up to 100 scheduled messages per chat and can be scheduled up to 365 days in advance.
- However, the Bot API does **not** expose a scheduling endpoint — bots cannot "schedule" messages via the API. All scheduling must be handled server-side by the management tool (i.e., ContentStudio stores the scheduled time and fires the API call at the right moment). This is the same pattern ContentStudio already uses for all other platforms.
- There is no Telegram-side scheduling to worry about from an API standpoint.

#### 3e. Rate Limits

| Limit | Value |
|---|---|
| Messages to the same chat | Max ~1/sec (enforced); 20/min for groups |
| Messages across all chats | Max 30 messages/sec |
| API requests overall | Max ~30 requests/sec |
| File uploads (standard) | Max 50 MB |
| File uploads (local Bot API server) | Up to 2 GB |
| Media captions | 1,024 characters (or 4,096 with Telegram Premium) |
| Text messages | 4,096 characters |
| Album size | Up to 10 items |
| Scheduled messages (native) | Up to 100 per chat, up to 365 days ahead |

If rate limits are exceeded, the API returns `HTTP 429` with a `Retry-After` header. ContentStudio should implement exponential backoff.

#### 3f. Notable Behaviors & Restrictions

1. **Bot must be admin**: For channels, the bot needs `can_post_messages` permission. For groups, it needs `can_send_messages`. Without admin rights, posting fails.
2. **No analytics API for basic stats**: Telegram does not expose per-post view counts or subscriber analytics via the Bot API. More advanced analytics (views, reach, subscriber growth) require either Telegram's own in-app stats (admin only) or third-party analytics tools like TGStat. This is a known limitation across all Telegram schedulers.
3. **Public vs. private channels**: Bots can post to both, but analytics and cross-referencing work differently.
4. **Inline buttons require a URL**: CTA buttons attached to messages must link to external URLs (not Telegram-internal deep links) for most use cases.
5. **Content protection**: The `protect_content` parameter prevents forwarding and saving of messages — useful for exclusive content.
6. **Silent notifications**: The `disable_notification` parameter sends a message without triggering push notifications — useful for non-urgent posts.
7. **Link preview control**: `disable_web_page_preview` prevents Telegram from generating a link preview card below the message.
8. **Media spoiler**: `has_spoiler` blurs media until the user taps to reveal — used for surprise announcements or sensitive content.

---

### 4. Common Patterns Across Tools

1. **Bot-based connection flow**: All tools (Publer, Nuelink, Postly) use a bot. User is instructed to: (a) add the tool's bot (e.g., @publer_bot) as admin to their channel/group, (b) enter the channel @username or invite link in the tool's UI. The tool validates the bot's admin status synchronously (public channels/groups) or via a background webhook event (private groups). Publer supports both public and private groups through this single input UX — the distinction is handled transparently on the backend.

2. **Channel + Group support from day one**: Every tool that supports Telegram supports both channels and groups — rarely just one.

3. **Text + image + video as minimum viable content types**: Every tool supports at minimum text, single image, and single video posting. Albums (multi-image) are supported by Publer, Nuelink, and Postly.

4. **Scheduling is server-side**: No tool relies on Telegram's native scheduled messages. All store the schedule on their side and fire at the right time.

5. **No native Telegram analytics from Bot API**: Since the API doesn't expose post analytics, most tools either (a) don't provide Telegram analytics, (b) show basic post-send confirmation, or (c) integrate with third-party Telegram analytics tools. Publer is notable for offering ad effectiveness and view count analytics.

6. **Cross-posting**: Most tools make it easy to compose once and cross-post to Telegram + other platforms. The "compose for all" approach is standard, with per-platform customization available.

7. **Pricing is not a differentiator for Telegram**: Publer offers Telegram at $5/mo; Nuelink at $18/mo. Telegram is not treated as a premium add-on — it's included in base plans.

---

### 5. Differentiators Worth Considering for ContentStudio

1. **Telegram-native options exposed in UI (Postly's approach)**: Postly surfaces Telegram-specific publish controls that most schedulers ignore: silent mode (no notification), disable link preview, content protection (prevent forwarding), spoiler blur for media, pin to top of channel, inline CTA buttons. These are differentiators because users who run serious Telegram channels know and want these features.

2. **Branded bot option (Postly)**: Let advanced users connect their own bot token instead of ContentStudio's shared bot. This keeps the bot's name matching the brand in the channel history. Premium UX feature.

3. **Follow-up comment scheduling (Publer, Nuelink)**: The ability to schedule a follow-up message/comment after the main post (e.g., "5 minutes after the main post, send: 'Reply below with your thoughts!'") is a distinctive Telegram engagement tactic. Publer supports up to 2 follow-up comments.

4. **Poll scheduling**: Nuelink and OnlySocial support scheduling polls to Telegram. This is often overlooked but highly engaging for community channels. ContentStudio's Composer could allow creating polls natively.

5. **Analytics integration**: ContentStudio already has an analytics module. Being the first major SMM tool to expose Telegram channel statistics (via integration with channels' native stats or third-party APIs) would be a clear differentiator. Telegram channels expose view counts for posts to channel admins — ContentStudio could pull these to show performance metrics.

6. **AI caption generation specific to Telegram tone**: ContentStudio has AI features. Telegram's tone is typically more conversational and direct than LinkedIn or Instagram. Offering Telegram-specific AI caption templates or tone suggestions would be differentiating.

7. **RSS-to-Telegram automation**: Publer supports auto-posting from RSS feeds to Telegram. ContentStudio already has RSS automation — extending it to Telegram would be natural and high-value for publishers.

---

### 6. User Expectations

**Table Stakes (what users assume ContentStudio will include from day 1):**
- Connect Telegram channels and groups (add as social accounts)
- Schedule text posts, single images, and videos to Telegram
- See scheduled Telegram posts in the content planner/calendar
- Basic post status (sent/failed) visibility
- Support for multiple Telegram channels/groups per workspace
- Telegram accounts listed alongside other social accounts in account selector

**Delighters (things that would surprise and delight users):**
- Album/carousel posting (up to 10 images/videos in one post)
- Inline CTA button keyboard on posts (e.g., "Buy Now → link")
- Silent post option (no push notification to subscribers)
- Disable link preview toggle
- Follow-up comment scheduling (post a reply to your own message, X minutes later)
- Poll creation and scheduling directly from Composer
- Telegram analytics — view counts, subscriber growth, top-performing posts
- RSS-to-Telegram automation
- Content protection toggle (prevent forwarding/saving)
- Branded bot support (connect your own @YourBot)
- AI caption generation optimized for Telegram's direct/conversational style

---

### 7. Recommended Approach for ContentStudio

**Phase 1 — v1 (Competitive parity with Publer + core use cases):**
1. Connect Telegram Channels and Groups as social accounts via ContentStudio's shared bot
2. Publish from Composer: text, single image, single video, multi-image albums (up to 10)
3. Schedule from the Planner calendar
4. Telegram account selector in Composer (alongside existing social accounts)
5. Telegram post preview in Composer (shows how message will look in Telegram)
6. Account management in Settings → Social Accounts (connect, disconnect, re-authorize)
7. RSS automation support for Telegram (extend existing RSS feature)
8. Post publish status in Planner (published / failed / scheduled)
9. Basic Telegram-specific options: disable link preview toggle, silent mode toggle

**Phase 2 — v2 (Differentiating, delight-tier):**
1. Inline CTA button keyboard (add URL buttons below post)
2. Follow-up comment scheduling (scheduled replies to your own message)
3. Poll creation and scheduling
4. Content protection toggle (prevent forwarding/saving)
5. Spoiler blur for media
6. Telegram channel analytics (post views, subscriber count trends)
7. Branded bot support (custom @bot integration)
8. AI caption generation with Telegram-optimized prompts

**Strategic rationale:** ContentStudio is a serious competitor to Publer in the mid-market. Publer's Telegram integration is the benchmark — matching it in Phase 1 removes a reason to choose Publer over ContentStudio. Phase 2 features (analytics, polls, inline buttons) would put ContentStudio clearly ahead of all schedulers on Telegram capability.

---

## Part B: Codebase Analysis

---

### 8. Existing Related Code

#### Backend (Laravel 10, PHP 8.3)

**Platform integration architecture:**

The codebase has a clear, consistent pattern for adding new social platforms. Each platform follows the same layered structure:

| Layer | Pattern | Relevant Files |
|---|---|---|
| Integration connector | `App\Strategy\Integrations\[Platform]Connector` | `BlueskyConnector.php`, `TwitterConnector.php`, etc. |
| Posting strategy | `App\Strategy\Planner\[Platform]Posting` | `BlueskyPosting.php`, `ThreadsPosting.php`, etc. |
| Model | `App\Models\Integrations\Platforms\Social\[Platform]Accounts` | `FacebookAccounts.php`, etc. |
| Controller | `App\Http\Controllers\Integrations\Platforms\Social\[Platform]Controller` | `BlueSkyController.php`, etc. |
| Config entry | `config/integrations.php` → `social_integrations.[platform]` | `integrations.php` |
| Platform list | `config/social_platforms.php` | `account_selection_fields` and `platforms` arrays |
| Store state | `contentstudio-frontend/src/modules/integration/store/states/platforms/social/[platform].js` | `bluesky.js`, `threads.js`, etc. |

**Platform config pattern (from `config/integrations.php`):**
- OAuth2 platforms (Threads, Twitter): store `client_key`, `client_secret`, `redirect_uri`, `scope`, `authorization_url`, `api_url`, `token_url`
- Non-OAuth platforms (Bluesky): only store `api_url` and other static config
- **Telegram would follow the Bluesky pattern** — no OAuth, just a bot token and API URL

**`SocialPosting.php`** (the main posting dispatcher at `app/Libraries/Publish/Posting/SocialPosting.php`):
- Calls each platform's posting class sequentially
- Currently handles: Facebook, Instagram, Twitter, LinkedIn, Pinterest, Tumblr, GMB, YouTube, TikTok, Threads, Bluesky
- Telegram would be added here with a condition: `if (($plan['account_selection']['telegram'] ?? false) && ...)` → `(new Posting('Telegram', $plan))->initializePosting()->performPosting()`

**`Strategy\Planner\Posting.php`** — the strategy router for newer platforms:
- Currently handles: Gmb, Youtube, TikTok, Threads, Bluesky
- Telegram would be registered here as `case 'Telegram': $this->post = new TelegramPosting($plan); break;`

**`Strategy\Integrations\Connector.php`** — integration connector router:
- Currently registered: Facebook, Instagram, Threads, Twitter, LinkedIn, Pinterest, GMB, YouTube, TikTok, Bluesky
- Telegram would be added: `case 'telegram': $this->connector = new TelegramConnector(); break;`

**`config/social_platforms.php`**:
- `account_selection_fields` array: needs `'telegram.platform_identifier'`
- `platforms` array: needs `'telegram'`

**`config/integrations.php`** `social_integrations_collection_channels`:
- Currently: `["bluesky", "threads"]`
- Telegram channels would be added here

**Plans model** (`app/Models/Publish/Planner/Plans.php`):
- The `$fillable` array currently includes `*_sharing_details` for each platform
- Would need: `'telegram_sharing_details'` added to `$fillable`

#### Frontend (Vue 3, TypeScript, `<script setup>`)

**Bluesky integration as the reference pattern** (most recently added platform, cleanest implementation):

| Component | Path | Purpose |
|---|---|---|
| Connection modal | `src/modules/integration/components/dialogs/AddBluesky.vue` | Modal for connecting a Bluesky account (username + app password) |
| Store state | `src/modules/integration/store/states/platforms/social/bluesky.js` | Vuex state for Bluesky accounts (`items`, `all_items`) |
| Composer preview | `src/modules/composer_v2/components/SocialPreviews/BlueskyPreview.vue` | How Bluesky posts appear in the Composer preview panel |
| Composer editor box | `src/modules/composer_v2/components/EditorBox/EditorBlueskyBox.vue` | Platform-specific composer options for Bluesky |

**Telegram-specific components to create:**
- `AddTelegram.vue` — connect modal: user adds @contentstudio_bot as admin, then enters @username (public) or invite link (private); backend validates synchronously for public or via `my_chat_member` webhook + Redis for private groups
- `src/modules/integration/store/states/platforms/social/telegram.js` — Vuex store state
- `TelegramPreview.vue` — composer preview showing Telegram-style message layout
- `EditorTelegramBox.vue` — Telegram-specific options panel in Composer (silent mode, disable link preview, etc.)

**Account type icon/asset:**
- `src/assets/img/integration/telegram-icon.svg` and `telegram-rounded.svg` would need to be added (currently present for Bluesky, Threads, etc.)

**Integration page** (`src/modules/integration/components/Integrations.vue`):
- Telegram would need to be added to the platform list/grid alongside existing channels

**Composer account selector:**
- The account selection component (used in Composer) would need Telegram accounts surfaced

**i18n:**
- New translation keys needed in `src/locales/*/integration.json` for Telegram-specific strings
- New keys in `src/locales/*/composer.json` for Telegram composer options

---

### 9. Reusable Components / Services

**Fully reusable as-is:**
- `SocialRepo` — the repository class used by all posting strategies for fetching account data
- `LogsBuilder` — logging used across all platform integrations
- `QueueSlotsHelper` — posting queue management
- The Vuex store state pattern (`bluesky.js` is a clean 30-line template)
- `SocialAccountsDatatable.vue` — the social_v2 datatable already handles multiple platforms

**Needs adaptation (not full rewrite):**
- `BlueskyConnector.php` → adapt to `TelegramConnector.php` (different auth model: bot token vs. app password, but similar structure: `setPayload`, `connect`, `reconnect`, `disconnect`)
- `BlueskyPosting.php` → adapt to `TelegramPosting.php` (same interface: `initializePosting`, `performPosting`, `generatePostingResponse`)
- `AddBluesky.vue` → adapt to `AddTelegram.vue` (different fields: bot connection flow instead of username/password)
- `BlueskyPreview.vue` → adapt to `TelegramPreview.vue` (different visual layout)

---

### 10. Integration Points

**Where Telegram plugs in:**

1. **Settings → Social Accounts**: Add Telegram tile to the social platform grid. Clicking it opens the `AddTelegram.vue` modal.

2. **Composer (account selector)**: Telegram channels/groups appear in the account selector alongside Facebook, Instagram, LinkedIn, etc. Selecting a Telegram account shows the `EditorTelegramBox.vue` options panel and enables `TelegramPreview.vue`.

3. **Planner (calendar)**: Scheduled Telegram posts appear on the calendar with Telegram's branding (blue paper plane icon), same as posts for any other platform.

4. **Publishing engine** (`SocialPosting.php`): Telegram added to the posting dispatcher. `TelegramPosting::performPosting()` is called when a plan includes Telegram account selection.

5. **RSS Automation** (`RssAutomation.php`): Telegram added as a supported destination for RSS auto-posting.

6. **Analytics** (Phase 2): A new `TelegramAnalytics` builder added under `app/Builders/Analytics/` once Telegram's stats API or webhook-based reporting is integrated.

---

### 11. Technical Considerations

**Authentication model:**
- Unlike OAuth platforms (Facebook, Twitter), Telegram uses a **bot token** model. ContentStudio would run a shared bot (e.g., @contentstudio_bot).
- Connection flow: User adds @contentstudio_bot as admin to their channel/group → enters the @username (public) or invite link (private) in ContentStudio → backend validates via `getChat` + `getChatMember` for public channels/groups, or resolves via a `my_chat_member` webhook event cached in Redis (30-min TTL) for private groups → account stored linked to the user's workspace.
- The `TelegramConnector` would implement `ConnectionInterface` but without OAuth redirects — just bot token + chat_id storage.
- Token storage: the `social_integrations` MongoDB collection (used by all platforms) would store `platform_identifier` (chat_id), `platform_name` (channel/group name), `platform_type` ('channel' or 'group'), `access_token` (bot token — shared across workspace or per workspace), `workspace_id`, `user_id`.

**Posting architecture:**
- `TelegramPosting` implements `PostingInterface` (same as `BlueskyPosting`, `ThreadsPosting`)
- Uses Laravel's `Http` facade (Guzzle) to call `https://api.telegram.org/bot{token}/sendMessage` (or `sendPhoto`, `sendVideo`, etc.)
- Content type routing: text-only → `sendMessage`; single image → `sendPhoto`; single video → `sendVideo`; multiple media → `sendMediaGroup`
- Error handling: 429 (rate limit) → queue retry with `Retry-After` delay; 400/403 (bot not admin, account disconnected) → mark account as invalid, trigger validity notification to user

**Plans model changes:**
- `telegram_sharing_details` field in `Plans` model `$fillable`
- `account_selection.telegram` array in the plan document structure
- `config/social_platforms.php` → add `'telegram.platform_identifier'` to `account_selection_fields`

**Queue/jobs:**
- ContentStudio already uses Laravel Horizon + queued jobs for publishing. Telegram posting would use the same `PublishPostJob` dispatcher, no new job types needed.

**Config:**
```php
// config/integrations.php → social_integrations
'telegram' => [
    'api_url' => 'https://api.telegram.org/bot',
    'bot_token' => env('TELEGRAM_BOT_TOKEN'),
    'webhook_url' => env('TELEGRAM_WEBHOOK_URL'),
],
```

**Rate limiting:**
- Telegram allows 1 message/sec per chat and 30 messages/sec across all chats. For ContentStudio's scale, this is not a concern for individual user posting, but the queue system should enforce a minimum 1-second gap between messages to the same chat. This can be handled via the existing `QueueSlotsHelper`.

**Media uploads:**
- Files up to 50 MB can be sent via bot API. ContentStudio's media library handles files in this range. No special handling needed unless supporting very large video files (which would require the local Bot API server setup — out of scope for v1).

**Analytics (Phase 2):**
- The Telegram Bot API does not expose post view counts or subscriber analytics. To show analytics, ContentStudio would need to use Telegram's `getChatMembersCount` (for subscriber count) and potentially leverage channel forwarding stats.
- Alternatively, for channels where the bot is admin, the channel's native "Statistics" (available via Telegram app for channels with 500+ subscribers) could be surfaced via a future API endpoint if Telegram exposes it.
- For v1: basic "post sent" confirmation and failure notification is sufficient.

---

### 12. Mobile (iOS/Android) Impact

Telegram integration is a **web-first feature** in v1. The mobile apps (iOS/Android) would be impacted in the following ways:

- **Account selector**: If a user has Telegram accounts connected, they should appear in the mobile Composer's account selector, allowing users to include Telegram when scheduling from mobile.
- **Planner calendar**: Telegram posts should be visible on the mobile planner calendar with the Telegram icon.
- **No Telegram-specific options panel needed on mobile for v1**: The advanced Telegram options (silent mode, inline buttons, etc.) can be web-only in v1, similar to how some platform-specific features are web-only.

Separate `[iOS]` and `[Android]` stories should be created for displaying Telegram accounts in the mobile account selector and showing Telegram posts in the mobile planner.

---

*Sources used: publer.com/integrations/telegram, postly.ai/telegram, blog.nuelink.com/social-media-tools-with-telegram-support, core.telegram.org/bots/api, limits.tginfo.me, onlysocial.io/platforms/telegram, help.zoho.com (Telegram inbox), socialbee.com/universal-posting, slashdot.org/software/social-media-management/for-telegram, contentstudio.frill.co (user feature requests)*
