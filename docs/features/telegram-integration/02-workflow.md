# Telegram Integration — Workflow Design

**Feature:** Telegram Integration for ContentStudio  
**Date:** April 1, 2026  
**Pipeline Step:** 02 — Workflow Design  

---

## 1. Feature Placement

Telegram integration surfaces across four areas of ContentStudio:

| Area | Entry Point | What Changes |
|---|---|---|
| **Settings → Social Accounts** | "Connect Social Account" flow | Telegram tile added to the social platform grid alongside Facebook, LinkedIn, X, TikTok, etc. |
| **Composer** | Account selector + options panel | Telegram channels/groups appear in the account picker; a Telegram-specific options panel (silent mode, disable link preview, protect content, spoiler blur) appears when a Telegram account is selected; a first comment section allows scheduling a follow-up reply; a Telegram-native preview renders in the right preview panel |
| **Planner (Calendar)** | Calendar event tiles | Scheduled Telegram posts appear on the calendar with the Telegram icon, same as any other platform |
| **RSS Automation** | Automation → RSS Feeds | Telegram channels/groups available as destinations in the RSS feed-to-post automation |

---

## 2. User Flow — Connecting a Telegram Account (Happy Path)

**Context:** User wants to add their Telegram channel or group to ContentStudio so they can schedule posts to it.

The connection flow is identical from the user's perspective regardless of whether the channel/group is public or private. The difference is handled transparently by the backend.

1. User navigates to **Settings → Social Accounts**.
2. User sees the Telegram tile in the platform grid (alongside Facebook, LinkedIn, etc.) and clicks **"Connect"**.
3. A modal opens: **"Connect Telegram"**. It contains:
   - A single input field labeled **"Telegram Group or Channel"** with placeholder text: *"@username or invite link"*
   - An info icon (ℹ) next to the label that, on hover, explains: "For public channels or groups, enter the @username (e.g. @mychannel). For private groups, paste the invite link from Telegram (e.g. https://t.me/+xxxx). Make sure you've added @contentstudio_bot as an admin first."
   - A toggle with the label: **"I confirm to have added @contentstudio_bot as an admin"**
   - **"Cancel"** and **"Add"** buttons. The "Add" button is disabled until the input is non-empty and the toggle is on.
4. User adds **@contentstudio_bot** as an administrator to their channel or group in Telegram (outside ContentStudio), then returns to the modal.
5. User enters their channel/group identifier and turns on the confirmation toggle, then clicks **"Add"**.
6. Backend resolves the account using one of two paths:
   - **Public channel/group (@username):** Calls `getChat` + `getChatMember` synchronously to verify the channel/group exists and the bot has admin permissions.
   - **Private group (invite link `https://t.me/+xxxx`):** When the bot was added as admin, Telegram fired a `my_chat_member` webhook update which ContentStudio stored in Redis (keyed by invite hash, 30-minute TTL). The backend extracts the hash from the invite link, looks up the `chat_id` in Redis, then calls `getChatMember` to verify admin status.
7. On success, the account is stored in the `social_integrations` collection.
8. The modal shows a success state: **"✓ [Channel Name] connected!"** with the channel's name and avatar.
9. User clicks **"Done"** and sees the Telegram account listed in their connected social accounts.
10. User can repeat to connect additional Telegram channels or groups.

---

## 3. User Flow — Publishing a Post to Telegram (Happy Path)

**Context:** User wants to create and schedule a post to a connected Telegram channel.

1. User opens the **Composer**.
2. In the account selector, user sees their Telegram channels/groups listed under a Telegram section (with the Telegram logo). User selects one or more.
3. The **Composer preview panel** on the right renders a Telegram-style preview: white message bubble with the channel name, avatar, and content as it will appear in Telegram.
4. User types their post content in the main text editor. If a URL is included, they see a toggle to **"Disable link preview"** in the Telegram options panel.
5. Below the main editor (or in a collapsible Telegram options panel), user sees Telegram-specific options:
   - **Silent post** toggle — "Send without push notification" (tooltip: "Your subscribers will receive this post but won't get a push notification. Useful for non-urgent updates like routine posts or late-night content.")
   - **Disable link preview** toggle — "Don't show link preview" (tooltip: "Telegram automatically shows a preview card for links in your post. Turn this off to keep the post cleaner if the link is already clear from context.")
   - **Protect content** toggle — "Prevent forwarding and saving" (tooltip: "When enabled, recipients will not be able to forward this post to other chats or save media from it. Useful for exclusive or paid community content.")
   - **Spoiler** toggle — "Blur media until viewed" (only visible when an image or video is attached; tooltip: "Adds a blur effect to your photo or video. Recipients see a blurred preview with a SPOILER label and tap to reveal the full content. Useful for reveals, surprises, or sensitive media."). Default: off.
6. User can add media: single image, single video, or an album of up to 10 images/videos.
7. Optionally, below the Telegram options panel, user can expand the **"First Comment"** section to schedule a follow-up reply. The user types the comment text (up to 4,096 characters); a character counter is shown. This comment is automatically posted as a reply to the original Telegram post immediately after it publishes.
8. User clicks **"Schedule"**, selects date/time, and clicks **"Confirm"**.
9. The post appears on the **Planner calendar** with the Telegram icon at the scheduled time.
10. At the scheduled time, ContentStudio fires `sendMessage` / `sendPhoto` / `sendVideo` / `sendMediaGroup` to the Telegram Bot API. If a first comment is set, a follow-up `sendMessage` with `reply_to_message_id` is posted immediately after.
11. Post status updates to **"Published"** in the Planner. If publishing fails, status shows **"Failed"** with an error message.

---

## 4. User Flow — Viewing Telegram Posts in the Planner

1. User navigates to **Planner (Calendar)**.
2. If filtering is enabled, user can filter by Telegram accounts.
3. Telegram posts appear as calendar tiles with the Telegram blue paper-plane icon.
4. Clicking a tile opens the post detail panel showing: content, scheduled time, Telegram account name, post status (Scheduled / Published / Failed).
5. For **Failed** posts, the panel shows the failure reason (e.g., "Bot is no longer an admin of this channel") and a **"Retry"** button.

---

## 5. Alternative Flows & Edge Cases

### 5a. Connection: Bot Not Added as Admin
- If the user submits the modal but @contentstudio_bot is not yet an admin in the specified channel/group, the `getChatMember` check fails and the modal shows an inline error below the input: "It looks like @contentstudio_bot isn't an admin of this channel yet. Please add it as an administrator in Telegram and try again."
- The modal remains open; the input and toggle are preserved so the user can retry without re-entering anything.

### 5b. Connection: Private Group — Invite Link Not Yet Seen by Bot
- If the user submits a private group invite link but the bot has not yet been added as admin (so no `my_chat_member` event was received and stored in Redis), the backend cannot resolve the `chat_id`.
- The modal shows an inline error: "We couldn't find this group. Make sure you've added @contentstudio_bot as an admin in Telegram first, then try again."
- Once the user adds the bot as admin in Telegram (which triggers the webhook), they can retry and the invite link will resolve successfully.
- If the Redis entry has expired (30-minute TTL), the user must remove and re-add the bot as admin to generate a fresh webhook event.

### 5c. Publishing: Bot Removed as Admin
- If between scheduling and publishing the bot was removed as admin from the channel, the publish job fails.
- The account's `validity` is set to `invalid` in the database.
- User receives an in-app notification: "Your Telegram account [Channel Name] needs to be reconnected. The bot was removed as an admin."
- The Telegram tile in Settings → Social Accounts shows a red "Reconnect" badge.

### 5h. Follow-up Comment: Main Post Failed
- If the main Telegram post fails to publish (any error), the first comment job is **not dispatched**. There is no original `message_id` to reply to.
- The first comment is discarded and the post is marked Failed with the main post's error reason.
- No separate error is shown for the first comment — the failure of the main post implicitly means the first comment will not be sent.

### 5i. Follow-up Comment: Rate Limit After Main Post
- If the main post succeeds but the follow-up `sendMessage` hits a rate limit, it retries using the same `Retry-After` backoff logic as main posts (up to 3 retries).
- If retries are exhausted, the post remains Published (main post succeeded) but the first comment is marked as Failed. A note is added to the post detail: "Your first comment could not be posted. Retry manually from the Planner."

### 5d. Publishing: Rate Limit Hit
- If Telegram returns HTTP 429, the posting job retries with exponential backoff using the `Retry-After` header value.
- If retry exhausted, the post is marked **Failed** with the message: "Telegram rate limit reached. Please retry manually."

### 5e. Publishing: File Too Large
- If media exceeds 50 MB, ContentStudio rejects it at the Composer level before submission with: "This file is too large for Telegram (max 50 MB). Please compress the file or use a smaller version."

### 5f. Multiple Telegram Accounts Selected
- User can select multiple Telegram channels/groups in the Composer simultaneously (same as selecting multiple Facebook pages).
- The post is queued and published to each channel independently. Each gets its own status (one can succeed while another fails).

### 5g. Account Disconnected Mid-Flow
- If a Telegram account is disconnected while the user has it selected in the Composer, the Composer shows a warning banner: "[Channel Name] has been disconnected. Please reconnect it in Settings before publishing."
- The account appears grayed out in the account selector with a warning icon.

---

## 6. Key Design Decisions

### Decision 1: Bot Connection Model — Shared ContentStudio Bot vs. User's Own Bot

**Option A (Recommended): Shared ContentStudio bot (@contentstudio_bot)**
- ContentStudio runs one shared bot. Users add this bot as admin to their channels/groups, then enter the @username or invite link in ContentStudio. The UX is identical for public and private — the backend handles both transparently:
  - **Public channels/groups:** Resolved synchronously via `getChat` + `getChatMember`.
  - **Private groups:** Resolved via a `my_chat_member` webhook event (fired when the bot is added as admin), with the `chat_id` cached in Redis for 30 minutes until the user submits the invite link.
- **Pros**: Single clean UX for all cases (matches Publer exactly), supports both public and private groups, immediate feedback on submission.
- **Cons**: ContentStudio's bot name appears in the channel admin list; some brands don't like "foreign" bots having admin access. Private group connection requires the webhook to have already fired (bot must be added before the user submits).

**Option B: User provides their own bot token**
- User creates a bot via @BotFather, gets a token, and enters it in ContentStudio.
- **Pros**: Brand has their own named bot (e.g., @mybrandbot); feels more enterprise/white-label.
- **Cons**: Significantly more complex UX; requires token validation, error handling for invalid tokens; only tech-savvy users will use it.

**Recommendation:** Ship Option A in v1 (shared bot). Add Option B as a premium feature in v2 — this aligns with how Postly differentiates their branded bot option.

---

### Decision 2: Telegram-Specific Options Placement — Inline vs. Dedicated Panel

**Option A (Recommended): Collapsible "Telegram Options" panel below the main composer**
- Same pattern used for other platform-specific options (e.g., LinkedIn first comment, Pinterest board selector).
- Shows only when a Telegram account is selected.
- Contains: silent mode toggle, disable link preview toggle.
- **Pros**: Consistent with existing UX patterns; doesn't clutter the main composer.
- **Cons**: Users might miss it if collapsed by default.

**Option B: Inline toggles in the platform-specific editor box**
- Options appear directly in the account selector row next to the Telegram account chip.
- **Pros**: More visible.
- **Cons**: Clutters the account selector; hard to scale when more options are added in v2.

**Recommendation:** Option A — collapsible panel, open by default when a Telegram account is first selected.

---

### Decision 3: Telegram Preview Fidelity — Pixel-Perfect vs. Functional

**Option A (Recommended): Functional Telegram-style preview**
- Shows the approximate layout: white bubble with channel name, avatar, message content, and media thumbnail.
- Truncates at Telegram's character limits (4,096 text / 1,024 caption) with a warning counter.
- Does not attempt to pixel-perfect replicate Telegram's exact font and spacing.
- **Pros**: Builds quickly, conveys critical information (content length, media layout), maintainable.
- **Cons**: Not exact — spacing and fonts differ slightly.

**Option B: Static mockup-style preview**
- A static gray card labeled "Telegram" showing only the text content, no channel chrome.
- **Pros**: Fastest to build.
- **Cons**: Doesn't give the user confidence their post will look right; inferior to what Publer offers.

**Recommendation:** Option A — functional preview with channel name, avatar, content, and character limit counter.

---

## 7. Integration with Existing Features

| Existing Feature | Integration |
|---|---|
| **Composer** | Telegram accounts appear in the account selector. The existing multi-platform composer dispatches to Telegram alongside other platforms. Platform-specific caption customization (existing feature) allows different text per platform — Telegram gets its own customized copy if needed. |
| **Planner / Calendar** | Telegram posts appear on the calendar with the Telegram icon. No calendar changes needed beyond adding Telegram as a known platform type. |
| **Media Library** | Files from the Media Library can be attached to Telegram posts the same as any other platform. The 50 MB limit is enforced at selection time. |
| **RSS Automation** | The existing RSS-to-post automation already supports multiple platforms as destinations. Telegram channels/groups are added to the destination picker. No structural changes to the automation engine — just an additional publishing destination. |
| **First Comment (Follow-up)** | ContentStudio already supports first comment scheduling for LinkedIn and other platforms. The same architecture is reused for Telegram: after the main post publishes, the stored `message_id` is used to post a `reply_to_message_id` follow-up via `sendMessage`. No changes to the first-comment scheduling engine — only a Telegram handler is added. |
| **Account Health / Notifications** | The existing validity checking system (which flags expired/invalid social accounts) is extended to flag Telegram accounts where the bot was removed as admin. |
| **Analytics** | v1: No Telegram analytics (Bot API doesn't expose them). The Analytics section would not show Telegram in v1. v2: Explore surfacing basic channel stats (subscriber count via `getChatMembersCount`). |
| **Content Approval Workflow** | Telegram posts go through the same approval flow as any other platform post — no changes needed. |
| **Bulk Scheduling / CSV Upload** | Telegram is available as a destination in bulk scheduling/CSV upload (same as other platforms). |

---

## 8. Scope Recommendation

### v1 — Launch Scope (This Epic)
- Connect Telegram Channels and Groups via shared @contentstudio_bot
- Publish text, single image, single video, and albums (up to 10 items)
- Telegram accounts in Composer account selector
- Telegram-style post preview in Composer
- Telegram options panel: silent mode toggle, disable link preview toggle, protect content toggle, spoiler blur toggle (media only)
- First comment (follow-up reply) scheduling from the Composer
- Planner calendar shows Telegram posts
- Settings → Social Accounts: Telegram tile (connect, disconnect, reconnect)
- RSS Automation: Telegram as a destination
- Account validity notifications (bot removed as admin)
- iOS/Android: display Telegram accounts in mobile account selector; show Telegram posts in mobile planner

### Defer to v2
- Inline CTA button keyboards (URL buttons below post)
- Poll creation and scheduling
- Telegram analytics (post views, subscriber count trends)
- Branded bot support (custom @bot token)
- AI caption generation with Telegram-optimized prompts
- Pin post to top of channel
