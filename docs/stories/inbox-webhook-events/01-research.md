# Inbox Webhook Events — Research

Date: August 3, 2026

Scope: emitting outbound public webhooks for inbox activity (new messages, comments, reviews, and triage actions), reusing the webhook delivery engine already built for publishing events.

---

## Current State

### The delivery engine is already built, and already cross-service

Everything below exists on `contentstudio-backend` branch **`features`** (merged via `feature/cont-2703-be-build-webhook-delivery-log-test-event-apis`). It is **not on `develop` or `master`**.

| Piece | Location |
|---|---|
| Endpoint CRUD, secret rotation, test event | `routes/api.php:157-171` → `WebhookEndpointController` |
| Delivery logs + CSV export, event-type catalog | `WebhookDeliveryLogController`, `WebhookEventTypeController` |
| Fan-out consumer | `app/Kafka/Handlers/WebhookEventHandler.php` |
| Delivery + retries + credit check | `app/Jobs/Webhooks/WebhookDeliveryJob.php` |
| Signing (Standard Webhooks headers) | `app/Services/Webhooks/WebhookSigner.php` |
| Event catalog | `app/Data/Webhooks/Enums/WebhookEventType.php` |
| Payload builder pattern to copy | `app/Services/Webhooks/PostWebhookPayload.php` |

The ingress topic is `contentstudio.api.webhooks.events` (`config/kafka.php:79`). The consumer's own docblock states the topic is **cross-service by design, with external producers publishing to it too**, which is exactly the seam this work needs. Nothing about delivery, signing, retries, logging, or the test button has to be rebuilt.

**Envelope the handler expects:**

```json
{
  "event_id": "uuid",
  "type": "message.received",
  "workspace_id": "...",
  "recipient_user_ids": ["..."],
  "data": { }
}
```

`event_id` is the dedupe key. `type` must exist in the enum or the message is dead-lettered. Routing is resolved server-side from current workspace membership and never from the envelope. Two modes: with `recipient_user_ids` the event targets those users, without it the event broadcasts to the workspace owner and admins.

Existing event types: `post.scheduled`, `post.published`, `post.failed`, `post.inreview`, `post.approved`, `post.rejected`, `post.comment`.

### The inbox service already has the exact emit points we need

`social-inbox-manager/app/utils/notification_helper.py` is the single place where "something happened in the inbox that a human should be told about" is decided. Its methods map almost one to one onto the event catalog:

| NotificationHelper method | Candidate event |
|---|---|
| `send_message_notification` | `message.received` |
| `send_comment_notification` | `comment.received` |
| `send_assigned_notification` | `inbox.item.assigned` |
| `send_unassigned_notification` | `inbox.item.unassigned` |
| `send_mark_done_notification` | `inbox.item.marked_done` |
| `send_archived_notification` | `inbox.item.archived` |
| `send_mentioned_notification` | `inbox.note.added` |

Called from the platform strategies, e.g. `facebook_strategy.py:839` (comments) and `:901` (messages).

The payload each method builds already carries what an envelope needs: `workspace_id`, `platform_id`, `element_id`, plus resource ids and `from`.

### Three gaps in that seam

1. **Reviews have no emit point.** `gmb_strategy.py` never calls `NotificationHelper`. `review.new` needs a new call site, it cannot piggyback an existing one.
2. **Notifications go to Redis, not Kafka.** These methods `rpush` onto `NOTIFICATION_QUEUE`. Webhook envelopes must go to the Kafka ingress topic, so this is a parallel emit rather than a reuse of the existing transport.
3. **A feature flag would silently disable webhooks.** Every method starts with `if SEND_EMAILS_AND_NOTIFICATION is False: return`. Emitting webhooks inside that guard means turning off notification emails also turns off customer webhooks. The webhook emit must sit outside it.

### The frontend webhooks UI does not exist yet

There is no webhook code in `contentstudio-frontend` on `develop`. The Webhooks tab, the endpoint list, and the event checkboxes are specified in the public-webhooks epic but unbuilt. The "Accounts · Inbox · Comments · Reviews / Coming soon" group is a spec, not shipped code.

This makes the FE work here **dependent on** `[FE] Build the Webhooks tab (manage webhooks + per-webhook delivery logs)` and `[FE] Restructure the API page into API and Webhooks tabs with a usage readout` from that epic.

---

## What Needs to Change

**Laravel (small):**
- Add inbox cases to `WebhookEventType` with labels, so the catalog endpoint and the UI pick them up automatically
- Add an `InboxWebhookPayload` builder mirroring `PostWebhookPayload`
- Decide the `post.comment` rename (see below)

**Social inbox manager (the real work):**
- A Kafka producer for `contentstudio.api.webhooks.events` and an envelope mapper
- Emit alongside each `NotificationHelper` call, outside the `SEND_EMAILS_AND_NOTIFICATION` guard
- A new emit point for GMB reviews
- Decide routing mode per event: arrivals broadcast, triage actions target the affected user

**Frontend (small, but blocked):**
- Render the new event groups in the Webhooks tab once that tab exists

---

## Naming Decision (locked by the user)

Resource-first, **no `inbox.` prefix**, matching the Zernio reference and our existing `post.*` convention. The UI groups events visually under Messages / Comments / Reviews headings rather than encoding the grouping in the event name.

Zernio's catalog for comparison: `message.received`, `message.sent`, `message.edited`, `message.deleted`, `message.delivered`, `message.read`, `message.failed`, `reaction.received`, `conversation.started`, `comment.received`, `review.new`, `review.updated`.

**`post.comment` is a collision risk.** It means an *internal* planner post discussion comment, which is a completely different thing from a social `comment.received`. Webhooks are not on `master` yet, so renaming it is still non-breaking. Flagged for the team rather than changed silently.

---

## Scope Limit

v1 covers only what we actually ingest today. Out of scope, with the reason:

| Zernio event | Why not v1 |
|---|---|
| `message.delivered`, `message.read` | Needs Meta `message_deliveries` / `message_reads`, not subscribed |
| `message.failed` | No failure signal captured today |
| `reaction.received` | Needs Meta `message_reactions`, not subscribed |
| `message.edited`, `message.deleted` | No edit or delete detection in ingestion today |
| `conversation.started` | No distinct first-message signal in the read model |
| `account.*`, `call.*`, `ad.*`, `whatsapp.*`, `verification.*` | Different domains, not inbox |

---

## Open Questions for the Stories

1. **Do our own outbound replies emit events?** If an agent replies from ContentStudio, or a customer automation replies through the Inbox Public API, does that emit a `message.sent`? If yes, an automation that replies on `message.received` can trigger itself. Needs a deliberate answer.
2. **Volume and credit cost.** Inbox events fire far more often than publishing events, and each delivery draws on the shared API credit pool. Worth sizing before enabling.

---

## Files Involved

**contentstudio-backend** (branch `features`)
- `app/Data/Webhooks/Enums/WebhookEventType.php`
- `app/Services/Webhooks/PostWebhookPayload.php` (pattern to copy)
- `app/Kafka/Handlers/WebhookEventHandler.php` (envelope contract, no change expected)
- `config/kafka.php`

**social-inbox-manager**
- `app/utils/notification_helper.py` (the seam)
- `app/social_sync/facebook_strategy.py`, `instagram_strategy.py`, `linkedin_strategy.py`, `youtube_strategy.py`, `gmb_strategy.py`, `whatsapp_strategy.py`
- `app/kafka/producer.py`, `app/config/kafka_config.py`
- `app/api/v1/routes/elements.py`, `sends.py` (triage and note actions)

**contentstudio-frontend**
- The Webhooks tab, once it exists
