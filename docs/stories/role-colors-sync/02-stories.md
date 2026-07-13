# Story: Consistent team-member role colors

## [FE] Use one consistent role color mapping across the platform

### Description

Team member role colors (for example Super Admin, Admin, Collaborator, Approver) are rendered inconsistently across the app. The same role can appear in different colors in different places. Centralize a single role-to-color mapping and reuse it everywhere roles are shown, so a role always appears in the same color.

### Workflow

1. Wherever a team member role is displayed (approval workflow, workspace manage-team, the team members page, and so on), the role appears in one consistent color.
2. Updating or adding a role color in the single source updates it everywhere the role appears.

### Acceptance criteria

- [ ] A single source defines the color for each team member role.
- [ ] The same role uses the same color across: the approval workflow, the all-workspace and manage-team sections, the team members page, the send-for-approval modal, and planner approval status.
- [ ] No screen shows a different color for the same role.
- [ ] Adding or changing a role reflects the same color everywhere it appears.
- [ ] Colors are theme-aware where applicable and are not hardcoded per component.

### UI copy

None.

### Impact on existing data

None. Presentation only.

### Impact on other products

Web focus. If mobile apps display roles, confirm they consume the same mapping or mirror it. Chrome extension: N/A.

### Dependencies

None.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only)
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default and white-label) — no dark mode
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- The team API already returns a `color` for roles/members (`contentstudio-frontend/src/api/team.ts`, `color: string`). Prefer that API color, or a single shared mapping/composable, as the one source of truth.
- Converge these current usages onto the shared source:
  - `modules/planner_v2/components/WorkflowApprovalLevels.vue`, `PlanApprovalStatus.vue`
  - `modules/composer_v2/components/SendForApprovalModal/ApprovalModal.vue`
  - `modules/common/components/planner-shared/FeedViewApprovalStatus.vue`
  - `modules/setting/components/workspace/team/AddTeamMember.vue`, `Team.vue`
  - `modules/setting/components/workspace/approval-workflows/ApprovalWorkflowsList.vue`
  - `modules/setting/components/manage-limits/tabs/TeamsTab.vue`
  - `modules/setting/components/organization/ManageTeamMember.vue`
- Prefer theme-aware classes or CSS variables; avoid hardcoded hex.
