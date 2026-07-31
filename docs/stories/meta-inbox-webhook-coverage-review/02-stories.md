# Meta Inbox Webhook Coverage Review — Stories

Date: July 29, 2026

Source doc (shared with dev): https://app.helpin.ai/share/8de4c73e2459d28b6e7ba050cc4d8e82

> **Note:** the doc was extended on July 29 to cover Instagram as well as Facebook. If re-uploading minted a new Helpin link, replace the three references below before the story is created.

---

## Epic

**Title:** Meta Inbox Webhook Coverage

**Summary:** Handle the Meta webhook events we already receive but discard, then expand coverage across Facebook Pages and Instagram.

**Description:**

ContentStudio's inbox ingests Facebook Page and Instagram activity through Meta's webhooks, but on both platforms we subscribe to fields we never handle. Facebook subscribes to `feed`, `messages`, and `conversations` and acts on two event shapes. Instagram subscribes to `comments`, `messages`, and `mentions` and also acts on two. Events Meta already delivers, and that we already have permission to receive, are discarded at the API layer.

The visible cost is inbox state that drifts from reality. A comment deleted on Facebook stays in our inbox indefinitely, so an agent can reply to content that no longer exists. A comment edited on Facebook still shows its original text. A comment hidden natively in Meta Business Suite leaves our hide button out of sync. Every Instagram mention of a customer's brand is received and silently dropped.

This epic covers the ingestion layer for both platforms: fixing the events we drop, subscribing to the events worth adding, and building the backfill needed to apply new subscriptions to accounts that are already connected. Facebook Pages and Instagram are separate webhook objects with different field catalogs, so the two platforms carry different work lists rather than one list applied twice. Several capabilities exist on only one of them.

Nearly all of this is reachable without new permissions. The OAuth scopes we already request cover every proposed addition, so there is no new App Review and no new consent screen for existing customers.

Scope note: user-facing surfaces that sit on top of this ingestion layer, specifically Facebook Page Recommendations in the inbox and a brand mentions surface, are tracked as separate epics because they need Design, frontend, and mobile work. This epic delivers the plumbing they depend on. All work here is gated on the review story completing first, since the audit's findings need engineering confirmation before anything is scheduled.

---

## Stories

---

## [BE] Review Meta inbox webhook coverage audit and validate proposed priorities

### Description:

As the engineer who owns the inbox ingestion pipeline, I want to review the Meta webhook coverage audit and give a written technical verdict on its findings, so that we commit to the right sequence of work before anyone starts building and we do not spend a sprint on gaps that turn out to be wrong, unaffordable, or unavailable from Meta.

An audit of our Facebook Page and Instagram inbox ingestion was completed on July 29, 2026. Its central claim is that on both platforms we subscribe to webhook fields we never handle, so events Meta already delivers are discarded at the API layer. Specifically it claims Facebook subscribes to `conversations` and ignores it, and Instagram subscribes to `mentions` and drops every one.

It also claims that Facebook and Instagram are **separate webhook objects with different field catalogs**, so several capabilities we would want on both exist only on one. If that is right, it changes what we can promise for Instagram.

All of this needs a second pair of eyes from someone who can run the code and check real payloads.

**Audit doc:** https://app.helpin.ai/share/8de4c73e2459d28b6e7ba050cc4d8e82

This story is a research and review task. No production code changes are expected from it. The output is a written response that we turn into implementation stories.

---

### Workflow:

1. Developer opens the shared audit doc and reads it end to end.
2. Developer works through Part 1 and checks each current-state claim against the running code and, where possible, against real webhook payloads from a test Facebook Page and a test Instagram account.
3. Developer confirms or refutes each numbered gap in Part 4, across all three groups, with a short note on what they actually observed.
4. Developer investigates the Instagram story mentions path described in section 1.3, which the audit believes is broken, and states whether it works in production today.
5. Developer checks the permissions claim: for each proposed new field on each platform, confirm whether our existing OAuth scopes are sufficient or whether App Review is genuinely required.
6. Developer assesses expected event volume for the noisier fields and flags any that would be too costly to subscribe at scale.
7. Developer notes anything the audit got wrong, overstated, or missed entirely, on either platform.
8. Developer gives an effort size per item and proposes a revised sequence if they disagree with the four waves in Part 5.
9. Developer answers the six open questions in Part 6, with the mentions inbox type question as the one that most needs an answer before design work starts.
10. Developer posts the response on this story.

---

### Acceptance criteria:

**Current state, both platforms**

- [ ] Each numbered gap in Part 4 of the audit has an explicit verdict of confirmed, partly confirmed, or not a real defect, with a one line reason
- [ ] The claim that only the first change in an entry is inspected is verified against a real batched payload on both Facebook and Instagram, and the response states whether comments are actually being lost today
- [ ] The claim that Instagram subscribes to `mentions` and drops every one is confirmed or refuted, and the response states what currently happens to an Instagram mention end to end
- [ ] The claim that a deleted Facebook comment stays in the inbox permanently is verified end to end, and the response states what a user currently sees after a comment is deleted on Facebook
- [ ] The response states whether the Facebook `conversations` field is genuinely unused, and recommends either handling it or removing it from the subscription

**Instagram story mentions**

- [ ] The response states whether Instagram story mentions work in production today
- [ ] The response confirms or refutes that no producer writes to the story insights topic its worker consumes
- [ ] The response explains what the hardcoded production URL in the story insights branch is for, and whether non-production environments are posting to production
- [ ] The response recommends either fixing the path or deleting the unused consumer, topic, and handler branch

**Feasibility and cost**

- [ ] For each proposed new field on each platform, the response confirms whether our current OAuth scopes cover it or whether App Review is required
- [ ] The response confirms whether existing connected accounts need a backfill subscription call, describes the shape of that job for both platforms, and states how many accounts are affected and what happens on failure
- [ ] The response gives an expected event volume assessment for the high frequency fields, at minimum Facebook `feed` reactions and `message_deliveries`, and recommends whether to subscribe, sample first, or skip
- [ ] The response confirms whether comments on Facebook ad posts and dark posts reach us today, and what currently happens when the parent post cannot be fetched
- [ ] The response confirms whether Instagram genuinely has no webhook for comment deletes and edits, and if so what polling would cost to reach parity with the Facebook fix
- [ ] The response confirms whether any Instagram change needs separate verification against both the direct and Facebook-connected connection paths

**Planning output**

- [ ] The response includes an effort size of small, medium, or large per item
- [ ] The response either endorses the four wave sequencing in Part 5 or proposes a different one with reasoning
- [ ] The response gives a technical recommendation on whether brand mentions should become a fourth inbox type or be handled as a filter on an existing type, and names what that choice touches in the read model, sidebar counts, filters, and mobile
- [ ] The response states which new event types should and should not be eligible for rule-based and AI auto-replies
- [ ] Any claim in the audit found to be wrong or overstated is called out explicitly so the doc can be corrected

---

### Mock-ups:

N/A. Research and review story with no UI.

---

### Impact on existing data:

None from this story itself. No schema changes and no writes.

Worth noting for the response: if the recommendations are accepted, the implementation that follows will change existing data. Handling comment deletes and edits means our stored comment rows start being removed or rewritten where today they are only ever added or updated. The reviewer should call out any risk they see in that, including whether removing a comment row breaks assigned, tagged, or marked done state attached to it.

---

### Impact on other products:

- **Mobile apps.** Mobile reads the same inbox API. A new inbox type for mentions would need matching mobile work, which is why that question in the acceptance criteria matters beyond web.
- **Auto replies.** Rule based and AI auto replies run off the same ingestion events. Every new event type needs an explicit in-or-out decision, or it silently becomes auto-reply eligible.
- **Social Listening.** Brand mentions overlap with the listening product. The reviewer should flag any duplication they see between the two ingestion paths.

---

### Dependencies:

- Access to the shared audit doc: https://app.helpin.ai/share/8de4c73e2459d28b6e7ba050cc4d8e82
- A test Facebook Page connected to a test workspace, with the ability to trigger comments, edits, deletes, hides, and reactions and observe the resulting webhook payloads
- A test Instagram business account on **both** connection paths, direct and Facebook-connected, with the ability to trigger comments, mentions, and story mentions
- Blocks: this review must be completed before any implementation story from the audit is written or scheduled

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories) — N/A, research story with no UI
- [ ] Multilingual support (frontend + backend, translations available or fallback handled) — N/A, no user facing strings
- [ ] UI theming support (default + white-label, design library components are being used) — N/A, no UI
- [ ] White-label domains impact review — N/A for the review itself
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Webhook receivers, where the routing gaps live:**

- `social-inbox-manager/app/api/webhooks/facebook_webhook_actions.py` — routes `messaging` and a first-change comment only
- `social-inbox-manager/app/api/webhooks/instagram_webhook_actions.py` — routes `messaging`, a first-change `comments` field, and a `story_insights` branch that posts to a hardcoded production URL. No `mentions` branch

**Workers and strategies:**

- `social-inbox-manager/app/workers/facebook_inbox_worker.py` and `instagram_inbox_worker.py` — both loop over all changes, so the first-change limitation is at the API layer only
- `social-inbox-manager/app/social_sync/facebook_strategy.py` — `process_webhook_comments` is where a fetch for a deleted comment returns nothing and the flow no-ops, and where a missing parent post logs a warning and continues
- `social-inbox-manager/app/utils/instagram_helper.py` — the two connection paths with different base URLs, auth, and field sets are set up in the constructor
- `social-inbox-manager/app/config/kafka_config.py` — the story insights topic and its consumer group are defined here. Research found no producer for it anywhere in the codebase

**Subscription and permissions:**

- `contentstudio-backend/config/integrations.php` — both OAuth scope strings and both `webhook_subscribed_fields` values
- `contentstudio-backend/app/Observers/Integrations/PlatformObserver.php` — the Facebook subscription call at connect time, plus a second hardcoded copy of the field list that can drift from config
- `contentstudio-backend/app/Helpers/Integrations/InstagramHelper.php` — the Instagram subscription path, which differs from Facebook's

**Frontend, relevant to the mentions inbox type question:**

- `contentstudio-frontend/src/modules/inbox-revamp/composables/useInboxUI.ts` — where the three inbox types are defined
- `contentstudio-frontend/src/modules/inbox-revamp/components/CommentBlock.vue` — where Graph capability flags gate the comment action buttons

**Existing behavior the audit found is already correct, so no change needed:**

- Direct message sending on both platforms already uses the `HUMAN_AGENT` tag after Meta deprecated `ACCOUNT_UPDATE` in April 2026, preserving the seven day reply window
- Both helpers split message fields into core and optional tiers, so a permission-gated field degrades gracefully instead of failing the whole call
- Comment liking is gated to Facebook only, which is correct rather than an oversight, because Meta does not expose it for Instagram
- Graph capability flags are correctly wired through to the UI and gate the hide, like, delete, and reply privately buttons

**Gotchas worth confirming:**

- Facebook's subscribed field list exists in two places and they can drift apart
- Meta documents that Facebook `feed` delivery is reduced for Pages above roughly ten thousand likes. Real time will never be complete for our largest accounts, so the periodic sync stays permanently as reconciliation. Sanity check any plan that assumes webhooks can replace polling
- Meta warns that some post and comment IDs returned by the Facebook `mention` field are not queryable with a Page token. Mention handling would need to work from the webhook payload rather than assume a follow up fetch succeeds
- Meta excludes ad posts from the Facebook `feed` field but includes comments on them. The audit reads this as an opportunity rather than a bug
- Instagram field names differ from their Facebook counterparts in ways that are easy to mistype: `message_edit`, `messaging_referral`, and `messaging_handover` are singular on Instagram and plural on Facebook
