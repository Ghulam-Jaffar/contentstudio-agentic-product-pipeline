# Research: Search Bar in Content Category Dropdown (Chrome Extension)

## Current State

The **Web App** already has the search bar implemented in `contentstudio-frontend/src/modules/publish/components/posting/social/ContentCategorySelection.vue`. It uses:
- `SearchInput` from `@contentstudio/ui` — an expandable search field inside the category dropdown header
- `categorySearchText` ref to filter the `getContentCategoryList` store data via a `filteredCategories` computed property
- Clear/close handlers (`handleCategorySearchClose`, `handleCategorySearchBlur`) that reset the search when the dropdown closes

The `ContentCategorySelection.vue` component is used across multiple contexts in the web app (publish, RSS automation, evergreen automation, bulk CSV automation).

The **Chrome Extension** is a separate codebase (not mounted in this pipeline repo). Based on the ticket description, the Chrome extension has its own content category dropdowns but **does not yet have a search bar** in those dropdowns.

## What Needs to Change

- Add a `SearchInput` (or equivalent search field) inside every content category dropdown in the Chrome Extension
- Filter the displayed category list in real time as the user types
- Clear the search when the dropdown closes
- Show a "No categories found" empty state when the search yields no results
- The UX pattern should match the Web App implementation (search field in the dropdown header, filtered list below)

## UX Reference

Web App pattern (already shipped): `SearchInput` component in the dropdown header area, expands when focused, filters the list inline. When no results match, shows an empty state message.

## Files Involved

- Chrome Extension codebase — content category dropdown component(s) (exact file paths to be confirmed by the Chrome extension dev team, as the codebase is not mounted here)
- `contentstudio-frontend/src/modules/publish/components/posting/social/ContentCategorySelection.vue` — reference implementation (Web App)
