# Epic + Stories: WhatsApp Inbox Integration

**Date:** 2026-03-08
**Epic:** WhatsApp Inbox Integration (existing epic #101850)

---

## Epic Description

WhatsApp Inbox Integration adds WhatsApp Business as a channel in ContentStudio's Inbox module across web, iOS, and Android. Users connect their WhatsApp Business account(s) via Meta's Embedded Signup, receive incoming customer messages in the unified Inbox, and reply with text and media — all within WhatsApp's 24-hour service window. This is a reply-only integration leveraging the WhatsApp Cloud API.

---

## Stories

---

### Story 1: [BE] Implement WhatsApp account connection via Meta Embedded Signup

**Group:** Backend
**Skill Set:** Backend
**Product Area:** Integrations
**Project:** Web App
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a social media manager, I want to connect my WhatsApp Business account to ContentStudio so that I can manage WhatsApp messages alongside my other social channels.

Build the backend OAuth flow for WhatsApp account connection using Meta's Embedded Signup. This follows the same redirect-based pattern used for Facebook and Instagram connections. Users are redirected to Facebook Business Login, complete the Embedded Signup (select/create WABA, add/verify phone number, grant permissions), and are redirected back to ContentStudio with an authorization code that is exchanged for an access token.

**Key implementation details:**
- Create `WhatsAppController` at `app/Http/Controllers/Integrations/Platforms/Social/WhatsAppController.php` following the pattern from `FacebookController`
- Configure Facebook Login for Business in Meta App Dashboard with Embedded Signup variation
- Request scopes: `whatsapp_business_management`, `whatsapp_business_messaging`
- Exchange auth code for System User access token (long-lived, no expiry)
- Store in `social_integrations` collection with `platform_type: 'whatsapp'` and fields: `whatsapp_business_account_id`, `phone_number_id`, `display_phone_number`, `display_name`, `channel_name`, `access_token` (encrypted via `SocialHelper::encryptToken()`)
- Register WhatsApp in `config/socialChannels.php` and `config/social_platforms.php`
- Support multiple WhatsApp accounts per workspace (different phone numbers)
- Validate that the same phone number isn't already connected in the workspace
- Subscribe to webhooks for the WABA after successful connection
- Add routes in `routes/web/integrations.php`

---

#### Workflow:
1. User clicks "Connect" next to WhatsApp in the social accounts connection modal
2. Backend generates the Facebook Business Login URL with WhatsApp Embedded Signup configuration and required scopes
3. User is redirected to Facebook where they complete the Embedded Signup flow
4. Facebook redirects back to ContentStudio's callback URL with an authorization code
5. Backend exchanges the code for an access token and retrieves the WABA ID and phone number ID
6. Backend stores the encrypted token and account details in `social_integrations`
7. Backend subscribes to webhooks for this WABA
8. Backend returns success with account details to the frontend

---

#### Acceptance criteria:
- [ ] `WhatsAppController` created with `redirectToProvider()` and `handleCallback()` methods
- [ ] Facebook Business Login URL is generated with correct scopes (`whatsapp_business_management`, `whatsapp_business_messaging`) and Embedded Signup configuration
- [ ] Authorization code is exchanged for a System User access token
- [ ] Access token is encrypted using `SocialHelper::encryptToken()` before storage
- [ ] Account record created in `social_integrations` with all required fields: `platform_type`, `whatsapp_business_account_id`, `phone_number_id`, `display_phone_number`, `display_name`, `channel_name`, `access_token`, `workspace_id`, `user_id`
- [ ] WhatsApp registered in `config/socialChannels.php` and `config/social_platforms.php`
- [ ] Multiple WhatsApp accounts can be connected per workspace (different phone numbers)
- [ ] Connecting the same phone number twice in the same workspace returns an error
- [ ] Webhook subscription created for the WABA after successful connection
- [ ] Routes added to `routes/web/integrations.php`
- [ ] Connection errors (cancelled login, missing permissions, already-used number) return appropriate error codes and messages

---

#### Mock-ups:
N/A — backend only

---

#### Impact on existing data:
- New records in `social_integrations` collection with `platform_type: 'whatsapp'`
- New fields on WhatsApp account records: `whatsapp_business_account_id`, `phone_number_id`, `display_phone_number`
- New entries in `config/socialChannels.php` and `config/social_platforms.php`

---

#### Impact on other products:
- Frontend needs the callback data to show the confirmation modal (depends on **[FE] Build WhatsApp account connection UI and channel name modal**)
- iOS and Android apps need the same connection endpoint (depends on **[iOS] Add WhatsApp account connection in iOS app** and **[Android] Add WhatsApp account connection in Android app**)

---

#### Dependencies:
- ContentStudio's Meta App must have `whatsapp_business_management` and `whatsapp_business_messaging` permissions approved in the App Dashboard
- Meta App must be configured with Facebook Login for Business + Embedded Signup variation

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

### Story 2: [FE] Build WhatsApp account connection UI and channel name modal

**Group:** Frontend
**Skill Set:** Frontend
**Product Area:** Integrations
**Project:** Web App
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a social media manager, I want to see WhatsApp as an option when connecting social accounts and give my WhatsApp channel a friendly name so that I can identify it easily in my inbox.

Add WhatsApp as a platform option in the existing social account connection modal and build the post-redirect confirmation modal where users name their WhatsApp channel.

---

#### Workflow:
1. User navigates to Settings → Social Accounts
2. User clicks "Connect a Social Account"
3. User sees "WhatsApp" in the platform list with a green WhatsApp icon alongside existing platforms
4. User clicks "Connect" next to WhatsApp
5. User is redirected to Facebook Business Login (handled by backend)
6. After completing Facebook setup, user is redirected back to ContentStudio
7. A "Choose Your Phone Number" modal appears with:
   - Phone number displayed (read-only, pre-filled)
   - "Channel Name" input field
   - "Connect" button
8. User enters a channel name (e.g., "Support Line") and clicks "Connect"
9. Success toast appears: "WhatsApp account connected successfully!"
10. WhatsApp account appears in the Social Accounts list with green WhatsApp icon, phone number, display name, and channel name
11. If user has multiple WhatsApp accounts, each is listed separately

---

#### UI Copy:

**Connection Modal — WhatsApp Entry:**
- Platform name: "WhatsApp"
- Icon: WhatsApp logo (green, `#25D366`)
- Button: "Connect"
- Tooltip on info icon (ℹ): "Connect your WhatsApp Business number to receive and reply to customer messages directly from your ContentStudio Inbox. You'll need a Facebook account and a phone number to get started."

**Post-Redirect Confirmation Modal:**
- Title: "Connect Your WhatsApp Number"
- Description: "Your WhatsApp Business number is ready to connect. Give it a name so you can easily identify it in your inbox."
- Phone number field:
  - Label: "Phone Number"
  - Value: Pre-filled, read-only (e.g., "+1 (555) 123-4567")
  - Helper text: "This is the number you verified with WhatsApp."
- Channel name field:
  - Label: "Channel Name"
  - Placeholder: "e.g., Support Line, Sales, Main Business"
  - Helper text: "This name will appear in your inbox and social accounts list. You can change it later."
  - Validation error (empty): "Please enter a name for this WhatsApp channel."
  - Validation error (too long, >50 chars): "Channel name must be 50 characters or less."
- Primary CTA: "Connect"
- Secondary CTA: "Cancel"

**Success Toast:**
- "WhatsApp account connected successfully!"

**Error Toasts:**
- Cancelled: "WhatsApp connection was cancelled."
- Missing permissions: "We need WhatsApp messaging permissions to connect your account. Please try again and grant the required permissions."
- Duplicate number: "This phone number is already connected to your workspace."

**Social Accounts List — WhatsApp Account Card:**
- Icon: WhatsApp logo (green)
- Line 1: Channel name (e.g., "Support Line")
- Line 2: Phone number + display name (e.g., "+1 (555) 123-4567 · Casper's Business")
- Status: "Connected" (green badge) or "Reconnection required" (red badge)
- Three-dot menu: Edit Channel Name, Reconnect (if token expired), Disconnect

---

#### Acceptance criteria:
- [ ] WhatsApp appears in the social account connection modal with correct icon and "Connect" button
- [ ] Clicking "Connect" triggers the backend redirect to Facebook Business Login
- [ ] Post-redirect confirmation modal displays the phone number (read-only) and channel name input
- [ ] Channel name is required — shows "Please enter a name for this WhatsApp channel." if empty
- [ ] Channel name is limited to 50 characters — shows validation error if exceeded
- [ ] Clicking "Connect" on the modal saves the channel name and shows success toast
- [ ] WhatsApp account card appears in Social Accounts list with icon, channel name, phone number, and display name
- [ ] Multiple WhatsApp accounts can be listed separately
- [ ] Error toasts display for cancelled login, missing permissions, and duplicate number scenarios
- [ ] All primary colors use `text-primary-cs-500`, `bg-primary-cs-50`, etc. — no hardcoded colors except WhatsApp green for the WhatsApp icon itself

---

#### Mock-ups:
See PRD section 7

---

#### Impact on existing data:
None — purely frontend

---

#### Impact on other products:
None

---

#### Dependencies:
- Depends on **[BE] Implement WhatsApp account connection via Meta Embedded Signup** for the redirect URL and callback handling

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

### Story 3: [BE] Build WhatsApp webhook receiver and message ingestion pipeline

**Group:** Backend
**Skill Set:** Backend
**Product Area:** Inbox
**Project:** Web App
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a social media manager, I want incoming WhatsApp messages from customers to appear in my ContentStudio Inbox so that I can see and manage them without switching to WhatsApp Business App.

Build the webhook endpoint that receives incoming WhatsApp messages from Meta, validates the signature, enqueues them for async processing, and stores them in the `inbox_details` collection.

**Key implementation details:**
- Create `WhatsAppWebhooksController` at `app/Http/Controllers/Emails/SocialWebhooks/WhatsAppWebhooksController.php` following the pattern from `InstagramWebhooksController`
- Public routes at `/whatsapp/webhook` (GET for verification, POST for events) in `routes/web.php`
- GET handler: validate `hub.verify_token` against stored token, return `hub.challenge`
- POST handler: validate `X-Hub-Signature-256` HMAC-SHA256 signature using app secret, then enqueue to Redis `dramatics:inbox:whatsapp:webhook`
- Create `WhatsAppWebhookExecuteJob` to process queued events
- Create `WhatsAppHelper` at `app/Libraries/Inbox/HelperClasses/WhatsAppHelper.php` for message processing
- Parse webhook payload: extract `phone_number_id` from metadata to identify the CS account, extract message content by type (text, image, video, audio, document, location, sticker, contacts, reaction)
- For media messages: download media from Meta's temporary URL using `GET /v21.0/{media_id}` → get download URL → fetch binary → store in ContentStudio's media storage (S3/equivalent)
- Create `inbox_details` record with `platform: 'whatsapp'`, `inbox_type: 'conversation'`, sender info from `contacts[].profile.name` and `wa_id`
- Message deduplication: check `wamid` in Redis (TTL ~2 hours) before processing
- Process status webhooks (sent, delivered, read) — update the corresponding outgoing message record
- Log all webhook events via `LogsBuilder`

---

#### Workflow:
1. Customer sends a message to the business's WhatsApp number
2. Meta sends a POST to `/whatsapp/webhook` with the message payload
3. Server validates `X-Hub-Signature-256` signature — rejects if invalid
4. Server checks message ID against Redis dedup cache — skips if already processed
5. Server enqueues the event to `dramatics:inbox:whatsapp:webhook` Redis queue
6. `WhatsAppWebhookExecuteJob` picks up the event, identifies the ContentStudio account by `phone_number_id`
7. Job parses the message type, downloads any media, and creates an `inbox_details` record
8. Message appears in the user's Inbox

---

#### Acceptance criteria:
- [ ] `WhatsAppWebhooksController` created with `webhookVerification()` (GET) and `handleWebhookEvent()` (POST) methods
- [ ] GET `/whatsapp/webhook` correctly validates verify token and returns challenge
- [ ] POST `/whatsapp/webhook` validates `X-Hub-Signature-256` HMAC-SHA256 signature and rejects invalid payloads with 401
- [ ] Valid webhook events are enqueued to `dramatics:inbox:whatsapp:webhook` Redis queue
- [ ] POST handler returns 200 immediately (before processing)
- [ ] `WhatsAppWebhookExecuteJob` processes queued events and creates `inbox_details` records
- [ ] Text messages stored with correct `inbox_details` fields (platform, workspace_id, platform_id, message content, sender name/phone, timestamp)
- [ ] Media messages (image, video, audio, document) — media downloaded from Meta URL and stored in ContentStudio media storage; inbox record includes media reference
- [ ] Location messages stored with latitude, longitude, name, address
- [ ] Sticker and voice note messages stored with media reference
- [ ] Contact messages stored with contact name and phone
- [ ] Message deduplication via Redis (`wamid` key with ~2 hour TTL) — duplicate webhook deliveries are skipped
- [ ] Status webhooks (sent, delivered, read) update the corresponding outgoing message record
- [ ] `WhatsAppHelper` created with methods for processing each message type
- [ ] All webhook events logged via `LogsBuilder`
- [ ] Routes added to `routes/web.php` (public, no auth required)

---

#### Mock-ups:
N/A — backend only

---

#### Impact on existing data:
- New records in `inbox_details` collection with `platform: 'whatsapp'`
- Media files stored in ContentStudio's media storage (S3/equivalent)
- Redis keys for message deduplication (`wamid:*` with TTL)

---

#### Impact on other products:
- Frontend inbox must support rendering WhatsApp messages (depends on **[FE] Add WhatsApp conversations to Inbox UI with platform filter and chat view**)
- Notification system should trigger on new WhatsApp messages (depends on **[BE] Extend notification system and auto-reply rules for WhatsApp**)

---

#### Dependencies:
- Depends on **[BE] Implement WhatsApp account connection via Meta Embedded Signup** — accounts must exist to map `phone_number_id` to workspace

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

### Story 4: [BE] Implement WhatsApp reply API with 24-hour window enforcement

**Group:** Backend
**Skill Set:** Backend
**Product Area:** Inbox
**Project:** Web App
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a social media manager, I want to reply to WhatsApp messages from ContentStudio so that I can respond to customers without leaving my unified inbox.

Build the reply endpoint that sends messages via the WhatsApp Cloud API, supporting text and single media attachment, with 24-hour service window enforcement.

**Key implementation details:**
- Add reply method to `WhatsAppHelper` at `app/Libraries/Inbox/HelperClasses/WhatsAppHelper.php`
- API call: `POST https://graph.facebook.com/v21.0/{phone_number_id}/messages` with bearer token
- Support message types: text (with `text.body`), image (with `image.link` + `image.caption`), video (with `video.link` + `video.caption`), document (with `document.link` + `document.caption` + `document.filename`)
- Use public CDN URL for media attachments (link method)
- Track last customer message timestamp per conversation in `inbox_details`
- Before sending: check if last customer message was within 24 hours. If expired, return error with `window_expired` code
- Validate media: file size (image ≤ 5MB, video ≤ 16MB, document ≤ 100MB), format (JPEG/PNG, MP4, PDF/DOC/DOCX/XLS/XLSX/PPT/PPTX/TXT)
- Store outgoing message in `inbox_details` with `direction: 'outgoing'` and initial status `sent`
- One attachment per message (WhatsApp limitation)
- Add inbox reply route for WhatsApp

---

#### Workflow:
1. User composes a reply in the inbox chat view (text and/or one attachment)
2. Frontend sends reply request to the backend WhatsApp reply endpoint
3. Backend validates the 24-hour window — if expired, returns `window_expired` error
4. Backend validates media attachment (if any) — checks size and format
5. Backend calls WhatsApp Cloud API to send the message
6. Backend stores the outgoing message in `inbox_details`
7. Backend returns success to frontend

---

#### Acceptance criteria:
- [ ] Reply endpoint accepts text message and sends via `POST /v21.0/{phone_number_id}/messages` with correct payload
- [ ] Reply endpoint accepts text + single media attachment (image, video, or document)
- [ ] Media sent using public URL link method (not upload)
- [ ] 24-hour window enforced: reply rejected with `window_expired` error code if last customer message > 24 hours ago
- [ ] Each customer reply resets the 24-hour window (timestamp updated on incoming message processing)
- [ ] Media validation: images ≤ 5MB, videos ≤ 16MB, documents ≤ 100MB — returns specific error if exceeded
- [ ] Format validation: rejects unsupported formats with descriptive error
- [ ] Only one attachment per message — returns error if multiple attachments sent
- [ ] Outgoing message stored in `inbox_details` with direction, content, timestamp, and initial status
- [ ] API errors from WhatsApp (rate limit, invalid token, etc.) are caught and returned with appropriate error codes
- [ ] Access token decrypted via `SocialHelper::decryptToken()` for API calls

---

#### Mock-ups:
N/A — backend only

---

#### Impact on existing data:
- New outgoing message records in `inbox_details`
- `last_customer_message_at` timestamp tracked per conversation

---

#### Impact on other products:
- Frontend reply composer depends on this endpoint (depends on **[FE] Add WhatsApp conversations to Inbox UI with platform filter and chat view**)

---

#### Dependencies:
- Depends on **[BE] Build WhatsApp webhook receiver and message ingestion pipeline** — conversations must exist to reply to
- Depends on **[BE] Implement WhatsApp account connection via Meta Embedded Signup** — access tokens must be stored

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

### Story 5: [FE] Add WhatsApp conversations to Inbox UI with platform filter and chat view

**Group:** Frontend
**Skill Set:** Frontend
**Product Area:** Inbox
**Project:** Web App
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a social media manager, I want to see and reply to WhatsApp conversations in my ContentStudio Inbox so that I can manage all my customer messages in one place.

Add WhatsApp support to the existing Inbox module — conversation list, chat view, reply composer, platform filter, and 24-hour window expired state.

---

#### Workflow:
1. User navigates to Inbox from the left sidebar
2. User sees WhatsApp conversations in the conversation list with:
   - Green WhatsApp icon as platform badge
   - Channel name (if multiple WhatsApp accounts, e.g., "Support Line")
   - Customer name (from WhatsApp profile) or phone number if unavailable
   - Message preview (truncated to ~60 characters)
   - Relative timestamp ("2m ago", "1h ago")
   - Unread indicator (bold text + blue dot) if not yet viewed
3. User clicks a WhatsApp conversation to open the chat view
4. Chat view shows the message thread — customer messages on the left, business replies on the right
5. Text messages display as text bubbles
6. Images display as thumbnails (clickable to expand in lightbox)
7. Videos display with a play button overlay (clickable to play)
8. Documents display as file cards with icon, filename, and size (clickable to download)
9. Audio/voice notes display as an inline audio player with play/pause and duration
10. Location messages display as a card with address text and a "View on Map" link (opens Google Maps)
11. Stickers display as inline images
12. User types a reply in the composer at the bottom
13. User optionally clicks the attachment icon to attach one file (image, video, or document)
14. Attachment preview shows above the composer with filename/thumbnail and an ✕ to remove
15. User clicks "Send" or presses Enter
16. Reply appears instantly in the thread with "Sending..." indicator
17. User can filter the conversation list by selecting "WhatsApp" in the platform filter dropdown
18. If multiple WhatsApp accounts connected, user can filter by specific account

**24-hour window expired:**
19. If the last customer message was > 24 hours ago, the reply composer area shows a warning banner instead of the input field
20. The "Send" button is disabled/hidden

---

#### UI Copy:

**Platform Filter:**
- Dropdown option: "WhatsApp" with WhatsApp icon

**Conversation List — WhatsApp Badge:**
- Platform icon: WhatsApp logo (green `#25D366` — this is the WhatsApp brand color, not a CS theme color)

**Chat View Header:**
- Customer name or phone number
- WhatsApp icon + channel name (e.g., "via Support Line")

**Reply Composer:**
- Placeholder: "Type your reply..."
- Send button label: "Send"
- Attachment icon tooltip: "Attach a file (one per message)"

**Attachment Preview:**
- Remove button: ✕ icon
- Tooltip on ✕: "Remove attachment"

**24-Hour Window Expired Banner:**
- Banner background: Light yellow/amber warning style (`bg-yellow-50 border-yellow-200`)
- Icon: Clock icon (Lucide `Clock`)
- Text: "The 24-hour reply window has expired. You can reply again once the customer sends a new message."
- Learn more link: "Learn more about WhatsApp's 24-hour messaging window" (links to help article)

**Error Toasts:**
- Send failed: "Message failed to send. Please try again."
- File too large (image): "This image is too large. Maximum size is 5 MB."
- File too large (video): "This video is too large. Maximum size is 16 MB."
- File too large (document): "This file is too large. Maximum size is 100 MB."
- Unsupported format: "This file format isn't supported on WhatsApp. Try JPEG, PNG, MP4, PDF, DOC, or XLSX."
- Multiple attachments: "WhatsApp allows one attachment per message. Send this one first, then attach another."

**Empty State (no WhatsApp conversations yet):**
- Headline: "No WhatsApp messages yet"
- Subtext: "When customers message your connected WhatsApp number, their conversations will appear here."
- No CTA button (messages come from customers, not initiated by the user)

**Loading State:**
- Skeleton loader for conversation list items (same as existing inbox loading)
- Spinner in chat view while messages load

**Error State (failed to load conversations):**
- Text: "Something went wrong loading your WhatsApp conversations. Please try again."
- CTA: "Retry"

---

#### Acceptance criteria:
- [ ] WhatsApp conversations appear in inbox conversation list with green WhatsApp icon badge
- [ ] Channel name shown on conversation list items when multiple WhatsApp accounts connected
- [ ] Customer name displayed from WhatsApp profile; phone number as fallback
- [ ] Message preview truncated to ~60 characters in conversation list
- [ ] Unread indicator (bold + blue dot) shown for unread conversations
- [ ] Chat view displays message thread with customer messages on left, business replies on right
- [ ] Text messages render as text bubbles with timestamps
- [ ] Images render as thumbnails, clickable to expand in lightbox
- [ ] Videos render with play button, clickable to play
- [ ] Documents render as file cards with icon, filename, clickable to download
- [ ] Audio/voice notes render as inline audio player with play/pause
- [ ] Location messages render as cards with address and "View on Map" link
- [ ] Stickers render as inline images
- [ ] Reply composer supports text input with "Send" button and Enter key
- [ ] Attachment button allows selecting one file; preview shown with ✕ to remove
- [ ] Only one attachment per message — second attachment shows toast error
- [ ] File size and format validated before upload — appropriate error toasts shown
- [ ] "WhatsApp" option available in platform filter dropdown
- [ ] Filter by specific WhatsApp account when multiple connected
- [ ] 24-hour expired banner shown when last customer message > 24 hours ago — composer disabled
- [ ] Empty state displayed when no WhatsApp conversations exist
- [ ] Loading skeleton shown while conversations load
- [ ] Error state with retry button shown on load failure
- [ ] All primary colors use theme classes (`text-primary-cs-500`, `bg-primary-cs-50`, etc.) except WhatsApp brand green for the icon

---

#### Mock-ups:
See PRD section 7

---

#### Impact on existing data:
None — frontend only

---

#### Impact on other products:
None

---

#### Dependencies:
- Depends on **[BE] Build WhatsApp webhook receiver and message ingestion pipeline** — for incoming message data
- Depends on **[BE] Implement WhatsApp reply API with 24-hour window enforcement** — for sending replies

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

### Story 6: [BE] Implement WhatsApp account management — edit, disconnect, reconnect, and token expiry handling

**Group:** Backend
**Skill Set:** Backend
**Product Area:** Integrations
**Project:** Web App
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a social media manager, I want to edit, disconnect, and reconnect my WhatsApp accounts so that I can manage my connections and recover from token expiry without contacting support.

Build backend endpoints for WhatsApp account management: update channel name, disconnect (remove webhook subscription + mark disconnected), reconnect (re-auth via Facebook Business Login + update token), and token expiry detection.

**Key implementation details:**
- Add endpoints to `WhatsAppController`: `updateChannelName()`, `disconnect()`, `reconnect()`
- Disconnect: unsubscribe from webhooks, mark account as disconnected in `social_integrations`, stop processing messages
- Reconnect: same redirect to Facebook Business Login, but re-authorizing the existing WABA. On callback, update the stored access token
- Token expiry detection: when WhatsApp API calls fail with auth errors (401/403), increment `invalid_tries` on the account. After threshold (e.g., 3 consecutive failures), mark account as `needs_reconnect` and trigger email notification using existing `sent_invalid_email` pattern
- API endpoint to check account health/status for frontend badge display

---

#### Workflow:
1. User requests channel name update → backend validates and updates `channel_name` in `social_integrations`
2. User requests disconnect → backend unsubscribes webhooks, marks account disconnected
3. Token becomes invalid (expired/revoked) → backend detects via failed API call → increments `invalid_tries` → after threshold, marks `needs_reconnect` and sends email notification
4. User clicks "Reconnect" → backend generates Facebook Business Login URL → user re-authorizes → callback updates token → clears `invalid_tries` and `needs_reconnect`

---

#### Acceptance criteria:
- [ ] PUT endpoint updates `channel_name` for a WhatsApp account — validates non-empty and ≤ 50 chars
- [ ] Disconnect endpoint unsubscribes webhooks for the WABA and marks account as disconnected
- [ ] Disconnected accounts stop receiving and processing webhook events
- [ ] Reconnect endpoint redirects to Facebook Business Login with same Embedded Signup configuration
- [ ] Reconnect callback updates the stored access token (encrypted) and clears `invalid_tries` and `needs_reconnect` flags
- [ ] Failed WhatsApp API calls (auth errors) increment `invalid_tries` on the account
- [ ] After 3 consecutive auth failures, account is marked `needs_reconnect: true`
- [ ] Email notification sent when account needs reconnection (reuses existing `sent_invalid_email` pattern)
- [ ] Health/status endpoint returns account status (connected, needs_reconnect, disconnected) for frontend
- [ ] All endpoints require workspace authentication and verify account ownership

---

#### Mock-ups:
N/A — backend only

---

#### Impact on existing data:
- Updates to existing `social_integrations` records: `channel_name`, `state`, `invalid_tries`, `sent_invalid_email`, `needs_reconnect`

---

#### Impact on other products:
- Frontend account management UI depends on these endpoints (depends on **[FE] Build WhatsApp account management UI — edit, disconnect, reconnect**)

---

#### Dependencies:
- Depends on **[BE] Implement WhatsApp account connection via Meta Embedded Signup** — accounts must exist

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

### Story 7: [FE] Build WhatsApp account management UI — edit, disconnect, reconnect

**Group:** Frontend
**Skill Set:** Frontend
**Product Area:** Integrations
**Project:** Web App
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a social media manager, I want to edit my WhatsApp channel name, disconnect accounts I no longer need, and reconnect accounts when the token expires so that I can manage my WhatsApp integration without contacting support.

Add account management actions to the WhatsApp account card in Social Accounts settings, including the token expiry warning badge and reconnect flow.

---

#### Workflow:
1. User navigates to Settings → Social Accounts
2. User sees their WhatsApp account(s) in the list
3. Connected accounts show a green "Connected" badge; accounts needing reconnection show a red "Reconnection required" badge
4. User clicks the three-dot menu (⋮) on a WhatsApp account card
5. Menu options:
   - "Edit Channel Name" — opens inline edit
   - "Reconnect" (visible only when token expired) — redirects to Facebook Business Login
   - "Disconnect" — opens confirmation modal
6. **Edit**: User edits the channel name inline and saves
7. **Disconnect**: Confirmation modal appears, user clicks "Disconnect", account is removed from active list
8. **Reconnect**: User is redirected to Facebook, re-authorizes, returns to ContentStudio, token refreshed, badge changes to "Connected"

**Inbox reconnect banner:**
9. When a WhatsApp account needs reconnection, conversations for that account show a banner in the chat view

---

#### UI Copy:

**Account Card Status Badges:**
- Connected: "Connected" (green badge, same style as other platforms)
- Needs reconnect: "Reconnection required" (red badge)

**Three-Dot Menu Options:**
- "Edit Channel Name"
- "Reconnect" (only when `needs_reconnect: true`)
- "Disconnect"

**Edit Channel Name:**
- Input field pre-filled with current name
- Save button: "Save"
- Cancel button: "Cancel"
- Validation error (empty): "Please enter a channel name."
- Validation error (>50 chars): "Channel name must be 50 characters or less."

**Disconnect Confirmation Modal:**
- Title: "Disconnect WhatsApp Account"
- Description: "Are you sure you want to disconnect **[channel name]** (+1 555-123-4567)? You'll stop receiving WhatsApp messages from this number in your Inbox."
- Primary CTA: "Disconnect" (red/destructive style)
- Secondary CTA: "Cancel"

**Disconnect Success Toast:**
- "WhatsApp account disconnected."

**Reconnect Success Toast:**
- "WhatsApp account reconnected successfully!"

**Inbox Reconnect Banner (in chat view for affected conversations):**
- Banner background: Light red warning (`bg-red-50 border-red-200`)
- Icon: AlertTriangle (Lucide)
- Text: "This WhatsApp account needs to be reconnected. Messages can't be sent or received until you reconnect."
- CTA link: "Go to Social Accounts →" (navigates to Settings → Social Accounts)

---

#### Acceptance criteria:
- [ ] WhatsApp account card shows "Connected" green badge or "Reconnection required" red badge based on account status
- [ ] Three-dot menu shows "Edit Channel Name" and "Disconnect" for connected accounts
- [ ] Three-dot menu shows "Reconnect" option when account needs reconnection
- [ ] Edit channel name: inline edit with save/cancel, validates non-empty and ≤ 50 chars
- [ ] Disconnect: confirmation modal with channel name and phone number, red "Disconnect" button
- [ ] Disconnect success removes account from active list and shows toast
- [ ] Reconnect: redirects to Facebook Business Login, on return updates badge to "Connected" and shows success toast
- [ ] Inbox chat view shows reconnect banner for conversations belonging to accounts that need reconnection
- [ ] Reconnect banner includes "Go to Social Accounts →" link
- [ ] All colors use theme classes — destructive actions use standard red/error styles

---

#### Mock-ups:
See PRD section 7

---

#### Impact on existing data:
None — frontend only

---

#### Impact on other products:
None

---

#### Dependencies:
- Depends on **[BE] Implement WhatsApp account management — edit, disconnect, reconnect, and token expiry handling**

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

### Story 8: [BE] Extend notification system and auto-reply rules for WhatsApp

**Group:** Backend
**Skill Set:** Backend
**Product Area:** Inbox
**Project:** Web App
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a social media manager, I want to receive notifications for new WhatsApp messages and set up auto-reply rules so that I never miss a message and common questions get answered automatically.

Extend the existing notification pipeline and auto-reply rules system to support WhatsApp as a platform.

**Key implementation details:**
- Notification: After a new WhatsApp message is stored in `inbox_details`, trigger the existing notification pipeline (in-app, browser push, mobile push, email) — same as Facebook/Instagram inbox messages. No new notification infrastructure needed, just register WhatsApp as a triggering platform.
- Auto-reply: Extend auto-reply rule creation/editing to accept `whatsapp` as a platform. When a new WhatsApp message arrives and matches a rule's keyword trigger, send the auto-reply text via the WhatsApp reply API. Auto-replies must respect the 24-hour window (should always be within it since the customer just messaged). Support selecting specific WhatsApp account(s) for a rule.

---

#### Workflow:
**Notifications:**
1. Customer sends a WhatsApp message
2. Webhook processes and stores the message
3. Notification pipeline triggers — sends in-app notification, browser push, mobile push, and/or email based on user's notification preferences
4. User sees the notification and navigates to Inbox

**Auto-Reply:**
1. User creates an auto-reply rule with platform "WhatsApp", keyword triggers, and a text response
2. Customer sends a message matching the trigger keywords
3. System auto-sends the reply text via WhatsApp Cloud API
4. Auto-reply appears in the conversation thread as an outgoing message

---

#### Acceptance criteria:
- [ ] New WhatsApp messages trigger in-app notifications to workspace members
- [ ] New WhatsApp messages trigger browser push notifications (if enabled by user)
- [ ] New WhatsApp messages trigger mobile push notifications to iOS/Android apps (if enabled)
- [ ] New WhatsApp messages trigger email notifications (based on existing email notification preferences)
- [ ] Notification content includes: customer name/phone, message preview, WhatsApp icon, channel name
- [ ] Auto-reply rules can be created with `whatsapp` as the platform
- [ ] Auto-reply rules support keyword-based triggers (same logic as existing platforms)
- [ ] Auto-reply rules support selecting specific WhatsApp account(s) when multiple connected
- [ ] Matched auto-replies are sent via WhatsApp Cloud API and stored as outgoing messages
- [ ] Auto-replies only fire within the 24-hour window (always true since customer just messaged, but validate as safety check)
- [ ] Auto-reply conflicts detected (same as existing conflict detection logic in `app/Services/Inbox/`)

---

#### Mock-ups:
N/A — backend only

---

#### Impact on existing data:
- New auto-reply rules in existing auto-reply collection with `platform: 'whatsapp'`
- Notification records in existing notification collections

---

#### Impact on other products:
- Frontend auto-reply UI needs to show WhatsApp option (depends on **[FE] Add WhatsApp to auto-reply rules UI**)
- Mobile apps receive push notifications for WhatsApp messages (same push infrastructure)

---

#### Dependencies:
- Depends on **[BE] Build WhatsApp webhook receiver and message ingestion pipeline** — messages must be ingested first
- Depends on **[BE] Implement WhatsApp reply API with 24-hour window enforcement** — for sending auto-replies

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

### Story 9: [FE] Add WhatsApp to auto-reply rules UI

**Group:** Frontend
**Skill Set:** Frontend
**Product Area:** Inbox
**Project:** Web App
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a social media manager, I want to create auto-reply rules for WhatsApp in ContentStudio so that common customer questions get answered automatically even when I'm away.

Add WhatsApp as a platform option in the existing auto-reply rule creation/editing modal.

---

#### Workflow:
1. User navigates to Inbox → Auto-Reply Rules
2. User clicks "Create Rule"
3. In the platform selection step, user sees "WhatsApp" alongside existing platforms
4. User selects WhatsApp
5. User selects which WhatsApp account(s) the rule applies to (dropdown showing channel names, relevant when multiple accounts connected)
6. User configures keyword triggers (same as existing)
7. User writes the auto-reply text response
8. User optionally sets active hours
9. User saves the rule
10. Rule appears in the auto-reply rules list with WhatsApp icon badge

---

#### UI Copy:

**Platform Selection:**
- Option: "WhatsApp" with WhatsApp icon

**Account Selection (visible when WhatsApp selected and multiple accounts exist):**
- Label: "WhatsApp Account"
- Placeholder: "Select an account"
- Helper text: "Choose which WhatsApp number this rule applies to. Select 'All accounts' to apply to all your WhatsApp numbers."
- Options: "All accounts" + list of channel names with phone numbers

**Response Field:**
- Label: "Auto-Reply Message"
- Placeholder: "e.g., Thanks for reaching out! We'll get back to you shortly."
- Helper text: "This text will be sent automatically when a customer message matches your trigger keywords. Keep it short and helpful."
- Character limit note: "WhatsApp messages can be up to 4,096 characters."

---

#### Acceptance criteria:
- [ ] "WhatsApp" appears as a platform option in auto-reply rule creation/editing
- [ ] WhatsApp icon displayed next to the platform option
- [ ] Account selector appears when WhatsApp is selected and multiple accounts are connected
- [ ] "All accounts" option available in account selector
- [ ] Keyword triggers work the same as existing platforms
- [ ] Text-only response field (no media attachments for auto-replies)
- [ ] Active hours scheduling works the same as existing platforms
- [ ] Created rules appear in the rules list with WhatsApp icon badge
- [ ] Rules can be edited and deleted (same as existing)

---

#### Mock-ups:
See PRD section 7

---

#### Impact on existing data:
None — frontend only

---

#### Impact on other products:
None

---

#### Dependencies:
- Depends on **[BE] Extend notification system and auto-reply rules for WhatsApp**

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

### Story 10: [FE] Add WhatsApp to onboarding social account connection step

**Group:** Frontend
**Skill Set:** Frontend
**Product Area:** Onboarding
**Project:** Web App
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a new user, I want to connect my WhatsApp Business account during onboarding so that I'm set up from day one without having to find the setting later.

Add WhatsApp as a connection option in the onboarding wizard's "Connect Social Accounts" step.

---

#### Workflow:
1. New user reaches the "Connect Social Accounts" step in the onboarding wizard
2. User sees WhatsApp listed alongside other platforms (Facebook, Instagram, Twitter, etc.)
3. User clicks "Connect" next to WhatsApp
4. User is redirected to Facebook Business Login for Embedded Signup (same flow as Settings → Social Accounts)
5. After completing the Facebook flow, user is redirected back to the onboarding wizard
6. The channel name confirmation modal appears (same as Story 2)
7. User enters a channel name and clicks "Connect"
8. WhatsApp shows as "Connected" with a green checkmark in the onboarding step
9. User can connect additional WhatsApp numbers or proceed to the next onboarding step

---

#### UI Copy:

**Onboarding Platform Entry:**
- Platform name: "WhatsApp"
- Icon: WhatsApp logo (green)
- Button: "Connect"
- Connected state: Green checkmark + channel name + phone number

**Tooltip on info icon (ℹ):**
- "Connect your WhatsApp Business number to manage customer messages in your ContentStudio Inbox."

---

#### Acceptance criteria:
- [ ] WhatsApp appears in the onboarding "Connect Social Accounts" step alongside other platforms
- [ ] "Connect" button triggers the same Facebook Business Login redirect as the Settings flow
- [ ] After redirect callback, channel name modal appears within the onboarding wizard context
- [ ] On successful connection, WhatsApp shows as "Connected" with checkmark, channel name, and phone number
- [ ] User can connect multiple WhatsApp numbers during onboarding
- [ ] User can skip WhatsApp and proceed to next onboarding step
- [ ] Redirect back from Facebook returns to onboarding wizard (not Settings page)

---

#### Mock-ups:
See PRD section 7

---

#### Impact on existing data:
None — frontend only

---

#### Impact on other products:
None

---

#### Dependencies:
- Depends on **[BE] Implement WhatsApp account connection via Meta Embedded Signup**
- Depends on **[FE] Build WhatsApp account connection UI and channel name modal** (reuses the channel name modal component)

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

### Story 11: [iOS] Add WhatsApp inbox and account connection to iOS app

**Group:** Frontend
**Skill Set:** Frontend
**Product Area:** iOS Mobile
**Project:** Mobile
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a social media manager using the ContentStudio iOS app, I want to view and reply to WhatsApp messages and connect WhatsApp accounts from my iPhone so that I can manage customer conversations on the go.

Add WhatsApp support to the iOS app's Inbox module and Social Account connection settings.

---

#### Workflow:

**Inbox:**
1. User opens the ContentStudio iOS app and navigates to Inbox
2. WhatsApp conversations appear in the conversation list with green WhatsApp icon badge
3. User taps a conversation to view the message thread
4. Messages display in chat view — customer on left, business replies on right
5. Text, images (thumbnails), videos (play button), documents (file cards), audio (player), and location (card) messages all render appropriately
6. User types a reply and optionally attaches one file from the device
7. User taps "Send"
8. If 24-hour window expired, composer is disabled with warning message

**Account Connection:**
9. User navigates to Settings → Social Accounts in the iOS app
10. User taps "Connect" next to WhatsApp
11. System browser opens with Facebook Business Login
12. User completes Embedded Signup flow
13. On completion, redirect back to the app
14. Channel name input screen appears
15. User enters channel name and taps "Connect"
16. WhatsApp account appears in the list

**Account Management:**
17. User can edit channel name, disconnect, and reconnect (when token expired) from the iOS account management screen

---

#### Acceptance criteria:
- [ ] WhatsApp conversations appear in iOS inbox conversation list with WhatsApp icon
- [ ] Chat view renders all message types: text, image, video, document, audio, location, sticker
- [ ] Reply composer supports text + single attachment
- [ ] 24-hour expired state disables composer with warning message
- [ ] Push notifications received for new WhatsApp messages
- [ ] WhatsApp appears in iOS social account connection screen
- [ ] Facebook Business Login opens in system browser for account connection
- [ ] Redirect back to app works after Facebook flow completion
- [ ] Channel name input screen shown after successful Facebook auth
- [ ] Account management: edit name, disconnect, reconnect available
- [ ] WhatsApp platform filter available in inbox
- [ ] WhatsApp available in iOS onboarding social account connection step

---

#### Mock-ups:
See PRD section 7 — adapted for iOS native UI patterns

---

#### Impact on existing data:
None — consumes existing backend APIs

---

#### Impact on other products:
None

---

#### Dependencies:
- Depends on **[BE] Implement WhatsApp account connection via Meta Embedded Signup**
- Depends on **[BE] Build WhatsApp webhook receiver and message ingestion pipeline**
- Depends on **[BE] Implement WhatsApp reply API with 24-hour window enforcement**
- Depends on **[BE] Implement WhatsApp account management — edit, disconnect, reconnect, and token expiry handling**

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, native iOS app
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

### Story 12: [Android] Add WhatsApp inbox and account connection to Android app

**Group:** Frontend
**Skill Set:** Frontend
**Product Area:** Android Mobile
**Project:** Mobile
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a social media manager using the ContentStudio Android app, I want to view and reply to WhatsApp messages and connect WhatsApp accounts from my Android phone so that I can manage customer conversations on the go.

Add WhatsApp support to the Android app's Inbox module and Social Account connection settings.

---

#### Workflow:

**Inbox:**
1. User opens the ContentStudio Android app and navigates to Inbox
2. WhatsApp conversations appear in the conversation list with green WhatsApp icon badge
3. User taps a conversation to view the message thread
4. Messages display in chat view — customer on left, business replies on right
5. Text, images (thumbnails), videos (play button), documents (file cards), audio (player), and location (card) messages all render appropriately
6. User types a reply and optionally attaches one file from the device
7. User taps "Send"
8. If 24-hour window expired, composer is disabled with warning message

**Account Connection:**
9. User navigates to Settings → Social Accounts in the Android app
10. User taps "Connect" next to WhatsApp
11. System browser opens with Facebook Business Login
12. User completes Embedded Signup flow
13. On completion, redirect back to the app
14. Channel name input screen appears
15. User enters channel name and taps "Connect"
16. WhatsApp account appears in the list

**Account Management:**
17. User can edit channel name, disconnect, and reconnect (when token expired) from the Android account management screen

---

#### Acceptance criteria:
- [ ] WhatsApp conversations appear in Android inbox conversation list with WhatsApp icon
- [ ] Chat view renders all message types: text, image, video, document, audio, location, sticker
- [ ] Reply composer supports text + single attachment
- [ ] 24-hour expired state disables composer with warning message
- [ ] Push notifications received for new WhatsApp messages
- [ ] WhatsApp appears in Android social account connection screen
- [ ] Facebook Business Login opens in system browser for account connection
- [ ] Redirect back to app works after Facebook flow completion (deep link / intent filter)
- [ ] Channel name input screen shown after successful Facebook auth
- [ ] Account management: edit name, disconnect, reconnect available
- [ ] WhatsApp platform filter available in inbox
- [ ] WhatsApp available in Android onboarding social account connection step

---

#### Mock-ups:
See PRD section 7 — adapted for Android native UI patterns (Material Design)

---

#### Impact on existing data:
None — consumes existing backend APIs

---

#### Impact on other products:
None

---

#### Dependencies:
- Depends on **[BE] Implement WhatsApp account connection via Meta Embedded Signup**
- Depends on **[BE] Build WhatsApp webhook receiver and message ingestion pipeline**
- Depends on **[BE] Implement WhatsApp reply API with 24-hour window enforcement**
- Depends on **[BE] Implement WhatsApp account management — edit, disconnect, reconnect, and token expiry handling**

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, native Android app
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

### Story 13: [BE] Add WhatsApp message status tracking (sent, delivered, read)

**Group:** Backend
**Skill Set:** Backend
**Product Area:** Inbox
**Project:** Web App
**Priority:** Low (P2 — Nice to Have)
**Type:** Feature

---

#### Description:
As a social media manager, I want to see whether my WhatsApp replies have been sent, delivered, and read so that I know if the customer has seen my response.

Process WhatsApp status webhooks and store message delivery status on outgoing messages. WhatsApp sends status updates (sent → delivered → read) via the same webhook endpoint.

**Key implementation details:**
- In `WhatsAppWebhookExecuteJob`, handle `statuses` array from webhook payload
- Status types: `sent`, `delivered`, `read`, `failed`
- Map status webhook `id` to the outgoing message record in `inbox_details`
- Update a `whatsapp_status` field on the message record: `sent` | `delivered` | `read` | `failed`
- For `failed` status, store the error code and message from the webhook payload
- Expose status in the inbox API response for frontend to display

---

#### Workflow:
1. Agent sends a reply from ContentStudio
2. WhatsApp Cloud API sends a `sent` status webhook → message status updated to `sent`
3. Message is delivered to customer's device → `delivered` status webhook → status updated
4. Customer opens and reads the message → `read` status webhook → status updated
5. Frontend displays the appropriate tick indicator

---

#### Acceptance criteria:
- [ ] `sent` status webhook updates outgoing message to `whatsapp_status: 'sent'`
- [ ] `delivered` status webhook updates outgoing message to `whatsapp_status: 'delivered'`
- [ ] `read` status webhook updates outgoing message to `whatsapp_status: 'read'`
- [ ] `failed` status webhook updates message to `whatsapp_status: 'failed'` with error details
- [ ] Status is only updated forward (sent → delivered → read), never backward
- [ ] Status field exposed in inbox API response for outgoing WhatsApp messages
- [ ] Status webhooks for messages not found in our system are gracefully ignored (no errors)

---

#### Mock-ups:
N/A — backend only

---

#### Impact on existing data:
- New `whatsapp_status` field on outgoing WhatsApp message records in `inbox_details`

---

#### Impact on other products:
- Frontend needs to display status indicators (depends on **[FE] Display WhatsApp message status ticks in chat view**)

---

#### Dependencies:
- Depends on **[BE] Build WhatsApp webhook receiver and message ingestion pipeline** — status webhooks arrive on same endpoint
- Depends on **[BE] Implement WhatsApp reply API with 24-hour window enforcement** — outgoing messages must exist to update

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend only
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend only
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

### Story 14: [FE] Display WhatsApp message status ticks in chat view

**Group:** Frontend
**Skill Set:** Frontend
**Product Area:** Inbox
**Project:** Web App
**Priority:** Low (P2 — Nice to Have)
**Type:** Feature

---

#### Description:
As a social media manager, I want to see delivery status ticks on my WhatsApp replies so that I know whether the customer has received and read my message.

Display WhatsApp-native status indicators on outgoing messages in the inbox chat view. This is a WhatsApp-only feature — other platforms don't show these ticks.

---

#### Workflow:
1. User views a WhatsApp conversation in the Inbox chat view
2. Outgoing messages show a status indicator next to the timestamp:
   - **Sending**: Clock icon (grey) while the message is being sent
   - **Sent**: Single check ✓ (grey)
   - **Delivered**: Double check ✓✓ (grey)
   - **Read**: Double check ✓✓ (blue, `text-primary-cs-500`)
   - **Failed**: Red exclamation mark with "Failed to send. Tap to retry."
3. Status updates in real-time as webhook data arrives (or on next inbox refresh)

---

#### UI Copy:

**Status Indicators (shown next to timestamp on outgoing messages):**
- Sending: Clock icon (Lucide `Clock`, grey `text-gray-400`)
- Sent: Single check icon (Lucide `Check`, grey `text-gray-400`)
- Delivered: Double check icon (Lucide `CheckCheck`, grey `text-gray-400`)
- Read: Double check icon (Lucide `CheckCheck`, blue `text-primary-cs-500`)
- Failed: Alert icon (Lucide `AlertCircle`, red `text-red-500`) + "Failed to send. Tap to retry."

**Tooltip on status icons:**
- Sending: "Sending..."
- Sent: "Sent"
- Delivered: "Delivered"
- Read: "Read"
- Failed: "Failed to send"

---

#### Acceptance criteria:
- [ ] Outgoing WhatsApp messages show status icon next to timestamp
- [ ] Clock icon shown while message is sending (before API response)
- [ ] Single grey check shown for `sent` status
- [ ] Double grey checks shown for `delivered` status
- [ ] Double blue checks shown for `read` status (using `text-primary-cs-500`)
- [ ] Red alert icon + "Failed to send. Tap to retry." shown for `failed` status
- [ ] Clicking "Tap to retry" on failed messages retries the send
- [ ] Tooltips shown on hover for each status icon
- [ ] Status ticks only appear on WhatsApp outgoing messages — not on other platforms
- [ ] Status updates reflect on next inbox data refresh (real-time via polling or websocket if available)
- [ ] Mobile apps (iOS/Android) also display status ticks on WhatsApp messages

---

#### Mock-ups:
See PRD section 7

---

#### Impact on existing data:
None — frontend only

---

#### Impact on other products:
- iOS and Android apps should also display these status ticks

---

#### Dependencies:
- Depends on **[BE] Add WhatsApp message status tracking (sent, delivered, read)**
- Depends on **[FE] Add WhatsApp conversations to Inbox UI with platform filter and chat view**

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

### Story 15: [Design] Create WhatsApp integration design assets and UI specifications

**Group:** Design
**Skill Set:** Design
**Product Area:** Throughout Product
**Project:** Web App
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a product team, we need design specifications and assets for the WhatsApp integration so that frontend and mobile developers have clear visual guidance for implementation.

Create Figma designs covering all WhatsApp integration touchpoints across web, iOS, and Android.

---

#### Workflow:
1. Designer reviews the PRD and workflow design document
2. Designer creates Figma frames for each screen/component:
   - Social Accounts: WhatsApp connection card, post-redirect confirmation modal, account management (edit/disconnect/reconnect), token expiry warning badge
   - Inbox: WhatsApp conversation list items, chat view with all message types (text, image, video, document, audio, location, sticker), reply composer with attachment, 24-hour expired banner, reconnect banner, empty state, loading state, error state
   - Auto-Reply: WhatsApp platform option in rule creation, account selector
   - Onboarding: WhatsApp in connect accounts step
   - Status ticks: sent/delivered/read indicators (nice-to-have)
   - Mobile: iOS and Android adaptations of all above screens
3. Designer hands off designs with annotations, spacing, and component specifications

---

#### Acceptance criteria:
- [ ] Figma designs created for WhatsApp social account connection modal and confirmation modal
- [ ] Figma designs for WhatsApp account card in Social Accounts (connected, reconnection required states)
- [ ] Figma designs for disconnect confirmation modal
- [ ] Figma designs for WhatsApp conversation list items in Inbox (unread, read states)
- [ ] Figma designs for WhatsApp chat view with all message types
- [ ] Figma designs for reply composer with attachment preview
- [ ] Figma designs for 24-hour expired banner
- [ ] Figma designs for reconnect warning banner in chat view
- [ ] Figma designs for empty state, loading state, and error state
- [ ] Figma designs for auto-reply rule creation with WhatsApp platform and account selector
- [ ] Figma designs for WhatsApp in onboarding connect accounts step
- [ ] Figma designs for message status ticks (sent, delivered, read, failed)
- [ ] iOS and Android adaptations of all screens
- [ ] WhatsApp brand green (`#25D366`) used only for the WhatsApp icon; all other UI colors follow CS theming
- [ ] All designs use existing CS design library components where possible

---

#### Mock-ups:
N/A — this story produces the mock-ups

---

#### Impact on existing data:
None

---

#### Impact on other products:
All frontend and mobile stories depend on these designs for visual guidance

---

#### Dependencies:
- PRD and workflow design documents must be approved (already done)

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support — N/A for design
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

### Story 16: [FE] Redesign WhatsApp post-connection modal to show all phone numbers in a single view

**Group:** Frontend
**Skill Set:** Frontend
**Product Area:** Integrations
**Project:** Web App
**Priority:** High (P0)
**Type:** Feature

---

#### Description:
As a social media manager connecting my WhatsApp Business account, I want to see all my phone numbers at once after the Meta redirect so that I can name the ones I want to connect without stepping through a confusing wizard.

This story replaces the current sequential wizard approach in the post-redirect modal (from **[FE] Build WhatsApp account connection UI and channel name modal**) with a single-view list layout. The wizard shows one phone number at a time with Skip/Continue buttons, which is a poor experience when Meta returns multiple numbers — the user has no visibility into how many numbers remain, can't go back, can't jump to a specific number, and must step through all numbers even if they only care about the last one.

The new design shows all returned phone numbers in one scrollable list. The user types a channel name for each number they want to connect and leaves the rest blank. No checkboxes, no wizard steps — the name input itself is the selection mechanism. If a name is filled in, that number gets connected. Empty = skipped.

**Reference patterns:**
- `contentstudio-frontend/src/modules/integration/components/dialogs/SaveSocialAccounts.vue` — multi-account list layout
- `contentstudio-frontend/src/modules/integration/components/dialogs/AddBluesky.vue` — name input with validation

**Components to use:**
- `Modal` from `@contentstudio/ui` — modal container
- `TextInput` from `@contentstudio/ui` — channel name input per row
- `Button` from `@contentstudio/ui` — primary and secondary CTAs
- `Icon` from `@contentstudio/ui` — WhatsApp icon, green checkmark indicator
- `Alert` from `@contentstudio/ui` — inline info message at top (if needed)

---

#### Workflow:
1. User completes Meta's Embedded Signup and is redirected back to ContentStudio
2. The connection modal opens showing all phone numbers returned from Meta in a single scrollable list
3. Each row displays the phone number (with country code) on the left and a channel name input on the right
4. Empty inputs show a placeholder suggesting a name (e.g., "e.g., Support Line")
5. User types a channel name into the input for any number they want to connect
6. As soon as a name is entered, a green checkmark appears on that row confirming it will be connected
7. If the user clears a previously entered name, the checkmark disappears — that number will be skipped
8. The primary CTA at the bottom updates dynamically to reflect how many numbers will be connected (e.g., "Connect 3 numbers")
9. User clicks the primary CTA
10. All named numbers are connected; unnamed numbers are skipped
11. Success toast confirms how many accounts were connected
12. If only one phone number was returned from Meta, the modal shows just that single row — same layout, no wizard difference

---

#### UI Copy:

**Modal Header:**
- Title: "Connect Your WhatsApp Numbers"
- Description (single number): "Name your WhatsApp number to connect it to your inbox."
- Description (multiple numbers): "Name the numbers you'd like to connect to your inbox. Leave a name blank to skip that number."

**Phone Number Row:**
- Phone number: Displayed with country code, formatted (e.g., "+1 (555) 123-4567") — read-only, `text-gray-900`, `text-sm`
- WhatsApp icon: Small green WhatsApp logo to the left of the number
- Green checkmark: Appears to the right of the input when a name is entered — `Icon` component, `text-green-500`

**Channel Name Input (per row):**
- Placeholder: "e.g., Support Line"
- Validation error (too long): "Channel name must be 50 characters or less."

**Primary CTA (dynamic):**
- No names entered (disabled): "Connect"
- 1 name entered: "Connect 1 number"
- Multiple names entered: "Connect 3 numbers" (dynamic count)

**Secondary CTA:**
- "Cancel"

**Success Toasts:**
- Single: "WhatsApp account connected successfully!"
- Multiple: "3 WhatsApp accounts connected successfully!" (dynamic count)

**Error Toasts:**
- Connection failed: "Something went wrong connecting your WhatsApp accounts. Please try again."
- Partial failure: "2 of 3 WhatsApp accounts connected. Some accounts failed — please try reconnecting them from Social Accounts."

**Loading state (while connecting):**
- Primary CTA shows `Loader` component inline: "Connecting..." (disabled during request)

---

#### Layout:

```
┌─────────────────────────────────────────────────────────────┐
│  ✕                                                          │
│                                                             │
│  Connect Your WhatsApp Numbers                              │
│  Name the numbers you'd like to connect to your inbox.      │
│  Leave a name blank to skip that number.                    │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  🟢 +1 (555) 123-4567      [Support Line____] ✓   │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │  🟢 +1 (555) 234-5678      [Sales___________] ✓   │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │  🟢 +44 20 7946 0958       [e.g., Support Li]     │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │  🟢 +92 300 123 4567       [e.g., Support Li]     │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │  🟢 +1 (555) 345-6789      [Main Business___] ✓   │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│                     [Cancel]    [Connect 3 numbers]          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

Each row: small WhatsApp icon (green) + phone number with country code (read-only, `text-gray-900`) on the left, `TextInput` for channel name + green checkmark icon (`text-green-500`, visible only when input has content) on the right.

CTA bar pinned to bottom if list scrolls.

---

#### Acceptance criteria:
- [ ] After Meta redirect, the modal opens showing all returned phone numbers in a single scrollable list — no wizard/stepper
- [ ] Each row shows the phone number with country code (read-only) on the left and a `TextInput` for the channel name on the right
- [ ] Empty inputs display placeholder text "e.g., Support Line"
- [ ] When a channel name is typed into an input, a green checkmark icon appears on that row
- [ ] When a channel name is cleared, the checkmark disappears
- [ ] The primary CTA text updates dynamically: "Connect 1 number", "Connect 2 numbers", "Connect 3 numbers", etc.
- [ ] The primary CTA is disabled when no channel names are entered (zero numbers selected)
- [ ] Clicking the primary CTA connects only the numbers that have a channel name filled in
- [ ] Numbers with empty channel name inputs are skipped — not connected
- [ ] Channel name validation: max 50 characters — shows inline error "Channel name must be 50 characters or less." if exceeded
- [ ] If the same phone number is already connected in the workspace, that row shows "This number is already connected to your workspace." with the input disabled
- [ ] Success toast shows dynamic count: "3 WhatsApp accounts connected successfully!"
- [ ] The modal is scrollable if more than 4-5 numbers are returned, with the CTA bar pinned to the bottom
- [ ] Single phone number scenario: same layout, just one row — no special UX
- [ ] Cancel button closes the modal without connecting anything
- [ ] Loading state: CTA shows "Connecting..." with inline `Loader` and is disabled during the request
- [ ] All colors use theme-aware classes (`text-primary-cs-500`, `bg-primary-cs-50`, etc.) — no hardcoded colors except WhatsApp brand green for the icon
- [ ] All user-facing strings use i18n keys — add to all locale directories

---

#### Mock-ups:
N/A — follow the layout above. Reference `SaveSocialAccounts.vue` for the multi-row list pattern inside a modal.

---

#### Impact on existing data:
No data or schema changes. This is a frontend-only redesign of the post-redirect modal UX.

---

#### Impact on other products:
- Web App: WhatsApp account connection flow only (Settings → Social Accounts and Onboarding)
- Mobile apps: iOS and Android use system browser for Meta Embedded Signup and have their own native post-redirect screens — this change does not affect them
- Chrome extension: No impact

---

#### Dependencies:
- Depends on **[BE] Implement WhatsApp account connection via Meta Embedded Signup** for the callback data (list of phone numbers, WABA details, access token)
- Updates the modal originally built in **[FE] Build WhatsApp account connection UI and channel name modal**

---

#### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness (modal should remain usable on smaller viewport widths — stack phone number above input if needed on narrow screens)
- [ ] Multilingual support (all modal copy uses i18n — add keys to all locale directories under `src/locales/`)
- [ ] UI theming support (use `@contentstudio/ui` components via props — no Tailwind color overrides except WhatsApp brand green for the icon)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

## Story Summary

| # | Story | Group | Priority | Project |
|---|---|---|---|---|
| 1 | [BE] Implement WhatsApp account connection via Meta Embedded Signup | Backend | P0 High | Web App |
| 2 | [FE] Build WhatsApp account connection UI and channel name modal | Frontend | P0 High | Web App |
| 3 | [BE] Build WhatsApp webhook receiver and message ingestion pipeline | Backend | P0 High | Web App |
| 4 | [BE] Implement WhatsApp reply API with 24-hour window enforcement | Backend | P0 High | Web App |
| 5 | [FE] Add WhatsApp conversations to Inbox UI with platform filter and chat view | Frontend | P0 High | Web App |
| 6 | [BE] Implement WhatsApp account management — edit, disconnect, reconnect, and token expiry handling | Backend | P0 High | Web App |
| 7 | [FE] Build WhatsApp account management UI — edit, disconnect, reconnect | Frontend | P0 High | Web App |
| 8 | [BE] Extend notification system and auto-reply rules for WhatsApp | Backend | P0 High | Web App |
| 9 | [FE] Add WhatsApp to auto-reply rules UI | Frontend | P0 High | Web App |
| 10 | [FE] Add WhatsApp to onboarding social account connection step | Frontend | P0 High | Web App |
| 11 | [iOS] Add WhatsApp inbox and account connection to iOS app | iOS | P0 High | Mobile |
| 12 | [Android] Add WhatsApp inbox and account connection to Android app | Android | P0 High | Mobile |
| 13 | [BE] Add WhatsApp message status tracking (sent, delivered, read) | Backend | P2 Low | Web App |
| 14 | [FE] Display WhatsApp message status ticks in chat view | Frontend | P2 Low | Web App |
| 15 | [Design] Create WhatsApp integration design assets and UI specifications | Design | P0 High | Web App |
| 16 | [FE] Redesign WhatsApp post-connection modal to show all phone numbers in a single view | Frontend | P0 High | Web App |

**Total: 16 stories** (14 P0 High, 2 P2 Low nice-to-have)
