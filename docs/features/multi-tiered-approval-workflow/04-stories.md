# Stories: Multi-Tiered Approval Workflow

**Epic:** https://app.shortcut.com/contentstudio-team/epic/47607
**Figma:** https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034
**Source documents:** 02-workflow.md, 03-prd.md

---

## Story Index

| # | Title | Type | Priority | Product Area |
|---|---|---|---|---|
| S-01 | [BE] Create approval_workflows collection and CRUD API | backend | P0 | settings |
| S-02 | [BE] Extend plans.approval schema for multi-level workflow state | backend | P0 | throughout_product |
| S-03 | [BE] Implement WorkflowApprovalBuilder for multi-level approval progression | backend | P0 | throughout_product |
| S-04 | [BE] Add new notification event types for multi-tiered approval | backend | P0 | throughout_product |
| S-05 | [BE] Implement workflow mid-flight modification rules | backend | P0 | throughout_product |
| S-06 | [BE] Fix rejected post silent status reset bug in ApprovalBuilder | backend | P0 | throughout_product |
| S-07 | [BE] Add "Allow Approval Workflow Management" permission to team member settings | backend | P1 | settings |
| S-08 | [BE] Handle approver removal from workspace with in-flight approval impacts | backend | P0 | settings |
| S-09 | [FE] Build Approval Workflows settings page (workflow list) | frontend | P0 | settings |
| S-10 | [FE] Build Create/Edit Workflow builder with levels, members, and drag-and-drop | frontend | P0 | settings |
| S-11 | [FE] Migrate Send for Approval from center modal to right sidebar with Users and Approval Workflows tabs | frontend | P0 | composer |
| S-12 | [FE] Build approval status tracking panel (planner hover popup and post preview modal) | frontend | P1 | planner |
| S-13 | [FE] Build edit confirmation dialogs for posts in approval (all 4 scenarios) | frontend | P0 | composer |
| S-14 | [FE] Build dedicated Approval Notifications icon and panel in the top bar | frontend | P1 | throughout_product |
| S-15 | [FE] Add "Manage Approval Workflows" permission toggle to Team Member Settings | frontend | P1 | settings |
| S-16 | [FE] Update Planner post status badges and filter labels for multi-level approval | frontend | P1 | planner |
| S-17 | [Design] Approval Workflow Settings, Create/Edit builder, and Send for Approval sidebar | design | P0 | throughout_product |
| S-18 | [iOS] Add Approval Workflows tab to Send for Approval in Composer | mobile | P0 | throughout_product |
| S-19 | [iOS] Add Pending Approvals bottom navigation item and view | mobile | P1 | throughout_product |
| S-20 | [Android] Add Approval Workflows tab to Send for Approval in Composer | mobile | P0 | throughout_product |
| S-21 | [Android] Add Pending Approvals bottom navigation item and view | mobile | P1 | throughout_product |
| S-22 | [BE] Implement approval override cleanup and cancellation notification dispatch | backend | P0 | throughout_product |
| S-23 | [FE] Add "Send for Approval" to single Planner post card and post detail view | frontend | P1 | planner |
| S-24 | [FE] Add approval override warning to Send for Approval sidebar, bulk Planner send, and external share link | frontend | P0 | throughout_product |

---

## S-01: [BE] Create approval_workflows collection and CRUD API

### Description:
As a workspace admin, I want to create, edit, duplicate, and delete named approval workflows so that my team can reuse predefined multi-level approval chains without re-selecting approvers every time.

---

### Workflow:
1. Admin opens Settings → Approval Workflows (new page)
2. Admin clicks "Create a Workflow", enters a name, defines up to 5 levels, assigns members to each, sets approval rule per level, and saves
3. Admin can edit any existing workflow, duplicate it (creates "[Name] (Copy)"), or delete it
4. Admin can mark one workflow as Default — it will be pre-selected in the Send for Approval sidebar
5. A workflow saved without all levels having at least 1 member is saved as "Draft" and hidden from the composer
6. Deleting a workflow with in-flight posts converts those posts to single-user ad-hoc approval before deleting

---

### Acceptance criteria:
- [ ] The `approval_workflows` resource must support the following fields: `_id`, `workspace_id`, `name`, `is_draft` (bool), `is_default` (bool), `levels[]` (each with: `level_number`, `title`, `rule` ["everyone"|"anyone"], `members[]` [workspace_member_ids]), `created_by`, `created_at`, `updated_at`
- [ ] `GET /workspaces/{workspace_id}/approval-workflows` returns all workflows for the workspace (excluding drafts when `include_drafts=false`)
- [ ] `POST /workspaces/{workspace_id}/approval-workflows` creates a new workflow; validates name is required; validates max 5 levels; validates each published (non-draft) level has at least 1 member
- [ ] `GET /approval-workflows/{id}` returns single workflow with all levels and member details
- [ ] `PUT /approval-workflows/{id}` updates workflow; applies mid-flight modification rules for in-flight posts (see S-05)
- [ ] `DELETE /approval-workflows/{id}` — if no in-flight posts: deletes immediately; if in-flight posts exist: returns `in_flight_count` and `affected_posts[]` for confirmation; on `?force=true`: converts in-flight posts to ad-hoc single-level approval using current active level's members, then deletes
- [ ] `POST /approval-workflows/{id}/duplicate` creates a copy named "[Name] (Copy)" with `is_default=false`
- [ ] `PUT /approval-workflows/{id}/set-default` sets this workflow as default and unsets any previous default in the workspace (only one default per workspace)
- [ ] `PUT /approval-workflows/{id}/remove-default` removes default flag
- [ ] All endpoints are gated by workspace role: Super Admin and Admin always have access; Collaborators/Approvers require `allow_workflow_management=true` on their workspace membership (see S-07)
- [ ] Workflow with `is_draft=true` is excluded from `GET .../approval-workflows` list unless `include_drafts=true` is passed (draft workflows must not appear in the composer)
- [ ] Saving a workflow where any level has 0 members results in `is_draft=true` automatically
- [ ] When a default workflow is deleted, the default is cleared (no auto-assignment of new default)
- [ ] Level title defaults: "Level 01:", "Level 02:", etc. (if user leaves title blank)
- [ ] Duplicate level creates a new level entry with same members and rule but a new `level_number` appended at end
- [ ] API response for list includes: id, name, is_draft, is_default, levels_count, members_count (total unique members across all levels)

---

### Mock-ups:
Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
- New `approval_workflows` resource — no impact on existing data
- Existing `plans.approval` data is not changed by this story

---

### Impact on other products:
- All products that submit plans for approval will eventually use workflows; this story creates the data layer they depend on
- iOS and Android apps consume the workflow list via this API

---

### Dependencies:
- Plan data and workspace member data must be accessible
- Workspace role/permission model must support the new `allow_workflow_management` field (see [BE] Add "Allow Approval Workflow Management" permission to team member settings)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A (backend-only story)
- [ ] Multilingual support — N/A for API; workflow names are user-defined strings
- [ ] UI theming support — N/A (backend-only story)
- [ ] White-label domains impact review — N/A; API is workspace-scoped
- [ ] Cross-product impact assessment — iOS, Android, and web all consume this API

---

## S-02: [BE] Extend plans.approval schema for multi-level workflow state

### Description:
As a backend developer, I want the `plans.approval` data structure to support multi-level workflow state tracking so that the system can persist the full progression of a workflow-based approval alongside the existing single-level ad-hoc approval without breaking backwards compatibility.

---

### Workflow:
1. When a post is sent for approval via a workflow, the system stores the workflow snapshot (workflow_id, all levels, all members) in `plans.approval` alongside the current approval state
2. As approvers act, the system updates per-level, per-member approval statuses
3. The `current_level` index tracks which level is active
4. The existing single-level approval fields remain unchanged for ad-hoc approvals
5. Approval history is permanently stored — it is never deleted, even after the post is published

---

### Acceptance criteria:
- [ ] The plan detail API must return the following new optional fields on `plans.approval` (backwards-compatible; single-level ad-hoc approval continues to use existing fields unchanged):
  - `workflow_id` (string|null) — reference to the approval workflow (null for ad-hoc)
  - `workflow_name` (string|null) — snapshot of workflow name at time of submission (for history if workflow is renamed/deleted)
  - `current_level` (int|null) — 1-based index of the currently active level (null for ad-hoc)
  - `total_levels` (int|null) — total number of levels in the workflow at time of submission
  - `workflow_levels[]` (array|null) — snapshot array of all levels at time of submission, each with: `level_number`, `title`, `rule`, `members[]` (each with: `member_id`, `status` ["pending"|"approved"|"rejected"|"no_action_needed"], `actioned_at`, `comment`)
  - `level_status[]` (array|null) — per-level aggregate status: ["not_started"|"in_progress"|"completed"|"rejected"]
- [ ] Existing `plans.approval.approvers[]` and `plans.approval.status` fields remain intact for ad-hoc approvals; they are populated for workflow approvals too (flattened list of current-level members) to maintain backwards compatibility with existing queries
- [ ] When a workflow post is fully approved, `plans.status` transitions appropriately (Scheduled stays Scheduled; Draft stays Draft — approval is an overlay)
- [ ] When a workflow post is rejected, `plans.status` is set to the existing `rejected` approval status value
- [ ] `plans.approval` records are never deleted — full history is preserved after publish
- [ ] The API must support efficient lookup of: "all plans in workspace X currently at workflow level Y"
- [ ] Pusher broadcast event `feed_approval` is updated to include `workflow_id`, `current_level`, and `level_status[]` fields for real-time UI updates
- [ ] `GET /plans/{id}` response includes the full `approval` object with all new fields when `with_approval=true` is passed

---

### Mock-ups:
N/A (backend schema change)

---

### Impact on existing data:
- Additive change — existing `plans.approval` records without new fields are treated as ad-hoc (null workflow_id); no migration required
- Existing approval status values and queries remain functional

---

### Impact on other products:
- iOS and Android clients must handle the new `workflow_id`, `current_level`, `workflow_levels[]` fields in the plan API response gracefully (ignore if not used yet — parsed in S-18 and S-20)

---

### Dependencies:
- Depends on: [BE] Create approval_workflows collection and CRUD API (workflow_id reference)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A
- [ ] Multilingual support — N/A (backend schema)
- [ ] UI theming support — N/A
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — all clients (web, iOS, Android) consuming the plans API are impacted; new fields are additive and must not break existing parsers

---

## S-03: [BE] Implement WorkflowApprovalBuilder for multi-level approval progression

### Description:
As a content creator, I want posts to progress automatically through workflow levels — notifying each tier only when the previous one is complete — so that I don't have to manually coordinate who reviews in what order.

---

### Workflow:
1. Creator sends post for approval using a workflow → Level 1 members are notified; `current_level = 1`, all Level 1 members set to `pending`
2. Level 1 members act (approve/reject). Based on rule:
   - "Everyone": all must approve → then advance to Level 2
   - "Anyone": first approval advances to Level 2; remaining Level 1 members set to `no_action_needed`
3. Level 2 members are notified; same logic repeats
4. Final level completes → post is fully approved; `plans.approval.status = fully_approved`
5. Rejection at any level → `plans.status = rejected`; remaining same-level pending members set to `no_action_needed`; workflow stops
6. Missed Review: approval continues; only current-level members notified of the missed status
7. Post deleted during approval: workflow terminates; only current-level members notified
8. Schedule date changed during approval: no effect on approval state
9. Revoke approval: approver can reset their status back to `pending` if their level is still active AND not already advanced via "Anyone" rule

---

### Acceptance criteria:
- [ ] Multi-level approval progression is handled server-side; the existing single-level ad-hoc approval flow is not modified (backwards-compatible)
- [ ] `POST /plans/{id}/approve` — records the approver's approval on the current level; if "Everyone" rule and all members approved → advances level; if "Anyone" rule → advances immediately, sets remaining members to `no_action_needed`; triggers appropriate notifications
- [ ] `POST /plans/{id}/reject` — records rejection with optional comment; terminates workflow at current level; sets remaining pending same-level members to `no_action_needed`; sets `plans.status = rejected`
- [ ] `POST /plans/{id}/revoke-approval` — resets approver's status to `pending`; only succeeds if: (a) approver's level is currently active, (b) post is not fully approved, (c) for "Anyone" rule: their approval did NOT advance the level (if it did, return 422 with message "This level has already advanced — your approval can no longer be revoked"); triggers `approval_revoked_creator` and `approval_revoked_approvers` notifications
- [ ] Level advancement: when Level N completes, `current_level` increments to N+1; Level N+1 members' statuses set to `pending`; `workflow_level_advanced` notification sent to Level N+1 members
- [ ] Final level completion: `current_level` set to null (or to `total_levels + 1` to indicate completion); `plans.approval.status = fully_approved`; `post_fully_approved` notification sent to creator
- [ ] Rejection: `plans.status = rejected`; `reject_approval` notification sent to creator; `rejection_no_action_needed` sent to other pending same-level members
- [ ] Missed Review during approval: when the post's scheduled time passes and it is still in approval, `plans.status = missed_review`; only current active level members are notified of missed review status; approve/reject actions remain available
- [ ] Reschedule prompt: when an approver acts on a Missed Review post, the API response includes `is_missed_review: true` so the frontend can show the reschedule prompt
- [ ] Post deleted (soft-deleted) while in approval: `post_deleted_approval` notification sent to current-level members only; workflow state is preserved in the plan's approval history (for audit)
- [ ] Schedule date change during approval: no approval state change; no notifications triggered
- [ ] Approval history: all `workflow_levels[]` data including timestamps and comments is permanently preserved on the plan document; never overwritten or deleted
- [ ] All workflow approval actions dispatch the appropriate notification events for all new event types
- [ ] `GET /plans/{id}/approval-history` returns full approval audit trail (all levels, all decisions, timestamps, comments)

---

### Mock-ups:
N/A (business logic)

---

### Impact on existing data:
- Single-level ad-hoc approval flow is unchanged

---

### Impact on other products:
- iOS and Android approval actions (approve/reject) call the same endpoints — must handle new response fields (`is_missed_review`, `workflow_level_advanced` etc.) gracefully

---

### Dependencies:
- Depends on: [BE] Extend plans.approval schema for multi-level workflow state
- Depends on: [BE] Add new notification event types for multi-tiered approval (for notification dispatch)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A
- [ ] Multilingual support — N/A (backend logic)
- [ ] UI theming support — N/A
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — iOS, Android, and web all call approve/reject endpoints; new response fields must not break existing mobile parsers

---

## S-04: [BE] Add new notification event types for multi-tiered approval

### Description:
As an approver or content creator, I want to receive precise notifications at the right moment — only when it's my turn or something changes that affects me — so that I'm not overwhelmed with irrelevant updates and never miss something that requires my attention.

---

### Workflow:
1. Events in the approval lifecycle dispatch specific notification types
2. All new notification event types are implemented server-side
3. All new notification copy from PRD Section 10 must be implemented
4. Email notifications receive new template variables for workflow context
5. Re-notify cooldown (2 hours) is enforced server-side

---

### Acceptance criteria:

**Updated existing notifications:**
- [ ] `pending_approval`: In-app description updated to append `(Level :level_number: :level_title)` for workflow submissions
- [ ] `pending_approval`: Email CTA text is "Please review the post and approve or reject it."
- [ ] `approve_approval`: In-app and email updated to include level completion info for workflows: "Level :level_number (:level_title) is complete."
- [ ] `reject_approval`: No copy changes needed; channel additions for mobile push confirmed

**New notification types:**
- [ ] `workflow_level_advanced` — sent to next level's members when previous level completes; in-app + email + mobile push; copy as per PRD Section 10.2
- [ ] `post_fully_approved` — sent to creator when final level completes; in-app + email + mobile push; copy as per PRD Section 10.2
- [ ] `level_no_action_needed` — sent to remaining same-level approvers when "Anyone" rule closes level; in-app + email only; copy as per PRD Section 10.2
- [ ] `rejection_no_action_needed` — sent to other pending same-level approvers when post is rejected; in-app + email only; copy as per PRD Section 10.2
- [ ] `post_deleted_approval` — sent to current active level approvers when post is deleted; in-app + email only; no CTA button in email; copy as per PRD Section 10.2
- [ ] `approval_revoked_creator` — sent to creator when approver revokes; in-app + email; copy as per PRD Section 10.2
- [ ] `approval_revoked_approvers` — sent to other same-level approvers when one revokes; in-app + email; copy as per PRD Section 10.2
- [ ] `content_updated_approval` — sent to relevant approvers when creator edits post during approval; in-app + email + mobile push; recipients depend on creator's dialog choice; copy as per PRD Section 10.2
- [ ] `re_notify` — manual reminder; in-app + email + mobile push; copy as per PRD Section 10.2; cooldown enforced
- [ ] `approver_removed_post` — sent to creator when a workspace member who was their approver is removed; in-app + email; copy as per PRD Section 10.2
- [ ] `level_removed_auto_advanced` — sent to creator when a workflow level they were at is deleted and their post auto-advances; in-app + email; copy as per PRD Section 10.2
- [ ] `level_removed_auto_approved` — sent to creator when the final workflow level is deleted and their post is auto-approved; in-app + email + mobile push; copy as per PRD Section 10.2
- [ ] `approval_cancelled_override` — sent to current active level approvers when a new approval overrides the existing one; in-app + email; copy as per PRD Section 10.2 (added in S-22)

**Re-notify cooldown:**
- [ ] `POST /plans/{plan_id}/re-notify/{member_id}` endpoint dispatches `re_notify` notification to the specified approver
- [ ] Cooldown: if a re-notify was sent to this approver on this plan within the last 2 hours, return 429 with `next_available_at` timestamp in response
- [ ] Cooldown enforced server-side with a 2-hour window per approver per plan

**Email template:**
- [ ] New template variables added (all optional, null-safe): `$workflow_name`, `$level_number`, `$level_title`, `$total_levels`, `$previous_level_number`, `$previous_level_title`, `$rejecter_name`, `$approver_name`, `$creator_name`, `$member_name`
- [ ] Conditional CTA button logic: "Review Post" for action-required types; "View Post" for info-only types; no button for `post_deleted_approval`
- [ ] All existing email notifications continue to work without new variables (null-safe rendering)

**Notification copy:**
- [ ] All new notification copy from PRD Section 10 must be implemented for all supported languages, with English fallback where translations are not yet available

---

### Mock-ups:
N/A (backend)

---

### Impact on existing data:
- Additive changes — new notification event types only; existing notifications unchanged

---

### Impact on other products:
- Mobile push notifications for new event types must be confirmed working with existing FCM (Android) and APNs (iOS) dispatch infrastructure

---

### Dependencies:
- Depends on: [BE] Implement WorkflowApprovalBuilder for multi-level approval progression (event triggers)
- Depends on: [BE] Extend plans.approval schema for multi-level workflow state (level context data)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A
- [ ] Multilingual support — all new notification copy added; English fallback for languages pending translation
- [ ] UI theming support — N/A (email template uses existing white-label compatible system)
- [ ] White-label domains impact review — email template uses existing white-label domain logic; no changes needed
- [ ] Cross-product impact assessment — mobile push affected; FCM/APNs dispatch must be verified for all new event types

---

## S-05: [BE] Implement workflow mid-flight modification rules

### Description:
As a workspace admin, I want changes I make to a workflow's levels and members to take effect intelligently on posts that are already mid-approval — following clear rules about what can and cannot change after a level has started or completed — so that in-flight approvals are never silently broken or orphaned.

---

### Workflow:
1. Admin opens an existing workflow in Settings and edits it (add/remove members, add/remove levels, change approval rules, rename)
2. The system evaluates each change against all in-flight posts using this workflow and applies the appropriate per-change rules
3. Changes are applied immediately; level removals with in-progress posts trigger auto-advance logic and notifications
4. Before saving, if the workflow has in-flight posts, the user must confirm a save confirmation dialog

---

### Acceptance criteria:

**Member changes (applied on `PUT /approval-workflows/{id}`):**
- [ ] Member added to a NOT STARTED level: member is added to all in-flight posts' future-level assignments; they will be notified when that level starts
- [ ] Member added to an IN PROGRESS level: member is added to the current level on all affected in-flight posts; `pending_approval` notification dispatched to them immediately
- [ ] Member added to a COMPLETED level: NOT added retroactively to any in-flight post; only applies to new posts submitted after this change
- [ ] Member removed from workflow and has NO pending posts in that role: removed cleanly from workflow definition
- [ ] Member removed from workflow and HAS pending posts in that level (level not started): their assignment on that future level is cancelled; `level_no_action_needed` style notification dispatched to them per affected post
- [ ] Member removed from workflow and HAS posts where that level IS IN PROGRESS: their assignment cancelled on in-flight posts; notification dispatched per affected post
- [ ] Member removed from workflow and HAS posts where that level IS COMPLETED: NOT removed from the completed level record on those posts (greyed-out in UI); the completed approval record is preserved; API returns a `completed_approvals_count` to help UI indicate this

**Level structure changes:**
- [ ] Level added to workflow: applies to all in-flight posts that have not yet reached this position (i.e., posts at earlier levels); inserted into the workflow progression for those posts
- [ ] Level removed from workflow where ALL in-flight posts are NOT at that level: removed cleanly
- [ ] Level removed where ANY in-flight post is currently AT that level:
  - Posts at that level **auto-advance to the next level** in the workflow; `workflow_level_advanced` notification dispatched to next level's approvers; `level_removed_auto_advanced` notification dispatched to each affected post's creator
  - If the removed level was the **last level**: posts are **auto-completed (fully approved)**; `level_removed_auto_approved` notification dispatched to each affected post's creator (in-app + email + mobile push)
- [ ] `level_removed_auto_advanced` notification spec: in-app title: `"Your post has been moved forward in [Workflow Name]"`; body: `"Level [N] ([Title]) was removed from [Workflow Name]. Your post has automatically advanced to Level [N+1] ([Title])."`; channels: in-app, email
- [ ] `level_removed_auto_approved` notification spec: in-app title: `"[Post title] has been fully approved"`; body: `"The final level of [Workflow Name] was removed. Your post has been automatically fully approved and will proceed as scheduled."`; channels: in-app, email, mobile push
- [ ] `PUT /approval-workflows/{id}` saves all changes together — mid-flight changes are applied after save confirmation; partial saves are not allowed

**Approval rule changes:**
- [ ] Rule changed "Everyone" → "Anyone" on a NOT STARTED level: applies to new posts and future traversal of that level on in-flight posts
- [ ] Rule changed "Everyone" → "Anyone" on an IN PROGRESS level AND at least one approval already exists: level auto-completes on affected in-flight posts; level advancement notifications dispatched
- [ ] Rule changed "Anyone" → "Everyone" on a COMPLETED level: no retroactive effect; stays completed
- [ ] Rule changed "Anyone" → "Everyone" on an IN PROGRESS level: all remaining members must now approve going forward (previously-received approvals are still counted)

**Name/title changes:**
- [ ] Workflow name renamed: all in-flight posts' `workflow_name` snapshot field is NOT updated (it's a historical snapshot); however, the live workflow name shown in active status panels is fetched from the workflow document itself, so it reflects immediately
- [ ] Level title renamed: similarly, in-flight `workflow_levels[].title` snapshot is not changed; the live title is shown from the current workflow definition

**Warning banner + save confirmation (API support):**
- [ ] `GET /approval-workflows/{id}` response includes `in_flight_posts_count` and `in_flight_level_breakdown[]` (per-level count of in-progress posts) so the frontend can display the mid-flight warning banner and build the save confirmation dialog copy
- [ ] `PUT /approval-workflows/{id}` accepts a `confirmed=true` query param; if `in_flight_posts_count > 0` and `confirmed` is not passed, return 422 with `requires_save_confirmation: true` and a `level_deletion_impact` object (number of posts that will auto-advance, number that will auto-complete) — frontend uses this data to build the confirmation dialog before retrying with `confirmed=true`

---

### Mock-ups:
N/A (backend logic)

---

### Impact on existing data:
- In-flight posts may have their `workflow_levels[]` member lists modified; all changes are applied together on save

---

### Impact on other products:
- Real-time updates via Pusher `feed_approval` events ensure the UI reflects changes without page refresh

---

### Dependencies:
- Depends on: [BE] Extend plans.approval schema for multi-level workflow state
- Depends on: [BE] Implement WorkflowApprovalBuilder (for level auto-completion logic)
- Depends on: [BE] Add new notification event types (for dispatching change-related notifications)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A
- [ ] Multilingual support — N/A (backend)
- [ ] UI theming support — N/A
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — mobile clients receiving Pusher events for plan approval state changes must handle new payload fields

---

## S-06: [BE] Fix rejected post silent status reset bug in ApprovalBuilder

### Description:
As a content creator, I want editing a rejected post to show me a confirmation dialog — not silently move the post back to draft or scheduled — so that I remain in control of the approval state and can choose whether to resend for approval or remove approval entirely.

---

### Workflow:
1. Creator opens a post that is in `rejected` approval status and edits it
2. Creator clicks "Save Draft" or "Save Schedule"
3. Instead of silently resetting the post status (current broken behavior), the API returns a flag indicating the post has rejected approval
4. Frontend intercepts this and shows the confirmation dialog (implemented in [FE] Build edit confirmation dialogs for posts in approval)
5. Creator chooses: "Resend for approval" or "Remove approval" or "Cancel"
6. API processes the chosen action

---

### Acceptance criteria:
- [ ] Bug is fixed: saving content on a rejected post no longer automatically changes `plans.status` to draft or scheduled
- [ ] `PUT /plans/{id}` (or equivalent save endpoint) response includes `approval_confirmation_required` object when the save operation involves a post with active or rejected approval state:
  ```json
  {
    "approval_confirmation_required": {
      "type": "rejected_single" | "rejected_workflow" | "active_single" | "active_workflow",
      "current_status": "...",
      "workflow_name": "...",   // null for single-user
      "current_level": 2,       // null for single-user
      "levels": [...]           // null for single-user
    }
  }
  ```
- [ ] `POST /plans/{id}/save-with-approval-action` endpoint processes the creator's dialog choice:
  - For rejected single-user: `action = "resend"` (resets all to pending, re-notifies) or `action = "remove"` (clears approval state)
  - For rejected workflow: `action = "restart"` (reset all levels), `action = "resume"` (reset only rejected level), or `action = "remove"` (clear approval)
  - For active single-user: `action = "renotify"`, `action = "keep"`, or `action = "remove"`
  - For active workflow: `action = "restart"`, `action = "renotify_current"`, `action = "keep"`, or `action = "remove"`
- [ ] Each action dispatches the appropriate notifications (content_updated_approval or rejection notifications)
- [ ] Existing behavior for non-approval posts (no approval state) is unchanged

---

### Mock-ups:
N/A (backend fix)

---

### Impact on existing data:
- Critical bug fix: existing rejected posts that were silently moved to draft/scheduled will now be properly handled going forward; historical data already in wrong state is not backfilled (out of scope)

---

### Impact on other products:
- iOS and Android save endpoints must also return `approval_confirmation_required` flag (they should share the same endpoint or have equivalent logic added)

---

### Dependencies:
- None (standalone bug fix, can be done independently of other stories)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A
- [ ] Multilingual support — N/A
- [ ] UI theming support — N/A
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — iOS/Android clients that allow editing posts must also handle the new `approval_confirmation_required` response to avoid silently resetting approval state

---

## S-07: [BE] Add "Allow Approval Workflow Management" permission to team member settings

### Description:
As a workspace admin, I want to control which non-admin team members can create and manage approval workflows — while still allowing all members to participate in approvals — so that workflow setup remains admin-controlled by default but can be delegated when needed.

---

### Workflow:
1. Admin opens a team member's settings (Collaborator or Approver role)
2. Admin sees a new "Manage Approval Workflows" toggle
3. When toggled on, the member gains the ability to create, edit, duplicate, delete, and set default workflows in Settings
4. Super Admins and Admins always have this ability regardless of the toggle (it is only shown for Collaborators/Approvers)
5. If the permission is revoked, the member's existing workflows remain active but they immediately lose edit access

---

### Acceptance criteria:
- [ ] Workspace member API must include the `allow_workflow_management` boolean field (default: `false` for Collaborators and Approvers; N/A for Super Admin and Admin who always have access)
- [ ] `PUT /workspace-members/{id}` endpoint accepts `allow_workflow_management` bool update
- [ ] All `approval-workflows` API endpoints (create, update, delete, duplicate, set-default) check the requesting user's permission: Super Admin and Admin always pass; Collaborators/Approvers pass only if `allow_workflow_management === true`; otherwise return 403
- [ ] Permission revocation takes immediate effect: existing JWT/session is not grandfathered; next API call is rejected
- [ ] Existing workflows created by a member whose permission is revoked: remain active and functional; the member simply cannot edit them via the API
- [ ] `GET /workspace-members/{id}` response includes `allow_workflow_management` field
- [ ] Super Admin and Admin role members: `allow_workflow_management` is not a settable field for them; attempts to set it are ignored (they always have access)

---

### Mock-ups:
N/A

---

### Impact on existing data:
- New field on workspace member records; defaults to `false` for all existing non-admin members (no existing member gets unintended access)

---

### Impact on other products:
- None; this is settings-layer permission only

---

### Dependencies:
- Depends on: [BE] Create approval_workflows collection and CRUD API (the endpoints being gated)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A
- [ ] Multilingual support — N/A
- [ ] UI theming support — N/A
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — no mobile impact; permission is settings-only

---

## S-08: [BE] Handle approver removal from workspace with in-flight approval impacts

### Description:
As a workspace admin, I want to be warned before removing a team member who is currently assigned as an approver on pending posts — and have their removal handled automatically without orphaning those posts — so that approval workflows continue without manual intervention.

---

### Workflow:
1. Admin attempts to remove a team member from the workspace
2. If the member is assigned as an approver on any in-flight posts (or is part of any workflow definition): show a warning with counts
3. Admin confirms removal
4. System automatically removes the member from all in-flight post approval assignments
5. Affected levels auto-continue (if the removed member was the only one on a "Everyone must approve" level, that level is considered passed)
6. Creator of each affected post is notified
7. Member is also removed from all workflow definitions (not just in-flight posts)

---

### Acceptance criteria:
- [ ] `DELETE /workspace-members/{id}` — before processing, checks for in-flight approvals and workflow memberships
- [ ] If affected, API returns a 200 preview response (not yet deleted) with: `in_flight_posts_count` (int), `workflow_definitions_count` (int), `requires_confirmation: true`
- [ ] `DELETE /workspace-members/{id}?confirmed=true` proceeds with full removal
- [ ] On confirmed removal: member is removed from `workflow_levels[].members[]` on all in-flight posts where they are currently `pending`
- [ ] For members already `approved` or `rejected` on completed levels: their record is preserved (not removed); they remain visible in the approval history
- [ ] Level auto-continue logic: after removing the member, if the remaining pending members at the current active level have all approved (or none are left), the level is marked complete and the workflow advances automatically
- [ ] Special case: if the removed member was the ONLY member on an "Everyone must approve" level AND all other members are also gone — that level is auto-passed
- [ ] `approver_removed_post` notification sent to the creator of each affected post (one notification per post)
- [ ] Member removed from all workflow definitions they appear in (from the `levels[].members[]` arrays)
- [ ] If removing the member leaves a workflow level with 0 members, that workflow is automatically set to `is_draft=true` (it can no longer be published until a new member is added)

---

### Mock-ups:
N/A

---

### Impact on existing data:
- Modifies in-flight plan approval records and workflow definitions; all changes are logged/auditable via the approval history

---

### Impact on other products:
- Pusher `feed_approval` events dispatched for affected posts to update real-time UI

---

### Dependencies:
- Depends on: [BE] Implement WorkflowApprovalBuilder (level auto-continue logic)
- Depends on: [BE] Add new notification event types (approver_removed_post notification)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A
- [ ] Multilingual support — N/A
- [ ] UI theming support — N/A
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — creators receive in-app notifications; no direct mobile action required

---

---

## S-09: [FE] Build Approval Workflows settings page (workflow list)

### Description:
As a workspace admin, I want a dedicated Approval Workflows page in Settings where I can see all my saved workflows, create new ones, and manage existing ones — so that I have a central place to control how content review is structured in my workspace.

---

### Workflow:
1. User navigates to Settings → Approval Workflows (new left nav item)
2. User sees a grid of workflow tiles plus one "Create a Workflow" tile at the start
3. User can click any workflow tile to edit it (opens Create/Edit builder — see S-10)
4. User can use the 3-dot menu on any tile to: Set as Default, Remove as Default, Edit, Duplicate, Delete
5. Deleting a workflow with in-flight posts shows a warning confirmation before proceeding
6. User with read-only access (no workflow management permission) sees the list but all management actions are hidden/disabled with an explanatory message

---

### Acceptance criteria:

**Layout and navigation:**
- [ ] "Approval Workflows" item appears in the Settings left sidebar navigation (below or near existing approval-related items)
- [ ] Page title is `Approval Workflows` (use `text-gray-900` heading, `text-xl font-semibold`)
- [ ] Workflows are displayed as a responsive grid of tiles (CSS Grid or Flexbox, wrapping on smaller screens)
- [ ] "Create a Workflow" tile is always the first/leftmost tile and is only visible to users with workflow management permission

**Empty state (no workflows yet):**
- [ ] Headline: "No Approval Workflows Yet"
- [ ] Subtext: "Create approval workflows to streamline your team's content review process. Set up multi-level approvals with custom rules for each stage."
- [ ] CTA button using `Button` component (primary variant): "Create Workflow"
- [ ] "Create a Workflow" tile still visible next to the empty state message

**Workflow tile:**
- [ ] Each tile shows: workflow name (truncated with tooltip if >30 chars), level count label ("3 Levels"), grouped member avatars from all levels (use `Avatar` component; overflow as "+N more" badge using `Badge` component)
- [ ] Draft workflows: show a `Badge` component with "Draft" label (yellow/warning variant) and subtext: "This workflow is incomplete and won't appear in the composer until all levels have at least one member."
- [ ] Default workflow: show a `Badge` component with "Default" label (primary-cs variant using `text-primary-cs-500` / `bg-primary-cs-50`)
- [ ] 3-dot menu on tile (use `ActionIcon` + `Dropdown` with `DropdownItem` components): "Set as Default" / "Remove as Default" / "Edit" / "Duplicate" / "Delete"
- [ ] "Set as Default" only shows when workflow is not already default and is not a draft
- [ ] "Remove as Default" only shows when this workflow is the current default

**Delete workflow — no in-flight posts:**
- [ ] Confirmation `Dialog` with:
  - Title: `Delete "[Workflow Name]"?`
  - Body: `This approval workflow will be permanently deleted. This action cannot be undone.`
  - Primary button (destructive/red variant `Button`): `Delete Workflow`
  - Secondary `Button`: `Cancel`

**Delete workflow — with in-flight posts:**
- [ ] Confirmation `Dialog` with:
  - Title: `Delete "[Workflow Name]"?`
  - Body: `[N] post(s) are currently in this approval workflow. Deleting it will convert those posts to single-user approval — the currently assigned approvers at each post's active level will remain assigned. You'll need to manage them manually.`
  - Primary `Button` (destructive): `Delete Workflow`
  - Secondary `Button`: `Cancel`

**No permission state:**
- [ ] "Create a Workflow" tile is hidden
- [ ] 3-dot menus on tiles are hidden
- [ ] `Alert` component (info variant) shown at top of page: "You don't have permission to manage approval workflows. Contact your workspace admin to make changes."

**Loading state:**
- [ ] Skeleton loading tiles (3–4 placeholder tiles) using Tailwind `animate-pulse` while API loads

---

### Mock-ups:
Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
- New Settings page; no impact on existing settings pages

---

### Impact on other products:
- None

---

### Dependencies:
- Depends on: [BE] Create approval_workflows collection and CRUD API
- Depends on: [BE] Add "Allow Approval Workflow Management" permission (for permission gating)
- Depends on: [Design] Approval Workflow Settings page (design spec)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — settings page should remain usable on tablet; tile grid wraps to fewer columns on smaller viewports
- [ ] Multilingual support — all UI copy must use i18n keys; no hardcoded strings
- [ ] UI theming support — uses `@contentstudio/ui` components; default badge uses `text-primary-cs-500` / `bg-primary-cs-50` (CSS variable-backed, white-label safe)
- [ ] White-label domains impact review — no white-label specific logic; theming is CSS variable-based
- [ ] Cross-product impact assessment — no Chrome extension or mobile impact

---

## S-10: [FE] Build Create/Edit Workflow builder with levels, members, and drag-and-drop

### Description:
As a workspace admin, I want a visual workflow builder where I can define named approval levels, assign team members by dragging them from a panel, set per-level approval rules, and reorder levels — so that setting up a complex approval chain is intuitive and fast.

---

### Workflow:
1. Admin clicks "Create a Workflow" tile or "Edit" on an existing workflow
2. Full-page builder opens with: workflow name input at top, level cards in the center, fixed right panel with team members
3. Admin drags members from the right panel into level cards OR clicks members to assign them
4. Admin sets approval rule per level (Everyone / Anyone) using radio buttons
5. Admin can add levels (up to 5), reorder via drag-and-drop, duplicate a level, or delete a level
6. Admin saves — if all levels have members: saved as published; if any level is empty: saved as Draft
7. Admin can explicitly click "Save as Draft" to save incomplete workflows without publishing

---

### Acceptance criteria:

**Page header:**
- [ ] Page title (create): "Create New Workflow"; page title (edit): "Edit Workflow"
- [ ] Workflow name `TextInput` with label "Workflow Name" and placeholder "Title of Workflow"; required field with validation error: "Workflow name is required"
- [ ] 3-dot menu in header (use `ActionIcon` + `Dropdown`): "Duplicate Workflow", "Delete Workflow"
- [ ] Draft `Badge` (warning variant) shown next to the name when workflow has empty levels; hover tooltip on draft badge: "This workflow is incomplete and won't appear in the composer until all levels have at least one member."
- [ ] Mid-flight warning `Alert` (warning variant) shown at top of page when editing a workflow with in-flight posts: "⚠ [N] post(s) are currently using this workflow. Some changes may affect those posts." (only in edit mode)
- [ ] Save `Button` (primary): "Save Workflow" (create) / "Save Changes" (edit)
- [ ] **Save confirmation dialog** — when clicking "Save Changes" on a workflow with in-flight posts, a `Dialog` is shown before the API call fires:
  - Base copy (no level deletions in this save):
    - Title: `"Save changes to "[Workflow Name]"?"`
    - Body: `"[N] post(s) are currently in approval using this workflow. Saving these changes will immediately affect their approval process."`
    - Primary `Button`: `"Save & Apply Changes"`
    - Secondary `Button`: `"Cancel"`
  - With level deletion(s) — append additional line:
    - `"[N] post(s) at the removed level(s) will automatically advance to the next level. If a removed level was the last, those posts will be fully approved."`
  - On "Save & Apply Changes": fire `PUT /approval-workflows/{id}?confirmed=true`; show `Loader` on button while saving
  - On "Cancel": dismiss dialog; no save; user stays in editor
  - This dialog is only shown in **edit** mode with in-flight posts; create mode and edit with zero in-flight posts save immediately without dialog
- [ ] "Save as Draft" `Button` (secondary ghost): "Save as Draft"
- [ ] Cancel `Button` (secondary): "Cancel"

**Level cards (center area):**
- [ ] Each level is shown as a card with: level number badge (e.g. "Level 01"), editable title `TextInput` with placeholder "Level 01:" / "Level 02:" etc., member assignment zone, approval rule selector, and 3-dot menu
- [ ] Member assignment zone shows assigned member `Avatar` components; overflow: "+N more" `Badge`; when empty: dashed-border placeholder with text "Drag and drop people from Members"
- [ ] Approval rule uses `Radio` components (group): "Post needs approval from:" label, then "Everyone" and "Anyone" options
  - "Everyone" tooltip (≥2 members): "All assigned team members must approve before this level is complete and the post moves to the next step."
  - "Everyone" tooltip (<2 members): "Add at least 2 team members to enable this option." (and the Everyone radio is disabled)
  - "Anyone" tooltip: "The level is complete as soon as any one team member approves. Other assigned members are notified that their review is no longer needed."
  - Use `CstPopup` for tooltips (no standalone tooltip component exists in the design library)
- [ ] Default rule on new level: "Everyone" pre-selected
- [ ] 3-dot menu per level (use `ActionIcon` + `Dropdown`): "Duplicate Level", "Delete Level"
- [ ] Delete level confirmation `Dialog`:
  - If level has **no in-progress posts**:
    - Title: "Delete this level?"
    - Body: "This will remove Level [N] and all its assigned members from the workflow."
    - Primary `Button`: "Delete Level"
    - Secondary `Button`: "Cancel"
  - If level has **in-progress posts** (not the last level):
    - Title: "Delete this level?"
    - Body: "[N] post(s) are currently at this level. Deleting it will automatically move them forward to the next level — those reviewers will be notified."
    - Primary `Button`: "Delete Level"
    - Secondary `Button`: "Cancel"
  - If level has **in-progress posts** and is the **last level**:
    - Title: "Delete this level?"
    - Body: "[N] post(s) are currently at this level and it's the final one. Deleting it will automatically fully approve those posts."
    - Primary `Button`: "Delete Level"
    - Secondary `Button`: "Cancel"
- [ ] Levels are draggable and reorderable using VueDraggable (or existing drag-and-drop library in the codebase); drag handle icon visible on each level card
- [ ] "Add New Level" `Button` (secondary/dashed outline style) at the bottom; disabled at 5 levels with tooltip: "Maximum 5 levels per workflow."

**Members right panel:**
- [ ] Fixed right sidebar panel with header "Members" and hint text "Drag & drop members to workflow levels"
- [ ] `SegmentedControl` component for filter tabs: "All" / "Team" / "Client"
- [ ] `SearchInput` component with placeholder "Search members..."
- [ ] Member list using `ListItem` components showing `Avatar` + name + role
- [ ] Empty search state: "No members found"
- [ ] Members are draggable from this panel into level cards
- [ ] Members can also be added by clicking — opens a mini-dropdown or clicking the member card directly assigns to the focused level
- [ ] Members already assigned to a level are shown with a visual indicator (check icon or highlighted) but can still be added to other levels

**Validation:**
- [ ] Workflow name: required; "Workflow name is required" if blank on save
- [ ] Empty level warning: on "Save Workflow" with empty levels, show inline validation message below the empty level card: "This level needs at least one team member before the workflow can be published." Offer "Save as Draft" alternative

**Loading/saving states:**
- [ ] Save button shows `Loader` component while saving
- [ ] Success: redirect to workflow list with a success toast: "Workflow saved successfully"
- [ ] Error: inline `Alert` (error variant) at top: "Something went wrong. Please try again."

---

### Mock-ups:
Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
- Editing an existing workflow applies mid-flight modification rules (handled by S-05)

---

### Impact on other products:
- None

---

### Dependencies:
- Depends on: [BE] Create approval_workflows collection and CRUD API
- Depends on: [BE] Implement workflow mid-flight modification rules (for edit behavior)
- Depends on: [FE] Build Approval Workflows settings page (workflow list) (navigates back to this page on save/cancel)
- Depends on: [Design] Approval Workflow Settings page (design spec)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — builder is a complex drag-and-drop interface; minimum usable at 1024px+ (tablet landscape and above); provide a simplified non-drag fallback for smaller screens if feasible
- [ ] Multilingual support — all copy via i18n keys; no hardcoded strings
- [ ] UI theming support — `@contentstudio/ui` components used throughout; `text-primary-cs-500`, `bg-primary-cs-50` for selected/default states
- [ ] White-label domains impact review — primary color theming is CSS variable-based; no hardcoded colors
- [ ] Cross-product impact assessment — no Chrome extension or mobile impact

---

## S-11: [FE] Migrate Send for Approval from center modal to right sidebar with Users and Approval Workflows tabs

### Description:
As a content creator, I want to send a post for approval using a right sidebar panel with two tabs — one for selecting individual users (the existing ad-hoc mode) and one for choosing a saved approval workflow — so that I can either use a pre-configured workflow with one click or still pick approvers manually when needed.

---

### Workflow:

**Via Composer:**
1. Creator finishes composing a post and clicks "Send for Approval"
2. A right sidebar panel slides in — titled "Send for Approval"
3. "Users" tab is shown first (existing ad-hoc mode, updated UI)
4. Creator can switch to "Approval Workflows" tab
5. On "Approval Workflows" tab: saved workflows shown as cards; default workflow is pre-selected; creator clicks a different card to switch selection
6. Creator optionally types a note in the Notes field at the bottom
7. Creator clicks "Send" — post enters approval process; sidebar closes; planner shows "Pending Approval" badge

**Via Planner bulk actions:**
- Same sidebar opens when creator selects multiple posts and clicks "Send for Approval" from the bulk action menu

**Via Automations:**
- Same sidebar component used in Automations "Send for Approval" action

---

### Acceptance criteria:

**Sidebar panel (new `SendForApprovalSidebar.vue` component, replaces `ApprovalModal.vue`):**
- [ ] Right sidebar using `CstSidebar` component; slides in from the right; does not overlay/replace the composer but sits alongside it
- [ ] Panel title: "Send for Approval" with an ℹ `ActionIcon` (tooltip using `CstPopup`: "Send this post to your team for review before publishing. The post won't be published until the approval process is complete.")
- [ ] Close `ActionIcon` (✕) at top right of sidebar
- [ ] Two tabs using `Tabs` component: "Users" (tab 1) and "Approval Workflows" (tab 2)
- [ ] "Send" `Button` (primary) and "Cancel" `Button` (secondary/ghost) — fixed at bottom of sidebar

**Tab 1 — Users (ad-hoc mode, updated UI):**
- [ ] `SearchInput` at top with placeholder: "Search by name..."
- [ ] No "hotlist" row (removed per design decision)
- [ ] Member list using `ListItem` components: `Avatar` + name + role; ordered alphabetically
- [ ] `Checkbox` component on each list item for selection
- [ ] Empty search state copy: "No team members found"
- [ ] Approval rule section below member list, label: "Post needs approval from:"; `Radio` group: "Everyone" and "Anyone"
  - "Everyone" disabled with tooltip when <2 members selected: "Select at least 2 team members to use this option."
  - "Everyone" tooltip (≥2 selected): "All selected team members must approve before the post is scheduled."
  - "Anyone" tooltip: "The post will be scheduled as soon as any one selected team member approves."
- [ ] Notes `Textarea` at bottom with placeholder: "Add a note for the approver(s)..."
- [ ] Validation: "Please select at least one approver." shown as inline error if user tries to Send with no one selected
- [ ] Restricted account warning: if any selected member has restricted social account access, show a warning icon on their list item with `CstPopup` tooltip: "This team member only has access to certain social accounts. Make sure the post's accounts align with their permissions."
- [ ] Permission warning `Dialog` on Send if restricted members are selected:
  - Title: "Restricted Access"
  - Body: "Some selected team members only have access to specific social accounts. Sending this post for approval may limit their ability to review all content. Do you want to proceed?"
  - Primary `Button`: "Send Anyway"
  - Secondary `Button`: "Cancel"

**Tab 2 — Approval Workflows:**
- [ ] Workflows loaded from API as a list of cards (vertically stacked, scrollable)
- [ ] Default workflow card is pre-selected when this tab opens (selected state: `border-primary-cs-200 bg-primary-cs-50/50`)
- [ ] Selected workflow has a visible selected indicator (border highlight using `border-primary-cs-200`)
- [ ] **Collapsed workflow card** shows: workflow name, level count ("3 Levels"), member avatars grouped by level with a visual separator/dash between levels (use `Avatar` components with stacking, `Badge` for overflow "+N more")
- [ ] "Default" `Badge` (primary-cs variant) on the default workflow card
- [ ] Expand/collapse chevron `ActionIcon` on each card; clicking expands to show level detail
- [ ] **Expanded workflow card** shows: each level as a row with "Level [N]: [Title]" label, member `Avatar` components for that level, and a rule label badge ("Everyone" / "Anyone" using `Badge` component)
- [ ] "Edit in Settings →" link text below the workflow list (above Notes); clicking shows a `Dialog`:
  - Title: "Leave the composer?"
  - Body: "You'll be redirected to Approval Workflow settings. Your composer content will be minimized but not lost — it will be here when you return."
  - Primary `Button`: "Go to Settings"
  - Secondary `Button`: "Stay Here"
  - On "Go to Settings": minimizes composer, navigates to Settings → Approval Workflows
- [ ] Notes `Textarea` at bottom with placeholder: "Add a note for this approval request..."
- [ ] Empty state (no workflows created):
  - Copy: "No approval workflows yet."
  - Subtext: "Create one in Settings → Approval Workflows."
  - Shows "Edit in Settings →" link prominently
- [ ] Validation on Send: "Please select a workflow." if no workflow is selected (can happen if user somehow deselects all)

**"Anyone can approve" level — 10+ member handling:**
- [ ] Show first N avatars (fitting the collapsed card width) + "+N more" `Badge`; hovering the "+N more" badge shows a `CstPopup` with scrollable full member list

**Loading state:**
- [ ] `Loader` component while workflows are loading; skeleton list while members load on Users tab

**Override warning (when post already has active approval):**
- [ ] When the sidebar opens for a post that already has an active approval, an `Alert` (warning variant) is shown at the top — before the tabs — with context-specific copy:
  - Ad-hoc internal: `"This post is already in an approval process. Starting a new one will cancel it — currently assigned approvers will be notified."`
  - Workflow internal: `"This post is in [Workflow Name] (Level [N]). Starting a new approval will cancel it — Level [N] approvers will be notified."`
  - External share link: `"This post has an active external approval via share link. Starting an internal approval will cancel it."`
- [ ] The warning is informational — the user can still proceed by filling in the sidebar and clicking Send; no extra confirmation step is needed for single-post override
- [ ] On Send with an override: the existing approval is cancelled server-side, cancellation notifications are dispatched, new approval begins (handled by S-22 backend)

**Existing `ApprovalModal.vue`:**
- [ ] Old center modal component is removed from Composer, Planner bulk actions, and Automations
- [ ] All three entry points now use the new `SendForApprovalSidebar.vue` component
- [ ] Backward compatibility: any existing state management in `useApproval.js` and `usePublishApprovalStore.ts` is extended, not replaced

---

### Mock-ups:
Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
- Existing ad-hoc approval logic is preserved in the Users tab; no behavioral change for existing single-level approvals

---

### Impact on other products:
- Automations "Send for Approval" action must also be updated to use this new sidebar component
- Planner bulk actions "Send for Approval" must use this component

---

### Dependencies:
- Depends on: [BE] Create approval_workflows collection and CRUD API
- Depends on: [BE] Extend plans.approval schema for multi-level workflow state
- Depends on: [Design] Approval Workflow Settings page, Create/Edit builder, and Send for Approval sidebar
- The new sidebar must be complete before Automations and Planner bulk action stories can reference it

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — sidebar is web-only in this story; mobile handled in S-18/S-20; ensure sidebar renders cleanly at 1024px+ minimum
- [ ] Multilingual support — all copy via i18n keys; Notes field is user input (untranslated); validation messages translated
- [ ] UI theming support — selected workflow card uses `bg-primary-cs-50/50` and `border-primary-cs-200`; Default badge uses `text-primary-cs-500` — all CSS variable-backed
- [ ] White-label domains impact review — no hardcoded colors
- [ ] Cross-product impact assessment — Automations team must be included in testing; Planner bulk actions must be verified

---

## S-12: [FE] Build approval status tracking panel (planner hover popup and post preview modal)

### Description:
As a content creator or approver, I want to see the full real-time approval status of any post — which level it's at, who has approved or is still pending, with the ability to re-notify a specific person or revoke my own approval — so that I always know exactly where a post stands and can take action if it's stalling.

---

### Workflow:
1. **In Planner (calendar, feed, list views):** Posts in approval show a status badge (e.g., "Pending Approval 1/3") on the post card; hovering/clicking the badge opens a small popup showing the level breakdown and per-user statuses with Re-notify buttons
2. **In Post Preview Modal (full detail view):** Approval Status panel is shown; full level-by-level breakdown; approver can Approve, Reject, Re-notify, or Revoke their own approval from here
3. Approver clicks "Approve" → confirmation dialog → approval recorded; if final approval in level, status advances
4. Approver clicks "Re-notify" → cooldown enforced; button disabled if within 2 hours with tooltip showing next available time
5. Approver clicks "Revoke approval" → confirmation dialog → if eligible, resets their status to pending

---

### Acceptance criteria:

**Planner post card badge:**
- [ ] Posts in any approval state (Pending Approval, Missed Review, Rejected) show a status badge on the post card using `Badge` component
- [ ] Workflow posts: badge shows progress counter "[approved]/[total at current level]" (e.g., "1/3")
- [ ] Badge color: Pending → orange/warning; Rejected → red/error; Missed Review → orange/warning
- [ ] Hover/click on badge opens a `CstPopup` popup (not a full modal) showing the compact approval status

**Planner hover popup:**
- [ ] Popup header: workflow name (if workflow-based) or "Approval Status" (if ad-hoc)
- [ ] Each level shown as a row: "Level [N]: [Title]" label + status badge ("Completed" in green, "In Progress" in orange, "Not Started" in gray) — use `Badge` component for each
- [ ] Per-user row under each level: `Avatar` + name + status chip: "Approved on [date]" (green), "Rejected on [date]" (red), "Awaiting approval" (gray)
- [ ] "Re-notify" `Button` (secondary small) per pending approver
- [ ] Re-notify disabled state: `Button` disabled; `CstPopup` tooltip: "Re-notified less than 2 hours ago. Available again at [HH:MM AM/PM]."
- [ ] Popup is dismissible by clicking outside or pressing Escape

**Post Preview Modal — Approval Status panel:**
- [ ] Panel title: "Approval Status"; workflow name subheading below (if workflow-based) or "Ad-hoc Approval"
- [ ] Creator's note section (if note was provided): "Request note from [Creator name]:" label, followed by the note text in a styled block
- [ ] Same level-by-level breakdown as hover popup but with more detail: per-user statuses, timestamps, and comments
- [ ] **Approve `Button`** (primary/success): visible to the current user if they are a pending approver at the current active level
  - Clicking opens a `Dialog`:
    - Title: "Approve Post"
    - Body: "You're about to approve this post. You can add an optional comment."
    - `Textarea` with placeholder: "Add your comment here (optional)"
    - Primary `Button`: "Yes, Approve"
    - Secondary `Button`: "Cancel"
- [ ] **Reject `Button`** (secondary/destructive): visible to the current user if they are a pending approver at the current active level
  - Clicking opens a `Dialog`:
    - Title: "Reject Post"
    - Body: "You're about to reject this post. Please add a comment explaining what needs to change."
    - `Textarea` with placeholder: "Explain what needs to change..." (required field)
    - Primary `Button` (destructive): "Reject Post"
    - Secondary `Button`: "Cancel"
    - Validation: reject comment is required; show "Please add a comment explaining what needs to change." if left empty
- [ ] **"Revoke approval" link/button**: visible to the current user ONLY if they have approved AND their level is the currently active level AND the post is not fully approved
  - Revoke unavailable tooltip (level locked via "Anyone" rule): "This level has already advanced — your approval can no longer be revoked."
  - Clicking opens a `Dialog`:
    - Title: "Revoke your approval?"
    - Body: "Your approval will be reset to pending. [Creator name] and other reviewers at this level will be notified."
    - Primary `Button`: "Revoke"
    - Secondary `Button`: "Cancel"
- [ ] **Re-notify per pending approver**: "Re-notify" `Button` next to each pending approver's row; same cooldown behavior as hover popup
- [ ] Missed Review prompt: when approver acts on a Missed Review post, after successful approve/reject API response, a `Dialog` is shown:
  - Title: "This post has missed its scheduled time"
  - Body: "Would you like to reschedule this post now?"
  - Primary `Button`: "Reschedule"
  - Secondary `Button`: "No, thanks"
  - "Reschedule" opens the existing reschedule picker (replicate existing behavior from the non-approval Missed Review flow)
- [ ] **Bulk approve/reject**: in Pending Approvals view, approver can select multiple posts and use bulk approve/reject:
  - Bulk approve `Dialog`:
    - Title: "Approve [N] Posts"
    - Body: "You're about to approve [N] posts. Add an optional comment."
    - `Textarea` with placeholder: "Add your comment here (optional)"
    - Primary `Button`: "Yes, Approve"
  - Bulk reject: same as single reject but title "Reject [N] Posts"; single comment applies to all
- [ ] Fully approved state: panel footer: "All levels completed — this post is approved." (no Approve/Reject/Revoke buttons)
- [ ] Post-publish state: panel footer: "This post has been published. Approval history is preserved for reference." (read-only, no actions)

**Real-time updates:**
- [ ] Approval status panel updates in real-time via existing Pusher `feed_approval` socket events; no page refresh needed
- [ ] Status badges on post cards also update in real-time

---

### Mock-ups:
Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
- Extends existing `PlanApprovalStatus.vue` and `FeedViewApprovalStatus.vue` components
- Existing single-level ad-hoc approval status display must continue to work unchanged

---

### Impact on other products:
- None (web only; mobile has its own approval status views)

---

### Dependencies:
- Depends on: [BE] Implement WorkflowApprovalBuilder (approve/reject/revoke/re-notify endpoints)
- Depends on: [BE] Add new notification event types (re-notify cooldown API)
- Depends on: [BE] Extend plans.approval schema (workflow_levels data in API response)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — planner is web-only; sidebar panel should be readable at 1024px+
- [ ] Multilingual support — all copy via i18n; timestamps use locale-aware date formatting
- [ ] UI theming support — status colors use Tailwind utility classes for green/orange/red (not primary-cs — these are semantic status colors, not brand primary color); `Badge` and `Button` components from `@contentstudio/ui`
- [ ] White-label domains impact review — semantic status colors (green/red/orange) are not primary brand colors; no white-label concern
- [ ] Cross-product impact assessment — no Chrome extension or mobile impact

---

## S-13: [FE] Build edit confirmation dialogs for posts in approval (all 4 scenarios)

### Description:
As a content creator, I want to see a clear confirmation dialog whenever I save changes to a post that is in approval or was rejected — giving me explicit control over what happens to the approval state — so that I never accidentally reset or break an approval without knowing about it.

---

### Workflow:
1. Creator edits a post that has an active or rejected approval state
2. Creator clicks "Save Draft" or "Save Schedule"
3. API returns `approval_confirmation_required` with the current approval state
4. Frontend intercepts this and shows the appropriate confirmation dialog (one of 4 scenarios)
5. Creator selects an option (radio button) and clicks "Confirm" — or clicks "Cancel" to return to editing without saving
6. API processes the save + the chosen approval action atomically

---

### Acceptance criteria:

**Scenario 1 — Post in active single-user approval:**
- [ ] `Dialog` component with:
  - Title: "This post is pending approval"
  - Body: Bulleted list of each approver with their status (e.g., "[Avatar] [User A] — Approved", "[Avatar] [User B] — Awaiting approval") using `Avatar` component per user
  - Question text: "You've updated the content. What would you like to do?"
  - `Radio` group (use `Radio` component):
    - "Re-notify approvers" — sub-label: "All pending and approved statuses reset. Approvers will be re-notified."
    - "Keep current approval status" — sub-label: "Save changes without affecting the approval process."
    - "Remove approval" — sub-label: "The post exits the approval process and reverts to draft/scheduled."
  - Primary `Button`: "Confirm"
  - Cancel link/text: "Cancel"

**Scenario 2 — Post in active workflow approval:**
- [ ] `Dialog` component with:
  - Title: "This post is in an approval workflow"
  - Body: Workflow name + expandable level breakdown using `Collapsible` component per level ("Level 1: Completed", "Level 2: In Progress — 1/3 approved, 1 pending", "Level 3: Not Started")
  - Question text: "You've updated the content. What would you like to do?"
  - `Radio` group:
    - "Restart from Level 1" — sub-label: "All approvals reset. The workflow starts over from the beginning."
    - "Re-notify current level" — sub-label: "Only Level [N] resets. Previous levels stay approved."
    - "Keep current approval status" — sub-label: "Save changes without affecting the workflow."
    - "Remove approval" — sub-label: "The post exits the approval workflow and reverts to draft/scheduled."
  - Primary `Button`: "Confirm"
  - Cancel link: "Cancel"

**Scenario 3 — Rejected post (single-user) — bug fix:**
- [ ] `Dialog` component with:
  - Title: "This post was rejected"
  - Body: Bulleted list of each approver with their status (showing "Rejected" in red, "Approved" in green per user); use `Badge` for status
  - Question text: "You've updated the content. What would you like to do?"
  - `Radio` group:
    - "Resend for approval" — sub-label: "All statuses reset to pending. Approvers will be re-notified."
    - "Remove approval" — sub-label: "The post moves to draft/scheduled without requiring approval."
  - Primary `Button`: "Confirm"
  - Cancel link: "Cancel"

**Scenario 4 — Rejected post (workflow):**
- [ ] `Dialog` component with:
  - Title: "This post was rejected at Level [N] of [Workflow Name]"
  - Body: Level breakdown showing which level rejected and who rejected with their comment (if any)
  - Question text: "You've updated the content. What would you like to do?"
  - `Radio` group:
    - "Restart from Level 1" — sub-label: "All approvals reset. The workflow starts over from the beginning."
    - "Resume from Level [N]" — sub-label: "Only Level [N] resets. Previous levels' approvals are preserved."
    - "Remove approval" — sub-label: "The post moves to draft/scheduled without requiring approval."
  - Primary `Button`: "Confirm"
  - Cancel link: "Cancel"

**Scenario 5 — Missed Review:**
- [ ] Behaves identically to Scenario 1 (single-user) or Scenario 2 (workflow) depending on approval type; same dialogs are shown

**General behavior across all scenarios:**
- [ ] First option in each dialog is pre-selected (sensible default)
- [ ] "Confirm" button is disabled until a radio option is selected (if no pre-selection desired by design, keep enabled with first option)
- [ ] "Cancel" does NOT save the post; editor stays open with unsaved changes
- [ ] On "Confirm": post is saved AND the approval action is processed in a single API call; loading state shown on Confirm button (use `Loader` component)
- [ ] Success: `CstToast` success notification: "Post saved successfully"
- [ ] Error: `Alert` (error variant) inside dialog: "Something went wrong. Please try again."
- [ ] Dialogs are triggered from the composer save flow; `useApproval.js` composable and `usePublishApprovalStore.ts` must be updated to intercept the save response and show the correct dialog

---

### Mock-ups:
Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
- Fixes existing bug: rejected posts no longer silently reset to draft/scheduled on save

---

### Impact on other products:
- None (web only; mobile edit flow is a separate story if needed)

---

### Dependencies:
- Depends on: [BE] Fix rejected post silent status reset bug in ApprovalBuilder (provides the `approval_confirmation_required` API flag)
- Depends on: [BE] Implement WorkflowApprovalBuilder (save-with-approval-action endpoint)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — dialogs use `Modal`/`Dialog` components which are mobile-aware; no special work needed
- [ ] Multilingual support — all copy via i18n; dynamic values (workflow name, level number, user names) injected at runtime
- [ ] UI theming support — `Radio`, `Dialog`, `Button`, `Badge`, `Avatar`, `Collapsible` all from `@contentstudio/ui`; no hardcoded colors
- [ ] White-label domains impact review — no brand color usage in these dialogs
- [ ] Cross-product impact assessment — no Chrome extension or mobile impact

---

## S-14: [FE] Build dedicated Approval Notifications icon and panel in the top bar

### Description:
As a content creator or approver, I want a dedicated approval notification area in the top bar — separate from general notifications — so that time-sensitive approval activity is always visible and I never miss a post waiting for my review.

---

### Workflow:
1. Approval events generate notifications that appear in a new dedicated icon in the top bar
2. A numeric unread badge on the icon shows pending count
3. User clicks the icon → panel slides open showing approval activity feed
4. User can filter by: All / Pending / Rejected / Updated / Comments tabs
5. Clicking a notification navigates to the relevant post in Planner
6. User can mark all as read
7. Cross-workspace: workspace switcher and workspace tiles show notification counts for approval activity in other workspaces

---

### Acceptance criteria:

**Top bar icon:**
- [ ] New approval notification `ActionIcon` added to the top bar, positioned to the LEFT of the existing general notifications bell icon
- [ ] Icon uses a relevant icon from the design system (e.g., checkmark/approval icon via `Icon` component)
- [ ] Unread count shown as a numeric `Badge` on the icon (same visual pattern as the general notifications bell badge)
- [ ] Hover `CstPopup` tooltip: "Approval notifications"
- [ ] Clicking icon opens/closes the approval panel (panel behavior same as general notifications panel — slides out or dropdown)

**Approval panel:**
- [ ] Panel title: "Approvals"
- [ ] `Tabs` component (or `SegmentedControl`) with filter tabs: "All" / "Pending" / "Rejected" / "Updated" / "Comments"
- [ ] "Mark all as read" `Button` (secondary ghost, small) at top right of panel
- [ ] Each notification item shows: icon (Approved/Pending/Rejected/Comment), actor avatar (`Avatar`), notification text (dynamic, from backend), timestamp (relative time e.g. "2h ago"), unread indicator dot
- [ ] Clicking a notification item: navigates to Planner with the relevant post focused/highlighted; closes the panel; marks notification as read
- [ ] "View post →" link text on each notification item (secondary link style)
- [ ] Filter tabs: "Pending" shows only unresolved approval requests assigned to the current user; "Rejected" shows rejected posts; "Updated" shows posts modified during approval; "Comments" shows new comments on posts in approval
- [ ] Clicking a category filter → navigates to Planner with the appropriate filter applied (the panel closes)

**Empty state:**
- [ ] Headline: "No approval activity yet"
- [ ] Subtext: "When posts are sent for approval or reviewed, activity will appear here."
- [ ] No illustration needed (matches existing notifications panel style)

**Loading state:**
- [ ] Skeleton notification items using `animate-pulse` while loading

**Cross-workspace notification badges:**
- [ ] Workspace switcher dropdown: each workspace tile shows an approval notification badge count (if > 0) — small `Badge` component (separate from the existing general notification badge)
- [ ] Workspace grid page: individual workspace tiles show approval notification count badge
- [ ] Clicking a workspace badge navigates to that workspace and opens the approval panel filtered to the relevant activity

**Real-time updates:**
- [ ] Approval panel updates in real-time via existing Pusher infrastructure; new approval events appear without page refresh
- [ ] Unread badge count on the icon updates in real-time

---

### Mock-ups:
Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
- New notifications stored alongside existing notifications but tagged as approval-type; general notifications panel is unchanged

---

### Impact on other products:
- Cross-workspace badges appear in workspace switcher used by all users; must not interfere with existing workspace switcher behavior

---

### Dependencies:
- Depends on: [BE] Add new notification event types for multi-tiered approval (for notification data)
- Depends on: [BE] Extend plans.approval schema (for real-time Pusher event data)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — top bar panel is web-only; mobile has push notifications; panel must be usable at 1024px+
- [ ] Multilingual support — all static copy via i18n; notification copy (coming from backend locale strings) also translatable
- [ ] UI theming support — `ActionIcon`, `Badge`, `Tabs`, `Avatar`, `Button` all from `@contentstudio/ui`; approval badge uses same styling as general notifications badge
- [ ] White-label domains impact review — icon and panel must be compatible with white-label top bar customization
- [ ] Cross-product impact assessment — workspace switcher is used by all products; badge additions must not break existing workspace switcher behavior for any product

---

## S-15: [FE] Add "Manage Approval Workflows" permission toggle to Team Member Settings

### Description:
As a workspace admin, I want a toggle in team member settings to grant or revoke approval workflow management access for Collaborators and Approvers — so that admins keep control over workflow setup by default while being able to delegate when needed.

---

### Workflow:
1. Admin opens Settings → Team Members → clicks on a Collaborator or Approver member
2. In the member's permission settings, admin sees a new "Manage Approval Workflows" toggle
3. Admin toggles it on → member can now create, edit, delete, and set default workflows
4. Admin toggles it off → member immediately loses workflow management access (existing workflows remain active)
5. Super Admin and Admin role members do not show this toggle (they always have access)

---

### Acceptance criteria:
- [ ] New `Switch` component (`@contentstudio/ui`) added to the team member permissions settings form for Collaborator and Approver roles only
- [ ] Toggle label: "Manage Approval Workflows"
- [ ] Toggle description below label: "Allow this member to create, edit, duplicate, delete, and set default approval workflows."
- [ ] Off state helper text shown below toggle: "Only admins can manage workflows" (when toggle is off)
- [ ] Toggle is hidden for Super Admin and Admin role members (they always have access; no toggle shown)
- [ ] Toggle state reflects current `allow_workflow_management` value from the API
- [ ] On toggle change: `PUT /workspace-members/{id}` called immediately (optimistic update with rollback on error)
- [ ] Success: `CstToast`: "Permissions updated successfully"
- [ ] Error: `Alert` (error) inline: "Failed to update permissions. Please try again." — toggle snaps back to previous state
- [ ] Loading state: `Loader` component shown while save is in progress; toggle is disabled during save

---

### Mock-ups:
Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
- New field in team member settings form; existing permission toggles unchanged

---

### Impact on other products:
- None

---

### Dependencies:
- Depends on: [BE] Add "Allow Approval Workflow Management" permission to team member settings

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — team member settings page adapts; toggle renders correctly at all breakpoints
- [ ] Multilingual support — all copy via i18n keys
- [ ] UI theming support — `Switch` component from `@contentstudio/ui`
- [ ] White-label domains impact review — no brand colors used in settings toggles
- [ ] Cross-product impact assessment — no mobile or Chrome extension impact

---

## S-16: [FE] Update Planner post status badges and filter labels for multi-level approval

### Description:
As a content creator, I want approval status badges on post cards in all Planner views (calendar, feed, list) and accurate filter labels — so that I can see at a glance which posts are in approval and filter my view to focus on what needs attention.

---

### Workflow:
1. User opens the Planner in any view (calendar, feed, list)
2. Posts in approval states show labeled status badges on their cards
3. User opens the filter sidebar and can filter by: "Pending Approval", "Missed Review", "Rejected"
4. "Under Review" label (if still present) is renamed to "Pending Approval"
5. Existing "My pending approvals" and "My requested approvals" custom views continue to work

---

### Acceptance criteria:

**Status badge updates:**
- [ ] All Planner views (calendar, feed, list) show status badges on post cards for approval states using `Badge` component
- [ ] Status labels and badge colors:
  - `pending_approval` / `review`: Label "Pending Approval" — orange/warning `Badge`
  - `missed_review`: Label "Missed Review" — orange/warning `Badge` with distinct visual treatment (e.g., dashed border or different icon) to differentiate from normal pending
  - `rejected`: Label "Rejected" — red/error `Badge`
- [ ] Workflow posts show an additional progress indicator on the badge: e.g., "Pending Approval · Level 2" or progress count "1/3" — exact display to match Figma design
- [ ] Fully approved posts show no approval badge (they revert to their underlying Scheduled/Draft status)

**Planner filter sidebar:**
- [ ] "Under Review" filter label renamed to "Pending Approval" throughout the filter UI and i18n keys (`planner.filter_sidebar.plan_approval_status.*`)
- [ ] Three approval status filters available: "Pending Approval", "Missed Review", "Rejected" (no "Partially Approved" — this status does not exist)
- [ ] Existing filter queries for `pending_approval` / `review` status work with the updated label
- [ ] Filter count badges on each filter option update in real-time via existing Pusher mechanism

**Planner "Approver Experience" — filtered view:**
- [ ] When an approver clicks a notification and is taken to Planner, only the post(s) assigned to them for review are shown (filtered view)
- [ ] Filter tag shown at top of feed: "Showing posts needing your approval" with a "Clear filter" link
- [ ] Multiple notified posts: all shown in feed view, filtered

**Custom views (existing):**
- [ ] "My pending approvals" and "My requested approvals" custom views continue to work; update their query to handle new workflow-based approval statuses
- [ ] These views now also surface workflow-based posts (not just single-user approval posts)

**Rescheduling approved posts:**
- [ ] Fully approved posts: the reschedule date/time picker works normally; no re-approval dialog shown; no notifications sent; existing save flow is unchanged
- [ ] Verify this is already the behavior (if broken, fix; if working, add a regression test in ACs)

---

### Mock-ups:
Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
- The "Under Review" → "Pending Approval" rename is a display-only change; underlying status values in the database are unchanged
- Existing posts with approval status continue to display correctly with updated labels

---

### Impact on other products:
- None (Planner is web-only in this context)

---

### Dependencies:
- Depends on: [BE] Extend plans.approval schema (for workflow level data in API response)
- Must be done after the labels/filters are confirmed in Figma designs

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — Planner is primarily a web interface; status labels should display cleanly on tablet (1024px)
- [ ] Multilingual support — all label changes via i18n; verify `planner.filter_sidebar.plan_approval_status` i18n file updated; no hardcoded label strings
- [ ] UI theming support — `Badge` component from `@contentstudio/ui`; status colors are semantic (not brand primary); use Tailwind `text-orange-500`, `text-red-500`, `text-green-500` (these are fixed semantic colors, not themed)
- [ ] White-label domains impact review — semantic status colors are not white-labeled; no concern
- [ ] Cross-product impact assessment — Planner filter API changes may affect Chrome extension if it uses approval filter queries; verify Chrome extension compatibility

---

## S-17: [Design] Approval Workflow Settings, Create/Edit builder, and Send for Approval sidebar

### Description:
As a designer, I want to produce final, implementation-ready designs for the Approval Workflows Settings page, the Create/Edit Workflow builder, and the Send for Approval right sidebar panel — so that the development team has precise, pixel-complete reference designs that use ContentStudio's design library components consistently throughout the entire feature.

---

### Workflow:
1. Designer reviews approved PRD (docs/features/multi-tiered-approval-workflow/03-prd.md) and workflow doc (02-workflow.md)
2. Designer produces designs in Figma covering all screens, states, and edge cases listed in the Acceptance Criteria
3. Designer shares Figma link with the product team for review
4. Product reviews designs, requests changes if needed
5. Final designs linked in the Shortcut epic and referenced by all development stories

---

### Acceptance criteria:

**Settings — Approval Workflows list page:**
- [ ] Workflow list page (populated state): tile grid with workflow cards showing name, level count, member avatars, Default badge, Draft badge
- [ ] Empty state: "No Approval Workflows Yet" with subtext and "Create Workflow" button
- [ ] "Create a Workflow" tile design
- [ ] 3-dot menu on workflow tile (expanded): Set as Default / Remove as Default / Edit / Duplicate / Delete
- [ ] Delete confirmation dialog (no in-flight posts variant)
- [ ] Delete confirmation dialog (in-flight posts variant with count)
- [ ] Read-only state (no permission): Alert banner

**Settings — Create/Edit Workflow builder:**
- [ ] Full-page builder layout: header (name, save buttons), level cards center, members right panel
- [ ] Level card: all states — empty (dashed placeholder), populated (member avatars), everyone/anyone radio, drag handle, expand/collapse
- [ ] Level card 3-dot menu: Duplicate Level / Delete Level
- [ ] Delete level confirmation dialog
- [ ] Members right panel: SegmentedControl tabs, search, member list
- [ ] Draft badge on workflow when levels are empty
- [ ] Mid-flight warning Alert banner (editing workflow with in-flight posts)
- [ ] Validation states: empty level error

**Send for Approval — right sidebar:**
- [ ] Sidebar panel: header with title + info icon + close button
- [ ] Tab 1 (Users): search, member list with checkboxes, Everyone/Anyone radios, notes textarea, Send/Cancel buttons
- [ ] Tab 2 (Approval Workflows): workflow card list (collapsed + expanded states), default pre-selected, Edit in Settings link, notes textarea
- [ ] Collapsed workflow card: name, level count, grouped avatar stacks with dashes
- [ ] Expanded workflow card: level rows with "Level N: Title", avatars, Everyone/Anyone badge
- [ ] Empty state (no workflows): "No approval workflows yet" with Settings link
- [ ] Edit redirect confirmation dialog: "Leave the composer?"
- [ ] Restricted member warning tooltip

**Approval status tracking panel:**
- [ ] Planner post card badge: Pending Approval, Missed Review, Rejected variants
- [ ] Hover popup: level breakdown with per-user statuses and Re-notify button
- [ ] Post preview modal: full approval status panel; Approve/Reject/Revoke/Re-notify states
- [ ] Confirmation dialogs: Approve, Reject, Revoke
- [ ] Missed Review reschedule prompt
- [ ] Bulk approve/reject dialogs

**Edit confirmation dialogs (all 4 scenarios):**
- [ ] Scenario 1: Active single-user approval dialog
- [ ] Scenario 2: Active workflow approval dialog (with level breakdown)
- [ ] Scenario 3: Rejected post single-user dialog
- [ ] Scenario 4: Rejected post workflow dialog

**Approval Notifications panel:**
- [ ] Top bar icon position (left of bell)
- [ ] Panel: title, filter tabs, notification item (with icon/avatar/text/timestamp), Mark all as read
- [ ] Empty state
- [ ] Cross-workspace badge on workspace switcher

**Design system notes:**
- [ ] All designs use `@contentstudio/ui` components (Button, Tabs, SegmentedControl, Checkbox, Radio, Avatar, Badge, Dialog, Collapsible, Switch, TextInput, SearchInput, Textarea, Alert, ActionIcon, Loader, Icon)
- [ ] No hardcoded color values — use `text-primary-cs-500`, `bg-primary-cs-50`, etc. for brand colors; gray utilities for neutral UI
- [ ] Consistent "Level N" naming (not "Step N") throughout all designs

---

### Mock-ups:
Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
N/A (design story)

---

### Impact on other products:
- Mobile designs (iOS and Android) are handled separately by mobile designers; this story covers web app only

---

### Dependencies:
- PRD (03-prd.md) and Workflow (02-workflow.md) must be final (already approved)
- Design system component library must be accessible in Figma

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — designs should include responsive breakpoints for 1024px and 1280px+ viewports for the Settings pages and sidebar
- [ ] Multilingual support — designs must leave room for text expansion (translated strings can be 30–40% longer than English)
- [ ] UI theming support — all component usage must reference design library components; no custom overrides with hardcoded brand colors
- [ ] White-label domains impact review — no white-label specific design variants needed; CSS variable theming handles it
- [ ] Cross-product impact assessment — N/A for design story

---

## S-18: [iOS] Add Approval Workflows tab to Send for Approval in Composer

### Description:
As a mobile user on iOS, I want to select a saved approval workflow when sending a post for approval in the Composer — just like I can on the web — so that I can use my team's pre-configured approval chains without switching to desktop.

---

### Workflow:
1. iOS user creates a post in the Composer and taps "Send for Approval"
2. The existing approval sheet now has two tabs: "Users" (existing) and "Approval Workflows" (new)
3. "Approval Workflows" tab shows the list of saved workflows; default is pre-selected
4. User can tap a workflow to select it; tap the expand chevron to see level details
5. User can add an optional note and tap "Send"
6. Post enters workflow-based approval; user sees confirmation

---

### Acceptance criteria:
- [ ] `ComposerApprovalView.swift` updated to show a two-tab interface: "Users" and "Approval Workflows"
- [ ] "Users" tab: existing single-user approval UI unchanged
- [ ] "Approval Workflows" tab:
  - Fetches workflows from `GET /workspaces/{id}/approval-workflows` (excluding drafts)
  - Shows list of workflow cells: name, level count, member avatar stacks per level with connecting visual separator
  - Default workflow is pre-selected when tab is opened
  - Tap to select/change selected workflow (selected state highlighted)
  - Tap expand to show level details: "Level N: [Title]", member avatars, Everyone/Anyone label
  - "+N more" overflow label for 10+ members in a level
- [ ] Notes text field visible on both tabs (existing behavior preserved for Users tab; new notes field on Workflows tab)
- [ ] "Send" button enabled only when a workflow is selected (or users are selected on Users tab)
- [ ] Empty state on Workflows tab (no workflows exist): "No approval workflows yet. Create one in Settings → Approval Workflows."
- [ ] API payload on send: includes `workflow_id` when a workflow is selected; follows same schema as web
- [ ] Loading state: spinner while workflows list loads
- [ ] Error state: "Failed to load workflows. Tap to retry."
- [ ] Existing push notification handling for all new approval event types (workflow_level_advanced, post_fully_approved, etc.) — verify APNs dispatch works for these new types from the backend

---

### Mock-ups:
Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034 (reference; mobile designer should produce iOS-specific screens)

---

### Impact on existing data:
- "Users" tab and existing single-user approval are unchanged; all changes are additive

---

### Impact on other products:
- None

---

### Dependencies:
- Depends on: [BE] Create approval_workflows collection and CRUD API (workflow list API)
- Depends on: [BE] Extend plans.approval schema (workflow submission payload)
- Depends on: [BE] Add new notification event types (APNs push notifications for new event types)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A (native iOS UI)
- [ ] Multilingual support — all strings in iOS Localizable.strings; English copy matches PRD Section 11
- [ ] UI theming support — use native iOS design system (UIKit/SwiftUI); match visual treatment of existing approval UI
- [ ] White-label domains impact review — N/A for native mobile UI
- [ ] Cross-product impact assessment — no web or Android impact

---

## S-19: [iOS] Add "Pending Approvals" bottom navigation item and view

### Description:
As an iOS user who is an approver, I want a dedicated "Pending Approvals" tab in the bottom navigation — filtered to posts needing my attention — so that I can review and approve content from my phone without hunting through all posts.

---

### Workflow:
1. Approver opens the ContentStudio iOS app
2. New "Pending Approvals" item appears in the bottom navigation bar (with badge count for unread pending approvals)
3. Tapping opens the Pending Approvals view with two segments: "Needs My Approval" and "My Requests"
4. "Needs My Approval": posts assigned to the current user for review, sorted by submission time
5. "My Requests": posts the current user has sent for approval (any state)
6. User can tap a post to view it, approve, reject, or add a comment
7. Bulk approve/reject: long-press or select mode to approve/reject multiple posts at once

---

### Acceptance criteria:
- [ ] New bottom navigation item: label "Pending Approvals", icon (approval/checkmark icon)
- [ ] Unread badge count on nav item showing number of posts pending the current user's approval (updates on Pusher event)
- [ ] Pending Approvals view with `UISegmentedControl` (or equivalent): "Needs My Approval" / "My Requests"
- [ ] "Needs My Approval" segment: fetches posts where current user is a pending approver at the current active level; sorted by time submitted
- [ ] "My Requests" segment: fetches posts where current user is the creator and post is in any approval state
- [ ] Post list cell: post title/preview thumbnail, social account, approval status, level progress (for workflow posts: "Level 2 of 3"), submission time, approver avatars
- [ ] Tapping a post: navigates to post detail view with approval status panel (existing `ApprovalStatusDialogVC.swift` extended to show workflow level breakdown)
- [ ] Approve and Reject buttons in post detail: existing behavior for current-level approvers (unchanged from S-03 API)
- [ ] Empty state — "Needs My Approval": "You're all caught up! No posts are waiting for your approval right now."
- [ ] Empty state — "My Requests": "No pending requests. Posts you've sent for approval will appear here."
- [ ] Loading state: spinner / skeleton cells while loading
- [ ] Error state: "Failed to load approvals. Pull to refresh."
- [ ] Pull-to-refresh on both segments
- [ ] Real-time update: Pusher/WebSocket event causes badge count and list to update without manual refresh

---

### Mock-ups:
Mobile designer to produce iOS-specific screens referencing the web Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
- Additive new tab; existing bottom navigation items are unaffected

---

### Impact on other products:
- None

---

### Dependencies:
- Depends on: [BE] Extend plans.approval schema (for workflow-aware plan queries)
- Depends on: [BE] Implement WorkflowApprovalBuilder (approve/reject endpoints)
- Depends on: [iOS] Add Approval Workflows tab to Send for Approval in Composer (shared approval infrastructure)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A (native iOS)
- [ ] Multilingual support — iOS Localizable.strings; matches PRD Section 11 copy
- [ ] UI theming support — native iOS; matches existing app design patterns
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — no web or Android impact

---

## S-20: [Android] Add Approval Workflows tab to Send for Approval in Composer

### Description:
As a mobile user on Android, I want to select a saved approval workflow when sending a post for approval in the Composer — just like on the web — so that I can use my team's pre-configured approval chains on Android without switching to desktop.

---

### Workflow:
1. Android user creates a post in the Composer and taps "Send for Approval"
2. The existing approval sheet (`ComposerApprovalFragment.java`) now shows two tabs: "Users" (existing) and "Approval Workflows" (new)
3. "Approval Workflows" tab loads saved workflows; default is pre-selected
4. User selects a workflow, optionally adds a note, and taps "Send"
5. Post enters workflow-based approval

---

### Acceptance criteria:
- [ ] `ComposerApprovalFragment.java` updated to show two tabs: "Users" and "Approval Workflows"
- [ ] "Users" tab: existing single-user approval UI unchanged
- [ ] "Approval Workflows" tab:
  - Fetches from `GET /workspaces/{id}/approval-workflows`
  - Shows workflow list: name, level count, member avatar stacks with level separator
  - Default workflow pre-selected
  - Tap to select; tap expand to see level details
  - "+N more" overflow for 10+ members in a level
- [ ] Notes input field on both tabs
- [ ] "Send" button enabled only when workflow is selected (or users selected)
- [ ] Empty state: "No approval workflows yet. Create one in Settings → Approval Workflows."
- [ ] API payload: includes `workflow_id` when workflow selected
- [ ] Loading/error states handled (spinner while loading; error toast with retry)
- [ ] Existing FCM push notification handling confirmed for all new approval event types (workflow_level_advanced, post_fully_approved, etc.)

---

### Mock-ups:
Figma reference: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034 (mobile designer to produce Android-specific screens)

---

### Impact on existing data:
- "Users" tab and existing single-user flow unchanged

---

### Impact on other products:
- None

---

### Dependencies:
- Depends on: [BE] Create approval_workflows collection and CRUD API
- Depends on: [BE] Extend plans.approval schema
- Depends on: [BE] Add new notification event types (FCM push notifications)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A (native Android)
- [ ] Multilingual support — Android strings.xml; matches PRD Section 11 copy
- [ ] UI theming support — native Android Material Design; matches existing app patterns
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — no web or iOS impact

---

## S-21: [Android] Add "Pending Approvals" bottom navigation item and view

### Description:
As an Android user who is an approver, I want a dedicated "Pending Approvals" section in the app — with a badge count showing how many posts need my attention — so that I can quickly review and approve content from my Android device.

---

### Workflow:
1. Approver opens the ContentStudio Android app
2. "Pending Approvals" item appears in the bottom navigation (with unread badge count)
3. Opens Pending Approvals view with two tabs: "Needs My Approval" and "My Requests"
4. User can tap a post to view details, approve/reject, or add a comment
5. Badge count updates in real-time

---

### Acceptance criteria:
- [ ] New bottom navigation item: label "Pending Approvals" with appropriate icon
- [ ] Unread badge count on nav item (posts pending current user's approval)
- [ ] Pending Approvals Fragment/Activity with `TabLayout`: "Needs My Approval" / "My Requests"
- [ ] "Needs My Approval": posts where current user is a pending approver at the current active level
- [ ] "My Requests": posts where current user is the creator and post is in any approval state
- [ ] Post list item: title/thumbnail, account, approval status, level progress (workflow posts), submission time
- [ ] Post detail: approval status panel with level breakdown; Approve / Reject buttons for current-level approvers
- [ ] Empty state — "Needs My Approval": "You're all caught up! No posts are waiting for your approval right now."
- [ ] Empty state — "My Requests": "No pending requests. Posts you've sent for approval will appear here."
- [ ] Loading state: shimmer/skeleton; error state with retry; pull-to-refresh
- [ ] Badge count updates on FCM push events

---

### Mock-ups:
Mobile designer to produce Android-specific screens; web Figma reference: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
- Additive new navigation item; existing items unaffected

---

### Impact on other products:
- None

---

### Dependencies:
- Depends on: [BE] Extend plans.approval schema
- Depends on: [BE] Implement WorkflowApprovalBuilder (approve/reject endpoints)
- Depends on: [Android] Add Approval Workflows tab to Send for Approval in Composer

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A (native Android)
- [ ] Multilingual support — Android strings.xml; PRD Section 11 copy
- [ ] UI theming support — native Android; matches existing app design
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessment — no web or iOS impact

---

## S-22: [BE] Implement approval override cleanup and cancellation notification dispatch

### Description:
As a content creator, I want starting a new approval on a post that already has one to properly cancel the old approval — including cleaning up all previous state and notifying the previously assigned approvers — so that no orphaned approval records exist and no approver is left waiting on a review that was silently cancelled.

---

### Workflow:
1. User initiates a new approval (any type: internal ad-hoc, internal workflow, or external share link) on a post that already has an active approval
2. Frontend has shown the appropriate warning (handled in S-24); user proceeds
3. API call to send for approval fires
4. Backend detects an existing active approval on the plan
5. Previous approvers at the current active level are notified of cancellation
6. All previous approval state is fully cleared from `plans.approval`
7. New approval is written and begins

---

### Acceptance criteria:
- [ ] Before writing a new approval to `plans.approval`, the system checks if the plan already has an active approval (`plans.approval.status` in: `pending`, `partially_approved`, `review`, `missed_review`, or any equivalent active state — including external approval state)
- [ ] If an active approval exists: dispatch `approval_cancelled_override` notification to current active level approvers only (same scope as `post_deleted_approval` — not past levels, not future levels)
- [ ] `approval_cancelled_override` notification spec:
  - In-app title: `"[Post title] has been sent for a new approval"`
  - In-app description: `"The approval process you were part of for [post title] in [workspace name] has been replaced. No further action is required from you."`
  - Email subject: `"No action needed — approval replaced for [post title]"`
  - Email body: `"[Creator name] has started a new approval process for [post title] in [workspace name]. The previous approval process has been cancelled. No further action is required from you."`
  - No CTA button
  - Channels: in-app, email
- [ ] After notifications dispatched: fully clear `plans.approval` (reset all internal approval fields: `workflow_id`, `current_level`, `workflow_levels`, `level_status`, `approvers`, `status`, etc.) before writing the new approval
- [ ] **Bug fix — external approval override**: when initiating an external share link approval, the API must fully clear any existing internal approval state before writing the new external approval. Currently this cleanup is missing. After this fix, internal→external override and external→internal override both clean up all previous state symmetrically.
- [ ] When a bulk internal approval overrides an existing active internal approval, `approval_cancelled_override` notifications must be dispatched to the previously active level approvers before the new approval is written
- [ ] Previous approval history must be preserved, not overwritten — prior approval cycle data is archived for audit trail
- [ ] All existing approval API endpoints continue to work for posts with no prior approval (no regression)

---

### Mock-ups:
N/A (backend)

---

### Impact on existing data:
- Fixes data inconsistency: existing posts that had external approval overlaid on internal without cleanup will not be retroactively fixed, but going forward all overrides are clean
- Previous approval history archiving is additive; no migration required

---

### Impact on other products:
- iOS and Android send-for-approval actions call the same endpoints; they will now receive cancellation notifications if they override existing approvals — no client-side changes needed

---

### Dependencies:
- Depends on: [BE] Add new notification event types (for notification dispatch infrastructure)
- Depends on: [BE] Extend plans.approval schema (for previous approval history archiving)

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A
- [ ] Multilingual support — `approval_cancelled_override` notification copy added for all supported languages, with English fallback
- [ ] UI theming support — N/A
- [ ] White-label domains impact review — email uses existing white-label template; no changes needed
- [ ] Cross-product impact assessment — all clients (web, iOS, Android) that trigger send-for-approval are affected; notification dispatch is server-side

---

## S-23: [FE] Add "Send for Approval" to single Planner post card and post detail view

### Description:
As a content creator, I want to send a single post for approval directly from the Planner — via the post card menu or the post detail view — so that I don't have to open the Composer just to route a post for review.

---

### Workflow:
1. User is in the Planner (any view: calendar, feed, list)
2. User sees a post in Draft or Scheduled state that needs approval
3. User clicks the 3-dot menu on the post card → "Send for Approval"
   — OR — opens the post detail/preview modal → clicks "Send for Approval" button
4. The `SendForApprovalSidebar` slides open (same component used in Composer — S-11)
5. If the post already has an active approval, the override warning is shown (S-24)
6. User selects approvers or a workflow, adds optional note, clicks Send

---

### Acceptance criteria:
- [ ] "Send for Approval" option added to the post card 3-dot (···) menu in all Planner views (calendar, feed, list) — visible for posts in `draft` or `scheduled` state; hidden for posts in `published`, `failed`, or `queued` states
- [ ] "Send for Approval" `Button` (secondary) added to the post detail / preview modal action area — visible when the post is NOT currently in any active approval state; replaced by the approval status panel (S-12) when approval is active
- [ ] Clicking either entry point opens the `SendForApprovalSidebar.vue` component (same as Composer — reused, not rebuilt)
- [ ] Sidebar context: the sidebar receives the `plan_id` of the selected post and operates identically to the Composer flow
- [ ] Override check: before opening the sidebar, check if the post has an active approval; if yes, the override warning `Alert` is shown at the top of the sidebar (handled in S-24; the sidebar accepts an `existingApproval` prop for this)
- [ ] On Send: post enters approval; sidebar closes; planner status badge updates in real-time via Pusher
- [ ] `CstToast` success notification: "Post sent for approval"
- [ ] Error state: `Alert` (error variant) inside sidebar: "Failed to send for approval. Please try again."
- [ ] The "Send for Approval" menu item is hidden for users who don't have permission to send for approval (collaborators without the relevant permission — follow existing permission checks)

---

### Mock-ups:
Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
- No data model changes; purely UI additions to existing Planner components

---

### Impact on other products:
- None

---

### Dependencies:
- Depends on: [FE] Migrate Send for Approval from center modal to right sidebar (S-11) — the sidebar component must exist before it can be embedded here
- Depends on: [FE] Add approval override warning (S-24) — the override warning shown inside the sidebar

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — Planner is web-only; "Send for Approval" on post card must be accessible at 1024px+
- [ ] Multilingual support — "Send for Approval" label via existing i18n key; success/error toasts via i18n
- [ ] UI theming support — `Button`, `ActionIcon`, `CstToast` from `@contentstudio/ui`; no hardcoded colors
- [ ] White-label domains impact review — no brand color usage in menu items
- [ ] Cross-product impact assessment — no Chrome extension or mobile impact

---

## S-24: [FE] Add approval override warning to Send for Approval sidebar, bulk Planner send, and external share link

### Description:
As a content creator, I want to see a clear warning whenever I'm about to start a new approval that will cancel an existing one — whether for a single post, multiple posts at once, or via the external share link flow — so that I'm never surprised that a previous approval process was cancelled without my knowledge.

---

### Workflow:

**Single post (Composer or Planner):**
1. User opens the Send for Approval sidebar for a post that already has an active approval
2. An Alert warning appears at the top of the sidebar explaining what will be cancelled
3. User can proceed (fill sidebar, click Send) or close the sidebar — no extra step required

**Bulk Planner send:**
1. User selects multiple posts and clicks "Send for Approval" in the Planner bulk action bar
2. System checks which selected posts have active approvals
3. If any do: a confirmation Dialog appears before the sidebar opens
4. User clicks "Continue" → sidebar opens; or "Cancel" → no action

**External share link:**
1. User opens the Share via Link modal for a post that has an active internal approval
2. In Step 2, user toggles on "Send for Approval" (adding external approver emails)
3. An Alert warning appears inside the share modal

---

### Acceptance criteria:

**Single post override warning (inside `SendForApprovalSidebar.vue`):**
- [ ] Sidebar accepts an `existingApproval` prop (object or null); when non-null, an `Alert` (warning variant) is rendered at the top of the sidebar above the tabs
- [ ] Copy varies by existing approval type:
  - Ad-hoc internal: `"This post is already in an approval process. Starting a new one will cancel it — currently assigned approvers will be notified."`
  - Workflow internal: `"This post is in [Workflow Name] (Level [N]). Starting a new approval will cancel it — Level [N] approvers will be notified."`
  - External share-link: `"This post has an active external approval via share link. Starting an internal approval will cancel it."`
- [ ] Alert is informational — Send button remains enabled; no additional confirmation step for single post
- [ ] `existingApproval` data is fetched as part of the plan detail API call that already runs when the sidebar opens (no extra round-trip needed if `plans.approval` is already in the response)

**Bulk send override confirmation Dialog:**
- [ ] Before `SendForApprovalSidebar` opens for a bulk action, check the approval state of each selected plan (use data already loaded in the Planner store; no extra API call if data is fresh)
- [ ] If any selected plans have `approval.status` in an active state:
  - `Dialog` (use `Dialog` component from `@contentstudio/ui`):
    - Title: `"Some posts are already in approval"`
    - Body: `"[N] of [M] selected posts already have an active approval process. Starting a new one will cancel their current approvals — assigned approvers will be notified."`
    - Primary `Button`: `"Continue"`
    - Secondary `Button`: `"Cancel"`
  - "Continue" → dismiss dialog, open sidebar
  - "Cancel" → dismiss dialog, keep selection, no sidebar
- [ ] If no selected plans have active approvals: skip the dialog entirely, open sidebar directly (no unnecessary friction)

**External share link warning (inside `SharePlanModal.vue`):**
- [ ] In `SharePlanModal.vue`, when the "Send for Approval" toggle is turned ON (user adds external approver emails) and the current plan has an active internal approval:
  - `Alert` (warning variant) rendered below the toggle: `"This post is currently in an internal approval process. Sending it for external approval will cancel it — currently assigned approvers will be notified."`
- [ ] Alert is only shown when both conditions are true: (1) approval emails are being added, AND (2) the plan has an active internal approval
- [ ] Alert is informational — the user can proceed; no blocking confirmation

**Post-send behaviour (all scenarios):**
- [ ] When Send is clicked and an override occurs: the backend (S-22) handles the cancellation and notifications; the frontend does not need to fire separate cancel API calls
- [ ] On success: `CstToast`: "Post sent for approval" (single) or "Posts sent for approval" (bulk)
- [ ] On error: `Alert` (error variant) in sidebar/dialog: "Something went wrong. Please try again."

---

### Mock-ups:
Figma: https://www.figma.com/file/zJEk0csU8yeDNKhUR7DIEo/ContentStudio-WebApp?node-id=9006%3A21034

---

### Impact on existing data:
- No data model changes; purely UI additions and conditional rendering

---

### Impact on other products:
- `SharePlanModal.vue` is used in Planner — confirm no other contexts use this component that would be inadvertently affected

---

### Dependencies:
- Depends on: [BE] Implement approval override cleanup and cancellation notification dispatch (S-22) — backend must handle the actual cancel + notify on Send
- Depends on: [FE] Migrate Send for Approval from center modal to right sidebar (S-11) — the sidebar component this story modifies
- For external share link: `SharePlanModal.vue` must have access to the plan's current `approval` state

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — sidebar and dialogs are web-only; `Dialog` and `Alert` components are responsive by default
- [ ] Multilingual support — all copy via i18n keys; dynamic values (workflow name, level number, post counts) injected at runtime
- [ ] UI theming support — `Alert`, `Dialog`, `Button` all from `@contentstudio/ui`; no hardcoded colors
- [ ] White-label domains impact review — no brand colors in warning alerts or dialogs
- [ ] Cross-product impact assessment — `SharePlanModal.vue` changes affect all Planner views; verify no regression in non-approval share link flows
