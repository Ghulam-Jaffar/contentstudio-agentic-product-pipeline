# Shortcut Links — Enable Publisher Sidebar for Approver Role

## Story
- [#116924 — [FE] Show publisher sidebar to approver role with planner access and permission-gated composer](https://app.shortcut.com/contentstudio-team/story/116924)

## Epic
- Q2 2026 Miscellaneous (id `115078`) — no dedicated epic for this quick fix

## Metadata
- **Iteration:** `20 April - 01 May - 2026` (id `115537`) — current active sprint
- **Workflow state:** Ready for Dev
- **Group:** Frontend
- **Project:** Web App
- **Priority:** Medium
- **Product area:** Planner
- **Skill set:** Frontend
- **Story template:** New Feature Template (5 checklist tasks added)

## Scope recap
Single-file change in [contentstudio-frontend/src/modules/publisher/views/PublisherMain.vue](contentstudio-frontend/src/modules/publisher/views/PublisherMain.vue) — expand `canAccessSidebar` computed to include approvers regardless of `approverCanCreatePost`. Optionally add an `isApprover()` helper in [usePermission.ts](contentstudio-frontend/src/composables/usePermission.ts). All internal guards inside `SidebarMain.vue` stay as-is.
