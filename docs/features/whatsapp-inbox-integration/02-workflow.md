# Workflow Design: WhatsApp Inbox Integration

**Date:** 2026-03-08

---

## 1. Feature Placement

### Navigation & Entry Points

| Entry Point | Location | Purpose |
|---|---|---|
| **Connect WhatsApp (Settings)** | Settings → Social Accounts → Connect a Social Account modal | Primary connection flow |
| **Connect WhatsApp (Onboarding)** | Onboarding wizard → Connect Social Accounts step | First-time setup |
| **WhatsApp Inbox** | Inbox (left sidebar) → Platform filter: "WhatsApp" | View & reply to WhatsApp conversations |
| **WhatsApp Conversations** | Inbox → Conversation list (WhatsApp icon badge on each conversation) | Individual conversation threads |

WhatsApp appears as a new platform option alongside Facebook, Instagram, Twitter, LinkedIn, and GMB in:
- Social Accounts settings page
- Onboarding connect accounts step
- Inbox platform filter
- iOS and Android social account connection screens

Users can connect **multiple WhatsApp accounts** (different phone numbers) to the same workspace.

---

## 2. User Flows

### Flow A: Connect WhatsApp Account

**Happy Path:**

1. User navigates to **Settings → Social Accounts** from the left sidebar (or encounters WhatsApp during **Onboarding → Connect Accounts** step)
2. User clicks **"Connect a Social Account"** button (existing modal)
3. In the connection modal, user sees a new **"WhatsApp"** option alongside existing platforms (Facebook, Instagram, Twitter, etc.)
4. User clicks **"Connect"** next to WhatsApp
5. User is redirected to **Facebook Business Login** (same redirect pattern as existing Facebook/Instagram connections)
6. On Facebook's page, user logs into their Facebook account (if not already logged in)
7. User selects or creates a **Meta Business Portfolio** for their business
8. User selects an existing **WhatsApp Business Account (WABA)** or creates a new one
9. User adds a phone number:
   - Enters their business phone number and country code
   - Sets a **WhatsApp Business display name** (must match business name per WhatsApp guidelines)
   - Chooses verification method: **Text message** or **Phone call**
   - Enters the verification code received
10. User grants permissions to ContentStudio (`whatsapp_business_management`, `whatsapp_business_messaging`)
11. User is redirected back to ContentStudio
12. ContentStudio shows a **"Choose Your Phone Number"** confirmation modal:
    - Phone number displayed (read-only, pre-filled from Meta)
    - **"Channel Name"** input field — user enters a friendly name for this WhatsApp account (used throughout ContentStudio)
    - "Connect" CTA button
13. User enters a channel name (e.g., "Support Line" or "Main Business") and clicks **"Connect"**
14. Success toast: **"WhatsApp account connected successfully!"**
15. WhatsApp account now appears in the Social Accounts list with a green WhatsApp icon, phone number, display name, and channel name

**Multiple Accounts:** User can repeat this flow to connect additional WhatsApp numbers. Each number is a separate channel in the inbox. The connection modal allows connecting another number even if one is already connected.

**Alternative Flows:**

- **A1: User cancels Facebook Login** → Redirected back to ContentStudio, no account created, toast: "WhatsApp connection was cancelled."
- **A2: User doesn't grant required permissions** → Error screen: "We need WhatsApp messaging permissions to connect your account. Please try again and grant the required permissions."
- **A3: Phone number already in use by another platform/BSP** → Meta shows error during signup. User must remove the number from the other platform first. We show a help article link.
- **A4: Phone verification fails** → User can retry verification (text or call). After 3 failed attempts, suggest trying the alternate method.
- **A5: User already has this phone number connected in this workspace** → Error: "This phone number is already connected to your workspace."
- **A6: Onboarding flow** → Same redirect-based connection, but after success the user returns to the onboarding wizard's next step instead of the Social Accounts page.

---

### Flow B: Receiving & Viewing WhatsApp Messages in Inbox

**Happy Path:**

1. A customer sends a message to the business's WhatsApp number
2. ContentStudio receives the message via webhook and stores it
3. User navigates to **Inbox** from the left sidebar
4. User sees the new WhatsApp conversation in the conversation list with:
   - **WhatsApp icon** (green) as the platform badge
   - **Channel name** shown if multiple WhatsApp accounts connected (e.g., "Support Line")
   - **Customer name** (from WhatsApp profile) or phone number if name unavailable
   - **Message preview** (truncated to ~60 chars)
   - **Timestamp** (relative: "2m ago", "1h ago", etc.)
   - **Unread indicator** (bold text + blue dot) if not yet viewed
5. User clicks the conversation to open it in the chat view
6. Chat view shows:
   - **Header**: Customer name/phone, WhatsApp icon, channel name (e.g., "Support Line")
   - **Message thread**: Customer messages on left, business replies on right (standard chat layout)
   - **Media inline**: Images display as thumbnails (clickable to expand), videos show play button, documents show file icon + name (clickable to download)
   - **Timestamp** on each message
7. User can **filter** the inbox conversation list by:
   - Platform: select "WhatsApp" to see only WhatsApp conversations
   - Specific WhatsApp account (if multiple connected)
   - Read/Unread status
   - Assigned team member
   - Tags

**Alternative Flows:**

- **B1: Customer sends media (image/video/document)** → Displayed inline in the chat thread. Images as thumbnails, videos with play button, documents as downloadable file cards.
- **B2: Customer sends location** → Shown as a location card with address text and a "View on Map" link
- **B3: Customer sends a sticker** → Displayed inline as an image
- **B4: Customer sends a voice note** → Shown as an audio player widget with play/pause and duration

---

### Flow C: Replying to WhatsApp Messages

**Happy Path:**

1. User is viewing a WhatsApp conversation in the Inbox chat view
2. User sees the **reply composer** at the bottom of the chat view (same as existing inbox reply area)
3. User types a text reply in the message input field
4. (Optional) User clicks the **attachment icon** (📎) to attach one file:
   - File picker opens — user selects an image, video, or document
   - Attachment preview appears above the input field showing filename/thumbnail
   - User can remove the attachment by clicking ✕ on the preview
5. User clicks **"Send"** button (or presses Enter)
6. Reply appears instantly in the chat thread with a "Sending..." indicator
7. The 24-hour window is tracked per conversation — no explicit countdown in UI since customer always initiates

**Alternative Flows:**

- **C1: 24-hour window has expired** → Reply composer shows a warning banner: "The 24-hour reply window has expired. You can no longer reply to this conversation until the customer messages again." Send button is disabled.
- **C2: Message fails to send** → Show error indicator on the message with "Failed to send. Tap to retry." Clicking retries the send.
- **C3: File exceeds size limit** → Toast: "This file is too large. Images must be under 5 MB, videos under 16 MB, and documents under 100 MB."
- **C4: Unsupported file format** → Toast: "This file format isn't supported. Supported formats: JPEG, PNG for images; MP4 for video; PDF, DOC, DOCX, XLS, XLSX for documents."
- **C5: User tries to attach multiple files** → Only one attachment per message allowed (WhatsApp limitation). Toast: "WhatsApp allows one attachment per message. Send this one first, then attach another."

---

### Flow D: Managing WhatsApp Account (Edit, Disconnect, Reconnect)

**Happy Path — Edit & Disconnect:**

1. User navigates to **Settings → Social Accounts**
2. User finds their WhatsApp account(s) in the list
3. User clicks the **three-dot menu** (⋮) on a WhatsApp account card
4. Options available:
   - **Edit Channel Name** → Inline edit of the friendly name
   - **Disconnect** → Confirmation modal: "Are you sure you want to disconnect this WhatsApp account? You'll stop receiving messages from this number in your Inbox." → "Disconnect" / "Cancel"
5. Disconnecting removes the webhook subscription and marks the account as disconnected

**Token Expiry & Reconnect Flow:**

6. If the WhatsApp access token becomes invalid (expired, revoked, permissions changed), ContentStudio detects this when a webhook event fails to process or when a reply API call returns an auth error
7. The WhatsApp account card in Social Accounts shows a **red warning badge**: "Reconnection required"
8. In the Inbox, conversations for this account show a banner: "This WhatsApp account needs to be reconnected. Go to Settings → Social Accounts to reconnect."
9. User clicks **"Reconnect"** on the account card (three-dot menu or a prominent "Reconnect" button on the warning badge)
10. User is redirected to Facebook Business Login again (same flow as initial connection, but re-authorizing the existing WABA + phone number)
11. On return, ContentStudio updates the stored access token
12. Success toast: "WhatsApp account reconnected successfully!"
13. Warning badge disappears, messages resume flowing

**Alternative Flows:**

- **D1: Token expires while user is away** → ContentStudio tracks failed API calls (existing `invalid_tries` + `sent_invalid_email` pattern). After N failed attempts, sends an email notification: "Your WhatsApp account [channel name] needs to be reconnected in ContentStudio."
- **D2: Multiple accounts, one needs reconnect** → Only the affected account shows the warning. Other accounts continue working normally.

---

### Flow E: Auto-Reply Rules for WhatsApp

**Happy Path:**

1. User navigates to **Inbox → Auto-Reply Rules** (existing feature)
2. User clicks **"Create Rule"**
3. In the rule creation modal, user now sees **"WhatsApp"** as an available platform option
4. User selects WhatsApp and configures:
   - **Trigger**: Keywords or "All messages" (same as existing platforms)
   - **Response**: Text reply (media attachments not supported in auto-replies)
   - **Active hours**: Optional schedule (e.g., outside business hours only)
   - **Account**: Select which WhatsApp account(s) the rule applies to (relevant when multiple accounts connected)
5. User saves the rule
6. When a customer message matches the trigger, the auto-reply is sent automatically within the 24-hour window

---

### Flow F: WhatsApp in Mobile Apps (iOS & Android)

**Inbox on Mobile:**

1. User opens the ContentStudio iOS/Android app
2. User navigates to **Inbox**
3. WhatsApp conversations appear in the conversation list with WhatsApp icon badge — same as web
4. User taps a conversation to view the message thread
5. User can reply with text + one attachment (same constraints as web)
6. 24-hour window expired state: reply composer disabled with banner message

**Social Account Connection on Mobile:**

1. User navigates to **Settings → Social Accounts** (or encounters it during onboarding)
2. User sees WhatsApp in the platform list
3. User taps "Connect" → opens Facebook Business Login in an in-app browser / system browser
4. Same Embedded Signup flow as web (Meta handles the responsive UI)
5. On completion, redirect back to the app → confirmation screen with channel name input
6. Account management (edit name, disconnect, reconnect) same as web but adapted to mobile UI patterns

---

## 3. Key Design Decisions

### Decision 1: Connection Flow — Redirect vs. Popup

| Option | Pros | Cons |
|---|---|---|
| **A: Redirect to Facebook (Recommended)** | Consistent with existing CS pattern for FB/IG, simpler implementation, works on all browsers and mobile | User leaves ContentStudio temporarily |
| B: Popup/Embedded iframe | User stays in ContentStudio | Popup blockers, more complex, inconsistent with existing CS flow, doesn't work well on mobile |

**Recommendation: Option A (Redirect)**. ContentStudio already uses redirects for Facebook and Instagram connections. Works consistently across web and mobile apps.

### Decision 2: 24-Hour Window Expired UX

| Option | Pros | Cons |
|---|---|---|
| **A: Disable reply + banner message (Recommended)** | Clear, prevents failed sends, simple to implement | User can't do anything until customer messages again |
| B: Show countdown timer + warning | User knows how much time is left | Adds complexity, timer may create anxiety, not actionable |
| C: Allow sending and show error after | Less restrictive | Bad UX — user writes a reply only to have it fail |

**Recommendation: Option A (Disable reply + banner)**. Track the last customer message timestamp per conversation. When > 24 hours, disable the composer and show a clear banner.

### Decision 3: Media in Replies — Upload to WhatsApp vs. Public URL

| Option | Pros | Cons |
|---|---|---|
| A: Upload to WhatsApp Media API first, then send | Works for any file, no public URL needed | Extra API call, media deleted after 30 days on Meta's side |
| **B: Use public URL (link method) (Recommended)** | Simpler, one API call, leverages CS media library | File must be publicly accessible |

**Recommendation: Option B (Public URL)**. ContentStudio's media library already stores files on CDN with public URLs. Avoids the extra upload-to-WhatsApp step.

### Decision 4: Message Status Ticks (Nice-to-Have)

| Option | Pros | Cons |
|---|---|---|
| **A: Add status ticks for WhatsApp only (Recommended)** | Native WhatsApp UX (users expect ticks), data comes free via status webhooks | Inconsistent with FB/IG inbox which don't show ticks, extra frontend work |
| B: Skip status ticks | Simpler, consistent with other platforms | Misses a WhatsApp-native feature users expect |

**Recommendation: Option A (Add ticks, WhatsApp only)**. WhatsApp users universally expect tick indicators. The data arrives automatically via status webhooks — we just need to store and display it. This is a separate, lower-priority story that can be cut if timeline is tight. Doesn't need to match FB/IG behavior since each platform has different native conventions.

Status indicators: Sent (single grey ✓), Delivered (double grey ✓✓), Read (double blue ✓✓).

---

## 4. Integration with Existing Features

| Existing Feature | Integration |
|---|---|
| **Inbox (Web)** | WhatsApp conversations appear alongside FB/IG/Twitter/LinkedIn conversations. Same chat UI, filtering, tagging, assignment. |
| **Inbox (iOS/Android)** | WhatsApp conversations in mobile inbox. Reply capability on mobile. |
| **Social Accounts (Web)** | WhatsApp in social accounts list with connect/edit/disconnect/reconnect. |
| **Social Accounts (Mobile)** | WhatsApp connection available in iOS/Android app settings. |
| **Onboarding** | WhatsApp as a connection option during new user onboarding wizard. |
| **Auto-Reply Rules** | WhatsApp added as a platform option in auto-reply rule creation. |
| **Team Collaboration** | WhatsApp conversations can be assigned to team members, tagged, and marked as resolved. |
| **Notifications** | New WhatsApp messages trigger the same notification system (in-app + browser push + mobile push). |

**Not integrated:** Composer (WhatsApp is inbox-only, no outbound publishing), Planner, Analytics.

---

## 5. Scope

- Connect WhatsApp Business account(s) via Embedded Signup (Facebook redirect) — **multiple accounts supported**
- Receive customer messages in Inbox (text, image, video, audio, document, location, sticker, voice note)
- Reply with text + single media attachment (image, video, document) within 24-hour window
- 24-hour window tracking — disable reply when expired
- Message status ticks (sent/delivered/read) — nice-to-have, separate story
- Platform filter for WhatsApp in Inbox
- Auto-reply rules for WhatsApp
- Account management: edit channel name, disconnect, reconnect on token expiry
- Token expiry detection + email notification + reconnect flow
- WhatsApp in onboarding connect accounts step
- iOS and Android: inbox view/reply + social account connection
- WhatsApp conversation assignment, tagging, read/unread

### Explicitly Out of Scope

- No conversation initiation (reply-only model)
- No message templates
- No interactive messages (buttons, lists) from inbox
- No WhatsApp Business Profile editing
- No WhatsApp-specific analytics
- No bulk messaging / broadcast
- No WhatsApp Commerce (product catalogs)
