# Stories: Settings UI/UX revamp (Account, Workspace, Integrations)

Two stories (one design, one build) to modernize three settings pages so they are consistent with the rest of the app and use the design library.

Pages in scope (all under Settings):
- Account settings
- Workspace settings
- Integrations (the "other integrations" page)

---

## [Design] Redesign the Account, Workspace, and Integrations settings pages

### Description

The Account settings, Workspace settings, and Integrations pages look dated and inconsistent with the rest of the app. Produce updated designs for all three that match current ContentStudio patterns and use design-library components, so they feel like one product.

### Acceptance criteria

- [ ] Updated designs delivered for all three pages: Account settings, Workspace settings, Integrations.
- [ ] Designs use existing design-library components and patterns (layout, headers, cards, form fields, buttons) rather than one-off styling.
- [ ] Designs are visually consistent with recently redesigned areas of the app.
- [ ] Designs cover the key states for each page: default, empty (where applicable), loading, and error/validation.
- [ ] Designs follow current spacing, typography, and color tokens (no dark mode, no RTL).
- [ ] Handoff includes any new component or pattern needs flagged for the frontend build.

### Dependencies

None. Feeds the frontend build story below.

### Global quality and compliance checklist

- [ ] Mobile responsiveness considered in the designs
- [ ] Multilingual support considered (text expansion, no fixed-width labels)
- [ ] UI theming supported (default and white-label) — no dark mode
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

## [FE] Rebuild the Account, Workspace, and Integrations settings pages to match the new design

### Description

Implement the redesigned Account settings, Workspace settings, and Integrations pages using design-library components, so all three are consistent with the rest of the app.

### Workflow

1. User opens Settings and navigates to Account settings, Workspace settings, or Integrations.
2. Each page presents the same fields and actions as before, in the updated, consistent layout.
3. All existing actions (editing profile/account details, workspace details, connecting/managing integrations) continue to work.

### Acceptance criteria

- [ ] All three pages match the approved design and use `@contentstudio/ui` (and shared) components instead of one-off markup.
- [ ] Every existing field, setting, and action on each page still works after the rebuild.
- [ ] Pages are visually consistent with each other and with the rest of the app.
- [ ] Loading, empty (where applicable), and error/validation states are implemented per the design.
- [ ] No hardcoded colors; theme-aware classes are used so white-label theming still applies.
- [ ] The pages are responsive.

### UI copy

Reuse existing labels, field copy, and messages on these pages. Only add copy where the design introduces a new element (empty states or helper text); flag any such new copy during the build.

### Impact on existing data

None. Presentation and layout only; no changes to what the settings store.

### Impact on other products

Web only (Settings). Mobile and Chrome extension: N/A.

### Dependencies

Depends on: **[Design] Redesign the Account, Workspace, and Integrations settings pages**.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only)
- [ ] Multilingual support verified (translations available or fallback handled)
- [ ] UI theming supported (default and white-label) — no dark mode
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- Account settings: `contentstudio-frontend/src/modules/setting/components/Profile.vue`.
- Workspace settings: `contentstudio-frontend/src/modules/setting/components/workspace/BasicSetting.vue` (and `ManageWorkspacesMain.vue` / `WorkspaceFields.vue`).
- Integrations page: the integrations settings surface (confirm exact path, for example under `modules/setting/components/` or the `integration` module).
- Use design-library components per `docs/ui-components.md`; prefer `@contentstudio/ui` over legacy `Cst*`. Keep changes scoped to layout/components, not the underlying settings logic.
