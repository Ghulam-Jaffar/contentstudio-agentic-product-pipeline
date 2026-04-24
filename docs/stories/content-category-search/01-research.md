# Research: Add Search Bar to Content Categories Dropdown in the Composer

## Current State

The Content Category selector in the Composer's account selection aside shows all categories as a flat scrollable list with no way to filter them. Users with many categories have to scroll to find the right one.

**Key file:** `contentstudio-frontend/src/modules/composer_v2/components/AccountSelectionAside.vue`

- Lines 48–148: `CstDropdown` component containing the category list
- `contentCategoryList` is a **prop** (array, passed in from parent — not locally owned)
- Items rendered via `v-for="(item, index) in contentCategoryList"` — no filtering applied
- Each item shows a colour swatch, category name, and an optional "Global" badge + crown icon

**Existing account search pattern (same file, lines 151–168):**
- Local data property `accountsSearchQuery: ''`
- Computed `sortedAccountsFlat` / `getChannelAccounts` filter by `accountsSearchQuery`
- Plain HTML `<input>` inside a styled wrapper with a search icon SVG and a clear button

The category search should follow this same approach: local `categorySearchQuery` data, a computed `filteredCategoryList` that filters by `item.name`, and a search input at the top of the dropdown body.

**i18n** — `src/locales/en/composer.json`, section `account_selection.content_category`:
```json
"content_category": {
  "title": "Content Category",
  "empty_state": "You don't have any content category",
  "create_category": "Create Category",
  "unselect_category": "Unselect Category",
  "global": "Global"
}
```
→ No `search_placeholder` or `no_results` keys yet. Need to add to all 8 locale files.

## What Needs to Change

- Add `categorySearchQuery: ''` to local component data in `AccountSelectionAside.vue`
- Add computed `filteredCategoryList` — filters `contentCategoryList` by `item.name.toLowerCase().includes(query)`
- Add search `<input>` inside the `CstDropdown` default slot, above the category items — only shown when `contentCategoryList.length > 0`
- Wire the `v-for` to use `filteredCategoryList` instead of `contentCategoryList`
- Add a "no results" message when `filteredCategoryList.length === 0` but `categorySearchQuery` is non-empty
- Reset `categorySearchQuery` to `''` when the dropdown closes (mirrors the accounts search UX)
- Add i18n keys `search_placeholder` and `no_results` under `account_selection.content_category` in all 8 locale files

## Files Involved

| File | Change |
|---|---|
| `contentstudio-frontend/src/modules/composer_v2/components/AccountSelectionAside.vue` | Add search input, `categorySearchQuery` data, `filteredCategoryList` computed |
| `contentstudio-frontend/src/locales/en/composer.json` | Add `content_category.search_placeholder`, `content_category.no_results` |
| `contentstudio-frontend/src/locales/de/composer.json` | Same |
| `contentstudio-frontend/src/locales/el/composer.json` | Same |
| `contentstudio-frontend/src/locales/es/composer.json` | Same |
| `contentstudio-frontend/src/locales/fr/composer.json` | Same |
| `contentstudio-frontend/src/locales/it/composer.json` | Same |
| `contentstudio-frontend/src/locales/pl/composer.json` | Same |
| `contentstudio-frontend/src/locales/zh/composer.json` | Same |
