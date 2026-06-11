# Research: Content Category Dropdown UX Improvements in Composer

## Current State

The Content Category dropdown with the sort filter lives in:
`contentstudio-frontend/src/modules/composer_v2/components/AccountSelectionAside.vue`

The outer container is a `CstDropdown` (legacy component) with `@on-close="resetCategorySearch"`. No max-height is set on this dropdown — it uses default height.

**Header area** (lines ~119–200): A `div.px-3.py-2` containing:
- `SearchInput` — full-width, `display-mode="full-width"`, `expand-on-focus="false"`
- Inner `Dropdown` (sort filter) with `placement="bottom-end"` — opens below and right of the trigger
- The sort button shows a static `ArrowUpDown` icon (line ~153), not the Planner filter icon
- A small blue dot indicator appears when `sortOrder !== 'az'` — but the icon itself never changes

**Sort options** (lines 866–884):
- A-Z (`ArrowDownAZ`), Newest (`ArrowDownWideNarrow`), Oldest (`ArrowUpNarrowWide`)
- `sortOrder` data property, default `'az'`

**Sort reset on close** (lines 1403–1408):
```js
resetCategorySearch() {
  this.clearCategorySearch()
  this.sortOrder = 'az'  // ← resets sort on every dropdown close
}
```
This is called via `@on-close="resetCategorySearch"` on the outer `CstDropdown`.

**Planner filter icon:**
A custom inline SVG (two horizontal sliders with circular handles) in `PlannerHeader.vue` lines 16–28. Fill color is `#4A4A4A` when inactive, `rgb(var(--cstu-primary-500))` when filters are active.

**Chrome Extension:** Separate codebase not mounted in this repo. Has its own version of the category dropdown with the same filter. Chrome Extension story should mirror web app fixes.

## What Needs to Change

1. **Dropdown height**: Expand the `CstDropdown` to 70% viewport height
2. **Persistent dropdown on filter selection**: Prevent the inner `Dropdown` (sort filter) from causing the outer `CstDropdown` to close when a sort option is clicked
3. **Filter state persistence**: Remove `this.sortOrder = 'az'` from `resetCategorySearch()` so the selected sort is retained across dropdown open/close cycles
4. **Sticky header with divider**: Pin the search bar + sort filter row to the top of the dropdown so they don't scroll away; add a visual divider between the header and the category list
5. **Filter dropdown positioning**: Change `placement="bottom-end"` to open the sort sub-dropdown to the right of the main panel, not below it within the list
6. **Reuse Planner filter icon**: Replace `ArrowUpDown` with the custom filter SVG from `PlannerHeader.vue`
7. **Dynamic filter icon swap**: When a sort option is selected, show that option's icon on the sort button (`ArrowDownAZ`, `ArrowDownWideNarrow`, or `ArrowUpNarrowWide`); when no sort is actively selected (or cleared), revert to the default Planner filter icon

## Files Involved

- `contentstudio-frontend/src/modules/composer_v2/components/AccountSelectionAside.vue` — primary file (template lines ~57–244, data lines ~865–885, `resetCategorySearch` lines ~1403–1408)
- `contentstudio-frontend/src/modules/planner_v2/components/PlannerHeader.vue` — reference for the filter SVG icon (lines 16–28)
