# Research — Global Content Category: Local vs Global Edit

## The requirement (clarified with user)

For a **global** content category in the Categories listing, the single edit (pencil) icon should become an **edit dropdown** with two options:

1. **Edit locally (this workspace)** → opens the **existing** Edit Category modal (name, color, Select Accounts, Allowed Team Members) and applies changes to the **current workspace only**.
2. **Edit globally (selected workspaces)** → opens a **new** modal modeled on the global *create* view but **without the Category Type selector** — it shows **Category Name**, **Choose Color**, and a **Select Workspaces** picker, and applies changes to the **selected workspaces**.

(Local categories keep their normal single edit pencil → existing modal. No change for them.)

> The user originally asked for two radio buttons inside the edit modal, then corrected the requirement to the edit-dropdown-with-two-modals design above.

---

## ⚠️ Important discrepancy: "Select Workspaces" is not in this checkout

The user's screenshot 2 shows the **Add Category** (global create) modal with a **"Select Workspaces"** dropdown ("16 workspaces selected") and copy reading *"Global content category will be created in **selected** workspaces…"*.

**The mounted codebase does not have this.** In the current code:
- The global-create branch of `AddCategory.vue` shows only a **warning box** — there is **no** workspace picker.
- The copy still reads *"created in **all** workspaces"* (`settings.add_category.global_category_warnings.workspace_creation`), and the backend creates the category in **every** workspace the user can access (`WorkspaceTeamRepo::getWorkspaceIdsByUserId`) — all-or-nothing, no selection.
- `grep` across `src/` finds **no** "Select Workspaces" / workspace-multiselect component anywhere.

So screenshot 2 reflects a **newer/design version** (or a parallel in-progress effort) that isn't in this checkout. **Implication:** the "Select Workspaces" picker for the "Edit globally" modal must be **built** as part of this work (or, if a create-flow effort ships it first, reused). This is flagged in the stories.

---

## Current State

### Where it lives
- **Listing + edit trigger:** `contentstudio-frontend/src/modules/setting/components/content-categories/ContentCategories.vue`. The global row currently renders an edit pencil (`icon-edit-cs`) and a delete icon; edit calls `editCategoryModal(category)` in `useContentCategories.ts`.
- **The modal (Add + Edit):** `contentstudio-frontend/src/modules/setting/components/content-categories/dialogs/AddCategory.vue` — title switches on `getContentCategoryAdd._id`; Category Type radios render only on create (`v-if="!getContentCategoryAdd._id"`).
- **Modal logic:** `contentstudio-frontend/src/modules/setting/composables/useAddCategory.ts` → `validateAndStoreCategory()` calls `storeCategory()` (local) or `storeGlobalCategory()` (global).
- **Store:** `contentstudio-frontend/src/stores/setting/useContentCategoryStore.ts` → `storeGlobalCategory()`.
- **API wrapper:** `contentstudio-frontend/src/api/content-categories.ts` → `storeGlobalCategory(workspaceId, category)` → POST `global/categories/store`.
- **Backend:** `contentstudio-backend/app/Http/Controllers/Settings/ContentCategories/GlobalContentCategoriesController.php` → `store()`.

### How a global edit behaves today (and why it's a problem)
When editing a global category, `store()` runs the branch where `global_content_category_id` is present:
1. **Name + color → applied to ALL workspaces** (`ContentCategoriesRepository::updateAllCategoriesNameAndColorById`).
2. **Accounts → applied to the CURRENT workspace only** (updates only the `$payload['_id']` category, and only `facebook, twitter, pinterest, linkedin, gmb, instagram`).

The scope is hardcoded and inconsistent, and the user has no control or visibility. The new design replaces this with two explicit, user-chosen paths (local vs global).

### Gaps / gotchas found
- The global-edit account update **omits** `threads, bluesky, telegram, youtube, tiktok` — those selections aren't persisted on a global edit today (pre-existing gap; fix alongside).
- `allowed_member_ids` is sent by the FE but the edit branch's `createOrUpdate` payload doesn't include it, so member-access changes may not persist through this endpoint (member access is otherwise managed via Team → category-access). Confirm during BE work.
- `GlobalContentCategories` model stores only `workspace_id, user_id, name, color`; the workspace membership is derived from `content_categories.global_content_category_id` (one linked category per workspace). "Select Workspaces" membership = which workspaces have a linked category.
- Backend building blocks already exist: `ContentCategoriesRepository::createOrUpdate` (add to a workspace), `removeById` / `removeByGlobalCategoryId` (remove), `getByGlobalCategoryId` (list linked), `updateAllCategoriesNameAndColorById` (propagate name/color).

---

## Target behavior

### Listing — edit dropdown (global categories only)
Replace the single edit pencil on a **global** category row with an edit action that opens a **`Dropdown`** menu with two **`DropdownItem`**s:
- **Edit locally (this workspace)** → opens the existing Edit Category modal, scoped to the current workspace.
- **Edit globally (selected workspaces)** → opens the new Edit Global Category modal.

### "Edit locally" modal (existing modal, scope = current workspace)
The existing modal (name, color, Select Accounts, Allowed Team Members). On save, changes apply **only to the current workspace's copy** — name/color no longer silently propagate to other workspaces.

### "Edit globally" modal (new, scope = selected workspaces)
Modeled on the global-create view (screenshot 2) **minus the Category Type selector**. Contains:
- **Category Name**
- **Choose Color**
- **Select Workspaces** (multi-select; pre-checked with the workspaces the global category currently belongs to)

On save:
- Update **name + color** across all selected workspaces.
- **Add** the category to newly-selected workspaces (create a linked category there, auto-selecting that workspace's accounts — same as create).
- **Remove** the category from de-selected workspaces (delete that workspace's linked category + its slots + member-access cleanup, same as the existing global delete path).

Account/member selection stays a **per-workspace** concern (handled in "Edit locally"), matching the create-flow model ("select/change accounts in each workspace individually after creation").

---

## UI components (from `docs/ui-components.md`)
- **Edit dropdown menu:** `Dropdown` + `DropdownItem` (already used in `AddCategory.vue`). Trigger via the existing edit icon / `ActionIcon`.
- **Select Workspaces picker:** `Dropdown` + `DropdownItem` + `Checkbox` + `SearchInput` — the exact pattern already used for Select Account(s) / Allowed Team Members in `AddCategory.vue`. Reuse it.
- **Modal:** `CstuModal` (as the existing modal uses) or `Modal`.
- No new component gaps — everything needed exists. The **Select Workspaces feature/data-flow** is the net-new part (not a new library component).

## Mobile Context
Not applicable. Content Category management is a **web-only** settings surface; no global-category editing exists in the iOS/Android apps. No mobile stories.

---

## Files Involved

**Frontend**
- `contentstudio-frontend/src/modules/setting/components/content-categories/ContentCategories.vue` — swap the global-row edit pencil for the edit `Dropdown` (two items)
- `contentstudio-frontend/src/modules/setting/composables/useContentCategories.ts` — `editCategoryModal` split into local vs global handlers
- `contentstudio-frontend/src/modules/setting/components/content-categories/dialogs/AddCategory.vue` — existing modal; ensure "Edit locally" saves current-workspace-only
- **New:** `.../content-categories/dialogs/EditGlobalCategory.vue` (name, color, Select Workspaces) + a `useEditGlobalCategory.ts` composable
- `contentstudio-frontend/src/modules/setting/composables/useAddCategory.ts` — local-scope save for a global category
- `contentstudio-frontend/src/stores/setting/useContentCategoryStore.ts` + `src/api/content-categories.ts` — pass scope / workspace list to the endpoint
- `contentstudio-frontend/src/locales/*/settings.json` — new copy keys (all locale dirs: de, el, en, es, fr, it, pl, zh)

**Backend**
- `contentstudio-backend/app/Http/Controllers/Settings/ContentCategories/GlobalContentCategoriesController.php` — support (a) local-only edit of a global category's copy and (b) global edit with name/color propagation + workspace add/remove
- `contentstudio-backend/app/Repository/Settings/ContentCategoriesRepository.php` — helpers for applying across / adding to / removing from linked workspaces

**Endpoint:** POST `global/categories/store` (existing, extended) — plus possibly a scoped variant for local-only edits.
