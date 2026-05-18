# Stories: Sort Filter Button in Content Category Dropdown

---

## Story 1

**Title:** `[FE] Add sort filter button to content category dropdown (Web App)`

---

### Description:

As a ContentStudio user selecting a content category while composing or scheduling a post, I want to be able to sort the category list by name or by when each category was created so that I can find the right category faster — whether I know it alphabetically or I'm looking for one I set up recently — especially when my workspace has many categories.

---

### Workflow:

1. User opens any content category dropdown in the Web App — this includes the Composer (all social channel tabs), RSS Automation save screen, Evergreen Automation create/edit screen, and Bulk CSV Automation save screen.
2. The dropdown header contains two controls side by side:
   - The existing **search field** (`SearchInput`) for text-based filtering
   - A new **sort icon button** immediately to the right of the search field
3. The sort button shows a sort/filter icon (e.g., `ListFilter` or `ArrowUpDown`). On hover, a tooltip reads: **"Sort categories"**.
4. User clicks the sort button — a small menu opens below (or above if near the bottom of the screen) containing three options:

   | Option | Description shown in menu |
   |---|---|
   | **A–Z** | Alphabetical order |
   | **Newest first** | Most recently created categories appear at the top |
   | **Oldest first** | Earliest created categories appear at the top |

5. The currently active sort (default: **A–Z**) is indicated with a checkmark or highlighted background so the user always knows which order is in effect.
6. User clicks a sort option:
   - The sort menu closes.
   - The category list reorders immediately — no API call, no loading state.
   - The sort button (or a small indicator next to it) updates to signal that a non-default sort is active, so the user can tell at a glance without reopening the menu.
7. If the user has an active search term, the sort is applied to the already-filtered results — the two controls compose cleanly without either resetting the other.
8. User selects a category from the sorted (and possibly filtered) list — the dropdown closes, the chosen category is applied to the post/automation, and both the sort and search state reset to defaults.
9. When the dropdown closes for any reason (category selected, clicked outside, or Escape key), the sort resets to **A–Z** and the search field clears — the next time the dropdown opens it starts fresh.

---

### Acceptance criteria:

**Sort button presence:**
- [ ] A sort icon button appears in the content category dropdown header in every context where the dropdown is rendered: Composer (all channel tabs), RSS Automation, Evergreen Automation, and Bulk CSV Automation
- [ ] The sort button is positioned immediately to the right of the existing search field within the dropdown header row
- [ ] On hover, the sort button displays a tooltip: **"Sort categories"**
- [ ] When the category list is empty (no categories exist in the workspace), the sort button is hidden or disabled — there is nothing to sort

**Sort menu:**
- [ ] Clicking the sort button opens a menu with exactly three labelled options:
  - **"A–Z"** — with subtext or aria-label: *"Alphabetical order"*
  - **"Newest first"** — with subtext or aria-label: *"Most recently created first"*
  - **"Oldest first"** — with subtext or aria-label: *"Earliest created first"*
- [ ] The currently active sort option is visually distinguished with a checkmark icon or highlighted background
- [ ] Clicking outside the sort menu without selecting an option closes the menu without changing the active sort

**Sort behaviour — A–Z:**
- [ ] When **A–Z** is selected, categories are ordered alphabetically by name, case-insensitively (e.g., "Analytics" before "Blog", "blog daily" and "Blog Weekly" are adjacent regardless of capitalisation)
- [ ] A–Z is the default sort every time the dropdown opens

**Sort behaviour — Newest first:**
- [ ] When **Newest first** is selected, the most recently created category appears at the top of the list
- [ ] Categories created on the same day appear in a consistent sub-order (e.g., by name)

**Sort behaviour — Oldest first:**
- [ ] When **Oldest first** is selected, the earliest created category appears at the top of the list
- [ ] Categories created on the same day appear in a consistent sub-order (e.g., by name)

**Sort + search composition:**
- [ ] When a search term is active and the user changes the sort, the filtered results re-sort immediately — the search text is preserved
- [ ] When a sort is active and the user types in the search field, the displayed results are filtered from the full list and sorted according to the active sort order
- [ ] Clearing the search field (via the clear button or backspace) returns the full sorted list — the active sort order is preserved

**Reset behaviour:**
- [ ] When the dropdown closes (by category selection, clicking outside, or pressing Escape), the sort resets to **A–Z** and the search field clears — the next open starts fresh

**Active sort indicator:**
- [ ] When a non-default sort (**Newest first** or **Oldest first**) is active, the sort button has a visible indicator (e.g., a filled icon, colour change, or dot badge) so the user can see a custom sort is applied without opening the menu

---

### UI copy reference

| Element | Copy |
|---|---|
| Sort button tooltip | `Sort categories` |
| Sort option 1 | `A–Z` |
| Sort option 1 subtext | `Alphabetical order` |
| Sort option 2 | `Newest first` |
| Sort option 2 subtext | `Most recently created first` |
| Sort option 3 | `Oldest first` |
| Sort option 3 subtext | `Earliest created first` |

---

### Mock-ups:

N/A

---

### Impact on existing data:

None. Sort is a client-side operation over the already-loaded `getContentCategoryList` Vuex store data. No API calls, no schema changes, no data mutations.

---

### Impact on other products:

Chrome Extension has its own category dropdown — covered by a separate story: **[FE] Add sort filter button to content category dropdown (Chrome Extension)**.

---

### Dependencies:

None. Can be implemented in parallel with the Chrome Extension story.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry point:**
- `contentstudio-frontend/src/modules/publish/components/posting/social/ContentCategorySelection.vue` — the sort button goes into the same dropdown header `<div>` that already holds the `SearchInput`. The existing `sortContentCategory()` method does A-Z via `sortBy(data, item => [item.name.toLowerCase()])` — extend `filteredCategories` to branch on a new `sortOrder` ref.

**State shape suggestion:**
```ts
const sortOrder = ref<'az' | 'newest' | 'oldest'>('az')
```
Reset both `sortOrder` and `categorySearchText` to defaults in the existing `handleCategorySearchClose` handler (already called on dropdown `@close`).

**Sort logic per option:**
- A–Z: `sortBy(list, item => item.name.toLowerCase())` — existing `sortContentCategory()`, no change needed
- Newest first: `[...list].sort((a, b) => (b._id > a._id ? 1 : -1))` — MongoDB ObjectId is lexicographically sortable; a higher string value = more recently created document
- Oldest first: `[...list].sort((a, b) => (a._id > b._id ? 1 : -1))`

**UI components to use:**
- `Dropdown` + `DropdownItem` from `@contentstudio/ui` — for the sort options menu
- `ActionIcon` from `@contentstudio/ui` — for the icon-only sort trigger button
- `Icon` from `@contentstudio/ui` — suggested icon names: `ListFilter` or `ArrowUpDown` (verify against the design system icon set)

**Closest existing pattern:**
- `contentstudio-frontend/src/modules/publisher/components/CreateCustomView.vue` — defines a `sortOrders` array with `{ name, key, icon }` shape, renders a `DropdownItem` per option, and tracks the active selection with a ref. The checkmark pattern for the active item is already implemented there — reuse that approach.

**i18n keys to add** (in `publish.json` namespace, all locale directories under `src/locales/`):
```
publish.content_category.sort_button_tooltip   → "Sort categories"
publish.content_category.sort_az               → "A–Z"
publish.content_category.sort_az_subtext       → "Alphabetical order"
publish.content_category.sort_newest           → "Newest first"
publish.content_category.sort_newest_subtext   → "Most recently created first"
publish.content_category.sort_oldest           → "Oldest first"
publish.content_category.sort_oldest_subtext   → "Earliest created first"
```

---

## Story 2

**Title:** `[FE] Add sort filter button to content category dropdown (Chrome Extension)`

---

### Description:

As a ContentStudio user composing or scheduling posts via the Chrome Extension, I want to sort the content category list by name or by when a category was created so that I can quickly locate the right category — whether I remember it alphabetically or know it's one I recently added — without scrolling through an unsorted list.

---

### Workflow:

1. User opens the Chrome Extension and navigates to any screen that includes a content category dropdown (e.g., the post composer, scheduling flow, or any other screen where a category can be assigned to a post).
2. The content category dropdown header contains two controls side by side:
   - A **search field** (already present or added alongside this feature) for text-based filtering
   - A **sort icon button** (new) immediately to the right of the search field
3. The sort button displays a sort/filter icon (e.g., an up-down arrow or list-filter icon). On hover, a tooltip reads: **"Sort categories"**.
4. User clicks the sort button — a small dropdown menu opens directly below (or above, if near the bottom of the screen) the button, containing three options:

   | Option | Description shown in menu |
   |---|---|
   | **A–Z** | Alphabetical order |
   | **Newest first** | Most recently created categories appear at the top |
   | **Oldest first** | Earliest created categories appear at the top |

5. The currently active sort option (default: **A–Z**) is indicated with a checkmark or a highlighted background so the user always knows which order is active.
6. User clicks a sort option:
   - The sort menu closes.
   - The category list reorders immediately to reflect the chosen sort — no page reload or API call.
   - The sort button icon (or a small indicator beside it) updates to reflect the active sort so the user can tell at a glance that a non-default sort is applied.
7. If the user has an active search term in the search field, the sort is applied to the filtered results — not the full list. The two controls compose cleanly.
8. User selects a category from the sorted (and possibly filtered) list — the dropdown closes, the selected category is applied to the post, and the sort/search state resets.
9. When the dropdown is closed for any reason (category selected, clicked outside, Escape key), both the search text and the sort order reset to defaults (A–Z, empty search) so the next time the dropdown opens it starts fresh.

---

### Acceptance criteria:

**Sort button presence:**
- [ ] A sort icon button appears in the content category dropdown header on every screen in the Chrome Extension where a content category dropdown is shown
- [ ] The sort button is positioned immediately adjacent to the search field within the dropdown header area
- [ ] On hover, the sort button shows a tooltip: **"Sort categories"**
- [ ] When the category list is empty (no categories exist in the workspace), the sort button is hidden or disabled — there is nothing to sort

**Sort menu:**
- [ ] Clicking the sort button opens a menu with exactly three labelled options:
  - **"A–Z"** — with subtext or aria-label: *"Alphabetical order"*
  - **"Newest first"** — with subtext or aria-label: *"Most recently created first"*
  - **"Oldest first"** — with subtext or aria-label: *"Earliest created first"*
- [ ] The currently active sort option is visually distinguished with a checkmark icon or highlighted background
- [ ] Clicking outside the sort menu (without selecting an option) closes the menu without changing the active sort

**Sort behavior — A–Z:**
- [ ] When **A–Z** is selected, categories are ordered alphabetically by name, case-insensitively (e.g., "Blog" comes before "Social", "blog daily" and "Blog Weekly" are adjacent regardless of capitalisation)

**Sort behavior — Newest first:**
- [ ] When **Newest first** is selected, the most recently created category appears at the top of the list
- [ ] Categories created on the same day appear in a consistent sub-order (e.g., by name)

**Sort behavior — Oldest first:**
- [ ] When **Oldest first** is selected, the earliest created category appears at the top of the list
- [ ] Categories created on the same day appear in a consistent sub-order (e.g., by name)

**Sort + search composition:**
- [ ] When a search term is active and the user changes the sort, the filtered results re-sort immediately — the search filter is preserved
- [ ] When a sort is active and the user types in the search field, the displayed results are filtered from the full list and sorted according to the active sort order

**Reset behaviour:**
- [ ] When the dropdown closes (by category selection, clicking outside, or pressing Escape), the sort resets to **A–Z** and the search field clears — the next open starts fresh

**Active sort indicator:**
- [ ] When a non-default sort (Newest first or Oldest first) is active, the sort button has a visible indicator (e.g., a dot, filled icon, or colour change) so the user can tell without opening the menu that a custom sort is applied

---

### UI copy reference

| Element | Copy |
|---|---|
| Sort button tooltip | `Sort categories` |
| Sort option 1 | `A–Z` |
| Sort option 2 | `Newest first` |
| Sort option 3 | `Oldest first` |

---

### Mock-ups:

N/A — follow the same visual pattern as the Web App (see **[FE] Add sort filter button to content category dropdown (Web App)**). The sort button sits in the dropdown header row to the right of the search field. The sort menu uses the same item style as the category list items.

---

### Impact on existing data:

None. Sort is a client-side operation over the already-loaded category list. No API calls, no schema changes, no data mutations.

---

### Impact on other products:

Web App is covered by a separate story: **[FE] Add sort filter button to content category dropdown (Web App)**. No impact on iOS or Android apps.

---

### Dependencies:

None. This story is independent of the Web App story and can be worked in parallel.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A (Chrome Extension runs in a browser, not a mobile app)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Reference implementation (Web App):**
- The Web App's content category dropdown — `contentstudio-frontend/src/modules/publish/components/posting/social/ContentCategorySelection.vue` — implements the same sort feature. It uses `SearchInput` from `@contentstudio/ui`, a `sortOrder` ref, and a `filteredCategories` computed that both filters by search text and sorts by the active order. The Chrome Extension dev can mirror this pattern using whatever component primitives the Extension codebase provides.

**Sort logic:**
- A–Z: `sortBy(list, item => item.name.toLowerCase())` — lodash-es `sortBy` or equivalent
- Newest first: `[...list].sort((a, b) => (b._id > a._id ? 1 : -1))` — MongoDB ObjectId is lexicographically sortable; a higher string value means a more recently created document
- Oldest first: `[...list].sort((a, b) => (a._id > b._id ? 1 : -1))`

**State shape suggestion:**
```
sortOrder: 'az' | 'newest' | 'oldest'   // default: 'az'
```
Reset both `sortOrder` and `searchText` to defaults in the dropdown `@close` handler.

**Closest existing pattern in the Web App:**
- `contentstudio-frontend/src/modules/publisher/components/CreateCustomView.vue` — defines a `sortOrders` array (`first_created`, `last_created`, etc.) and renders `DropdownItem` per option with an active-state check mark. Adapt this pattern for the three options above.
