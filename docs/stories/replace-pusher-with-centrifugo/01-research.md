# Research — Replace Pusher with Centrifugo across the frontend

## Current State
The frontend uses **three** real-time mechanisms today: Pusher, Socket, and Centrifugo. **Centrifugo is already in use for Inbox.** The goal is to move every Pusher-powered real-time feature onto Centrifugo and then remove Pusher from the frontend entirely, leaving a single, consistent real-time stack.

Pusher appears in **~64 frontend files** (`contentstudio-frontend/src/`). The user-facing real-time areas it currently powers:

- **Planner** — `src/modules/planner_v2/*` and `src/modules/planner/*`: live post updates (status changes, new/updated/deleted posts) in calendar, list, and feed views.
- **Analytics** — `src/modules/analytics_v3/*` and `src/modules/analytics/*`: report-ready notifications and live-loading of report sections (e.g. `usePusherReportNotifications.js`, `usePusherAnalytics.js`, `useOverviewPusherAnalytics.js`, per-network `MainComponent.vue` views).
- **Composer & collaboration** — `src/modules/composer_v2/*` and `src/modules/publish/components/posting/collaboration/*`: live comments, tasks, and activity updates.
- **Discovery / Feeder** — `src/modules/discovery/components/feeder/*`: import progress and feeder group updates.
- **Listening** — `src/modules/listening/*`: live mention-loading progress.
- **Approval notifications** — `src/modules/approval-workflows/composables/useApprovalNotifications.ts`.
- **Banner notifications** — `src/modules/common/components/dialogs/BannerNotificationModal.vue`.
- **Core plumbing** — `src/composables/useAccount.ts`, `useApproval.ts`, `useAuthHydration.ts`, `src/stores/core/useProfileStore.ts`, `src/stores/publish/usePublishCommentStore.ts`, plus reset/plugin stores.

Key shared files:
- `src/modules/common/lib/centrifugo.ts` — the existing Centrifugo client (already wired for Inbox) → the pattern to reuse.
- `src/modules/common/lib/pusher.ts` — the Pusher client to remove.
- `src/config/api-utils.js` and `src/env.d.ts` — Pusher config/keys to remove.

## What Needs to Change
- Migrate every Pusher-powered real-time feature (Planner, Analytics, Composer/collaboration, Discovery, Listening, Approval, banner notifications) onto Centrifugo, reusing the existing Inbox/Centrifugo pattern.
- Remove the Pusher client, its config/keys, and all Pusher-specific code paths from the frontend.
- Preserve the real-time connection lifecycle (connect on login, tear down on logout, re-subscribe on workspace switch) so there are no stale/duplicate connections or missed updates.

## Dependency / boundary note
For any area not already publishing over Centrifugo, the backend must emit the same real-time events on Centrifugo. That's a backend concern handled separately — this story is the **frontend** migration + Pusher removal.

## Files Involved
See the module list above. Anchor files: `src/modules/common/lib/centrifugo.ts` (reuse), `src/modules/common/lib/pusher.ts` (remove), `src/config/api-utils.js`, `src/env.d.ts`, and the per-module composables/views listed.
