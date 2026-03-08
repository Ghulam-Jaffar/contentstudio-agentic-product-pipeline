# PRD: WhatsApp Inbox Integration

**Author:** Product Team
**Last Updated:** 2026-03-08
**Status:** In Review
**Target Release:** Q2 2026

---

## 1. Overview

WhatsApp Inbox Integration adds WhatsApp Business as a channel in ContentStudio's Inbox module. Users connect their WhatsApp Business account via Meta's Embedded Signup (Facebook Business Login redirect), then receive and reply to customer messages directly from ContentStudio's Inbox — on web, iOS, and Android. This is a **reply-only** integration: customers initiate conversations by messaging the business's WhatsApp number, and agents respond within WhatsApp's 24-hour service window. The integration leverages the WhatsApp Cloud API (hosted by Meta) and reuses ContentStudio's existing inbox infrastructure, requiring no self-hosted WhatsApp infrastructure.

---

## 2. Problem Statement

**What problem are we solving?**

ContentStudio users who receive customer messages on WhatsApp must currently switch to WhatsApp Business App or a separate tool to respond. This fragments their inbox workflow — they already manage Facebook, Instagram, Twitter, LinkedIn, and GMB messages in ContentStudio's Inbox, but WhatsApp (the world's most popular messaging platform with 2B+ users) is missing. Users toggle between tools, miss messages, and lose response time.

**Who has this problem?**

- **Social media managers** handling customer support across multiple channels — they need one inbox for everything
- **Small businesses and agencies** using WhatsApp as a primary customer communication channel (especially in LATAM, SEA, MENA, and Europe where WhatsApp dominates)
- **E-commerce brands** receiving order inquiries, support requests, and product questions via WhatsApp

**What happens if we don't solve it?**

- Users continue fragmenting their workflow across ContentStudio + WhatsApp Business App
- Competitive disadvantage: Sprout Social, Hootsuite, and Zoho Social already offer WhatsApp inbox integration
- Potential churn from users who need unified inbox for all messaging channels
- Missed upsell opportunity: WhatsApp is a premium feature in most competitor pricing tiers
- Ignoring one of our most highly requested features — users have been actively asking for WhatsApp inbox integration

---

## 3. Goals & Success Metrics

| Goal | Metric | Target | How We'll Measure |
|---|---|---|---|
| Drive WhatsApp adoption | WhatsApp accounts connected | 500 workspaces in 90 days | Product analytics: social_integrations where platform_type = 'whatsapp' |
| Unify inbox workflow | % of inbox users who also use WhatsApp inbox | 25% of active inbox users | Product analytics: inbox sessions with WhatsApp filter |
| Maintain response quality | Average reply time for WhatsApp conversations | < 2 hours | Inbox analytics: time between customer message and first reply |
| No reliability regressions | Webhook processing success rate | > 99.5% | Logs: WhatsApp webhook job success/failure rate |

---

## 4. Target Users

**Primary Persona:**
Social Media Manager — manages 3-10 social accounts across platforms, handles customer inquiries daily, values having a single inbox for all channels. Moderate technical skill, familiar with ContentStudio's existing inbox workflow.

**Secondary Persona:**
Small Business Owner — runs their own social media, uses WhatsApp as primary customer contact, needs a simple way to manage WhatsApp messages alongside other social channels. Low technical skill, needs a straightforward connection flow.

**Non-Users (explicitly out of scope):**
- Businesses wanting to send bulk promotional messages via WhatsApp (no outbound/broadcast)
- Businesses needing WhatsApp chatbot automation beyond simple keyword auto-replies
- Businesses wanting to initiate conversations with customers (reply-only model)

---

## 5. User Stories / Jobs to Be Done

| ID | As a... | I want to... | So that... | Priority |
|---|---|---|---|---|
| US-1 | Social media manager | connect my WhatsApp Business account to ContentStudio | I can manage WhatsApp messages alongside my other social channels | Must Have |
| US-2 | Social media manager | see incoming WhatsApp messages in my ContentStudio Inbox | I don't have to switch to WhatsApp Business App to check messages | Must Have |
| US-3 | Social media manager | reply to WhatsApp messages with text and attachments from ContentStudio | I can respond to customers without leaving my unified inbox | Must Have |
| US-4 | Social media manager | know when the 24-hour reply window has expired | I don't waste time composing a reply that will fail to send | Must Have |
| US-5 | Social media manager | connect multiple WhatsApp numbers to my workspace | I can manage separate support, sales, and marketing WhatsApp lines in one place | Must Have |
| US-6 | Social media manager | filter my inbox to show only WhatsApp conversations | I can focus on WhatsApp messages when needed | Must Have |
| US-7 | Social media manager | set up auto-reply rules for WhatsApp | common questions get answered automatically even when I'm away | Should Have |
| US-8 | Social media manager | see message delivery status (sent, delivered, read) on WhatsApp | I know whether the customer has seen my reply | Nice to Have |
| US-9 | Social media manager | reconnect my WhatsApp account when the token expires | my inbox keeps working without manual intervention from support | Must Have |
| US-10 | Social media manager | manage WhatsApp messages from the ContentStudio mobile app | I can respond to urgent messages on the go | Must Have |
| US-11 | New user | connect WhatsApp during onboarding | I'm set up from day one without having to find the setting later | Should Have |

---

## 6. Requirements

### 6.1 Must Have (P0)

- **Account Connection**: Connect WhatsApp Business account via Meta Embedded Signup (Facebook Business Login redirect). Store WABA ID, phone number ID, access token (encrypted), display phone number, display name, and user-provided channel name in `social_integrations` collection.
- **Multiple Accounts**: Support connecting multiple WhatsApp phone numbers per workspace. Each number is a separate channel.
- **Webhook Receiver**: Public endpoint `/whatsapp/webhook` for Meta webhook verification (GET) and event reception (POST). Validate `X-Hub-Signature-256` HMAC-SHA256 signature on all incoming webhooks. Enqueue to Redis queue `dramatics:inbox:whatsapp:webhook` for async processing.
- **Message Ingestion**: Process incoming webhook events — extract message content (text, image, video, audio, document, location, sticker, voice note, contacts), download media from Meta's temporary URLs, store in ContentStudio's media storage, create `inbox_details` records with `platform: 'whatsapp'`.
- **Reply API**: Send text replies and single-media replies (image, video, document) via `POST /v21.0/{phone_number_id}/messages` using the stored access token. Use public CDN URL for media (link method).
- **24-Hour Window Tracking**: Track the timestamp of the last customer message per conversation. When > 24 hours since last customer message, block replies — return an error to the frontend so it can disable the composer and show the expired banner.
- **Inbox UI (Web)**: Display WhatsApp conversations in the existing inbox conversation list with WhatsApp platform badge. Chat view shows message thread with text, inline media, timestamps. Reply composer supports text + single attachment. Platform filter includes "WhatsApp" option. Show expired-window banner when applicable.
- **Account Management**: Edit channel name, disconnect account (removes webhook subscription, marks as disconnected), reconnect on token expiry (re-auth via Facebook Business Login).
- **Token Expiry Handling**: Detect invalid tokens via failed API calls (increment `invalid_tries`). Show warning badge on account card + inbox banner. Send email notification after repeated failures (reuse existing `sent_invalid_email` pattern). Provide "Reconnect" action.
- **Inbox on Mobile (iOS & Android)**: WhatsApp conversations visible in mobile inbox with reply capability (text + single attachment). Same 24-hour window behavior.
- **Social Account Connection on Mobile**: WhatsApp connection available in iOS/Android app settings via system browser for Facebook Business Login.
- **Platform Configuration**: Register WhatsApp in `config/socialChannels.php` and `config/social_platforms.php`. Add WhatsApp helper class for message processing and reply sending.
- **Message Deduplication**: Use message ID (`wamid`) as deduplication key in Redis with TTL to prevent duplicate processing from webhook retries.
- **Auto-Reply Rules**: WhatsApp as a platform option in existing auto-reply rule system. Support keyword-based triggers and text-only responses. Support account selection for workspaces with multiple WhatsApp accounts.
- **Onboarding**: WhatsApp as a connection option in the new-user onboarding wizard (Connect Social Accounts step).
- **Team Collaboration**: WhatsApp conversations support assignment to team members, tagging, mark as done/resolved, and read/unread marking — same as existing inbox platforms. Reuse existing assignment and resolution logic.
- **Notifications**: New WhatsApp messages trigger in-app, browser push, and mobile push notifications + email notifications via existing notification system — same behavior as other inbox platforms.

### 6.2 Should Have (P1)

(None — all functional requirements moved to P0 or P2.)

### 6.3 Nice to Have (P2)

- **Message Status Ticks**: Display sent (✓), delivered (✓✓), and read (blue ✓✓) indicators on outgoing messages. Data sourced from WhatsApp status webhooks. WhatsApp-only feature — not added to other platforms.
- **Rich Media Display**: Location messages shown as map cards with "View on Map" link. Voice notes shown as inline audio player. Stickers displayed as images. Contact cards shown with name and phone number.

### 6.4 Explicitly Out of Scope

- No conversation initiation — reply-only model, customers always message first
- No message template management or sending
- No interactive messages (buttons, lists) sent from inbox
- No WhatsApp Business Profile editing from ContentStudio
- No WhatsApp-specific analytics or reporting dashboard
- No bulk messaging or broadcast campaigns
- No WhatsApp Commerce (product catalogs, carts, payments)
- No AI-powered auto-replies (AI is web-only and out of scope for this integration)

---

## 7. User Flow (High Level)

### Connect Account
1. User goes to Settings → Social Accounts (or Onboarding, or mobile app settings)
2. User clicks "Connect" next to WhatsApp
3. User is redirected to Facebook Business Login
4. User completes Meta Embedded Signup (login → select/create WABA → add/verify phone → grant permissions)
5. User is redirected back to ContentStudio
6. User enters a channel name in the confirmation modal and clicks "Connect"
7. WhatsApp account appears in Social Accounts list

### Receive & Reply
1. Customer sends a WhatsApp message to the business number
2. Message appears in ContentStudio Inbox (web + mobile) with WhatsApp badge
3. User opens conversation, sees message thread
4. User types reply (optional: attaches one file) and clicks "Send"
5. Reply is delivered to customer on WhatsApp

### Account Management
1. User finds WhatsApp account in Settings → Social Accounts
2. User can edit channel name, disconnect, or reconnect (if token expired)
3. Token expiry: account shows warning badge, user clicks "Reconnect" → Facebook re-auth → token refreshed

---

## 8. Business Rules & Constraints

| Rule ID | Rule | Rationale |
|---|---|---|
| BR-1 | Replies can only be sent within 24 hours of the last customer message | WhatsApp Cloud API enforces a 24-hour customer service window. After expiry, only pre-approved templates can be sent (not in scope). |
| BR-2 | Each customer reply resets the 24-hour window timer | WhatsApp's window rule: timer resets on every customer message. An active conversation stays open indefinitely as long as the customer keeps responding. |
| BR-3 | Only one attachment per reply message | WhatsApp API limitation — each message can contain one media object. |
| BR-4 | Media size limits: images ≤ 5 MB, video/audio ≤ 16 MB, documents ≤ 100 MB | WhatsApp Cloud API enforced limits. Validate before sending. |
| BR-5 | Supported media formats: JPEG/PNG (images), MP4 H.264 (video), AAC/AMR/MP3/M4A/OGG (audio), PDF/DOC/DOCX/XLS/XLSX/PPT/PPTX/TXT (documents) | WhatsApp Cloud API supported formats. Reject unsupported formats with clear error message. |
| BR-6 | Webhook payloads must pass HMAC-SHA256 signature verification | Security requirement — `X-Hub-Signature-256` header validated against app secret. Reject unverified payloads. |
| BR-7 | Same phone number cannot be connected twice in the same workspace | Prevent duplicate accounts. Check before saving. |
| BR-8 | Service messages (replies within 24h window) are free — no per-message cost to ContentStudio users | WhatsApp pricing as of July 2025. Reply-only model = all service conversations = free. |
| BR-9 | A phone number registered with Cloud API cannot simultaneously be used on WhatsApp personal app or WhatsApp Business App (unless coexistence mode at reduced 20 MPS) | Meta platform constraint. User must understand they're dedicating this number to the API. |
| BR-10 | Auto-reply rules fire only within the 24-hour window | Auto-replies are still subject to the service window. If window expired, auto-reply is suppressed. |

---

## 9. Open Questions

| Question | Options | Owner | Due Date | Decision |
|---|---|---|---|---|
| Should WhatsApp be available on all pricing plans or only premium? | All plans / Business+ / Enterprise only | Product | Before dev starts | Pending |
| Do we need a WhatsApp-specific plan limit (e.g., max N WhatsApp accounts per plan)? | Unlimited / Plan-based limit / Add-on | Product | Before dev starts | Pending |
| Should we show a "reply window closes in X hours" countdown in the UI? | Yes (countdown) / No (just disable when expired) | Design | During design phase | Pending — recommended: no countdown in v1, just disable |
| How do we handle the edge case where a customer messages multiple connected WhatsApp numbers in the same workspace? | Separate conversations per number / Merge somehow | Engineering | During dev | Pending — recommended: separate conversations, each number is independent |

---

## 10. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Meta changes Embedded Signup flow or API version | Medium | High | Pin to a specific Graph API version (v21.0). Monitor Meta's changelog. Abstract the connection flow so changes are isolated to one controller. |
| Webhook volume spikes overwhelm queue processing | Low | Medium | Async processing via Redis queue with configurable concurrency. Rate limit webhook endpoint. Monitor queue depth with alerts. |
| Users don't understand the 24-hour window and get frustrated when reply is disabled | Medium | Medium | Clear banner copy explaining why reply is disabled and what to do. Tooltip on the disabled composer. Help article link. |
| Token expiry goes unnoticed, messages stop flowing | Medium | High | Proactive detection via failed API calls → email notification → in-app warning badge → "Reconnect" CTA. Same pattern proven for Facebook/Instagram token expiry. |
| Phone number conflicts — user's number already registered with another BSP | Low | Low | Meta handles this during Embedded Signup. We show a clear error with a help article on how to migrate the number. |
| Media download from Meta URLs fails (URLs expire) | Low | Medium | Download media immediately on webhook receipt. Retry with exponential backoff. Log failures for monitoring. Store a "media unavailable" placeholder if download ultimately fails. |

---

## 11. Dependencies

**Internal:**
- **Inbox module** (`app/Models/Inbox/InboxDetails.php`, `app/Repository/Inbox/InboxDetailsRepository.php`) — WhatsApp conversations stored in same `inbox_details` collection
- **Social Integrations model** (`app/Models/Integrations/Platforms/SocialIntegrations.php`) — WhatsApp accounts stored with `platform_type: 'whatsapp'`
- **Token encryption** (`SocialHelper::encryptToken()`) — reused for WhatsApp access tokens
- **Queue system** (`EnqueueDequeue::enqueue()`) — new queue `dramatics:inbox:whatsapp:webhook`
- **Webhook infrastructure** — follows pattern from `InstagramWebhooksController.php`
- **Auto-reply service** (`app/Services/Inbox/`) — extend to support WhatsApp platform
- **Frontend inbox module** (`src/modules/inbox-revamp/`) — add WhatsApp platform support
- **Frontend integration module** (`src/modules/integration/components/platforms/social_v2/`) — add WhatsApp connection card
- **iOS app** (`contentstudio-ios-v2/`) — inbox view + social account connection
- **Android app** (`contentstudio-android-v2/`) — inbox view + social account connection
- **Notification system** — extend to trigger on WhatsApp messages
- **Onboarding wizard** — add WhatsApp to social account connection step

**External:**
- **WhatsApp Cloud API** (Meta) — `graph.facebook.com/v21.0` for sending messages and media
- **Meta Embedded Signup** — for account connection (Facebook Business Login)
- **Meta Webhooks** — for receiving messages and status updates
- **Meta App Dashboard** — ContentStudio must be registered as a Tech Provider with `whatsapp_business_management` and `whatsapp_business_messaging` permissions approved

**Blockers:**
- ContentStudio's Meta App must have WhatsApp permissions approved in the App Dashboard before development can begin on the connection flow
- A test WhatsApp Business Account and phone number are needed for development and QA

---

## 12. Appendix

- [WhatsApp Cloud API Overview](https://developers.facebook.com/docs/whatsapp/cloud-api/overview)
- [WhatsApp Embedded Signup](https://developers.facebook.com/documentation/business-messaging/whatsapp/embedded-signup/overview/)
- [WhatsApp Webhooks Reference](https://developers.facebook.com/documentation/business-messaging/whatsapp/webhooks/reference/messages/)
- [WhatsApp Send Messages Guide](https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages/)
- [WhatsApp Business Platform Pricing](https://business.whatsapp.com/products/platform-pricing)
- Feature Research: `docs/features/whatsapp-inbox-integration/01-research.md`
- Workflow Design: `docs/features/whatsapp-inbox-integration/02-workflow.md`

---

## Changelog

| Date | Author | Changes |
|---|---|---|
| 2026-03-08 | Product Team | Initial draft |
