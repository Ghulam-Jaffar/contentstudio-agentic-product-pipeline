# Stories: Search Bar in Content Category Dropdown (Chrome Extension)

---

## Story 1

**Title:** `[FE] Add search bar to content category dropdowns in Chrome Extension`

---

### Description:

As a ContentStudio user composing posts via the Chrome Extension, I want a search bar inside the content category dropdown so that I can quickly find and select the right category without having to scroll through a long list.

---

### Workflow:

1. User opens the Chrome Extension and navigates to the post composer or any screen where a content category dropdown is available.
2. User clicks the content category dropdown — the dropdown panel opens, showing the full list of categories.
3. At the top of the dropdown, a search field is visible (or expands when the user clicks/focuses on the search icon).
4. User types part of a category name (e.g., "blog") — the list below filters in real time to show only matching categories.
5. User selects the desired category from the filtered results — the dropdown closes and the selected category is reflected in the field.
6. If the user types a term that matches no categories, an empty state message appears: **"No categories found"**.
7. When the dropdown closes (by selection, by clicking outside, or by pressing Escape), the search field resets to empty — the full category list is shown the next time the dropdown is opened.

---

### Acceptance criteria:

- [ ] Every content category dropdown in the Chrome Extension includes a search field at the top of the dropdown panel
- [ ] Typing in the search field filters the category list in real time, case-insensitively (e.g., typing "blog" matches "Blog Posts", "blog_weekly", etc.)
- [ ] The search field is cleared and the full category list is restored when the dropdown closes (whether by category selection, clicking outside, or pressing Escape)
- [ ] When the search query matches no categories, the list is replaced with the message: **"No categories found"**
- [ ] The search field placeholder text reads: **"Search categories..."**
- [ ] When there are no categories configured in the workspace at all (empty list), the existing empty state message is shown instead of the search field returning no results
- [ ] Selecting a category from the filtered list works identically to selecting from the unfiltered list — the category is applied and the dropdown closes
- [ ] The search field does not persist state between dropdown openings — each time the user opens the dropdown, search starts fresh

---

### Mock-ups:

N/A — follow the Web App pattern already implemented in `contentstudio-frontend/src/modules/publish/components/posting/social/ContentCategorySelection.vue`.

---

### Impact on existing data:

None. The search is a client-side filter over the already-loaded category list. No schema or API changes required.

---

### Impact on other products:

This change is scoped to the Chrome Extension only. The Web App already has this feature. No impact on iOS, Android, or Web App.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A (Chrome Extension is a browser extension, not a mobile app)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Reference implementation (Web App):**
- `contentstudio-frontend/src/modules/publish/components/posting/social/ContentCategorySelection.vue` — the exact same feature already shipped on the Web App. Uses `SearchInput` from `@contentstudio/ui` inside the `Dropdown` component header slot, with `categorySearchText` ref, `filteredCategories` computed (client-side filter), and `handleCategorySearchClose` / `handleCategorySearchBlur` handlers to reset on close.

**UI component:**
- `SearchInput` from `@contentstudio/ui` — confirmed available in the component catalog. Use `variant="filled"`, set an `expanded-width` appropriate to the dropdown panel width, and wire `@clear` to reset `categorySearchText`.

**Pattern to replicate:**
- Filter logic: `getContentCategoryList.filter(c => c.name.toLowerCase().includes(searchText.toLowerCase()))`
- On dropdown `@close` event → reset search text to `''`
- Show "No categories found" when filtered list is empty but `searchText` is non-empty; show the existing "No categories added" empty state when the list is empty and `searchText` is empty
