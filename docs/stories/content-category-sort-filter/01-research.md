# Research: Sort Filter Button in Content Category Dropdown

## Current State

The content category dropdown is implemented in:
- **Web App:** `contentstudio-frontend/src/modules/publish/components/posting/social/ContentCategorySelection.vue`
  - Currently sorts A-Z by default using `sortBy(data, item => [item.name.toLowerCase()])` (lodash-es)
  - Already has `SearchInput` from `@contentstudio/ui` for text filtering
  - No sort order toggle exists yet

The **Chrome Extension** has its own category dropdown (separate codebase, not mounted here).

## Existing Sort Patterns in Codebase

The closest existing pattern is in `contentstudio-frontend/src/modules/publisher/components/CreateCustomView.vue`, which uses a `sortOrders` array with keys like `first_created` / `last_created` (oldest/newest by creation time). Sort is done by `_id` (MongoDB ObjectId encodes creation timestamp — higher ObjectId = newer).

**Category object fields available for sorting:**
- `_id` — MongoDB ObjectId; encodes creation time (can sort oldest/newest)
- `name` — string; used for A-Z sort

**Confirmed available sort options for this feature:**
| Label | Sort logic |
|---|---|
| A–Z | `sortBy(list, item => item.name.toLowerCase())` (current default) |
| Newest first | Sort by `_id` descending (newer ObjectId = more recent) |
| Oldest first | Sort by `_id` ascending |

## What Needs to Change

**Web App (`ContentCategorySelection.vue`):**
- Add a sort icon button next to the `SearchInput` in the dropdown header
- Clicking it reveals a small sort menu (3 options: A–Z, Newest first, Oldest first)
- Selected sort is applied to `filteredCategories` computed
- Default sort remains A–Z (current behavior)
- Active sort option is visually highlighted

**Chrome Extension:**
- Same UX, same 3 sort options — implemented in the Chrome Extension's category dropdown component(s)

## UI Components Available

From `docs/ui-components.md`:
- `Dropdown` + `DropdownItem` — for the sort options menu
- `ActionIcon` — icon-only clickable trigger for the sort button
- `Icon` — for sort icon (e.g., `ArrowUpDown` or `ListFilter`)

No new components needed.

## Files Involved

**Web App:**
- `contentstudio-frontend/src/modules/publish/components/posting/social/ContentCategorySelection.vue` — primary change
- `contentstudio-frontend/src/locales/*/publish.json` — i18n keys for sort option labels

**Chrome Extension:**
- Chrome Extension's content category dropdown component(s) — exact paths to be confirmed by dev team
