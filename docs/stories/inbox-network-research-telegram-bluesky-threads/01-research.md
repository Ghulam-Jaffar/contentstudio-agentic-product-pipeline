# Research — Inbox network expansion: Telegram, Bluesky and Threads

Items 26, 27 and 28 of the 7 Aug 2026 backlog batch. All three were requested as **research tickets first**, with development tickets to follow once feasibility and scope are confirmed. This document is the pre-work that scopes those research stories.

## Current state

### What Social Inbox supports today

`social-inbox-manager/app/social_sync/` holds one strategy per supported network, on a shared base:

`base_strategy.py`, `facebook_strategy.py`, `instagram_strategy.py`, `twitter_strategy.py`, `linkedin_strategy.py`, `youtube_strategy.py`, `gmb_strategy.py`

None of Telegram, Bluesky or Threads is present.

The service architecture the research has to fit into: `app/api/webhooks/` for inbound platform events, `app/kafka/` for fan-out, `app/workers/` for per-platform processing, `app/database/mongo/` and `app/database/redis/`, `app/services/auto_reply/` for AI auto-replies, and realtime UI updates pushed to the frontend. Frontend consumer is `contentstudio-frontend/src/modules/inbox-revamp/`.

### All three already connect for publishing

- **Telegram** — `contentstudio-backend/app/Strategy/Integrations/TelegramConnector.php`, `app/Strategy/Planner/TelegramPosting.php`
- **Bluesky** — `app/Strategy/Integrations/BlueskyConnector.php`, `app/Strategy/Planner/BlueskyPosting.php`
- **Threads** — `app/Strategy/Integrations/ThreadsConnector.php`, `app/Strategy/Planner/ThreadsPosting.php`

### Telegram: the connection question is already half answered

Item 26 asks specifically whether the Telegram integration we use for publishing can also serve Inbox. The code gives a strong starting answer.

`TelegramConnector.php` uses the **Bot API** with a bot token, not OAuth (line 54 notes "OAuth is not applicable for Telegram bot token flow"). Its `validateToken()` calls `getMe()` and already stores two flags that are precisely the ones that determine Inbox feasibility:

- `can_join_groups`
- `can_read_all_group_messages`

`can_read_all_group_messages` is Telegram's privacy-mode flag. With privacy mode on, a bot in a group receives only commands and replies directed at it, not all messages. That single flag largely determines whether a Telegram Inbox can see group conversation at all.

The existing feature research at `docs/features/telegram-integration/01-research.md` adds more:

- We integrate via the **Bot API**, not MTProto. A bot must be an administrator of a channel or group to post.
- Line 91's permissions table already records the constraint that matters most for Inbox: **"Bot can only message users who started a chat with the bot."** Telegram does not allow a bot to initiate a conversation.
- Line 46 notes **Zoho Social added Telegram to its unified inbox in 2024** for message management and responses, without publishing. So a competitor has shipped this, which is useful evidence that it is viable and worth studying.
- Line 127 notes Telegram exposes no per-post view or subscriber analytics via the Bot API, which is an analytics limitation rather than an inbox one, but confirms the Bot API's surface is narrower than a typical platform API.
- Rate limits recorded: roughly 1 message per second per chat, 20 per minute for groups, 30 per second across all chats.

So the Telegram research is less "is it possible" and more "what exactly can we see and do, given privacy mode, admin requirements, and the bot-cannot-initiate rule". That is a narrower and more answerable question than the other two.

### Bluesky and Threads

Neither has an inbox precedent in our codebase. Both need the full feasibility question answered: what interactions exist, what can be fetched, what actions we can take, and how updates arrive.

Note that "comments and replies" means different things on these networks than on Facebook. Bluesky's reply model is a thread graph over an open protocol. Threads is a Meta surface with its own reply model. Neither maps cleanly onto the comment-on-a-post model our Facebook and Instagram strategies were built around, which is itself a finding the research needs to surface.

## The templates to follow

We have done this exact kind of research three times recently, and the research stories should produce a document of the same shape:

- `docs/features/tiktok-inbox-comments/` — `01-research.md` has the cleanest structure: what the feature is, platform API capabilities, competitor analysis, table-stakes versus delighters, recommended v1 approach, codebase analysis with integration points, and technical gotchas.
- `docs/features/x-social-inbox-integration/`
- `docs/features/whatsapp-inbox-integration/`, with `docs/stories/whatsapp-inbox-feature-flag-rollout/` showing how the rollout followed.

Also relevant: `docs/stories/inbox-module-architecture-sync-findings/`, `docs/stories/inbox-stability-hardening/`, `docs/stories/meta-inbox-webhook-coverage-review/` and `docs/stories/inbox-webhook-events/`, all of which describe the current inbox architecture's known constraints. A new network's research should be consistent with those findings rather than assuming a clean slate.

## What each research story must produce

Same deliverable shape for all three, differing by network:

1. **Interaction inventory.** What conversation, message, comment or reply types exist on this network, and which of them the API exposes to us.
2. **Read capability.** What we can fetch, how far back, and with what completeness.
3. **Write capability.** What actions we can perform: reply, react, hide, delete, mark as read, assign. Anything we cannot do that users expect.
4. **Update mechanism.** Webhooks, polling, or a subscription. Latency, reliability, and whether we can guarantee we do not miss events.
5. **Authentication and permissions.** Whether our existing publishing connection is sufficient, or whether a new or expanded authorisation, an app review, or a different API entirely is required.
6. **Constraints.** Rate limits, message-window restrictions, media support, retention, and anything that would make the inbox incomplete in a way users would notice.
7. **Fit against our inbox architecture.** Whether the network fits the existing per-network strategy pattern, and what does not fit.
8. **Auto-reply applicability.** Whether AI auto-replies can work on this network, since that is a paid feature users will expect to extend.
9. **Recommended scope.** What a realistic v1 supports, what is deferred, and what we should tell users we cannot do.

## Why three separate stories

Three different protocols, three different access models, three different answers. Telegram is a bot-token API with privacy-mode and initiation constraints. Bluesky is an open protocol. Threads is behind Meta review. A combined document would be vague about all three. They are grouped into one epic for scheduling, not merged.

## Systems the research will touch

- `social-inbox-manager/app/social_sync/base_strategy.py` and the six existing strategies, to establish what a new strategy has to provide
- `social-inbox-manager/app/api/webhooks/`, `app/kafka/`, `app/workers/`
- `social-inbox-manager/app/services/auto_reply/`
- `contentstudio-backend/app/Strategy/Integrations/{Telegram,Bluesky,Threads}Connector.php`
- `contentstudio-frontend/src/modules/inbox-revamp/`

No code changes ship from any of the three stories.

## Mobile

The mobile apps have their own inbox screens. Whether a new network reaches them is decided after the scope documents exist, and would be a separate `[Flutter]` story.
