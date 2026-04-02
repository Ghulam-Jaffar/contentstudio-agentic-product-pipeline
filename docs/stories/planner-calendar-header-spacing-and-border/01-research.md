# Research: Calendar View Header Spacing & Border

## Current State

The planner has 3 view modes sharing the same header component (`PlannerHeader.vue`) rendered inside `MainPlanner.vue`. The header is wrapped in a sticky `div` with `bg-white p-[0.156rem]` at `MainPlanner.vue:20`.

**List view** — `DataTable.vue` applies `shadow-sm` (line 225) to the table container, which creates a subtle visual separation between header and content.

**Feed view** — The feed cards have their own card-style borders providing visual separation.

**Calendar view** — `CalenderView.vue` has no border, shadow, or visual separator. Combined with the header having no `border-b`, the result is a gap of white space between the header and the calendar grid with no visual divider.

The `PlannerHeader.vue` component has **no `border-b`** class anywhere — the bottom border/divider is completely absent.

## What Needs to Change

1. **Reduce white space** between the planner header and the calendar component below it — the `p-[0.156rem]` padding on the header wrapper and any margin/gap between `#mainHeader` and `#mainView` in `MainPlanner.vue` contributes to excess spacing in calendar view
2. **Add a bottom border or divider** to the planner header (or top border on the calendar) to visually separate header from content — consistent with how list and feed views have visual separation

## Files Involved

- `src/modules/planner_v2/views/MainPlanner.vue` — header wrapper div (line 20) padding and spacing between `#mainHeader` and `#mainView`
- `src/modules/planner_v2/components/PlannerHeader.vue` — add bottom border to the header
- `src/modules/planner_v2/views/CalenderView.vue` — potentially reduce top padding/margin on the calendar container
