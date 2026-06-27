# Story — Replace Pusher with Centrifugo across the frontend

Standalone story for the **Q2 - 2026: Miscellaneous** epic. Carries its own Shortcut fields block. Nothing is pushed to Shortcut — the Product Owner creates it manually using the **New Feature Template**.

---

## [FE] Replace Pusher with Centrifugo across the frontend

### Description:
As ContentStudio, we want all live (real-time) updates in the web app to run through a **single** real-time service — Centrifugo — and to remove **Pusher** from the frontend entirely. Centrifugo is already used for the Inbox; we want everywhere else that still relies on Pusher to use Centrifugo too, so the app runs on one consistent real-time stack instead of several. This is simpler to maintain and removes the duplicate dependency.

This is a behind-the-scenes change: the live updates users already see should keep working exactly as they do today — just powered by Centrifugo instead of Pusher.

---

### Workflow:
1. Wherever the app shows live updates today (the post planner updating as posts are scheduled/published/changed, analytics reports notifying when they're ready and loading section by section, comments/tasks/activity appearing in real time on a post, content import progress, listening mentions loading, approval notifications, and in-app banner notifications), those updates continue to work — now delivered through Centrifugo.
2. The user experiences no visible difference: the same things update live, at the same moments, with the same on-screen behavior.
3. Pusher is no longer loaded or connected anywhere in the frontend.

---

### Acceptance criteria:

- [ ] Every live-update feature that uses Pusher today is moved to Centrifugo, with **no change the user can notice** — the same updates still happen in real time.
- [ ] The migrated areas cover everywhere Pusher is used today: **Planner** (calendar, list, and feed views — live post status and new/updated/deleted posts), **Analytics** (report-ready notifications and live-loading report sections), **Composer & collaboration** (live comments, tasks, and activity updates), **Discovery / content feeds** (e.g. import progress), **Listening** (live mention-loading progress), **Approval notifications**, and **in-app banner notifications**.
- [ ] Inbox real-time (already on Centrifugo) continues to work with no regression.
- [ ] Pusher is fully removed from the frontend — the Pusher client, its setup/keys/config, and any Pusher-specific code are deleted, and the app no longer loads or connects to Pusher.
- [ ] Live connections start on login, are torn down on logout, and re-subscribe correctly on workspace switch — with no stale or duplicate connections and no missed updates.
- [ ] After a dropped connection, live updates reconnect on their own.
- [ ] No console errors related to real-time connections after the change.

---

### Mock-ups:
N/A — no visual change; existing live-update behavior and copy are unchanged.

---

### Impact on existing data:
None.

---

### Impact on other products:
- **Web app frontend only.** No visual or behavioral change for users.
- **Mobile apps** use their own real-time handling and are not affected by removing Pusher from the web frontend.
- **Chrome extension:** not in scope for this story.
- **White-label:** live updates must keep working on white-label domains after the switch.

---

### Dependencies:
For any area not already delivering live updates over Centrifugo, the backend must publish the same real-time events on Centrifugo. That backend work is handled separately — this story is the frontend migration and Pusher removal.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, no UI/layout change
- [ ] Multilingual support — N/A, no new user-facing copy
- [ ] UI theming support — N/A, no UI change
- [ ] White-label domains impact review (live updates must keep working on white-label domains)
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Reuse the existing Centrifugo setup:**
- `contentstudio-frontend/src/modules/common/lib/centrifugo.ts` — the Centrifugo client already wired for Inbox. The Inbox modules (`src/modules/inbox-revamp/*`) are the reference for connection lifecycle (connect/disconnect, subscribe per workspace).

**Remove these Pusher pieces:**
- `contentstudio-frontend/src/modules/common/lib/pusher.ts` (the Pusher client)
- Pusher config/keys in `contentstudio-frontend/src/config/api-utils.js` and the type in `contentstudio-frontend/src/env.d.ts`

**Areas still on Pusher today (migrate each — ~64 files):**
- Planner: `src/modules/planner_v2/*` (e.g. `MainPlanner.vue`, `CalenderView.vue`, `FeedView.vue`, `DataTable.vue`, `composables/usePlannerHelper.ts`, `composables/useUpdatePostHelper.js`) and `src/modules/planner/*`
- Analytics: `src/modules/analytics_v3/*` (`usePusherReportNotifications.js`, `MainAnalyticsHeader.vue`, competitor report views) and `src/modules/analytics/*` (`usePusherAnalytics.js`, `useOverviewPusherAnalytics.js`, `usePinterestPusherAnalytics.js`, per-network `MainComponent.vue`)
- Composer & collaboration: `src/modules/composer_v2/components/{Comments,Tasks,Activities}.vue` and `src/modules/publish/components/posting/collaboration/{Comments,Tasks,Activities}.vue`
- Discovery feeder: `src/modules/discovery/components/feeder/*`
- Listening: `src/modules/listening/*`
- Approval notifications: `src/modules/approval-workflows/composables/useApprovalNotifications.ts`
- Banner notifications: `src/modules/common/components/dialogs/BannerNotificationModal.vue`
- Core composables/stores: `src/composables/{useAccount,useApproval,useAuthHydration}.ts`, `src/stores/core/useProfileStore.ts`, `src/stores/publish/usePublishCommentStore.ts`, and the store reset/plugin files

**Gotcha:** the connection lifecycle is the main risk area — make sure subscriptions are cleaned up on logout and re-established on workspace switch so there are no duplicate listeners or missed events (mirror how Inbox already does it).

---

### Shortcut fields
- **Template:** New Feature Template
- **Story type:** Chore
- **Project:** Web App
- **Group:** Frontend
- **Epic:** Q2 - 2026: Miscellaneous
- **Priority:** Medium
- **Product area:** Throughout Product
- **Skill set:** Frontend
- **Estimate:** _(empty — devs estimate at sprint planning)_
- **Labels:** none
- **Iteration:** assigned by PO at creation
