# Research — Global Content Category: per-workspace access/visibility

## The request

An agency owner with 50 workspaces creates a **Global Content Category** but only wants it to appear in **5 of the 50** workspaces. The other 45 should have **no access or visibility** to that global category unless it is explicitly granted to them later. Today there is no way to choose — a global category is force-pushed into **every** workspace the creator belongs to.

## Current State

**What a "Content Category" is:** ContentStudio lets users group social accounts + recurring posting slots into reusable evergreen buckets ("publishing queues"). Two kinds exist, distinguished by the `category_state` field:
- **Local** (`category_state: 'local'`) — lives in a single workspace.
- **Global** (`category_state: 'global'`) — created once, mirrored across the creator's workspaces. Only a **super admin** (or a team-membership **admin**) can create one — gated by `isEnableGlobalCategoryOption`.

**Data model (MongoDB):**
- `global_content_categories` — the master record (`GlobalContentCategories`): `workspace_id` (where it was created), `user_id`, `name`, `color`. `hasMany` `ContentCategories` via `global_content_category_id`.
- `content_categories` — per-workspace category docs (`ContentCategories`): one is created **per workspace** for a global category, each linked back via `global_content_category_id`, with `category_state: 'global'` plus the per-platform account arrays and slots.

**Current creation behavior (the root cause):** In `GlobalContentCategoriesController::store()`, when a new global category is created, the controller fetches **all** workspaces the creator belongs to and writes a `content_categories` doc into every one of them:

```php
$workspace_ids = WorkspaceTeamRepo::getWorkspaceIdsByUserId(Auth::id()); // ALL the user's workspaces (excludes approver role)
foreach ($workspace_ids as $ws_id) { /* create a ContentCategories doc per workspace */ }
```

There is **no parameter to limit which workspaces** receive it. For the 50-workspace user, all 50 get it.

**Editing a global category** (payload carries `global_content_category_id`): only updates name/color across all linked docs and updates the channels for the **current** workspace's doc. It cannot add or drop workspaces.

**Deleting a global category:** removes the linked `content_categories` doc (+ slots + member-access cleanup) from every workspace, then deletes the master record.

**Important distinction — this is NOT the existing member-access feature.** The codebase already has `ContentCategoryAccessService` / `ContentCategoryAccessController` (the "Allowed Team Members" dropdown), which controls **which members _within_ a workspace** can see a category (stored in `WorkspaceTeam.permissions.content_categories`). That is member-level. **This request is workspace-level** — which of the creator's workspaces the global category exists in at all. The two are independent and both should keep working.

**Frontend today:**
- `AddCategory.vue` modal (logic in `useAddCategory.ts`): name, color, a local/global radio (global disabled unless `isEnableGlobalCategoryOption`), account selection, and the "Allowed Team Members" multi-select. When creating a **new global** category, the right column shows a **static warning list** (`global_category_warnings.*`) that literally says it will be created in **all** workspaces — this becomes inaccurate once scoping exists.
- `ContentCategories.vue` lists the active workspace's categories; globals show a Crown icon; edit/remove of globals is gated by `isEnableGlobalCategoryOption`.
- Store `useContentCategoryStore.ts` → `storeGlobalCategory()` posts to `global/categories/store`; `deleteGlobalCategory()` posts to `global/categories/delete`.
- The frontend **already has the candidate workspace list** in `useWorkspaceStore` (`getWorkspaces.items` — each with `_id`, `name`, logo), so a workspace picker needs no new "list workspaces" endpoint.

## What Needs to Change

**Backend (`global/categories/store` + a small read endpoint):**
- Accept a selected **`workspace_ids`** array in the store payload instead of always using every workspace the creator belongs to.
- **Create:** write the per-workspace `content_categories` doc only into the selected workspaces (still auto-filling that workspace's connected accounts as it does today).
- **Edit:** sync the selection — create the category in newly selected workspaces, **remove** it (doc + slots + member-access cleanup) from de-selected workspaces, and update name/color on the ones that stay.
- Guard so a creator can only target workspaces they actually have access to.
- Add a read endpoint that returns the workspace IDs a given global category currently lives in (so the edit modal can pre-select them) — derivable from `ContentCategoriesRepository::getByGlobalCategoryId()`.

**Frontend (`AddCategory.vue` + `useAddCategory.ts`):**
- Add an **"Apply to workspaces"** multi-select (reuse the existing Allowed-Team-Members `Dropdown` + `Checkbox` + `SearchInput` + select-all pattern in the same modal), shown whenever the category type is **Global**.
- Replace the now-inaccurate "created in all workspaces" warning copy.
- On **edit**, pre-select the workspaces the global category already lives in (from the new read endpoint).
- Warn before saving when the edit **removes** workspaces (that deletes their queue + slots).
- Send the selected `workspace_ids` to the backend.

**Product decision (confirmed):** On the **create** modal, the workspace picker defaults to **all of the creator's workspaces pre-selected** (today's behavior). The admin then **deselects** down to the few they want before saving. (Considered alternatives: current-workspace-only, or none-pre-selected — both rejected in favor of preserving current behavior.)

## Scope / impact

- **No mobile stories.** Global-category management is a web-only admin/settings flow. Mobile apps and the Chrome extension only *consume* the categories that exist in a given workspace; once scoping is applied they simply see fewer categories in non-selected workspaces — driven entirely by backend data, no app changes.
- **Not AI, not analytics-pipeline** — no `contentstudio-ai-agents` / `social-analytics-go` involvement.
- **Story split:** `[BE]` (workspace-scoped create/edit/sync + read endpoint) + `[FE]` (workspace picker UI + copy). 2 stories.

## Files Involved

**Backend**
- `contentstudio-backend/app/Http/Controllers/Settings/ContentCategories/GlobalContentCategoriesController.php` — `store()` (create/edit), `delete()`
- `contentstudio-backend/app/Repository/Settings/ContentCategoriesRepository.php` — `getByGlobalCategoryId()`, `removeByGlobalCategoryId()`, `createOrUpdate()`
- `contentstudio-backend/app/Repository/Settings/GlobalContentCategoriesRepository.php`
- `contentstudio-backend/app/Services/Settings/ContentCategoryAccessService.php` — `removeCategoryFromAllMembers()` (reuse on workspace removal)
- `contentstudio-backend/app/Repository/Settings/WorkspaceTeamRepo.php` — `getWorkspaceIdsByUserId()`
- `contentstudio-backend/routes/web/settings.php` — `global/categories/*` route group (~line 208)

**Frontend**
- `contentstudio-frontend/src/modules/setting/components/content-categories/dialogs/AddCategory.vue`
- `contentstudio-frontend/src/modules/setting/composables/useAddCategory.ts`
- `contentstudio-frontend/src/stores/setting/useContentCategoryStore.ts` — `storeGlobalCategory()`
- `contentstudio-frontend/src/api/content-categories.ts` — `storeGlobalCategory()`, plus a new "fetch workspaces for global category" call
- `contentstudio-frontend/src/modules/setting/config/api-utils.js` — URL definitions
- `contentstudio-frontend/src/stores/setting/useWorkspaceStore.ts` — `getWorkspaces.items` (candidate workspace list, already available)
