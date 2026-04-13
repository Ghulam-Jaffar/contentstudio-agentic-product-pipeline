# Telegram Integration — Workflow Design

**Feature:** Telegram Integration for ContentStudio  
**Last Updated:** April 6, 2026  
**Pipeline Step:** 02 — Workflow Design  

---

## 1. Feature Placement

Telegram integration surfaces across four areas of ContentStudio:

| Area | Entry Point | What Changes |
|---|---|---|
| **Settings → Social Accounts** | "Connect Social Account" flow | Telegram tile added to the platform grid. Clicking opens a 3-step bot-token connection modal. |
| **Composer** | Account selector + options panel | Telegram channels/groups appear in the account picker; a Telegram-specific options panel (silent message, disable link preview, pin message) appears when a Telegram account is selected; a Telegram-native preview renders in the right preview panel |
| **Planner (Calendar)** | Calendar event tiles | Scheduled Telegram posts appear on the calendar with the Telegram icon, same as any other platform |
| **RSS Automation** | Automation → RSS Feeds | Telegram channels/groups available as destinations in the RSS feed-to-post automation |

---

## 2. User Flow — Connecting a Telegram Account (Happy Path)

**Context:** User wants to add their Telegram channel or group to ContentStudio. Unlike OAuth-based platforms, Telegram uses a bot token model — the user creates their own bot via @BotFather, gives it admin access to their channel/group, and provides the token to ContentStudio.

### Step 1 — Enter Bot Token

1. User navigates to **Settings → Social Accounts**.
2. User sees the **Telegram** tile in the platform grid and clicks **"Connect"**.
3. A modal opens: **"Connect Telegram"** (Step 1 of 3). It contains:
   - A field labeled **"Bot Token"** with placeholder: *"Paste your bot token from @BotFather"*
   - A collapsible helper section: **"How to create a Telegram Bot"** (5-step instructions)
   - A **"Validate"** primary button
4. User creates a bot via @BotFather if they haven't already, copies the token, pastes it in, and clicks **"Validate"**.
5. Frontend calls `POST /telegram/validate-bot` with `{ bot_token, workspace_id }`.
6. While the request is in-flight, the "Validate" button shows a loading spinner.
7. On success: backend returns the bot's name and @username. The modal shows a bot info card (bot name, @username) and auto-advances to Step 2.
8. On error: inline error below the input — "Invalid bot token. Please check and try again." Modal stays on Step 1.

### Step 2 — Select Channels & Groups

1. On mount, frontend calls `POST /telegram/discover-chats` with `{ bot_token, workspace_id }` to fetch all chats the bot has already been added to.
2. A loading spinner shows while discovery runs.
3. Discovered chats appear as a list. Each item shows: chat title (bold), type badge ("Channel", "Group", or "Supergroup"), member count (if available), and a checkbox for selection.
4. A **"Select all / Deselect all"** toggle is provided above the list.
5. A separator line followed by a **"Can't find your channel or group?"** section:
   - Instructions: "Make sure the bot is added as an admin, then enter the channel/group username or ID below"
   - An input field with placeholder: *"@channelusername or chat ID"*
   - An **"Add"** button; on click: calls `POST /telegram/validate-chat` with `{ bot_token, chat_identifier, workspace_id }`. On success: appends the validated chat to the list with checkbox pre-selected. On error: inline error below the input.
6. **Empty state** (no chats discovered and none manually added):
   - Icon + message: "No channels or groups found"
   - Instructions: "1. Add your bot as an admin to a Telegram channel or group. 2. Send any message in that channel/group. 3. Click 'Refresh' or enter the username/ID below."
   - "Refresh" button to re-call the discover endpoint.
7. The **"Next"** button is disabled until at least 1 chat is checked.
8. User selects the desired chats and clicks **"Next"**.

### Step 3 — Confirm & Connect

1. The modal shows a summary: "Confirm connection" — lists selected chats with title and type badge.
2. A note: "Each channel/group will count as one social account in your plan."
3. User clicks **"Connect"** (primary action). Loading state while saving.
4. Frontend calls `POST /telegram/add-chats` with `{ bot_token, workspace_id, chats: [...] }`.
5. On success:
   - Success toast: "Telegram channels/groups connected successfully"
   - Modal closes
   - Social accounts list refreshes — new Telegram accounts appear
6. On error (subscription limit): limit-exceeded message with upgrade link.
7. On other error: error toast.

---

## 3. User Flow — Managing Connected Chats

Each connected Telegram chat appears in the integrations settings page as a card/row showing:
- Chat title
- Type badge (Channel / Group)
- Bot username used (e.g., "via @MyContentBot")
- **"Remove"** button → confirmation dialog → `DELETE /telegram/remove-chat/{id}` → remove from list + success toast
- **"Add more"** button → re-opens the connection modal at Step 2 (same or different bot)

---

## 4. User Flow — Publishing a Post to Telegram (Happy Path)

**Context:** User wants to create and schedule a post to a connected Telegram channel.

1. User opens the **Composer**.
2. In the account selector, user sees their Telegram channels/groups listed under a Telegram section. User selects one or more.
3. The **Composer preview panel** on the right renders a Telegram-style preview (`TelegramPreview.vue`): white message bubble with the channel name, avatar, and content as it will appear in Telegram.
4. User types their post content in the main text editor.
   - **Text-only posts:** Character counter shows `X / 4096`.
   - **Media posts (image, video, PDF):** Counter switches to `X / 1024` (caption limit) as soon as media is attached.
5. Below the main editor, the **"Telegram Options"** collapsible panel (`TelegramOptions.vue`) is visible when at least one Telegram account is selected:
   - **"Silent message"** toggle — "Post without sending a notification to channel/group members". Default: off.
   - **"Disable link preview"** toggle — "Prevent automatic link preview expansion in the post". Default: off.
   - **"Pin message"** toggle — "Pin this message in the channel/group after posting". Default: off.
6. User can add media in one of two modes (mutually exclusive):
   - **Images / Videos mode:** drag & drop, media library, URL. Up to 10 items. Mixed images and videos allowed.
   - **PDF Document mode:** single .pdf file, up to 50 MB. When a PDF is attached, image/video upload is disabled and vice versa.
7. User clicks **"Schedule"**, selects date/time, and clicks **"Confirm"**.
8. The post appears on the **Planner calendar** with the Telegram icon at the scheduled time.
9. At the scheduled time, ContentStudio fires the appropriate Telegram Bot API method using the stored bot token:
   - Text only → `sendMessage`
   - Single image → `sendPhoto`
   - Single video → `sendVideo`
   - 2–10 media items → `sendMediaGroup`
   - PDF document → `sendDocument`
   - If `pin_message: true`, ContentStudio calls `pinChatMessage` immediately after a successful post.
10. Post status updates to **"Published"** in the Planner. If publishing fails, status shows **"Failed"** with an error message.

---

## 5. User Flow — Viewing Telegram Posts in the Planner

1. User navigates to **Planner (Calendar)**.
2. If filtering is enabled, user can filter by Telegram accounts.
3. Telegram posts appear as calendar tiles with the Telegram blue paper-plane icon.
4. Clicking a tile opens the post detail panel showing: content, scheduled time, Telegram account name, post status (Scheduled / Published / Failed).
5. For **Failed** posts, the panel shows the failure reason (e.g., "The bot no longer has admin permissions in this channel") and a **"Retry"** button.

---

## 6. Alternative Flows & Edge Cases

### 6a. Connection: Bot Token Invalid
- If the user submits a malformed or revoked bot token, `POST /telegram/validate-bot` returns an error.
- Inline error below the input: "Invalid bot token. Please check and try again."
- Modal stays on Step 1; input is preserved.

### 6b. Connection: No Chats Discovered
- If `POST /telegram/discover-chats` returns an empty list, the modal shows the empty state on Step 2 with instructions for how to add the bot as admin.
- "Refresh" button lets users re-poll after adding the bot to their channel.

### 6c. Connection: Manual Chat Add Fails
- If `POST /telegram/validate-chat` fails (bot not admin, chat not found), an inline error appears below the manual input: specific message from the API (e.g., "Bot is not an admin of this chat" or "Chat not found").

### 6d. Connection: Subscription Limit on Connect
- If `POST /telegram/add-chats` fails due to plan limits, a limit-exceeded message is shown with an upgrade link (same pattern as other platforms).

### 6e. Publishing: Bot No Longer Admin
- If between scheduling and publishing the bot was removed as admin from the channel, the publish job fails.
- The account's `validity` is set to `invalid` in the database.
- User receives an in-app notification: "Your Telegram account [Channel Name] needs to be reconnected. The bot no longer has admin permissions."
- The Telegram tile in Settings → Social Accounts shows a warning "Reconnect" badge.

### 6f. Publishing: Rate Limit Hit
- If Telegram returns HTTP 429, the posting job retries with exponential backoff using the `Retry-After` header value.
- After 3 retries, the post is marked **Failed** with the message: "Telegram rate limit reached. Please retry manually."

### 6g. Publishing: File Too Large
- If media exceeds 50 MB, ContentStudio rejects it at the Composer level before submission: "This file is too large for Telegram (max 50 MB). Please compress the file or use a smaller version."

### 6h. Multiple Telegram Accounts Selected
- User can select multiple Telegram channels/groups in the Composer simultaneously.
- The post is queued and published to each channel independently via its own bot token. Each gets its own status.

### 6i. PDF + Images/Videos Together
- ContentStudio enforces media type exclusivity at the Composer level.
- When a PDF is attached, image/video upload buttons are disabled (and vice versa).
- If a user switches from common_box_status to per-platform and the common editor has images, Telegram's per-platform editor shows an info banner: "Switch to PDF mode to attach a document — existing images will be used instead."

### 6j. PDF in Common Box Mode
- PDFs are Telegram-specific and are not included in `common_sharing_details`.
- When `common_box_status: true`, the PDF upload option is not available for Telegram.
- A tooltip/hint is shown: "Switch to platform-specific mode to attach a PDF for Telegram."

### 6k. Pin Message Failure
- If `pinChatMessage` fails after a successful post (e.g., bot lacks pin permissions), the post is still marked as **Published**.
- A non-blocking warning is shown in the post detail panel: "Post was published but could not be pinned. Check that your bot has the 'Pin Messages' permission."

### 6l. EasyConnect — Client Connecting Telegram via Agency Link

**Context:** EasyConnect is ContentStudio's white-label feature where agencies generate a shareable link and send it to clients. Clients visit the link and connect their own social accounts directly — without the agency needing to handle credentials. All other platforms (Facebook, Instagram, LinkedIn, etc.) use OAuth redirects for this. Telegram does not use OAuth; it uses the user's own bot token.

**Flow:**

1. Agency navigates to **Settings → Integrations → EasyConnect** and creates a shareable link (or uses an existing one).
2. Agency shares the link with a client who owns a Telegram channel or group.
3. Client visits the EasyConnect URL (e.g., `app.contentstudio.io/easy-connect/{slug}`).
4. Client sees the platform connection page showing any accounts they've already connected via this link.
5. Client clicks **"+ Connect Social Account"** — the platform selector modal opens.
6. Client selects **Telegram** from the platform grid.
7. Instead of an OAuth redirect (which doesn't apply to Telegram), the **3-step bot token connection modal** opens inline on the EasyConnect page — same UX as connecting Telegram from Settings → Social Accounts:
   - **Step 1:** Enter bot token — client pastes their @BotFather token and clicks "Validate"
   - **Step 2:** Discover and select chats — chats the bot is already admin in are listed; client selects desired channels/groups
   - **Step 3:** Confirm — client clicks "Connect"
8. Backend saves the connected chats with `connection_via_link: true` and `connection_link_id` referencing the agency's EasyConnect link.
9. Modal closes. The EasyConnect page refreshes and shows the newly connected Telegram accounts.
10. The agency's workspace now has the client's Telegram channels/groups available for publishing.

**Edge cases within EasyConnect for Telegram:**
- If the client's bot token is invalid: Step 1 shows inline error "Invalid bot token. Please check and try again." — modal stays on Step 1.
- If no chats are discovered (bot not yet added as admin): empty state on Step 2 shows setup instructions + Refresh button.
- If adding a chat manually fails (bot not admin, chat not found): inline error below the manual input field.
- If the workspace has hit its account limit: limit-exceeded message with upgrade link shown after Step 3 attempt.

---

## 7. Key Design Decisions

### Decision 1: Bot Connection Model — User's Own Bot Token ✅

**Decision (final):** Users connect their own Telegram bot token obtained from @BotFather.

- User creates a bot via @BotFather, adds it as admin to their channel/group, and enters the token in ContentStudio.
- ContentStudio stores the token per workspace account and uses it for all publishing calls.
- **Pros:** Brand has their own named bot (e.g., @mybrandbot); no shared bot risk (rate limits, abuse, ContentStudio bot name in channel admin list); enterprise/white-label friendly.
- **Cons:** More steps for the user than a shared bot approach; requires user familiarity with @BotFather.
- **Why chosen over shared bot:** Aligns with how Publer implements their enterprise tier; removes the risk of a shared bot being rate-limited or abused by one workspace affecting all others; more secure token isolation.

---

### Decision 2: Telegram-Specific Options — TelegramOptions.vue Panel

**Decision:** Dedicated `TelegramOptions.vue` component with 3 toggle switches: silent message, disable link preview, pin message.

- Shown as a collapsible panel below the main composer when at least one Telegram account is selected.
- Options: `silent_message`, `disable_link_preview`, `pin_message`.
- **Removed vs. prior design:** protect_content and spoiler blur — these were evaluated but descoped to reduce v1 scope and because they're low-adoption features that add implementation complexity (spoiler requires special handling per media type in sendMediaGroup).
- Pins are a high-value channel management feature and worth including in v1.

---

### Decision 3: Telegram Preview Fidelity — Functional Preview

**Decision:** `TelegramPreview.vue` renders a functional Telegram-style message bubble — not pixel-perfect, but shows channel name, avatar, content, media thumbnail, and character counter.

- Dynamic character counter updates based on whether media is attached (4096 text-only / 1024 media caption).
- Link preview card shown when URL is present and "Disable link preview" is off.
- Pin and silent indicators shown when the respective toggles are on.

---

### Decision 4: PDF Document Support in v1

**Decision:** Include PDF document posting (`sendDocument`) in v1.

- Single PDF per post, up to 50 MB.
- Mutually exclusive with images/videos.
- Not available when `common_box_status: true` (PDF is Telegram-specific, not common media).
- Rationale: Many Telegram channel operators regularly share documents (reports, guides, catalogs). This is a key differentiator vs. Publer which only supports images/videos.

---

## 8. Integration with Existing Features

| Existing Feature | Integration |
|---|---|
| **Composer** | Telegram accounts appear in the account selector. The existing multi-platform composer dispatches to Telegram alongside other platforms. Platform-specific caption customization allows different text per platform — Telegram gets its own customized copy if needed. |
| **Planner / Calendar** | Telegram posts appear on the calendar with the Telegram icon. No calendar changes beyond adding Telegram as a known platform type. |
| **Media Library** | Files from the Media Library can be attached to Telegram posts the same as any other platform. The 50 MB limit is enforced at selection time. |
| **RSS Automation** | The existing RSS-to-post automation supports multiple platforms as destinations. Telegram channels/groups are added to the destination picker. No structural changes to the automation engine — just an additional publishing destination. |
| **Account Health / Notifications** | The existing validity checking system is extended to flag Telegram accounts where the bot was removed as admin. |
| **Analytics** | v1: No Telegram analytics (Bot API doesn't expose them). v2: Explore surfacing basic channel stats (`getChatMembersCount`). |
| **Content Approval Workflow** | Telegram posts go through the same approval flow as any other platform post. |
| **Bulk Scheduling / CSV Upload** | Telegram is available as a destination in bulk scheduling/CSV upload. |

---

## 9. Scope

### v1 — Launch Scope (This Epic)
- Connect Telegram Channels and Groups via user's own bot token (3-step: validate token → discover chats → confirm)
- Manage connected chats: add more, remove
- Publish text, single image, single video, albums (up to 10 items), and PDF documents
- Telegram accounts in Composer account selector
- Telegram-style post preview in Composer (`TelegramPreview.vue`)
- Telegram options panel: silent message, disable link preview, pin message (`TelegramOptions.vue`)
- PDF document mode with media type exclusivity
- Dynamic character counter (4096 text / 1024 caption)
- Planner calendar shows Telegram posts
- Settings → Social Accounts: Telegram tile (connect, manage, disconnect)
- RSS Automation: Telegram as a destination
- Account validity notifications (bot removed as admin)
- iOS/Android: display Telegram accounts in mobile account selector; show Telegram posts in mobile planner

### Defer to v2
- Protect content toggle (`protect_content: true`)
- Spoiler blur for media (`has_spoiler: true`)
- First comment / follow-up reply scheduling
- Inline CTA button keyboards (URL buttons below post)
- Poll creation and scheduling
- Telegram analytics (post views, subscriber count trends)
- AI caption generation with Telegram-optimized prompts
