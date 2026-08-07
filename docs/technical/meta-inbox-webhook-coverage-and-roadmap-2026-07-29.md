# Meta Inbox: Webhook Coverage Audit and Expansion Roadmap

Date: July 29, 2026

Audience: Product, Engineering Leadership

Scope: What the ContentStudio inbox currently ingests and acts on for Facebook Pages and Instagram business accounts, what Meta's two webhook surfaces actually offer, and a priority-ranked view of what we should add, filtered through what matters for our users.

References:

- [Meta Page Webhooks Reference](https://developers.facebook.com/docs/graph-api/webhooks/reference/page/) — Facebook Pages only
- [Meta Instagram Webhooks Reference](https://developers.facebook.com/docs/graph-api/webhooks/reference/instagram/) — Instagram only, a separate object with a different and much smaller field catalog

**These are two different webhook objects.** They do not share a field list. Several capabilities that exist on one do not exist on the other at all. Any plan that treats "Meta" as one surface will schedule work Meta does not expose.

---

## TL;DR

1. On both platforms we subscribe to fields we never handle. Facebook subscribes to `feed`, `messages`, `conversations` and acts on two event shapes. Instagram subscribes to `comments`, `messages`, `mentions` and acts on two. **Both are discarding events at the API layer that Meta is already delivering.**
2. **Instagram already subscribes to `mentions` and drops every single one.** There is no branch for it in the handler. This is the cheapest item in this document: the subscription exists, the permission exists, only the handler is missing.
3. On Facebook the highest-value work is not new subscriptions. It is finishing the `feed` field we already receive: comment deletes, edits, native hide and unhide, and new-post events. A comment deleted on Facebook currently stays in our inbox forever.
4. Every proposed Facebook addition is already covered by the OAuth scopes we request. **No new App Review, no new consent screen.** The cost is subscription config, a backfill job, and handler work.
5. The two platforms need different roadmaps. Facebook Page Recommendations, message echoes, and feed-level delete and edit events **have no Instagram equivalent**. Instagram live comments have no Facebook equivalent in our scope.
6. The Instagram story mentions path does not appear to work. No producer writes to the topic its worker consumes, and the handler branch that should feed it posts a payload that fails its own validation schema.

---

## Part 1: What We Capture Today

### 1.1 Subscriptions versus actual handling

| | Facebook Page | Instagram |
|---|---|---|
| Subscribed fields | `feed`, `messages`, `conversations` | `comments`, `messages`, `mentions` |
| Config location | `contentstudio-backend/config/integrations.php:25` | `contentstudio-backend/config/integrations.php:540` |
| Handled in code | inbound DMs; comments, but only if the comment is the **first** change in the entry | inbound DMs; comments, but only the **first** change; plus a `story_insights` branch for a field we do not subscribe to |
| Subscribed but unhandled | **`conversations`** | **`mentions`** |
| Handler | `social-inbox-manager/app/api/webhooks/facebook_webhook_actions.py:95` | `social-inbox-manager/app/api/webhooks/instagram_webhook_actions.py:118` |

Facebook's routing:

```python
if "messaging" in entry:
    await WebhookHandler.process_message(entry)
elif "changes" in entry and entry["changes"][0]["value"]["item"] == "comment":
    await WebhookHandler.process_comment(entry)
```

Instagram's routing:

```python
if "messaging" in entry:
    await WebhookHandler.process_message(entry)
elif "changes" in entry:
    change_value = entry["changes"][0]
    if change_value["field"] == "comments":
        await WebhookHandler.process_comment(entry)
    elif change_value["field"] == "story_insights":
        # HTTP redirect to a hardcoded production URL
```

A `mentions` event matches neither branch, falls through, and the endpoint returns `{"success": True}`. Meta considers it delivered. We never see it.

### 1.2 Defects visible in both handlers

- **Only `changes[0]` is inspected, on both platforms.** Meta batches multiple changes into one entry. If any other change type sits ahead of a comment in the same entry, the comment is silently dropped. Both downstream workers do loop over all changes, so this is purely an API-layer filter bug and is cheap to fix.
- **No change-type or `verb` handling on Facebook.** `verb` tells us `add`, `edited`, `remove`, `hide`, `unhide`. We treat every comment event as "go fetch this comment". On a delete the fetch returns nothing and we no-op, leaving the stale comment in the inbox permanently.

### 1.3 The Instagram story mentions path looks broken

Three things do not line up, and they should be checked together:

1. `story_insights` is **not** in our subscribed field list, yet the handler has a live branch for it.
2. That branch does not produce to Kafka. It makes an HTTP POST to a **hardcoded** `https://api-prod.contentstudio.io/instagram/webhook`, sending `{"entry": ...}` without the `object` key that its own `InstagramWebhookPayload` schema requires. The receiving end should reject it.
3. `INSTAGRAM_STORY_INSIGHTS_TOPIC` has a worker consuming it and a consumer group configured, but **no producer anywhere in the codebase writes to it**.

So there is a fully built story mentions consumer waiting on a topic nothing feeds. This needs confirming against a live environment before we treat it as definitive, but on a read of the code the feature cannot be working. The hardcoded production URL also means any non-production environment posts to production.

### 1.4 Periodic sync, the actual safety net

Both platforms poll on identical limits (`app/jobs/facebook_inbox_job.py:22`, `app/jobs/instagram_inbox_job.py:22`):

| Job | Per-run limit |
|---|---|
| conversations | 50 |
| messages | 500 |
| posts | 50 |
| comments | 500 |

This is what makes new posts, reaction counts, and natively-sent replies eventually appear. It is also why users perceive the inbox as lagging.

### 1.5 What agents can do today

| Action | Facebook | Instagram |
|---|---|---|
| Reply to comment | Yes | Yes |
| Reply privately to a commenter | Yes | Yes |
| Hide / unhide comment | Yes | Yes |
| Delete comment | Yes | Yes |
| **Like / unlike comment** | Yes | **No.** Gated to Facebook only in `comment_actions.py:107`, and Meta does not expose it for Instagram |
| Send DM text and media | Yes | Yes |
| Tag, assign, note, bookmark, archive, mark done, saved replies, AI auto-replies | Yes | Yes |

Two pieces of existing engineering worth preserving:

- DM sending uses `messaging_type: MESSAGE_TAG` with `tag: HUMAN_AGENT` on both platforms, correct after Meta deprecated `ACCOUNT_UPDATE` in April 2026, giving a seven-day reply window.
- Both helpers split message fields into core and optional tiers so a permission-gated field like `reactions` degrades gracefully instead of failing the whole call.

Instagram additionally runs **two connection paths** with different base URLs, auth styles, and field sets: direct accounts on `graph.instagram.com` with a Bearer token, and Facebook-connected accounts on `graph.facebook.com` with `appsecret_proof` (`instagram_helper.py:62-109`). Any Instagram change has to be verified against both.

---

## Part 2: The Facebook Page Webhook Surface

### 2.1 Content and engagement

| Field | Fires on | Subscribed | Handled | Fit |
|---|---|---|---|---|
| `feed` | posts, comments, reactions, shares, edits, deletes, hides across the Page feed | Yes | Comments only | **Core.** Underused |
| `mention` | the Page is tagged in someone else's post or comment | No | No | **High.** New surface |
| `ratings` | Page Recommendations created, edited, deleted, commented on, reacted to | No | No | **High.** Reuses existing UI |
| `group_feed` | comments on Page posts inside Facebook Groups | No | No | Low. Meta's Groups API is heavily restricted |

`feed` sub-values: `item` can be `post`, `comment`, `reaction`, `share`, `status`, `photo`, `video`, `link`, `album`, `story`, `mention`, `rating`, `event` and more. `verb` can be `add`, `edit`, `edited`, `remove`, `delete`, `hide`, `unhide`, `block`, `unblock`, `follow`, `mute`, `update`.

### 2.2 Messaging

| Field | Fires on | Subscribed | Handled | Fit |
|---|---|---|---|---|
| `messages` | inbound customer DM | Yes | Yes | Core |
| `conversations` | thread-level update | Yes | **No** | Medium. Already paid for |
| `message_echoes` | a message the Page sent, including from Meta Business Suite or another tool | No | No | **High.** Prevents double-replies |
| `message_reactions` | customer reacts to one of our messages | No | No | **High** |
| `message_edits` | customer edits a message they sent | No | No | Medium |
| `messaging_referrals` | DM started from an ad or m.me link, carries `ad_id` and `ref` | No | No | **High.** Attribution |
| `message_deliveries` | our message was delivered | No | No | Medium |
| `message_reads` | customer read our message | No | No | Medium |
| `messaging_policy_enforcement` | Meta warned or acted on our sending | No | No | Medium. Ops visibility |
| `messaging_handovers` | thread ownership moved between apps | No | No | Medium. Multi-tool Pages |
| `standby` | thread idle for a secondary-receiver app | No | No | Medium. Pairs with handovers |
| `messaging_postbacks` | Get Started, persistent menu, quick reply clicks | No | No | Low. Only if we ship bot flows |
| `messaging_optins` | plugin opt-ins, customer matching | No | No | Low |
| `message_context` | Meta's own detection and context signals | No | No | Low |
| `inbox_labels` | labels applied in the native Page inbox | No | No | Low. Could mirror our tags |
| `send_cart`, `calls`, `call_settings_update`, `call_permission_reply` | commerce and calling | No | No | Not our market |
| `messaging_customer_information`, `message_template_status_update` | structured data collection, template approvals | No | No | Not applicable |

### 2.3 Page administration

`leadgen` (lead ad submissions, a different product surface), `live_videos` and `videos` (publishing and analytics), `name`, `picture`, `category`, `page_change_proposal`, `page_upcoming_change` (could feed account-health warnings, not inbox), `business_integrity` (compliance notices, ops-only).

---

## Part 3: The Instagram Webhook Surface

Twelve fields against the Page object's roughly thirty-five. The expansion room here is genuinely narrower, and that is a platform fact rather than a scoping choice.

| Field | Fires on | Subscribed | Handled | Fit |
|---|---|---|---|---|
| `comments` | comment on our media | Yes | Yes | Core |
| `messages` | inbound DM | Yes | Yes | Core |
| `mentions` | someone @mentions us in their comment or caption | **Yes** | **No** | **Highest.** Already paid for, silently dropped |
| `live_comments` | comments during an Instagram Live | No | No | **Medium to high.** No Facebook equivalent. Strong creator fit |
| `message_reactions` | customer reacts to one of our messages | No | No | High |
| `message_edit` | customer edits a message. Note the singular, Facebook uses `message_edits` | No | No | Medium |
| `messaging_referral` | referral context on a conversation. Note the singular | No | No | Medium |
| `messaging_seen` | customer read our message. Instagram's equivalent of `message_reads` | No | No | Medium |
| `messaging_handover` | thread control transfer. Note the singular | No | No | Medium |
| `messaging_postbacks` | postback interactions | No | No | Low |
| `standby` | secondary-receiver app standby | No | No | Low |
| `story_insights` | fires when a story expires, with metrics. Counts below 5 return as `-1` | No | Branch exists, path appears broken | See 1.3 |

### 3.1 What Instagram does not have

This is the part that matters most for planning, because these are the items that cannot be built regardless of priority:

- **No `feed` field.** Instagram has `comments` only. There is no reaction, share, edit, hide, or delete event stream. Facebook's entire P0 "handle the verbs" workstream has no Instagram counterpart, and detecting a deleted or edited Instagram comment may only be possible by polling.
- **No `ratings`.** Instagram has no reviews or recommendations. The Page Recommendations win is Facebook-only.
- **No `message_echoes`.** The double-reply prevention win is Facebook-only.
- **No `message_deliveries` or `message_reads`.** Only `messaging_seen`.
- **No comment liking**, at the API level, so our Facebook-only gating is correct rather than an oversight.

---

## Part 4: Gaps, Grouped by Where They Apply

Effort is S / M / L across the combined backend, inbox-service, and frontend change.

### Group A: Shared, both platforms

| # | Gap | Why it matters | Effort |
|---|---|---|---|
| A1 | **Only `changes[0]` is inspected** | Batched entries silently drop real comments on both platforms. Data loss, not a missing feature. Cheapest correctness fix in this document | S |
| A2 | **Backfill subscription for existing accounts** | Subscription happens only at connect time. Adding fields to config does nothing for the installed base. Needs a one-off job re-subscribing every connected account. **This blocks every new field from reaching current customers on both platforms** | S |
| A3 | **Field lists can drift** | Facebook's list is defined twice, in config and hardcoded in `PlatformObserver`. Instagram's lives in config but is applied through a different helper. Collapse to one source per platform | S |
| A4 | **Auto-reply eligibility for new event types** | Rule-based and AI auto-replies run off the same ingestion events. Every new event type needs an explicit in-or-out decision, or it silently becomes auto-reply eligible | S |

### Group B: Facebook only

**B-P0, finish the `feed` field we already receive.** No new subscription, no new permission.

| # | Gap | Why it matters | Effort |
|---|---|---|---|
| B1 | Comment deleted (`verb: remove`) is ignored | Deleted comments live in the inbox forever. Agents reply to content that no longer exists | S |
| B2 | Comment edited (`verb: edited`) is ignored | We show the original text. A customer who edits a complaint sees a reply to a version they retracted | S |
| B3 | Native hide and unhide are ignored | Someone moderates in Meta Business Suite, our state drifts, our hide button then fails or double-toggles | S |
| B4 | New post events are ignored | Posts only enter the inbox at the next sync, so early comments arrive with no parent post context. This is the visible symptom users complain about | M |
| B5 | Ad-post comment handling | Meta excludes ad posts from `feed` but **includes their comments**. We receive them, then fail to fetch the dark post as parent and log a warning. Dark-post comment moderation is a feature competitors gate behind higher tiers | M |
| B6 | `conversations` subscribed but unhandled | We pay the delivery cost and use none of it. Either handle it or drop it from the subscription | S |
| B7 | Post reactions (`item: reaction`) ignored | Counts come only from sync, so they are stale between runs. Live counters are a small visible quality win | M |

**B-P1, new subscriptions.** All covered by our existing scope string (`config/integrations.php:17` already requests `pages_read_engagement`, `pages_read_user_content`, `pages_manage_engagement`, `pages_manage_metadata`, `pages_messaging`, `pages_manage_ads`).

| # | Addition | Why it matters | Effort |
|---|---|---|---|
| B8 | **`ratings`, Page Recommendations** | We already built the `review` inbox type for GMB end to end: filter chip, list row, detail view, reply, auto-reply. Facebook recommendations drop into it with almost no new UI. **Cheapest customer-visible win on Facebook** | M |
| B9 | **`mention`, brand mentions** | Engagement we currently never see. The most common gap versus Sprout and Hootsuite. Also feeds Social Listening | L |
| B10 | `messaging_referrals` | Carries `ad_id` and `ref`. Label a DM "came from ad X" and route it. Agencies proving campaign ROI care, and nobody at our price point does it well | M |
| B11 | `message_echoes` | A reply sent from Meta Business Suite appears in our thread immediately. Prevents two agents answering the same customer | M |
| B12 | `message_reactions` | A thumbs-up on a resolution is an implicit close signal. Invisible today | S |
| B13 | `message_edits` | Same staleness class as comment edits, on the DM side | S |

**B-P2, reliability.** `messaging_policy_enforcement` (surface why a send was blocked, S), `messaging_handovers` plus `standby` (Pages also running ManyChat may never route to us with no diagnostic, M), `message_deliveries` plus `message_reads` (sent, delivered, read state, M), `inbox_labels` (mirror native labels into our tags, M).

**Facebook not now:** `leadgen`, `live_videos`, `videos`, `name`, `picture`, `category`, `page_change_proposal`, `page_upcoming_change`, `send_cart`, calling fields, `message_template_status_update`, `messaging_customer_information`, `business_integrity`, `group_feed`, `messaging_postbacks`, `messaging_optins`, `message_context`.

### Group C: Instagram only

| # | Gap | Why it matters | Effort |
|---|---|---|---|
| C1 | **`mentions` subscribed and dropped** | Subscription exists, permission exists, only the handler branch is missing. Highest value-to-effort ratio in this entire document | M |
| C2 | **Story mentions path appears broken** | Built consumer, no producer, hardcoded production URL, payload that fails its own schema. Either fix it or delete the dead path. Needs live confirmation first | M |
| C3 | `live_comments` | Comments during an Instagram Live. No Facebook equivalent, strong fit for the creator segment, and a genuine differentiator | M |
| C4 | `message_reactions` | Same value as Facebook B12 | S |
| C5 | `messaging_seen` | Instagram's read receipts | S |
| C6 | `message_edit`, `messaging_referral` | Singular field names on Instagram. Same value as the Facebook equivalents | S |
| C7 | Comment delete and edit detection | Instagram gives no webhook for these. If we want parity with the Facebook fix, it has to come from polling. **Scope this before promising parity** | L |
| C8 | Verify every change against both connection paths | Direct and Facebook-connected accounts use different base URLs, auth, and field sets. A fix verified on one may not hold on the other | S per change |

---

## Part 5: Suggested Sequencing

**Wave 1, trust.** A1 through A4, B1 through B4, B6. Nothing new appears in the UI. What is already there stops being wrong. Prerequisite for everything else and the fastest route to reducing the "inbox is out of sync" complaint volume.

**Wave 2, cheap wins already paid for.** C1 (Instagram mentions), B8 (Facebook Recommendations), B5 (ad-post comments), B12 and C4 (message reactions). Every item here is either already subscribed or lands in UI we have already built.

**Wave 3, differentiation.** B9 (Facebook mentions, pairing with C1 to give one cross-platform mentions surface), B10, B11, C3 (Instagram Live comments), B7.

**Wave 4, ops.** B-P2, C5, C6. Do this when inbox support load justifies it.

**Unscheduled pending investigation:** C2 (story mentions) and C7 (Instagram delete and edit detection).

Note that mentions arrive in two different waves by design: Instagram's is a handler branch on an existing subscription, Facebook's is a new subscription plus a likely new inbox type. Shipping Instagram's first is a cheap way to learn what the surface should look like before committing to the Facebook build.

---

## Part 6: Open Questions

1. **Mentions: new inbox type or new filter?** A mention is not a conversation, post, or review. It may warrant a fourth `inbox_type`, which touches the sidebar, counts, filters, the read model, and mobile. This needs answering before Wave 2, because C1 is the first thing that lands in it.
2. **Do we act on Recommendations, or only display them?** Meta lets a Page comment on a recommendation via its `open_graph_story_id`. If we ship reply, review auto-reply already exists and would extend to Facebook for free.
3. **Is the Instagram story mentions feature meant to be live?** If yes it needs fixing, if no the consumer, topic, and handler branch should be deleted rather than left as a trap.
4. **Volume and cost.** Facebook `feed` reactions and `message_deliveries` on a large Page are high-frequency, low-information events. Sample before subscribing at scale.
5. **Ad-comment scope.** Full dark-post moderation may deserve to be a paid tier rather than a bug fix. A pricing call, not an engineering one.
6. **Do we promise Instagram parity on deletes and edits?** Meta does not expose the events. Parity means polling, which is a materially bigger build than the Facebook fix.

---

## Part 7: Constraints to Plan Around

- Meta documents that **`feed` delivery is reduced for Pages above roughly 10K likes**. Real-time will never be complete for our largest Facebook accounts. The periodic sync stays permanently as reconciliation. Do not plan a "webhooks replace polling" migration.
- The `mention` docs warn that some returned post and comment IDs are **not queryable** with a Page token. Mention handling must degrade to the webhook payload rather than assume a follow-up fetch succeeds.
- Instagram's **two connection paths** double the verification surface for every Instagram change.
- Handling deletes changes our write pattern. Comment rows today are only ever added or updated. Removing them raises the question of what happens to assigned, tagged, or marked-done state attached to a comment that no longer exists.

---

## Evidence

| Claim | Location |
|---|---|
| Facebook subscribes to `feed,messages,conversations` | `contentstudio-backend/config/integrations.php:25` |
| Instagram subscribes to `comments,messages,mentions` | `contentstudio-backend/config/integrations.php:540` |
| Duplicate hardcoded Facebook field list on connect | `contentstudio-backend/app/Observers/Integrations/PlatformObserver.php:15` |
| Facebook handler routes messaging and `changes[0]` comments only | `social-inbox-manager/app/api/webhooks/facebook_webhook_actions.py:95` |
| Instagram handler has no `mentions` branch | `social-inbox-manager/app/api/webhooks/instagram_webhook_actions.py:118` |
| Instagram `story_insights` posts to a hardcoded production URL | `social-inbox-manager/app/api/webhooks/instagram_webhook_actions.py:137` |
| No producer writes to the story insights topic its worker consumes | `app/config/kafka_config.py:115`, consumed at `app/workers/instagram_inbox_worker.py:154` |
| Facebook worker loops all changes but never reads `verb` | `social-inbox-manager/app/workers/facebook_inbox_worker.py:318` |
| Deleted comment fetch no-ops silently | `social-inbox-manager/app/social_sync/facebook_strategy.py:669` |
| Missing parent post logs a warning and continues | `social-inbox-manager/app/social_sync/facebook_strategy.py:730` |
| Comment like is gated to Facebook only | `social-inbox-manager/app/api/v1/routes/comment_actions.py:107` |
| Instagram runs two connection paths with different field sets | `social-inbox-manager/app/utils/instagram_helper.py:62-109` |
| Graph field sets, Facebook | `social-inbox-manager/app/utils/facebook_helper.py:59` |
| DM send uses HUMAN_AGENT after ACCOUNT_UPDATE deprecation | `social-inbox-manager/app/utils/facebook_helper.py:1064` |
| Sync limits per run | `app/jobs/facebook_inbox_job.py:22`, `app/jobs/instagram_inbox_job.py:22` |
| Inbox types are conversation, post, review | `contentstudio-frontend/src/modules/inbox-revamp/composables/useInboxUI.ts:189` |
| Capability flags gate the comment action buttons | `contentstudio-frontend/src/modules/inbox-revamp/components/CommentBlock.vue:80` |
| Storage model and collections | `social-inbox-manager/docs/DATA_STORAGE.md` |

Related reading: [Inbox Module Architecture and Sync Findings](inbox-module-architecture-and-sync-findings-2026-03-24.md), which covers the read-model and realtime-contract problems this document assumes as given.
