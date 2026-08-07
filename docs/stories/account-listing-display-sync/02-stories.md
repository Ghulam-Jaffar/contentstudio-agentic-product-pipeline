# Story: Consistent social account listing

## [FE] Standardize how connected social accounts are displayed in lists

### Description

Connected social account rows look different in different parts of the app. Standardize each account row to show the profile picture with a small platform badge, followed by the account name and its type, and reuse a single component everywhere accounts are listed (composer, planner, inbox, analytics).

### Workflow

1. Anywhere connected social accounts are listed, each row shows: profile picture with a small platform badge, then the account name and account type.
2. The same layout appears consistently across composer, planner, inbox, and analytics.

### Acceptance criteria

- [ ] A single reusable account-row component renders: profile picture plus platform badge, then account name and type.
- [ ] The component is used in account listings across composer, planner, inbox, and analytics.
- [ ] Account rows look visually consistent across those areas.
- [ ] A missing profile picture falls back sensibly (initials or a default avatar).
- [ ] The platform badge correctly reflects the account platform for every supported platform.

### UI copy

Row content is the account name and its type; no new copy strings beyond existing account-type labels.

### Impact on existing data

None. Presentation only.

### Impact on other products

Web focus (composer, planner, inbox, analytics). Confirm mobile separately if it shows account lists. Chrome extension: N/A.

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

- Existing item component to promote and reuse: `contentstudio-frontend/src/modules/planner_v2/components/SocialAccountListItem.vue`.
- Listing surfaces to converge on the shared component:
  - `modules/composer_v2/components/AccountSelectionDropdown.vue`, `AccountSelectionAside.vue`
  - `modules/planner_v2/components/SocialAccountsDrawer.vue`, `FilterSidebar.vue`
  - `modules/analytics_v3/components/ReportTile.vue`
  - `components/dashboard/SocialAccountsCard.vue`
  - the inbox account lists (inbox-revamp)
- If the component needs to live outside `planner_v2`, move it to `src/components/common` so all modules can import it.
