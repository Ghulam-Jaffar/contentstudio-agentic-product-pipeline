# Stories: Add Search Bar to Content Categories Dropdown in the Composer

---

## [FE] Add search bar to the Content Categories dropdown in the Composer

### Description:

As a user who has many Content Categories in their workspace, I want to search and filter the category list in the Composer so that I can quickly find and select the right category without scrolling through a long list.

**What needs to be built** in `contentstudio-frontend/src/modules/composer_v2/components/AccountSelectionAside.vue`:

1. Add `categorySearchQuery: ''` to local component data
2. Add a computed property `filteredCategoryList` that filters the `contentCategoryList` prop by `item.name` case-insensitively using the search query
3. Inside the `CstDropdown` default slot (above the category `v-for` items), render a search input — visible only when `contentCategoryList.length > 0`
4. Replace `v-for="(item, index) in contentCategoryList"` with `v-for="(item, index) in filteredCategoryList"` on the `CstDropdownItem` list
5. Add a "no results" message shown when `filteredCategoryList.length === 0` and `categorySearchQuery` is non-empty
6. Reset `categorySearchQuery` to `''` when the dropdown closes

The search input should mirror the existing account search input pattern already in this file (lines 151–168) — same styled wrapper with a search icon and a clear (×) button.

Add i18n keys `composer.account_selection.content_category.search_placeholder` and `composer.account_selection.content_category.no_results` to all 8 locale files under `src/locales/`.

---

### Workflow:

1. User opens the Composer
2. User clicks the **Content Category** dropdown in the account selection sidebar
3. The dropdown opens showing the full list of categories
4. A search input appears at the top of the dropdown list (below the "Unselect Category" action if one is selected)
5. User starts typing a category name (e.g., "Blog") — the list filters in real time to show only matching categories
6. User clicks the matching category to select it — the dropdown closes and the search query is cleared
7. If the user types a query with no matches, a "No categories found" message is shown instead of the list
8. The user can click the × button in the search input to clear the query and restore the full list

---

### Acceptance criteria:

- [ ] A search input is visible inside the Content Category dropdown when at least one category exists — it is not shown when the list is empty (the "empty state" / "Create Category" state)
- [ ] Typing in the search input filters the category list in real time, case-insensitively, matching against the category name
- [ ] The selected-category highlight (`bg-indigo-100` + checkmark) still applies correctly on filtered results
- [ ] When no categories match the search query, a "No categories found" message is displayed instead of the list — the "Unselect Category" action remains visible if a category is currently selected
- [ ] Clicking the × clear button resets the search query and restores the full category list
- [ ] When the dropdown closes (either by selecting a category or clicking outside), the search query is reset to empty
- [ ] The search input renders correctly in all 8 supported locales — no missing translation keys
- [ ] The "Unselect Category" action is always shown at the top (when applicable) and is not affected by the search filter
- [ ] The existing empty state ("You don't have any content category" + "Create Category" button) is unchanged and unaffected by this change

---

### Mock-ups:

N/A — The search input should match the existing account search pattern already in `AccountSelectionAside.vue` (lines 151–168): a light gray rounded wrapper, search icon on the left, plain text input, and a × clear button on the right when a query is active.

---

### UI Copy:

**Search input placeholder:**
> Search categories

**No results message** (shown when query returns 0 matches):
> No categories found

**i18n keys to add** under `composer.account_selection.content_category`:

| Key | `en` | `de` | `es` | `fr` | `it` | `pl` | `el` | `zh` |
|---|---|---|---|---|---|---|---|---|
| `search_placeholder` | Search categories | Kategorien suchen | Buscar categorías | Rechercher des catégories | Cerca categorie | Szukaj kategorii | Αναζήτηση κατηγοριών | 搜索类别 |
| `no_results` | No categories found | Keine Kategorien gefunden | No se encontraron categorías | Aucune catégorie trouvée | Nessuna categoria trovata | Nie znaleziono kategorii | Δεν βρέθηκαν κατηγορίες | 未找到类别 |

---

### Impact on existing data:

None. This is a purely client-side filter on data already fetched — no API changes, no new endpoints, no schema changes.

---

### Impact on other products:

- **Mobile apps (iOS / Android):** The Content Category selector in the mobile apps is a separate native component — not impacted by this change
- **Chrome extension:** Has its own composer implementation — not impacted
- **White-label:** No impact — no hardcoded colours introduced; the search input follows existing Tailwind utility patterns used elsewhere in the same file

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
