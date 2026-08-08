# Epic: Inbox network research, Telegram, Bluesky and Threads

## Problem

Three networks we already publish to have no presence in Social Inbox: Telegram, Bluesky and Threads. All three were requested as **research first**, with development tickets to follow once feasibility and scope are confirmed.

Research genuinely is the right first step here, because none of the three resembles the networks our inbox was built around. Telegram is a bot-token API with a privacy mode that can hide group messages from us and a hard rule that a bot cannot start a conversation. Bluesky is an open protocol whose reply model is a thread graph rather than comments on a post. Threads sits behind Meta's review process. Writing dev stories against any of them today would mean writing them twice.

## Goal

Three scope documents, one per network, each concrete enough that a PO can write dev stories from it without a second investigation. Each answers what we can read, what we can do, how updates reach us, whether our existing connection suffices, whether AI auto-replies can work there, and what a realistic v1 looks like.

## Scope

Three research stories. No code ships from this epic.

## Rules

- **Answer the connection question first.** For each network, state whether the publishing connection we already have is sufficient for Inbox. That is the question that most changes the size of the work.
- **Be specific about what we cannot do.** A document that lists capabilities without listing the gaps produces an inbox that quietly does less than users expect. Every action a user would reasonably try and cannot must be named.
- **Cover AI auto-replies.** They are a paid feature and users will expect them to extend to any new inbox network. If they cannot work, say so and say why.
- **Fit the existing architecture, or say why not.** Our inbox is a per-network strategy pattern with webhook fan-out. If a network does not fit that, that is a finding, not a detail.
- **Follow the established document shape.** We have done this research three times recently; the TikTok inbox research is the cleanest template.

## Sequencing

Independent, can run in parallel. Telegram is the most tractable and has the most existing internal evidence, so it is a good first one if capacity is limited. Threads should start early if Meta review turns out to be on the critical path.

## Out of scope

- Any implementation.
- Social Listening coverage of these networks, which already exists for Bluesky and Threads and is a different thing (public mention monitoring, not owned-account conversations).
- Mobile. Whether a new network reaches the apps is decided after the scope documents exist.

## Stories

1. `[Research] Telegram Inbox feasibility and recommended scope`
2. `[Research] Bluesky comments and replies in Inbox: feasibility and recommended scope`
3. `[Research] Threads comments and replies in Inbox: feasibility and recommended scope`

---
---

# 1. [Research] Telegram Inbox feasibility and recommended scope

### Description

We publish to Telegram channels and groups, and users are asking whether Telegram conversations can come into Social Inbox. The specific question is whether the Telegram integration we already use for publishing can also serve Inbox, or whether Inbox needs something different.

There is more internal evidence for this one than the other two, and it points at a real constraint rather than an unknown: our integration uses the Bot API with a bot token, and a Telegram bot has a privacy mode that limits what it can see in groups and cannot start a conversation with a user who has not messaged it first. So the useful research is not "is it possible" but "exactly what can we see and do, given those constraints, and is the resulting inbox worth shipping".

The deliverable is a feasibility and scope document a PO can write dev stories from.

### Workflow

*(Research story. The deliverable is a document, not a user-facing change.)*

1. Establish what the Bot API lets a bot receive: message types, chat types, and events.
2. Establish what privacy mode changes, and whether we can turn it off for our bot or need users to configure something.
3. Establish what a bot can send, and to whom, given that it cannot initiate a conversation.
4. Determine whether our existing bot-token connection is sufficient, or whether Inbox needs a different bot configuration, a different API, or per-user bots.
5. Establish how updates arrive and how we avoid missing them.
6. Assess fit against our existing inbox architecture, and whether AI auto-replies can work.
7. Recommend a realistic v1 scope, then review with product before any dev story is written.

### Acceptance criteria

- [ ] A document exists stating which Telegram message and update types the Bot API delivers to us, across private chats, groups and channels.
- [ ] The document states clearly whether the Telegram integration we already use for publishing is sufficient for Inbox, or what would have to change.
- [ ] The effect of bot privacy mode is documented: what a bot in a group can and cannot see with it on, whether we can disable it for our bot, and whether that requires anything from the user.
- [ ] The consequence of the bot-cannot-initiate rule is documented, in user-facing terms: which conversations can appear in Inbox at all, and which are structurally unreachable.
- [ ] Whether channel comments and discussion-group messages are receivable is documented, since channels are our primary publishing target and their comment threads are the closest analogue to what users expect from Inbox.
- [ ] What we can send is documented: replies, media, reactions, edits, deletes, and whether a reply appears attributed to the brand rather than to a bot.
- [ ] Whether the reply appears to end users as coming from the brand or visibly from a bot is documented, because that materially affects whether the feature is usable for customer conversations.
- [ ] The update mechanism is documented, including whether webhooks or polling is appropriate for us, and how we guarantee we do not lose events during a deploy or an outage.
- [ ] Rate limits are documented, with an assessment of whether they support inbox-style reply volume rather than only scheduled publishing volume.
- [ ] Media support is documented: what media types arrive, size limits, and how we would store and display them.
- [ ] Whether one shared ContentStudio bot or a per-customer bot is the right model is assessed, including the white-label implication of a reply appearing to come from a bot named after us.
- [ ] Fit against our inbox architecture is assessed against the existing per-network strategy pattern and webhook fan-out, and anything that does not fit is named.
- [ ] Whether AI auto-replies can work on Telegram is assessed, including whether the initiation rule and privacy mode constrain them.
- [ ] The Zoho Social Telegram inbox implementation is reviewed as the one known competitor precedent, and what it does and does not support is recorded.
- [ ] The document ends with a recommended v1 scope, what is deferred, and what we should tell users is not possible and why.
- [ ] The document is reviewed with product, and the review outcome is recorded, before any development story is created.

### Mock-ups

N/A. Research deliverable. If the recommended scope needs an inbox conversation type our current UI cannot represent, the document should say so, so a design story can be raised.

### Impact on existing data

None. Research only.

### Impact on other products

- None directly. Informs future work in the inbox service, the backend integration, and the web app.
- If a per-customer bot model is recommended, that has an onboarding and white-label consequence the PO needs to know early.

### Dependencies

None. Can start immediately, and is the most tractable of the three.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, research
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A, research
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A, research
- [ ] White-label domains impact reviewed — the shared-bot versus per-customer-bot question is a white-label question and must be addressed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — the document should note whether the recommended scope has any mobile implication

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- **Start with our own code, not Telegram's docs.** `contentstudio-backend/app/Strategy/Integrations/TelegramConnector.php` already answers much of the connection question: it uses a bot token rather than OAuth (noted explicitly at line 54), and `validateToken()` calls `getMe()` and stores `can_join_groups` and `can_read_all_group_messages` (around lines 160 to 161). That second flag is Telegram's privacy-mode indicator and is close to the single most important input to this research.
- **Then read our existing Telegram integration research.** It already records several relevant constraints: we use the Bot API rather than MTProto, a bot must be a channel or group administrator to post, and its permissions table states that a bot can only message users who started a chat with it. It also notes Zoho Social shipped a Telegram inbox in 2024 without publishing, and records rate limits of roughly 1 message per second per chat, 20 per minute for groups, and 30 per second overall.
- The document shape to follow is the one used for the TikTok inbox comments research: what the feature is, platform API capabilities, competitor analysis, table stakes versus delighters, recommended v1, codebase analysis with integration points, and technical gotchas.
- The architecture to fit is `social-inbox-manager/app/social_sync/base_strategy.py` plus the six existing strategies (`facebook_strategy.py`, `instagram_strategy.py`, `twitter_strategy.py`, `linkedin_strategy.py`, `youtube_strategy.py`, `gmb_strategy.py`). Reading `base_strategy.py` tells you exactly what a new network has to provide.
- Inbound events land via `social-inbox-manager/app/api/webhooks/` and fan out through `app/kafka/` to `app/workers/`.
- Auto-replies are `social-inbox-manager/app/services/auto_reply/`, entry point `engine.py`.
- Stay consistent with the findings already recorded by the inbox module architecture sync review, the inbox stability hardening work, and the inbox webhook events work.

---
---

# 2. [Research] Bluesky comments and replies in Inbox: feasibility and recommended scope

### Description

We publish to Bluesky and offer no way to see or respond to the replies our users' posts receive. Before building anything, we need to know what Bluesky's APIs let us retrieve and act on.

Bluesky is unlike the networks our inbox was built for. It is an open protocol rather than a single platform API, replies form a thread graph rather than comments attached to a post, and accounts can live on different providers. Each of those is a potential mismatch with our per-network strategy pattern, and identifying them is a large part of this story's value.

The deliverable is a feasibility and scope document a PO can write dev stories from.

### Workflow

*(Research story. The deliverable is a document, not a user-facing change.)*

1. Establish what interaction types exist on Bluesky and which the APIs expose.
2. Establish what we can retrieve for a connected account's posts: replies, mentions, quotes, likes, reposts.
3. Establish what actions we can perform, and whether they are attributed to the user's account.
4. Establish how we learn about new interactions, and the latency and reliability of that mechanism.
5. Determine whether our existing publishing connection is sufficient.
6. Assess fit against our inbox architecture, given the thread-graph reply model and the open-protocol account model.
7. Assess whether AI auto-replies can work.
8. Recommend a realistic v1 scope, then review with product.

### Acceptance criteria

- [ ] A document exists inventorying Bluesky's interaction types and stating which the APIs expose to us for a connected account.
- [ ] What we can retrieve is documented: replies to the account's posts, mentions of the account, quote posts, likes and reposts, with how far back each goes and how complete it is.
- [ ] The reply model is documented in terms of how it differs from the comment-on-a-post model our existing strategies assume, and what that means for how a conversation would be presented in Inbox.
- [ ] What we can do is documented: replying, liking, reposting, deleting our own replies, and anything else. Actions users would expect and we cannot perform are named explicitly.
- [ ] Whether replies are attributed to the user's own account rather than an intermediary is confirmed.
- [ ] Whether direct messages exist on Bluesky and whether they are accessible to us is addressed, since users will ask.
- [ ] The update mechanism is documented: whether we can subscribe to events, must poll, or can use the protocol's firehose, with latency, reliability and the cost of each approach.
- [ ] Whether we can guarantee not missing interactions during a deploy or outage is addressed.
- [ ] Authentication requirements are documented, including whether the existing Bluesky publishing connection is sufficient or a new or expanded authorisation is needed, and whether existing connected accounts would need re-authorising.
- [ ] Rate limits are documented, with an assessment of whether they support inbox-style read and reply volume across our connected account base.
- [ ] The open-protocol implication is documented: whether accounts hosted on different providers are equally readable and writable, and whether any part of the recommended scope only works for accounts on the main provider.
- [ ] Media and rich-content support in replies is documented.
- [ ] Fit against our inbox architecture is assessed against the existing per-network strategy pattern and webhook fan-out, and anything that does not fit is named.
- [ ] Whether AI auto-replies can work on Bluesky is assessed.
- [ ] Competitor coverage is reviewed: which comparable tools support Bluesky in an inbox, and what they support.
- [ ] The document ends with a recommended v1 scope, what is deferred, and what we should tell users is not possible and why.
- [ ] The document is reviewed with product, and the review outcome is recorded, before any development story is created.

### Mock-ups

N/A. Research deliverable. If the thread-graph model needs an inbox presentation our current UI cannot represent, the document should say so, so a design story can be raised.

### Impact on existing data

None. Research only.

### Impact on other products

- None directly. Informs future work in the inbox service, the backend integration, and the web app.
- If existing connected accounts would need re-authorising, that has a user-facing consequence to flag early.

### Dependencies

None. Can start immediately.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, research
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A, research
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A, research
- [ ] White-label domains impact reviewed — N/A, research
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — the document should note whether the recommended scope has any mobile implication

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- Start with `contentstudio-backend/app/Strategy/Integrations/BlueskyConnector.php` to see exactly what the current connection obtains, and `app/Strategy/Planner/BlueskyPosting.php` for what we already do with it. That answers the connection-sufficiency question faster than the protocol docs.
- The architecture to fit is `social-inbox-manager/app/social_sync/base_strategy.py` and the six existing strategies. `twitter_strategy.py` is probably the closest existing analogue, since X's reply model is nearer to Bluesky's than Facebook's comment model is.
- Inbound events land via `social-inbox-manager/app/api/webhooks/` and fan out through `app/kafka/`. If Bluesky offers no webhook equivalent, the research needs to say what replaces that stage, because it is load-bearing in the current design.
- Auto-replies are `social-inbox-manager/app/services/auto_reply/`, entry point `engine.py`.
- The document shape to follow is the one used for the TikTok inbox comments research. The X social inbox integration is the closest precedent by reply model.
- Social Listening already monitors Bluesky as a public-mention source. That is not inbox access, but the researcher should check whether any of that access is reusable before concluding something is unavailable.
- Stay consistent with the findings already recorded by the inbox module architecture sync review, the inbox stability hardening work, and the inbox webhook events work.

---
---

# 3. [Research] Threads comments and replies in Inbox: feasibility and recommended scope

### Description

We publish to Threads and offer no way to see or respond to the replies our users' posts receive. Threads is a Meta API surface, so the questions that decide feasibility are Meta's: which permissions we need, whether they require app review, what webhooks exist, and what rate limits apply. We already run Facebook and Instagram inbox integrations through Meta APIs, so there is existing infrastructure and existing accumulated knowledge to check before treating this as new ground.

The deliverable is a feasibility and scope document a PO can write dev stories from.

### Workflow

*(Research story. The deliverable is a document, not a user-facing change.)*

1. Establish which Threads interaction types the APIs expose: replies, mentions, quotes, reposts.
2. Establish what we can retrieve for a connected account's posts, and how far back.
3. Establish what actions we can perform: reply, hide, delete, react.
4. Establish the permission model and whether Meta app review is required.
5. Establish the update mechanism, webhooks or otherwise, and its reliability.
6. Determine whether our existing Threads publishing connection is sufficient.
7. Assess whether the existing Meta inbox path can be extended rather than duplicated, and whether AI auto-replies can work.
8. Recommend a realistic v1 scope, then review with product.

### Acceptance criteria

- [ ] A document exists inventorying Threads interaction types and stating which the APIs expose to us for a connected account.
- [ ] What we can retrieve is documented: replies to the account's posts, mentions, quote posts and reposts, with how far back each goes and how complete it is.
- [ ] What we can do is documented: replying, hiding, deleting, reacting, and anything else. Actions users would expect and we cannot perform are named explicitly.
- [ ] Whether replies are attributed to the user's own Threads account is confirmed.
- [ ] Required permissions and scopes are documented, along with whether they need Meta app review, what that review requires from us, and a realistic expectation of how long it takes.
- [ ] Whether the existing Threads publishing connection is sufficient is stated explicitly, along with whether existing connected accounts would need re-authorising.
- [ ] Whether the existing Meta inbox integration path can be extended for Threads, rather than a new strategy built from scratch, is assessed. This materially affects the size of the dev work.
- [ ] Webhook availability is documented: which events Meta pushes for Threads, whether they cover everything we need, and what has to be polled if they do not.
- [ ] Webhook reliability is addressed, including whether the known coverage gaps in our existing Meta webhook handling apply here too.
- [ ] Rate limits are documented, with an assessment of whether they support inbox-style read and reply volume.
- [ ] Account-type constraints are documented: whether inbox access is available for all Threads account types we can connect, or only some.
- [ ] Whether Threads direct messages exist and are accessible is addressed, since users will ask.
- [ ] Media support in replies is documented.
- [ ] Fit against our inbox architecture is assessed, and anything that does not fit the per-network strategy pattern is named.
- [ ] Whether AI auto-replies can work on Threads is assessed.
- [ ] Competitor coverage is reviewed: which comparable tools support Threads in an inbox, and what they support.
- [ ] The document ends with a recommended v1 scope, what is deferred, and what we should tell users is not possible and why.
- [ ] Any dependency on Meta approval that could block delivery is flagged as a schedule risk, not buried in the constraints section.
- [ ] The document is reviewed with product, and the review outcome is recorded, before any development story is created.

### Mock-ups

N/A. Research deliverable. If the recommended scope needs an inbox conversation type our current UI cannot represent, the document should say so, so a design story can be raised.

### Impact on existing data

None. Research only.

### Impact on other products

- None directly. Informs future work in the inbox service, the backend integration, and the web app.
- If Meta app review is required, that has a delivery-timeline consequence the PO needs to know early.

### Dependencies

None. Can start immediately, and should start early if Meta review turns out to be on the critical path.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, research
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A, research
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A, research
- [ ] White-label domains impact reviewed — N/A, research
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — the document should note whether the recommended scope has any mobile implication

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- Start with `contentstudio-backend/app/Strategy/Integrations/ThreadsConnector.php` to see what the current connection obtains, and `app/Strategy/Planner/ThreadsPosting.php` for what we already do with it.
- The highest-value question is whether the existing Meta path extends. `social-inbox-manager/app/social_sync/facebook_strategy.py` and `instagram_strategy.py` are the closest existing implementations, and `base_strategy.py` defines what a strategy has to provide. If Threads slots into the Instagram path, the dev work shrinks considerably.
- Our accumulated Meta webhook knowledge is directly relevant and some of it is already written down: the Meta inbox webhook coverage review records where our Meta webhook coverage has gaps, and the inbox webhook events work covers the event surface. Check both before assuming Threads webhooks will behave.
- Inbound events land via `social-inbox-manager/app/api/webhooks/` and fan out through `app/kafka/`.
- Auto-replies are `social-inbox-manager/app/services/auto_reply/`, entry point `engine.py`.
- The document shape to follow is the one used for the TikTok inbox comments research. The WhatsApp inbox integration is the closest precedent for a Meta-permission-gated inbox network, including how its rollout was staged behind a feature flag.
- Social Listening already monitors Threads as a public-mention source. Worth checking for reusable access.
- Stay consistent with the findings already recorded by the inbox module architecture sync review and the inbox stability hardening work.
