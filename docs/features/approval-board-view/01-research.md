# Approval Board: Research

**Feature:** Approval board, a new kanban-style view in the Planner, grouped by approval state
**Date:** 2026-10-02
**Design canvas (agreed with PO):** https://claude.ai/artifact/ACAtyDsDKYbBGc31FyY69e

---

## 0. Settled PO decisions (input to this pipeline)

| # | Decision |
|---|---|
| D1 | Approval-only board, not a general post-status kanban. Post statuses are system-driven (you can't drag a post into "Published" or "Failed"), so a status board would be read-only and duplicate List and Calendar |
| D2 | It is a **Planner view** called **"Approval board"**, picked from the view dropdown at the top right next to Calendar, List, Compact list, Feed, Instagram grid and TikTok grid. The PO renamed it from "Board" so it reads as approval-related |
| D3 | Columns: **Awaiting approval**, **Missed review**, **Changes requested**, **Approved** |
| D4 | Scope switch: **Assigned to me / Requested by me / All approvals** |
| D5 | **Drag and drop is required.** Every drag action also exists as a button on the card |
| D6 | Authors can still approve their own posts. That completes the current level, same as today |
| D7 | Multi-level workflow posts stay in Awaiting until the last level is approved. The card shows the level and a progress bar |
| D8 | The Approved column is limited to the planner's date range |
| D9 | **New** approvers land on Approval board by default. **Existing** users keep their saved view and get an in-app announcement instead |
| D10 | **The client share-link page also gets Approval board (v1, Must have)**, with 3 columns: Awaiting approval, Changes requested, Approved |

---

## 1. Competitor & Industry Research

### 1.1 What is this feature?

An approval board is a kanban view of the content planner. Each post is a card, the columns are approval states, and moving a card between columns *is* the approval action. Three roles want it for different reasons:

- **Approvers** (managers, clients, legal) want one queue of "what is waiting on me", with fast triage and bulk sign-off. A calendar buries that queue among published and scheduled posts.
- **Submitters** (creators, junior marketers) want to see where each post they sent is stuck, what came back with changes, and what is cleared to go.
- **Agency admins** want a pipeline view across the team, to spot bottlenecks such as posts whose publish time passed without a decision.

### 1.2 Competitor Analysis Table

| Competitor | Has Feature? | Key Capabilities | Pricing Tier | UX Approach | Unique Differentiator |
|---|---|---|---|---|---|
| **Planable** | **Yes** | "My approvals" board with 3 columns (Awaiting approval, Feedback left, Approved). "Approve all" bulk action, "Review X posts" step-through, count badge in the sidebar, aggregates across workspaces | Bulk approval on Pro/Enterprise. Multi-level approval is Enterprise only | Approver-centric personal board | Spans every client workspace. Review focus mode |
| **Sked Social** | **Yes** (launched 2026-08-05) | Approvals Board: status columns, drag-and-drop changes status, click to preview, Cmd+Click to open the editor, right-click quick actions, collapsible columns | Included with Approvals | Default view of the Approvals section. Cards look like the content | Closest direct analogue. Rich card interactions |
| **Kontentino** | **Yes** (status board) | Board View with status columns and drag-and-drop. Column picker, label and type filters, bulk "Time Savers" | Unverified | "Board" toggle from the calendar | You choose which columns are visible. Dragging to Assigned delegates the post |
| **Publer** | **Partial** | Kanban for Ideas (To Do, In Progress, In Review, Done) with columns you can add and edit | Unverified | Trello-like board for ideas, not scheduled posts | Custom columns. Not tied to approval state |
| **Vista Social** | **Partial** | Board, List and Calendar inside "Vista Work" project management. Approvals appear as "Pending review" on the calendar | Unverified | Board for tasks, calendar for approvals | Custom statuses per client |
| **Metricool** | **Partial** (list and calendar) | "Review requests" page with "For you / Assigned by you" tabs, bulk approve/reject, criteria (any or all reviewers), external email reviewers, 3-hour grace period | Advanced and Custom plans | List or calendar, no board | The For you / Assigned by you split matches our scope switch |
| **Sprout Social** | **No board** | Multi-step and multi-user approval paths, "Needs Approval" tab | Professional and Advanced | Tabbed list | Parallel and multi-stage paths |
| **Hootsuite** | **No board** | 2 to 3 tier approvals, Approvals tab in Planner and on mobile, recording of external approvals | Unverified | Calendar plus an Approvals tab | Audit trail for approvals given outside the tool |
| **Buffer** | **No board** | "Awaiting Approval" tab for drafts, single approver role | Team plan | Tab | Simplest model |
| **Later** | **No board** | Reviewer role, external reviewers without an account | Unverified | Calendar and post-level | Approval without a login |
| **Loomly** | **No board** (unverified) | Statuses include "Requires Edits" and "Pending Approval", changed from List View | Unverified | List with status filters | "Requires Edits" is the same idea as Changes requested |
| **Sendible** | **No board** | Tasks with "For Me / For Others", bulk send for approval | Unverified | Task list | Approval as tasks |
| **Agorapulse** | **No board** | Multi-step approval chains (2025), bulk assign and approve | Unverified | Calendar and list | Post stays in the workflow until every step is done |
| **SocialBee** | **No** | Single-approver Draft to Approved flow | Unverified | Post-level | None |

Sources: planable.io/guides/content-approvals-in-planable, skedsocial.com/blog/approvals-board, help.kontentino.com (welcome-workflow, collaboration-and-approvals), publer.com blog, vistasocial.com/insights/vista-work, help.metricool.com (review requests), support.sproutsocial.com (approval workflows), blog.hootsuite.com, support.buffer.com, help.later.com, loomly.zendesk.com, support.sendible.com, agorapulse.com 2025 features, socialbee.com, gainapp.com, sprinklr.com help. Pricing tiers marked Unverified could not be confirmed.

### 1.3 Common Patterns

- **Columns** come down to Pending, Changes requested or Rejected, and Approved. Planable uses 3 approver-centric columns.
- **Cards look like the content:** thumbnail, caption preview, channels, scheduled time, requester.
- **Bulk approve** is standard (Planable "Approve all", Metricool bulk dropdown) and often a paid-tier feature.
- **A drop changes the status** (Sked, Kontentino). Nobody documents drag restrictions or confirmation dialogs.
- **"For me / by me" scope** is common: Metricool, Sendible, Planable "My approvals".
- Several competitors have **multi-level approvals**, but **none shows step progress on a board card**.

### 1.4 Differentiators for ContentStudio

1. **A Missed review column.** No competitor surfaces posts whose publish time passed without a decision. Metricool silently stops publishing after a grace period.
2. **Multi-level progress on the card**, for example "Level 2 of 3: Legal" with a segmented bar and a "Your turn" badge.
3. **Rejecting by drag asks for a note**, which turns the drag into structured feedback.
4. **One board serves three personas** through the scope switch, instead of a separate approver-only page.

### 1.5 User Expectations

- **Table stakes:** columns with counts, cards that look like the content, preview on click, approve or reject from the card, account and date filters, an "assigned to me" scope, bulk approve, rejection feedback visible on the card.
- **Delighters:** drag to approve, missed-review surfacing, multi-level progress, collapsible columns, a step-through review mode.

### 1.6 Recommended Approach

Build it as a Planner view, so it gets the existing filters, saved views and URL params for free. Use the 4 columns with counts. Only allow drops the user is permitted to make, and fade columns they can't drop into as soon as a drag starts. Dropping on Changes requested asks for a note. Dropping a missed post on Approved asks for a new time. Every drag action has a button equivalent. Cards stay in Awaiting until the last level. Approve all only covers posts waiting on you.

**Where we differ from the competitor recommendation:** competitors suggest an undo toast after a drop. We use a **confirm step** instead, because a last-level approval schedules the post and can't be revoked (revoke is blocked once fully approved), so an undo would promise something we can't deliver.

---

## 2. Codebase Analysis

Paths are relative to the mounted repos. FE = `contentstudio-frontend/src`, BE = `contentstudio-backend`.

### 2.1 Two approval systems run side by side

- **Legacy single-level** `plan.approval`: `approve_option` (anyone or everyone), `approvers[]` with per-approver status. Overall status is `pending_approval`, `completed_approval` or `rejected_approval`. Logic: `BE/app/Builders/Approval/ApprovalBuilder.php` `approvalAction()` (~219-257).
- **Multi-level workflow** `plan.approval_workflow`: `workflow_id`, `current_level`, `total_levels`, `workflow_levels[]` (rule, members with status), `level_status[]`, `submitted_by`. Shape: `BE/app/Data/Posts/ApprovalWorkflowStateData.php`.
  - Engine: `BE/app/Builders/Approval/ApprovalWorkflowBuilder.php`. Workflow status is one of `pending`, `partially_approved`, `fully_approved`, `rejected`. A rule of "anyone" lets one approval complete the level.
  - On final approval the post goes to `scheduled`, or stays `draft` if it was a draft, and `publish_time_options.plan_status` becomes "Approved".
  - Any single rejection sets the post to `rejected`.
  - The post's creator can approve or reject even when not on the current level, which force-completes that level (D6).
  - Revoke isn't possible once fully approved.
  - Saved workflows: `BE/app/Models/Settings/ApprovalWorkflow.php`, up to 5 levels.
- **Column mapping must handle both systems:**

| Column | Legacy | Workflow |
|---|---|---|
| Awaiting approval | `status=review` with a future time, or a draft with `approval.status=pending_approval` | `approval_workflow.status` is `pending` or `partially_approved`, with time in the future or no time set |
| Missed review | `status=review` and the time has passed | Same |
| Changes requested | `status=rejected` | `approval_workflow.status=rejected` |
| Approved | `approval.status=completed_approval` | `approval_workflow.status=fully_approved` |

  Under review and missed review are computed, not stored: `BE/app/Models/Publish/Planner/Plans.php` ~273-311, filters at `PlansRepository.php` ~717-750.
- **Approved needs a new filter.** There is no "approved" status filter today. It means approval complete, whatever the post status is now (scheduled, published or draft), limited to the date range.

### 2.2 Planner views and the view dropdown

- **Routes:** `FE/modules/publisher/config/routes/publisher.ts` ~186-265. Six child routes: `planner_list_v2`, `planner_list_compact_v2`, `planner_calendar_v2`, `planner_feed_v2`, `planner_instagram_grid_view`, `planner_tiktok_grid_view`. Shell: `FE/modules/planner_v2/views/MainPlanner.vue`.
- **View dropdown:** `FE/modules/planner_v2/components/PlannerHeader.vue:159-204`. It is a `Dropdown` with one `DropdownItem` per entry in `visibleViewOptions` (computed at :459-464, which filters out views not allowed on mobile). Icons come from `FE/components/common/IconComponents/` (:334-372). Test ids: `VIEW_TESTID_SUFFIX` (:353-360).
- **To add a view:**
  - Add it to `PLANNER_ROUTE_NAMES` (:26), `PlannerProfileViewId` (:57), `PROFILE_VIEW_TO_ROUTE` and `PLANNER_VIEWS` (:91-128) in `FE/modules/planner_v2/constants/plannerRoutes.ts`.
  - Possibly also `ROUTE_NAME_TO_PLANNER_VIEW` and `PlannerView` in `queries/types.ts`.
  - Add the icon mapping in `composables/usePlannerViewSwitcher.ts:110-117`, a new icon component, the testid suffix, and an i18n key under `planner.planner_header.views.*`.
  - `__tests__/usePlannerViewSwitcher.spec.ts` asserts exactly 6 views and must be updated.
- **Filters:** `FE/stores/planner/usePlannerFilterStore.ts`. They're kept in the URL: statuses, labels, campaigns, content_category, members, created_by, approval_requested_by, approval_assigned_to, date and others. Status options are in `FilterSidebar.vue` ~461-516.
- **Default saved views:** `BE/app/Services/Planner/DefaultPlannerViewsService.php` ships "My Pending Approvals" and "My Approval Requests", both in the Feed view.
  - Line :116-122 uses `ucfirst(planner_default_view)` as the saved-view type, so the new view id needs a valid type there.
  - Moving these two views to Approval board for new workspaces is a candidate (see open questions).

### 2.3 Default view: `planner_default_view`

- **Storage:** a field on the User document (`BE/app/Data/Auth/UserData.php:68`). Set by `UsersRepository::setPlannerDefaultView` (:810) via `POST /preferences/setPlannerDefaultView` (`BE/routes/web/settings.php:71`, `UserPreferencesController.php:30`).
  - **The validator at :35 only allows `feed,list,compact_list,calendar,grid_instagram,grid_tiktok`.** The new view id must be added.
- **Every view switch saves the default:** `persistDefaultView` in `FE/modules/planner_v2/composables/usePlannerHeader.ts:240-251`. So once a user picks Approval board, it becomes their default, which is how existing users opt in (D9).
- **Fallbacks:** the FE falls back to `'list'` (`FE/stores/core/useProfileStore.ts:73`). Nothing sets the field at signup or invite.
- **New approvers (D9):** approver creation goes through `TeamController::processSingleMemberAddition` (`BE/app/Http/Controllers/Settings/Team/TeamController.php:215`; the approver block is at :261-265, creating the user via `registerAccount`). Set the board default there.
  - An existing user who is added as an approver takes a different branch. Only set the default when `planner_default_view` is empty, so nobody's saved choice gets overwritten.
  - `UsersRepository::updateUserDetails` :169-173 already sets `preferences.default_landing_page = "dashboard"` for approvers. That is precedent for role-based defaults.
- **Other places that switch on the view id:**
  - `getDefaultPlannerRoute` (`publisher.ts:94-135`). Mobile always goes to list.
  - `FE/modules/publisher/composables/usePlannerState.ts:76, :285`
  - `FE/modules/publisher/components/SidebarMain.vue:766`
  - `FE/composables/useWorkspaceSwitcher.ts:456`

### 2.4 Approver role routing

- Approvers have an explicit route allowlist: `FE/composables/usePermission.ts:268-291`. It includes feed, list, calendar and both grids, but not compact.
- The guard at `FE/router.ts:964-978` redirects anything else to `planner_list_v2`. **The new route must be added to the allowlist**, or approvers get bounced to List.
- Approver permissions: `approverCanCreatePost`, `approverCanEditPost`, `approverCanAddNotes`. Approvers log in at `/approverLogin`.

### 2.5 Data fetching and counts

- **Plans:** `POST /fetchPlans` (`BE/routes/web/planner.php:38`) calls `PlanController::fetchPlans` (:80), then `PlansRepository::getV2PlansByFilters` (:3039).
  - That path only runs for route names in the allowlist at :95. **The board route must be added there**, and to `getPlannerSpecificResponse` (:1700).
  - Pagination is by page and limit over one combined list (~:3138-3145). FE: `queries/usePlansInfiniteQuery.ts:146`.
- **Counts:** `POST /fetchPlansCount` calls `PlanController::fetchPlansCount` (:235), then `PlansRepository::fetchPlansCounts` (:866).
  - It already returns `review`, `missed` and `rejected`.
  - **The `requested_by_me` / `assigned_to_me` counts are commented out** (:891-892, :1031-1062). There is no "approved" count.
  - FE: `queries/usePlanStatusCountsQuery.ts`.
- **The board needs:** a page of posts per column (each column paginates on its own, with infinite scroll inside the column), and counts per column that follow the active scope and filters.
- **Approval fields are already in the response:** `PlannerViewService::extractCommonFields` (`BE/app/Services/PlannerViewService.php:167`, ~190-200) returns the raw `approval`, `approval_workflow` (current_level, levels, members with statuses) and `approval_workflow_history`. `components/WorkflowApprovalLevels.vue` and `ApproverList.vue` already render levels.

### 2.6 Scope filter semantics

- **`approval_assigned_to`** (`PlansRepository.php:413-446`) matches **any level**, not just the current one, and ignores the member's status:
  ```php
  $q->where('approval.approvers', 'elemMatch', ['user_id' => ['$in' => $ids]])
    ->orWhereIn('approval_workflow.workflow_levels.members.user_id', $ids);
  ```
  So "Assigned to me" already includes posts where my level hasn't started yet, which matches the mock ("Waiting on Level 1. Your turn comes at Level 3."). It also includes posts where I already approved and the post moved on. That's correct for the board: those cards sit in Awaiting (waiting on a later level) or in Approved.
- **"Your turn"** has to be worked out per card: am I a pending member of the current level? This logic exists in `PlannerViewService::isCurrentLevelPendingWorkflowMember` (:414+). The FE has `canActOnWorkflow` (`FE/composables/useApproval.ts:237`).
- **`approval_requested_by`** requires a legacy approvers list or a `workflow_id`, then matches `updated_by_id` (or `user_id` when that is null).
  - **Gotcha:** it matches on the last editor. If an approver edits a post, it may move from the author's "Requested by me" to the approver's. Worth confirming against `approval_workflow.submitted_by`.

### 2.7 Approve, reject and missed review

- **Single post, both systems:** `processPlanApproval` in `FE/composables/useApproval.ts:291+`.
  - **Workflow plans:** `changePlanStatusMutation` with status `scheduled` or `rejected` plus a comment (required for reject). Server path: `PlanController::changePlanStatus` (:319) → `maybeRouteThroughWorkflow` (~:382) → `ApprovalWorkflowBuilder` → `WorkflowNotificationDispatcher`.
  - **Legacy plans:** `planApprovalActionApi` (`FE/api/planner.ts:278`).
  - **Drag and drop should reuse this path.**
- **Dedicated REST routes** exist (`plans/{planId}/approve|reject|revoke-approval|re-notify`, `BE/routes/api/approval-workflows.php:61-63`) but have no FE caller today.
- **Bulk approve (the most important finding):** `PlanController::processPlanBulkOperation` (:777-1130) **never references `approval_workflow`**.
  - It filters on `publish_time_options.plan_status = 'Under Review'`, updates legacy approver status when `plan.approval` is set, and otherwise sets `status` to `rejected` or `scheduled` directly (~:906, ~:1016).
  - For a workflow post that likely **skips its levels and notifications**.
  - The FE bulk bar (`PlannerBulkActionsBar.vue` ~87, `usePlanApprovalFlow.ts` ~371) only enables Approve when the status filter is exactly `under_review`.
  - **Approve all on the board must not ship on this endpoint as it is.** It needs a BE story to route every post through the workflow engine (legacy posts keep the legacy path).
- **Missed review:**
  - The modal is `<PublishingTimeFeedView type="missed_reviewed">` (mounted in `MainPlanner.vue:64`, logic in `composables/usePublishingTimeFeedView.ts:347`). For workflow posts it calls `changePlanStatus({status:'scheduled', publish_time_options})`.
  - It's opened by `initializeMissedReviewedDatePicker` (`useApproval.ts:563`), from `changePlanStatusMethod` (:259-271) when the execution time is in the past.
  - BE: `ApprovalWorkflowBuilder::approve(?comment, ?publishTimeOptions)` (:73, ~191-202) applies the new time.
  - **Unverified:** the workflow branch of `processPlanApproval` may skip the missed-review date check. The board must always ask for a new time on a missed post.
- **Notifications:** approval notifications go through `WorkflowNotificationDispatcher`. Review-missed emails link to `/planner/list-view?statuses=missed_review` (`BE/app/Notifications/Publish/PlanNotification.php:292`). Workflow emails link to `/publisher/planner/list-view?plan_ids=` (`BE/app/Notifications/Approval/WorkflowNotification.php:230`).
- **The bell panel** (`FE/modules/approval-workflows/components/ApprovalNotificationsPanel.vue` ~239-325) opens `planner_feed_v2?plan_ids=<id>` plus the preview. Its "Open planner" opens the feed view with under_review and missed_review, plus requested-by and assigned-to set to me.

### 2.8 Real-time

- The planner uses Socket.IO (`window.socket`), not Pusher. `dispatchApprovalSocket` (`useApproval.ts:652`) emits `feed_approval`, but only on the legacy path.
- `FE/composables/useListener.ts:380-400` listens for `post_approval`, `plan_approval`, `bulk_post_approval`, `missed_review` and `review_missed`, but only to show notifications.
- **The planner doesn't refresh when someone else approves.** Refetching only happens on a notification click (`useNotificationHandler.ts:170`). The board should refetch the affected columns after the user's own actions. Live refresh from other users' actions is a P1 or P2 candidate.

### 2.9 Drag and drop and UI components

- `vuedraggable` 4.1.0 is installed and already used for list reordering (`SidebarMain`, `PlannerCustomViewsSidebarDropdown`, `CustomizeSidebarModal`, composer, `ApprovalWorkflowBuilder`). Its `group` / `put` / `pull` and `move` callback can enforce allowed columns.
- The calendar uses FullCalendar's interaction plugin for rescheduling by drag (`useCalendarView.ts:279`). That's not relevant here.
- **No kanban component exists** anywhere in the frontend.
- **Available in `@contentstudio/ui`:**
  - `SegmentedControl` for the scope switch
  - `Badge` for counts, "Your turn" and "Time passed"
  - `Progress` for level progress (check that it supports segments; otherwise compose it from bars)
  - `Avatar` for accounts and requester
  - `Modal` / `Dialog` for the confirm dialogs
  - `Textarea` for notes
  - `Button`
  - `Dropdown` / `DropdownItem`
  - `Loader`
- **Gaps:** there is no generic Card or kanban column component, no Tooltip, and no Pill/Chip. These need a [Design] story.

### 2.10 Usermaven

There is no planner, view-switch or approval tracking today. The only global event is `pageview` (`FE/router.ts:1127`), which will cover route visits to the new view. Track with `trackUserMaven` from `useUserMaven`.

### 2.11 Public developer surfaces

- `GET workspaces/{id}/posts` (`BE/routes/api/v1.php:316`, `PostIndexRequest.php:27-34`) already accepts `under_review`, `missed_review` and `rejected`, plus `approval_assigned_to[]` and `approval_requested_by[]`.
- `POST workspaces/{id}/posts/{post_id}/approval` (~:368) already routes workflow plans through `handleWorkflowApproval` (PostController ~:3143).
- **Gaps:** there's no "approved" status filter and no counts endpoint. If the board adds an approved filter to the internal query, the public posts list should accept it too. The `contentstudio` CLI and the MCP `fetch_posts` tool would then pick it up.

### 2.11b Client share-link page

- **Page:** `FE/modules/planner_v2/views/SharePlans.vue` (guest route `share-plans`). It has a List and a Calendar view, switched in `components/share-plans/SharePlansControls.vue`. It also has `SharePlansListView.vue`, the bulk bar `SharePlansBulkActionsBar.vue`, and a preview with approve and reject with comment.
- **Backend:** the model is `BE/app/Models/Publish/Planner/ShareLink.php`, with `approval_flow`, `approval_emails`, `approval_tokens` and `allow_external_approval_actions`. Each email has its own token. The public endpoints are `shareLink/get`, `fetchPlans`, `comment` and `action` (`BE/routes/web/planner.php` ~22-31).
- **Client approvals are legacy only:** `BE/app/Builders/Approval/ExternalApprovalBuilder.php` updates `plan.approval` and records `plan.external_actions`. **It doesn't touch `approval_workflow`.**
- **Gotcha:** when a client approves a post whose time has passed (~:126-133), the approval becomes `completed_approval`, but the status stays `review` because it only schedules when the time is still in the future. The post is stuck. The client board puts missed posts in Awaiting with a "Time passed" badge, and the team board shows them in Approved as "Approved, needs a new time".
- **Columns per client** should come from that client's own entry in the approvers list or `external_actions`, so each client on a shared link sees their own decision.

### 2.12 Mobile

- Out of scope for v1. The board is a desktop pattern.
- The responsive web planner should still work at phone width: the columns become tabs, as in the design canvas.
- The Flutter app has `lib/features/approval_workflows/` but no kanban. No [Flutter] story is planned.

---

## 3. Risks and gotchas surfaced by research

1. **Bulk approve on workflow posts** likely bypasses levels today (2.7). This is a blocker for Approve all.
2. **The default view is saved on every switch** (2.3). That's fine for opting in, but the release must not write `board` for existing users.
3. **Approvers get redirected** unless the new route is added to their allowlist (2.4).
4. **Requested by me matches the last editor**, not the submitter (2.6).
5. **No live refresh** (2.8). Two approvers on the same board can act on a stale card, so the server must return a clear error when the post was already acted on.
6. **Paging columns separately** needs a new or extended endpoint (2.5).
