# Telegram Integration — Epic & Stories

**Feature:** Telegram Integration for ContentStudio  
**Date:** April 1, 2026  
**Pipeline Step:** 04 — Epic & Stories

---

## Epic

**Title:** Telegram Integration

**Description:**

ContentStudio currently supports publishing to Facebook, Instagram, LinkedIn, X (Twitter), TikTok, Pinterest, YouTube, Threads, and Bluesky — but not Telegram. With Telegram surpassing 1 billion monthly active users in 2025, this gap is a growing competitive disadvantage: of ContentStudio's 10 direct competitors, only Publer has full native Telegram integration, and it is one of the most-upvoted missing features among ContentStudio users.

This epic delivers v1 Telegram integration: users can connect their Telegram Channels and Groups to ContentStudio via a shared ContentStudio bot, schedule and publish text, image, video, and album posts from the Composer, see Telegram posts on the Planner calendar, manage Telegram accounts in Settings → Social Accounts, and use Telegram as a destination in RSS automation. A Telegram-style Composer preview and platform-specific options (silent mode, disable link preview, protect content) bring the publishing experience in line with ContentStudio's best-in-class platform integrations.

The implementation follows the same architecture as recently added platforms (Bluesky, Threads): a `TelegramConnector` for account management, a `TelegramPosting` strategy for publishing, and matching frontend components following the Bluesky reference pattern.

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

This story covers: adding Telegram to the platform config, creating the `TelegramAccounts` model (MongoDB, `social_integrations` collection), creating the `TelegramRepo` repository implementing `SocialPlatformsInterface`, and registering Telegram in the integration connector router (`Connector.php`). This is the data layer foundation that all other Telegram stories depend on.

---

**Workflow:**

1. Developer adds `telegram` to `config/social_platforms.php` → `platforms` array and adds `telegram.platform_identifier` to `account_selection_fields`.
2. Developer adds `telegram` config entry to `config/integrations.php` → `social_integrations` section: `bot_token` (from `TELEGRAM_BOT_TOKEN` env var), `api_url` (`https://api.telegram.org/bot`), `webhook_url` (from `TELEGRAM_WEBHOOK_URL` env var).
3. Developer creates `app/Models/Integrations/Platforms/Social/TelegramAccounts.php`:
   - Extends Eloquent, uses `social_integrations` MongoDB collection.
   - Global scope filters by `platform_type = 'telegram'`.
   - Fillable fields: `telegram_id`, `chat_id`, `bot_username`, `platform_identifier`, `platform_name`, `platform_type` (channel/group), `access_token`, `platform_logo`, `image`, `name`, `user_id`, `workspace_id`, `added_by`, `state`, `validity`, `validity_status`, `invalid_tries`, `QueueSlots`, `user_details`.
   - `access_token` stored encrypted (same pattern as `TwitterAccounts`).
4. Developer creates `app/Repository/Integrations/Platforms/Social/TelegramRepo.php`:
   - Implements `SocialPlatformsInterface`.
   - Methods: `getItems($filters)`, `getItem()`, `fetchQueueSlots($filters)`, `updateQueueSlots()`.
   - Uses `TelegramAccounts::where('platform_type', 'telegram')`.
5. Developer registers `TelegramConnector` in `app/Strategy/Integrations/Connector.php` → `case 'telegram': $this->connector = new TelegramConnector(); break;`
6. Developer adds `TELEGRAM_BOT_TOKEN` and `TELEGRAM_WEBHOOK_URL` to `.env.example`.
7. Developer adds `telegram_sharing_details` to `$fillable` in `app/Models/Publish/Planner/Plans.php`.

---

**Acceptance criteria:**

- [ ] `config/social_platforms.php` includes `'telegram'` in the `platforms` array
- [ ] `config/social_platforms.php` includes `'telegram.platform_identifier'` in `account_selection_fields`
- [ ] `config/integrations.php` includes a `telegram` key in `social_integrations` with `bot_token`, `api_url`, and `webhook_url`
- [ ] `TELEGRAM_BOT_TOKEN` and `TELEGRAM_WEBHOOK_URL` are present in `.env.example`
- [ ] `TelegramAccounts` model exists at `app/Models/Integrations/Platforms/Social/TelegramAccounts.php`
- [ ] `TelegramAccounts` global scope filters records by `platform_type = 'telegram'`
- [ ] `access_token` is encrypted on write and decrypted on read (same pattern as Twitter)
- [ ] `TelegramRepo` exists at `app/Repository/Integrations/Platforms/Social/TelegramRepo.php` and implements `SocialPlatformsInterface`
- [ ] `TelegramRepo::getItems()` returns Telegram accounts for a given workspace, paginated
- [ ] `TelegramRepo::getItem()` returns a single Telegram account by chat_id and workspace_id
- [ ] `Connector.php` routes `'telegram'` to `TelegramConnector` without breaking existing platform routing
- [ ] `Plans` model `$fillable` includes `telegram_sharing_details`
- [ ] Unit tests: `TelegramRepo::getItems()` returns only Telegram accounts (not accounts of other platforms)

---

**Mock-ups:** N/A — backend only

---

**Impact on existing data:**

- `social_integrations` MongoDB collection: New documents with `platform_type: 'telegram'` will be added. No changes to existing documents.
- `plans` MongoDB collection: `telegram_sharing_details` field added to `$fillable`; no schema migration required (MongoDB is schemaless). Existing plan documents are unaffected.

---

**Impact on other products:**

- No impact on existing platform integrations. The new `TelegramConnector` case in `Connector.php` is additive only.
- Chrome extension, mobile apps: No impact at this stage.

---

**Dependencies:** None — this is the foundational story.

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (backend only)
- [ ] Multilingual support — N/A (no user-facing strings in this story)
- [ ] UI theming support — N/A (backend only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 2: [BE] Implement Telegram account connection API

**Group:** Backend  
**Project:** Web App  
**Priority:** High (P0)  
**Product Area:** Integrations  
**Skill Set:** Backend  
**Story Type:** feature

---

**Description:**

As a user, I want to connect my Telegram channel or group to ContentStudio — whether it's public or private — so that I can schedule and publish posts to it alongside my other social accounts.

This story implements the server-side account connection flow supporting both public channels/groups and private groups. The UX is a single input field (same for both cases); the backend handles them differently:

- **Public channels/groups (@username):** Validated synchronously via `getChat` + `getChatMember`.
- **Private groups (invite link `https://t.me/+xxxx`):** Telegram fires a `my_chat_member` webhook update when the bot is added as admin. The backend stores `{ chat_id, invite_hash }` in Redis (30-minute TTL). When the user submits the invite link, the backend extracts the hash, looks up the `chat_id` in Redis, then confirms admin status via `getChatMember`.

The story also covers account disconnection and reconnection when the bot's admin status is revoked.

---

**Workflow:**

1. User adds @contentstudio_bot as admin to their Telegram channel/group in the Telegram app, then opens Settings → Social Accounts in ContentStudio and clicks "Connect" on the Telegram tile.
2. User enters their channel @username or invite link, confirms the toggle, and clicks "Add".
3. Frontend calls `POST /integrations/telegram/connect` with `{ channel_input: "@mychannel", workspace_id: "..." }`.
4. Backend detects input type:

   **Path A — Public (@username):**
   a. Strips leading `@` if present.
   b. Calls `getChat(@username)` to resolve channel/group metadata (name, type, `chat_id`, photo).
   c. If not found, returns HTTP 422 `{ error: "channel_not_found" }`.
   d. Calls `getChatMember` with the bot's own user ID to verify admin status.
   e. If not admin, returns HTTP 422 `{ error: "bot_not_admin" }`.

   **Path B — Private (invite link `https://t.me/+xxxx`):**
   a. Extracts the invite hash from the URL (the part after `https://t.me/+`).
   b. Looks up `telegram:chat:{invite_hash}` in Redis. If not found (bot not yet added, or TTL expired), returns HTTP 422 `{ error: "private_group_not_found", message: "We couldn't find this group. Make sure you've added @contentstudio_bot as an admin in Telegram first, then try again." }`.
   c. Retrieves `chat_id` from Redis.
   d. Calls `getChatMember` with the bot's own user ID to verify admin status.
   e. If not admin, returns HTTP 422 `{ error: "bot_not_admin" }`.

   **Both paths continue:**
   f. If already connected to this workspace, updates the existing record rather than creating a duplicate.
   g. If the channel is connected to a different workspace, returns HTTP 409 `{ error: "already_connected", message: "This channel is already connected to another ContentStudio workspace." }`.
   h. On success: stores the account in `social_integrations` with `platform_type: 'telegram'`, `chat_id`, `platform_name`, `platform_identifier` (username if public, `chat_id` string if private), `image` (photo URL), `chat_type` (channel/group), `access_token` (encrypted bot token), `state: 'Added'`, `validity: 'valid'`, `workspace_id`, `user_id`. Returns HTTP 200 with the account object.

5. **Webhook handler** (`POST /integrations/telegram/webhook`) — listens for `my_chat_member` updates:
   a. When the bot is added as admin to any chat, extract `chat_id` and `invite_link` (if present).
   b. If the chat has an invite link, extract the hash and store `telegram:chat:{invite_hash} → { chat_id, chat_title, chat_type, photo }` in Redis with a 30-minute TTL.
   c. This runs as a background handler — no user-facing response.

6. User can disconnect: `DELETE /integrations/telegram/disconnect/{account_id}` → sets `state: 'Removed'` and removes the account from the active list.
7. Reconnect flow: If an account has `validity: 'invalid'`, the frontend opens the same connection modal. On successful reconnection, `validity` is reset to `valid` and `invalid_tries` reset to 0.

---

**Acceptance criteria:**

- [ ] `POST /integrations/telegram/connect` accepts `channel_input` (username or invite link) and `workspace_id`
- [ ] If `channel_input` starts with `@` or is a plain username, the backend uses Path A (public — `getChat` + `getChatMember`)
- [ ] If `channel_input` starts with `https://t.me/+`, the backend uses Path B (private — Redis lookup + `getChatMember`)
- [ ] Path A: if `getChat` returns not-found, returns HTTP 422 with `error: "channel_not_found"`
- [ ] Path B: if Redis has no entry for the invite hash, returns HTTP 422 with `error: "private_group_not_found"` and user-readable message
- [ ] Both paths: if bot is not an admin, returns HTTP 422 with `error: "bot_not_admin"` and user-readable message
- [ ] Both paths: if channel already connected to a different workspace, returns HTTP 409 with `error: "already_connected"`
- [ ] Connecting the same channel to the same workspace a second time updates the existing record rather than creating a duplicate
- [ ] On success: account stored with correct `platform_type`, `chat_id`, `platform_name`, `validity: 'valid'`, `state: 'Added'`, encrypted `access_token`, `workspace_id`, `user_id`
- [ ] `POST /integrations/telegram/webhook` handles `my_chat_member` updates: when bot is added as admin to a chat that has an invite link, stores `{ chat_id, chat_title, chat_type, photo }` in Redis keyed by invite hash with 30-minute TTL
- [ ] Redis key expires after 30 minutes (TTL enforced)
- [ ] `DELETE /integrations/telegram/disconnect/{account_id}` sets account `state` to `'Removed'` and removes it from the active list
- [ ] Multiple Telegram accounts (different channels/groups) can be connected to the same workspace
- [ ] `access_token` (bot token) is encrypted at rest
- [ ] Unit tests: public connection success, private connection success, channel_not_found (public), private_group_not_found (private — no Redis entry), bot_not_admin, already_connected, duplicate handling

---

**Mock-ups:** N/A — backend only

---

**Impact on existing data:**

- `social_integrations` collection: new Telegram account documents created on successful connection. No changes to existing documents.
- Redis: `my_chat_member` events stored temporarily as `telegram:chat:{invite_hash}` with 30-minute TTL. Keys are short-lived and self-expiring.

---

**Impact on other products:**

- Chrome extension: No impact.
- Mobile apps: No impact at this stage — the connection flow is web-only; mobile apps will display connected accounts (covered in mobile stories).

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure**
- @contentstudio_bot must be created via @BotFather and `TELEGRAM_BOT_TOKEN` provisioned in the environment before development can be tested end-to-end.
- `TELEGRAM_WEBHOOK_URL` must be registered with Telegram's `setWebhook` API — required for the `my_chat_member` handler that enables private group support. This is a deployment step.

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (backend only)
- [ ] Multilingual support — N/A (backend only; error messages are English-only for v1)
- [ ] UI theming support — N/A (backend only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 3: [BE] Implement Telegram post publishing (text, image, video, album)

**Group:** Backend  
**Project:** Web App  
**Priority:** High (P0)  
**Product Area:** Integrations  
**Skill Set:** Backend  
**Story Type:** feature

---

**Description:**

As a user, I want to publish text, image, video, and album posts to my Telegram channel or group from ContentStudio so that my scheduled content is delivered to Telegram automatically without manual effort.

This story implements `TelegramPosting.php` (the posting strategy class), registers it in the `Posting.php` strategy router and the `SocialPosting.php` dispatcher, and handles all four content types: text-only (`sendMessage`), single image (`sendPhoto`), single video (`sendVideo`), and albums of 2–10 media items (`sendMediaGroup`). It handles all publishing options (silent mode, disable link preview, protect content, spoiler blur), error classification, validity invalidation on permission errors, and retry-on-rate-limit behavior. The `message_id` from every successful post is stored to enable the follow-up comment feature (see **[BE] Implement Telegram first comment publishing**).

---

**Workflow:**

1. A plan is scheduled for publishing. At the scheduled time, `PlanPostingJob` is dispatched.
2. `SocialPosting::processSocialPosting($plan)` detects `account_selection.telegram` is set and calls `(new Posting('Telegram', $plan))->initializePosting()->performPosting()`.
3. `Posting.php` routes to `new TelegramPosting($plan)`.
4. `TelegramPosting::performPosting()` executes:
   a. Fetches the Telegram account from `TelegramRepo::getItem()` using `chat_id` from `telegram_sharing_details`.
   b. Decrypts the bot token from `access_token`.
   c. Determines content type from the plan:
      - Text only → `sendMessage` with `text` parameter and Markdown formatting if applicable.
      - Single image → `sendPhoto` with `photo` (file or URL) and `caption`.
      - Single video → `sendVideo` with `video` (file or URL) and `caption`.
      - 2–10 media items → `sendMediaGroup` with array of `InputMedia` objects.
   d. Applies publishing options from `telegram_sharing_details`:
      - `silent_mode: true` → `disable_notification: true`
      - `disable_link_preview: true` → `disable_web_page_preview: true` (on `sendMessage`)
      - `protect_content: true` → `protect_content: true` (on all message types: `sendMessage`, `sendPhoto`, `sendVideo`, `sendMediaGroup`)
      - `spoiler: true` → `has_spoiler: true` (on `sendPhoto`, `sendVideo`, and each `InputMediaPhoto`/`InputMediaVideo` item in `sendMediaGroup` only — not applicable to text-only `sendMessage`)
   e. Sends the API request to `https://api.telegram.org/bot{token}/{method}`.
   f. On success (HTTP 200, `ok: true`): stores the `message_id` from the response in the posting log (required for first comment follow-up). Marks post as published.
   g. On `HTTP 429` (rate limit): reads `Retry-After` header, re-queues with that delay (up to 3 retries, then marks as failed).
   h. On `HTTP 400/403` (bot not admin, chat not found, token invalid): marks the account as `validity: 'invalid'`, `invalid_tries++`; dispatches account-invalid notification; marks post as failed with appropriate error message.
   i. On other errors: marks post as failed with the Telegram error description.
5. Posting response is merged into the overall `processSocialPosting` response.

---

**Acceptance criteria:**

- [ ] Text-only posts are published to Telegram via `sendMessage`; the full text appears in the Telegram channel/group
- [ ] Single image posts are published via `sendPhoto`; caption (up to 1,024 chars) appears below the image
- [ ] Single video posts are published via `sendVideo`; caption appears below the video
- [ ] Albums of 2–10 media items are published via `sendMediaGroup`; all items appear as a single album in Telegram
- [ ] `silent_mode: true` in `telegram_sharing_details` results in `disable_notification: true` in the API call
- [ ] `disable_link_preview: true` results in `disable_web_page_preview: true` in the `sendMessage` API call
- [ ] `protect_content: true` results in `protect_content: true` in all API calls (`sendMessage`, `sendPhoto`, `sendVideo`, `sendMediaGroup`)
- [ ] `spoiler: true` results in `has_spoiler: true` on `sendPhoto` and `sendVideo` API calls
- [ ] `spoiler: true` results in `has_spoiler: true` on each `InputMediaPhoto`/`InputMediaVideo` item in `sendMediaGroup`
- [ ] `spoiler: true` has no effect on `sendMessage` (text-only posts — `has_spoiler` is not a valid parameter there)
- [ ] Successful publication stores the `message_id` from Telegram's response in the posting log for use by the first comment follow-up
- [ ] HTTP 429 responses trigger a retry after the `Retry-After` delay; after 3 failed retries, the post is marked as failed
- [ ] HTTP 403 (bot removed as admin) marks the account as `validity: 'invalid'` and the post as failed with message: "The ContentStudio bot no longer has permission to post in this channel. Please reconnect your Telegram account."
- [ ] HTTP 400 (bad request / chat not found) marks the post as failed with the Telegram error description
- [ ] Invalid token marks the account as `validity: 'invalid'`
- [ ] Successful publication records the `message_id` from Telegram's response in the posting log
- [ ] Telegram posting does not affect publishing to other platforms in the same plan (failures are isolated)
- [ ] `SocialPosting.php` correctly conditionally calls `TelegramPosting` only when `account_selection.telegram` is set in the plan
- [ ] `Posting.php` correctly routes `'Telegram'` to `TelegramPosting`
- [ ] Unit tests: text posting, image posting, rate-limit retry, 403 handling

---

**Mock-ups:** N/A — backend only

---

**Impact on existing data:**

- `plans` collection: `telegram_sharing_details` is read during publishing. No structural changes to existing plan documents.
- `social_integrations` collection: `validity` and `invalid_tries` are updated on failed posts due to auth errors.
- Publish logs: Telegram posting results logged alongside other platforms.

---

**Impact on other products:**

- `SocialPosting.php` change is additive; existing platform publishing is unaffected.
- Chrome extension: No impact.
- Mobile apps: No impact — publishing is fully server-side.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure**
- Depends on: **[BE] Implement Telegram account connection API** (accounts must be connectable before publishing can be tested end-to-end)

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (backend only)
- [ ] Multilingual support — N/A (backend only; error messages are English-only for v1)
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

This story extends the existing RSS-to-post automation engine to include Telegram as a supported publishing destination. Users who set up an RSS automation can now select their connected Telegram channels/groups as destinations. No changes to the RSS parsing or feed management logic are needed — only the destination list and the publishing dispatch need to be extended.

---

**Workflow:**

1. User opens Automation → RSS Feeds and creates or edits an RSS automation.
2. In the "Post to" destination selector, user sees their connected Telegram channels/groups listed alongside other social accounts (enabled by the frontend story **[FE] Add Telegram account connection flow in Settings → Social Accounts**).
3. User selects a Telegram channel as a destination and saves the automation.
4. At the configured RSS check interval, the RSS automation job fetches new feed items.
5. For each new item, the job creates a plan document with `telegram_sharing_details` populated from the automation settings and `account_selection.telegram` set.
6. The plan is dispatched to the publishing queue, which calls `TelegramPosting::performPosting()` (already implemented in **[BE] Implement Telegram post publishing**).
7. The RSS item is published to the Telegram channel as a text post (title + URL, optionally with featured image if present in the feed item).

---

**Acceptance criteria:**

- [ ] RSS automation job includes Telegram accounts in the destination set when Telegram channels/groups are selected in the automation configuration
- [ ] A plan document is correctly created with `account_selection.telegram` and `telegram_sharing_details` populated from the automation settings
- [ ] The plan is dispatched to the publishing queue and `TelegramPosting::performPosting()` is called
- [ ] RSS items are published to the Telegram channel with the post title, description (truncated to fit Telegram's 4,096-char limit), and source URL
- [ ] If the RSS item has a featured image and the automation is configured to include images, the post is published as a photo post (`sendPhoto`) with the title+URL as caption
- [ ] Failed Telegram publications from RSS automation are logged and the failure reason is surfaced in the automation error log
- [ ] Publishing to other platforms in the same RSS automation is not affected by adding Telegram as a destination

---

**Mock-ups:** N/A — backend only

---

**Impact on existing data:**

- Plans created by RSS automation will now include `telegram_sharing_details` when Telegram is selected as a destination.
- No changes to existing RSS automation configurations.

---

**Impact on other products:**

- No impact on Chrome extension, iOS, or Android apps.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure**
- Depends on: **[BE] Implement Telegram post publishing (text, image, video, album)**

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (backend only)
- [ ] Multilingual support — N/A (backend only)
- [ ] UI theming support — N/A (backend only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 5: [FE] Add Telegram account connection flow in Settings → Social Accounts

**Group:** Frontend  
**Project:** Web App  
**Priority:** High (P0)  
**Product Area:** Integrations  
**Skill Set:** Frontend  
**Story Type:** feature

---

**Description:**

As a user, I want to connect my Telegram channel or group to ContentStudio from the Settings → Social Accounts page so that I can start scheduling posts to Telegram alongside my other social accounts.

This story implements the Telegram tile in the platform grid, the `AddTelegram.vue` connection modal (username/invite link input + confirmation toggle), success/error states, and the connected accounts list with disconnect and reconnect actions.

---

**Workflow:**

1. User navigates to **Settings → Social Accounts**.
2. User sees a **"Telegram"** tile in the platform grid with the Telegram logo. The tile shows "Connect" if no accounts are connected, or "X connected" (e.g., "2 connected") if accounts exist.
3. User clicks the Telegram tile.
4. The **"Connect Telegram"** modal opens. It contains:
   - A field labeled **"Telegram Group or Channel"** with an ℹ icon. Hovering the ℹ icon shows: "For public channels or groups, enter the @username (e.g. @mychannel). For private groups, paste the invite link from Telegram (e.g. https://t.me/+xxxx). Make sure you've added @contentstudio_bot as an admin first."
   - An input field with placeholder: *"@username or invite link"*
   - A toggle: **"I confirm to have added @contentstudio_bot as an admin"**
   - **"Cancel"** and **"Add"** buttons. The "Add" button is disabled until the input is non-empty and the toggle is on.
5. User adds @contentstudio_bot as admin to their Telegram channel/group (outside ContentStudio), returns to the modal, enters the username or invite link, turns on the confirmation toggle, and clicks **"Add"**.
6. Frontend calls `POST /integrations/telegram/connect` with the `channel_input`. While the request is in-flight, the "Add" button shows a loading spinner and is disabled.
7. On success:
   - The modal transitions to a success state: channel/group avatar + **"✓ [Channel Name] connected!"** headline + subtext: "You can now schedule posts to this channel from the Composer."
   - Two buttons: **"Connect Another"** (resets the modal to the input form) and **"Done"** (closes the modal).
8. Connected Telegram accounts appear in the connected accounts list. Each account shows: avatar, name, type badge ("Channel" or "Group"), and action buttons: **"Reconnect"** (if invalid) or **"Disconnect"** (if valid).
9. If an account has `validity: 'invalid'` (bot removed as admin), the account row shows a yellow warning badge and a **"Reconnect"** button that reopens the connection modal.
10. Clicking **"Disconnect"** shows a confirmation dialog before removing the account.

---

**Acceptance criteria:**

- [ ] Telegram tile appears in Settings → Social Accounts platform grid with the Telegram logo
- [ ] Telegram tile shows "Connect" when no accounts are connected, and "X connected" when accounts exist
- [ ] Clicking the tile opens the "Connect Telegram" modal
- [ ] Modal contains: a "Telegram Group or Channel" labeled input field with ℹ tooltip, placeholder "@username or invite link", a confirmation toggle, and Cancel/Add buttons
- [ ] "Add" button is disabled until the input is non-empty and the confirmation toggle is on
- [ ] While `POST /integrations/telegram/connect` is in-flight, the "Add" button shows a loading spinner and is disabled
- [ ] On API success, modal transitions to success state showing channel/group avatar, name, and "✓ [Channel Name] connected!" headline
- [ ] "Connect Another" button resets the modal back to the input form (clears input, resets toggle)
- [ ] "Done" button closes the modal; new account appears in the connected accounts list
- [ ] If API returns `error: "bot_not_admin"`, an inline error appears below the input: "It looks like @contentstudio_bot isn't an admin of this channel yet. Please add it as an administrator in Telegram and try again." The modal remains open.
- [ ] If API returns `error: "channel_not_found"`, inline error: "We couldn't find this channel or group. Please check the @username and try again."
- [ ] If API returns `error: "private_group_not_found"`, inline error: "We couldn't find this group. Make sure you've added @contentstudio_bot as an admin in Telegram first, then try again."
- [ ] If API returns `error: "already_connected"`, inline error: "This channel is already connected to another ContentStudio workspace."
- [ ] Connected Telegram accounts list shows avatar, name, type badge ("Channel" or "Group") for each account
- [ ] Accounts with `validity: 'invalid'` show a yellow "Needs reconnection" badge and a "Reconnect" button
- [ ] "Reconnect" opens the connection modal (same form, empty input)
- [ ] "Disconnect" shows a confirmation dialog with the channel name and a warning about scheduled posts
- [ ] Confirming disconnect removes the account from the list and shows a success toast: "[Channel Name] has been disconnected."
- [ ] All user-facing strings use `$t()` with new keys in `src/locales/*/integration.json`

---

**UI Copy & Components:**

**Modal — "Connect Telegram":**
- **Title:** "Connect Telegram"
- **Field label:** "Telegram Group or Channel" (with ℹ icon)
- **ℹ tooltip:** "For public channels or groups, enter the @username (e.g. @mychannel). For private groups, paste the invite link from Telegram (e.g. https://t.me/+xxxx). Make sure you've added @contentstudio_bot as an admin first."
- **Input placeholder:** "@username or invite link"
- **Toggle label:** "I confirm to have added @contentstudio_bot as an admin"
- **Cancel button:** "Cancel" (`Button` secondary variant)
- **Add button:** "Add" (`Button` primary variant, disabled until input + toggle filled; shows `Loader` small when loading)

**Modal — Success state:**
- **Title:** "✓ Channel Connected!" (or "✓ Group Connected!" depending on type)
- **Subtext:** "[Channel Name] is now connected to ContentStudio. You can select it in the Composer when creating posts."
- Primary button: "Connect Another" (`Button` primary variant)
- Secondary button: "Done" (`Button` secondary variant)

**Inline error (below input field, use `CstAlert` or a red helper text):**
- `bot_not_admin`: "It looks like @contentstudio_bot isn't an admin of this channel yet. Please add it as an administrator in Telegram and try again."
- `channel_not_found`: "We couldn't find this channel or group. Please check the @username and try again."
- `private_group_not_found`: "We couldn't find this group. Make sure you've added @contentstudio_bot as an admin in Telegram first, then try again."
- `already_connected`: "This channel is already connected to another ContentStudio workspace."
- Generic API error: "Something went wrong. Please try again."

**Disconnect confirmation dialog (use `Dialog` component):**
- **Title:** "Disconnect [Channel Name]?"
- **Body:** "This will remove [Channel Name] from ContentStudio. Any posts scheduled to this channel that haven't been published yet will be cancelled."
- Confirm button: "Yes, Disconnect" (`Button` danger/destructive variant)
- Cancel button: "Cancel" (`Button` secondary variant)

**Account row badges (use `Badge` component):**
- Channel type badge: "Channel" (gray/neutral)
- Group type badge: "Group" (gray/neutral)
- Invalid status badge: "Needs reconnection" (yellow/warning)

**Toast notifications (use `CstToast`):**
- Success: "[Channel Name] has been disconnected."
- Error on disconnect failure: "Failed to disconnect. Please try again."

**Components used:** `Modal`, `Dialog`, `Button`, `Loader`, `Badge`, `CstToast`, `CstPopup` (for ℹ tooltip), `Switch` (for confirmation toggle), `ActionIcon` (for close button)

**Component gap — None.** All required components are available in `@contentstudio/ui` or legacy `Cst*`.

---

**Empty / Error / Loading states:**

- **Empty state (Telegram tile, no accounts):** Tile shows "Connect" CTA button.
- **Loading (Add in-flight):** "Add" button shows `Loader` (small, inline) and is disabled.
- **Inline error (API error):** Red helper text below the input field with the relevant error message (see UI Copy above). Modal stays open; input and toggle are preserved so user can fix and retry.
- **Generic network error:** "Something went wrong. Please try again." inline below the input.

---

**Mock-ups:** See PRD section 7 (workflow description). Figma designs to be provided by Design team.

---

**Impact on existing data:** None — purely additive frontend changes. No existing integration UI is modified.

---

**Impact on other products:** None on Chrome extension or mobile apps.

---

**Dependencies:**

- Depends on: **[BE] Implement Telegram account connection API** (the `/connect` endpoint must exist)
- Requires Telegram SVG icon asset (`telegram-icon.svg`, `telegram-rounded.svg`) in `src/assets/img/integration/` — to be provided by Design or sourced from Telegram's brand assets.

---

**Global quality & compliance:**

- [ ] Mobile responsiveness (the Integrations settings page is desktop-first but must not break on tablet viewports)
- [ ] Multilingual support (all user-facing strings must use `$t()` with new keys in `src/locales/*/integration.json`)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 6: [FE] Add Telegram to Composer — account selector, preview, and options panel

**Group:** Frontend  
**Project:** Web App  
**Priority:** High (P0)  
**Product Area:** Composer  
**Skill Set:** Frontend  
**Story Type:** feature

---

**Description:**

As a user, I want to select my Telegram channels and groups in the Composer, see a Telegram-style preview of my post, set Telegram-specific publishing options (silent mode, disable link preview, protect content, spoiler blur), and schedule a first comment that auto-posts as a reply immediately after publishing — so that I can create, customize, and schedule Telegram posts with confidence from within my existing content workflow.

This story implements: (1) Telegram accounts in the existing multi-platform account selector; (2) `TelegramPreview.vue` in the right preview panel; (3) `EditorTelegramBox.vue` options panel with silent mode, disable link preview, protect content, and spoiler toggles; (4) a First Comment section (following the existing LinkedIn first comment pattern) where users can write a follow-up reply to be posted immediately after the main post.

---

**Workflow:**

1. User opens the **Composer** (Create New Post / Edit Draft).
2. In the **account selector**, user sees a "Telegram" section with their connected Telegram channels and groups listed. Each account shows the Telegram logo, channel/group avatar, and name. The `CstAccountCheckBox` component is used for selection (consistent with other platforms).
3. User selects one or more Telegram accounts. The Composer marks Telegram as an active platform.
4. The **right preview panel** renders a **Telegram-style preview** (`TelegramPreview.vue`):
   - A message bubble styled in Telegram's visual language: white bubble on a light background, channel avatar (circle) + channel name in bold at the top left, message content below, media thumbnail (if any) above the text, timestamp (showing "Just now") at the bottom right.
   - A **character counter** below the preview content: shows remaining characters out of 4,096 (text-only) or 1,024 (media caption). Counter turns yellow at 80% and red at 100% (use `text-yellow-500` and `text-red-500` — these are semantic, not primary-color, and acceptable).
   - If a URL is in the text and "Disable link preview" is off, the preview shows a link preview card below the message (a simplified mockup card with the URL domain).
5. A **"Telegram Options"** collapsible panel (`EditorTelegramBox.vue`) appears below the main text area (or alongside the platform-specific sections) when at least one Telegram account is selected. The panel is open by default on first selection.
   - **"Silent post"** toggle (`Switch` component from `@contentstudio/ui`):
     - Label: "Silent post"
     - Subtext: "Send without push notification"
     - Tooltip (shown on `ℹ` icon hover via `CstPopup`): "Your subscribers will receive this post but won't get a push notification. Useful for non-urgent updates like late-night posts or routine content — so you don't interrupt your audience at odd hours."
     - Default: off
   - **"Disable link preview"** toggle (`Switch` component):
     - Label: "Disable link preview"
     - Subtext: "Don't show the link preview card"
     - Tooltip: "Telegram automatically shows a preview card (with image and title) when your post contains a URL. Turn this off if the link context is already clear from your text, or if you want a cleaner-looking post without the preview box."
     - This toggle appears only when the Composer detects a URL in the Telegram post text (auto-detect).
     - Default: off
   - **"Protect content"** toggle (`Switch` component):
     - Label: "Prevent forwarding and saving"
     - Subtext: "Restrict recipients from forwarding or saving this post"
     - Tooltip (shown on `ℹ` icon hover via `CstPopup`): "When enabled, recipients cannot forward this post to other chats or save media from it. Useful for exclusive or paid community content you want to keep within your channel."
     - Default: off
   - **"Spoiler"** toggle (`Switch` component):
     - Label: "Spoiler"
     - Subtext: "Blur media until viewed"
     - Tooltip: "Adds a blur effect to your photo or video. Recipients see a blurred preview with a SPOILER label and tap to reveal the full content. Useful for reveals, surprises, or sensitive media."
     - Visible only when an image or video is attached to the post. Hidden for text-only posts.
     - Default: off
6. Below the Telegram Options panel, a **"First Comment"** collapsible section is available (same UX pattern as LinkedIn first comment in the existing Composer). When expanded:
   - A text area labeled **"First Comment"** with placeholder: "Write your first comment…"
   - A character counter showing remaining characters out of 4,096.
   - The text is stored as `telegram_sharing_details.first_comment.text` with `telegram_sharing_details.first_comment.enabled: true`.
   - When collapsed/cleared, `first_comment.enabled: false`.
7. When the user edits the post text for Telegram (using per-platform customization if enabled), the preview updates in real time.
8. The `telegram_sharing_details` object saved with the plan includes: `silent_mode: boolean`, `disable_link_preview: boolean`, `protect_content: boolean`, `spoiler: boolean`, `chat_ids: string[]`, `first_comment: { enabled: boolean, text: string }`.
8. If a selected Telegram account has `validity: 'invalid'`, the account selector shows a yellow warning icon next to the account name, and hovering shows: "This Telegram account needs reconnecting. Go to Settings → Social Accounts to fix it."
9. When the user deselects all Telegram accounts, the Telegram Options panel collapses and the TelegramPreview is hidden.

---

**Acceptance criteria:**

- [ ] Connected Telegram channels and groups appear in the Composer account selector under a "Telegram" section
- [ ] Selecting a Telegram account activates the TelegramPreview in the right panel
- [ ] TelegramPreview renders: channel/group avatar, channel/group name, message text, media thumbnail (if applicable)
- [ ] Character counter shows remaining characters; turns yellow at 80% of limit, red at 100%+
- [ ] Character limit for text-only posts is 4,096; for media posts (caption) is 1,024
- [ ] Link preview mock card appears in TelegramPreview when a URL is present and "Disable link preview" is off
- [ ] "Telegram Options" panel is visible and open by default when at least one Telegram account is selected
- [ ] "Silent post" toggle is present with the correct label, subtext, and tooltip; defaults to off
- [ ] "Disable link preview" toggle appears only when a URL is detected in the post text
- [ ] "Disable link preview" toggle has the correct label, subtext, and tooltip; defaults to off
- [ ] "Protect content" toggle is present with the correct label, subtext, and tooltip; defaults to off
- [ ] "Spoiler" toggle appears in the Telegram Options panel only when an image or video is attached; hidden for text-only posts
- [ ] "Spoiler" toggle has the correct label ("Spoiler"), subtext ("Blur media until viewed"), and tooltip; defaults to off
- [ ] When Spoiler is on, `spoiler: true` is saved in `telegram_sharing_details`
- [ ] "First Comment" collapsible section is present below the Telegram Options panel, following the existing LinkedIn first comment UX pattern
- [ ] "First Comment" text area has placeholder "Write your first comment…" and a 4,096-character counter
- [ ] When a first comment is entered and enabled, `first_comment: { enabled: true, text: "..." }` is saved in `telegram_sharing_details`
- [ ] When First Comment section is collapsed or text is cleared, `first_comment.enabled: false`
- [ ] Toggle states (`silent_mode`, `disable_link_preview`, `protect_content`, `spoiler`, `first_comment`) are saved in `telegram_sharing_details` when the post is scheduled/saved
- [ ] "Telegram Options" panel collapses when all Telegram accounts are deselected
- [ ] Invalid Telegram accounts (validity: invalid) show a yellow warning icon in the account selector
- [ ] Hovering the warning icon shows: "This Telegram account needs reconnecting. Go to Settings → Social Accounts to fix it."
- [ ] Multiple Telegram accounts can be selected simultaneously; TelegramPreview shows the first selected account's name/avatar (or a combined indicator for multiple)
- [ ] All user-facing strings in this story are wrapped in `$t()` with keys in `src/locales/*/composer.json`

---

**UI Copy & Components:**

**"Telegram Options" panel header:**
- Label: "Telegram Options"
- Use `Collapsible` component (`@contentstudio/ui`)

**"Silent post" toggle:**
- Label: "Silent post"
- Subtext: "Send without push notification"
- Tooltip (`ℹ` icon via `CstPopup`): "Your subscribers will receive this post but won't get a push notification. Useful for non-urgent updates like late-night posts or routine content — so you don't interrupt your audience at odd hours."
- Component: `Switch` (`@contentstudio/ui`)

**"Disable link preview" toggle:**
- Label: "Disable link preview"
- Subtext: "Don't show the link preview card below your post"
- Tooltip (`ℹ` icon via `CstPopup`): "Telegram automatically shows a preview box (with image and title) when your post contains a URL. Turn this off if the link speaks for itself or if you want a cleaner post. Example: A post that says 'Check our blog: example.com/post' doesn't need a preview box."
- Component: `Switch` (`@contentstudio/ui`)

**"Prevent forwarding and saving" toggle:**
- Label: "Prevent forwarding and saving"
- Subtext: "Restrict recipients from forwarding or saving this post"
- Tooltip (`ℹ` icon via `CstPopup`): "When enabled, recipients cannot forward this post to other chats or save media from it. Useful for exclusive or paid community content you want to keep within your channel."
- Component: `Switch` (`@contentstudio/ui`)
- Default: off

**Character counter:**
- Below the text area (same location as the existing character counter for other platforms)
- Shows: `[remaining] / [limit] characters` (e.g., "3,891 / 4,096 characters")
- At 80%+ used: `text-yellow-500` (semantic warning color)
- At 100% (over limit): `text-red-500` (semantic error color)

**Warning icon on invalid account:**
- Use `Icon` component (`@contentstudio/ui`) with warning/alert icon, `text-yellow-500`
- Tooltip via `CstPopup`: "This Telegram account needs reconnecting. Go to Settings → Social Accounts to fix it."

**Component gap — Tooltip:** No standalone `Tooltip` component exists in `@contentstudio/ui`. Use `CstPopup` for all tooltip hover content. This is the existing pattern across the Composer.

---

**Empty / Error / Loading states:**

- **No Telegram accounts connected:** The Telegram section in the account selector shows: "No Telegram accounts connected. [Connect one →]" (link to Settings → Social Accounts). This matches the existing pattern for unconnected platforms.
- **Preview loading:** `Loader` (small) shown briefly in the preview panel while the account's avatar is fetching. If avatar fails to load, show the Telegram logo as fallback.

---

**Mock-ups:** See PRD section 7 and workflow design. Figma designs to be provided by Design team.

---

**Impact on existing data:** None — the Composer reads `telegram_sharing_details` from the plan document. This field was added to `Plans.$fillable` by **[BE] Create Telegram account integration infrastructure**.

---

**Impact on other products:** None directly. The account selector and editor box are loaded conditionally — no changes affect other platform sections.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure** (Vuex store needs Telegram accounts data structure)
- Depends on: **[FE] Add Telegram account connection flow in Settings → Social Accounts** (accounts must be connectable before they can appear in the Composer selector — logically required for QA)
- Requires `telegram.js` Vuex store state module at `src/modules/integration/store/states/platforms/social/telegram.js` (should be created as part of this story, following the `bluesky.js` pattern)

---

**Global quality & compliance:**

- [ ] Mobile responsiveness (Composer is desktop-first; Telegram section should not break on tablet viewport)
- [ ] Multilingual support (all strings via `$t()` in `src/locales/*/composer.json`)
- [ ] UI theming support (default + white-label, design library components are being used — `Switch`, `Collapsible`, `Icon` from `@contentstudio/ui`)
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

This story adds Telegram post tiles to the Planner calendar and list view, including the Telegram platform icon, correct status display (Scheduled / Published / Failed), a failed post reason display, and Telegram in the platform filter list.

---

**Workflow:**

1. User navigates to **Planner** (Calendar or List view).
2. Scheduled Telegram posts appear as tiles on the calendar at their scheduled date/time. Each tile shows: Telegram logo, channel/group name (truncated if long), post text preview (first ~60 characters), and status indicator (colored dot: blue = Scheduled, green = Published, red = Failed).
3. User can filter the calendar by platform. **Telegram** appears in the platform filter list with the Telegram logo.
4. Clicking a Telegram post tile opens the **post detail panel** on the right:
   - Platform: Telegram logo + "Telegram"
   - Account: channel/group avatar + name + type badge ("Channel" or "Group")
   - Status: "Scheduled" / "Published" / "Failed"
   - Scheduled time (or published time)
   - Post content preview (full text, media thumbnail if applicable)
   - **If status is "Failed":** The failure reason is shown in a red `Alert` component (e.g., "The ContentStudio bot no longer has permission to post in this channel. Please reconnect your Telegram account in Settings → Social Accounts.")
   - **If status is "Failed":** A **"Retry"** button (`Button` primary variant) is shown to re-queue the post.
   - Standard action buttons: Edit, Duplicate, Delete (same as any other platform post).
5. If the user has filtered by "Telegram" only, the calendar shows only Telegram posts.

---

**Acceptance criteria:**

- [ ] Telegram posts appear on the Planner calendar at their scheduled date/time with the Telegram logo
- [ ] Calendar tiles show: Telegram logo, channel/group name (truncated to ~30 chars), status indicator dot (blue = Scheduled, green = Published, red = Failed)
- [ ] "Telegram" appears in the platform filter list with the Telegram logo; filtering by Telegram shows only Telegram posts
- [ ] Clicking a Telegram tile opens the post detail panel
- [ ] Post detail panel shows: platform icon, account name + type badge, status, scheduled/published time, post content preview
- [ ] Failed posts show the failure reason in a red `Alert` component inside the detail panel
- [ ] Failed posts show a "Retry" button that re-queues the post
- [ ] "Retry" shows a loading state (button disabled + `Loader` inside button) while the API request is in flight
- [ ] After successful retry queue, button shows "✓ Queued for retry" and then returns to normal state after 3 seconds
- [ ] Telegram posts in List view also show the Telegram logo and correct status
- [ ] All user-facing strings are wrapped in `$t()` with keys in `src/locales/*/planner.json`

---

**UI Copy & Components:**

**Failed post `Alert` (use `Alert` component from `@contentstudio/ui`, variant: danger/error):**
- "The ContentStudio bot no longer has permission to post in this channel or group. Please go to Settings → Social Accounts and reconnect your Telegram account."
- "Your Telegram account was disconnected. Please reconnect it in Settings → Social Accounts."
- "This post could not be delivered due to a temporary Telegram error. Please retry below."

**"Retry" button (use `Button`, primary variant):**
- Default label: "Retry Post"
- Loading state: `Loader` (x-small) inside button, disabled
- Success state: "✓ Queued for retry" (3 seconds, then reverts)

**Type badge (use `Badge` component):**
- "Channel" (neutral/gray)
- "Group" (neutral/gray)

---

**Empty / Error / Loading states:**

- **Empty state (Telegram filter active, no posts):** Calendar shows: "No Telegram posts scheduled" with subtext "Select Telegram accounts in the Composer to start scheduling posts to Telegram." No dedicated illustration needed — use the existing empty calendar state.
- **Loading state:** Existing Planner skeleton/loading pattern (no change needed for Telegram-specific loading).

---

**Mock-ups:** Follows existing Planner calendar tile and detail panel patterns. Figma designs to be provided by Design team.

---

**Impact on existing data:** None — reads existing plan documents. Telegram plans stored the same as any other platform plan.

---

**Impact on other products:** None — additive changes to the Planner.

---

**Dependencies:**

- Depends on: **[BE] Implement Telegram post publishing (text, image, video, album)** (posts must be publishable before the full Published/Failed status flow can be tested)
- Depends on: **[FE] Add Telegram to Composer — account selector, preview, and options panel** (posts must be schedulable before they appear in the Planner)

---

**Global quality & compliance:**

- [ ] Mobile responsiveness (Planner is desktop-first but Telegram tiles should render correctly on tablet)
- [ ] Multilingual support (all strings via `$t()` in `src/locales/*/planner.json`)
- [ ] UI theming support (default + white-label, `Alert`, `Button`, `Badge`, `Loader` from `@contentstudio/ui`)
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

This story covers the iOS-specific changes needed to surface Telegram integration in the existing mobile Composer and Planner screens. Telegram-specific publishing options (silent mode, disable link preview) and the connection flow are out of scope for iOS — these are web-only in v1. AI functionality is web-only and not applicable here.

---

**Workflow:**

1. User opens the ContentStudio iOS app.
2. User taps **Create Post** to open the mobile Composer.
3. In the account selector, **Telegram** section appears with the user's connected Telegram channels and groups listed. Each account shows the Telegram logo, account avatar, and name. Tapping an account selects/deselects it.
4. With a Telegram account selected, the user can write their post text and attach media (image or video) as they would for any other platform. Platform-specific Telegram options (silent mode, etc.) are not shown on mobile in v1.
5. User schedules the post (sets date/time) and confirms. The post is queued for publishing.
6. User navigates to the **Planner** tab. Telegram posts appear on the mobile calendar view with the Telegram logo icon on the event tile.
7. Tapping a Telegram event tile shows the post detail: platform (Telegram), account name, status (Scheduled / Published / Failed), scheduled time, and post content preview.
8. If the post has **Failed** status, the detail view shows the failure reason (e.g., "Bot no longer has permission").

---

**Acceptance criteria:**

- [ ] Telegram section appears in the iOS mobile Composer account selector with connected channels/groups
- [ ] Each Telegram account shows the Telegram logo, account avatar, and name
- [ ] Tapping an account correctly selects/deselects it for the post
- [ ] Posts with Telegram selected can be scheduled successfully from the iOS Composer
- [ ] Telegram posts appear on the iOS Planner calendar with the Telegram logo
- [ ] Tapping a Telegram calendar event shows: platform, account name, status, scheduled time, content preview
- [ ] Failed Telegram posts show the failure reason in the iOS post detail view
- [ ] No Telegram-specific options (silent mode, link preview) appear in the iOS Composer in v1 — these are web-only

---

**Mock-ups:** Follow existing iOS Composer account selector and Planner calendar patterns. See PRD section 7 for full feature context.

---

**Impact on existing data:** None — reads the same plan documents as the web app.

---

**Impact on other products:** No impact on web app, Android, or Chrome extension.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure** (Telegram accounts must exist in the API)
- Depends on: **[BE] Implement Telegram post publishing (text, image, video, album)** (publishing must be functional before end-to-end testing on iOS)
- The iOS app must support Telegram's `platform_type` in the social accounts API response — confirm with backend that the API returns Telegram accounts in the same format as other platforms.

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (this is the native iOS app)
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

This story covers the Android-specific changes needed to surface Telegram integration in the existing mobile Composer and Planner screens. Telegram-specific publishing options (silent mode, disable link preview) and the connection flow are out of scope for Android — these are web-only in v1. AI functionality is web-only and not applicable here.

---

**Workflow:**

1. User opens the ContentStudio Android app.
2. User taps **Create Post** to open the mobile Composer.
3. In the account selector, **Telegram** section appears with the user's connected Telegram channels and groups. Each account shows the Telegram logo, account avatar, and name. Tapping selects/deselects it.
4. With a Telegram account selected, the user writes their post and optionally attaches media. Platform-specific Telegram options are not shown on Android in v1.
5. User schedules the post and confirms.
6. User navigates to the **Planner**. Telegram posts appear on the Android calendar view with the Telegram logo.
7. Tapping a Telegram event shows the post detail: platform, account name, status, scheduled time, content preview.
8. Failed posts show the failure reason.

---

**Acceptance criteria:**

- [ ] Telegram section appears in the Android mobile Composer account selector with connected channels/groups
- [ ] Each Telegram account shows the Telegram logo, account avatar, and name
- [ ] Tapping an account correctly selects/deselects it for the post
- [ ] Posts with Telegram selected can be scheduled successfully from the Android Composer
- [ ] Telegram posts appear on the Android Planner calendar with the Telegram logo
- [ ] Tapping a Telegram calendar event shows: platform, account name, status, scheduled time, content preview
- [ ] Failed Telegram posts show the failure reason in the Android post detail view
- [ ] No Telegram-specific options (silent mode, link preview) appear in the Android Composer in v1 — these are web-only

---

**Mock-ups:** Follow existing Android Composer account selector and Planner calendar patterns.

---

**Impact on existing data:** None — reads the same plan documents as the web app.

---

**Impact on other products:** No impact on web app, iOS, or Chrome extension.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure**
- Depends on: **[BE] Implement Telegram post publishing (text, image, video, album)**
- The Android app must support Telegram's `platform_type` in the social accounts API response — confirm with backend.

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (native Android app)
- [ ] Multilingual support (all strings use the Android localization system)
- [ ] UI theming support (follow existing Android design system patterns)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 10: [BE] Register Telegram in platform registries, permissions, automation configs, and account validity job

**Group:** Backend
**Project:** Web App
**Priority:** High (P0)
**Product Area:** Throughout Product
**Skill Set:** Backend
**Story Type:** feature

---

**Description:**

Several backend platform registries, permission guards, validation rules, automation configs, and background jobs need to know Telegram is a supported platform before the full integration can ship. Without these updates: Telegram accounts won't appear in search results, workspace permission toggles won't include Telegram, CSV/Evergreen automations won't offer Telegram as a destination, and the nightly account-validity job won't detect when the bot is removed as admin.

This story is entirely additive — every change is adding `'telegram'` to an existing list or switch statement, following the same pattern used when Bluesky and Threads were added.

---

**Workflow:**

1. Developer adds `'telegram'` to `app/Libraries/Integrations/Integrations.php`:
   - `getSocialIntegrationsNames()` return array
   - `getIdentifierKey()` switch: `case 'telegram': return 'platform_identifier';`
2. Developer adds `'telegram'` to `app/Models/Scopes.php`:
   - `$social_platforms` array
   - `$social_platforms_selection` with field mapping: `'telegram' => ['telegram_id', 'chat_id', 'platform_identifier', 'platform_name', 'image', 'chat_type']`
3. Developer adds `'telegram'` to `app/Libraries/Permission/PermissionHelper.php` → `$social_platforms` array.
4. Developer adds `'telegram'` to the `in:` validation rule in:
   - `app/Http/Requests/Settings/Team/SocialAccountAccessRequest.php`
   - `app/Http/Requests/Integrations/SearchSocialAccountRequest.php`
5. Developer adds `'telegram' => []` to `$defaultAccountSelection` (and `default_account_selection`) in:
   - `config/csvAutomation.php`
   - `config/evergreenAutomation.php`
6. Developer adds a Telegram validity check to `app/Jobs/Integrations/Validate/SocialAccountsJob.php`:
   - For each connected Telegram account in the workspace, call Telegram Bot API `getChatMember` (bot user ID + stored `chat_id`) to verify the bot is still an admin.
   - If the bot has been removed as admin: set `validity: 'invalid'`, increment `invalid_tries`, dispatch the workspace account-invalid notification (same pattern as other platforms).
   - If the bot token itself is invalid (403 from Telegram): same invalidity handling.

---

**Acceptance criteria:**

- [ ] `Integrations::getSocialIntegrationsNames()` includes `'telegram'`
- [ ] `Integrations::getIdentifierKey('telegram')` returns `'platform_identifier'`
- [ ] `Scopes::$social_platforms` includes `'telegram'`
- [ ] `Scopes::$social_platforms_selection` includes `'telegram'` with the correct field array
- [ ] `PermissionHelper::$social_platforms` includes `'telegram'`
- [ ] `SocialAccountAccessRequest` allows `'telegram'` as a valid platform value without validation error
- [ ] `SearchSocialAccountRequest` allows `'telegram'` as a valid platform value without validation error
- [ ] `config/csvAutomation.php` `$defaultAccountSelection` includes a `'telegram'` key (empty array default)
- [ ] `config/evergreenAutomation.php` `$defaultAccountSelection` includes a `'telegram'` key (empty array default)
- [ ] `SocialAccountsJob` runs a Telegram validity check: calls `getChatMember` for each connected Telegram account
- [ ] If bot is no longer admin, account is set to `validity: 'invalid'`, `invalid_tries` is incremented, and a workspace notification is dispatched
- [ ] If bot token is invalid (Telegram returns 403), same invalidity flow applies
- [ ] No existing platform's validation or job logic is broken by these additive changes

---

**Mock-ups:** N/A — backend only

---

**Impact on existing data:** None — all changes are additive to existing arrays and switch statements. No existing records are modified.

---

**Impact on other products:**

- CSV and Evergreen automations will now render a Telegram account selection option (requires the matching FE story **[FE] Add Telegram to onboarding, automations, settings, and product-wide platform lists**).
- The validity job change means Telegram bot removal will be caught automatically on the next job run and surfaced to users via existing notification channels.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure** (TelegramAccounts model and TelegramRepo must exist for the validity job to query them)
- `@contentstudio_bot` must be provisioned (via @BotFather) and `TELEGRAM_BOT_TOKEN` set in the environment before the validity job can be tested end-to-end

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (backend only)
- [ ] Multilingual support — N/A (no user-facing strings in this story)
- [ ] UI theming support — N/A (backend only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 11: [FE] Add Telegram to onboarding, automations, settings, and product-wide platform lists

**Group:** Frontend
**Project:** Web App
**Priority:** High (P0)
**Product Area:** Throughout Product
**Skill Set:** Frontend
**Story Type:** feature

---

**Description:**

As Telegram integration ships, it needs to appear everywhere other social platforms appear in the product. This story registers Telegram in all frontend platform registries and composables, adds icon assets, and surfaces Telegram in five specific product areas that have explicit platform lists not covered by other stories: the onboarding wizard, CSV and Evergreen automations, workspace team permissions, the Manage Limits settings tab, and the billing plan comparison.

The reference pattern is Bluesky — the most recently added platform. Every file touched in this story follows the same additive pattern used for Bluesky.

---

**Workflow:**

1. Designer provides `telegram-icon.svg` and `telegram-rounded.svg` (same dimensions and style as `bluesky-icon.svg` / `bluesky-rounded.svg`). Developer places them in `src/assets/img/integration/`.
2. Developer registers Telegram in the core platform composables and utilities:
   - `src/modules/integration/components/platforms/social_v2/composables/useSocialPlatforms.js` — add `telegram` to platforms array and mapping object
   - `src/modules/common/composables/useSocialChannels.js` — add telegram to channel name and type mappings
   - `src/modules/common/lib/integrations.js` — add `'telegram'` to both `socialPlatformNames` functions (deprecated and current)
   - `src/composables/usePlatform.js` — add `getTelegramPlatformByID()` and `telegramImage()` methods, following the pattern of `getBlueskyPlatformByID()` and `blueskyImage()`
   - `src/components/common/SocialIcon.vue` — add Telegram SVG path to the `ICONS` object so the icon renders when `type="telegram"` is passed
   - `src/modules/setting/components/workspace/team/platformConfigs.js` — add telegram SVG icon import and a `telegram` config entry (type, label, iconSrc)
3. **Onboarding:** A first-time user going through the onboarding wizard (`SocialPlatform.vue`) sees a Telegram tile. Clicking it opens the existing Connect Telegram modal (`AddTelegram.vue`). On successful connection, Telegram is marked as connected in the onboarding checklist.
4. **CSV Bulk Upload automation:** A user creating a CSV automation sees Telegram in the account selection step. They can pick connected Telegram channels/groups as publishing destinations.
5. **Evergreen Automation:** A user creating an Evergreen automation sees Telegram in the account selection step (`EvergreenAccountSelection.vue`). They can pick connected Telegram channels/groups.
6. **Workspace Team Permissions:** A workspace admin opens Settings → Members → selects a team member → clicks "Social Account Access". The modal shows a Telegram toggle alongside Facebook, Instagram, etc. Toggling it on/off grants or revokes that member's ability to post to Telegram accounts.
7. **Manage Limits:** A workspace admin opens Settings → Manage Limits → Social Accounts tab. A Telegram row appears showing the workspace's current Telegram account count vs. the plan limit.
8. **Billing plan comparison:** On the pricing/upgrade page, Telegram appears in the "Supported social networks" list in the plan feature comparison table.

---

**Acceptance criteria:**

- [ ] `telegram-icon.svg` and `telegram-rounded.svg` exist in `src/assets/img/integration/` and match the style of other platform icons
- [ ] `useSocialPlatforms.js` includes `telegram` in both the platforms array and mapping object
- [ ] `useSocialChannels.js` includes `telegram` in both channel name and type mappings
- [ ] `integrations.js` `socialPlatformNames()` includes `'telegram'`
- [ ] `usePlatform.js` exposes `getTelegramPlatformByID()` and `telegramImage()` methods
- [ ] `SocialIcon.vue` renders the Telegram icon when `type="telegram"` is passed
- [ ] `platformConfigs.js` includes a `telegram` entry with correct SVG import, label (`'Telegram'`), and type
- [ ] **Onboarding:** Telegram tile appears in `SocialPlatform.vue`; clicking it opens the Connect Telegram modal; successful connection marks the onboarding step as complete
- [ ] **CSV Automation:** Telegram accounts appear in the account selection step of CSV bulk upload automation creation
- [ ] **Evergreen Automation:** Telegram accounts appear in the account selection step of Evergreen automation creation
- [ ] **Team Permissions:** Social account access modal for a team member includes a Telegram toggle; granting/revoking access works correctly
- [ ] **Manage Limits:** `SocialTab.vue` shows a Telegram row with current account count and the workspace's plan limit
- [ ] **Billing:** `plansComparison.js` includes `'Telegram'` in the `socialNetworks` array shown in plan feature comparison
- [ ] No existing platform's rendering or behaviour is broken by these additive changes

---

**Mock-ups:**

- Onboarding tile: follow the Bluesky tile pattern in `SocialPlatform.vue`
- Team permissions modal: follow the existing per-platform toggle pattern (see Facebook/Instagram rows for reference)
- Manage Limits row: follow existing platform row pattern in `SocialTab.vue`
- All other changes are list additions with no new UI components required

---

**UI copy:**

- **Onboarding tile label:** "Telegram"
- **Onboarding tile subtext:** "Channels & Groups"
- **Team permissions toggle label:** "Telegram"
- **Manage Limits row label:** "Telegram Accounts"

---

**Impact on existing data:** None — all additive.

---

**Impact on other products:**

- The onboarding change affects new user registration flow — must be tested with a fresh workspace.
- Team permissions change is workspace-scoped — no cross-workspace impact.

---

**Dependencies:**

- Depends on: **[BE] Register Telegram in platform registries, permissions, automation configs, and account validity job** (backend must accept `'telegram'` in validation rules before FE calls go through)
- Depends on: **[FE] Add Telegram account connection flow in Settings → Social Accounts** (`AddTelegram.vue` must exist before the onboarding step can open it)
- **[Design] dependency:** `telegram-icon.svg` and `telegram-rounded.svg` assets must be delivered before this story can be completed

---

**Global quality & compliance:**

- [ ] Mobile responsiveness tested (onboarding, settings, billing pages must be responsive)
- [ ] Multilingual support verified (all new string keys added to locale files; translations available or fallback handled)
- [ ] UI theming supported (default + white-label; design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 12: [BE] Audit and verify Telegram support in ContentStudio public API

**Group:** Backend
**Project:** Web App
**Priority:** Medium (P1)
**Product Area:** Integrations
**Skill Set:** Backend
**Story Type:** feature

---

**Description:**

ContentStudio's public API powers third-party integrations including Zapier, Make.com, and direct API consumers. As Telegram accounts are added to the platform, the public API must be audited to confirm that Telegram accounts are returned by the social accounts endpoint, that Telegram posts can be scheduled via the API, and that no explicit platform allowlists block Telegram requests.

This story is investigative first: if the API already works for Telegram through the backend additions in the other stories, the work is documenting and verifying that. If gaps exist, they must be patched and the OpenAPI/Swagger docs updated.

---

**Workflow:**

1. Developer audits the public API routes (`routes/api.php`) and all controllers involved in:
   - `GET /api/v1/social-accounts` (or equivalent) — does it return Telegram accounts?
   - `POST /api/v1/posts` (or equivalent) — does it accept Telegram `account_ids` and schedule the post correctly?
   - Any endpoint that filters by `platform` or `platform_type` — are there explicit allowlists that need updating?
2. Developer checks `storage/api-docs/api-docs.json` (Swagger/OpenAPI spec) for any platform-specific enumeration that excludes Telegram.
3. Developer tests end-to-end via the API:
   - Connect a Telegram account via the connection flow
   - Call `GET /api/v1/social-accounts` — confirm Telegram account is returned
   - Call the schedule post endpoint with the Telegram `account_id` — confirm the post is queued
4. If any gaps are found (allowlist blocks, missing platform case, validation rule exclusion): developer patches the relevant controller/validator and adds Telegram.
5. Developer updates `api-docs.json` to include Telegram as a supported platform in all relevant endpoint schemas.
6. Developer documents in the story comments whether Make.com and Zapier integrations (which consume the public API) work automatically or require separate app-module updates by the integrations team.

---

**Acceptance criteria:**

- [ ] `GET /api/v1/social-accounts` (or equivalent) returns connected Telegram accounts when the workspace has them
- [ ] The post scheduling endpoint accepts Telegram `account_id` values and successfully queues the post
- [ ] No explicit platform allowlist in the public API routes or controllers blocks `'telegram'`
- [ ] `api-docs.json` (Swagger/OpenAPI) is updated: `'telegram'` appears as a valid value in all platform-related enum fields
- [ ] End-to-end test passes: connect Telegram account → schedule post via API → post appears in the planner as scheduled
- [ ] A comment on this story documents whether Make.com / Zapier require separate integration module updates, with a follow-up story created if so

---

**Mock-ups:** N/A — backend/API only

---

**Impact on existing data:** None — all additive or documentation changes.

---

**Impact on other products:**

- If Make.com / Zapier modules hardcode a platform list (outside ContentStudio's codebase), those integrations will need separate updates by whoever manages the Make.com/Zapier app modules. This story surfaces that finding.

---

**Dependencies:**

- Depends on: **[BE] Create Telegram account integration infrastructure**
- Depends on: **[BE] Implement Telegram account connection API**
- Depends on: **[BE] Implement Telegram post publishing (text, image, video, album)**
- Depends on: **[BE] Register Telegram in platform registries, permissions, automation configs, and account validity job** (platform registries must include Telegram before API validation passes)

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (API only)
- [ ] Multilingual support — N/A (no user-facing strings)
- [ ] UI theming support — N/A (API only)
- [ ] White-label domains impact reviewed (public API is used by white-label clients — confirm Telegram is not blocked in white-label API configs)
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 13: [BE] Implement Telegram first comment publishing

**Group:** Backend
**Project:** Web App
**Priority:** Medium (P1)
**Product Area:** Integrations
**Skill Set:** Backend
**Story Type:** feature

---

**Description:**

As a user, I want ContentStudio to automatically post a follow-up reply to my Telegram post immediately after it publishes so that I can add context, credits, or a call-to-action as a threaded message without doing it manually.

ContentStudio already supports first comment scheduling for LinkedIn and other platforms. This story extends that existing architecture to Telegram. After the main Telegram post publishes successfully, if `telegram_sharing_details.first_comment.enabled` is `true` and `first_comment.text` is non-empty, a follow-up `sendMessage` call is made with `reply_to_message_id` set to the `message_id` returned by the original post's publishing response.

---

**Workflow:**

1. `TelegramPosting::performPosting()` publishes the main post and receives a response containing `message_id` from Telegram.
2. Developer ensures `message_id` is stored in the posting log/plan response (already handled by **[BE] Implement Telegram post publishing (text, image, video, album)** — confirm it persists accessibly).
3. After the main post is marked as published, `TelegramPosting` checks `telegram_sharing_details.first_comment`:
   - If `enabled: true` and `text` is non-empty: dispatches `TelegramFirstCommentJob` (or calls inline) with `chat_id`, `message_id`, and `first_comment.text`.
   - If `enabled: false` or `text` is empty: skips silently.
4. `TelegramFirstCommentJob` calls `POST https://api.telegram.org/bot{token}/sendMessage` with:
   - `chat_id`: the channel/group chat ID
   - `text`: `first_comment.text`
   - `reply_to_message_id`: the `message_id` from the main post
5. On success: marks first comment as published in the plan/posting log.
6. On HTTP 429: retries with `Retry-After` backoff (up to 3 retries). If exhausted: marks first comment as failed and adds a note to the plan: "First comment could not be posted due to rate limiting."
7. On HTTP 400/403: marks first comment as failed with the Telegram error description. The main post status remains Published.
8. If the main post itself failed (no `message_id` available): first comment job is never dispatched.

---

**Acceptance criteria:**

- [ ] If `telegram_sharing_details.first_comment.enabled` is `true` and `text` is non-empty, a `sendMessage` call is made after the main post publishes successfully
- [ ] The follow-up `sendMessage` includes `reply_to_message_id` set to the `message_id` of the main post
- [ ] If the main post fails (no `message_id`), the first comment is not dispatched
- [ ] HTTP 429 on first comment triggers retry with `Retry-After` backoff; after 3 retries, first comment is marked failed; main post status is unaffected
- [ ] HTTP 400/403 on first comment marks first comment as failed with error description; main post status remains Published
- [ ] If `first_comment.enabled` is `false` or `text` is empty, no follow-up call is made
- [ ] First comment failure does not affect the Published status of the main post
- [ ] Unit tests: first comment sent after successful main post; first comment skipped when main post fails; rate limit retry on first comment

---

**Mock-ups:** N/A — backend only

---

**Impact on existing data:**

- `plans` collection: reads `telegram_sharing_details.first_comment` — no schema migration needed (MongoDB is schemaless).
- Posting logs: first comment result appended to the Telegram posting log entry.

---

**Impact on other products:** None — additive only.

---

**Dependencies:**

- Depends on: **[BE] Implement Telegram post publishing (text, image, video, album)** — `message_id` must be stored in the posting response before this story can use it
- Depends on: **[FE] Add first comment and spoiler blur to Telegram Composer** — frontend must write `first_comment` into `telegram_sharing_details` for this story to have anything to process

---

**Global quality & compliance:**

- [ ] Mobile responsiveness — N/A (backend only)
- [ ] Multilingual support — N/A (backend only; error messages English-only for v1)
- [ ] UI theming support — N/A (backend only)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

---

### Story 14: [FE] Add first comment and spoiler blur to Telegram Composer

**Group:** Frontend
**Project:** Web App
**Priority:** Medium (P1)
**Product Area:** Composer
**Skill Set:** Frontend
**Story Type:** feature

---

**Description:**

As a user, I want to write a first comment directly in the Telegram Composer that ContentStudio will auto-post as a reply after my main post publishes, and I want to optionally blur media as a spoiler — so that I can add context and manage content visibility without extra manual steps.

This story adds two features to `EditorTelegramBox.vue` (or its equivalent):
1. A **First Comment** collapsible section (following the existing LinkedIn first comment UX pattern in the Composer).
2. A **Spoiler** toggle in the Telegram Options panel (only visible when media is attached).

Both features save their state into `telegram_sharing_details` which is already read by the backend publishing stories.

---

**Workflow:**

1. User selects a Telegram account in the Composer.
2. The Telegram Options panel (`EditorTelegramBox.vue`) shows existing toggles (silent mode, disable link preview, protect content) plus the new **Spoiler** toggle at the bottom of the options panel.
3. **Spoiler toggle** (visible only when an image or video is attached):
   - User attaches a photo or video to the post. The Spoiler toggle appears.
   - User turns on the Spoiler toggle. The TelegramPreview updates to show the media with a blur overlay and a "SPOILER" label.
   - User turns off the toggle. Preview returns to normal.
   - If user removes all media, the Spoiler toggle is hidden (and its value reset to `false`).
4. Below the Telegram Options panel, a **"First Comment"** collapsible section is present (collapsed by default, same as the existing LinkedIn first comment pattern).
5. User expands the First Comment section. A text area appears with:
   - Placeholder: "Write your first comment…"
   - Character counter: remaining characters out of 4,096
   - The section header shows a character count badge when text is present (same as LinkedIn first comment UX)
6. User types their first comment. The state is immediately reflected in `telegram_sharing_details.first_comment`.
7. User schedules the post. `telegram_sharing_details` is saved with `first_comment: { enabled: true, text: "..." }`.
8. If user clears the first comment text area or collapses the section without entering text, `first_comment.enabled` is `false`.

---

**Acceptance criteria:**

- [ ] **Spoiler toggle** appears in the Telegram Options panel only when an image or video is attached; not shown for text-only posts
- [ ] Spoiler toggle label: "Spoiler", subtext: "Blur media until viewed", with ℹ tooltip: "Adds a blur effect to your photo or video. Recipients see a blurred preview with a SPOILER label and tap to reveal the full content. Useful for reveals, surprises, or sensitive media."
- [ ] Spoiler toggle defaults to off
- [ ] When Spoiler is toggled on, `TelegramPreview` shows the attached media blurred with a SPOILER overlay
- [ ] When Spoiler is toggled on, `telegram_sharing_details.spoiler: true` is saved with the plan
- [ ] When media is removed, Spoiler toggle is hidden and `spoiler` resets to `false`
- [ ] **First Comment section** is a collapsible panel below the Telegram Options, collapsed by default
- [ ] First Comment text area has placeholder "Write your first comment…" and a live 4,096-character counter
- [ ] Character counter turns yellow at 80% and red at 100% (same pattern as main text counter)
- [ ] When text is entered, `first_comment: { enabled: true, text: "..." }` is saved in `telegram_sharing_details`
- [ ] When the section is collapsed with no text (or text is cleared), `first_comment.enabled: false`
- [ ] First Comment section follows the existing LinkedIn first comment UX pattern for consistency (reuse components where possible)
- [ ] All new strings are wrapped in `$t()` with keys added to `src/locales/*/composer.json`

---

**UI copy:**

| Element | Copy |
|---|---|
| Spoiler toggle label | Spoiler |
| Spoiler toggle subtext | Blur media until viewed |
| Spoiler tooltip | Adds a blur effect to your photo or video. Recipients see a blurred preview with a SPOILER label and tap to reveal the full content. Useful for reveals, surprises, or sensitive media. |
| First Comment section header | First Comment |
| First Comment placeholder | Write your first comment… |
| First Comment character counter | [n] / 4,096 characters |
| TelegramPreview spoiler label | SPOILER |

---

**Mock-ups:**

- Spoiler toggle: follows the same layout as the "Protect content" toggle above it in `EditorTelegramBox.vue`
- TelegramPreview spoiler state: media shown with a CSS blur filter and a centered "SPOILER" badge overlay
- First Comment section: follow the existing `EditorLinkedinBox.vue` or equivalent first comment component pattern

---

**Impact on existing data:** None — reads/writes `telegram_sharing_details` which already exists in `Plans.$fillable`.

---

**Impact on other products:** None.

---

**Dependencies:**

- Depends on: **[FE] Add Telegram to Composer — account selector, preview, and options panel** (`EditorTelegramBox.vue` must exist)
- Depends on: **[BE] Implement Telegram first comment publishing** for end-to-end testing of the first comment flow

---

**Global quality & compliance:**

- [ ] Mobile responsiveness tested (Composer is desktop-first but tablet must not break)
- [ ] Multilingual support verified (all new string keys in locale files)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

## Story Summary

| # | Title | Group | Project | Priority | Product Area |
|---|---|---|---|---|---|
| 1 | [BE] Create Telegram account integration infrastructure | Backend | Web App | High (P0) | Integrations |
| 2 | [BE] Implement Telegram account connection API | Backend | Web App | High (P0) | Integrations |
| 3 | [BE] Implement Telegram post publishing (text, image, video, album) | Backend | Web App | High (P0) | Integrations |
| 4 | [BE] Add Telegram as RSS automation publishing destination | Backend | Web App | Medium (P1) | Automation |
| 5 | [FE] Add Telegram account connection flow in Settings → Social Accounts | Frontend | Web App | High (P0) | Integrations |
| 6 | [FE] Add Telegram to Composer — account selector, preview, and options panel | Frontend | Web App | High (P0) | Composer |
| 7 | [FE] Display Telegram posts on the Planner calendar | Frontend | Web App | High (P0) | Planner |
| 8 | [iOS] Add Telegram accounts to mobile Composer account selector and Planner calendar | Product | Mobile | Medium (P1) | iOS Mobile |
| 9 | [Android] Add Telegram accounts to mobile Composer account selector and Planner calendar | Product | Mobile | Medium (P1) | Android Mobile |
| 10 | [BE] Register Telegram in platform registries, permissions, automation configs, and account validity job | Backend | Web App | High (P0) | Throughout Product |
| 11 | [FE] Add Telegram to onboarding, automations, settings, and product-wide platform lists | Frontend | Web App | High (P0) | Throughout Product |
| 12 | [BE] Audit and verify Telegram support in ContentStudio public API | Backend | Web App | Medium (P1) | Integrations |
| 13 | [BE] Implement Telegram first comment publishing | Backend | Web App | Medium (P1) | Integrations |
| 14 | [FE] Add first comment and spoiler blur to Telegram Composer | Frontend | Web App | Medium (P1) | Composer |
