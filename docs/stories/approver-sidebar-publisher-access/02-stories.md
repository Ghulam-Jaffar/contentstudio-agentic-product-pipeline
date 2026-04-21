# Story — Enable Publisher Sidebar for Approver Role

Single `[FE]` story. No backend, no mobile, no design work.

## [FE] Show publisher sidebar to approver role with planner access and permission-gated composer

### Description
As an approver, I want to see the left sidebar inside the publisher so that I can access the planner and my custom views even if I don't have permission to create posts — the sidebar is hidden from me entirely today, so I have no way to browse or apply my saved planner views.

If I do have create-post permission, the sidebar also shows the Composer entry (with the "Use template" option) at the top. I should never see AI Studio, Automations, or Planner Settings — those remain owner/admin/collaborator-only.

---

### Workflow
1. User with role = **Approver** (with or without `approverCanCreatePost`) logs in and opens the **Publisher** from the top navigation.
2. The left sidebar is now visible (today it's hidden for approvers without `approverCanCreatePost`).
3. If the approver has `approverCanCreatePost = true`, the **Compose** split button is shown at the top of the sidebar with the same dropdown approvers-with-create-post already see today — Social Post action + "Use template" option. Blog post action stays hidden.
4. If the approver has `approverCanCreatePost = false`, no Compose button is shown in the sidebar — but the sidebar itself still renders.
5. Directly below the Compose button (or at the top when Compose is hidden), the **Planner** collapsible is shown — expanded by default — containing:
   - The "All Posts" static default view
   - The list of the approver's saved planner custom views (with drag-to-reorder, actions menu, set-default, edit, delete)
   - The "Create new view" affordance
6. Nothing is shown below Planner. The AI Studio, Automations, and Planner Settings sections remain hidden for approvers.
7. The collapsed (shrunk) sidebar state mirrors the expanded behavior: Compose icon only if `approverCanCreatePost = true`; Planner dropdown always visible; no other icons.
8. Clicking any custom view in the sidebar applies it to the planner content — same behavior as other roles.

---

### Acceptance criteria
- [ ] Approvers without `approverCanCreatePost` now see the publisher sidebar (today it is hidden). Verified by logging in as an approver with `approverCanCreatePost = false` and opening the publisher.
- [ ] Approvers with `approverCanCreatePost = true` continue to see the sidebar exactly as they do today (no regression). Verified by logging in with `approverCanCreatePost = true`.
- [ ] Non-approver users (owner, admin, collaborator, super_admin) see the sidebar exactly as they do today (no regression).
- [ ] For approvers, the `canAccessSidebar` computed in `contentstudio-frontend/src/modules/publisher/views/PublisherMain.vue` returns `true` regardless of `approverCanCreatePost`. Preferred implementation: add an `isApprover()` helper in `contentstudio-frontend/src/composables/usePermission.ts` and use it in the computed — centralises role checks.
- [ ] Compose button at the top of the sidebar (expanded and shrunk) is shown for an approver **only if** `hasPermission('can_create_post')` is true. This is already the behavior of `canAccessTopHeader` inside `SidebarMain.vue`; this story must not change it.
- [ ] Compose dropdown for an approver with create-post permission includes the "Use template" option (already controlled by `:show-template-attachment="true"` passed into `ComposeActionsDropdown`) — no change needed, just verified.
- [ ] Compose dropdown for an approver does **not** include the Blog post action (already gated by `:can-access-blog-composer="hasFullAccess"` — `hasFullAccess` is false for approvers).
- [ ] Planner collapsible (expanded sidebar) and Planner dropdown (shrunk sidebar) are visible for every approver — with or without create-post permission.
- [ ] All Planner custom view actions — click to apply, drag to reorder, open actions menu (set default / remove default / edit / delete), "Create new view", show more / show less — continue to work for approvers just like they do for other roles.
- [ ] AI Studio collapsible is **not** visible for any approver, regardless of `approverCanCreatePost` (already gated by `hasFullAccess` — verified).
- [ ] Automations collapsible is **not** visible for any approver (already gated by `hasFullAccess` — verified).
- [ ] Planner Settings collapsible is **not** visible for any approver (already gated by `hasFullAccess` — verified).
- [ ] Shrunk sidebar for an approver without create-post permission shows only the Planner icon — no compose icon, no other items.
- [ ] Sidebar collapse / expand toggle continues to work for approvers.
- [ ] No layout shift or empty-column visual glitch on the approver's publisher landing — the router-view width adjusts to the sidebar width the same way it does for other roles.

---

### Mock-ups
N/A — this story exposes an already-designed sidebar to an additional role. No new UI elements, no copy changes, no visual design work.

---

### Impact on existing data
None. Pure UI gating change. No migrations, no API calls, no persisted preferences touched.

---

### Impact on other products
- **Web app:** the only product affected. Approvers will now see the sidebar on the `/publisher/*` route tree.
- **Mobile apps (iOS / Android):** no impact. The mobile planner has its own navigation and is not driven by this sidebar.
- **Chrome extension:** no impact.
- **White-label domains:** no impact — same theming stack applies.

---

### Dependencies
None. Self-contained change in `PublisherMain.vue` (and optionally `usePermission.ts`).

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — verify the small-screen breakpoint (`small-screen-range:hidden`) on `PublisherMain.vue:3` still hides the sidebar on mobile widths for approvers, same as other roles
- [ ] Multilingual support — N/A, no user-facing strings introduced or changed; existing i18n keys already cover the sidebar
- [ ] UI theming support — N/A, no new components or styles introduced; the existing sidebar already uses `@contentstudio/ui` components and CSS variables for theming (confirmed across white-label domains on other roles)
- [ ] White-label domains impact review — N/A, theming untouched
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension) — N/A for mobile / Chrome (unaffected); web-only change

---

## Metadata for push

| Field | Value |
|---|---|
| Prefix | `[FE]` |
| Project | Web App |
| Group | Frontend |
| Priority | Medium |
| Product area | Planner |
| Skill set | Frontend |
| Epic | Q2 2026 Miscellaneous (no dedicated epic) |
| Iteration | Current iteration — confirm at push time |
| Story template | New Feature Template |
