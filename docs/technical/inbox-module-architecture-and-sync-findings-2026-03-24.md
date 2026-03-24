# Inbox Module Architecture and Sync Findings

Date: March 24, 2026

Audience: CTO, CEO, Engineering Leadership

Scope: Static code analysis of the inbox stack across `contentstudio-frontend`, `contentstudio-backend`, and `social-inbox-manager`, with emphasis on architecture, data flow, sync behavior, realtime behavior, and failure modes.

## Executive Summary

The current inbox stack is not failing because of one isolated bug. It is failing because the system is built around a fragmented data model, best-effort realtime updates, and frontend reconciliation logic that depends heavily on timing.

At a high level:

- The frontend inbox revamp talks directly to a separate inbox service for most inbox reads and writes.
- Laravel is still involved, but mainly for Pusher auth and adjacent inbox features such as auto-replies.
- The separate inbox service stores conversations, messages, and comments in different collections, then reconstructs a unified inbox view dynamically at read time.
- Realtime updates are not authoritative. Many push events are only notifications that force the frontend to refetch data and hope the read model is ready.
- Frontend state for the inbox is implemented as shared module-level mutable state, which makes race conditions and stale UI state likely.

The result is a system that behaves like eventual consistency without the controls normally required to make eventual consistency reliable. This is why users experience:

- new data not appearing quickly or consistently
- list items and sidebar counts drifting apart
- send/reply actions appearing delayed
- state transitions not reflecting immediately
- intermittent `undefined` payloads in the UI
- different behavior depending on route, filter state, or timing

## Bottom Line

The inbox needs to be treated as a platform reliability problem, not a UI polish task.

The highest-leverage issue is architectural: there is no single canonical read model for the inbox list. Until that is corrected, frontend fixes alone will keep reducing symptoms without removing the source of drift.

## Current Architecture

### 1. Frontend

The revamp inbox UI in `contentstudio-frontend` uses `VITE_INBOX_API_URL` and sends most inbox requests directly to the separate inbox service, not to Laravel.

Responsibilities observed in the frontend:

- load sidebar counts and filtered inbox rows
- open conversations and post threads
- send messages, comments, and state-change actions
- subscribe to Pusher events
- reconcile partial realtime events into current UI state

### 2. Laravel Backend

Laravel is not the main inbox data service for the revamp module. Its current role is mostly:

- Pusher auth
- internal inbox-adjacent routes
- auto-reply and brand-help-doc APIs

This means Laravel is part of the operational chain, but it is not the primary source of the sync and state-consistency issues.

### 3. Social Inbox Manager Service

`social-inbox-manager` is the real inbox backend for sync, persistence, and most inbox APIs.

Observed responsibilities:

- ingest webhooks
- trigger manual sync jobs
- publish and consume Kafka messages
- persist inbox entities in MongoDB
- expose inbox APIs for list/details/messages/comments/actions
- emit Pusher notifications

### 4. Data Storage Shape

The service persists inbox data in a split model:

- `inbox_details`: conversations, reviews, posts
- `inbox_messages`: conversation messages and action messages
- `inbox_comments`: comments and replies

The inbox list shown in the UI is not backed by one canonical denormalized record. Instead, the service reconstructs a mixed list of conversations, reviews, posts, and comments during read requests.

That design is the core source of drift.

## Main Findings

### 1. The inbox list is assembled dynamically from multiple stores

The primary list endpoint fetches records from multiple Mongo collections, transforms comment documents into pseudo-inbox rows, adds reply trees, applies search and action filters, then sorts and paginates in application code.

Implications:

- list freshness depends on multiple write paths completing correctly
- query latency increases as data volume grows
- counts and rows can diverge
- ordering becomes unstable when some writes update `updated_at` and others do not
- the same entity can be represented differently depending on which path produced it

This is not a normal read path for a high-volume inbox. It is expensive and structurally hard to keep correct.

### 2. Realtime updates are notifications, not authoritative state

The `new_element` event sent to the frontend is intentionally lightweight. The frontend receives it and performs a targeted refetch to reconstruct the actual row.

Implications:

- realtime depends on the DB write already being visible
- realtime depends on the current filter state accepting the row
- if the row does not match current route conditions at the moment of the push, it may not be rendered
- if the materialized row is not available yet, the push appears to do nothing

In practice, this is not true realtime. It is notify-then-requery.

### 3. Frontend inbox state is globally shared mutable state

The inbox composable uses module-level refs for key state such as:

- selected channels
- selected filters
- conversation list
- current conversation
- Pusher instance and channel

This behaves like a process-wide singleton inside the SPA. That makes state leakage between route changes, workspace changes, remounts, and asynchronous requests very likely.

Implications:

- stale rows can survive longer than expected
- a late response can overwrite fresher UI state
- Pusher updates can reconcile against the wrong in-memory state
- the same bug can appear intermittent because timing changes behavior

### 4. Mutation endpoints often return stale or non-canonical data

Several action endpoints read records first, update them in Mongo, then return transformed copies of the pre-update records instead of refetching the updated entities.

Implications:

- the UI can receive old values after a successful action
- frontend code must rely on optimistic local mutation or follow-up refetches
- users perceive actions as delayed or inconsistent

This is especially damaging in a system that already depends on async reconciliation.

### 5. Status handling is inconsistent across entity types

The codebase uses different status representations for comments versus conversations and posts. For example:

- top-level comment read state
- nested `comments_details`
- transformed `comments_detail.read_status`
- nested action state inside `inbox_details`

Implications:

- read/unread behavior is easier to desynchronize
- one endpoint can update a field that another endpoint does not read as the source of truth
- derived list rows can diverge from stored documents

### 6. Sidebar counts and list results are derived differently

Sidebar counts are computed through a different backend path from the main inbox list. Because the underlying logic is not unified, counts can disagree with what is visible in the list.

Implications:

- trust in inbox metrics erodes
- user actions appear to fail when badges do not clear
- support diagnosis becomes harder because both views may be technically consistent with their own logic while still disagreeing with each other

### 7. `undefined` payloads are a real contract problem

The frontend store expects nested response shapes such as `message.data.payload` in some send paths. The service returns helper responses that are not strongly normalized. There are also frontend branches that do not return a safe default on unexpected non-200 flows.

Implications:

- UI code can receive `undefined` despite the request succeeding partially
- send/reply flows can fail silently or inconsistently
- different platforms can behave differently if helper response shapes vary

This is not only a UI issue. It is an API contract problem.

### 8. Outbound message/comment persistence is not authoritative enough

The send endpoints update action-related state, but the code path suggests the canonical sent entity often depends on webhook echo or later sync to appear in the inbox data model.

Implications:

- users can send a message and not see it appear immediately
- success responses and visible state can drift apart
- network or webhook lag turns into perceived product unreliability

This is one of the clearest causes of "I sent something, but it did not show up."

### 9. Realtime event delivery is implemented in multiple inconsistent ways

Some events are sent directly to Pusher from repositories. Other events, especially Facebook comments, are sent to Kafka and then emitted by a batch Pusher worker.

Implications:

- event timing differs by entity type and platform
- ordering is not guaranteed across all inbox activity
- reasoning about event arrival becomes difficult during debugging

Mixed delivery models are manageable only with strict sequencing and versioning. That control is not visible in the current implementation.

### 10. There are concrete sync-worker bugs, not just systemic weakness

The analysis found at least two concrete implementation issues:

- Instagram sync uses shared mutable metadata objects inside threadpool workers, which can leak conversation or post identifiers across concurrent tasks.
- The manual sync trigger path contains a clear repository mismatch in the YouTube branch, updating `GmbAccountRepository`.

Implications:

- some sync errors are deterministic bugs
- platform-specific correctness can degrade independently of the general architecture

## Root Causes

The major root causes can be summarized as follows:

### 1. No canonical inbox read model

The UI consumes a stitched view instead of a purpose-built, denormalized list model.

### 2. Best-effort eventual consistency without safety rails

The system behaves asynchronously but lacks the versioning, materialization guarantees, and clear source-of-truth contracts needed to make async state reliable.

### 3. Frontend reconciliation is too stateful and timing-sensitive

The frontend does too much repair work on top of incomplete events and mutable shared state.

### 4. API contracts are weak

The response model is not normalized tightly enough for a complex multi-platform inbox.

### 5. Operational paths differ by platform and entity type

Comments, messages, posts, and conversations do not move through one coherent sync and notification model.

## Business Impact

If left as-is, the current architecture will continue to create:

- unreliable user trust in the inbox
- support overhead around "missing" or "late" data
- slower feature development because each new feature must work around unstable state
- higher regression risk when introducing AI replies, auto-replies, or new platform integrations
- scaling risk as inbox volume grows

This is not only a technical debt problem. It affects product credibility and expansion velocity.

## Recommended Plan

### Phase 1. Stabilize Correctness

Goal: stop obvious drift and contract failures.

Recommended actions:

- refetch and return canonical post-write entities from every mutation endpoint
- update `updated_at` consistently for every state-changing action
- normalize response contracts for send/comment/message endpoints
- remove unsafe frontend assumptions about nested response shapes
- fix concrete sync-worker bugs in Instagram concurrency and the YouTube trigger path

Expected outcome:

- fewer `undefined` states
- fewer delayed state transitions
- more predictable action results

### Phase 2. Stabilize Frontend State

Goal: remove timing-sensitive state leakage.

Recommended actions:

- move inbox state out of module-level singleton refs into store-scoped or view-scoped state
- introduce request-generation guards so late responses cannot overwrite fresher state
- tighten Pusher lifecycle management around workspace and route changes

Expected outcome:

- fewer stale rows
- fewer intermittent UI-only inconsistencies
- easier debugging

### Phase 3. Redesign Realtime Contract

Goal: make realtime reliable instead of advisory.

Recommended actions:

- stop relying on lightweight push plus blind refetch for correctness
- emit authoritative row payloads or versioned entity events
- unify Pusher delivery strategy instead of mixing direct repository sends and Kafka-batched sends without sequencing guarantees

Expected outcome:

- lower time-to-visibility for new activity
- fewer "realtime but not really" failures

### Phase 4. Build a Canonical Inbox Read Model

Goal: remove the main architectural source of drift.

Recommended actions:

- create a dedicated denormalized inbox-list document per visible row
- use the same read model for both sidebar counts and list rendering
- keep comments, messages, and platform-specific payloads as supporting detail records, not as the list model itself

Expected outcome:

- lower query cost
- lower reconciliation complexity
- far more reliable filtering, ordering, and counts

## Suggested Success Metrics

To confirm improvement, the team should track:

- p95 time from inbound webhook or sync event to visible inbox row
- p95 time from outbound send action to visible sent item
- list/sidebar count drift rate
- percentage of inbox API responses with missing required fields
- frontend error rate for undefined/null message or comment payloads
- sync reconciliation failures by platform

## Representative Evidence

The findings above are based on the following code paths.

### Frontend

- `contentstudio-frontend/src/config/api-utils.js`
- `contentstudio-frontend/src/stores/inbox/useInboxRevampStore.ts`
- `contentstudio-frontend/src/modules/inbox-revamp/composables/useInbox.js`
- `contentstudio-frontend/src/modules/inbox-revamp/views/InboxView.vue`
- `contentstudio-frontend/src/modules/inbox-revamp/components/PostView.vue`
- `contentstudio-frontend/src/modules/inbox-revamp/components/MessageComposer.vue`
- `contentstudio-frontend/src/modules/common/lib/pusher.js`

### Laravel

- `contentstudio-backend/routes/web.php`
- `contentstudio-backend/app/Libraries/Inbox/InboxPusherBroadcast.php`
- `contentstudio-backend/routes/web/inbox.php`

### Social Inbox Manager

- `social-inbox-manager/app/api/main.py`
- `social-inbox-manager/app/api/routes/inbox.py`
- `social-inbox-manager/app/database/mongo/repository/inbox_details_repository.py`
- `social-inbox-manager/app/database/mongo/repository/inbox_messages_repository.py`
- `social-inbox-manager/app/database/mongo/repository/inbox_comments_repository.py`
- `social-inbox-manager/app/social_sync/facebook_strategy.py`
- `social-inbox-manager/app/social_sync/instagram_strategy.py`
- `social-inbox-manager/app/workers/pusher_notification_worker.py`
- `social-inbox-manager/docs/ARCHITECTURE.md`

## Notes and Limitations

This document is based on source-code analysis, not on production telemetry, live tracing, or database inspection. The conclusions are still strong because the instability is visible in the design itself, but the recommended next step is to pair these findings with runtime measurements from production or staging.

## Recommendation for Leadership

Treat inbox reliability as a funded engineering initiative with explicit ownership across frontend and backend, rather than as a sequence of isolated bug fixes.

A reasonable leadership framing is:

- Phase 1 fixes product trust
- Phase 2 reduces operational noise
- Phase 3 and Phase 4 create a scalable inbox foundation for future AI and automation features
