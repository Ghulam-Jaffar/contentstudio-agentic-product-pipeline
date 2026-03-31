# Multi-Tiered Approval Workflow — Shortcut Links

## Epic
- **Epic:** [Multi-Tiered Approval Workflow](https://app.shortcut.com/contentstudio-team/epic/47607)

## Shortcut Docs
- **Workflow Design:** [69c4c66f-2ad5-4acc-bd65-772bbd686eed](https://app.shortcut.com/contentstudio-team/documents/69c4c66f-2ad5-4acc-bd65-772bbd686eed)
- **PRD:** [69c4c67d-0ee9-4aa5-aea0-147aab8da8f5](https://app.shortcut.com/contentstudio-team/documents/69c4c67d-0ee9-4aa5-aea0-147aab8da8f5)
- **Frontend Data Contract:** [69caeb16-c1b6-4433-b485-4d4d27f71a49](https://app.shortcut.com/contentstudio-team/write/IkRvYyI6I3V1aWQgIjY5Y2FlYjE2LWMxYjYtNDQzMy1iNDg1LTRkNGQyN2Y3MWE0OSI=)
- **Final QA Checklist:** [69caeb17-faf3-4209-ae97-d2203dc39a26](https://app.shortcut.com/contentstudio-team/write/IkRvYyI6I3V1aWQgIjY5Y2FlYjE3LWZhZjMtNDIwOS1hZTk3LWQyMjAzZGMzOWEyNiI=)
- **Frontend Implementation Map:** [69caeb18-f4b1-438c-93cc-3d06e5ec810e](https://app.shortcut.com/contentstudio-team/write/IkRvYyI6I3V1aWQgIjY5Y2FlYjE4LWY0YjEtNDM4Yy05M2NjLTNkMDZlNWVjODEwZSI=)

## Local Working Docs
- **Frontend Data Contract:** [06-frontend-data-contract.md](./06-frontend-data-contract.md)
- **Final QA Checklist:** [07-final-qa-checklist.md](./07-final-qa-checklist.md)
- **Frontend Implementation Map:** [08-frontend-implementation-map.md](./08-frontend-implementation-map.md)

---

## Stories

### Frontend Stories — Current Iteration (113539, 23 Mar–03 Apr 2026)

| Story ID | Shortcut | Title |
|---|---|---|
| S-09 | [sc-114804](https://app.shortcut.com/contentstudio-team/story/114804) | [FE] Build Approval Workflows settings page (workflow list) |
| S-10 | [sc-114810](https://app.shortcut.com/contentstudio-team/story/114810) | [FE] Build Create/Edit Workflow builder with levels, members, and drag-and-drop |
| S-11 | [sc-114816](https://app.shortcut.com/contentstudio-team/story/114816) | [FE] Migrate Send for Approval from center modal to right sidebar with Users and Approval Workflows tabs |
| S-12 | [sc-114822](https://app.shortcut.com/contentstudio-team/story/114822) | [FE] Build approval status tracking panel (planner hover popup and post preview modal) |
| S-13 | [sc-114828](https://app.shortcut.com/contentstudio-team/story/114828) | [FE] Build edit confirmation dialogs for posts in approval (all 4 scenarios) |
| S-14 | [sc-114834](https://app.shortcut.com/contentstudio-team/story/114834) | [FE] Build dedicated Approval Notifications icon and panel in the top bar |
| S-15 | [sc-114840](https://app.shortcut.com/contentstudio-team/story/114840) | [FE] Add "Manage Approval Workflows" permission toggle to Team Member Settings |
| S-16 | [sc-114846](https://app.shortcut.com/contentstudio-team/story/114846) | [FE] Update Planner post status badges and filter labels for multi-level approval |
| S-17 | [sc-114852](https://app.shortcut.com/contentstudio-team/story/114852) | [Design] Approval Workflow Settings, Create/Edit builder, and Send for Approval sidebar |
| S-23 | [sc-114858](https://app.shortcut.com/contentstudio-team/story/114858) | [FE] Add "Send for Approval" to single Planner post card and post detail view |
| S-24 | [sc-114864](https://app.shortcut.com/contentstudio-team/story/114864) | [FE] Add approval override warning to Send for Approval sidebar, bulk Planner send, and external share link |

### Backend Stories — Next Iteration (114517, 06 Apr–17 Apr 2026)

| Story ID | Shortcut | Title |
|---|---|---|
| S-01 | [sc-114870](https://app.shortcut.com/contentstudio-team/story/114870) | [BE] Create approval_workflows collection and CRUD API |
| S-02 | [sc-114876](https://app.shortcut.com/contentstudio-team/story/114876) | [BE] Extend plans.approval schema for multi-level workflow state |
| S-03 | [sc-114882](https://app.shortcut.com/contentstudio-team/story/114882) | [BE] Implement WorkflowApprovalBuilder for multi-level approval progression |
| S-04 | [sc-114888](https://app.shortcut.com/contentstudio-team/story/114888) | [BE] Add new notification event types for multi-tiered approval |
| S-05 | [sc-114894](https://app.shortcut.com/contentstudio-team/story/114894) | [BE] Implement workflow mid-flight modification rules |
| S-06 | [sc-114900](https://app.shortcut.com/contentstudio-team/story/114900) | [BE] Fix rejected post silent status reset bug in ApprovalBuilder |
| S-07 | [sc-114906](https://app.shortcut.com/contentstudio-team/story/114906) | [BE] Add "Allow Approval Workflow Management" permission to team member settings |
| S-08 | [sc-114912](https://app.shortcut.com/contentstudio-team/story/114912) | [BE] Handle approver removal from workspace with in-flight approval impacts |
| S-22 | [sc-114918](https://app.shortcut.com/contentstudio-team/story/114918) | [BE] Implement approval override cleanup and cancellation notification dispatch |

### Mobile Stories — Next Iteration (114517, 06 Apr–17 Apr 2026)

| Story ID | Shortcut | Title |
|---|---|---|
| S-18 | [sc-114924](https://app.shortcut.com/contentstudio-team/story/114924) | [iOS] Add Approval Workflows tab to Send for Approval in Composer |
| S-19 | [sc-114930](https://app.shortcut.com/contentstudio-team/story/114930) | [iOS] Add "Pending Approvals" bottom navigation item and view |
| S-20 | [sc-114936](https://app.shortcut.com/contentstudio-team/story/114936) | [Android] Add Approval Workflows tab to Send for Approval in Composer |
| S-21 | [sc-114942](https://app.shortcut.com/contentstudio-team/story/114942) | [Android] Add "Pending Approvals" bottom navigation item and view |
