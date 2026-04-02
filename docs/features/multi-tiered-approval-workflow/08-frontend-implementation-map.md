# Frontend Implementation Map: Multi-Tiered Approval Workflow

## Purpose

This document explains **what was built in frontend**, **where it lives**, **how the parts talk to each other**, and **what must change during backend integration**.

Use this together with:

- `06-frontend-data-contract.md` for backend payload shape
- `07-final-qa-checklist.md` for live verification

This is the doc a backend dev or a follow-up frontend dev should read when asking:

- what files are involved?
- what components were added or changed?
- where is mock data injected?
- what event names and locale keys are in use?
- what needs to be replaced when real APIs arrive?

---

## 1. Feature Areas

The frontend implementation is split across five areas:

1. Workspace Settings
2. Composer
3. Planner
4. Notifications
5. Shared approval-workflow module

---

## 2. Shared Core Module

These files are the core of the feature:

### Composables

- `contentstudio-frontend/src/modules/approval-workflows/composables/useApprovalWorkflows.ts`
  - central workflow CRUD composable
  - owns `MOCK_ENABLED`
  - owns `APPROVAL_MOCK_DEMO`
  - owns mock workflow store
  - exports `getMockApprovalPlan()`
- `contentstudio-frontend/src/modules/approval-workflows/composables/useWorkflowBuilder.ts`
  - builder-specific level/member state and operations
- `contentstudio-frontend/src/modules/approval-workflows/composables/useApprovalNotifications.ts`
  - approval bell notification store
  - currently seeds mock notifications when demo flag is on
- `contentstudio-frontend/src/modules/approval-workflows/composables/usePlannerApprovalDemo.ts`
  - injects demo workflow approval into planner items for visual testing
- `contentstudio-frontend/src/modules/approval-workflows/composables/usePlanStatusDisplay.ts`
  - shared status text for planner rows/cards

### Shared Components

- `contentstudio-frontend/src/modules/approval-workflows/components/ApprovalWorkflowCard.vue`
  - shared workflow card used across settings list and send-for-approval sidebar
- `contentstudio-frontend/src/modules/approval-workflows/components/SendForApprovalSidebar.vue`
  - new main approval UI
  - replaces the old center modal flow
  - supports Users and Approval Workflows tabs
  - supports composer, planner bulk, planner single-post, and CSV/automation usage
- `contentstudio-frontend/src/modules/approval-workflows/components/ApprovalEditConfirmationDialog.vue`
  - composer save/edit confirmation flow for active/rejected approval cases
- `contentstudio-frontend/src/modules/approval-workflows/components/ApprovalNotificationsPanel.vue`
  - approval bell dropdown panel in top bar
- `contentstudio-frontend/src/modules/approval-workflows/components/PlannerApprovalCompactBadge.vue`
  - compact workflow badge + popup used in planner surfaces
- `contentstudio-frontend/src/modules/approval-workflows/components/ApprovalWorkflowBuilder.vue`
  - create/edit workflow page
- `contentstudio-frontend/src/modules/approval-workflows/components/WorkflowLevelCard.vue`
  - builder level card UI

---

## 3. Settings Integration

### Routes and navigation

- `contentstudio-frontend/src/modules/setting/config/routes/setting.js`
  - settings route for workflow list
  - settings route for workflow builder
- `contentstudio-frontend/src/modules/setting/components/SettingSidebar.vue`
  - settings sidebar item for Approval Workflows
- `contentstudio-frontend/src/components/layout/TopHeaderBar.vue`
  - topbar menu item for Approval Workflows under Social Accounts

### Settings pages

- `contentstudio-frontend/src/modules/setting/components/workspace/approval-workflows/ApprovalWorkflowsList.vue`
  - workflow list page
  - tile click -> edit flow
  - delete workflow confirmation
  - empty state
- `contentstudio-frontend/src/modules/setting/components/workspace/approval-workflows/ApprovalWorkflowTile.vue`
  - wrapper around shared workflow card for settings list
- `contentstudio-frontend/src/modules/setting/components/workspace/team/AddTeamMember.vue`
  - permission toggle path for workflow-management access

### Main settings behaviors

- show/create/edit/delete/duplicate/set default workflows
- workflow tiles are clickable
- tooltip on hover
- standard confirm box pattern for delete flows
- localized empty state

---

## 4. Composer Integration

### Main files

- `contentstudio-frontend/src/modules/composer_v2/views/SocialModal.vue`
  - mounts `SendForApprovalSidebar`
  - mounts `ApprovalEditConfirmationDialog`
  - listens to `open-approval-sidebar`
  - owns save interception logic for approval edit scenarios
  - first-time send-for-approval closes composer normally; blocking success confirmation is reserved for explicit edit-approval actions
- `contentstudio-frontend/src/modules/composer_v2/components/MainComposer.vue`
  - shows approval summary/status block
  - `Change approval`
  - `Remove approval`
  - opens sidebar through EventBus
- `contentstudio-frontend/src/modules/composer_v2/components/MainComposerFooter.vue`
  - footer approval button/access checks

### Old component still present

- `contentstudio-frontend/src/modules/composer_v2/components/SendForApprovalModal/ApprovalModal.vue`
  - legacy component
  - no longer the main intended path for this feature
  - should eventually be removed when fully safe

### Composer event flow

1. User clicks `Change approval` in `MainComposer.vue`
2. `MainComposer.vue` emits EventBus event:
   - `open-approval-sidebar`
3. `SocialModal.vue` receives it, sets:
   - `approvalSidebarExistingData`
   - `approvalSidebarIsNotQueued`
   - `showApprovalSidebar = true`
4. `SendForApprovalSidebar.vue` opens and prefills based on existing approval
5. Sidebar emits:
   - `approvalData`
   - EventBus `approval-selected`
6. `SocialModal.vue` stores updated `approval`
7. Save path in `SocialModal.vue` sends approval payload with post save

### Important composer checks

- active approval edit confirmation
- rejected approval edit confirmation
- remove approval confirmation
- workflow preselection on change flow
- send disabled until actual change is made
- first-time send-for-approval should not show the blocking success modal

---

## 5. Planner Integration

### Main files

- `contentstudio-frontend/src/modules/planner_v2/views/MainPlanner.vue`
  - mounts planner send-for-approval sidebars
  - applies planner mock approval decoration
- `contentstudio-frontend/src/modules/planner_v2/components/DataTable.vue`
  - desktop list view
  - compact badge under status
- `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue`
  - calendar badge placement
- `contentstudio-frontend/src/modules/planner_v2/components/SocialMediaViewer/Instagram/FeedItem.vue`
  - feed/grid-like tile badge placement
- `contentstudio-frontend/src/modules/planner_v2/components/DataRow.vue`
  - supporting row renderer
- `contentstudio-frontend/src/modules/planner_v2/components/DataCardMobile.vue`
  - mobile/tablet card
- `contentstudio-frontend/src/modules/planner_v2/components/DataRowCardMobile.vue`
  - mobile row card
- `contentstudio-frontend/src/modules/planner_v2/components/ActionButtons.vue`
  - single-post actions like Send/Change Approval
- `contentstudio-frontend/src/modules/planner_v2/components/PlannerPostPreview_v2.vue`
  - full preview modal renders approval panel
- `contentstudio-frontend/src/modules/planner_v2/components/PlannerAside.vue`
  - aside renderer for approval status
- `contentstudio-frontend/src/modules/planner_v2/components/PlanApprovalStatus.vue`
  - main full approval panel
- `contentstudio-frontend/src/modules/planner_v2/components/WorkflowApprovalLevels.vue`
  - workflow level renderer
- `contentstudio-frontend/src/modules/planner_v2/components/SharePlanModal.vue`
  - override warning for external share-link approval

### Planner behavior split

- compact status:
  - `PlannerApprovalCompactBadge.vue`
- full detail:
  - `PlanApprovalStatus.vue`
- text badge/status label:
  - `usePlanStatusDisplay.ts`

### Current compact badge rules

- compact approval popup is shown for both:
  - workflow approvals (`workflow_levels`)
  - ad-hoc approvals (`approvers`)
- popup is intentionally limited to approval-related planner states:
  - `review` / `reviewed`
  - `missedReview` / `missed_review`
  - `rejected`
- when the compact popup is shown in list/compact-list, the duplicate plain-text status is hidden
- the closed pill uses planner-facing labels/colors:
  - `In Review`
  - `Missed Review`
  - `Rejected`

### Planner mock behavior

- `MainPlanner.vue` decorates eligible plans using `usePlannerApprovalDemo.ts`
- `PlanApprovalStatus.vue` still has a local demo fallback if no real `workflow_levels` exist

That second fallback is important: it means planner demo is **not fully centralized yet**.

---

## 6. Notifications Integration

### Files

- `contentstudio-frontend/src/components/layout/TopHeaderBar.vue`
  - approval bell
  - unread badge
  - panel toggle
- `contentstudio-frontend/src/modules/approval-workflows/components/ApprovalNotificationsPanel.vue`
  - panel UI
- `contentstudio-frontend/src/modules/approval-workflows/composables/useApprovalNotifications.ts`
  - notification store and mock seeding

### Current notification event types in FE

- `approval_request`
- `approved`
- `rejected`
- `level_passed`
- `reminder`

### Important note

Spec includes more backend notification scenarios than current FE mock support.

Examples not yet modeled in FE notification mock types:

- approver removed from workspace
- level removed and auto-advanced
- auto-fully-approved on final level removal

Those will need backend payload + FE notification extension later.

---

## 7. Locale Namespaces

These are the main locale namespaces involved:

### Settings

- `settings.approval_workflows.*`
- settings sidebar label for approval workflows
- team permission label/description for workflow management

### Composer

- `composer.approval_sidebar.*`
- `composer.approval_section.*`
- `composer.approval_edit_dialog.*`
- `composer.approval_status.*`

### Planner

- `planner.filter_sidebar.plan_approval_status.*`

### Approval notifications

- `approval.notifications.*`

### Where locale files live

- `contentstudio-frontend/src/locales/*/settings.json`
- `contentstudio-frontend/src/locales/*/composer.json`
- `contentstudio-frontend/src/locales/*/planner.json`
- `contentstudio-frontend/src/locales/*/approval.json`

---

## 8. Mock / Demo Centralization Status

## What is centralized well

- workflow CRUD mock state:
  - `useApprovalWorkflows.ts`
- planner demo decorator:
  - `usePlannerApprovalDemo.ts`
- mock approval object factory:
  - `getMockApprovalPlan()`

## What is still scattered

- composer injects mock approval directly:
  - `SocialModal.vue`
- planner full panel injects demo fallback directly:
  - `PlanApprovalStatus.vue`
- notifications seed mock data directly:
  - `useApprovalNotifications.ts`
- some UI visibility/access checks directly read `APPROVAL_MOCK_DEMO`:
  - `MainComposer.vue`
  - `MainComposerFooter.vue`
  - `ApprovalWorkflowBuilder.vue`
  - `PlannerAside.vue`
  - other touched surfaces

## Practical answer

Backend cutover is manageable, but not one-switch clean yet.

If cleanup is wanted before backend integration, the next refactor should be:

1. create `useApprovalFeatureFlags()`
2. move all mock/demo booleans behind it
3. keep only 3 explicit mock entry points:
   - workflow CRUD mock store
   - planner demo decorator
   - notification seed

---

## 9. Event / Data Flow Inventory

### EventBus events used by this feature

- `open-approval-sidebar`
  - composer opens approval sidebar
- `approval-selected`
  - sidebar emits selected approval payload after send
- `refreshPlannerTableV2`
  - planner refresh after bulk approval send
- `minimize-composer`
  - sidebar “go to settings” flow in composer context

### Prop / emit patterns to know

- `SendForApprovalSidebar.vue`
  - props:
    - `existingApproval`
    - `accountSelection`
    - `module`
    - `selectedPlanIds`
    - `isNotQueued`
    - `singlePlan`
    - `selectedPlansData`
  - emits:
    - `update:modelValue`
    - `approvalData`
    - `closed`

### Key approval payload builders

- ad-hoc approval payload built in:
  - `SendForApprovalSidebar.vue`
- workflow approval payload built in:
  - `SendForApprovalSidebar.vue`
- approval edit action payload built in:
  - `SocialModal.vue`

---

## 10. Backend Integration Touchpoints

These are the places backend integration will most directly affect:

### Workflow CRUD

- `useApprovalWorkflows.ts`

Replace:

- `MOCK_ENABLED` branches
- module-level mock store

### Planner demo

- `usePlannerApprovalDemo.ts`
- `MainPlanner.vue`
- `PlanApprovalStatus.vue`

Replace/remove:

- visual demo decoration
- local demo fallback in full panel

### Composer demo/state assumptions

- `SocialModal.vue`
- `MainComposer.vue`
- `MainComposerFooter.vue`

Replace/remove:

- direct demo approval injection
- UI visibility checks that only exist for mock mode

### Notifications

- `useApprovalNotifications.ts`
- `ApprovalNotificationsPanel.vue`
- `TopHeaderBar.vue`

Replace/remove:

- seeded notifications
- wire real fetch/mark-read/live update flow

### Workspace member removal scenario

Current state:

- defined in spec
- not implemented as a dedicated FE path yet

Backend integration will likely need:

- updated `plans.approval`
- new notification payload
- possible real-time update path

---

## 11. File Inventory by Responsibility

### Approval workflows module

- `contentstudio-frontend/src/modules/approval-workflows/composables/useApprovalWorkflows.ts`
- `contentstudio-frontend/src/modules/approval-workflows/composables/useWorkflowBuilder.ts`
- `contentstudio-frontend/src/modules/approval-workflows/composables/useApprovalNotifications.ts`
- `contentstudio-frontend/src/modules/approval-workflows/composables/usePlannerApprovalDemo.ts`
- `contentstudio-frontend/src/modules/approval-workflows/composables/usePlanStatusDisplay.ts`
- `contentstudio-frontend/src/modules/approval-workflows/components/ApprovalWorkflowBuilder.vue`
- `contentstudio-frontend/src/modules/approval-workflows/components/WorkflowLevelCard.vue`
- `contentstudio-frontend/src/modules/approval-workflows/components/ApprovalWorkflowCard.vue`
- `contentstudio-frontend/src/modules/approval-workflows/components/SendForApprovalSidebar.vue`
- `contentstudio-frontend/src/modules/approval-workflows/components/ApprovalEditConfirmationDialog.vue`
- `contentstudio-frontend/src/modules/approval-workflows/components/ApprovalNotificationsPanel.vue`
- `contentstudio-frontend/src/modules/approval-workflows/components/PlannerApprovalCompactBadge.vue`

### Settings integration

- `contentstudio-frontend/src/modules/setting/config/routes/setting.js`
- `contentstudio-frontend/src/modules/setting/components/SettingSidebar.vue`
- `contentstudio-frontend/src/modules/setting/components/workspace/approval-workflows/ApprovalWorkflowsList.vue`
- `contentstudio-frontend/src/modules/setting/components/workspace/approval-workflows/ApprovalWorkflowTile.vue`
- `contentstudio-frontend/src/modules/setting/components/workspace/team/AddTeamMember.vue`
- `contentstudio-frontend/src/components/layout/TopHeaderBar.vue`

### Composer integration

- `contentstudio-frontend/src/modules/composer_v2/views/SocialModal.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/MainComposer.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/MainComposerFooter.vue`
- `contentstudio-frontend/src/modules/composer_v2/components/SendForApprovalModal/ApprovalModal.vue`

### Planner integration

- `contentstudio-frontend/src/modules/planner_v2/views/MainPlanner.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/DataTable.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/DataRow.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/DataCardMobile.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/DataRowCardMobile.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/ActionButtons.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/SocialMediaViewer/Instagram/FeedItem.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/PlannerPostPreview_v2.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/PlannerAside.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/PlanApprovalStatus.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/WorkflowApprovalLevels.vue`
- `contentstudio-frontend/src/modules/planner_v2/components/SharePlanModal.vue`

---

## 12. What a Backend Dev Should Read First

Recommended order:

1. `06-frontend-data-contract.md`
2. this file
3. `03-prd.md`
4. `07-final-qa-checklist.md`

Then in code:

1. `useApprovalWorkflows.ts`
2. `SendForApprovalSidebar.vue`
3. `SocialModal.vue`
4. `PlanApprovalStatus.vue`
5. `useApprovalNotifications.ts`

That is the fastest path to understanding the feature without reading the whole repo.
