# Research: Filter Button Applied Count Badge

## Current State

### Planner Header
The filter button in `PlannerHeader.vue` toggles between `text`/`secondary` (no filters) and `outline`/`primary` (filters applied) using the `areFiltersApplied` boolean prop. No count is shown.

`areFiltersApplied` is computed in `MainPlanner.vue:633` — it checks route query params for 8 filter keys (`statuses`, `members`, `created_by`, `type`, `labels`, `campaigns`, `content_category`, `date`) and returns `true` if any have a value. This could be changed to return a count instead (count of params with non-empty values).

The filter button appears twice in `PlannerHeader.vue`:
- Mobile layout (line ~7): `Button` with filter icon
- Desktop layout (line ~303): `Button` with filter icon + "Filters" text label

Neither shows a count badge.

### Inbox FilterDrawer
The filter button in `FilterDrawer.vue` (inbox-revamp) uses `hasFilterSelected` boolean — same pattern: toggles button variant/color but shows no count.

The `FilterDrawer.vue` is rendered in `InboxView.vue:19` in the left sidebar header area.

### Reference: Custom View filter count
The user mentioned the create/edit custom view shows a filter count. This would be in the publisher/planner custom view components — the pattern for showing a count badge on a button already exists somewhere in the codebase.

## What Needs to Change

1. **Planner header filter button** — Show a count badge (e.g., `Badge` component or inline counter) on the filter button when filters are applied. The count should reflect how many distinct filter categories are active (e.g., if statuses + labels are set → count = 2). Modify `areFiltersApplied` in `MainPlanner.vue` to compute a count instead of just a boolean, and pass it down to `PlannerHeader.vue`.

2. **Inbox filter button** — Show the same count badge on the filter button in `FilterDrawer.vue` when filters are applied. Compute a count from `hasFilterSelected` logic.

## Files Involved

### Planner
- `src/modules/planner_v2/views/MainPlanner.vue` — change `areFiltersApplied` to also compute filter count
- `src/modules/planner_v2/components/PlannerHeader.vue` — add count badge to filter button (both mobile and desktop instances)

### Inbox
- `src/modules/inbox-revamp/components/FilterDrawer.vue` — compute filter count from selected filters and show badge on the filter button
- `src/modules/inbox-revamp/views/InboxView.vue` — no changes needed (FilterDrawer is self-contained)
