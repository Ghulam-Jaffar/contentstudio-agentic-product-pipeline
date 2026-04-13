# Telegram Integration — Epic & Stories

**Feature:** Telegram Integration for ContentStudio  
**Last Updated:** April 6, 2026  
**Pipeline Step:** 04 — Epic & Stories

---

## Epic

**Title:** Telegram Integration

**Description:**

ContentStudio currently supports publishing to Facebook, Instagram, LinkedIn, X (Twitter), TikTok, Pinterest, YouTube, Threads, and Bluesky — but not Telegram. With Telegram surpassing 1 billion monthly active users in 2025, this gap is a growing competitive disadvantage: of ContentStudio's 10 direct competitors, only Publer has full native Telegram integration, and it is one of the most-upvoted missing features among ContentStudio users.

This epic delivers v1 Telegram integration: users connect their own Telegram bot token (obtained from @BotFather) through a 3-step guided modal that validates the token, discovers channels/groups the bot already has admin access to, and saves the selected accounts. From the Composer, users can schedule text, image, video, album, and PDF document posts, with Telegram-specific options (silent message, disable link preview, pin message) and a live Telegram-style preview. Telegram posts appear on the Planner calendar. Settings → Social Accounts shows the Telegram card for managing connected chats.

The user's own bot model (vs. a shared ContentStudio bot) gives each workspace their own named bot, eliminates shared-bot rate-limit risks, and is the enterprise-grade approach. Implementation follows the same architecture as recently added platforms (Bluesky, Threads): a `TelegramConnector` for account management, a `TelegramPosting` strategy for publishing, and matching frontend components following the Bluesky reference pattern.

---

## Stories

---

### Story 1: [BE] Create Telegram account integration infrastructure

**Group:** Backend  
**Project:** Web App  
**Priority:** High (P0)  
**Product Area:** Integrations  
**Skill Set:** Backend  
**Story Type:** feature

---

**Description:**

As a backend engineer, I need to build the foundational Telegram integration infrastructure so that the ContentStudio platform can store, retrieve, and manage Telegram accounts (channels and groups) in the same way it manages all other social platform accounts.

This story covers: adding Telegram to the platform config, creating the `TelegramAccounts` model (MongoDB, `social_integrations` collection), creating the `TelegramRepo` repository implementing `SocialPlatformsInterface`, and registering Telegram in the integration connector router (`Connector.php`). Note: unlike other integrations, there is no global Telegram bot token — each workspace stores its own user-provided bot token in `access_token` (encrypted).

---

**Workflow:**

1. Developer adds `telegram` to `config/social_platforms.php` → `platforms` array and adds `telegram.platform_identifier` to `account_selection_fields`.
2. Developer adds a `telegram` config entry to `config/integrations.php` → `social_integrations` section: `api_url` (`https://api.telegram.org/bot`). No global bot token — tokens are per-account.
3. Developer creates `app/Models/Integrations/Platforms/Social/TelegramAccounts.php`:
   - Extends Eloquent, uses `social_integrations` MongoDB collection.
   - Global scope filters by `platform_type = 'telegram'`.
   - Fillable fields: `telegram_id`, `chat_id`, `bot_username`, `platform_identifier`, `platform_name`, `platform_type` (channel/group/supergroup), `access_token` (encrypted per-account bot token), `platform_logo`, `image`, `name`, `user_id`, `workspace_id`, `added_by`, `state`, `validity`, `validity_status`, `invalid_tries`, `QueueSlots`, `user_details`.
   - `access_token` stored encrypted (same pattern as `TwitterAccounts`).
4. Developer creates `app/Repository/Integrations/Platforms/Social/TelegramRepo.php`:
   - Implements `SocialPlatformsInterface`.
   - Methods: `getItems($filters)`, `getItem()`, `fetchQueueSlots($filters)`, `updateQueueSlots()`.
5. Developer registers `TelegramConnector` in `app/Strategy/Integrations/Connector.php` → `case 'telegram': $this->connector = new TelegramConnector(); break;`
6. Developer adds `telegram_sharing_details` to `$fillable` in `app/Models/Publish/Planner/Plans.php`.

---

**Acceptance criteria:**

- [ ] `config/social_platforms.php` includes `'telegram'` in the `platforms` array
- [ ] `config/social_platforms.php` includes `'telegram.platform_identifier'` in `account_selection_fields`
- [ ] `config/integrations.php` includes a `telegram` key in `social_integrations` with `api_url`; no global bot token
- [ ] `TelegramAccounts` model exists at `app/Models/Integrations/Platforms/Social/TelegramAccounts.php`
- [ ] `TelegramAccounts` global scope filters records by `platform_type = 'telegram'`
- [ ] `access_token` is encrypted on write and decrypted on read (same pattern as Twitter)
- [ ] `TelegramRepo` exists at `app/Repository/Integrations/Platforms/Social/TelegramRepo.php` and implements `SocialPlatformsInterface`
- [ ] `TelegramRepo::getItems()` returns Telegram accounts for a given workspace, paginated
- [ ] `TelegramRepo::getItem()` returns a single Telegram account by chat_id and workspace_id
- [ ] `Connector.php` routes `'telegram'` to `TelegramConnector` without breaking existing platform routing
- [ ] `Plans` model `$fillable` includes `telegram_sharing_details`
- [ ] Unit tests: `TelegramRepo::getItems()` returns only Telegram accounts (not other platforms)

---

**Mock-ups:** N/A — backend only

---

**Impact on existing data:**

- `social_integrations` MongoDB collection: New documents with `platform_type: 'telegram'` will be added. No changes to existing documents.
- `plans` MongoDB collection: `telegram_sharing_details` field added to `$fillable`; no schema migration required. Existing plan documents unaffected.

---

**Impact on other products:** No impact on existing platform integrations. Chrome extension, mobile apps: No impact at this stage.

---

**Dependencies:** None — this is the foundational story.

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (backend only)
- [ ] Multilingual support — N/A (no user-facing strings)
- [ ] UI theming support — N/A (backend only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 2: [BE] Implement Telegram bot token connection API

**Group:** Backend  
**Project:** Web App  
**Priority:** High (P0)  
**Product Area:** Integrations  
**Skill Set:** Backend  
**Story Type:** feature

---

**Description:**

As a user, I want to connect my Telegram channels and groups to ContentStudio by providing my own Telegram bot token so that I can schedule and publish posts to my channels without using a shared bot.

This story implements the server-side API for the 3-step connection flow: validating the user's bot token, discovering chats the bot is already admin in, manually validating individual chats, saving the selected chats as connected accounts, and removing individual chats.

**API Endpoints:**
- `POST /telegram/validate-bot` — Validate a bot token and return bot info
- `POST /telegram/discover-chats` — Discover all chats the bot has admin access to
- `POST /telegram/validate-chat` — Validate a manually entered chat (by @username or chat ID)
- `POST /telegram/add-chats` — Save selected chats as connected social accounts
- `DELETE /telegram/remove-chat/{id}` — Remove a connected Telegram chat

---

**Workflow:**

**POST /telegram/validate-bot** — Step 1

1. Frontend sends `{ bot_token: "123456:ABC...", workspace_id: "..." }`.
2. Backend calls `GET https://api.telegram.org/bot{token}/getMe`.
3. On success (`ok: true`): returns `{ bot_id, bot_name, bot_username }`.
4. On failure (`ok: false` or HTTP error): returns HTTP 422 `{ error: "invalid_token", message: "Invalid bot token. Please check and try again." }`.

**POST /telegram/discover-chats** — Step 2

1. Frontend sends `{ bot_token, workspace_id }`.
2. Backend calls `GET /getUpdates?limit=100&allowed_updates=["my_chat_member"]` to retrieve recent `my_chat_member` updates where the bot was added as admin. Also calls `GET /getMyCommands` or other available endpoints to enumerate known chats.
3. For each chat discovered, calls `GET /getChatMember?chat_id={chat_id}&user_id={bot_id}` to verify the bot still has admin permissions.
4. Returns array of chats: `{ chat_id, title, type ("channel"/"group"/"supergroup"), member_count, username (if public) }`. Returns empty array if none found.
5. If the bot token was just used in `validate-bot`, the `bot_id` should be cached in the request session or passed in the body to avoid calling `getMe` again.

**POST /telegram/validate-chat** — Step 2 (manual add)

1. Frontend sends `{ bot_token, chat_identifier: "@channelusername" | "chat_id", workspace_id }`.
2. Backend resolves the chat:
   - If `chat_identifier` starts with `@` or is a username string: calls `getChat(@username)` to resolve metadata.
   - If `chat_identifier` is a numeric string: uses it directly as `chat_id`.
3. Calls `getChatMember` with the bot's own user ID to verify admin status.
4. On success: returns `{ chat_id, title, type, member_count, username }`.
5. On error:
   - Bot not admin: HTTP 422 `{ error: "bot_not_admin", message: "The bot is not an admin of this chat. Please add it as an administrator and try again." }`
   - Chat not found: HTTP 422 `{ error: "chat_not_found", message: "Chat not found. Check the username or ID and try again." }`

**POST /telegram/add-chats** — Step 3

1. Frontend sends `{ bot_token, workspace_id, chats: [{ chat_id, title, type, username, member_count }] }`.
2. Backend validates subscription limits. If exceeded: HTTP 402 `{ error: "subscription_limit_exceeded" }`.
3. For each chat, creates or updates a record in `social_integrations`:
   - `platform_type: 'telegram'`
   - `chat_id`
   - `access_token` = encrypted `bot_token`
   - `bot_username` (from getMe result)
   - `platform_name` = chat title
   - `platform_identifier` = @username if public, chat_id as string if private
   - `chat_type` = 'channel' / 'group' / 'supergroup'
   - `member_count`
   - `state: 'Added'`, `validity: 'valid'`
   - `workspace_id`, `user_id`
4. If a chat is already connected to this workspace (same chat_id + workspace_id): updates the existing record rather than creating a duplicate.
5. Returns HTTP 200 with array of created/updated account objects.

**DELETE /telegram/remove-chat/{id}**

1. Finds the account by ID, verifies it belongs to the requesting workspace.
2. Sets `state: 'Removed'` and removes it from the active account list.
3. Cancels or marks as cancelled any pending scheduled posts for this account (same behaviour as other platform disconnect flows).
4. Returns HTTP 200.

---

**Acceptance criteria:**

- [ ] `POST /telegram/validate-bot` calls `getMe` and returns bot info on success; returns `error: "invalid_token"` on failure
- [ ] `POST /telegram/discover-chats` returns an array of chats the bot has admin access to (or empty array if none)
- [ ] `POST /telegram/discover-chats` verifies bot admin status per chat via `getChatMember`; chats where bot lost admin are excluded
- [ ] `POST /telegram/validate-chat` accepts @username or numeric chat ID; returns chat info if bot is admin
- [ ] `POST /telegram/validate-chat` returns `error: "bot_not_admin"` if bot is not an admin of the chat
- [ ] `POST /telegram/validate-chat` returns `error: "chat_not_found"` if chat cannot be resolved
- [ ] `POST /telegram/add-chats` creates `social_integrations` records for each selected chat with correct fields
- [ ] `access_token` (bot token) stored encrypted; decrypts correctly for publish-time use
- [ ] `POST /telegram/add-chats` updates existing record on duplicate (same chat_id + workspace_id); no duplicate accounts created
- [ ] `POST /telegram/add-chats` returns HTTP 402 with `error: "subscription_limit_exceeded"` when plan limits are reached
- [ ] `DELETE /telegram/remove-chat/{id}` sets `state: 'Removed'` and removes account from active list
- [ ] `DELETE /telegram/remove-chat/{id}` returns 403 if account does not belong to requesting workspace
- [ ] All endpoints are authenticated (workspace-scoped, authorized request only)
- [ ] Unit tests: validate-bot success, validate-bot invalid token, validate-chat success, validate-chat bot_not_admin, validate-chat chat_not_found, add-chats creates correctly, add-chats updates on duplicate, add-chats subscription limit, remove-chat success

---

**Mock-ups:** N/A — backend only

---

**Impact on existing data:**

- `social_integrations` collection: new Telegram account documents created on successful connection. No changes to existing documents.

---

**Impact on other products:**

- Chrome extension: No impact.
- Mobile apps: No impact at this stage — connection flow is web-only.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure**

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (backend only)
- [ ] Multilingual support — N/A (backend only; error codes are machine-readable, user-facing messages in frontend)
- [ ] UI theming support — N/A (backend only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 3: [BE] Implement Telegram post publishing (text, image, video, album, PDF)

**Group:** Backend  
**Project:** Web App  
**Priority:** High (P0)  
**Product Area:** Integrations  
**Skill Set:** Backend  
**Story Type:** feature

---

**Description:**

As a user, I want to publish text, image, video, album, and PDF document posts to my Telegram channel or group from ContentStudio so that my scheduled content is delivered automatically without manual effort.

This story implements `TelegramPosting.php`, registers it in the `Posting.php` strategy router and `SocialPosting.php` dispatcher, and handles five content types: text-only (`sendMessage`), single image (`sendPhoto`), single video (`sendVideo`), albums of 2–10 items (`sendMediaGroup`), and PDF documents (`sendDocument`). It handles Telegram options (silent message, disable link preview, pin message), error classification, account validity invalidation on permission errors, and retry-on-rate-limit behaviour.

---

**Workflow:**

1. A plan is scheduled for publishing. At the scheduled time, `PlanPostingJob` is dispatched.
2. `SocialPosting::processSocialPosting($plan)` detects `account_selection.telegram` is set and calls `(new Posting('Telegram', $plan))->initializePosting()->performPosting()`.
3. `Posting.php` routes to `new TelegramPosting($plan)`.
4. `TelegramPosting::performPosting()` executes:
   a. Fetches the Telegram account from `TelegramRepo::getItem()` using `chat_id` from `telegram_sharing_details`.
   b. Decrypts the bot token from `access_token`.
   c. Determines content type:
      - Text only → `sendMessage` with `text`, Markdown formatting if applicable.
      - Single image → `sendPhoto` with `photo` (file URL or upload) and `caption`.
      - Single video → `sendVideo` with `video` and `caption`.
      - 2–10 media items → `sendMediaGroup` with array of `InputMedia` objects.
      - PDF document → `sendDocument` with `document` (file URL or upload) and `caption`.
   d. Applies options from `telegram_options`:
      - `silent_message: true` → `disable_notification: true` (all message types)
      - `disable_link_preview: true` → `disable_web_page_preview: true` (`sendMessage` only)
      - `pin_message: true` → after successful post, calls `pinChatMessage` with the returned `message_id`
   e. Sends the API request to `https://api.telegram.org/bot{token}/{method}`.
   f. On success (`ok: true`): stores `message_id` from response. If `pin_message: true`, calls `pinChatMessage`; if pin fails, logs a non-blocking warning (post remains Published). Marks post as Published.
   g. On HTTP 429 (rate limit): reads `Retry-After` header, re-queues with that delay. Up to 3 retries, then marks Failed.
   h. On HTTP 400/403 (bot not admin, chat not found, token invalid): marks account as `validity: 'invalid'`, `invalid_tries++`; dispatches account-invalid notification; marks post Failed with error message.
   i. On other errors: marks post Failed with the Telegram error description.
5. Posting response is merged into the overall `processSocialPosting` response.

---

**Acceptance criteria:**

- [ ] Text-only posts published via `sendMessage`; full text appears in the channel/group
- [ ] Single image posts published via `sendPhoto`; caption (≤ 1,024 chars) appears
- [ ] Single video posts published via `sendVideo`; caption appears
- [ ] Albums of 2–10 items published via `sendMediaGroup`; all items appear as a single album
- [ ] PDF documents published via `sendDocument`; document arrives with caption
- [ ] `silent_message: true` results in `disable_notification: true` in API call (all message types)
- [ ] `disable_link_preview: true` results in `disable_web_page_preview: true` on `sendMessage` calls
- [ ] `pin_message: true` calls `pinChatMessage` with the `message_id` after a successful post
- [ ] If `pinChatMessage` fails, the post is still marked Published; a non-blocking warning is stored in the posting log
- [ ] HTTP 429 triggers retry after `Retry-After` delay; after 3 retries post is marked Failed
- [ ] HTTP 403 (bot removed as admin / token invalid) marks account `validity: 'invalid'` and post as Failed with message: "The bot no longer has permission to post in this channel. Please reconnect your Telegram account in Settings."
- [ ] HTTP 400 marks post as Failed with the Telegram error description
- [ ] Telegram posting does not affect publishing to other platforms in the same plan (failures isolated)
- [ ] `SocialPosting.php` calls `TelegramPosting` only when `account_selection.telegram` is set
- [ ] `Posting.php` correctly routes `'Telegram'` to `TelegramPosting`
- [ ] Unit tests: text posting, image posting, video posting, PDF posting, silent message, pin message, pin failure non-blocking, rate-limit retry, 403 handling

---

**Mock-ups:** N/A — backend only

---

**Impact on existing data:**

- `social_integrations` collection: `validity` and `invalid_tries` updated on failed posts.
- Publish logs: Telegram posting results logged alongside other platforms.

---

**Impact on other products:** `SocialPosting.php` change is additive. Chrome extension: No impact. Mobile apps: Publishing is server-side.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure**
- Depends on: **[BE] Implement Telegram bot token connection API** (accounts must exist before publishing can be tested)

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (backend only)
- [ ] Multilingual support — N/A (error messages are English-only in v1)
- [ ] UI theming support — N/A (backend only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 4: [BE] Add Telegram as RSS automation publishing destination

**Group:** Backend  
**Project:** Web App  
**Priority:** Medium (P1)  
**Product Area:** Automation  
**Skill Set:** Backend  
**Story Type:** feature

---

**Description:**

As a content publisher, I want to use ContentStudio's RSS automation feature to automatically post items from an RSS feed to my Telegram channel so that I can distribute content to my Telegram audience without manual effort.

This story extends the existing RSS-to-post automation engine to include Telegram as a supported publishing destination. No changes to the RSS parsing or feed management logic — only the destination list and publishing dispatch need to be extended.

---

**Workflow:**

1. User opens Automation → RSS Feeds and creates or edits an RSS automation.
2. In the "Post to" destination selector, connected Telegram channels/groups appear alongside other social accounts.
3. User selects a Telegram channel as a destination and saves the automation.
4. At the configured RSS check interval, the RSS job fetches new feed items.
5. For each new item, the job creates a plan document with `telegram_sharing_details` populated and `account_selection.telegram` set.
6. The plan is dispatched to the publishing queue, which calls `TelegramPosting::performPosting()`.
7. The RSS item is published to the Telegram channel as a text post (title + URL, optionally with featured image if present in the feed item).

---

**Acceptance criteria:**

- [ ] Connected Telegram channels/groups appear in the RSS automation destination selector
- [ ] A plan document is correctly created with `account_selection.telegram` and `telegram_sharing_details` populated
- [ ] `TelegramPosting::performPosting()` is called for the plan
- [ ] RSS items published to Telegram with post title, truncated description (≤ 4,096 chars), and source URL
- [ ] If RSS item has a featured image and automation is configured to include images, post is published via `sendPhoto` with title+URL as caption
- [ ] Failed Telegram publications from RSS automation are logged and surfaced in the automation error log
- [ ] Publishing to other platforms in the same RSS automation is not affected

---

**Mock-ups:** N/A — backend only

---

**Impact on existing data:** Plans created by RSS automation will include `telegram_sharing_details` when Telegram is selected. No changes to existing RSS automation configurations.

---

**Impact on other products:** No impact on Chrome extension, iOS, or Android apps.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure**
- Depends on: **[BE] Implement Telegram post publishing (text, image, video, album, PDF)**

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (backend only)
- [ ] Multilingual support — N/A (backend only)
- [ ] UI theming support — N/A (backend only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 5: [FE] Add Telegram bot token connection flow in Settings → Social Accounts

**Group:** Frontend  
**Project:** Web App  
**Priority:** High (P0)  
**Product Area:** Integrations  
**Skill Set:** Frontend  
**Story Type:** feature

---

**Description:**

As a user, I want to connect my Telegram channels and groups to ContentStudio by providing my own Telegram bot token so that I can start scheduling posts to Telegram alongside my other social accounts.

This story implements the Telegram card in the integrations settings page, the 3-step "Connect Telegram" modal (bot token validation → chat selection → confirm & connect), success/error states, and the connected accounts management view (add more, remove individual chats).

---

**Workflow:**

**Step 1 — Enter Bot Token**

1. User navigates to **Settings → Social Accounts** (or the integrations settings page).
2. User sees a **Telegram** card in the platform grid and clicks it.
3. The **"Connect Telegram"** modal opens at Step 1 of 3. It contains:
   - A field labeled **"Bot Token"** with placeholder: *"Paste your bot token from @BotFather"*
   - A collapsible **"How to create a Telegram Bot"** section with 5 steps: (1) Open Telegram, search for @BotFather. (2) Send `/newbot`. (3) Choose a display name. (4) Choose a username ending in `bot`. (5) Copy the token BotFather gives you.
   - A **"Validate"** primary button. Disabled until the input is non-empty.
4. User pastes their bot token and clicks "Validate".
5. While `POST /telegram/validate-bot` is in-flight, the button shows a loading spinner and is disabled.
6. On success: a bot info card appears (bot name, @username). Modal auto-advances to Step 2.
7. On error: inline error below the input: "Invalid bot token. Please check and try again." Modal stays on Step 1.

**Step 2 — Select Channels & Groups**

1. On mount (or after advancing from Step 1), `POST /telegram/discover-chats` is called. A loading spinner shows.
2. Discovered chats appear as a list with: chat title (bold), type badge ("Channel" / "Group" / "Supergroup"), member count (if available), checkbox for selection.
3. A **"Select all / Deselect all"** toggle appears above the list.
4. A separator + **"Can't find your channel or group?"** section below the list:
   - Instructions: "Make sure the bot is added as an admin, then enter the channel/group username or ID below"
   - Input with placeholder: *"@channelusername or chat ID"*
   - **"Add"** button → calls `POST /telegram/validate-chat`. On success: appends chat to list with checkbox pre-selected. On error: inline error below the input with the API message.
5. **Empty state** (no chats, none manually added): Icon + "No channels or groups found" + instructions + **"Refresh"** button (re-calls discover endpoint).
6. **"Next"** button is disabled until at least 1 chat is checked. On click: advances to Step 3.

**Step 3 — Confirm & Connect**

1. Modal shows "Confirm connection" — summary list of selected chats with title and type badge.
2. Note: "Each channel/group will count as one social account in your plan."
3. **"Connect"** button (primary, with loading state). Calls `POST /telegram/add-chats`.
4. On success: success toast "Telegram channels/groups connected successfully", modal closes, social accounts list refreshes.
5. On subscription limit error: limit-exceeded message with upgrade link.
6. On other error: error toast.

**Manage Connected Chats**

- Each connected Telegram chat appears as a row: chat title, type badge, bot username (e.g., "via @MyContentBot"), **"Remove"** button.
- **"Remove"** → confirmation dialog → `DELETE /telegram/remove-chat/{id}` → row removed + success toast.
- **"Add more"** button → re-opens the modal (starting at Step 1 if no token on file, or Step 2 if same bot).
- Accounts with `validity: 'invalid'` show a yellow warning badge and a **"Reconnect"** button that reopens the modal.

---

**Acceptance criteria:**

- [ ] Telegram card appears in the integrations settings page with the Telegram logo
- [ ] Clicking the card opens the "Connect Telegram" modal at Step 1
- [ ] Step 1: Bot Token field with placeholder, collapsible "How to create a Telegram Bot" guide (5 steps), "Validate" button disabled until input non-empty
- [ ] Step 1: Validate button shows loading spinner while `POST /telegram/validate-bot` is in-flight
- [ ] Step 1: On success, bot info card (bot name, @username) appears and modal advances to Step 2
- [ ] Step 1: On error, inline error: "Invalid bot token. Please check and try again." Modal stays on Step 1
- [ ] Step 2: Loading spinner shown while `POST /telegram/discover-chats` runs
- [ ] Step 2: Discovered chats shown as a list with title, type badge, member count, checkbox
- [ ] Step 2: Select all / Deselect all toggle works
- [ ] Step 2: Manual chat add — input + "Add" button → calls `POST /telegram/validate-chat` → appends to list with checkbox pre-selected
- [ ] Step 2: Manual chat add — inline error shown on API failure
- [ ] Step 2: Empty state shown when no chats discovered and none manually added; "Refresh" button re-calls discover endpoint
- [ ] Step 2: "Next" button disabled until at least 1 chat selected
- [ ] Step 3: Summary list of selected chats with title and type badge
- [ ] Step 3: Plan note: "Each channel/group will count as one social account in your plan."
- [ ] Step 3: "Connect" button calls `POST /telegram/add-chats` with loading state; modal closes on success with success toast
- [ ] Step 3: Subscription limit error shows limit-exceeded message with upgrade link
- [ ] Connected chats list shows: title, type badge, bot username ("via @botname"), Remove button
- [ ] "Remove" shows confirmation dialog; on confirm calls `DELETE /telegram/remove-chat/{id}` and removes row with success toast
- [ ] "Add more" opens the modal
- [ ] Invalid accounts show yellow warning badge and "Reconnect" button
- [ ] All user-facing strings use `$t()` with new keys in `src/locales/*/integration.json`

---

**UI Copy & Components:**

**Step 1 — Modal title:** "Connect Telegram" (Step 1 of 3)  
**Bot Token field label:** "Bot Token"  
**Bot Token placeholder:** "Paste your bot token from @BotFather"  
**Collapsible guide title:** "How to create a Telegram Bot"  
**Guide steps:**
1. "Open Telegram and search for @BotFather"
2. "Send /newbot"
3. "Choose a display name for your bot (e.g., My Brand Bot)"
4. "Choose a username ending in 'bot' (e.g., mybrand_bot)"
5. "Copy the token BotFather gives you and paste it above"

**Step 2 — Modal title:** "Select channels and groups" (Step 2 of 3)  
**Empty state heading:** "No channels or groups found"  
**Empty state instructions:** "1. Add your bot as an admin to a Telegram channel or group. 2. Send any message in that channel/group. 3. Click 'Refresh' below."  
**Manual add label:** "Can't find your channel or group?"  
**Manual add hint:** "Make sure the bot is added as an admin, then enter the username or ID below"  
**Manual add placeholder:** "@channelusername or chat ID"  

**Step 3 — Modal title:** "Confirm connection" (Step 3 of 3)  
**Plan note:** "Each channel/group will count as one social account in your plan."  

**Remove confirmation dialog:**
- **Title:** "Remove [Chat Name]?"
- **Body:** "This will remove [Chat Name] from ContentStudio. Any posts scheduled to this chat that haven't been published yet will be cancelled."
- Confirm: "Yes, Remove" (`Button` destructive variant)
- Cancel: "Cancel" (`Button` secondary variant)

**Type badges:** "Channel" / "Group" / "Supergroup" (use `Badge` component, neutral)  
**Invalid badge:** "Needs reconnection" (`Badge`, warning/yellow)  
**Bot username label:** "via @{botUsername}"  
**Toast — success:** "Telegram channels/groups connected successfully"  
**Toast — remove success:** "[Chat Name] has been removed."  

**Components used:** `Modal`, `Dialog`, `Button`, `Loader`, `Badge`, `CstToast`, `Collapsible` (for How to guide), `Switch` (for Select all toggle)

---

**Empty / Error / Loading states:**

- **Empty state (Step 2, no chats):** Icon + message + instructions + Refresh button
- **Loading (Validate / Connect in-flight):** Button shows `Loader` (small, inline), disabled
- **Inline error (Step 1, invalid token):** Red helper text below input: "Invalid bot token. Please check and try again."
- **Inline error (Step 2, manual add failed):** Red helper text below the manual input with specific API message
- **Generic network error:** "Something went wrong. Please try again."

---

**Mock-ups:** Figma designs to be provided by Design team.

---

**Impact on existing data:** None — purely additive changes. No existing integration UI modified.

---

**Impact on other products:** None on Chrome extension or mobile apps.

---

**Dependencies:**

- Depends on: **[BE] Implement Telegram bot token connection API** (all 5 endpoints must exist)
- Requires Telegram SVG icon asset in `src/assets/img/integration/`

---

**Global quality & compliance:**

- [ ] Mobile responsiveness (Integrations settings page is desktop-first; must not break on tablet)
- [ ] Multilingual support (all strings via `$t()` with new keys in `src/locales/*/integration.json`)
- [ ] UI theming support (default + white-label, design library components used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 6: [FE] Add Telegram to Composer — account selector, preview, options panel, and PDF media mode

**Group:** Frontend  
**Project:** Web App  
**Priority:** High (P0)  
**Product Area:** Composer  
**Skill Set:** Frontend  
**Story Type:** feature

---

**Description:**

As a user, I want to select my Telegram channels and groups in the Composer, see a Telegram-style preview of my post, set Telegram-specific publishing options (silent message, disable link preview, pin message), and choose between images/videos or PDF document media mode — so that I can create, customize, and schedule Telegram posts with confidence from within my existing content workflow.

This story implements:
1. Telegram platform registration in the Composer (initial state, constants, helper mappings)
2. Telegram platform section in `AccountSelectionAside.vue`
3. New `TelegramOptions.vue` — 3-toggle options panel
4. New `TelegramPreview.vue` — Telegram-style post preview in the sidebar
5. `SocialModal.vue` changes — Telegram validation, payload, error aggregation
6. Media type exclusivity (PDF mode vs. images/videos mode) with dynamic character counter
7. Common box status handling (PDF not available in common mode)

---

**Workflow:**

1. User opens the **Composer** (Create New Post / Edit Draft).

**Platform registration & initial state:**
- `composerInitialState.js` has `defaultTelegramSharingDetails` (message, image[], video, document, url, media_type: 'text') and `defaultTelegramOptions` (silent_message: false, disable_link_preview: false, pin_message: false).
- `defaultAccountSelection` includes `telegram: []`.
- `SocialModal.vue` data initializes `telegramOptions` and `telegram_sharing_details` on mount.

**Account selection:**
2. In the account selector (`AccountSelectionAside.vue`), a **Telegram** section appears with connected channels and groups. Each item shows: Telegram logo, chat type indicator ("Channel" or "Group"), chat title, checkbox (using `CstAccountCheckBox`).
3. A "Connect Telegram" link appears if no accounts are connected (links to Settings → Social Accounts).
4. Invalid accounts show a yellow warning icon; hovering via `CstPopup` shows: "This Telegram account needs reconnecting. Go to Settings → Social Accounts to fix it."

**TelegramOptions.vue panel:**
5. When at least one Telegram account is selected, the `TelegramOptions.vue` panel appears (open by default on first selection) with 3 `Switch` toggles:
   - **Silent message** — Label: "Silent message", Tooltip: "Post without sending a notification to channel/group members"
   - **Disable link preview** — Label: "Disable link preview", Tooltip: "Prevent automatic link preview expansion in the post"
   - **Pin message** — Label: "Pin message", Tooltip: "Pin this message in the channel/group after posting"
6. The panel collapses when all Telegram accounts are deselected.

**TelegramPreview.vue:**
7. The right preview panel renders a Telegram-style message bubble: channel avatar + channel name (bold), message content, media (image/video thumbnail or document icon + filename), timestamp "Just now".
8. A **character counter** appears below the preview: `X / 4096` for text-only posts, switches to `X / 1024` when any media is attached. Counter turns yellow at 80% and red at 100%.
9. When a URL is present and "Disable link preview" is off, the preview shows a simplified link preview card.
10. When pin_message is on, a pin indicator is shown in the preview. When silent_message is on, a mute indicator is shown.

**Media mode (PDF vs images/videos):**
11. When Telegram is selected and `common_box_status` is `false` (per-platform mode), the Telegram-specific editor area shows a media type selector: **"Images / Videos"** | **"PDF Document"** (segmented control / radio group).
12. **Images / Videos mode** (default): standard image/video upload (drag & drop, media library, URL), up to 10 items, mixed OK. On attachment, character counter switches to 1024 limit.
13. **PDF Document mode**: single .pdf upload, up to 50 MB. After upload shows: PDF icon + filename + file size + "Replace" and "Remove" buttons. Character counter switches to 1024 limit.
14. When PDF mode is active, image/video upload is disabled. When images/videos are present, PDF upload is disabled.
15. When `common_box_status` is `true`: PDF mode is not available for Telegram. A hint is shown: "Switch to platform-specific mode to attach a PDF for Telegram." Images/videos from the common editor are used for Telegram.

**Validation (SocialModal.vue):**
16. `telegramErrors` computed property checks:
    - Character limit (4096 text-only, 1024 with media)
    - PDF + images/videos together not allowed
    - Max 10 images/videos (warning)
    - Missing content (no message, no image, no video, no document)
17. Errors/warnings added to `socialPostErrors()` when Telegram accounts are selected.
18. `validateSocialShare()` blocks submission if Telegram has no content.

**Payload:**
19. `generatePayload()` includes `telegram_sharing_details` and `telegram_options` in the plan payload.

**Edit post flow:**
20. `editPublication()` populates `telegramOptions` and `telegram_sharing_details` from fetched plan data.
21. When editing an album post, a warning banner is shown: "Telegram does not support editing album posts. Changes will require deleting and re-posting."

---

**Acceptance criteria:**

- [ ] `composerInitialState.js` exports `defaultTelegramSharingDetails` with fields: message, image[], video, document, url, media_type ('text')
- [ ] `composerInitialState.js` exports `defaultTelegramOptions` with fields: silent_message, disable_link_preview, pin_message (all false)
- [ ] `defaultAccountSelection` includes `telegram: []`
- [ ] `SUPPORTED_PLATFORMS` enum in `composer.js` includes `TELEGRAM: 'telegram'`
- [ ] `'telegram'` added to `MULTIMEDIA_ALLOWED_PLATFORMS` in `composer.js`
- [ ] `socialIntegrationsConfigurations.telegram` validation config added to `api-utils.js` (characters: 4096, caption_characters: 1024, image.max_size: 10MB, video.max_size: 50MB, document.max_size: 50MB, pdf only)
- [ ] Telegram section appears in `AccountSelectionAside.vue` with connected channels/groups listed
- [ ] Each Telegram account shows: Telegram logo, type indicator (Channel/Group), chat title, checkbox
- [ ] "Connect Telegram" link shown when no accounts connected
- [ ] Invalid accounts show yellow warning icon with reconnect tooltip via `CstPopup`
- [ ] `TelegramOptions.vue` appears when Telegram account is selected; hidden when deselected
- [ ] Silent message toggle: correct label, tooltip, defaults to off; `silent_message` saved in telegramOptions
- [ ] Disable link preview toggle: correct label, tooltip, defaults to off; `disable_link_preview` saved
- [ ] Pin message toggle: correct label, tooltip, defaults to off; `pin_message` saved
- [ ] `TelegramPreview.vue` renders channel avatar, name, message content, media thumbnail, timestamp
- [ ] Character counter shows `X / 4096` for text-only; switches to `X / 1024` when media attached
- [ ] Counter turns yellow at 80%; red at 100%
- [ ] Link preview card appears in preview when URL present and disable_link_preview is off
- [ ] Pin indicator in preview when pin_message is on; mute indicator when silent_message is on
- [ ] Media mode selector (Images/Videos | PDF Document) shown in per-platform Telegram editor
- [ ] PDF mode: accepts .pdf files only, max 50MB; shows filename + filesize + Replace/Remove
- [ ] PDF attached → image/video upload disabled; images/videos attached → PDF upload disabled
- [ ] In common mode (common_box_status: true): PDF mode unavailable; hint shown
- [ ] `telegramErrors` catches: over character limit (correct limit for content type), PDF + media mix, missing content
- [ ] Warning shown for >10 images/videos: "Telegram allows a maximum of 10 images/videos per post"
- [ ] `generatePayload()` includes `telegram_sharing_details` and `telegram_options`
- [ ] Edit mode: `telegramOptions` and `telegram_sharing_details` restored from plan data
- [ ] Album edit warning banner: "Telegram does not support editing album posts. Changes will require deleting and re-posting."
- [ ] All user-facing strings use `$t()` with keys in `src/locales/*/composer.json`

---

**UI Copy & Components:**

**"Telegram Options" panel header:** "Telegram Options" — use `Collapsible` (`@contentstudio/ui`)

**Silent message toggle:** Label: "Silent message" · Subtext: "Post without push notification" · Tooltip (`CstPopup`): "Post without sending a notification to channel/group members"

**Disable link preview toggle:** Label: "Disable link preview" · Subtext: "Don't expand URLs in this post" · Tooltip (`CstPopup`): "Prevent automatic link preview expansion in the post"

**Pin message toggle:** Label: "Pin message" · Subtext: "Pin after posting" · Tooltip (`CstPopup`): "Pin this message in the channel/group after posting"

**Character counter:** `[used] / [limit] characters` — yellow at 80%, red at 100%

**PDF upload area (PDF mode active):** "Drop a PDF here or click to upload" · accept=".pdf" · max 50 MB  
**PDF attached state:** PDF icon + filename + formatted file size + "Replace" + "Remove" buttons

**Common box PDF hint:** "Switch to platform-specific mode to attach a PDF for Telegram" (use `Alert` or inline info text)

**Album edit warning:** (use `Alert`, warning variant): "Telegram does not support editing album posts. Changes will require deleting and re-posting."

**No accounts empty state:** "No Telegram accounts connected. [Connect one →]" (link to Settings)

**Components used:** `Switch`, `Collapsible`, `Alert`, `Button`, `Loader`, `Icon`, `Badge`, `CstPopup`, `CstAccountCheckBox`

**i18n keys to add** (`src/locales/en/composer.json` + 10 other locale files):

```json
{
  "composer.telegram_options.title": "Telegram Options",
  "composer.telegram_options.silent_message": "Silent message",
  "composer.telegram_options.silent_message_tooltip": "Post without sending a notification to channel/group members",
  "composer.telegram_options.disable_link_preview": "Disable link preview",
  "composer.telegram_options.disable_link_preview_tooltip": "Prevent automatic link preview expansion in the post",
  "composer.telegram_options.pin_message": "Pin message",
  "composer.telegram_options.pin_message_tooltip": "Pin this message in the channel/group after posting",
  "composer.errors.telegram_message_too_long": "Message exceeds Telegram's 4096 character limit",
  "composer.errors.telegram_caption_too_long": "Caption exceeds Telegram's 1024 character limit for media posts",
  "composer.errors.telegram_document_with_media": "PDF posts cannot include images or videos",
  "composer.errors.telegram_no_content": "Telegram requires a message, image, video, or document",
  "composer.errors.telegram_album_edit": "Telegram does not support editing album posts. Delete and re-post to make changes.",
  "composer.warnings.telegram_max_images": "Telegram allows a maximum of 10 images/videos per post"
}
```

---

**Empty / Error / Loading states:**

- **No Telegram accounts connected:** "No Telegram accounts connected. [Connect one →]" in account selector
- **Loading (account avatar fetch):** `Loader` small shown briefly in preview panel; fallback to Telegram logo
- **Validation error (over limit):** Red character counter + error message in composer error summary
- **Album edit warning:** `Alert` (warning) in composer editor area when editing an existing album post

---

**Mock-ups:** Figma designs to be provided by Design team.

---

**Impact on existing data:** None — reads `telegram_sharing_details` from plan documents. No changes to existing plans.

---

**Impact on other products:** None directly. All changes are conditional on Telegram being selected.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure**
- Depends on: **[FE] Add Telegram bot token connection flow in Settings → Social Accounts** (accounts must be connectable before they appear in the selector — logically required for QA)

---

**Global quality & compliance:**

- [ ] Mobile responsiveness (Composer is desktop-first; Telegram section should not break on tablet)
- [ ] Multilingual support (all strings via `$t()` in `src/locales/*/composer.json`; i18n keys for all 11 locale files listed above)
- [ ] UI theming support (default + white-label; `Switch`, `Collapsible`, `Alert`, `Icon` from `@contentstudio/ui`)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 7: [FE] Display Telegram posts on the Planner calendar

**Group:** Frontend  
**Project:** Web App  
**Priority:** High (P0)  
**Product Area:** Planner  
**Skill Set:** Frontend  
**Story Type:** feature

---

**Description:**

As a user, I want to see my scheduled and published Telegram posts on the ContentStudio Planner calendar so that I have a complete view of my cross-platform content schedule including Telegram.

This story adds Telegram post tiles to the Planner calendar and list view, correct status display, failed post reason display, and Telegram in the platform filter list.

---

**Workflow:**

1. User navigates to **Planner** (Calendar or List view).
2. Scheduled Telegram posts appear as tiles at their scheduled date/time: Telegram logo, channel/group name (truncated), post text preview (~60 chars), status dot (blue = Scheduled, green = Published, red = Failed).
3. **Telegram** appears in the platform filter list. Filtering by Telegram shows only Telegram posts.
4. Clicking a tile opens the **post detail panel**: platform (Telegram logo + "Telegram"), account name + type badge (Channel/Group), status, scheduled/published time, post content preview.
5. **Failed posts:** Red `Alert` in detail panel with failure reason. A **"Retry"** button re-queues the post.
6. Standard action buttons: Edit, Duplicate, Delete.

---

**Acceptance criteria:**

- [ ] Telegram posts appear on the Planner calendar at scheduled date/time with the Telegram logo
- [ ] Tiles show: Telegram logo, channel/group name (truncated to ~30 chars), status dot
- [ ] "Telegram" appears in the platform filter list; filtering shows only Telegram posts
- [ ] Clicking a tile opens the post detail panel with: platform icon, account name + type badge, status, time, content preview
- [ ] Failed posts show the failure reason in a red `Alert` in the detail panel
- [ ] Failed posts show a "Retry" button (loading state + `Loader` while API call in-flight)
- [ ] After successful retry queue: "✓ Queued for retry" (3 seconds, then reverts)
- [ ] Telegram posts in List view also show the Telegram logo and correct status
- [ ] All strings via `$t()` with keys in `src/locales/*/planner.json`

---

**UI Copy & Components:**

**Failed post `Alert`** (use `Alert`, variant: danger):
- "The bot no longer has permission to post in this channel or group. Please go to Settings → Social Accounts and reconnect your Telegram account."
- "Your Telegram account was disconnected. Please reconnect it in Settings → Social Accounts."
- "This post could not be delivered due to a temporary Telegram error. Please retry below."

**"Retry" button** (use `Button`, primary): Default: "Retry Post" · Loading: `Loader` (x-small) + disabled · Success: "✓ Queued for retry"

**Components used:** `Alert`, `Button`, `Loader`, `Badge`

---

**Mock-ups:** Follows existing Planner calendar tile and detail panel patterns.

---

**Impact on existing data:** None — reads existing plan documents.

---

**Impact on other products:** None — additive Planner changes.

---

**Dependencies:**

- Depends on: **[BE] Implement Telegram post publishing (text, image, video, album, PDF)**
- Depends on: **[FE] Add Telegram to Composer — account selector, preview, options panel, and PDF media mode**

---

**Global quality & compliance:**

- [ ] Mobile responsiveness (Planner is desktop-first; Telegram tiles must render on tablet)
- [ ] Multilingual support (all strings via `$t()` in `src/locales/*/planner.json`)
- [ ] UI theming support (default + white-label; `Alert`, `Button`, `Badge`, `Loader` from `@contentstudio/ui`)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 8: [iOS] Add Telegram accounts to mobile Composer account selector and Planner calendar

**Group:** Product  
**Project:** Mobile  
**Priority:** Medium (P1)  
**Product Area:** iOS Mobile  
**Skill Set:** Frontend  
**Story Type:** feature

---

**Description:**

As a mobile user on iOS, I want to see my connected Telegram channels and groups in the iOS Composer account selector and view Telegram posts on the iOS Planner calendar so that I can manage my Telegram content schedule from my iPhone.

Connection flow and Telegram-specific publishing options (silent message, disable link preview, pin message, PDF document) are out of scope for iOS — these are web-only in v1. AI functionality is web-only and not applicable here.

---

**Workflow:**

1. User opens the ContentStudio iOS app and taps **Create Post**.
2. In the account selector, a **Telegram** section appears with connected channels and groups. Each account shows the Telegram logo, avatar, and name.
3. Tapping selects/deselects the account. With Telegram selected, user writes content and attaches images/videos as usual. Telegram-specific options are not shown on iOS in v1.
4. User schedules the post. It is queued for server-side publishing.
5. User navigates to **Planner**. Telegram posts appear on the mobile calendar with the Telegram logo.
6. Tapping a Telegram event shows: platform (Telegram), account name, status (Scheduled/Published/Failed), scheduled time, content preview.
7. Failed posts show the failure reason (e.g., "The bot no longer has permission to post in this channel").

---

**Acceptance criteria:**

- [ ] Telegram section appears in the iOS Composer account selector with connected channels/groups
- [ ] Each account shows Telegram logo, avatar, and name
- [ ] Tapping an account selects/deselects it correctly
- [ ] Posts with Telegram selected can be scheduled from the iOS Composer
- [ ] Telegram posts appear on the iOS Planner calendar with the Telegram logo
- [ ] Tapping a Telegram event shows: platform, account name, status, time, content preview
- [ ] Failed posts show the failure reason in the iOS post detail view
- [ ] No Telegram-specific options (silent message, link preview, pin, PDF) appear in the iOS Composer in v1

---

**Impact on existing data:** None — reads the same plan documents as the web app.

**Impact on other products:** No impact on web app, Android, or Chrome extension.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure**
- Depends on: **[BE] Implement Telegram post publishing (text, image, video, album, PDF)**

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (native iOS app)
- [ ] Multilingual support (all user-facing strings use the iOS localization system)
- [ ] UI theming support (follow existing iOS design system patterns)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 9: [Android] Add Telegram accounts to mobile Composer account selector and Planner calendar

**Group:** Product  
**Project:** Mobile  
**Priority:** Medium (P1)  
**Product Area:** Android Mobile  
**Skill Set:** Frontend  
**Story Type:** feature

---

**Description:**

As a mobile user on Android, I want to see my connected Telegram channels and groups in the Android Composer account selector and view Telegram posts on the Android Planner calendar so that I can manage my Telegram content schedule from my Android device.

Connection flow and Telegram-specific publishing options are out of scope for Android — these are web-only in v1. AI functionality is web-only and not applicable here.

---

**Workflow:**

1. User opens the ContentStudio Android app and taps **Create Post**.
2. In the account selector, a **Telegram** section appears with connected channels and groups. Each item shows the Telegram logo, avatar, and name.
3. Tapping selects/deselects the account. With Telegram selected, user writes content and attaches media. Telegram-specific options are not shown on Android in v1.
4. User schedules the post.
5. User navigates to **Planner**. Telegram posts appear on the Android calendar with the Telegram logo.
6. Tapping a Telegram event shows: platform, account name, status, time, content preview.
7. Failed posts show the failure reason.

---

**Acceptance criteria:**

- [ ] Telegram section appears in the Android Composer account selector with connected channels/groups
- [ ] Each account shows Telegram logo, avatar, and name
- [ ] Tapping an account selects/deselects it correctly
- [ ] Posts with Telegram selected can be scheduled from the Android Composer
- [ ] Telegram posts appear on the Android Planner calendar with the Telegram logo
- [ ] Tapping a Telegram event shows: platform, account name, status, time, content preview
- [ ] Failed posts show the failure reason in the Android post detail view
- [ ] No Telegram-specific options appear in the Android Composer in v1

---

**Impact on existing data:** None — reads the same plan documents as the web app.

**Impact on other products:** No impact on web app, iOS, or Chrome extension.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure**
- Depends on: **[BE] Implement Telegram post publishing (text, image, video, album, PDF)**

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (native Android app)
- [ ] Multilingual support (all strings use the Android localization system)
- [ ] UI theming support (follow existing Android design system patterns)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 15: [BE] Add Telegram to EasyConnect — agency client account connection via bot token

**Group:** Backend
**Project:** Web App
**Priority:** Medium (P1)
**Product Area:** Integrations
**Skill Set:** Backend
**Story Type:** feature

---

**Description:**

As an agency account manager, I want my clients to be able to connect their own Telegram channels and groups through my EasyConnect link — so I can onboard client Telegram accounts without ever needing to handle their bot tokens myself.

EasyConnect is ContentStudio's white-label feature where agencies generate a shareable link that clients visit to connect their social accounts. All other platforms use OAuth redirects; Telegram is different because it uses a bot token model. This story adds the backend plumbing to support Telegram in EasyConnect: the `ExternalLinkIntegration` model relationship, eager loading in the repo, and three EasyConnect-scoped API routes that run the same bot validation logic as the main connection flow but through the `EasyConnectMiddleware` (token-based auth, not session-based).

---

**Workflow:**

1. Agency creates an EasyConnect link in Settings → Integrations → EasyConnect (existing feature — no changes here)
2. Agency shares the link with a client who owns a Telegram channel or group
3. Client visits the EasyConnect URL and clicks "+ Connect Social Account" → selects Telegram
4. Client enters their bot token in the 3-step modal — frontend calls `POST /telegram/validate-bot/easy-connect`
5. Backend validates the token via the Telegram Bot API and returns bot info; response identical to `POST /telegram/validate-bot`
6. Client selects chats — frontend calls `POST /telegram/discover-chats/easy-connect`
7. Client optionally adds a chat manually — frontend calls `POST /telegram/validate-chat/easy-connect`
8. Client confirms — frontend calls `POST /telegram/add-chats/easy-connect`
9. Backend saves each chat as a `SocialIntegrations` document with `connection_via_link: true` and `connection_link_id` set to the EasyConnect link's `_id`
10. The agency's workspace now has the client's Telegram channels/groups connected and available for publishing

---

**Acceptance criteria:**

- [ ] `ExternalLinkIntegration` model has a `telegramAccounts()` relationship method that queries `SocialIntegrations` where `platform_type = 'telegram'` AND `connection_via_link = true` AND `connection_link_id = {link._id}`, selecting only the fields needed for display (`connection_link_id`, `connection_via_link`, `platform_name`, `platform_logo`, `validity`, `platform_identifier`)
- [ ] `ExternalLinkIntegrationRepo.php` eager-loads `telegramAccounts` alongside all existing platform accounts in `getExternalIntegrationLink()`, `getLinkSecureStatus()`, `verifyLinkPassword()`, and `getLinkDetails()` — result includes `telegram_accounts` array in the response payload
- [ ] `POST /telegram/validate-bot/easy-connect` route exists under `EasyConnectMiddleware`; delegates to `TelegramController@validateBot` with the workspace/user context injected by the middleware (not session); returns identical response shape to `POST /telegram/validate-bot`
- [ ] `POST /telegram/discover-chats/easy-connect` route exists under `EasyConnectMiddleware`; delegates to `TelegramController@discoverChats`; returns identical response shape to `POST /telegram/discover-chats`
- [ ] `POST /telegram/validate-chat/easy-connect` route exists under `EasyConnectMiddleware`; delegates to `TelegramController@validateChat`; returns identical response shape
- [ ] `POST /telegram/add-chats/easy-connect` route exists under `EasyConnectMiddleware`; delegates to `TelegramController@addChats` with `connection_via_link: true` and `connection_link_id` injected from the middleware's decoded token payload — saved accounts include both flags
- [ ] All five EasyConnect routes are grouped under `EasyConnectMiddleware` (not `auth` middleware); the middleware handles identity resolution from the `external-link-token` header
- [ ] If the EasyConnect token is expired or invalid, all five routes return 401 (standard `EasyConnectMiddleware` behavior)
- [ ] If the workspace account limit is reached, `POST /telegram/add-chats/easy-connect` returns the same limit-exceeded error as the main `add-chats` endpoint
- [ ] Existing `POST /telegram/validate-bot`, `POST /telegram/discover-chats`, `POST /telegram/validate-chat`, `POST /telegram/add-chats` routes (standard auth) are unaffected by these changes

---

**Mock-ups:** N/A — backend only

---

**Impact on existing data:**

- `social_integrations` documents created via EasyConnect will have `connection_via_link: true` and `connection_link_id` set — same pattern as all other EasyConnect-connected platforms. No changes to existing documents.
- `ExternalLinkIntegration` documents: no schema changes. The `telegram_accounts` array is assembled at read time via the new relationship.

---

**Impact on other products:** No impact on the standard Telegram connection flow, Composer, or Planner. EasyConnect change is additive.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure**
- Depends on: **[BE] Implement Telegram bot webhook and account connection API**

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (backend only)
- [ ] Multilingual support — N/A (no user-facing strings in this story)
- [ ] UI theming support — N/A (backend only)
- [ ] White-label domains impact review — EasyConnect is the primary white-label feature; ensure `connection_link_id` is correctly scoped per workspace and cannot leak Telegram accounts across workspaces
- [ ] Cross-product impact assessed — EasyConnect page is separate from Settings; no composer or planner impact

---

---

### Story 16: [FE] Add Telegram to EasyConnect client-facing account connection page

**Group:** Frontend
**Project:** Web App
**Priority:** Medium (P1)
**Product Area:** Integrations
**Skill Set:** Frontend
**Story Type:** feature

---

**Description:**

As an agency client, I want to connect my Telegram channel or group through the EasyConnect link my agency shared with me — so I can give my agency access to publish to my Telegram without handing over my credentials.

This story adds Telegram to the EasyConnect client-facing page. Unlike every other platform on EasyConnect (which redirect the client to an OAuth screen), Telegram uses a bot token. So clicking "Telegram" on the EasyConnect platform grid must open the existing 3-step bot token connection modal inline — not trigger an OAuth redirect. After connecting, the client's Telegram accounts appear on the EasyConnect page alongside other connected accounts.

---

**Workflow:**

1. Client opens the EasyConnect link their agency shared
2. Client sees the EasyConnect account management page — it shows any accounts already connected via this link
3. Client clicks **"+ Connect Social Account"**
4. The platform selector modal opens — Telegram appears in the platform grid
5. Client clicks **Telegram** — the **3-step bot token connection modal** opens (not an OAuth redirect):
   - **Step 1:** "Enter your bot token" — client pastes token, clicks "Validate" — frontend calls `POST /telegram/validate-bot/easy-connect`
   - **Step 2:** "Select your channels & groups" — discovered chats listed with checkboxes; manual add input; frontend calls `POST /telegram/discover-chats/easy-connect` and optionally `POST /telegram/validate-chat/easy-connect`
   - **Step 3:** "Confirm connection" — summary list; client clicks "Connect" — frontend calls `POST /telegram/add-chats/easy-connect`
6. On success: modal closes, EasyConnect page refreshes, newly connected Telegram accounts appear in the account list
7. Client sees their Telegram channels/groups listed with the Telegram icon, account name, and type badge (Channel / Group)

---

**Acceptance criteria:**

**`useSocialAccounts.js` composable:**
- [ ] `'telegram'` added to the `PLATFORMS` array
- [ ] `telegramAccounts: []` added to the `easyConnectAccounts` reactive state object
- [ ] `updateAccountData()` (or equivalent account refresh function) maps `response.data.telegram_accounts` to `easyConnectAccounts.telegramAccounts`

**`ExternalCloudConnect.vue` (v2 — primary EasyConnect page):**
- [ ] Telegram accounts from `easyConnectAccounts.telegramAccounts` are included in the `BulkReconnectedAccounts` (or equivalent account list) spread so connected Telegram accounts appear on the page
- [ ] The platform selector modal includes Telegram in the platform grid
- [ ] Clicking Telegram in the platform selector **opens the bot token connection modal** (reuses the existing `AddTelegram.vue` / Telegram connection modal component) — it does NOT trigger an OAuth redirect or call `getAuthorizationUrl`
- [ ] The modal is initialized in EasyConnect mode: all API calls inside the modal use the `/easy-connect` route variants (`POST /telegram/validate-bot/easy-connect`, etc.) with the `external-link-token` header
- [ ] On successful connection (`POST /telegram/add-chats/easy-connect` returns 200): modal closes, `updateAccountData()` is called to refresh the account list
- [ ] Account list items for Telegram show: Telegram icon, account name (channel/group name), type badge ("Channel" or "Group" using `Badge` component)

**`ExternalCloudConnectLink.vue` (v1 legacy EasyConnect page):**
- [ ] `AccountListing` entry for `type="telegram"` added with `external-link: true` and `:external-link-account="data.telegramAccounts"` props
- [ ] `updateExternalLinkData()` maps `response.data.telegram_accounts` to `data.telegramAccounts`
- [ ] Telegram accounts display correctly in the legacy account listing with name and type

**Copy:**
- [ ] Platform grid tile label: "Telegram"
- [ ] Platform grid tile subtitle (if shown): "Channels & Groups"
- [ ] Empty state (no Telegram accounts connected yet): no dedicated empty state needed — Telegram accounts simply don't appear in the list until connected

**Error states:**
- [ ] If bot token validation fails (Step 1 error): inline error "Invalid bot token. Please check and try again." — modal stays on Step 1
- [ ] If no chats discovered (Step 2 empty): empty state with instructions shown — same copy as the standard connection modal
- [ ] If account limit reached: `Alert` component (warning variant): "You've reached the account limit for your plan. Upgrade to connect more accounts." — with upgrade link

**General:**
- [ ] All new user-facing strings use `$t()` i18n keys; add to all locale files
- [ ] No hardcoded colors — use `text-primary-cs-500`, `bg-primary-cs-50` etc. for primary-colored elements

**UI component summary:**
- `Badge` — Channel / Group type indicator in the account list
- `Alert` (warning) — account limit error
- Reuses existing `AddTelegram.vue` modal component (no new modal needed — just wire it into the EasyConnect context)

---

**Mock-ups:**
See `docs/features/telegram-integration/02-workflow.md` section 6l for the full EasyConnect flow description.

---

**Impact on existing data:** None — reads `telegram_accounts` from the EasyConnect link details response. Additive.

---

**Impact on other products:** No impact on standard Telegram connection flow (Settings → Social Accounts), Composer, or Planner.

---

**Dependencies:**

- Depends on: **[BE] Add Telegram to EasyConnect — agency client account connection via bot token**
- Depends on: **[FE] Add Telegram account connection flow in Settings → Social Accounts** (reuses the bot token modal component)
- Depends on: **[FE] Add Telegram to onboarding, automations, settings, and product-wide platform lists** (Telegram tile in platform grids)

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — EasyConnect page is used on mobile browsers by clients; Telegram connection modal and account list must be fully usable at 375px width
- [ ] Multilingual support — All labels, error messages, and copy must use `$t()` i18n keys; add to all locale files
- [ ] UI theming support — Use CSS variable-backed Tailwind classes for all primary colors; no hardcoded colors
- [ ] White-label domains impact review — EasyConnect is the core white-label surface; Telegram brand color (blue) must not bleed into white-label theme
- [ ] Cross-product impact assessed — change is isolated to EasyConnect page; no impact on Composer, Planner, or Settings
