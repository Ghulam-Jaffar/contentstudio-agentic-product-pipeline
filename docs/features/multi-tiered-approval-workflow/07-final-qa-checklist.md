# Final QA Checklist: Multi-Tiered Approval Workflow

## Purpose

This checklist is for the final live verification pass of the frontend-first, mock-data implementation.

Use this in `yarn run dev` before backend integration starts.

---

## 1. Settings: Approval Workflows

- Open Workspace Settings -> Approval Workflows and confirm the menu item appears under Social Accounts.
- Confirm the list page loads without layout overflow on desktop and mobile widths.
- Confirm workflow tiles open the builder on direct click, not only from the 3-dots menu.
- Hover a workflow tile and confirm the localized edit tooltip appears.
- Confirm default and draft badges display correctly.
- Confirm the empty state explains the feature and has a clear CTA.
- Create a workflow with multiple levels and confirm:
  - members can be added from the drawer
  - level ordering can be dragged and re-sorted
  - level member chips/rows render correctly
  - Anyone / Everyone controls use the same UI and tooltip language as Send for Approval
- Delete a level and confirm the standard `msgBoxConfirm` pattern is used.
- Delete a workflow and confirm the standard `msgBoxConfirm` pattern is used, including the highlighted warning block for in-review posts.
- From builder `Cancel`, confirm navigation returns to the correct workspace settings route.

## 2. Composer: Send / Change / Remove Approval

- Open composer on a post without approval and confirm Send for Approval opens the right sidebar.
- In Users tab:
  - select users
  - switch Anyone / Everyone
  - add note
  - send successfully
- In Approval Workflows tab:
  - confirm default workflow preselection for new submissions
  - expand/collapse workflow cards
  - select another workflow and send
- On a post already using workflow approval:
  - click `Change approval`
  - confirm sidebar opens on `Approval Workflows`
  - confirm the current workflow is preselected
  - confirm `Send for Approval` stays disabled until an actual change is made
  - change workflow and confirm override confirmation appears before applying
- On a post already using custom approval:
  - click `Change approval`
  - confirm sidebar opens on `Users`
  - confirm existing users/rule/note are prefilled
- Click `Remove approval` and confirm:
  - a standard confirm box appears
  - the warning copy matches workflow vs custom approval context
  - approval is not silently removed without confirmation
- Edit a post in approval and save. Confirm the edit-confirmation flow works for:
  - active approval
  - workflow approval
  - rejected approval
  - missed review
- Confirm canceling the edit-confirmation dialog preserves unsaved composer state.
- Confirm first-time `Send for Approval` closes composer normally without a blocking success modal.
- Confirm blocking success confirmation remains limited to the explicit edit-approval action flow.

## 3. Planner: List / Calendar / Feed / Preview

- In list view, confirm approval posts show:
  - one approval status control, not duplicated plain text + pill
  - compact approval badge under status
- In list view, confirm both workflow and custom approval posts use the compact approval badge.
- In calendar view, confirm approval posts show the compact approval badge.
- In feed view, confirm approval posts show the compact approval badge.
- Decide whether the current grid treatment is acceptable or should stay lighter than feed/list/calendar.
- Click the compact badge and confirm the popup shows:
  - workflow name / current level / per-level summary for workflow approvals
  - flat approver list for custom approvals
  - approver states
- Confirm the compact badge is shown only on approval-related states:
  - `In Review`
  - `Missed Review`
  - `Rejected`
- Confirm the compact badge pill uses the planner-facing labels/colors:
  - `In Review`
  - `Missed Review`
  - `Rejected`
- Confirm the compact badge tooltip shows progress detail, not only the approval rule.
- Click the post and confirm full preview shows the detailed approval status panel.
- In preview/aside, confirm workflow mode shows workflow levels as the primary UI, without the redundant activity block above them.
- In preview/aside, confirm custom approvals still use the legacy user-list renderer.
- Confirm planner actions work visually and semantically for:
  - Change Approval / Send for Approval
  - Re-notify
  - Restart approval
  - Remove approval

## 4. Planner: Override Flows

- Bulk-select planner posts and use Send for Approval.
- Confirm override warning appears when some selected posts are already in review.
- Use single-post planner Send for Approval on a post already in approval and confirm warning behavior.
- Open Share via Link on a post already under internal approval, enable Send for Approval, and confirm the internal-override warning appears.

## 5. Notifications

- Confirm the top-bar approval bell appears and unread count updates.
- Open the panel and verify `Unread` and `All` tabs.
- Mark a single notification as read.
- Use `Mark all read`.
- Click a notification and confirm it routes into planner and opens the relevant post context.
- Confirm workflow and custom approval notifications are visually distinguishable.

## 6. Visual / Interaction Consistency

- Confirm all delete/override/removal flows use the shared confirm-box pattern.
- Confirm approval workflow cards keep the same design language across:
  - settings list
  - send-for-approval sidebar
- Confirm primary token classes are used instead of raw `rgb(var(--cstu-primary-...))` utilities in the touched approval-workflow surfaces.
- Confirm hover states work for:
  - workflow tiles
  - workflow builder member rows
  - builder drag handles
  - row remove actions

## 7. Mock-Mode Expectations

- Confirm mock approval data only appears where intended for demo visibility.
- Confirm planner demo states do not create obviously contradictory UI between compact badge, preview panel, and composer summary.
- Identify any mock-only behavior that must be removed or feature-flagged before backend rollout.

## 8. Backend Handoff Readiness

- Review [06-frontend-data-contract.md](./06-frontend-data-contract.md) with backend.
- Confirm exact field names for:
  - workflow CRUD responses
  - `plans.approval`
  - approval action flags
  - notifications payloads
- Confirm nullable/fallback rules before backend starts implementation.
