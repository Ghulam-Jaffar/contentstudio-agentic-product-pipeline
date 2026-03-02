# Research: Shared Link Feedback Controls (Comments + Approve/Reject Toggles)

## Current State

The shared planner link flow is implemented in `planner_v2` and already has one toggle for sending approval requests, but no persistent permission flags for what external users can do after opening the link.

Key frontend behavior today:

- **Create/edit share link modal** (`src/modules/planner_v2/components/SharePlanModal.vue`)
  - Supports `approvalFlow` toggle (`send_for_approval`) and `approvalRule`.
  - Sends these in update payload as `approval_flow`, `approval_emails`, `approval_option`.
  - Does **not** expose/comment on any dedicated `allow_comments` or `allow_approval_actions` style permissions.

- **Public share link page** (`src/modules/planner_v2/views/SharePlans.vue`)
  - Bulk actions always render in list view: Approve / Reject / Comment.
  - Reads link metadata from `shareLink/get` and stores `approverEmail`, `isApprovalFlow`, but current UI does not use `isApprovalFlow` to hide/disable action controls.

- **Per-row action buttons**
  - Desktop: `src/modules/planner_v2/components/DataRow.vue`
  - Mobile: `src/modules/planner_v2/components/DataRowCardMobile.vue`
  - Approve/Reject visibility depends on post status and completion state, not on link-level permission flags.
  - Comment action is always available.

- **Preview modal external actions** (`src/modules/planner_v2/components/PlannerPostPreview_v2.vue`)
  - External mode uses `CommentsAndNotes` with `comment-type="external"`.
  - Footer approve/reject buttons are shown for eligible statuses unless approval already completed.
  - No link-level permission switch check for comments/approve/reject.

- **Action submission path**
  - Approve/reject through `ExternalActionsModal.vue` -> `shareLink/action`.
  - Comments through `CommentsAndNotes.vue` (external) or modal -> `shareLink/comment`.
  - Frontend currently assumes both endpoints are available for the link.

## What Needs to Change

### 1) Add explicit link permissions in share-link data model (frontend contract)

Introduce two booleans in link payload/response contract:

- `allow_external_comments` (default `true`)
- `allow_external_approval_actions` (default `true` or derived from approval flow)

These must be:

- Sent from create/edit modal when saving a link.
- Returned by `shareLink/get` and `shareLink/fetch`.
- Preserved in Manage Links edit payload so toggles do not get reset accidentally.

### 2) Add toggles in SharePlanModal

In `SharePlanModal.vue`, add two independent toggles in the form state:

- `Allow comments`
- `Allow approve/reject`

Suggested UX behavior:

- If `allow_external_approval_actions = false`, hide/disable approval email requirement and approval rule controls.
- If `allow_external_comments = false` and approval actions are also false, show a validation error (a fully view-only link is acceptable only if intended; decide product behavior).
- Prefill these values in edit mode from `props.editLink`.

### 3) Enforce permissions across all external action entry points

Use link-level flags from `SharePlans.vue` state (fetched from `shareLink/get`) and pass to child components where needed.

Action entry points to gate:

- `SharePlans.vue`: bulk dropdown items (approve/reject/comment)
- `DataRow.vue`: hover action buttons
- `DataRowCardMobile.vue`: action buttons
- `PlannerPostPreview_v2.vue`: external comments panel + approve/reject footer
- `ExternalActionsModal.vue`: reject submit if link disallows action (defensive UX)
- `CommentsAndNotes.vue` (external mode): disable/hide submit form when comments disallowed

### 4) Backend-backed enforcement is required (not only UI hiding)

Frontend-only checks are bypassable by direct API requests. For complete behavior:

- `shareLink/comment` should reject when `allow_external_comments = false`.
- `shareLink/action` should reject when `allow_external_approval_actions = false`.

Frontend should surface returned errors with existing toast paths.

## Suggested Implementation Approaches

### Option A (recommended): Two independent toggles

- Supports exact user ask: turn comments off, approvals off, or either one.
- Clear and future-proof for client-specific review workflows.

### Option B (minimal): Single "Feedback enabled" master toggle

- Simpler UI but less flexible.
- Does not satisfy use case where approvals are off but comments on (or vice versa).

## Files Involved (frontend)

- `src/modules/planner_v2/components/SharePlanModal.vue`
- `src/modules/planner_v2/views/SharePlans.vue`
- `src/modules/planner_v2/components/DataRow.vue`
- `src/modules/planner_v2/components/DataRowCardMobile.vue`
- `src/modules/planner_v2/components/PlannerPostPreview_v2.vue`
- `src/modules/planner_v2/components/ExternalActionsModal.vue`
- `src/modules/planner_v2/components/CommentsAndNotes.vue`
- `src/modules/planner_v2/components/ManageLinksModal.vue` (include new fields in update payload helper)
- `src/locales/en/planner.json` (+ other locales for parity)

## Risk Notes

- Existing links without new fields should default to current behavior (`true/true`) to avoid regressions.
- Edit flow currently updates only a subset of link fields; adding new permission fields without carefully merging payload could unintentionally reset settings.
- If only frontend gating is done, users can still perform restricted actions via direct API calls.

