# Research: WhatsApp Inbox Integration

**Date:** 2026-03-08
**Feature:** WhatsApp Business Inbox Integration for ContentStudio

---

## 1. What is WhatsApp Business Integration?

WhatsApp Business Platform allows companies to receive and respond to customer messages on WhatsApp through the Cloud API hosted by Meta. For a social media management tool like ContentStudio, this means adding WhatsApp as a channel in the Inbox module — users connect their WhatsApp Business account, receive incoming customer messages, and reply directly from the ContentStudio Inbox.

ContentStudio's integration will be **reply-only** — we won't allow users to initiate conversations. Customers message the business first, opening a 24-hour service window during which the business can reply freely.

---

## 2. WhatsApp Cloud API Architecture

### Key Concepts

| Concept | Description |
|---|---|
| **WhatsApp Business Account (WABA)** | Top-level container for a business's WhatsApp presence. Sits under a Meta Business Portfolio. |
| **Phone Number ID** | Unique ID for each registered WhatsApp phone number. Used in all API calls to send messages. |
| **Business Phone Number** | The actual phone number customers message. Must be verified and registered with Cloud API. |
| **Cloud API** | Meta-hosted REST API — no self-hosted infrastructure needed. Uses Graph API for sending, Webhooks for receiving. |
| **System User Token** | Long-lived token generated via Business Manager for production API access. |

### Message Flow

```
Customer sends message to business WhatsApp number
  → Meta servers receive it
  → Webhook POST to ContentStudio's endpoint (JSON payload)
  → ContentStudio processes and stores in inbox_details
  → Agent replies from ContentStudio Inbox
  → ContentStudio calls POST /v21.0/{phone_number_id}/messages
  → Meta delivers reply to customer on WhatsApp
```

### API Base URL

```
https://graph.facebook.com/v21.0/{phone_number_id}/messages
```

Authentication: `Authorization: Bearer {ACCESS_TOKEN}` header.

---

## 3. Account Connection Flow (Embedded Signup)

WhatsApp uses **Embedded Signup** — a Meta-provided flow that lets platforms (like ContentStudio) onboard business customers to WhatsApp Business Platform directly from our UI.

### How It Works

1. **ContentStudio registers as a Tech Provider / Solution Partner** with Meta
2. **Configure Facebook Login for Business** in Meta App Dashboard:
   - Create a "Facebook Login for Business" configuration
   - Select the "Embedded Signup" sign-in variation
   - Request permissions: `whatsapp_business_management`, `whatsapp_business_messaging`
3. **User clicks "Connect WhatsApp" in ContentStudio** → redirected to Facebook Business Login
4. **User goes through Meta's Embedded Signup flow:**
   - Logs into Facebook
   - Creates or selects a Meta Business Portfolio
   - Creates or selects a WhatsApp Business Account (WABA)
   - Adds and verifies a phone number (text/call verification)
   - Sets a WhatsApp Business display name
   - Grants permissions to ContentStudio
5. **Callback returns to ContentStudio** with:
   - Authorization code (exchanged for access token)
   - WABA ID
   - Phone Number ID
6. **ContentStudio stores** the access token (encrypted), WABA ID, phone number ID, and display name in `social_integrations`
7. **ContentStudio subscribes to webhooks** for this WABA to receive incoming messages

### Required Permissions (Scopes)

- `whatsapp_business_management` — manage WABA settings, phone numbers, templates
- `whatsapp_business_messaging` — send and receive messages via Cloud API

### Token Management

- **System User Tokens** (recommended for production): Long-lived, scoped to individual onboarded customer WABAs, don't require re-authentication
- **User Tokens**: Short-lived, need refresh flow
- For ContentStudio, use **Business Integration System User tokens** — they persist and work for automated actions without requiring user re-authentication

---

## 4. Webhooks & Receiving Messages

### Webhook Verification (GET)

When registering the webhook URL, Meta sends a GET request:

```
GET /whatsapp/webhook?hub.mode=subscribe&hub.verify_token={our_token}&hub.challenge={challenge}
```

Server must validate `hub.verify_token` and respond with `hub.challenge` value + HTTP 200.

### Incoming Message Payload (POST)

```json
{
  "object": "whatsapp_business_account",
  "entry": [{
    "id": "WABA_ID",
    "changes": [{
      "field": "messages",
      "value": {
        "messaging_product": "whatsapp",
        "metadata": {
          "display_phone_number": "15551234567",
          "phone_number_id": "PHONE_NUMBER_ID"
        },
        "contacts": [{
          "profile": { "name": "Customer Name" },
          "wa_id": "CUSTOMER_PHONE"
        }],
        "messages": [{
          "id": "wamid.xxx",
          "from": "CUSTOMER_PHONE",
          "timestamp": "1234567890",
          "type": "text",
          "text": { "body": "Hello, I need help" }
        }]
      }
    }]
  }]
}
```

### Message Types Received via Webhook

| Type | Fields |
|---|---|
| `text` | `text.body` |
| `image` | `image.id`, `image.mime_type`, `image.caption` |
| `video` | `video.id`, `video.mime_type`, `video.caption` |
| `audio` | `audio.id`, `audio.mime_type` |
| `document` | `document.id`, `document.filename`, `document.mime_type`, `document.caption` |
| `location` | `location.latitude`, `location.longitude`, `location.name`, `location.address` |
| `contacts` | `contacts[].name`, `contacts[].phones[]` |
| `sticker` | `sticker.id`, `sticker.mime_type` |
| `reaction` | `reaction.message_id`, `reaction.emoji` |
| `interactive` | `interactive.type`, `interactive.button_reply` or `interactive.list_reply` |

### Status Webhooks

```json
{
  "statuses": [{
    "id": "wamid.xxx",
    "status": "delivered",  // sent | delivered | read | failed
    "timestamp": "1234567890",
    "recipient_id": "CUSTOMER_PHONE"
  }]
}
```

### Security: Signature Verification

All webhook POSTs include an `X-Hub-Signature-256` header — HMAC-SHA256 of the raw body using the app secret. **Must validate in production** to prevent spoofed payloads.

### Retry Behavior

If webhook returns non-200, Meta retries with decreasing frequency for up to 7 days. Handlers should be **idempotent** — use message IDs for deduplication (Redis with TTL).

---

## 5. Sending Messages & Replies

### API Endpoint

```
POST https://graph.facebook.com/v21.0/{phone_number_id}/messages
Authorization: Bearer {ACCESS_TOKEN}
Content-Type: application/json
```

### Text Reply

```json
{
  "messaging_product": "whatsapp",
  "recipient_type": "individual",
  "to": "CUSTOMER_PHONE",
  "type": "text",
  "text": {
    "preview_url": false,
    "body": "Thank you for reaching out! How can we help?"
  }
}
```

### Image Reply

```json
{
  "messaging_product": "whatsapp",
  "to": "CUSTOMER_PHONE",
  "type": "image",
  "image": {
    "link": "https://example.com/image.jpg",
    "caption": "Here's the info you requested"
  }
}
```

### Video Reply

```json
{
  "messaging_product": "whatsapp",
  "to": "CUSTOMER_PHONE",
  "type": "video",
  "video": {
    "link": "https://example.com/video.mp4",
    "caption": "Product demo"
  }
}
```

### Document Reply

```json
{
  "messaging_product": "whatsapp",
  "to": "CUSTOMER_PHONE",
  "type": "document",
  "document": {
    "link": "https://example.com/catalog.pdf",
    "caption": "Our product catalog",
    "filename": "catalog.pdf"
  }
}
```

### Interactive Buttons

```json
{
  "messaging_product": "whatsapp",
  "to": "CUSTOMER_PHONE",
  "type": "interactive",
  "interactive": {
    "type": "button",
    "body": { "text": "How can we help?" },
    "action": {
      "buttons": [
        { "type": "reply", "reply": { "id": "btn_support", "title": "Support" } },
        { "type": "reply", "reply": { "id": "btn_sales", "title": "Sales" } }
      ]
    }
  }
}
```

### The 24-Hour Customer Service Window (Critical Rule)

- **When a customer messages your business**, a 24-hour window opens
- **During this window**: You can send any message type freely (text, media, interactive) — these are "service conversations" and are **free of charge**
- **Each customer reply resets the timer** — active conversations stay free indefinitely as long as the customer keeps responding
- **After the window expires**: You can ONLY send pre-approved **Message Templates** (see section 6)
- **Click-to-WhatsApp ads**: If customer initiates via a CTWA ad or Facebook Page CTA, the window extends to **72 hours** and all messages are free

**For ContentStudio's reply-only model**: Since customers always initiate, every conversation starts with an open window. The key constraint is ensuring agents reply within 24 hours of the last customer message.

---

## 6. Message Templates

### When Required

Templates are the **only way** to message a customer outside the 24-hour service window. Since ContentStudio is reply-only, templates are mainly needed for:
- Follow-up messages after the window expires (e.g., "We have an update on your request")
- Re-engaging a customer who hasn't responded in 24+ hours

### How They Work

1. Business creates a template via the WhatsApp Business Management API or Meta Business Manager
2. Meta reviews and approves the template (~24 hours)
3. Template can then be sent via the Cloud API
4. When customer replies to a template, it opens a new 24-hour service window

### Template Categories

| Category | Use Case | Cost |
|---|---|---|
| **Service** (utility) | Order updates, appointment reminders | Low cost |
| **Marketing** | Promotions, offers | Higher cost |
| **Authentication** | OTP, verification codes | Low cost |

### V1 Scope for ContentStudio

For v1, we can defer template management and focus purely on the reply-within-24-hours model. Template support can be a v2 feature for users who need to follow up after the window closes.

---

## 7. Media Handling

### Supported Formats & Size Limits

| Media Type | Formats | Max Size |
|---|---|---|
| **Image** | JPEG, PNG | 5 MB |
| **Video** | MP4 (H.264 video + AAC audio) | 16 MB |
| **Audio** | AAC, AMR, MP3, M4A, OGG (Opus) | 16 MB |
| **Document** | PDF, DOC, DOCX, XLS, XLSX, PPT, PPTX, TXT | 100 MB |
| **Sticker** | WebP | 100 KB (static), 500 KB (animated) |

### Uploading Media

```
POST https://graph.facebook.com/v21.0/{phone_number_id}/media
Content-Type: multipart/form-data

file: <binary>
type: <mime_type>
messaging_product: whatsapp
```

Returns a `media_id` that can be used in message payloads instead of a URL link.

### Downloading Media (from incoming messages)

```
GET https://graph.facebook.com/v21.0/{media_id}
Authorization: Bearer {ACCESS_TOKEN}
```

Returns a temporary download URL. **Media URLs expire** — download promptly and store in ContentStudio's media library.

**Uploaded files are auto-deleted after 30 days** by Meta.

### For ContentStudio

- **Incoming media**: Download from Meta URL immediately on webhook receipt, store in our media storage (S3/equivalent)
- **Outgoing media**: Upload to WhatsApp via media endpoint or use a publicly accessible URL (link method)
- **One attachment per message** — matches Zoho Social's behavior and WhatsApp's native UX

---

## 8. Pricing Model (as of July 2025)

WhatsApp shifted from conversation-based to **per-message pricing** effective July 1, 2025.

| Message Category | Cost Range (varies by country) | Notes |
|---|---|---|
| **Service (customer-initiated)** | **FREE** | Replies within the 24-hour window |
| **Marketing** | $0.025–$0.1365/msg | Promotions, offers |
| **Utility** | $0.004–$0.0456/msg | Order updates, reminders |
| **Authentication** | $0.004–$0.0456/msg | OTP codes |

### Key Pricing Facts for ContentStudio

- **Reply-only model = all service messages = FREE** for our users (Meta charges nothing for replies within the 24-hour window)
- Volume-based discounts available for utility/authentication messages
- Click-to-WhatsApp ad conversations: free for 72 hours
- ContentStudio doesn't need to handle billing pass-through for v1 since service messages are free

---

## 9. Rate Limits & Throughput

| Metric | Default | Upgraded |
|---|---|---|
| **Messages per second (MPS)** | 80 MPS per phone number | Up to 1,000 MPS (eligible accounts) |
| **Coexistence numbers** (Cloud API + WhatsApp Business App) | 20 MPS fixed | Cannot upgrade |

### Upgrade Eligibility (to 1,000 MPS)

- Phone number registered with Cloud API
- Can initiate unlimited unique conversations in 24 hours
- Medium or higher quality rating

### Error Handling

- Exceeding rate limit returns error code `130429`
- Implement exponential backoff and queue-based sending

### For ContentStudio

80 MPS default is more than sufficient for inbox reply use cases. Rate limiting is mainly a concern for bulk template sends (v2).

---

## 10. Phone Number Requirements

- **New number or existing**: Can register a new number or migrate an existing one
- **Cannot be on WhatsApp personal or WhatsApp Business App simultaneously** (unless using "coexistence" mode at reduced 20 MPS)
- **Verification**: Via SMS or phone call — one-time during Embedded Signup
- **Display Name**: Must match business name, follows WhatsApp's display name guidelines
- **One phone number = one channel** in ContentStudio

---

## 11. Key Technical Considerations for ContentStudio

### Token Management
- Use **System User tokens** (long-lived, no expiry) — different from Facebook's 60-day page tokens
- Store encrypted in `social_integrations` collection using existing `SocialHelper::encryptToken()`
- No refresh token flow needed initially

### Webhook Routing (Multi-Tenant)
- Single webhook endpoint receives events for ALL connected WhatsApp accounts
- Route by `phone_number_id` in webhook metadata → map to workspace/account in `social_integrations`
- Must handle high volume — enqueue to Redis immediately, process async via jobs

### Webhook Security
- **HMAC-SHA256 signature validation** via `X-Hub-Signature-256` header (stricter than Instagram's simple token verification)
- Must validate before processing any webhook event

### Message Deduplication
- WhatsApp may retry webhook delivery — use `message.id` (wamid) as deduplication key
- Store processed message IDs in Redis with TTL (few hours)

### Media Storage
- Download incoming media immediately (Meta URLs expire)
- Store in ContentStudio's existing media storage
- For replies with attachments: upload to WhatsApp media endpoint or use public URL

### Conversation Threading
- Thread by sender phone number (`from` field) + our phone number ID
- Maps cleanly to existing `inbox_details` model with `element_details.conversation_id`

---

## 12. Competitor Snapshot (Minimal)

| Platform | WhatsApp Inbox? | Connection Flow | Key Notes |
|---|---|---|---|
| **Zoho Social** | Yes | Facebook Business Login → select WABA → verify phone → name channel | Reply-only inbox, one attachment at a time, straightforward chat UI |
| **Sprout Social** | Yes | Similar Meta Embedded Signup | Full inbox integration with smart inbox, automated rules |
| **Hootsuite** | Yes | Via Sparkcentral (acquired) | Enterprise-focused, separate product for messaging |

All competitors follow the same Meta Embedded Signup pattern. The inbox UX is consistently a chat-style interface similar to Facebook Messenger / Instagram DMs — which ContentStudio's existing inbox already supports.

---

## 13. Recommended Approach for ContentStudio

### V1 Scope (MVP)
1. **Account Connection**: Embedded Signup via Facebook Business Login (same redirect pattern as existing Facebook/Instagram connections)
2. **Receive Messages**: Webhook handler → Redis queue → job processor → store in `inbox_details`
3. **Reply to Messages**: Text + single media attachment (image, video, document) within 24-hour window
4. **Inbox UI**: Add WhatsApp as a platform filter in existing inbox, show conversations in chat view
5. **Message Statuses**: Show sent/delivered/read indicators via status webhooks
6. **Auto-Reply Rules**: Extend existing auto-reply system to support WhatsApp conversations

### V2 (Defer)
- Message template management (create, submit for approval, send)
- Interactive messages (buttons, lists) from inbox
- WhatsApp Business Profile management
- Analytics/reporting on WhatsApp conversations
- Quick replies / canned responses specific to WhatsApp
- Contact management / customer profiles from WhatsApp

### Why This Approach Works
- **Reuses 80%+ of existing infrastructure**: Same MongoDB collections, same webhook pattern, same queue system, same inbox UI components
- **Reply-only = free for users**: No pricing complexity in v1 since service messages are free
- **Matches Zoho Social's approach**: Proven UX pattern, minimal innovation risk
- **Clean integration**: WhatsApp fits naturally as another platform in the existing multi-platform inbox

---

## Codebase Analysis Summary

### Existing Infrastructure (Reusable)

| Component | File Path | Reuse Strategy |
|---|---|---|
| Social account model | `app/Models/Integrations/Platforms/SocialIntegrations.php` | Add `platform_type: 'whatsapp'` |
| Token encryption | `SocialHelper::encryptToken()` | Use as-is for WhatsApp tokens |
| Inbox data model | `app/Models/Inbox/InboxDetails.php` | Same collection, add `platform: 'whatsapp'` |
| Inbox repository | `app/Repository/Inbox/InboxDetailsRepository.php` | Use existing CRUD operations |
| Webhook pattern | `app/Http/Controllers/Emails/SocialWebhooks/InstagramWebhooksController.php` | Clone pattern for WhatsApp |
| Queue system | `EnqueueDequeue::enqueue()` | New queue: `dramatics:inbox:whatsapp:webhook` |
| Auto-reply rules | `app/Services/Inbox/` | Extend to support WhatsApp |
| Logging | `LogsBuilder` | Use as-is |
| Frontend inbox | `src/modules/inbox-revamp/` | Add WhatsApp platform filter + icon |
| Integration settings | `src/modules/integration/components/platforms/social_v2/` | Add WhatsApp connection card |

### New Components Needed

| Component | Path | Purpose |
|---|---|---|
| WhatsApp webhook controller | `app/Http/Controllers/Emails/SocialWebhooks/WhatsAppWebhooksController.php` | Webhook verification + event handling |
| WhatsApp helper | `app/Libraries/Inbox/HelperClasses/WhatsAppHelper.php` | Message processing, reply sending |
| WhatsApp webhook job | `app/Jobs/WhatsAppWebhookExecuteJob.php` | Async message processing from queue |
| WhatsApp connection controller | `app/Http/Controllers/Integrations/Platforms/Social/WhatsAppController.php` | Embedded Signup OAuth flow |
| Platform config entries | `config/socialChannels.php`, `config/social_platforms.php` | Register WhatsApp as supported platform |
| Webhook route | `routes/web.php` | Public `/whatsapp/webhook` endpoint |

### Database Impact
- **No new collections** — uses existing `social_integrations` and `inbox_details`
- Add compound index on `platform_type: 'whatsapp'` + `workspace_id` for efficient queries
- New fields in social account: `whatsapp_business_account_id`, `phone_number_id`, `display_phone_number`
