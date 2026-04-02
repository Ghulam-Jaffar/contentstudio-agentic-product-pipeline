# Frontend Data Contract: Multi-Tiered Approval Workflow

## Purpose

This document captures the **actual frontend data contract** currently required by the mock-data implementation in `contentstudio-frontend`.

Use this as the backend handoff reference for:

- approval workflow settings APIs
- plan/post approval payloads
- planner approval rendering
- composer approval editing
- approval notifications

This is intentionally FE-first. It reflects the fields the current UI already reads, writes, or depends on.

## Scope

This contract currently covers:

- workspace approval workflow CRUD
- assigning approval to a post
- workflow-based approval status
- custom approval status
- planner compact badge + full approval panel
- composer change/remove approval flows
- approval notifications bell/panel

This does **not** define:

- mobile native payload specifics
- final backend persistence model
- permissions/auth implementation details

Important scope note:

- the current mobile story set covers workflow selection and pending-approvals views
- it does **not** define mobile parity for the web-only composer edit-confirmation flow when editing a post already in approval

## 1. Workflow Settings Contract

### Workflow object

The settings list, builder, and send-for-approval workflow picker expect this shape:

```ts
interface WorkflowMember {
  user_id: string
  user?: {
    firstname?: string
    lastname?: string
    email?: string
    image?: string
  }
  role?: string
  membership?: string
  permissions?: Record<string, string[]>
}

interface WorkflowLevel {
  level_number: number
  title: string
  rule: 'everyone' | 'anyone'
  members: WorkflowMember[]
}

interface ApprovalWorkflow {
  _id: string
  name: string
  is_draft: boolean
  is_default: boolean
  levels_count: number
  members_count: number
  levels: WorkflowLevel[]
  created_at?: string
  updated_at?: string
}
```

### Required behavior

- `levels_count` and `members_count` should be sent by backend, not recomputed in FE.
- `is_default` is used in settings list and workflow picker.
- `is_draft` is used in the settings list/builder and should be reliable.
- `levels` must be fully populated in the list response because the current cards/sidebar use them directly.

### CRUD payload used by FE

Create/update payload currently aligns to:

```ts
interface WorkflowPayload {
  name: string
  levels: Array<{
    level_number: number
    title: string
    rule: 'everyone' | 'anyone'
    members: WorkflowMember[]
  }>
}
```

### Workflow endpoints FE currently expects conceptually

- `GET /workspaces/:workspaceId/approval-workflows`
- `GET /approval-workflows/:id`
- `POST /workspaces/:workspaceId/approval-workflows`
- `PUT /approval-workflows/:id`
- `DELETE /approval-workflows/:id`
- `POST /approval-workflows/:id/duplicate`
- `PUT /approval-workflows/:id/set-default`
- `PUT /approval-workflows/:id/remove-default`

### Delete workflow response

The settings list/builder confirmation UX needs the delete response to support active-review impact messaging.

Preferred response:

```ts
{
  status: boolean
  message?: string
  in_review_posts_count?: number
  auto_advance_summary?: string
}
```

If deletion requires explicit override/force:

```ts
DELETE /approval-workflows/:id?force=true
```

or equivalent body/query pattern.

## 2. Plan/Post Approval Contract

The post-level approval object is the most important backend handoff item.

The FE currently has two modes:

- workflow approval
- custom approval

### Canonical approval object

```ts
interface ApproverStatus {
  user_id: string
  status: 'approve' | 'reject' | 'pending'
  last_action_time?: string | { date?: string }
  last_action_note?: string
}

interface ApprovalLevel {
  level: number
  name: string
  rule: 'everyone' | 'anyone'
  members: string[]
}

interface LevelStatus {
  level: number
  approvers: ApproverStatus[]
}

interface PlanApproval {
  status?: string

  // shared
  notes?: string
  members?: string[]
  approvers?: ApproverStatus[]
  approve_option?: 'anyone' | 'everyone'
  is_external?: boolean

  // workflow mode
  workflow_id?: string
  workflow_name?: string
  current_level?: number
  total_levels?: number
  levels_count?: number
  workflow_levels?: ApprovalLevel[]
  level_status?: LevelStatus[]

  // action flags sent back by FE on edit/save flows
  remove_approval?: boolean
  renotify_approvers?: boolean
  renotify_current_level?: boolean
  reset_workflow?: boolean
  restart_from_level?: number
  resend_approval?: boolean
  resume_workflow?: boolean
  resume_from_level?: number
}
```

### Status values currently handled in FE

The FE currently checks for these values:

- `pending`
- `pending_approval`
- `under_review`
- `review`
- `partially_approved`
- `missed_review`
- `rejected`
- `rejected_approval`

Preferred backend direction:

- keep one canonical list
- do not send near-duplicate status values unless intentionally required

But until FE cleanup is complete, backend should assume these may still appear.

### Critical workflow-specific requirements

For workflow-backed posts, backend should always send:

- `workflow_id`
- `workflow_name`
- `current_level`
- `total_levels`
- `workflow_levels`
- `level_status`

These are required for:

- planner compact badge
- planner full approval panel
- composer approval summary
- composer change-approval preselection

### Strong recommendation

For workflow-backed posts, backend should also continue sending:

- `approvers`
- `members`
- `approve_option`

Reason:

- some older/shared FE paths still fall back to flat approver data
- this avoids partial regressions while workflow UI and legacy UI coexist

### Workflow level structure

For planner/composer workflow rendering, FE expects:

```ts
workflow_levels: [
  {
    level: 1,
    name: 'Content Review',
    rule: 'everyone',
    members: ['user-1', 'user-2']
  }
]
```

and:

```ts
level_status: [
  {
    level: 1,
    approvers: [
      {
        user_id: 'user-1',
        status: 'approve',
        last_action_time: '2026-03-28T10:00:00Z',
        last_action_note: 'Looks good'
      },
      {
        user_id: 'user-2',
        status: 'pending'
      }
    ]
  }
]
```

### Important fallback note

The FE still has some fallback support for `levels_count`, but backend should prefer `total_levels` for post approval objects.

### Planner compact-status note

For compact planner rendering, FE now supports both:

- workflow approvals via `workflow_levels` + `level_status`
- custom approvals via `approvers`

The compact approval popup is intentionally shown only for approval-related planner states:

- `review` / `reviewed`
- `missedReview` / `missed_review`
- `rejected`

FE normalizes underscores/spaces/casing in those state checks, but backend should still prefer one consistent canonical shape.

## 3. Approval Assignment Payloads

### Custom approval assignment

When FE sends selected users:

```ts
{
  type: 'custom',
  members: string[],
  approvers: Array<{
    user_id: string
    status: 'pending'
    last_action_time: ''
    last_action_note: ''
  }>,
  approve_option: 'anyone' | 'everyone',
  notes?: string,
  status: 'pending_approval'
}
```

### Workflow approval assignment

When FE sends a reusable workflow:

```ts
{
  type: 'workflow',
  workflow_id: string,
  workflow_name: string,
  current_level: 1,
  total_levels: number,
  workflow_levels: ApprovalLevel[],
  level_status: LevelStatus[],
  approvers: ApproverStatus[],
  members: string[],
  approve_option: 'anyone' | 'everyone',
  notes?: string,
  status: 'pending_approval'
}
```

Backend does not have to persist the exact FE-generated placeholder statuses as-is, but it should return the normalized approval object immediately after save.

## 4. Composer Edit/Save Action Flags

When editing a post already in approval, FE may send the existing approval object back with one or more action flags set.

### Flags currently emitted by FE

- `remove_approval: true`
- `renotify_approvers: true`
- `renotify_current_level: true`
- `reset_workflow: true`
- `restart_from_level: 1`
- `resend_approval: true`
- `resume_workflow: true`
- `resume_from_level: current_level`

### Backend expectation

The save/update response should return the **full updated `approval` object**, not only a success boolean.

This matters because the composer, planner preview, and approval panels all rely on the post-save approval state being immediately renderable.

### Success-state note

The first-time "Send for Approval" flow is now expected to close composer normally without a blocking success modal.

The blocking success confirmation is reserved for edit-save flows where the user explicitly chose an approval-handling action from the approval edit dialog.

## 5. Bulk Send-for-Approval Contract

The bulk planner flow currently sends:

```ts
{
  workspace_id: string,
  plans: string[],
  approval: PlanApproval
}
```

Preferred response:

```ts
{
  status: boolean
  message?: string
  updated_count?: number
  skipped_count?: number
  overridden_count?: number
}
```

## 6. Planner Rendering Requirements

### Compact badge

List/calendar/feed planner views need enough approval data to show:

- workflow name
- current level
- total levels
- per-level status
- current-level pending/approved users
- custom approval approver list and statuses when no workflow levels exist

Required minimum fields:

- `status`
- `post_state`
- `workflow_name`
- `current_level`
- `total_levels`
- `workflow_levels`
- `level_status`
- `approvers` for custom approval rows/cards

### Full preview panel

Planner preview/aside uses the same workflow object plus:

- `approve_option`
- `approvers`
- `notes` if present
- action timestamps and notes when available

## 7. Approval Notifications Contract

### Notification shape used by FE

```ts
type NotificationType =
  | 'approval_request'
  | 'approved'
  | 'rejected'
  | 'level_passed'
  | 'reminder'

interface ApprovalNotification {
  id: string
  type: NotificationType
  read: boolean
  created_at: string

  plan_id: string
  plan_title: string

  actor?: string
  actor_role?: string

  target: 'workflow' | 'users'
  workflow_name?: string
  level?: number
  level_name?: string
  approvers_count?: number

  summary?: string
  note?: string
}
```

### Required behavior

- bell count uses unread items only
- click-through needs `plan_id`
- workflow notifications should include `workflow_name`
- workflow level events should include `level`
- `level_name` is optional but strongly preferred

### Notification endpoints/events

Backend can implement this through polling, embedded header payload, or sockets, but FE eventually needs:

- fetch notifications
- mark one as read
- mark all as read
- real-time insert/update event support

## 8. Nullability and Fallback Rules

### Workflow-backed approval

Backend should avoid returning workflow approvals with only `workflow_name` and no `workflow_id`.

The FE has a name fallback for now, but:

- `workflow_id` is preferred
- `workflow_name` alone is not ideal

### Timestamps

`last_action_time` currently tolerates both:

- ISO string
- `{ date: ISOString }`

Backend should standardize to a plain ISO string if possible.

### Members and user resolution

All approval user references should use workspace member `user_id` values that match the active workspace member list.

That is how FE resolves:

- avatar
- full name
- initials
- role badges

## 9. Minimum Backend Contract to Unblock FE

If backend wants the smallest possible first pass, these are the highest-priority fields to get right:

### Workflow settings

- `_id`
- `name`
- `is_draft`
- `is_default`
- `levels_count`
- `members_count`
- `levels[]`

### Post approval

- `status`
- `workflow_id`
- `workflow_name`
- `current_level`
- `total_levels`
- `workflow_levels[]`
- `level_status[]`
- `approvers[]`
- `approve_option`
- `notes`

### Notifications

- `id`
- `type`
- `read`
- `created_at`
- `plan_id`
- `plan_title`
- `target`
- `workflow_name` for workflow notifications

## 10. Recommended Backend Response Principle

For every mutation, return the **fully normalized object the FE should render next**.

That applies to:

- workflow create/update/delete/default actions
- post save with approval changes
- approval actions like re-notify/restart/remove
- bulk send-for-approval

Avoid returning only:

- `status: true`
- `message: "done"`

That forces FE to guess or fake intermediate state.

## 11. Current FE Source References

These files are the main source of truth for the current FE contract:

- `contentstudio-frontend/src/modules/approval-workflows/composables/useApprovalWorkflows.ts`
- `contentstudio-frontend/src/modules/approval-workflows/components/SendForApprovalSidebar.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/PlanApprovalStatus.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/WorkflowApprovalLevels.vue`
- `contentstudio-frontend/src/modules/approval-workflows/components/PlannerApprovalCompactBadge.vue`
- `contentstudio-frontend/src/modules/approval-workflows/composables/useApprovalNotifications.ts`
- `contentstudio-frontend/src/modules/approval-workflows/components/ApprovalNotificationsPanel.vue`
- `contentstudio-frontend/src/modules/composer_v2/views/SocialModal.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/MainComposer.vue`
